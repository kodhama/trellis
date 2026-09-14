package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestStalenessHook exercises the plugin's SessionStart staleness hook — decision-0039
// rule 1 (the surface is a SessionStart hook emitting additionalContext) as reworked by
// decision-0043 / kodhama-0007 slice 4 (#120), with the compared path moved by
// decision-0051 (the authority split): the check is a binary-free, git-free
// file-to-file comparison of the project's .trellis/internal/version stamp against the
// installed plugin's ${CLAUDE_PLUGIN_ROOT}/reference/version payload stamp. Silent
// when they match (or nothing is comparable); a refresh nudge when they differ. A
// stamp found only at the legacy flat path (.trellis/version — pre-decision-0051
// layouts, and before them the plugin@<sha>/CLI-semver stamps) always draws the
// nudge: the layout itself is stale, and a refresh is the migration vehicle. With
// `trellis status` retired (#120), this hook is the only user-facing drift surface
// (decision-0035: drift is made visible, not silent). Runs the real shell script
// against a temp "plugin root".
func TestStalenessHook(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(hook); err != nil {
		t.Fatalf("hook script missing: %v", err)
	}

	// The payload stamp the plugin actually ships (kept current by
	// TestVendoredPayloadIsCurrent); the hook compares against this file.
	shipped := payloadFile(t, "version")
	current := strings.TrimSpace(shipped)

	// A plain-directory plugin root — no git repo, on purpose: the file compare
	// must not depend on the plugin being a git checkout (decision-0043).
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginRoot, "reference", "version"), []byte(shipped), 0o644); err != nil {
		t.Fatal(err)
	}
	// The rest of the payload, copied from the real bundle rather than stubbed.
	// This fixture used to carry `version` alone, which silently disarmed every
	// assertion downstream of a payload read: the hook would bail with
	// TRELLIS_RULES_NOT_LOADED before it could inject anything, so a test
	// asserting "no rules were injected" passed for the wrong reason. Mutation
	// found it — making the announcing turn inject the full rule set left this
	// file green while a hand-run of the same mutation leaked twelve slugs.
	for _, f := range []string{"rules.md", "trellis.md", "invariants.md"} {
		src := filepath.Join(vendoredBundleAbs(t), "reference", f)
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("reading %s from the shipped bundle: %v", f, err)
		}
		if err := os.WriteFile(filepath.Join(pluginRoot, "reference", f), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// run executes the hook in a fresh project dir; stampRel names where the stamp
	// file is written (".trellis/internal/version", the legacy ".trellis/version",
	// or "" for no overlay at all).
	run := func(t *testing.T, stampRel, stamp string) string {
		proj := t.TempDir()
		if stampRel != "" {
			p := filepath.Join(proj, filepath.FromSlash(stampRel))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(stamp+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			// A real vendored overlay carries the payload files the managed
			// block imports, not just the stamp. The hook now checks they are
			// present before treating the old transport as healthy, so a
			// stamp-only fixture models a shape that cannot exist — and one
			// that, in the wild, is a silently broken install.
			if strings.Contains(stampRel, "internal") {
				for name, body := range map[string]string{
					"trellis.md": payloadFile(t, "trellis.md"),
					"rules.md":   payloadFile(t, "rules.md"),
				} {
					q := filepath.Join(filepath.Dir(p), name)
					if err := os.WriteFile(q, []byte(body), 0o644); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v) — a hook must never fail the session: %s", err, out)
		}
		return strings.TrimSpace(string(out))
	}

	nudge := func(t *testing.T, out string) string {
		t.Helper()
		if out == "" {
			t.Fatal("want a staleness message, got silence")
		}
		// The envelope must be nested. A bare top-level {"additionalContext": ...}
		// parses fine but is silently discarded by the host, so decoding the flat
		// shape here is what let the hook ship un-delivered: the test passed while
		// no model ever saw the nudge. Decode only the shape the host reads.
		var v struct {
			HookSpecificOutput struct {
				HookEventName     string `json:"hookEventName"`
				AdditionalContext string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal([]byte(out), &v); err != nil {
			t.Fatalf("output is not valid JSON: %v (%q)", err, out)
		}
		if v.HookSpecificOutput.HookEventName != "SessionStart" {
			t.Errorf("want nested SessionStart envelope, got %q", out)
		}
		// decision-0072 retired /trellis:setup, which every nudge used to name as
		// the remedy. The property this assertion actually guards is that a nudge
		// is ACTIONABLE — it must still say what to remove and what survives, or
		// it is a notification with no way out of the state it reports.
		ctx := v.HookSpecificOutput.AdditionalContext
		if !strings.Contains(ctx, "delete") || !strings.Contains(ctx, ".trellis/rules.toml") {
			t.Errorf("a nudge must name the manual migration — what to delete, and that rules.toml is kept: %q", ctx)
		}
		if strings.Contains(ctx, "/trellis:setup") {
			t.Errorf("nudge still points at the setup skill, retired by decision-0072: %q", ctx)
		}
		return v.HookSpecificOutput.AdditionalContext
	}

	// decision-0070 D4 replaced silence here with an announcement. A user-scoped
	// plugin in a project with no rules.toml used to say nothing at all, which
	// meant the developer never learned Trellis was about to govern the repo. It
	// now says so and offers the way out — and injects NO rules on that turn, so
	// "will be governed" stays a true statement rather than a fait accompli.
	t.Run("an unadopted project is told, and governed by nothing yet", func(t *testing.T) {
		out := run(t, "", "")
		if !strings.Contains(out, "TRELLIS_NOT_YET_GOVERNING") {
			t.Fatalf("decision-0070 D4: an unadopted project must be TOLD, not silently skipped; got %q", out)
		}
		if !strings.Contains(out, "governed = false") {
			t.Errorf("the announcement must name the exact way out, or declining is guesswork: %q", out)
		}
		if got := hookSlugs(out); len(got) > 0 {
			t.Errorf("no rule may be injected on the announcing turn — the message promises \"will be governed\", so governing already would make it false; found %v in %q", keysOfBool(got), out)
		}
		// AE6. The accept instruction quotes the file TRL-97 writes, byte for
		// byte, and nothing in it names strictness or a preset: an accepted
		// project holds only what has effect.
		ctx := nudgeContext(t, out)
		if got := quotedNewRulesFile(t, ctx); got != newRulesFile {
			t.Errorf("the accept instruction must quote the new file byte for byte\n got: %q\nwant: %q", got, newRulesFile)
		}
		for _, stale := range []string{"strictness", "preset", "rules-b.toml", "seeded_from", "active = true"} {
			if strings.Contains(ctx, stale) {
				t.Errorf("the announcement still names %q; a new file carries only what has effect:\n%s", stale, ctx)
			}
		}
		for _, want := range []string{"rules, followed by default", "the project is never governed"} {
			if !strings.Contains(ctx, want) {
				t.Errorf("the announcement lost %q:\n%s", want, ctx)
			}
		}
	})

	// D5: an explicit refusal outranks every default, and NOT GOVERNED MEANS NOT
	// GOVERNED — the two floor- rules go too. The floors are a floor on
	// CONFIGURATION (a row cannot dial a rule to zero while the project is
	// governed), not a claim on a project that declined to be governed at all.
	//
	// This comment previously argued the opposite, and sat directly above the
	// assertion that refutes it — left behind by a revert. Kept as a note rather
	// than deleted, because the boundary is genuinely easy to slide off: it was
	// gotten wrong here in both directions before it was gotten right.
	t.Run("governed = false injects nothing at all, floors included", func(t *testing.T) {
		if out := run(t, ".trellis/rules.toml", "governed = false"); out != "" {
			t.Errorf("not governed means NOT GOVERNED: no rule may be injected, the two floor- rules included; got %q", out)
		}
	})
	t.Run("current stamp at internal/version is silent", func(t *testing.T) {
		if out := run(t, ".trellis/internal/version", current); out != "" {
			t.Errorf("want silent, got %q", out)
		}
	})
	t.Run("older stamp at internal/version surfaces a refresh nudge", func(t *testing.T) {
		msg := nudge(t, run(t, ".trellis/internal/version", "payload@000000000000"))
		if !strings.Contains(msg, "payload@000000000000") || !strings.Contains(msg, current) {
			t.Errorf("message should name both stamps: %q", msg)
		}
	})
	t.Run("legacy flat-layout stamp surfaces a migration nudge", func(t *testing.T) {
		// A stamp at the pre-decision-0051 path means the overlay itself predates
		// the internal/ layout — the nudge fires even if the stamp text happens to
		// match the shipped payload, because the layout is what's stale.
		for _, stamp := range []string{current, "payload@000000000000", ""} {
			msg := nudge(t, run(t, ".trellis/version", stamp))
			// KTD10: an overlay this old may predate .trellis/rules.toml, and the
			// remedy names the new file rather than a preset.
			if strings.Contains(msg, "rules-b.toml") {
				t.Errorf("the legacy migration nudge still names the retired preset: %q", msg)
			}
			if got := quotedNewRulesFile(t, msg); got != newRulesFile {
				t.Errorf("the legacy migration nudge must quote the new file byte for byte\n got: %q\nwant: %q", got, newRulesFile)
			}
		}
	})
	t.Run("legacy plugin@sha stamp surfaces a refresh nudge", func(t *testing.T) {
		// Pre-#120 installs are stamped plugin@<short-sha> (decision-0039 rule 2,
		// superseded in part by decision-0043); they sit at the flat path, so the
		// one-time nudge migrates them onto the payload vocabulary and layout.
		nudge(t, run(t, ".trellis/version", "plugin@0000000"))
	})
	t.Run("legacy CLI-version stamp surfaces a refresh nudge", func(t *testing.T) {
		nudge(t, run(t, ".trellis/version", "0.2.16"))
	})
	t.Run("internal stamp wins over a leftover legacy file", func(t *testing.T) {
		// Mid-migration robustness: when both exist, the new layout's stamp is the
		// one compared — a current internal/version stays silent even if the old
		// flat file was not cleaned up.
		proj := t.TempDir()
		for rel, stamp := range map[string]string{
			".trellis/internal/version": current,
			".trellis/version":          "payload@000000000000",
		} {
			p := filepath.Join(proj, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(stamp+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		writeVendoredPayload(t, filepath.Join(proj, ".trellis", "internal"))
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v): %s", err, out)
		}
		if strings.TrimSpace(string(out)) != "" {
			t.Errorf("want silent (internal/version is current; the leftover flat file must not fire), got %q", out)
		}
	})
	// INVERTED by TRL-33. This subtest asserted the defect: it required SILENCE
	// from a broken vendored overlay, on the same path where a MISSING stamp had
	// always refused loudly. Absent-vs-empty was never chosen; it fell out of
	// `[ ! -f ]` for one and `[ -n "$overlay" ] || exit 0` for the other, and the
	// two sibling files in the same directory already refused loudly when empty.
	// Reproduced before changing anything: an empty .trellis/internal/version
	// produced 0 bytes of stdout at exit 0.
	t.Run("an empty overlay stamp is a broken overlay, said out loud", func(t *testing.T) {
		out := run(t, ".trellis/internal/version", "")
		ctx := nudgeContext(t, out)
		if !strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
			t.Errorf("an empty overlay stamp must refuse loudly, as a missing one already did:\n%s", ctx)
		}
		if !strings.Contains(ctx, ".trellis/internal/version") {
			t.Errorf("the refusal must name the file that could not be read:\n%s", ctx)
		}
	})
	// The legacy flat overlay is stale BY ITS LAYOUT, so the migration nudge is
	// correct whether or not either stamp can be read. It used to vanish instead.
	t.Run("an empty legacy stamp still draws the migration nudge", func(t *testing.T) {
		ctx := nudgeContext(t, run(t, ".trellis/version", ""))
		if !strings.Contains(ctx, "predates the .trellis/internal/ layout") {
			t.Errorf("a legacy overlay must still be told to migrate when its stamp cannot be read:\n%s", ctx)
		}
		if strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
			t.Errorf("a legacy overlay still governs the session — this is a migration nudge, not a blackout:\n%s", ctx)
		}
	})
	// INVERTED by TRL-34, which this subtest pinned: it required SILENCE when the
	// hook could not read the installed plugin's own reference/version, so the
	// staleness warning this hook exists to produce was withheld with no signal
	// at all. Reproduced against main before changing anything — mode 000 on that
	// file, a healthy vendored overlay present: 0 bytes of stdout at exit 0.
	//
	// The marker is deliberately NOT TRELLIS_RULES_NOT_LOADED. The overlay is
	// intact and the session IS governed by it; what is withheld is a warning,
	// not governance. Reusing the blackout marker would be the over-correction
	// this suite has already caught twice on this hook.
	t.Run("an unreadable plugin stamp says so instead of vanishing", func(t *testing.T) {
		// The REAL stamp, which is also what the project's overlay carries in
		// every case below. The corruption fixtures are built around it on
		// purpose: with the pre-fix `head -n1 | tr -d '[:space:]'` read, each
		// of the last four reduced to exactly this value, compared EQUAL to the
		// project's, and produced zero bytes of output. Hardcoding some other
		// hash would have made them compare unequal and draw a stale nudge — a
		// visible wrong answer instead of the silent one that actually shipped,
		// so the fixture would have tested the wrong failure.
		real := strings.TrimSpace(payloadFile(t, "version"))
		for _, tc := range []struct {
			name  string
			stamp string // "" means: do not create reference/ at all
		}{
			{"absent", ""},
			{"empty", "\n"},
			{"truncated to a short hash", "payload@abc\n"},
			{"a stamp from the pre-payload@ era", "plugin@abcdef123456\n"},
			{"a stamp with non-hex in it", "payload@zzzzzzzzzzzz\n"},
			// The four below are the P2 from review on #262. `head -n1` threw
			// away every byte after line 1 and `tr -d '[:space:]'` every space
			// INSIDE the stamp, both before any check ran, so a corrupted
			// payload stamp was accepted as authoritative. Reproduced at zero
			// bytes of stdout before the fix; pinned here so it cannot return.
			{"a valid stamp followed by garbage", real + "\nGARBAGE\n"},
			{"a valid stamp followed by a decoy stamp", real + "\npayload@ffffffffffff\n"},
			{"whitespace inside the stamp", "payload@" + real[8:12] + " " + real[12:] + "\n"},
			{"a stamp preceded by a garbage line", "GARBAGE\n" + real + "\n"},
			// The other direction, in the same table so it cannot be forgotten:
			// these are HEALTHY files an editor or a core.autocrlf=true
			// checkout produces, and refusing them would be the over-correction.
			// They must draw NO marker at all — the stamps match.
			{"CRLF line ending is healthy", real + "\r\n"},
			{"trailing blank lines are healthy", real + "\n\n\n"},
			{"trailing whitespace on the stamp line is healthy", real + "  \n"},
		} {
			healthy := strings.HasSuffix(tc.name, "healthy")
			t.Run(tc.name, func(t *testing.T) {
				proj := t.TempDir()
				p := filepath.Join(proj, ".trellis", "internal", "version")
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(payloadFile(t, "version")), 0o644); err != nil {
					t.Fatal(err)
				}
				writeVendoredPayload(t, filepath.Dir(p))
				broken := t.TempDir()
				if tc.stamp != "" {
					if err := os.MkdirAll(filepath.Join(broken, "reference"), 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(broken, "reference", "version"), []byte(tc.stamp), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				cmd := exec.Command(hook)
				cmd.Dir = proj
				cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+broken)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("hook exited non-zero (%v): %s", err, out)
				}
				raw := strings.TrimSpace(string(out))
				if healthy {
					// The stamps match, so silence is the CORRECT answer here —
					// and it is the one answer the corruption rows must never
					// get. Same table, opposite direction, so a guard that
					// over-tightens the stamp shape fails here rather than
					// reaching a consumer.
					if raw != "" {
						t.Fatalf("a %s reference/version is a healthy file and its stamp matches — refusing or nudging over it is the over-correction:\n%s", tc.name, raw)
					}
					return
				}
				ctx := nudgeContext(t, raw)
				if !strings.Contains(ctx, "TRELLIS_STALENESS_UNKNOWN") {
					t.Errorf("an unusable plugin stamp must say staleness could not be checked, not vanish:\n%s", ctx)
				}
				if strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
					t.Errorf("the overlay is intact and governs this session — this must not read as a blackout:\n%s", ctx)
				}
				if strings.Contains(ctx, "may be stale") {
					t.Errorf("nothing may be reported STALE against a stamp that could not be read:\n%s", ctx)
				}
			})
		}
	})

	// Path B — plugin-native delivery. A project that carries only the config
	// file gets the rules injected from the plugin's own payload. This is the
	// mode that lets a consumer stop vendoring `.trellis/internal/` entirely.
	t.Run("config-only project receives the rules from the plugin", func(t *testing.T) {
		full := t.TempDir()
		if err := os.MkdirAll(filepath.Join(full, "reference"), 0o755); err != nil {
			t.Fatal(err)
		}
		for name, body := range payloadFiles() {
			if err := os.WriteFile(filepath.Join(full, "reference", name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		proj := t.TempDir()
		toml := filepath.Join(proj, ".trellis", "rules.toml")
		if err := os.MkdirAll(filepath.Dir(toml), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(toml, []byte(configOnlyProjectRules), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+full)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v): %s", err, out)
		}
		ctx := nudgeContext(t, strings.TrimSpace(string(out)))

		// The always-loaded chain, assembled at runtime instead of via imports.
		if strings.Contains(ctx, "@rules.md") {
			t.Error("the @rules.md import must be resolved, not passed through")
		}
		for _, want := range []string{
			"How to work in this project", // posture header
			"inv-directional-flow",        // a rule body from rules.md
			"[rules]",                     // the project's own file, echoed
			"floor-intent-gate",
		} {
			if !strings.Contains(ctx, want) {
				t.Errorf("injected context missing %q", want)
			}
		}
		if !governedSwitchingOff(ctx, configOnlyProjectRow) {
			t.Errorf("the computed sentence must name the one rule the project file switches off:\n%s", ctx)
		}
		// The vendored path does not exist in this mode; the pointer must name
		// the plugin's copy, which is the payload this session is running.
		if strings.Contains(ctx, ".trellis/internal/invariants.md") {
			t.Error("invariants pointer still names the vendored path")
		}
		if !strings.Contains(ctx, filepath.Join(full, "reference", "invariants.md")) {
			t.Error("invariants pointer does not name the plugin's copy")
		}
	})

	// The two paths are mutually exclusive: a vendored project must never also
	// receive the rules, or every session would carry them twice.
	t.Run("vendored project never also receives the rules", func(t *testing.T) {
		full := t.TempDir()
		if err := os.MkdirAll(filepath.Join(full, "reference"), 0o755); err != nil {
			t.Fatal(err)
		}
		for name, body := range payloadFiles() {
			if err := os.WriteFile(filepath.Join(full, "reference", name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		proj := t.TempDir()
		for rel, body := range map[string]string{
			".trellis/internal/version": "payload@stale\n",
			".trellis/rules.toml":       newRulesFile,
		} {
			p := filepath.Join(proj, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+full)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v): %s", err, out)
		}
		ctx := nudgeContext(t, strings.TrimSpace(string(out)))
		if strings.Contains(ctx, "inv-directional-flow") {
			t.Error("vendored project received the rule bodies as well as the nudge")
		}
	})

	// A current stamp is not proof the overlay can load. Deleting a payload file
	// leaves the managed block importing something that is not there, and the
	// stamp says nothing about it — this project was silently ungoverned.
	t.Run("current stamp with a gutted payload fails loudly", func(t *testing.T) {
		proj := t.TempDir()
		dir := filepath.Join(proj, ".trellis", "internal")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "version"), []byte(shipped), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "trellis.md"), []byte(payloadFile(t, "trellis.md")), 0o644); err != nil {
			t.Fatal(err)
		}
		// rules.md deliberately absent.
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v): %s", err, out)
		}
		ctx := nudgeContext(t, strings.TrimSpace(string(out)))
		if !strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
			t.Errorf("want a loud failure for a gutted overlay, got %q", ctx)
		}
		if !strings.Contains(ctx, "rules.md") {
			t.Error("the failure must name the missing file")
		}
	})
}

// nudgeContext decodes the nested SessionStart envelope the host actually reads.
func nudgeContext(t *testing.T, out string) string {
	t.Helper()
	if out == "" {
		t.Fatal("want injected context, got silence")
	}
	var v struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("output is not valid JSON: %v (%q)", err, out)
	}
	if v.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Fatalf("want nested SessionStart envelope, got %q", out)
	}
	return v.HookSpecificOutput.AdditionalContext

}

// writeVendoredPayload gives a fixture the payload files a real vendored overlay
// carries. The hook checks they exist before trusting the stamp, so a stamp-only
// fixture models a shape that cannot exist in the wild.
func writeVendoredPayload(t *testing.T, internalDir string) {
	t.Helper()
	for _, name := range []string{"trellis.md", "rules.md"} {
		if err := os.WriteFile(filepath.Join(internalDir, name), []byte(payloadFile(t, name)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// decision-0068 D10 / spec-0005 AC2b. The install path renders
// `.claude/rules/trellis.md`, which Claude Code loads at launch on its own. If
// the plugin is ALSO present its hook would inject the same rules a second time
// — measured, not predicted: both present delivers the rule bodies twice, once
// in the project-instructions block and once in additionalContext.
//
// The discriminator is the FILE, not the directory holding it. `.claude/rules/`
// is a shared directory any project may use for unrelated rules; only
// `trellis.md` inside it means Trellis is already delivered. That is the mirror
// of decision-0065's argument for the vendored overlay, where the DIRECTORY is
// the artifact and the file inside it is not.
// renderedFile builds the fixture from the SHIPPED PAYLOAD BYTES, not from Go
// string literals. The literal version drifted immediately — it emitted only the
// H1 where install.sh emits the header's whole head, so the fixture asserted a
// footer referring to "the posture sentence above" over a file that had none —
// and no test could notice, because the fixture was the only definition of
// correct. Derived from the same payload install.sh reads, it drifts only if the
// payload does, which is the point. The footer carries framing only (TRL-97):
// the activation rule lives once, in rules.md.
func renderedFile(t *testing.T, stamp string) string {
	t.Helper()
	head, tail, _ := strings.Cut(payloadFile(t, "trellis.md"), "@rules.md\n")
	return "<!-- trellis:rendered-begin -->\n" + head + payloadFile(t, "rules.md") + tail +
		"<!-- trellis:rendered-footer -->\n" +
		"\n## Project rule activation\n\n@../../.trellis/rules.toml\n" +
		"\n<!-- trellis:rendered-from " + stamp + " -->\n"
}

// TestEveryDestructiveInstructionIsGated: a Codex P2 on #227, and then a Codex
// P2 on the GUARD ITSELF, which is the more useful of the two.
//
// staleness.sh's emit strings are injected straight into the agent's context, so
// "delete .trellis/internal/ and the managed block" is an instruction an
// autonomous agent can act on immediately, against tracked files. The retired
// /trellis:setup offered exactly this migration behind a confirmation
// (floor-intent-gate). Retiring the skill silently retired the gate with it.
//
// The first version of this guard matched the literal word "delete" — and a
// remedy saying "drop the unknown ones" slipped past it, instructing the removal
// of a consumer-owned row with no confirmation, while the reseed remedy two
// clauses later WAS gated for exactly that risk. A guard that recognises one
// verb is a guard against one verb.
//
// So the verb list below is the weak point, and is written to fail loudly rather
// than quietly: if it ever matches fewer messages than it does today, the
// filter has broken and the test says so instead of passing on an empty set.
var slashCommandRe = regexp.MustCompile(`/trellis:[a-z-]+`)

var destructiveVerbs = []string{
	"delete", "drop", "remove", "overwrite", "replace", "reset", "discard", "rm ",
}

// payloadAssembly returns the shell source of the payload="$( ... )" block in
// body. Both destructive-instruction guards below scan it: the rule activation
// framing and the rules.toml warnings ride into the agent's context through
// printfs inside this block rather than through emit "...", so a guard that
// reads only emit strings never sees them. The safety argument for shipping
// them ungated is "no deletion verb reaches the agent" — that argument only
// holds if every channel reaching the agent is enforced, not just one.
func payloadAssembly(t *testing.T, body string) string {
	t.Helper()
	start := strings.Index(body, `payload="$(`)
	if start < 0 {
		t.Fatal("payload=\"$(...)\" assembly not found in staleness.sh — the scan is broken")
	}
	rest := body[start:]
	end := strings.Index(rest, "\n)\"")
	if end < 0 {
		t.Fatal("payload=\"$(...)\" assembly has no closing )\" — the scan is broken")
	}
	return rest[:end]
}

// quotedSpanRe matches one bash-quoted span, single or double. staleness.sh's
// payload printf strings sometimes carry a literal apostrophe by splitting
// across three adjacent spans — 'This project'"'"'s ...' is the concatenation
// 'This project' + "'" + 's ...', bash's standard trick for embedding a
// single quote inside a single-quoted string. Concatenating every span's
// inner text, in source order, reconstructs the literal message exactly,
// that trick included.
var quotedSpanRe = regexp.MustCompile(`'[^']*'|"[^"]*"`)

// payloadPrintfMessages extracts the literal text of every `printf '...'`
// call in block (normally payloadAssembly's output), reconstructed from its
// quoted spans. A passthrough like `printf '%s\n' "$rules_prose"` reconstructs
// to something like "%s\n$rules_prose" — inert for verb-scanning, so it is not
// filtered out specially; what it carries at runtime is either payload prose or
// the project's own file, which the hook echoes rather than instructs.
func payloadPrintfMessages(block string) []string {
	var msgs []string
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "printf ") {
			continue
		}
		spans := quotedSpanRe.FindAllString(trimmed, -1)
		if len(spans) == 0 {
			continue
		}
		var sb strings.Builder
		for _, s := range spans {
			sb.WriteString(s[1 : len(s)-1])
		}
		msgs = append(msgs, sb.String())
	}
	return msgs
}

// ungatedDestructiveMessages scans msgs against destructiveVerbs and reports,
// for each message that instructs one, whether it carries the confirmation
// gate the message is required to. gated is every message that hit a verb
// (the running total TestEveryDestructiveInstructionIsGated's `known` floor
// pins); violations is the subset of those missing "explicit confirmation".
// Factored out so the same scan the real script runs through can also run
// against a deliberately mutated copy of it — see
// TestEveryDestructiveInstructionIsGated's "the payload channel is actually
// enforced" subtest, which proves this scan catches an ungated destructive
// message injected into the payload printf channel, not merely counts
// messages that channel happens to contribute today.
func ungatedDestructiveMessages(msgs []string) (violations []string, gated int) {
	for _, msg := range msgs {
		// Pointing at a slash command is not instructing a mutation: /trellis:remove
		// is a skill that runs its own confirmation. Scanning the raw text matched
		// its NAME and demanded a gate on a message that only names it, which would
		// have taught the next reader that the guard cries wolf.
		scan := strings.ToLower(slashCommandRe.ReplaceAllString(msg, " "))
		hit := ""
		for _, v := range destructiveVerbs {
			if strings.Contains(scan, v) {
				hit = v
				break
			}
		}
		if hit == "" {
			continue
		}
		gated++
		if !strings.Contains(msg, "explicit confirmation") {
			violations = append(violations, msg)
		}
	}
	return violations, gated
}

// codexQuotedSpanRe matches one JS-quoted span — single, double, or template
// literal (backtick) — the JS counterpart of quotedSpanRe for
// codex-context.mjs. JS template literals delimit with backticks (bash has
// no such form), so a third alternative is needed here; the scanned functions'
// text never contains a literal backtick, so this naive (no escape handling)
// span is exact for that source.
var codexQuotedSpanRe = regexp.MustCompile("'[^']*'|\"[^\"]*\"|`[^`]*`")

// codexPayloadAssembly returns the JS source of the codex-context.mjs functions
// named in codexPayloadFunctions — the Codex-side counterpart to
// payloadAssembly's shell payload="$( ... )" region, so the "no deletion verb
// reaches the agent" argument staleness.sh's guards enforce covers this channel
// too. Every function in codex-context.mjs that concatenates literal text into
// the agent's context is named here. TRL-97 retired the two that carried the
// repair mandate (repairMandate and provenanceOmittedNotice) with the
// reconciler: activationSection now builds the framing lines and the too-large
// line, and ruleWarnings the rules.toml warnings. A new channel into the agent's
// context that these guards cannot see is precisely the hole they exist to close.
var codexPayloadFunctions = []string{"activationSection", "ruleWarnings"}

func codexPayloadAssembly(t *testing.T, body string) string {
	t.Helper()
	var blocks []string
	for _, name := range codexPayloadFunctions {
		marker := "\nfunction " + name + "("
		start := strings.Index(body, marker)
		if start < 0 {
			t.Fatalf("%s(...) not found in codex-context.mjs — the scan is broken", name)
		}
		rest := body[start:]
		end := strings.Index(rest, "\n}\n")
		if end < 0 {
			t.Fatalf("%s(...) has no closing '}' — the scan is broken", name)
		}
		blocks = append(blocks, rest[:end])
	}
	return strings.Join(blocks, "\n")
}

// codexPayloadMessages extracts the literal text of every quoted-string line
// in block (normally codexPayloadAssembly's output), reconstructed from its
// quoted spans — the JS counterpart of payloadPrintfMessages. A comment line
// is skipped; every other line carrying a quoted span is a fragment of the
// mandate text codex-context.mjs concatenates into the agent's context, so
// each is scanned as its own message, the same granularity as one bash
// printf call. An interpolation like `${stamp}` reconstructs literally (as
// the text "${stamp}"), which is inert for verb-scanning — the same
// treatment payloadPrintfMessages gives a bash `"$rules_prose"` passthrough.
func codexPayloadMessages(block string) []string {
	var msgs []string
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		spans := codexQuotedSpanRe.FindAllString(trimmed, -1)
		if len(spans) == 0 {
			continue
		}
		var sb strings.Builder
		for _, s := range spans {
			sb.WriteString(s[1 : len(s)-1])
		}
		msgs = append(msgs, sb.String())
	}
	return msgs
}

func TestEveryDestructiveInstructionIsGated(t *testing.T) {
	body, err := os.ReadFile("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	emits := regexp.MustCompile(`(?m)^\s*emit "((?:[^"\\]|\\.)*)"`).FindAllStringSubmatch(string(body), -1)
	if len(emits) < 8 {
		t.Fatalf("found only %d emit strings — the scan is broken, and a guard that reads nothing passes", len(emits))
	}
	var msgs []string
	for _, m := range emits {
		msgs = append(msgs, m[1])
	}
	payloadMsgs := payloadPrintfMessages(payloadAssembly(t, string(body)))
	if len(payloadMsgs) < 5 {
		t.Fatalf("found only %d payload printf messages — the scan is broken, and a guard that reads nothing passes", len(payloadMsgs))
	}
	msgs = append(msgs, payloadMsgs...)
	// Fix round 2: the "actually enforced" subtest below recomputes its own
	// mutated payloadMsgs independently (payloadAssembly -> payloadPrintfMessages
	// -> ungatedDestructiveMessages on a copy of body), so it proves those
	// helpers work but never exercises the append above — the actual data flow
	// the scan below runs on. The re-reviewer proved the gap by mutation:
	// replacing the append with `_ = payloadMsgs` left `gated` unchanged (12,
	// since none of the real payload messages carry a verb) and the subtest
	// still green; only a log line moved from 28 to 19 messages, with nothing
	// asserting on it. This check ties the guard directly to msgs, the exact
	// slice ungatedDestructiveMessages is about to scan: every payload message
	// computed above must actually be IN it, or the payload channel is
	// computed but not wired into what this test asserts on.
	for _, pm := range payloadMsgs {
		found := false
		for _, m := range msgs {
			if m == pm {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("a payload printf message was computed but never entered the scanned set (msgs) — "+
				"the payload channel is computed but not wired into this guard's assertions:\n%s", pm)
		}
	}
	// codex-context.mjs's framing and rules.toml warnings are the Codex
	// counterpart of the payload="$( ... )" channel above. Scanned the same way,
	// wired the same way, for the same reason: the safety argument for shipping
	// them ungated only holds if this channel is actually enforced, not merely
	// present.
	codexBody, err := os.ReadFile("../plugins/trellis/hooks/codex-context.mjs")
	if err != nil {
		t.Fatal(err)
	}
	codexMsgs := codexPayloadMessages(codexPayloadAssembly(t, string(codexBody)))
	if len(codexMsgs) < 5 {
		t.Fatalf("found only %d codex payload messages — the scan is broken, and a guard that reads nothing passes", len(codexMsgs))
	}
	msgs = append(msgs, codexMsgs...)
	for _, cm := range codexMsgs {
		found := false
		for _, m := range msgs {
			if m == cm {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("a codex payload message was computed but never entered the scanned set (msgs) — "+
				"the codex payload channel is computed but not wired into this guard's assertions:\n%s", cm)
		}
	}
	violations, gated := ungatedDestructiveMessages(msgs)
	for _, msg := range violations {
		t.Errorf("this message instructs a mutation with no confirmation gate — an autonomous "+
			"agent can act on it against files the consumer owns (floor-intent-gate):\n%s", msg)
	}
	// A floor, not a ceiling: the count only ever grows as remedies are added, so
	// a drop means the regex or the verb list stopped matching, not that the
	// script got safer. Advanced 11 → 13 when decision-0073 D2 added the two
	// inline-shape messages (the S4 refusal and the inline+rendered conflict).
	// Retreated 13 → 12 when TRL-20 retired the gated mismatch remedy ("for
	// unknown:, remove those rows; for duplicate:, delete the extra
	// occurrences"). Unmoved by TRL-97: the payload printf channel now carries
	// the rule activation framing and the rules.toml warnings, none of which asks
	// for the file to change, so none carries a verb in the list.
	const known = 12
	if gated < known {
		t.Fatalf("matched %d destructive messages, expected at least %d — the filter broke; "+
			"a guard that matches nothing passes silently", gated, known)
	}
	t.Logf("checked %d destructive messages of %d (emit + payload printf)", gated, len(msgs))

	// Fix round 1, finding 2: `known` above never moves on the payload
	// channel's account, because every real message there is clean by design
	// (Ruling B). That left this test unable to tell "the payload channel has
	// nothing to gate" apart from "the payload channel stopped being scanned"
	// — payloadAssembly binding the wrong region, or payloadMsgs silently
	// dropping out of msgs, would both still leave `gated` at exactly `known`.
	// This subtest gives the scan a positive case: inject a synthetic,
	// UNGATED destructive instruction into a copy of the real script's
	// payload printf text and run it through the identical
	// payloadAssembly -> payloadPrintfMessages -> ungatedDestructiveMessages
	// pipeline the scan above uses. If that pipeline is broken in any of the
	// ways above, this injected message is never reached and the subtest
	// passes on nothing — the same failure mode this whole finding is about
	// — so the premise check below is the load-bearing part: it fails loudly
	// if the mutation itself did not land.
	//
	// One subtest per kind of payload printf, re-anchored by TRL-97 when the
	// repair mandate they used to mutate retired: the computed sentence, and a
	// warning template (KTD15).
	for _, tc := range []struct {
		name, marker, mutation, injected string
	}{{
		name:     "the payload channel is actually enforced, not merely counted",
		marker:   `printf 'The project file .trellis/rules.toml switches no rule off, so every rule above applies.\n'`,
		mutation: `printf 'Delete the unknown rows now. The project file .trellis/rules.toml switches no rule off, so every rule above applies.\n'`,
		injected: "Delete the unknown rows now",
	}, {
		name:     "the payload warning templates are actually enforced, not merely counted",
		marker:   `is a later row for %s, so it is ignored;`,
		mutation: `is a later row for %s; remove it now, so it is ignored;`,
		injected: "remove it now",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			mutated := strings.Replace(string(body), tc.marker, tc.mutation, 1)
			if mutated == string(body) {
				t.Fatalf("premise: %q was not found in staleness.sh — the case would prove nothing", tc.marker)
			}
			mutatedMsgs := payloadPrintfMessages(payloadAssembly(t, mutated))
			mutatedViolations, _ := ungatedDestructiveMessages(mutatedMsgs)
			found := false
			for _, v := range mutatedViolations {
				if strings.Contains(v, tc.injected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("an ungated instruction (%q) injected into the payload printf channel was not caught — "+
					"the payload channel is not actually being scanned, only its message count is being checked", tc.injected)
			}
		})
	}

	// The Codex counterparts of the subtest above: prove codexPayloadAssembly ->
	// codexPayloadMessages -> ungatedDestructiveMessages actually catches an
	// injected ungated instruction, rather than merely counting
	// codex-context.mjs's messages without scanning them. One per scanned
	// function, re-anchored by TRL-97 when the repair mandate they used to mutate
	// retired: the framing sentence in activationSection, and a warning template
	// in ruleWarnings (KTD15).
	for _, tc := range []struct {
		name, marker, mutation, injected string
	}{{
		name:     "the codex payload channel is actually enforced, not merely counted",
		marker:   `"The project file .trellis/rules.toml switches no rule off, so every rule above applies."`,
		mutation: `"Delete the unknown rows now. The project file .trellis/rules.toml switches no rule off, so every rule above applies."`,
		injected: "Delete the unknown rows now",
	}, {
		name:     "the codex warning templates are actually enforced, not merely counted",
		marker:   "is a later row for ${name}, so it is ignored;",
		mutation: "is a later row for ${name}; remove it now, so it is ignored;",
		injected: "remove it now",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			mutated := strings.Replace(string(codexBody), tc.marker, tc.mutation, 1)
			if mutated == string(codexBody) {
				t.Fatalf("premise: %q was not found in codex-context.mjs — the case would prove nothing", tc.marker)
			}
			mutatedMsgs := codexPayloadMessages(codexPayloadAssembly(t, mutated))
			mutatedViolations, _ := ungatedDestructiveMessages(mutatedMsgs)
			found := false
			for _, v := range mutatedViolations {
				if strings.Contains(v, tc.injected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("an ungated instruction (%q) injected into codex-context.mjs was not caught — "+
					"the codex payload channel is not actually being scanned, only its message count is being checked", tc.injected)
			}
		})
	}
}

// TestEveryDeletionInstructionIsGated: a Codex P2 on #227, and the SIXTH
// appearance of one class on this PR — every finding here has been a remedy that
// told an agent to do something destructive or shape-wrong without a gate.
//
// staleness.sh's emit strings are injected straight into the agent's context, so
// "delete .trellis/internal/ and the managed block" is an instruction an
// autonomous agent can act on immediately, against TRACKED files. The retired
// /trellis:setup offered exactly this migration and required confirmation
// (floor-intent-gate). Retiring the skill silently retired the gate with it.
//
// This is a source-level check on purpose. A per-branch behavioural test would
// pin the six remedies that exist today; the defect is that a SEVENTH can be
// added without a gate, so the guard reads every emit string in the script —
// and, since Ruling B (TRL-20 Task 2), the payload="$( ... )" assembly's
// printf strings too, the other channel that reaches the agent's context.
func TestEveryDeletionInstructionIsGated(t *testing.T) {
	body, err := os.ReadFile("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	// Each emit "..." payload, which is what reaches the agent.
	emits := regexp.MustCompile(`(?m)^\s*emit "((?:[^"\\]|\\.)*)"`).FindAllStringSubmatch(string(body), -1)
	if len(emits) < 8 {
		t.Fatalf("found only %d emit strings — the scan is broken, and a guard that reads nothing passes", len(emits))
	}
	var msgs []string
	for _, m := range emits {
		msgs = append(msgs, m[1])
	}
	payloadMsgs := payloadPrintfMessages(payloadAssembly(t, string(body)))
	if len(payloadMsgs) < 5 {
		t.Fatalf("found only %d payload printf messages — the scan is broken, and a guard that reads nothing passes", len(payloadMsgs))
	}
	msgs = append(msgs, payloadMsgs...)
	// Scan codex-context.mjs's framing and warnings the same way — see the
	// sibling wiring in TestEveryDestructiveInstructionIsGated for why.
	codexBody, err := os.ReadFile("../plugins/trellis/hooks/codex-context.mjs")
	if err != nil {
		t.Fatal(err)
	}
	codexMsgs := codexPayloadMessages(codexPayloadAssembly(t, string(codexBody)))
	if len(codexMsgs) < 5 {
		t.Fatalf("found only %d codex payload messages — the scan is broken, and a guard that reads nothing passes", len(codexMsgs))
	}
	msgs = append(msgs, codexMsgs...)
	gated := 0
	for _, msg := range msgs {
		// Fix round 1 (TRL-30 task 3): case-insensitive, matching the same
		// sibling scan in TestEveryDestructiveInstructionIsGated
		// (plugin_hook_test.go:596) and the payload/codex-payload-specific
		// loops below. Pre-existing hole, on the Claude emit channel, closed
		// while here: a case-sensitive check here is exactly what let a
		// capitalized "Delete the unknown rows..." land undetected on the
		// payload channel before that was fixed (see the comment on the
		// payload-loop check below) — this is the same defect class on the
		// channel that scan didn't cover.
		if !strings.Contains(strings.ToLower(msg), "delete") {
			continue
		}
		gated++
		if !strings.Contains(msg, "explicit confirmation") {
			t.Errorf("this message instructs a deletion with no confirmation gate — an autonomous "+
				"agent can act on it against tracked files (floor-intent-gate):\n%s", msg)
		}
	}
	if gated == 0 {
		t.Fatal("no deletion-instructing message was found at all — the filter is wrong, not the script")
	}
	// The payload printf messages must contribute ZERO deletion hits —
	// that is Ruling B's whole point, enforced here rather than merely argued in
	// a comment. If this ever fires, a deletion verb reached the agent through
	// the one channel that must stay ungated; the fix is to reword the printf,
	// never to weaken this guard.
	//
	// Lowercased before matching, same as the sibling scan in
	// TestEveryDestructiveInstructionIsGated (plugin_hook_test.go:596): a
	// case-sensitive check here was proven (by mutation, fix round 1) to let a
	// capitalized "Delete the unknown rows once you get explicit confirmation,
	// then write …" land in this channel undetected. Unlike the main loop
	// above, there is no confirmation-clause exception — a gate phrase does
	// not make a deletion verb acceptable HERE, because the payload is text
	// injected into the session, not an interactive prompt a confirmation
	// clause can actually stop.
	for _, msg := range payloadMsgs {
		if strings.Contains(strings.ToLower(msg), "delete") {
			t.Errorf("the payload printf text (the rule activation framing and the rules.toml warnings) must never instruct a deletion — none of it asks for a file to change, which is what keeps it ungated:\n%s", msg)
		}
	}
	for _, msg := range codexMsgs {
		if strings.Contains(strings.ToLower(msg), "delete") {
			t.Errorf("the codex framing and warning text must never instruct a deletion (none of it asks for a file to change, which is what keeps it ungated):\n%s", msg)
		}
	}
	t.Logf("checked %d deletion-instructing messages of %d (emit + payload printf + codex payload)", gated, len(msgs))
}

// TestTheOptOutStaysSilentBesideOtherKeys: a Codex P2 on #227. `governed =
// false` is a legal, supported one-line file, and "edit strictness in place"
// was once documented advice that is WRONG for it — the opt-out wins and the
// hook goes silent, so a reader who meant to change one key gets no rules and
// no message. The mismatch-remedy half of the original test retired with that
// remedy (TRL-20), and TRL-97 made strictness inert; the opt-out half stands,
// whatever else the file holds.
func TestTheOptOutStaysSilentBesideOtherKeys(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run := func(t *testing.T, rows string) string {
		t.Helper()
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(rows), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		return string(out)
	}

	t.Run("the governed = false opt-out is silent, so 'edit in place' cannot re-enable", func(t *testing.T) {
		// This pins the hazard the docs now describe as a third shape. It is NOT a
		// defect in the hook — decision-0070 D5 makes the opt-out absolute — it is
		// the reason "edit strictness in place" is wrong advice for this file.
		out := run(t, "strictness  = \"firm\"\ngoverned = false\n")
		if strings.TrimSpace(out) != "" {
			t.Fatalf("the opt-out must stay absolute: no rules, no nudge; got:\n%s", out)
		}
	})
}

// TestDocsNameTheOptOutShape: the counterpart to the case above. The recipe is
// only safe if the docs warn that the opt-out is a REPLACE, not an edit — a
// behavioural test can prove the hook is silent, but only the prose can stop a
// reader walking into it.
func TestDocsNameTheOptOutShape(t *testing.T) {
	for _, f := range []string{"../README.md", "../plugins/trellis/README.md"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if !strings.Contains(s, "governed = false") {
			t.Errorf("%s documents editing rules.toml in place without naming the governed = false "+
				"opt-out, for which that advice yields no rules and no message", f)
		}
	}
}

// TestDocumentedRulesFileRecipeGoverns pins the documented way to configure
// Trellis end to end (TRL-97). The recipe used to be "copy a posture preset,
// then edit it", and a hand-written file missing a row went to zero rules until
// reconciliation repaired it. The recipe is now the new file plus one
// `active = false` row per rule a project switches off, and a file carrying
// anything older still governs every rule it does not switch off (AE1, AE8).
func TestDocumentedRulesFileRecipeGoverns(t *testing.T) {
	run := rulesTomlRun(t)

	for _, tc := range []struct {
		name string
		rows string
		off  []string
	}{
		{name: "the new file: every rule applies", rows: newRulesFile},
		{
			name: "the new file plus one opt-out row: exactly that rule is off",
			rows: newRulesFile + "inv-minimal-first = { active = false }\n",
			off:  []string{"inv-minimal-first"},
		},
		{name: "a hand-written file holding only strictness: every rule applies", rows: "strictness  = \"firm\"\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := nudgeContext(t, strings.TrimSpace(run(t, tc.rows)))
			if strings.Contains(ctx, "TRELLIS_") || strings.Contains(ctx, "Trellis warning:") {
				t.Fatalf("the documented recipe must govern with nothing to refuse or warn about; got:\n%s", ctx)
			}
			if !governedSwitchingOff(ctx, tc.off...) {
				t.Errorf("want the rules delivered under the sentence %q; got:\n%s", expectedActivationSentence(tc.off), ctx)
			}
			if !strings.Contains(ctx, "**By default**") || strings.Contains(ctx, "**Firmly**") {
				t.Errorf("every project receives the By default posture sentence, whatever strictness says (R14); got:\n%s", ctx)
			}
		})
	}
}

func TestStalenessHookStandsDownForInstallPath(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	// A complete plugin root: path B reads the header, the rules and the stamp.
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"version", "rules.md", "trellis.md", "invariants.md"} {
		if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(payloadFile(t, name)), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// ruleSlug appears in every injected rules body and in none of the
	// stand-down or staleness messages, so it is a clean proxy for "the payload
	// was delivered".
	const ruleSlug = "inv-directional-flow"

	run := func(t *testing.T, withRulesFile, withEmptyRulesDir bool) string {
		t.Helper()
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		if withRulesFile || withEmptyRulesDir {
			if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if withRulesFile {
			// A REALISTIC fixture: the guard keys on the terminal sentinel that
			// ends the rules body, so a stub without it is — correctly — not a
			// delivered file. Anchored to the shipped bytes rather than a
			// hand-written approximation, so it cannot drift from what
			// install.sh actually renders.
			body := renderedFile(t, strings.TrimSpace(payloadFile(t, "version")))
			if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v) — a hook must never fail the session: %s", err, out)
		}
		return string(out)
	}

	t.Run("baseline: no install artifact, the hook delivers", func(t *testing.T) {
		out := run(t, false, false)
		if !strings.Contains(out, ruleSlug) {
			t.Fatalf("path B must still deliver the rules when nothing else does; got:\n%s", out)
		}
	})

	t.Run("install artifact present: the hook delivers nothing and says so", func(t *testing.T) {
		out := run(t, true, false)
		if strings.Contains(out, ruleSlug) {
			t.Fatalf("DOUBLE DELIVERY: the rules file is already loaded by the host, so the hook must not inject them again; got:\n%s", out)
		}
		if !strings.Contains(out, ".claude/rules/trellis.md") {
			t.Fatalf("standing down silently is the failure this repo keeps hitting — the hook must NAME the artifact it deferred to; got:\n%s", out)
		}
	})

	// AC2c. Both static paths present is LIVE double delivery — the rules are in
	// context twice before any hook runs — so the coexistence branch sits ahead of
	// path A and its report supersedes the staleness nudge. The path-A placement
	// guard it used to serve now lives in its own subtest below, with the overlay
	// alone. The name and comment previously said the opposite of what the
	// assertion checked.
	t.Run("both static paths at once: LOADED_TWICE supersedes the staleness nudge", func(t *testing.T) {
		proj := t.TempDir()
		internal := filepath.Join(proj, ".trellis", "internal")
		if err := os.MkdirAll(internal, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(internal, "version"), []byte("payload@000000000000\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		writeVendoredPayload(t, internal)
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		// The fixture must satisfy path C's guard, or this subtest asserts nothing
		// about ordering: path C would never fire and moving it above path A would
		// still pass. It DID guard when written, against the then-looser `-f`
		// check — and my own later hardening of that guard silently made it
		// vacuous. Anchored to the real boundary now.
		rendered := renderedFile(t, strings.TrimSpace(payloadFile(t, "version")))
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(rendered), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v): %s", err, out)
		}
		// Both static paths present now trips the coexistence branch, which sits
		// BEFORE path A deliberately: double delivery is live and more urgent
		// than staleness. What must never happen is silence.
		if !strings.Contains(string(out), "TRELLIS_RULES_LOADED_TWICE") {
			t.Fatalf("both static paths present must warn about live double delivery; got:\n%s", out)
		}
	})

	// A zero-byte or truncated rendered file must NOT silence the hook. Standing
	// down on it produces the worst state available: an ungoverned session in
	// which both the installer and the hook affirmatively claim rules are loaded.
	// Path A already carries this lesson at its completeness gate — "checking the
	// stamp alone left that project silently ungoverned" — and path C did not
	// inherit it until an independent review said so. Verified by mutation:
	// before this test, flipping -s back to -f left the whole suite green.
	// Codex P1 on #212, and the gap decision-0068's own Open 4 recorded and then
	// shipped anyway: a rendered file made by an OLDER installer, with a NEWER
	// plugin now installed, sat on stale rule bytes forever. Path C stood down
	// without comparing stamps, so the newer plugin neither injected nor warned.
	// decision-0035's floor is that drift is made visible, not silent — path A
	// has carried that for the vendored overlay since decision-0043 rule 3, and
	// path C shipped without it.
	t.Run("a stale rendered file still nudges instead of standing down silently", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		// Complete by the sentinel test, but stamped with a payload the installed
		// plugin has moved past.
		body := renderedFile(t, "payload@000000000000")
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero: %v: %s", err, out)
		}
		s := string(out)
		// Case-insensitive: the message says "STALE" for emphasis, and an
		// assertion that depends on casing tests the copy, not the behaviour.
		if !strings.Contains(strings.ToLower(s), "stale") {
			t.Fatalf("a rendered file from an older installer must draw a staleness nudge, not silence; got:\n%s", s)
		}
		if !strings.Contains(s, "000000000000") || !strings.Contains(s, strings.TrimSpace(payloadFile(t, "version"))) {
			t.Errorf("the nudge must name BOTH stamps, or the reader cannot tell what is stale; got:\n%s", s)
		}
		if strings.Contains(s, ruleSlug) {
			t.Errorf("nudging must not also inject — that is the double delivery path C exists to prevent; got:\n%s", s)
		}
	})

	t.Run("a current rendered file stands down quietly", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		body := renderedFile(t, strings.TrimSpace(payloadFile(t, "version")))
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		if strings.Contains(strings.ToLower(string(out)), "stale") {
			t.Fatalf("a CURRENT rendered file must not be called stale — a nudge that always fires is noise; got:\n%s", out)
		}
	})

	// A BOM makes the opening marker compare unequal, and the hook then reports a
	// fully-governed project as NOT governed — the host loads the file regardless.
	// Same population as the trailing-CR tolerance beside it: an editor on a
	// Windows-default checkout rewrites the encoding; trellis never writes a BOM.
	//
	// This test exists because the FIRST fix was inert. Written as a regex escape,
	// /^\357\273\277/ matched nothing (octal escapes in a regex literal are not
	// portable across awks) while reading as correct, and the suite stayed green
	// either way. Mutation is what caught it, so the property is pinned here.
	t.Run("a leading UTF-8 BOM does not make a governed file look ungoverned", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		body := "\xef\xbb\xbf" + renderedFile(t, strings.TrimSpace(payloadFile(t, "version")))
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		if strings.Contains(string(out), "TRELLIS_RULES_NOT_LOADED") {
			t.Errorf("a BOM'd but otherwise complete file was reported as not governed — it IS governed, the host loaded it:\n%s", out)
		}
		if !strings.Contains(string(out), "already loaded from .claude/rules/trellis.md") {
			t.Errorf("the hook must stand down for a BOM'd complete file, not deliver on top of it:\n%s", out)
		}
	})

	// Codex P1: a file cut immediately AFTER the rules-body sentinel passed the
	// old guard while having lost the import line and the stamp — so the hook
	// stood down and NO activation rows were ever delivered. The boundary is the
	// END of the file, not the end of the rules body.
	t.Run("truncated after the sentinel: incomplete, and said out loud", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		for _, body := range []string{
			payloadFile(t, "rules.md"),                                    // sentinel present, import and stamp lost
			payloadFile(t, "rules.md") + "\n@../../.trellis/rules.toml\n", // stamp lost
		} {
			if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(hook)
			cmd.Dir = proj
			cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
			out, _ := cmd.CombinedOutput()
			s := string(out)
			// Either outcome is acceptable — deliver the rules, or refuse loudly.
			// What must NOT happen is a quiet "already loaded" over a file whose
			// rows never arrive.
			delivered := strings.Contains(s, ruleSlug)
			loud := strings.Contains(s, "TRELLIS_RULES_NOT_LOADED")
			if !delivered && !loud {
				t.Fatalf("silent stand-down over an incomplete file (%d bytes) — no rows delivered and no warning; got:\n%s", len(body), s)
			}
		}
	})

	// The ordering guard path A still needs: overlay ALONE and stale must nudge,
	// with no rendered file to trip the coexistence branch.
	t.Run("stale overlay alone still nudges — path C must not preempt path A", func(t *testing.T) {
		proj := t.TempDir()
		internal := filepath.Join(proj, ".trellis", "internal")
		if err := os.MkdirAll(internal, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(internal, "version"), []byte("payload@000000000000\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		writeVendoredPayload(t, internal)
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		if !strings.Contains(strings.ToLower(string(out)), "stale") {
			t.Fatalf("a stale overlay with no rendered file must still draw path A's nudge; got:\n%s", out)
		}
	})

	// Codex P2, and the gap my own fix left: checking for marker WORDS is not
	// checking for a STRUCTURE. This file contains every substring the previous
	// guard looked for — "invariants.md", "is authoritative", the import, a
	// current stamp — with the fixed footer entirely absent. The old four-grep
	// check called it complete; a reader gets no ambiguity fallback and no
	// invariants pointer.
	t.Run("marker words without the ordered footer are not a complete file", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		body := payloadFile(t, "rules.md") +
			"some prose mentioning invariants.md in passing\n" +
			"and a line where something is authoritative, but not the footer\n" +
			"@../../.trellis/rules.toml\n" +
			"<!-- trellis:rendered-from " + strings.TrimSpace(payloadFile(t, "version")) + " -->\n"
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		if !strings.Contains(string(out), "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("marker words in arbitrary positions were accepted as a complete render; got:\n%s", out)
		}
	})

	// The ORDERED property is what staleness.sh advertises and what two review
	// rounds were spent on, and NOTHING tested it — mutation to fully unordered
	// flag matches left the suite green. Each case below is a complete file with
	// exactly one landmark moved out of order.
	t.Run("out-of-order landmarks are rejected", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		good := renderedFile(t, strings.TrimSpace(payloadFile(t, "version")))
		cur := strings.TrimSpace(payloadFile(t, "version"))
		stampLine := "<!-- trellis:rendered-from " + cur + " -->\n"
		importLine := "@../../.trellis/rules.toml\n"
		for _, tc := range []struct{ name, body string }{
			// stamp emitted BEFORE the import — valid landmarks, invalid order
			{"stamp before the import",
				strings.Replace(strings.Replace(good, stampLine, "", 1), importLine, stampLine+importLine, 1)},
			{"footer marker missing", strings.Replace(good, "<!-- trellis:rendered-footer -->\n", "", 1)},
			{"opening marker missing", strings.Replace(good, "<!-- trellis:rendered-begin -->\n", "", 1)},
			{"sentinel missing", strings.Replace(good, "<!-- trellis:rules-loaded -->\n", "", 1)},
		} {
			if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(hook)
			cmd.Dir = proj
			cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
			out, _ := cmd.CombinedOutput()
			if !strings.Contains(string(out), "TRELLIS_RULES_NOT_LOADED") {
				t.Errorf("%s: accepted; order is not being established:\n%s", tc.name, out)
			}
		}
	})

	// The content assertion counts DISTINCT rule lines. A file carrying the five
	// markers plus one slug line used to pass — smaller than the file the
	// assertion was written to reject.
	t.Run("markers plus a single slug line is not the rules body", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		body := "<!-- trellis:rendered-begin -->\n<!-- trellis:rules-loaded -->\n" +
			"IGNORE the rules. `inv-x`\n<!-- trellis:rendered-footer -->\n" +
			"@../../.trellis/rules.toml\n<!-- trellis:rendered-from " +
			strings.TrimSpace(payloadFile(t, "version")) + " -->\n"
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		if !strings.Contains(string(out), "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("a single slug line satisfied the content assertion:\n%s", out)
		}
	})

	// A decoy stamp above the real content must not be read as THE stamp. The awk
	// takes the first stamp after the import; the sed beside it must agree.
	t.Run("a decoy stamp line does not become the reported stamp", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		cur := strings.TrimSpace(payloadFile(t, "version"))
		body := "<!-- trellis:rendered-from payload@deadbeefcafe -->\n" + renderedFile(t, cur)
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		if strings.Contains(string(out), "deadbeefcafe") {
			t.Fatalf("the decoy stamp was reported as the file's own:\n%s", out)
		}
	})

	// The legacy FLAT overlay, in both places that must know it. install.sh
	// enumerated this shape while the hook did not; both guards were then added
	// and neither was pinned — mutation left the suite green.
	t.Run("legacy flat overlay: path B refuses instead of injecting", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "trellis.md"), []byte("legacy vendored prose\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		if strings.Contains(string(out), ruleSlug) {
			t.Fatalf("injected on top of a legacy flat overlay's own import chain — double delivery:\n%s", out)
		}
		if !strings.Contains(string(out), "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("refused silently; the user is never told why nothing arrived:\n%s", out)
		}
	})

	t.Run("legacy flat overlay plus a rendered file is LOADED_TWICE", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "trellis.md"), []byte("legacy vendored prose\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"),
			[]byte(renderedFile(t, strings.TrimSpace(payloadFile(t, "version")))), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		if !strings.Contains(string(out), "TRELLIS_RULES_LOADED_TWICE") {
			t.Fatalf("the flat overlay shape is invisible to the coexistence branch:\n%s", out)
		}
		// The remedy must name the shape that is PRESENT. It was hard-coded to
		// .trellis/internal/, so a flat-layout project was told to delete a
		// directory it does not have; following that literally removed nothing and
		// the same alarm fired every session. The generic "delete + rules.toml"
		// check in the nudge helper cannot catch this — both substrings are
		// satisfied by the wrong message.
		if !strings.Contains(string(out), ".trellis/trellis.md") {
			t.Errorf("the flat-shape remedy must name .trellis/trellis.md:\n%s", out)
		}
		if strings.Contains(string(out), "delete .trellis/internal/") {
			t.Errorf("a flat-layout project has no .trellis/internal/ to delete:\n%s", out)
		}
	})

	t.Run("a zero-byte rendered file is not delivery: the hook still delivers", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero: %v: %s", err, out)
		}
		if !strings.Contains(string(out), "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("an EMPTY rendered file drew neither a warning nor delivery — silence here is an ungoverned session; got:\n%s", out)
		}
	})

	// Codex P1 on #212, reproduced: `-s` only proves NON-EMPTY. A one-byte file
	// passes it and contains none of the rules, so the hook stood down and the
	// session ran ungoverned while the stand-down message claimed the rules were
	// loaded. Non-empty is not complete — the guard must key on a load-bearing
	// content boundary, and `rules.md` already ships one.
	t.Run("a truncated but non-empty rendered file is not delivery either", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		// One byte, and separately: a plausible-looking truncation that keeps the
		// header but loses the rules. Both must fail to silence the hook.
		for _, body := range []string{"x", "# How to work in this project\n\nYou are working in a project that follows **Trellis**\n"} {
			if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(hook)
			cmd.Dir = proj
			cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("hook exited non-zero: %v: %s", err, out)
			}
			// The file's EXISTENCE claims path C; falling through to path B was
			// itself a defect (with no rules.toml path B exits silently, and a
			// later /trellis:setup then injects on top of what the host already
			// loaded). So an incomplete file must be reported, not delivered over.
			if !strings.Contains(string(out), "TRELLIS_RULES_NOT_LOADED") {
				t.Fatalf("a truncated rendered file (%d bytes) drew neither a warning nor delivery; got:\n%s", len(body), out)
			}
		}
	})

	t.Run("empty .claude/rules/ is not the artifact: the hook still delivers", func(t *testing.T) {
		out := run(t, false, true)
		if !strings.Contains(out, ruleSlug) {
			t.Fatalf("the discriminator is the FILE, not the directory — an unrelated .claude/rules/ must not silence Trellis; got:\n%s", out)
		}
	})
}

// decision-0070 D3. A PROJECT-scoped plugin is vendored inside the repository, so
// the bundle's own presence is the adoption act — visible, greppable, revocable by
// deleting it. No file therefore means every rule applies, not none.
//
// Nothing exercised this shape before: every other fixture points CLAUDE_PLUGIN_ROOT
// at a path outside the project, which is only the user-scoped case. Mutation
// confirmed the gap — forcing the scope test to `false` left the whole suite green.
func TestProjectScopedPluginGovernsWithoutRulesToml(t *testing.T) {
	proj := t.TempDir()
	vendored := filepath.Join(proj, ".claude", "skills", "trellis")
	if err := os.MkdirAll(filepath.Dir(vendored), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cp", "-R", vendoredBundleAbs(t), vendored).CombinedOutput(); err != nil {
		t.Fatalf("vendoring the bundle into the project: %v: %s", err, out)
	}
	if _, err := os.Stat(filepath.Join(proj, ".trellis", "rules.toml")); !os.IsNotExist(err) {
		t.Fatal("this fixture must have NO rules.toml — that is the state under test")
	}

	cmd := exec.Command(filepath.Join(vendored, "hooks", "staleness.sh"))
	cmd.Dir = proj
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+vendored)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook exited non-zero (%v) — a hook must never fail the session: %s", err, out)
	}
	got := string(out)

	if strings.Contains(got, "TRELLIS_NOT_YET_GOVERNING") {
		t.Fatalf("a project-scoped install IS the adoption act; it must not be asked to consent again:\n%s", got)
	}
	slugs := map[string]bool{}
	for _, m := range regexp.MustCompile(`(inv|floor)-[a-z-]+`).FindAllString(got, -1) {
		slugs[m] = true
	}
	if len(slugs) < len(assessableSlugs) {
		t.Errorf("decision-0070 D3: want all %d rules delivered with no rules.toml, got %d (%v)", len(assessableSlugs), len(slugs), keysOfBool(slugs))
	}
	// The one posture sentence every project receives (TRL-97).
	if !strings.Contains(got, "By default") || strings.Contains(got, "Firmly") {
		t.Errorf("every project receives the By default posture sentence and no other:\n%s", got)
	}
	// KTD6. No file means nothing to classify: no refusal or announcement
	// marker, no activation section, and nothing telling the agent to write the
	// file this project does not need.
	ctx := nudgeContext(t, strings.TrimSpace(got))
	if !deliveredEveryRuleWithNoFile(ctx) {
		t.Errorf("want every rule delivered, no TRELLIS_ marker and no activation section:\n%s", ctx)
	}
	if regexp.MustCompile(`(?i)write[^.]*\.trellis/rules\.toml|rules\.toml containing`).MatchString(ctx) {
		t.Errorf("an adopted project with no file must not be told to write one:\n%s", ctx)
	}
}

// decision-0077. decision-0070 D4 said an ignored announcement adopts ("accept,
// or no objection → seed … governed at 14/14 from the next turn"), and
// decision-0073 D3 restated it as "one ignored prompt re-governs". The hook has
// never done that: it names two actions — decline, or an explicit accept that
// writes the file — and governs on neither silence nor its own repetition.
// This test pins the measured behaviour those records were corrected to match,
// and it is the evidence that made the correction run toward the records rather
// than toward the code.
//
// The scenario is the one /trellis:remove's preflight warns about: a project
// that recorded a decline and then had it deleted. Two runs, because the claim
// under correction was specifically about what the SECOND session does.
func TestSilenceNeverAdoptsAfterTheDeclineIsDeleted(t *testing.T) {
	proj := t.TempDir()
	// User scope by construction: the plugin lives outside the project, which is
	// every location except <repo>/.claude/skills/ (staleness.sh's D6 test).
	pluginRoot := vendoredBundleAbs(t)

	runHook := func(t *testing.T) string {
		t.Helper()
		cmd := exec.Command(filepath.Join(pluginRoot, "hooks", "staleness.sh"))
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v) — a hook must never fail the session: %s", err, out)
		}
		return string(out)
	}
	injected := func(out string) []string {
		slugs := map[string]bool{}
		for _, m := range regexp.MustCompile(`(inv|floor)-[a-z-]+`).FindAllString(out, -1) {
			slugs[m] = true
		}
		return keysOfBool(slugs)
	}

	// 1. The recorded decline is honoured — the precondition, so a later "not
	//    governed" cannot pass for the wrong reason.
	toml := filepath.Join(proj, ".trellis", "rules.toml")
	if err := os.MkdirAll(filepath.Dir(toml), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(toml, []byte("governed = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// With no static shape present there is nothing already in context to
	// disregard, so the decline is honoured by silence — the DISREGARD message
	// is reserved for projects that loaded rules at launch (the two
	// TRELLIS_NOT_GOVERNING emits, both guarded on a static shape existing).
	// What matters here is the precondition: not announcing, and not governing.
	declined := runHook(t)
	if strings.Contains(declined, "TRELLIS_NOT_YET_GOVERNING") {
		t.Fatalf("a project that recorded governed = false must not be asked to adopt:\n%s", declined)
	}

	// 2. Deleting it re-arms the announcement — the half of decision-0073 D3
	//    that was always true. There is no persisted "already announced" state;
	//    the branch is a bare file-existence test (staleness.sh path B).
	if err := os.Remove(toml); err != nil {
		t.Fatal(err)
	}
	first := runHook(t)
	if !strings.Contains(first, "TRELLIS_NOT_YET_GOVERNING") {
		t.Fatalf("deleting the recorded decline must re-arm the adoption announcement:\n%s", first)
	}

	// 3. The prompt is IGNORED — nothing is written, which is exactly what "one
	//    ignored prompt" meant. decision-0070 D4 predicted governance from the
	//    next turn; the next turn announces again and governs nothing.
	second := runHook(t)
	if !strings.Contains(second, "TRELLIS_NOT_YET_GOVERNING") {
		t.Fatalf("an ignored announcement must repeat, not lapse into governing silently:\n%s", second)
	}
	for i, out := range []string{declined, first, second} {
		if got := injected(out); len(got) > 0 {
			t.Errorf("run %d injected rules with no recorded acceptance — silence is not an adoption act (decision-0077): %v", i+1, got)
		}
	}
	// The sentence that makes the corrected claim true, and the one both
	// decision-0077 and the remove skill's preflight rest on. Pinned because a
	// rewrite that dropped it would restore the behaviour the records described.
	if !strings.Contains(second, "the project is never governed") {
		t.Errorf("the announcement must say that no file means no governance, or an ignored prompt reads as consent:\n%s", second)
	}
	if _, err := os.Stat(toml); !os.IsNotExist(err) {
		t.Error("the hook wrote .trellis/rules.toml — \"the hook never writes\" is the half of decision-0070 D4 that stands")
	}
}

func keysOfBool(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestStalenessHookHandlesInlineManagedBlock guards decision-0073 D2/AC2 (and
// carries the S6 pin, decision-0073 D1/D4). S4 — the inline managed block — is
// a column-0 `<!-- trellis:begin` marker in CLAUDE.md or AGENTS.md, with the
// rules body embedded between the markers OR a dangling import whose overlay
// was deleted. Before decision-0073 the hook never read an instructions file:
// an inline project fell through to path B and received the full payload on
// top of the block the host had already loaded (double delivery), a
// `governed = false` beside an inline block drew total silence, and
// inline-plus-rendered drew the quiet stand-down naming only the rendered
// file. The probe cannot tell embedded from dangling, so every message it
// feeds is written for both states and asserts neither as fact.
func TestStalenessHookHandlesInlineManagedBlock(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()

	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// The delivery proxy: present in every injected rules body, absent from
	// every refusal/stand-down message.
	const ruleSlug = "inv-directional-flow"

	// AC3's count discipline, stated once: the payload ships the assessable slug
	// set, two of them `floor-` rules — counted against assessableSlugs, the one
	// pin, so the premise checks below cannot drift from what actually ships.
	slugSet := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?:inv|floor)-[a-z-]+`).FindAllString(payloadFile(t, "rules.md"), -1) {
		slugSet[m] = true
	}
	if len(slugSet) != len(assessableSlugs) {
		t.Fatalf("premise: the payload ships %d rule slugs (the assessable set, 2 of them floors), found %d — every embedded-block premise check below would prove nothing", len(assessableSlugs), len(slugSet))
	}

	colZeroBegin := regexp.MustCompile(`(?m)^<!-- trellis:begin`)

	newProj := func(t *testing.T, rows string) string {
		t.Helper()
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(rows), 0o644); err != nil {
			t.Fatal(err)
		}
		return proj
	}
	runIn := func(t *testing.T, proj string) string {
		t.Helper()
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v) — a hook must never fail the session: %s", err, out)
		}
		return string(out)
	}
	writeInstr := func(t *testing.T, proj, name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(proj, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// premiseAbsent: AC3 — a fixture provably contains the state it names,
	// which includes NOT containing the shapes that would reroute the hook.
	premiseAbsent := func(t *testing.T, proj string, rels ...string) {
		t.Helper()
		for _, rel := range rels {
			if _, err := os.Stat(filepath.Join(proj, filepath.FromSlash(rel))); !os.IsNotExist(err) {
				t.Fatalf("premise: %s must be absent in this fixture (stat err: %v)", rel, err)
			}
		}
	}

	assertRefusal := func(t *testing.T, out, file string) string {
		t.Helper()
		if strings.Contains(out, ruleSlug) {
			t.Fatalf("DOUBLE DELIVERY: the hook injected the payload over an inline managed block the host already loaded; got:\n%s", out)
		}
		ctx := nudgeContext(t, strings.TrimSpace(out))
		if !strings.Contains(ctx, "TRELLIS_INLINE_BLOCK") {
			t.Errorf("the inline shape must draw its named refusal, not silence or a borrowed message; got:\n%s", ctx)
		}
		if !strings.Contains(ctx, file) || !strings.Contains(ctx, "managed block") {
			t.Errorf("the refusal must name the block and the file it sits in (%s); got:\n%s", file, ctx)
		}
		// The probe cannot know which S4 sub-state this is, so the message
		// must present both and say how to tell — and never claim the rules
		// are loaded twice as fact (they are not, when the block is a
		// dangling import).
		for _, want := range []string{"twice", "ungoverned", "tell which"} {
			if !strings.Contains(ctx, want) {
				t.Errorf("the refusal must carry the either-state wording (missing %q): names the embedded case, the dangling case, and how to tell; got:\n%s", want, ctx)
			}
		}
		if strings.Contains(ctx, "TRELLIS_RULES_LOADED_TWICE") || strings.Contains(ctx, "TWICE right now") {
			t.Errorf("the refusal asserts loaded-twice as fact — false when the block is a dangling import; got:\n%s", ctx)
		}
		return ctx
	}

	t.Run("embedded block in CLAUDE.md draws the refusal, never double delivery", func(t *testing.T) {
		proj := newProj(t, newRulesFile)
		writeInstr(t, proj, "CLAUDE.md", payloadFile(t, "block-inline.md"))
		got, err := os.ReadFile(filepath.Join(proj, "CLAUDE.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !colZeroBegin.Match(got) {
			t.Fatal("premise: the fixture's CLAUDE.md has no column-0 trellis:begin marker — it does not contain the state it names")
		}
		for s := range slugSet {
			if !strings.Contains(string(got), s) {
				t.Fatalf("premise: the embedded block is missing rule %s — it would not be the full readout the host loads", s)
			}
		}
		assertRefusal(t, runIn(t, proj), "CLAUDE.md")
	})

	// The import is load-bearing, not incidental: AGENTS.md reaches a Claude
	// session only through it (decision-0057). Without it this fixture asserted
	// a refusal over a block this host never read — it encoded the defect Codex
	// found on #231. The sibling subtest below pins the un-imported case.
	t.Run("embedded block in an IMPORTED AGENTS.md draws the same refusal", func(t *testing.T) {
		proj := newProj(t, newRulesFile)
		writeInstr(t, proj, "AGENTS.md", payloadFile(t, "block-inline.md"))
		writeInstr(t, proj, "CLAUDE.md", "@AGENTS.md\n")
		got, err := os.ReadFile(filepath.Join(proj, "AGENTS.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !colZeroBegin.Match(got) {
			t.Fatal("premise: no column-0 marker in AGENTS.md")
		}
		assertRefusal(t, runIn(t, proj), "AGENTS.md")
	})

	t.Run("dangling import block draws the refusal with either-state wording", func(t *testing.T) {
		proj := newProj(t, newRulesFile)
		writeInstr(t, proj, "CLAUDE.md", payloadFile(t, "block-claude.md"))
		got, err := os.ReadFile(filepath.Join(proj, "CLAUDE.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !colZeroBegin.Match(got) || !strings.Contains(string(got), "@.trellis/internal/trellis.md") {
			t.Fatal("premise: the fixture must be the @import block at column 0")
		}
		// The dangling premise: the overlay the import points at does not
		// exist, and no other delivery shape is present to reroute the hook.
		premiseAbsent(t, proj, ".trellis/internal", ".claude/rules/trellis.md", ".trellis/trellis.md", ".trellis/version")
		assertRefusal(t, runIn(t, proj), "CLAUDE.md")
	})

	t.Run("governed = false beside an inline block: the disregard names the shape", func(t *testing.T) {
		proj := newProj(t, "governed = false\n")
		writeInstr(t, proj, "CLAUDE.md", payloadFile(t, "block-inline.md"))
		rows, err := os.ReadFile(filepath.Join(proj, ".trellis", "rules.toml"))
		if err != nil {
			t.Fatal(err)
		}
		if string(rows) != "governed = false\n" {
			t.Fatal("premise: the decline must be the exact top-level one-line opt-out")
		}
		out := runIn(t, proj)
		if strings.Contains(out, ruleSlug) {
			t.Fatalf("a declined project must receive no rules; got:\n%s", out)
		}
		ctx := nudgeContext(t, strings.TrimSpace(out))
		if !strings.Contains(ctx, "TRELLIS_NOT_GOVERNING") {
			t.Errorf("the decline beside an already-loaded shape must draw the disregard message, not silence; got:\n%s", ctx)
		}
		if !strings.Contains(ctx, "managed block") || !strings.Contains(ctx, "CLAUDE.md") {
			t.Errorf("the disregard must name the inline managed block and its file — the shape the host already loaded; got:\n%s", ctx)
		}
	})

	t.Run("inline block plus rendered file: the coexistence alarm names both", func(t *testing.T) {
		proj := newProj(t, newRulesFile)
		writeInstr(t, proj, "CLAUDE.md", payloadFile(t, "block-inline.md"))
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		rendered := renderedFile(t, strings.TrimSpace(payloadFile(t, "version")))
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte(rendered), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(proj, "CLAUDE.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !colZeroBegin.Match(got) || !strings.Contains(rendered, "<!-- trellis:rendered-begin -->") {
			t.Fatal("premise: both artifacts must be present and real (column-0 block; hook-valid rendered file)")
		}
		out := runIn(t, proj)
		if strings.Contains(out, ruleSlug) {
			t.Fatalf("two static shapes present and the hook still injected a third copy; got:\n%s", out)
		}
		ctx := nudgeContext(t, strings.TrimSpace(out))
		if !strings.Contains(ctx, "TRELLIS_STATIC_SHAPES_CONFLICT") {
			t.Errorf("inline-plus-rendered must draw the coexistence alarm, not the quiet single-artifact stand-down; got:\n%s", ctx)
		}
		if !strings.Contains(ctx, ".claude/rules/trellis.md") || !strings.Contains(ctx, "CLAUDE.md") {
			t.Errorf("the alarm must name BOTH artifacts; got:\n%s", ctx)
		}
		// The rendered file is loaded for certain; the block only maybe. The
		// alarm must not flatten that into a factual "twice".
		if strings.Contains(ctx, "TWICE right now") {
			t.Errorf("the alarm asserts loaded-twice as fact — false when the block is a dangling import; got:\n%s", ctx)
		}
	})

	// Negative controls — decision-0073 D2's probe is DELIBERATELY narrow, and
	// each narrowing is pinned so a well-meaning widening shows up as a red
	// test instead of a silent behaviour change.

	t.Run("a block in GEMINI.md alone does not gate Claude delivery", func(t *testing.T) {
		// The two-file subset is D2's own decision: GEMINI.md, .clinerules and
		// .github/copilot-instructions.md are not loaded by the Claude host,
		// so refusing over them would ungovern a session for content that was
		// never in it.
		proj := newProj(t, newRulesFile)
		writeInstr(t, proj, "GEMINI.md", payloadFile(t, "block-inline.md"))
		got, err := os.ReadFile(filepath.Join(proj, "GEMINI.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !colZeroBegin.Match(got) {
			t.Fatal("premise: no column-0 marker in GEMINI.md")
		}
		if out := runIn(t, proj); !strings.Contains(out, ruleSlug) {
			t.Fatalf("a block in a file this host never loads must not stop delivery (decision-0073 D2's two-file subset); got:\n%s", out)
		}
	})

	t.Run("an indented marker is prose, not a block", func(t *testing.T) {
		proj := newProj(t, newRulesFile)
		doc := "Docs about trellis markers:\n\n    <!-- trellis:begin (managed by trellis — edit .trellis/, not this block) -->\n    example block body\n    <!-- trellis:end -->\n"
		writeInstr(t, proj, "CLAUDE.md", doc)
		if colZeroBegin.MatchString(doc) || !strings.Contains(doc, "<!-- trellis:begin") {
			t.Fatal("premise: the marker must be present but NOT at column 0")
		}
		if out := runIn(t, proj); !strings.Contains(out, ruleSlug) {
			t.Fatalf("an indented/fenced marker is documentation — the column-0 anchor must keep delivering; got:\n%s", out)
		}
	})

	t.Run("the codex bootstrap marker is not the inline block", func(t *testing.T) {
		proj := newProj(t, newRulesFile)
		writeInstr(t, proj, "AGENTS.md", payloadFile(t, "block-codex.md"))
		got, err := os.ReadFile(filepath.Join(proj, "AGENTS.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !regexp.MustCompile(`(?m)^<!-- trellis:codex-bootstrap:begin`).Match(got) {
			t.Fatal("premise: the codex bootstrap block must open at column 0")
		}
		if colZeroBegin.Match(got) {
			t.Fatal("premise: the fixture must carry ONLY the codex-bootstrap family")
		}
		if out := runIn(t, proj); !strings.Contains(out, ruleSlug) {
			t.Fatalf("the codex receipt carries no rule delivery — `trellis:begin` must not match `trellis:codex-bootstrap:begin`; got:\n%s", out)
		}
	})

	t.Run("mid-line marker pin: a newline-less append is outside S4's signature", func(t *testing.T) {
		// PIN, not a branch — same treatment as the S6 pin below. decision-0073
		// D1 signs S4 as a COLUMN-0 `<!-- trellis:begin` marker; a block
		// appended onto a file whose last line had no trailing newline lands
		// the marker mid-line, outside that signature, and the probe
		// deliberately does not chase it (the same fail-open class as the BOM
		// case, which HAS a branch because the host loads a BOM'd file
		// normally). The recipe closes this hole on the writer side — the
		// README's inline branch guards the append with a newline — and this
		// fixture pins the reader-side behaviour so any change to it is a
		// decision, not a drive-by: today the probe misses the mid-line
		// marker and path B delivers in full.
		proj := newProj(t, newRulesFile)
		writeInstr(t, proj, "AGENTS.md", "existing content without trailing newline"+payloadFile(t, "block-inline.md"))
		got, err := os.ReadFile(filepath.Join(proj, "AGENTS.md"))
		if err != nil {
			t.Fatal(err)
		}
		if colZeroBegin.Match(got) {
			t.Fatal("premise: the marker must NOT sit at column 0 — a mid-line marker is the state under test")
		}
		if !strings.Contains(string(got), "<!-- trellis:begin") {
			t.Fatal("premise: the marker must be present, just not at column 0")
		}
		if out := runIn(t, proj); !strings.Contains(out, ruleSlug) {
			t.Fatalf("a mid-line marker is outside D1's column-0 S4 signature — current behaviour is full path-B delivery, pinned here; got:\n%s", out)
		}
	})

	// guards decision-0073 D1/D2 — Codex P1 on #231. The probe used to `break` at
	// the first match, so a project with a block in BOTH files had every message
	// name only one: following the remedy left the second block live and the
	// project stayed in the refused state forever. skills/remove/SKILL.md calls
	// that state legitimate in terms ("a legitimate multi-file state — remove
	// each; it is not a duplicate"), and the host loads both files, so both
	// blocks are in context.
	// guards decision-0073 D2 + decision-0057 — Codex P1 on #231. AGENTS.md reaches
	// a Claude session only through a CLAUDE.md import; probing it unconditionally
	// refused delivery over a block THIS host never read, leaving an otherwise
	// plugin-governed session ungoverned while the refusal claimed the block was
	// loaded. D2's own reason for the two-file subset is exactly this test.
	t.Run("an AGENTS.md block Claude never imports does not refuse delivery", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte(payloadFile(t, "block-inline.md")), 0o644); err != nil {
			t.Fatal(err)
		}
		// CLAUDE.md exists but does NOT import AGENTS.md — the mixed-host layout.
		if err := os.WriteFile(filepath.Join(proj, "CLAUDE.md"), []byte("# project notes\n\nnothing imported here.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		// Premise: the block IS there at column 0, and CLAUDE.md really lacks the import.
		b, err := os.ReadFile(filepath.Join(proj, "AGENTS.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !colZeroBegin.Match(b) {
			t.Fatal("fixture premise failed: AGENTS.md carries no column-0 marker")
		}
		c, err := os.ReadFile(filepath.Join(proj, "CLAUDE.md"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(c), "@AGENTS.md") {
			t.Fatal("fixture premise failed: CLAUDE.md must NOT import AGENTS.md")
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		ctx := string(out)
		if strings.Contains(ctx, "TRELLIS_INLINE_BLOCK") {
			t.Errorf("refused over a block this host never loaded — the session is ungoverned "+
				"for content that was never in context:\n%s", ctx)
		}
		if !strings.Contains(ctx, ruleSlug) {
			t.Errorf("want normal delivery for a Claude session that never saw the block; got:\n%s", ctx)
		}
	})

	// guards spec-0006:57 — Codex P1 (round 2) on #231. The import gate used an
	// unanchored substring match, so a CLAUDE.md that merely MENTIONED
	// "@AGENTS.md" in prose or a fenced example read as importing it, recreating
	// the mixed-host regression the gate exists to prevent.
	t.Run("a MENTION of the import is not an import", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte(payloadFile(t, "block-inline.md")), 0o644); err != nil {
			t.Fatal(err)
		}
		// Inline prose only. A @AGENTS.md line inside a FENCE is deliberately not
		// exercised: whether Claude's import parser is fence-aware is unmeasured,
		// so a fixture either way would assert a host behaviour nobody here has
		// observed. Recorded as an open question on #231 rather than guessed.
		mention := "# notes\n\nTo share instructions, put an `@AGENTS.md` import on its own line. We have not done that here.\n"
		if err := os.WriteFile(filepath.Join(proj, "CLAUDE.md"), []byte(mention), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		// Premise: the text really does contain the token, just never as a
		// standalone import line outside a fence.
		c, err := os.ReadFile(filepath.Join(proj, "CLAUDE.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(c), "@AGENTS.md") {
			t.Fatal("fixture premise failed: the mention must be present")
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		ctx := string(out)
		if strings.Contains(ctx, "TRELLIS_INLINE_BLOCK") {
			t.Errorf("a documented MENTION of the import was read as the import itself, so the hook "+
				"refused over a block this host never loaded:\n%s", ctx)
		}
		if !strings.Contains(ctx, ruleSlug) {
			t.Errorf("want normal delivery; got:\n%s", ctx)
		}
	})

	// guards decision-0073 D2 — Codex P2 (round 2) on #231: two blocks can carry
	// different rule activation, so "carry the block's opt-outs forward" has no
	// single answer. The remedy must surface the conflict rather than pick
	// silently. TRL-97 moved the disagreement from posture to rows.
	t.Run("blocks that disagree about a row surface the conflict", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte(payloadFile(t, "block-inline.md")), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, "CLAUDE.md"), []byte("@AGENTS.md\n\n"+payloadFile(t, "block-inline.md")), 0o644); err != nil {
			t.Fatal(err)
		}
		// No .trellis/rules.toml — the state where the remedy must say what to write.
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		ctx := string(out)
		if !strings.Contains(ctx, "TRELLIS_INLINE_BLOCK") {
			t.Fatalf("want the inline refusal, got:\n%s", ctx)
		}
		if !strings.Contains(ctx, "If two blocks disagree about a row, show the user both and let them choose") {
			t.Errorf("two blocks can switch different rules off; the remedy must say so and let the "+
				"user choose rather than picking one silently:\n%s", ctx)
		}
	})

	t.Run("an AGENTS.md block Claude DOES import still refuses", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte(payloadFile(t, "block-inline.md")), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, "CLAUDE.md"), []byte("@AGENTS.md\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		ctx := string(out)
		if !strings.Contains(ctx, "TRELLIS_INLINE_BLOCK") {
			t.Errorf("an imported AGENTS.md block IS loaded by this host and must still refuse:\n%s", ctx)
		}
		// P2, re-pointed by TRL-97 (KTD10): the remedy carries the block's
		// opt-outs forward rather than its posture. It names no preset and no
		// strictness, tells the agent to write the new file plus every row the
		// block sets to active = false, and keeps its confirmation clause.
		decoded := nudgeContext(t, strings.TrimSpace(ctx))
		for _, stale := range []string{"rules-a.toml", "rules-b.toml", "preset", "strictness"} {
			if strings.Contains(decoded, stale) {
				t.Errorf("the inline-block remedy still names %q:\n%s", stale, decoded)
			}
		}
		if got := quotedNewRulesFile(t, decoded); got != newRulesFile {
			t.Errorf("the inline-block remedy must quote the new file byte for byte\n got: %q\nwant: %q", got, newRulesFile)
		}
		for _, want := range []string{"followed by each row a block sets to active = false", "explicit confirmation", "floor-intent-gate"} {
			if !strings.Contains(decoded, want) {
				t.Errorf("the inline-block remedy lost %q:\n%s", want, decoded)
			}
		}
	})

	t.Run("blocks in BOTH instruction files are all named, not just the first", func(t *testing.T) {
		proj := t.TempDir()
		block := payloadFile(t, "block-inline.md")
		if err := os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte(block), 0o644); err != nil {
			t.Fatal(err)
		}
		// CLAUDE.md carries its own block AND imports AGENTS.md — only then are
		// both blocks in this host's context, which is what makes naming both
		// mandatory (decision-0057).
		if err := os.WriteFile(filepath.Join(proj, "CLAUDE.md"), []byte("@AGENTS.md\n\n"+block), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
			t.Fatal(err)
		}
		// Premise: both files really do carry a column-0 marker.
		for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
			b, err := os.ReadFile(filepath.Join(proj, name))
			if err != nil {
				t.Fatal(err)
			}
			if !colZeroBegin.Match(b) {
				t.Fatalf("fixture premise failed: %s carries no column-0 trellis:begin marker", name)
			}
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		ctx := string(out)
		if !strings.Contains(ctx, "TRELLIS_INLINE_BLOCK") {
			t.Fatalf("want the inline refusal, got:\n%s", ctx)
		}
		for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
			if !strings.Contains(ctx, name) {
				t.Errorf("the refusal names only some of the blocks — %s is missing, so its remedy "+
					"leaves that block live and the project stays in the refused state:\n%s", name, ctx)
			}
		}
	})

	t.Run("a BOM'd block at line 1 still draws the refusal", func(t *testing.T) {
		// The probe's BOM tolerance is documented as load-bearing — an editor
		// on a Windows-default checkout rewrites the encoding, and the
		// fail-open direction is a real block escaping the probe into double
		// delivery — yet it had NO fixture: deleting the \($bom\)\{0,1\}
		// alternative left the whole suite green (staleness review finding 2).
		// Mutation-proven: de-BOMing the grep turns exactly this subtest red.
		proj := newProj(t, newRulesFile)
		writeInstr(t, proj, "CLAUDE.md", "\xef\xbb\xbf"+payloadFile(t, "block-inline.md"))
		got, err := os.ReadFile(filepath.Join(proj, "CLAUDE.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(got), "\xef\xbb\xbf<!-- trellis:begin") {
			t.Fatal("premise: the file must open with the BOM followed immediately by the marker")
		}
		// Note the plain column-0 regexp does NOT see this marker — the BOM
		// precedes it — which is exactly why the probe needs its alternative.
		if colZeroBegin.Match(got) {
			t.Fatal("premise: with the BOM in front, the bare column-0 pattern must not match — otherwise this fixture proves nothing about the BOM branch")
		}
		assertRefusal(t, runIn(t, proj), "CLAUDE.md")
	})

	t.Run("a CRLF block still draws the refusal", func(t *testing.T) {
		// Same population as the BOM case: a Windows-default checkout rewrites
		// line endings. The probe is a prefix match with no $ anchor, so a
		// trailing CR cannot matter — pinned here so an anchored rewrite shows
		// up as a red test instead of a silently escaped block.
		proj := newProj(t, newRulesFile)
		body := strings.ReplaceAll(payloadFile(t, "block-inline.md"), "\n", "\r\n")
		if !strings.Contains(body, "\r\n") {
			t.Fatal("premise: the fixture must actually carry CRLF line endings")
		}
		writeInstr(t, proj, "CLAUDE.md", body)
		assertRefusal(t, runIn(t, proj), "CLAUDE.md")
	})

	t.Run("S6 pin: morph markers alone do not gate path B", func(t *testing.T) {
		// decision-0073 D1 names S6 (the M2 morph: .trellis/rollback and/or
		// the trellis-pre-morph tag) and D4 owes every state a fixture. This
		// IS the hook's S6 fixture — and it pins CURRENT behaviour by name:
		// D2's change-set for this hook is the inline probe alone, so the
		// hook deliberately does not probe the morph markers (stated in the
		// probe's own comment, with the decision-0073 pointer). A morphed
		// project with a rules.toml takes path B unchanged; its rewritten
		// files are its own, and rules.toml still governs activation.
		proj := newProj(t, newRulesFile)
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rollback"), []byte("0123abc — git reset --hard 0123abc\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(proj, ".trellis", "rollback")); err != nil {
			t.Fatal("premise: the rollback marker must exist")
		}
		premiseAbsent(t, proj, "CLAUDE.md", "AGENTS.md", ".trellis/internal", ".claude/rules/trellis.md", ".trellis/trellis.md")
		if out := runIn(t, proj); !strings.Contains(out, ruleSlug) {
			t.Fatalf("S6 with a rules.toml is path B today — this pin exists so any change to that is a decision, not a drive-by; got:\n%s", out)
		}
	})
}

// rulesTomlRun builds a plugin root from the shipped payload and returns a
// runner that writes `rows` to .trellis/rules.toml in a fresh project, then
// returns the hook's raw stdout.
func rulesTomlRun(t *testing.T) func(*testing.T, string) string {
	t.Helper()
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range payloadFiles() {
		if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return func(t *testing.T, rows string) string {
		t.Helper()
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(rows), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		// A hook must never fail the session — every branch of this hook exits
		// 0, loud message or not, because a non-zero SessionStart hook is a
		// broken session rather than a governed one. That was true of every
		// path here and verified only by READING the source: this runner
		// discarded the exit code and no test in this file asserted one, so a
		// branch that started exiting non-zero (a stray `set -e` interaction, a
		// failing command at the end of a new branch) would have gone green.
		// Pinned behaviourally instead, across every fixture that runs through
		// this helper: missing rows, unknown rows, duplicates, an empty file
		// and the broken-payload branch.
		if err != nil {
			t.Fatalf("a hook must never fail the session, but the hook exited non-zero (%v); output:\n%s", err, out)
		}
		return string(out)
	}
}

// hookSlugs returns every distinct rule slug appearing anywhere in the hook's
// output. It scrapes the WHOLE output, including the injected rules.md prose
// (each rule ends with its slug in backticks) and the project's echoed file, so
// its presence is not proof a rule governs — only that the slug was mentioned
// somewhere. Fine for proving absence (an empty set really does mean the slug
// appears nowhere); use governedSwitchingOff to prove a governed delivery.
func hookSlugs(out string) map[string]bool {
	slugs := map[string]bool{}
	for _, m := range regexp.MustCompile(`(inv|floor)-[a-z-]+`).FindAllString(out, -1) {
		slugs[m] = true
	}
	return slugs
}

// newRulesFile is the file a project adopting Trellis writes (TRL-97, KTD9): a
// comment naming the opt-out row form and an empty [rules] table, each line
// ending in a newline. The hook's accept instruction and its remedies quote
// these bytes, and install.sh seeds them.
const newRulesFile = "# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }\n[rules]\n"

// quotedNewRulesFile returns the file a hook message tells the agent to write:
// the two backticked lines that follow "containing exactly these two lines",
// each given back its newline. Fatal when the message quotes no such file, so a
// reworded message fails here rather than comparing an empty string.
func quotedNewRulesFile(t *testing.T, msg string) string {
	t.Helper()
	const lead = "containing exactly these two lines"
	i := strings.Index(msg, lead)
	if i < 0 {
		t.Fatalf("the message quotes no file %q:\n%s", lead, msg)
	}
	spans := regexp.MustCompile("`([^`]*)`").FindAllStringSubmatch(msg[i:], 2)
	if len(spans) != 2 {
		t.Fatalf("the message names %q but quotes %d backticked lines after it, not two:\n%s", lead, len(spans), msg)
	}
	return spans[0][1] + "\n" + spans[1][1] + "\n"
}

// governedSwitchingOff reports whether context (a DECODED additionalContext —
// see nudgeContext) delivered the rules and framed the project file under the
// computed sentence naming exactly off (KTD1). It replaced deliveredRow, which
// looked for a `slug = { active = ...` row line: now that a row only switches a
// rule off, a row line proves nothing about which rules govern, and the
// sentence is what states it.
func governedSwitchingOff(context string, off ...string) bool {
	return strings.Contains(context, rulesLoadedSentinel) &&
		strings.Contains(context, "\n"+activationHeading+"\n\n"+expectedActivationSentence(off)+"\n")
}

// TestNoSlugsInPayloadFailsLoudly pins "fail loudly rather than govern silently
// on a partial payload" against a rules.md that tags no slug at all. Against an
// empty slug set every row would name an unknown rule and nothing could be
// switched off, so the session would look governed with no way to tell which
// rules it names. An earlier version of this hook quarantined every row against
// that empty set and ran the session ungoverned at exit 0.
func TestNoSlugsInPayloadFailsLoudly(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()
	// The one thing this fixture must break: no backticked trailing `inv-`/
	// `floor-` slug anywhere, which is exactly what staleness.sh's validator
	// scans rules.md for. Everything else about the payload stays valid —
	// including the terminator, which the hook now checks BEFORE it derives a
	// slug set (matching codex-context.mjs's order). Without this line the
	// terminator gate would capture the fixture and this test would pass for
	// the wrong reason, pinning nothing about no-slugs-in-payload.
	files["rules.md"] = "# Rules\n\nThis payload carries no rule slugs at all.\n" +
		rulesLoadedSentinel + "\n"

	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	proj := t.TempDir()
	if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
		t.Fatal(err)
	}
	// An ordinary, valid file — the defect is not in it.
	if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(hook)
	cmd.Dir = proj
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook exited non-zero (%v) — a hook must never fail the session: %s", err, out)
	}
	ctx := string(out)

	if !strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
		t.Fatalf("a payload with no rule slugs must fail loudly, not run ungoverned; got:\n%s", ctx)
	}
	if strings.Contains(ctx, activationHeading) {
		t.Errorf("a payload with no slugs must never reach the activation section — there is nothing to classify the file against; got:\n%s", ctx)
	}
	if got := hookSlugs(ctx); len(got) > 0 {
		t.Errorf("no rows may be injected when the payload itself cannot be validated against; got %v in:\n%s", keysOfBool(got), ctx)
	}
}

// TestAnUnreadableRulesInputIsRefusedLoudly pins two doors into one hazard: a
// file this hook needs exists but cannot be opened. An earlier version read
// both positionally, where awk dies printing nothing, and read that empty
// output as a mismatch to repair: every row was quarantined and the session ran
// ungoverned at exit 0, and an unreadable rules.toml fell through to a `cat`
// that failed just as quietly.
//
// The payload's rules.md is a broken install. The project's rules.toml is not,
// but a governed = false the hook cannot read cannot be ruled out (KTD5), so
// both exit through the loud door and deliver no rule and no activation section.
func TestAnUnreadableRulesInputIsRefusedLoudly(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: mode 0000 does not deny reads, so the fixture cannot be built")
	}
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()

	for _, tc := range []struct {
		name string
		// unreadable names the file the fixture chmods to 0000, relative to
		// whichever root owns it.
		pluginRel  string
		projectRel string
	}{
		{name: "the payload rules.md exists but cannot be read", pluginRel: "reference/rules.md"},
		{name: "the project rules.toml exists but cannot be read", projectRel: ".trellis/rules.toml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pluginRoot := t.TempDir()
			if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
				t.Fatal(err)
			}
			for name, body := range files {
				if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			proj := t.TempDir()
			if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
				t.Fatal(err)
			}
			// A valid file: the defect is never in it.
			if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(newRulesFile), 0o644); err != nil {
				t.Fatal(err)
			}

			target := filepath.Join(proj, filepath.FromSlash(tc.projectRel))
			if tc.pluginRel != "" {
				target = filepath.Join(pluginRoot, filepath.FromSlash(tc.pluginRel))
			}
			if err := os.Chmod(target, 0o000); err != nil {
				t.Fatal(err)
			}
			// Restore before TempDir cleanup, which must still be able to remove it.
			t.Cleanup(func() { _ = os.Chmod(target, 0o644) })
			if _, err := os.ReadFile(target); err == nil {
				t.Skipf("premise: %s is still readable at mode 0000 (root, or a filesystem without POSIX modes)", target)
			}

			cmd := exec.Command(hook)
			cmd.Dir = proj
			cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
			// stdout and stderr are read SEPARATELY here, unlike rulesTomlRun:
			// awk announces its own fatal "can't open file" on stderr, and the
			// host reads only stdout, which must still be exactly one JSON
			// envelope. Combining them would make the fixture look like
			// malformed output when it is in fact the diagnostic doing its job.
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("a hook must never fail the session, but the hook exited non-zero (%v); stdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
			}
			raw := stdout.String()
			if !strings.Contains(raw, "TRELLIS_RULES_NOT_LOADED") {
				t.Fatalf("an unreadable rules input must fail loudly:\n%s", raw)
			}
			ctx := nudgeContext(t, strings.TrimSpace(raw))
			if strings.Contains(ctx, rulesLoadedSentinel) || strings.Contains(ctx, activationHeading) {
				t.Errorf("no rule and no activation section may be delivered when a rules input could not be read:\n%s", ctx)
			}
			if got := hookSlugs(ctx); len(got) > 0 {
				t.Errorf("no rule slug may appear at all on this path; got %v in:\n%s", keysOfBool(got), ctx)
			}
		})
	}
}

// TestUnreadableHeaderNeverShipsRowsWithoutRules is the fifth instance of the
// class the two guards above close, and the worst-looking one: the payload
// assembly awk read $header POSITIONALLY, one line below the fixes for the
// other two, behind the same bare `-f` existence check. A $header that exists
// but yields nothing dies fatally and prints nothing, while the printfs and the
// row block around it carry on inside the same command substitution.
//
// Measured on the pre-fix hook four ways — mode 000, zero-byte, truncated above
// the `@rules.md` import, and the firm posture header since retired — the hook emitted
// sixteen activation rows, ZERO rules prose, no loud marker and exit 0. It is
// MORE dangerous than the reconciler blackouts, not less: the payload looks
// substantive, so nothing signals a problem. The agent is told exactly which
// sixteen rules are active and handed none of them. No permission trickery is
// required either — a header left truncated by an interrupted install.sh does
// it.
//
// The Codex hook has always refused this shape (readRequired's unreadable-file
// / missing-file, plus its explicit empty-prose check, plus
// invalid-placeholder-count for the truncated-above-the-import case), which is
// what makes the Claude-side gap an oversight rather than a design choice. This
// pins the matching behaviour on both halves.
func TestUnreadableHeaderNeverShipsRowsWithoutRules(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()
	// The import line the assembly resolves; a header truncated above it is
	// non-empty and still frames an activation section with no rules above it.
	const importLine = "@rules.md\n"
	head, _, found := strings.Cut(payloadFile(t, "trellis.md"), importLine)
	if !found || head == "" {
		t.Fatal("premise: trellis.md must carry an @rules.md import with prose above it")
	}

	// One header ships (TRL-97), so the firm-posture row this table carried
	// retired with the second header.
	for _, tc := range []struct {
		name    string
		mode    os.FileMode
		content string // "" with mode 0 means: leave the bytes, deny the read
	}{
		{name: "the header exists but cannot be read", mode: 0o000},
		{name: "the header is zero bytes", content: "", mode: 0o644},
		{name: "the header is truncated above its @rules.md import", content: head, mode: 0o644},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mode == 0o000 && os.Geteuid() == 0 {
				t.Skip("running as root: mode 0000 does not deny reads, so the fixture cannot be built")
			}
			pluginRoot := t.TempDir()
			if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
				t.Fatal(err)
			}
			for name, body := range files {
				if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			target := filepath.Join(pluginRoot, "reference", "trellis.md")
			if tc.mode != 0o000 {
				if err := os.WriteFile(target, []byte(tc.content), tc.mode); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Chmod(target, 0o000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(target, 0o644) })
				if _, err := os.ReadFile(target); err == nil {
					t.Skipf("premise: %s is still readable at mode 0000", target)
				}
			}

			proj := t.TempDir()
			if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
				t.Fatal(err)
			}
			// A valid project file every time: it is never the defect here,
			// which is the point — framing it alone is the failure.
			if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(configOnlyProjectRules), 0o644); err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command(hook)
			cmd.Dir = proj
			cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("a hook must never fail the session, but the hook exited non-zero (%v); stdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
			}
			raw := stdout.String()
			if !strings.Contains(raw, "TRELLIS_RULES_NOT_LOADED") {
				t.Fatalf("an unusable header must fail loudly, not frame the project file with no rules above it:\n%s", raw)
			}
			ctx := nudgeContext(t, strings.TrimSpace(raw))
			// The two halves of the blackout, asserted separately: no activation
			// section, and no file echoed inside one. Either alone would pass
			// over the pre-fix output, which carried a perfectly ordinary-looking
			// heading above a perfectly ordinary-looking row list.
			if strings.Contains(ctx, activationHeading) {
				t.Errorf("no activation section may be delivered when the rules prose could not be assembled:\n%s", ctx)
			}
			if strings.Contains(ctx, configOnlyProjectRules) {
				t.Errorf("the project file was echoed with no rules prose to govern by:\n%s", ctx)
			}
			if strings.Contains(ctx, rulesLoadedSentinel) {
				t.Errorf("nothing may claim the rules were loaded on this path:\n%s", ctx)
			}
		})
	}
}

// truncateRulesMdAfter returns the payload's rules.md cut immediately after its
// nth backticked slug tag — a file that is NON-EMPTY and carries real slugs, so
// an empty-slug-set test alone passes it, and that has lost its terminator, as
// a real truncation does.
func truncateRulesMdAfter(t *testing.T, n int) string {
	t.Helper()
	tag := regexp.MustCompile("`(inv|floor)-[a-z-]+`[ \t]*$")
	var kept []string
	seen := 0
	for _, line := range strings.Split(payloadFile(t, "rules.md"), "\n") {
		kept = append(kept, line)
		if tag.MatchString(line) {
			seen++
			if seen == n {
				break
			}
		}
	}
	if seen != n {
		t.Fatalf("premise: the payload rules.md must carry at least %d slug tags, found %d", n, seen)
	}
	return strings.Join(kept, "\n") + "\n"
}

// TestTruncatedRulesMdIsRefusedByItsOwnTerminator adds the check Codex has had
// since it shipped and the Claude hook did not: rules.md must carry exactly one
// `<!-- trellis:rules-loaded -->` terminator, as its final line.
//
// rules.md carries the terminator on its last line, so ANY truncation loses it
// — no slug arithmetic required, and no second payload file to compare against.
//
// Ordered BEFORE the slug derivation, matching codex-context.mjs and its
// comment's reasoning: a payload validated after it is trusted produces
// verdicts about the consumer's file for a defect that is the plugin's.
func TestTruncatedRulesMdIsRefusedByItsOwnTerminator(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()
	if !strings.HasSuffix(payloadFile(t, "rules.md"), rulesLoadedSentinel+"\n") {
		t.Fatal("premise: the shipped rules.md must end with the terminator this gate requires")
	}

	run := func(t *testing.T, rulesMd string) string {
		t.Helper()
		pluginRoot := t.TempDir()
		if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
			t.Fatal(err)
		}
		for name, body := range files {
			if name == "rules.md" {
				body = rulesMd
			}
			if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(configOnlyProjectRules), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("a hook must never fail the session, but the hook exited non-zero (%v); stdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
		}
		return nudgeContext(t, strings.TrimSpace(stdout.String()))
	}

	assertRefused := func(t *testing.T, ctx string) {
		t.Helper()
		if !strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("a payload without its terminator must be refused, not believed:\n%s", ctx)
		}
		if strings.Contains(ctx, activationHeading) {
			t.Errorf("no activation section may be delivered from a truncated payload:\n%s", ctx)
		}
	}

	t.Run("a rules.md cut short of its terminator is refused", func(t *testing.T) {
		assertRefused(t, run(t, truncateRulesMdAfter(t, 2)))
	})

	t.Run("a doubled terminator is refused too", func(t *testing.T) {
		assertRefused(t, run(t, payloadFile(t, "rules.md")+rulesLoadedSentinel+"\n"))
	})

	// The control: the shipped payload must sail through the new gate, and its
	// slug set must still find the rule the project file switches off.
	t.Run("the shipped payload passes its own terminator check", func(t *testing.T) {
		ctx := run(t, payloadFile(t, "rules.md"))
		if strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("the shipped payload must not be refused by the gate meant for broken ones:\n%s", ctx)
		}
		if !governedSwitchingOff(ctx, configOnlyProjectRow) {
			t.Errorf("the shipped payload must still govern, switching off the one rule the file names:\n%s", ctx)
		}
	})

	// The gate compared awk's RAW record against an exact ASCII string, so a
	// rules.md checked out or packaged with CRLF normalization carried a
	// trailing \r on its last line and was reported `not-last` — a full
	// blackout on a COMPLETE, CORRECT payload, which is the mirror image of
	// every defect this branch has been closing and just as useless to a
	// consumer, who is told nothing loaded and has nothing to fix.
	//
	// The fixture exists so that blindness cannot come back.
	t.Run("a healthy CRLF payload governs and is not refused", func(t *testing.T) {
		crlf := strings.ReplaceAll(payloadFile(t, "rules.md"), "\n", "\r\n")
		if !strings.Contains(crlf, "\r\n") || strings.Contains(payloadFile(t, "rules.md"), "\r") {
			t.Fatal("premise: the fixture must convert LF to CRLF and the shipped file must not already be CRLF")
		}
		ctx := run(t, crlf)
		if strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("a complete payload with CRLF line endings must govern, not black out:\n%s", ctx)
		}
		// The switched-off rule proves the slug set survived CRLF, not merely
		// that some prose arrived: a slug still carrying its \r would draw an
		// unknown-slug warning and switch nothing off.
		if !governedSwitchingOff(ctx, configOnlyProjectRow) || strings.Contains(ctx, "Trellis warning:") {
			t.Errorf("a CRLF payload must still govern, switching off the one rule the file names:\n%s", ctx)
		}
	})

	// ...and tolerating \r must not make the gate blind: a CRLF payload that
	// really is truncated is still refused.
	t.Run("a truncated CRLF payload is still refused", func(t *testing.T) {
		crlf := strings.ReplaceAll(truncateRulesMdAfter(t, 2), "\n", "\r\n")
		assertRefused(t, run(t, crlf))
	})
}

// TestPluginRootWithABackslashStillDeliversTheRules is the eighth instance of
// the silent-read class: `@rules.md` expanded through `while ((getline line <
// rules) > 0)` with the return value discarded.
//
// The trigger is not a permission but the `-v` channel. `awk -v`
// ESCAPE-PROCESSES its value — `awk -v v='/a\tb/c'` yields length 6, not 7 — so
// a CLAUDE_PLUGIN_ROOT containing a backslash reached that awk as a DIFFERENT
// path than the one every `-f` test and every positional read used, all of
// which passed. Measured with a root named `plug\tools`: 16 activation rows,
// 0 rules prose, exit 0, no marker — verbatim the damage shape the posture
// header guard exists to stop.
//
// The fix passes the paths through ENVIRON, which does no escape processing, so
// such a root now WORKS rather than merely failing loudly; the rc check remains
// for genuine read failures. The invariants pointer is asserted verbatim
// because it rode the identical mangling silently, and because the gsub-based
// substitution that carried it was half-wrong on this awk in the other
// direction — an unescaped `&` IS expanded, an unescaped backslash is NOT, so
// the escaping meant to protect both corrupted one. Substituting by index
// invokes no replacement semantics at all.
func TestPluginRootWithABackslashStillDeliversTheRules(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()

	for _, tc := range []struct {
		name string
		dir  string
	}{
		// The path the rc check never covered.
		{name: "a backslash in the plugin root", dir: "plug\\tools"},
		// An ampersand corrupted the invariants pointer through the same awk by
		// the opposite mechanism; both are asserted here so a fix for either
		// cannot silently break the other.
		{name: "an ampersand in the plugin root", dir: "R&D"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pluginRoot := filepath.Join(t.TempDir(), tc.dir)
			if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
				t.Fatal(err)
			}
			for name, body := range files {
				if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			proj := t.TempDir()
			if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(configOnlyProjectRules), 0o644); err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command(hook)
			cmd.Dir = proj
			cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("a hook must never fail the session, but the hook exited non-zero (%v); stdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
			}
			ctx := nudgeContext(t, strings.TrimSpace(stdout.String()))

			if strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
				t.Fatalf("a plugin root whose NAME contains a metacharacter is legitimate and every file on it is readable — it must govern, not refuse:\n%s", ctx)
			}
			// The blackout, asserted as itself: an activation section without
			// rules is the shape, so both halves are required.
			if !strings.Contains(ctx, rulesLoadedSentinel) {
				t.Errorf("the rules prose was not imported — this is the rows-without-rules blackout:\n%s", ctx)
			}
			if !governedSwitchingOff(ctx, configOnlyProjectRow) {
				t.Errorf("the project file was not classified on such a root:\n%s", ctx)
			}
			// Verbatim, not merely present: the escape hazard mangled this
			// pointer without changing its shape.
			wantPointer := filepath.Join(pluginRoot, "reference", "invariants.md")
			if !strings.Contains(ctx, "`"+wantPointer+"`") {
				t.Errorf("the invariants pointer must name the real path verbatim, want %q in:\n%s", wantPointer, ctx)
			}
		})
	}
}

// TestEveryLegitimateShapeStillGoverns is the over-refusal sweep for the payload
// guards this hook carries. Every one of them refuses loudly, and two of them
// once shipped defects that refused a HEALTHY input — a CRLF payload, and an
// unreadable comparison file. That direction is as bad for a consumer as the
// blackouts the guards exist to stop: TRELLIS_RULES_NOT_LOADED with nothing
// wrong to fix.
//
// So the ordinary shapes are pinned as a set rather than one at a time. None of
// them may produce that marker, and TRL-97 adds the other half of the claim:
// none of them is asked to change the file either (R15, R16). A file written for
// an older release governs every rule it does not switch off (AE7, AE8).
func TestEveryLegitimateShapeStillGoverns(t *testing.T) {
	run := rulesTomlRun(t)
	thisRepo := readFileT(t, "../.trellis/rules.toml")
	const oneTrueRow = "inv-minimal-first = { active = true }\n"

	// The fourteen-row set that predates the two newest rules.
	var fourteen strings.Builder
	fourteen.WriteString("[rules]\n")
	for _, slug := range assessableSlugs {
		if slug == "inv-deliberate-succession" || slug == "inv-no-orphan-followups" {
			continue
		}
		fourteen.WriteString(slug + " = { active = true }\n")
	}

	for _, tc := range []struct {
		name string
		rows string
		off  []string
		// declines marks the opt-out, whose correct answer is silence.
		declines bool
	}{
		{name: "the new file", rows: newRulesFile},
		{name: "this repository's own file, strictness and seeded_from included", rows: thisRepo},
		{name: "fourteen true rows from an older release", rows: fourteen.String()},
		{name: "an unknown row set true", rows: newRulesFile + "inv-not-a-real-rule = { active = true }\n"},
		{name: "a duplicate row", rows: newRulesFile + oneTrueRow + oneTrueRow},
		{
			name: "an opt-out beside a slug this plugin does not ship",
			rows: newRulesFile + "inv-renamed-slug = { active = false }\ninv-minimal-first = { active = false }\n",
			off:  []string{"inv-minimal-first"},
		},
		{
			// A file an earlier release repaired carries that release's
			// comments; they are comments, and draw nothing.
			name: "a file already carrying a quarantined row",
			rows: newRulesFile +
				"# inv-not-a-real-rule = { active = true }  # quarantined 2026-01-01: not in payload@000000000000.\n",
		},
		{name: "a hand-written file: a posture and no rows", rows: "strictness = \"firm\"\n"},
		{name: "an empty file", rows: ""},
		{name: "an explicit opt-out", rows: "governed = false\n", declines: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := run(t, tc.rows)
			if strings.Contains(out, "TRELLIS_RULES_NOT_LOADED") {
				t.Fatalf("a legitimate shape must never be refused as a broken payload:\n%s", out)
			}
			if tc.declines {
				// `governed = false` in a project with no static shape emits
				// NOTHING, by design: there is nothing already loaded to tell
				// the agent to disregard.
				if strings.TrimSpace(out) != "" {
					t.Errorf("a project that declined must be delivered nothing:\n%s", out)
				}
				return
			}
			ctx := nudgeContext(t, strings.TrimSpace(out))
			if !governedSwitchingOff(ctx, tc.off...) {
				t.Fatalf("want the rules delivered under the sentence %q:\n%s", expectedActivationSentence(tc.off), ctx)
			}
			// R15: whatever the file holds, nothing the hook says about it asks
			// for it to change. The echoed file is the project's own text and is
			// excluded.
			section := ctx[strings.Index(ctx, "\n"+activationHeading+"\n"):]
			if m := writeInstructionRe.FindString(strings.Replace(section, tc.rows, "", 1)); m != "" {
				t.Errorf("the hook said %q about a legitimate file; nothing may ask for .trellis/rules.toml to change:\n%s", m, section)
			}
		})
	}

	// THE VENDORED-BUNDLE PATH, which every shape above misses: rulesTomlRun
	// always seeds a project file. With the bundle under the project's own
	// .claude/skills/ and no file, every rule applies and there is no file to
	// frame (KTD6). This path used to stand the shipped default rows in for the
	// project file, and the refusals that guarded that stand-in retired with it.
	t.Run("no project file, bundle vendored in the project: every rule governs", func(t *testing.T) {
		pluginRoot, proj := writeDefaultsShapeProject(t)
		ctx := claudeContextFor(t, pluginRoot, proj)
		if !deliveredEveryRuleWithNoFile(ctx) {
			t.Fatalf("want every rule delivered, no TRELLIS_ marker and no activation section:\n%s", ctx)
		}
		if _, err := os.Stat(filepath.Join(proj, ".trellis", "rules.toml")); !os.IsNotExist(err) {
			t.Errorf("the hook wrote a .trellis/rules.toml into a project that had none (stat err: %v)", err)
		}
	})
}

// TestBrokenPayloadIsNeverSilent is the behavioural half of the guard TRL-33
// asks for, and the primary deliverable of that issue.
//
// Fifteen defects, counted 2026-09-03, shared one shape: an absent, empty,
// truncated or unreadable payload input reached downstream logic and the
// session ran ungoverned at exit 0 with nothing signalling a problem. The
// count is decision-0087's own measurement, not a total either predecessor
// states — decision-0083 records eight and decision-0084 several more without
// adding them up, and four were still open when it was taken.
//
// ALMOST NONE was found by this suite or by reading the code — every one was
// found by a reviewer RUNNING the hook against a deliberately broken input,
// one file at a time, as they thought of it. This test is that reviewer,
// written down.
//
// THE FILE LIST IS READ FROM THE BUNDLE, never hardcoded. That is what makes
// the NEXT instance caught by construction: a payload file added to
// plugins/trellis/reference/ later joins this matrix without anyone
// remembering, and if the hook starts reading it unguarded, this test says so.
//
// DELIBERATE SUBSET of decision-0073 D1's delivery states, stated here where it
// is done, per D1's own per-component relevance clause. It runs on the two
// states where the hook DELIVERS — config-only and vendored-defaults — because
// only there is "silence is wrong" unconditional. On path A a healthy, current
// overlay is legitimately silent, and a payload file that path never reads
// would make an unconditional assertion there simply false; that path is
// covered by its own named subtests in TestStalenessHook instead.
//
// The healthy and CRLF rows run in this same table ON PURPOSE. Two of the
// fifteen were the INVERSE defect — a guard that refused a healthy payload, one
// of them over CRLF line endings exactly — so a matrix that only proved the
// hook refuses broken input would be half a guard, and the missing half is the
// one that has shipped twice.
func TestBrokenPayloadIsNeverSilent(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	bundleRef := filepath.Join(vendoredBundleAbs(t), "reference")
	entries, err := os.ReadDir(bundleRef)
	if err != nil {
		t.Fatal(err)
	}
	var payloadNames []string
	for _, e := range entries {
		if !e.IsDir() {
			payloadNames = append(payloadNames, e.Name())
		}
	}
	sort.Strings(payloadNames)
	if len(payloadNames) < 10 {
		t.Fatalf("found only %d payload files in %s — the enumeration is broken, and a matrix over nothing passes", len(payloadNames), bundleRef)
	}

	// breaks are applied to ONE payload file at a time. "healthy" and "crlf"
	// break nothing; they are the discrimination controls.
	breaks := []string{"healthy", "crlf", "absent", "zero-byte", "unreadable", "truncated"}

	// shapes are the two delivering project layouts. Both put the plugin root
	// inside the repo under .claude/skills/, which decision-0070 D6 requires
	// for the defaults path; config-only adds the project file that claims
	// path B.
	shapes := []string{"config-only", "vendored-defaults"}

	for _, name := range payloadNames {
		for _, brk := range breaks {
			for _, shape := range shapes {
				t.Run(name+"/"+brk+"/"+shape, func(t *testing.T) {
					if brk == "unreadable" && os.Geteuid() == 0 {
						t.Skip("running as root: mode 0000 does not deny reads, so the fixture cannot be built")
					}
					proj := t.TempDir()
					pluginRoot := filepath.Join(proj, ".claude", "skills", "trellis")
					refDir := filepath.Join(pluginRoot, "reference")
					if err := os.MkdirAll(refDir, 0o755); err != nil {
						t.Fatal(err)
					}
					for _, f := range payloadNames {
						body := readFileT(t, filepath.Join(bundleRef, f))
						target := filepath.Join(refDir, f)
						if f != name {
							if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
								t.Fatal(err)
							}
							continue
						}
						switch brk {
						case "healthy":
							if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
								t.Fatal(err)
							}
						case "crlf":
							crlf := strings.ReplaceAll(body, "\n", "\r\n")
							if err := os.WriteFile(target, []byte(crlf), 0o644); err != nil {
								t.Fatal(err)
							}
						case "absent":
							// written by no one — that IS this row
						case "zero-byte":
							if err := os.WriteFile(target, nil, 0o644); err != nil {
								t.Fatal(err)
							}
						case "unreadable":
							if err := os.WriteFile(target, []byte(body), 0o000); err != nil {
								t.Fatal(err)
							}
							t.Cleanup(func() { _ = os.Chmod(target, 0o644) })
							// The premise, checked rather than assumed: a
							// filesystem without POSIX modes would make this
							// row silently test the healthy case instead.
							if _, err := os.ReadFile(target); err == nil {
								t.Skipf("premise: %s is still readable at mode 0000", target)
							}
						case "truncated":
							if err := os.WriteFile(target, []byte(body[:len(body)/2]), 0o644); err != nil {
								t.Fatal(err)
							}
						}
					}
					if shape == "config-only" {
						if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(configOnlyProjectRules), 0o644); err != nil {
							t.Fatal(err)
						}
					}

					cmd := exec.Command(hook)
					cmd.Dir = proj
					cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
					var stdout, stderr bytes.Buffer
					cmd.Stdout = &stdout
					cmd.Stderr = &stderr
					if err := cmd.Run(); err != nil {
						t.Fatalf("a hook must never fail the session, but it exited non-zero (%v)\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
					}
					raw := strings.TrimSpace(stdout.String())
					if raw == "" {
						t.Fatalf("SILENT: breaking %s (%s) on the %s path produced no output at all — this is the whole defect class, and it is what TRL-33 measured on main", name, brk, shape)
					}
					ctx := nudgeContext(t, raw)

					loud := strings.Contains(ctx, "TRELLIS_")
					// A COMPLETE injection is the rules body AND, where there is
					// a project file, its activation section naming the rule it
					// switches off. The damage shape this catches is neither of
					// the two obvious ones: a payload that looks substantive and
					// delivers an activation list with NO RULES above it —
					// measured four ways on the posture header, at exit 0 with
					// no marker. With no project file there is no section.
					governed := strings.Contains(ctx, "`floor-intent-gate`") &&
						strings.Contains(ctx, "`inv-minimal-first`")
					if shape == "config-only" {
						governed = governed && governedSwitchingOff(ctx, configOnlyProjectRow)
					} else {
						governed = governed && deliveredEveryRuleWithNoFile(ctx)
					}

					if !loud && !governed {
						t.Fatalf("breaking %s (%s) on the %s path produced neither a complete governed injection nor a loud refusal — a partial delivery with nothing signalling it:\n%s", name, brk, shape, ctx)
					}

					// The other direction, which has shipped twice: a healthy
					// payload must NOT be refused. CRLF is not a corruption —
					// a collaborator on core.autocrlf=true has one by default.
					if brk == "healthy" || brk == "crlf" {
						if !governed {
							t.Fatalf("a %s payload must govern, and %s was not delivered complete on the %s path:\n%s", brk, name, shape, ctx)
						}
						if strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
							t.Fatalf("a %s payload was refused as broken — this is the over-correction, and it is as bad for a consumer as the silence:\n%s", brk, ctx)
						}
					}
				})
			}
		}
	}
}

// TestNoPayloadReadBypassesTheGateway is the structural half of the guard
// TRL-33 asks for. TestBrokenPayloadIsNeverSilent is the behavioural half and
// the primary one; this half is what makes a read added LATER safe, at the
// moment it is written rather than the moment someone breaks it.
//
// Two rules, and between them they have no false positives on the shapes this
// hook legitimately uses:
//
//  1. Every shell variable assigned a path under the installed plugin must be
//     passed to payload_read. This is what catches a NEW payload file — you
//     cannot read one without naming it, and naming it puts you in the scan.
//  2. No literal "$plugin/..." may be a file operand outside payload_read.
//     This catches the shortcut around rule 1: reading the path inline without
//     ever binding it to a variable.
//
// What the rules deliberately do NOT forbid is a SECOND read of a variable
// that has already been through the gateway — $rules is read positionally by
// awk after payload_read has proved it readable and non-empty, and that
// is the safe pattern rather than a bypass. The hazard the fifteen instances
// shared was an UNCHECKED first read, not a checked one repeated.
func TestNoPayloadReadBypassesTheGateway(t *testing.T) {
	body, err := os.ReadFile("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	src := string(body)

	gateStart := strings.Index(src, "\npayload_read() {")
	if gateStart < 0 {
		t.Fatal("payload_read() not found in staleness.sh — the scan is broken, and a guard that reads nothing passes")
	}
	gateEnd := strings.Index(src[gateStart:], "\n}\n")
	if gateEnd < 0 {
		t.Fatal("payload_read() has no closing brace at column 0 — the scan is broken")
	}
	gateLo := strings.Count(src[:gateStart], "\n") + 1
	gateHi := gateLo + strings.Count(src[gateStart:gateStart+gateEnd], "\n")

	// Rule 1.
	//
	// ANCHORED ON A DELIMITER, NOT ON LINE START. `^[ \t]*([a-z_]+)="\$plugin/`
	// was the first version and it had a blind spot review found on #262: the
	// posture header was assigned inside a `case` arm —
	// `firm) header="$plugin/reference/<firm header>" ;;` — so `$header`, a real
	// payload path variable, was invisible to the scan. It happened to be
	// guarded anyway, but a NEW payload path introduced the same way would have
	// bypassed both rules undetected, which is the one thing this test exists to
	// prevent. The mutation that proved the guard used a line-start assignment,
	// so it never exercised the shape that was missing: a guard is only known to
	// work against the mutations you actually try.
	assignRe := regexp.MustCompile(`(?m)(?:^|[ \t;)&|])([a-z_]+)="\$plugin/`)
	found := assignRe.FindAllStringSubmatch(src, -1)
	seen := map[string]bool{}
	var names []string
	for _, m := range found {
		if seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		names = append(names, m[1])
	}
	sort.Strings(names)
	// Named, not counted. A count says nothing about WHICH variables were found,
	// and the previous version's floor passed while missing $header because an
	// unrelated variable filled the slot its own failure message attributed to
	// the posture header. Every $plugin/-rooted payload path staleness.sh reads
	// today is listed; a scan that stops seeing one of them fails here rather
	// than passing quietly on a smaller set.
	//
	// $plugin/-ROOTED is the scope of both this floor and the scan above, and
	// saying "every payload path" overstated it. The gateway's own definition
	// (staleness.sh:122-123) counts a vendored copy inside the consuming
	// repository as payload too, and two such reads — payload_read
	// "$internal/$f" and payload_read "$legacy" — are invisible here because
	// neither variable is assigned from $plugin/. They go through the gateway
	// today; what is unguarded is a FUTURE one that does not, and rule 2 below
	// has the same blind spot for the same reason. Widening this list is not
	// the fix (adding `internal` fails the guard immediately); widening both
	// patterns to the overlay root would be, and is not attempted here.
	//
	// TRL-97 dropped $preset and $toml from this list: the comparison preset
	// retired, and $toml is the project's own file on every path, never a path
	// inside the plugin.
	for _, want := range []string{"header", "inv", "ref", "rules"} {
		if !seen[want] {
			t.Errorf("the payload-path scan no longer sees $%s — it found %v, and a guard that stops seeing a payload read passes on nothing", want, names)
		}
	}
	for _, name := range names {
		if !strings.Contains(src, `payload_read "$`+name+`"`) {
			t.Errorf("$%s holds a path inside the installed plugin but is never passed to payload_read — every payload read goes through the gateway, so that an absent, empty or unreadable file is refused loudly instead of reaching downstream logic", name)
		}
	}

	// Rule 2. A literal "$plugin/... as the operand of a read command or a file
	// test. `emit` lines are excluded: they NAME payload paths in prose (a
	// refusal tells the reader which file it could not read), which is text,
	// not a read.
	inlineRe := regexp.MustCompile(`(cat|head|tail|sed|awk|grep|wc|sort|\[ *!? *-[fser])\b[^\n]*"\$plugin/`)
	for i, line := range strings.Split(src, "\n") {
		n := i + 1
		if n >= gateLo && n <= gateHi {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "emit ") {
			continue
		}
		if inlineRe.MatchString(line) {
			t.Errorf("staleness.sh:%d opens a payload path inline instead of through payload_read:\n  %s", n, trimmed)
		}
	}
}

// TestEveryReadRequiredStatesWhatEmptyMeans is the Codex half of the same
// guard. readRequired has been the model for this all along — it reports
// missing-file and unreadable-file where the shell hook used to swallow both —
// but a zero-byte file came back as { value: "" }, a SUCCESS, and emptiness was
// caught only by post-checks each caller remembered to write. That is guarded
// by remembering, which is the failure mode TRL-33 was filed about — see
// decision-0087 for the count and how it was arrived at.
//
// The third argument makes the default loud: a call that says nothing gets
// empty-file. A call site where empty is legitimate — the project's own
// .trellis/rules.toml, the supported hand-written-partial shape — says so in
// its own source, where a reader sees it.
func TestEveryReadRequiredStatesWhatEmptyMeans(t *testing.T) {
	body, err := os.ReadFile("../plugins/trellis/hooks/codex-context.mjs")
	if err != nil {
		t.Fatal(err)
	}
	src := string(body)

	// Every call site is an assignment, which is what separates them from the
	// `function readRequired(...)` definition.
	callRe := regexp.MustCompile(`(?s)=\s*readRequired\(([^;]*?)\);`)
	calls := callRe.FindAllStringSubmatch(src, -1)
	if len(calls) < 2 {
		t.Fatalf("found only %d readRequired call sites — the scan is broken, and a guard that reads nothing passes", len(calls))
	}
	// The scan must see every occurrence but the definition. A call whose result
	// is not assigned would otherwise slip past the pattern above unexamined,
	// which is the failure mode of a source guard: passing on what it cannot see.
	if occurrences := strings.Count(src, "readRequired("); occurrences != len(calls)+1 {
		t.Fatalf("codex-context.mjs mentions readRequired( %d times but the scan matched %d call sites plus one definition — something is calling it in a shape this guard cannot read", occurrences, len(calls))
	}
	for _, m := range calls {
		args := strings.Join(strings.Fields(m[1]), " ")
		if !strings.Contains(args, "emptyError") && !strings.Contains(args, "emptyIsValid") {
			t.Errorf("a readRequired call site does not say what an empty read means — pass { emptyError: ... } or { emptyIsValid: true }:\n  readRequired(%s)", args)
		}
	}
}

// TRL-43. The hook read .trellis/rules.toml for the `governed` key before every
// delivery path, and that read opened the file before anything checked it was
// a regular file — so a FIFO at that path blocked the SessionStart hook until
// the host gave up on it. Reproduced on main (49d938b): `timeout 10` on the
// real hook against a `mkfifo .trellis/rules.toml` project exits 124 every
// time. Every read of that file is now behind a regular-file check, and a path
// that exists but is not a regular file is refused loudly rather than read as
// "no rules.toml" (which it is not) or as an opt-out (which it is not either).
//
// A real FIFO, not a stand-in: the defect is the open blocking, and only a FIFO
// blocks. The directory row cannot hang, so it pins the disposition alone.
// Verified by mutation: with the `-f && -r` guard removed from the governed
// read, the FIFO row hangs and this test reports it at the timeout.
func TestStalenessHookDoesNotBlockOnNonRegularRulesToml(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range payloadFiles() {
		if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name string
		make func(t *testing.T, path string)
	}{
		{"a FIFO", func(t *testing.T, path string) {
			if err := syscall.Mkfifo(path, 0o644); err != nil {
				t.Skipf("cannot create a FIFO here: %v", err)
			}
		}},
		{"a directory", func(t *testing.T, path string) {
			if err := os.Mkdir(path, 0o755); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name+" at .trellis/rules.toml", func(t *testing.T) {
			proj := t.TempDir()
			if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
				t.Fatal(err)
			}
			tc.make(t, filepath.Join(proj, ".trellis", "rules.toml"))

			cmd := exec.Command(hook)
			cmd.Dir = proj
			cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
			// Its own process group, so a timeout kills the sed blocked on the
			// FIFO too and not just the shell that spawned it — an orphaned
			// reader would hold the output pipe and stall Wait forever.
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			var out bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &out
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("a hook must never fail the session, but the hook exited non-zero (%v); output:\n%s", err, out.String())
				}
			case <-time.After(30 * time.Second):
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				// Wait joins the copier goroutines that write into out; reading
				// the buffer before it returns is a data race. The group kill
				// above is what makes this Wait return at all.
				<-done
				t.Fatalf("the hook hung on %s at .trellis/rules.toml — a read opened the path before checking it is a regular file\noutput so far: %s", tc.name, out.String())
			}

			got := out.String()
			if !strings.Contains(got, "TRELLIS_RULES_NOT_LOADED") {
				t.Errorf("%s at .trellis/rules.toml must be refused loudly; got:\n%s", tc.name, got)
			}
			if !strings.Contains(got, "not a regular file") {
				t.Errorf("the refusal must say what is wrong with the path — that it is not a regular file; got:\n%s", got)
			}
			// Neither of the two things it is not: the "no rules.toml" branch
			// (project-scope defaults, or the user-scope announcement) would be
			// wrong about the reader's state, and its remedy — write the file —
			// would block on the very FIFO it is describing.
			for _, wrong := range []string{"has no .trellis/rules.toml", "TRELLIS_NOT_YET_GOVERNING", "TRELLIS_NOT_GOVERNING", "shipped defaults"} {
				if strings.Contains(got, wrong) {
					t.Errorf("%s at .trellis/rules.toml must not be read as a missing file or as an opt-out (found %q); got:\n%s", tc.name, wrong, got)
				}
			}
			// The refusal names floor-intent-gate in its remedy, as every
			// destructive remedy in this hook must, so a slug scrape is the
			// wrong probe here: assert on the delivery shape instead — the
			// posture header, the rules terminator, and an activation section,
			// on the DECODED context.
			ctx := nudgeContext(t, strings.TrimSpace(got))
			for _, s := range []string{"**How strictly to follow them:**", "<!-- trellis:rules-loaded -->", activationHeading} {
				if strings.Contains(ctx, s) {
					t.Errorf("no rule may be delivered over a rules.toml the hook could not read (found %q); got:\n%s", s, ctx)
				}
			}
			// KTD10: the remedy names the new file rather than a preset, and
			// keeps its confirmation clause.
			if strings.Contains(ctx, "rules-b.toml") {
				t.Errorf("the remedy still names the retired preset:\n%s", ctx)
			}
			if quoted := quotedNewRulesFile(t, ctx); quoted != newRulesFile {
				t.Errorf("the remedy must quote the new file byte for byte\n got: %q\nwant: %q", quoted, newRulesFile)
			}
			if !strings.Contains(ctx, "explicit confirmation") {
				t.Errorf("the remedy lost its confirmation clause:\n%s", ctx)
			}
		})
	}
}

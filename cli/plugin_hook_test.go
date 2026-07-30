package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
	shipped := payloadFiles()["version"]
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
					"trellis.md": payloadFiles()["trellis-b.md"],
					"rules.md":   payloadFiles()["rules.md"],
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
		if !strings.Contains(v.HookSpecificOutput.AdditionalContext, "/trellis:setup") {
			t.Errorf("message should point at /trellis:setup: %q", v.HookSpecificOutput.AdditionalContext)
		}
		return v.HookSpecificOutput.AdditionalContext
	}

	t.Run("no overlay is silent", func(t *testing.T) {
		if out := run(t, "", ""); out != "" {
			t.Errorf("want silent, got %q", out)
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
		nudge(t, run(t, ".trellis/version", current))
		nudge(t, run(t, ".trellis/version", "payload@000000000000"))
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
	t.Run("empty stamp is silent", func(t *testing.T) {
		if out := run(t, ".trellis/internal/version", ""); out != "" {
			t.Errorf("want silent (nothing to compare), got %q", out)
		}
		if out := run(t, ".trellis/version", ""); out != "" {
			t.Errorf("want silent (empty legacy stamp), got %q", out)
		}
	})
	t.Run("unreadable plugin reference is silent", func(t *testing.T) {
		proj := t.TempDir()
		p := filepath.Join(proj, ".trellis", "internal", "version")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("payload@abc\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		writeVendoredPayload(t, filepath.Dir(p))
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+t.TempDir())
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v): %s", err, out)
		}
		if strings.TrimSpace(string(out)) != "" {
			t.Errorf("want silent (can't read the installed payload stamp), got %q", out)
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
		if err := os.WriteFile(toml, []byte(payloadFiles()["rules-b.toml"]), 0o644); err != nil {
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
			"[rules]",                     // the project's live rows
			"floor-intent-gate",
		} {
			if !strings.Contains(ctx, want) {
				t.Errorf("injected context missing %q", want)
			}
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
			".trellis/rules.toml":       payloadFiles()["rules-b.toml"],
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
		if err := os.WriteFile(filepath.Join(dir, "trellis.md"), []byte(payloadFiles()["trellis-b.md"]), 0o644); err != nil {
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
	for name, key := range map[string]string{"trellis.md": "trellis-b.md", "rules.md": "rules.md"} {
		if err := os.WriteFile(filepath.Join(internalDir, name), []byte(payloadFiles()[key]), 0o644); err != nil {
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
func TestStalenessHookStandsDownForInstallPath(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()

	// A complete plugin root: path B reads the header, the rules and the stamp.
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"version", "rules.md", "trellis-a.md", "trellis-b.md", "invariants.md"} {
		if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(files[name]), 0o644); err != nil {
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
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
			body := "# rendered by install.sh\n\n" + files["rules.md"] +
				"\n@../../.trellis/rules.toml\n" +
				"\n<!-- trellis:rendered-from " + strings.TrimSpace(files["version"]) + " -->\n"
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

	// Placement guard. A project MIGRATING off a vendored overlay can hold both
	// artifacts at once, and path A's staleness nudge is the only signal it gets.
	// decision-0035's floor is that drift is made visible, not silent — so path C
	// must sit AFTER path A, not before it.
	t.Run("migrating project: a stale overlay still nudges, even with the install artifact", func(t *testing.T) {
		proj := t.TempDir()
		internal := filepath.Join(proj, ".trellis", "internal")
		if err := os.MkdirAll(internal, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(internal, "version"), []byte("payload@000000000000\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		for name, key := range map[string]string{"trellis.md": "trellis-b.md", "rules.md": "rules.md"} {
			if err := os.WriteFile(filepath.Join(internal, name), []byte(files[key]), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".claude", "rules", "trellis.md"), []byte("# rendered\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v): %s", err, out)
		}
		if !strings.Contains(string(out), "stale") {
			t.Fatalf("a migrating project loses its ONLY drift signal if path C preempts path A; got:\n%s", out)
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
		body := files["rules.md"] + "\n@../../.trellis/rules.toml\n" +
			"\n<!-- trellis:rendered-from payload@000000000000 -->\n"
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
		if !strings.Contains(s, "000000000000") || !strings.Contains(s, strings.TrimSpace(files["version"])) {
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
		body := files["rules.md"] + "\n@../../.trellis/rules.toml\n" +
			"\n<!-- trellis:rendered-from " + strings.TrimSpace(files["version"]) + " -->\n"
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

	// Codex P1: a file cut immediately AFTER the rules-body sentinel passed the
	// old guard while having lost the import line and the stamp — so the hook
	// stood down and NO activation rows were ever delivered. The boundary is the
	// END of the file, not the end of the rules body.
	t.Run("truncated after the sentinel: incomplete, and said out loud", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		for _, body := range []string{
			files["rules.md"], // sentinel present, import and stamp lost
			files["rules.md"] + "\n@../../.trellis/rules.toml\n", // stamp lost
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

	t.Run("a zero-byte rendered file is not delivery: the hook still delivers", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
		if !strings.Contains(string(out), ruleSlug) {
			t.Fatalf("an EMPTY rendered file silenced the hook — the session would run ungoverned while both the installer and the hook claim rules are loaded; got:\n%s", out)
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
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
			if !strings.Contains(string(out), ruleSlug) {
				t.Fatalf("a truncated rendered file (%d bytes) silenced the hook — ungoverned session, with the stand-down message claiming rules are loaded; got:\n%s", len(body), out)
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

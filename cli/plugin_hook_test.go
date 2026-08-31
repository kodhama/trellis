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
	// The rest of the payload, copied from the real bundle rather than stubbed.
	// This fixture used to carry `version` alone, which silently disarmed every
	// assertion downstream of a payload read: the hook would bail with
	// TRELLIS_RULES_NOT_LOADED before it could inject anything, so a test
	// asserting "no rules were injected" passed for the wrong reason. Mutation
	// found it — making the announcing turn inject the full rule set left this
	// file green while a hand-run of the same mutation leaked twelve slugs.
	for _, f := range []string{"rules.md", "trellis-a.md", "trellis-b.md", "rules-b.toml", "invariants.md"} {
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
		if strings.Contains(out, "inv-directional-flow") {
			t.Errorf("no rule may be injected on the announcing turn — the message promises \"will be governed\", so governing already would make it false: %q", out)
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
// renderedFile builds the fixture from the SHIPPED PAYLOAD BYTES, not from Go
// string literals. The literal version drifted immediately — it emitted only the
// H1 where install.sh emits trellis-b.md's whole head, so the fixture asserted a
// footer referring to "the posture sentence above" over a file that had none —
// and no test could notice, because the fixture was the only definition of
// correct. Derived from the same payload install.sh reads, it drifts only if the
// payload does, which is the point.
func renderedFile(files map[string]string, stamp string) string {
	head, tail, _ := strings.Cut(files["trellis-b.md"], "@rules.md\n")
	return "<!-- trellis:rendered-begin -->\n" + head + files["rules.md"] + tail +
		"<!-- trellis:rendered-footer -->\n" +
		"**If the posture sentence above and the rows below disagree, the rows win:**\n" +
		"the `strictness` key in `.trellis/rules.toml` is authoritative.\n" +
		"\n## Project rule activation\n\n@../../.trellis/rules.toml\n" +
		"\n<!-- trellis:rendered-from " + stamp + " -->\n"
}

// TestRowMismatchRemedyIsNotDestructive: a Claude review finding on #227. The
// row-mismatch branch used to name /trellis:setup as the remedy, and the
// retirement rewrote that to "copy reference/rules-b.toml over
// .trellis/rules.toml".
//
// rules-b.toml is the ADAPTIVE preset with every row active, and `strictness`
// selects which header the hook injects. So that instruction silently flips a
// firm-posture project back to adaptive and turns every hand-disabled rule back
// on — no diff, no confirmation. The skill it replaced diffed and asked, per
// floor-intent-gate. And this branch fires on the ordinary upgrade path: any
// time the shipped catalog gains a rule, every existing rules.toml mismatches.
//
// Retiring a confirm-gated writer is fine. Replacing its remedy with an
// unconditional clobber is not, and it is the kind of loss that shows up as a
// consumer's governance quietly resetting rather than as a failure.
// TestDocumentedPostureRecipeActuallyGoverns: a Codex P1 on #227, and the
// sharpest finding on it — the documented replacement for the retired skill
// BROKE governance if followed literally.
//
// decision-0072's first draft said the replacement was "one sentence: edit
// .trellis/rules.toml — strictness = \"firm\"". But the hook validates the row
// set against the shipped catalog and injects NOTHING when a slug is missing, so
// a project-scope install sitting at a governed full row set goes to ZERO the moment
// someone hand-writes a file containing strictness alone. The retired skill's §1
// copied a whole preset first; that step was the part that had to survive, and
// the record had described it away.
//
// This test pins the recipe end to end, in both directions: the partial file
// must fail loudly, and the documented copy-then-edit must deliver every rule
// rules at the requested posture.
//
// Narrowed by the reconciliation change (TRL-20/TRL-2/TRL-27, same as
// TestRepairRemedyCoversEveryMismatchKind above): "the partial file must fail
// loudly" no longer holds — a hand-written partial file is exactly a `missing:`
// mismatch, and that is now reconciled rather than refused. Its half of this
// pin is retired; the surviving subtests (copy-then-edit, at both postures,
// and disabling a row) are unaffected, since none of them exercises a
// mismatch. The retired half's intent survives in
// TestSlugMismatchStillDeliversEveryRule's "a missing row does not black out
// the other rules".
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
// body. Ruling B (TRL-20 Task 2) requires both destructive-instruction guards
// below to scan it: the repair mandate rides into the agent's context through
// a printf inside this block rather than through emit "...", so a guard that
// reads only emit strings never saw it. The whole safety argument for leaving
// the repair ungated is "no deletion verb reaches the agent" — that argument
// only holds if every channel reaching the agent is enforced, not just one.
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
// quoted spans. A passthrough like `printf '%s\n' "$reconciled"` reconstructs
// to something like "%s\n$reconciled" — inert for verb-scanning, so it is not
// filtered out specially; the data it actually carries at runtime (TOML rows,
// a slug report) is not English prose and cannot contain an instruction.
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
	violations, gated := ungatedDestructiveMessages(msgs)
	for _, msg := range violations {
		t.Errorf("this message instructs a mutation with no confirmation gate — an autonomous "+
			"agent can act on it against files the consumer owns (floor-intent-gate):\n%s", msg)
	}
	// A floor, not a ceiling: the count only ever grows as remedies are added, so
	// a drop means the regex or the verb list stopped matching, not that the
	// script got safer. Advanced 11 → 13 when decision-0073 D2 added the two
	// inline-shape messages (the S4 refusal and the inline+rendered conflict).
	// Retreated 13 → 12 for the reconciliation change (TRL-20): the gated
	// TRELLIS_RULES_NOT_LOADED mismatch remedy this counted — "for missing:, add
	// those slugs; for unknown:, remove those rows; for duplicate:, delete the
	// extra occurrences" — is GONE, not merely reworded; a mismatch is now
	// reconciled in memory (add/quarantine, never delete) rather than refused, so
	// there is nothing destructive left to gate on that path. Unmoved by adding
	// the payload printf messages (Ruling B, TRL-20 Task 2): the repair mandate
	// they carry is written to avoid every verb in the list, by construction —
	// see TestReconciledRepairIsMandatedAndReported's "no deletion verb enters
	// the emit" subtest for the behavioural half of that claim.
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
	// mandate printf text and run it through the identical
	// payloadAssembly -> payloadPrintfMessages -> ungatedDestructiveMessages
	// pipeline the scan above uses. If that pipeline is broken in any of the
	// ways above, this injected message is never reached and the subtest
	// passes on nothing — the same failure mode this whole finding is about
	// — so the premise check below is the load-bearing part: it fails loudly
	// if the mutation itself did not land.
	t.Run("the payload channel is actually enforced, not merely counted", func(t *testing.T) {
		const marker = `printf 'Write .trellis/rules.toml with exactly the rows shown above`
		mutated := strings.Replace(string(body), marker,
			`printf 'Delete the unknown rows now, then write .trellis/rules.toml with exactly the rows shown above`, 1)
		if mutated == string(body) {
			t.Fatal("premise: the mandate printf text to mutate was not found in staleness.sh — the case would prove nothing")
		}
		mutatedMsgs := payloadPrintfMessages(payloadAssembly(t, mutated))
		mutatedViolations, _ := ungatedDestructiveMessages(mutatedMsgs)
		found := false
		for _, v := range mutatedViolations {
			if strings.Contains(v, "Delete the unknown rows now") {
				found = true
				break
			}
		}
		if !found {
			t.Error("an ungated deletion instruction injected into the payload printf channel was not caught — " +
				"the payload channel is not actually being scanned, only its message count is being checked")
		}
	})
}

// TestDocumentedPostureRecipeActuallyGoverns: a Codex P1 on #227, and the
// sharpest finding on it — the documented replacement for the retired skill
// BROKE governance if followed literally.
//
// decision-0072's first draft said the replacement was "one sentence: edit
// .trellis/rules.toml — strictness = \"firm\"". But the hook validates the row
// set against the shipped catalog and injects NOTHING when a slug is missing, so
// a project-scope install sitting at a governed full row set goes to ZERO the moment
// someone hand-writes a file containing strictness alone. The retired skill's §1
// copied a whole preset first; that step was the part that had to survive, and
// the record had described it away.
//
// This test pins the recipe end to end, in both directions: the partial file
// must fail loudly, and the documented copy-then-edit must deliver every rule
// at the requested posture.
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
	gated := 0
	for _, msg := range msgs {
		if !strings.Contains(msg, "delete") {
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
	// The repair mandate's printf messages must contribute ZERO deletion hits —
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
			t.Errorf("the repair mandate's payload printf text must never instruct a deletion (the reconciliation is additive/commenting-only, which is what keeps it ungated):\n%s", msg)
		}
	}
	t.Logf("checked %d deletion-instructing messages of %d (emit + payload printf)", gated, len(msgs))
}

// TestRepairRemedyCoversEveryMismatchKind and the opt-out shape: two Codex P2s
// on #227, both the same class as the five before them — a remedy that does not
// cover the state the reader is actually in.
//
// $slug_report emits THREE kinds (missing:, unknown:, duplicate:) and the repair
// remedy explained two. Following it on a duplicate leaves the hook injecting
// nothing, so the advertised minimal repair was a dead end for that third of the
// cases.
//
// The opt-out is the same shape of gap in the docs: `governed = false` is a
// legal, supported one-line file, and the documented "edit strictness in place"
// branch is WRONG for it — the opt-out wins and the hook goes silent, so the
// consumer who asked for the firm posture gets no rules and no message either.
//
// Narrowed by the reconciliation change (TRL-20/TRL-2/TRL-27): the remedy this
// guarded — "for missing:, add those slugs; for unknown:, remove those rows;
// for duplicate:, delete the extra occurrences" — no longer exists, because a
// mismatch is now reconciled rather than refused. Its intent survives in
// TestSlugMismatchStillDeliversEveryRule, which asserts every mismatch kind is
// RESOLVED rather than merely explained. The `governed = false` subtest below
// is unrelated to the remedy and stays.
func TestRepairRemedyCoversEveryMismatchKind(t *testing.T) {
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

func TestDocumentedPostureRecipeActuallyGoverns(t *testing.T) {
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

	t.Run("copy the firm preset, then edit: the full rule set at the firm posture", func(t *testing.T) {
		out := run(t, files["rules-a.toml"])
		if strings.Contains(out, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("the DOCUMENTED recipe must produce a governed project; got:\n%s", out)
		}
		slugs := map[string]bool{}
		for _, m := range regexp.MustCompile(`(?:inv|floor)-[a-z-]+`).FindAllString(out, -1) {
			slugs[m] = true
		}
		if len(slugs) < len(assessableSlugs) {
			t.Errorf("want all %d rules delivered from the copied preset, got %d (%v)", len(assessableSlugs), len(slugs), keysOfBool(slugs))
		}
	})

	t.Run("copy the adaptive preset: same, at the default posture", func(t *testing.T) {
		out := run(t, files["rules-b.toml"])
		if strings.Contains(out, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("the DOCUMENTED recipe must produce a governed project; got:\n%s", out)
		}
	})

	t.Run("copy a preset then disable one row: every other rule still arrives", func(t *testing.T) {
		edited := strings.Replace(files["rules-b.toml"], "inv-minimal-first         = { active = true }", "inv-minimal-first         = { active = false }", 1)
		if edited == files["rules-b.toml"] {
			t.Fatal("fixture did not contain the row it claims to edit — the case would prove nothing")
		}
		out := run(t, edited)
		if strings.Contains(out, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("turning a row off is the documented edit and must stay valid; got:\n%s", out)
		}
	})

	// Positive replacement for the retired "hand-written partial file governs
	// nothing" subtest — not just its negation (no blackout), but the actual P1
	// claim this test exists to pin: the undocumented, but now CORRECT, recipe
	// governs the full rule set at the requested posture from a one-line file.
	// A strictly weaker input (one missing row) is already covered by
	// TestSlugMismatchStillDeliversEveryRule; this is the ALL-SIXTEEN-missing
	// case decision-0072's first draft actually described.
	t.Run("hand-written partial file reconciles to the full rule set at the requested posture", func(t *testing.T) {
		out := run(t, "strictness  = \"firm\"\n")
		if strings.Contains(out, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("a hand-written partial file must reconcile, not black out delivery; got:\n%s", out)
		}
		if !strings.Contains(out, "RECONCILED") {
			t.Errorf("this is a genuine mismatch (sixteen missing rows) and must say it reconciled; got:\n%s", out)
		}
		if !strings.Contains(out, "added 16 row(s)") {
			t.Errorf("all sixteen rows are missing, so the repair summary must say added 16 row(s); got:\n%s", out)
		}
		context := nudgeContext(t, out)
		for _, slug := range assessableSlugs {
			if !deliveredRow(context, slug) {
				t.Errorf("rule %s's row was not actually delivered after reconciling a fully partial file:\n%s", slug, out)
			}
		}
		if !strings.Contains(context, `strictness  = "firm"`) {
			t.Errorf("the requested posture must survive verbatim; got:\n%s", context)
		}
	})
}

// TestRowMismatchRemedyIsNotDestructive originally pinned the retired
// blackout-and-explain remedy's non-destructive INSTRUCTIONS — preserve
// strictness, gate any reseed behind confirmation. Narrowed by the
// reconciliation change (TRL-20/TRL-2/TRL-27, same as
// TestRepairRemedyCoversEveryMismatchKind above): there is no remedy text to
// instruct an agent through any more, because the hook reconciles the
// mismatch itself rather than refusing and explaining. What this test still
// proves, on the same fixture, is the property its name promises: the repair
// is non-destructive by construction, not merely by instruction. Every row
// the consumer wrote — including the unknown one — survives verbatim; the
// unknown row is quarantined (commented out), never deleted, matching this
// task's binding constraint that a repair is additive or commenting and no
// row's value is ever lost.
func TestRowMismatchRemedyIsNotDestructive(t *testing.T) {
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

	proj := t.TempDir()
	if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A firm-posture project carrying a slug the shipped catalog does not know —
	// the same state an upgrade produces from the other direction.
	rows := "seeded_from = \"conductor\"\nstrictness  = \"firm\"\ninv-not-a-real-rule = true\n"
	if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(rows), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(hook)
	cmd.Dir = proj
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
	out, _ := cmd.CombinedOutput()
	ctx := string(out)

	// What the consumer chose must survive verbatim — nothing reseeded, nothing
	// dropped. ctx is the hook's raw JSON stdout, so the fixture's own quote
	// characters come back \"-escaped; match that form, not the unescaped one.
	for _, want := range []string{`seeded_from = \"conductor\"`, `strictness  = \"firm\"`} {
		if !strings.Contains(ctx, want) {
			t.Errorf("the repair must preserve %q verbatim rather than reseed or drop it; got:\n%s", want, ctx)
		}
	}
	// The unknown row is quarantined, not deleted: its original text survives,
	// commented out, with dated provenance a reader can tell it was not simply
	// removed.
	if !strings.Contains(ctx, "# inv-not-a-real-rule = true") {
		t.Errorf("the unknown row must be quarantined by commenting, not deleted; got:\n%s", ctx)
	}
	if !strings.Contains(ctx, "quarantined") {
		t.Errorf("the quarantine must be labelled so a reader can tell why; got:\n%s", ctx)
	}
}

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
			body := renderedFile(files, strings.TrimSpace(files["version"]))
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
		for name, key := range map[string]string{"trellis.md": "trellis-b.md", "rules.md": "rules.md"} {
			if err := os.WriteFile(filepath.Join(internal, name), []byte(files[key]), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		// The fixture must satisfy path C's guard, or this subtest asserts nothing
		// about ordering: path C would never fire and moving it above path A would
		// still pass. It DID guard when written, against the then-looser `-f`
		// check — and my own later hardening of that guard silently made it
		// vacuous. Anchored to the real boundary now.
		rendered := renderedFile(files, strings.TrimSpace(files["version"]))
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
		body := renderedFile(files, "payload@000000000000")
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
		body := renderedFile(files, strings.TrimSpace(files["version"]))
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
		body := "\xef\xbb\xbf" + renderedFile(files, strings.TrimSpace(files["version"]))
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
		for name, key := range map[string]string{"trellis.md": "trellis-b.md", "rules.md": "rules.md"} {
			if err := os.WriteFile(filepath.Join(internal, name), []byte(files[key]), 0o644); err != nil {
				t.Fatal(err)
			}
		}
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
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		body := files["rules.md"] +
			"some prose mentioning invariants.md in passing\n" +
			"and a line where something is authoritative, but not the footer\n" +
			"@../../.trellis/rules.toml\n" +
			"<!-- trellis:rendered-from " + strings.TrimSpace(files["version"]) + " -->\n"
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
		good := renderedFile(files, strings.TrimSpace(files["version"]))
		cur := strings.TrimSpace(files["version"])
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
			strings.TrimSpace(files["version"]) + " -->\n"
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
		cur := strings.TrimSpace(files["version"])
		body := "<!-- trellis:rendered-from payload@deadbeefcafe -->\n" + renderedFile(files, cur)
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
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
			[]byte(renderedFile(files, strings.TrimSpace(files["version"]))), 0o644); err != nil {
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
// deleting it. Absent rows therefore mean the standard set, not none.
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
	// Posture B — the lenient one, and the same default install.sh renders.
	if !strings.Contains(got, "By default") {
		t.Errorf("the default posture must be B (author-adapt, \"By default\"), not the firm one:\n%s", got)
	}
	if strings.Contains(got, "Firmly") {
		t.Errorf("posture A leaked into the no-rules.toml default:\n%s", got)
	}
}

// decision-0077. decision-0070 D4 said an ignored announcement adopts ("accept,
// or no objection → seed … governed at 14/14 from the next turn"), and
// decision-0073 D3 restated it as "one ignored prompt re-governs". The hook has
// never done that: it names two actions — decline, or an explicit accept that
// copies the preset — and governs on neither silence nor its own repetition.
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

// reportSection returns just the "(...)" defect report the hook prints after
// "the rules the installed plugin ships", excluding the remedy prose that
// follows. The remedy names every category by name, so assertions against the
// whole message cannot tell a one-category report from a three-category one.
func reportSection(t *testing.T, out string) string {
	t.Helper()
	m := regexp.MustCompile(`installed plugin ships \(([^)]*)\)`).FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("no defect report found in the hook output:\n%s", out)
	}
	return m[1]
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
	for _, m := range regexp.MustCompile(`(?:inv|floor)-[a-z-]+`).FindAllString(files["rules.md"], -1) {
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
		proj := newProj(t, files["rules-b.toml"])
		writeInstr(t, proj, "CLAUDE.md", files["block-inline-b.md"])
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
		proj := newProj(t, files["rules-b.toml"])
		writeInstr(t, proj, "AGENTS.md", files["block-inline-b.md"])
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
		proj := newProj(t, files["rules-b.toml"])
		writeInstr(t, proj, "CLAUDE.md", files["block-claude.md"])
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
		writeInstr(t, proj, "CLAUDE.md", files["block-inline-b.md"])
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
		proj := newProj(t, files["rules-b.toml"])
		writeInstr(t, proj, "CLAUDE.md", files["block-inline-b.md"])
		if err := os.MkdirAll(filepath.Join(proj, ".claude", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		rendered := renderedFile(files, strings.TrimSpace(files["version"]))
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
		proj := newProj(t, files["rules-b.toml"])
		writeInstr(t, proj, "GEMINI.md", files["block-inline-b.md"])
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
		proj := newProj(t, files["rules-b.toml"])
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
		proj := newProj(t, files["rules-b.toml"])
		writeInstr(t, proj, "AGENTS.md", files["block-codex.md"])
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
		proj := newProj(t, files["rules-b.toml"])
		writeInstr(t, proj, "AGENTS.md", "existing content without trailing newline"+files["block-inline-b.md"])
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
		if err := os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte(files["block-inline-b.md"]), 0o644); err != nil {
			t.Fatal(err)
		}
		// CLAUDE.md exists but does NOT import AGENTS.md — the mixed-host layout.
		if err := os.WriteFile(filepath.Join(proj, "CLAUDE.md"), []byte("# project notes\n\nnothing imported here.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
		if err := os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte(files["block-inline-b.md"]), 0o644); err != nil {
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
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
	// different postures, so "copy the preset that matches" has no answer. The
	// remedy must surface the conflict rather than pick silently.
	t.Run("blocks disagreeing on strictness surface the conflict", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte(files["block-inline-a.md"]), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, "CLAUDE.md"), []byte("@AGENTS.md\n\n"+files["block-inline-b.md"]), 0o644); err != nil {
			t.Fatal(err)
		}
		// No .trellis/rules.toml — the state where a preset must be chosen.
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		ctx := string(out)
		if !strings.Contains(ctx, "TRELLIS_INLINE_BLOCK") {
			t.Fatalf("want the inline refusal, got:\n%s", ctx)
		}
		if !strings.Contains(ctx, "disagree on strictness") {
			t.Errorf("two blocks can carry different postures; the remedy must say so and let the "+
				"user choose rather than picking one silently:\n%s", ctx)
		}
	})

	t.Run("an AGENTS.md block Claude DOES import still refuses", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte(files["block-inline-b.md"]), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, "CLAUDE.md"), []byte("@AGENTS.md\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
		// P2: the remedy must not silently downgrade a firm project.
		if !strings.Contains(ctx, "rules-a.toml") || !strings.Contains(ctx, "strictness") {
			t.Errorf("the missing-rules.toml remedy must preserve the block's own posture "+
				"(name rules-a for firm), not assume adaptive:\n%s", ctx)
		}
	})

	t.Run("blocks in BOTH instruction files are all named, not just the first", func(t *testing.T) {
		proj := t.TempDir()
		block := files["block-inline-b.md"]
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
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
		proj := newProj(t, files["rules-b.toml"])
		writeInstr(t, proj, "CLAUDE.md", "\xef\xbb\xbf"+files["block-inline-b.md"])
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
		proj := newProj(t, files["rules-b.toml"])
		body := strings.ReplaceAll(files["block-inline-b.md"], "\n", "\r\n")
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
		proj := newProj(t, files["rules-b.toml"])
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
// output — the same shape as the `injected` closure at plugin_hook_test.go:1458.
// It scrapes the WHOLE output, including the injected rules.md prose (each rule
// ends with its slug in backticks) and any quarantined/commented row, so its
// presence is not proof the rule was actually DELIVERED as a governed row —
// only that the slug was mentioned somewhere. Fine for proving absence (an
// empty set really does mean the slug appears nowhere); use deliveredRow to
// prove presence of an actual row.
func hookSlugs(out string) map[string]bool {
	slugs := map[string]bool{}
	for _, m := range regexp.MustCompile(`(inv|floor)-[a-z-]+`).FindAllString(out, -1) {
		slugs[m] = true
	}
	return slugs
}

// deliveredRow reports whether context (a DECODED additionalContext — real
// newlines, not the JSON-escaped `\n` a raw hook stdout carries — see
// nudgeContext) contains an actual TOML row for slug: a line shaped
// `slug = { active = ...`, anchored at line start — as opposed to the slug
// merely appearing in the rules.md prose or in a quarantined, commented-out
// row (which starts with `# `, so cannot match here). This is the
// discriminator hookSlugs cannot make: a reconciler that quarantined EVERY
// legitimate row (the no-slugs-in-payload defect) still left every slug name
// somewhere in the prose, so a hookSlugs-only assertion passed silently over a
// session running fully ungoverned.
func deliveredRow(context, slug string) bool {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(slug) + `[ \t]*=[ \t]*\{[ \t]*active`)
	return re.MatchString(context)
}

// A slug mismatch used to inject NOTHING — one bad row cost all sixteen rules,
// every session, until a human edited the file (TRL-20). Delivery now
// reconciles in memory, so a mismatch degrades to a repair notice rather than
// a blackout.
func TestSlugMismatchStillDeliversEveryRule(t *testing.T) {
	run := rulesTomlRun(t)
	files := payloadFiles()

	t.Run("a missing row does not black out the other rules", func(t *testing.T) {
		short := strings.Replace(files["rules-b.toml"],
			"inv-minimal-first         = { active = true }\n", "", 1)
		if short == files["rules-b.toml"] {
			t.Fatal("fixture removed nothing — the case would prove nothing")
		}
		out := run(t, short)
		// "Nothing was injected" was the retired blackout message's own wording;
		// asserting its absence proved nothing once that string left the script.
		// TRELLIS_RULES_NOT_LOADED is the one that would actually fire on a
		// refusal, and RECONCILED is the one the new preamble prints instead.
		if strings.Contains(out, "TRELLIS_RULES_NOT_LOADED") {
			t.Errorf("a missing row must not black out delivery:\n%s", out)
		}
		if !strings.Contains(out, "RECONCILED") {
			t.Errorf("a mismatch must reconcile and say so, not deliver as if nothing were wrong:\n%s", out)
		}
		// hookSlugs would stay true even if every row were quarantined — the slug
		// still appears in the rules.md prose either way. deliveredRow checks the
		// actual TOML row, which a quarantined (`# `-prefixed) line cannot match.
		// It needs real newlines to anchor on, so decode the raw JSON stdout first.
		context := nudgeContext(t, out)
		for _, slug := range assessableSlugs {
			if !deliveredRow(context, slug) {
				t.Errorf("rule %s's row was not actually delivered after reconciliation:\n%s", slug, out)
			}
		}
	})

	t.Run("the missing row is reconciled to active = true", func(t *testing.T) {
		short := strings.Replace(files["rules-b.toml"],
			"inv-minimal-first         = { active = true }\n", "", 1)
		out := run(t, short)
		if !strings.Contains(out, "inv-minimal-first = { active = true }") {
			t.Errorf("the reconciled rows must add the missing slug as active:\n%s", out)
		}
	})

	t.Run("an unknown row is quarantined, never dropped", func(t *testing.T) {
		bogus := files["rules-b.toml"] + "inv-not-a-real-rule       = { active = false }\n"
		out := run(t, bogus)
		if !strings.Contains(out, "# inv-not-a-real-rule") {
			t.Errorf("an unknown row must survive as a commented-out row:\n%s", out)
		}
		if !strings.Contains(out, "quarantined") {
			t.Errorf("the quarantine must be labelled so a reader can tell why:\n%s", out)
		}
		if !strings.Contains(out, "claude plugin update trellis@kodhama") {
			t.Errorf("quarantine provenance must name the stale-plugin cause (TRL-27):\n%s", out)
		}
	})

	t.Run("a duplicate keeps the first occurrence and quarantines the extra", func(t *testing.T) {
		dup := files["rules-b.toml"] + "inv-minimal-first         = { active = false }\n"
		out := run(t, dup)
		if !strings.Contains(out, "# inv-minimal-first         = { active = false }") {
			t.Errorf("the extra occurrence must be quarantined, not deleted:\n%s", out)
		}
		if !strings.Contains(out, "inv-minimal-first         = { active = true }") {
			t.Errorf("the FIRST occurrence must survive verbatim:\n%s", out)
		}
	})

	t.Run("a rename is both kinds at once and both are reconciled", func(t *testing.T) {
		renamed := strings.Replace(files["rules-b.toml"],
			"inv-minimal-first         = { active = true }",
			"inv-renamed-first         = { active = true }", 1)
		if renamed == files["rules-b.toml"] {
			t.Fatal("fixture did not rename anything — the case would prove nothing")
		}
		out := run(t, renamed)
		if !strings.Contains(out, "# inv-renamed-first") {
			t.Errorf("the stale slug must be quarantined:\n%s", out)
		}
		if !strings.Contains(out, "inv-minimal-first = { active = true }") {
			t.Errorf("the new slug must be added:\n%s", out)
		}
	})

	t.Run("a quarantined row is invisible next session — the repair is idempotent", func(t *testing.T) {
		quarantined := files["rules-b.toml"] +
			"# inv-not-a-real-rule = { active = false }  # quarantined 2026-08-30: not in payload@test\n"
		out := run(t, quarantined)
		// "does not match the rules the installed plugin ships" was the retired
		// blackout message's own wording, which this same commit deleted — the
		// assertion passed unconditionally regardless of behaviour. RECONCILED is
		// the string the new preamble actually prints when a repair notice fires;
		// its absence is what "no second notice" means now.
		if strings.Contains(out, "RECONCILED") {
			t.Errorf("an already-repaired file must draw no repair notice at all:\n%s", out)
		}
	})
}

// The repair is applied and REPORTED, not proposed and gated. decision-0072's
// finding #6 — "retiring a confirm-gated writer silently retires the gate" —
// is answered by quarantine semantics: no prior value is ever lost, so the
// gate that guarded destructive writes is not engaged. What must never be
// lost is the loudness.
func TestReconciledRepairIsMandatedAndReported(t *testing.T) {
	run := rulesTomlRun(t)
	files := payloadFiles()
	short := strings.Replace(files["rules-b.toml"],
		"inv-minimal-first         = { active = true }\n", "", 1)
	out := run(t, short)

	t.Run("the agent is told to write the file, not to ask", func(t *testing.T) {
		if !strings.Contains(out, "Write .trellis/rules.toml") {
			t.Errorf("the emit must mandate the write:\n%s", out)
		}
		if strings.Contains(out, "get explicit confirmation before writing") {
			t.Errorf("the confirm-first remedy is retired by this change:\n%s", out)
		}
	})

	t.Run("the agent must report what changed, before other work", func(t *testing.T) {
		if !strings.Contains(out, "Tell the user") {
			t.Errorf("a silent repair is the failure this design exists to prevent:\n%s", out)
		}
		if !strings.Contains(out, "added 1 row(s)") {
			t.Errorf("the emit must state what it reconciled, per row:\n%s", out)
		}
	})

	t.Run("the emit carries the literal file content to write", func(t *testing.T) {
		// The agent re-deriving the repair is how a wrong one lands. The hook
		// computes it once and shows exactly the bytes to save. The added
		// row's provenance (date, stamp, count) now sits in a single header
		// above the block rather than repeated per row (Ruling 6, TRL-20 task
		// 3 — Codex's own context budget left too little headroom for sixteen
		// per-row copies), so the pin checks the header plus the bare row
		// rather than a comment trailing the row itself.
		if !strings.Contains(out, "# added 1 row(s) below on") {
			t.Errorf("the emit must carry the single reconciliation header:\n%s", out)
		}
		if !strings.Contains(out, "inv-minimal-first = { active = true }") {
			t.Errorf("the emit must quote the exact reconciled file:\n%s", out)
		}
	})

	t.Run("no deletion verb enters the emit", func(t *testing.T) {
		// Keeping the repair non-destructive is what keeps it ungated —
		// TestEveryDeletionInstructionIsGated would otherwise demand a
		// confirmation clause and re-impose the gate this change removes.
		for _, verb := range []string{"delete those rows", "remove those rows", "drop the unknown"} {
			if strings.Contains(out, verb) {
				t.Errorf("the reconciled remedy must never instruct a deletion, found %q:\n%s", verb, out)
			}
		}
	})
}

// TestRepairSummaryCountsThisSessionOnly pins the spoken summary to the work
// THIS run did. Quarantine notes and the `# added N row(s)` header are
// persisted provenance — they stay in the file after a repair is applied, by
// design — so counting them out of the reconciled text counts every earlier
// session's repairs alongside this one. Measured on the pre-fix hook with the
// fixture below: "added 2 row(s); quarantined 1 row(s)" for a session that
// added exactly 1 and quarantined 0.
//
// It is only the summary that inflated; the in-file provenance was right
// either way. That is precisely why it matters — the summary is the string the
// agent reads back to the user, and this whole design rests on that report
// being trustworthy. An inflated count tells the user rows were touched that
// were not, in the one channel built to prevent silent repairs.
func TestRepairSummaryCountsThisSessionOnly(t *testing.T) {
	run := rulesTomlRun(t)
	files := payloadFiles()

	// A partially repaired file: an earlier session quarantined one row and
	// added one, and both marks are still on disk (that is what "reversible
	// from the file itself" means). One further row is missing now.
	partiallyRepaired := strings.Replace(files["rules-b.toml"],
		"inv-minimal-first         = { active = true }\n", "", 1)
	partiallyRepaired = strings.Replace(partiallyRepaired,
		"[rules]  # one row per assessable catalog slug (signature-catalog-v1)\n",
		"[rules]  # one row per assessable catalog slug (signature-catalog-v1)\n"+
			"# inv-since-retired = { active = true }  # quarantined 2026-08-01: not in payload@aaaaaaaaaaaa. If a newer Trellis ships this slug, run `claude plugin update trellis@kodhama` and uncomment.\n"+
			"# added 1 row(s) below on 2026-08-01 (missing from payload@aaaaaaaaaaaa)\n", 1)

	out := run(t, partiallyRepaired)
	if !strings.Contains(out, "RECONCILED") {
		t.Fatalf("premise: this fixture must reconcile, or the case proves nothing:\n%s", out)
	}
	// Premise: the earlier session's marks must survive into the reconciled
	// text. If they did not, this fixture could not distinguish a per-session
	// count from a cumulative one and would pass for the wrong reason.
	if !strings.Contains(out, "quarantined 2026-08-01") || !strings.Contains(out, "# added 1 row(s) below on 2026-08-01") {
		t.Fatalf("premise: the prior session's provenance must be carried through unchanged (quarantine never deletes):\n%s", out)
	}

	if !strings.Contains(out, "added 1 row(s); quarantined 0 row(s)") {
		t.Errorf("the summary must count only this session: this run added 1 row and quarantined 0, whatever earlier repairs the file already records:\n%s", out)
	}
	if strings.Contains(out, "added 2 row(s)") || strings.Contains(out, "quarantined 1 row(s)") {
		t.Errorf("the summary is counting earlier sessions' marks as this session's work:\n%s", out)
	}
	// The counts reach the shell on a trailer line the reconciler prints; it
	// must be stripped before delivery, or it lands in the rows the agent is
	// told to write verbatim — a top-level comment is harmless TOML but it is
	// still hook bookkeeping leaking into a consumer-owned file.
	if strings.Contains(out, "trellis-reconcile-counts") {
		t.Errorf("the reconciler's internal counts trailer must never reach the delivered text:\n%s", out)
	}
}

// TestNoSlugsInPayloadFailsLoudly pins the invariant staleness.sh states 40
// lines above the reconciler ("Fail loudly rather than govern silently on a
// partial payload") against the one verdict the reconciler must NEVER touch.
// `no-slugs-in-payload` (staleness.sh:596) means the validator found NOTHING
// to check rows against — the payload's own rules.md is unreadable or
// malformed — which is a different failure than a project's rows not
// matching a valid payload. Before the fix, the reconciler's guard was
// `[ "$slug_report" != "ok" ]`, so this verdict entered it with an EMPTY want
// set: every legitimate row got quarantined and the session ran silently
// ungoverned with exit 0. Reproduced during review: RECONCILED …
// (added 0 row(s); quarantined 16 row(s)), all sixteen rows commented out.
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
	// An ordinary, fully valid row set — the defect is not in these rows.
	if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
	if strings.Contains(ctx, "RECONCILED") {
		t.Errorf("no-slugs-in-payload must never enter the reconciler — there is nothing to reconcile against; got:\n%s", ctx)
	}
	if got := hookSlugs(ctx); len(got) > 0 {
		t.Errorf("no rows may be injected when the payload itself cannot be validated against; got %v in:\n%s", keysOfBool(got), ctx)
	}
}

// TestUnreadablePayloadFileNeverReconcilesToBlackout pins the SECOND door into
// the same hazard the note at staleness.sh:624 names, and the fourth time this
// branch produced it: an empty want set reaching the reconciler, every
// legitimate row quarantined, and the session running ungoverned at exit 0.
//
// The validator awk reads two files POSITIONALLY — the payload's rules.md and
// the project's rules.toml — and a positional file that exists but cannot be
// opened is a fatal awk error: it prints nothing, so `$slug_report` is the
// EMPTY STRING, not `no-slugs-in-payload`. The old guard tested only
// `!= ok && != no-slugs-in-payload`, so "" read as a mismatch to repair. Inside
// the reconciler the want set is filled by a REDIRECTED getline, which returns
// -1 silently on a failed open rather than dying, so want[] stayed empty, every
// row failed `row in want`, and the hook shipped `added 0 row(s); quarantined
// 16 row(s)` together with a mandate to write that all-commented-out file to
// disk. The unreadable-rules.toml variant is the same defect one step further
// on: the reconcile awk died too, `$reconciled` stayed empty, and delivery fell
// back to a `cat` that failed just as quietly — the rows heading with nothing
// under it and no warning at all.
//
// Both are broken-payload/broken-config shapes, so both must exit through the
// loud door, deliver no governed row, and never claim a reconciliation.
func TestUnreadablePayloadFileNeverReconcilesToBlackout(t *testing.T) {
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
			// A fully valid row set: the defect is never in these rows, which is
			// the whole point — quarantining them is the failure.
			if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
				t.Fatalf("an unreadable payload file must fail loudly, not reconcile against an empty want set:\n%s", raw)
			}
			ctx := nudgeContext(t, strings.TrimSpace(raw))
			if strings.Contains(ctx, "RECONCILED") || strings.Contains(ctx, "quarantined") {
				t.Errorf("nothing may be reconciled or quarantined when the want set could not be read at all:\n%s", ctx)
			}
			// hookSlugs alone cannot tell a governed row from a slug named in
			// prose or in a commented-out quarantine line; deliveredRow can.
			for _, slug := range []string{"floor-transparency", "floor-intent-gate", "inv-minimal-first"} {
				if deliveredRow(ctx, slug) {
					t.Errorf("no row may be delivered when the payload could not be validated; got a row for %s in:\n%s", slug, ctx)
				}
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
// the `@rules.md` import, and the firm-posture trellis-a.md — the hook emitted
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
	// non-empty and still delivers rows with no rules under them.
	const importLine = "@rules.md\n"
	head, _, found := strings.Cut(files["trellis-b.md"], importLine)
	if !found || head == "" {
		t.Fatal("premise: trellis-b.md must carry an @rules.md import with prose above it")
	}

	for _, tc := range []struct {
		name string
		// rows selects the posture, which selects WHICH header file is read.
		rows string
		// header names the payload file the fixture breaks, and how.
		header  string
		mode    os.FileMode
		content string // "" with mode 0 means: leave the bytes, deny the read
	}{
		{
			name:   "the adaptive posture header exists but cannot be read",
			rows:   files["rules-b.toml"],
			header: "trellis-b.md",
			mode:   0o000,
		},
		{
			name:    "the posture header is zero bytes",
			rows:    files["rules-b.toml"],
			header:  "trellis-b.md",
			content: "",
			mode:    0o644,
		},
		{
			name:    "the posture header is truncated above its @rules.md import",
			rows:    files["rules-b.toml"],
			header:  "trellis-b.md",
			content: head,
			mode:    0o644,
		},
		{
			// Posture selects the header, so the firm side needs its own row:
			// a fix that guarded only trellis-b.md would leave firm projects
			// blacked out and every adaptive test green.
			name:   "the firm posture header exists but cannot be read",
			rows:   files["rules-a.toml"],
			header: "trellis-a.md",
			mode:   0o000,
		},
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
			target := filepath.Join(pluginRoot, "reference", tc.header)
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
			// A fully valid row set every time: rows are never the defect here,
			// which is the point — delivering them alone is the failure.
			if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(tc.rows), 0o644); err != nil {
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
				t.Fatalf("an unusable posture header must fail loudly, not ship rows with no rules under them:\n%s", raw)
			}
			ctx := nudgeContext(t, strings.TrimSpace(raw))
			// The two halves of the blackout, asserted separately: no activation
			// list, and no row inside one. Either alone would pass over the
			// pre-fix output, which carried a perfectly ordinary-looking
			// heading above sixteen perfectly ordinary-looking rows.
			if strings.Contains(ctx, "## Project rule activation") {
				t.Errorf("no activation list may be delivered when the rules prose could not be assembled:\n%s", ctx)
			}
			for _, slug := range []string{"floor-transparency", "floor-intent-gate", "inv-minimal-first"} {
				if deliveredRow(ctx, slug) {
					t.Errorf("row for %s delivered with no rules prose to govern by:\n%s", slug, ctx)
				}
			}
			if strings.Contains(ctx, rulesLoadedSentinel) {
				t.Errorf("nothing may claim the rules were loaded on this path:\n%s", ctx)
			}
		})
	}
}

// truncateRulesMdAfter returns the payload's rules.md cut immediately after its
// nth backticked slug tag — a file that is NON-EMPTY and carries real slugs, so
// the validator's only broken-payload test, `length(want) == 0`, passes it.
//
// keepTerminator re-appends the `<!-- trellis:rules-loaded -->` line. The
// coherence test needs it TRUE: the terminator gate runs earlier and would
// otherwise catch this fixture first, leaving the coherence guard pinned by
// nothing. The terminator test needs it FALSE — that is the shape it is about.
// A real truncation loses the terminator, so keepTerminator: true models the
// narrower case of a payload whose rule list was corrupted with its ending
// intact, which is exactly the residue the terminator gate cannot see.
func truncateRulesMdAfter(t *testing.T, n int, keepTerminator bool) string {
	t.Helper()
	tag := regexp.MustCompile("`(inv|floor)-[a-z-]+`[ \t]*$")
	var kept []string
	seen := 0
	for _, line := range strings.Split(payloadFiles()["rules.md"], "\n") {
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
	out := strings.Join(kept, "\n") + "\n"
	if keepTerminator {
		out += rulesLoadedSentinel + "\n"
	}
	return out
}

// TestIncoherentPayloadNeverMandatesQuarantiningTheProjectsRows closes the one
// member of this family that is worse in KIND than the rest.
//
// Every guard above withholds governance for a session when the payload is
// broken. This one PERSISTS DAMAGE. The validator's only test for a broken
// rules.md is `length(want) == 0`, so a rules.md truncated BELOW its first slug
// is non-empty, passes, and is then treated as authoritative. Measured with a
// nine-line, two-slug payload: the hook reported `quarantined 14 row(s)`,
// commented out BOTH floor rules, and instructed the agent to write that file
// to .trellis/rules.toml — at exit 0, with no loud marker. The whole safety
// argument for repairing without a gate is that quarantine loses nothing; a
// truncated payload turns it into fourteen deletions-in-effect in a file the
// consumer owns.
//
// It is introduced by this branch, not inherited: `main` has no reconciler and
// no quarantine at all.
//
// The check is payload-vs-PAYLOAD. The payload states its rule set twice —
// rules.md tags sixteen slugs, reference/rules-b.toml carries sixteen rows —
// and the two are identical by construction, so a disagreement between them is
// provable internal corruption. That is exactly what distinguishes it from the
// stale-plugin case quarantine legitimately exists for, where the payload is
// coherent and the PROJECT is out of step. The second subtest is the control
// for that distinction: it must still reconcile, unchanged.
func TestIncoherentPayloadNeverMandatesQuarantiningTheProjectsRows(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()

	run := func(t *testing.T, rulesMd, rows string) string {
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
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(rows), 0o644); err != nil {
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

	t.Run("a truncated rules.md must not drive a quarantine mandate", func(t *testing.T) {
		truncated := truncateRulesMdAfter(t, 2, true)
		// The premise the validator could not see: non-empty, real slugs, and
		// still not the payload's rule set.
		if strings.TrimSpace(truncated) == "" {
			t.Fatal("premise: the fixture must be non-empty, or no-slugs-in-payload would catch it for the wrong reason")
		}
		ctx := run(t, truncated, files["rules-b.toml"])

		if !strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("a payload that disagrees with itself must be refused, not treated as the authority on the consumer's rows:\n%s", ctx)
		}
		// The damage, asserted as the damage: a quarantine count, a commented
		// floor row, and the mandate to persist it.
		if strings.Contains(ctx, "quarantined") {
			t.Errorf("nothing may be quarantined against a payload that cannot say what the rule set is:\n%s", ctx)
		}
		for _, floor := range []string{"floor-transparency", "floor-intent-gate"} {
			if strings.Contains(ctx, "# "+floor) {
				t.Errorf("%s was commented out on the strength of a truncated payload:\n%s", floor, ctx)
			}
		}
		if strings.Contains(ctx, "Write .trellis/rules.toml with exactly the rows shown above") {
			t.Errorf("the repair mandate must never fire against an incoherent payload — it is the half that persists the damage:\n%s", ctx)
		}
		if strings.Contains(ctx, "RECONCILED") {
			t.Errorf("nothing may be reconciled against a payload that disagrees with itself:\n%s", ctx)
		}
		for _, slug := range []string{"floor-transparency", "inv-minimal-first"} {
			if deliveredRow(ctx, slug) {
				t.Errorf("no row may be delivered from an incoherent payload; got %s in:\n%s", slug, ctx)
			}
		}
	})

	// The control that keeps the guard honest. It is payload-vs-payload
	// disagreement being caught, never payload-vs-project: a project out of
	// step with a COHERENT payload is the case reconciliation exists for, and
	// it must behave exactly as it did before this guard.
	t.Run("a coherent payload still reconciles a project that is out of step", func(t *testing.T) {
		short := strings.Replace(files["rules-b.toml"],
			"inv-minimal-first         = { active = true }\n", "", 1)
		if short == files["rules-b.toml"] {
			t.Fatal("fixture removed nothing — the control would prove nothing")
		}
		ctx := run(t, files["rules.md"], short)

		if strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("a coherent payload must still reconcile a project whose rows are out of step:\n%s", ctx)
		}
		if !strings.Contains(ctx, "RECONCILED") {
			t.Errorf("the ordinary mismatch must still reconcile and say so:\n%s", ctx)
		}
		for _, slug := range []string{"inv-minimal-first", "floor-transparency", "floor-intent-gate"} {
			if !deliveredRow(ctx, slug) {
				t.Errorf("reconciliation must still deliver %s as a governed row:\n%s", slug, ctx)
			}
		}
	})
}

// reconciledRowsFromContext extracts the reconciled `.trellis/rules.toml` text
// from a decoded additionalContext (see nudgeContext) — the block between the
// "Project rule activation" preamble's fixed trailing sentence and whichever
// comes first: the repair mandate's "## Rule activation was reconciled this
// session" heading (Ruling B / Task 2 — present whenever $reconciled is, which
// is the only case this helper is called for), or the fixed "Delivered by the
// Trellis plugin" footer, for a payload with no reconciliation at all. This is
// exactly what a preamble carrying RECONCILED describes as "the file on disk
// still differs": what an agent applying the repair would write to
// .trellis/rules.toml — the row block only, not the mandate prose that follows
// it and is never valid TOML.
func reconciledRowsFromContext(t *testing.T, context string) string {
	t.Helper()
	m := regexp.MustCompile(`(?s)apply regardless of their row\.\n\n(.*?)\n\n(?:## Rule activation was reconciled this session|Delivered by the Trellis plugin)`).
		FindStringSubmatch(context)
	if m == nil {
		t.Fatalf("could not find the row block in the hook's decoded context:\n%s", context)
	}
	return m[1]
}

// TestReconciledRowsParseForCodexToo is the end-to-end regression for the
// [rules]-table defect the reviewer found in the reconciler: staleness.sh
// appended missing rows at EOF with no awareness of whether a `[rules]` table
// already opened one. For the hand-written-partial shape (just
// `strictness = "firm"`, no rows at all) that put sixteen rows at TOP LEVEL —
// parseRulesToml in codex-context.mjs accepts inv-/floor- keys only INSIDE
// `[rules]`, rejecting any other top-level key as invalid-rules. Once an agent
// writes the reconciled text to disk (staleness.sh itself never writes —
// decision-0070 D4), the identical file would read invalid-rules (0 rules)
// under Codex while Claude's own hook — a naive line scanner that does not
// care about TOML table scoping — kept governing normally from it: the exact
// host divergence those files' own comments exist to prevent.
//
// This closes the loop for real, not by re-deriving the fix in Go: run
// Claude's hook to get the actual reconciled text (the reviewer's exact
// reproduction fixture — just `strictness = "firm"`, all sixteen rows
// missing, no [rules] table at all — the only shape a single inserted
// [rules] header can fully cover: any row already present before the
// insertion point would remain outside the table it opens, so this is not
// an arbitrary choice of fixture, it is the one this specific fix actually
// solves), write THAT text to .trellis/rules.toml (what applying the repair
// means), then run Codex's own hook against the identical file and require
// it to parse.
//
// Codex's hook is run in VENDORED mode with minimal placeholder overlay
// prose/rules/version files, not the real payload — deliberately, and unlike
// every other codex_hook_test.go case. Reconciling all sixteen rows with this
// reconciler's per-row provenance comments used to assemble to roughly 1.4KB
// on its own; added to the REAL rules.md + trellis-a.md (~6.7KB), the total
// cleared Claude's 32768-byte budget comfortably but exceeded Codex's much
// tighter 8000-byte one — confirmed separately: the plain firm preset alone
// (no reconciliation comments) already assembled to 7876 bytes under Codex,
// leaving under 200 bytes of headroom. That was a genuine, separate finding,
// resolved by TRL-20 task 3 (Ruling 6): the reconciler now states an added
// row's provenance once, in a header above the block, instead of once per
// row — see TestReconciledCodexPayloadFitsContextBudget for the dedicated,
// real-payload pin. It is still not what THIS test exists to prove, and
// letting the real payload's size gate it would make it fail (or pass) for
// the wrong reason. Minimal placeholders isolate the one property under
// test: does the SAME .trellis/rules.toml Claude governs from parse under
// Codex's real, unmodified parseRulesToml.
//
// The placeholder rules.md is no longer free to be arbitrarily short,
// though: TRL-20 task 3 made Codex derive its slug set FROM this file's
// content (parity with Claude's reconciler), so it must still carry a
// slug-tag line for every slug the reconciled rules.toml below carries, or
// parseRulesToml would reject every row as unknown for a reason unrelated
// to the property this test checks. It stays minimal in every other way —
// bare backticked slug lines, no rule prose.
// The second case covers the same class through the OPPOSITE mistake, found by
// the final whole-branch review. staleness.sh detected the existing table with
// an anchored `^\[rules\]`, while parseRulesToml TRIMS each line before matching
// its section regex — so Codex read an indented `  [rules]` as opening the table
// and Claude did not. A file with an indented table plus any missing row got a
// SECOND `[rules]` appended, and a second table header is exactly what
// parseRulesToml rejects: the repaired file read invalid-rules on Codex.
// Nothing was lost — such a file was already Codex-invalid before the repair,
// so this is not a regression — but the mandate the hook prints promises the
// written file "matches what governs", and a file Codex refuses does not.
// Both cases are the one property: the table structure the reconciler produces
// must be legal for the parser on the other host, whatever the input shape.
func TestReconciledRowsParseForCodexToo(t *testing.T) {
	run := rulesTomlRun(t)

	for _, tc := range []struct {
		name  string
		rows  string
		added string
	}{
		{
			name:  "no [rules] table at all: sixteen rows must not land at top level",
			rows:  "strictness  = \"firm\"\n",
			added: "added 16 row(s)",
		},
		{
			name:  "an INDENTED [rules] table is the table, not a reason to append a second",
			rows:  "strictness = \"firm\"\n  [rules]\n  inv-directional-flow = { active = false }\n",
			added: "added 15 row(s)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := run(t, tc.rows)
			context := nudgeContext(t, out)
			if !strings.Contains(out, "RECONCILED") {
				t.Fatalf("premise: this fixture must reconcile, or the case proves nothing:\n%s", out)
			}
			if !strings.Contains(out, tc.added) {
				t.Fatalf("premise: this fixture must report %q, or it is not the shape the fix covers:\n%s", tc.added, out)
			}
			reconciled := reconciledRowsFromContext(t, context)
			// Premise, not the assertion under test: [rules] must appear BEFORE
			// the first row (it need not be the first LINE — strictness precedes
			// it), or writing this to disk would trivially fail for every parser,
			// proving nothing about the specific defect below.
			rulesIdx := strings.Index(reconciled, "[rules]")
			firstRow := regexp.MustCompile(`(?m)^[ \t]*(inv|floor)-[a-z-]+[ \t]*=`).FindStringIndex(reconciled)
			if rulesIdx < 0 || firstRow == nil || rulesIdx > firstRow[0] {
				t.Fatalf("premise: the reconciled text must open a [rules] table before its first row, or the case would prove nothing; got:\n%s", reconciled)
			}
			// The direct pin, stated in Go as well as through Codex below: one
			// table header, indented or not. Reading it off the text names the
			// defect when it returns; the Codex run proves it actually matters.
			if n := len(regexp.MustCompile(`(?m)^[ \t]*\[rules\]`).FindAllString(reconciled, -1)); n != 1 {
				t.Fatalf("the reconciled text must carry exactly one [rules] table header, got %d — a second one is a fatal duplicate section for Codex:\n%s", n, reconciled)
			}

			project := newGitProject(t)
			internal := filepath.Join(project, ".trellis", "internal")
			if err := os.MkdirAll(internal, 0o755); err != nil {
				t.Fatal(err)
			}
			// Minimal, valid-shape placeholders — short on purpose (see doc
			// comment): the property under test is .trellis/rules.toml, not
			// rules.md/trellis.md content, and the real payload's size is
			// exactly what must NOT gate this test. rules.md still needs one
			// bare slug-tag line per assessable slug (see doc comment) so
			// Codex's derived slug set matches the reconciled file's sixteen
			// rows.
			if err := os.WriteFile(filepath.Join(internal, "trellis.md"), []byte("@rules.md\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			var minimalRulesMd strings.Builder
			for _, slug := range assessableSlugs {
				minimalRulesMd.WriteString("`" + slug + "`\n")
			}
			minimalRulesMd.WriteString(rulesLoadedSentinel + "\n")
			if err := os.WriteFile(filepath.Join(internal, "rules.md"), []byte(minimalRulesMd.String()), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(internal, "version"), []byte("payload@000000000000\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			// What an agent applying the repair actually writes to disk.
			if err := os.WriteFile(filepath.Join(project, ".trellis", "rules.toml"), []byte(reconciled), 0o644); err != nil {
				t.Fatal(err)
			}

			raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
			if got.HookSpecificOutput == nil || got.SystemMessage != "" {
				t.Fatalf("the reconciled rows must parse for Codex too — a file Claude governs normally from must not read invalid-rules (0 rules) under Codex: %s", raw)
			}
		})
	}
}

// TestTruncatedRulesMdIsRefusedByItsOwnTerminator adds the check Codex has had
// since it shipped and the Claude hook did not: rules.md must carry exactly one
// `<!-- trellis:rules-loaded -->` terminator, as its final line.
//
// It closes the same truncated-payload hole the coherence guard does, by a
// shorter and less conditional route. rules.md is 39 lines with the terminator
// on line 39, so ANY truncation loses it — no slug arithmetic required. And
// unlike the coherence guard it has no second-file dependency: that guard needs
// reference/rules-b.toml and SKIPS itself when the file is absent, which
// reopens the full hole (measured: `quarantined 14 row(s)`, both floor rules
// commented out, exit 0, no marker). The third subtest is that exact
// combination, and it is the reason this gate is worth its line count.
//
// Ordered BEFORE the slug derivation, matching codex-context.mjs and its
// comment's reasoning: a payload validated after it is trusted produces
// verdicts about the consumer's rows for a defect that is the plugin's.
func TestTruncatedRulesMdIsRefusedByItsOwnTerminator(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()
	if !strings.HasSuffix(files["rules.md"], rulesLoadedSentinel+"\n") {
		t.Fatal("premise: the shipped rules.md must end with the terminator this gate requires")
	}

	run := func(t *testing.T, rulesMd string, dropPresetFile bool) string {
		t.Helper()
		pluginRoot := t.TempDir()
		if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
			t.Fatal(err)
		}
		for name, body := range files {
			if name == "rules.md" {
				body = rulesMd
			}
			if dropPresetFile && name == "rules-b.toml" {
				continue
			}
			if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(files["rules-b.toml"]), 0o644); err != nil {
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
		if strings.Contains(ctx, "quarantined") {
			t.Errorf("nothing may be quarantined against a truncated payload:\n%s", ctx)
		}
		for _, floor := range []string{"floor-transparency", "floor-intent-gate"} {
			if strings.Contains(ctx, "# "+floor) {
				t.Errorf("%s was commented out on the strength of a truncated payload:\n%s", floor, ctx)
			}
		}
		if strings.Contains(ctx, "Write .trellis/rules.toml with exactly the rows shown above") {
			t.Errorf("the repair mandate must never fire against a truncated payload:\n%s", ctx)
		}
		for _, slug := range []string{"floor-transparency", "inv-minimal-first"} {
			if deliveredRow(ctx, slug) {
				t.Errorf("no row may be delivered from a truncated payload; got %s in:\n%s", slug, ctx)
			}
		}
	}

	t.Run("a rules.md cut short of its terminator is refused", func(t *testing.T) {
		assertRefused(t, run(t, truncateRulesMdAfter(t, 2, false), false))
	})

	t.Run("a doubled terminator is refused too", func(t *testing.T) {
		assertRefused(t, run(t, files["rules.md"]+rulesLoadedSentinel+"\n", false))
	})

	// The reason this gate earns its place beside the coherence guard: with
	// rules-b.toml absent, coherence skips itself and the hole is fully open.
	// Measured on the pre-gate hook, exactly this fixture produced
	// `quarantined 14 row(s)` with both floor rules commented out.
	t.Run("a truncated rules.md is refused even with no rules-b.toml to compare against", func(t *testing.T) {
		assertRefused(t, run(t, truncateRulesMdAfter(t, 2, false), true))
	})

	// The control: the shipped payload must sail through the new gate.
	t.Run("the shipped payload passes its own terminator check", func(t *testing.T) {
		ctx := run(t, files["rules.md"], false)
		if strings.Contains(ctx, "TRELLIS_RULES_NOT_LOADED") {
			t.Fatalf("the shipped payload must not be refused by the gate meant for broken ones:\n%s", ctx)
		}
		for _, slug := range []string{"floor-transparency", "inv-minimal-first"} {
			if !deliveredRow(ctx, slug) {
				t.Errorf("the shipped payload must still deliver %s:\n%s", slug, ctx)
			}
		}
	})
}

// TestPluginRootWithABackslashStillDeliversTheRules is the eighth instance of
// the silent-read class, and the twin of the reconciler getline this branch
// already fixed: `@rules.md` expanded through `while ((getline line < rules) >
// 0)` with the return value discarded, on the `slug_report == "ok"` path, 186
// lines from the sibling that checks it.
//
// The trigger is not a permission but the `-v` channel. `awk -v`
// ESCAPE-PROCESSES its value — `awk -v v='/a\tb/c'` yields length 6, not 7 — so
// a CLAUDE_PLUGIN_ROOT containing a backslash reached that awk as a DIFFERENT
// path than the one every `-f` test and every positional read used, all of
// which passed. Measured with a root named `plug\tools`: 16 activation rows,
// 0 rules prose, exit 0, no marker — verbatim the damage shape the posture
// header guard exists to stop. The discriminator that made it unambiguous: the
// SAME root with a mismatched rules.toml refused loudly, from the sibling that
// checks rc.
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
		name       string
		dir        string
		rows       string
		reconciles bool
	}{
		{
			// The path the sibling rc check never covered.
			name: "a backslash in the plugin root, rows already matching",
			dir:  "plug\\tools",
			rows: files["rules-b.toml"],
		},
		{
			// An ampersand corrupted the invariants pointer through the same
			// awk by the opposite mechanism; both are asserted here so a fix
			// for either cannot silently break the other.
			name: "an ampersand in the plugin root, rows already matching",
			dir:  "R&D",
			rows: files["rules-b.toml"],
		},
		{
			// The discriminator: this path already refused loudly, because the
			// reconciler checks rc. It must now DELIVER, since the root is
			// legitimate and every file on it is readable.
			name:       "a backslash in the plugin root, rows needing reconciliation",
			dir:        "plug\\tools",
			rows:       strings.Replace(files["rules-b.toml"], "inv-minimal-first         = { active = true }\n", "", 1),
			reconciles: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.reconciles && tc.rows == files["rules-b.toml"] {
				t.Fatal("fixture removed nothing — the reconcile case would prove nothing")
			}
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
			if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(tc.rows), 0o644); err != nil {
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
			// The blackout, asserted as itself: rows without rules is the
			// shape, so both halves are required.
			if !strings.Contains(ctx, rulesLoadedSentinel) {
				t.Errorf("the rules prose was not imported — this is the rows-without-rules blackout:\n%s", ctx)
			}
			for _, slug := range []string{"floor-transparency", "floor-intent-gate", "inv-minimal-first"} {
				if !deliveredRow(ctx, slug) {
					t.Errorf("row for %s was not delivered:\n%s", slug, ctx)
				}
			}
			// Verbatim, not merely present: the escape hazard mangled this
			// pointer without changing its shape.
			wantPointer := filepath.Join(pluginRoot, "reference", "invariants.md")
			if !strings.Contains(ctx, "`"+wantPointer+"`") {
				t.Errorf("the invariants pointer must name the real path verbatim, want %q in:\n%s", wantPointer, ctx)
			}
			if tc.reconciles && !strings.Contains(ctx, "RECONCILED") {
				t.Errorf("a project out of step must still reconcile on such a root:\n%s", ctx)
			}
		})
	}
}

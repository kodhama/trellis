package main

import (
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

func TestEveryDestructiveInstructionIsGated(t *testing.T) {
	body, err := os.ReadFile("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	emits := regexp.MustCompile(`(?m)^\s*emit "((?:[^"\\]|\\.)*)"`).FindAllStringSubmatch(string(body), -1)
	if len(emits) < 8 {
		t.Fatalf("found only %d emit strings — the scan is broken, and a guard that reads nothing passes", len(emits))
	}
	gated := 0
	for _, m := range emits {
		msg := m[1]
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
			t.Errorf("this message instructs a mutation (%q) with no confirmation gate — an autonomous "+
				"agent can act on it against files the consumer owns (floor-intent-gate):\n%s", hit, msg)
		}
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
	// there is nothing destructive left to gate on that path.
	const known = 12
	if gated < known {
		t.Fatalf("matched %d destructive messages, expected at least %d — the filter broke; "+
			"a guard that matches nothing passes silently", gated, known)
	}
	t.Logf("checked %d destructive messages of %d emits", gated, len(emits))
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
// added without a gate, so the guard reads every emit string in the script.
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
	gated := 0
	for _, m := range emits {
		msg := m[1]
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
	t.Logf("checked %d deletion-instructing messages of %d emits", gated, len(emits))
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
		out, _ := cmd.CombinedOutput()
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
	// scans rules.md for. Everything else about the payload stays valid.
	files["rules.md"] = "# Rules\n\nThis payload carries no rule slugs at all.\n"

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

// reconciledRowsFromContext extracts the reconciled `.trellis/rules.toml` text
// from a decoded additionalContext (see nudgeContext) — the block between the
// "Project rule activation" preamble's fixed trailing sentence and the fixed
// "Delivered by the Trellis plugin" footer. This is exactly what a preamble
// carrying RECONCILED describes as "the file on disk still differs": what an
// agent applying the repair would write to .trellis/rules.toml.
func reconciledRowsFromContext(t *testing.T, context string) string {
	t.Helper()
	m := regexp.MustCompile(`(?s)apply regardless of their row\.\n\n(.*)\n\nDelivered by the Trellis plugin`).
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
// reconciler's verbose per-row provenance comments assembles to roughly 1.4KB
// on its own; added to the REAL rules.md + trellis-a.md (~6.7KB), the total
// clears Claude's 32768-byte budget comfortably but exceeds Codex's much
// tighter 8000-byte one — confirmed separately: the plain firm preset alone
// (no reconciliation comments) already assembles to 7876 bytes under Codex,
// leaving under 200 bytes of headroom. That is a genuine, separate finding —
// reported alongside this fix, not fixed by it, and its failure mode is
// Codex's own loud, already-defined "context-over-budget" path, not a silent
// one — but it is not what THIS test exists to prove, and letting the real
// payload's size gate it would make it fail (or pass) for the wrong reason.
// Minimal placeholders isolate the one property under test: does the
// SAME .trellis/rules.toml Claude governs from parse under Codex's real,
// unmodified parseRulesToml.
func TestReconciledRowsParseForCodexToo(t *testing.T) {
	run := rulesTomlRun(t)
	out := run(t, "strictness  = \"firm\"\n")
	context := nudgeContext(t, out)
	if !strings.Contains(out, "RECONCILED") {
		t.Fatalf("premise: this fixture must reconcile, or the case proves nothing:\n%s", out)
	}
	if !strings.Contains(out, "added 16 row(s)") {
		t.Fatalf("premise: all sixteen rows must be missing, or this is not the shape the fix actually covers:\n%s", out)
	}
	reconciled := reconciledRowsFromContext(t, context)
	// Premise, not the assertion under test: [rules] must appear BEFORE the
	// first row (it need not be the first LINE — strictness precedes it), or
	// writing this to disk would trivially fail for every parser, proving
	// nothing about the specific defect below.
	rulesIdx := strings.Index(reconciled, "[rules]")
	firstRow := regexp.MustCompile(`(?m)^(inv|floor)-[a-z-]+[ \t]*=`).FindStringIndex(reconciled)
	if rulesIdx < 0 || firstRow == nil || rulesIdx > firstRow[0] {
		t.Fatalf("premise: the reconciled text must open a [rules] table before its first row, or the case would prove nothing; got:\n%s", reconciled)
	}

	project := newGitProject(t)
	internal := filepath.Join(project, ".trellis", "internal")
	if err := os.MkdirAll(internal, 0o755); err != nil {
		t.Fatal(err)
	}
	// Minimal, valid-shape placeholders — short on purpose (see doc comment):
	// the property under test is .trellis/rules.toml, not rules.md/trellis.md
	// content, and the real payload's size is exactly what must NOT gate this
	// test.
	if err := os.WriteFile(filepath.Join(internal, "trellis.md"), []byte("@rules.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(internal, "rules.md"), []byte(rulesLoadedSentinel+"\n"), 0o644); err != nil {
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
}

package main

// TRL-52. The invariants pointer the posture prose ships — the token
// `.trellis/internal/invariants.md` — is a PLACEHOLDER, not an address. Which
// address it resolves to is decided per delivery channel, and until this file
// existed one channel decided nothing.
//
// decision-0065:106-111 states the rule, and it is scoped to a SHAPE rather
// than to a host: "For a project holding `.trellis/rules.toml` and no
// `.trellis/internal/`, the hook injects the always-loaded chain from the
// plugin payload ... the one edit repoints the invariants pointer at the
// plugin's own copy, which is where the file is in this mode and which
// therefore cannot go stale." Both hooks deliver that chain on that shape;
// only staleness.sh performed the edit. codex-context.mjs shipped the token
// verbatim, so a Codex session was told, in the last line of its context, to
// read a path that mode cannot have.
//
// The converse shape is why the fix is not "repoint unconditionally". A
// VENDORED project's `.trellis/internal/invariants.md` is a real file: the
// retired setup skill copied it there (its SKILL.md at 2f6a21a:121,
// `cp "${CLAUDE_PLUGIN_ROOT}/reference/invariants.md" .trellis/internal/invariants.md`),
// and decision-0051:75-78 specifies it as part of the trellis-authoritative
// half of the overlay. plugins/trellis/README.md:27 then binds the delivery:
// "Where a vendored `.trellis/internal/` still exists it remains
// authoritative ... the plugin's `reference/` files stay installation sources
// rather than runtime substitutes." Repointing a vendored project at the
// plugin's copy is exactly the runtime substitution that sentence forbids, so
// the token survives there — which is also what staleness.sh does, by never
// injecting into a vendored project at all.
//
// decision-0028's guard-per-pair applies because the repoint is now duplicated
// across two hosts in two languages (awk in staleness.sh, JS in
// codex-context.mjs). TestBothHostsRepointTheInvariantsPointerIdentically is
// that guard, and it is behavioural rather than textual: the two
// implementations share no line to diff, so the only thing that can be pinned
// is the address they hand the model.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// invariantsToken is the unresolved placeholder as it ships inside
// reference/trellis-{a,b}.md. Backticked because that is how the prose carries
// it and how both hooks match it — an unbackticked search would also hit this
// file's own prose in a grep, and more importantly would not prove the
// substitution replaced the whole delivered span.
const invariantsToken = "`.trellis/internal/invariants.md`"

// writeDualHostPluginRoot builds one plugin root that BOTH hooks accept: the
// real payload under reference/ (so the pointer's target is a real file, not a
// string that merely looks right) plus the .codex-plugin manifest
// codex-context.mjs's validPluginRoot requires.
func writeDualHostPluginRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range payloadFiles() {
		if err := os.WriteFile(filepath.Join(root, "reference", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, ".codex-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".codex-plugin", "plugin.json"), []byte(`{"name":"trellis"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// writeConfigOnlyProject is the decision-0065 shape: rules.toml and no
// .trellis/internal/. The .git directory writePluginNativeProject supplies is
// needed because codex-context.mjs bounds its overlay search at the nearest git
// boundary; staleness.sh does not need it and does not mind it.
//
// The firm posture is this helper's own choice, not a property of the shape —
// TRL-55 lifted the body into writePluginNativeProject so a test can pick the
// other one. Every caller here wants a posture it does not care about, so they
// keep asking for the shape and letting it default.
func writeConfigOnlyProject(t *testing.T) string {
	t.Helper()
	return writePluginNativeProject(t, "a")
}

// codexContextFor runs codex-context.mjs and returns the injected context. It
// fails loudly on the systemMessage branch rather than returning "", because a
// silent empty string would let every assertion below pass vacuously — the
// failure mode plugin_hook_test.go's own fixture comment records paying for.
func codexContextFor(t *testing.T, pluginRoot, project string) string {
	t.Helper()
	raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("the hook injected nothing; assertions below would pass vacuously: %s", raw)
	}
	return got.HookSpecificOutput.AdditionalContext
}

// claudeContextFor runs staleness.sh on the same shapes.
func claudeContextFor(t *testing.T, pluginRoot, project string) string {
	t.Helper()
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(hook)
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+project, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("staleness.sh exited non-zero (%v): %s", err, out)
	}
	return nudgeContext(t, strings.TrimSpace(string(out)))
}

// pointerIn extracts the backticked path the delivered prose tells the model to
// read. Anchored on the trigger's own surrounding words rather than on a path
// shape, so a channel that delivers a MANGLED pointer (the escape hazard
// staleness.sh's own comment block records) fails here instead of being
// matched by a laxer pattern.
func pointerIn(t *testing.T, context string) string {
	t.Helper()
	const lead = "read its entry in `"
	i := strings.Index(context, lead)
	if i < 0 {
		t.Fatalf("the delivered context carries no invariants trigger at all:\n%s", context)
	}
	rest := context[i+len(lead):]
	j := strings.Index(rest, "`")
	if j < 0 {
		t.Fatalf("the invariants trigger's path is not closed by a backtick:\n%s", context)
	}
	return rest[:j]
}

// TestCodexRepointsTheInvariantsPointerOnThePluginPath is the defect TRL-52
// reported, asserted as itself. Note what the last assertion checks: not that
// the string looks like a path, but that the file is THERE. "Names a path that
// exists on the path it is running" is the issue's own done-when, and only a
// stat proves it.
func TestCodexRepointsTheInvariantsPointerOnThePluginPath(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := writeConfigOnlyProject(t)
	context := codexContextFor(t, pluginRoot, project)

	if strings.Contains(context, invariantsToken) {
		t.Errorf("Codex was handed the unresolved placeholder; this mode has no .trellis/internal/ for it to name (decision-0065:106-111):\n%s", context)
	}
	want := filepath.Join(pluginRoot, "reference", "invariants.md")
	if got := pointerIn(t, context); got != want {
		t.Errorf("the repointed invariants pointer is wrong\nwant: %s\ngot:  %s", want, got)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("the pointer names a file that is not there — the defect moved rather than closed: %v", err)
	}
}

// TestCodexLeavesAVendoredInvariantsPointerAlone is the other half of the
// property, and it is the half a careless fix breaks. Without it, "repoint the
// pointer" reads as "repoint it always", which overrides a real vendored file
// with the plugin's copy — the runtime substitution README.md:27 rules out.
func TestCodexLeavesAVendoredInvariantsPointerAlone(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	// The file the retired setup skill wrote beside the rest of the overlay.
	// Present here because the assertion is about a pointer that RESOLVES, and
	// a fixture without it would be arguing for a pointer into thin air.
	if err := os.WriteFile(
		filepath.Join(project, ".trellis", "internal", "invariants.md"),
		[]byte(payloadFiles()["invariants.md"]), 0o644); err != nil {
		t.Fatal(err)
	}

	context := codexContextFor(t, pluginRoot, project)
	if !strings.Contains(context, invariantsToken) {
		t.Errorf("a vendored overlay is authoritative and carries its own invariants.md; the pointer must keep naming it (README.md:27):\n%s", context)
	}
	if strings.Contains(context, pluginRoot) {
		t.Errorf("a vendored project must not be pointed at the plugin's reference/ — those stay installation sources, not runtime substitutes:\n%s", context)
	}
}

// TestBothHostsRepointTheInvariantsPointerIdentically is decision-0028's guard
// for the pair. Run on ONE scratch repo, the way
// TestVendorAgentsBlockRefusalFollowsTheImportGate is, because the defect this
// exists to catch is invisible in either host alone: each hook can be
// self-consistent while the two hand the model different addresses for the
// same file in the same project.
func TestBothHostsRepointTheInvariantsPointerIdentically(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := writeConfigOnlyProject(t)

	claude := pointerIn(t, claudeContextFor(t, pluginRoot, project))
	codex := pointerIn(t, codexContextFor(t, pluginRoot, project))
	if claude != codex {
		t.Errorf("the two hosts name different invariants files for the same project — a fifth answer to a question that has one (decision-0028)\nstaleness.sh:      %s\ncodex-context.mjs: %s", claude, codex)
	}
	if _, err := os.Stat(claude); err != nil {
		t.Errorf("both hosts agree on a path that is not there: %v", err)
	}
}

// TestNoDeliveryChannelShipsTheUnresolvedPointer is the "a fourth consumer
// cannot invent a fifth answer" pin. The three channels below are the complete
// set that renders the posture prose for a model to read, and each is asserted
// through its own real output rather than by reading its source. A fourth
// channel is added HERE, with its resolved address, or this test says nothing
// about it — which is why the set is named in one place instead of three.
func TestNoDeliveryChannelShipsTheUnresolvedPointer(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := writeConfigOnlyProject(t)

	channels := map[string]string{
		"staleness.sh (Claude, plugin path)":            claudeContextFor(t, pluginRoot, project),
		"codex-context.mjs (Codex, plugin path)":        codexContextFor(t, pluginRoot, project),
		"reference/block-inline-tail.md (inline block)": payloadFiles()["block-inline-tail.md"],
	}
	for name, delivered := range channels {
		if strings.Contains(delivered, invariantsToken) {
			t.Errorf("%s delivers the unresolved placeholder to the model:\n%s", name, delivered)
		}
	}

	// install.sh's rendered file is the fourth channel. It is pinned by
	// TestInstallScriptRendersRulesFile (install_script_test.go:1066-1074),
	// which asserts both halves — the token gone, `.claude/skills/trellis/
	// reference/invariants.md` present — against a real install. Named rather
	// than re-run: vendoring the bundle costs a shell install per run, and a
	// second assertion of the same property would drift from it silently.
	if !strings.Contains(readFileT(t, installScriptPath(t)), "`.claude/skills/trellis/reference/invariants.md`") {
		t.Error("install.sh no longer rewrites the pointer to the copy it vendors; the channel named above has gone unpinned")
	}
}

// TestCodexRepointsWhenAVendoredOverlayLacksInvariants is TRL-58, and it is
// the third shape — the one the pair above does not cover between them.
//
// TestCodexLeavesAVendoredInvariantsPointerAlone writes invariants.md into the
// overlay before asserting, and says so: "a fixture without it would be
// arguing for a pointer into thin air." That fixture is the argument for
// leaving the token alone, and it is sound — WHERE THE FILE IS THERE. The
// overlay writer was the retired setup skill, and a skill is
// model-executed instructions rather than a script, so an overlay whose
// session skipped the `cp` is a real shape nothing rules out. Reached through
// writeValidCodexOverlay unmodified, which writes the three VENDORED_PAYLOAD
// files and no fourth.
//
// The delivery source does not move: prose, rules and version still come from
// the overlay, and this test asserts that below. Only the CONSULTED reference
// falls back, and only when the overlay has none — which is why the hook's own
// doctrine at codex-context.mjs:907-912 ("a missing one is a broken overlay
// that must fail loudly rather than silently falling through") does not reach
// it. That rule governs the three DELIVERED files, whose absence would make
// the injected chain wrong; invariants.md is read on demand, when a rule seems
// ambiguous, and a session that never hits an ambiguous rule never opens it.
// Failing a whole session closed over an on-demand reference trades a dead
// pointer for no governance at all.
//
// README.md:27's "installation sources rather than runtime substitutes" is
// likewise not in tension: nothing is being substituted FOR. There is no
// authoritative local copy to override, which is the whole premise of the
// sentence.
func TestCodexRepointsWhenAVendoredOverlayLacksInvariants(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	// Deliberately no invariants.md — the incomplete overlay. Asserted rather
	// than assumed, so this test cannot quietly become a duplicate of the
	// vendored test above if the helper ever starts writing a fourth file.
	if _, err := os.Stat(filepath.Join(project, ".trellis", "internal", "invariants.md")); err == nil {
		t.Fatal("fixture drift: writeValidCodexOverlay now writes invariants.md, so this test no longer covers the incomplete overlay")
	}

	context := codexContextFor(t, pluginRoot, project)

	if strings.Contains(context, invariantsToken) {
		t.Errorf("the overlay has no invariants.md, so the shipped token names nothing — TRL-52's defect surviving on the vendored arm:\n%s", context)
	}
	want := filepath.Join(pluginRoot, "reference", "invariants.md")
	if got := pointerIn(t, context); got != want {
		t.Errorf("the fallback pointer is wrong\nwant: %s\ngot:  %s", want, got)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("the pointer names a file that is not there — the defect moved rather than closed: %v", err)
	}

	// The delivery source must NOT have moved. If the rules body came from the
	// plugin rather than the overlay, this stopped being a pointer fallback and
	// became the silent mode switch :907-912 forbids.
	overlayRules := readFileT(t, filepath.Join(project, ".trellis", "internal", "rules.md"))
	if first := strings.SplitN(strings.TrimSpace(overlayRules), "\n", 2)[0]; first != "" && !strings.Contains(context, first) {
		t.Errorf("the injected rules no longer come from the overlay — the fallback switched delivery mode, not just the pointer\nwant a line from: %s", first)
	}
}

// TestCodexDoesNotFallBackOnAnUnusablePluginCopy is the other half of the test
// above, and it exists because that one cannot see this: it builds a HEALTHY
// plugin root, so the second half of the fallback condition is satisfied in
// every case it runs and is never actually exercised.
//
// decision-0093:2 makes both halves of that condition load-bearing and states
// what the second one is for: it "stops the fallback replacing one dead pointer
// with another". A bare stat does not deliver that — statSync needs no read
// permission, so a zero-byte or mode-0000 plugin copy satisfied existingFile,
// the fallback fired, and a vendored project's pointer was moved off its own
// authoritative address onto a file that yields nothing to read. That is the
// substitution the condition exists to prevent, performed by the check meant to
// prevent it. decision-0094:4 rules that this half asks payloadDefect instead.
//
// Reverting that one call to existingFile is a one-line mutant that the rest of
// the suite does not kill, which is what this test is for. The repo has paid
// for the shape before: TRL-52 shipped a broken pointer on a branch no fixture
// reached.
//
// The token SURVIVING is the assertion, and on this branch that is the right
// outcome rather than a lesser evil: a vendored project really does have a
// `.trellis/internal/`, so naming it is not the lie it would be on the
// plugin-native arm. Both addresses are dead here — that residue is TRL-71's,
// not this test's.
func TestCodexDoesNotFallBackOnAnUnusablePluginCopy(t *testing.T) {
	for _, tc := range invariantsFaults() {
		// The healthy row is the control for the test above, not for this one:
		// with a good plugin copy the fallback SHOULD fire. Skipping it here
		// keeps this test a statement about unusable copies only.
		if tc.why == "" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			pluginRoot := writeDualHostPluginRoot(t)
			project := newGitProject(t)
			writeValidCodexOverlay(t, project)
			overlayCopy := filepath.Join(project, ".trellis", "internal", "invariants.md")
			if _, err := os.Stat(overlayCopy); err == nil {
				t.Fatal("fixture drift: writeValidCodexOverlay now writes invariants.md, so the fallback arm is no longer reached")
			}
			tc.breakIt(t, filepath.Join(pluginRoot, "reference", "invariants.md"))

			context := codexContextFor(t, pluginRoot, project)

			// Delivery is untouched: only the pointer was ever in question
			// (decision-0093:1).
			if !strings.Contains(context, rulesLoadedSentinel) {
				t.Fatalf("the overlay's rules must still be delivered whole:\n%s", context)
			}
			if got := pointerIn(t, context); got != ".trellis/internal/invariants.md" {
				t.Errorf("the fallback fired onto a plugin copy that %s — decision-0093:2's second half exists to stop exactly this substitution\nwant the overlay's own address to survive: .trellis/internal/invariants.md\ngot:  %s", tc.why, got)
			}
		})
	}
}

// invariantsReportLead is the phrase BOTH hosts must produce when the plugin
// payload has no readable invariants copy, immediately followed by the absolute
// path. Pinning a shared PHRASE rather than a shared path is what makes the
// assertion non-vacuous on the Claude side: staleness.sh delivers its report
// inside the very context that already carries the repointed pointer, so
// "the output names the path" would be satisfied by the pointer alone.
const invariantsReportLead = "no readable "

// invariantsReportClaim is the second phrase the two hosts must share, and
// invariantsReportFalseClaim is the wording it replaced.
//
// The two reports are NOT identical sentences and must not be — Claude says
// "the rules above" where Codex says "the context just injected", because the
// report sits in different places. What they may not differ on is the CLAIM
// they make about the file. #295 changed "names a file that cannot be read" to
// "yields nothing to read" on the Claude side, because the first is false for
// exactly one of the four classifications: an empty file reads fine, it just
// yields nothing. It left Codex alone on the stated ground that "Claude is the
// only host that fires on `empty`" — a premise TRL-72 makes false, so Codex
// shipped `(it is empty), so … names a file that cannot be read` until #296.
//
// Pinned here because nothing else can see it: the pair guard compares the lead
// phrase and the classification, and both stayed correct while the sentence
// between them contradicted itself.
const invariantsReportClaim = "yields nothing to read"
const invariantsReportFalseClaim = "cannot be read"

// invariantsFault is one way the plugin's `reference/invariants.md` can be
// broken, paired with the classification BOTH hosts must hand the reader for
// it.
//
// TRL-72 made this table shared. Until then codex-context.mjs asked
// existingFile — a bare `statSync().isFile()`, which succeeds on a zero-byte
// file and on one at mode 0000, because stat needs no read permission — so two
// of the four shapes below were reported by Claude and passed over by Codex in
// silence, with the pair guard green. The two hosts now ask the same question
// and answer it in the same words, and one table drives both guards.
type invariantsFault struct {
	name string
	// breakIt applies exactly one fault to the plugin's invariants copy. nil is
	// the discrimination control: a healthy payload reports nothing.
	breakIt func(t *testing.T, target string)
	// why is the phrase the report must carry, VERBATIM. It is staleness.sh's
	// $payload_why, and codex-context.mjs mirrors those four strings byte for
	// byte. The shared vocabulary is what makes "both hosts said the same
	// thing" a check rather than a hope — decision-0028's guard-per-pair,
	// applied to a string table that now lives in two files in two languages
	// and can only be pinned behaviourally. Empty means the payload is healthy
	// and neither host may say anything.
	why string
}

// classification is `why` without the remedy hint that follows the em dash.
//
// Derived rather than stored, so the two halves cannot drift. Exclusion is
// asserted on this STEM and not on the full phrase: a report that flattened the
// four cases into a bare "is not a readable file" with no tail would satisfy an
// exclusion keyed on the full phrase while telling the reader nothing about
// which remedy applies, which is the failure both guards below exist to catch.
func (f invariantsFault) classification() string {
	return strings.SplitN(f.why, " — ", 2)[0]
}

// invariantsFaults is every shape payload_read classifies, applied to the
// plugin's own copy of the consulted reference. Listed in ONE place so a case
// cannot quietly stop being covered by one of the two guards below, and so a
// classification added to payload_read has exactly one place to be added here.
//
// The table carries three cases that are not duplicates of the zero-byte one,
// and each exists to kill a different way of writing the emptiness test wrong.
// payload_read reaches "empty" through `$(cat "$1")`, and `$( )` discards
// exactly two byte values: TRAILING NEWLINES and NUL. So:
//
//   - newlines only, and NUL only, are EMPTY on the Claude side. A Codex mirror
//     written `readFileSync(…).length === 0` — the obvious one — agrees on the
//     zero-byte file and diverges on both.
//   - a single SPACE is HEALTHY on both hosts, and it is in this table as a
//     control for the opposite mistake: a scan that widened to whitespace
//     (`… && chunk[i] !== 0x20`) matches payload_read on every broken shape and
//     diverges only here. An earlier version of this comment argued the space
//     case out of the table BECAUSE both hosts call it healthy, which is
//     backwards — that is precisely what makes it the discriminating control.
//     Review of #296 killed a whitespace-widening mutant that survived without
//     it.
//
// The MIXED NUL shapes ("a\x00b", "\x00a") stay out, and that exclusion is not
// the same mistake: they agree on every shell measured, but POSIX leaves NUL
// handling in `$( )` unspecified, so pinning them would assert a portability
// property nobody has checked rather than discriminate between implementations.
func invariantsFaults() []invariantsFault {
	return []invariantsFault{{
		name: "a complete payload",
	}, {
		name:    "a copy holding a single space is healthy on both hosts",
		breakIt: func(t *testing.T, target string) { writeFileT(t, target, " ") },
	}, {
		name:    "an absent copy is reported as missing",
		breakIt: func(t *testing.T, target string) { removeFileT(t, target) },
		why:     "is missing",
	}, {
		name:    "a zero-byte copy is reported as empty, not as missing",
		breakIt: func(t *testing.T, target string) { writeFileT(t, target, "") },
		why:     "is empty",
	}, {
		name:    "a copy holding nothing but newlines is reported as empty",
		breakIt: func(t *testing.T, target string) { writeFileT(t, target, "\n\n\n") },
		why:     "is empty",
	}, {
		// The post-crash zero-fill. `$( )` discards NUL bytes as well as
		// trailing newlines — a shell variable holds a C string and cannot
		// carry a NUL — so this file is empty on the Claude side too. Measured
		// EMPTY on bash (the sibling hook's shebang), sh and dash; only zsh,
		// which does not run either hook, disagrees. The MIXED shapes ("a\x00b")
		// are deliberately not pinned: they agree on every shell measured, but
		// POSIX leaves NUL handling in `$( )` unspecified and a truncating
		// shell would diverge, so pinning them would assert portability nobody
		// has measured.
		name:    "a copy holding nothing but NUL bytes is reported as empty",
		breakIt: func(t *testing.T, target string) { writeFileT(t, target, "\x00\x00\x00") },
		why:     "is empty",
	}, {
		name: "an unreadable copy is reported as a permission fault, not as missing",
		breakIt: func(t *testing.T, target string) {
			if os.Geteuid() == 0 {
				t.Skip("running as root: mode 0000 does not deny reads, so the fixture cannot be built")
			}
			if err := os.Chmod(target, 0o000); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(target, 0o644) })
			if _, err := os.ReadFile(target); err == nil {
				t.Skipf("premise: %s is still readable at mode 0000", target)
			}
		},
		why: "exists but could not be read — a permission mode, a stale ACL, or a symlink whose target is gone",
	}, {
		name: "a directory at the path is reported as not a readable file",
		breakIt: func(t *testing.T, target string) {
			removeFileT(t, target)
			if err := os.Mkdir(target, 0o755); err != nil {
				t.Fatal(err)
			}
		},
		why: "is not a readable file — a directory or a device sits at that path",
	}}
}

// invariantsClassifications is every classification the table can produce,
// deduplicated (the two empty shapes share one). Derived from the table rather
// than listed again, so a classification added to payload_read and to the table
// is excluded everywhere at once — and one added to payload_read and NOT to the
// table narrows the exclusions visibly, in the table, instead of in a second
// slice nobody thinks to update.
func invariantsClassifications() []string {
	seen := map[string]bool{}
	var all []string
	for _, f := range invariantsFaults() {
		if f.why == "" || seen[f.classification()] {
			continue
		}
		seen[f.classification()] = true
		all = append(all, f.classification())
	}
	return all
}

// assertInvariantsReport is the shared body of the two guards below: given what
// one host SAID about one fault, it checks the host said the right thing and
// only that thing.
//
// The absence half is not belt-and-braces, it is the whole assertion. An
// earlier version of the Claude guard checked only that the expected phrase was
// present, and review killed it with a one-line mutant: a report that names
// every classification at once —
//
//	inv_defect="is missing, is empty, is not a readable file, or exists but
//	            could not be read -- this report does not know which"
//
// passed every case and the whole package, while doing exactly the flattening
// these guards exist to prevent. Presence alone cannot distinguish "told the
// reader which" from "told the reader all four"; only exclusion can.
func assertInvariantsReport(t *testing.T, host, report, target string, fault invariantsFault) {
	t.Helper()
	said := strings.Contains(report, invariantsReportLead+target)
	if fault.why == "" {
		if said {
			t.Errorf("%s reports a broken payload on a complete one — the over-correction is as bad for a reader as the silence\ngot: %q", host, report)
		}
		// None of the classifications may appear either. This is also the
		// premise every exclusion below rests on: if a phrase turned up in a
		// healthy delivery, excluding it there would prove nothing.
		for _, why := range invariantsClassifications() {
			if strings.Contains(report, why) {
				t.Errorf("%s already carries %q on a healthy payload, so excluding it elsewhere would prove nothing\ngot: %q", host, why, report)
			}
		}
		return
	}
	if !said {
		t.Errorf("%s repointed at a file it cannot read and said nothing about it — the address guard is green and the hosts disagree on the diagnosis (TRL-70, TRL-72)\nwant a report naming: %s%s\ngot: %q", host, invariantsReportLead, target, report)
		return
	}
	if !strings.Contains(report, fault.why) {
		t.Errorf("%s flattened the classification; the reader is told the file is unusable without being told which remedy applies\nwant: %q\ngot: %q", host, fault.why, report)
	}
	// The claim the sentence makes about the file, which both hosts share even
	// though the sentences around it differ.
	if !strings.Contains(report, invariantsReportClaim) {
		t.Errorf("%s does not say what is actually true of all four classifications\nwant: %q\ngot: %q", host, invariantsReportClaim, report)
	}
	if strings.Contains(report, invariantsReportFalseClaim) {
		t.Errorf("%s says the file %q while classifying it %q — false for an empty file, which reads fine and yields nothing (#295 corrected this on the other host)\ngot: %q", host, invariantsReportFalseClaim, fault.why, report)
	}
	for _, why := range invariantsClassifications() {
		if why == fault.classification() || !strings.Contains(report, why) {
			continue
		}
		t.Errorf("%s names %q as well as %q — a report that lists every classification has told the reader nothing about which remedy applies, and presence alone cannot catch that\ngot: %q", host, why, fault.classification(), report)
	}
}

// TestBothHostsReportAMissingInvariantsTarget is TRL-70 and TRL-72, and it is
// the half of decision-0028's guard-per-pair that
// TestBothHostsRepointTheInvariantsPointerIdentically does not cover. That test
// pins the ADDRESS the two hosts hand the model, on a healthy fixture. It
// cannot see a divergence in what they SAY about a broken one — which is
// exactly how they diverged twice: #294 taught codex-context.mjs to report a
// missing `reference/invariants.md` while staleness.sh went on repointing at it
// in silence, and #295 then made Claude the stricter host on the two shapes a
// bare stat cannot see. Both times the address guard was green.
//
// The two hosts use DIFFERENT CHANNELS, deliberately, and this test records
// that rather than forcing them together:
//
//   - codex-context.mjs reports on `systemMessage`, a field that rides
//     alongside a successful delivery, which is where its floor-row warning
//     already goes.
//
//   - staleness.sh has no such field. `emit` writes one key,
//     hookSpecificOutput.additionalContext, and every TRELLIS_ marker it can
//     print is emitted INSTEAD of a payload — a marker means this hook injected
//     nothing. So its report goes inside the delivered context, beside the
//     provenance line that already degrades there when the version stamp cannot
//     be read (TRL-34).
//
//     An earlier version of this comment said a marker is "a refusal, not a
//     note", and review measured that false: TRELLIS_STALENESS_UNKNOWN
//     (staleness.sh:523 and :663) is a note on a session that IS governed —
//     "Nothing is wrong with this project and no rules are missing". The
//     injects-nothing half is true of every marker and carries the argument on
//     its own; the refusal half did not, and overstating it here would have
//     made the design read stronger than what was measured.
//
// What must match is the SUBSTANCE, and that is what is asserted: on every
// shape payload_read classifies, both hosts name the unreadable absolute path,
// give it the SAME classification verbatim and no other, still deliver a
// governing context, and still repoint the pointer at that same path; on a
// complete payload neither says anything.
//
// TRL-72 removed the divergence this test used to record in a comment rather
// than pin. The classification cases are no longer the Claude host's alone: the
// table drives both hosts, so a host that stats where the other reads fails
// here, at the shape it cannot see.
func TestBothHostsReportAMissingInvariantsTarget(t *testing.T) {
	for _, tc := range invariantsFaults() {
		t.Run(tc.name, func(t *testing.T) {
			pluginRoot := writeDualHostPluginRoot(t)
			project := writeConfigOnlyProject(t)
			target := filepath.Join(pluginRoot, "reference", "invariants.md")

			// The Codex baseline, taken on THIS plugin root before the fault is
			// applied, so the two contexts are comparable byte for byte. A
			// second fixture would differ in its temp path alone.
			_, healthy := runCodexHook(t, pluginRoot, startupInput(t, project))
			if healthy.HookSpecificOutput == nil {
				t.Fatalf("the healthy fixture delivers nothing, so the comparison below would measure one absence against another")
			}

			if tc.breakIt != nil {
				tc.breakIt(t, target)
			}

			claude := claudeContextFor(t, pluginRoot, project)
			raw, codex := runCodexHook(t, pluginRoot, startupInput(t, project))
			if codex.HookSpecificOutput == nil {
				t.Fatalf("a consulted reference is not a delivered payload (decision-0093:1): Codex must still inject the rules: %s", raw)
			}
			// The Codex report costs the injected context NOTHING: it rides
			// systemMessage, outside the MAX_CONTEXT_BYTES bound on `context`.
			// That is the property TestTheMissingInvariantsReportNeverCosts-
			// TheSessionItsRules pins on the Claude side by ordering the report
			// after the budget check, and it is the one way a report about a
			// CONSULTED file could fail a session closed (decision-0093:1). On
			// Codex it holds structurally, so the way to keep it is to pin the
			// structure: moving this warning into the context breaks the
			// equality below rather than waiting for a payload large enough to
			// notice.
			if got, want := codex.HookSpecificOutput.AdditionalContext, healthy.HookSpecificOutput.AdditionalContext; got != want {
				t.Errorf("the Codex report changed the injected context; a report about a consulted file must not spend the context budget (decision-0093:1)\nwant %d bytes identical to the healthy delivery, got %d", len(want), len(got))
			}

			for _, host := range []struct {
				name string
				// context is what the model was handed.
				context string
				// report is the channel that host says a broken install on.
				// For staleness.sh it IS the context, because emit() is the
				// only channel it has.
				report string
			}{
				{name: "staleness.sh (Claude)", context: claude, report: claude},
				{name: "codex-context.mjs (Codex)", context: codex.HookSpecificOutput.AdditionalContext, report: codex.SystemMessage},
			} {
				// Delivery survives every fault, on both hosts: invariants.md
				// is consulted on demand and never injected, and
				// decision-0093:1 rules that failing a session closed over such
				// a file trades a dead pointer for no governance at all.
				if !strings.Contains(host.context, rulesLoadedSentinel) || !deliveredRow(host.context, "floor-intent-gate") {
					t.Errorf("%s did not deliver a complete governed context; a broken consulted reference must not cost the session its rules (decision-0093:1):\n%s", host.name, host.context)
				}
				// The pointer moves either way. This is the assertion the
				// symmetric guard fails: leaving the raw token would ship
				// `.trellis/internal/` to the one mode defined by not having it.
				if strings.Contains(host.context, invariantsToken) {
					t.Errorf("%s shipped the unresolved placeholder; a broken target must not degrade the pointer back to a directory this mode cannot have:\n%s", host.name, host.context)
				}
				if got := pointerIn(t, host.context); got != target {
					t.Errorf("%s names the wrong invariants file\nwant: %s\ngot:  %s", host.name, target, got)
				}
				assertInvariantsReport(t, host.name, host.report, target, tc)
			}
		})
	}
}

// TestStalenessNamesWhyTheInvariantsTargetCannotBeRead is the Claude-side
// statement of the same table, and it pins the two things the pair guard cannot
// see: WHERE the report sits in the delivered context, and how many times it
// appears. Those exist only on this host, because staleness.sh has no channel
// beside its payload and appends the report to the context itself.
//
// The classification assertions are shared with the pair guard on purpose
// rather than duplicated by accident: this test states what the Claude host
// must say, the pair guard states that both hosts say the same thing, and the
// two fail with different messages when one host alone regresses.
func TestStalenessNamesWhyTheInvariantsTargetCannotBeRead(t *testing.T) {
	for _, tc := range invariantsFaults() {
		t.Run(tc.name, func(t *testing.T) {
			pluginRoot := writeDualHostPluginRoot(t)
			project := writeConfigOnlyProject(t)
			target := filepath.Join(pluginRoot, "reference", "invariants.md")
			if tc.breakIt != nil {
				tc.breakIt(t, target)
			}

			context := claudeContextFor(t, pluginRoot, project)
			if !strings.Contains(context, rulesLoadedSentinel) || !deliveredRow(context, "floor-intent-gate") {
				t.Fatalf("the rules must still be delivered whole; only the consulted reference is broken (decision-0093:1):\n%s", context)
			}
			if got := pointerIn(t, context); got != target {
				t.Errorf("the pointer names the wrong invariants file\nwant: %s\ngot:  %s", target, got)
			}
			assertInvariantsReport(t, "staleness.sh", context, target, tc)
			if tc.why == "" || !strings.Contains(context, invariantsReportLead+target) {
				// Position and count both index on this phrase, and
				// strings.Index returns -1 when it is absent — which would make
				// the position assertion pass on a report that is not there at
				// all. assertInvariantsReport has already said so.
				return
			}
			// The report is APPENDED, and its own words depend on that: it says
			// "the invariants pointer in the rules above" and "The rules and
			// rows above are complete". Review pointed out that a prepending
			// implementation passed every assertion here while making both
			// sentences false, and that a double-append passed too. Position
			// and count are cheap to state and were being left to luck.
			if strings.Index(context, invariantsReportLead+target) < strings.Index(context, rulesLoadedSentinel) {
				t.Errorf("the report leads the context instead of following the rules it refers to as being above it:\n%s", context)
			}
			if n := strings.Count(context, invariantsReportLead+target); n != 1 {
				t.Errorf("the report appears %d times; one broken file earns one report", n)
			}
		})
	}
}

// TestTheMissingInvariantsReportNeverCostsTheSessionItsRules is the guard for
// the ORDERING of the report against the 32768-byte injection budget, which is
// the one thing about this report that can fail a session closed.
//
// Found in review of the first version of this change, which assembled the
// report INSIDE the `payload="$( … )"` block, before the budget check. On a
// project whose rules.toml put the payload 32573 bytes into the budget, the
// session was fully governed — and deleting `reference/invariants.md`, changing
// nothing else, pushed it to 33149 and turned the whole thing into a
// TRELLIS_RULES_NOT_LOADED refusal with no rules and no rows. A consulted
// reference is not a delivered payload (decision-0093:1): trading a dead
// pointer for no governance at all is exactly what that rule forbids, and the
// hook's own comments claimed it could not happen.
//
// It is also the property the two hosts must agree on. codex-context.mjs bounds
// `context` alone (:1194) and pushes the same warning onto systemMessage
// afterwards (:1333), so Codex can never lose its context to this warning.
//
// The fixture CALIBRATES rather than hardcoding a padding size: it measures the
// payload at two padding sizes, derives the per-line cost, and solves for a
// rules.toml that lands the healthy payload just inside the budget. A hardcoded
// count would drift out of the dangerous band the first time the shipped
// payload changed size, and this test would then pass while proving nothing —
// which is why the band itself is asserted.
func TestTheMissingInvariantsReportNeverCostsTheSessionItsRules(t *testing.T) {
	// staleness.sh's own budget, READ OUT OF THE HOOK rather than duplicated.
	//
	// An earlier version hardcoded 32768 and argued that a budget moving
	// underneath it would fail loudly. Review measured that true in one
	// direction only. Raised to 65536 — the likelier direction as context
	// windows grow — the calibration went on aiming at the stale constant, the
	// broken payload landed comfortably inside the hook's real budget, nothing
	// refused, and every assertion below passed; with the reviewed in-budget
	// ordering restored at the same time, the whole four-test set went green
	// and the regression this test exists for went undetected. A test that
	// measures against a number the code no longer uses is not pinned to the
	// code. Parsed the way TestNoPayloadReadBypassesTheGateway parses this same
	// file, and fatal if the declaration cannot be found.
	limit := stalenessInjectionBudget(t)
	const padLine = "# padding, so this fixture sits exactly where the budget bites\n"

	pluginRoot := writeDualHostPluginRoot(t)
	project := writeConfigOnlyProject(t)
	toml := filepath.Join(project, ".trellis", "rules.toml")
	base := readFileT(t, toml)

	measure := func(lines int) int {
		writeFileT(t, toml, base+strings.Repeat(padLine, lines))
		return len(claudeContextFor(t, pluginRoot, project))
	}

	at0, at64 := measure(0), measure(64)
	perLine := (at64 - at0) / 64
	if perLine <= 0 {
		t.Fatalf("calibration failed: 64 padding lines moved the payload by %d bytes, so the fixture cannot be aimed", at64-at0)
	}
	// Aim just under the budget, then correct — integer division and the row
	// reconciler's own wording make one pass approximate.
	//
	// The band below is an AIMING AID, not the property: the decisive assertion
	// is that the broken context exceeds the budget, and it fails loudly on its
	// own terms if the aim lands short. The one thing the aim must not do is
	// panic before saying anything, which a negative count in strings.Repeat
	// would — reachable if the shipped payload ever grows past the budget on
	// its own.
	lines := (limit - 300 - at0) / perLine
	if lines <= 0 {
		t.Fatalf("the payload is already %d bytes against a %d-byte budget, so there is no room to pad a fixture toward the band — this test can no longer be aimed and is not silently passing instead", at0, limit)
	}
	healthy := measure(lines)
	for i := 0; i < 8 && (healthy > limit || healthy < limit-700); i++ {
		lines += (limit - 300 - healthy) / perLine
		healthy = measure(lines)
	}
	if healthy > limit || healthy < limit-700 {
		t.Fatalf("could not aim the fixture into the band: %d bytes against a %d-byte budget, after %d padding lines — a fixture outside the band proves nothing", healthy, limit, lines)
	}
	if !strings.Contains(claudeContextFor(t, pluginRoot, project), rulesLoadedSentinel) {
		t.Fatalf("the calibrated fixture is not governed even with a healthy payload; the assertion below would prove nothing")
	}

	// The whole change: this file going missing must cost the session nothing
	// but the pointer it names.
	target := filepath.Join(pluginRoot, "reference", "invariants.md")
	removeFileT(t, target)
	broken := claudeContextFor(t, pluginRoot, project)

	if strings.Contains(broken, "TRELLIS_RULES_NOT_LOADED") {
		t.Fatalf("deleting a CONSULTED reference refused the session — a dead pointer traded for no governance at all (decision-0093:1), and the budget is the only thing that changed:\n%s", broken)
	}
	if !strings.Contains(broken, rulesLoadedSentinel) || !deliveredRow(broken, "floor-intent-gate") {
		t.Errorf("the rules and rows must survive a missing consulted reference:\n%s", broken)
	}
	if !strings.Contains(broken, invariantsReportLead+target) {
		t.Errorf("the report was dropped rather than delivered outside the budget:\n%s", broken)
	}
	// THE assertion, and the one the earlier version was missing. `len` on a Go
	// string is bytes, the same measure `wc -c` gives the hook. A delivered
	// context LARGER than the budget can only mean the report was appended
	// after the check: had it been assembled inside, the hook would have
	// refused instead of delivering this. Without it the test asserted only
	// that the context grew, which is true of an appended report anywhere,
	// under any budget.
	if len(broken) <= limit {
		t.Fatalf("the fixture never crossed the budget: healthy %d bytes plus the report is %d, still inside the %d-byte budget — an implementation that assembled the report INSIDE the budget would pass this test unchanged, so it proves nothing as it stands", healthy, len(broken), limit)
	}
}

// stalenessInjectionBudget reads the injection budget out of staleness.sh, so a
// test calibrating a fixture against it cannot drift onto a number the hook has
// stopped using. Fatal rather than defaulted: a fixture aimed at a guessed
// budget is exactly the silent pass this exists to prevent.
func stalenessInjectionBudget(t *testing.T) int {
	t.Helper()
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^limit=([0-9]+)$`).FindStringSubmatch(readFileT(t, hook))
	if m == nil {
		t.Fatal("staleness.sh no longer declares `limit=<n>` at column 0, so the budget this fixture aims at cannot be read from the code it is meant to pin")
	}
	limit, err := strconv.Atoi(m[1])
	if err != nil || limit <= 0 {
		t.Fatalf("staleness.sh declares an unusable injection budget %q: %v", m[1], err)
	}
	return limit
}

// writeDefaultsShapeProject is the OTHER arm of path B: decision-0070's
// shipped-defaults shape, with no `.trellis/rules.toml` at all and the plugin
// vendored inside the repository — which is the adoption act that makes the
// shipped rows apply (decision-0070 D6, staleness.sh's own `rows_are_default`).
// It reaches the same delivery block, and therefore the same repoint and the
// same report, through a different door.
func writeDefaultsShapeProject(t *testing.T) (pluginRoot, project string) {
	t.Helper()
	project = t.TempDir()
	pluginRoot = filepath.Join(project, ".claude", "skills", "trellis")
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range payloadFiles() {
		writeFileT(t, filepath.Join(pluginRoot, "reference", name), body)
	}
	return pluginRoot, project
}

// TestTheMissingInvariantsReportFiresOnTheDefaultsShapeToo covers the arm the
// two tests above do not reach. Both build their project with
// writeConfigOnlyProject, so both enter path B through `.trellis/rules.toml`;
// a project with none of it, governed by the rows the plugin ships, enters the
// same block by a different route.
//
// Worth its own case because this repo has already paid for the assumption that
// one door covers the other: TRL-52 shipped a broken pointer on the Codex
// plugin-native branch precisely because no fixture reached that branch, which
// codex_plugin_native_test.go's header records at length. The report is a
// property of the delivery block, not of the door, and this asserts that rather
// than assuming it.
func TestTheMissingInvariantsReportFiresOnTheDefaultsShapeToo(t *testing.T) {
	pluginRoot, project := writeDefaultsShapeProject(t)
	target := filepath.Join(pluginRoot, "reference", "invariants.md")

	healthy := claudeContextFor(t, pluginRoot, project)
	if !strings.Contains(healthy, rulesLoadedSentinel) || !deliveredRow(healthy, "floor-intent-gate") {
		t.Fatalf("fixture drift: this shape no longer delivers the shipped defaults, so the assertions below would prove nothing:\n%s", healthy)
	}
	if strings.Contains(healthy, invariantsReportLead+target) {
		t.Errorf("a complete payload was reported as broken on the defaults shape:\n%s", healthy)
	}

	removeFileT(t, target)
	broken := claudeContextFor(t, pluginRoot, project)
	if !strings.Contains(broken, rulesLoadedSentinel) || !deliveredRow(broken, "floor-intent-gate") {
		t.Errorf("the shipped defaults must still govern a project whose plugin lacks its invariants copy:\n%s", broken)
	}
	if got := pointerIn(t, broken); got != target {
		t.Errorf("the pointer names the wrong invariants file on the defaults shape\nwant: %s\ngot:  %s", target, got)
	}
	if !strings.Contains(broken, invariantsReportLead+target) {
		t.Errorf("the defaults shape reaches the same delivery block and must report the same broken payload:\n%s", broken)
	}
}

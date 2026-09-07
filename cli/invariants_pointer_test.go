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
	"bytes"
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
	return nudgeContext(t, claudeStdoutFor(t, pluginRoot, project))
}

// claudeStdoutFor is claudeContextFor's raw half: staleness.sh's stdout,
// trimmed, WITHOUT requiring that it parse as a delivered context.
//
// Split out for the vendored fixtures, where staleness.sh legitimately emits
// NOTHING -- path A opens at `if [ -d "$internal" ]` (staleness.sh:482) and
// every route out of it exits, unconditionally at :529, long before the repoint
// at :1403. An earlier version of this comment cited :559, which is a comment
// inside path C; review caught it. nudgeContext fails on
// that empty string, which is right for every caller that expects a delivery
// and wrong for a caller asserting the absence of one.
func claudeStdoutFor(t *testing.T, pluginRoot, project string) string {
	t.Helper()
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(hook)
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+project, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
	// THE TWO STREAMS ARE READ SEPARATELY, and that is a correctness fix rather
	// than tidying. This used to be CombinedOutput, which conflates them — and
	// the hook's contract is exactly one JSON object on STDOUT, so anything the
	// shell says on stderr was being spliced into the payload under test.
	//
	// Measured, on the NUL fixture: bash >= 4.4 prints
	//
	//   staleness.sh: line 192: warning: command substitution: ignored null
	//   byte in input
	//
	// which is bash reporting its OWN discarding of NULs — the very behaviour
	// payload_read relies on to call that file empty. It goes to stderr, the
	// hook's JSON is untouched on stdout, and both hosts still classify the
	// file correctly. macOS ships bash 3.2, which is silent, so this reproduced
	// only on the Linux CI runner: a genuine platform split that CombinedOutput
	// turned into "output is not valid JSON" and would have misattributed to
	// the hook.
	//
	// stderr is not asserted on, only reported: a shell warning is not this
	// suite's to police, but a reader debugging a stdout failure wants to see
	// it.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("staleness.sh exited non-zero (%v)\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	if stderr.Len() > 0 {
		t.Logf("staleness.sh wrote to stderr (not a failure; stdout carries the contract):\n%s", stderr.String())
	}
	return strings.TrimSpace(stdout.String())
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

// TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting is TRL-73,
// and it is the guard decision-0095:3 named as
// TestCodexSaysNothingWhenTheOverlayCarriesItsOwnInvariants. It is not a
// parallel guard: it keeps that test's fixture, its shape table and its
// subject — what the hook does when the overlay has its own copy, in any state.
// What changed is that four of the five shapes stopped asserting SILENCE and
// started asserting the two properties silence was standing in for.
//
// decision-0095:3 could pin those four on silence because the eligibility check
// had not been widened and nothing else could fire. decision-0096 rules that a
// dead pointer is reported on this arm too, so silence is no longer available as
// a proxy — and the property it proxied for is now asserted DIRECTLY and more
// strongly: the pointer must still name the OVERLAY's own address. A mutation
// widening `existingFile` to `payloadDefect` moves that address to the plugin's
// copy and fails here.
//
// The healthy row keeps asserting silence, and that assertion is untouched: it
// is decision-0095:3's real over-correction guard, added after review found a
// mutant firing the report on every vendored project survived the whole suite.
// Nothing else covers it — TestCodexLeavesAVendoredInvariantsPointerAlone reads
// through codexContextFor, which discards systemMessage;
// TestCodexDoesNotFallBackOnAnUnusablePluginCopy skips the healthy row outright;
// and assertInvariantsReport's over-correction arm runs only on the pair guard's
// config-only fixture, where this arm cannot fire at all.
//
// EACH SHAPE RUNS THE HOOK TWICE, against a healthy plugin copy and then a
// missing one, and requires the same answer both times. That is decision-0096:2
// stated as a check rather than as prose: on this arm the plugin's copy is
// ineligible BY RULE, not unavailable, so nothing about its state may reach the
// reader. It is also what separates this cell from decision-0095's, where the
// plugin copy's own fault is load-bearing and is named.
func TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting(t *testing.T) {
	for _, shape := range overlayOwnCopyShapes() {
		t.Run(shape.name, func(t *testing.T) {
			pluginRoot := writeDualHostPluginRoot(t)
			project := newGitProject(t)
			writeValidCodexOverlay(t, project)
			overlayCopy := filepath.Join(project, ".trellis", "internal", "invariants.md")
			shape.write(t, overlayCopy)

			// The plugin's copy is HEALTHY on the first run, and that is the row
			// where substituting is most tempting and most wrong: a usable copy
			// sits one directory away and the overlay still owns the address
			// (decision-0096:1). The silence-era fixture removed it before the
			// only run it made, so this row was never covered.
			withCopy := assertOverlayOwnCopyReport(t, pluginRoot, project, overlayCopy, shape)
			removeFileT(t, filepath.Join(pluginRoot, "reference", "invariants.md"))
			withoutCopy := assertOverlayOwnCopyReport(t, pluginRoot, project, overlayCopy, shape)

			if withCopy != withoutCopy {
				t.Errorf("the report changed with the plugin copy's state, but that copy may not stand in for this one whatever state it is in (decision-0096:2)\nwith a healthy plugin copy: %q\nwith none:                  %q", withCopy, withoutCopy)
			}
		})
	}
}

// overlayOwnCopyShapes is every state the overlay's OWN copy can be in while
// still satisfying `existingFile` — which is what decides eligibility, and is
// deliberately untouched by decision-0096. `why` is the classification the
// report must carry, and "" is the healthy control that must be reported on at
// all.
//
// The four unusable shapes are the four decision-0094's "What is not decided
// here" and TRL-73 name — its Consequences list the PLUGIN-side four, which is a
// different table — and they are not duplicates of each other: payload_read reaches "empty"
// through `$(cat)`, which strips trailing newlines AND NUL bytes, so a
// newline-only and an all-NUL file are empty on the other host too. A JS mirror
// written as `readFileSync(...).length === 0` passes the zero-byte row and fails
// the other two.
func overlayOwnCopyShapes() []struct {
	name  string
	why   string
	write func(t *testing.T, target string)
} {
	return []struct {
		name  string
		why   string
		write func(t *testing.T, target string)
	}{{
		name:  "a healthy overlay copy",
		why:   "",
		write: func(t *testing.T, target string) { writeFileT(t, target, payloadFiles()["invariants.md"]) },
	}, {
		name:  "a zero-byte overlay copy",
		why:   "is empty",
		write: func(t *testing.T, target string) { writeFileT(t, target, "") },
	}, {
		name:  "an overlay copy holding nothing but newlines",
		why:   "is empty",
		write: func(t *testing.T, target string) { writeFileT(t, target, "\n\n\n") },
	}, {
		name:  "an overlay copy holding nothing but NUL bytes",
		why:   "is empty",
		write: func(t *testing.T, target string) { writeFileT(t, target, "\x00\x00\x00") },
	}, {
		name: "an unreadable overlay copy",
		why:  "exists but could not be read — a permission mode, a stale ACL, or a symlink whose target is gone",
		write: func(t *testing.T, target string) {
			if os.Geteuid() == 0 {
				t.Skip("running as root: mode 0000 does not deny reads, so the fixture cannot be built")
			}
			writeFileT(t, target, payloadFiles()["invariants.md"])
			if err := os.Chmod(target, 0o000); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(target, 0o644) })
		},
	}}
}

// assertOverlayOwnCopyReport runs the hook once on an overlay that has its own
// copy of invariants.md and checks everything decision-0096 rules about that
// cell. It returns the systemMessage so the caller can require the two plugin
// states to agree.
func assertOverlayOwnCopyReport(t *testing.T, pluginRoot, project, overlayCopy string, shape struct {
	name  string
	why   string
	write func(t *testing.T, target string)
}) string {
	t.Helper()
	raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("a consulted reference is not a delivered payload (decision-0093:1): the overlay still governs and must still be injected: %s", raw)
	}
	context := got.HookSpecificOutput.AdditionalContext

	// THE ELIGIBILITY ASSERTION, and the reason this test no longer needs
	// silence to make it. The overlay owns this address whatever state its copy
	// is in (decision-0094:5, kept by decision-0096:1), so the shipped token
	// must survive rather than being rewritten to the plugin's copy.
	if !strings.Contains(context, invariantsToken) {
		t.Errorf("the overlay's own address must survive whatever state its copy is in — a report is not a licence to substitute (decision-0096:1):\n%s", context)
	}

	if shape.why == "" {
		if got.SystemMessage != "" {
			t.Errorf("the overlay's copy is healthy, so the pointer resolves and there is nothing to report — the over-correction is as bad for a reader as the silence\ngot: %q", got.SystemMessage)
		}
		return got.SystemMessage
	}

	// decision-0096:2. The lead is decision-0095's, byte for byte: both cells
	// blame the overlay, and an operator who meets one and then the other must
	// not have to reconcile two vocabularies for one class of fault.
	if !strings.HasPrefix(got.SystemMessage, vendoredInvariantsReportLead) {
		t.Errorf("the report must blame the overlay: on this branch the plugin's reference/ is not the delivery source and may not stand in for it (decision-0096:2)\nwant the lead: %q\ngot: %q", vendoredInvariantsReportLead, got.SystemMessage)
	}
	// The classification is the fixture's, not a constant: hardcoding one here
	// would leave payloadDefect(overlayInvariants) replaceable by that literal
	// with the suite still green — the mutant review caught on the sibling arm.
	want := invariantsReportLead + overlayCopy + " (it " + shape.why + ")"
	if !strings.Contains(got.SystemMessage, want) {
		t.Errorf("the hook delivered a pointer it had just proved dead and did not name the overlay's own unusable file — TRL-52's defect, surviving on the last arm that never had to answer for it\nwant: %s\ngot:  %q", want, got.SystemMessage)
	}
	if !strings.Contains(got.SystemMessage, invariantsReportClaim) {
		t.Errorf("the report does not say what is actually true of every classification\nwant: %q\ngot: %q", invariantsReportClaim, got.SystemMessage)
	}
	if strings.Contains(got.SystemMessage, invariantsReportFalseClaim) {
		t.Errorf("the report says the file %q, false for an empty one which reads fine (#295)\ngot: %q", invariantsReportFalseClaim, got.SystemMessage)
	}
	// Flattening is refused the same way it is on every other arm: no
	// classification beyond the one this file actually has.
	for _, other := range invariantsClassifications() {
		if other == strings.SplitN(shape.why, " — ", 2)[0] || !strings.Contains(got.SystemMessage, other) {
			continue
		}
		t.Errorf("the report names %q as well, so the reader is told the file is broken in more ways than it is\ngot: %q", other, got.SystemMessage)
	}

	// decision-0096:2's negative half, and it is the whole difference between
	// this report and decision-0095's. There the plugin's copy was an ELIGIBLE
	// substitute that happened to be unavailable, so naming it told the reader
	// whether reinstalling would have helped. Here it is ineligible whatever
	// state it is in, so naming it — or offering the reinstall — sends the
	// reader to a remedy that cannot work.
	if strings.Contains(got.SystemMessage, pluginRoot) {
		t.Errorf("the report names the plugin root, which may not stand in for the overlay whatever state it is in — the reader is pointed at a file that is not the fix (decision-0096:2)\ngot: %q", got.SystemMessage)
	}
	if strings.Contains(strings.ToLower(got.SystemMessage), "reinstall") {
		t.Errorf("the report offers a reinstall, which cannot repair this cell: the overlay is authoritative and the plugin's copy is ineligible, not merely broken (decision-0096:2)\ngot: %q", got.SystemMessage)
	}
	const wantRemedy = "Putting a readable invariants.md at that overlay path"
	if !strings.Contains(got.SystemMessage, wantRemedy) {
		t.Errorf("the report dropped the one repair that works on this arm (decision-0096:2)\nwant: %q\ngot: %q", wantRemedy, got.SystemMessage)
	}
	// The POSITIVE half, and it is pinned because review showed it droppable in
	// silence: deleting this sentence left the whole guard green. The negative
	// assertions above pin what must NOT appear; without this one, nothing pins
	// the sentence that tells the reader WHY no substitute is offered — which is
	// the only place decision-0096:1's ruling reaches a person. The sibling arm
	// sets the precedent for exactly this hazard (both of its remedy clauses are
	// pinned, "because review found the second one droppable in silence").
	// THE WHOLE SENTENCE, not its opening clause. Review planted the mutant that
	// proves the difference: swapping the tail for "when that copy is also broken"
	// asserts the OPPOSITE of decision-0096:2 — it makes the plugin's copy sound
	// merely unavailable rather than ineligible — and it survived a prefix-only
	// assertion, because it keeps the lead, the classification, "yields nothing to
	// read", the remedy, names no pluginRoot and offers no reinstall. The tail is
	// where the ruling actually lives.
	const wantWhyNoSubstitute = "A vendored overlay is authoritative, so the plugin's own copy does not stand in for it whatever state that copy is in"
	if !strings.Contains(got.SystemMessage, wantWhyNoSubstitute) {
		t.Errorf("the report does not say why no substitute is offered, so a reader meeting the sibling report's reinstall advice has no way to tell why it is absent here (decision-0096:1)\nwant: %q\ngot: %q", wantWhyNoSubstitute, got.SystemMessage)
	}
	// decision-0093:1 discharged TO THE READER, and pinned because review found
	// it droppable in silence. Everything else in this guard establishes that a
	// dead pointer is REPORTED; this is the only sentence telling the person who
	// reads the report that the session is still governed -- that the rules
	// arrived, and that a consulted reference is not a delivered payload. Without
	// it the report reads as "your governance is broken", which is precisely the
	// reading decision-0093:1 exists to prevent, and no other assertion notices.
	const wantStillGoverns = "The rules themselves were delivered and govern this session normally"
	if !strings.Contains(got.SystemMessage, wantStillGoverns) {
		t.Errorf("the report does not tell the reader the session is still governed, so a dead consulted reference reads as a governance failure (decision-0093:1)\nwant: %q\ngot: %q", wantStillGoverns, got.SystemMessage)
	}

	// The report costs the injected context NOTHING: it rides systemMessage,
	// outside the MAX_CONTEXT_BYTES bound on `context`. That is the one way a
	// report about a CONSULTED file could fail a session closed
	// (decision-0093:1), and here it holds structurally, so the structure is
	// what is pinned.
	if strings.Contains(context, invariantsReportLead+overlayCopy) {
		t.Errorf("the report was assembled into the injected context; a report about a consulted file must not spend the context budget (decision-0093:1)\n%s", context)
	}
	return got.SystemMessage
}

// TestCodexSaysNothingWhenTheVendoredProseCarriesNoPointer is the second
// condition review put on decision-0096's arm, and it is a guard against the
// report asserting something the hook never checked.
//
// The report states as FACT that "the invariants pointer in the context just
// injected names a file that yields nothing to read". On the vendored branch the
// prose is the PROJECT's own .trellis/internal/trellis.md, and the only thing
// the hook validates about it is its @rules.md placeholder count — nothing
// requires it to carry the invariants pointer at all. A project that edited the
// pointer out would otherwise draw a report about a pointer that was never
// delivered, sending the reader to repair a file nothing names. That is the
// right-diagnosis-wrong-artifact class this whole thread has been closing, so it
// may not be reintroduced by the change that closes its last cell.
//
// The overlay's own copy is zero-byte here, so EVERY other condition for the
// report is satisfied: this fixture isolates the pointer-presence check alone,
// and reverting it turns this test red while the rest of the suite stays green.
func TestCodexSaysNothingWhenTheVendoredProseCarriesNoPointer(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)

	prose := filepath.Join(project, ".trellis", "internal", "trellis.md")
	before := readFileT(t, prose)
	if !strings.Contains(before, invariantsToken) {
		t.Fatal("fixture drift: the vendored prose no longer carries the invariants pointer, so removing it proves nothing")
	}
	writeFileT(t, prose, strings.ReplaceAll(before, invariantsToken, "`the invariants reference`"))
	// Every other precondition for the report holds, so a firing here can only
	// come from the pointer-presence check being absent.
	writeFileT(t, filepath.Join(project, ".trellis", "internal", "invariants.md"), "")

	raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("the overlay still governs and must still be injected (decision-0093:1): %s", raw)
	}
	if strings.Contains(got.HookSpecificOutput.AdditionalContext, invariantsToken) {
		t.Fatal("fixture drift: the delivered context still carries the pointer, so this test is not exercising its subject")
	}
	if got.SystemMessage != "" {
		t.Errorf("the delivered prose names no invariants pointer, so there is no dead pointer to report and this report claims one was injected — a repair for a file nothing points at\ngot: %q", got.SystemMessage)
	}
}

// TestCodexReportsWhenTheVendoredPointerArrivesThroughTheRulesHalf is the other
// half of the pointer-presence check, and it exists because review found the
// check written against only one of the two files it needed to cover.
//
// The delivered context is `trellis.replace("@rules.md", rules)`
// (codex-context.mjs:1455). On the vendored branch BOTH halves are the project's
// own unvalidated files: `.trellis/internal/trellis.md` and
// `.trellis/internal/rules.md`. A guard that asks only whether `trellis` carries
// the pointer therefore goes silent on a project that moved the pointer into its
// rules half — while every substantive precondition for reporting holds and the
// delivered context really does carry a dead pointer. That is the same
// right-diagnosis-wrong-artifact defect the check was added to prevent, missed
// by half.
//
// The mirror mutant is what makes this test necessary rather than decorative:
// widening the production check to `(trellis + rules).includes(...)` is
// unobservable to the rest of the suite, so the check's SCOPE is unpinned in
// both directions without a fixture that puts the pointer in the rules half.
func TestCodexReportsWhenTheVendoredPointerArrivesThroughTheRulesHalf(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)

	// Move the pointer out of the prose half and into the rules half. Inserted
	// at the START of rules.md so the sentinel gate — exactly one sentinel, and
	// the file ending with it — is untouched.
	prose := filepath.Join(project, ".trellis", "internal", "trellis.md")
	proseBody := readFileT(t, prose)
	if !strings.Contains(proseBody, invariantsToken) {
		t.Fatal("fixture drift: the vendored prose no longer carries the invariants pointer, so moving it proves nothing")
	}
	writeFileT(t, prose, strings.ReplaceAll(proseBody, invariantsToken, "`the invariants reference`"))
	rules := filepath.Join(project, ".trellis", "internal", "rules.md")
	writeFileT(t, rules, "Consult "+invariantsToken+" when a rule seems ambiguous.\n\n"+readFileT(t, rules))

	// Every other precondition for the report holds.
	writeFileT(t, filepath.Join(project, ".trellis", "internal", "invariants.md"), "")

	raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("a consulted reference is not a delivered payload (decision-0093:1): the overlay still governs and must still be injected: %s", raw)
	}
	// The premise: the pointer really is delivered, through the rules half.
	if !strings.Contains(got.HookSpecificOutput.AdditionalContext, invariantsToken) {
		t.Fatalf("fixture drift: the delivered context carries no pointer, so this test is not exercising its subject:\n%s", got.HookSpecificOutput.AdditionalContext)
	}
	if !strings.HasPrefix(got.SystemMessage, vendoredInvariantsReportLead) {
		t.Errorf("the delivered context carries a pointer this hook has just proved dead, and the hook said nothing — the pointer-presence check reads only the prose half, while the delivery is trellis.replace(\"@rules.md\", rules)\ngot: %q", got.SystemMessage)
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
// doctrine at codex-context.mjs:1025-1030 ("a missing one is a broken overlay
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
	// became the silent mode switch :1025-1030 forbids.
	overlayRules := readFileT(t, filepath.Join(project, ".trellis", "internal", "rules.md"))
	if first := strings.SplitN(strings.TrimSpace(overlayRules), "\n", 2)[0]; first != "" && !strings.Contains(context, first) {
		t.Errorf("the injected rules no longer come from the overlay — the fallback switched delivery mode, not just the pointer\nwant a line from: %s", first)
	}

	// TRL-71's discrimination control, and the reason it lives HERE: this is
	// the fixture where the fallback actually fires. A hook that reported an
	// incomplete overlay unconditionally would satisfy every assertion in
	// TestCodexDoesNotFallBackOnAnUnusablePluginCopy and still be wrong on the
	// shape that is not broken at all -- the over-correction
	// assertInvariantsReport guards against on the other arm.
	_, got := runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.SystemMessage != "" {
		t.Errorf("the fallback fired onto a healthy plugin copy, so nothing is broken and nothing may be reported\ngot: %q", got.SystemMessage)
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
// plugin-native arm.
//
// TRL-71 is the other half, and it is this test's now rather than a residue it
// hands on. Both addresses ARE dead here: the pointer survives at an address
// the hook has just proved nothing is at, and until this change the hook said
// nothing. #294 scoped its report away from this arm on a reason that is sound
// about WORDING and silent about SILENCE — "on the vendored branch it is not
// the source: a plugin root with no reference/ at all is a legitimate shape
// there" — so what is asserted below is a report that blames the OVERLAY and
// names the plugin copy only as the substitute that was unavailable.
//
// The classification assertions are the same shared table and the same
// assertInvariantsReport as the pair guard, against the PLUGIN path, because
// that is the copy the classification describes. The overlay's own address is
// asserted separately: it is the address the model was actually handed.
// overlayAbsences is how the OVERLAY's own copy can be missing-as-a-file while
// still reaching this arm. Both rows fail existingFile, which is what the
// eligibility check asks; they differ in what payloadDefect then calls them, and
// that difference is the whole reason decision-0095:3 classifies this half.
//
// The directory row exists because review found its absence: with only the
// absent row, `payloadDefect(overlayInvariants)` could be replaced by the string
// literal "is missing" and the entire suite still passed. The shape is not
// exotic — it is the one decision-0095:3 names as the case where a bare
// "restore invariants.md" would be incomplete until the directory is gone.
func overlayAbsences() []struct {
	name string
	make func(t *testing.T, target string)
	why  string
} {
	return []struct {
		name string
		make func(t *testing.T, target string)
		why  string
	}{{
		name: "the overlay simply has no copy",
		make: func(t *testing.T, target string) {},
		why:  "is missing",
	}, {
		name: "a directory sits at the overlay's own path",
		make: func(t *testing.T, target string) {
			if err := os.Mkdir(target, 0o755); err != nil {
				t.Fatal(err)
			}
		},
		why: "is not a readable file — a directory or a device sits at that path",
	}}
}

func TestCodexDoesNotFallBackOnAnUnusablePluginCopy(t *testing.T) {
	for _, overlay := range overlayAbsences() {
		for _, tc := range invariantsFaults() {
			// The healthy row is the control for the test above, not for this
			// one: with a good plugin copy the fallback SHOULD fire. Skipping it
			// here keeps this test a statement about unusable copies only.
			if tc.why == "" {
				continue
			}
			runUnusablePluginCopyCase(t, overlay.name+" / "+tc.name, overlay.make, overlay.why, tc)
		}
	}
}

// runUnusablePluginCopyCase is one cell of the overlay-shape x plugin-fault
// matrix above, lifted out so the two loops stay readable.
func runUnusablePluginCopyCase(t *testing.T, name string, makeOverlay func(*testing.T, string), overlayWhy string, tc invariantsFault) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		pluginRoot := writeDualHostPluginRoot(t)
		project := newGitProject(t)
		writeValidCodexOverlay(t, project)
		overlayCopy := filepath.Join(project, ".trellis", "internal", "invariants.md")
		if _, err := os.Stat(overlayCopy); err == nil {
			t.Fatal("fixture drift: writeValidCodexOverlay now writes invariants.md, so the fallback arm is no longer reached")
		}
		makeOverlay(t, overlayCopy)
		pluginCopy := filepath.Join(pluginRoot, "reference", "invariants.md")
		tc.breakIt(t, pluginCopy)

		raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
		if got.HookSpecificOutput == nil {
			t.Fatalf("a consulted reference is not a delivered payload (decision-0093:1): Codex must still inject the overlay's rules: %s", raw)
		}
		context := got.HookSpecificOutput.AdditionalContext

		// Delivery is untouched: only the pointer was ever in question
		// (decision-0093:1).
		if !strings.Contains(context, rulesLoadedSentinel) {
			t.Fatalf("the overlay's rules must still be delivered whole:\n%s", context)
		}
		if got := pointerIn(t, context); got != ".trellis/internal/invariants.md" {
			t.Errorf("the fallback fired onto a plugin copy that %s — decision-0093:2's second half exists to stop exactly this substitution\nwant the overlay's own address to survive: .trellis/internal/invariants.md\ngot:  %s", tc.why, got)
		}

		// TRL-71. The pointer above survives at an address this hook has
		// just proved nothing is at, and the overlay is what the reader
		// must repair — the plugin's reference/ is not this delivery's
		// source, so blaming the install would name the wrong artifact.
		//
		// BOTH halves are asserted as path-and-classification PAIRS rather
		// than through assertInvariantsReport, and that is forced by the
		// report naming two files. That helper's exclusion arm treats a
		// second classification as flattening, which is right when one file
		// is described and wrong here: two files with two different faults
		// SHOULD carry two classifications. Pinning the pairs is also
		// strictly stronger than the helper's separate contains-checks,
		// which a report that attached the wrong defect to the wrong path
		// would satisfy.
		// decision-0095:2's ruling, asserted HERE rather than left to the
		// strip helper's HasPrefix in five tests about other things. Review
		// showed that swapping this lead for the plugin-native one — the
		// exact advice that ruling rejects — failed only those five, with
		// messages blaming unrelated subjects, and would go entirely
		// unpinned if writeCodexPluginRoot were ever healed.
		if !strings.HasPrefix(got.SystemMessage, vendoredInvariantsReportLead) {
			t.Errorf("the report must blame the overlay, not the install: on this branch the plugin's reference/ is not the delivery source (decision-0095:2)\nwant the lead: %q\ngot: %q", vendoredInvariantsReportLead, got.SystemMessage)
		}
		overlayTarget := filepath.Join(project, ".trellis", "internal", "invariants.md")
		// The overlay's own classification is the fixture's, not a
		// constant: review found that hardcoding "is missing" here left
		// payloadDefect(overlayInvariants) replaceable by that same literal
		// with the whole suite still green.
		wantOverlay := invariantsReportLead + overlayTarget + " (it " + overlayWhy + ")"
		if !strings.Contains(got.SystemMessage, wantOverlay) {
			t.Errorf("the hook kept a pointer it had just proved dead and did not name the overlay's own missing file — TRL-52's defect, surviving on the arm #294 did not change\nwant: %s\ngot:  %q", wantOverlay, got.SystemMessage)
		}
		wantPlugin := invariantsReportLead + pluginCopy + " either (it " + tc.why + ")"
		if !strings.Contains(got.SystemMessage, wantPlugin) {
			t.Errorf("the plugin copy is why no substitute was available, and which fault it is decides whether reinstalling would have helped\nwant: %s\ngot:  %q", wantPlugin, got.SystemMessage)
		}
		// Flattening is still refused: no classification may appear beyond
		// the two these two files actually have.
		for _, why := range invariantsClassifications() {
			if why == strings.SplitN(overlayWhy, " — ", 2)[0] || why == tc.classification() || !strings.Contains(got.SystemMessage, why) {
				continue
			}
			t.Errorf("the report names %q as well, so the reader is told the two files are broken in more ways than they are\ngot: %q", why, got.SystemMessage)
		}
		// The claim both hosts share about a file in this state, and the
		// wording #295 corrected because it is false for an empty one.
		if !strings.Contains(got.SystemMessage, invariantsReportClaim) {
			t.Errorf("the report does not say what is actually true of every classification\nwant: %q\ngot: %q", invariantsReportClaim, got.SystemMessage)
		}
		if strings.Contains(got.SystemMessage, invariantsReportFalseClaim) {
			t.Errorf("the report says the file %q, false for an empty one which reads fine (#295)\ngot: %q", invariantsReportFalseClaim, got.SystemMessage)
		}
		// The remedy is the actionable half of "blame the overlay", and it
		// is pinned because review found it unpinned: swapping it for the
		// plugin-native remedy — the very advice decision-0095:2 rules out,
		// since it sends the reader to reinstall an install that may be
		// fine — passed the whole suite.
		// BOTH clauses, because review found the second one droppable in
		// silence: wantRemedy pinned only the first, and the strip helper
		// needs no more than "is the likely fix." to exist somewhere.
		for _, wantRemedy := range []string{
			"Putting a readable invariants.md at that overlay path",
			"reinstalling the Trellis plugin so its copy can stand in",
		} {
			if !strings.Contains(got.SystemMessage, wantRemedy) {
				t.Errorf("the report dropped half its remedy; both repairs are real and the reader needs both (decision-0095:2)\nwant: %q\ngot: %q", wantRemedy, got.SystemMessage)
			}
		}

		// The report costs the injected context NOTHING: it rides
		// systemMessage, outside the MAX_CONTEXT_BYTES bound on `context`.
		// That is the one way a report about a CONSULTED file could fail a
		// session closed (decision-0093:1), and on this host it holds
		// structurally — so the structure is what is pinned.
		if strings.Contains(context, invariantsReportLead+overlayTarget) {
			t.Errorf("the report was assembled into the injected context; a report about a consulted file must not spend the context budget (decision-0093:1)\n%s", context)
		}

		// The asymmetry is deliberate and is pinned here rather than left
		// as a reading of the other hook: staleness.sh's path A exits at
		// the vendored branch, so Claude never delivers an invariants
		// pointer into a vendored project and has nothing to report about.
		// A vendored row in the pair guard would therefore have to assert
		// Claude delivers NOTHING, contradicting that guard's own delivery
		// assertion — which is why this cell is guarded here instead.
		if out := claudeStdoutFor(t, pluginRoot, project); strings.Contains(out, "read its entry in `") {
			t.Errorf("staleness.sh delivered an invariants pointer into a vendored project; this cell is Codex-only by construction and the pair guard's shape depends on it:\n%s", out)
		}
	})
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

// vendoredInvariantsReportLead opens the report codex-context.mjs makes when a
// vendored overlay has no invariants.md and the plugin's copy cannot stand in
// for it (TRL-71). It names the OVERLAY, because the overlay is what the reader
// must repair; the plugin's copy appears later in the same sentence pair as the
// substitute that was unavailable.
const vendoredInvariantsReportLead = "Trellis warning: this project's vendored overlay has no readable "

// warningsBesideVendoredInvariants returns a Codex systemMessage with that
// report removed, so a test whose subject is something else can still assert
// "and nothing else was said".
//
// It exists because a large family of vendored fixtures is in TRL-71's cell BY
// CONSTRUCTION and always was: writeCodexPluginRoot builds a plugin root
// carrying only .codex-plugin/plugin.json, with no reference/ at all, and
// writeValidCodexOverlay writes no invariants.md. Both addresses are therefore
// dead on every one of them, so the report is CORRECT there — it is simply not
// what those tests are about. Measured when the report was first wired up: 58
// hook invocations across the suite fire it, and exactly five assertions
// noticed. FOUR are `SystemMessage != ""` standing in for "the hook did not fail
// closed"; the fifth, TestCodexHookFalseFloorRowsWarnButSucceed, is an equality
// against the floor warning, which the report displaces by prefixing it. An
// earlier version of this comment called all five silence proxies; review
// measured that false.
//
// THE 58 DOES NOT REPRODUCE, and is flagged rather than re-endorsed. Measured on
// this branch and on main alike, by logging every hook stdout the suite drives:
// 155 invocations on main and 161 here, of which THIS report fires 65 on both.
// The count is pre-existing (it predates this branch and is quoted in TRL-75's
// own title), so it is left as the historical figure it is and corrected here
// rather than silently carried forward. What the number was used to argue is
// unaffected: whatever the exact count, every one of those invocations is in
// this cell by construction and the narrowing below is still the cheaper option.
//
// decision-0096's guard changes that count by ZERO — its ten new invocations all
// carry an overlay invariants.md and so never reach this arm at all, which is the
// same fact stated two paragraphs down.
//
// Narrowing those five is deliberately preferred over healing
// writeCodexPluginRoot. Giving that helper a reference/invariants.md would move
// all 58 onto the decision-0093 fallback path, changing the pointer every one
// of them delivers from a 33-byte token to an absolute temp path — a silent
// change to what they cover, and to the byte profile two of them measure.
//
// What is NOT relaxed, stated precisely because an earlier version of this
// comment was not: a spurious report is caught by
// TestCodexRepointsWhenAVendoredOverlayLacksInvariants for a HEALTHY plugin copy
// (the fallback's own cell), and by
// TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting — named
// TestCodexSaysNothingWhenTheOverlayCarriesItsOwnInvariants until
// decision-0096 — for an overlay whose own copy is HEALTHY. The second of those
// was added because review showed the first does not cover it: mutating the
// reporting arm's condition to `true` survived the whole suite without it.
// assertInvariantsReport's over-correction arm does NOT cover this arm at all —
// it runs only on the pair guard's config-only fixture, where this arm cannot
// fire.
//
// Its four UNUSABLE rows no longer assert silence, and this helper is why that
// costs nothing here: decision-0096's report shares this one's lead and its
// trailing "is the likely fix.", so the strip below removes either. The two can
// never both fire — they sit on mutually exclusive arms — and no fixture
// reached by this helper is in decision-0096's cell at all, since
// writeValidCodexOverlay writes no invariants.md. Measured: wiring that report up
// fired it on 4 hook invocations suite-wide against the PRE-REWRITE fixture, and
// on 8 against the merged one, because decision-0096's guard runs each shape
// twice — against a healthy plugin copy and a missing one. Every firing is inside
// that guard on either count.
func warningsBesideVendoredInvariants(t *testing.T, msg string) string {
	t.Helper()
	if !strings.HasPrefix(msg, vendoredInvariantsReportLead) {
		return msg
	}
	// The report is pushed first and the warnings are joined by one space, so
	// the remainder starts after this report's own final sentence.
	const tail = "is the likely fix."
	i := strings.Index(msg, tail)
	if i < 0 {
		t.Fatalf("the vendored invariants report no longer ends with %q, so this helper would swallow the whole systemMessage instead of just that report\ngot: %q", tail, msg)
	}
	return strings.TrimPrefix(msg[i+len(tail):], " ")
}

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
		//
		// THIS ROW IS ALSO THE ONE THAT SPLITS BY PLATFORM, which is why
		// claudeContextFor reads stdout and stderr separately. bash >= 4.4
		// announces its own NUL-discarding on stderr ("warning: command
		// substitution: ignored null byte in input"); macOS ships bash 3.2 and
		// says nothing. The classification is identical either way — this is
		// the shell narrating the behaviour payload_read depends on — but the
		// helper used to conflate the streams, so the warning landed in front
		// of the JSON and the row failed on Linux CI alone, blaming the hook
		// for output it never wrote to stdout.
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
// `context` alone (:1328) and pushes the same warning onto systemMessage
// afterwards (:1482), so Codex can never lose its context to this warning.
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

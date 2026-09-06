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

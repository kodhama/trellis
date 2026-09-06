package main

// TRL-55. codex-context.mjs resolves its payload down one of two branches, and
// a single directory test decides which (codex-context.mjs:912): a project
// holding `.trellis/internal/` is VENDORED and the overlay is the source;
// a project without one is PLUGIN-NATIVE and the plugin's own `reference/` is.
//
// Measured on 1984862 by instrumenting the branch and running the whole
// package: the suite executed this hook 99 times and took the plugin-native
// branch in 5 of them (5%), spread over 5 tests that each reach it once —
// three invariants-pointer tests (TRL-52/#283, all posture a) plus one
// incidental control in TestCodexHookHonoursGovernedFalse and one posture
// sub-test in TestCodexReconcilesInsteadOfFailingClosed. Every other fixture in
// the suite pairs writeCodexPluginRoot (which writes only
// `.codex-plugin/plugin.json`, never `reference/`) with an overlay, so it runs
// the vendored branch. That is how TRL-52 shipped: the tests were not weak on
// the plugin-native path, they never reached it.
//
// What this file adds is the half of that branch the pointer tests do not
// touch — which FILE the branch reads, and which file it NAMES when that file
// is broken. Both are things only this branch can get wrong, because both are
// read off `sources`, the one value the branch decides.
//
// What it deliberately does NOT do is re-run the vendored suite on this branch.
// Everything downstream of `sources` — row reconciliation, quarantine, floor
// warnings, provenance, budget degradation, the `governed = false` opt-out and
// the whole stdin/cwd/PLUGIN_ROOT input vocabulary — reads `payload.*` and
// `rulesToml` and never consults `vendored` again; the input checks run before
// the branch is even computed. Feeding those the same bytes down the other
// branch would execute the identical instructions for the identical result:
// runtime spent to imply coverage that is not there.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The one line that differs between reference/trellis-a.md and
// reference/trellis-b.md. `comm` on the two files returns exactly this pair, so
// each marker is present in its own posture's prose and absent from the other's
// — which is what makes them usable as evidence of WHICH file was read.
const (
	firmProseMarker     = "**Firmly**"
	adaptiveProseMarker = "**By default**"
)

// removeFileT is the readFileT/writeFileT sibling this file needs: a fault
// injected by DELETING a payload file, checked rather than ignored so a case
// that silently removed nothing cannot pass by asserting the healthy path.
func removeFileT(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

// writePluginNativeProject is writeConfigOnlyProject with the posture dial the
// plugin-native branch actually turns. writeConfigOnlyProject hardcodes
// rules-a.toml, so every fixture built from it computes posture "a" and reads
// reference/trellis-a.md; nothing could reach the "b" arm of
// `reference/trellis-${posture}.md` through it.
//
// Posture is a PLUGIN-NATIVE-ONLY dial: codex-context.mjs computes it for every
// run but only the plugin-native arm of `sources` consumes it, so this helper
// deliberately has no vendored twin. A symmetric helper would imply the dial
// works on both branches; TestCodexVendoredDeliveryIgnoresTheProjectsPosture
// asserts that it does not.
func writePluginNativeProject(t *testing.T, posture string) string {
	t.Helper()
	project := newGitProject(t)
	toml := filepath.Join(project, ".trellis", "rules.toml")
	if err := os.MkdirAll(filepath.Dir(toml), 0o755); err != nil {
		t.Fatal(err)
	}
	body, ok := payloadFiles()["rules-"+posture+".toml"]
	if !ok {
		t.Fatalf("no payload rules-%s.toml; the posture dial names a file that does not ship", posture)
	}
	if err := os.WriteFile(toml, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return project
}

// TestCodexPluginNativeServesAFirmProjectThePluginRootsOwnPayload is the
// positive statement of this branch, which it did not have. It asserts two
// things about ONE run because they are two facts about one delivery, and
// splitting them would mean building the same fixture and executing the same
// instructions twice to look at different bytes of the same output:
//
//   - WHICH prose file the posture selected (the firm half of the dial), and
//   - that the rules body and version stamp delivered are the plugin root's own
//     bytes (ORIGIN, not merely shape).
//
// On the FIRM half: the adaptive half is already run by
// TestCodexReconcilesInsteadOfFailingClosed's "a missing strictness falls back
// to adaptive, as Claude does — posture" sub-test, which reaches this branch
// with a real bundle and asserts the same `**By default**` header. Only the firm
// direction was unpinned, and with the dial pinned on one side a selector stuck
// at "b" served adaptive prose to every firm project with the suite green.
//
// On ORIGIN: the three invariants-pointer tests run this branch but assert only
// the pointer, and TestCodexPluginNativePayloadFaultsNameTheFileActuallyRead
// below proves the three payload paths are STAT'd — neither proves the bytes at
// those paths are the bytes the model receives. The vendored branch has that
// assertion and this one did not: TestCodexHookValidStartupAndLiveRows pins that
// a vendored project must NOT source rule content from the plugin payload;
// nothing stated the converse. The gap is not hypothetical — this hook already
// rewrites one payload value between reading it and delivering it (the
// invariants repoint rewrites `trellis` in place), so "read the file, deliver
// something else" is a shape the file's own structure invites.
//
// Each of the three files is marked distinguishably in a throwaway copy of the
// plugin root, inside the constraints the hook enforces on that file: the prose
// keeps its single @rules.md, the rules body keeps its single trailing
// sentinel, and the version stamp stays a well-formed but different stamp.
func TestCodexPluginNativeServesAFirmProjectThePluginRootsOwnPayload(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := writePluginNativeProject(t, "a")

	const proseMark = "TRL55-PROSE-FROM-PLUGIN-ROOT"
	const rulesMark = "TRL55-RULES-FROM-PLUGIN-ROOT"
	// A valid stamp (`^payload@[0-9a-f]{12}$`) that is not the shipped one, so
	// asserting it proves the delivered stamp tracked this file rather than
	// coincidentally matching the payload the test package already holds.
	const versionMark = "payload@0123456789ab"
	if versionMark == strings.TrimSpace(payloadFiles()["version"]) {
		t.Fatal("fixture drift: the marker stamp equals the shipped stamp, so the assertion below would pass either way")
	}

	ref := func(name string) string { return filepath.Join(pluginRoot, "reference", name) }

	// Ahead of @rules.md, so the placeholder count the hook enforces is untouched.
	prose := readFileT(t, ref("trellis-a.md"))
	marked := strings.Replace(prose, "@rules.md", proseMark+"\n\n@rules.md", 1)
	if marked == prose {
		t.Fatal("fixture drift: reference/trellis-a.md no longer carries @rules.md")
	}
	writeFileT(t, ref("trellis-a.md"), marked)

	// Ahead of the terminal sentinel, so the rules body still ends with exactly
	// one of them and slug derivation is unaffected.
	rules := readFileT(t, ref("rules.md"))
	markedRules := strings.Replace(rules, rulesLoadedSentinel, rulesMark+"\n\n"+rulesLoadedSentinel, 1)
	if markedRules == rules {
		t.Fatal("fixture drift: reference/rules.md no longer carries the loaded sentinel")
	}
	writeFileT(t, ref("rules.md"), markedRules)

	writeFileT(t, ref("version"), versionMark+"\n")

	context := codexContextFor(t, pluginRoot, project)

	// Which file the posture selected. The marker was written into
	// reference/trellis-a.md only, so its presence identifies the file read.
	if !strings.Contains(context, firmProseMarker) {
		t.Errorf("a firm project must be served reference/trellis-a.md:\n%s", context)
	}
	if strings.Contains(context, adaptiveProseMarker) {
		t.Errorf("a firm project was served the adaptive prose — the posture selector is not reading the project's strictness:\n%s", context)
	}

	// That the bytes at the three resolved paths are the bytes delivered.
	for _, want := range []string{proseMark, rulesMark, versionMark} {
		if !strings.Contains(context, want) {
			t.Errorf("the plugin-native branch validated the plugin root's payload but delivered something else — %q is missing from the injected context:\n%s", want, context)
		}
	}
}

// TestCodexVendoredDeliveryIgnoresTheProjectsPosture is the other half of the
// dial, and it is the half that makes `posture` a PLUGIN-NATIVE-ONLY value
// rather than a global one. codex-context.mjs computes posture on every run but
// only the plugin-native arm of `sources` consumes it; the vendored arm spreads
// VENDORED_PAYLOAD, whose prose is the overlay's single trellis.md.
//
// The overlay is authoritative (plugins/trellis/README.md:27), so its own prose
// is delivered whatever the project declares. This is the arm a careless "make
// posture work everywhere" change breaks: the overlay ships one trellis.md and
// has no trellis-{a,b}.md for a posture to select between, so honouring posture
// here can only mean reading a file that is not there — or, as the mutation that
// proves this test bites does, quietly sourcing the prose from the plugin
// instead of the overlay that outranks it.
func TestCodexVendoredDeliveryIgnoresTheProjectsPosture(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := newGitProject(t)
	// writeValidCodexOverlay writes trellis-a.md as the overlay prose and
	// rules-a.toml as the project config. Overwriting only the config makes the
	// two disagree: the project declares adaptive, the overlay carries firm
	// prose. Without that disagreement the fixture could not tell the two
	// branches apart.
	writeValidCodexOverlay(t, project)
	writeFileT(t, filepath.Join(project, ".trellis", "rules.toml"), payloadFiles()["rules-b.toml"])

	context := codexContextFor(t, pluginRoot, project)
	if !strings.Contains(context, firmProseMarker) {
		t.Errorf("the vendored overlay's own trellis.md must be delivered regardless of the project's declared posture:\n%s", context)
	}
	if strings.Contains(context, adaptiveProseMarker) {
		t.Errorf("posture reached the vendored branch and selected prose there; the overlay is authoritative and has no posture variants to choose between:\n%s", context)
	}
}

// TestCodexPluginNativePayloadFaultsNameTheFileActuallyRead is the assertion
// this branch never had. Every `fail()` on a payload fault reports
// `sources.prose` / `sources.rules` / `sources.version`, and those three values
// ARE the branch — `.trellis/internal/*` when vendored,
// `reference/trellis-${posture}.md` and `reference/{rules.md,version}` when
// plugin-native. The hook's own comment records this going wrong once: the
// labels were hardcoded to the vendored paths, "so on the plugin-native path it
// reported a failure against a file that was never read". The fix shipped
// unpinned — the whole vocabulary is asserted only through
// TestCodexHookFailureVocabularyAndIsolation, whose fixture is vendored, so a
// regression to the hardcoded path is invisible to this suite.
//
// One case per REACHABLE `fail(sources.*)` call site rather than a re-run of
// that test's table: each site is a separate place the wrong path can be
// written, and the classes themselves are not branch-sensitive. The vendored
// half of the vocabulary is deliberately not repeated here; it is pinned
// literally next door, and asserting it twice would state one property in two
// places that could then drift apart.
//
// Each case gets its own throwaway plugin root, because breaking a payload file
// to see how the hook names it is not something to do to a shared fixture.
func TestCodexPluginNativePayloadFaultsNameTheFileActuallyRead(t *testing.T) {
	for _, tc := range []struct {
		name    string
		posture string
		// break mutates the plugin root's reference/ into the fault under test.
		breakIt func(t *testing.T, ref func(string) string)
		label   string
		class   string
	}{{
		name:    "a missing prose file names the posture prose, not the overlay's trellis.md",
		posture: "a",
		breakIt: func(t *testing.T, ref func(string) string) { removeFileT(t, ref("trellis-a.md")) },
		label:   "reference/trellis-a.md",
		class:   "missing-file",
	}, {
		name:    "prose with no @rules.md import names the posture prose",
		posture: "a",
		breakIt: func(t *testing.T, ref func(string) string) {
			writeFileT(t, ref("trellis-a.md"), "no import\n")
		},
		label: "reference/trellis-a.md",
		class: "invalid-placeholder-count",
	}, {
		name:    "an empty rules payload names reference/rules.md",
		posture: "a",
		breakIt: func(t *testing.T, ref func(string) string) { writeFileT(t, ref("rules.md"), "") },
		label:   "reference/rules.md",
		class:   "empty-prose",
	}, {
		name:    "a rules payload with no sentinel names reference/rules.md",
		posture: "a",
		breakIt: func(t *testing.T, ref func(string) string) {
			writeFileT(t, ref("rules.md"), "no sentinel\n")
		},
		label: "reference/rules.md",
		class: "invalid-rules",
	}, {
		name:    "a missing version stamp names reference/version",
		posture: "a",
		breakIt: func(t *testing.T, ref func(string) string) { removeFileT(t, ref("version")) },
		label:   "reference/version",
		class:   "missing-file",
	}, {
		name:    "a malformed version stamp names reference/version",
		posture: "a",
		breakIt: func(t *testing.T, ref func(string) string) {
			writeFileT(t, ref("version"), "plugin@abcdef123456\n")
		},
		label: "reference/version",
		class: "invalid-version",
	}, {
		// The label is templated on the posture, so an adaptive project's broken
		// prose must name trellis-b.md. Without this case a label frozen at
		// trellis-a.md passes every other case in this table.
		name:    "an adaptive project's broken prose names trellis-b.md, not trellis-a.md",
		posture: "b",
		breakIt: func(t *testing.T, ref func(string) string) { removeFileT(t, ref("trellis-b.md")) },
		label:   "reference/trellis-b.md",
		class:   "missing-file",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			pluginRoot := writeDualHostPluginRoot(t)
			project := writePluginNativeProject(t, tc.posture)
			tc.breakIt(t, func(name string) string {
				return filepath.Join(pluginRoot, "reference", name)
			})

			raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
			want := fmt.Sprintf(`{"systemMessage":"Trellis hook did not load rules: %s: %s. The AGENTS.md bootstrap must attempt the installed overlay."}`, tc.label, tc.class)
			if raw != want {
				t.Errorf("the plugin-native failure names the wrong file\n got: %s\nwant: %s", raw, want)
			}
			if got.HookSpecificOutput != nil {
				t.Error("a payload fault must not also inject context")
			}
		})
	}
}

package main

// TRL-55. codex-context.mjs resolves its payload down one of two branches, and
// a single directory test decides which (the `vendored` const): a project
// holding `.trellis/internal/` is VENDORED and the overlay is the source; a
// project without one is PLUGIN-NATIVE and the plugin's own `reference/` is.
//
// Measured on 1984862 by instrumenting the hook and running the whole package:
// it was invoked 110 times, of which 99 got as far as this branch decision —
// the other 11 exit earlier, on a bad stdin/event/cwd, on no project root, or
// on `governed = false`, all of which are settled before `vendored` is
// computed. Of those 99, the plugin-native branch was taken 5 times (5%). Every
// other fixture that reached the branch wrote an overlay, so it went vendored
// whatever plugin root it was handed. That is how TRL-52 shipped: the tests were
// not weak on the plugin-native path, they never reached it. (TRL-97's parity
// table, TestBothHostsClassifyRulesRowsIdentically, now runs this branch on
// every row it has.)
//
// What this file adds is the half of that branch the pointer tests do not
// touch — which FILE the branch reads, and which file it NAMES when that file
// is broken. Both are things only this branch can get wrong, because both are
// read off `sources`, the one value the branch decides.
//
// What it deliberately does NOT do is re-run the vendored suite on this branch.
// Everything downstream of `sources` — row classification, the warnings, the
// `governed = false` opt-out and the whole stdin/cwd/PLUGIN_ROOT input
// vocabulary — reads `payload.*` and `rulesToml` and never consults `vendored`
// again; the input checks run before the branch is even computed.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The one line that tells the shipped header from the header a vendored overlay
// froze before TRL-97. The shipped reference/trellis.md carries the By default
// sentence; frozenFirmOverlayHeader carries the Firmly one. Each marker is in its
// own file and absent from the other, which is what makes them evidence of
// WHICH file was read.
const (
	frozenOverlayProseMarker = "**Firmly**"
	shippedProseMarker       = "**By default**"
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

// TestCodexPluginNativeServesThePluginRootsOwnPayload is the positive statement
// of this branch. It asserts two things about ONE run because they are two facts
// about one delivery:
//
//   - WHICH prose file the branch read: reference/trellis.md, the one header
//     TRL-97 ships, whatever `strictness` the project declares, and
//   - that the rules body and version stamp delivered are the plugin root's own
//     bytes (ORIGIN, not merely shape).
//
// The project declares `strictness = 'firm'`, in the literal string form that
// once selected the retired firm header, so a selector that came back would
// have something to select.
//
// Each of the three files is marked distinguishably in a throwaway copy of the
// plugin root, inside the constraints the hook enforces on that file: the prose
// keeps its single @rules.md, the rules body keeps its single trailing
// sentinel, and the version stamp stays a well-formed but different stamp.
func TestCodexPluginNativeServesThePluginRootsOwnPayload(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := writeConfigOnlyProject(t)
	writeFileT(t, filepath.Join(project, ".trellis", "rules.toml"), "strictness = 'firm'\n"+configOnlyProjectRules)

	const proseMark = "TRL55-PROSE-FROM-PLUGIN-ROOT"
	const rulesMark = "TRL55-RULES-FROM-PLUGIN-ROOT"
	// A valid stamp (`^payload@[0-9a-f]{12}$`) that is not the shipped one, so
	// asserting it proves the delivered stamp tracked this file rather than
	// coincidentally matching the payload the test package already holds.
	const versionMark = "payload@0123456789ab"
	if versionMark == strings.TrimSpace(payloadFile(t, "version")) {
		t.Fatal("fixture drift: the marker stamp equals the shipped stamp, so the assertion below would pass either way")
	}

	ref := func(name string) string { return filepath.Join(pluginRoot, "reference", name) }

	// Ahead of @rules.md, so the placeholder count the hook enforces is untouched.
	prose := readFileT(t, ref("trellis.md"))
	marked := strings.Replace(prose, "@rules.md", proseMark+"\n\n@rules.md", 1)
	if marked == prose {
		t.Fatal("fixture drift: reference/trellis.md no longer carries @rules.md")
	}
	writeFileT(t, ref("trellis.md"), marked)

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

	if !strings.Contains(context, shippedProseMarker) {
		t.Errorf("the plugin-native branch must serve reference/trellis.md, the one header that ships:\n%s", context)
	}
	if strings.Contains(context, frozenOverlayProseMarker) {
		t.Errorf("a project declaring strictness = 'firm' was served the Firmly sentence; strictness selects nothing (TRL-97):\n%s", context)
	}

	// That the bytes at the three resolved paths are the bytes delivered.
	for _, want := range []string{proseMark, rulesMark, versionMark} {
		if !strings.Contains(context, want) {
			t.Errorf("the plugin-native branch validated the plugin root's payload but delivered something else — %q is missing from the injected context:\n%s", want, context)
		}
	}

	// The rules body, WHOLE and verbatim. A single marker only proves the file
	// was touched: a defect that read reference/rules.md and delivered it with
	// any other line rewritten keeps the marker and passes the loop above. The
	// hook splices this payload in one piece and unmodified
	// (`trellis.replace("@rules.md", rules)`), so the whole body is a legitimate
	// thing to demand, and demanding it is what makes this ORIGIN rather than
	// evidence of contact.
	if !strings.Contains(context, markedRules) {
		t.Errorf("the delivered rules body is not reference/rules.md's bytes — the hook validated one thing and delivered another:\n%s", context)
	}
}

// TestCodexVendoredProseOutranksThePlugins is the other half of the branch: a
// vendored overlay's own trellis.md is delivered, not the plugin's
// reference/trellis.md. The overlay is authoritative (plugins/trellis/README.md:27),
// so its frozen prose governs even where the plugin ships a newer header.
//
// The overlay here carries the header vendored overlays froze before TRL-97,
// and the plugin root carries the shipped one, so the two branches deliver
// different posture sentences and the fixture can tell them apart. The mutation
// this bites is the careless "serve the one shipped header everywhere" change,
// which quietly sources the prose from the plugin instead of the overlay that
// outranks it.
func TestCodexVendoredProseOutranksThePlugins(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	overlay := readFileT(t, filepath.Join(project, ".trellis", "internal", "trellis.md"))
	if !strings.Contains(overlay, frozenOverlayProseMarker) || strings.Contains(payloadFile(t, "trellis.md"), frozenOverlayProseMarker) {
		t.Fatal("fixture drift: the overlay's frozen header and the shipped header no longer differ by the posture sentence, so this test cannot tell the branches apart")
	}

	context := codexContextFor(t, pluginRoot, project)
	if !strings.Contains(context, frozenOverlayProseMarker) {
		t.Errorf("the vendored overlay's own trellis.md must be delivered:\n%s", context)
	}
	if strings.Contains(context, shippedProseMarker) {
		t.Errorf("the plugin's reference/trellis.md reached a vendored project; the overlay is authoritative:\n%s", context)
	}
}

// TestCodexPluginNativePayloadFaultsNameTheFileActuallyRead is the assertion
// this branch never had. Every `fail()` on a payload fault reports
// `sources.prose` / `sources.rules` / `sources.version`, and those three values
// ARE the branch — `.trellis/internal/*` when vendored, `reference/trellis.md`
// and `reference/{rules.md,version}` when plugin-native. The hook's own comment
// records this going wrong once: the labels were hardcoded to the vendored
// paths, "so on the plugin-native path it reported a failure against a file that
// was never read".
//
// One case per REACHABLE `fail(sources.*)` call site: the payload read loop
// (exercised once per role, since the label it reports is `sources[key]` and the
// role is what varies), invalid-placeholder-count, invalid-version,
// invalid-rules and no-slugs-in-payload. The two `empty-prose` post-checks are
// the second lock on a door readRequired already shut; a zero-byte file is
// refused at the read and reported from the loop, so the `empty-prose` CLASS is
// covered there.
//
// The vendored half of the vocabulary is deliberately not repeated here; it is
// pinned literally next door.
//
// Each case gets its own throwaway plugin root, because breaking a payload file
// to see how the hook names it is not something to do to a shared fixture.
func TestCodexPluginNativePayloadFaultsNameTheFileActuallyRead(t *testing.T) {
	for _, tc := range []struct {
		name string
		// breakIt mutates the plugin root's reference/ into the fault under test.
		breakIt func(t *testing.T, ref func(string) string)
		label   string
		class   string
	}{{
		name:    "a missing prose file names reference/trellis.md, not the overlay's trellis.md",
		breakIt: func(t *testing.T, ref func(string) string) { removeFileT(t, ref("trellis.md")) },
		label:   "reference/trellis.md",
		class:   "missing-file",
	}, {
		name: "prose with no @rules.md import names reference/trellis.md",
		breakIt: func(t *testing.T, ref func(string) string) {
			writeFileT(t, ref("trellis.md"), "no import\n")
		},
		label: "reference/trellis.md",
		class: "invalid-placeholder-count",
	}, {
		name:    "an empty rules payload names reference/rules.md",
		breakIt: func(t *testing.T, ref func(string) string) { writeFileT(t, ref("rules.md"), "") },
		label:   "reference/rules.md",
		class:   "empty-prose",
	}, {
		name: "a rules payload with no sentinel names reference/rules.md",
		breakIt: func(t *testing.T, ref func(string) string) {
			writeFileT(t, ref("rules.md"), "no sentinel\n")
		},
		label: "reference/rules.md",
		class: "invalid-rules",
	}, {
		name:    "a missing version stamp names reference/version",
		breakIt: func(t *testing.T, ref func(string) string) { removeFileT(t, ref("version")) },
		label:   "reference/version",
		class:   "missing-file",
	}, {
		name: "a malformed version stamp names reference/version",
		breakIt: func(t *testing.T, ref func(string) string) {
			writeFileT(t, ref("version"), "plugin@abcdef123456\n")
		},
		label: "reference/version",
		class: "invalid-version",
	}, {
		// The site furthest from the read: a rules payload that is well-formed
		// enough to pass the sentinel gate but carries no backticked rule tags, so
		// the derived slug set comes back empty. It is the easiest to leave
		// hardcoded, because everything around it concerns the PROJECT's
		// rules.toml rather than the payload. TestCodexRejectsAnEmptyDerivedSlugSet
		// pins this class on the vendored branch with the overlay path written out
		// literally, so it cannot see this label move.
		name: "a rules payload with no rule tags names reference/rules.md",
		breakIt: func(t *testing.T, ref func(string) string) {
			writeFileT(t, ref("rules.md"),
				"Prose with no backticked rule tags, so no slug is derived.\n\n"+rulesLoadedSentinel+"\n")
		},
		label: "reference/rules.md",
		class: "no-slugs-in-payload",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			pluginRoot := writeDualHostPluginRoot(t)
			project := writeConfigOnlyProject(t)
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

// TestCodexReportsAPluginPayloadMissingItsInvariantsCopy is TRL-69: the fourth
// payload file's fault behaviour, which
// TestCodexPluginNativePayloadFaultsNameTheFileActuallyRead above has no case
// for because `invariants.md` is not READ by this hook at all. It is pointed
// at — the plugin-native arm rewrites the shipped token to the plugin's own
// copy (TRL-52, decision-0065:106-111) — and until this test existed nothing
// observed what happens when that copy is not there.
//
// Three cases, and each pins something the other two cannot:
//
//   - The pointer MOVES even when its target is missing. This is the assertion
//     that fails the symmetric guard — the obvious fix, and the wrong one,
//     because the raw token names `.trellis/internal/` in the mode defined by
//     not having it. Asserted in every case, so no case can drift off it.
//   - A COMPLETE payload warns about nothing. Without it, "warn on a missing
//     copy" reads as "warn always" and an unconditional message passes.
//   - Both warnings at once JOIN rather than clobber. The assembly collects
//     into an array precisely because independent conditions can hold
//     together: here the missing copy and a floor row set false in the
//     project's file (TRL-97 folded that warning into the rules.toml warnings).
//     An implementation that went back to assigning response.systemMessage
//     directly would drop one message and pass the whole suite.
//
// Delivery survives all three: the context is asserted non-nil in every case,
// because invariants.md is consulted on demand and decision-0093:1 rules that
// failing a session closed over such a file trades a dead pointer for no
// governance at all.
func TestCodexReportsAPluginPayloadMissingItsInvariantsCopy(t *testing.T) {
	floorWarning := warnFloorRow(2, "floor-intent-gate")
	for _, tc := range []struct {
		name string
		// removeInvariants makes the plugin payload the broken one: three valid
		// delivered files and no fourth, consulted one.
		removeInvariants bool
		// falseFloor sets floor-intent-gate active = false in the PROJECT's
		// rules.toml, the other condition that warns on a successful delivery.
		falseFloor bool
	}{{
		name: "a complete plugin payload warns about nothing",
	}, {
		name:             "a plugin payload with no invariants.md is reported, and the rules still land",
		removeInvariants: true,
	}, {
		name:             "both warnings fire at once and neither is dropped",
		removeInvariants: true,
		falseFloor:       true,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			pluginRoot := writeDualHostPluginRoot(t)
			project := writeConfigOnlyProject(t)
			pointer := filepath.Join(pluginRoot, "reference", "invariants.md")
			if tc.removeInvariants {
				removeFileT(t, pointer)
			}
			if tc.falseFloor {
				writeFileT(t, filepath.Join(project, ".trellis", "rules.toml"), "[rules]\nfloor-intent-gate = { active = false }\n")
			}

			raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
			if got.HookSpecificOutput == nil {
				t.Fatalf("a consulted reference is not a delivered payload (decision-0093:1): the rules must still be injected: %s", raw)
			}
			context := got.HookSpecificOutput.AdditionalContext

			// The pointer moves either way. This is the assertion that makes the
			// symmetric guard fail here rather than pass quietly.
			if strings.Contains(context, invariantsToken) {
				t.Errorf("the raw token survived; this mode has no .trellis/internal/ for it to name, so a missing plugin copy must not degrade the pointer back to it:\n%s", context)
			}
			if got := pointerIn(t, context); got != pointer {
				t.Errorf("the pointer names the wrong file\nwant: %s\ngot:  %s", pointer, got)
			}

			if !tc.removeInvariants {
				if got.SystemMessage != "" {
					t.Errorf("a plugin payload with every file present must warn about nothing: %q", got.SystemMessage)
				}
				return
			}
			// Naming the absolute path is the point of the message: it is what
			// tells the operator WHICH install is broken, and it is the only part
			// of that wording worth pinning.
			if !strings.Contains(got.SystemMessage, pointer) {
				t.Errorf("a plugin payload missing its invariants copy must be reported, naming the file that is not there\ngot: %q", got.SystemMessage)
			}
			if !tc.falseFloor {
				return
			}
			if !strings.Contains(got.SystemMessage, floorWarning) {
				t.Errorf("the floor warning was dropped: a second writer to systemMessage overwrote the first instead of joining\ngot: %q", got.SystemMessage)
			}
			// Order and separator, not merely co-presence: the payload warning
			// reports a broken install and leads, and the join is one space after
			// the preceding sentence's full stop.
			if !strings.Contains(got.SystemMessage, ". "+floorWarning) {
				t.Errorf("the two warnings are not joined by a single space after the payload warning\ngot: %q", got.SystemMessage)
			}
		})
	}
}

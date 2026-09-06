package main

// TRL-55. codex-context.mjs resolves its payload down one of two branches, and
// a single directory test decides which (codex-context.mjs:1030): a project
// holding `.trellis/internal/` is VENDORED and the overlay is the source;
// a project without one is PLUGIN-NATIVE and the plugin's own `reference/` is.
//
// Measured on 1984862 by instrumenting the hook and running the whole package:
// it was invoked 110 times, of which 99 got as far as this branch decision —
// the other 11 exit earlier, on a bad stdin/event/cwd, on no project root, or
// on `governed = false`, all of which are settled before `vendored` is
// computed. Of those 99, the plugin-native branch was taken 5 times (5%),
// spread over 5 tests that each reach it once: three invariants-pointer tests
// (TRL-52/#283, all posture a) plus one incidental control in
// TestCodexHookHonoursGovernedFalse and one posture sub-test in
// TestCodexReconcilesInsteadOfFailingClosed. Every other fixture that reaches
// the branch writes an overlay, so it goes vendored whatever plugin root it is
// handed — most pair writeCodexPluginRoot, which writes only
// `.codex-plugin/plugin.json` and no `reference/` at all, and two vendored
// invariants tests pair writeDualHostPluginRoot, whose `reference/` is real but
// goes unread because the overlay outranks it. That is how TRL-52 shipped: the
// tests were not weak on the plugin-native path, they never reached it.
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
// those paths are the bytes the model receives. Nor does the nearest thing on
// the vendored branch: TestCodexHookValidStartupAndLiveRows asserts that one
// hardcoded relative literal, "../plugins/trellis/reference", never appears in a
// vendored project's context — a check on a leaked path string rather than on
// where the delivered bytes came from, and one its own fixture could not fail
// anyway, since that fixture's plugin root is a t.TempDir() whose path bears no
// resemblance to the literal.
// So no test on either branch pinned delivery to origin, and this one does it
// where the branch decides the origin. The gap is not hypothetical — this hook
// already rewrites one payload value between reading it and delivering it (the
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

// TestCodexFirmPostureIsSelectedFromEitherTomlStringForm pins the OTHER half of
// the posture regex on the only branch that consumes posture.
//
// codex-context.mjs:1034 matches `strictness = "firm"` or `strictness = 'firm'`,
// and its own comment records why both are there: "matching only the basic form
// served a firm project the adaptive posture without saying so." That is a
// shipped defect with a recorded fix — and the fix was unpinned in effect.
// codex_hook_test.go's strict-schema test does feed the literal form, but on a
// VENDORED fixture, where posture is computed and then discarded, and it asserts
// only that the hook does not fail. Nothing observed which prose the literal
// form selected, so deleting `|'firm'` from the regex left the whole suite
// green while every literal-quoted firm project silently got adaptive prose.
//
// Separate from the test above rather than a case inside it: that one builds a
// marked plugin root to prove ORIGIN, and none of that machinery is needed to
// ask which of two files a quoting form selects.
func TestCodexFirmPostureIsSelectedFromEitherTomlStringForm(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	project := writePluginNativeProject(t, "a")

	rules := filepath.Join(project, ".trellis", "rules.toml")
	basic := readFileT(t, rules)
	literal := strings.Replace(basic, `strictness  = "firm"`, `strictness  = 'firm'`, 1)
	if literal == basic {
		t.Fatal("fixture drift: rules-a.toml no longer carries a basic-string strictness to convert")
	}
	writeFileT(t, rules, literal)

	context := codexContextFor(t, pluginRoot, project)
	if !strings.Contains(context, firmProseMarker) {
		t.Errorf("a firm project declared with a TOML literal string must still select reference/trellis-a.md:\n%s", context)
	}
	if strings.Contains(context, adaptiveProseMarker) {
		t.Errorf("the literal-quoted firm posture fell through to adaptive — the regex's `'firm'` arm is not doing anything:\n%s", context)
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
	config := filepath.Join(project, ".trellis", "rules.toml")
	// Asserted rather than assumed, in the idiom
	// TestCodexRepointsWhenAVendoredOverlayLacksInvariants already uses. If the
	// helper ever seeded rules-b.toml itself, the overwrite below would be a
	// no-op, the project and the overlay would agree, and this test would go on
	// passing while proving nothing — the overlay's prose is firm either way.
	if readFileT(t, config) == payloadFiles()["rules-b.toml"] {
		t.Fatal("fixture drift: writeValidCodexOverlay now seeds rules-b.toml, so this fixture no longer makes the project's posture and the overlay's prose disagree")
	}
	writeFileT(t, config, payloadFiles()["rules-b.toml"])

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
// written, and the classes themselves are not branch-sensitive. The hook has
// seven sites whose label comes from `sources` — six match a grep for
// `fail(sources.`, and the seventh is the read loop at :944, which reaches it
// through `sources[key]`. Five are reachable and each has a case here: the payload
// read loop (:944, exercised once per role, since the label it reports is
// `sources[key]` and the role is what varies), :980
// invalid-placeholder-count, :970 invalid-version, :1060 invalid-rules and
// :1093 no-slugs-in-payload. The other two — :962 and :966, both
// `empty-prose` — are the second lock on a door readRequired already shut,
// as the hook's own comment above them says; a zero-byte file is refused at
// the read and reported from :944, so those two cannot be reached to be
// asserted. The `empty-prose` CLASS is still covered, via the read loop.
//
// The vendored half of the vocabulary is deliberately not repeated here; it is
// pinned literally next door, and asserting it twice would state one property in
// two places that could then drift apart.
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
		// The seventh site, and the one furthest from the read: a rules payload
		// that is well-formed enough to pass the sentinel gate but carries no
		// backticked rule tags, so the derived slug set comes back empty. It is
		// the last `fail(sources.*)` in the file (codex-context.mjs:1093) and the
		// easiest to leave hardcoded, because everything around it reports the
		// PROJECT's rules.toml rather than the payload.
		// TestCodexRejectsAnEmptyDerivedSlugSet pins this class on the vendored
		// branch with the overlay path written out literally, so it cannot see
		// this label move.
		name:    "a rules payload with no rule tags names reference/rules.md",
		posture: "a",
		breakIt: func(t *testing.T, ref func(string) string) {
			writeFileT(t, ref("rules.md"),
				"Prose with no backticked rule tags, so no slug is derived.\n\n"+rulesLoadedSentinel+"\n")
		},
		label: "reference/rules.md",
		class: "no-slugs-in-payload",
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

// TestCodexReportsAPluginPayloadMissingItsInvariantsCopy is TRL-69: the fourth
// payload file's fault behaviour, which
// TestCodexPluginNativePayloadFaultsNameTheFileActuallyRead above has no case
// for because `invariants.md` is not READ by this hook at all. It is pointed
// at — the plugin-native arm rewrites the shipped token to the plugin's own
// copy (TRL-52, decision-0065:106-111) — and until this test existed nothing
// observed what happens when that copy is not there.
//
// The argument for reporting rather than guarding is at the repoint itself
// (codex-context.mjs, the TRL-69 comment) and is not restated here; what
// belongs here is what the cases pin. Three of them, and each pins something
// the other two cannot:
//
//   - The pointer MOVES even when its target is missing. This is the assertion
//     that fails the symmetric guard — the obvious fix, and the wrong one,
//     because the raw token names `.trellis/internal/` in the mode defined by
//     not having it. Asserted in every case, so no case can drift off it.
//   - A COMPLETE payload warns about nothing. Without it, "warn on a missing
//     copy" reads as "warn always" and an unconditional message passes.
//   - Both warnings at once JOIN rather than clobber. The assembly collects
//     into an array precisely because two independent conditions can hold
//     together, and nothing else in the package makes them: the floor-row test
//     is vendored, where this warning cannot fire, and the two cases above
//     carry no false floor row. An implementation that went back to assigning
//     response.systemMessage directly would drop one message and pass the
//     whole suite.
//
// Delivery survives all three: the context is asserted non-nil in every case,
// because invariants.md is consulted on demand and decision-0093:1 rules that
// failing a session closed over such a file trades a dead pointer for no
// governance at all.
func TestCodexReportsAPluginPayloadMissingItsInvariantsCopy(t *testing.T) {
	// The floor warning as codex-context.mjs composes it, matched from its own
	// start so the separator assertion below can anchor on it.
	const floorWarning = "Trellis warning: floor rows set active = false are overridden-by-floor and remain active: floor-intent-gate."
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
			project := writePluginNativeProject(t, "a")
			pointer := filepath.Join(pluginRoot, "reference", "invariants.md")
			if tc.removeInvariants {
				removeFileT(t, pointer)
			}
			if tc.falseFloor {
				config := filepath.Join(project, ".trellis", "rules.toml")
				rows := readFileT(t, config)
				disabled := setRuleActive(t, rows, "floor-intent-gate", false)
				if disabled == rows {
					t.Fatal("fixture drift: setting floor-intent-gate inactive changed nothing, so the floor warning would not fire")
				}
				writeFileT(t, config, disabled)
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

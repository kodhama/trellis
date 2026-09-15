package main

// TRL-97 U2. Both hooks read the same .trellis/rules.toml on the plugin-native
// path, and R12 requires them to reach the same answer: the same rules switched
// off, the same file segment, the same warnings, in the same words. The two
// classifiers share no line of code (awk in staleness.sh, JS in
// codex-context.mjs), so the only thing that can be pinned is what each hands
// the model. This table is that pin (decision-0028, a guard per pair).
//
// Every row runs each hook on its own fresh project against one dual-host
// plugin root, asserts that host's delivery against the golden text below, and
// then compares the two hosts with each other. Subtests are named
// <row>/claude, <row>/codex and <row>/parity, so one host can be run alone.
//
// The golden text lives in this file and nowhere else in the test suite. A test
// that read the hook's own templates could not catch the templates changing.
//
// The maintainer's answers to the five open questions each sit behind one named
// place here:
//
//   - any false row wins, and a repeated row draws no warning (the rows below)
//   - five named warnings in total, real-rule attempts first (warnMore)
//   - one bound on the file plus its warnings (rulesShowMaxBytes,
//     expectedTooLargeLine)
//   - a symbolic link is classified and never shown (expectedSymlinkLine,
//     warnSymlinkCounted)
//
// Out of scope on purpose: CR-only line endings and invalid UTF-8, the two
// divergences TRL-97's follow-up issue records, and every shape where no hook
// classifies the file (the stand-down paths, KTD7).

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// rulesShowMaxBytes is B, the bound both hooks share (codex-context.mjs
// RULES_ECHO_MAX_BYTES, staleness.sh rules_echo_max). The file is echoed
// verbatim only when its bytes plus the bytes of its rendered warning block fit
// B; otherwise the too-large line stands in for it. Measured so the largest
// section either branch can build leaves codexContextMargin under
// codexContextCap (the two "largest" rows below).
const rulesShowMaxBytes = 1800

// rulesReadMaxBytes is the read bound both hooks put on the project file
// (codex-context.mjs MAX_PROJECT_CONFIG_BYTES, staleness.sh rules_file_max): a
// larger file is refused rather than read.
const rulesReadMaxBytes = 1024 * 1024

// codexContextCap is codex-context.mjs's MAX_CONTEXT_BYTES, written out rather
// than read from the hook so a changed cap is a visible edit here.
const codexContextCap = 9500

// codexContextMargin is the headroom the largest section must leave under
// codexContextCap on this suite's own plugin root, for an install whose plugin
// root path, repointed into the prose, is longer.
const codexContextMargin = 250

const activationHeading = "## Project rule activation"

// expectedActivationSentence is the second framing line (KTD1). It names the
// rules the file switches off in the order rules.md lists them.
func expectedActivationSentence(off []string) string {
	if len(off) == 0 {
		return "The project file .trellis/rules.toml switches no rule off, so every rule above applies."
	}
	return "The project file .trellis/rules.toml switches these rules off: " + strings.Join(off, ", ") + ". Every other rule above applies."
}

// expectedTooLargeLine stands in for a file whose bytes plus its warning block
// exceed B.
func expectedTooLargeLine() string {
	return "The project file and its warnings are too large to show here; the sentence above names every rule it switches off."
}

// expectedSymlinkLine stands in for a file reached through a symbolic link, at
// any size.
func expectedSymlinkLine() string {
	return "The project file is a symbolic link, so its contents are not shown here; the sentence above names every rule it switches off."
}

// The warning texts (KTD4). Each names the line and the slug, the key, or only
// the entry type, and never quotes the raw line.
func warnLine(line int, rest string) string {
	return fmt.Sprintf("Trellis warning: line %d of .trellis/rules.toml %s", line, rest)
}

func warnUnknownSlug(line int, slug string) string {
	return warnLine(line, "sets "+slug+" to active = false, but this plugin ships no rule by that name, so the row is ignored; another plugin version may ship it.")
}

func warnFloorRow(line int, slug string) string {
	return warnLine(line, "sets the floor rule "+slug+" to active = false, but floor rules cannot be switched off, so the row is ignored and the rule applies.")
}

func warnRowAboveRules(line int, slug string) string {
	return warnLine(line, "is a row for "+slug+" above the [rules] table, so it is ignored; rows count only under [rules].")
}

func warnRulesAgain(line int) string {
	return warnLine(line, "opens [rules] a second time; the rows under it still count.")
}

func warnOtherTable(line int) string {
	return warnLine(line, "opens a table other than [rules], so every line under it is ignored.")
}

func warnGoverned(line int) string {
	return warnLine(line, "sets governed but does not opt out; only one governed = false line, above every table, opts a project out, so this project stays governed.")
}

func warnUnknownKey(line int, key string) string {
	return warnLine(line, "sets "+key+", which is not a key this file defines, so it is ignored and switches no rule off; a row under [rules] such as slug = { active = false } switches a rule off.")
}

func warnTopLevelLine(line int) string {
	return warnLine(line, "is not an entry this file defines, so it is ignored and switches no rule off.")
}

// warnMalformedRow names the bare key before `=` when the line has one.
func warnMalformedRow(line int, key string) string {
	if key == "" {
		return warnLine(line, "is a malformed row, so it is ignored and switches no rule off.")
	}
	return warnLine(line, "is a malformed row naming "+key+", so it is ignored and switches no rule off; any rule it names still applies.")
}

func warnNulByte() string {
	return "Trellis warning: .trellis/rules.toml contains a NUL byte, so none of its rows take effect and it is not shown; every rule applies."
}

// warnMore is the count line after the five named warnings: n entries not
// named, m of which try to switch a rule off.
func warnMore(n, m int) string {
	const lead = "Trellis warning: .trellis/rules.toml has "
	if n == 1 {
		if m == 1 {
			return lead + "1 more ignored entry, and it tries to switch a rule off."
		}
		return lead + "1 more ignored entry, and it does not try to switch a rule off."
	}
	switch m {
	case 0:
		return fmt.Sprintf(lead+"%d more ignored entries, none of which tries to switch a rule off.", n)
	case 1:
		return fmt.Sprintf(lead+"%d more ignored entries, 1 of which tries to switch a rule off.", n)
	}
	return fmt.Sprintf(lead+"%d more ignored entries, %d of which try to switch a rule off.", n, m)
}

// warnSymlinkCounted is the only warning a symlinked file draws: a count, and
// nothing read from the file it points at.
func warnSymlinkCounted(n, m int) string {
	const lead = "Trellis warning: .trellis/rules.toml is a symbolic link, so its "
	if n == 1 {
		if m == 1 {
			return lead + "1 ignored entry is counted but not named; it tries to switch a rule off."
		}
		return lead + "1 ignored entry is counted but not named; it does not try to switch a rule off."
	}
	switch m {
	case 0:
		return fmt.Sprintf(lead+"%d ignored entries are counted but not named; none of them tries to switch a rule off.", n)
	case 1:
		return fmt.Sprintf(lead+"%d ignored entries are counted but not named; 1 of them tries to switch a rule off.", n)
	}
	return fmt.Sprintf(lead+"%d ignored entries are counted but not named; %d of them try to switch a rule off.", n, m)
}

// warningBlockBytes is what the bound charges for the warnings: every warning
// line with its newline, the count line included.
func warningBlockBytes(warnings []string) int {
	if len(warnings) == 0 {
		return 0
	}
	return len(strings.Join(warnings, "\n")) + 1
}

type rulesSegment int

const (
	segmentEcho rulesSegment = iota
	segmentTooLarge
	segmentSymlink
	segmentNone
)

// rulesLink says how the project reaches its rules file.
type rulesLink int

const (
	linkNone rulesLink = iota
	// linkFile: .trellis/rules.toml is a symbolic link to a file outside the project.
	linkFile
	// linkDir: .trellis is a symbolic link to a directory outside the project.
	linkDir
)

type rulesOutcome string

const (
	outcomeDeliver    rulesOutcome = "deliver"
	outcomeRefuse     rulesOutcome = "refuse"
	outcomeUngoverned rulesOutcome = "ungoverned"
)

type rulesRowsCase struct {
	name       string
	toml       string
	unreadable bool
	link       rulesLink
	outcome    rulesOutcome // "" means deliver
	off        []string
	segment    rulesSegment
	warnings   []string
	// codexRefusal is the failure class a refused file draws from Codex; ""
	// means unreadable-file. Claude refuses every such file as
	// TRELLIS_RULES_NOT_LOADED.
	codexRefusal string
	// absent is text the file holds that neither host may put anywhere in its
	// output, the Codex systemMessage included.
	absent []string
	// largest marks a row built to the largest section its branch can deliver:
	// the Codex context must leave codexContextMargin under the cap.
	largest bool
}

func (c rulesRowsCase) wantOutcome() rulesOutcome {
	if c.outcome == "" {
		return outcomeDeliver
	}
	return c.outcome
}

// expectedSegment is the file part of the region: the file verbatim (with a
// newline supplied when it has none), the too-large line, the symlink line, or
// nothing.
func (c rulesRowsCase) expectedSegment() string {
	switch c.segment {
	case segmentTooLarge:
		return expectedTooLargeLine() + "\n"
	case segmentSymlink:
		return expectedSymlinkLine() + "\n"
	case segmentNone:
		return ""
	}
	// The hook output is decoded as JSON, which turns every byte that is not
	// valid UTF-8 into U+FFFD, one per byte; converting through runes does the
	// same, and leaves a valid file unchanged.
	file := string([]rune(c.toml))
	if file == "" || strings.HasSuffix(file, "\n") {
		return file
	}
	return file + "\n"
}

// expectedRegion is everything between the heading's blank line and the host's
// footer: the sentence, the segment, then the warnings, one blank line apart.
func (c rulesRowsCase) expectedRegion() string {
	parts := []string{expectedActivationSentence(c.off) + "\n"}
	if seg := c.expectedSegment(); seg != "" {
		parts = append(parts, seg)
	}
	if len(c.warnings) > 0 {
		parts = append(parts, strings.Join(c.warnings, "\n")+"\n")
	}
	return strings.Join(parts, "\n")
}

// checkBoundPremise keeps the table honest about the bound: a row that declares
// the echo or the too-large line must be on that side of B, so a golden cannot
// pin the wrong branch by mistake.
func (c rulesRowsCase) checkBoundPremise(t *testing.T) {
	t.Helper()
	if c.wantOutcome() != outcomeDeliver || c.link != linkNone || (c.segment != segmentEcho && c.segment != segmentTooLarge) {
		return
	}
	charged := len(c.toml) + warningBlockBytes(c.warnings)
	want := segmentEcho
	if charged > rulesShowMaxBytes {
		want = segmentTooLarge
	}
	if c.segment != want {
		t.Fatalf("premise: the file and its warnings are %d bytes against a bound of %d, so the row must declare segment %d, not %d", charged, rulesShowMaxBytes, want, c.segment)
	}
}

// rulesFileOfBytes pads body with one comment line to exactly n bytes.
func rulesFileOfBytes(t *testing.T, body string, n int) string {
	t.Helper()
	pad := n - len(body)
	if pad < 3 {
		t.Fatalf("cannot pad a %d-byte body to %d bytes with a comment line", len(body), n)
	}
	out := body + "# " + strings.Repeat("x", pad-3) + "\n"
	if len(out) != n {
		t.Fatalf("padded fixture is %d bytes, wanted %d", len(out), n)
	}
	return out
}

// shippedNonFloorSlugs is the payload's non-floor slugs in rules.md order, the
// order the computed sentence names them in.
func shippedNonFloorSlugs(t *testing.T) []string {
	t.Helper()
	var slugs []string
	for _, m := range regexp.MustCompile("(?m)`((?:inv|floor)-[a-z-]+)`[ \t]*$").FindAllStringSubmatch(payloadFiles()["rules.md"], -1) {
		if !strings.HasPrefix(m[1], "floor-") {
			slugs = append(slugs, m[1])
		}
	}
	if len(slugs) != 14 {
		t.Fatalf("premise: the payload ships fourteen non-floor rules, found %d", len(slugs))
	}
	return slugs
}

func rulesRowsCases(t *testing.T) []rulesRowsCase {
	t.Helper()
	nonFloor := shippedNonFloorSlugs(t)

	var fourteen strings.Builder
	fourteen.WriteString("[rules]  # one row per assessable catalog slug\n")
	for _, slug := range assessableSlugs {
		// The fourteen-row set that predates these two rules.
		if slug == "inv-deliberate-succession" || slug == "inv-no-orphan-followups" {
			continue
		}
		fmt.Fprintf(&fourteen, "%-25s = { active = true }\n", slug)
	}

	thisRepo := readFileT(t, "../.trellis/rules.toml")
	if !strings.Contains(thisRepo, `strictness  = "firm"`) || !strings.Contains(thisRepo, "seeded_from") ||
		strings.Count(thisRepo, "{ active = true }") != 16 || !strings.HasSuffix(thisRepo, "\n") {
		t.Fatal("premise: AE8 is this repository's file with strictness = \"firm\", seeded_from and sixteen true rows, ending in a newline; it no longer is")
	}

	const optOut = "[rules]\ninv-minimal-first = { active = false }\n"
	longSlug := "inv-" + strings.Repeat("a", 60)

	// The retired rules-b.toml preset as it last shipped (4ef640a^), with eight
	// false rows appended: the code review's reproduction of a Codex session that
	// loaded no rules at all.
	presetB := strings.Replace(strings.Replace(legacyFirmRulesToml,
		`seeded_from = "conductor"`, `seeded_from = "author-adapt"`, 1),
		`strictness  = "firm"`, `strictness  = "adaptive"`, 1)
	presetBPlusEight := presetB
	for _, slug := range nonFloor[:8] {
		presetBPlusEight += slug + " = { active = false }\n"
	}
	if len(presetB) != 1149 || len(presetBPlusEight) != 1493 {
		t.Fatalf("premise: rules-b.toml is 1149 bytes and 1493 with eight false rows; got %d and %d", len(presetB), len(presetBPlusEight))
	}

	// This repository's file with six mixed warnings appended, about 1.3 KB.
	base := strings.Count(thisRepo, "\n")
	mixed := thisRepo +
		"floor-transparency = { active = false }\n" +
		"inv-not-a-rule = { active = false }\n" +
		"inv-bounded-context = off\n" +
		"governed = false\n" +
		"[rules]\n" +
		"[tool]\n"

	var allAbove strings.Builder
	var allAboveWarnings []string
	for i, slug := range nonFloor {
		allAbove.WriteString(slug + " = { active = false }\n")
		if i < 5 {
			allAboveWarnings = append(allAboveWarnings, warnRowAboveRules(i+1, slug))
		}
	}
	allAbove.WriteString("[rules]\n")
	allAboveWarnings = append(allAboveWarnings, warnMore(9, 9))

	withWarning := "[rules]\nfloor-transparency = { active = false }\ninv-minimal-first = { active = false }\n"
	withWarningWarnings := []string{warnFloorRow(2, "floor-transparency")}
	fitsWithWarning := rulesShowMaxBytes - warningBlockBytes(withWarningWarnings)

	// The largest echoed section: every non-floor rule switched off, two warnings
	// with names at the cut, and the file padded so it and its warnings are
	// exactly B, with no trailing newline, so the hook supplies one more byte.
	longKey := strings.Repeat("k", 45)
	cutKey := longKey[:40] + "..."
	var largestBody strings.Builder
	largestBody.WriteString(longKey + "1 = 1\n" + longKey + "2 = 1\n[rules]\n")
	for _, slug := range nonFloor {
		largestBody.WriteString(slug + "={active=false}\n")
	}
	largestWarnings := []string{warnUnknownKey(1, cutKey), warnUnknownKey(2, cutKey)}
	largestEcho := strings.TrimSuffix(rulesFileOfBytes(t, largestBody.String(), rulesShowMaxBytes-warningBlockBytes(largestWarnings)+1), "\n")

	// The largest too-large section: every non-floor rule switched off and the
	// longest warning kind five times over, at seven-digit line numbers, beside a
	// count line.
	var runaway strings.Builder
	runaway.WriteString(strings.Repeat("\n", 999_999))
	var runawayWarnings []string
	for i := 0; i < 8; i++ {
		fmt.Fprintf(&runaway, "%s%d = 1\n", longKey, i)
		if i < 5 {
			runawayWarnings = append(runawayWarnings, warnUnknownKey(1_000_000+i, cutKey))
		}
	}
	runaway.WriteString("[rules]\n")
	for _, slug := range nonFloor {
		runaway.WriteString(slug + "={active=false}\n")
	}
	runawayWarnings = append(runawayWarnings, warnMore(3, 0))

	// A file a symbolic link points at, holding what must never be shown.
	const canaryKey = "trellis_canary_unknown_key"
	const canarySecret = "wJalrXUtnFEMIK7MDENGbPxRfiCYTRELLISCANARY"
	credentials := canaryKey + " = 1\n" +
		"aws_secret_access_key = " + canarySecret + "\n" +
		"[rules]\n" +
		"inv-minimal-first = { active = false }\n" +
		"floor-transparency = { active = false }\n"
	canaries := []string{canaryKey, "aws_secret_access_key", canarySecret}

	return []rulesRowsCase{{
		name: "AE1 a single false row switches that rule off",
		toml: optOut,
		off:  []string{"inv-minimal-first"},
	}, {
		name:     "AE3 an unknown slug set false is ignored and named",
		toml:     "[rules]\ninv-not-a-rule = { active = false }\n",
		warnings: []string{warnUnknownSlug(2, "inv-not-a-rule")},
	}, {
		name:     "AE4 a floor row set false is ignored",
		toml:     "[rules]\nfloor-transparency = { active = false }\n",
		warnings: []string{warnFloorRow(2, "floor-transparency")},
	}, {
		name:     "AE5 a malformed row costs only itself",
		toml:     "[rules]\ninv-bounded-context = off\ninv-minimal-first = { active = false }\n",
		off:      []string{"inv-minimal-first"},
		warnings: []string{warnMalformedRow(2, "inv-bounded-context")},
	}, {
		name: "AE7 fourteen true rows from an older release",
		toml: fourteen.String(),
	}, {
		name: "AE8 this repository's own file",
		toml: thisRepo,
	}, {
		name: "an empty file",
		toml: "",
	}, {
		name: "only a [rules] header",
		toml: "[rules]\n",
	}, {
		name: "a [rules] header with a trailing comment",
		toml: "[rules]  # one row per rule\n",
	}, {
		name: "CRLF line endings keep an opt-out",
		toml: "[rules]\r\ninv-minimal-first = { active = false }\r\n",
		off:  []string{"inv-minimal-first"},
	}, {
		name: "a byte-order mark before [rules] keeps an opt-out",
		toml: "\ufeff" + optOut,
		off:  []string{"inv-minimal-first"},
	}, {
		name: "a file with no trailing newline",
		toml: strings.TrimSuffix(optOut, "\n"),
		off:  []string{"inv-minimal-first"},
	}, {
		name: "accepted row spellings",
		toml: "  [rules]\n" +
			"inv-directional-flow={active=false}\n" +
			"inv-handover-points\t=\t{\tactive\t=\tfalse\t}\n" +
			"inv-intent-locus = { active = false }  # off for now\n" +
			"inv-ratifiable-artifacts = { active = false }   \n" +
			"\tinv-graph-maintenance = { active = false }\n",
		off: []string{"inv-directional-flow", "inv-handover-points", "inv-intent-locus", "inv-ratifiable-artifacts", "inv-graph-maintenance"},
	}, {
		name: "rejected row spellings",
		toml: "[rules]\n" +
			"\u00a0inv-graph-maintenance = { active = false }\n" +
			"inv-self-improvement = { active = False }\n" +
			"\"inv-deliberate-succession\" = { active = false }\n" +
			"inv-no-orphan-followups.active = false\n",
		warnings: []string{
			warnMalformedRow(2, ""),
			warnMalformedRow(3, "inv-self-improvement"),
			warnMalformedRow(4, ""),
			warnMalformedRow(5, ""),
		},
	}, {
		name: "a multi-line inline table is three malformed lines",
		toml: "[rules]\ninv-gate-at-handover = {\n  active = false\n}\n",
		warnings: []string{
			warnMalformedRow(2, "inv-gate-at-handover"),
			warnMalformedRow(3, "active"),
			warnMalformedRow(4, ""),
		},
	}, {
		// Q4: any false row wins, and a repeated row is never an error.
		name: "a duplicate slug, false then true, is off and silent",
		toml: optOut + "inv-minimal-first = { active = true }\n",
		off:  []string{"inv-minimal-first"},
	}, {
		name: "a duplicate slug, true then false, is off and silent",
		toml: "[rules]\ninv-minimal-first = { active = true }\ninv-minimal-first = { active = false }\n",
		off:  []string{"inv-minimal-first"},
	}, {
		name: "three rows for one slug are off and silent",
		toml: "[rules]\ninv-minimal-first = { active = true }\ninv-minimal-first = { active = false }\ninv-minimal-first = { active = true }\n",
		off:  []string{"inv-minimal-first"},
	}, {
		name: "repeated true rows are silent",
		toml: "[rules]\ninv-minimal-first = { active = true }\ninv-minimal-first = { active = true }\n",
	}, {
		name:     "a floor row set false twice draws one warning, at the first line",
		toml:     "[rules]\nfloor-transparency = { active = false }\nfloor-transparency = { active = false }\n",
		warnings: []string{warnFloorRow(2, "floor-transparency")},
	}, {
		// The true row draws nothing, so the first line of the floor-row kind is
		// the false row.
		name:     "a floor row true then false draws its warning at the false row",
		toml:     "[rules]\nfloor-intent-gate = { active = true }\nfloor-intent-gate = { active = false }\n",
		warnings: []string{warnFloorRow(3, "floor-intent-gate")},
	}, {
		name:     "an unknown slug set false twice draws one warning",
		toml:     "[rules]\ninv-not-a-rule = { active = false }\ninv-not-a-rule = { active = true }\ninv-not-a-rule = { active = false }\n",
		warnings: []string{warnUnknownSlug(2, "inv-not-a-rule")},
	}, {
		name:     "a slug repeated above [rules] draws one warning",
		toml:     "inv-minimal-first = { active = true }\ninv-minimal-first = { active = false }\n[rules]\n",
		warnings: []string{warnRowAboveRules(1, "inv-minimal-first")},
	}, {
		name: "the retired rules-b preset plus eight false rows switches eight rules off and is shown",
		toml: presetBPlusEight,
		off:  nonFloor[:8],
	}, {
		name:     "a second [rules] header continues the table",
		toml:     optOut + "[rules]\ninv-bounded-context = { active = false }\n",
		off:      []string{"inv-bounded-context", "inv-minimal-first"},
		warnings: []string{warnRulesAgain(3)},
	}, {
		name:     "a false row under another table",
		toml:     "[rules]\n[tool]\ninv-minimal-first = { active = false }\nname = \"x\"\n",
		warnings: []string{warnOtherTable(2)},
	}, {
		name:     "a false row above [rules]",
		toml:     "inv-minimal-first = { active = false }\n[rules]\n",
		warnings: []string{warnRowAboveRules(1, "inv-minimal-first")},
	}, {
		name:     "an unknown top-level key",
		toml:     "disabled = [\"inv-minimal-first\"]\n[rules]\n",
		warnings: []string{warnUnknownKey(1, "disabled")},
	}, {
		name:     "a top-level line that is not a key",
		toml:     "just some words\n[rules]\n",
		warnings: []string{warnTopLevelLine(1)},
	}, {
		name: "strictness and seeded_from are silent whatever they hold",
		toml: "strictness = \"loose\"\nseeded_from = 'anything'\nstrictness = 3\n" + optOut,
		off:  []string{"inv-minimal-first"},
	}, {
		name: "an unknown slug set true is silent",
		toml: "[rules]\ninv-retired-rule = { active = true }\n",
	}, {
		name: "commented-out rows and old quarantine notes are silent",
		toml: "[rules]\n# inv-gone = { active = false }  # quarantined 2026-01-01: not in payload@000000000000.\ninv-minimal-first = { active = true }\n",
	}, {
		name:     "a long unknown slug is named in part",
		toml:     "[rules]\n" + longSlug + " = { active = false }\n",
		warnings: []string{warnUnknownSlug(2, longSlug[:40]+"...")},
	}, {
		name:     "governed = false under [rules] does not opt out",
		toml:     optOut + "governed = false\n",
		off:      []string{"inv-minimal-first"},
		warnings: []string{warnGoverned(3)},
	}, {
		name:     "governed = false twice does not opt out",
		toml:     "governed = false\ngoverned = false\n[rules]\n",
		warnings: []string{warnGoverned(1), warnGoverned(2)},
	}, {
		name:     "governed = \"false\" does not opt out",
		toml:     "governed = \"false\"\n[rules]\n",
		warnings: []string{warnGoverned(1)},
	}, {
		name: "governed = true is silent",
		toml: "governed = true\n" + optOut,
		off:  []string{"inv-minimal-first"},
	}, {
		name:    "AE2 governed = false alone",
		toml:    "governed = false\n",
		outcome: outcomeUngoverned,
	}, {
		name:    "AE2 governed = false beside an unknown key",
		toml:    "governed = false\nprofile = \"b\"\n[rules]\n",
		outcome: outcomeUngoverned,
	}, {
		name:    "AE2 governed = false beside a malformed row",
		toml:    "governed = false\n[rules]\ninv-bounded-context = off\n",
		outcome: outcomeUngoverned,
	}, {
		name:    "AE2 governed = false beside a NUL byte",
		toml:    "governed = false\n[rules]\n\x00\n",
		outcome: outcomeUngoverned,
	}, {
		name:     "a NUL byte and no opt-out",
		toml:     optOut + "\x00\n",
		segment:  segmentNone,
		warnings: []string{warnNulByte()},
	}, {
		// A byte that is not valid UTF-8 costs nothing either: under a UTF-8 locale
		// the Claude hook's JSON escaping once stopped at it and dropped the rest of
		// the file, every warning and the footer.
		name:     "a Latin-1 byte in a comment keeps the whole section",
		toml:     "# Cr\xe9\xe9 par l'\xe9quipe\n[rules]\nfloor-intent-gate = { active = false }\ninv-minimal-first = { active = false }\n",
		off:      []string{"inv-minimal-first"},
		warnings: []string{warnFloorRow(3, "floor-intent-gate")},
	}, {
		// Q2: five named warnings in total, then one count line.
		name: "twelve malformed rows name five and count the rest",
		toml: "[rules]\n" + strings.Repeat("{ active = false }\n", 12),
		warnings: []string{
			warnMalformedRow(2, ""), warnMalformedRow(3, ""), warnMalformedRow(4, ""),
			warnMalformedRow(5, ""), warnMalformedRow(6, ""), warnMore(7, 0),
		},
	}, {
		// The two rows above [rules] for inv-directional-flow are one entry, and
		// its false row makes it a real-rule attempt, so it is named at line 2.
		name: "real-rule attempts are named first, then the other warnings in line order",
		toml: "disabled = [\"inv-minimal-first\"]\n" +
			"inv-directional-flow = { active = true }\n" +
			"inv-directional-flow = { active = false }\n" +
			"inv-handover-points = { active = false }\n" +
			"[rules]\n" +
			"not a row\n" +
			"floor-transparency = { active = false }\n" +
			"inv-not-a-rule = { active = false }\n" +
			"floor-intent-gate = { active = true }\n" +
			"floor-intent-gate = { active = false }\n" +
			"inv-intent-locus = { active = false }\n" +
			"inv-graph-maintenance = { active = true }\n" +
			"inv-graph-maintenance = { active = false }\n" +
			"inv-graph-maintenance = { active = true }\n" +
			"[tool]\n" +
			"inv-ratifiable-artifacts = { active = false }\n",
		off: []string{"inv-intent-locus", "inv-graph-maintenance"},
		warnings: []string{
			warnRowAboveRules(2, "inv-directional-flow"),
			warnRowAboveRules(4, "inv-handover-points"),
			warnFloorRow(7, "floor-transparency"),
			warnFloorRow(10, "floor-intent-gate"),
			warnUnknownKey(1, "disabled"),
			warnMore(3, 0),
		},
	}, {
		name: "seven real-rule attempts: the count line says how many of the rest try",
		toml: "strictness = \"firm\"\n" +
			"mystery = 1\n" +
			"inv-directional-flow = { active = false }\n" +
			"inv-handover-points = { active = false }\n" +
			"inv-intent-locus = { active = false }\n" +
			"inv-ratifiable-artifacts = { active = false }\n" +
			"inv-graph-maintenance = { active = false }\n" +
			"[rules]\n" +
			"floor-transparency = { active = false }\n" +
			"junk\n" +
			"floor-intent-gate = { active = false }\n" +
			"inv-not-a-rule = { active = false }\n",
		warnings: []string{
			warnRowAboveRules(3, "inv-directional-flow"),
			warnRowAboveRules(4, "inv-handover-points"),
			warnRowAboveRules(5, "inv-intent-locus"),
			warnRowAboveRules(6, "inv-ratifiable-artifacts"),
			warnRowAboveRules(7, "inv-graph-maintenance"),
			warnMore(5, 2),
		},
	}, {
		name: "six real-rule attempts beside one other warning",
		toml: "other = 1\n" +
			"inv-directional-flow = { active = false }\n" +
			"inv-handover-points = { active = false }\n" +
			"inv-intent-locus = { active = false }\n" +
			"inv-ratifiable-artifacts = { active = false }\n" +
			"[rules]\n" +
			"floor-transparency = { active = false }\n" +
			"floor-intent-gate = { active = false }\n",
		warnings: []string{
			warnRowAboveRules(2, "inv-directional-flow"),
			warnRowAboveRules(3, "inv-handover-points"),
			warnRowAboveRules(4, "inv-intent-locus"),
			warnRowAboveRules(5, "inv-ratifiable-artifacts"),
			warnFloorRow(7, "floor-transparency"),
			warnMore(2, 1),
		},
	}, {
		name: "six real-rule attempts alone",
		toml: "inv-directional-flow = { active = false }\n" +
			"inv-handover-points = { active = false }\n" +
			"inv-intent-locus = { active = false }\n" +
			"inv-ratifiable-artifacts = { active = false }\n" +
			"[rules]\n" +
			"floor-transparency = { active = false }\n" +
			"floor-intent-gate = { active = false }\n",
		warnings: []string{
			warnRowAboveRules(1, "inv-directional-flow"),
			warnRowAboveRules(2, "inv-handover-points"),
			warnRowAboveRules(3, "inv-intent-locus"),
			warnRowAboveRules(4, "inv-ratifiable-artifacts"),
			warnFloorRow(6, "floor-transparency"),
			warnMore(1, 1),
		},
	}, {
		name:     "every non-floor rule set false above [rules] is ignored, named five times and counted",
		toml:     allAbove.String(),
		warnings: allAboveWarnings,
	}, {
		// Q1: about 1.3 KB and six warnings is past the bound, so the too-large
		// line stands in for the file and both hosts still deliver.
		name:    "a 1.3 KB file with six mixed warnings is not shown and still delivers",
		toml:    mixed,
		segment: segmentTooLarge,
		warnings: []string{
			warnFloorRow(base+1, "floor-transparency"),
			warnUnknownSlug(base+2, "inv-not-a-rule"),
			warnMalformedRow(base+3, "inv-bounded-context"),
			warnGoverned(base + 4),
			warnRulesAgain(base + 5),
			warnMore(1, 0),
		},
	}, {
		name: "a file of exactly B bytes and no warning is echoed",
		toml: rulesFileOfBytes(t, optOut, rulesShowMaxBytes),
		off:  []string{"inv-minimal-first"},
	}, {
		name:    "a file one byte over B and no warning is not echoed and still switches its rule off",
		toml:    rulesFileOfBytes(t, optOut, rulesShowMaxBytes+1),
		off:     []string{"inv-minimal-first"},
		segment: segmentTooLarge,
	}, {
		name:     "a file and its warnings of exactly B bytes are echoed",
		toml:     rulesFileOfBytes(t, withWarning, fitsWithWarning),
		off:      []string{"inv-minimal-first"},
		warnings: withWarningWarnings,
	}, {
		name:     "a file and its warnings one byte over B are not echoed",
		toml:     rulesFileOfBytes(t, withWarning, fitsWithWarning+1),
		off:      []string{"inv-minimal-first"},
		segment:  segmentTooLarge,
		warnings: withWarningWarnings,
	}, {
		name:     "the largest echoed section fits the Codex context with its margin",
		toml:     largestEcho,
		off:      nonFloor,
		warnings: largestWarnings,
		largest:  true,
	}, {
		name:     "the largest too-large section fits the Codex context with its margin",
		toml:     runaway.String(),
		off:      nonFloor,
		segment:  segmentTooLarge,
		warnings: runawayWarnings,
		largest:  true,
	}, {
		// Q5: a symbolic link is classified, and nothing read from its target is
		// shown or named.
		name:     "a symlinked rules file is classified and not shown",
		toml:     credentials,
		link:     linkFile,
		off:      []string{"inv-minimal-first"},
		segment:  segmentSymlink,
		warnings: []string{warnSymlinkCounted(3, 1)},
		absent:   canaries,
	}, {
		name:     "a symlinked .trellis directory is classified and not shown",
		toml:     credentials,
		link:     linkDir,
		off:      []string{"inv-minimal-first"},
		segment:  segmentSymlink,
		warnings: []string{warnSymlinkCounted(3, 1)},
		absent:   canaries,
	}, {
		name:    "a symlinked rules file with no ignored entry draws no warning",
		toml:    optOut,
		link:    linkFile,
		off:     []string{"inv-minimal-first"},
		segment: segmentSymlink,
	}, {
		name:    "an empty symlinked rules file is still not shown",
		toml:    "",
		link:    linkFile,
		segment: segmentSymlink,
	}, {
		name:     "a symlinked file with one ignored entry that tries nothing",
		toml:     "[rules]\n" + canaryKey + "\n",
		link:     linkFile,
		segment:  segmentSymlink,
		warnings: []string{warnSymlinkCounted(1, 0)},
		absent:   []string{canaryKey},
	}, {
		name:     "a symlinked directory with one ignored entry that tries to switch a rule off",
		toml:     "[rules]\nfloor-intent-gate = { active = false }\n",
		link:     linkDir,
		segment:  segmentSymlink,
		warnings: []string{warnSymlinkCounted(1, 1)},
	}, {
		name:     "a symlinked file counts every ignored entry, past five too",
		toml:     "[rules]\n" + strings.Repeat("not a row\n", 12),
		link:     linkFile,
		segment:  segmentSymlink,
		warnings: []string{warnSymlinkCounted(12, 0)},
	}, {
		name:     "a symlinked file whose ignored entries include two attempts",
		toml:     "[rules]\nfloor-transparency = { active = false }\nfloor-intent-gate = { active = false }\n" + canaryKey + " = 1\n",
		link:     linkFile,
		segment:  segmentSymlink,
		warnings: []string{warnSymlinkCounted(3, 2)},
		absent:   []string{canaryKey},
	}, {
		name:     "a symlinked file holding a NUL byte keeps the NUL handling",
		toml:     credentials + "\x00\n",
		link:     linkFile,
		segment:  segmentNone,
		warnings: []string{warnNulByte()},
		absent:   canaries,
	}, {
		name:    "a symlinked governed = false still opts out",
		toml:    "governed = false\n",
		link:    linkFile,
		outcome: outcomeUngoverned,
	}, {
		name:       "an unreadable file is refused",
		toml:       optOut,
		unreadable: true,
		outcome:    outcomeRefuse,
	}, {
		// The read bound: a file of exactly rulesReadMaxBytes is still
		// classified, and one byte more is refused on both hosts.
		name:    "a file of exactly the read bound is classified and not echoed",
		toml:    rulesFileOfBytes(t, optOut, rulesReadMaxBytes),
		off:     []string{"inv-minimal-first"},
		segment: segmentTooLarge,
	}, {
		name:         "a file one byte over the read bound is refused",
		toml:         rulesFileOfBytes(t, optOut, rulesReadMaxBytes+1),
		outcome:      outcomeRefuse,
		codexRefusal: ".trellis/rules.toml: context-over-budget",
	}}
}

// hostRulesResult is what one hook did with one file.
type hostRulesResult struct {
	outcome       rulesOutcome
	context       string
	region        string
	systemMessage string
	raw           string
}

// activationRegion returns the text between the heading's blank line and the
// host's footer. The footer is matched from the end, because an echoed file is
// free to contain anything.
func activationRegion(context, footer string) (string, bool) {
	lead := "\n" + activationHeading + "\n\n"
	i := strings.Index(context, lead)
	if i < 0 {
		return "", false
	}
	start := i + len(lead)
	j := strings.LastIndex(context, footer)
	if j < start {
		return "", false
	}
	return context[start:j], true
}

func runRulesRowsHost(t *testing.T, host, pluginRoot, project string) hostRulesResult {
	t.Helper()
	switch host {
	case "claude":
		out := claudeStdoutFor(t, pluginRoot, project)
		r := hostRulesResult{raw: out}
		if out == "" {
			r.outcome = outcomeUngoverned
			return r
		}
		ctx := nudgeContext(t, out)
		r.context = ctx
		switch {
		case strings.Contains(ctx, rulesLoadedSentinel):
			r.outcome = outcomeDeliver
			r.region, _ = activationRegion(ctx, "\nDelivered by the Trellis plugin")
		case strings.Contains(ctx, "TRELLIS_NOT_GOVERNING"):
			r.outcome = outcomeUngoverned
		default:
			r.outcome = outcomeRefuse
		}
		return r
	case "codex":
		raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
		r := hostRulesResult{raw: raw, systemMessage: got.SystemMessage}
		switch {
		case raw == "":
			r.outcome = outcomeUngoverned
		case got.HookSpecificOutput == nil:
			r.outcome = outcomeRefuse
		default:
			r.outcome = outcomeDeliver
			r.context = got.HookSpecificOutput.AdditionalContext
			r.region, _ = activationRegion(r.context, "\nTrellis hook loaded installed overlay: ")
		}
		return r
	}
	t.Fatalf("unknown host %q", host)
	return hostRulesResult{}
}

// splitActivationRegion pulls the three parts back out of a region, so a
// mismatch says which part moved. The whole-region equality is the assertion;
// this is its diagnosis.
func splitActivationRegion(region string) (sentence, segment string, warnings []string) {
	lines := strings.SplitAfter(region, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return "", "", nil
	}
	sentence = strings.TrimSuffix(lines[0], "\n")
	rest := lines[1:]
	if len(rest) > 0 && rest[0] == "\n" {
		rest = rest[1:]
	}
	n := len(rest)
	for n > 0 && strings.HasPrefix(rest[n-1], "Trellis warning: ") {
		n--
	}
	for _, w := range rest[n:] {
		warnings = append(warnings, strings.TrimSuffix(w, "\n"))
	}
	body := rest[:n]
	if len(warnings) > 0 && len(body) > 0 && body[len(body)-1] == "\n" {
		body = body[:len(body)-1]
	}
	return sentence, strings.Join(body, ""), warnings
}

// writeInstructionRe is what the hooks' own framing and warnings may never say:
// an instruction to change the file, or the vocabulary of the retired repair.
var writeInstructionRe = regexp.MustCompile(`(?i)\b(write|rewrite|edit|uncomment)\b|reconcil|quarantin`)

func assertRulesRowsHost(t *testing.T, host string, c rulesRowsCase, r hostRulesResult) {
	t.Helper()
	// Q5: nothing read from a symlinked file reaches either host's output, on
	// any outcome. Checked first, because a leak is worse than a wrong branch.
	for _, s := range c.absent {
		if !strings.Contains(c.toml, s) {
			t.Fatalf("premise: the file must hold %q for its absence to mean anything", s)
		}
		if strings.Contains(r.raw, s) || strings.Contains(r.context, s) || strings.Contains(r.systemMessage, s) {
			t.Errorf("%s: text from the file a symbolic link points at reached the output: %q\n%s", host, s, r.raw)
		}
	}
	if r.outcome != c.wantOutcome() {
		t.Fatalf("%s: want outcome %q, got %q\n%s", host, c.wantOutcome(), r.outcome, r.raw)
	}
	switch r.outcome {
	case outcomeRefuse:
		wantCodex := ".trellis/rules.toml: unreadable-file"
		if c.codexRefusal != "" {
			wantCodex = c.codexRefusal
		}
		if host == "codex" && !strings.Contains(r.raw, wantCodex) {
			t.Errorf("codex: this file must be refused as %q:\n%s", wantCodex, r.raw)
		}
		if host == "claude" && !strings.Contains(r.context, "TRELLIS_RULES_NOT_LOADED") {
			t.Errorf("claude: this file must be refused as TRELLIS_RULES_NOT_LOADED:\n%s", r.context)
		}
		return
	case outcomeUngoverned:
		// AE2: no rule loads, and nothing delivered says a floor rule applies.
		if strings.Contains(r.raw, "trellis:rules-loaded") || strings.Contains(r.raw, "Project rule activation") {
			t.Errorf("%s: a governed = false project was handed rules:\n%s", host, r.raw)
		}
		return
	}

	if !strings.Contains(r.context, rulesLoadedSentinel) {
		t.Fatalf("%s: the rules themselves were not delivered:\n%s", host, r.context)
	}
	// R14 and AE8: one posture sentence for every project, whatever the file says.
	if !strings.Contains(r.context, "**By default**") || strings.Contains(r.context, "**Firmly**") {
		t.Errorf("%s: every project receives the By default posture sentence and no other:\n%s", host, r.context)
	}
	if r.region == "" && !strings.Contains(r.context, "\n"+activationHeading+"\n\n") {
		t.Fatalf("%s: the context carries no %q section:\n%s", host, activationHeading, r.context)
	}

	want := c.expectedRegion()
	sentence, segment, warnings := splitActivationRegion(r.region)
	if wantSentence := expectedActivationSentence(c.off); sentence != wantSentence {
		t.Errorf("%s: computed sentence\n got: %q\nwant: %q", host, sentence, wantSentence)
	}
	if wantSegment := c.expectedSegment(); segment != wantSegment {
		t.Errorf("%s: file segment\n got: %q\nwant: %q", host, segment, wantSegment)
	}
	if strings.Join(warnings, "\n") != strings.Join(c.warnings, "\n") {
		t.Errorf("%s: warnings, in order\n got: %q\nwant: %q", host, warnings, c.warnings)
	}
	if r.region != want {
		t.Errorf("%s: the activation section differs from the golden text\n got: %q\nwant: %q", host, r.region, want)
	}

	// R15: nothing the hook itself wrote asks for the file to change. The echoed
	// file is the project's own text and is excluded.
	hookText := strings.Replace(r.region, c.expectedSegment(), "", 1)
	if m := writeInstructionRe.FindString(hookText); m != "" {
		t.Errorf("%s: the hook's framing or warnings say %q, and nothing may ask for .trellis/rules.toml to change:\n%s", host, m, hookText)
	}

	if host == "codex" {
		n := len(r.context)
		if n > codexContextCap {
			t.Errorf("codex: the context is %d bytes, over MAX_CONTEXT_BYTES (%d)", n, codexContextCap)
		}
		// Q1: no project file can cost a Codex session its rules, and the largest
		// section the bound allows leaves room for a longer plugin root.
		if c.largest {
			t.Logf("codex: the largest section of this kind assembles a %d-byte context, %d under MAX_CONTEXT_BYTES", n, codexContextCap-n)
			if codexContextCap-n < codexContextMargin {
				t.Errorf("codex: the largest section leaves %d bytes under MAX_CONTEXT_BYTES, less than the %d-byte margin a longer plugin root needs", codexContextCap-n, codexContextMargin)
			}
		}
		// KTD4: the warnings are mirrored to systemMessage, in the same order and
		// nothing else: the plugin root here is complete, so no invariants report
		// shares the message, and an extra or repeated warning is a failure.
		if wantMessage := strings.Join(c.warnings, " "); r.systemMessage != wantMessage {
			t.Errorf("codex: systemMessage\n got: %q\nwant: %q", r.systemMessage, wantMessage)
		}
	}
}

// writeRulesRowsProject writes the case's file where the case says the project
// reaches it: in place, or through a symbolic link to a place outside the project.
func writeRulesRowsProject(t *testing.T, c rulesRowsCase) (project, path string) {
	t.Helper()
	project = newGitProject(t)
	path = filepath.Join(project, ".trellis", "rules.toml")
	switch c.link {
	case linkFile:
		target := filepath.Join(t.TempDir(), "credentials")
		writeFileT(t, target, c.toml)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, path); err != nil {
			t.Fatal(err)
		}
	case linkDir:
		outside := filepath.Join(t.TempDir(), "elsewhere")
		writeFileT(t, filepath.Join(outside, "rules.toml"), c.toml)
		if err := os.Symlink(outside, filepath.Join(project, ".trellis")); err != nil {
			t.Fatal(err)
		}
	default:
		writeFileT(t, path, c.toml)
	}
	return project, path
}

// TestBothHostsClassifyRulesRowsIdentically is R12 and KTD3 as one table.
func TestBothHostsClassifyRulesRowsIdentically(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	for _, c := range rulesRowsCases(t) {
		t.Run(c.name, func(t *testing.T) {
			c.checkBoundPremise(t)
			results := map[string]hostRulesResult{}
			for _, host := range []string{"claude", "codex"} {
				t.Run(host, func(t *testing.T) {
					project, path := writeRulesRowsProject(t, c)
					if c.unreadable {
						if os.Geteuid() == 0 {
							t.Skip("running as root: mode 0000 does not deny reads")
						}
						if err := os.Chmod(path, 0o000); err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
					}

					r := runRulesRowsHost(t, host, pluginRoot, project)
					results[host] = r
					assertRulesRowsHost(t, host, c, r)

					// Neither hook writes the file, whatever it says about it.
					if !c.unreadable {
						if after := readFileT(t, path); after != c.toml {
							t.Errorf("%s: the hook changed .trellis/rules.toml\nbefore: %q\nafter:  %q", host, c.toml, after)
						}
					}
				})
			}
			t.Run("parity", func(t *testing.T) {
				claude, okClaude := results["claude"]
				codex, okCodex := results["codex"]
				if !okClaude || !okCodex {
					t.Fatal("a host produced no result to compare")
				}
				if claude.outcome != codex.outcome {
					t.Fatalf("the hosts disagree about the outcome: claude %q, codex %q", claude.outcome, codex.outcome)
				}
				if claude.outcome == outcomeDeliver && claude.region != codex.region {
					t.Errorf("the hosts delivered different activation sections for the same file\nclaude: %q\ncodex:  %q", claude.region, codex.region)
				}
			})
		})
	}
}

package main

// The rows of TestBothHostsClassifyRulesRowsIdentically (rules_rows_parity_test.go),
// which holds the golden text, the harness and the assertions.

import (
	"fmt"
	"strings"
	"testing"
)

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
		// A byte-order mark is three bytes of the file on both hosts, so it counts
		// toward B like any other.
		name: "a file with a byte-order mark of exactly B bytes is echoed",
		toml: "\ufeff" + rulesFileOfBytes(t, optOut, rulesShowMaxBytes-3),
		off:  []string{"inv-minimal-first"},
	}, {
		name:    "a file with a byte-order mark one byte over B is not echoed",
		toml:    "\ufeff" + rulesFileOfBytes(t, optOut, rulesShowMaxBytes+1-3),
		off:     []string{"inv-minimal-first"},
		segment: segmentTooLarge,
	}, {
		// Code review round 2: a Latin-1 comment fits B as read, but each of its
		// bytes enters the Codex context as three, and a context over
		// MAX_CONTEXT_BYTES would cost the session every rule.
		name:         "a Latin-1 file that fits B only as read still delivers on Codex",
		toml:         optOut + "# " + strings.Repeat("\xe9", rulesShowMaxBytes-len(optOut)-3) + "\n",
		off:          []string{"inv-minimal-first"},
		decodedOverB: true,
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
		// Only .trellis and its rules file are checked for a link: a project
		// reached through one is an ordinary project, and its file is shown.
		name: "a project reached through a symlinked ancestor is shown as usual",
		toml: optOut,
		link: linkAncestor,
		off:  []string{"inv-minimal-first"},
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

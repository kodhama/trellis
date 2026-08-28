package main

// Textual guards over plugins/trellis/skills/remove/SKILL.md — decision-0073
// D3/D4/AC1. The skill is model-interpreted prose, so the honest assertion is
// textual (D4's assertion-mode split): these tests pin what the text CLAIMS —
// that every state in decision-0073 D1's closed set is inventoried, transacted
// and reported by name — not what a session does with it. They follow the
// section-isolation pattern of TestRemoveSkillEnumeratesTheRenderedRulesFile:
// anchor on a section heading, fail loudly when the anchor is gone, and assert
// inside the window so a mention in the wrong section cannot satisfy a claim
// about the right one.

import (
	"strings"
	"testing"
)

const removeSkillPath = "../plugins/trellis/skills/remove/SKILL.md"

// removeSkillSection isolates the block between two anchors, failing loudly if
// either is missing or out of order — an assertion over a window that silently
// became empty passes for the wrong reason.
func removeSkillSection(t *testing.T, body, from, to string) string {
	t.Helper()
	i := strings.Index(body, from)
	j := strings.Index(body, to)
	if i < 0 || j < 0 || j < i {
		t.Fatalf("cannot isolate the %q..%q window (i=%d, j=%d) — every assertion over it would silently not run", from, to, i, j)
	}
	return body[i:j]
}

// normalizeWS collapses every whitespace run to one space. SKILL.md hard-wraps
// its prose, so a multi-word needle routinely straddles a line break — two of
// this file's negative needles shipped never having matched the pre-fix bytes
// for exactly that reason (code-review finding A on this run: the sub-guards
// were green against the very sentences they pin). Line wrapping is not
// semantic; every multi-word Contains over skill prose goes through here.
func normalizeWS(s string) string { return strings.Join(strings.Fields(s), " ") }

// noOpPredicateWindow returns the no-op predicate clause: from the anchor
// ("only when there is no" appears in the predicate and nowhere else — same
// anchor as TestRemoveSkillNoOpPredicateCountsTheRenderedFile) to the next
// section heading. It was a fixed 700-byte window, which ended six bytes short
// of "## Reversing" — a latent wrong-section satisfaction waiting for the next
// edit (code-review finding D). Structure, not a byte count, bounds it now.
func noOpPredicateWindow(t *testing.T, body string) string {
	t.Helper()
	const anchor = "only when there is no"
	j := strings.Index(body, anchor)
	if j < 0 {
		t.Fatalf("cannot locate the no-op predicate clause — this assertion would silently not run")
	}
	rest := body[j:]
	k := strings.Index(rest, "\n## ")
	if k < 0 {
		t.Fatalf("cannot bound the predicate window at the next section heading — this assertion would silently read into nothing")
	}
	return rest[:k]
}

// TestRemoveSkillCoversEveryDeliveryState guards decision-0073 D3/D4/AC1: the
// table below IS decision-0073 D1's closed set, S0 through S6 — with S0
// present, because a test that enumerates S1–S5 and forgets the none-state
// repeats the defect the record was written to end (D1's first closure
// clause). Each row asserts, per section, that the skill names the state's
// on-disk artifact where it inventories (§1), and that the no-op predicate and
// the §5 report count it.
func TestRemoveSkillCoversEveryDeliveryState(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	preflight := removeSkillSection(t, body, "## 1.", "## 2.")
	report := removeSkillSection(t, body, "## 5.", "## Reversing")
	predicate := noOpPredicateWindow(t, body)

	for _, tc := range []struct {
		state   string
		window  string
		name    string
		needles []string
	}{
		// S0-unadopted — the one true "already absent": the predicate must say
		// so by name, so no removable state can ever satisfy it.
		{"S0-unadopted", predicate, "no-op predicate", []string{"S0-unadopted"}},
		// S0-config-only — a bare rules.toml is the plugin-native mode's
		// removable state, never absence (AC1's quantifier).
		{"S0-config-only", predicate, "no-op predicate", []string{"rules.toml"}},
		{"S0-config-only", report, "§5 report", []string{".trellis/rules.toml"}},
		// S1 — the rendered file (also guarded by the older
		// TestRemoveSkillEnumeratesTheRenderedRulesFile; kept here so this
		// table mirrors the whole set).
		{"S1", preflight, "§1 preflight", []string{".claude/rules/trellis.md"}},
		{"S1", predicate, "no-op predicate", []string{".claude/rules/trellis.md"}},
		// S2 — the internal overlay.
		{"S2", predicate, "no-op predicate", []string{".trellis/internal/"}},
		{"S2", report, "§5 report", []string{".trellis/internal/"}},
		// S3 — the legacy flat overlay.
		{"S3", predicate, "no-op predicate", []string{"legacy flat"}},
		{"S3", report, "§5 report", []string{".trellis/trellis.md"}},
		// S4 — inline managed blocks across the five documented files, with
		// marker cardinality stated PER FILE (code-review F5: a block in
		// CLAUDE.md and another in AGENTS.md is a legitimate multi-file state,
		// not a duplicate).
		{"S4", preflight, "§1 preflight", []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md", ".github/copilot-instructions.md", ".clinerules", "per file"}},
		{"S4", predicate, "no-op predicate", []string{"managed block"}},
		// S5 — the vendored bundle, with the $HOME carve-out from
		// decision-0070 D6 (that path is the user-scope location, not this
		// project's adoption act).
		{"S5", preflight, "§1 preflight", []string{".claude/skills/trellis/", "$HOME"}},
		{"S5", predicate, "no-op predicate", []string{".claude/skills/trellis/"}},
		{"S5", report, "§5 report", []string{".claude/skills/trellis/"}},
		// S6 — both morph markers, in preflight (before any write — F2) and in
		// the predicate: a marker-only project is removable state, not absence.
		{"S6", preflight, "§1 preflight", []string{".trellis/rollback", "trellis-pre-morph"}},
		{"S6", predicate, "no-op predicate", []string{".trellis/rollback", "trellis-pre-morph"}},
		// spec-adversary N2, both halves: the predicate claimed closure while
		// missing two removable states. An ignore entry is removable residue —
		// a project whose sole leftover is a .trellis/ line in .prettierignore
		// is not "already absent". And the bundle clause needs the $HOME
		// carve-out (decision-0070 D6): without it a $HOME-rooted project can
		// never report already-absent, because the user-scope install at that
		// same path is not this project's artifact.
		{"ignore entries", predicate, "no-op predicate", []string{"ignore entry"}},
		{"S5 ($HOME carve-out)", predicate, "no-op predicate", []string{"$HOME"}},
	} {
		for _, needle := range tc.needles {
			if !strings.Contains(tc.window, needle) {
				t.Errorf("%s: the %s does not name %q — a project in that state would be handled by a document that cannot see it (decision-0073 D1/D3)", tc.state, tc.name, needle)
			}
		}
	}
}

// TestRemoveSkillBundleTransactionPosition guards decision-0073 D3 (first
// bullet): the bundle is deleted behind the existing explicit-confirmation
// gate, positioned WITH OR BEFORE the .trellis/ deletion — never after it, so
// no interruption window leaves the adoption-act artifact as the sole
// survivor. And only the trellis directory: .claude/skills/ is a shared
// directory this operation must never claim.
func TestRemoveSkillBundleTransactionPosition(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	tx := removeSkillSection(t, body, "## 4.", "## 5.")

	iBundle := strings.Index(tx, ".claude/skills/trellis/")
	if iBundle < 0 {
		t.Fatalf("§4 (the only mutating section) never deletes the bundle — the adoption-act artifact survives every removal (decision-0073 D3)")
	}
	iOverlay := strings.Index(tx, "`.trellis/` overlay")
	if iOverlay < 0 {
		t.Fatalf("cannot locate the .trellis/ deletion step in §4 — the ordering assertion would silently not run")
	}
	if iBundle > iOverlay {
		t.Errorf("the bundle is deleted AFTER .trellis/ — an interruption between the two leaves the adoption-act artifact as the sole survivor (decision-0073 D3 pins with-or-before)")
	}
	if !strings.Contains(tx, "never `.claude/skills/`") {
		t.Errorf("§4 must state the file-not-directory boundary: delete only .claude/skills/trellis/, never the shared .claude/skills/")
	}
}

// TestRemoveSkillSurfacesDeclineAndConsumerRows guards decision-0073 D3
// (second bullet, as corrected by decision-0076) and code-review F6: the
// recorded governed = false decline is surfaced BY NAME before any deletion,
// with its consequence (deleting it re-arms the adoption announcement on a
// user-scope machine and returns the project to unadopted; an explicit accept
// governs at 15/15, an ignored prompt does not) — and the consumer's own
// rules.toml rows are shown before the directory holding them is destroyed,
// closing the consent asymmetry where lesser artifacts got consent and the
// consumer's rows did not.
func TestRemoveSkillSurfacesDeclineAndConsumerRows(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	preflight := removeSkillSection(t, body, "## 1.", "## 2.")

	// The count needle tracks LIVE behaviour, not decision-0073's wording: 0073 D3 is
	// append-only and correctly frozen at "14/14", but this skill text tells a user what
	// will happen in their repo now, so it moves with the rule count (decision-0074).
	for _, needle := range []string{"governed = false", "re-arms", "15/15"} {
		if !strings.Contains(preflight, needle) {
			t.Errorf("§1 does not carry %q — the decline artifact must be surfaced by name, with its consequence, before any write (decision-0073 D3)", needle)
		}
	}
	// decision-0076. The count needles above pass either way, so they cannot tell
	// the corrected consequence from the one it replaced. These can: the surviving
	// half is that the announcement RETURNS, and what governs is an answer, never
	// silence. A surface that promises re-governing on an ignored prompt overstates
	// the risk of deleting a decline, which is a consent warning telling a user
	// something untrue about their own repository.
	if !strings.Contains(preflight, "unadopted") {
		t.Errorf("§1 must say deleting the decline returns the project to the UNADOPTED state — the announcement returns, governance does not (decision-0076)")
	}
	if !strings.Contains(preflight, "silence is not an adoption act") {
		t.Errorf("§1 must state that ignoring the re-armed announcement does not govern the project (decision-0076)")
	}
	if strings.Contains(preflight, "one ignored prompt re-governs") {
		t.Errorf("§1 still carries decision-0070 D4's superseded silence-adopts claim, which staleness.sh never implemented (decision-0076)")
	}
	if !strings.Contains(preflight, "Show them the rows") {
		t.Errorf("§1 does not surface the consumer's own rules.toml rows before .trellis/ is deleted — the consent asymmetry of code-review F6")
	}
}

// TestRemoveSkillMorphMarkerHonesty guards decision-0073 AC1's surfacing rule
// and conformance item 5: the morph section presents the markers AS MARKERS,
// never the morph as asserted fact (the tag can outlive the state), says how
// to tell (does the rewritten content still stand?), how to clear a stale tag,
// and — when both markers are absent — says it cannot locate a rollback point
// rather than guess (spec-0004 §2/AC3).
func TestRemoveSkillMorphMarkerHonesty(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	i := strings.Index(body, "## Reversing")
	if i < 0 {
		t.Fatalf("cannot locate the morph section — every assertion below would silently not run")
	}
	morph := body[i:]

	for _, tc := range []struct{ element, needle string }{
		{"the how-to-tell question (marker-as-marker honesty)", "still stand"},
		{"the stale-tag clearing instruction", "git tag -d trellis-pre-morph"},
		{"the both-markers-absent instruction (spec-0004 §2/AC3)", "cannot locate a rollback point"},
	} {
		if !strings.Contains(morph, tc.needle) {
			t.Errorf("the morph section is missing %s — no %q in it", tc.element, tc.needle)
		}
	}
}

// TestRemoveSkillReportIsTrueUnderEitherBundleAnswer guards decision-0073 D3's
// retained-wording rule: whether the vendored skills load is decision-0068's
// open question 5, so the report's bundle wording must be true under either
// answer — the bundle is the adoption-act artifact and the product's skills;
// its hook delivery was measured inert on the headless surface only.
func TestRemoveSkillReportIsTrueUnderEitherBundleAnswer(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	report := removeSkillSection(t, body, "## 5.", "## Reversing")

	for _, needle := range []string{"adoption-act artifact", "measured inert", "headless"} {
		if !strings.Contains(report, needle) {
			t.Errorf("§5's bundle wording is missing %q — the retained-report must be true under either OQ5 answer (decision-0073 D3)", needle)
		}
	}
}

// TestRemoveSkillHandlesManuallyPastedBlocks guards decision-0073 D3
// (code-review F7): the separator-newline instruction assumed every block was
// appended by a setup;
// a manually pasted block may carry no separator newline at all, and removing
// "the one separator newline" over it eats a byte the user wrote.
func TestRemoveSkillHandlesManuallyPastedBlocks(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	staging := removeSkillSection(t, body, "## 2.", "## 3.")
	if !strings.Contains(staging, "manually pasted") {
		t.Errorf("§2's separator phrasing does not cover manually pasted blocks — code-review F7")
	}
}

// TestRemoveSkillFalseSentencesRemoved guards decision-0073 D3/AC1: it pins
// the three recorded false sentences out of the file (code-review F3,
// conformance items 10/11, the decision-adversary's F1) plus §3's undecidable
// provenance predicate (F4). Negative needles quoted verbatim from the
// pre-decision-0073 file.
func TestRemoveSkillFalseSentencesRemoved(t *testing.T) {
	// Whitespace-normalized on purpose: the pre-fix file line-wrapped two of
	// these sentences ("has the rendered\nfile…", "leaves **at least one**\n
	// delivery path intact"), so space-carrying needles over the raw bytes
	// never matched them and those sub-guards were green against the exact
	// defect they pin — demonstrated by reverting the fix (code-review A).
	body := normalizeWS(readFileT(t, removeSkillPath))

	for _, tc := range []struct{ why, needle string }{
		{"frontmatter 'touching nothing else' contradicts the body (conformance item 11)", "touching nothing else"},
		{"install.sh seeds rules.toml, so a fresh curl install carries TWO shapes (conformance item 10 / adversary F1)", "has the rendered file and neither of the other two"},
		{"false for the post-step-2 window — keep the ordering, fix the argument (code-review F3)", "leaves **at least one** delivery path intact"},
		{"provenance is undecidable from disk; consent-on-ambiguity, not a claimed pedigree (code-review F4)", "known to have been added by Trellis"},
	} {
		if strings.Contains(body, tc.needle) {
			t.Errorf("SKILL.md still carries %q — %s", tc.needle, tc.why)
		}
	}

	// The plugin README repeats the same three-shape 'touching nothing else'
	// description of remove; shipping it unchanged would be a fresh private,
	// partial enumeration — the exact class decision-0073 ends.
	readme := normalizeWS(readFileT(t, "../plugins/trellis/README.md"))
	if strings.Contains(readme, "touching nothing else") {
		t.Errorf("plugins/trellis/README.md still describes remove as 'touching nothing else' — stale against decision-0073 D3")
	}
}

// TestRemoveSkillConsentModelIsDefinedOnce guards decision-0073 D3
// (spec-adversary N1/N3/N9): the file used to say both "withheld consent stops
// with the snapshot unchanged" (§1) and "the bundle may be retained by
// withheld consent while the rest proceeds" (§5), with a denied rows-consent
// making the .trellis/ step unexecutable and no defined partial-transaction
// behavior. The consent model must be stated once: integrity problems stop
// everything; a DENIED item-scoped consent narrows the transaction and is
// reported retained; only unobtainable consent stops with the snapshot
// unchanged. And no predicate may decide provenance from bytes: file deletion
// joins the consent list, the ignore evidence is the user's confirmation, and
// unrecognized files inside .trellis/ are surfaced before the wholesale step.
func TestRemoveSkillConsentModelIsDefinedOnce(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	preflight := normalizeWS(removeSkillSection(t, body, "## 1.", "## 2."))
	staging := normalizeWS(removeSkillSection(t, body, "## 2.", "## 3."))
	ignores := normalizeWS(removeSkillSection(t, body, "## 3.", "## 4."))

	for _, tc := range []struct{ element, needle string }{
		{"the stop-everything class (integrity problems)", "stop everything"},
		{"the item-scoped class (a denial narrows, never aborts)", "narrow the transaction"},
		{"file deletion on the consent list (N3: a deletion was gated on provenance)", "becomes empty"},
		{"unrecognized .trellis/ files surfaced before the wholesale step (N9)", "anything unrecognized"},
		// Round 3 (R2): the stop's every-item-retained quantifier must carve
		// out the stopping artifact itself, or the ambiguous category is
		// unreachable from the class that exists to feed it.
		{"the ambiguous carve-out on the stop-everything class (R2)", "reported **ambiguous**"},
		// Round 5 (F1/F2/F3, the convergent blocker): the carve-out must
		// EXCEPT the artifact from the retained quantifier (not coordinate it
		// into both categories), must be conditional (a nobody-to-ask stop
		// has no stopping artifact), and must scope by trigger — a §4
		// mismatch's mandatory category is verification-failed, never
		// ambiguous, or one artifact carries two mandatory categories in an
		// exclusive grammar.
		{"the excepting construction in §1's stop (F2)", "save the stopping artifact"},
		{"the conditional artifact clause — nobody-to-ask has none (F3)", "where the stop has one"},
		{"the mismatch category scoped inside the stop class (F1)", "verification-failed"},
	} {
		if !strings.Contains(preflight, tc.needle) {
			t.Errorf("§1's consent model is missing %s — no %q in it", tc.element, tc.needle)
		}
	}

	// The three provenance-from-bytes predicates (N3), pinned out.
	if strings.Contains(staging, "because Trellis created it") {
		t.Errorf("§2 still gates a file DELETION on creation provenance, which cannot be proven from bytes (N3)")
	}
	if !strings.Contains(staging, "consented in the preflight") {
		t.Errorf("§2's empty-file deletion must route to the preflight consent, not a provenance claim")
	}
	if strings.Contains(ignores, "demonstrably wrote") {
		t.Errorf("§3 still decides provenance from surrounding bytes (N3)")
	}
	if strings.Contains(ignores, "it may be removed") {
		t.Errorf("§3 still carries the operative nondeterministic 'may' (N7) — the empty-file outcome must be deterministic")
	}
	if !strings.Contains(ignores, "never guess") {
		t.Errorf("§3 lost its never-guess rule — the consent routing must keep it")
	}
}

// TestRemoveSkillVerificationAndReportSemantics guards decision-0073 D3
// (spec-adversary N4, prior F7; code-review E/G; N6): a §4 snapshot mismatch
// had no specified behavior and no report bucket, and a preflight stop did not
// say whether it reports. Both are specified now: a mismatch stops the
// transaction where it stands with a verification-failed category, and the §5
// report is owed on every exit. The §4 rationale describes the rendered file
// one way only (import-form — the shape install.sh actually renders), the
// numbered-step collision ("resolved in step 1") is gone, and self-deletion
// at project scope is acknowledged.
func TestRemoveSkillVerificationAndReportSemantics(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	tx := normalizeWS(removeSkillSection(t, body, "## 4.", "## 5."))
	report := normalizeWS(removeSkillSection(t, body, "## 5.", "## Reversing"))

	for _, tc := range []struct{ where, element, needle, window string }{
		{"§4", "the mismatch-stop behavior (N4)", "stops the transaction where it stands", tx},
		{"§4", "the verification-failed category (N4)", "verification-failed", tx},
		{"§4", "the single import-form rendered-file characterization (N6)", "imported activation rows", tx},
		{"§4", "the self-deletion acknowledgement (code-review E)", "this very skill file", tx},
		// Round 3 (spec-adversary R1/R3): the entry threshold must read
		// resolves — a denied consent narrows the transaction, it never bars
		// entry — and verification must be timed per target, immediately
		// before its own write or delete: the verify-after-the-fact reading
		// makes the mismatch branch dead code and destroys a concurrent user
		// edit before reporting clean. Round 5 (F4) widened the wording to
		// cover deletions and, doing so, traded the per-TARGET anchor for
		// "that step" — this document's term of art for the five numbered
		// transaction items, two of which BATCH targets (five instruction
		// files, up to four ignore files). Batch-verify-at-the-step-boundary
		// is still "before the fact" and reopens the very window the next
		// sentence names, so round 6 re-anchors to the individual target and
		// this needle moves with it: the anchor is the target, never the step.
		{"§4", "the resolves entry threshold (R1)", "preflight and consent resolves", tx},
		{"§4", "per-target (not per-step) verification timing (R3/F4/N1)", "immediately before that target's own write or delete", tx},
		{"§4", "deletion steps inside the verification scope (F4)", "or deleted", tx},
		{"§5", "the verification-failed report bucket (N4)", "verification-failed", report},
		{"§5", "the report owed on every exit, a stop included (N4)", "every untouched item retained", report},
		// Round 3 (R2/R4): a stop must leave the ambiguous category reachable
		// (the artifact whose duplicate markers stopped the run is its whole
		// purpose), and the every-exit list must cover the already-absent
		// no-op (whose predicate sentence is itself the report) and name the
		// morph hand-off as an exit, not a resume.
		{"§5", "the reachable ambiguous category on a stop (R2)", "reported **ambiguous**", report},
		{"§5", "the no-op exit in the every-exit list (R4)", "itself that exit's report", report},
		{"§5", "the morph hand-off named as an exit (R4)", "morph hand-off", report},
		// Round 5 (F1/F3/F5 in §5's mirror): the stop sentence must route a
		// first-target mismatch — a pre-write stop under per-target timing —
		// to verification-failed, carry the same conditional
		// stopping-artifact clause as §1, and scope the hand-off exit to a
		// STANDING morph so the stale-tag resume does not falsify the
		// every-exit completeness claim.
		{"§5", "the mismatch category in the stop sentence (F1 mirror)", "first-target mismatch", report},
		{"§5", "the conditional stopping-artifact clause (F3)", "where the stop has one", report},
		{"§5", "the standing-morph scoping of the hand-off exit (F5)", "standing morph", report},
		// Round 6 (code-review M1): the stop-class gate said "before any
		// write", but this round's own F4 ruling is that a write is not a
		// delete — so a §4 mismatch on step 2's first target, after step 1
		// already deleted the rendered file, was literally "a stop before any
		// write" and mandated every-item-retained while §4 mandates completed
		// steps removed. Two statuses for one artifact on a mundane
		// concurrency path.
		{"§5", "the stop gate covering applied deletions too (M1)", "stop before any step is applied", report},
	} {
		if !strings.Contains(tc.window, tc.needle) {
			t.Errorf("%s is missing %s — no %q in it", tc.where, tc.element, tc.needle)
		}
	}
	if strings.Contains(tx, "resolved in step 1") {
		t.Errorf("§4 still says \"resolved in step 1\" inside its own numbered list — that reads as transaction step 1, not the preflight (code-review G)")
	}
	if strings.Contains(tx, "consent succeeds") {
		t.Errorf("§4's threshold still says \"consent succeeds\" — the abort-on-denial model the consent rework replaced (R1)")
	}
}

// TestRemoveSkillMorphReentryAndTrigger guards decision-0073 D3/AC1
// (spec-adversary N5/N8): the both-markers-absent branch was unreachable —
// the morph section was entered only when a marker was present — so its
// cannot-locate-a-rollback-point instruction could never fire; the user's own
// assertion of a morph is its trigger. And a completed git reversal rewrites
// the tree, so the operation must re-enter through a fresh preflight rather
// than continue on stale snapshots and consents.
func TestRemoveSkillMorphReentryAndTrigger(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	i := strings.Index(body, "## Reversing")
	if i < 0 {
		t.Fatalf("cannot locate the morph section — every assertion below would silently not run")
	}
	morph := normalizeWS(body[i:])

	if !strings.Contains(morph, "the user says this project was morphed") {
		t.Errorf("the both-markers-absent branch has no reachable trigger — the user's assertion must open it (N8)")
	}
	if !strings.Contains(morph, "start over from the preflight") {
		t.Errorf("post-reversal re-entry is unspecified — a reversal rewrites the tree, so the fresh preflight must be ordered (N5)")
	}
	// Round 5 (F5): the stale-tag branch shows no options and locates
	// nothing, so it fits neither the exit definition nor a stated resume —
	// its routing was a live guess. Clearing (or declining to clear) a stale
	// tag is not a hand-off: there is no morph left to hand off, and the
	// removal continues as an ordinary one.
	if !strings.Contains(morph, "continues from §2") {
		t.Errorf("the stale-tag branch has no stated routing — with the morph already reversed the removal must continue from §2, not exit (F5)")
	}
}

// TestRemoveSkillStopCategoryBindings guards decision-0073 D3 (code-review M2,
// this round's own lesson). Rounds 4 and 5 were spent making one artifact
// carry exactly one mandatory category per trigger — and the guards that
// landed pinned only the PRESENCE of each category word, so swapping the two
// bindings (an ambiguity stop reported verification-failed, a §4 mismatch
// reported ambiguous) survived every needle in this file. Presence is not
// binding, the same lesson TestRemoveSkillEnumeratesTheRenderedRulesFile
// records for presence-vs-deletion. Each needle below is a complete
// trigger-plus-category phrase, so the swap mutation kills it.
//
// Whitespace-normalized: both sentences wrap, and two of this file's needles
// once shipped never having matched wrapped prose (code-review A).
func TestRemoveSkillStopCategoryBindings(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	preflight := normalizeWS(removeSkillSection(t, body, "## 1.", "## 2."))
	report := normalizeWS(removeSkillSection(t, body, "## 5.", "## Reversing"))

	for _, tc := range []struct{ where, binding, needle, window string }{
		{"§1", "an ambiguity/preflight stop binds ambiguous", "preflight stop's artifact is reported **ambiguous**", preflight},
		{"§1", "a §4 mismatch binds verification-failed", "mismatch's target is reported **verification-failed**", preflight},
		{"§5", "an ambiguity/preflight stop binds ambiguous", "**ambiguous** after an ambiguity or preflight stop", report},
		{"§5", "a first-target mismatch binds verification-failed", "**verification-failed** after a first-target mismatch", report},
	} {
		if !strings.Contains(tc.window, tc.needle) {
			t.Errorf("%s no longer binds its category to its trigger (%s) — no %q in it. The categories are mutually exclusive per trigger; a guard that only checks both words are present passes a swap.",
				tc.where, tc.binding, tc.needle)
		}
	}
}

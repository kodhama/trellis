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

// noOpPredicateWindow returns the no-op predicate clause. Same anchor as
// TestRemoveSkillNoOpPredicateCountsTheRenderedFile ("only when there is no"
// appears in the predicate and nowhere else); the window is wider because the
// predicate now counts every removable D1 state, not three shapes.
func noOpPredicateWindow(t *testing.T, body string) string {
	t.Helper()
	const anchor = "only when there is no"
	j := strings.Index(body, anchor)
	if j < 0 {
		t.Fatalf("cannot locate the no-op predicate clause — this assertion would silently not run")
	}
	end := j + 700
	if end > len(body) {
		end = len(body)
	}
	return body[j:end]
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
// (second bullet) and code-review F6: the recorded governed = false decline is
// surfaced BY NAME before any deletion, with its consequence (deleting the
// recorded decline re-arms the adoption announcement on a user-scope machine;
// one ignored prompt re-governs at 14/14) — and the consumer's own rules.toml
// rows are shown before the directory holding them is destroyed, closing the
// consent asymmetry where lesser artifacts got consent and the consumer's rows
// did not.
func TestRemoveSkillSurfacesDeclineAndConsumerRows(t *testing.T) {
	body := readFileT(t, removeSkillPath)
	preflight := removeSkillSection(t, body, "## 1.", "## 2.")

	for _, needle := range []string{"governed = false", "re-arms", "14/14"} {
		if !strings.Contains(preflight, needle) {
			t.Errorf("§1 does not carry %q — the decline artifact must be surfaced by name, with its consequence, before any write (decision-0073 D3)", needle)
		}
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
	body := readFileT(t, removeSkillPath)

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
	readme := readFileT(t, "../plugins/trellis/README.md")
	if strings.Contains(readme, "touching nothing else") {
		t.Errorf("plugins/trellis/README.md still describes remove as 'touching nothing else' — stale against decision-0073 D3")
	}
}

package main

import (
	"strings"
	"testing"
)

// TestInvariantRulesCoverCatalog guards decision-0026: the always-loaded rules are
// parsed from the bundled catalog, so every assessable invariant must yield a rule.
func TestInvariantRulesCoverCatalog(t *testing.T) {
	rules := invariantRules()
	if len(rules) != len(assessableSlugs) {
		t.Errorf("expected %d invariant rules parsed from the catalog, got %d: %v", len(assessableSlugs), len(rules), sortedKeys(rules))
	}
	// Every pinned slug, not a four-slug spot-check: a count-only assertion passes
	// when one slug is swapped for another, and the spot-check it replaces could not
	// see a new row arrive unparsed.
	for _, slug := range assessableSlugs {
		if rules[slug] == "" {
			t.Errorf("no rule extracted for %s", slug)
		}
	}
}

// TestInvariantPrimaryFailureCoverCatalog guards decision-0031: every assessable
// invariant must yield a primary failure example (its first `violated` case) for the
// always-loaded grounding line.
func TestInvariantPrimaryFailureCoverCatalog(t *testing.T) {
	fails := invariantPrimaryFailure()
	if len(fails) != len(assessableSlugs) {
		t.Errorf("expected %d primary failures parsed from the catalog, got %d: %v", len(assessableSlugs), len(fails), sortedKeys(fails))
	}
	// spot-check: the first violated example for directional-flow, not the second.
	if got := fails["inv-directional-flow"]; !strings.Contains(got, "still being edited") {
		t.Errorf("inv-directional-flow primary failure looks wrong: %q", got)
	}
}

// TestCatalogSlugOrder guards decision-0051 rule 4: the assembled readout
// concatenates fragments "in catalog order" — the order the entries appear in the
// bundled catalog document (structural → operating → floors), which the parser must
// preserve. The set must be exactly the assessable slugs the other parsers cover.
func TestCatalogSlugOrder(t *testing.T) {
	order := catalogSlugOrder()
	if len(order) != len(assessableSlugs) {
		t.Fatalf("expected %d slugs in catalog order, got %d: %v", len(assessableSlugs), len(order), order)
	}
	if order[0] != "inv-directional-flow" {
		t.Errorf("catalog order must open with the structural set (inv-directional-flow), got %s", order[0])
	}
	// Indexed from the end, not hardcoded at 13/14: an operating row inserted before
	// the floors shifts every absolute index, and the previous form made that a
	// silent arithmetic edit on every row addition (decision-0078).
	if tail := order[len(order)-2:]; tail[0] != "floor-transparency" || tail[1] != "floor-intent-gate" {
		t.Errorf("catalog order must close with the floors (floor-transparency, floor-intent-gate), got %v", tail)
	}
	dirs := invariantDirectives()
	for _, slug := range order {
		if dirs[slug] == "" {
			t.Errorf("catalog-order slug %s has no directive — order and directive parsers disagree", slug)
		}
	}
}

// TestDeliberateSuccessionCarriesEntropyLean guards decision-0074, which supersedes
// decision-0052 in part: the entropy lean moves out of inv-self-improvement into
// inv-deliberate-succession, which renders BOTH directions — forward (what a new
// pattern leaves outside it) and backward (what the old base supplies to a gap in the
// new design). 0052's format constraints still hold at the new home.
func TestDeliberateSuccessionCarriesEntropyLean(t *testing.T) {
	d := invariantDirectives()["inv-deliberate-succession"]
	for _, want := range []string{
		"migrate it, or name the exemption and ask",
		"say why it fits the new model or choose differently",
	} {
		if !strings.Contains(d, want) {
			t.Errorf("inv-deliberate-succession directive missing %q: %q", want, d)
		}
	}
	if frag := ruleFragment("inv-deliberate-succession"); strings.Count(frag, "\u2717") != 1 {
		t.Errorf("decision-0052 point 4, still binding at the new home: exactly one failure bullet, got: %q", frag)
	}
	// Scoped to the ENTRY BODY, not the whole catalog. Searching invariantsRef whole
	// proved only that the strings existed somewhere — the assertion passed with both
	// *(structure)* bullets moved back into inv-self-improvement, i.e. with the very
	// move decision-0074 exists to make left un-made. Caught by mutation testing at
	// this change's independent review.
	entry := strings.Join(strings.Fields(catalogEntry("inv-deliberate-succession")), " ")
	if !strings.Contains(entry, "migrate or exempt?") || !strings.Contains(entry, "two conventions in one tree") {
		t.Errorf("the *(structure)* honored/violated pair did not travel to inv-deliberate-succession: %q", entry)
	}
	if !strings.Contains(entry, "two thresholds, silently two different measures") {
		t.Errorf("inv-deliberate-succession missing decision-0074's backward-direction violated example: %q", entry)
	}
	if old := strings.Join(strings.Fields(catalogEntry("inv-self-improvement")), " "); strings.Contains(old, "two conventions in one tree") {
		t.Errorf("the *(structure)* pair is still at its OLD home, inv-self-improvement: %q", old)
	}
}

// catalogEntry returns one catalog entry's raw body — from its `- **`slug`**` line to
// the next entry's — so a test can assert that content lives at a PARTICULAR entry
// rather than anywhere in the document.
func catalogEntry(slug string) string {
	lines := strings.Split(invariantsRef, "\n")
	start := -1
	for i, ln := range lines {
		if strings.HasPrefix(ln, "- **`"+slug+"`**") {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "- **`") {
			return strings.Join(lines[start:i], "\n")
		}
	}
	return strings.Join(lines[start:], "\n")
}

// TestSelfImprovementRevertsToReactiveFace guards decision-0074 point 2: the lean is
// gone from inv-self-improvement, which keeps only friction to root cause to retire.
func TestSelfImprovementRevertsToReactiveFace(t *testing.T) {
	d := invariantDirectives()["inv-self-improvement"]
	for _, gone := range []string{
		"And notice the friction you are about to create",
		"migrate it, or name the exemption and ask",
	} {
		if strings.Contains(d, gone) {
			t.Errorf("inv-self-improvement still carries the entropy lean after decision-0074: %q", gone)
		}
	}
}

// TestInvariantDirectivesCoverCatalog guards decision-0034: every invariant carries an
// imperative, host-agent-facing directive for the block — and it must not leak the
// Trellis-internal codes a host agent can't resolve.
func TestInvariantDirectivesCoverCatalog(t *testing.T) {
	dirs := invariantDirectives()
	if len(dirs) != len(assessableSlugs) {
		t.Errorf("expected %d directives parsed from the catalog, got %d: %v", len(assessableSlugs), len(dirs), sortedKeys(dirs))
	}
	for slug, d := range dirs {
		if d == "" {
			t.Errorf("%s has an empty directive", slug)
		}
		for _, code := range []string{"(A1)", "(A2)", "(A3)", "(A4)", "(B2)", "(C2)", "(D1)", "decision-0", "invariant B"} {
			if strings.Contains(d, code) {
				t.Errorf("directive for %s leaks internal code %q: %s", slug, code, d)
			}
		}
	}
}

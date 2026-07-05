package main

import (
	"strings"
	"testing"
)

// TestInvariantRulesCoverCatalog guards decision-0026: the always-loaded rules are
// parsed from the bundled catalog, so every assessable invariant must yield a rule.
func TestInvariantRulesCoverCatalog(t *testing.T) {
	rules := invariantRules()
	if len(rules) != 14 {
		t.Errorf("expected 14 invariant rules parsed from the catalog, got %d: %v", len(rules), sortedKeys(rules))
	}
	for _, slug := range []string{"inv-directional-flow", "floor-transparency", "floor-intent-gate", "inv-self-improvement"} {
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
	if len(fails) != 14 {
		t.Errorf("expected 14 primary failures parsed from the catalog, got %d: %v", len(fails), sortedKeys(fails))
	}
	// spot-check: the first violated example for directional-flow, not the second.
	if got := fails["inv-directional-flow"]; !strings.Contains(got, "still being edited") {
		t.Errorf("inv-directional-flow primary failure looks wrong: %q", got)
	}
}

// TestInvariantDirectivesCoverCatalog guards decision-0034: every invariant carries an
// imperative, host-agent-facing directive for the block — and it must not leak the
// Trellis-internal codes a host agent can't resolve.
func TestInvariantDirectivesCoverCatalog(t *testing.T) {
	dirs := invariantDirectives()
	if len(dirs) != 14 {
		t.Errorf("expected 14 directives parsed from the catalog, got %d: %v", len(dirs), sortedKeys(dirs))
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

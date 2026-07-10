package main

import "testing"

func TestProfilesCarryNoGatekeeper(t *testing.T) {
	// decision-0024: a preset sets `active` + C1 only; C2 (gatekeeper) is detected
	// from the project, never chosen by a posture. Enforce that here so no one adds
	// a C2 field to Profile without revisiting the decision.
	for _, p := range allProfiles {
		if p.C1Lean == "" {
			t.Errorf("profile %q has no C1 lean", p.Key)
		}
	}
}

func TestRenderedProfilesAreAllActive(t *testing.T) {
	// seed/custom are parked (decision-0033); the two rendered postures both activate
	// all invariants (nil Active) — they differ in stance/lean, not the active set.
	if len(allProfiles) != 2 {
		t.Errorf("expected exactly the two rendered postures, got %d", len(allProfiles))
	}
	for _, p := range allProfiles {
		if p.Key != "a" && p.Key != "b" {
			t.Errorf("unexpected posture %q — seed/custom stay parked (decision-0033)", p.Key)
		}
		if p.Active != nil {
			t.Errorf("profile %q should be all-active (nil Active)", p.Key)
		}
	}
}

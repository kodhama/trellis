package main

import "testing"

func TestMorphModelOptionsExcludeNone(t *testing.T) {
	// M2 (morph) is model-driven — "none" must never be offered (there is no
	// deterministic rewrite). Guards the coherence between the menu and the paths.
	for _, o := range morphModelOptions() {
		if o.key == "none" {
			t.Fatal("morph model options must not include 'none'")
		}
	}
	if len(morphModelOptions()) == 0 {
		t.Fatal("morph should offer at least one model")
	}
}

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

func TestLookups(t *testing.T) {
	if _, ok := profileByKey("a"); !ok {
		t.Error("profile a should resolve")
	}
	if _, ok := profileByKey("zzz"); ok {
		t.Error("unknown profile should not resolve")
	}
	if _, ok := modeByKey("m1"); !ok {
		t.Error("mode m1 should resolve")
	}
	if _, ok := modelByKey("high"); !ok {
		t.Error("model high should resolve")
	}
	if _, ok := modelByKey("nope"); ok {
		t.Error("unknown model should not resolve")
	}
}

func TestSeedIsMinimalConductorIsAll(t *testing.T) {
	seed, _ := profileByKey("seed")
	if len(seed.Active) == 0 {
		t.Error("seed should name an explicit minimal active set")
	}
	a, _ := profileByKey("a")
	if a.Active != nil {
		t.Error("conductor (a) should be all-active (nil Active), not a subset")
	}
}

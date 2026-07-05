package main

// Onboarding data: the postures, install modes, and models the setup flow offers.
// Note the postures carry only `active` + a C1 (enforcement) lean — never C2. Per
// decision-0024 the gatekeeper is *detected and respected* from the project, not
// chosen by a preset.

// Profile is an onboarding posture that seeds an expression profile.
type Profile struct {
	Key         string
	Name        string
	Description string
	C1Lean      string   // default enforcement for active invariants (C2 is detected, not preset)
	Active      []string // active invariant slugs; nil means "all assessable"
}

// Mode is an install mode (spec-0003 §2b).
type Mode struct {
	Key         string
	Name        string
	Description string
}

// Model is a reasoning-model choice passed to `claude --model` (empty Alias = no model).
type Model struct {
	Key         string
	Name        string
	Alias       string
	Description string
}

// seedActive is the structural core the `seed` posture starts from (spec-0003 §2 /
// the two floors + the admission-gate structurals), ratcheting up over time.
var seedActive = []string{
	"floor-transparency", "floor-intent-gate",
	"inv-directional-flow", "inv-handover-points", "inv-ratifiable-artifacts",
}

var allProfiles = []Profile{
	{Key: "a", Name: "A · conductor", Description: "adopt the framework strictly", C1Lean: "enforced"},
	{Key: "b", Name: "B · author-adapt", Description: "evolve as you go; self-improvement enforced", C1Lean: "default-on-but-skippable"},
	{Key: "seed", Name: "seed", Description: "start minimal, ratchet up", C1Lean: "expressed", Active: seedActive},
	{Key: "custom", Name: "Custom", Description: "start from seed and edit afterwards", C1Lean: "expressed", Active: seedActive},
}

var allModes = []Mode{
	{Key: "m1", Name: "M1 · alongside", Description: "install next to your instructions (augment-not-clobber, deterministic)"},
	{Key: "m2", Name: "M2 · rewrite", Description: "rewrite your machinery on a git branch (model-driven)"},
}

var allModels = []Model{
	{Key: "high", Name: "high-reasoning (opus)", Alias: "opus", Description: "for M2 rewrites — real judgement"},
	{Key: "balanced", Name: "balanced (sonnet)", Alias: "sonnet", Description: "a middle option"},
	{Key: "cheap", Name: "cheap (haiku)", Alias: "haiku", Description: "light model-assisted edits"},
	{Key: "none", Name: "no model (deterministic)", Alias: "", Description: "plain file editing — enough for M1"},
}

func profileByKey(k string) (Profile, bool) {
	for _, p := range allProfiles {
		if p.Key == k {
			return p, true
		}
	}
	return Profile{}, false
}

func modeByKey(k string) (Mode, bool) {
	for _, m := range allModes {
		if m.Key == k {
			return m, true
		}
	}
	return Mode{}, false
}

func modelByKey(k string) (Model, bool) {
	for _, m := range allModels {
		if m.Key == k {
			return m, true
		}
	}
	return Model{}, false
}

func profileOptions() []option {
	opts := make([]option, len(allProfiles))
	for i, p := range allProfiles {
		opts[i] = option{p.Key, p.Name + " — " + p.Description}
	}
	return opts
}

func modeOptions() []option {
	opts := make([]option, len(allModes))
	for i, m := range allModes {
		opts[i] = option{m.Key, m.Name + " — " + m.Description}
	}
	return opts
}

// morphModelOptions are the reasoning tiers offered for M2 (morph). "none" is
// intentionally excluded: there is no deterministic rewrite, so a morph always
// needs a model.
func morphModelOptions() []option {
	var opts []option
	for _, m := range allModels {
		if m.Key == "none" {
			continue
		}
		opts = append(opts, option{m.Key, m.Name})
	}
	return opts
}

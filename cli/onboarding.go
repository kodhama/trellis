package main

// Onboarding data: the postures, install modes, and models the setup flow offers.
// Note the postures carry only `active` + a C1 (enforcement) lean — never C2. Per
// decision-0024 the gatekeeper is *detected and respected* from the project, not
// chosen by a preset.

// Profile is an onboarding posture that seeds an expression profile. Name is the full
// label used in the plan/summary; Short is the bold word shown in the picker.
type Profile struct {
	Key         string
	Name        string
	Short       string
	Description string
	C1Lean      string   // default enforcement for active invariants (C2 is detected, not preset)
	Active      []string // active invariant slugs; nil means "all assessable"
}

// Mode is an install mode (spec-0003 §2b). Short is the picker label.
type Mode struct {
	Key         string
	Name        string
	Short       string
	Description string
}

// Model is a reasoning-model choice passed to `claude --model` (empty Alias = no model).
type Model struct {
	Key         string
	Name        string
	Alias       string
	Description string
}

// Two opinionated postures are offered. `seed` (minimal-start) and `custom` (a per-dial
// dialog) are parked until they earn their own UI (decision-0033). Both postures below
// activate all invariants — they differ in stance/lean, not (yet) in the active set;
// the `Active` subset field stays for when `seed` returns.
var allProfiles = []Profile{
	{Key: "a", Name: "A · conductor", Short: "conductor", Description: "hold every invariant firmly — strict, by-the-book", C1Lean: "enforced"},
	{Key: "b", Name: "B · author-adapt", Short: "author-adapt", Description: "same invariants, but adapt as you go — self-improvement leads", C1Lean: "default-on-but-skippable"},
}

var allModes = []Mode{
	{Key: "m1", Name: "M1 · alongside", Short: "alongside", Description: "install next to your instructions (augment-not-clobber, deterministic)"},
	{Key: "m2", Name: "M2 · rewrite", Short: "rewrite", Description: "rewrite your machinery on a git branch (model-driven)"},
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
		opts[i] = option{p.Key, p.Short, p.Description}
	}
	return opts
}

func modeOptions() []option {
	opts := make([]option, len(allModes))
	for i, m := range allModes {
		opts[i] = option{m.Key, m.Short, m.Description}
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
		opts = append(opts, option{m.Key, m.Alias, m.Description})
	}
	return opts
}

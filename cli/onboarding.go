package main

// The postures the payload is rendered for. Note the postures carry only `active` +
// a C1 (enforcement) lean — never C2. Per decision-0024 the gatekeeper is *detected
// and respected* from the project, not chosen by a preset.
//
// Since #120 (decision-0043) these feed the release-time render only: the posture
// question is asked by the plugin setup skill (or answered by the hand-owned
// expression.md frontmatter), never by this binary.

// Profile is a posture that seeds an expression profile. Name is the full label;
// Short is the one-word form.
type Profile struct {
	Key         string
	Name        string
	Short       string
	Description string
	C1Lean      string   // default enforcement for active invariants (C2 is detected, not preset)
	Active      []string // active invariant slugs; nil means "all assessable"
}

// Two opinionated postures are rendered. `seed` (minimal-start) and `custom` (a
// per-dial dialog) are parked (decision-0033). Both postures below activate all
// invariants — they differ in stance/lean, not (yet) in the active set; the `Active`
// subset field stays for when `seed` returns.
var allProfiles = []Profile{
	{Key: "a", Name: "A · conductor", Short: "conductor", Description: "hold every invariant firmly — strict, by-the-book", C1Lean: "enforced"},
	{Key: "b", Name: "B · author-adapt", Short: "author-adapt", Description: "same invariants, but adapt as you go — self-improvement leads", C1Lean: "default-on-but-skippable"},
}

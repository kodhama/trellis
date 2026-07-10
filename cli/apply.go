package main

// The render core of the payload generator (kodhama-0007 rule 1: render once, at
// release). Everything here renders bundle content for payloadFiles(); no code in
// this package writes into a consuming project anymore — the install-time writers
// (the CLI's applyM1/M2) retired with the binary channel (decision-0043, #120), and
// the live writers are mechanical copiers of the pre-rendered payload: the plugin
// setup skill and the documented manual copy path.

import (
	_ "embed"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// The invariant reference shipped in the overlay. Kept in sync from the single
// source in core/ by the generate step below (run `go generate ./...` in cli/).
//
//go:generate cp ../core/catalog/signature-catalog-v1.md assets/invariants.md
//
//go:embed assets/invariants.md
var invariantsRef string

// The M1 overlay writes into CLAUDE.md between these markers only. Everything
// outside them is the host's and is never touched (augment-never-clobber); a
// re-run replaces what is between them (idempotent). The markers are rendered
// into the payload's block files; the copiers paste between them.
const (
	trellisBegin = "<!-- trellis:begin (managed by trellis — edit .trellis/, not this block) -->"
	trellisEnd   = "<!-- trellis:end -->"
)

// strengthLine turns the profile's C1 lean into a plain-language instruction the host
// agent can act on — no jargon (decision-0034).
func strengthLine(c1 string) string {
	switch c1 {
	case "enforced":
		return "**Firmly** — treat these as hard requirements. Follow them as written; don't skip or soften one without the human's explicit say-so."
	case "expressed":
		return "**As guidance** — keep these front of mind and lean toward them; they are the intent, not hard gates."
	default: // default-on-but-skippable
		return "**By default** — follow them unless you have a clear, specific reason not to, and when you deviate say so out loud rather than doing it silently."
	}
}

// governanceHeader is the imperative framing shared by the CLAUDE.md header and the
// inline (AGENTS.md) block: what this is, that the agent must follow it, and how
// strictly — self-contained, no Trellis-internal codes (decision-0034).
func governanceHeader(p Profile) string {
	return "# How to work in this project\n\n" +
		"You are working in a project that follows **Trellis** — a small, load-bearing set of working rules on top of the project's own process. **Follow the rules below as you work here.** They add guardrails; they don't replace this project's own instructions.\n\n" +
		"**How strictly to follow them:** " + strengthLine(p.C1Lean) + "\n"
}

// invariantsTrigger is the always-on pointer at the full reference, phrased as a
// trigger rather than a description (kodhama-0007 rider): the moment to read the
// detail is when a rule is ambiguous or in tension, before deviating.
const invariantsTrigger = "If a rule seems ambiguous, or in tension with this project's own instructions, read its entry in `.trellis/invariants.md` — the description and with/without examples — before deviating."

// rulesBody is the active rules (each an imperative directive + the ✗ failure it
// prevents). Shared by profile.md and the inline block.
func rulesBody(p Profile) string {
	return "## The rules — do these\n\n" +
		"Each is a rule to follow, then the ✗ failure it prevents:\n\n" +
		activeRuleLines(p)
}

// renderInlineBlock is the M1 footprint for instruction files WITHOUT @import support
// (e.g. AGENTS.md): the whole thing is inlined and self-contained. The reasoning +
// examples still live in .trellis/invariants.md, but the block stands on its own.
func renderInlineBlock(p Profile) string {
	return trellisBegin + "\n" +
		governanceHeader(p) + "\n" +
		rulesBody(p) +
		"\n" + invariantsTrigger + " After changing the profile, refresh the overlay — re-copy it from the Trellis payload (repo README, Install).\n" +
		trellisEnd
}

// renderClaudeBlock is the minimal CLAUDE.md footprint: a human-readable line plus
// a native @import of the header. Everything else lives in .trellis/.
func renderClaudeBlock() string {
	return trellisBegin + "\n" +
		"This project follows **Trellis** — working rules you are expected to follow while you work here. They are imported below:\n" +
		"@.trellis/trellis.md\n" +
		trellisEnd
}

// renderHeader is the entry point CLAUDE.md imports: the intro + the governance
// behavior, then it pulls in the profile, the project's own expression, and points
// at the invariant reference. Ordering is rules first, then the expression
// (kodhama-0007 rule 4 via #119 — always-loaded, matching how projects actually
// used it: the hand-authored expression sat below the generated rules).
// expression.md is hand-owned: every writer seeds it from the payload skeleton on
// first run only and never rewrites an existing one.
func renderHeader(p Profile) string {
	return governanceHeader(p) + "\n" +
		"@profile.md\n" +
		"@expression.md\n\n" +
		"---\n" + invariantsTrigger + "\n"
}

// renderExpressionSkeleton is the seed content for `.trellis/expression.md`, the
// bundle's one hand-owned file (kodhama-0007 rule 4 via #119). It is payload content
// like everything else — rendered per posture with the machine-read frontmatter
// pre-filled, so every writer (the setup skill and the manual copy path, via the
// payload's expression-<p>.md) copies it verbatim with nothing left to fill: one
// render, many copiers, applied to the skeleton itself. Seeded when absent, never
// rewritten; the installed file is excluded from install-time checksum verification
// because the project owns it from the first edit on.
func renderExpressionSkeleton(p Profile) string {
	return "---\nprofile: " + p.Key + "\n---\n\n" +
		"# Trellis expression\n\n" +
		"<!-- This file is yours (hand-owned; kodhama-0007 rule 4). Setup seeded it\n" +
		"once and will never rewrite it; it is excluded from install-time checksum\n" +
		"verification. The `profile:` key above (a = conductor · b = author-adapt)\n" +
		"is the only machine-read line — a refresh reads it and asks nothing. Record\n" +
		"below how this project expresses the invariants: dials, mappings, gate\n" +
		"tables. Agents and humans read the body; machinery never parses it. -->\n"
}

// renderProfile is the tunable readout: posture, active invariants, dials. The
// governance behavior lives in the header (single source), not here.
// renderProfile is auto-imported (always in context). Per decision-0026 it carries
// the active invariants as concise *rules* — not just names — so they genuinely
// govern every turn. The full why + examples stay on-demand in invariants.md.
//
// The trailing "(Generated from your profile …)" line is the generated-content
// sentinel the setup skill's overwrite guard keys on (SKILL.md step 2) — keep its
// prefix stable.
func renderProfile(p Profile) string {
	return rulesBody(p) +
		"\n(Generated from your profile — edit `.trellis/` and refresh the overlay (`/trellis:setup`, or the manual copy path) to change these.)\n"
}

// activeRuleLines renders the active invariants as imperative, self-contained directives
// (decision-0034 — no internal codes/slugs), each with its primary ✗ failure for
// grounding (decision-0031). Shared by the profile and the inline overlay block.
func activeRuleLines(p Profile) string {
	dirs := invariantDirectives()
	fails := invariantPrimaryFailure()
	active := p.Active
	if len(active) == 0 { // postures A/B: all assessable invariants
		active = sortedKeys(dirs)
	}
	var b strings.Builder
	for _, slug := range active {
		d := dirs[slug]
		if d == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("- %s\n", d))
		if f := fails[slug]; f != "" {
			b.WriteString(fmt.Sprintf("    ✗ %s\n", f))
		}
	}
	return b.String()
}

// invariantDirectives parses the bundled catalog for each invariant's `directive` — the
// imperative, host-agent-facing instruction shown in the always-loaded block (decision-0034).
func invariantDirectives() map[string]string {
	slugRe := regexp.MustCompile("^- \\*\\*`([a-z][a-z-]*)`\\*\\*")
	dirs := map[string]string{}
	var cur string
	var buf []string
	collecting := false
	flush := func() {
		if cur != "" && len(buf) > 0 {
			dirs[cur] = strings.TrimSpace(strings.Join(buf, " "))
		}
		buf, collecting = nil, false
	}
	for _, ln := range strings.Split(invariantsRef, "\n") {
		if m := slugRe.FindStringSubmatch(ln); m != nil {
			flush()
			cur = m[1]
			continue
		}
		t := strings.TrimSpace(ln)
		switch {
		case strings.HasPrefix(t, "- directive:"):
			buf = []string{strings.TrimSpace(strings.TrimPrefix(t, "- directive:"))}
			collecting = true
		case collecting && strings.HasPrefix(t, "- "): // next field ends the directive
			flush()
		case collecting && t != "":
			buf = append(buf, t)
		}
	}
	flush()
	return dirs
}

// invariantPrimaryFailure parses the bundled catalog for each invariant's FIRST
// `violated` example — the primary failure to avoid, always-loaded as one line of
// grounding under the rule (decision-0031). Curation is by ordering: the example we
// want always-loaded is placed first; only one is pulled, to stay terse.
func invariantPrimaryFailure() map[string]string {
	slugRe := regexp.MustCompile("^- \\*\\*`([a-z][a-z-]*)`\\*\\*")
	tagRe := regexp.MustCompile(`^- \*\([^)]*\)\* (.*)`)
	fails := map[string]string{}
	var cur string
	var buf []string
	inViolated, have := false, false
	flush := func() {
		if cur != "" && len(buf) > 0 && fails[cur] == "" {
			fails[cur] = strings.TrimSpace(strings.Join(buf, " "))
		}
		buf = nil
	}
	for _, ln := range strings.Split(invariantsRef, "\n") {
		if m := slugRe.FindStringSubmatch(ln); m != nil {
			flush()
			cur, inViolated, have = m[1], false, false
			continue
		}
		t := strings.TrimSpace(ln)
		switch {
		case t == "- violated:":
			inViolated = true
		case inViolated && !have && tagRe.MatchString(t):
			buf = []string{tagRe.FindStringSubmatch(t)[1]}
			have = true
		case have && strings.HasPrefix(t, "- "): // the 2nd example or the next field ends it
			flush()
			inViolated = false
		case have && t != "": // a continuation line of the first example
			buf = append(buf, t)
		}
	}
	flush()
	return fails
}

// invariantRules parses the bundled catalog into slug → its one-line `what` rule —
// the single source, so the always-loaded rules can't drift from the reference.
func invariantRules() map[string]string {
	slugRe := regexp.MustCompile("^- \\*\\*`([a-z][a-z-]*)`\\*\\*")
	rules := map[string]string{}
	var cur string
	var buf []string
	collecting := false
	flush := func() {
		if cur != "" && len(buf) > 0 {
			rules[cur] = strings.TrimSpace(strings.Join(buf, " "))
		}
		buf, collecting = nil, false
	}
	for _, ln := range strings.Split(invariantsRef, "\n") {
		if m := slugRe.FindStringSubmatch(ln); m != nil {
			flush()
			cur = m[1]
			continue
		}
		t := strings.TrimSpace(ln)
		switch {
		case strings.HasPrefix(t, "- what:"):
			buf = []string{strings.TrimSpace(strings.TrimPrefix(t, "- what:"))}
			collecting = true
		case collecting && strings.HasPrefix(t, "- "): // next field ends the `what`
			flush()
		case collecting && t != "":
			buf = append(buf, t)
		}
	}
	flush()
	return rules
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

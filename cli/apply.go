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
// detail is when a rule is ambiguous or in tension, before deviating. The reference
// lives under .trellis/internal/ since decision-0051 (the authority split).
const invariantsTrigger = "If a rule seems ambiguous, or in tension with this project's own instructions, read its entry in `.trellis/internal/invariants.md` — the description and with/without examples — before deviating."

// The readout's top matter (decision-0053 point 2): the authority header — the
// eval-tested AUTHORITY_HEADER of research-0012, verbatim except one word adapted
// for the channel split ("inlined" → "loaded": the inline block inlines the rows
// below the rules; the import block loads them below the rules via the managed
// block's @.trellis/rules.toml import) — then the heading, then the live-rows
// preamble (research-0012's header_arm_readout transform, same one-word
// adaptation). The tested wording is the shipped wording (decision-0053 context;
// trellis#170 watch-out). The absence-era assembly preamble and the "(Generated
// from your `rules.toml` …)" footer retired with decision-0053 points 4+5 — no
// shipped text claims refresh-time semantics for rows, and nothing closes the
// readout below its last rule.
const (
	rulesAuthorityHeader = "**Rule activation is governed by `.trellis/rules.toml` (its rows are loaded below the rules):** apply each rule below ONLY if its row says `active = true`. A rule whose row is `active = false` does not apply in this project — do not follow it. The two `floor-` rows apply regardless of their row value.\n"
	rulesReadoutHeader   = rulesAuthorityHeader +
		"\n## The rules — do these\n\n" +
		"Each rule below ends with its row's slug. Whether a rule applies is governed by its row in `.trellis/rules.toml` (see the authority note above; the rows are loaded below the rules). Each is a rule to follow, then the ✗ failure it prevents:\n\n"
)

// The inline managed block: the rows-inlined sandwich (decision-0053 point 2,
// inline channel) — head + the complete readout + the rows section + tail, exactly
// the experiment's annotation/control-arm shape (research-0012 run.sh's overlay
// build), now the shipped shape. The shipped block-inline-<p>.md is the seed-state
// instance; on refresh an inline install rebuilds the rows section from the
// consumer's actual rules.toml (its general update cadence — row edits themselves
// take effect at read time). The head carries the posture's strictness line; the
// tail is posture-independent, so one tail file ships.

// renderInlineBlockHead is everything above the readout: the begin marker and the
// governance header (the one per-posture part).
func renderInlineBlockHead(p Profile) string {
	return trellisBegin + "\n" + governanceHeader(p) + "\n"
}

// renderInlineBlockTail is everything below the rows section: the invariants
// trigger with the live-rows closing sentence (research-0012's header_arm_tail
// wording — the last thing the model reads must not claim refresh-time row
// semantics, decision-0053 point 4), and the end marker.
func renderInlineBlockTail() string {
	return "\n" + invariantsTrigger + " Rule activation follows the rows in `.trellis/rules.toml` directly (see the authority note above).\n" +
		trellisEnd
}

// renderRowsSection wraps a rules.toml's content as the inline block's rows section
// — byte-for-byte the wrapper the experiment tested (research-0012 run.sh's
// sandwich printf). The argument is whichever rules.toml governs the install: the
// posture seed at render time, the consumer's actual file on a refresh re-paste.
func renderRowsSection(toml string) string {
	return "\n## Active rows (`.trellis/rules.toml`)\n\n```toml\n" + toml + "```\n"
}

// renderInlineBlock is the M1 footprint for instruction files WITHOUT @import support
// (e.g. AGENTS.md): the whole thing is inlined and self-contained — the seed-state
// instance of the head + readout + rows + tail sandwich (decision-0053 point 2).
// The reasoning + examples still live in .trellis/internal/invariants.md, but the
// block stands on its own.
func renderInlineBlock(p Profile) string {
	return renderInlineBlockHead(p) + renderRulesReadout() + renderRowsSection(renderRulesToml(p)) + renderInlineBlockTail()
}

// renderClaudeBlock is the minimal CLAUDE.md footprint: a human-readable line plus
// two native block-level @imports — the generated header and the consumer's
// rules.toml (decision-0053 point 2: the empirically tested shape; the rows load
// below the rules, so row edits govern each session at read time with no
// nested-import dependency). expression.md stays retired (decision-0051 amendment):
// a project's governance prose belongs in its own instructions file.
func renderClaudeBlock() string {
	return trellisBegin + "\n" +
		"This project follows **Trellis** — working rules you are expected to follow while you work here. They are imported below:\n" +
		"@.trellis/internal/trellis.md\n" +
		"@.trellis/rules.toml\n" +
		trellisEnd
}

// renderHeader is the entry point the managed block imports (installed at
// .trellis/internal/trellis.md): the intro + the governance behavior, then it pulls
// in its sibling rules.md — the assembled readout — and points at the invariant
// reference. It imports only the sibling (paths resolve relative to the importing
// file; decision-0051 rule 1).
func renderHeader(p Profile) string {
	return governanceHeader(p) + "\n" +
		"@rules.md\n\n" +
		"---\n" + invariantsTrigger + "\n"
}

// catalogSlugOrder parses the bundled catalog for the assessable slugs in document
// order (structural → operating → floors) — the "catalog order" decision-0051
// rule 4 assembles fragments in.
func catalogSlugOrder() []string {
	slugRe := regexp.MustCompile("^- \\*\\*`([a-z][a-z-]*)`\\*\\*")
	var order []string
	for _, ln := range strings.Split(invariantsRef, "\n") {
		if m := slugRe.FindStringSubmatch(ln); m != nil {
			order = append(order, m[1])
		}
	}
	return order
}

// ruleFragment renders one rule's readout entry: the imperative directive plus its
// primary ✗ failure for grounding (decision-0031) — exactly the bytes the complete
// readout carries for that rule, newline-terminated so concatenation is seamless.
// (The per-rule fragment FILES left the shipped payload with decision-0053 — no
// consumer remained once assembly retired — but this render source survives: the
// readout is still built rule by rule from the catalog.) The directive line ends
// with the rule's catalog slug in backticks (rows-as-truth legibility, maintainer
// addendum to decision-0051, load-bearing under live rows: the slug is how a reader
// matches a rules.md bullet ↔ its rules.toml row ↔ its invariants.md entry). This
// narrows decision-0034's no-internal-codes rule deliberately: a slug that resolves
// in two consumer-visible files is a cross-reference, not an unresolvable internal
// code.
func ruleFragment(slug string) string {
	d := invariantDirectives()[slug]
	if d == "" {
		return ""
	}
	s := fmt.Sprintf("- %s `%s`\n", d, slug)
	if f := invariantPrimaryFailure()[slug]; f != "" {
		s += fmt.Sprintf("    ✗ %s\n", f)
	}
	return s
}

// renderRulesReadout is the complete readout (installed at
// .trellis/internal/rules.md on every install — decision-0053 point 1: the one
// readout every install carries; per-row subset assembly retired): the authority
// header + heading + live-rows preamble, then every rule in catalog order, nothing
// below the last rule. Which rules APPLY is decided at read time by the
// rules.toml rows the authority header points at.
func renderRulesReadout() string {
	var b strings.Builder
	b.WriteString(rulesReadoutHeader)
	for _, slug := range catalogSlugOrder() {
		b.WriteString(ruleFragment(slug))
	}
	return b.String()
}

// renderRulesToml renders a posture's rules.toml seed (decision-0051 rule 2:
// posture-as-seed, rows-as-truth): explicit rows, one per assessable catalog slug,
// all active; seeded_from is provenance only; strictness is the one instance-level
// key (rule 7 — no per-row dials until something enforces them). The comments are
// the live-rows wording research-0012's header_arm_toml tested (decision-0053
// point 4): rows govern at read time, and the floor rows apply regardless of their
// value (decision-0051 rule 3, now held by the readout's authority header plus
// setup's loud validation rather than by assembly).
func renderRulesToml(p Profile) string {
	strictness := "adaptive"
	if p.C1Lean == "enforced" {
		strictness = "firm"
	}
	slugs := catalogSlugOrder()
	width := 0
	for _, s := range slugs {
		if len(s) > width {
			width = len(s)
		}
	}
	var b strings.Builder
	b.WriteString("# Rows govern rule activation live (see the authority note in the project instructions).\n\n")
	fmt.Fprintf(&b, "seeded_from = %q  # provenance only — the rows below win if they diverge\n", p.Short)
	fmt.Fprintf(&b, "strictness  = %q  # firm (a·conductor) | adaptive (b·author-adapt)\n", strictness)
	b.WriteString("\n[rules]  # one row per assessable catalog slug (signature-catalog-v1)\n")
	for _, slug := range slugs {
		fmt.Fprintf(&b, "%-*s = { active = true }", width, slug)
		if strings.HasPrefix(slug, "floor-") {
			b.WriteString("  # floor — applies regardless of this row")
		}
		b.WriteString("\n")
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

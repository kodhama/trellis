package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
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
// re-run replaces what is between them (idempotent).
const (
	trellisBegin = "<!-- trellis:begin (managed by trellis — edit .trellis/, not this block) -->"
	trellisEnd   = "<!-- trellis:end -->"
)

// applyM1 performs the deterministic M1 overlay: write the .trellis/ bundle
// (a header, the profile, the invariant reference) and add a minimal import of
// the header to CLAUDE.md. No model — plain file editing.
//
// Layout (imports resolve relative to the importing file — verified against the
// Claude Code docs):
//
//	CLAUDE.md      -> @.trellis/trellis.md      (header, auto-loaded)
//	.trellis/trellis.md -> @profile.md          (profile, auto-loaded)
//	                    -> `.trellis/invariants.md` (backticked = read on demand)
func applyM1(dir string, plan Plan) (string, error) {
	tdir := filepath.Join(dir, ".trellis")
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		return "", fmt.Errorf("creating .trellis/: %w", err)
	}
	bundle := map[string]string{
		"trellis.md":    renderHeader(plan),
		"profile.md":    renderProfile(plan),
		"invariants.md": invariantsRef,
	}
	for name, content := range bundle {
		if err := os.WriteFile(filepath.Join(tdir, name), []byte(content), 0o644); err != nil {
			return "", fmt.Errorf("writing .trellis/%s: %w", name, err)
		}
	}

	target := plan.Target
	if target.Name == "" {
		target = instructionFiles[0] // default CLAUDE.md (e.g. a plan with no target set)
	}
	targetPath := filepath.Join(dir, target.Name)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil { // e.g. .github/ for Copilot
		return "", fmt.Errorf("creating parent dir for %s: %w", target.Name, err)
	}
	existing := ""
	if b, err := os.ReadFile(targetPath); err == nil {
		existing = string(b)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("reading %s: %w", target.Name, err)
	}

	// Files with @import get a one-line import; others get the rules inlined, since
	// there is nothing to import (decision-0029 follow-up).
	block, attach := renderClaudeBlock(), "imports .trellis/trellis.md"
	if !target.Imports {
		block, attach = renderInlineBlock(plan), "inlines the rules (no @import)"
	}
	if err := os.WriteFile(targetPath, []byte(upsertBlock(existing, block)), 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", target.Name, err)
	}

	return fmt.Sprintf("applied (M1 overlay):\n"+
		"  wrote .trellis/{trellis,profile,invariants}.md\n"+
		"  updated %s (%s)\n", target.Name, attach), nil
}

// renderInlineBlock is the M1 footprint for instruction files WITHOUT @import support
// (e.g. AGENTS.md): the governance behavior + active rules are inlined directly, since
// there is nothing to import. The full why + examples still live in .trellis/invariants.md
// (a plain path a reader can open). Re-run `trellis setup` to refresh after a profile change.
func renderInlineBlock(plan Plan) string {
	return trellisBegin + "\n" +
		"This project is governed by **Trellis** (see `.trellis/`). Its rules are inlined here — this file does not support `@import`:\n\n" +
		"**Key behavior:** surface any **human-gated handover performed without its human approval** (invariant B2). Agent-gated handovers proceed silently. Gatekeepers are whatever this project already declares — respected, not imposed (decision-0024).\n\n" +
		"**Active invariants — follow these:**\n" +
		activeRuleLines(plan) +
		"\nFull *why* + with/without examples: `.trellis/invariants.md`. Re-run `trellis setup` to refresh these after a profile change.\n" +
		trellisEnd
}

// upsertBlock replaces the delimited trellis block in content if present, else
// appends it. Content outside the markers is preserved exactly.
func upsertBlock(content, block string) string {
	i := strings.Index(content, trellisBegin)
	j := strings.Index(content, trellisEnd)
	if i >= 0 && j > i {
		return content[:i] + block + content[j+len(trellisEnd):]
	}
	if strings.TrimSpace(content) == "" {
		return block + "\n"
	}
	return strings.TrimRight(content, "\n") + "\n\n" + block + "\n"
}

// renderClaudeBlock is the minimal CLAUDE.md footprint: a human-readable line plus
// a native @import of the header. Everything else lives in .trellis/.
func renderClaudeBlock() string {
	return trellisBegin + "\n" +
		"This project is governed by **Trellis** (see the `.trellis/` folder). Its rules are imported here:\n" +
		"@.trellis/trellis.md\n" +
		trellisEnd
}

// renderHeader is the entry point CLAUDE.md imports: the intro + the governance
// behavior, then it pulls in the profile and points at the invariant reference.
func renderHeader(plan Plan) string {
	return "# Trellis governance\n\n" +
		"This project is supervised by **Trellis**, a governance layer over your existing process: it holds a small set of invariants at the strengths this project has adopted, and otherwise respects your methodology.\n\n" +
		"**Key behavior:** surface any **human-gated handover performed without its human approval** (invariant B2). Agent-gated handovers proceed silently. Gatekeepers are whatever this project already declares — respected, not imposed (decision-0024).\n\n" +
		"## This project's profile\n\n" +
		"@profile.md\n\n" +
		"## Reference\n\n" +
		"The **active rules are in the profile above** — always in context, each with its primary **✗ failure** for grounding, so they govern every turn. The full *why* and the rest of the with/without pairs, plus the invariants not active here, live in `.trellis/invariants.md`; read it for the detail behind a rule.\n"
}

// renderProfile is the tunable readout: posture, active invariants, dials. The
// governance behavior lives in the header (single source), not here.
// renderProfile is auto-imported (always in context). Per decision-0026 it carries
// the active invariants as concise *rules* — not just names — so they genuinely
// govern every turn. The full why + examples stay on-demand in invariants.md.
func renderProfile(plan Plan) string {
	var b strings.Builder
	b.WriteString("# Trellis expression profile\n\n")
	b.WriteString(fmt.Sprintf("- posture: %s — %s\n", plan.Profile.Name, plan.Profile.Description))
	b.WriteString(fmt.Sprintf("- enforcement (C1) lean: `%s`\n", plan.Profile.C1Lean))
	b.WriteString("- gatekeeper (C2): detected from this project, not preset (decision-0024)\n")
	b.WriteString(fmt.Sprintf("- install mode: %s\n", plan.Mode.Name))
	b.WriteString("\n## Active invariants — follow these\n\n")
	b.WriteString("In force for this project, at the enforcement lean above — each with its primary " +
		"**✗ failure to avoid** inline for grounding. The full *why* and the rest of the with/without " +
		"pairs (and the invariants not active here) are in `.trellis/invariants.md`.\n\n")
	b.WriteString(activeRuleLines(plan))
	b.WriteString("\nEdit this file to tune the profile; `CLAUDE.md` imports `.trellis/trellis.md`, which imports this.\n")
	return b.String()
}

// activeRuleLines renders the active invariants as "- **slug** — <rule>" lines from the
// bundled catalog (decision-0026: rules always in context). Shared by the profile and
// the inline overlay block.
func activeRuleLines(plan Plan) string {
	rules := invariantRules()
	fails := invariantPrimaryFailure()
	active := plan.Profile.Active
	if len(active) == 0 { // postures A/B: all assessable invariants
		active = sortedKeys(rules)
	}
	var b strings.Builder
	for _, slug := range active {
		if r := rules[slug]; r != "" {
			b.WriteString(fmt.Sprintf("- **%s** — %s\n", slug, r))
		} else {
			b.WriteString(fmt.Sprintf("- **%s**\n", slug))
		}
		if f := fails[slug]; f != "" {
			b.WriteString(fmt.Sprintf("    ✗ %s\n", f)) // the primary failure to avoid (decision-0031)
		}
	}
	return b.String()
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

func activeList(plan Plan) string {
	if len(plan.Profile.Active) == 0 {
		return "all assessable invariants"
	}
	return strings.Join(plan.Profile.Active, ", ")
}

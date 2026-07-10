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
//	                    -> @expression.md       (the project's hand-owned expression, auto-loaded;
//	                       kodhama-0007 rule 4 — seeded from the payload skeleton if absent,
//	                       never overwritten)
//	                    -> `.trellis/invariants.md` (backticked = read on demand)
func applyM1(dir string, plan Plan) (string, error) {
	tdir := filepath.Join(dir, ".trellis")
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		return "", fmt.Errorf("creating .trellis/: %w", err)
	}

	// Read before we clobber: profile.md has no begin/end markers like the target
	// instructions file does, so any hand-appended content here is about to be
	// silently destroyed by the write below. Warn before that happens (kodhama/trellis#112).
	warning := warnOrphanedProfileContent(tdir)

	bundle := map[string]string{
		"trellis.md":    renderHeader(plan),
		"profile.md":    renderProfile(plan),
		"invariants.md": invariantsRef,
		// The version that generated this overlay — the D1 staleness marker `trellis
		// status` reads (decision-0035). Kept out of the rendered content so the repo's
		// sync-guard diffs behavior, not the build number.
		"version": version + "\n",
	}
	for name, content := range bundle {
		if err := os.WriteFile(filepath.Join(tdir, name), []byte(content), 0o644); err != nil {
			return "", fmt.Errorf("writing .trellis/%s: %w", name, err)
		}
	}

	// Seed the hand-owned declaration file on first run only (kodhama-0007 rule 4 via
	// #119): the header written above imports @expression.md, and this writer stays
	// live until the CLI channel retires (#120), so it must seed like any other copier
	// — a verbatim copy of the payload skeleton, never a rewrite of an existing one.
	seeded := ""
	expPath := filepath.Join(tdir, "expression.md")
	if _, err := os.Stat(expPath); os.IsNotExist(err) {
		if werr := os.WriteFile(expPath, []byte(renderExpressionSkeleton(plan)), 0o644); werr != nil {
			return "", fmt.Errorf("seeding .trellis/expression.md: %w", werr)
		}
		seeded = "  seeded .trellis/expression.md (hand-owned — record the project's expression there)\n"
	} else if err != nil {
		return "", fmt.Errorf("checking .trellis/expression.md: %w", err)
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

	return warning + fmt.Sprintf("applied (M1 overlay):\n"+
		"  wrote .trellis/{trellis,profile,invariants}.md\n"+
		seeded+
		"  updated %s (%s)\n", target.Name, attach), nil
}

// profileGeneratedSentinel is the exact trailing line renderProfile always emits.
// The generator never writes anything after it, so content found past it in an
// existing profile.md is unambiguously hand-authored, not generated.
const profileGeneratedSentinel = "(Generated from your profile — edit `.trellis/` and re-run `trellis setup` to change these.)\n"

// warnOrphanedProfileContent detects hand-appended content after the generated
// sentinel in an existing .trellis/profile.md and returns a warning describing what
// is about to be overwritten, or "" if the file is absent or carries none.
//
// profile.md is a pure generated snapshot (decision-0035): unlike the target
// instructions file, which is only ever touched between its trellis:begin/end
// markers (upsertBlock), profile.md has no such markers and applyM1 always rewrites
// it whole. That's silent data loss for anyone who hand-appends project-specific
// content below the generated block — which has happened for real, twice: in this
// repo (#106 → reverted by #111) and downstream (kodhama/trellis#112). This does not
// preserve the content — decision-0035 requires profile.md stay byte-identical to
// the generator's output — it only makes the loss visible before it happens.
func warnOrphanedProfileContent(tdir string) string {
	b, err := os.ReadFile(filepath.Join(tdir, "profile.md"))
	if err != nil {
		return ""
	}
	i := strings.Index(string(b), profileGeneratedSentinel)
	if i < 0 {
		return ""
	}
	after := strings.TrimSpace(string(b)[i+len(profileGeneratedSentinel):])
	if after == "" {
		return ""
	}
	// The move-it-here pointer names .trellis/expression.md — the bundle's one
	// hand-owned file (kodhama-0007 rule 4, #119). This supersedes PR #114's original
	// wording, which pointed at the instructions file (e.g. CLAUDE.md) before the
	// hand-owned home existed.
	return "⚠ .trellis/profile.md has hand-authored content below the generated block, about to be overwritten and lost:\n\n" +
		after + "\n\n" +
		"profile.md is a pure generated snapshot (decision-0035) — trellis setup always rewrites it whole. Move project-specific content into the body of .trellis/expression.md instead (the bundle's hand-owned file, kodhama-0007 rule 4): setup seeds it once, never rewrites it, and the overlay header keeps it always-loaded.\n\n"
}

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
func governanceHeader(plan Plan) string {
	return "# How to work in this project\n\n" +
		"You are working in a project that follows **Trellis** — a small, load-bearing set of working rules on top of the project's own process. **Follow the rules below as you work here.** They add guardrails; they don't replace this project's own instructions.\n\n" +
		"**How strictly to follow them:** " + strengthLine(plan.Profile.C1Lean) + "\n"
}

// invariantsTrigger is the always-on pointer at the full reference, phrased as a
// trigger rather than a description (kodhama-0007 rider): the moment to read the
// detail is when a rule is ambiguous or in tension, before deviating.
const invariantsTrigger = "If a rule seems ambiguous, or in tension with this project's own instructions, read its entry in `.trellis/invariants.md` — the description and with/without examples — before deviating."

// rulesBody is the active rules (each an imperative directive + the ✗ failure it
// prevents). Shared by profile.md and the inline block.
func rulesBody(plan Plan) string {
	return "## The rules — do these\n\n" +
		"Each is a rule to follow, then the ✗ failure it prevents:\n\n" +
		activeRuleLines(plan)
}

// renderInlineBlock is the M1 footprint for instruction files WITHOUT @import support
// (e.g. AGENTS.md): the whole thing is inlined and self-contained. The reasoning +
// examples still live in .trellis/invariants.md, but the block stands on its own.
func renderInlineBlock(plan Plan) string {
	return trellisBegin + "\n" +
		governanceHeader(plan) + "\n" +
		rulesBody(plan) +
		"\n" + invariantsTrigger + " Re-run `trellis setup` after changing the profile.\n" +
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
// first run only (renderExpressionSkeleton) and never rewrites an existing one.
func renderHeader(plan Plan) string {
	return governanceHeader(plan) + "\n" +
		"@profile.md\n" +
		"@expression.md\n\n" +
		"---\n" + invariantsTrigger + "\n"
}

// renderExpressionSkeleton is the seed content for `.trellis/expression.md`, the
// bundle's one hand-owned file (kodhama-0007 rule 4 via #119). It is payload content
// like everything else — rendered per posture with the machine-read frontmatter
// pre-filled, so every writer (applyM1 here, the setup skill via the payload's
// expression-<p>.md) copies it verbatim with nothing left to fill: one render, many
// copiers, applied to the skeleton itself. Seeded when absent, never rewritten; the
// installed file is excluded from install-time checksum verification because the
// project owns it from the first edit on.
func renderExpressionSkeleton(plan Plan) string {
	return "---\nprofile: " + plan.Profile.Key + "\n---\n\n" +
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
func renderProfile(plan Plan) string {
	return rulesBody(plan) +
		"\n(Generated from your profile — edit `.trellis/` and re-run `trellis setup` to change these.)\n"
}

// activeRuleLines renders the active invariants as imperative, self-contained directives
// (decision-0034 — no internal codes/slugs), each with its primary ✗ failure for
// grounding (decision-0031). Shared by the profile and the inline overlay block.
func activeRuleLines(plan Plan) string {
	dirs := invariantDirectives()
	fails := invariantPrimaryFailure()
	active := plan.Profile.Active
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

func activeList(plan Plan) string {
	if len(plan.Profile.Active) == 0 {
		return "all assessable invariants"
	}
	return strings.Join(plan.Profile.Active, ", ")
}

---
id: research-0001
type: research-note
status: draft
depends_on: [decision-0006, invariants-v0]
owner: gundi
date: 2026-06-29
---

# Research 0001 — Target landscape (Step 0: target identification)

> **Method.** Deep-research pass: 6 search angles → 25 sources fetched → 118 claims →
> top 25 verified by 3-vote adversarial verification (need 2/3 refutes to kill) → 25
> confirmed, 0 killed → synthesized to 10 findings. Confidence tags below are the
> verification verdicts, not assertions. Adoption numbers are **point-in-time snapshots**
> (fast-moving) and stars are an imperfect proxy. Full transcript: workflow `wf_c9d9078c-ddf`.

## Purpose

Identify the most worthwhile **AI-agentic-dev methodologies** to deep-test against
`invariants-v0` in Step 1 (decision `0006`). Not a ranking of "best tool" — a ranking of
*most informative to test*: real, distinct stages/gates/artifacts/roles, decent maturity,
and archetype diversity.

## The landscape, by archetype

### A. Spec-driven (spec-before-code, spec as versioned artifact)
- **GitHub Spec Kit** — `confidence: high`. GitHub's official open-source Python CLI for
  Spec-Driven Development. Explicit sequenced slash-command workflow: `constitution →
  specify → plan → tasks → implement → converge`, plus optional gates `clarify` (pre-plan),
  `analyze` (read-only cross-artifact consistency post-tasks), `checklist`. **Strict
  left-to-right dependency order** (can't plan before specify, can't implement before
  tasks). ~93k→116k stars (May→Jun 2026), v0.11.9, 30+ supported agents. *Adaptable shell,
  opinionated flow.* Sources: github.com/github/spec-kit; github.github.com/spec-kit.
- **OpenSpec** (Fission AI) — `confidence: high`. Lightweight SDD; captures what/why/how
  before code. Git-versioned `openspec/` with two artifact types: a **living spec library**
  (`specs/`, by capability) and a **changes system** (`changes/`, proposals with
  design/tasks/spec deltas). Four-stage flow (`new → ff → apply → archive`; core profile:
  propose/explore/apply/sync/archive). MIT, 21 tools, ~27k→57.6k stars in <6 months.
  Sources: openspec.pro; github.com/Fission-AI/OpenSpec; YC launch.

### B. PRD/story-driven multi-agent lifecycle (role-decomposed)
- **BMAD-METHOD** — `confidence: high`. 12+ named domain-expert roles (PM, Architect,
  Developer, UX, …) guiding brainstorming→deployment; 34+ core workflows producing PRDs,
  UX specs, PRFAQs; "Party Mode" multi-agent collaboration; extensibility modules. ~49.8k
  stars, 37 releases, v6.9.0 (Jun 2026). *Opinionated, rich.* Source:
  github.com/bmad-code-org/bmad-method.

### C. Standards-injection / spec-shaping config
- **Agent OS** (Brian Casel / Builder Methods) — `confidence: high`. Four capabilities:
  `Discover Standards → Deploy Standards → Shape Spec → Index Standards` — extracts codebase
  conventions and injects them into spec-driven builds (Claude Code, Cursor, Codex). ~5k
  stars, v3.0.0 (Jan 2026). Source: github.com/buildermethods/agent-os.

### D. Vendor/IDE-native orchestration primitives
- **Anthropic Claude Agent SDK / Claude Code subagents** — `confidence: high`. Agent SDK
  prescribes a three-stage loop `gather context → take action → verify work → repeat`;
  multi-agent system documents an **orchestrator-worker** pattern (lead agent delegates to
  parallel subagents, each given objective/output-format/tool-guidance/task-boundaries).
  Claude Code subagents = Markdown + YAML frontmatter, isolated context, scoped tools.
  *Caveat: framed as **patterns/guidance**, softer than a fixed methodology — tests
  differently than a rigid staged framework.* **Note: this is the substrate Bonsai itself
  runs on — testing it is partly testing our own foundation.** Sources: anthropic.com
  engineering posts; code.claude.com/docs/sub-agents.
- **claude-sub-agent** (zhsama) — `confidence: high`. Community SDD on Claude Code subagents
  with **numeric quality gates**: Planning →95%, Development →80%, Validation →85%. *Low
  maturity (592 stars, 10 commits, dormant ~10mo)* — structurally interesting, low priority.
- **VoltAgent/awesome-claude-code-subagents** — `confidence: high`. A *catalog* (10 role
  categories incl. Meta & Orchestration), not a methodology. Reference for the design space.

### Controls (non-AI — flagged, not primary targets)
- **Shape Up** — `confidence: high`. Pitch → betting-table gate → fixed 6-week cycle →
  2-week cool-down. Stable author-canonical baseline. Source: basecamp.com/shapeup.
- **Scrum** — `confidence: high`. Declared "immutable" yet "purposefully incomplete."
  Prescriptive baseline to contrast against adaptable AI methods. Source: scrumguides.org.

## ⚠ Gaps (loud failure — invariant 7: do not rank on no data)

- **Kiro (Amazon)** — **required candidate, ZERO surviving verified claims** this batch.
  It is a real spec-driven AI IDE, but its stages/artifacts/gates/maturity are *unconfirmed
  here.* Needs a dedicated verification pass before ranking.
- **spec-swarm** — **required candidate, ZERO verified claims.** Existence as a distinct,
  maintained framework is *unconfirmed.* Verify before ranking.
- **Thin archetypes** (in-scope for discovery, no verified claims): IDE rules/config packs
  (**Cursor** rules/workflows, **Windsurf**, **Aider**, **Cline/Roo**) and autonomous
  commercial agents (**Devin/Cognition**, **Tessl**, **Conductor**). The rules/config-pack
  and autonomous-agent archetypes are currently underrepresented.

## Ranked shortlist for Step 1 (recommendation)

**Tier 1 — test first (distinct archetypes, real structure, strong maturity):**
1. **GitHub Spec Kit** — spec-driven; its strict stage ordering is a live instance of
   invariant 1 (directional handover) with explicit gates (invariant 2).
2. **BMAD-METHOD** — PRD/role-driven; named roles are a live instance of invariant 4
   (authority split / who-owns-what).
3. **OpenSpec** — lightweight spec-driven; versioned specs + changes + archive is a live
   instance of invariant 3 (auditable archive).

**Tier 2 — test next (archetype coverage):**
4. **Agent OS** — standards-injection (a different shape; tests whether the gate fits
   non-staged config systems).
5. **Anthropic Agent SDK / Claude Code subagents** — SDK/IDE-native orchestration; also the
   self-reference case (our own substrate). Its `verify work` step is a live instance of
   invariant 5 (independent verification) worth probing.

**Controls — map "which invariant subset do they satisfy" (per decision `0006`), not pass/fail:**
- Shape Up, Scrum.

**Before finalizing Step 1:** run a targeted verification on **Kiro** and **spec-swarm**,
and one pass on the **rules/config-pack** archetype (Cursor) so all archetypes are
represented or explicitly excluded with reason.

## Acceptance criteria

- Every named target carries a confidence tag and a primary source.
- Required candidates that could not be verified are flagged as gaps, not dropped silently.
- The shortlist spans distinct archetypes, so Step 1 tests the invariants against variety,
  not five clones of one shape.

## Open questions (carried into Step 1)

- **Kiro / spec-swarm:** dedicated verification — real stages/artifacts/gates + maturity?
- **Archetype coverage:** do we need a rules/config-pack (Cursor) and an autonomous-agent
  (Devin) target for the gate to be tested against the full design space?
- **Preview signal (to confirm in Step 1, currently `inferred` not verified):** several
  targets already exhibit invariant correlates — Spec Kit↔1/2, OpenSpec↔3, BMAD↔4,
  claude-sub-agent↔2, Agent-SDK `verify`↔5. Does this survive close gate-testing, or do the
  correlates break under detail? (This is exactly Step 1's job, and seeds Step 2.)
- **Soft vs hard methodologies:** Anthropic's "patterns/guidance" framing means it isn't a
  rigid staged process — does the admission gate even apply to pattern-level guidance, or
  only to methodologies with explicit stages? (A real boundary question for the gate.)

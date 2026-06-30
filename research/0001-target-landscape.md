---
id: research-0001
type: research-note
status: ratified
ratified: 2026-06-30
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
>
> **Addendum (targeted gap-fill, 2026-06-29):** SpecSwarm, Kiro, Cursor, Devin added via
> direct primary-source fetch + cross-check (lighter than the 3-vote pass); confidence
> tagged per source quality (`medium` where only a vendor self-describes).

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
- **Kiro (AWS)** — `confidence: high` (vendor + independent). Agentic IDE/CLI/web for
  spec-driven dev. Three artifacts **`requirements.md → design.md → tasks.md`**, then
  implementation, then **verification via property-based tests** (catch edge cases unit
  tests miss). **Approval gates** between phases (Requirements-First / Design-First variants;
  a "Quick Plan" *skips* gates); tasks run as a dependency-graph in concurrent "waves." AWS
  frames the docs as "guardrails." Moderately opinionated; supports ACP / AGENTS.md / MCP.
  Production, paid (Pro Max $100/mo). Sources: kiro.dev; kiro.dev/docs/specs; AWS Builder
  Center.
- **SpecSwarm** (Marty Bonacci) — `confidence: high`. Spec-driven *for Claude Code* with
  multi-agent orchestration. 5-command core loop `init → build → fix → modify → ship`
  (build = spec → plan → tasks → implement → quality-score). Quality-threshold **gate**
  (default 80/100), per-task **verifier** subagent, adversarial **`spec-mentor`**
  spec-vs-code check, version-controlled specs in `.specswarm/`, a constitution with
  warn/block severities. MIT; 22 releases (v7.11.0, May 2026), 219 commits, ~63★ — *active
  but low adoption.* (`specswarm.com` returned HTTP 403; data from the repo.) Source:
  github.com/MartyBonacci/specswarm.

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

### E. Rules/config packs (substrate, not a process)
- **Cursor Rules** — `confidence: high`. System-level instructions injected into agent
  context (`.mdc` files in `.cursor/rules`, user rules, `AGENTS.md`). Explicitly *"persistent
  configuration and instructions, not a methodology"* — no workflow stages, gates, or
  artifact transitions. The boundary case / **negative control** for the admission gate.
  Source: cursor.com/docs.

### F. Autonomous agents (open-ended)
- **Devin (Cognition)** — `confidence: medium` (vendor self-description only). "The AI
  software engineer" — autonomous, plans and executes end-to-end (migrations, on-call),
  *open-ended rather than a fixed methodology.* Human oversight is a **merge-time approval**
  ("a human is kept in the loop just to… approve Devin's changes"). Production at enterprise
  scale (Nubank). Source: devin.ai.

### Controls (non-AI — flagged, not primary targets)
- **Shape Up** — `confidence: high`. Pitch → betting-table gate → fixed 6-week cycle →
  2-week cool-down. Stable author-canonical baseline. Source: basecamp.com/shapeup.
- **Scrum** — `confidence: high`. Declared "immutable" yet "purposefully incomplete."
  Prescriptive baseline to contrast against adaptable AI methods. Source: scrumguides.org.

## Gaps — status after targeted follow-up

- **Kiro (AWS)** — ✅ resolved (now under archetype A; vendor + independent sources).
- **SpecSwarm** — ✅ resolved (now under archetype A; confirmed via GitHub repo).
  `specswarm.com` returned HTTP 403, so the marketing site is unverified — repo data used.
- **Cursor / Devin** — ✅ added (archetypes E, F), filling the config-pack and autonomous
  archetypes.
- **Still unsurveyed (low priority):** Windsurf, Aider, Cline/Roo, Tessl, Conductor — further
  instances of archetypes already represented by Cursor and Devin. Survey only if Step 1
  reveals a coverage gap.

## Resolution — does the gate apply to pattern-level guidance? (maintainer: yes, if it carries guardrails)

The Cursor ↔ Devin ↔ Agent-SDK contrast settles the boundary question Step 0 raised. The
admission gate keys on the **invariant properties acting as guardrails**, *not* on the
presence of fixed/named steps:

- **Agent SDK** — pattern-level (`gather → act → verify → repeat`), no rigid stages, but its
  `verify` embeds invariant 5 as a guardrail → **admissible**.
- **Devin** — open-ended, but a human merge-time approval embeds part of invariant 4 →
  **partial** (likely fails directional-flow / gate-at-every-handover).
- **Cursor Rules** — embeds *no* flow or gate, only injected standards → **fails** — rightly
  so. It fails for lacking guardrails, not for "being patterns/config."

**Candidate refinement (for Step 2):** reword invariant 1's gate face — "directional flow
exists" should *not* require fixed or named stages, only one-way decreasing-ambiguity flow.
Patterns that carry the guardrails pass; pure config that carries none does not.

## Ranked shortlist for Step 1 (recommendation)

Chosen for **archetype + boundary coverage** — including cases that *should* fail — so Step 1
tests the invariants against variety, not five clones of one shape.

**Tier 1 — staged methodologies, real structure + maturity:**
1. **GitHub Spec Kit** — spec-driven; strict stage ordering ↔ invariant 1, explicit gates ↔ 2.
2. **Kiro (AWS)** — commercial IDE spec-driven; `requirements→design→tasks→verify` with
   **approval gates** ↔ invariants 1, 2, 4; property-based verification ↔ 5.
3. **BMAD-METHOD** — PRD/role-driven; named roles ↔ invariant 4.
4. **OpenSpec** — lightweight spec-driven; versioned specs/changes/archive ↔ invariant 3.

**Tier 2 — boundary & archetype cases (often the most informative tests):**
5. **SpecSwarm** — Claude-Code-native; structurally the richest invariant-correlate
   (quality-gate ↔ 2, `spec-mentor` adversarial check ↔ 5, versioned specs ↔ 3, constitution
   ↔ 4) — *low adoption; test for **coverage**, not popularity.*
6. **Anthropic Agent SDK / Claude Code subagents** — pattern-level (not staged); the
   self-substrate case. Tests whether the gate admits guardrail-bearing *patterns* (see
   resolution above). `verify work` ↔ invariant 5.
7. **Devin (Cognition)** — autonomous, open-ended; only a merge-time human gate. Tests how
   the gate handles agents with no staged directional flow.
8. **Cursor Rules** — **negative control:** pure config, *explicitly not a methodology*.
   Should **fail** the gate — confirming the gate discriminates.

**Controls — map which invariant subset they satisfy (decision `0006`), not pass/fail:**
- Shape Up, Scrum.

**Still unsurveyed (optional, low priority):** Windsurf, Aider, Cline/Roo (more config-pack
shapes), Tessl, Conductor (more autonomous agents) — fill only if Step 1 finds a gap.

## Acceptance criteria

- Every named target carries a confidence tag and a primary source.
- Required candidates that could not be verified are flagged as gaps, not dropped silently.
- The shortlist spans distinct archetypes, so Step 1 tests the invariants against variety,
  not five clones of one shape.

## Open questions (carried into Step 1)

- **Preview signal (tag `inferred`, to confirm in Step 1):** targets already exhibit
  invariant correlates — Spec Kit↔1/2, Kiro↔1/2/4/5, OpenSpec↔3, BMAD↔4, SpecSwarm↔2/3/4/5,
  Agent-SDK `verify`↔5. Does this survive close gate-testing, or do correlates break under
  detail? (Step 1's job; seeds Step 2.)
- **Negative control:** does Cursor Rules cleanly **fail** the gate, and Devin land
  *partial*? If a pure-config pack somehow "passes," the gate is too loose.
- **Skippable gates:** Kiro's "Quick Plan" and SpecSwarm's `--quick` *skip* gates/approvals.
  Does an **optional** gate still satisfy invariant 2 ("a gate at every handover"), or does
  skippability violate it? (A real test of how strict the gate invariant is.)

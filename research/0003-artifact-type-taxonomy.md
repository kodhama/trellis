---
id: research-0003
type: research-note
status: ratified
ratified: 2026-06-30
depends_on: [research-0001, research-0002, invariants-v1]
owner: gundi
---

# Research 0003 — Artifact types across methodologies (are types invariant?)

> **Method & honesty.** This is a **consolidation of type data already captured incidentally**
> in `research-0001`/`research-0002` — **not** a dedicated research pass. Steps 0–1 targeted
> the *gate invariants*, not artifact types, so every row below is `confidence: inferred`
> (from each framework's primary docs as cited in `0001`/`0002`), not independently verified
> for completeness. The gap is logged in Open questions.

## The question

The maintainer asked: are **artifact types** (decision, spec, research, feedback, …) invariant
across methodologies — should Bonsai aim for a fixed type *set*?

## What the methodologies use (from `0001`/`0002`)

| Methodology | Artifact types it defines |
|---|---|
| **Spec Kit** | constitution · spec · plan · tasks |
| **Kiro** | requirements · design · tasks |
| **BMAD** | PRD · architecture · UX-spec · story · PRFAQ (+ roles) |
| **OpenSpec** | specs (living library) · changes (proposals: proposal/design/tasks/deltas) |
| **SpecSwarm** | spec · plan · tasks · constitution · verification |
| **Agent OS** | standards · spec |
| **math-quest** | decision/adr · spec · discovery · feedback · rubric |

## Finding

**Types are *not* invariant the way the process properties are.** What recurs is the
*function*, under wildly varying names and granularity:

- **Spec / contract / requirements** — *near-universal* (spec, requirements, PRD, changes,
  spec). The strongest recurring function: *a ratified contract before build.*
- **Plan / design** — *wide* (plan, design, architecture) — sometimes merged into the spec.
- **Tasks** — common in spec-driven ones.
- **Standing principles** (constitution / standards) — several (Spec Kit, SpecSwarm, Agent OS).
- **Decision record** — the *concept* recurs, but as an explicit *type* it's rarer (math-quest).
- **Research / discovery** and **feedback** — soft, mostly math-quest.

The same function appears as `spec` / `requirements` / `PRD` / `changes` — different names,
different splits. So the **recurring thing is the artifact *contract* (frontmatter +
lifecycle + a rubric), not a fixed type *enum*.**

## Conclusion (consistent with `decision-0003`, methodology-agnostic)

- **The contract is the invariant; `type` is an *open* field the methodology populates.**
  Bonsai should **not** impose a closed type set — it accepts whatever types a methodology
  defines and requires only that each satisfies the contract (and carries a rubric).
- There *is* a **soft recurring spine** — *spec/contract · plan · tasks · standing-principles ·
  decision-record* — worth shipping as **seed/default types**, while staying extensible.
- This refines `spec-0001`: make `type` open-with-seeds + a `scope` tag, not a closed enum.

## Acceptance criteria

*(research-note: Open questions + sources/confidence; no acceptance-criteria gate per the
per-type rule.)*

## Open questions

- **The gap:** type-recurrence was never *systematically* studied — this is consolidated
  incidental data (`inferred`). A dedicated pass (full artifact inventory per methodology,
  incl. ones not yet examined: Cursor/Devin/Claude-SDK have *no* artifact types — pattern-
  level) would either confirm the soft spine or refine it. Run only if the seed-type set
  proves contentious.
- **Is "standing principles" (constitution/standards) a seed type, or Bonsai-product content
  (like `invariant-set`)?** It straddles — a project's constitution is per-project; Bonsai's
  invariants are product. Decide when seeding types.
- **Does the soft spine hold for non-spec-driven archetypes** (multi-agent/BMAD roles,
  autonomous agents)? BMAD's roles aren't artifacts; autonomous agents have ~none.

## Sources

Primary-source citations are in `research-0001` (landscape) and `research-0002` (gate-test);
this note adds no new sources — it re-reads them for type data. `confidence: inferred`
throughout.

---
id: decision-0014
type: decision
status: ratified
depends_on: [invariants-v1, decision-0005, decision-0013]
owner: gundi
date: 2026-07-01
ratified: 2026-07-01
---

# 0014 — Invariant set: a compiled spec + change-ADRs (not a versioned, history-laden file)

## Context

Invariants are Trellis's core **architecture**, so a change to one is an architecture
*decision* — the natural fit is an **ADR** (append-only, context/decision/consequences),
which reuses our existing `decisions/` infrastructure and beats the alternatives (git history
isn't first-class/auditable; a plain changelog records *what* but not *why*; versioned files
proliferate). But the invariant file had been mixing **current truth** (the invariants) with
**change-history** (the `Changes from v0→v1` table, the amendment log) — which *violates our
own B4* (immutable history separated from consolidated current-truth). And now that `core/`
ships (`decision-0005`), the file should be a clean, installable, **compiled** spec.

## Decision

- **The invariant set is a compiled, revise-in-place spec** (`core/invariants/…`) — *current
  truth*. It sits at `ratified` and is **not re-ratified per change** (codebase-like: you
  ratify the decisions that change it, not the whole set each time).
- **Significant invariant changes are recorded as ADRs** (append-only, Layer B — *not*
  shipped). "Significant" = changes an invariant's meaning or structure. Minor/editorial edits
  (rewording, adding an example) are just spec edits — no ADR (minimal-first, B7).
- The spec carries **rationale by reference** (links to its governing ADRs), not inline
  change-narrative.
- **`invariants-v0` is superseded** (its id stays resolvable for append-only decisions that
  cite it; *retiring the file* is deferred — needs a check-4 registry-resolution refinement).

## Consequences

- The shipped invariant file becomes a clean compiled spec (structure + invariants + examples
  + slug registry + provenance) — no inline change log.
- B4 is now honored by the invariant set itself; the *why* lives in ADRs, one click away.
- Commits us to an ADR per significant invariant change **going forward** (we'd been lax —
  most past changes lived in PRs + the inline table; those are absorbed below).

## Migrated change history (was inline in `invariants-v1`)

**v0 → v1 (with rationale + source):**

| # | Change | Why | Source |
|---|---|---|---|
| 1 | Split flat list into **A/B + dials + floors** | Step 1: frameworks have structure, not enforcement | `research-0002` |
| 2 | **Reclassified Independent Verification** (v0-5) `methodology` → `trellis-design` | spec-driven tools lack it; Trellis supplies it; SpecSwarm proves it implementable | `research-0002` |
| 3 | Split "gate at every handover" (v0-2) into **A2 (points exist)** + **B2 (enforced)** | enforced ≠ merely defined; skippable gates puncture it | `research-0002` |
| 4 | Added **A4 ratifiable/checkable artifacts** | it's what B1/B3/B4 act on | synthesis |
| 5 | **Enforcement demoted to a dial (C1)** | strictness was a hypothesis; keep speed-first users | `decision-0008` |
| 6 | Added **gatekeeper dial (C2) `{agent\|human\|none}`** | who enforces is configurable | `decision-0008` |
| 7 | **Elevated loud-failure (v0-7) → transparency floor (D1)** | the true floor is surfacing, not enforcing | `decision-0008` |
| 8 | Named the **intent-gate floor (D2)** | the one non-configurable gate | v0-4 + `0008` |
| 9 | **Split B7**: minimal-first kept; reference-not-adoption → **B8** | strict single-framework adopt is legitimate | `decision-0002` + Spec Kit case |
| 10 | Added **B9 clarify-before-commit** | near-universal; absent from v0 | `research-0001/0002` |
| 11 | **B11 epistemic integrity merged into B3** (intent face) | same principle one layer up | maintainer + B3/D2 |
| 12 | **bounded-correction candidate dropped** → D1 | not invariant-worthy on reflection | minimal-first |

**Post-ratification amendments (2026-06-30 → 2026-07-01):**
- **`B1 → Directional-graph maintenance`** — consolidated former B1 (flow) + B6
  (self-improvement) + backprop + forward re-propagation (B6 → slug-supersede to
  `inv-graph-maintenance`). Surfaced by the spine's own drift-and-repair + graph-hygiene scan.
- **Examples convention added** — every invariant carries a concrete example (iron rule §7 on
  the set itself); surfaced by backprop (B1's abstraction revealed it).
- **Stable slugs** (`decision-0013`) — canonical ids; ordinals demoted to display labels.

## Supersedes / superseded by

— (none)

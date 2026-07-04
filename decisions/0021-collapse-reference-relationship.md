---
id: decision-0021
type: decision
status: ratified
depends_on: [invariants-v1, decision-0002, decision-0018, decision-0020, brief-§7]
owner: gundi
date: 2026-07-04
ratified: 2026-07-04
---

# 0021 — Collapse `inv-reference-relationship` (B8): no distinct mechanism

**Raised by:** the maintainer — a **prune review** of the invariant set, triggered by the recent
changes to it (self-improvement restored, the "no silent drift" theme added). Same discipline as the
prior merges (B11→B3, B10→D1, and — the other way — B6 restored).

## Context

**The rubric (consistent with `decision-0018`'s *keep B6*): collapse an invariant when it has no
distinct mechanism — when it is an existing floor/rule applied to a specific object; keep it when it
introduces a mechanism the others don't.** That rubric restored B6 (self-improvement has a real engine
graph-maintenance lacks) and, applied honestly, retires B8.

**What B8 says** (fact, `invariants-v1`): *the adopt-vs-adapt relationship is a free choice, but the
choice and every divergence from a reference are recorded and surfaced, never silent drift.*
Decomposed (inference):

- the **free adopt/adapt choice** is already **`decision-0002`** — a *dial*, not an invariant;
- **"recorded"** is **B4** (auditable archive);
- **"surfaced, never silent drift"** is **D1** / the *no-silent-drift* theme (`decision-0020`).

So B8 = D1 + B4 + the dial, **applied to the object "external frameworks."** It introduces **no
mechanism of its own**, and it is already tagged the most `provisional` in the set.

**The rest of the review, honestly (no manufactured collapses):** A1/A2 are a tight pair but distinct
dimensions (direction vs. discreteness); A3↔D2 are the same principle at two layers by *deliberate*
design (admission gate vs. floor); B9 leans on D1/A1 but adds a distinct near-universal *act*
(clarify-before-commit). Every other invariant carries a mechanism the others don't. **Exactly one
genuine collapse: B8.**

## Decision

1. **Retire `inv-reference-relationship` (B8)** as a standalone invariant. Its content re-homes, it is
   not lost:
   - the *surface-divergence-from-a-framework* facet → a **few-shot example under D1 / no-silent-drift**;
   - the *adopt/adapt choice* → **`decision-0002`**'s dial (unchanged; it now carries this role explicitly).
   Registry (`decision-0013`): `inv-reference-relationship` becomes a **retired slug** resolving to
   `floor-transparency` (D1) — old references (e.g. `decision-0002`) still resolve. **15 → 14**
   assessable slugs (B6 in, B8 out).
2. **Record the trigger that fired this** (a `B6` self-improvement trigger): *a **significant change to
   the invariant set** triggers a **prune / collapse review**.* This is graph-maintenance in the rules
   layer (when the set changes, re-review it for redundancy) — **not time-based** (a review with no
   change yields nothing new; "periodically" was loose phrasing). It rode the existing change, as the
   loop demands.

## Consequences

- **`invariants-v1` revised** — B8 section retired; registry marks it retired→D1; the "divergence from
  a framework you claim to follow" moves into D1 as a few-shot; count 15 → 14.
- **Catalog** drops the `inv-reference-relationship` entry (→ 14 assessable); D1's entry gains the
  framework-divergence example; the 15→14 cascade across `spec-0002` / rubric / agent.
- **`decision-0002`** stands and now explicitly carries the adopt/adapt role B8 was fronting.
- **Executed in increment 2** (the catalog `why`/`honored`/`violated` pass, `decision-0020`), **after
  the un-merge (PR #41) merges** — so the un-merge (+B6) and this collapse (−B8) land in sequence, not
  in conflict.

## Open questions

- **Independent review via a *non-Claude* model (maintainer).** Running the collapse / conformance
  review through a strong non-Claude reasoner would give genuinely independent judgment — it avoids
  *correlated* blind spots a single model family shares, which a same-family "independent" agent does
  not. A real strengthening of **B3** (and the B3 positive-control open question). **Owed when the
  tooling exists** — this environment runs Claude only; recorded as the target, not done.
- Does retiring B8 lose *explicit* pedagogy about the framework relationship? Mitigated: it becomes a
  prominent D1 example + the `decision-0002` dial, rather than a redundant standalone rule.

## Supersedes / superseded by

— (retires `inv-reference-relationship`/B8; content absorbed into `floor-transparency`/D1 +
`decision-0002`; not a document supersede)

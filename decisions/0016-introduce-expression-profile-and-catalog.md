---
id: decision-0016
type: decision
status: draft
depends_on: [research-0003, research-0005, research-0006, research-0007, research-0009, decision-0009]
owner: gundi
date: 2026-07-03
---

# 0016 — Introduce two typed artifacts: the invariant-signature catalog and the expression profile

**Raised by:** the genetics/DES research (`research-0005/0006/0007`) and the issue triage
(`research-0009`), which found that four of five issue clusters (#22/#23/#24/#27/#28) all consume one
shared object Trellis does not yet name as an artifact.

## Context

`research-0009` surfaced that the application front-end (#23), cross-instance validation (#28),
tutoring (#27), and partial application (#22) are not four bespoke features — they all read/write
**one per-instance object** (which invariants are active, how strongly, gatekept by whom) plus **one
product-level dictionary** (what each invariant looks like when honored). Both lenses converged on
it: genetics calls it the *expression profile* (`research-0005`), SCT the *control map*
(`research-0006`), and #23's comment already half-named the dictionary the *invariant-signature
catalog*. `research-0003` established that Trellis's artifact `type` is an **open field, extensible
by a recorded decision** — this is that decision.

## Decision

Introduce two `type`s (schema deferred to `spec-0002`; this decision fixes their *existence, scope,
and relationship*):

- **`signature-catalog`** — *scope `trellis-product`* (one, shipped). The genome annotation: for each
  invariant (by stable slug), what it *is*, its observable **signature** (the tells it is honored
  implicitly), its scope, and its **default C1/C2**. Consumed by Assess (#23) and tutoring (#27).
  `decision-0009` already anticipates "catalog" as core content.
- **`expression-profile`** — *scope `core-methodology`* (one **per instance**). The per-instance
  readout: for each invariant, `active?`, `C1` strength, `C2` gatekeeper, and delivery axes A/B
  (`research-0007`). Produced by Assess, **ratified by the human (D2)**, consumed by Apply, diffed
  across instances (#28), and minimized into "Trellis-lite" (#22).

Relationship: **catalog : profile :: reference-genome-annotation : single-cell-readout** — the
catalog is the dictionary; the profile is one instance's expression against it.

## Consequences

- Unblocks the `research-0009` build order: profile + catalog are the foundation clusters 1/3/4/5
  depend on.
- **`spec-0002` owed** — the schema + lifecycle for both (fields; how a profile references catalog
  slugs; how Assess populates + confidence-tags; the ratification flow). This decision is the gate
  `spec-0002` consumes.
- The conformance check (`spec-0001`) must learn the two new types + their required sections.
- `draft` until the underlying research (0005–0009) is ratified or `spec-0002` forces revision.

## Supersedes / superseded by

— (none)

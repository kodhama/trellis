---
id: decision-0001
type: decision
status: ratified
ratified: 2026-06-29
depends_on: [brief-§9.1, brief-§7]
owner: gundi
date: 2026-06-29
---

# 0001 — Product form: a shippable, portable pack, validated by dogfooding

**Fork (brief §9.1):** product vs consulting.

## Context

The onboarding invariant (Pillar III: teach in situ) encodes a *self-serve telos* — it is
the mechanism that removes the human teacher from the loop. So the invariant set itself
points to a product end-state, with self-serve stopping at the intent gate (invariant 4).
The product/consulting question therefore maps onto the authority split, not onto a
business-model fork.

## Decision

Trellis is built as a **self-serve, shippable, portable pack** — something that can be
dropped into other projects. We **validate it by dogfooding our own project**, not by
running a consulting practice. Reaching external instances #2/#3 remains the key risk to
attack (brief §7), but it is *not* gated behind a consulting engagement.

## Consequences

- The deliverable's *form* is a pack (instructions, gates, sub-agents), not a service.
- N=1 validation (our own project) is accepted *for now*; instance #2 is an open need,
  tracked in the invariant set's open questions.
- Distribution / go-to-market is explicitly deferred — not decided here.

## Supersedes / superseded by

— (none)

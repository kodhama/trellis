---
id: decision-0002
type: decision
status: ratified
ratified: 2026-06-29
depends_on: [brief-§9.2, brief-§5, decision-0001]
owner: gundi
date: 2026-06-29
---

# 0002 — Adaptation is a user-controlled dial, not a mode choice

**Fork (brief §9.2):** conductor-first vs author-first.

## Context

The brief framed conductor mode (run an existing methodology) and author mode (generate a
fitted one) as two on-ramps to choose between. In practice they are the two *ends of one
axis*: how much the user adapts a baseline.

## Decision

Treat conductor↔author as a **single continuous dial the user controls** — how much they
want to adapt a baseline methodology. Trellis must support sitting anywhere on it. In the
limit, the project carries **enough seeds to spawn a coherent methodology organically**,
using others only as reference.

## Consequences

- No "pick the mode" fork; the dial is a product feature, not a configuration we hard-set.
- The build must support both extremes from one mechanism, so neither is special-cased.
- "Enough seeds to spawn a coherent methodology" becomes a design target for the parts
  catalog (relates to decision `0003`).

## Refinement (2026-06-29) — two coarse modes, not a per-area dial

Testing the dial concretely (maintainer pushback): a *per-area* adherence setting (conduct
framework X for the build loop, author your own planning) risks being **brittle**. Working
stance instead — two coarse modes, the real variable being *single-source fidelity vs
multi-source synthesis*:

1. **Adopt** — one framework, faithfully, for the whole process (conductor end).
2. **Adapt** — synthesize from one *or several* references, evolving (author end).

**Deliberately not formalized into the invariant model yet** — it is abstract and hard to
test, so per minimal-first it stays a recorded stance until instance #2 shows what lands.
`invariants-v1` encodes only the durable part: the reference relationship (adopt/adapt/
diverge) is explicit, recorded, and surfaced (B8).

## Supersedes / superseded by

— (none)

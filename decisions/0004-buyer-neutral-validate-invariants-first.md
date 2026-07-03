---
id: decision-0004
type: decision
status: ratified
ratified: 2026-06-29
depends_on: [brief-§9.4, brief-§4, decision-0003]
owner: gundi
date: 2026-06-29
---

# 0004 — Buyer-neutral by construction; validate the invariants first

**Fork (brief §9.4):** target buyer — startup (speed+safety) vs enterprise (trust+audit).

## Context

Startups and enterprises differ in the *kind* of gates, steps, and artifacts they need —
but the premise is that **both ride the same invariants**. If that holds, Trellis is
organically suited to both, and buyer selection is downstream parameterization, not a fork
to resolve now. This is an untested hypothesis — which is precisely the point of the
project.

## Decision

Keep **both buyers in view**; build **buyer-neutral**. The startup/enterprise difference
lives in gates/steps/artifacts layered *on top of* the shared invariants. The **starting
point is to validate and identify a set of reasonable, useful invariants** (`invariants-v0`)
— everything else (mode, framework, buyer) is parameterized by them and deferred until
they are proven.

## Consequences

- The invariant set is the project's first deliverable and gate; the spine and all buyer-
  specific machinery `depends_on` a ratified version of it.
- **Explicitly out of scope now:** the "meta of relaxing invariants" (studying which
  invariants can be loosened and at what cost). Named as a known later layer, not pursued.

## Supersedes / superseded by

— (none)

---
id: decision-0011
type: decision
status: ratified
depends_on: [invariants-v1]
owner: gundi
date: 2026-06-29
ratified: 2026-06-29
---

# 0011 — Add a spec (behavioral-contract) stage between decisions and build

**Raised by:** the maintainer — other projects (math-quest) author **specs between ADRs and
build**, and tests derive from them. We were about to jump decisions → build for the spine.

## Context

Invariant A1's stages are *research → decisions → **contracts** → implementation →
validation*. "Contracts" **is** the spec stage — and we'd been skipping it. Friction (about to
build the spine with no contract) revealed the boundary (B7 / self-improvement loop B6).

## Decision

- Our process **and the product** gain an explicit **spec stage**: *research → decisions →
  **spec** → implementation → validation* (A1 made literal). New artifact `type: spec`.
- A spec is a **behavioral contract** carrying **acceptance criteria from which tests/checks
  derive** — harvested from math-quest's `adr-0008-spec-stage` and `spec-quality` rubric.
- Implementation consumes only a **ratified/approved** spec (B1); the spec is gated before any
  build (A2 / D2).

## Consequences

- The spine's first build artifact is a **spec**, not code.
- Adds a `spec-quality` rubric (to harvest) and `type: spec` to the artifact contract — which
  the spine will formalize. (Recursion: the spine's own spec is the first user of the new
  stage.)
- Keeps tests grounded in an approved upstream rather than invented post-hoc.

## Open questions

- **Spec granularity** — plan-inline-by-default (per `adr-0008`)? Does *every* change need a
  spec, or only non-trivial ones (minimal-first → likely a threshold)?

## Supersedes / superseded by

— (none)

---
id: decision-0003
type: decision
status: ratified
ratified: 2026-06-29
depends_on: [brief-§9.3, brief-§5, decision-0002]
owner: gundi
date: 2026-06-29
---

# 0003 — Methodology-agnostic: inspect or be told, then supervise

**Fork (brief §9.3):** which reference framework (spec-kit vs BMAD) to support first.

## Context

The fork assumed Trellis supports one framework first. But the design intent is for Trellis to
supervise *any* methodology that honors the invariants — not to privilege one framework.

## Decision

Trellis is **methodology-agnostic by construction**. It determines the methodology in play
either by **inspecting the project** or by being **told via a standardized instruction
file**, then acts as tutor/supervisor for it — provided that methodology satisfies
`invariants-v0`. A built-in **default reference** *may* be included, but is optional and
not a privileged "first framework."

## Consequences

- The first real machinery owed by this decision is a **standardized way to describe a
  methodology to Trellis** (the instruction-file format) — not a spec-kit or BMAD adapter.
- The empirical question moves to the foreground: *do real methodologies actually satisfy
  the invariants?* — answerable only by pointing Trellis at many of them. This is the same
  test named in `invariants-v0` open questions.
- spec-kit / BMAD become *reference points to harvest from* (invariant 10), not adoption
  targets — recorded as decisions if/when borrowed.

## Supersedes / superseded by

— (none)

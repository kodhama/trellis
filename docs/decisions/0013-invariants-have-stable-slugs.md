---
id: decision-0013
type: decision
status: ratified
depends_on: [invariants-v1, decision-0004]
superseded_in_part_by: [decision-0038]
owner: gundi
date: 2026-07-01
ratified: 2026-07-01
---

# 0013 — Invariants carry stable slugs; references use slugs

## Context

Consolidating the invariant set (the `B6 → B1` merge) forced an awkward **tombstone**:
decisions cite invariants by **ordinal** (`B6`, `B8`), so renumbering would either edit
ratified decisions (which append-only forbids) or leave those references fragile — and the
conformance check can't verify prose references, so a bad update would be *silent* drift.
Root cause: **ordinal numbers are unstable identifiers.** The friction has now recurred twice
(the `B10`/`B11` churn during v1 drafting; this tombstone) — by minimal-first, the boundary is
revealed. This decision is `inv-graph-maintenance` (B1) firing on itself: downstream friction
(the tombstone) surfaced an upstream gap (no stable ids) → repair upstream.

## Decision

- **Every invariant has a stable `slug`** (e.g. `inv-graph-maintenance`) — its *canonical
  identifier*, recorded in the `invariants-v1` **Identifiers registry**.
- The `A/B/C/D`+number is a **display label**: convenient, **frozen** for existing invariants
  (never renumbered), but **not** a reference.
- **References use slugs.** When invariants merge/retire, the absorbed slug is `superseded_by`
  the survivor's in the registry, so any old reference — *including an ordinal* — resolves
  through it. **No citing artifact is edited to chase a rename** (the same historical-reference
  exemption already used for artifact supersedes; append-only stays intact).

## Consequences

- **Retires the tombstone hack:** a merge is now a proper slug-supersede, not an ordinal gap.
- Future invariant refactors are clean and non-destructive to the decision record.
- **Deferred (minimal-first):** inline slugs in each invariant heading (the registry suffices
  for now); migrating machinery references (`spec-0001`/rubric/agent cite `A1`/`B1`/`D1`) from
  ordinals to slugs (grandfathered while ordinals stay frozen).

## Supersedes / superseded by

— (none)

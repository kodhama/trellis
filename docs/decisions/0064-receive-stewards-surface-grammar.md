---
id: decision-0064
type: decision
status: approved
depends_on: [decision-0061, kodhama/kodhama-0022-propagate-collective-strategy, kodhama/kodhama-0023-separate-operational-availability-from-support]
superseded_in_part_by: [kodhama/kodhama-0025-retire-the-surface-matrix]  # 2026-07-28 — the received upstream only. "Stewards remains authoritative for both" and "the shared strategy" no longer have a referent; 0025 supersedes kodhama-0023 in full. What stands: this receipt authorized nothing, and the Consequences line reserving migration to a future product decision is now the live clause, not a dormant one
owner: agent
updated: 2026-07-28
---

# 0064 — receive the Stewards surface grammar

> **Forward pointer.** Stewards
> [`kodhama-0025`](https://github.com/kodhama/stewards/blob/main/decisions/0025-retire-the-surface-matrix.md)
> (approved 2026-07-27) supersedes `kodhama-0023` in full: no family contract
> requires any plugin to carry `surfaces.json`, exact surface rows,
> `availability_state`, or `support_claim`. This receipt received a strategy
> that no longer exists, so its Context — *"Stewards remains authoritative for
> both"* — holds only for `kodhama-0022`.
>
> The Consequences clause below is unchanged and now load-bearing: *"A future
> product decision must define and authorize any migration from
> `behavior_state`."* Retiring the upstream mandate does not retire Trellis's
> own `plugins/trellis/surfaces.json`, which is still shipped and still
> referenced by `decision-0061` §3 and `decision-0063`. Whether Trellis keeps
> it is a Trellis decision, and it is not this one.

## Context

Trellis owns a current Kodhama plugin, so the approved Stewards
[`kodhama-0023`](https://github.com/kodhama/stewards/blob/main/decisions/0023-separate-operational-availability-from-support.md)
strategy applies here. This receipt follows the approved
[`kodhama-0022`](https://github.com/kodhama/stewards/blob/main/decisions/0022-propagate-collective-strategy.md)
propagation decision. Stewards remains authoritative for both.

## Decision

Trellis records receipt of the shared strategy. This cross-link communicates
the upstream constraint without restating it.

Reconciliation of Trellis's current `behavior_state` metadata is separate
product work. This receipt authorizes no metadata, schema, behavior, setup,
package, release, distribution, evidence, validation, or support change.

## Consequences

The shared strategy is discoverable in Trellis's local decision graph without
being copied or redefined.

Decision 0061 and Trellis's current product surfaces remain unchanged. A
future product decision must define and authorize any migration from
`behavior_state`; that work cannot be inferred from this receipt.

The maintainer ratified the upstream strategy and directed its receipt memos
to roll out on 2026-07-26. An independent decision adversary returned `SOUND`
for exact commit `dc5a601`. That rollout direction is the human intent act;
`approved` records it.

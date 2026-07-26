---
id: decision-0064
type: decision
status: gated
depends_on: [decision-0061, kodhama/kodhama-0022-propagate-collective-strategy, kodhama/kodhama-0023-separate-operational-availability-from-support]
owner: agent
updated: 2026-07-26
---

# 0064 — receive the Stewards surface grammar

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
to roll out on 2026-07-26. This artifact remains gated until independent
review confirms that it is only the authorized thin receipt.

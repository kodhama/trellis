---
id: decision-0062
type: decision
status: gated
depends_on: [decision-0061, kodhama/kodhama-0021-separate-adoption-posture-from-support, kodhama/kodhama-0022-propagate-collective-strategy]
owner: agent
updated: 2026-07-25
---

# 0062 — receive the Stewards adoption-posture strategy

## Context

Trellis owns a current Kodhama plugin, so the approved Stewards
[`kodhama-0021`](https://github.com/kodhama/stewards/blob/main/decisions/0021-separate-adoption-posture-from-support.md)
strategy applies here. This receipt follows the approved
[`kodhama-0022`](https://github.com/kodhama/stewards/blob/main/decisions/0022-propagate-collective-strategy.md)
propagation decision. Stewards remains authoritative for both.

## Decision

Trellis records receipt of the shared strategy. This is communication only: it
selects no adoption posture, changes no product-owned support claim, and
authorizes no implementation, package, release, catalog, distribution, or
support work.

Any future Trellis adoption-posture choice is a separate local decision. That
follow-up is parked until a concrete need arises.

## Consequences

The shared strategy is discoverable in Trellis's local decision graph without
being copied or redefined. No existing Trellis decision or specification is
superseded or amended.

The maintainer authorized rolling out the cross-link ADRs on 2026-07-25. This
artifact is self-checked and `gated` for independent soundness review; exact
ratification remains a later human intent act.

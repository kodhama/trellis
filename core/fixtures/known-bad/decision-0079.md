---
id: decision-0079
type: decision
depends_on: [decision-0001]
owner: gundi
---

# Fixture decision-0079 — the retired-artifacts registry

## Context

Valid on purpose. The conformance check finds this artifact by its id, `decision-0079`, and reads
retired artifact ids from the 3a table below, as it does against the live record.

## Decision

**3a. Retired-artifacts registry.**

| Retired id | Was |
|---|---|
| `spec-0001` | a retired fixture spec |
| `spec-0007` | another retired fixture spec |

## Consequences

- `valid-refs.md` depends on `spec-0007` and on `spec-0001@v1`, which resolve only through this table.

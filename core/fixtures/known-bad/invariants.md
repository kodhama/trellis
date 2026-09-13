---
id: invariants-v1
type: invariant-set
depends_on: [decision-0001]
owner: gundi
---

# Fixture invariant set — the Identifiers registry

Valid on purpose. The conformance check finds this artifact by its id, `invariants-v1`, and reads
retired ids from the section below, as it does against the live invariant set. `valid-refs.md`
depends on `invariants-v0`, `inv-old-slug` and `B8`, which resolve only through this registry.

## Identifiers (stable slugs)

**Legacy code map (retired display labels → slugs):**

| Slug | Legacy code (retired) | Note |
|---|---|---|
| `inv-kept` | A1 | |
| `inv-old-slug` | B8 | **collapsed → `inv-kept`**; id resolves to `inv-kept` |

**Retired artifact ids:**

| Retired id | → Successor |
|---|---|
| `invariants-v0` | `invariants-v1` |

## Acceptance criteria

- The registry above resolves every retired id the fixture corpus cites.

## Open questions

- None.

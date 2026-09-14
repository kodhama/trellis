---
id: profile-fixture-duplicate-c2
type: expression-profile
depends_on: [signature-catalog-v1, invariants-v1]
owner: gundi
scope: core-methodology
---

# Fixture profile with a duplicated C2 column

Seeded violation: the `## Profile` table carries `C2` twice. A reader that kept only the last column
would see `human` for `inv-gate` and miss the `none` in the first, a forbidden `C2: none` (rule 11).
The duplicate is the only finding; no row is read.

## Delivery

- **delivery_relationship:** `supervisor`

## Profile

| slug | active | C1 | C2 | C2 | basis | confidence | evidence |
|---|---|---|---|---|---|---|---|
| `inv-gate` | true | enforced | none | human | honored-implicitly | verified | seeded: the first `C2` is `none`, the second `human` |

## Assessment notes

- The duplicated `C2` column is the seeded violation.

## Open questions

- None.

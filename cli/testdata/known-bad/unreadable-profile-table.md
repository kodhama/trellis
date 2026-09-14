---
id: profile-fixture-no-c2
type: expression-profile
depends_on: [signature-catalog-v1, invariants-v1]
owner: gundi
scope: core-methodology
---

# Fixture profile with no C2 column

Seeded violation: the `## Profile` table has no `C2` column, so rule 11 cannot be applied to any row.
The table's problem is the only finding; no row is read.

## Delivery

- **delivery_relationship:** `supervisor`

## Profile

| slug | active | C1 | basis | confidence | evidence |
|---|---|---|---|---|---|
| `inv-good` | true | enforced | honored-implicitly | verified | valid: every other column is present |

## Assessment notes

- The missing `C2` column is the seeded violation.

## Open questions

- None.

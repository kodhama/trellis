---
id: profile-fixture
type: expression-profile
depends_on: [signature-catalog-v1, invariants-v1]
owner: gundi
scope: core-methodology
---

# Fixture expression profile

Seeded violations for rules 9, 10a and 11, beside rows that are valid on purpose. Every row resolves
against `catalog.md` except `inv-unknown`.

## Delivery

- **delivery_relationship:** `supervisor`

## Profile

| slug | active | C1 | C2 | basis | confidence | evidence |
|---|---|---|---|---|---|---|
| `inv-good` | true | enforced | independent-agent | honored-implicitly | verified | valid: `valid-refs.md` exercises every accepted reference form |
| `inv-gate` | true | enforced | none | honored-implicitly | verified | seeded, rule 11: `C2: none` on an entry whose `intent_locus` is true |
| `inv-misaligned` | true | enforced | none | honored-implicitly | verified | valid: `C2: none` on an entry that is not an intent locus |
| `inv-one-pair` | true | expressed | human | to-be-added | | |
| `inv-no-why` | true | enforced | human | honored-implicitly | verified | |
| `inv-no-c2` | true | enforced | human | honored-implicitly | likely | seeded, rule 10a: `likely` is not a confidence tag |
| `inv-unknown` | true | enforced | human | honored-implicitly | verified | seeded, rule 9: no catalog entry carries this slug |

## Assessment notes

- `inv-one-pair` is valid on purpose: a `to-be-added` row owes no confidence or evidence.
- `inv-no-why` is seeded for rule 10a: an `honored-implicitly` row with no evidence.

## Open questions

- None.

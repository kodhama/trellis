---
id: signature-catalog-shadow
type: signature-catalog
depends_on: [invariants-v1]
owner: gundi
---

# Fixture shadow signature catalog

Seeded violation, rule 8e: the corpus holds a second signature catalog, and exactly one,
`signature-catalog-v1`, is allowed. Its file name sorts before `catalog.md`, and its `inv-gate` entry
is valid but sets `intent_locus: false`. A check that resolved profiles through the first catalog it
read would take this entry, and `profile.md`'s `C2: none` on `inv-gate` (rule 11) would pass unseen.

## Entries

- **`inv-gate`**
  - what: an intent-locus entry, shadowed with its intent locus cleared.
  - directive: Get the human approval before you ship.
  - why: a human owns the goal.
  - signature: a recorded human approval.
  - honored:
    - *(ADR)* a human ratifies the decision.
    - *(release)* a human approves the deploy.
  - violated:
    - *(ADR)* a decision ratifies itself.
    - *(release)* a deploy promotes itself on green tests alone.
  - class: `floor`  ·  mechanizable: `false`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `human`

## Acceptance criteria

- The check reports this file under rule 8e and resolves no profile through it.

## Open questions

- None.

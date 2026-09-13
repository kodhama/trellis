---
id: signature-catalog-v1
type: signature-catalog
depends_on: [invariants-v1]
owner: gundi
---

# Fixture signature catalog

Seeded violations for rules 8b and 8c, and the entries `profile.md` resolves against. `inv-good`
and `inv-gate` are valid on purpose, in the shapes the live catalog uses: prose wrapped onto the
next line, `·`-joined fields, a middle dot inside a prose field and inside a parenthetical,
`intent_locus` wrapped onto a line of its own, bold `intent_locus`, and uppercase tags.

## Entries

### Valid entries

- **`inv-good`** *(valid: its header parenthetical wraps
  onto a second line)*
  - what: a valid entry whose prose wraps onto
    a second line.
  - directive: Do the valid thing.
  - why: the check reads this entry and reports nothing.
  - signature: every field is present · even when a prose field carries a middle dot.
  - honored:
    - *(CI)* a check fires on every change.
    - *(docs)* a guide names its source, and this example wraps
      onto a second line.
  - violated:
    - *(CI)* a check is skipped silently.
    - *(docs)* a guide names no source.
  - class: `methodology`  ·  mechanizable: `false` (a parenthetical · with a middle dot)
    ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-gate`**
  - what: an intent-locus entry.
  - directive: Get the human approval before you ship.
  - why: a human owns the goal.
  - signature: a recorded human approval.
  - honored:
    - *(ADR)* a human ratifies the decision.
    - *(release)* a human approves the deploy.
  - violated:
    - *(ADR)* a decision ratifies itself.
    - *(release)* a deploy promotes itself on green tests alone.
  - class: `floor`  ·  mechanizable: `false`  ·  **intent_locus: `true`**
  - default_C1: `enforced`  ·  default_C2: `human` (never `none`)

### Seeded violations

This line belongs to no entry.

- **`inv-no-why`**
  - what: an entry with no `why` field, rule 8b.
  - directive: Say why.
  - signature: a missing field.
  - honored:
    - *(docs)* a rule states its goal.
    - *(code)* a comment states its reason.
  - violated:
    - *(docs)* a rule states no goal.
    - *(code)* a comment states no reason.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `human`

- **`inv-no-c2`**
  - what: an entry whose combined dial line has no `default_C2`, rule 8b.
  - directive: Name the gatekeeper.
  - why: a default needs both dials.
  - signature: a combined line with one dial.
  - honored:
    - *(ops)* a runbook names who approves.
    - *(CI)* a workflow names its reviewer.
  - violated:
    - *(ops)* a runbook names nobody.
    - *(CI)* a workflow names no reviewer.
  - class: `trellis-design`  ·  mechanizable: `false`  ·  intent_locus: `false`
  - default_C1: `enforced`

- **`inv-one-pair`**
  - what: an entry with one honored and one violated example, rule 8c.
  - directive: Show two pairs.
  - why: one example reads as domain-specific.
  - signature: a single pair.
  - honored:
    - *(docs)* the one honored example.
  - violated:
    - *(docs)* the one violated example.
  - class: `trellis-design`  ·  mechanizable: `false`  ·  intent_locus: `false`
  - default_C1: `expressed`  ·  default_C2: `human`

- **`inv-misaligned`**
  - what: an entry whose second pair's tags differ, rule 8c.
  - directive: Pair each example with its twin.
  - why: a pair shows one situation broken and then fixed.
  - signature: tags that do not align.
  - honored:
    - *(CI)* the first honored example.
    - *(CI)* the second honored example.
  - violated:
    - *(CI)* the first violated example.
    - *(ops)* the second violated example, under a different tag.
  - class: `trellis-design`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

## Acceptance criteria

- `inv-good` and `inv-gate` yield no finding; each seeded entry yields the finding its `what` names.

## Open questions

- None.

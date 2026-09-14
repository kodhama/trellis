---
id: decision-0203
type: decision
depends_on: [decision-0001]
owner: gundi
---

# Suffixed and fenced headings

## Context

Two heading forms. A lowercase heading carrying a parenthetical suffix satisfies `Decision`, so no
finding follows for it. The only `Consequences` heading sits inside a fenced code block, so rule 6a
reports `Consequences` as missing.

## decision (draft)

Nothing is decided here.

```markdown
## Consequences

- An example inside a fence, not a section of this file.
```

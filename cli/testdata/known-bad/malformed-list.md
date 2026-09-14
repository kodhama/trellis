---
id: decision-0206
type: decision
depends_on: [decision-0001
owner: gundi
---

# Malformed flow list

## Context

Seeded violation, rule 1c: `depends_on` opens a flow list that never closes with `]`. Its entries are
not resolved, so this is the only finding the file yields.

## Decision

Nothing is decided here.

## Consequences

- None.

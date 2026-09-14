---
id: fixture-repeated-and-empty
type:
depends_on: [decision-0001]
owner: gundi
owner: agent
---

# Repeated and empty required fields

Seeded violations, rule 1b: `owner` appears twice, and `type` has no value. With no type, neither
check 2 nor check 6 applies to this file.

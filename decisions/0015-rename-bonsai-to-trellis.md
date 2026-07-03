---
id: decision-0015
type: decision
status: draft
depends_on: [research-0008, brief-§4, invariants-v1]
owner: gundi
date: 2026-07-03
---

# 0015 — Rename the product Bonsai → Trellis (option B: now, not staged)

**Raised by:** the maintainer, after the genetics/DES research (`research-0005/0006/0007`) clarified
the product's essence and `research-0008` surfaced the name.

## Context

`research-0008` found the old name **Bonsai** carried a *topiary* metaphor — *"shape to the
specimen,"* imposing an external form — which is the **opposite** of the essence the research
clarified: a *maximally-permissive supervisor* that constrains a space of allowed behaviors and
permits maximal freedom **within** (`research-0006` Result 3). `research-0008` recommended
**staging** the rename (option A) on the project's own naming guardrail (`brief-§4`: naming is a
deliberate-later act) plus the N=1 / genealogical-taint cautions, with **Trellis** the pre-committed
lead candidate.

## Decision

Rename **now — option B**, overriding the staged recommendation. **Trellis** keeps the horticultural
lineage while *correcting* the exact flaw: a trellis is structure that **enables growth rather than
dictating form** — the permissive supervisor as a garden object (topiary → trellis = shape-to-a-form
→ support-growth-within-bounds). Maintainer's rationale for *now* rather than staged: **stop the old
name spreading into more places** while the corpus is still small and internal, so the churn is
cheapest today — the counter-weight to the guardrail's caution, judged decisive at this size.

Scope executed: all repo content + three `bonsai-`named files renamed; the GitHub repo
`gundisalwa/bonsai → gundisalwa/trellis` (local `origin` updated; GitHub keeps a redirect from the
old URL). **Unchanged on purpose:** the invariant-set framing stays *"our synthesis, v1"* (only the
possessive updates — "Trellis's invariants"), so the guardrail's actual target is untouched; and
artifact `id`s (`invariants-v1`, `decision-*`, `research-*`) carry no product name, so no
`depends_on` broke.

## Consequences

- The topiary/permissive mismatch is resolved at the name level; new contributors meet a metaphor
  that matches the control philosophy.
- **Follow-ups still owed:** (a) the local working directory (`…/Projects/bonsai`) — deliberately
  *not* renamed, to avoid breaking the live session; (b) auto-memory files (outside the repo) still
  say "Bonsai"; (c) open GitHub issue bodies (#22–#28) still say "bonsai" (historical; optional).
- The staged path (option A) and its guardrail reasoning are preserved in `research-0008` as the
  record of what was weighed — the recommendation was A; the maintainer chose B.
- **`draft` pending ratification at the intent gate (D2):** a rename is an intent-layer act, so the
  builder does not self-ratify — the maintainer ratifies.

## Supersedes / superseded by

— (none)

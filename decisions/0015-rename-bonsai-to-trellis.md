---
id: decision-0015
type: decision
status: ratified
depends_on: [brief-§4, invariants-v1]
owner: gundi
date: 2026-07-03
ratified: 2026-07-03
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
- **Ratified at the intent gate (D2), 2026-07-03**, by the maintainer's explicit permission this
  session — a rename is an intent-layer act, so the builder does not self-ratify. Note:
  `research-0008` (which informed this) remains `draft`; it is *informing context*, not a consumed
  upstream, so this ratified decision depends only on ratified/external refs (`brief-§4`,
  `invariants-v1`) — no ratified→draft edge.

## Supersedes / superseded by

— (none)

---

> **Amendment (2026-08-22, append-only — the maintainer's correction, raised in conversation
> while auditing the leftover mentions; the merge of this PR ratifies it).**
> **Follow-up (c) was misclassified, and is owed rather than optional.** The original text calls
> the leftover "bonsai" in open issues #22–#28 *"(historical; optional)"*. An audit of all 46
> issues found **33 occurrences across 7 issues, none of them historical**: there is not one dated
> quotation, "renamed from" construction, or as-it-was-then gloss among them. They are ordinary
> product prose about a live product, written 2026-07-01/02 — one and two days *before* this
> decision ratified — which is why they read as current rather than retrospective.
>
> Three are not cosmetic at all but **dangling pointers to renamed artifacts**: #28 cites
> `bonsai-self` (now `profiles/trellis-self.md`), #23 cites `bonsai-core` ×3 (now `trellis-core`),
> and #26 glosses `decision-0009` as *"how bonsai improves itself"* while that record's heading
> reads *"How Trellis improves itself"* — an open issue misquoting the current title of the
> decision it points at. "Optional" was the wrong call for a broken reference.
>
> **This is an amendment, not a supersession** (`decision-0051`'s correction precedent): the
> judgement was loose when made, not outgrown by later events. **The original text above is not
> edited** — what was decided on 2026-07-03, and that it was decided wrongly on this one point,
> both stay legible.
>
> **The genuine historical set is named here so the sweep cannot swallow it:** this record itself
> (filename and body — it *is* the record of the rename); `research/0008`, whose §Part 2 declares
> its own "Bonsai" mentions load-bearing as the subject of its analysis; merged PR titles #1, #10
> and #31; and git history. Issue **comments** are also exempt — they are dated utterances by a
> person, not specification, and are not rewritten. Follow-up (a) is now **done** (the working
> copy is `Projects/trellis`); (b) is untouched by this amendment.
>
> **Two boundary rules the sweep needed, decided by the maintainer and recorded so the next
> rename does not re-litigate them:**
> 1. **A closed issue is frozen.** Closing is what "historical" is *for*, so a closed issue is not
>    swept regardless of what its text says. #25 is therefore exempt, and its live-in-kind
>    `bonsai-native` mention stays.
> 2. **A quotation follows its source when the source is revise-in-place; it pins when the source
>    is append-only.** #22 quotes `invariants-v1` on *"Bonsai's value"*, and the rename rewrote that
>    sentence in the source. `invariants-v1` declares itself *"the compiled current-truth spec
>    (revise-in-place)"* — so a citation of it tracks it, and the fixed text is the accurate
>    citation; anyone wanting the 2026-07-01 wording cites a git rev. Had the quoted source been a
>    `decisions/` record, the opposite would hold: append-only text never moves, so the quote pins.

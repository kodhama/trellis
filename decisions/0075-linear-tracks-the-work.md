---
id: decision-0075
type: decision
status: approved  # maintainer's intent act 2026-08-23, in-conversation ("approve") after reading the drafted record — this flip records it (decision-0046, decision-0022). Author (agent) != approver (maintainer). The migration itself was directed step by step in the same session
depends_on: [decision-0028]
informed_by: [decision-0027, decision-0072, decision-0074]
owner: agent
date: 2026-08-23
---

> **Provenance.** Directed by the maintainer in session on 2026-08-23, following `kodhama/math-quest`,
> which moved on 2026-08-20. Prior art was read from that migration and confirmed with the session
> that performed it, rather than inferred from its artifacts — which mattered, because its
> `CLAUDE.md` still describes the migration as partial and that description is stale.

# 0075 — Linear tracks the work; GitHub hosts the code

## Context

- **The backlog stopped carrying information.** At migration time trellis had 27 open GitHub
  issues, **26 of them at `stage: triage`** — including issues filed seven weeks earlier and worked
  on since. Only one had ever advanced. The `kodhama:issues` convention says to advance stage on any
  issue you touch; in practice nobody did, so the field recorded when an issue was *filed*, not how
  far it got. A field that never moves is not a weak signal, it is a false one.

- **math-quest had already moved, and the record of it was wrong.** Its `CLAUDE.md` says *"The
  migration is partial. Surviving GitHub issues were not ported."* The session that performed it
  confirmed the opposite: **106 open issues went to 0**, with roughly 39 closed in triage, 15 ported
  as bugs, ~30 as improvements, and 25 folded into an ideas document. The line was written
  mid-migration and never updated. Recorded here because this decision would have inherited a false
  premise from it.

- **The citation problem is specific to this repo.** `decisions/` is append-only and cites issue
  numbers heavily — `#165`, `#201`, `#206`, `#212` and others. Those citations can never be
  rewritten, so `github.com/kodhama/trellis/issues/N` must keep resolving no matter what moves.

## Decision

**1. Linear tracks the work; GitHub keeps pull requests, CI and code.** Kodhama workspace, team
**Trellis** (`TRL-*`). No super-team: label inheritance is a *workspace* property, verified — a team
created minutes before the port already returned all seven workspace labels with identical IDs, with
no parent involved.

**2. trellis follows the Linear taxonomy, and the `kodhama:issues` convention no longer applies
here.** Type becomes the `Bug` / `Feature` / `Improvement` labels; stage becomes Linear's workflow
states; `facing:` is dropped as unused in practice. Severity maps onto the workspace `Severity`
group, whose descriptions are load-bearing and were adopted verbatim from math-quest:
`blocker → High`, `broken-feature → Medium`, `papercut → Low`. `Critical` is reserved for its
stated meaning — outage, data loss, or security exposure — and nothing in the backlog qualified.

**3. Ideas are a document, not issues** — the Linear [Ideas
document](https://linear.app/kodhama/document/ideas-4220c523c6a1), team Trellis. An idea filed as an
issue is a to-do nobody agreed to; an idea written as prose is a claim someone can challenge. Each entry carries the **trigger** that
would promote it — a condition that revisits its priority, distinct from an *open question*, which
is something undecided about work already scoped and is resolved by doing that work.

**4. Closed GitHub issues carry a banner in the body, not a comment.** Every migrated issue was
closed with a banner prepended to its **body**, marked `<!-- trellis:linear-migration -->` so
re-running is idempotent. A closing comment can be scrolled past, and a body containing standing
instructions keeps reading as true after it closes. Three banner texts, because they closed for
three different reasons — *moved*, *parked as an idea*, *resolved*. A "moved to Linear" banner on an
issue that was actually resolved would be false, and these are permanent.

**5. A Linear issue carries content here — the same rule as math-quest, correctly applied.** The
rule is *content lives where the work lives*; thin pointers are what that **produces** when a repo
artifact exists, not the rule itself. math-quest's own `linear-bug-report` skill already states the
carve-out: *"The contentless-pointer rule in `CLAUDE.md` governs **epic and story** pointers, whose
content lives in the repo. A bug has no repo artifact behind it, so its detail belongs in the issue
itself… There is nothing for a stub to point at."* trellis is that bug-shaped case applied to a whole
repo — the GitHub body *was* the content, and it is now closed — so each Linear issue carries a real
summary plus a link.

*An earlier draft of this record called this a deliberate divergence. That was wrong, and wrong in a
way that mattered: framing a correct application as a divergence invites a future session to "fix" it
back to stubs. Corrected on the evidence above, supplied by the session that wrote the original rule.*

**The drift this rule exists to prevent, stated so it is watched for:** the day a trellis issue grows
a repo artifact behind it, that issue's body must **become** a pointer rather than staying a second
copy. Dual-homing does not announce itself (`inv-deliberate-succession`, backward direction).

## Consequences

- **The backlog was triaged before porting, not after.** 27 issues → 19 ported, 8 closed (30%): two
  completed (the work had landed, both clauses verified against the working tree), one on a dead
  premise (`/trellis:setup`, retired by `decision-0072`), and five parked as ideas. The line used:
  *is this issue the unpaid debt of a decision that was actually made?* Yes → port. No → idea.
- **`stage:` labels were not ported.** Porting a field that records filing date as if it recorded
  progress would make the new board look triaged when it is not. State was set deliberately per
  issue instead.
- **This drops a family convention for one repo.** `kodhama:issues` implements the family record
  `kodhama-0026-issue-taxonomy`, and trellis leaving GitHub issues makes that record inapplicable
  here. **This decision does not amend the family record** — it cannot; that is a kodhama-level act.
  The maintainer's stated intent is to move the family to Linear progressively, trellis first. Until
  that lands, trellis is a declared exception rather than a silent one.
- **Old issue numbers resolve forever**, so no `decisions/` citation was touched, and none needed to
  be. Verified after the port: `#165`, `#166`, `#212` all still resolve.
- **The Kodhama team is owed an archive.** It holds only Linear's four onboarding stubs. Team
  structure has no MCP mutation, so moving Trellis and Math Quest out and archiving it is UI work,
  and is not done by this change.
- **Nothing in trellis resolves a Linear team by name yet, and that is worth keeping true.** Verified
  at migration: **no hook, skill or test performs a Linear team lookup at all**, by name or
  otherwise. The literal string `Trellis` is of course everywhere — it is the product's name — and
  no count of it is stated here, because two earlier drafts of this bullet each stated one and each
  got it wrong. The claim that matters is the absence of a lookup, not the presence of a word.
  math-quest's
  automation resolves its team by the literal `Math Quest` in eight places, where a **rename** —
  not a key change — would break resolution silently, because an unmatched team yields nothing found
  rather than an error. If trellis acquires Linear automation, resolve by team **id**, not name.
- **Research issues carry no type label.** Linear's set is Bug / Feature / Improvement; a research
  question is none of them. Leaving the label off was judged better than minting one for four
  issues — revisit if the count grows.

## Open questions

- **Does the ideas document stay honest, or become a graveyard?** math-quest's answer is the
  promotion trigger on every entry. Untested here.
- **What happens to `kodhama-0026-issue-taxonomy` when the last repo leaves GitHub issues?** Retired,
  or rewritten against Linear. Not this record's to decide.
- **Is `Critical` reachable for a process-tooling product at all?** Its definition is outage, data
  loss or security exposure. If no trellis defect can ever qualify, the four-level scale is
  effectively three here.

## Self-check (gate)

The prior art was confirmed with the session that produced it rather than read from its artifacts,
which caught a stale claim that would otherwise have been inherited (`inv-independent-judgment`).
What looked like a divergence from math-quest's pointer rule was checked against that rule's own
source and turned out to be the rule correctly applied; the record says so and keeps the wrong
reading visible rather than quietly replacing it. The family-convention exception is declared rather
than left to be discovered, and `AGENTS.md` gains the filing rule in this same change — without it a
future agent would read the old convention and keep filing GitHub issues (`decision-0028`,
`inv-deliberate-succession`, `floor-transparency`). Two mapping judgements — severity levels and
unlabelled research issues — are recorded as the author's calls, not presented as derived. Flipped to
`approved` on the maintainer's intent act after reading the drafted framing, not on the author's
judgement (`decision-0046`); the record of that act is in the frontmatter.

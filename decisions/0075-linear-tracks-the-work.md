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

**3. Ideas are a document, not issues.** An idea filed as an issue is a to-do nobody agreed to; an
idea written as prose is a claim someone can challenge. Each entry carries the **trigger** that
would promote it — a condition that revisits its priority, distinct from an *open question*, which
is something undecided about work already scoped and is resolved by doing that work.

**4. Closed GitHub issues carry a banner in the body, not a comment.** Every migrated issue was
closed with a banner prepended to its **body**, marked `<!-- trellis:linear-migration -->` so
re-running is idempotent. A closing comment can be scrolled past, and a body containing standing
instructions keeps reading as true after it closes. Three banner texts, because they closed for
three different reasons — *moved*, *parked as an idea*, *resolved*. A "moved to Linear" banner on an
issue that was actually resolved would be false, and these are permanent.

**5. A pointer carries content here, diverging from math-quest deliberately.** That repo's rule is
that a Linear issue is a thin pointer, because a repo artifact holds the content. trellis has no such
artifact — the GitHub issue body *was* the content, and it is now closed. So each Linear issue
carries a real summary plus a link to the original. The rule was not inherited by default
(`inv-deliberate-succession`).

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
The divergence from math-quest's pointer rule is named with its reason rather than taken silently,
and the family-convention exception is declared rather than left to be discovered
(`inv-deliberate-succession`, `floor-transparency`). Two mapping judgements — severity levels and
unlabelled research issues — are recorded as the author's calls, not presented as derived. Flipped to
`approved` on the maintainer's intent act after reading the drafted framing, not on the author's
judgement (`decision-0046`); the record of that act is in the frontmatter.

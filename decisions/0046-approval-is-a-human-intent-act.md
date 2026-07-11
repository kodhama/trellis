---
id: decision-0046
type: decision
status: gated  # self-checked 2026-07-11; ratified post-merge (old mechanic — see Self-check)
depends_on: [decision-0042, decision-0037, spec-0001, invariants-v1]
owner: agent
date: 2026-07-11
---

> Shaped interactively with the maintainer (2026-07-11), resolving
> kodhama/trellis#142 — the **intent-layer** approval-act half of
> `spec-0001` §2's deferred "execution-layer `approved`" question (the
> does-code-conform half is trellis#25). The maintainer chose the
> verifiable-in-PR-flip *model* but to **retire the guard now** and defer a
> family-wide guard to grove#38 — streamlining over building refined
> machinery for trellis-self alone.

# 0046 — the approval act is a human intent act, not the merge; retire ratify-guard, defer a family-wide guard

## Context

The current mechanic (`decision-0022` + `decision-0042`): *"merging is the
ratification act; a post-merge bump commit records `approved`; nobody
writes `approved` into a PR's own diff for a new artifact"* — and
`ratify-guard` enforces the last part (it fails a *ready* PR that adds a
decision/spec already `status: approved`).

This uses **"only a merge can flip to `approved`"** as a *proxy* for
**"only the human can approve"** (`floor-intent-gate`). It works — it keeps
agents from self-approving — but at a cost: it **conflates the approval
*act* with the merge**, and mislabels a real conversational human approval
as not-yet-approved. Worked example from grove#22 (the originating
exchange): the maintainer's instruction *"execute it and fold it into the
PR"* **was** the approval act, but under the proxy the artifact stayed
`gated` and execution *looked* like it preceded approval — when it followed
it. The proxy mislabeled a real human approval as not-yet-approved.

## Decision

1. **The approval act is a human intent act, not intrinsically the merge.**
   A human's approval — in conversation, in review, or by merging — is the
   ratification act (`floor-intent-gate`; `spec-0001` §2: "some ratified
   state is a human act — or a human-authorized, recorded ratchet").
   Flipping `status: approved` **records** that act; the merge is **one way
   to perform it, not the only way**.

2. **An agent writing `approved` with no human act is forbidden.** This —
   not "in-PR vs post-merge" — is the real line `floor-intent-gate` draws
   and the thing `ratify-guard` was really trying to catch. The distinction
   that matters is **human-act vs agent-act**.

3. **In-PR `gated → approved` flips are legitimate when they record a human
   approval act** — the same practice grove already uses (adr-0004/0005/
   0006). trellis adopts it, so the family runs **one** approval mechanic,
   not two.

4. **Retire `ratify-guard` (for now).** Under the intent-act model the
   guard's "a new artifact must not arrive `approved`" rule is superseded —
   it would now block a *legitimate* human-authorized flip. Rather than
   build a refined machine-checkable guard (e.g. gating an in-PR flip on a
   GitHub PR-approval review) **just for trellis-self**, retire it; the
   human's **review-before-merge** is the backstop against agent
   self-approval — the same backstop grove already relies on, guard-less. A
   family-wide re-hardening is **deferred, not dropped** (grove#38).

5. **Clarify intent when ambiguous.** When it is ambiguous whether a human
   instruction *is* the approval act, **clarify before treating it as
   one** — an agent must not **infer** the gate has opened, nor **stall** a
   real approval by failing to recognize it. (Companion intent-gate rule,
   directly from the originating exchange.)

## Considered and rejected

- **A refined verifiable guard now** (allow an in-PR `approved` flip *iff* a
  machine-checkable human PR-approval review exists) — **deferred, not
  adopted.** It is the right *shape* for a family-wide rollout, but building
  and maintaining it for trellis-self alone is heavier than the
  streamlining is worth today. Homed as **grove#38** for the family-wide
  consideration (Consequences).
- **Keep the guard + reframe only** (the minimal option) — rejected: it
  leaves the artifact reading `gated` in the PR window, so the mislabeling
  trellis#142 names persists. The maintainer chose the in-PR-flip model.
- **Merge = the sole approval act** (status quo) — rejected at the root: it
  conflates a human intent act with a mechanical landing.
- **Two family mechanics** (grove flips in-PR, trellis merges-then-bumps) —
  rejected: `kodhama-0004` already set uniformity as the family direction;
  one mechanic is simpler and honest.

## Consequences

1. **`ratify-guard.yml` is removed** in this decision's PR. This
   **supersedes-in-part `decision-0042`** — its "post-merge bump / no
   in-diff `approved` for a new artifact" rule and the guard that enforced
   it. `decision-0042`'s core — *merge is A valid ratification
   performance*, and the family `draft → gated → approved` enum — **stands**
   (the merge is no longer the *only* way, not "no longer valid").
   `decision-0042` gains `superseded_in_part_by: [decision-0046]`.
2. **grove#38 opened** — consider a family-wide *verifiable* approval guard
   (or something like it) rolled out to all kodhama repos: the deferred
   re-hardening.
3. **`spec-0001` needs no amendment.** Its *portable* contract already says
   "some ratified state is a human act — or a human-authorized, recorded
   ratchet" (§2). This decision refines trellis's **self-application**
   mechanic (`decision-0022`/`0042` + the guard), not the portable
   contract.
4. **This decision is itself ratified under the *old* mechanic** — shipped
   `gated`, merged (ratification), then a post-merge bump to `approved`. The
   new in-PR-flip rule cannot apply to its own ratification before it
   exists; this is the **last** trellis artifact ratified the
   post-merge-bump way. Henceforth, in-PR flips.
5. **Pairs with trellis#25** (execution-layer `approved` — does an
   implementation conform to its upstream). This decision settles the
   **intent-layer** half; #25 is the execution-layer half.

## Open questions (parked, ≤3)

- **Family-wide guard shape** — GitHub-PR-review gate vs signed-commit
  marker vs other; and whose channel (grove-propagated CI vs kodhama
  rollout vs trellis overlay), given a CI workflow is not the charter
  channel grove currently propagates through. Parked to **grove#38**.
- **The clarify-when-ambiguous rule as agent behavior** — it is an
  agent-behavior rule; it may belong stated in grove's agent charters /
  managed CLAUDE.md primer, not only in this trellis decision. Flagged for
  the grove charter-execution / propagation pass (adr-0006).

## Self-check (gate)

- **Frontmatter**: `id`/`type`/`status`/`depends_on`/`owner`/`date`
  present, well-typed. PASS.
- **`depends_on` resolution**: `decision-0042` (`approved`), `decision-0037`
  (`ratified`), `spec-0001` (`ratified`), `invariants-v1` (`ratified`) —
  all resolve, none `draft`. PASS.
- **Directional flow**: this is `gated`; every dependency is
  ratified/approved, not draft. PASS.
- **Required body sections** (`spec-0001` §4): Context/Decision/Consequences
  present, plus Considered-and-rejected, Open questions, Self-check. PASS.
- **Append-only discipline**: new artifact; nothing edited in substance;
  it **supersedes-in-part `decision-0042`**, which gains
  `superseded_in_part_by: [decision-0046]` — a marking, not a substance
  edit (`spec-0001` §2 / `decision-0040`). PASS.
- **Approval mechanic**: left `gated`, not flipped. Ratified the **old**
  way (merge + post-merge bump) — this decision cannot ride its own new
  rule before it exists (Consequences 4). `ratify-guard` passes a `gated`
  artifact, so this PR is guard-clean even before the guard's removal takes
  effect. PASS.

**Overall: internally sound, consumable, and `gated`** — self-checked,
awaiting the maintainer's approval and the post-merge bump, which closes
kodhama/trellis#142 (the intent-layer half) and hands the family-wide guard
to grove#38.

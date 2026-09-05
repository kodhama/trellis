---
id: decision-0046
type: decision
status: approved  # 2026-07-11 — in-PR flip recording the maintainer's pre-merge intent act; the first decision ratified under its own rule (see Self-check)
depends_on: [decision-0042, decision-0037, spec-0001, invariants-v1]
superseded_in_part_by: [decision-0082]  # decision-0082: D1-D4 — the `status: approved` flip that recorded the act, the in-PR flip rules, and the kept draft-landing half of ratify-guard all go with the field. STANDS: D5 (clarify before treating an ambiguous instruction as approval, and do not stall a real one) as general agent behavior, and the principle that the approval act is the human's, not the merge mechanic's
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

4. **Retire `ratify-guard`'s *self-approval* check; keep its *draft-landing*
   check.** The guard did two things: **(a)** fail a ready PR that leaves a
   `status: draft` decision/spec on `main`, and **(b)** fail a *new*
   artifact born `approved`. Only **(b)** — the merge-as-approval proxy —
   is in tension with the intent-act model (it would now block a
   *legitimate* human-authorized flip), so **(b) is retired**. **(a)
   stays** — it enforces `decision-0042`'s still-standing "no draft on
   `main`" core, which this decision does *not* supersede (dropping it
   would be collateral damage, not a decision). Rather than build a refined
   machine-checkable *approval* guard (e.g. gating an in-PR flip on a GitHub
   PR-approval review) **just for trellis-self**, the self-approval
   enforcement is simply dropped; the human's **review-before-merge** is
   the backstop against agent self-approval — the same backstop grove
   already relies on, guard-less. A family-wide re-hardening of the
   *approval* guard is **deferred, not dropped** (grove#38).

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

1. **`ratify-guard.yml` is slimmed** in this decision's PR — its
   self-approval check removed, its **draft-landing check kept**. This
   **supersedes-in-part `decision-0042`** — its "post-merge bump / no
   in-diff `approved` for a new artifact" rule and the self-approval
   enforcement. `decision-0042`'s core — *merge is A valid ratification
   performance*, the family `draft → gated → approved` enum, and **"no
   draft on `main`"** — **stands** (the merge is no longer the *only* way;
   the draft guard is untouched). `decision-0042` gains
   `superseded_in_part_by: [decision-0046]`.
2. **grove#38 opened** — consider a family-wide *verifiable* approval guard
   (or something like it) rolled out to all kodhama repos: the deferred
   re-hardening.
3. **`spec-0001` needs no amendment.** Its *portable* contract already says
   "some ratified state is a human act — or a human-authorized, recorded
   ratchet" (§2). This decision refines trellis's **self-application**
   mechanic (`decision-0022`/`0042` + the guard), not the portable
   contract.
4. **This decision is ratified under its *own* new rule** — the **first**
   trellis artifact so ratified. The maintainer's pre-merge approval
   (2026-07-11) is the human intent act; the `status: approved` flip **in
   this PR** records it; the merge lands it. There is no bootstrap paradox:
   `floor-intent-gate` is satisfied by the human act itself, which does not
   depend on this decision — the decision only recognizes that the act can
   be performed pre-merge. The slimmed guard (no self-approval check)
   permits the flip. The decision demonstrates its own mechanism.
5. **Pairs with trellis#25** (execution-layer `approved` — does an
   implementation conform to its upstream). This decision settles the
   **intent-layer** half; #25 is the execution-layer half.
6. **Dependent operating docs updated in this PR** — the guard change's
   blast radius, per "update everything that depends on it":
   - `CLAUDE.md`'s ratification rule — the "never write `approved` in the
     PR itself / a post-merge bump records it" clause is **reversed**
     (in-PR flips recording a human act are legitimate); the "no draft is
     left on `main`" clause **stands**.
   - the `conformance-reviewer` charter's trellis-vendored
     `<PR_CONTRACT_SECTIONS>` resolution — drops the retired "must not add a
     new artifact already `approved`" clause; keeps the draft clause. (This
     is a trellis placeholder-resolution, not grove-canonical body — grove's
     charter carries a `<PR_CONTRACT_SECTIONS>` placeholder there — so
     editing the trellis copy causes no canonical drift.)
   - the other three grove-vendored agent files (`run-resumer`,
     `propagation-remediator`, `agents/README`) reference `ratify-guard`
     *generically* as a must-pass check — **still valid**, since the
     slimmed guard still runs. Any deeper mechanic-description in these
     grove-vendored copies otherwise propagates via grove (the charter
     execution / re-vendoring), not per-repo hand-edits.

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
- **Append-only discipline**: new artifact; no append-only *decision*
  edited in substance. It **supersedes-in-part `decision-0042`**, which
  gains `superseded_in_part_by: [decision-0046]` — a marking, not a
  substance edit (`spec-0001` §2 / `decision-0040`). The `CLAUDE.md` and
  `conformance-reviewer` charter edits (Consequences 6) are to
  **revise-in-place operating docs**, not append-only artifacts — a
  legitimate in-place update, the same class `decision-0042` made to
  `CLAUDE.md`. PASS.
- **Approval mechanic**: flipped to `approved` **in this PR**, recording
  the maintainer's pre-merge intent act (2026-07-11) — the decision's own
  new rule, applied to itself (Consequences 4). The slimmed guard has no
  self-approval check, so the flip is guard-clean. PASS.
- **Adversary round**: `spec-adversary` returned NEEDS-REVISION. Premise
  cleared (`floor-intent-gate` intact — it governs gatekeeper *identity*,
  not mechanical enforcement; `spec-0001` §2 no-amendment verified;
  supersession coherent; no bootstrap paradox). Two findings fixed: **(A)**
  the guard's live dependents were left dangling → the two that describe the
  *retired* check (`CLAUDE.md`, `conformance-reviewer`) are updated here,
  the three generic ones stay valid (Consequences 6); **(B)** deleting the
  whole guard silently dropped `decision-0042`'s still-in-force
  draft-landing check → the guard is now **slimmed, not removed** (draft
  check kept, Decision 4). PASS (revised).

**Overall: internally sound, consumable, and `approved`** — self-checked,
survived two adversary rounds (revised), ratified by the maintainer's
pre-merge intent act via in-PR flip (its own new rule, applied to itself).
Closes kodhama/trellis#142 (the intent-layer half); hands the family-wide
guard to grove#38.

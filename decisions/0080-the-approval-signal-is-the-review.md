---
id: decision-0080
type: decision
status: gated  # authored + self-checked by agent; the approved flip is the maintainer's intent act (decision-0046)
depends_on: [decision-0046]
informed_by: [decision-0022, decision-0076, decision-0078]
owner: agent
date: 2026-08-29
---

# 0080 — The approval signal is the review, not the merge

## Context

`decision-0046` made the approval act a **human intent act** and left recording it a hand
edit. It named one backstop against agent self-approval: *"the human's review-before-merge …
the same backstop grove already relies on, guard-less."* It considered a machine-checkable
guard — *allow the flip iff a human PR-approval review exists* — and **deferred it to
grove#38**.

Two things have since broken that arrangement.

- **The backstop does not hold on the path that now dominates.** Agents in this repo act with
  the maintainer's GitHub token, so an agent merge is recorded as `merged_by: gundisalwa` and
  is **indistinguishable from the maintainer's**. On 2026-08-29 an agent opened, reviewed
  nothing, and merged `#255` and `#256`; both read as maintainer merges. There was no
  review-before-merge, and nothing could have observed its absence.
- **The deferral lost its consumer.** `decision-0076` retired grove from this repo, so
  grove#38 no longer re-presents anything here. Under `decision-0078` that is an orphan
  follow-up — the exact shape that invariant forbids.

The maintainer's report of the friction was that *nothing automatically flips the status*.
The measurement says the opposite is the real defect: **the gate has no sensor**, and
automating a flip on merge would have made an unverified approval automatic and invisible —
strictly worse than a hand edit an agent must consciously choose to write.

## Decision

**1. The approval signal is an APPROVED pull-request review from an account that is not the
PR author.** Not the merge. A merge is a delivery act that any agent holding the token can
perform; a review by a *different* account is the smallest signal this repo can actually
check. This does not narrow `decision-0046` — approval in conversation remains a valid intent
act, recorded by an in-PR flip as before. It gives the *automated* path one verifiable trigger
instead of none.

**2. `ratify-flip` records that signal, and only that.**
`.github/workflows/ratify-flip.yml` runs on merge to `main`, finds decision records the PR
left at `gated`, and flips them to `approved` **only** when such a review exists — stamping
the approver and PR number into the record. With no qualifying review it flips nothing, warns
on the run, and leaves the records `gated`.

**3. It is dormant today, on purpose, and that is the honest state.** Agents author, review
and merge as the maintainer, so *"not the author"* is never satisfied and the workflow will
never flip anything. It will report gated records and stop. **A guard that flipped anyway
would certify agent self-approval as the maintainer's** — the precise failure
`floor-intent-gate` exists to prevent.

**4. The activating change is an agent identity, and it is the maintainer's to make.** Giving
agents their own GitHub App or bot account makes `merged_by` and the review author differ from
the maintainer, at which point this guard begins enforcing with **no change to the workflow
file**. Recorded here as the named next step; its consumer is this record plus the workflow's
own warning on every merge that leaves a gated record, which satisfies `decision-0078` —
neither is "someone reading this file later."

## Consequences

- **The grove#38 orphan is discharged.** The guard `decision-0046` deferred exists.
- **`gated` stops being a state records rot in silently.** Every merge that leaves one now
  says so, loudly, on the run.
- **Cost accepted:** until an agent identity exists, the flip stays manual for
  conversationally-approved records — unchanged from today. This buys a correct sensor, not
  less typing.
- **`ratify-guard` is untouched.** It still fails a ready PR that leaves a `draft` record;
  this workflow governs `gated → approved` after merge. The two do not overlap.

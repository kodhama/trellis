---
id: decision-0090
type: decision
depends_on: [decision-0089]
changes: [decision-0089]
informed_by: [decision-0028, decision-0081, decision-0085]
owner: agent
date: 2026-09-04
---

> **Provenance.** A review pass over `kodhama/trellis#266` after it merged, on **TRL-40**. Two of the
> five findings were **false passes reproduced against the running script**, not readings of it. The
> guard `decision-0089` shipped is corrected here rather than edited in place, because `#266` merged
> before the review landed and `decisions/` is append-only once merged.

# 0090 — a decision-id claim is judged by its destination, and a branch can collide with itself

## Context

`decision-0089` shipped `decision-id-guard` and defined a claim in its point 4:

> **A claim is an *added* file with the `decisions/NNNN-*.md` shape — nothing else.** Modifying
> `decision-0087` does not claim `0087`; **a rename is a file that already exists**; …

**Two branches pass that definition while colliding.** Both were run against the shipped script
before being accepted, not inferred from reading it:

| | Fixture | Shipped behaviour |
| --- | --- | --- |
| **A** | one PR adds `decisions/0090-a.md` **and** `decisions/0090-b.md` | `PR #300 claims: 0090 0090` … `— free.` twice, **exit 0** |
| **B** | PR #300 renames `0086-old.md` → `0090-new.md` while open PR #299 adds `decisions/0090-rival.md` | `PR #300 adds no decisions/NNNN-*.md file`, **exit 0** |

**A is a gap in coverage, not in reasoning.** The base check and the rival check each compare this
PR's ids against *another* source. Nothing compared the branch against itself, so a diff carrying one
id twice had no check pointed at it.

**B is a reasoning error, and the phrasing shows where it went wrong.** *"A rename is a file that
already exists somewhere"* is true of the **old** path and false of the **new** one. GitHub reports
a rename with the destination in `filename` and the source in `previous_filename`; renaming
`0088-old.md` to `0090-new.md` puts a file at `decisions/0090-*.md` exactly as adding it would.
`decision-0089` reasoned about the wrong end of the operation and excluded the whole status.

**The shape both share is worth more than either instance.** A guard is only as good as the cases it
declines to treat as claims, and both misses were in that declining — an over-wide exclusion is
**silent by construction**: it produces a green tick and no output at all. `decision-0089`'s own test
suite had a case named `TestDecisionIDGuardIgnoresNonAddedFiles` that *asserted* the rename false
pass as correct behaviour.

**Three further paths would have failed open**, each read out of the source and then confirmed:

- **`gh pr list --limit 200` caps silently.** Past the cap it returns a short list and exits 0, so an
  id held by an unlisted PR reads as free — contradicting `decision-0089`'s own *"every open PR"*
  claim with no trace in the log.
- **`set -u` is not `pipefail`, and this script is `sh`.** `awk … | sort -u > mine` takes the exit
  status of `sort`; a failing `awk` left an empty file, and the guard printed *"no id claimed,
  nothing to check"* and exited 0. An internal error rendered as a clean pass.
- **The workflow/script pair test asserted substrings that also appear in the workflow's own
  comment.** `.github/scripts/decision-id-guard.sh` occurs at `decision-id-guard.yml:11` (prose) and
  `:26` (the `paths:` list), so rewriting `run:` to `run: true` left the test green while CI invoked
  nothing — the exact failure `decision-0089` cited that test as preventing.

## Decision

**1. `decision-0089` point 4 is replaced. A claim is a file the branch puts at a *new*
`decisions/NNNN-*.md` path, judged by the destination.**

- `added` and `copied` claim outright.
- **`renamed` claims its destination id.** `0088-old.md` → `0090-new.md` claims `0090`.
- **A slug-only rename claims nothing.** `0087-one-gateway.md` → `0087-a-better-slug.md` keeps its
  own number. This is why the rule cannot simply be *"renamed counts as added"*: `0087` is on the
  base branch by definition, so counting it would fail every legitimate retitle.
- `modified`, `removed`, `changed` never claim. Editing `decision-0087` does not take `0087`.
- `decisions/README.md` and `decisions/0089.md` are not claims — four digits and a dash is the whole
  shape. *(Unchanged from `decision-0089`.)*

**2. Whether a rename *releases* the source id is left undecided, in the code comment as well as
here.** Once the PR merges the old path is gone and the number looks free; until then the source
record is still on the base branch. Handing that number to a second branch on the strength of an
unmerged rename would cause the collision the guard exists to prevent, so **the source id stays
taken** — wrong, if ever, by one wasted number rather than by a duplicate. Named rather than silently
resolved, which is the disposal `decision-0089` did not give it.

**3. A branch may collide with itself, and that is red on that PR alone.** Two files claiming one id
in one diff draws
`::error::PR #300 adds two files claiming decision-0090: decisions/0090-a.md decisions/0090-b.md — one id, one record.`
**No tie-break applies** — both files are on the same branch, and the author picks which one moves.
This extends `decision-0089` point 1; its points 2 (the tie-break) and 3 (state the rule, name the
winner) are untouched.

**4. A processing failure is exit 2, never exit 0.** `pipefail` is not portably available in `sh`
(ubuntu-latest's `/bin/sh` is dash), so every processing step writes to a file whose status is
checked instead. Open PRs are enumerated with
`gh api "repos/$repo/pulls" -f state=open -f per_page=100 --paginate`, which has no cap; both forms
were verified to return the same list against this repository (`266 264 263 262 245 208`) before the
swap.

**5. Every "this shape is not a claim" arm carries its own test.** That is the rule the two false
passes buy, and it is the one worth keeping: an exclusion that is too wide emits nothing, so only a
test aimed at it can see it. `TestDecisionIDGuardIgnoresNonAddedFiles` was rewritten — it had
asserted the false pass — and split, so `renamed` now has a test on each side of its rule.

## Consequences

- **`decision-0089` is changed in part, not superseded.** Its points 2, 3, 5, 6 and 7 stand exactly
  as written; point 1 gains a third clause and point 4 is replaced. The guard, the tie-break, the
  script-not-YAML argument and the closure of `decision-0078`'s parked observation are all
  unaffected. The forward pointer on `0089` records that scope.

- **A rename into a fresh id now trips the guard on branches that previously passed.** That is the
  intended behaviour and it is a real behaviour change: a branch that renumbers a record by renaming
  its file — which is exactly what a collision remedy looks like — is now checked, where before it
  was invisible. Renaming *to an id that is genuinely free* stays green.

- **Three mutations pin the fixes** (`decision-0089`'s own standard, applied to its corrections):
  dropping the duplicate scan turns the self-collision test red and makes the mutant print `— free.`
  twice for one id; dropping the `renamed` arm turns the rename test red with *"adds no
  decisions/NNNN-\*.md file"*; rewriting `run:` to `run: true` turns the workflow-pair test red,
  where the old assertion stayed green.

- **The live `gh` path is still not unit-tested**, and the pagination fix lands inside it. It is
  pinned by a *source-level* assertion — no `gh pr list` command, `--paginate` present — which binds
  the fix and not the behaviour. Labelled as such in the test rather than left to look like
  behavioural coverage.

- **This record exists because `#266` merged before the review landed.** Had the review arrived on
  the open PR, the same corrections would have been an edit to an unmerged record and no second
  record would exist. Under `decision-0081` that is the cheap-reversal case; once merged, the
  append-only rule applies and the cost is one more file. **The lesson is not "review faster" but
  that the merge is the point where correction cost jumps**, which is the same fact `decision-0089`
  is built on — merging is what makes an id, or a rule, real.

## Self-check (gate)

- **The two false passes were reproduced, not accepted on a reviewer's word.** Both fixtures were run
  against the shipped script and their actual output is quoted in the table above; the controls
  (`lower-numbered rival`, `id already on base`) were run in the same sweep and stayed red, so the
  fixes are not a general loosening. The `run: true` finding was confirmed by checking that the
  script path still appears at `decision-id-guard.yml:11` and `:26` with the `run:` step gone.

- **The over-wide exclusion is stated as the general defect, and this record contains an instance of
  it.** The test forbidding `gh pr list` first went red against *the script's own comment explaining
  why it does not use `gh pr list`* — the identical self-match trap `agent-workflow-parity.yml`
  documents at its `USES_RE` (*"This file names it in its own comments; a substring match flags
  itself."*). Reading that warning in a neighbouring file did not stop it recurring one file over,
  which is that workflow's own thesis about point-of-use comments not transferring.

- **Point 2 is a disclosed non-decision, not an omission.** `decision-0089` was corrected for
  generalising past its mandate (`decision-0085`'s defect); this record therefore names the rename-
  release question and declines it, in the record *and* in the code comment, rather than picking an
  answer nobody asked for.

- **Not independently reviewed at the time of writing.** The author wrote the corrections, their
  tests and this record. The `corpus-reviewer` run and the PR review are the independent passes
  (`inv-independent-judgment`) — and on `decision-0089` those two passes found five defects between
  them that the author's own reviews had not, which is the reason this section claims nothing on
  their behalf.

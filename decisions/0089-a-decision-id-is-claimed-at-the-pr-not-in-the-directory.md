---
id: decision-0089
type: decision
depends_on: [decision-0078, decision-0082]
changes: [decision-0078]
informed_by: [decision-0005, decision-0010, decision-0028, decision-0077, decision-0085]
owner: agent
date: 2026-09-03
---

> **Provenance.** Filed as **TRL-40** by an agent session from the `TRL-33` branch — the third
> recurrence — after a `corpus-reviewer` pass found it. The maintainer ruled on 2026-09-03: *a CI
> check on the PR*, with the lower-numbered open PR keeping a contested id. The two other shapes the
> ticket offered (allocate at merge; make the renumber cheap) were **not taken**.

# 0089 — a decision id is claimed at the pull request, not in the directory

## Context

**`decisions/` on `main` is not the allocation authority; the merge queue is.** Every branch that
collided did the correct thing in isolation — list `decisions/`, take the next free number. A branch
cut before a sibling merges is reading a stale world correctly, and nothing in the repository told
it so.

Three recurrences, each recorded at the time:

| # | Collision | How it surfaced |
| --- | --- | --- |
| 1 | `decision-0077` claimed by two branches at once | the loser noticed CI had never reported, went to `main`'s log, and found the other record already merged |
| 2 | trellis#252 open on an already-taken `decision-0076` | not recorded — `decision-0078` states the collision as a standing fact while arguing #1 was not a one-off |
| 3 | `decision-0086` claimed by both trellis#262 (TRL-33) and trellis#263 (TRL-29) | a coordinating session noticed; #262 renumbered to `0087` before merge |

**`decision-0078` recorded #1 and #2 and armed a trigger**, at
`decisions/0078-no-orphan-followups.md:147-156` *(TRL-40 cites it as `:146-155`; the forward pointer
this change adds to that record's frontmatter displaced it by one — `decision-0085`'s treatment of
the same problem)*: *"it is dropped, on this record, because no one has
agreed to do it; if it recurs a third time, that is the trigger to file it."* That is the rule of
`inv-no-orphan-followups` applied to itself — a deferral with a named consumer and a stated
condition. The condition fired, so this record is the discharge of an obligation, not a new
proposal.

**The cost of a collision is small each time and entirely wasted.** A renumber touches the id, the
filename, the title, every forward and back pointer, the planning artifacts, the PR body and the
issue comments — seven files on recurrence #3. None of it is work; all of it is bookkeeping caused
by two branches being unable to see each other.

**Why a local check could not have caught any of the three.** A pre-commit hook, a Go test, a
`corpus-reviewer` pass — each reads the working tree and `main`, which is exactly the stale world
the losing branch already read. Seeing a sibling *branch's* claim needs the GitHub API, which no
check in this repository used before this one.

## Decision

**1. A CI check on the pull request, `decision-id-guard`.** It fails when a `decisions/NNNN-*.md`
file **newly added** by the branch carries an id that is already on the base branch, or that is also
newly added by a **lower-numbered open pull request**.

**2. The tie-break is "the older claim was there first".** Two open PRs claiming one id is a
collision on both, but only the **higher-numbered** one goes red. Failing both would leave neither
able to merge without first talking to the other, which is the coordination cost the check exists to
remove. The lower-numbered PR gets a `::notice` and stays green.

**3. The check states the rule and names the winner in its own output.** A red that only says
"taken" sends the reader back to the merge queue to work out who moves. Concretely, PR #300 against
an older #299 gets:

> `::error file=decisions/0089-the-newer-claim.md::decision-0089 is also claimed by open PR #299`
> `(decisions/0089-the-older-claim.md). TIE-BREAK: the older claim wins — the lower-numbered open PR`
> `keeps the id, so #299 keeps decision-0089 and this PR (#300) renumbers.`

and #299, against a newer #301, gets the mirror image as a notice: *"this PR (#300) keeps
decision-0089 and #301 renumbers. Reported here, red there."*

**4. A claim is an *added* file with the `decisions/NNNN-*.md` shape — nothing else.** Modifying
`decision-0087` does not claim `0087`; a rename is a file that already exists; `decisions/README.md`
and `decisions/0089.md` are not claims. This is why the status filter lives in the script rather
than in the workflow's `gh --jq` expression: the rule has to sit in the half the tests can reach.

**5. The logic is a script, not inline YAML** — `.github/scripts/decision-id-guard.sh`, executed by
`.github/workflows/decision-id-guard.yml` and by `cli/decision_id_guard_test.go`. Every input is
injectable (`GUARD_MAIN_FILES`, `GUARD_PR_FILES`, `GUARD_PR_NUMBER`), so the tests run offline
against fixtures while the workflow feeds the same script real `git` and `gh` output. This is the
repo's existing pattern for external scripts — `cli/plugin_hook_test.go` on
`plugins/trellis/hooks/staleness.sh`, `cli/install_script_test.go` on `install.sh` — and it is the
reason `go test -count=1` is not optional here.

**6. `decision-0078`'s parked observation is closed, and only that.** It gains
`superseded_in_part_by: [decision-0089]` scoped to the dropped observation at `:147-156`, plus a
dated note beside the original claim (`decision-0085` §5 shape: a note beside the claim, never an
edit that makes the record look prescient). **`inv-no-orphan-followups` itself is untouched** — this
record does not reopen the rule, only the example it parked.

## Consequences

- **The check is advisory, like `agent-workflow-parity`.** `main` carries no branch protection here,
  so a red does not block a merge; it replaces a human noticing. The maintainer can merge a red PR
  deliberately — for instance when he wants the *higher*-numbered PR to keep the id after all.

- **It runs only on PRs that touch `decisions/`** (plus the guard's own two files, so a change to it
  runs against itself). A branch that adds no record cannot claim an id, and the guard says so in one
  line rather than staying mute: *"PR #N adds no `decisions/NNNN-*.md` file — no id claimed, nothing
  to check."*

- **Two ids on one branch are judged separately.** A branch adding `0088` (taken) and `0090` (free)
  is red for `0088` and prints `- decision-0090 (…) — free.` for the other, so "renumber" is never
  ambiguous about which file.

- **The tie-break can still cost a renumber; it just makes it deterministic and early.** The
  higher-numbered PR moves. What is removed is the discovery cost: not one of the three was found by
  anything failing. #1 was found because CI happened not to report, which sent its author to `main`'s
  log — and by then the rival record had already merged. #3 was found by a coordinating session
  reading two branches at once. #2's discovery route is not recorded anywhere, which is itself the
  point: nothing was watching.

- **The guard is only as fresh as the last CI run on each PR.** PR #299 merging does not re-run the
  check on #300; #300 sees it on its next push, or at merge time if the maintainer re-runs. This is
  the same staleness `agent-workflow-parity` carries and is disclosed rather than fixed — a
  `pull_request_target` cron over every open PR is a bigger mechanism than the problem.

- **The workflow/script pair is guarded** (`decision-0028`): `TestDecisionIDGuardWorkflowRunsTheScript`
  asserts the workflow still invokes the tested script with the inputs it refuses to run without. Without
  that assertion the Go suite would keep passing while CI ran nothing — the failure mode
  `agent-workflow-parity` was built for, one file over.

- **The live `gh` path is not covered by the tests, and this is a real gap.** The tests build the
  script's environment from scratch, which proves they never reach the network — and equally proves
  they never exercise `gh pr list` / `gh api`. The parse-and-compare half is fully covered; the
  fetch half is verified only by real runs. Its first one, on this record's own PR (#266), fetched
  `main`, walked all six open PRs and printed `- decision-0089 (…) — free.` The alternative (a `gh`
  stub on `PATH`) would test the stub.

- **This is Layer B, and it does not touch `decision-0010`'s no-runtime rule.** The guard lives at
  the repo root — the methodology used to *build* Trellis (`decision-0005`) — not in `core/`, and
  nothing ships it to a consumer. `decision-0010` permits exactly this: *"Any deterministic helper a
  project wants for hard CI gating is written in the target project's own stack — never a runtime
  Trellis imposes."* Artifact **conformance** stays agent-applied via `corpus-reviewer`; this check
  is about id allocation, which is a merge-queue fact no rubric can see.

- **The count of guards grows by one, and `inv-minimal-first` cuts against that.** The argument for
  paying it is that the cheaper options were tried and failed: the observation was *recorded twice*
  in `decision-0078` and recurred anyway, because a record is not a check. `AGENTS.md`'s iron rule is
  the same point — prefer producing a checkable artifact over writing about one.

- **This record's own id was allocated by the manual loop the check ends** — `git ls-tree origin/main
  decisions/` plus the added files of all six open PRs (#265 → `0088`, #264 → none, #263 → `0086`,
  #262 → `0087`, #245 → none, #208 → `0067`, which is a genuine gap on `main`), then a line on TRL-40
  saying `0089` was taken. Re-verified immediately before pushing. That sequence takes minutes and is
  exactly what the guard replaces.

## Self-check (gate)

- **The three recurrences are quoted from where they were recorded**, not reconstructed:
  `decision-0078:147-156` for #1 and #2, TRL-40's own table for #3. The trigger sentence is quoted
  verbatim because the whole warrant for this record is that a prior record armed it.

- **The scope of the change to `decision-0078` is stated narrowly and is checkable.** The forward
  pointer names the dropped observation at `:147-156` and nothing else; `## Decision` there is
  untouched, and `inv-no-orphan-followups` remains exactly as minted. A reader who wants to reject
  this record can do so without disturbing that rule.

- **Every abstract claim above carries its concrete instance** (iron rule): the tie-break shows both
  the error and the notice as emitted; the not-a-claim rule names `README.md`, `0089.md` and a
  rename; the two-ids case names `0088`/`0090`; the id allocation lists the six PRs it checked.

- **Mutation-checked, not merely tested.** Flipping the tie-break comparison (`-lt` → `-gt`) turns
  both tie-break tests red in opposite directions — the lower-numbered rival stops failing, the
  higher-numbered one starts. Restored and re-run green. A check nobody has broken on purpose is a
  check nobody has verified.

- **Every frontmatter edge is argued in the body, or it is not here.** `decision-0082` is why this
  record carries no `status` and treats the forward pointer as the whole supersession mark;
  `decision-0078` is both dependency and subject; `0005` / `0010` are the layer and no-runtime
  arguments above; `0028` the pair guard; `0077` the first collision; `0085` the note-beside-the-claim
  shape. A sixth entry (`decision-0075`) was drafted and **dropped rather than left decorative**, the
  disposal `decision-0069` used for the same defect.

- **The independent pass ran, and it caught something.** The `corpus-reviewer` returned PASS on all
  eleven rubric checks and one real finding the numbered checks do not reach: the citation of
  `decision-0078`'s parked observation was written as `:146-155` in four places, correct against
  `main` and wrong by one against the state a reader sees — because the forward pointer this change
  adds to that record's frontmatter displaces its body. **Line 156 is the trigger sentence this
  record quotes as its entire warrant**, so the scope pointer excluded the line that fired. Corrected
  to `:147-156` at all four sites. `decision-0085:16-17` had hit the identical off-by-one and
  disclosed it; citing that record's §5 without carrying its handling of this exact problem is the
  defect. The author wrote the script, the workflow and the tests, so the PR review is still the only
  pass on the code (`inv-independent-judgment`).

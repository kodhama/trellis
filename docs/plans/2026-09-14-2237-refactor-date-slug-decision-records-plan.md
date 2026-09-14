---
title: Decision records are named by date and slug - Plan
type: refactor
date: 2026-09-14
topic: date-slug-decision-records
artifact_contract: ce-unified-plan/v1
product_contract_source: ce-brainstorm
execution: code
---

# Decision records are named by date and slug - Plan

## Goal Capsule

- **Objective:** An agent writing a new decision record names it without reserving an id or waiting on a CI check, and no record still describes the id-reservation guard as a live check without a note saying it no longer holds.
- **Means:** new records are named by date and slug, and that name is their id; the decision-id guard is deleted; the records that define or describe it get dated notes.
- **Product authority:** the maintainer, through `trellis-simplify`, whose answers of 2026-09-14 settle the id form and how the guard's records are marked. The Product Contract's requirements win on behaviour; the Key Technical Decisions win on mechanism. PR 1 of idea 1a (TRL-98, #313) and idea 1b are not active scope.
- **Open blockers:** none. #313 merged into `main` as `85ad446` on 2026-09-14.
- **Execution profile:** one pull request against `main`, four implementation units run inline, no plugin payload change.
- **Stop conditions:** a change under `plugins/trellis/`, a new decision record, or an edit to `AGENTS.md`'s release line (idea 5 owns it).

---

## Product Contract

### Summary

New decision records get a date-and-slug filename, and `decision-` plus that filename's stem is their id, so no id needs reserving. The decision-id guard and `AGENTS.md`'s reservation clause are deleted. The records that define the guard get dated notes saying their decisions no longer hold, and the records that describe it as a live check get notes on those sentences.

### Problem Frame

Records are numbered. Two branches cut before either merges can both take the next free number under different filenames. That happened three times: `decision-0077`; `decision-0076` on #252; and `decision-0086` on both #262 and #263. A person noticed each one, and the third renumber touched seven files (`decision-0089`'s Context). The response was a CI check that reads every open pull request through the GitHub API (`decision-0089`, refined by `decision-0092` and moved to `docs/decisions/` by `decision-0099`). It costs 337 lines of shell, a 49-line workflow, a 697-line Go test, and a clause in `AGENTS.md` that every record author has to follow.

Two things have changed since. TRL-98 made records rare, so the guard now protects a step that almost never happens. And a record named by date and slug breaks the guard, which reads a filename's leading four digits and dash as the id, so every such record claims `decision-2026`. A probe on 2026-09-14 ran the script on injected inputs. One such record passed; two in one pull request, two open pull requests each adding one, or one already on `main` each went red with an instruction to renumber. The guard is advisory and cannot block a merge. Once the first date-and-slug record lands, which TRL-97 plans, the guard sends a false red to the author of every later one.

### Requirements

**Naming new records**

- R1. A new decision record is `docs/decisions/YYYY-MM-DD-<slug>.md`, dated the day it is written and with a slug of lowercase words and digits joined by single hyphens, as numbered records' slugs are; its `id` is always `decision-` followed by that filename's stem, so `docs/decisions/2026-09-20-rules-toml-is-optional.md` has `id: decision-2026-09-20-rules-toml-is-optional`.
- R2. Records `decision-0001` to `decision-0100` keep their filenames, their ids and every citation of them.
- R3. `AGENTS.md` states R1's naming rule where it tells an agent how to record a choice, in place of the id-reservation clause.
- R4. `AGENTS.md` gives, in one sentence, the reason no check replaces the guard: two pull requests adding the same path conflict at merge, so git stops the second, and the guard existed only because numbered ids let differently named files claim one number.
- R5. A debt marker can cite a date-and-slug record, as `TODO(decision-2026-09-20-rules-toml-is-optional)` does, and pass the TODO check as `TODO(decision-0042)` does.

**Retiring the guard**

- R6. The decision-id guard's script, workflow and Go test are deleted.
- R7. Anything that names the guard as its reason is updated in the same PR; a check that still earns its place keeps running, with a reason that no longer cites the guard.

**Marking what retires**

- R8. `AGENTS.md` states one rule for a record retired in full, both where its operating method describes retirement and in its table row on superseding a record: with a successor record it carries `superseded_by`, and with no successor record it gets a dated note instead.
- R9. `decision-0089` and `decision-0092` each get a dated note, in `AGENTS.md`'s shape, saying the whole decision no longer holds and naming `AGENTS.md` as the place record naming is stated now.
- R10. Every other record with a point or sentence that deleting the guard makes untrue gets a dated note on that point, among them `decision-0098`'s sentence that `decision-id-guard` remains a separate check and `decision-0099`'s point 2, the guard's claim rule.

### Key Decisions

- **The id is the date plus the slug, and always matches the filename.** Governs R1, R4. (session-settled: user-directed — chosen over a full stamp with the time of day and over a slug-only id: citations stay of medium length, and git already stops two pull requests adding the same path.)
- **The guard's records get dated notes, and the rule allowing that ships in this PR.** Governs R8, R9, R10. (session-settled: user-directed — chosen over adding the sentence to #313 and over a successor record: #313 stays as approved, and the rule ships with its first use.)
- **No check replaces the guard, and none rejects a new numbered name.** Governs R4, R6. (session-settled: user-approved — chosen over a conformance assertion that numbered records stop at `0100`: `AGENTS.md` carries the rule, as it carries the citation rule the maintainer kept out of CI.)
- **This PR writes no decision record.** (session-settled: user-approved — chosen over a successor record for the guard: retiring an advisory check is cheap to undo and quick to notice, so it fails `AGENTS.md`'s record test.)
- **The TODO check accepts the new id form.** Governs R5. (session-settled: user-approved — chosen over keeping markers limited to numbered records: TRL-97's record will be the first date-and-slug record a marker can cite.)

<!-- ce-section: work-relationships -->
### How This Work Fits Together

This plan covers PR 2 of idea 1a in the Simplify Trellis project. The list below is the current understanding, not a committed roadmap.

- PR 1 of idea 1a (TRL-98, #313, merged as `85ad446`): the record test, dated notes, and no new `superseded_in_part_by`.
  - Enables this plan: R8 extends the retirement rule #313 put on `main`.
- TRL-97 (simplify-2): writes a decision record that 14 dated notes will name.
  - Shares R1's naming, which it adopts before this PR merges.
  - Can proceed independently of this plan: nothing on `main` reacts to its record except the advisory guard.
- Idea 1b: consolidates records into living specs and re-points cited decision ids.
  - Still to decide: whether numbered records are renamed then. This plan keeps them.

### Acceptance Examples

- AE1. **Covers R1, R2.** **Given** TRL-97's record `docs/decisions/2026-09-15-<slug>.md` with `id: decision-2026-09-15-<slug>` merged before this PR, **when** this PR's checks run, **then** none fails on that record, an artifact listing that id in `depends_on` resolves, and every numbered record keeps its filename and id.
- AE2. **Covers R4.** **Given** two open pull requests each adding `docs/decisions/2026-09-20-same-slug.md` with different contents, **when** the first merges, **then** the second shows a merge conflict on that path, and no CI check is involved.
- AE3. **Covers R1, R6.** **Given** this PR merged, **when** a pull request adds two date-and-slug records from the same year, **then** no check reports an id collision.
- AE4. **Covers R5.** **Given** a source file containing `TODO(decision-2026-09-20-rules-toml-is-optional)`, **when** the TODO check runs, **then** it passes, and a bare `TODO` in the same file still fails.
- AE5. **Covers R8, R9, R10.** **Given** an agent reading `decision-0089` or `decision-0099` after this PR, **when** it reaches the note under the title, **then** it learns which of the record's rules about the guard no longer hold and where record naming is stated now, without opening another record.

### Success Criteria

- The repository loses the guard's three files, 1,083 lines, and nothing replaces them.
- An agent that has read only `AGENTS.md` names a new record correctly and knows why no guard exists.

### Scope Boundaries

- Renaming or renumbering existing records, or re-pointing citations of numbered ids (idea 1b).
- Any check that new records use date-and-slug names.
- `docs/plans/` naming, which keeps its time of day and type.
- Changes under `plugins/trellis/`: nothing there reads record filenames or ids.
- `AGENTS.md`'s release line (idea 5).

### Dependencies / Assumptions

- #313 merged into `main` as `85ad446`, so `AGENTS.md`'s record test, dated-note shape, citation rule and retirement rules are current truth this plan builds on.
- #312 (TRL-96) removed `cli-ci`'s path filter and `cli/ci_paths_guard_test.go`, so no workflow lists the guard's files.
- `decision-id-guard` is not a required status check on `main`. The required checks are `Analyze (go)`, `Analyze (javascript)`, `build-test` and `release-guard` (branch rules, read 2026-09-14), so deleting the guard needs no settings change.
- The conformance tests compare ids as strings with no numeric shape, and the full `cli` suite passed with a date-and-slug record present (probe, 2026-09-14).
- `decision-0092`'s line citations into `decision-0089` are dated by `decision-0092` itself, whose self-check says they are "given for the file as this change leaves it", so R9's note on `decision-0089` forces no citation corrections there.

### Sources

- Linear TRL-99, and idea 1 of `docs/ideation/2026-09-14-repo-simplification-ideation.html`.
- `decision-0089` (the three collisions and the guard), `decision-0092` (what counts as a claim), `decision-0098` and `decision-0099` (where the guard is described as live).
- `.github/scripts/decision-id-guard.sh`, whose `decision_id` function reads the first four digits of any `docs/decisions/[0-9][0-9][0-9][0-9]-*.md` path as the id.
- `AGENTS.md`'s *Operating method* on `main` since #313: the record test, the dated-note shape and the citation rule.

---

## Planning Contract

### Product Contract preservation

Product Contract unchanged, except that its three Deferred to Planning questions are answered by KTD3, KTD4 and KTD5 and removed from Outstanding Questions.

### Key Technical Decisions

- KTD1. **`AGENTS.md` states the naming rule in its own bold paragraph in *Operating method*, directly after the record-test paragraph, and the "record a significant choice" table row points to it.** Governs R1, R3, R4. The paragraph carries R1's name and id form with one example, says records `0001` to `0100` keep their numbers, and gives R4's reason in one sentence. It reads R1's "the day it is written" as the day the file is first created, fixed through later revisions and matching the record's frontmatter `date:`, so a revised record keeps its id (code review, 2026-09-15). The table row loses the reservation clause and its `decision-0089` and `decision-0092` citations. The table stays a router to rules stated once above it, as #313 left it.
- KTD2. **R8 rewrites the three places `AGENTS.md` states full retirement, not a sentence added beside them.** Governs R8. The append-only paragraph's "or `superseded_by` when the record is retired in full", the status paragraph's bold "A record retired in full carries `superseded_by`", and the "supersede a record, in full or in part" table row each gain the condition: `superseded_by` when a successor record exists, a dated note when none does. A sentence beside them would leave three statements contradicting it (document review, 2026-09-14). The status paragraph names `decision-0089` and `decision-0092` as the rule's example. Rubric check 7 states the same rule, so it gains the no-successor case, with its new digest and guidance row `7e` in `cli/corpus_conformance_test.go` (code review, 2026-09-15).
- KTD3. **The TODO check's filter gains one alternative for the date-and-slug id, built from R1's grammar.** Governs R5. The alternative is a four-digit year, two-digit month and day, then one or more groups of lowercase letters and digits joined by single hyphens. An empty `TODO(decision-)`, a trailing hyphen, capitals or a one-digit month match neither alternative and still fail. The script's error message and `README.md`'s sentence listing accepted markers show the new form beside `TODO(decision-0042)`, because both describe what the filter accepts (`decision-0028`).
- KTD4. **Six records get dated notes; every other mention of the guard is history and stays as written.** Governs R8, R9, R10.

  | Record | What its note names | Why |
  |---|---|---|
  | `decision-0089` | the whole decision | R9 |
  | `decision-0092` | the whole decision | R9 |
  | `decision-0098` | the supersession list's clause that `decision-id-guard` remains a separate check, and the Consequences sentence that both conclusions still hold because `decision-id-guard` is not a required check | present-tense claims about a check that no longer exists |
  | `decision-0099` | point 2 whole; point 4's clause that once the move merges the guard sees `docs/decisions/`; point 6's clause that `decision-0089` is not superseded because its surviving clauses are true once swept | rules and forward claims about the guard |
  | `decision-0087` | the provenance aside's clause that `decision-0089`'s CI guard now fails the higher-numbered claimant | an undated present-tense claim |
  | `decision-0082` | point 3's "Supersession is marked by the forward pointer" and #313's note that full retirement still uses `superseded_by` | R8 makes both untrue for a record with no successor (code review, 2026-09-15) |

  These stay as written:
  - `decision-0078`'s description of the guard's three files, which sits inside its dated note of 2026-09-03. `AGENTS.md` counts a dated note as a dated inventory, a passage that reports what was true on a stated date, so what it reports is not a current-truth claim a later note must mark.
  - `decision-0087`'s statement that numbers are allocated by whichever branch merges first, which stays true of the numbered records.
  - `decision-0099`'s Context, which describes the situation the record answered on its date.
  - `decision-0098`'s account of the comment it corrected in the guard's workflow, and its self-check, which report what that change did.

  Each note names `AGENTS.md`'s *Operating method* as the place current behaviour is stated. On `decision-0098`, `decision-0099` and `decision-0082` the new note goes below the note already there.
- KTD5. **One live line citation is corrected: `decision-0100`'s Consequences citation of `decision-0098:133-134`.** The lines it means, which name `core/fixtures/known-bad/` and `core/rubrics/`, sit at `decision-0098:135-136` on `main` because #312's note moved them. KTD4's note on `decision-0098` moves them two lines further, so the citation is corrected in place to their final position, checked against the edited file. These stay as written, because each is a dated inventory under `AGENTS.md`'s citation rule:
  - #312's note on `decision-0100`, which cites `decision-0098:135-136`.
  - `decision-0092`'s citations of `decision-0089`, `decision-0087:28` and `AGENTS.md:51`, which it dates itself and lists in its sweep table.
  - `decision-0081`'s and `decision-0082`'s citations of `AGENTS.md` lines, dated as #313 found.

  No file cites `decision-0099` by line, and `decision-0082` is cited by line only from a plan.
- KTD6. **`cli/selfapply_test.go` keeps its assertion that no root `decisions/` directory exists, and its comment and failure message name the conformance corpus instead of the guard.** Governs R7. The assertion still catches a record landing outside `docs/decisions/`, the only record path `corpusRoots` in `cli/corpus_conformance_test.go` reads.
- KTD7. **`package.json`'s `lint:shell` drops its `.github/scripts/*.sh` pattern.** Governs R6. The guard's script is the only file in `.github/scripts/`, and an unmatched pattern reaches shellcheck as a literal path, which fails the quality gate CI runs. Nothing pins that command's file list.
- KTD8. **One new test, `cli/check_todos_test.go`.** `TestCheckTodosMarkerForms` runs `scripts/check-todos.sh` over every marker form the script's error message and `README.md` show, which must pass, and over U3's malformed samples, which must fail. This plan first chose no test, because the project is retiring guards (idea 4). Codex's review of PR #315 pointed out that `decision-0028`, still current on `main`, asks for a guard on each source↔derivative pair, and idea 4 has not landed, so the test was added. The guard's deletion is proved by the Go suite still building and passing, and by a search showing that nothing outside records, plans and the ideation names the guard.
- KTD9. **The notes cite PR #315**, the next number when this plan was written, since #314 is the newest issue or pull request. The PR corrects them before review if another pull request takes that number first, as `AGENTS.md` requires.

### Assumptions

- TRL-97's record may merge to `main` before this PR. If it does, `main` is merged into the branch, the checks run again, and AE1 is checked against the real record.
- No open pull request adds another record sentence that describes the guard as live. A search just before the PR opens confirms it.
- `shellcheck` is not installed locally, so `npm run quality` stops at `lint:shell` here. CI installs it and runs the gate, including over the edited `scripts/check-todos.sh`.

### Sequencing

U1, U2 and U3 are independent. U4 follows U2, because its notes name the paragraph U2 adds. All four land in one pull request.

---

## Implementation Units

### U1. Retire the decision-id guard

- **Goal:** delete the guard and update what named it.
- **Requirements:** R6, R7; AE3; KTD6, KTD7.
- **Dependencies:** none.
- **Files:**
  - delete `.github/scripts/decision-id-guard.sh`
  - delete `.github/workflows/decision-id-guard.yml`
  - delete `cli/decision_id_guard_test.go`
  - modify `cli/selfapply_test.go`
  - modify `package.json`
- **Approach:**
  1. Delete the three files; `.github/scripts/` leaves the tree with them.
  2. Drop the `.github/scripts/*.sh` pattern from `lint:shell` (KTD7).
  3. Reword the root-`decisions/` assertion's comment and failure message (KTD6).
- **Patterns to follow:** #312 deleted `cli/ci_paths_guard_test.go` together with everything that named it, in one pull request.
- **Test scenarios:**
  - The Go suite builds and passes with the three files gone, because no other test file calls the deleted test's helpers.
  - The self-apply test still fails when a root `decisions/` directory exists, and its message names the conformance corpus rather than the guard.
  - A case-insensitive search for `decision-id-guard`, `decision-id guard`, `decision_id_guard` and `DecisionIDGuard` outside `docs/decisions/`, `docs/plans/` and `docs/ideation/` finds nothing. The spaced form is how `cli/selfapply_test.go` named the guard, so the search also shows KTD6's rewording landed.
  - `lint:shell`'s command names no pattern that matches no file.
- **Verification:** the U1 checks in the Verification Contract pass.

### U2. State the naming and retirement rules in AGENTS.md

- **Goal:** `AGENTS.md` tells an agent how to name a record, why no guard exists, and how a full retirement is marked, and rubric check 7 states the same retirement rule.
- **Requirements:** R1, R3, R4, R8; KTD1, KTD2.
- **Dependencies:** none.
- **Files:** modify `AGENTS.md`, `docs/rubrics/artifact-contract.md` and `cli/corpus_conformance_test.go`.
- **Approach:**
  1. Add the naming paragraph after the record-test paragraph (KTD1).
  2. Point the "record a significant choice" table row at it, in place of the reservation clause.
  3. Rewrite the three full-retirement statements, with their example (KTD2).
  4. Add the no-successor case to rubric check 7, then record its new digest and guidance row (KTD2).
  5. Leave *Checks and review*'s release line as it is.
- **Patterns to follow:** #313's operating-method block: a bold lead sentence per rule, and at least one concrete example per rule (the iron rule). #313's edit to rubric check 7 and its digest.
- **Test scenarios:**
  - `cli/docs_consistency_test.go` and `cli/selfapply_test.go`, which read `AGENTS.md`, pass. The new text contains no `trellis` followed by a lowercase word and no `/trellis:`, which the docs-consistency test reads as a command claim.
  - Covers the second Success Criterion: reading only `AGENTS.md`, an agent derives `docs/decisions/2026-09-20-rules-toml-is-optional.md` with `id: decision-2026-09-20-rules-toml-is-optional`, and nothing tells it to reserve an id or check other pull requests.
  - None of the three full-retirement statements says `superseded_by` without the successor-record condition.
  - The artifact-contract guard fails on rubric check 7's old digest and passes once the new digest is recorded.
- **Verification:** the Go suite passes, and `AGENTS.md` no longer mentions reserving an id, a free id, a claimant or `decision-0089`'s guard.

### U3. Accept the new id form in the TODO check

- **Goal:** a debt marker can cite a date-and-slug record.
- **Requirements:** R5; AE4; KTD3, KTD8.
- **Dependencies:** none.
- **Files:** modify `scripts/check-todos.sh` and `README.md`; add `cli/check_todos_test.go` (KTD8).
- **Approach:**
  1. Add the date-and-slug alternative to the filter (KTD3).
  2. Show the new form in the script's error message and in `README.md`'s list of accepted markers.
- **Execution note:** prove the filter on sample marker lines before and after the edit. The script scans tracked files only, so an untracked scratch file does not exercise it.
- **Test scenarios:**
  - Covers AE4. `TODO(decision-2026-09-20-rules-toml-is-optional)` passes.
  - `TODO(decision-0042)`, `TODO(TRL-123)`, `TODO(#45)` and `FIXME(decision-2026-09-20-rules-toml-is-optional)` still pass.
  - A bare `TODO`, `TODO(decision-)`, `TODO(decision-2026-09-20-)`, `TODO(decision-2026-09-20-Rules)` and `TODO(decision-2026-9-20-rules)` fail.
  - `npm run check:todos` passes on the branch.
- **Verification:** each sample behaves as listed, and `check:todos` passes.

### U4. Mark the guard's records with dated notes

- **Goal:** every record that defines or describes the guard as live, or states the retirement rule R8 changes, says what no longer holds.
- **Requirements:** R8, R9, R10; AE5; KTD4, KTD5, KTD9.
- **Dependencies:** U2.
- **Files:** modify
  - `docs/decisions/0082-retire-the-status-field.md`
  - `docs/decisions/0087-one-gateway-for-every-payload-read.md`
  - `docs/decisions/0089-a-decision-id-is-claimed-at-the-pr-not-in-the-directory.md`
  - `docs/decisions/0092-a-claim-is-a-new-record-path.md`
  - `docs/decisions/0098-artifact-conformance-is-a-ci-check.md`
  - `docs/decisions/0099-the-governance-corpus-lives-under-docs.md`
  - `docs/decisions/0100-the-artifact-contract-is-repository-internal.md`
- **Approach:**
  1. Add one note in `AGENTS.md`'s shape directly under each of the six records' titles, below any note already there, naming what KTD4's table lists.
  2. Correct `decision-0100`'s Consequences citation against the edited `decision-0098` (KTD5).
  3. Change nothing else in the records, frontmatter pointers included.
- **Patterns to follow:** #313's notes on `decision-0014`, `decision-0040` and `decision-0082`, and #312's multi-clause note on `decision-0098`.
- **Test scenarios:**
  - `TestCorpusConform` and `TestArtifactContract` pass: each record still opens with its frontmatter and keeps its required sections.
  - Covers AE5. Reading `decision-0099` from the top, an agent meets the note before point 2 and learns the claim rule no longer holds and that `AGENTS.md` states record naming now.
  - `decision-0100`'s corrected citation lands on the two lines of `decision-0098` that name `core/fixtures/known-bad/` and `core/rubrics/`.
  - Each record's diff only adds lines, except `decision-0100`'s one-line citation correction.
- **Verification:** the conformance tests pass, and the diff under `docs/decisions/` is six added notes and one corrected citation.

---

## Verification Contract

Every command runs from the repository root.

- **Build and vet** (U1): `go -C cli build ./...` and `go -C cli vet ./...` succeed, which shows nothing depended on the deleted test's helpers.
- **Go suite** (U1, U2, U4): `go -C cli test -count=1 ./...` passes, including the self-apply and docs-consistency tests.
- **Conformance** (U2, U4): the conformance and artifact-contract tests pass when run on their own, with `go -C cli test -count=1 -run 'TestCorpusConform|TestArtifactContract' .`.
- **Quality gate** (U1, U3): `npm run quality` passes through `lint:js`; `lint:shell` needs `shellcheck`, which CI installs.
- **Guard search** (U1): `git grep -n -i -E 'decision-id[- ]guard|decision_id_guard|DecisionIDGuard' -- . ':!docs/decisions' ':!docs/plans' ':!docs/ideation'` returns nothing.
- **TODO samples** (U3): `TestCheckTodosMarkerForms` in the Go suite runs `scripts/check-todos.sh` over U3's sample markers and the forms its error message and `README.md` show, and passes and fails them as U3 lists.
- **Payload** (all): `git diff origin/main...HEAD -- plugins/trellis` is empty once the work is committed. The three-dot form diffs from the merge base, so changes merged to `main` since the branch was cut do not show.

---

## Definition of Done

- R1 to R10 hold, and AE3 to AE5 hold on the branch. AE1 holds against TRL-97's record if that record merged first, and otherwise rests on the 2026-09-14 probe named in Dependencies. AE2 rests on git's conflict on a path two pull requests both add, and needs no check.
- Every Verification Contract check passes, with `lint:shell` left to CI where `shellcheck` is missing.
- The PR body lists the six records it put dated notes on and the one citation it corrected, as `AGENTS.md` requires.
- No abandoned or experimental edits remain in the diff.
- PR 2 is open against `main`, and every review finding is fixed or answered on the PR.

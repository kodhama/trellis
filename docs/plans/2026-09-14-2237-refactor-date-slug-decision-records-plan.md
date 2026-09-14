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
- **Product authority:** the maintainer, through `trellis-simplify`, whose answers of 2026-09-14 settle the id form and how the guard's records are marked. PR 1 of idea 1a (TRL-98, #313) and idea 1b are not active scope.
- **Open blockers:** none. #313 merged into `main` as `85ad446` on 2026-09-14.
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

### Outstanding Questions

**Deferred to Planning**

- The TODO check's exact accepted pattern for the new id form, including whether an empty `TODO(decision-)` still fails.
- The full set of records R10 reaches. Candidates beyond `decision-0098`'s separate-check sentence and `decision-0099`'s point 2: `decision-0099`'s points 4 and 6 (point 6 says `decision-0089` is not superseded), `decision-0098`'s sentence that its conclusions hold because `decision-id-guard` is not a required check, and `decision-0087`'s provenance line that the guard fails the higher-numbered claimant.
- Whether `decision-0100`'s citation of `decision-0098:133-134` is live under `AGENTS.md`'s citation rule, and so corrected when R10's note shifts those lines.

### Sources

- Linear TRL-99, and idea 1 of `docs/ideation/2026-09-14-repo-simplification-ideation.html`.
- `decision-0089` (the three collisions and the guard), `decision-0092` (what counts as a claim), `decision-0098` and `decision-0099` (where the guard is described as live).
- `.github/scripts/decision-id-guard.sh`, whose `decision_id` function reads the first four digits of any `docs/decisions/[0-9][0-9][0-9][0-9]-*.md` path as the id.
- `AGENTS.md`'s *Operating method* on `main` since #313: the record test, the dated-note shape and the citation rule.

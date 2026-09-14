---
title: Decision records become rare - Plan
type: docs
date: 2026-09-14
topic: decision-records-become-rare
artifact_contract: ce-unified-plan/v1
product_contract_source: ce-brainstorm
execution: code
---

# Decision records become rare - Plan

## Goal Capsule

- **Objective:** An agent working in this repository can tell from `AGENTS.md` alone whether its change needs a new decision record and whether it owes dated notes on existing records. From then on new records are rare, and no change adds a partial-supersession chain.
- **Means:** PR 1 of Simplify Trellis idea 1a (TRL-98) rewrites the operating method in `AGENTS.md` and the surfaces that restate it (KTD1).
- **Product authority:** the maintainer, through `trellis-simplify`, who approved these requirements on 2026-09-14. The Product Contract's requirements win on behaviour; the Key Technical Decisions win on mechanism. PR 2 of idea 1a (TRL-99) and idea 1b are not active scope.
- **Execution profile:** one documentation PR with no code behaviour change, built by the `simplify-1a` session and merged by the maintainer.
- **Stop conditions:** stop and ask `trellis-simplify` before any change would touch `plugins/trellis/`, add a decision record, or edit `AGENTS.md`'s release line about when installs re-pull.
- **Open blockers:** none.

---

## Product Contract

### Summary

`AGENTS.md` admits a new decision record only when a wrong call would be expensive to undo, slow to notice, or seriously damaging while it stands. Any other change keeps its reasoning in its requirements doc, plan and PR, and puts a dated note under the title of each record it makes untrue. No new `superseded_in_part_by` pointer is written. PR 1 adds no record of its own and puts dated notes on the records whose account of the method it makes untrue.

### Problem Frame

`AGENTS.md:29-31` makes the records the method's home: *"a decision is the current truth"*, `docs/decisions/` is append-only, and you *"supersede with a forward pointer, never edit."* Changing any rule a record states therefore means writing a successor record and a pointer. The maintainer's ruling of 2026-09-05 made that explicit: *"Factual corrections get dated notes. Rule divergences get successors"* (`decision-0092:16-17`).

The corpus now holds 98 records and about 150,000 words. By the ideation's count, 53 records carry a supersession pointer, 47 of them partial. `decision-0089`'s pointer comment alone is a clause-by-clause diff of about 500 words (`decision-0089:6`). Fifty-seven records were added in July and fifteen in the first half of September. Of the 15 records added between 2026-08-31 and 2026-09-13, eight record a plugin behaviour change and landed in the same commit as a `plugins/trellis/VERSION` bump. The maintainer's reading: *"Should have more specs and less decisions."*

`decision-0081` does not remove the successor record. It scales **who** may retire a record by cost of reversal, but it still retires by writing a successor record (`decision-0081:154-164`), and its heading reads *"(proposal, not a decision)"*.

Ideas 2, 3, 5 and 6 are being worked on in parallel now, so the method has to change before they land. Under today's method each would add successor records and partial pointers, which idea 1b would then have to fold back in when it consolidates the records into living specs.

### Requirements

**When a change gets a record**

- R1. `AGENTS.md` admits a new decision record only when a wrong call would be expensive to undo, slow to notice, or the damage while it stands would be serious; any one of the three is enough.
- R2. `AGENTS.md` states that test in full with at least one example on each side, and the examples include both sides of a plugin behaviour change that ships to consumers: one where a wrong call could silently stop rules reaching consumers' sessions gets a record, and an error-message, reporting, diagnostic or cosmetic change does not.
- R3. When an agent cannot tell whether a change meets the test, `AGENTS.md` tells it to ask the maintainer instead of deciding either way.

**Where the reasoning goes without a record**

- R4. `AGENTS.md` says a change that does not meet the test keeps its reasoning in its requirements doc, its plan under `docs/plans/`, and its PR body, and that the PR body lists the records the change marks with dated notes or says it found none.

**Marking records a change makes untrue**

- R5. When a merged change makes a point or sentence of an existing record untrue, the same PR adds one dated note to that record, directly under its title heading, in this shape:

  ```markdown
  > **Dated note, YYYY-MM-DD — TRL-<n> (PR #<n>):** <what is no longer true, naming the point or sentence>. <where the current behaviour is stated now>.
  ```

- R6. `AGENTS.md` carries the dated-note shape from R5 word for word, and the citation rule from R10.
- R7. No PR adds a `superseded_in_part_by` entry; existing entries stay as they are and keep resolving.
- R8. A record written under R1 marks each record it retires in part with a dated note that names the new record, and each record it retires in full with `superseded_by`, as today.
- R9. A dated note adds text above the record's body and never rewrites the record's original text.
- R10. The PR that adds a dated note corrects every live line citation into that record, whether in another record, a code comment, a test or `AGENTS.md`; citations in plans and in dated inventories stay as they are.

**Where current truth lives until idea 1b**

- R11. `AGENTS.md` says a record is current truth except where a dated note says a point no longer holds; for that point, the place the note names is current truth.
- R12. The place a dated note names is a file on `main` that is kept current in place, never a plan or a PR body; when no such file states the new behaviour, the same PR states it in `AGENTS.md`.

**The rule applied to its own PR**

- R13. PR 1 adds no decision record.
- R14. PR 1 adds a dated note to each existing record whose statement of the method it makes untrue.
- R15. PR 1 updates every surface outside `docs/decisions/` that restates the method, starting with check 7 of `docs/rubrics/artifact-contract.md` and `README.md`'s two statements that decisions are append-only and superseded by a forward pointer.

### Key Decisions

- **Two PRs, and this plan is the first.** Naming new records by date and slug and deleting the decision-id guard go to PR 2 (TRL-99), after simplify-6 (TRL-96) merges. (session-settled: user-directed — chosen over keeping naming and the guard for 1b, and over one combined PR: the rule is urgent while other ideas are running, and PR 2 would collide with simplify-6's `cli-ci.yml` edits.)
- **Dated notes mark out-of-date records.** Governs R5, R6, R9, R11, R12. (session-settled: user-directed — chosen over an `AGENTS.md` precedence rule that leaves records untouched, and over a single interim list of overridden records: a reader sees the mark in the record itself, and idea 1b gets a hint about which points are spent rather than pointers to unwind.)
- **A new record marks the records it retires in part with dated notes.** Governs R8. (session-settled: user-approved — chosen over a new `superseded_in_part_by` pointer: confirmed through simplify-2's `decision-0102` question, and `decision-0102` is written that way.)
- **Full retirement keeps `superseded_by`.** Governs R8. (session-settled: user-directed — chosen over dated notes for full retirement too: the conformance check keeps resolving the pointer, and idea 1b's hand-over record uses it. Only the partial pointer retires.)
- **The PR that adds a note corrects the live citations it shifts.** Governs R10. (session-settled: user-directed — chosen over leaving shifted citations for idea 1b to re-point: a live citation that lands on the wrong line misleads readers now, while plans and dated inventories are history and stay.)
- **Idea 1a writes no record of its own.** Governs R13. (session-settled: user-directed — chosen over recording `decision-0101`: by the test it is cheap to reverse, since undoing it means editing `AGENTS.md` back, a wrong call shows on every PR, and nothing leaves the repository; landing without a record is the rule's first worked example.)
- **The test is `decision-0081`'s three parts, stated in full in `AGENTS.md`.** Governs R1, R2, R3. (session-settled: user-directed — chosen over "consumer-visible changes only" and "strategic forks only": it covers internal choices that are costly, such as idea 1b's hand-over record, and asking when unsure prevents silent calls. It is stated in full because `decision-0081` addressed who may retire a record, not when to write one.) R1 keeps `decision-0081`'s own size threshold, *"the damage while it stands would be serious"* (`decision-0081:328-329`), which the maintainer approved after seeing it differed from his "harmful".
- **A shipped plugin behaviour change gets a record only when a wrong call could silently stop rules reaching consumers' sessions.** Governs R2. (session-settled: user-directed — chosen over "always", since every consumer gets it, and over "never", since the next release undoes it: this keeps records for changes that can quietly switch governance off in another project and drops them for error-message, reporting, diagnostic and cosmetic changes.)
- **A dated note points at a file someone keeps current.** Governs R12. Plans are history that nobody maintains (`decision-0097` point 5), and a PR body lives outside the repository, so a note pointing at either would make stale text current truth.
- **Existing pointers keep their conformance checks, and nothing checks dated notes.** Governs R7. The 53 records that carry pointers still need them to resolve. A check for notes would add a test that idea 4 plans to delete with the other corpus checks.

<!-- ce-section: work-relationships -->
### How This Work Fits Together

This plan covers PR 1 of idea 1a (TRL-98). The rest is the current understanding of the Simplify Trellis project, not a committed roadmap.

- PR 2 of idea 1a (TRL-99): new records named by date and slug, and the decision-id guard retired.
  - Depends on this plan landing, and on simplify-6 (TRL-96) merging.
- Idea 1b: consolidate the records into living specs, write one hand-over record, and re-point the decision ids cited from code and tests.
  - Depends on this plan; reads the dated notes as evidence of which points are spent.
- Ideas 2, 3, 5, 6 and 7: follow this plan's rule once it is on `main`.
  - Shares the dated-note shape and the citation rule, which simplify-2 and simplify-6 are already using.
  - Can proceed independently of this plan's merge: the maintainer approved TRL-96 without waiting for PR 1, and a PR that lands first cites his TRL-98 ruling in its body.

### Acceptance Examples

- AE1. **Covers R1, R4, R5.** **Given** a Simplify Trellis PR that deletes a CI path filter. **When** its agent applies the test, one workflow edit reverses the deletion, the next PR's CI shows a wrong call, and nothing ships to consumers. **Then** it writes no record, keeps its reasoning in its requirements, plan and PR, adds a dated note under the title of each record whose text the deletion makes untrue, and lists those records in its PR body.
- AE2. **Covers R1, R8.** **Given** a PR that replaces the consumer-facing `rules.toml` format. **When** its agent applies the test, consumers who migrated would have to migrate back, so undoing it is expensive. **Then** it writes a record, and each existing record that describes the old format, such as `decision-0083`, gets a dated note naming the new record and no partial pointer.
- AE3. **Covers R2.** **Given** two plugin releases, one changing how the session-start hook reconciles `.trellis/rules.toml` and one rewording the hook's error message. **When** each agent applies the test. **Then** the first gets a record, because a wrong reconcile could silently leave a consumer's sessions without rules, and the second gets none.
- AE4. **Covers R3.** **Given** an agent that cannot judge whether a change would do serious damage while it stands. **Then** it asks the maintainer before writing or skipping a record.
- AE5. **Covers R5, R9, R11.** **Given** `decision-0040` after PR 1 merges. **When** an agent reads it, a dated note under its title says partial retirements now use dated notes and points to `AGENTS.md`. **Then** point 5's original text is unchanged, and the agent follows `AGENTS.md`.
- AE6. **Covers R7.** **Given** `decision-0089`'s existing `superseded_in_part_by` entry. **Then** it stays and still resolves under conformance check 7b.
- AE7. **Covers R12.** **Given** a change without a record that alters a rule no maintained file restates. **Then** the same PR states the new rule in `AGENTS.md`, and the dated note names `AGENTS.md`, not the change's plan.
- AE8. **Covers R10.** **Given** a record that receives a dated note, whose line numbers are cited both in a test comment and in a plan under `docs/plans/`. **Then** the same PR corrects the test comment's citation to the shifted lines and leaves the plan's citation as written.

### Success Criteria

- A fresh agent session can tell from `AGENTS.md` alone whether its change needs a new record and whether it owes dated notes, where some changes need both and some neither, without opening `decision-0040`, `decision-0081` or `decision-0082`. Finding the records a note is due on, and the citations into them, still means searching the repository.
- No PR merged after PR 1 adds a `superseded_in_part_by` entry.
- PR 1 stays reviewable on a phone: `AGENTS.md`, `README.md`, the rubric text R15 reaches, a few one-paragraph notes, and any citation corrections they force.

### Scope Boundaries

- **PR 2 (TRL-99):** naming new records by date and slug, and deleting the decision-id guard with `AGENTS.md`'s id-reservation clause. PR 1 leaves that clause as written.
- **Idea 1b:** consolidating records into living specs, the hand-over record, re-pointing cited decision ids, and removing existing pointers.
- **Idea 5:** `AGENTS.md`'s release line about when an existing install re-pulls the plugin. It is incomplete for installs pinned per checkout, and PR 1 neither edits it nor restates it.
- **Not changed:**
  - existing `superseded_by` and `superseded_in_part_by` entries and their conformance checks
  - any record's substance; the only edits to another record's text are the citation corrections R10 requires
  - git history
  - anything under `plugins/trellis/`, including the shipped rules text, so no `VERSION` bump
- **Not enforced by a check:** the admission test, the dated-note shape and the citation rule; the PR reviewer applies all three.

### Dependencies / Assumptions

- No Go test reads `decision-0014`, `decision-0015`, `decision-0040` or `decision-0082` by path or pins their lines, and no file outside `docs/` cites their line numbers (grep, 2026-09-14). So PR 1's notes break no test and force no edit under `plugins/trellis/`.
- Inside `docs/`, the only line citations into the records PR 1 notes (KTD3) are `decision-0081`'s citations of `decision-0040`, which that record dates to 2026-08-28 in its own header, and one citation of `decision-0082` in a plan (grep, 2026-09-14). Both are history under R10, so PR 1 corrects no citation.
- `trellis-simplify` counts at least 153 `decision-NNNN:line` citations across the repository, some in `cli/` tests, `install.sh` and `plugins/trellis/`. A later PR whose note shifts a record cited under `plugins/trellis/` therefore edits the payload, and `AGENTS.md` makes that a release that bumps `VERSION`.
- `cli/artifact_contract_guard_test.go` pins each rubric check's text by digest, so rewording check 7 means updating its outcome rows in the same PR.

### Sources

- Ideation, idea 1: `docs/ideation/2026-09-14-repo-simplification-ideation.html` (PR #311) and the Linear document "Ideation: simplifying the Trellis repository".
- `AGENTS.md:29-36` (operating method) and `:48-51` (the supersede, retire and record rows).
- `decision-0081:17` (heading), `:154-164` (retirement by successor record), `:324-329` (proposed directive).
- `decision-0082` Decision point 3; `decision-0040` Decision point 5; `decision-0092:16-17` (the 2026-09-05 ruling); `decision-0097` point 5 (plans are history); `decision-0014` Decision bullet 2; `decision-0015:117-135` (the rename-sweep rule and its open question).
- `docs/rubrics/artifact-contract.md` check 7; `cli/corpus_conformance_test.go` check 7b.

---

## Planning Contract

**Product Contract preservation:** requirements, Key Decisions and Acceptance Examples unchanged. Changed by planning evidence, with no scope change: the Goal Capsule gains execution profile and stop conditions; the second Dependencies bullet now states which citations exist (planning found none that PR 1 must correct); the deferred planning questions are resolved below and their list is removed.

### Key Technical Decisions

- KTD1. **`AGENTS.md` states the decision-record method in one block inside `## Operating method`, replacing the opening paragraph and the supersession sentence.** The rules must be readable in one place (Success Criteria), and the current paragraph at `AGENTS.md:29-31` is what reviewers read as forbidding dated notes: on #312 a Codex review took *"append-only: you supersede with a forward pointer, never edit"* as barring them. The new block says in plain words that a dated note is the only text added to a record, and that a record's original text is never rewritten except for the live line-citation corrections R10 requires. The table rows keep their shape; the supersede and record rows point at the block, and the record row keeps its id-reservation clause for PR 2. Nothing in the `## Checks and review` section changes, including the release line.
- KTD2. **A record's notes read oldest first.** A later note goes below the notes already under the title, so the note directly under the title is always the first, and parallel PRs that both add notes conflict at one predictable place. R5 places notes directly under the title; this fixes their order among themselves.
- KTD3. **PR 1 notes three records: `decision-0014`, `decision-0040` and `decision-0082`.** Each states a rule PR 1 changes:
  - `decision-0014`'s second Decision bullet and its Consequence make every change to an invariant's meaning or structure a record; R1's test now decides.
  - `decision-0040` point 5 made `superseded_in_part_by` the way to mark partial supersession; R7 and R8 retire it for new work.
  - `decision-0082` point 3 marks partial supersession by that pointer; R8 marks it by a dated note.

  Checked and left without a note:
  - `decision-0015`: its open question on appended blocks is answered, but nothing it states becomes untrue, and a note would shift the rename-sweep rule at lines 117-121 that five live records cite.
  - `decision-0018:142`: it restates `decision-0014` inside a dated aside and cites it by id, so a reader lands on `decision-0014`'s note.
  - `decision-0081`: its test moves into `AGENTS.md`, but its own claims are about who may retire a record, which PR 1 does not change.
  - `decision-0089` and `decision-0092`: their "successor record" sentences are a dated note and a quoted ruling, not rules.
- KTD4. **R15 reaches `README.md` and rubric check 7, and nothing else outside `docs/decisions/`.** Two surfaces change:
  - `README.md:301` and `README.md:326-331` describe this repository's own records.
  - Rubric check 7 describes partial supersession only by `superseded_in_part_by`.

  Checked and unchanged:
  - `profiles/trellis-self.md:105`: "`docs/decisions/` append-only" still holds, because append-only guarantees a record's decisions are not re-made, not that its bytes never change (`decision-0015:117-121`), and dated notes add text without re-making anything.
  - The catalog, `cli/assets/invariants.md`, `site/invariants.html` and the plugin reference: their "append-only decision records" signatures describe a consumer project's own ADR practice and ship to consumers.
  - `core/invariants/trellis-invariants-v1.md`, rubric line 71 and check 7's exemption: they concern retired-id resolution and dependencies of append-only records, which still hold.
  - Comments in `cli/` tests and the `cli/testdata/known-bad/` fixtures: they exercise existing checks.
- KTD5. **A line citation is live unless it sits in a plan or in a dated inventory.** A dated inventory is a passage that reports what was true on a stated date: a sweep table, a self-check count, a dated note, or a whole record that dates its own citations, as `decision-0081`'s header does. `AGENTS.md` carries this test with R10's rule, so every PR applies R10 the same way.
- KTD6. **PR 1's notes cite the PR number the repository will assign next.** The number is read from the repository just before the PR opens. If another PR takes it first, a follow-up commit on the branch corrects the three notes before review. `TRL-98` is known now.
- KTD7. **Check 7 gains one sentence and keeps its opening words.** The sentence says new partial retirements are dated notes (`AGENTS.md`) and existing `superseded_in_part_by` entries still resolve. `cli/artifact_contract_guard_test.go` finds checks by the literal `7. **Supersede integrity.**`, so that stays. The check's digest in `cli/corpus_conformance_test.go` is updated after rows 7a-7c are re-read against the new text; they stay accurate, because the Go check still resolves existing pointers and checks nothing about notes.
- KTD8. **No new Go test.** The admission test, the note shape and the citation rule are reviewer judgment (Scope Boundaries), and idea 4 plans to delete corpus guards, not add them. The existing suite covers what PR 1 can break: the artifact-contract guard (check 7's text), the corpus conformance tests (the three notes keep frontmatter first and sections intact), the docs-consistency tests (`AGENTS.md` and `README.md` claim no command that does not exist) and the self-apply test (`AGENTS.md` keeps its routing statements and headings).

### Assumptions

- The PR number read just before opening is still free when the PR opens; KTD6's follow-up commit covers the case where it is not.
- simplify-2 and simplify-6 may add notes to records PR 1 also notes. A conflict resolves by merging `origin/main` into the branch and ordering the notes per KTD2.
- The docs-consistency test treats `trellis` followed by a lowercase word, and a `/trellis:` skill name, in `AGENTS.md` and `README.md` as a command claim (`cli/docs_consistency_test.go:166-167`). New prose avoids both forms except for real commands.

### Sequencing

U1 first, because U3's notes name the `AGENTS.md` block U1 writes. U2 does not depend on either.

---

## Implementation Units

### U1. Rewrite the decision-record method in AGENTS.md

- **Goal:** `AGENTS.md` alone tells an agent whether a change needs a record, whether it owes dated notes, and how to write them.
- **Requirements:** R1-R12; Key Decisions governing R1, R2, R3, R5, R6, R8, R9, R10, R11 and R12; KTD1, KTD5.
- **Dependencies:** none.
- **Files:** `AGENTS.md`.
- **Approach:**
  1. Replace `AGENTS.md:29-31` with the block KTD1 describes, in this order: current truth (R11, R12), the admission test with its examples and the ask-when-unsure line (R1-R3), where reasoning goes and the PR-body list (R4), and dated notes with the shape word for word, the sentence KTD1 specifies (a dated note is the only text added, and original text is never rewritten except for R10's citation corrections), the placement and order (KTD2) and the citation rule with KTD5's test (R5, R6, R9, R10).
  2. Replace the supersession sentence at `AGENTS.md:35-36` with R7 and R8: full retirement keeps `superseded_by`, partial retirement is a dated note, and no new `superseded_in_part_by`. Keep the rest of that paragraph, about the `status` field, as it is.
  3. Point the "supersede a record" and "record a significant choice" table rows at the block. Keep the record row's id-reservation clause and the `decision-0081` and `decision-0074` citations in the retire row.
- **Patterns to follow:** the iron rule's "every abstract instruction carries at least one concrete example"; the existing table-row style (verb phrase, then citations).
- **Test scenarios:**
  - Covers AE3. An agent reading only `AGENTS.md` classifies a change to how the session-start hook reconciles `.trellis/rules.toml` as a record, and a rewording of the hook's error message as no record.
  - Covers AE1. The same agent classifies deleting a CI path filter as no record plus dated notes, listed in the PR body.
  - Covers AE7. The block says a note names `AGENTS.md`, code, a test or workflow, a rubric or a record, never a plan or PR body.
  - The self-apply test still finds `# Trellis — operating method`, `## Operating method` and the six routing statements in the maintaining-instructions section.
  - The docs-consistency test finds no `trellis <word>` or `/trellis:<name>` form that names a command the code lacks.
- **Verification:** the diff touches only the operating-method paragraph, the supersession sentence and two table rows; the release line and `## Checks and review` are unchanged.

### U2. Update README.md and rubric check 7

- **Goal:** the two restatements outside `docs/decisions/` match the new method.
- **Requirements:** R15; KTD4, KTD7.
- **Dependencies:** none.
- **Files:** `README.md`, `docs/rubrics/artifact-contract.md`, `cli/corpus_conformance_test.go`.
- **Approach:**
  1. `README.md:301`: the repo-map row describes the records without claiming supersession is only by pointer.
  2. `README.md:326-331`: the "How we work" sentence says decisions are rare, keep their text, and are marked by dated notes or `superseded_by`, pointing to `AGENTS.md`.
  3. Rubric check 7: add the sentence KTD7 names.
  4. Update check 7's digest in the outcome table after re-reading rows 7a-7c.
- **Execution note:** edit the rubric first and run the artifact-contract tests to see check 7's digest fail, then update the digest; the failing run is the evidence the pin still bites.
- **Patterns to follow:** earlier rubric rewords that updated a check's digest with its rows in the same change (`decision-0098`, `TRL-95`).
- **Test scenarios:**
  - After the rubric edit and before the digest update, the artifact-contract guard fails naming rubric check 7.
  - After the digest update, the artifact-contract guard and the corpus conformance tests pass.
  - The docs-consistency test passes on the changed `README.md`.
- **Verification:** check 7 still opens with `7. **Supersede integrity.**`, and rows 7a-7c are unchanged in text.

### U3. Add dated notes to the three records

- **Goal:** readers of `decision-0014`, `decision-0040` and `decision-0082` see which point no longer holds and where the current rule is.
- **Requirements:** R5, R9, R10, R12, R13, R14; KTD2, KTD3, KTD6. Covers AE5.
- **Dependencies:** U1.
- **Files:** `docs/decisions/0014-invariants-spec-vs-decisions.md`, `docs/decisions/0040-reverse-ports-from-instance-1.md`, `docs/decisions/0082-retire-the-status-field.md`.
- **Approach:**
  1. Under each title heading, add one note in R5's shape: `TRL-98`, the PR number per KTD6, the point that no longer holds (KTD3), and `AGENTS.md`'s operating-method block as where the rule is stated now.
  2. Run the line-citation search once more for the three records and apply R10 to anything live. The planning search found only history (Dependencies / Assumptions).
- **Patterns to follow:** the note shape simplify-2 and simplify-6 use in their PRs, which R5 fixes.
- **Test scenarios:**
  - Covers AE5. `decision-0040`'s diff is additions only: the note and its surrounding blank lines, with point 5's text unchanged.
  - The corpus conformance tests pass on all three records: frontmatter still opens each file, and the Context, Decision and Consequences sections are still found.
- **Verification:** `git diff` for each record shows added lines only, and each note is a single paragraph that names one point and `AGENTS.md`.

---

## Verification Contract

| Check | Command | Proves |
| --- | --- | --- |
| Build and typecheck | `go -C cli build ./... && go -C cli vet ./...` | nothing in `cli/` broke |
| Full Go suite | `go -C cli test -count=1 ./...` | self-apply, docs-consistency and hook tests still pass |
| Corpus conformance | `go -C cli test -count=1 -run TestCorpusConform .` | the three notes keep their records conformant |
| Rubric pin | `go -C cli test -count=1 -run TestArtifactContract .` | check 7's new text is pinned by its updated digest |
| Repository quality | `npm run quality` | formatting, static analysis, file size, TODO markers and the plugin smoke test |
| Payload untouched | `git diff --stat origin/main...HEAD -- plugins/trellis`, after U1-U3 are committed | no release, no `VERSION` bump |
| Records additive | `git diff origin/main...HEAD -- docs/decisions`, after U1-U3 are committed | added lines only |

A reviewer confirms the judgement parts: AE1-AE8 can be answered from `AGENTS.md` alone.

---

## Definition of Done

- U1, U2 and U3 are in the branch, and every command in the Verification Contract passes.
- No file under `plugins/trellis/` changed, and no decision record was added.
- The PR body lists `decision-0014`, `decision-0040` and `decision-0082` as the records it notes, and states that no line citation needed correcting.
- The PR body names the maintainer's rulings it rests on and lists the calls made without asking.
- No leftover edits from abandoned attempts remain in the diff.

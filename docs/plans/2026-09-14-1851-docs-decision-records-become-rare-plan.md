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
- **Means:** PR 1 of Simplify Trellis idea 1a (TRL-98) rewrites the operating method in `AGENTS.md` and the surfaces that restate it.
- **Product authority:** the maintainer, through `trellis-simplify`, who approved these requirements on 2026-09-14. PR 2 of idea 1a (TRL-99) and idea 1b are not active scope.
- **Open blockers:** none.

---

## Product Contract

### Summary

`AGENTS.md` admits a new decision record only when a wrong call would be expensive to undo, slow to notice, or seriously damaging while it stands. Any other change keeps its reasoning in its requirements doc, plan and PR, and puts a dated note under the title of each record it makes untrue. No new `superseded_in_part_by` pointer is written. PR 1 adds no record of its own and puts dated notes on the records whose account of the method it makes untrue.

### Problem Frame

`AGENTS.md:29-31` makes the records the method's home: *"a decision is the current truth"*, `docs/decisions/` is append-only, and you *"supersede with a forward pointer, never edit."* Changing any rule a record states therefore means writing a successor record and a pointer. The maintainer's ruling of 2026-09-05 made that explicit: *"Factual corrections get dated notes. Rule divergences get successors"* (`decision-0092:16-17`).

The corpus now holds 98 records and about 150,000 words. By the ideation's count, 53 records carry a supersession pointer, 47 of them partial. `decision-0089`'s pointer comment alone is a clause-by-clause diff of about 500 words (`decision-0089:6`). Forty records were added in July and fifteen in the first half of September. Of the 15 records added between 2026-08-31 and 2026-09-13, eight record a plugin behaviour change and landed in the same commit as a `plugins/trellis/VERSION` bump. The maintainer's reading: *"Should have more specs and less decisions."*

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
- PR 1 stays reviewable on a phone: `AGENTS.md`, `README.md`, the rubric text R15 reaches, a few one-paragraph notes, and the citation corrections they force.

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
- Inside `docs/`, the records `decision-0081`, `decision-0092`, `decision-0097`, `decision-0098`, `decision-0099` and `decision-0100` cite line numbers of those four records, and so do three plans (grep, 2026-09-14). The plans stay; the live citations in the records are R10's work in PR 1.
- `trellis-simplify` counts at least 153 `decision-NNNN:line` citations across the repository, some in `cli/` tests, `install.sh` and `plugins/trellis/`. A later PR whose note shifts a record cited under `plugins/trellis/` therefore edits the payload, and `AGENTS.md` makes that a release that bumps `VERSION`.
- `cli/artifact_contract_guard_test.go` pins each rubric check's text by digest, so rewording check 7 means updating its outcome rows in the same PR.

### Outstanding Questions

**Deferred to planning**

- The exact wording in `AGENTS.md`, and which real changes serve as the example on each side of the test.
- How a reader tells a live citation from a dated inventory inside a record, so R10 is applied the same way by every PR.
- The full set of records R14 reaches. Known candidates:
  - `decision-0040` point 5, which introduced partial supersession
  - `decision-0082` point 3, where supersession is marked by the forward pointer, including the partial form
  - `decision-0014`'s criterion that every "significant invariant change" gets a record
  - `decision-0015`'s open question on whether an appended block counts as an edit
- The full set of surfaces R15 reaches, including whether `profiles/trellis-self.md`'s evidence cell "`docs/decisions/` append-only" still holds.
- How PR 1's own dated notes cite a PR number that exists only once the PR is open.

### Sources

- Ideation, idea 1: `docs/ideation/2026-09-14-repo-simplification-ideation.html` (PR #311) and the Linear document "Ideation: simplifying the Trellis repository".
- `AGENTS.md:29-36` (operating method) and `:48-51` (the supersede, retire and record rows).
- `decision-0081:17` (heading), `:154-164` (retirement by successor record), `:324-329` (proposed directive).
- `decision-0082` Decision point 3; `decision-0040` Decision point 5; `decision-0092:16-17` (the 2026-09-05 ruling); `decision-0097` point 5 (plans are history); `decision-0014` Decision bullet 2; `decision-0015:125-135` (open question).
- `docs/rubrics/artifact-contract.md` check 7; `cli/corpus_conformance_test.go` check 7b.

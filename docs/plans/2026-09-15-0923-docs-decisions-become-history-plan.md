---
title: Decision records become history; living specs hold current truth - Plan
type: docs
date: 2026-09-15
topic: decisions-become-history
artifact_contract: ce-unified-plan/v1
product_contract_source: ce-brainstorm
execution: code
---

# Decision records become history; living specs hold current truth - Plan

## Goal Capsule

- **Objective:** An agent or the maintainer finds any rule of Trellis, or of how this repository is worked on, in one of three living specs, without reading the 99 decision records and their notes. The records stay in the tree as history.
- **Means:** Simplify Trellis idea 1b (TRL-103). The in-force rules of every record move into specs under `docs/specs/`, and one hand-over record marks every record superseded.
- **Product authority:** the maintainer, through the `trellis-simplify` orchestrator. He answered Q1 to Q4 on 2026-09-15 and has not approved these requirements. Ideas 3, 4, 5 and 7 are not active scope.
- **Open blockers:** Q5, Q6, Q7 and Q8 under Outstanding Questions, then the maintainer's approval. The project is paused until 2026-09-20.

---

## Product Contract

### Summary

Three living specs under `docs/specs/` (rules, delivery and repository) state every rule from the decision records that holds today. Each rule gets one line of why and pointers to the code and tests that carry it. One date-and-slug record hands authority to the specs and marks every record superseded. After it lands, a change to a rule edits a spec, and a record is written only when a change passes the record test.

### Problem Frame

`AGENTS.md` makes each decision record current truth, except for a point a dated note on it says no longer holds. To know one rule, an agent finds every record that touches it and reads each with its notes and pointer comments. The corpus is 99 records and about 154,000 words.

Most of that text is no longer a rule. A first-pass sort of every record found 7 records whose points all still hold, 73 that mix live and retired points, and 19 with nothing live. Five subagents did the sort, and it has not been cross-checked (`docs/plans/trl-103-sort/`). About 70 live rules are written nowhere except their record. One is test-first discipline for non-trivial logic (`docs/decisions/0023-code-tech-stack-and-dev-cycle.md:21`), which neither `AGENTS.md` nor `README.md` states.

Retirements are marked unevenly. 47 records carry clause-level `superseded_in_part_by` comments, and the sort found retirements with no mark at all.

Records and the tree disagree. `decision-0066:177` says the Codex startup claim "remains supported", while `README.md:132` says "Codex is not supported yet". `decision-0068:431-442` says the plugin carries no version-bump policy, while `.github/workflows/release-guard.yml:63` fails a payload change without a `VERSION` bump. The sort found about 30 such disagreements between records or living documents and the code.

Idea 1a (TRL-98, TRL-99) stopped new records and new partial pointers, but every existing rule still lives in a record. The maintainer on 2026-09-14: "Should have more specs and less decisions."

### Requirements

**The specs**

- R1. Three specs live in `docs/specs/`: rules, delivery, and repository (method, CI and release).
- R2. The rules spec points to the invariant set (`core/invariants/trellis-invariants-v1.md`) and the catalog for what they state, and does not restate them.
- R3. Positioning stays in `README.md`, and no spec states it.
- R4. A spec states each rule that holds today once, with one line of why, and points to the code and tests for the mechanics.
- R5. A spec cites the records each rule came from, as history.
- R6. No spec states what idea 3 retires: `profiles/`, `core/lexicon.md` and the profile schema section.

**What is not carried**

- R7. A commitment that was never built is dropped, and the PR that writes the spec lists each dropped commitment with its record.
- R8. Where a record and the code disagree and the code looks wrong, the spec states what the code does, and the case goes on one list for the maintainer to triage.

**The records**

- R9. Every record stays at its path, and the hand-over record marks each one with `superseded_by` naming the hand-over record.
- R10. Existing `superseded_in_part_by` lines and dated notes stay as history.
- R11. A Linear issue, blocked until idea 1b's last PR merges, asks the maintainer whether to delete the records from the tree once the specs have held.
- R12. The hand-over record is named by date and slug under `AGENTS.md`'s naming rule, and states that the specs are the authority and the records are history.

**The method after the hand-over**

- R13. A PR that changes a rule edits the spec that states it.
- R14. A record is written only when a change passes `AGENTS.md`'s record test, and it is history from its merge.
- R15. `AGENTS.md` no longer carries the rule that puts dated notes on records, or the rule that corrects line citations into records.

### Key Decisions

- **Three specs by subject, in `docs/specs/`.** Governs R1, R3. (session-settled: user-directed — chosen over five specs, one spec with sections, and folding the rules into `AGENTS.md` and existing documents: rules, delivery and repository hold 86 of the 99 records, idea 3 makes `README.md` the home of positioning, and one file per subject keeps each PR phone-sized.)
- **The rules spec points to the invariant set instead of restating it.** Governs R2. (session-settled: user-approved — the orchestrator's caveat on Q1, seen by the maintainer when he chose Q1 (a).)
- **A spec states what holds today, and never-built commitments are dropped.** Governs R4, R7, R8. (session-settled: user-directed — chosen over the records' full intent with "not built" markers, and over moving never-built commitments to the Ideas doc: delivery is the largest subject and idea 5 will rewrite it, so specs that restate mechanism would churn. The maintainer asked for one triage list rather than about 30 Linear issues.)
- **Records stay in place, and deleting them is decided later.** Governs R9, R10, R11. (session-settled: user-directed — chosen over stripping the partial pointer lines and over deleting the records now: deleting later stays cheap, while deleting before the sort is verified is costly to undo.)
- **Specs are the authority, and a record is history from its merge.** Governs R13, R14, R15. (session-settled: user-directed — chosen over keeping dated notes on records a later change makes untrue: a note on history has no reader.)
- **These specs are current truth, not the retired spec stage.** `decision-0079` retired specs that sat between decisions and code, duplicated the records and grew amendment pointers. These specs replace the records, are edited in place, and carry no amendment pointers. Compound Engineering plans under `docs/plans/` stay per-change history (`decision-0097` point 5). The invariant set already works this way (`core/invariants/trellis-invariants-v1.md:18-19`).

<!-- ce-section: work-relationships -->
### How This Work Fits Together

This plan covers idea 1b. The rest is the current understanding of the Simplify Trellis project, not a committed roadmap.

- Idea 3 (simplify-3): owns `README.md` positioning and retires `profiles/`, `core/lexicon.md` and the profile schema section.
  - Can proceed independently of this plan; whichever lands second adjusts to the other.
- Idea 4: deletes the corpus conformance tests and the artifact-contract rubric.
  - Depends on this plan's last PR merging, because those tests check this migration.
- Idea 5: one delivery path.
  - Edits the delivery spec in place once it exists.
- Idea 7: comments state current behaviour.
  - Removes many decision-id citations from code, which shrinks the sweep in Q6.

### Acceptance Examples

- AE1. **Covers R2, R4.** **Given** an agent that needs to know what happens when `.trellis/rules.toml` switches a rule off. **When** it reads the rules and delivery specs. **Then** it finds the rule, one line of why and pointers to the hook and its tests, without opening `decision-0083`, `decision-0084` or `decision-2026-09-15-rule-rows-only-switch-rules-off`.
- AE2. **Covers R7.** **Given** `decision-0016`'s Assess and Apply sub-agents, which were never built. **Then** no spec states them, and the rules spec's PR lists them as dropped with `decision-0016`.
- AE3. **Covers R4, R8.** **Given** `decision-0066:177` calling the Codex startup claim supported and `README.md:132` saying Codex is not supported. **Then** the delivery spec states that Codex is not supported, and the case stays off the triage list because the record is out of date, not the code.
- AE4. **Covers R8.** **Given** a record's rule that the code no longer follows and that no record, note or maintained file retired. **Then** the spec states what the code does, and the case goes on the triage list.
- AE5. **Covers R9, R10.** **Given** `decision-0083` after the hand-over merges. **Then** its frontmatter carries `superseded_by` naming the hand-over record, and the rest of its text, its notes included, is unchanged.
- AE6. **Covers R13, R14.** **Given** a PR after the hand-over that changes how `install.sh` renders rules and fails the record test. **Then** it edits the delivery spec and adds no record and no dated note.

### Success Criteria

- A fresh agent session answers "what is the rule for X" in rules, delivery or repository work from `docs/specs/` and `AGENTS.md`, without opening `docs/decisions/`.
- Every dropped rule is listed with its record and reason, and reviewed before it is dropped.
- Each PR stays reviewable on a phone.

### Scope Boundaries

- Git history.
- The text of existing records, apart from one `superseded_by` line each.
- The corpus conformance tests, which idea 4 retires after this work.
- Positioning and the identity sentence (idea 3).
- Deleting records from the tree (R11).
- Fixing code where a record and the code disagree (R8's triage list decides).
- Idea 5's delivery rewrite and idea 7's comment pass.

### Dependencies / Assumptions

- The sort under `docs/plans/trl-103-sort/` is a first pass and not cross-checked; its counts may move.
- `cli/corpus_conformance_test.go:715-757` reads the retired-artifact registry from `decision-0079` by id, and halts if the record or table is missing. R9 keeps the record in place, so the lookup keeps working.
- Check 7c (`cli/corpus_conformance_test.go:1076-1105`) reports an artifact other than a decision whose `depends_on` names a record carrying `superseded_by`. Once every record carries it, the check shows which artifacts still depend on a record.
- Rubric check 6 already recognises the `spec` type, with Acceptance criteria and Open questions as required sections (`docs/rubrics/artifact-contract.md:106`).
- `docs/specs/` is not a corpus root today (`cli/corpus_conformance_test.go:45-49`), and `cli/artifact_contract_guard_test.go:543-545` pins the rubric's corpus paragraph against the roots.
- `cli/docs_consistency_test.go:72,80` skips the top-level `docs/` directory, so its live-command check would not read `docs/specs/`.
- `cli/selfapply_test.go:204-208` fails if a `decisions/` or `research/` directory exists at the repository root.
- The shipped plugin carries no records, yet `plugins/trellis/hooks/staleness.sh:541,544` emit text citing `decision-0051` into sessions.

### Outstanding Questions

#### Resolve Before Planning

- Q5. **How does the maintainer approve which rules are carried and which are dropped, without reading 99 records?** 73 of 99 records are mixed, so the sort is per rule, not per record. The risk is a live rule dropped by mistake, and kept rules are visible as spec text.
  - (a) Each spec PR carries the spec plus a "not carried" table: every dropped rule, its record and a one-line reason. A second model (Codex Sol) checks each dropped rule first, and disagreements go at the top. The maintainer reads the spec and the table.
  - (b) One ledger for all 99 records, in the hand-over PR: one long review at the end.
  - (c) The maintainer sees only the rules where the two models disagree or confidence is low: least reading, and it trusts two-model agreement.
  - Recommendation: (a). His attention goes where the risk is, one subject at a time, and the hand-over record keeps the combined ledger as history.
- Q6. **How far does the decision-id citation sweep reach?** Record ids appear in the Go tests (320 comment lines, 131 other lines), `install.sh`, CI workflows and the shipped plugin (both hooks, the invariants reference, the remove skill). Any plugin edit is a release. Idea 7 rewrites these comments, and idea 5 deletes much of the hook code.
  - (a) Re-point what claims current authority: artifact frontmatter, `AGENTS.md`, `README.md`, the rubric, `core/` documents and the conformance test's lookups. Comment ids in code, tests and the plugin stay, and still resolve from the record through `superseded_by` to the hand-over ledger and the spec. No release.
  - (b) Re-point every citation, plugin included: a full sweep, a plugin release, and work ideas 5 and 7 redo.
  - (c) Like (a), plus comments in `cli/` and `install.sh`, still with no plugin release.
  - Recommendation: (a). Check 7c fails while any artifact still depends on a superseded record, so it proves the load-bearing edges moved. The rest is history that ideas 5 and 7 rewrite. This departs from the ideation's full sweep.
- Q7. **How does the work split into PRs?** Every PR must be reviewable on a phone. The ideation put citations first, but citations have nothing to point at until a spec exists.
  - (a) One PR per spec (rules, repository, then delivery), each re-pointing its subject's citations, with a spec current truth from its merge where it covers a record. Then a hand-over PR marks every record superseded and rewrites `AGENTS.md`'s method: about four PRs, and specs are usable early.
  - (b) All specs first with no authority, then one hand-over PR that switches authority at once: a clean switch, but specs and records drift apart while PRs wait.
  - (c) One PR: too large for a phone.
  - Recommendation: (a). Each PR stands alone, and authority moves with the content.
- Q8. **Do plans keep dated notes?** This question came up while folding in Q4. R15 retires dated notes on records, but `decision-0097` point 5 also corrects a plan that turns out wrong about what shipped with a dated note beside the claim, and the repository spec has to state one rule or the other.
  - (a) Plans keep dated notes; only the rule for records retires. A plan stays an honest account of its work.
  - (b) Dated notes retire everywhere, and plans are left as written. The method gets one rule fewer.
  - Recommendation: (b). Nobody maintains plans (`decision-0097` point 3), and the spec and the PR carry what shipped.

#### Deferred to Planning

- Where the retired-artifact registry in `decision-0079` lives once records are history, and how the conformance test finds it.
- Adding `docs/specs/` to the conformance corpus: the rubric's corpus paragraph, the test's roots and the guard.
- Whether the live-command check in `cli/docs_consistency_test.go` should read `docs/specs/`.
- The format and home of R8's triage list.
- The hand-over record's slug.

### Sources

- Ideation, idea 1: `docs/ideation/2026-09-14-repo-simplification-ideation.html` and the Linear document "Ideation: simplifying the Trellis repository".
- Idea 1a plans: `docs/plans/2026-09-14-1851-docs-decision-records-become-rare-plan.md` and `docs/plans/2026-09-14-2237-refactor-date-slug-decision-records-plan.md`.
- `AGENTS.md` *Operating method*; `decision-0014` (the invariant set as a revise-in-place spec); `decision-0079` (the spec stage retired); `decision-0082` (no `status`; merge is acceptance); `decision-0097` (Compound Engineering plans are history).
- The first-pass sort: `docs/plans/trl-103-sort/batch-1.md` to `batch-5.md`.

---
title: A record written from a plan copies summary prose that its unit detail contradicts
date: 2026-10-20
category: workflow-issues
module: docs/decisions
problem_type: workflow_issue
component: documentation
severity: high
applies_when:
  - Writing a decision record from a merged plan's Key Decisions, KTDs or High-Level Technical Design prose
  - Restating a plan's summary sentence as a normative rule in a record
  - Superseding in part a record whose tables define states or verbs, such as decision-0073 D1
  - Applying a review finding to a record written from a plan
tags: [decision-records, plans, summary-drift, normative-wording, unit-test-scenarios, verb-tables, supersession, review]
---

# A record written from a plan copies summary prose that its unit detail contradicts

## Context

TRL-122 wrote `docs/decisions/2026-10-20-trellis-delivers-rules-as-committed-files.md` from the vendored-delivery plan, `docs/plans/2026-10-09-1945-refactor-trellis-delivery-direction-plan.md`, which merged in PR #339. The plan's unit U4 told the author to "Write the record from this plan's Key Decisions and KTD4–KTD18" (plan:528).

A plan states one fact at several levels. A Key Technical Decision (KTD) or a High-Level Technical Design paragraph says it in a sentence. The verb table, the state diagram and each unit's test scenarios say it case by case. In this plan the levels did not always agree, and where they differed the summary was the looser one.

The record kept inheriting the summary wording. The plan step accepted U4 as enough to build from, so nothing reconciled the summary against the unit detail before the record was written. The corpus conformance test passed before any reviewer had read the record: it checks a record's shape, not its fidelity to the plan (session history).

Reviewers then found the drifts one at a time. The session review (ce-code-review with its Codex adversarial pass, ce-doc-review, and a standalone Codex Sol pass) found the first four drifts below, the fourth only in part. A Codex Sol re-check of the first fix found the rest of the fourth, and the Codex reviewer on PR #340 found the fifth after both passes. Each fix moved a sentence toward the plan's most specific statement, and none changed the design.

## Guidance

For each normative sentence in a record you write from a plan:

1. Find every place the plan states the same thing: the KTD, the High-Level Technical Design, the verb or state table, the state diagram, and the test scenarios of the unit that will build it.
2. Read what the governing records say, above all the record the new one supersedes in part. The new record often keeps its definitions. The record's point 5 keeps decision-0073 D1's table as the definition of S0–S6, so every term it imports means what D1's table says. Where the plan describes existing code, read the code. The plan said `cli/apply.go` "writes" the codex-bootstrap block (plan:225), but the file's header says no code in that package writes into a consuming project (`cli/apply.go:3-8`). The record's point 5 now says the block is what `cli/apply.go` "renders into the payload".
3. Write the sentence from the most specific statement the plan's authority order allows. Plan:18 sets that order: "the Product Contract's requirements win on behaviour; a KTD wins on mechanism within them; a unit overrides neither." Inside it, a more specific statement refines a broader one: a KTD's detailed bullet outranks the one-line summary elsewhere that cites it. A unit's test scenario does not outrank a KTD. Where two statements contradict rather than one refining the other, do not choose between them: raise the contradiction with the lead, as the plan's stop conditions have a unit do "when a KTD cannot hold" (plan:20).
4. Check each scope word in the sentence, such as "any of S0–S6", "outside S6" or "the repo's files". List the cases the specific source enumerates, and confirm the sentence gets each one right.
5. When a reviewer reports one drift, re-derive the whole sentence from the full table row or column. Do not patch only the case the finding names.

Cite the record you write, and the record it supersedes, by point or table row rather than by line. A dated note added later moves every line below the title (see Related).

## Why This Matters

Once merged, the record is current truth and the plan is not. AGENTS.md lets a dated note point only at a file kept current on `main`, "never a plan or a PR body". A summary sentence that was harmless in the plan, because the unit detail sat beside it, becomes normative on its own in the record.

The next units build and test against the plan's detail. U5's test scenarios include a tag-only morph (plan:547). A record that defines the probe as reading "the repo's files" contradicts that test, and the implementer has to choose between them.

These drifts change behaviour, not only wording. Four of the examples below describe a run that does the wrong thing: an empty repo reported as adopted with nothing written, opt-outs lost in a crash, a morph partly modified before rollback, a tag-only morph migrated. ogham pins the record's point 6 outcome contract, so after the first tag that point changes only through a new record.

Review catches these one at a time. Here it took three review passes. Deriving each sentence from the detail while writing is cheaper than waiting for a reviewer to find each one.

## When to Apply

- You write a decision record, a contract or any other normative text from a plan that states the same facts at more than one level.
- The plan's own instruction points you at its summary sections, as U4 pointed at the KTDs (plan:528).
- The new record supersedes another in part and keeps some of its definitions.
- You apply a review finding to such a record.

Context and provenance prose that sets no behaviour does not need this check.

## Examples

"First draft" line numbers refer to the record at 612ee5b, its first commit; the later drafts are in PR #340's commit history. "Point N" names a point of the record as it stands now. Plan lines are the merged plan's.

**1. An empty repo read as `already-present`.** KTD6 says "a repo in any of S0–S6 gets `already-present <state>`" (plan:203). Decision-0073 D1's S0 row includes the empty repo: "with nothing present it is the unadopted state". The verb table gives `none` its own row, "`installed` / writes all" (plan:303), separate from the S0–S5 row (plan:305).

The first draft followed KTD6 and D1. It said that with nothing present a repo "is in D1's S0-unadopted state, which the CLI calls `none`" (first draft:115-116), and that "`install` in a repo in S0–S6 or codex-bootstrap reports `already-present <state>` and changes nothing" (first draft:147-148). Read together, an empty repo gets a success line and no files. The session review merged this as a P1: the Codex adversarial pass rated it P1 and the correctness reviewer P2.

The record's point 5 now orders the states so that `none` is its own state, and point 6 gives `install`'s outcome row by row: "In `none` it reports `installed` and writes every file; only this case writes."

**2. Migration's write order.** The design section summarises: "Every conversion follows KTD8's order: it writes the new artifacts first, then removes the old ones, then writes the lock" (plan:317). KTD8's detailed bullet adds the exception: "Where an old artifact holds opt-outs that no `rules.toml` does yet (S3's overlay, or an S4 block in a repo with no `rules.toml`), migration writes `rules.toml` first, before the section replaces any old block" (plan:214).

The first draft kept the fixed order, which puts the `AGENTS.md` section before `rules.toml`. It softened the exception to "`rules.toml` carries them before any old artifact goes" (first draft:173-177). Writing the section replaces an S4 block (plan:327), so a crash between the section and `rules.toml` loses the only copy of the opt-outs. The Codex adversarial pass inside ce-code-review rated this P1, though that run's validator rejected it, and the standalone Codex Sol pass rated it P2. The fix applied it all the same: the record's point 7 now names the exception and the cases it covers.

**3. The git index.** KTD8 says the CLI "never reads HEAD or touches the git index" (plan:212). The first draft paraphrased it as "The CLI never reads HEAD or the git index" (first draft:169), turning "touches" into "reads". KTD15 lists nested `AGENTS.md` files with `git ls-files -z --cached --others` (plan:239), and `--cached` lists paths from the index, so the paraphrase was false of the plan's own design. The session's ce-doc-review flagged it as a P2. The first fix pass narrowed it, and point 7 now says the CLI "reads file content from the working tree, never from HEAD or the git index, and it never changes the index".

**4. A fix that patched only the reported case.** The verb table lets `remove` act in only four states: `none`, where it does nothing, plus `declined`, `vendored-incomplete` and `vendored` (plan:303-308). In S0–S5 and codex-bootstrap it reports "run install --migrate" (plan:305). In S6 it reports "roll back first" (plan:306). A mixed state answers "run install --migrate" (plan:315).

The first draft said `remove` "deletes everything Trellis wrote except `rules.toml`" (first draft:203-204). That was R6's own wording, which names no states: "`remove` deletes every file and section Trellis wrote except `.trellis/rules.toml`" (plan:67). The session review flagged the S6 case as a P1. The first fix added "Outside S6 (point 6)", which patched only that case. The Codex Sol re-check then found `remove` still acting in S0–S5, codex-bootstrap and mixed repos. The record's point 9 now reads R6 through the verb table's `remove` column, and names the column's acting states instead of an exception: "In a `vendored`, `vendored-incomplete` or `declined` repo, it deletes … In `none` it reports `done` and changes nothing, and in S0–S6, codex-bootstrap or a mixed state it changes nothing and reports `error` (point 6)." That is a reading of R6, not R6's text.

**5. A probe defined by files alone.** KTD11's closing paragraph says: "The probe is a pure function from the repo's files to a state, plus the evidence that decided it" (plan:229). U5's test scenario says: "A `trellis-pre-morph` tag with no `.trellis/rollback` reads as S6 … The tag reaches the function as an input, so it stays pure" (plan:547). Decision-0073 D1's S6 row defines S6 as "`.trellis/rollback` and/or the `trellis-pre-morph` tag".

The record copied plan:229 almost word for word, and two review passes left it. A git tag is not a file. A probe built to that sentence cannot see a tag-only morph, so it would read the repo as `none` or an old state, and `install --migrate` could modify a repo the record says it refuses. The Codex reviewer on PR #340 flagged the sentence as a P1. The record's point 5 now says the probe "takes the repo's files and whether the `trellis-pre-morph` tag exists", passes the tag in as an input so the function stays pure, and reads a tag with no `.trellis/rollback` as S6. The fix is in PR #340.

## Related

- `docs/solutions/workflow-issues/a-dated-note-moves-line-citations-a-one-line-search-misses.md`: the same area's other trap. A record's text silently diverges from what it points at, CI does not catch it, and only an independent review does. Its "open the target, never add the delta" is the citation-level form of this doc's "write from the most specific statement".
- `docs/rubrics/artifact-contract.md`, check 5: decides which records go in `depends_on`, but not which plan text a record's wording comes from.

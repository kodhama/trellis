---
title: Move the artifact-contract rubric to docs/rubrics and its fixtures to cli/testdata - Plan
type: refactor
date: 2026-09-14
artifact_contract: ce-unified-plan/v1
product_contract_source: ce-plan-bootstrap
execution: code
---

# Move the artifact-contract rubric to docs/rubrics and its fixtures to cli/testdata - Plan

## Goal Capsule

- **Objective:** The artifact contract, which nothing Trellis ships needs, sits beside the governance corpus it governs instead of in `core/`. The CI conformance check still gates that corpus and still rejects the seeded violations.
- **Means:** `git mv` of the rubric and the fixture directory, the Go path constants and `cli-ci` filters re-pointed, a token-only path sweep, and the rubric's agent wording rewritten (KTD1, KTD2, KTD3, KTD5, KTD6).
- **Authority:** Linear TRL-95 and the maintainer's ruling recorded in `decision-0098` point 4, as relayed by `trellis-d1`; then the records on `main` (`decision-0005`, `0010`, `0015`, `0028`, `0040`, `0082`, `0098`, `0099`); then this plan.
- **Stop conditions:** send a DECISION to `trellis-d1` when a record needs more than a path token or a forward pointer, an unlisted mention might fit a KTD3 keep class, an edit would move a live cited line (KTD4), a file under `plugins/trellis/` changes, `decision-0100` is taken, or a merge conflict needs judgment.
- **Execution profile:** red then green on the Go conformance tests around the rename (U1, U2), with mutation checks that the relocated tests still fail when pointed at the wrong place. The rest is mechanical, proved by grep and the full suite.
- **Who finishes:** the `trl-95` worker session implements, runs review and repair, and opens a ready-for-review PR. The maintainer merges.

---

## Product Contract

### Summary

Move `core/rubrics/artifact-contract.md` to `docs/rubrics/artifact-contract.md` with `scope: trellis-meta`, and move `core/fixtures/` to `cli/testdata/`.
Point the Go conformance tests, `cli-ci`'s path filters, live instructions, records and planning records at the new paths.
Rewrite the rubric's sub-agent wording so it describes a repository-internal contract that a Go test enforces.
Whether a record is owed for taking the rubric out of `core/` waits on `trellis-d1` (Open Questions).

### Problem Frame

On 2026-09-14 the maintainer ruled that the rubric leaves `core/`: *"I think it should, unless it's needed by the plugin itself."* The plugin does not need it. Nothing under `plugins/` or in `install.sh` names it, the shipped `plugins/trellis/reference/` holds no rubric, and the CLI's one non-test mention is a comment at `cli/payload.go:29` saying the payload leaves artifact-contract metadata out. `decision-0098` point 4 records the ruling and names TRL-95 as the move.

`core/` is Layer A, the shippable product (`decision-0005`). The rubric and its known-bad fixtures sit there, though only this repository's Go tests and documents read them. PR #309 kept the rubric's sub-agent wording for that reason: its repository note at `core/rubrics/artifact-contract.md:39-40` says the wording "describes the product's form for a consumer project".

Two records still describe the contract as a product resource. `decision-0010` Decision bullet 1 lists "the artifact contract and its conformance check" among Trellis's resources, and bullet 2 calls the validator "a conformance sub-agent applying a rubric". `decision-0098`'s forward pointer on `decision-0010` says bullet 2 stands for the product "provisionally, while decision-0098's open question on whether the rubric leaves core/ is unanswered". `decision-0098` now marks that question answered, and lists "what TRL-95 owes the records" as not decided there.

### Requirements

**Location**

- R1. The rubric is at `docs/rubrics/artifact-contract.md` and the fixture corpus at `cli/testdata/known-bad/`, each with its git history. `core/rubrics/` and `core/fixtures/` no longer exist, and `core/schemas/typed-artifacts.md` stays.
- R2. The rubric declares `scope: trellis-meta`.

**Enforcement**

- R3. The conformance check reads the live corpus with `docs/rubrics/` in it and passes, and it rejects the corpus at `cli/testdata/known-bad/` with exactly the expected findings.
- R4. A PR that touches only the rubric or only the fixtures still runs `cli-ci`'s `build-test`.

**Pointers and wording**

- R5. Every mention of the two moved paths follows the move per `decision-0015:117-121`, except the keep classes in KTD3.
- R6. The rubric describes a contract internal to this repository and enforced by the Go tests, with no wording that presents it as a sub-agent gate for consumer projects.
- R7. `core/README.md` no longer lists `rubrics/` or `fixtures/` among `core/`'s contents.
- R8. Every `file:line` citation into a file this change edits still points at the same content (KTD4).

**Records and payload**

- R9. No record is written or superseded until `trellis-d1` answers the record question in Open Questions.
- R10. Nothing under `plugins/trellis/` changes, so `plugins/trellis/VERSION` does not move.

### Key Decisions

- **The rubric goes to `docs/rubrics/` with `scope: trellis-meta`, and the fixtures go to `cli/testdata/known-bad/`.** (session-settled: user-directed — chosen over keeping both in `core/`: the plugin does not need the rubric, and only the Go control test reads the fixtures.) Governs R1, R2.
- **The typed-artifacts schema stays in `core/`.** (session-settled: user-directed — chosen over moving it with the rubric: it defines the catalog and profile, which are product types.) Governs R1.
- **The sweep reaches append-only records, changes only the path token, and leaves commit-pinned references alone.** (session-settled: user-directed — chosen over leaving records unswept: the maintainer chose this for PR #306's `site/` move and PR #308's `docs/` move under `decision-0015:117-121`.) Governs R5.

### Scope Boundaries

- No record substance changes beyond a path token, unless `trellis-d1` answers the record question with a record (U7).
- No numbered rubric check changes, so no outcome-row digest moves.
- No tripwire test for `core/rubrics/` or `core/fixtures/` reappearing (A5).
- The known-bad fixtures keep their synthetic `decision-0100` id. That corpus is self-contained, and the decision-id guard counts only `docs/decisions/NNNN-*.md` paths (`.github/scripts/decision-id-guard.sh:75-80`).
- `core/README.md`'s Contents list omits `catalog/` and `lexicon.md` today. That gap predates this change and stays.
- `docs/decisions/0082-retire-the-status-field.md:145` cites `core/fixtures/known-bad.md`, a file #309 already moved into `known-bad/`. The sweep changes its token only.
- Links from outside the repository to `blob/main/core/rubrics/…` break, because GitHub does not redirect moved paths. Linear's TRL-88, TRL-92 and TRL-95 text is not swept.

---

## Planning Contract

### Key Technical Decisions

- KTD1. **The move lands as a pure rename commit.** `core/rubrics/artifact-contract.md` and the whole `core/fixtures/` directory move with `git mv` in a commit with no content change, so `git log --follow` traces every file. `core/fixtures/README.md` becomes `cli/testdata/README.md` and `core/fixtures/known-bad/` becomes `cli/testdata/known-bad/` (A3). At that commit the conformance tests fail, because their constants still name the old paths. That is the red state U2 turns green.
- KTD2. **The fixture root stays a `../`-prefixed literal: `../cli/testdata/known-bad`, not Go's bare `testdata/known-bad`.** `corpusDisplayPath` strips `../` so findings render repository-relative paths, and `contractRootSpelling` spells the rubric's corpus paragraph the same way. A bare literal would render `testdata/known-bad/…` and force the rubric to name `testdata/`, which is not repository-relative. `cli/ci_paths_guard_test.go` reads the literal as `cli/testdata/known-bad`, which `cli/**` covers. In `cli/corpus_conformance_test.go`, `corpusRoots` swaps `../core/rubrics` for `../docs/rubrics`, `artifactContractPath` becomes `../docs/rubrics/artifact-contract.md`, and `wantRendered` becomes `cli/testdata/known-bad/known-bad.md:4: check 4 (4i): `. In `cli/artifact_contract_guard_test.go`, the seeded case expecting `"core/fixtures/"` expects `"cli/testdata/"`.
- KTD3. **Sweep classes.** The token map is `core/rubrics/` → `docs/rubrics/` and `core/fixtures/` → `cli/testdata/`. It applies wherever the token appears, globs and code literals in planning records included. A line the swap makes anachronistic is still swept, because `decision-0015:117-121` holds that reasoning moves with the token. A mention is kept only when it is:
  1. text that describes this move: `decision-0098:132-134` (point 4's destinations and "Until TRL-95 lands"), this plan, and `decision-0100` if written;
  2. not a path: "known-bad" as a word (`core/invariants/trellis-invariants-v1.md:334`, `cli/codex_hook_test.go:1096`), the fixture id `fixture-known-bad`, and the artifact id `rubric-artifact-contract`;
  3. a different rubric: `agentic-dev-meta-layer-brief.md:321` names `rubrics/spec-quality.md`, which never lived in `core/rubrics/`.

  These follow `decision-0099` point 8's classes. A commit-pinned mention would also be kept; none was found at planning time, and U6 re-checks. The Appendix lists every line swept and kept.
- KTD4. **Edits keep cited lines in place.** Records and plans cite fixed lines in files this change edits, among them `artifact-contract.md:5,20,39,53-57,91`, rubric line 81 in `docs/superpowers/specs/2026-09-03-row-set-guard-design.md:32`, `cli-ci.yml:11-28`, `cli-ci.yml:17` and `AGENTS.md:103`. Every edit to a pre-existing file is a same-line substitution or a rewording that keeps each hunk at its line position. Swept Markdown is never re-wrapped. Two edits touch cited lines, and neither breaks a live citation, because earlier changes already rewrote the cited content. `core/README.md` loses its two Contents lines, which shifts the lines `decision-0076:174` cites as `core/README.md:13–16`. KTD6 rewrites the rubric's `:39-40`, which `decision-0076:94` cites. Neither triggers the stop condition.
- KTD5. **`cli-ci`'s two filter lists drop `core/rubrics/**` and `core/fixtures/**`, with no replacement.** `docs/**` already selects `docs/rubrics/` and `cli/**` selects `cli/testdata/`. Nothing reads the old paths, and `cli/ci_paths_guard_test.go` only requires that read paths are covered. The header comment at `cli-ci.yml:19-22` names the rubric under `docs/**` and the fixtures under `cli/**`, keeping its line count. The entries `decisions/**`, `research/**` and `.grove/**` stay: `cli/selfapply_test.go` checks those paths are absent.
- KTD6. **The rubric's wording names the Go test and drops the consumer framing, and its section headings stay.** Header, grading and acceptance text sit outside every numbered check's digest window (`cli/artifact_contract_guard_test.go:58-61`), so no outcome row moves. On `main` at `a3ffa70`, the lines that change are:
  1. `:13`: the Go conformance test applies the gate, not "the conformance sub-agent";
  2. `:39-40`: the sentence that says the sub-agent wording describes the product's form is replaced by a statement that the contract is internal to this repository, because nothing Trellis ships reads it (`decision-0098` point 4);
  3. `:166`: the test reports one finding per violation, as `path:line: check N (rule): detail`, and a run with no finding passes;
  4. `:173`: the acceptance criterion names the Go test that `cli-ci` runs (`decision-0098`), in place of "applied by an agent with no runtime (`0010`)".

  The "How it is graded" row of `contractClauseOutcomes` in `cli/corpus_conformance_test.go` drops "the section's consumer-facing agent wording stays".
- KTD7. **`core/README.md` drops its `rubrics/` and `fixtures/` Contents lines, and its conformance paragraph (`:14-21`) stops presenting an artifact-contract sub-agent as a product resource.** The paragraph says the artifact contract, `docs/rubrics/artifact-contract.md`, is internal to this repository and applied by `cli/corpus_conformance_test.go`. Its opening no longer says a conformance sub-agent "applies these", and the `core/agents/` product-home sentence no longer reads as the home of an artifact-contract sub-agent. The paragraph keeps its line count. The general line about product resources at `:3-4` ("rubrics, sub-agents, conventions") stays, because it names resource kinds.

### Assumptions

These are agent bets that no one has confirmed. `trellis-d1` can override any of them.

- A1. The record question is answered with a record, `decision-0100` (Open Questions). If `trellis-d1` answers otherwise, U7 is dropped and `decision-0010`'s pointer stays as it is.
- A2. `AGENTS.md:102`'s corpus parenthetical gains `docs/rubrics/`, not just the token swap, because after the move the rubric's own directory is no longer among "the `core/` artifacts listed there". The `inv-gate-at-handover` row in `profiles/trellis-self.md:103` quotes that sentence word for word, so the quote changes the same way (`decision-0015:117-121`: a quotation tracks its source). The alternative is a token-only swap in both places, which leaves the parenthetical silent about `docs/rubrics/`.
- A3. The fixture README moves with its directory to `cli/testdata/README.md`. The ruling names only `known-bad/`. The README cannot go inside `known-bad/`: the check walks every `.md` under its root (`cli/corpus_conformance_test.go:442-446`), and a file without frontmatter is a rule 1a finding the expected set does not name.
- A5. No tripwire test is added for `core/rubrics/` or `core/fixtures/` reappearing. TRL-89 added one for root `decisions/` because the decision-id guard cannot see a root record. A stray old path here gets no such blindness guard, and the risk is a stale branch adding a fixture outside the control test's root.
- A6. The rubric keeps its legacy `status: ratified`, `ratified: 2026-07-03` and `owner: gundi` lines (`decision-0082` preserves `status:` on artifacts that predate it).

### Open Questions

- **Blocking for U7 only: is a record owed for taking the rubric out of `core/`?** Sent to `trellis-d1` as a DECISION.
  - (a) No record. `decision-0098` point 4 already records the ruling. `decision-0010`'s pointer keeps saying bullet 2 stands for the product "provisionally" on a question `decision-0098` calls answered. Bullet 1's parenthetical, which the same pointer marks as standing, keeps naming the artifact contract a product resource.
  - (b) `decision-0100`. It records the rubric as repository-internal (`scope: trellis-meta`, Layer B under `decision-0005`, which stands) and supersedes in part `decision-0010` Decision bullet 1's parenthetical "(including the artifact contract and its conformance check)" and bullet 2 as it describes the product. `decision-0010` gains the forward pointer.
  - The author recommends (b). A forward pointer must name a record, and `decision-0098` says it did not decide what TRL-95 owes the records. TRL-89's move was also recorded, as `decision-0099`, though that record had a guard change to supersede as well. U1–U6 do not wait on the answer.
- **Not blocking, for `trellis-d1`: `core/invariants/trellis-invariants-v1.md:148`.** This `trellis-product` artifact names `rubric-artifact-contract` as "the deferred conformance-to-upstream check", so after the move a product artifact names a repository-internal one. The sweep matches paths, not ids, so no unit touches the line, and a change under `core/invariants/` is the maintainer's call. It may belong in `decision-0100`'s scope, or stay as written.

---

## Implementation Units

### U1. Rename the rubric and the fixture directory

- **Goal:** both artifacts are at their new paths with history intact (KTD1).
- **Requirements:** R1.
- **Dependencies:** none.
- **Files:** `core/rubrics/artifact-contract.md` → `docs/rubrics/artifact-contract.md`; `core/fixtures/README.md` → `cli/testdata/README.md`; `core/fixtures/known-bad/` → `cli/testdata/known-bad/`.
- **Approach:** a rename-only commit. Record the conformance tests' failing output at this commit for the PR's Validation section.
- **Execution note:** this is the red state. The conformance tests must fail here because they cannot read the rubric at its old path.
- **Test expectation:** none -- rename only; U2 carries the tests.
- **Verification:** `git log --follow` traces the rubric and a fixture file past the rename, and `core/rubrics/` and `core/fixtures/` are gone.

### U2. Point the conformance tests and the rubric's corpus paragraph at the new homes

- **Goal:** the conformance suite passes against the moved files (KTD2).
- **Requirements:** R3, R5.
- **Dependencies:** U1.
- **Files:** `cli/corpus_conformance_test.go`, `cli/artifact_contract_guard_test.go`, `docs/rubrics/artifact-contract.md` (the **Corpus:** paragraph at `:20-21` and the positive-control path at `:168`).
- **Approach:**
  1. Change the constants, `wantRendered` and the seeded guard case per KTD2.
  2. Swap the path tokens in the comments at `cli/corpus_conformance_test.go:4,14,54` and `cli/artifact_contract_guard_test.go:16`.
  3. Swap `core/rubrics/` and `core/fixtures/` in the rubric's **Corpus:** paragraph, keeping the bold label and the `**Derived resource` anchor that `contractWindow` matches.
- **Patterns to follow:** PR #308's one-line `corpusRoots` change for the `docs/` move.
- **Test scenarios:**
  - `TestCorpusConformanceRejectsKnownBadFixture` over `../cli/testdata/known-bad` yields exactly `knownBadExpected`, and one finding starts with `cli/testdata/known-bad/known-bad.md:4: check 4 (4i): `.
  - The live corpus run over `corpusRoots` yields no finding, and `docs/rubrics/artifact-contract.md` is among the files it loads.
  - `TestArtifactContractCorpusMatchesCorpusRoots` passes with the paragraph naming `docs/rubrics/` and `cli/testdata/`.
  - `TestArtifactContractGuardComparisonsFail`'s "names no path" case reports `"cli/testdata/"`.
  - `TestCIPathFilterCoversEveryPathTheSuiteReads` passes, reading `cli/testdata/known-bad` and `docs/rubrics`.
  - Mutation, not committed: with `knownBadRoot` set back to `../core/fixtures/known-bad`, the control test fails, by halting or by naming missing expected findings.
  - Mutation, not committed: with `../docs/rubrics` removed from `corpusRoots`, the corpus guard fails naming `docs/rubrics/`.
- **Verification:** `go test -count=1 -run 'TestCorpusConform|TestArtifactContract' .` in `cli/` goes from U1's red to green, and both mutations turn it red.

### U3. Update `cli-ci`'s path filters and header comment

- **Goal:** the filters select the new homes and name no removed path (KTD5).
- **Requirements:** R4, R8.
- **Dependencies:** U2.
- **Files:** `.github/workflows/cli-ci.yml`.
- **Approach:** remove `core/rubrics/**` and `core/fixtures/**` from both lists, which stay identical. Reword the comment at `:19-22` without changing the line count.
- **Test scenarios:**
  - `TestCIPathFilterCoversEveryPathTheSuiteReads` passes, and it checks both lists.
  - The `pull_request` and `push` lists are byte-identical.
- **Verification:** the workflow file's line count is unchanged, and no `core/rubrics` or `core/fixtures` token remains in it.

### U4. Set the rubric's scope and rewrite its agent wording

- **Goal:** the rubric reads as a repository-internal contract (KTD6).
- **Requirements:** R2, R6, R8.
- **Dependencies:** U2.
- **Files:** `docs/rubrics/artifact-contract.md` (`:8`, `:13`, `:39-40`, `:166`, `:173`), `cli/corpus_conformance_test.go` (the "How it is graded" row of `contractClauseOutcomes`).
- **Approach:** `scope: trellis-product` becomes `scope: trellis-meta` on the same line. Rewrite the four passages per KTD6, each within its current line count. Leave the legacy lines (A6), `depends_on`, and every numbered check untouched.
- **Test scenarios:**
  - The live conformance run passes, so rule 2b accepts `trellis-meta` on the rubric.
  - The numbered-check digest test passes with no digest changed in `contractOutcomeTable`.
  - A search of the rubric for "sub-agent" and "agent" finds no wording that presents the contract as a consumer-facing agent gate.
- **Verification:** every hunk in the rubric keeps its line position (KTD4).

### U5. Sweep live instructions, product docs and the profile

- **Goal:** every live mention points at the new paths (KTD3, KTD7, A2).
- **Requirements:** R5, R7, R8.
- **Dependencies:** U1, U2.
- **Files:** `AGENTS.md:102`, `README.md:283`, `core/README.md:10,12,14-21`, `cli/testdata/README.md:1`, `core/schemas/typed-artifacts.md:20`, `profiles/trellis-self.md:101,103,112`.
- **Approach:** token swaps per KTD3, with three exceptions. `core/README.md` drops two lines and rewords its conformance paragraph (KTD7). `AGENTS.md:102` and the profile's quote at `:103` both gain `docs/rubrics/` (A2). The fixture README's title becomes `` `cli/testdata/` ``. Profile cells change only their path tokens and the tracked quote, so each row still shows the tell its verdict claims.
- **Test scenarios:**
  - The live conformance run passes, including rule 10's evidence-present check on `profiles/trellis-self.md` and the schema's own checks.
  - `cli/row_set_guard_test.go` and `cli/selfapply_test.go` pass against the edited `profiles/trellis-self.md` and `AGENTS.md`.
- **Verification:** no live file outside the Appendix keep list mentions `core/rubrics` or `core/fixtures`, and every edited file except `core/README.md` keeps its hunks at their line positions.

### U6. Sweep records and planning records

- **Goal:** citations in append-only records and plans follow the move (KTD3).
- **Requirements:** R5, R8.
- **Dependencies:** U1, U2.
- **Files:** the records and plans in the Appendix's swept table.
- **Approach:** replace path tokens only, with no re-wrap. Re-run the repository-wide search for both tokens. A hit the Appendix does not list is swept by default. The unit stops only when such a hit might fit one of KTD3's keep classes (Goal Capsule stop conditions).
- **Test expectation:** none -- records only; the sweep search and the live conformance run prove this unit.
- **Verification:** the only remaining hits are the Appendix's kept lines, this plan, and `decision-0100` if written.

### U7. Record the move as `decision-0100` (only if `trellis-d1` answers (b))

- **Goal:** the record graph states that the artifact contract is not a product resource.
- **Requirements:** R9.
- **Dependencies:** U4, U5, and the answer to the Open Question.
- **Files:** `docs/decisions/0100-<slug>.md` (new), `docs/decisions/0010-no-runtime-agent-instructions.md` (the frontmatter pointer only), `docs/rubrics/artifact-contract.md` (`depends_on` gains `decision-0100`).
- **Approach:**
  1. Confirm `decision-0100` is free on `main` and in every open PR, and record the check in the PR body.
  2. Write the record without a `status` field, shaped like `decision-0098`: Context, Decision, Consequences, Open questions and Self-check. It records the location and scope, names what it supersedes in `decision-0010` word for word with a STANDS clause for the rest, and states that `decision-0005` stands.
  3. Append `decision-0100` to `decision-0010`'s `superseded_in_part_by` on the same line, with its own clause. `decision-0098`'s existing clause stays as written.
- **Test scenarios:**
  - The live conformance run passes with the new record, which includes the partial-supersession pointer check (7b) and reference resolution (check 4).
  - `decision-id-guard` on the PR reports no clash for `0100`.
- **Verification:** `decision-0010`'s line count is unchanged, and the rubric's `depends_on` together with `decision-0100`'s `changes:` form the benign pair that rubric check 5 names.

---

## Verification Contract

| Gate | Command or check | Proves |
|---|---|---|
| Build, vet, full suite | `cd cli && go build ./... && go vet ./... && go test -count=1 ./...` | R3, R4, R10 (the plugin package and release tests stay green) |
| Conformance | `cd cli && go test -count=1 -run 'TestCorpusConform\|TestArtifactContract' .` | R3; red at U1 and green from U2 |
| Repository quality | `npm run quality` from the repository root; ShellCheck may be missing locally, and CI runs it | the repository's quality gates |
| Sweep | `git grep -n -e 'core/rubrics' -e 'core/fixtures'` | R5: only the Appendix's kept lines, this plan, and `decision-0100` if written |
| History | `git log --follow` on `docs/rubrics/artifact-contract.md` and on `cli/testdata/known-bad/known-bad.md` | R1 |
| Line positions | `git diff -M -U0 origin/main` over modified pre-existing files; new files are out of scope | R8: every hunk's old and new start line and length match, except `core/README.md`'s Contents removal (KTD4) |
| Payload | `git diff --name-only origin/main -- plugins/` is empty | R10 |
| Mutations | the two U2 mutations, run and reverted | R3: the relocated tests can still fail |

## Definition of Done

- U1–U6 are complete, and U7 is complete or dropped per `trellis-d1`'s answer.
- Every Verification Contract gate holds, and the PR's Validation section records the U1 red run and both mutation results.
- No mutation or experimental edit is left in the diff.
- The PR is open and ready for review, with the kept-line list and the assumptions `trellis-d1` did not override in its body. The maintainer merges.

---

## Appendix

Line numbers are from `main` at `a3ffa70`, before this change. U6 re-verifies them.

### Swept

| File | Lines |
|---|---|
| `.github/workflows/cli-ci.yml` | `19-22` (comment); `35`, `38` (entries removed, KTD5) |
| `AGENTS.md` | `102` |
| `README.md` | `283` |
| `cli/artifact_contract_guard_test.go` | `16`, `733` |
| `cli/corpus_conformance_test.go` | `4`, `14`, `49`, `54`, `55`, `58`, `1270` |
| `core/README.md` | `10`, `12` (removed), `14-21` (KTD7) |
| `cli/testdata/README.md` (was `core/fixtures/README.md`) | `1` |
| `docs/rubrics/artifact-contract.md` | `20-21`, `168` |
| `core/schemas/typed-artifacts.md` | `20` |
| `profiles/trellis-self.md` | `101`, `103`, `112` |
| `docs/decisions/0037-statuses-are-methodology-defined.md` | `79` |
| `docs/decisions/0038-retire-display-codes-slugs-only.md` | `49` |
| `docs/decisions/0044-cross-repo-depends-on-convention.md` | `180` |
| `docs/decisions/0066-retire-the-surface-matrix.md` | `354` |
| `docs/decisions/0074-deliberate-succession.md` | `123` |
| `docs/decisions/0076-retire-grove.md` | `61`, `94`, `136`, `203` |
| `docs/decisions/0078-no-orphan-followups.md` | `96` |
| `docs/decisions/0079-retire-the-spec-stage.md` | `43`, `64` |
| `docs/decisions/0081-supersession-authority-by-cost-of-reversal.md` | `29`, `132`, `372` |
| `docs/decisions/0082-retire-the-status-field.md` | `66`, `143`, `145` |
| `docs/decisions/0098-artifact-conformance-is-a-ci-check.md` | `22`, `59`, `176`, `178`, `179`, `183`, `198` |
| `docs/plans/2026-09-13-2321-feat-corpus-conformance-ci-check-plan.md` | `35`, `95`, `159`, `161`, `192`, `195`, `233`, `253`, `255`, `258`, `301`, `387`, `393` |
| `docs/plans/2026-09-13-2321-refactor-corpus-under-docs-plan.md` | `80`, `98`, `242` |
| `docs/superpowers/plans/2026-09-03-row-set-guard.md` | `297`, `320` |
| `docs/superpowers/specs/2026-09-03-row-set-guard-design.md` | `32` |

### Kept

| Class (KTD3) | Lines |
|---|---|
| Describes this move | `decision-0098:132-134`; this plan; `decision-0100` if written |
| Not a path | `core/invariants/trellis-invariants-v1.md:334`; `cli/codex_hook_test.go:1096`; `cli/testdata/known-bad/known-bad.md:2` (`fixture-known-bad`) |
| A different rubric | `agentic-dev-meta-layer-brief.md:321` |

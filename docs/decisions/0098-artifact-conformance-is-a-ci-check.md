---
id: decision-0098
type: decision
depends_on: [decision-0007, decision-0010, decision-0047, decision-0076, decision-0082, decision-0089, decision-0097, decision-0099, schema-typed-artifacts]  # coupling under decision-0047's test. 0010, 0076, 0089 and 0099 are the records superseded in part, and 0010's fourth Decision bullet is the authority for point 1. 0082 is why 5b is dropped, and 0047 is why 5d is judgment rather than code (point 2). schema-typed-artifacts fixes C2's three values, which point 7's framing rests on. 0007's review workflow and 0097's ce-code-review are the gatekeeper point 7 names, and ce-code-review's standards reviewer is why the guidance lives in AGENTS.md (point 2)
changes: [decision-0010, decision-0076, decision-0089, decision-0099, profile-trellis-self, rubric-artifact-contract]  # the rubric is here and not in depends_on, although points 2 and 6 read its check text. This change adds decision-0098 to the rubric's depends_on, and rubric check 5 calls "an artifact both depending on its authorizing decision and named in that decision's changes:" a benign pair, not a cycle. A depends_on edge back from this record would be the cycle
informed_by: [decision-0028, decision-0040, decision-0081, decision-0092]
owner: agent
date: 2026-09-14
---

> **Provenance.** TRL-88 records the maintainer's decision of 2026-09-13 to move `corpus-reviewer`'s
> checks into a deterministic script that runs in CI, and to retire the agent. On 2026-09-14 the
> maintainer settled the supersession scope, the check outcomes, the profile framing and the
> provisional consumer impact, and later that day ruled that the rubric leaves `core/`. Those rulings reached the author as relayed by the supervising
> session `trellis-d1`, and are recorded here as relayed, not quoted. The plan is
> `docs/plans/2026-09-13-2321-feat-corpus-conformance-ci-check-plan.md`.

# 0098 — Artifact conformance here is a CI check, and `corpus-reviewer` retires

> **Dated note, 2026-09-14 — TRL-96 (PR #312):** Four statements no longer hold: point 1's "that touches a corpus path, the rubric, the fixtures or the check itself", point 5's "does not block a merge" and "`build-test`, the job that runs the check, is not among them", and the Consequences' "`cli/ci_paths_guard_test.go` forces the filter to cover every path the suite reads" and "the `cli-ci` filter mirror it, and guards pin both". `cli-ci` has no path filter, that guard is deleted, and `build-test` runs on every pull request as a required check on `main`, as `.github/workflows/cli-ci.yml` and `AGENTS.md`'s conformance bullet state.

## Context

Before this change, `.claude/agents/corpus-reviewer.md` applied `docs/rubrics/artifact-contract.md`
to `docs/decisions/`, `docs/research/`, `core/` and `profiles/`. It was an LLM sub-agent, and `AGENTS.md` asked
authors to invoke it before merging, the rule `decision-0076` point 6 added. Three things made that
gate weak:

- **Nothing ran it.** It ran when an author remembered to. Until `decision-0099` moved the corpus
  under `docs/`, `cli-ci`'s paths filter on `main` selected neither `decisions/**` nor `research/**`,
  so a decision-only PR did not run the Go suite either.
- **Its report left no trace.** The agent reported in prose, to the session that invoked it. The PR
  carried no record of a run beyond what an author chose to write in a self-check.
- **Its reading drifted from the rubric.** The TRL-60 review found the charter and the rubric
  disagreeing in eight places, and `cli/artifact_contract_guard_test.go` was added to pin three of
  their shared enumerations. Rules stated only in prose stayed unpinned.

Three records state the agent-applied gate as current truth:

- `decision-0010`, Decision bullet 2: *"The artifact-contract "validator" is a **conformance
  sub-agent applying a rubric**, failing loudly (B3 / D1) — **not** a program."*
- `decision-0076`, point 6: *"invoke the repo-owned `corpus-reviewer` before merging a change to
  `docs/decisions/`, `specs/`, `docs/research/` or `core/`."*
- `decision-0089`, Consequences: *"Artifact **conformance** stays agent-applied via
  `corpus-reviewer`; this check is about id allocation, which is a merge-queue fact no rubric can
  see."*

`decision-0010`'s fourth bullet already allows the alternative: *"Any deterministic helper a project
wants for hard CI gating is written in **the target project's own stack** — never a runtime Trellis
imposes."* This repository's CI guards are Go tests in `cli/`, run by `cli-ci`.

## Decision

**1. A Go test applies the artifact contract to this repository's corpus, and `corpus-reviewer`
retires.**

- `cli/corpus_conformance_test.go` holds the corpus loader, checks 1–7 and the outcome table.
  `TestCorpusConformsToArtifactContract` runs every rule over the live corpus, one subtest per check,
  and reports each violation as `<path>:<line>: check <N> (<rule-id>): <detail>`.
  `TestCorpusConformanceRejectsKnownBadFixture` is the positive control: it runs the same rules over
  `cli/testdata/known-bad/` and requires exactly the expected findings, none missing and none extra.
- `cli/corpus_conformance_typed_test.go` holds checks 8–11, over the catalog and the profiles.
- A shared input that is missing or unparseable halts the run by name, never a partial pass.
- `cli/artifact_contract_guard_test.go` pins the check to the rubric. A numbered rubric check with no
  outcome row, or whose text changed without its rows being reviewed, fails it.
- The check runs in `cli-ci`'s `build-test` job, on every pull request and every push to `main` that
  touches a corpus path, the rubric, the fixtures or the check itself.

The authority is `decision-0010`'s fourth bullet: a deterministic helper for CI gating, written in
the project's own stack. `.claude/agents/corpus-reviewer.md` is deleted.

**2. Every part of every rubric check has a recorded outcome, and the table lives in the check.**
`contractOutcomeTable` and `contractClauseOutcomes` in `cli/corpus_conformance_test.go` carry one row
per rule, with its outcome and reason. This record does not restate them. Most rules are implemented,
accepted as a form the rubric admits, or covered by another rule or test. Check 6's research-note
habits have no rule, because the rubric says they are not gated, and check 12 stays retired. The
maintainer accepted every outcome on 2026-09-14.

Four parts are dropped:

- **2c, a type's scope and rubric "may be declared centrally".** No central registry exists; where
  to declare them is still the rubric's own open question. Before this record, 107 of the corpus's
  112 artifacts carried no per-file `scope`, so requiring a declaration would reject valid artifacts.
- **2d, the list of recognized typed artifacts and their scopes.** It describes and does not
  constrain. The typed checks key on `type`, and 2b validates any `scope` a file declares.
- **5b, the status-lifecycle form**, *"no gated/approved artifact `depends_on` a draft one"*. The
  rubric applies it only where a methodology declares a status lifecycle, and this repository
  declares none (`decision-0082`).
- **The charter's "Derive your checklist yourself".** That was an instruction to an agent not to
  take its checklist from the author of the artifacts. The checklist is now code pinned to the
  rubric, and CI runs it whoever authored the change.

Two parts are judgment, and stay as review guidance:

- **5d, a coupling relabeled as `informed_by`.** Whether an artifact's correctness rests on a source
  is the test `decision-0047` sets, and a parser cannot apply it.
- **10b, whether a row's evidence supports its claim.** The check confirms that `evidence` and
  `confidence` are present. Whether the pointer shows the tell takes reading it.

Both are written in `AGENTS.md`'s conformance bullet, each citing its rubric check.
`ce-code-review`'s project-standards reviewer takes its criteria from the standards files that
govern a changed file, root `AGENTS.md` among them, and *"Every finding you report must cite a
specific rule from a specific standards file"* (`references/personas/project-standards-reviewer.md`,
Compound Engineering 3.25.0). A rule written only in the rubric would give that reviewer nothing to
cite.

**3. Four records are superseded in part, each by a clause-scoped forward pointer** in
`superseded_in_part_by` (`decision-0040`). Only frontmatter changes; no body is edited.

- **`decision-0010`: Decision bullet 2, as it governs this repository's own gate.** Bullet 2 stands
  as a description of Trellis's product for a consumer project; whether moving the rubric out of
  `core/` (point 4) changes that is left to TRL-95. Bullet 4
  stands, and is this record's authority. Open question 2, how the conformance sub-agent is invoked
  in CI without a runtime, is answered for this repository only: it is not invoked, because a test in
  the repository's own stack replaced it. The question stays open for a consumer.
- **`decision-0076`: point 6's standing rule** to invoke `corpus-reviewer` before merging, with its
  ground that no CI check covered the contract. Point 6's account of why the rule was declared rather
  than smuggled stands.
- **`decision-0089`: the first clause of the Consequences sentence quoted above.** Its second clause
  stands: id allocation is a merge-queue fact, and `decision-id-guard` remains a separate check.
- **`decision-0099`: the one Consequences bullet saying the rubric's corpus paragraph and
  `corpus-reviewer`'s charter name the new paths together.** The charter is deleted, and
  `cli/artifact_contract_guard_test.go` now compares the paragraph with `corpusRoots`. Every Decision
  point of `decision-0099` and its other Consequences stand. This pointer was added after TRL-89
  merged, following the maintainer's choice of clause-scoped pointers for the three records above,
  as relayed by `trellis-d1`.

**4. Nothing Trellis ships changes, and the rubric leaves `core/`.** Nothing under
`plugins/trellis/` names the rubric, the fixtures or the agent, and neither does `install.sh`. The
one mention in the CLI's non-test sources, `cli/payload.go`, is a comment saying the payload leaves
artifact-contract metadata out. There is no payload change and no `plugins/trellis/VERSION` bump.
On 2026-09-14 the maintainer ruled that the rubric leaves `core/`, because the plugin does not need
it; the ruling reached the author as relayed by `trellis-d1`. The move is TRL-95, which waits for
this change: the rubric goes to `docs/rubrics/artifact-contract.md` with `scope: trellis-meta`, and
`core/fixtures/known-bad/` goes to `cli/testdata/known-bad/`. The typed-artifacts schema stays in
`core/`. Until TRL-95 lands, the rubric sits at `core/rubrics/` with its check text and its
consumer-facing agent wording unchanged. This change touches only its **Derived resource**
paragraph, a short note that this repository enforces it with a Go test, and its `depends_on`,
which gains this record.

**5. The check runs on every corpus PR but does not block a merge.** `main`'s ruleset,
`main protection`, requires `release-guard`, `Analyze (go)`, `Analyze (javascript)` and `hygiene`.
`build-test`, the job that runs the check, is not among them. A red check is a signal to the
maintainer, not a gate GitHub enforces. Whether to make it required is a repository setting, not
decided here.

**6. Two verdicts `corpus-reviewer` recorded as failures now pass, and the check has no exemption
for either.**

- **`docs/decisions/0044-cross-repo-depends-on-convention.md:5`** carries
  `kodhama/kodhama-0004-uniform-lifecycle`. Check 4 accepts a `<repo>/<id>` entry on shape and
  registry membership, and `kodhama` is in the rubric's registry list. The FAIL repeated in
  `decision-0092`'s self-check, after `decision-0076`, `decision-0088` and `decision-0090`, was
  against the earlier bare `kodhama-0004-uniform-lifecycle`, which matched no accepted form. That
  verdict was right when it was given. TRL-51 (#282) qualified the entry on 2026-09-06, the day after
  `decision-0092` merged.
- **`docs/research/0010-agent-instruction-file-landscape.md`** carries `informed_by: [decision-0029]`. The
  plan `docs/superpowers/plans/2026-08-30-codex-reconciliation-parity.md` listed `docs/research/0010:5` as
  a known pre-existing violation. On that date line 5 read `depends_on: [decision-0029]`. The plan
  does not say which check failed, but `decision-0029` exists, so it was not a dangling reference.
  The likelier reading is the edge's kind, which TRL-57 (#285) named in its title on 2026-09-06:
  *"research-0010's edge to decision-0029 is provenance, not a dependency"*. The entry now resolves
  under check 4. Whether that relabel was right is the 5d judgment point 2 leaves to review.

**7. In `profiles/trellis-self.md`, the gatekeeper is a reviewer, and the check is supporting
evidence.** Where a row's C2 stays `independent-agent`, the gatekeeper named is the Claude Code Review
workflow (`decision-0007`) together with `ce-code-review` (`decision-0097`). The deterministic check
is cited as supporting evidence, never as the gatekeeper, because C2 admits only
`independent-agent`, `human` or `none` (`schema-typed-artifacts`), and a test is none of the three.
The rows that cited `corpus-reviewer` keep `confidence: verified` on re-argued evidence, as the
maintainer ruled.

## Consequences

- **Changed in the same PR (`decision-0028`):**
  - `AGENTS.md`'s conformance bullet: the CI check replaces the invoke-before-merge rule, and the
    bullet states 5d and 10b.
  - `docs/rubrics/artifact-contract.md`: the **Derived resource** paragraph, the repository note and
    `depends_on`.
  - `core/README.md`, and `cli/testdata/README.md`. `cli/testdata/known-bad.md` moves into
    `cli/testdata/known-bad/`.
  - `profiles/trellis-self.md`: the rows citing `corpus-reviewer` as evidence, repaired under the
    profile's own repair convention.
  - `.github/workflows/cli-ci.yml`: the paths filter gains `core/schemas/**`, `core/lexicon.md` and
    `cli/testdata/**`, and its comment names the test. `docs/**`, which `decision-0099` added,
    already selects the corpus.
  - The comments in `.github/workflows/repo-hygiene.yml` and `.github/workflows/decision-id-guard.yml`.
  - `cli/selfapply_test.go` drops the charter from its bounded references.
  - `eval/experiments/annotation-vs-absence/README.md` keeps its historical "corpus-reviewer PASS"
    and gains a pointer to this record.
  - `.claude/agents/corpus-reviewer.md` is deleted.
- **Historical mentions of `corpus-reviewer` stay as written**: the append-only decisions,
  `docs/research/0012`'s `status:` comment, the plans under `docs/superpowers/`, and the dated repair
  notes in `profiles/trellis-self.md`.
- **A record-only PR runs the whole `build-test` job**, as it has since `decision-0099`: `npm ci`,
  ShellCheck, `npm run quality` with staticcheck, then Go build, vet, test and the coverage floor. The
  check runs inside that Go suite, where `cli/ci_paths_guard_test.go` forces the filter to cover every
  path the suite reads.
- **The new filter entries overlap TRL-92**, the broader gap: `cli/docs_consistency_test.go` walks
  files the `cli-ci` filter does not select. TRL-92 lists `core/schemas/` and `cli/testdata/` among
  them, and this change selects both because the conformance test reads them. TRL-92's other paths,
  and its root fix, stay open there.
- **TRL-89's move of the corpus under `docs/` (`decision-0099`) was a one-line change in the check.**
  The corpus is named once, in `corpusRoots`, which now reads `docs/decisions/` and `docs/research/`.
  The rubric's **Corpus:** paragraph and the `cli-ci` filter mirror it, and guards pin both.
- **`decision-0089`'s "`main` carries no branch protection here" is out of date**, and so was the
  same line in `decision-id-guard.yml`, which this change corrects. `main` now has a ruleset. Both
  conclusions still hold, because `decision-id-guard` is not a required check. This record names the
  sentence and does not supersede it.
- **`decision-0099`'s consequence that the rubric's corpus paragraph and `corpus-reviewer`'s charter
  "name the new paths together" no longer holds,** and this record supersedes that bullet in part
  (point 3). The charter is deleted, and `cli/artifact_contract_guard_test.go` now compares the
  paragraph with `corpusRoots`.
- **Forward pointers:** `decision-0010`, `decision-0076`, `decision-0089` and `decision-0099` each gain
  `decision-0098` in `superseded_in_part_by`.
- **Not decided here:** whether `build-test` becomes a required check, TRL-92's root fix, and what
  TRL-95 owes the records when it moves the rubric out of `core/` (point 4).

## Open questions

- None open. The question this record carried, whether the artifact-contract rubric leaves `core/`,
  was answered by the maintainer on 2026-09-14 in review of this change: it does (point 4), and
  TRL-95 moves it. The check reads the rubric through one path constant, `artifactContractPath`, so
  that move is a one-line change there.

## Self-check

- **The author is an agent**: the U3 implementation worker under session `trl-88`, running Claude
  Opus 5. Independent review follows: `ce-code-review` on the branch, and `decision-0007`'s automated
  review on the PR. This record does not rule itself correct.
- **What the author checked, and how:**
  - Every frontmatter reference resolves: each id was found as an `id:` line in `docs/decisions/`,
    `core/` or `profiles/`.
  - The four pointer edits touch frontmatter only, confirmed with `git diff`. `decision-0089`'s
    sentence was copied from the file. `decision-0076` gains a frontmatter line, which shifts its
    body by one line; a search found no citation of it by line number.
  - The ruleset's required checks were read from the GitHub API on 2026-09-14, and `main`'s paths
    filter with `git show origin/main:.github/workflows/cli-ci.yml`.
  - The 107-of-112 count comes from a script over the corpus roots, counting the files that open
    with frontmatter and those that declare `scope:`.
  - Point 4's "nothing ships it" rests on a search of `plugins/trellis/`, `install.sh` and the CLI's
    non-test Go sources for the rubric, fixture and agent paths.
  - Point 6's history comes from `git log -p` on both files.
  - The id was unclaimed: on 2026-09-14 no pull request was open, and no ref carried a
    `decisions/0098-*` file. Re-checked after TRL-89 merged, with `main` at `2581691`: `main` holds
    `docs/decisions/0097-*` and `docs/decisions/0099-*` and no `0098`, and #309 was the only open
    pull request.
- **Checked by the orchestrating session, not the author:** the Go suite, including the live
  conformance run that reads this record, passed on the integrated tree (`go test -count=1 ./...`),
  and point 1's description of the guard was read against the rewritten
  `cli/artifact_contract_guard_test.go`. CI runs the suite again on the pull request.
- **After TRL-89 merged,** this branch merged `origin/main` without a rebase, took this record to
  `docs/decisions/`, pointed `corpusRoots` at `docs/decisions/` and `docs/research/`, and swept this
  record's own path citations (`decision-0015:117-121`). Two mentions keep the old path because
  they describe a past state: the pre-move filter in Context and the id check of 2026-09-14.
- **Point 6 corrects the plan it came from.** The plan framed both verdicts as the agent's reading
  drifting from the rubric. Git history does not support that for either entry: both were changed
  at source on 2026-09-06, and every recorded verdict on them predates the change. The drift evidence
  that stands is TRL-60's eight disagreements.
- **The rubric is in `changes`, not `depends_on`, and that is a judgment call.** Points 2 and 6 read
  its check text, so the coupling is real. The rubric also gains `decision-0098` in its own
  `depends_on` in this change, and rubric check 5 names that shape, not a two-way `depends_on`, as the
  benign pair. A reviewer who reads the back-edge as required should flag it.
- **Supersession authority.** This record marks four records superseded in part, and its author is
  an agent. Under `decision-0092`'s reading of `decision-0081`, merging it is the maintainer's act of
  supersession, not the author's.

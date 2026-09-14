---
id: decision-0100
type: decision
depends_on: [decision-0005, decision-0010, decision-0097, decision-0098, decision-0099]  # coupling under decision-0047's test. 0098 point 4 is the ruling this record carries out, and 0010 is the record it supersedes in part. 0005 is the layer split point 1 places the rubric under. 0097 point 4 and 0099 are the docs/ exemption and the docs/ location point 4 rests on
changes: [decision-0010, invariants-v1, rubric-artifact-contract]
informed_by: [decision-0015, decision-0028, decision-0040, decision-0082]
owner: agent
date: 2026-09-14
---

> **Provenance.** TRL-95 carries out the ruling `decision-0098` point 4 records: on 2026-09-14 the
> maintainer ruled that the artifact-contract rubric leaves `core/`, because the plugin does not
> need it. `decision-0098` left "what TRL-95 owes the records" undecided. Later that day the
> maintainer answered: a record is owed, it supersedes `decision-0010` in the two clauses point 2
> names, and `invariants-v1` drops the rubric's id. `trellis-d1` settled the docs-consistency walk
> under `decision-0097` point 4. Those answers reached the author as relayed by `trellis-d1`, and are
> recorded here as relayed, not quoted. The plan is
> `docs/plans/2026-09-14-0930-refactor-artifact-contract-out-of-core-plan.md`.

# 0100 — The artifact contract is repository-internal

## Context

TRL-95 moved the rubric from `core/rubrics/artifact-contract.md` to `docs/rubrics/artifact-contract.md`
and its known-bad fixtures from `core/fixtures/` to `cli/testdata/`. Nothing Trellis ships reads
either: nothing under `plugins/` or in `install.sh` names them (`decision-0098` point 4).

Four texts still called the contract part of the product:

- `decision-0010`, Decision bullet 1, lists Trellis's resources "(including the artifact contract and
  its conformance check)".
- `decision-0010`, Decision bullet 2: *"The artifact-contract "validator" is a **conformance sub-agent
  applying a rubric**, failing loudly (B3 / D1) — **not** a program."* `decision-0098` superseded it
  for this repository's gate and left it standing for the product "provisionally, while
  decision-0098's open question on whether the rubric leaves core/ is unanswered".
- `core/invariants/trellis-invariants-v1.md:147-148` named `rubric-artifact-contract` as the deferred
  conformance-to-upstream check.
- `README.md`'s repo map described `core/` as "The shippable product: invariants, the conformance
  rubric, the signature catalog, the lexicon."

## Decision

**1. The artifact contract is repository-internal.** The rubric declares `scope: trellis-meta` and
lives beside the governance corpus `decision-0099` placed under `docs/`. Its positive control lives
in `cli/testdata/known-bad/`, where only the Go control test reads it. `decision-0005` stands: it
draws the line between the product and the build methodology, and this record places one artifact on
the methodology side of that line. The typed-artifacts schema stays in `core/`, because it defines the
catalog and the profile, which are product types.

**2. `decision-0010` is superseded in part, in two clauses.** The clauses are bullet 1's parenthetical
"(including the artifact contract and its conformance check)", and bullet 2 as it describes the
product. With `decision-0098`'s clause for this repository, bullet 2 is superseded whole. The rest of
bullet 1, bullets 3 and 4, and Open question 1 stand. `decision-0010` gains `decision-0100` in
`superseded_in_part_by` with its own clause. That clause also records that this record settles the
provisional wording in `decision-0098`'s clause, which stays as written.

**3. `invariants-v1` stops naming the repository-internal id.** The forward re-propagation example at
`core/invariants/trellis-invariants-v1.md:147-148` read *"(Full content-consistency enforcement is the
deferred conformance-to-upstream check, `rubric-artifact-contract`.)"* It now ends at "check". The
operation and its deferral are unchanged. Only the id goes, because a product artifact should not
name a repository-internal one. This record is the authority for that edit to the ratified set.
Nothing under `plugins/`, in `install.sh` or in the CLI's non-test sources reads the file, so the
edit ships nothing.

**4. The docs-consistency walk's `docs/` exemption covers `docs/rubrics/`.** `decision-0097` point 4
rests the exemption on location: the walk guards documents that describe the product to its users,
and `docs/` holds none. The rubric is repository-internal, so `TestDocsClaimOnlyRealCommands` and
`TestNoUnqualifiedSetupClaims` no longer reading it follows that rule. The comment at
`cli/docs_consistency_test.go:63-71` names `docs/rubrics/` with the governance corpus.

## Consequences

- **Changed in the same PR (`decision-0028`):**
  - The rubric and the fixture directory move with `git mv`.
  - `cli/corpus_conformance_test.go` and `cli/artifact_contract_guard_test.go` read the new paths.
  - `cli-ci`'s two filter lists drop `core/rubrics/**` and `core/fixtures/**`; `docs/**` and `cli/**`
    already cover the new homes.
  - The rubric's corpus paragraph, `scope`, `depends_on`, header, repository note, "How it is graded"
    and acceptance criterion change.
  - `AGENTS.md`, `README.md`, `core/README.md`, `core/schemas/typed-artifacts.md`,
    `cli/testdata/README.md` and `profiles/trellis-self.md` take the new paths, and `README.md`'s repo
    map stops listing the rubric under `core/`.
  - Eleven records and four planning records take the new path token, per `decision-0015:117-121`.
    `decision-0098:133-134`, which describe this move, keep the old paths.
  - `decision-0010`'s forward pointer, `core/invariants/trellis-invariants-v1.md:148` and the comment in
    `cli/docs_consistency_test.go` change as points 2 to 4 state.
- **No payload change**, so `plugins/trellis/VERSION` does not move.
- **`decision-0010`'s Open question 2 has no product subject.** Trellis ships no artifact-contract
  validator to invoke in CI. `decision-0098` answered the question for this repository, and this
  record does not supersede it.
- **No test fails if `core/rubrics/` or `core/fixtures/` reappears.** Nothing reads those paths and
  `cli-ci` no longer selects them, so a stale branch that recreates one lands unchecked. That was
  accepted, not guarded.

## Open questions

- None open.

## Self-check

- **The author is an agent**: the `trl-95` worker session, running Claude Opus 5. Independent review
  follows: `ce-code-review` on the branch with a Codex cross-model pass, and `decision-0007`'s
  automated review on the PR. This record does not rule itself correct.
- **What the author checked, and how:**
  - `decision-0100` was free: on 2026-09-14 `main` (`a3ffa70`) ended at `decision-0099`, and no pull
    request was open.
  - Every frontmatter reference resolves to an `id:` line in `docs/decisions/`, `docs/rubrics/` or
    `core/`.
  - `decision-0010`'s edit stays on its one frontmatter line, and the `invariants-v1` edit rewrites line
    148 in place, so no cited line moves.
  - Point 3's "ships nothing" rests on a search for `trellis-invariants-v1` across `cli/*.go`,
    `plugins/`, `install.sh`, `scripts/` and `.github/`: only `cli/row_set_guard_test.go` reads the file.
  - `cd cli && go test -count=1 ./...` passes, the conformance tests and `cli/row_set_guard_test.go`
    among them.

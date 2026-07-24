---
id: spec-0006
type: spec
status: gated
superseded_in_part_by: [decision-0058]
depends_on: [decision-0057]
implements: decision-0057
owner: agent
rubric: rubric-artifact-contract
date: 2026-07-23
---

# Spec 0006 — Shared project-instruction entrypoints for Trellis itself

## Purpose

Give Trellis one authoritative home for project instructions shared by Claude
Code and Codex, while preserving the existing Claude-specific Trellis managed
import block and its live `.trellis/rules.toml` behavior.

## Scope

**In scope:** this repository's root `AGENTS.md` and `CLAUDE.md`, current-truth
references to the shared instruction home, the self-application sync guard, and
the CI path filter that runs that guard.

**Out of scope:** consumer-wide setup/refresh/remove migration; plugin-native
delivery of the Trellis overlay to Codex; `.codex/agents`; moving, rewriting,
or removing the existing Trellis managed block.

## Canonical state

### 1. `AGENTS.md` owns shared project instructions

Root `AGENTS.md` contains the Layer-B operating method and Grove managed block
that previously lived in `CLAUDE.md`. It contains no
`<!-- trellis:begin`/`<!-- trellis:end -->` block and no
`@.trellis/...` imports.

Before the managed blocks, it contains a `## Maintaining project instructions`
section that tells every host:

- `AGENTS.md` is the canonical home for shared project instructions;
- new shared rules are edited there, outside managed blocks;
- `CLAUDE.md` is the Claude adapter and is not a shared-rule edit surface;
- genuinely Claude-only rules belong in `.claude/rules/`;
- Grove and Trellis project choices remain in their named
  `.grove/`/`.trellis/` configuration files; and
- managed blocks are not hand-edited.

The wording may be concise, but every routing statement above is required and
is asserted by the sync guard.

### 2. `CLAUDE.md` adapts shared prose and retains Trellis delivery

Root `CLAUDE.md` contains, in order:

1. one standalone `@AGENTS.md` line;
2. one blank separator line; and
3. the existing Trellis managed block, byte-identical to
   `plugins/trellis/reference/block-claude.md`, followed by one newline.

It contains no Layer-B or Grove prose outside that adapter and Trellis block.
The Trellis block remains import style and continues to load both
`.trellis/internal/trellis.md` and `.trellis/rules.toml`; this change does not
alter its bytes or live-row semantics.

### 3. Current-truth references follow the authority move

The bounded current-truth set for this change is exactly the following files;
references in them that tell agents where the repo's method, conventions, open
questions, or Grove stamp live point to `AGENTS.md`:

- root `README.md`;
- `profiles/trellis-self.md`;
- `.grove/config.toml`;
- `.grove/README.md`; and
- `.claude/agents/corpus-reviewer.md`.

Historical records retain their original prose and receive the forward
pointers required by `decision-0057`.

References specifically describing Claude's Trellis import adapter continue to
name `CLAUDE.md`.

### 4. Drift guard and CI trigger

`cli/selfapply_test.go` verifies:

- the exact `@AGENTS.md` adapter occurrence and placement;
- the required editor-routing statements in `AGENTS.md`;
- zero Trellis markers/imports in `AGENTS.md`;
- exactly one matched Trellis block in `CLAUDE.md`;
- byte equality of that block with `block-claude.md`; and
- absence of duplicated Layer-B/Grove prose in `CLAUDE.md`.

The test keeps its upstream comment (`spec-0006`) as the de facto test ledger.

`.github/workflows/cli-ci.yml` includes `AGENTS.md`, `CLAUDE.md`, and
`.trellis/**` in both pull-request and main-push path filters so an
instruction-only drift can run the self-application guard.

## Acceptance criteria

- **AC1 — one shared authority.** The full Layer-B method and Grove block exist
  in `AGENTS.md` and not in `CLAUDE.md`.
- **AC2 — explicit edit routing.** Claude, after following `@AGENTS.md`, is
  explicitly told to edit shared rules in `AGENTS.md`; all six routing
  statements in §1 are present.
- **AC3 — Trellis block preserved.** The pre-change and post-change managed
  Trellis blocks in `CLAUDE.md` are byte-identical, and the post-change block
  matches `block-claude.md`.
- **AC4 — exact adapter shape.** `CLAUDE.md` is exactly the adapter, separator,
  Trellis block, and final newline described in §2.
- **AC5 — no false Codex import claim.** `AGENTS.md` contains no Trellis block
  or `@.trellis/...` imports, and no changed current-truth prose claims this PR
  delivers the Trellis overlay to Codex.
- **AC6 — current truth propagated.** Every bounded current-truth surface in
  §3 names `AGENTS.md` for shared conventions while preserving
  Claude-adapter-specific references.
- **AC7 — historical integrity.** Decisions 0005, 0028, and 0035 carry the
  exact partial-supersession pointers required by decision-0057.
- **AC8 — deterministic guard.** Each canonical-state obligation in §4 has an
  automated assertion, and instruction/overlay-only PRs trigger the workflow.
- **AC9 — no scope creep.** No consumer setup, refresh, or remove skill;
  plugin manifest; generated payload; or `.codex/agents` file changes.
- **AC10 — existing gates remain green.** `go test ./...`, `go build ./...`,
  `go vet ./...`, and `git diff --check` exit zero in their applicable
  directories, and `gofmt -l selfapply_test.go` under `cli/` produces empty
  output. Repo-wide `gofmt -l .` is not a clean baseline (`apply.go` and
  `sync_test.go` predate this change); this PR does not silently absorb that
  unrelated stock.

## Test coverage

| Test | Contract |
|---|---|
| Self-application entrypoint test | AC1–AC5, AC8 |
| Bounded reference grep/assertions in the self-application test | AC6 |
| Artifact/corpus review | AC7 |
| Staged path inventory | AC9 |
| Existing CLI gate suite | AC10 |

## Open questions

None.

## Rubric check

Self-checked against `core/rubrics/artifact-contract.md`: required frontmatter
and sections are present; the one approved fidelity upstream is declared;
scope and exclusions are explicit; every acceptance criterion is observable;
and the implementation/test surfaces are named without prescribing unrelated
consumer behavior. Independent spec-adversary review remains required.

---

> **Current-truth pointer (2026-07-24).** `decision-0058` supersedes AC5/AC9's
> zero-Trellis-block assertion only enough to permit a separately marked,
> generated Codex receipt/fallback in `AGENTS.md`. The full rule payload and
> Claude import block remain prohibited there; all other acceptance results
> and the shared-entrypoint implementation stand.

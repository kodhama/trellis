---
id: decision-0099
type: decision
depends_on: [decision-0015, decision-0092, decision-0097]
changes: [decision-0092]
informed_by: [decision-0005, decision-0028, decision-0040, decision-0076, decision-0082, decision-0089]
owner: agent
date: 2026-09-14
---

# 0099 — The governance corpus lives under `docs/`

## Context

On 2026-09-13 the maintainer decided that `decisions/` and `research/` move inside `docs/`, the artifact root Compound Engineering uses (`decision-0097`). TRL-89 records the decision and its scope:

> The maintainer decided on 2026-09-13 that these folders move inside `docs/`, where Compound Engineering keeps its artifacts.

and, of `eval/`, *"It stays at the repo root, because `docs/` shouldn't hold runnable code."*

The move also closes a CI gap. `marketplaceCommands` in `cli/docs_consistency_test.go` walks the whole tree, but `cli-ci`'s path filter never selected `decisions/`, `research/` or `eval/`, so a PR touching only those skipped `build-test`.

Two records name the old location as more than a citation:

- `decision-0092` point 1 defines a decision-id claim as a file at a `decisions/NNNN-*.md` path, and `.github/scripts/decision-id-guard.sh` enforces that shape. Moving the folder changes what the guard matches.
- `decision-0005` split Trellis-core from the build methodology, and its Consequences said to "keep build methodology at root". `AGENTS.md` places that methodology at "the repo root: this file, `decisions/` and `research/`".

The maintainer answered four questions on 2026-09-13, relayed by the supervising session for TRL-89: the guard reads `docs/decisions/` only; a root tripwire is added; the guard's rule clauses in `decision-0089` and `decision-0092` are swept like any citation, and `decision-0089` is not superseded; `decision-0005` is not superseded. Because the guard cannot see `main`'s root ids during the move, the move PR carries a one-off check that this record's id is free.

## Decision

**1. The corpus lives at `docs/decisions/` and `docs/research/`.** Both folders moved with `git mv`, so each file's history follows it. `eval/` stays at the repository root.

**2. A decision-id claim is a file a pull request puts at a `docs/decisions/NNNN-*.md` path.** Everything else `decision-0092` states about claims stands at the new path: `added` and `copied` claim, a rename claims its destination id when it differs from its source, a pull request can collide with itself, a rename does not release its source id, and exit 2 means the guard could not run. A file at the root `decisions/NNNN-*.md` path claims nothing, and the guard reads the base branch's ids from `docs/decisions/` only.

Concretely: `docs/decisions/0100-x.md` added claims `0100`; `decisions/0100-x.md` added claims nothing.

**3. A root `decisions/` or `research/` directory fails the Go suite.** A test asserts both are absent, the same shape `decision-0076` gave `.grove/`. `cli-ci`'s filters keep `decisions/**` and `research/**`, so the pull request that recreates either directory runs that test. This, not the guard, is what catches a record that a branch cut before the move adds at the old path, or a merge that resolves git's directory-rename conflict toward it.

**4. The move's own pull request is checked by hand for `decision-0099`.** On that pull request the base branch still holds its records at the root, so the guard reads an empty `docs/decisions/` and reports every moved id free. Its green run is not evidence. The pull request's Validation shows a one-off check that `decision-0099` is taken by no record on `main`, under either prefix, and by no other open pull request. The check does not recur: once this merges, the base branch holds `docs/decisions/` and the guard sees it.

**5. `cli-ci` selects `docs/**`, `eval/**`, `decisions/**` and `research/**`.** The first covers the moved corpus, the second closes the `eval/` gap, and the last two exist for point 3.

**6. `decision-0092` is superseded in part.** Point 1's claim path and point 4's exit-2 condition for a record path containing a space now read `docs/decisions/`, and a root path no longer claims. The rename sweep (`decision-0015:117-121`) rewrites both records' text to the new path like any other citation, so the forward pointer records that the guard changed rather than marking text as wrong. `decision-0089` is not superseded: its surviving clauses are true once swept.

**7. `decision-0005` stands.** Its Decision puts Trellis-core in its own namespace, `core/`. "Keep build methodology at root" appears only in its Consequences, describing a reorganization as opposed to `core/`. `docs/decisions/` and `docs/research/` are outside `core/`, so the build methodology still does not leak into the product. `AGENTS.md` takes the new path and nothing else.

**8. What the rename sweep kept.** Every mention of the two folders follows the move, in every artifact class, append-only records and planning records included (`decision-0015:117-121`, `decision-0097` §5). A mention stays as it was when it is a path in another repository, text that is not a path, text pinned to a commit or a named past run, text that describes this move, or a root path the tripwire and `cli-ci`'s filters name on purpose. The pull request lists every kept line.

## Consequences

- **Record-only and `eval/`-only pull requests run `build-test`.** That costs CI time on corpus changes and removes the gap `marketplaceCommands` had.
- **`docSurfacesIn` no longer skips directories named `decisions` or `research`.** The top-level `docs/` path skip from `decision-0097` §4 covers the corpus, so the name skips would have skipped nothing.
- **The plugin payload named the folders once, in a comment in `hooks/codex-context.mjs`.** Sweeping it is a release (`decision-0028`), shipped in the same pull request.
- **The artifact contract's corpus paragraph and `corpus-reviewer`'s charter name the new paths together,** because `cli/artifact_contract_guard_test.go` compares them.
- **Line citations survive the move.** Every edit to a file a record cites by line was a same-line substitution, and swept Markdown was not re-wrapped. `decision-0092` gains one frontmatter line, so its body moves down by one; no `decision-0092:NN` citation existed on `main` when this was written.
- **Links from outside the repository break.** GitHub does not redirect moved paths, so a link to `blob/main/decisions/…` stops resolving. No such link was found in other kodhama repositories; links from Linear were not checked.
- **Not decided here:**
  - the other text files `marketplaceCommands` reads that `cli-ci`'s filter does not select, which the maintainer directed to a separate issue, TRL-92;
  - retiring `corpus-reviewer`, which is TRL-88's.

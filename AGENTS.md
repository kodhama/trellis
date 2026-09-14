# Trellis — operating method (seed)

> **What Trellis is.** A shippable, portable pack that supervises an *agentic software-
> development process* — it fits, teaches, adapts, and guards whatever methodology a
> project uses, while enforcing a small set of invariants. It is **not** a process; it is
> the layer above the steps. See `agentic-dev-meta-layer-brief.md` (in this repo) for the
> full thesis and `core/invariants/trellis-invariants-v1.md` for the load-bearing core (ratified).

> **We build Trellis with Trellis.** This file is a deliberately tiny instance of the seed
> operating method (brief §12). It dogfoods our own invariants from commit one. Friction
> we hit while following it *is product research* — record it, don't route around it.

> **Which layer is this? (`decision-0005`.)** **Trellis-core** — the shippable product —
> lives in `core/`. **The methodology used to build Trellis** is the repo root: this file,
> `docs/decisions/` and `docs/research/`. *This file is Layer B — instance #1, the first methodology
> Trellis supervises — not Trellis's product agent-instructions.*

## The iron rule (most important design constraint)

Trellis must **always ground out in concrete, project-specific artifacts** — a real
instructions file, real gates, real sub-agents. If it ever just *describes* process
instead of *producing and enforcing* it, it has failed. Same rule applies to this repo:
prefer producing a checkable artifact over writing about one. **And it applies to our own
rules:** every invariant or abstract instruction carries ≥1 concrete example (few-shot) — *a
rule you can't exemplify is probably vaporware.*

## Operating method

The method lives in `docs/decisions/` and in the rules below. **A record is current truth, except for
a point a dated note on it says no longer holds; for that point, the place the note names is current
truth.** Read a record and its notes before relying on a rule.

**A change gets a new decision record only when a wrong call would be expensive to undo, slow to
notice, or the damage while it stands would be serious.** Any one of the three is enough. When you
cannot tell, ask the maintainer rather than writing a record or skipping one. *Examples.* The four
strategic forks, `decision-0001`–`0004`, get records. So does a plugin change where a wrong call
could silently stop rules reaching consumers' sessions, such as how the session-start hook reconciles
`.trellis/rules.toml` (`decision-0083`). An error-message, reporting, diagnostic or cosmetic plugin
change gets none: a vendored delivery naming its own dead pointer (`decision-0095`) would get no record
today, and nor would a move such as `decision-0099` putting the corpus under `docs/`. Both were
recorded before this test existed.

**A new record is named `docs/decisions/YYYY-MM-DD-<slug>.md`, and its `id` is `decision-` followed
by that filename's stem.** The date is the day the record is written, and the slug is lowercase words
and digits joined by single hyphens, as the numbered records' slugs are. *Example:*
`docs/decisions/2026-09-20-rules-toml-is-optional.md` has `id: decision-2026-09-20-rules-toml-is-optional`.
Records `0001` to `0100` keep their numbers. No id is reserved and no check guards the name: two pull
requests adding the same path conflict at merge, so git stops the second, and a guard was needed only
while numbered ids let differently named files claim one number.

**A change without a record keeps its reasoning in its requirements doc, its plan under
`docs/plans/` and its PR body.** The PR body lists the records the change put dated notes on, or says
it found none.

**`docs/decisions/` stays append-only: a record's decisions are never re-made or rewritten.** The only
text added to a record is a dated note, or `superseded_by` when a successor record retires it in full. The only
edits to its existing text are correcting a line citation or a note's PR number, and a rename sweep
(`decision-0015`). When a merged change makes a point or sentence of a record untrue, the same PR adds
one note directly under the record's title, below any notes already there:

```markdown
> **Dated note, YYYY-MM-DD — TRL-<n> (PR #<n>):** <what is no longer true, naming the point or sentence>. <where the current behaviour is stated now>.
```

A note cites its own PR's number. A note written before the PR opens uses the number the repository
will assign next, and the same PR corrects it if another PR takes that number first.

The place a note names is a file on `main` kept current in place (this file, a rubric, the code, test
or workflow that carries the behaviour, or a newer record), never a plan or a PR body; when no such
file states the new behaviour, the same PR states it here. The PR that adds a note also corrects every
live line citation into that record, whether in another record, a code comment, a test or this file.
A citation is live unless it sits in a plan or in a dated inventory: a passage that reports what was
true on a stated date, such as a sweep table, a self-check count, a dated note, or a record that dates
its own citations, as `decision-0081` does.

**There is no `status` field** (`decision-0082`). **Merging to `main` is the acceptance:** an
artifact on `main` is current truth and may be consumed; one not yet merged may not. Nothing to
flip, ever. **A record retired in full carries `superseded_by` when a successor record retires it and a
dated note when none does; a record retired in part gets a dated note**, whether a newer record or another
change retires it. No change adds a new
`superseded_in_part_by`; the entries already in records stay and still resolve. *Artifacts predating `decision-0082` keep
their `status:` lines as history — several carry the maintainer's intent act in a trailing
comment. Read them as accepted; do not add the field to anything new, and do not strip it from
anything old.*

**Never tell the maintainer a change is blocked on an artifact's recorded state.** If he has
asked for a PR, open it. An agent still may not merge on his behalf without his act
(`floor-intent-gate`) — the gate did not move; only the bookkeeping around it went away.

| Before you… | Read |
|---|---|
| write or change an artifact — frontmatter, per-type body sections | `decision-0082` (no `status`; the merge is the acceptance) · `decision-0042` (family lifecycle) · `decision-0037` (`owner: agent` carries *authorship*, not accountability — that stays with the maintainer) |
| supersede a record, in full or in part | the rules above — `superseded_by` for a full retirement by a successor record, a dated note for a partial retirement or for a full one with no successor |
| retire something, or draw a boundary with what came before | `decision-0081` (supersession authority scales with cost of reversal) · `decision-0074` |
| change a source that has derivatives — the catalog, the CLI's command set | `decision-0028` (update derivatives in the same change; a guard per pair) |
| record a significant choice | only when it meets the record test above, named by date and slug as the naming rule above says — the four strategic forks are `0001–0004` |
| plan a build between a decision and the code | the **Compound Engineering** skills (`ce-brainstorm` or `ce-plan`, then `ce-work`, `ce-code-review`, `ce-commit-push-pr`) — `decision-0097`; the spec stage retired in `decision-0079`, and `specs/` with it |
| implement or debug in an area a past fix touched | `docs/solutions/` — documented solutions to past problems (bugs, best practices, workflow patterns), organized by category with YAML frontmatter (`module`, `tags`, `problem_type`) |
| record a next step | `decision-0078` — name the consumer that will re-present it, or drop it |
| pick up or file work | `decision-0075` — see *Where work lives* below |

Beyond the records: **one logical change per PR**; descriptive, linear history; diffs small
enough to review on a phone. When friction reveals a missing rule, add it *where it fires*,
**prefer retiring to adding**, and keep it subordinate to the work (`inv-self-improvement`).

After a solved, verified problem, automatically invoke the `ce-compound` skill with `mode:non-interactive` at the completion checkpoint only when the work produced durable project reasoning that is not readily recoverable from the final code, tests, types, comments, or existing documentation, and losing it would plausibly cause recurrence, material risk, or substantial rediscovery. Apply this counterfactual: if the learning document disappeared, would a future engineer reading the final implementation still be likely to repeat the mistake or redo substantial investigation? If not, do not invoke it. Completion, effort, and diff size alone are not enough. Capture at the checkpoint so a qualifying learning can ship in the PR that produced it, and only where the repository treats captured learnings as tracked, committed knowledge.

Write every report, summary, or handoff to the user through the `ce-noslop` skill. This applies when you are the top-level agent writing to the user, not when you are a subagent reporting to its caller. Do not apply it to code, config, verbatim quotes, or text the user asked to post as written.

The invariants themselves — transparency, independent judgment, the rest — are delivered live by
the Trellis plugin at session start, not hand-written here (`decision-0035`, `decision-0071`). A
behavior that reads like a bare invariant belongs in the catalog, not in this file.

## Naming guardrail (research discipline, applied to ourselves)

If we ever name the invariant set authoritatively, attribute it clearly as **our
synthesis** — never imply pre-existing provenance. For now it is exactly *"Trellis's
invariants — our synthesis, v1."* Eponymous framing is a deliberate *later* decision, made
only once the set's durability is proven across multiple instances.

## Where work lives

**This project is managed in Linear** — Kodhama workspace, team **Trellis** (`TRL-*`). Issues,
stages, priorities and their history live there; this repository hosts the code, pull requests
and CI (`decision-0075`).

- **Ideas are a document, not issues** — the team's Ideas doc, each entry carrying the trigger
  that would promote it. An idea filed as an issue is a to-do nobody agreed to.
- **Resolve the team by id, not by name.** A rename breaks name resolution silently — an
  unmatched team yields *nothing found* rather than an error.

## Checks and review

- **Tests / typecheck:** `cd cli && go test -count=1 ./...` · `cd cli && go build ./... && go vet ./...`
  (build + vet are Go's typecheck; the `cli-ci` workflow runs the same). **`-count=1` is not
  optional here:** the hook tests execute `plugins/trellis/hooks/*` as external files, which Go's
  test cache does not track — edit a hook with no `.go` change and a cached PASS replays over the
  mutation. No dedicated linter is configured; `gofmt -l cli/` is available locally but is not
  CI-enforced.
- **A payload change is a release** (`decision-0028`, applied to the plugin package). Bump
  `plugins/trellis/VERSION` in the same PR as any change under `plugins/trellis/` — an existing
  install re-pulls only when the version string moves, so an unbumped payload reaches fresh
  installs and no one else. `release-guard` fails that pair, keyed on the whole shipped bundle
  rather than `reference/`: #275 edited `hooks/staleness.sh` and was a release (0.10.0 → 0.11.0)
  though `reference/version`, which hashes `reference/` only, never moved.
- **Artifact conformance is CI-applied** (`decision-0098`). A Go test checks the corpus that
  `docs/rubrics/artifact-contract.md` names (`docs/decisions/`, `docs/research/`, `docs/rubrics/`, `profiles/` and the `core/`
  artifacts listed there) against that rubric, in `cli-ci`'s `build-test` job on every PR.
  `build-test` is a required status check on `main`, so a red blocks the merge. Run it locally with
  `cd cli && go test -count=1 -run 'TestCorpusConform|TestArtifactContract' .`. Two of the
  contract's rules take judgment, and the reviewer of a corpus change applies them:
  - **A coupling is not provenance** (rubric check 5, `decision-0047`): a source the artifact's
    correctness is contingent on belongs in `depends_on`, and an `informed_by` edge that is really
    such a coupling is non-conformant. *Example:* `decision-0098` keeps `decision-0082` in
    `depends_on`, because it drops rule 5b on 0082's authority; `research-0010`'s edge to
    `decision-0029` is provenance, so it sits in `informed_by` (TRL-57).
  - **Evidence shows the tell** (rubric check 10): a profile row's `evidence` must actually show
    the tell its verdict claims, because the check confirms only that the field is present.
    *Example:* when `decision-0098` deleted `.claude/agents/corpus-reviewer.md`, four
    `profiles/trellis-self.md` rows still cited it as evidence. The check passed them, since each
    cell was non-empty; review had to re-point them.
- **Branch names are `<category>/<slug>`** — `decision/0075-linear-tracks-the-work`, `research/…`,
  `fix/…`. Since the Linear migration, `feature/*` branches also carry the issue key
  (`feature/trl-22-…`), so **whether a branch is findable by issue number depends on its category**:
  all three current `feature/*` branches are, no `decision/*` branch is.
- **Grove is retired** (`decision-0076`). The plugin, its `grove:<role>` subagents and the
  `/grove:*` commands are gone, and `.grove/` with them. Citations to **grove-the-repo**
  (`grove/adr-00NN`) are a different thing and remain live and load-bearing — `decision-0045`
  is superseded in part by `grove/adr-0010`. Do not "clean up" those.

## Maintaining project instructions

`AGENTS.md` is the canonical home for shared project instructions. Edit new shared rules
here, outside managed blocks. `CLAUDE.md` is the Claude adapter, not a shared-rule edit
surface. Genuinely Claude-only rules belong in `.claude/rules/`.

Trellis project choices remain in `.trellis/` configuration files. Do not hand-edit
managed blocks.

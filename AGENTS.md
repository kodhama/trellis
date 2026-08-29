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
> `decisions/` and `research/`. *This file is Layer B — instance #1, the first methodology
> Trellis supervises — not Trellis's product agent-instructions.*

## The iron rule (most important design constraint)

Trellis must **always ground out in concrete, project-specific artifacts** — a real
instructions file, real gates, real sub-agents. If it ever just *describes* process
instead of *producing and enforcing* it, it has failed. Same rule applies to this repo:
prefer producing a checkable artifact over writing about one. **And it applies to our own
rules:** every invariant or abstract instruction carries ≥1 concrete example (few-shot) — *a
rule you can't exemplify is probably vaporware.*

## Operating method

The method lives in `decisions/`, **not restated here** — a decision is the current truth, and a
summary in this file only goes stale against it. Read the record before relying on a rule.
`decisions/` is append-only: you *supersede* with a forward pointer, never edit.

**There is no `status` field** (`decision-0082`). **Merging to `main` is the acceptance:** an
artifact on `main` is current truth and may be consumed; one not yet merged may not. Nothing to
flip, ever. **Supersession is marked by the forward pointer** — `superseded_by`, or
`superseded_in_part_by` when the remainder is live. *Artifacts written before 2026-08-29 keep
their `status:` lines as history — several carry the maintainer's intent act in a trailing
comment. Read them as accepted; do not add the field to anything new, and do not strip it from
anything old.*

**Never tell the maintainer a change is blocked on an artifact's recorded state.** If he has
asked for a PR, open it. An agent still may not merge on his behalf without his act
(`floor-intent-gate`) — the gate did not move; only the bookkeeping around it went away.

| Before you… | Read |
|---|---|
| write or change an artifact — frontmatter, per-type body sections | `decision-0082` (no `status`; the merge is the acceptance) · `decision-0042` (family lifecycle) · `decision-0037` (`owner: agent` carries *authorship*, not accountability — that stays with the maintainer) |
| supersede a record | `decision-0082` — the forward pointer *is* the mark; `decision-0040` for the partial form |
| retire something, or draw a boundary with what came before | `decision-0081` (supersession authority scales with cost of reversal) · `decision-0074` |
| change a source that has derivatives — the catalog, the CLI's command set | `decision-0028` (update derivatives in the same change; a guard per pair) |
| record a significant choice | append to `decisions/` — the four strategic forks are `0001–0004` |
| plan a build between a decision and the code | the **superpowers** skills (`brainstorming`, `writing-plans`, `executing-plans`) — the spec stage retired in `decision-0079`, and `specs/` with it |
| record a next step | `decision-0078` — name the consumer that will re-present it, or drop it |
| pick up or file work | `decision-0075` — see *Where work lives* below |

Beyond the records: **one logical change per PR**; descriptive, linear history; diffs small
enough to review on a phone. When friction reveals a missing rule, add it *where it fires*,
**prefer retiring to adding**, and keep it subordinate to the work (`inv-self-improvement`).

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

- **Tests / typecheck:** `cd cli && go test ./...` · `cd cli && go build ./... && go vet ./...`
  (build + vet are Go's typecheck; the `cli-ci` workflow runs the same). No dedicated linter is
  configured; `gofmt -l cli/` is available locally but is not CI-enforced.
- **Artifact conformance is agent-applied, not CI-applied** (`decision-0010` — the contract and
  its conformance check ship as agent instructions with no runtime). Invoke the repo-owned
  `corpus-reviewer` (`.claude/agents/`) before merging a change to `decisions/`, `research/`
  or `core/`. It checks the corpus against `core/rubrics/artifact-contract.md`, and is
  read-only by charter — it reports, never fixes.
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

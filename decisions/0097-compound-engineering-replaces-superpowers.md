---
id: decision-0097
type: decision
depends_on: [decision-0079, decision-0085]
changes: [decision-0079, decision-0085]
informed_by: [decision-0015, decision-0028, decision-0078, decision-0081, decision-0082]
owner: agent
date: 2026-09-13
---

# 0097 — Compound Engineering replaces superpowers here; its artifacts are records under `docs/`

## Context

`TRL-22` adopted the superpowers plugin as this repository's working method. `decision-0079` §1
routed planning to its skills, and `decision-0085` retained its specs and plans. `0085` also named
its own end: *"If superpowers is ever replaced, this record retires with it rather than binding its
successor."*

On 2026-09-13 the maintainer replaced it (`TRL-86`):

> "Hi, I'm using compound engineering plugin in math-quest and considering using it globally...
> Can we uninstall superpowers, install ce plugin, and track it in a linear issue?"

He chose where CE's artifacts go, and moved the landing page out of the way (`TRL-87`):

> "install ce setup to docs but move the landing page to another folder"

And he set CE's skills as the way work proceeds, for the first two issues filed under it
(`TRL-88`, `TRL-89`):

> "use CE skills when working on them, up to review/ repair cycles and PR."

### What Compound Engineering does

Quoted from the plugin as installed (`compound-engineering` 3.25.0, `skills/`):

- **Artifact root.** `ce-setup/SKILL.md` reads `docs_root` from `.compound-engineering/config.yaml`,
  and *"Unset -> `<root>` is `docs`"*. It is unset here, so `<root>` is `docs/`.
- **Plans.** `ce-brainstorm` writes to *"`<root>/plans/`"*; `ce-plan` and `ce-work` read and write
  the same tree.
- **Learnings.** `ce-compound`: *"one qualifying solved problem is written as a durable learning
  under `<root>/solutions/`"*. `ce-ideate` writes `<root>/ideation/`.
- **Scratch.** `ce-setup/references/repo-fixes.md`, Step 8: *"Skills that keep local scratch write
  it under `.context/compound-engineering/`"*, which setup offers to git-ignore.

**What it does not say.** I found no CE instruction to commit a plan or learning, and none to delete
one. `ce-setup` offers its compounding directive *"only when the repository treats the store as
tracked, committed knowledge"*. **CE leaves retention to the repository.** So point 3 below is
inferred, not quoted, and the inference is stated: CE keeps scratch in a separate, ignorable
directory and writes its artifacts beside the repository's documents; the maintainer put that root
in `docs/`; and `TRL-89` moves the governance corpus next to it.

## Decision

**1. Compound Engineering replaces superpowers in this repository.** `.claude/settings.json`
enables `compound-engineering@compound-engineering-plugin` at project scope and declares its
marketplace, so fresh clones and cloud sessions resolve it. `superpowers@claude-plugins-official`
is no longer enabled here.

**2. Work between a decision and a merged change uses CE's skills.** That means `ce-brainstorm` or
`ce-plan`, then `ce-work`, then `ce-code-review` before the PR, with fixes through
`ce-resolve-pr-feedback`, and `ce-commit-push-pr` to open it. **This supersedes the sentence in
`decision-0079` §1 that routes planning to the superpowers skills.** Apart from the Consequences
sentence `0085` had already overridden (point 7), nothing else in `0079` changes: the spec stage
stays retired and `specs/` stays deleted. The maintainer still merges (`floor-intent-gate`).

**3. CE's artifacts are committed and kept** under `docs/`: `docs/plans/`, `docs/solutions/`,
`docs/ideation/`, and any other tree a CE skill writes under its root. `.context/compound-engineering/`
is scratch and is not committed. The basis is the inference stated above, not a CE instruction.

**4. `docs/` as a whole is a record root, exempt from `TestDocsClaimOnlyRealCommands`'s walk.** After
`TRL-87`, `docs/` holds no live doc surface. It holds planning records now, and the governance
corpus after `TRL-89`. **The ground is restated here, not inherited from `0085`:** a record
legitimately names artifacts that later retired, and a retirement must not force an edit to the
record of the work before it. The exemption covers only that walk. `marketplaceCommands` still reads
`docs/`, and `cli-ci`'s path filter keeps `docs/**` for it.

**5. Planning records are history, not contracts.** When one turns out wrong about what shipped, the
correction is a dated note beside the claim, never an edit that makes the record look prescient. A
rename sweep under `decision-0015:117-121` is not a correction, and it still applies. This carries
`0085` point 5 forward on its own footing.

**6. The superpowers specs and plans already in `docs/superpowers/` stay as they are.** They record
work done under the previous tooling, so they are neither migrated to CE's layout nor deleted.

**7. `decision-0085` is superseded in full.**
- Points 2 and 3 were about superpowers, and they retire with it. `.superpowers/` is deleted and its
  `.gitignore` line removed.
- Points 4 and 5 are restated above, on their own ground.
- Point 1 restated `0079` and stays true through `0079`.
- **Its correction of `0079` does not lapse.** `0085` overrode the sentence in `0079`'s
  Consequences that dropped the plan-retention question. Point 3 above keeps planning artifacts
  retained, so this record now carries that override, and `0079`'s sentence stays overridden.

## Consequences

- **`AGENTS.md`'s "plan a build" row** names the CE skills in place of the superpowers ones.
- **`cli/docs_consistency_test.go`** replaces its `superpowers` and `.superpowers` name skips with a
  skip of the top-level `docs/` directory, and its comment cites this record. A nested directory
  named `docs` is still walked.
- **The guard walks in `cli/` skip `.context/`**, CE's git-ignored scratch, as structurally outside the
  checkout. A worker session writes scratch inside its own worktree, and walking it failed a local
  run on the agent's notes while CI, which never sees them, passed.
- **`.gitignore`** drops `.superpowers/`.
- **Forward pointers:** `decision-0079` gains `decision-0097` in `superseded_in_part_by`, and
  `decision-0085` gains `superseded_by`.
- **Not decided here:**
  - retiring `corpus-reviewer` for a CI script (`TRL-88`);
  - moving `decisions/` and `research/` under `docs/` (`TRL-89`);
  - whether planning artifacts are worth their upkeep in general. As in `0085`, that is a judgment
    about tooling, and this record would retire with CE rather than bind a successor.
- **Superpowers stays installed in `review-kit`.** This record covers this repository only.

## Self-check

- **The maintainer's words are quoted**, including their typos, rather than paraphrased. `0085`
  exists because a paraphrase widened.
- **Point 3's evidence is weaker than `0085` had for superpowers' specs**, and it is labelled
  inference. A reader who rejects the inference can reject point 3 without rejecting points 1, 2 or 4.
- **Supersession authority.** This record marks two records superseded, one fully, and its author is
  an agent. Under `decision-0092`'s reading of `decision-0081`, merging it is the maintainer's act of
  supersession, not the author's.

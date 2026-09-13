---
id: decision-0097
type: decision
depends_on: [decision-0015, decision-0079, decision-0085]
changes: [decision-0079, decision-0085]
informed_by: [decision-0025, decision-0028, decision-0078, decision-0081, decision-0082]
owner: agent
date: 2026-09-13
---

# 0097 — Compound Engineering replaces superpowers here; its artifacts live under `docs/`

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
- **Plans are committed and never deleted.** `ce-brainstorm` writes to *"`<root>/plans/`"*, and
  `ce-plan` and `ce-work` read and write the same tree. `ce-work` records a `plan_checkpoint`,
  *"the disclosed checkpoint commit when the selected plan was the only canonical dirt"*
  (`ce-work/references/return-to-caller.md:23`). A search of `ce-plan` and `ce-work` for any
  instruction to archive or delete a plan found none.
- **Learnings are committed and kept current.** `ce-compound` writes *"a durable learning under
  `<root>/solutions/`"*. `ce-compound-refresh` audits those learnings *"against the current
  codebase"* and delivers *"committed changes"* (`ce-compound-refresh/SKILL.md`). It keeps each one,
  updates it in place, consolidates it into another, replaces it, or deletes it. On deletion:
  *"Delete the file — git history is the archive; there is no `_archived/`"*
  (`ce-compound-refresh/references/classify.md:11`).
- **Vocabulary.** `CONCEPTS.md` *"Lives at the repo root"*
  (`ce-compound-refresh/references/concepts-vocabulary.md:3`). `ce-compound` and
  `ce-compound-refresh` add, refine, retire or delete its entries as a side effect.
- **Scratch.** `ce-setup/references/repo-fixes.md`, Step 8: *"Skills that keep local scratch write
  it under `.context/compound-engineering/`"*, which setup offers to git-ignore.

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

**3. CE's artifacts are committed, and each kind follows CE's own lifecycle for it.**
- **Plans** (`docs/plans/`) and **ideation** (`docs/ideation/`) are kept as records of the work.
  They are neither deleted nor maintained against the code (point 5).
- **Learnings** (`docs/solutions/`) are a store CE maintains. `ce-compound-refresh` may update,
  consolidate, replace or delete a learning when the evidence supports it, and git history is the
  archive. A refresh that deletes a stale learning is this rule working, not a breach of it.
- **`CONCEPTS.md`**, at the repo root, is a live document that CE maintains the same way.
- **`.context/compound-engineering/`** is scratch and is never committed.

The learnings' lifecycle is quoted. For plans it is partly inference, from `ce-work`'s checkpoint
commit and the absence of any deletion instruction. Ideation is treated like plans on inference alone.

**4. The top-level `docs/` is exempt from `TestDocsClaimOnlyRealCommands`'s walk, and the line is
its location.** That walk guards the documents that describe the product to its users, so none of
them advertises a command the code doesn't have (`decision-0025`). After `TRL-87`, `docs/` holds
none of those documents. It holds planning records and CE's learnings, and it will hold the
governance corpus if `TRL-89` lands. **The ground is restated here, not inherited from `0085`:**
- A planning record legitimately names artifacts that later retired, and a retirement must not force
  an edit to the record of the work before it.
- A learning records how a past problem was solved. `ce-compound-refresh` reconciles it with the
  code, and a retirement elsewhere should not fail on it.

The exemption covers only that walk. `marketplaceCommands` still reads `docs/`, and `cli-ci`'s path
filter keeps `docs/**` for it. **`CONCEPTS.md` gets no exemption.** CE maintains it too, but it
lives at the repo root, outside `docs/`, so the walk reads it like any other document there. If an
entry trips a guard, the entry is what gets fixed.

**5. Planning records are history, not contracts.** This covers plans, ideation and the superpowers
records in `docs/superpowers/`. When one turns out wrong about what shipped, the correction is a
dated note beside the claim, never an edit that makes the record look prescient. A rename sweep under
`decision-0015:117-121` is not a correction, and it still applies. Learnings are not history: they
follow point 3. This carries `0085` point 5 forward on its own footing.

**6. The superpowers specs and plans already in `docs/superpowers/` stay as they are.** They record
work done under the previous tooling, so they are neither migrated to CE's layout nor deleted.

**7. `decision-0085` is superseded in full.**
- **What was about superpowers retires with it:** point 2's retention rule for superpowers' own
  specs, plans and SDD workspace, and point 3 (`.superpowers/` git-ignored). `.superpowers/` is
  deleted and its `.gitignore` line removed. The specs and plans already committed stay, under
  point 6.
- Points 4 and 5 are restated above, on their own ground.
- Point 1 restated `0079` and stays true through `0079`.
- **Point 2's correction of `0079` does not lapse.** `0085` overrode the sentence in `0079`'s
  Consequences that dropped the plan-retention question, *"only as to superpowers artifacts"*. This
  record carries that override and extends it to CE's planning artifacts: point 6 keeps the
  superpowers records, and point 3 keeps CE's plans and ideation. `0079`'s sentence stays overridden.

## Consequences

- **`AGENTS.md`'s "plan a build" row** names the CE skills in place of the superpowers ones. The same
  change adds the maintainer-approved output of `ce-setup`: a row pointing at `docs/solutions/`,
  the compounding directive, and the `ce-noslop` directive.
- **`cli/docs_consistency_test.go`** replaces its `superpowers` and `.superpowers` name skips with a
  skip of the top-level `docs/` directory, and its comment cites this record. A nested directory
  named `docs` is still walked.
- **The guard walks in `cli/` skip `.context/`**, CE's git-ignored scratch, as structurally outside the
  checkout. A worker session writes scratch inside its own worktree, and walking it failed a local
  run on the agent's notes while CI, which never sees them, passed.
- **`CONCEPTS.md` is walked** once a CE skill creates it. A stale live claim there, such as one about
  the retired setup skill (`decision-0072`), fails the docs-consistency tests the same way it would
  in `README.md`.
- **`.gitignore`** drops `.superpowers/` and ignores `.context/compound-engineering/`.
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
- **This record corrects its own first draft.** The draft said CE gave no instruction to commit or
  delete a plan or learning. An independent Codex review of the pull request found
  `ce-compound-refresh`'s commit and delete instructions. The search behind the claim had covered
  only `ce-plan`, `ce-brainstorm` and `ce-compound`, so the claim was never established. The quoted
  bullets in Context replace it, and point 3 now distinguishes what CE maintains from what it keeps as
  history.
- **A second `corpus-reviewer` pass found internal contradictions, now resolved.**
  - Point 7 retired `0085`'s point 2 whole while keeping that point's correction of `0079`. Point 7
    now separates the two.
  - Point 4 gave a reason for exempting learnings that would also have exempted `CONCEPTS.md`. The
    exemption now rests on location, which is the line actually drawn.
- **Point 3's evidence is uneven, and it is labelled.** The learnings' lifecycle is quoted. Plans rest
  partly on inference, and ideation on inference alone. A reader who rejects that inference can
  reject point 3's first bullet without rejecting the rest.
- **Supersession authority.** This record marks two records superseded, one fully, and its author is
  an agent. Under `decision-0092`'s reading of `decision-0081`, merging it is the maintainer's act of
  supersession, not the author's.

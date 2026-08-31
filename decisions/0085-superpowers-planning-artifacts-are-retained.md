---
id: decision-0085
type: decision
depends_on: [decision-0079]
changes: [decision-0079]
informed_by: [decision-0011, decision-0028]
owner: agent
date: 2026-08-31
---

# 0085 — superpowers planning artifacts are retained; `decision-0079` recorded a wider ruling than was given

## Context

`decision-0079` retired the `spec-*` artifact stage and deleted `specs/`. Its Consequences section
also closed a question it had raised, at `decisions/0079-retire-the-spec-stage.md:105-107`
*(line numbers as of this record; the frontmatter pointer this change adds displaced them by one)*:

> **No orphan follow-ups (`decision-0078`).** This record parks nothing. The one question it
> raised — whether superpowers plans should be retained in-repo — is **dropped**, not deferred:
> plans are session scaffolding, and what survives a change is the decision and the diff.

**That sentence records a ruling the maintainer did not make.** Asked directly on 2026-08-31 what
`0079` was meant to settle, his answer was that he asked for *the old specs* to be deleted to avoid
confusion, and said nothing about superpowers' own planning artifacts:

> "I just asked to delete the old specs to avoid confusion, but didn't say anything about the new
> specs. Just follow what superpowers does."

So the retirement of the ratified `spec-*` stage stands exactly as written. What does not stand is
the generalisation from it to a standing rule about a different artifact class produced by different
tooling.

**The generalisation was not harmless.** It was cited — correctly, against the text as written — by
two independent automated reviewers on `kodhama/trellis#258`, each reading it as this repository's
settled position that committing a superpowers spec or plan is a violation. A record that overstates
its own mandate does not merely sit there; it gets enforced.

### What superpowers actually does

The instruction is *"follow what superpowers does"*, and superpowers retains both artifacts and
deletes neither. **The evidence is stronger for the spec than for the plan, and the difference is
stated rather than smoothed over** — the whole defect this record corrects is a paraphrase that
widened, so it would be self-defeating to widen one here.

**Quoted, for the spec:**

- `superpowers:brainstorming` — *"Write the validated design (spec) to
  `docs/superpowers/specs/YYYY-MM-DD-<topic>-design.md`"*, followed at `SKILL.md:209` by an explicit
  *"Commit the design document to git"*.

**Inferred, for the plan** — `superpowers:writing-plans` gives a location and no commit instruction:

- *"**Save plans to:** `docs/superpowers/plans/YYYY-MM-DD-<feature-name>.md`"* (`SKILL.md:18`). Its
  other mentions of committing are the task template's steps for the *feature's* commits, not the
  plan's.
- The location is what carries the argument: `docs/superpowers/` is tracked, while the working state
  the same tooling creates goes to `.superpowers/`, which this repository git-ignores. A tool that
  wanted plans discarded had a git-ignored directory available and did not use it.
- `superpowers:subagent-driven-development`, at Finish — *"delete this plan's workspace*
  *(`rm -rf <workspace>`) — the git history is the record now"* (`SKILL.md:483-484`; emphasis on
  *workspace* added). The deletion is scoped to `.superpowers/sdd/<plan>/`, named at `SKILL.md:137-138`.
  Nothing instructs deleting the plan document.

The only thing the tooling deletes is the SDD **workspace** — the ledger, task briefs and review
packages under `.superpowers/sdd/`. That is the scratch. Nothing in any of the three skills instructs
deleting or archiving the spec or the plan, and there is no archive step to invoke: git history *is*
the archive.

`0079`'s *"plans are session scaffolding"* is therefore true of the workspace and false of the
plan. The two were conflated.

## Decision

**1. `decision-0079`'s retirement of the `spec-*` stage stands, in full.** `specs/` is gone, the
ratified spec artifact type is retired, `decision-0011` remains changed by it, and the retired-artifact
registry it established is unaffected. Nothing in this record reopens any of that.

**2. Its disposal of the plan-retention question does not stand.** Superpowers specs under
`docs/superpowers/specs/` and plans under `docs/superpowers/plans/` are **committed and retained**,
per the tooling's own contract. The SDD workspace under `.superpowers/sdd/` is scratch and is deleted
when its plan completes.

**3. `.superpowers/` is git-ignored; `docs/superpowers/` is not.** That split is the mechanical
expression of point 2 and is already in place.

**4. Planning artifacts are exempt from `TestDocsClaimOnlyRealCommands`, on the same ground as the
governance corpus.** `docSurfaces` in `cli/docs_consistency_test.go` skips `decisions`, `specs` and
`research` because those records legitimately name artifacts that later retired. A planning document
is the same shape: it records what was believed and decided at a moment, and a change that retires a
command must not force an edit to the record of the work that preceded it. `superpowers` and
`.superpowers` join that list for that reason — **not**, as the exemption was first argued, because
`docs/superpowers/` inherits `specs/`'s exemption by succession. That argument assumed `0079`
permitted retention, which as written it did not.

**The derivative is corrected in this same change (`decision-0028`).** The rationale lived in a
comment above the skip list in `cli/docs_consistency_test.go`, stating the succession argument
verbatim. Leaving it would have left a source and its derivative disagreeing, with the derivative
carrying the reasoning this record repudiates — the precise pair `decision-0028` requires be fixed
together. The exemption is unchanged; only its stated ground is.

**5. Planning artifacts are records, not contracts.** They are not maintained against the code after
the work lands. When a plan or spec turns out to have been wrong about what shipped, the correction
is a **dated note appended beside the original claim**, never an edit that makes the record look
prescient. Five such notes already sit on this branch's own spec and plan, and that is the intended shape.

## Consequences

- **`decision-0079` is changed in part, not superseded.** Its `## Decision` section is untouched;
  only the Consequences sentence quoted above is corrected, and only as to superpowers artifacts.
  The forward pointer records the scope.
- **The two open pull requests stop carrying a standing objection.** `#258` and `#259` each commit a
  spec and a plan under `docs/superpowers/`; under this record that is conformant rather than a
  violation, and the reviewers' finding is answered by correcting the rule rather than by deleting
  the artifacts.
- **What is *not* claimed here:** that planning artifacts are useful enough to be worth their upkeep
  in general. That is a judgement about tooling, and the tooling has been adopted (`TRL-22`). If
  superpowers is ever replaced, this record retires with it rather than binding its successor.

## Self-check

- **This record was written by the agent that made the mistake it corrects.** The exemption in point
  4 was originally argued on a premise this record now shows to be false, and the argument survived
  three of that agent's own reviews before two external reviewers caught it independently. Point 4 is
  therefore restated on its own footing rather than carried forward; a reader who disagrees with the
  new footing should reject point 4 without needing to reject points 1–3.
- **The maintainer's words are quoted rather than paraphrased**, because the whole defect being
  corrected is a paraphrase that widened.
- **`decision-0081`'s cost-of-reversal framing applies and is cited honestly** — that record calls
  itself *"(proposal, not a decision)"* at `decisions/0081-supersession-authority-by-cost-of-reversal.md:17`. Under it, correcting an over-wide
  record before anything depends on it is cheap; leaving it costs a further reviewer objection on
  every subsequent change that uses the tooling.

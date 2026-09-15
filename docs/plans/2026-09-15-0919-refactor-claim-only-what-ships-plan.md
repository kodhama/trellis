---
title: Claim only what ships; keep only evidence that ran - Plan
type: refactor
date: 2026-09-15
topic: claim-only-what-ships
artifact_contract: ce-unified-plan/v1
product_contract_source: ce-brainstorm
execution: code
---

# Claim only what ships; keep only evidence that ran - Plan

## Goal Capsule

- **Objective:** Whoever reads what Trellis is, in the README, the marketplace listing, the landing site or `AGENTS.md`, is told only what installing it gives them today. The repository keeps only evaluation evidence that a result or decision used, and everything it retires can be restored from a commit it names.
- **Means:** One approved identity sentence replaces the old claim on every surface, with README as the only place that describes Trellis at length. The discovery brief moves to `docs/research/` as the dated founding thesis. The self-profile, its schema section, the lexicon and rubric checks 9 to 11 retire. The stalled `does-trellis-help` suite and its scorecard sync are deleted, and annotation-vs-absence's raw run files leave the tree, cited by commit.
- **Product authority:** the maintainer (`gundisalwa`), through the Simplify Trellis orchestrator, issue TRL-102; six scoping answers on 2026-09-15. Idea 1b (TRL-103, living specs), idea 4 (guards follow their subject) and idea 5 (one delivery path) are not active scope.
- **Execution profile:** three PRs in order (evidence, scaffolding, positioning), each tracked by its own sub-issue of TRL-102. Only the positioning PR is a plugin release.
- **Stop conditions:** a check marked for removal turns out to cover decision records; evidence that a session-settled decision cannot work; a review finding whose fix would change consumer-visible text beyond what R1 to R7 approve. Each goes to the Simplify Trellis orchestrator as a question.
- **Open blockers:** this is a draft, not yet reviewed or approved: `ce-doc-review` has not run, and the maintainer has not approved the requirements. The Resolve Before Planning question below is also open.

---

## Product Contract

### Summary

Trellis describes itself by what ships: "Trellis is a small set of working rules for AI coding agents. A Claude Code plugin loads them into every session, and a project can switch individual rules off." README is the one place that says what Trellis is at length, and every other surface repeats that sentence or points to README. The artifacts built for features that never shipped retire, and so do the stalled `does-trellis-help` suite and the raw run files of the one experiment that ran, each restorable from a commit the repository names.

### Problem Frame

The identity statement describes a product that was not built. `AGENTS.md:3-6` says Trellis "fits, teaches, adapts, and guards whatever methodology a project uses", README opens with a layer that "learns its shape", and the plugin listing and landing site say the same. Of the discovery brief's §8 MVP, the conformance sub-agent was retired (`decision-0098`) and Author-mode v0 was never built. What a consumer installs is the rules text, the session-start hook that loads it, the `.trellis/rules.toml` opt-out and the remove skill. Since #316, the one per-rule setting a project has is switching a rule off, so "held as firmly as you choose" is no longer true either.

Artifacts for the unbuilt half cost repair on every rename. `profiles/trellis-self.md` was "authored by hand (Assess does not exist yet, cluster 1)" (`:75-76`), and four stacked repair preambles sit above its body. Only tests read it, and only the corpus conformance test reads `core/lexicon.md`.

The evaluation tree holds evidence that did not run. `does-trellis-help` ran 2 of its 11 tasks, and `eval/experiments/README.md` has said "full run pending" since 2026-07-20. Its runner reads the shipped payload, so #316 had to edit it for a suite nobody runs. The one experiment that did run, annotation-vs-absence, keeps 181 raw run files (187 KB) in the tree; `research-0012:145` cites their directory, and no file cites a single run.

### Key Decisions

- **Describe today's product; the direction stands.** (session-settled: user-directed — chosen over narrowing the product with a new record that retires the "fit any methodology" direction of `decision-0002` and `decision-0003`: it removes every false claim without re-making a strategic fork.) No decision record is written for the repositioning. Governs R2, R5, R8.
- **The identity sentence is the plain-rules wording.** (session-settled: user-directed — chosen over "a governance layer for AI coding agents" and a sentence listing example rules: it matches the maintainer's own description of Trellis as "a set of rules", names only the supported host, and carries no count that could go stale.) Codex stays unnamed because README states it "is not supported yet". Governs R1.
- **The sentence goes on every surface.** (session-settled: user-directed — chosen over leaving the site to a separate issue, or leaving the plugin listing and site unchanged: the marketplace listing and the landing page are where outsiders read the claim.) This makes the positioning PR a plugin release, and both manifests carry the sentence verbatim, the Codex manifest included. Governs R2, R3, R4, R6, R7.
- **On the site, the approved scope is the hero and meta lines plus unshipped-feature claims, with no redesign.** This plan reads "drop claims about features that never shipped" as covering every section of the landing page, because the question put to the maintainer named a line outside the hero ("a conformance check it runs on itself"). Sections, layout and install tabs stay. Governs R6, R7.
- **`does-trellis-help` is deleted, restorable from a named commit.** (session-settled: user-directed — chosen over keeping a runnable minimum of about 40 files, or keeping every file marked stopped: the suite cost upkeep on payload releases and had not run in two months.) Governs R11, R12, R13.
- **The run data is cited by commit, with no tag.** (session-settled: user-directed — chosen over an annotated archive tag pushed to origin, or keeping the files in the tree: a commit on `main` cannot be force-pushed away or deleted under ruleset 18526979, while no ruleset covers tags, and it needs no repository action outside the PR.) Governs R14.
- **Checks 9 to 11 retire with the files they check.** (session-settled: user-directed — chosen over keeping the schema's profile section and checks 9 to 11 running on test fixtures until idea 4, or deferring the whole retirement to idea 4: idea 4 waits on idea 1, and a check with nothing left to check is the scaffolding this work removes.) Checks 9 to 11 run only on `expression-profile` artifacts, so none covers decision records. Governs R17, R18, R19.
- **No decision record for the retirements.** Every deleted file is restorable from a named commit, nothing a consumer's session receives changes, and a wrong deletion would show up in the next review or test run. By `AGENTS.md`'s record test that earns no record; the reasoning lives in this plan and the PR bodies. Governs R22.
- **Three PRs, each with its own sub-issue.** Evidence, scaffolding and positioning are separate logical changes, and splitting them keeps each diff reviewable on a phone without changing what ships, since only the positioning PR reaches consumers. The evidence PR goes first so no later PR edits the catalog while the scorecard sync still watches it. Governs R24.

### Requirements

**Positioning**

- R1. The identity sentence reads exactly: "Trellis is a small set of working rules for AI coding agents. A Claude Code plugin loads them into every session, and a project can switch individual rules off."
- R2. README opens with R1's sentence, and every statement README makes about what Trellis does or includes is true of what a consumer gets by installing it today.
- R3. `AGENTS.md` makes no identity statement of its own: its "What Trellis is" paragraph, and any other sentence describing what Trellis is, gives way to a pointer to README.
- R4. The descriptions in `plugins/trellis/.claude-plugin/plugin.json` and `plugins/trellis/.codex-plugin/plugin.json`, and the opening line of `plugins/trellis/README.md`, carry R1's sentence, and `plugins/trellis/VERSION` moves in the same PR.
- R5. README mentions what has not shipped only as direction, in one place that points to the founding thesis and to `decision-0002` and `decision-0003`; no section presents an unbuilt capability as shipped or in progress.
- R6. On the landing site, in both `site/index.html` and `site/lp-content.md`, the hero subtitle and the meta description carry R1's sentence, and the page title and eyebrow claim nothing R1 does not.
- R7. Every other line of the landing site that presents an unshipped capability as part of Trellis is corrected or removed inside its existing section.
- R8. The discovery brief moves under `docs/research/` as the dated founding thesis. Its existing text and section numbers stay unchanged, it is marked with the date it was written and as a statement of intent rather than a description of the product, and it passes the corpus conformance test.
- R9. Every live link to `agentic-dev-meta-layer-brief.md` points to the brief's new path.
- R10. The payload generator's usage text makes no claim beyond R1.

**The stopped eval and the run data**

- R11. `eval/experiments/does-trellis-help/` (50 files) and the `eval-scorecard-sync` workflow with its `check-scorecard.sh` are deleted.
- R12. `research-0011` records the suite as stopped: the date, the reason (2 of 11 tasks ran, the full run pending since 2026-07-20, the runner needing edits whenever the payload changed), and the commit on `main` that restores it.
- R13. Every citation of the suite's task-03 notes, in `research-0011` and `research-0012`, points to that commit instead of a path the deletion removes.
- R14. The 181 files under `eval/experiments/annotation-vs-absence/runs/` are deleted, and `research-0012` names commit `30f53c1` as where they live, with the command that restores them.
- R15. annotation-vs-absence keeps its runner, task, fixture, scorecard and aggregator, and the shared `eval/fill.py` and `eval/prompts/reviewer.md` it reads; `eval/README.md`, `eval/experiments/README.md` and the experiment's README describe the tree as it is afterwards.

**Scaffolding for the unbuilt half**

- R16. `profiles/trellis-self.md` (the only file in `profiles/`), `core/lexicon.md`, and the expression-profile schema in `core/schemas/typed-artifacts.md` (section 2 and its profile-only lifecycle text) are deleted. The catalog schema, including its `default_C1` and `default_C2` fields, stays.
- R17. Rubric checks 9, 10 and 11 retire, and check 6 loses its `expression-profile` and `lexicon` rows. The three check numbers stay, each pointing to this change, as check 12's does.
- R18. Every test, fixture, guard and rubric passage that reads a deleted file or enforces a retired check goes with it, and the checks that remain (1 to 8) keep running unchanged against the live corpus.
- R19. `AGENTS.md`'s *Checks and review* names the corpus as it is afterwards, and its passage on the contract's rules that take judgment keeps only the coupling-versus-provenance rule.
- R20. `core/README.md` lists what `core/` holds afterwards and names no unbuilt component as belonging there.
- R21. No live text outside the records names a deleted file as current: README's repo map, the catalog's note on what reads it (and its generated copy), and the rubric's corpus list included.

**Records**

- R22. Each record whose sentence this work makes untrue gets one dated note, and every live line citation into a noted record is corrected, as `AGENTS.md`'s *Operating method* sets out on `main` at merge. Each PR body lists the records it noted, and no new decision record is written.
- R23. A research note that states a deleted file or path as current is corrected in place, and one that records a plan made at the time stays as written.

**Delivery**

- R24. The work ships as three PRs merged in order: the stopped eval and run data (R11 to R15), the scaffolding (R16 to R21), then positioning (R1 to R10). Each leaves `main` with no live reference to what it deleted, and each carries the R22 and R23 changes its deletions cause.

### Acceptance Examples

- AE1. **Covers R2, R5.** **Given** README after the positioning PR, **when** a reader looks for supervisor mode, the structural gate, per-rule strictness, the lexicon or the expression profile, **then** none is described as shipped or in progress, and the direction appears once, pointing to the founding thesis and `decision-0002` and `decision-0003`.
- AE2. **Covers R4.** **Given** the plugin at its new `VERSION`, **when** a user reads either manifest's description, **then** it is R1's sentence.
- AE3. **Covers R12, R14.** **Given** a checkout of `main` after the evidence PR, **when** someone runs the restore command `research-0011` or `research-0012` gives, **then** the deleted suite or run files come back byte-identical at their old paths.
- AE4. **Covers R15.** **Given** a full clone after the evidence PR, **when** annotation-vs-absence's runner starts, **then** every input it reads exists, and it numbers new runs from 1 in a fresh `runs/`.
- AE5. **Covers R17, R18.** **Given** the scaffolding PR, **when** the corpus conformance tests run, **then** they pass, checks 1 to 8 still fail their known-bad fixtures as before, and no test reads `profiles/`, `core/lexicon.md` or the deleted schema section.
- AE6. **Covers R8.** **Given** `decision-0002`, whose frontmatter cites `brief-§9.2`, **after** the brief moves, **then** the conformance test still accepts the reference and §9.2 of the moved brief still heads the same text.

<!-- ce-section: work-relationships -->
### How This Work Fits Together

This plan covers idea 3 of the Simplify Trellis ideation. The breakdown below is the project's current understanding, not a committed roadmap.

- Idea 1b (TRL-103), decision records consolidate into living specs — **shares** `docs/decisions/`, `AGENTS.md` and the rubric; this plan's dated notes follow whatever `AGENTS.md` says on `main` when each PR merges.
- Idea 4, guards follow their subject — **depends on** idea 1 and later deletes the rest of the corpus conformance tests and the rubric; this plan retires checks 9 to 11 first.
- Idea 5, one delivery path — **can proceed independently of** this plan; README's install and host-support sections are its to change.
- Idea 2 (#316, merged) — already removed the inert Active-subset code the ideation listed under this idea.

### Scope Boundaries

- README's install, host-support and local-quality sections, apart from any identity claim inside them; "Codex is not supported yet" stays.
- A redesign of the landing site, and `site/invariants.html`, which is rendered from the catalog.
- The license section's conditional line on future paid services and the site's matching "Free & open" lede, which state a policy (`decision-0019`) rather than a shipped feature.
- The catalog's entries, the shipped rules text, and the C1/C2 dials, which the shipped invariants reference shows.
- `decision-0002` and `decision-0003`, which stand as the product's direction.
- An archive tag, and the local-only `archive/research-0013` tag.
- The remaining conformance tests and the rubric's checks 1 to 8 (idea 4).
- Repository settings, including GitHub's social-preview image; `.github/social-preview.png` reads "GOVERNANCE FOR AGENTIC DEVELOPMENT", and whether it is the configured preview is unverified.

### Dependencies / Assumptions

- `30f53c1` is an ancestor of `origin/main` (checked 2026-09-15), and ruleset 18526979 rejects force-pushes to and deletion of `main`.
- Checks 9 to 11 run only on `expression-profile` artifacts: `checkProfile` is called only for that type (`cli/corpus_conformance_typed_test.go:176`, `:223`).
- The commit that restores `does-trellis-help` is the newest `main` commit holding the suite when the evidence PR opens; it is re-checked before merge in case another PR edits the suite first.
- Linear's GitHub integration may close an issue when a PR linked to it merges (unverified), which is why each PR carries its own sub-issue.
- simplify-1b may change the dated-note rule before these PRs merge; the rule on `main` at merge applies.

### Outstanding Questions

**Resolve Before Planning**

- Should stopping `does-trellis-help` add an entry to the team's Linear Ideas doc, with a trigger such as "a second project wants evidence that Trellis helps", so something can bring the suite back? The doc already carries a trigger for profiles ("Adaptive application front-end"). The entry would be a change outside the repository.

**Deferred to Planning**

- The brief's new filename and `id`, and how its frontmatter and a `## Open questions` section meet the research-note contract without changing its existing text.
- Which records get a dated note, and which of their sentences are history rather than current claims.
- Which site and README lines count as unshipped-feature claims under R2, R5 and R7.

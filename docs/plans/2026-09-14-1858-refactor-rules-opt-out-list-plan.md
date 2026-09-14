---
title: Rule configuration shrinks to an opt-out list - Plan
type: refactor
date: 2026-09-14
topic: rules-opt-out-list
artifact_contract: ce-unified-plan/v1
product_contract_source: ce-brainstorm
execution: code
---

# Rule configuration shrinks to an opt-out list - Plan

## Goal Capsule

- **Objective:** A project that adopts Trellis configures it by naming only the rules it switches off, and a plugin release that adds, renames or retires a rule can never cost that project its rules or rewrite its file.
- **Means:** `.trellis/rules.toml` keeps its `[rules]` rows, but only a row set `active = false` has any effect. `strictness` and `seeded_from` are ignored, the posture presets and posture headers stop shipping, and both hooks' reconcile, quarantine and rewrite machinery retires. No existing file has to change. `decision-0102` records the change.
- **Product authority:** the maintainer (`gundisalwa`), through the Simplify Trellis project, issue TRL-97; requirements approved 2026-09-14. Idea 5 (one delivery path) and idea 3 (claim only what ships, including the eval) are not active scope.
- **Open blockers:** none.

---

## Product Contract

### Summary

`.trellis/rules.toml` keeps its `[rules]` rows, and only a row set `active = false` has any effect. A rule with no row applies, and `active = true` rows, `strictness` and `seeded_from` are ignored. `governed = false` still opts a project out. The posture presets and headers retire with the hook machinery that kept every file's rows in step with the plugin, and every existing file keeps working unchanged.

### Problem Frame

The file carries three things, and only the rows decide which rules apply.

A valid `strictness` selects one sentence in the delivered text. An invalid one makes Codex refuse the whole file and load no rules, while Claude reads it as adaptive. `decision-0033:28-30` records that the lean "is descriptive, not enforced (nothing acts on it yet)", and the two shipped presets are byte-identical in all sixteen rows. `seeded_from` has no behavioural reader anywhere.

The rows must list every rule the plugin ships, so every catalog change leaves every existing file out of step. `decision-0083:25` records the cost: "A single unmatched row cost all sixteen rules, every session, until a human edited the file." The repair added reconciliation, quarantine and a write-back mandate to both hooks. It then needed seven guards against a broken payload driving that repair (`decision-0083`'s entry-point table), a byte-budget degradation path on Codex, and cross-host parity tests. On 2026-09-14, four of the nine Kodhama consumer files were still on the fourteen-row set and reconciled every session.

What does work is switching a rule off. `research-0012` measured it on one rule and one task: control 95%, row off 0%, a +95-point effect. No known consumer uses it yet, since every known Kodhama file switches every rule on, but it is the one lever worth keeping.

### Key Decisions

- **Sparse rows, not a new list.** (session-settled: user-directed — chosen over a `disabled` list with a transition reader and follow-on migration PRs, or both forms permanently: every existing file stays valid under every installed plugin version, so no repository migrates.) The maintainer accepted the cost: existing Kodhama files keep rows that no longer do anything, and this repository's file keeps `strictness = "firm"` while its sessions receive the "By default" sentence. Governs R1, R2, R4, R16, R17.
- **Every project receives the "By default" posture sentence.** (session-settled: user-directed — chosen over "Firmly" for everyone, dropping the sentence, or keeping `strictness` as an optional key: eight of the nine known consumers already receive it, so no untested wording reaches them; this repository moves from "Firmly".) Governs R14.
- **`governed = false` stays.** (session-settled: user-directed — chosen over retiring it or replacing it with new syntax: a decline stays a committed, reviewable project fact, as `decision-0070` ruled.) Governs R3, R7, R9.
- **The records this change makes out of date carry one dated note each, not new pointers.** (session-settled: user-directed — chosen over new `superseded_in_part_by` pointers: the maintainer's convention for records a change makes out of date.) The maintainer set the scope at every record the change makes untrue, fourteen in all, over the five first named. A record whose every Decision point retires is retired in full and carries `superseded_by` instead, as his answer on TRL-98 keeps it; the orchestrator reconciled the two answers that way. Governs R24, R25.
- **`decision-0102` is written.** The maintainer's rule for plugin behaviour changes (relayed from TRL-98) asks for a record when a wrong call could silently stop rules reaching consumers' sessions. This change replaces the logic that decides which rules reach every session on both hosts, and deletes the reconciler and the guards `decision-0083` and `decision-0084` built around it against silent blackouts; a wrong deletion there is exactly that failure. Governs R23.
- **The file stays the adoption marker wherever adoption has no other sign.** Where the plugin's files live outside the repository — a user-scope install, or a marketplace install at project scope, whose files sit in the user's plugin cache — and on Codex, a project adopts by having the file (`decision-0070` D4 and D7, `decision-0077`). "No file means every rule applies" holds only where adoption is otherwise evidenced: a plugin vendored into the repository, or a curl install that seeds the file. Governs R8, R9, R10, R11.
- **A bad entry costs that entry, never the session.** A floor or an unknown slug set `active = false`, a malformed row, or an unknown key is ignored with a warning, and the file's valid rows keep their effect; only a file that cannot be parsed at all degrades to every rule applying. A declared `governed = false` still means no rule applies. This follows `decision-0083` §5: over-governance the session announces beats under-governance nobody sees. Governs R5, R6, R7.
- **Nothing reconciles, so nothing is rewritten.** With a missing row meaning active and an unknown row ignored, a file cannot fall out of step with the plugin. Governs R15.
- **New files hold only what has effect.** (session-settled: user-approved — chosen over new files carrying every shipped rule as an `active = true` row: such a file is only written when a repository adopts Trellis, and no older install loses rules silently on it.) A file written from now on carries no `strictness`, no `seeded_from` and no `active = true` rows. Installed plugin versions are pinned per checkout and update independently of this release, and older ones read such a file differently: measured on 2026-09-14, 0.6.0 and older refuse it loudly and load no rules, and 0.7.0 to 0.23.x keep its `active = false` rows but reconcile it and instruct writing every row back. Full `active = true` rows would read cleanly on 0.6.0 and later but not on 0.3.0 or 0.5.0, and would keep spreading inert rows. Governs R9, R11.
- **The annotation mechanism stays; filtering switched-off rules out of the text is idea 5's.** The hook keeps delivering the full rules text with the activation rule and the rows, the shape `research-0012`'s annotation arm measured. Governs R13.
- **Legacy delivery shapes keep their frozen text and their file.** A vendored overlay, an inline managed block, or a rendered file from an older installer carries rules text that applies a rule only when its row says `active = true`, so a project on one of those shapes must keep a full row set. Nothing here writes into those projects; idea 5 migrates those shapes.

### Requirements

**The file**

- R1. `.trellis/rules.toml` keeps its `[rules]` table of `slug = { active = … }` rows, and a row set `active = false` switches that rule off.
- R2. A rule with no row applies, a rule whose row says `active = true` applies, and an empty file means every shipped rule applies.
- R3. `governed = false` keeps its meaning on both hosts: no rule applies, floor rules included, whatever else the file holds (`decision-0070` D5). It is recognized as both hooks recognize it today: exactly one uncommented top-level `governed` assignment above any table header, read before the rest of the file.
- R4. `strictness` and `seeded_from` are ignored without comment.
- R5. A row that sets a slug the installed plugin does not ship to `active = false` is ignored, and the session is told which slug and that another plugin version may ship it.
- R6. A floor row set `active = false` is ignored, and the session is told floor rules cannot be switched off.
- R7. A malformed row, or a key the format does not define, is ignored with a warning, and the file's other rows keep their effect; a file that cannot be parsed at all delivers every rule and the session is told why. A file that declares `governed = false` falls under R3 in either case.

**Adoption**

- R8. With the plugin vendored under the project's own `.claude/skills/` and no file, every rule applies (`decision-0070` D3, unchanged).
- R9. With the plugin's files outside the repository and no file, the hook keeps announcing and loading nothing (`decision-0070` D4, `decision-0077`); on acceptance the agent writes a file holding only what has effect, and on decline one holding `governed = false`.
- R10. On Codex, the file's presence keeps locating and adopting the project (`decision-0070` D7).
- R11. `install.sh` seeds a missing file holding only what has effect, and leaves an existing file untouched.

**Delivery**

- R12. Given the same file, both hooks deliver the same rules text, the same set of rules switched off, and the same warnings.
- R13. The delivered rules text states the activation rule once: a rule applies unless its row says `active = false`, and floor rules always apply while the project is governed.
- R14. Every project receives the "By default" posture sentence, whatever `strictness` its file carries.
- R15. Neither hook reconciles, quarantines, or tells the agent to rewrite `.trellis/rules.toml`.

**Existing files**

- R16. Every existing `.trellis/rules.toml` stays valid as it is, and nothing in this change asks a project to edit it.
- R17. This repository's own `.trellis/rules.toml` is left unchanged: `strictness = "firm"` and sixteen `active = true` rows, all now inert.

**What retires, and what follows it**

- R18. The posture presets and posture variants stop shipping: `rules-a.toml` and `rules-b.toml` go, `trellis-a.md` and `trellis-b.md` become one header carrying the "By default" sentence, and the per-posture inline blocks become one of each.
- R19. Every reader of a retired file and every restatement of the old semantics moves in the same change: the payload generator, `install.sh` with its bundle manifest and rendered footer, both hooks' remedy messages, the Codex fallback block, the remove skill, both READMEs, the landing page, the catalog's row-set note and its mirror, and `scripts/plugin-smoke.mjs`. Where those consumer docs explain the rows, they say that removing rows from a file is safe only once every install that opens the repository runs this release.
- R20. Tests that guard retired behaviour retire with it, including the check that compares the rules payload against the default rows preset and the guard that read that preset as a project's rows. Tests that guard behaviour that survives (adoption, the opt-out, the rules payload's own terminator and slug list, the read gateway, delivery shapes) are re-pointed rather than dropped.
- R21. The two eval runners that read retired payload files still run against the inputs they were designed for.
- R22. The change ships as one plugin release, with `plugins/trellis/VERSION` bumped.

**Records**

- R23. `decision-0102` records what this change retires and why.
- R24. Each record this change makes out of date is marked, and none gets a new `superseded_in_part_by`: `decision-0033`, `0051`, `0053`, `0058`, `0068`, `0070`, `0072`, `0074`, `0078`, `0083`, `0084`, `0086`, `0087` and `0088`. A record whose every Decision point retires carries `superseded_by: [decision-0102]`; `decision-0086` and `decision-0088` are the candidates, and `decision-0088` stays partial if its D3 or D4 still describes what `install.sh` reads. Every other record gets one dated note directly under the title, below any notes already there, pointing at `decision-0102`, in the shape `> **Dated note, YYYY-MM-DD — TRL-<n> (PR #<n>):** <what is no longer true, naming the point or sentence>. <where the current behaviour is stated now>.` The PR body lists each record marked and whether it was marked in full or in part.
- R25. Live line citations into those records (in other live records, code comments, tests and `AGENTS.md`) are corrected in the same change; citations in plans and dated inventories stay as they are, such as `decision-0081`, which dates its own citations. A grep on 2026-09-14 counted twenty across ten files before that exclusion.

### Acceptance Examples

- AE1. **Covers R1, R2.** **Given** a file whose only row is `inv-minimal-first = { active = false }`, **then** every other shipped rule applies and `inv-minimal-first` does not.
- AE2. **Covers R3, R7, R13.** **Given** a file holding `governed = false`, alone, beside a key the format does not define, or beside a row that does not parse, **then** no rule loads, floor rules included, and no delivered text says floor rules apply.
- AE3. **Covers R5.** **Given** a row `inv-not-a-rule = { active = false }`, **then** every rule applies and the session is told `inv-not-a-rule` is not a rule this plugin ships.
- AE4. **Covers R6.** **Given** a row `floor-transparency = { active = false }`, **then** every rule applies, `floor-transparency` included, and the session is told floor rules cannot be switched off.
- AE5. **Covers R7.** **Given** a malformed row `inv-bounded-context = off` beside `inv-minimal-first = { active = false }`, **then** every rule except `inv-minimal-first` applies and the session is told which row is malformed.
- AE6. **Covers R9.** **Given** the plugin's files outside the repository and no file, **then** no rules load and the session is told Trellis will govern once accepted; **after** acceptance the file carries no `strictness`, no `seeded_from` and no `active = true` rows.
- AE7. **Covers R2, R15.** **Given** a plugin release that adds a rule, and a project whose file has fourteen `active = true` rows, **then** the next session is governed by every rule including the new one, and nothing asks for the file to change.
- AE8. **Covers R4, R14, R17.** **Given** this repository's file, with `strictness = "firm"` and sixteen `active = true` rows, **then** every rule applies under the "By default" sentence and nothing is said about either key.

<!-- ce-section: work-relationships -->
### How This Work Fits Together

This plan covers idea 2 of the Simplify Trellis ideation. The breakdown below is the project's current understanding, not a committed roadmap.

- Idea 5, one delivery path and a hook that checks adoption and prints the rules minus opt-outs — **depends on** these semantics.
- Idea 3, claim only what ships — **shares** the two eval experiments this change touches; retiring or tagging them is idea 3's.
- Idea 1a, decision records become rare — **shares** the record convention this change follows: one new record and dated notes, no new pointers (R23 to R25).

### Scope Boundaries

- Tidying the inert rows and keys out of existing files, this repository's included. Nothing owes it. Removing rows is safe only once every install that opens a repository runs this release: installs are pinned per checkout, and plugin 0.6.0 and older refuse a file with missing rows.
- Filtering switched-off rules out of the delivered text, and consolidating delivery onto one path.
- Migrating legacy delivery shapes: vendored overlay, flat overlay, inline managed block, rendered file from an older installer.
- Retiring the eval experiments, the C1/C2 dials, `profiles/` or `core/lexicon.md`.
- Rewriting hook comments beyond the sections this change deletes (idea 7).

### Dependencies / Assumptions

- `research-0012` measured an `active = false` row, the form this change keeps, on one rule and one task; that it holds for other rules is assumed.
- A renamed rule's opt-out lapses: the old slug's row is ignored with R5's warning, and the new slug applies until the project adds a row for it. No slug has been renamed so far.
- The Kodhama repositories are the known consumers: nine on 2026-09-14, with wisp and review-kit since being retired by the maintainer. No record counts installs outside them.
- Installed plugin versions are pinned per checkout and update independently of this release. On 2026-09-14 the Codex install ran 0.23.0 and the user-scope Claude install 0.22.0; at about 20:11 UTC that day the trellis, math-quest, stewards and kodhama checkouts were updated to 0.23.1 at project scope, taking effect as their sessions restart.
- Codex remains carried rather than supported (`plugins/trellis/README.md:137-141`); R12 keeps the two hosts in step regardless.

### Outstanding Questions

#### Deferred to Planning

- Whether the hooks deliver the file verbatim or a normalized statement of which rules are switched off (R12, R13).
- The exact content of a newly written file: an empty `[rules]` table, and whether a comment names the opt-out form (R9, R11).
- Which row decides when a file repeats a slug; today's reconciler keeps the first (R1).
- Which faults make a file one that cannot be parsed at all, such as an unreadable file or a second `[rules]` table (R7).
- How the eval runners keep their inputs: reading them at the commit each experiment recorded, or a frozen copy (R21).

### Sources / Research

- Consumer scan (`gh api`, 2026-09-14): nine `kodhama/*` repositories carried `.trellis/rules.toml`; every row was active; `trellis` was `firm` and the rest `adaptive`; four were on fourteen rows; none used a legacy delivery shape.
- Installed versions (2026-09-14, before about 20:11 UTC): `~/.claude/plugins/installed_plugins.json` pinned per checkout path; the trellis, stewards, kodhama and wisp main checkouts ran 0.6.0 as marketplace installs at project scope with files in the user's plugin cache; math-quest 0.8.0; review-kit 0.11.0; worktrees ranged from 0.3.0 to 0.23.1; the user-scope Claude install ran 0.22.0; `~/.codex/plugins/cache/kodhama/trellis/` held only 0.23.0. Shipped slug counts: 0.3.0 fourteen, 0.5.0 fifteen, 0.6.0 onward sixteen.
- Measurements (2026-09-14): against a file holding only `[rules]`, and one whose only row is `inv-minimal-first = { active = false }`, the cached 0.6.0 `staleness.sh` refuses and loads no rules; the `staleness.sh` and `codex-context.mjs` on `main` at 83adb8c keep the `active = false` row, reconcile, and instruct writing the rows back.
- Record sweep (fresh-context agent, 2026-09-14; entries spot-checked against the records by this session and by the orchestrator): the fourteen records in R24 carry decision or consequence text this change makes untrue; `decision-0077` and the research records do not.
- `docs/decisions/0033-park-seed-and-custom-postures.md`, `0053-live-rows-assembly-retires.md`, `0070-adoption-is-the-consent-act-not-installation.md`, `0077-silence-is-not-an-adoption-act.md`, `0083-rules-toml-reconciles-itself.md`, `0084-codex-reaches-reconciliation-parity.md`.
- `docs/research/0012-annotation-vs-absence-mechanism.md:148-164`.
- `plugins/trellis/hooks/staleness.sh` path B and its reconciler; `plugins/trellis/hooks/codex-context.mjs` `parseRulesToml` and `reconcileRows`; `cli/payload.go`, `cli/apply.go`, `cli/onboarding.go`; `install.sh`.

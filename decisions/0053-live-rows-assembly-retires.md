---
id: decision-0053
type: decision
status: approved  # maintainer's intent act 2026-07-20, in-conversation ("flip and fold executor run in the same PR") — this flip records it (decision-0046); pre-gate: corpus-reviewer no-violation (stats recomputed from raw runs), its open check closed by git diff, precision+scope notes applied
depends_on: [decision-0051, research-0012]
owner: agent
date: 2026-07-20
---

> **Provenance.** The maintainer's original target for `rules.toml` (2026-07-19, during
> `decision-0051`'s shaping): rows governing **live** — "editing the rules.toml would
> directly influence its application." Deferred then on an untested salience-leak prior;
> `research-0012` was designed to decide it on data (maintainer: "amend decision-0051
> only on this data"), the experiment ran 2026-07-19, and its pre-registered decision
> rule returned **AMEND**. This record is that amendment, as a decision superseding
> `decision-0051` **in part** — 0051 is long-merged, so the append-note instrument its
> own same-day amendment used is not available; a forward pointer lands on 0051 at this
> record's `approved` flip.

# 0053 — live rows: the readout ships complete, `rules.toml` governs at read time; assembly retires

## Context

- **The evidence (`research-0012`, batch 1 — REPEATS=20, all gates valid, zero
  exclusions):** leak (annotation − absence) **+0%**, 95% CI [−16%, +16%], Fisher
  p = 1.000 — twenty workers read the deactivated rule in full and *none* applied it;
  the salience-leak prior that justified deactivation-by-absence is refuted in this
  setting. Row effect (control − annotation, the clean row-only pair): **+95 points,
  p = 3×10⁻¹⁰** (exact; the script prints 0.000) — a one-token row flip in otherwise byte-identical context is
  near-binding in both directions. Verdict per the pre-registered rule: **amend**
  (point estimate ≤ +10 and CI upper bound +16 < +25 — equivalence, not mere
  non-detection).
- **The prerequisite holds:** `@import` of a `.toml` file loads into context —
  verified empirically 2026-07-20 (import cell recites the file's content with tools
  forbidden; the no-import negative control returns NOT-FOUND).
- **What 0051 got right stands; what it deferred now lands.** The authority split
  (consumer-owned `rules.toml` + generated `internal/`), rows-as-truth,
  posture-as-seed, the naming de-collision, the slug tags — all stand, and the slug
  tags matter *more* under live rows (they are how a reader matches a row to its
  rule). What retires is the **write-time selection mechanism**: 0051 rule 4's
  fragment assembly existed solely to produce *subset* readouts, and its honest note
  ("no per-session reader") described the constraint the experiment has now removed.
  Selection moves from write time to read time; the write-time machinery has no job
  left (`inv-minimal-first`: this amendment *deletes* machinery).
- **The tested wording is the shipped wording (trellis#170 watch-out).** The
  experiment validated a specific authority header, a rows-inlined-below-the-rules
  layout for the inline channel, and live-rows seed comments (`header_arm_toml`).
  The build ships text as close to the tested content as the payload format allows —
  what the data validated is the artifact, not a paraphrase of it.

## Decision

**1. The readout ships complete.** `rules.md` (the assembled all-active render) is the
one readout every install carries; per-row subset assembly at refresh **retires**.
Fragments (`reference/rules/*.md`) remain the **generator's** release-time render
source — they leave the shipped payload only if nothing else consumes them (executor
verifies; the manifest follows either way).

**2. Rows govern at read time.** The readout carries the authority header (the
eval-tested wording, adapted): rules apply **only** where `rules.toml` marks the row
`active = true`; a `false` row's rule does not apply; **the two `floor-` rows apply
regardless of their row value**. Delivery of the rows into context:
- **Import channel** (Claude Code): the managed block imports both
  `@.trellis/internal/trellis.md` and `@.trellis/rules.toml` — block-level imports,
  the empirically tested shape; no nested-import dependency.
- **Inline channel** (no-import harnesses): the block inlines the rows below the rules
  — exactly the experiment's annotation/control-arm sandwich, now the shipped shape.

**3. Row edits take effect immediately** (the import reads the current file each
session; inline installs re-paste the block's rows section on refresh, their general
update cadence). `/trellis:setup` refresh remains the payload-update path and still
**validates** `rules.toml` on every run: unknown slugs, missing rows, or floor rows
set `false` are surfaced loudly (floor rows: named as overridden-by-floor, never
silently honored) — validation survives assembly's retirement.

**4. The absence-era wording retires with the mechanism.** The readout
preamble/footer ("assembled from the active rows… refresh to re-assemble"), the toml
seeds' "no effect until refresh" comments, and the block tail's "refresh the overlay —
re-assemble" sentence all become their live-rows variants (the experiment's
`header_arm_*` transforms are the model). No shipped text may claim refresh-time
semantics for rows.

**5. Superseded in part, stands in part.** This record supersedes `decision-0051`'s
rule 4 (fragment assembly; the "no per-session reader" delivery stance), the
refresh-time-effect semantics in its rules 2–3, and rule 5's absence-era
closing-line specification ("(Generated from your `rules.toml` …)" — retired with
the footer it lived in); 0051's rules 1, 6, 7, the rest of rule 5 (the naming
de-collision), and its amendment (expression.md retirement) stand. `decision-0051` gains
`superseded_in_part_by: [decision-0053]` + an append-only pointer note **at this
record's `approved` flip, in the same change**.

## Consequences

- **Machinery deleted, not added:** `SKILL.md` step 4's selection `cat` → a plain copy
  of `rules.md`; the floor-held assembly override → the header's floor sentence + the
  step-3 validation; the two-oracle verify (concatenation equality; head+readout+tail
  sandwich) → byte-compares against static payload files; the `#112` sentinel
  ("Generated from your…") is re-anchored or retired with the footer it lived in
  (executor resolves against the guard's actual text, loudly).
- **Payload changes:** authority header lands in the readout/header content; block
  gains the `@.trellis/rules.toml` import (import style) / the rows section (inline
  style); seeds' comments become the live-rows wording; `checksums`, `version` stamp,
  `install.sh` manifest, scorecard (unchanged content-wise unless the ✗/directive
  render moves — executor verifies), `docs/invariants.html` untouched (no catalog
  change), repo overlay refreshed. Family and consumers pick it all up on refresh.
- **The experiment machinery is an archive, not a derivative** —
  `eval/experiments/annotation-vs-absence/` stays byte-untouched (its runs are the
  evidence; its eval-local transforms were the prototype of what now ships).
  `research-0012` carries the results appendix (landed with this record).
- **trellis#166 unparks after this build** — the trigger-format experiment was
  sequenced behind this decision's implementation; it also serves as the replication
  vehicle for the one-rule/one-task generalization limit.
- **No sequencing hazard remains:** no experiment is running; batch 1 is complete and
  recorded. A future batch-2/replication runs against the new shipped state and says
  so in its provenance.

## Open questions

- **Does the header's floor sentence hold floors as strongly as mechanical assembly
  did?** The experiment tested rule deactivation, not floor-override attempts. If a
  consumer sets a floor row `false` and agents honor the floor sentence, fine; if not,
  that is a new experiment in the established harness — cheap to run, same shape.
- **One rule, one task** (inherited from research-0012): the zero-leak result
  generalizes by assumption; #166 doubles as replication.

## Self-check (gate)

Evidence-driven by a pre-registered decision rule (`research-0012`, approved), with
the prerequisite verified before drafting — not vibes (`inv-independent-judgment`).
Supersede-in-part via a new record with a forward pointer, never an edit to the merged
0051 (`inv-auditable-archive`); the pointer obligation is named for the flip
(`decision-0040`/0051-on-0033 precedent). Deletes machinery rather than adding it
(`inv-minimal-first`), and ships the *tested* wording rather than a paraphrase
(`floor-transparency` applied to evidence). `depends_on` carries genuine coupling only
(the record amended; the evidence consumed — both approved, no draft consumed,
`inv-directional-flow`). Left at `draft`: the author does not grade its own decision —
the gate and the `approved` flip are the maintainer's intent act (`decision-0046`),
ideally after an independent pass.

---

> **Amendment (2026-07-20, append-only — added in the introducing PR pre-merge, the
> same instrument as `decision-0051`'s expression amendment; ratified in-conversation,
> the maintainer's merge seals it).** The conformance gate on this decision's build
> surfaced a propagation this record missed: a pre-0053 install's `rules.toml` carries
> seed *comments* claiming refresh-time row semantics, and under point 2's import
> channel those comments now ride into the very context that carries the authority
> header — the exact contradiction point 4 retires from shipped text, made permanent
> because the file is consumer-owned and never refreshed. **Setup's refresh therefore
> offers — never imposes (`decision-0049` consent pattern) — to update a legacy
> `rules.toml`'s retired refresh-semantics comments to the current seed wording,
> touching comments only, never a row.** Declined → noted and moved on; autonomous →
> file left untouched, the stale comments reported loudly. The seed-once/never-clobber
> ownership rule is unchanged: this is a consented comment repair, the same class as
> `decision-0049`'s ignore-file offer.

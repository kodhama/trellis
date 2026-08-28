---
id: decision-0078
type: decision
status: approved  # maintainer's intent act, relayed on TRL-24: "The maintainer has ruled: the ledger gets reviewed and pruned periodically, and combining moral cousins is that review's job, not this issue's" — the mint was decided before this record was drafted, and the framing quoted below is the maintainer's. This flip records that act (decision-0046, decision-0022). Author (agent) != approver (maintainer). Scope of the act: the mint. The wording, the pair set, the release call and the scope additions below are the maintainer's to accept at merge
depends_on: [invariants-v1, signature-catalog-v1]
informed_by: [decision-0018, decision-0021, decision-0027, decision-0028, decision-0040, decision-0074]
owner: agent
date: 2026-08-28
---

> **Provenance.** Raised by the maintainer: *"If it's not on an action pipeline it's like it doesn't
> exist."* Filed as `kodhama/trellis#250`, then as **TRL-24**, which carries the argument and the
> ruling. The evidence is a measurement in `kodhama/math-quest` on 2026-08-28, filed there as
> **MQ-77**.

# 0078 — `inv-no-orphan-followups`: a deferral with no consumer is not tracked work

> **Decided: mint it.** The maintainer ruled before drafting began, and ruled twice — that the row
> is added rather than merged into a neighbour, and that combining moral cousins belongs to a
> periodic review of the whole set, not to this change. That ruling is why the compression argument
> below is *recorded* rather than *acted on*.

## Context

- **The measurement.** `kodhama/math-quest`, 2026-08-28:
  `output/implementation-artifacts/deferred-work.md` held **61 entries** against **3 lines marking
  anything resolved** — and all three were written that same day, by the audit that found the ratio.
  Real defects were sitting in it: a grading path rounding through `Number()` where the repo forbids
  it, a keyboard path bypassing its own key mask, pad keys announcing nothing to a screen reader.
  All three were fixed that week **because a fresh review re-derived them**, not because anyone read
  the ledger.

- **Why this is invariant-shaped and not one repo's sloppiness.** Two instructions, each correct on
  its own terms, manufactured the orphans *between* them. The repo says tracking lives in the
  tracker. Its build workflow's review step says every `defer` finding is appended to that markdown
  file, and explicitly not to check for duplicates. A designed consumer existed — a sweep skill that
  partitions the ledger — and had **never been switched on**. Nobody did anything wrong; the graph
  had a **write with no read**, and no rule made that visible.

- **Why the existing rows do not cover it** (the three nearest, each checked against its own text):
  - `inv-graph-maintenance` is about **coherence** — update what depends on a change. A ledger entry
    with no dependents is perfectly coherent and still lost. Its `(ops)` violated case does name *"a
    parked item sits in a channel its executor never reads, trips nothing, and rots"* — the closest
    existing text — but that clause is about **one home per kind of information**, i.e. a copy in
    the wrong place, not a deferral with no consumer anywhere.
  - `inv-self-improvement` is the nearest neighbour (*"improvement signals are surfaced and acted
    on"*), and its `(process)` violated case is *"a PR raises the same open question every time,
    with no follow-up, and it rots unowned."* But it is about **learning from friction**: a glitch
    becomes a fix so it cannot recur. A deferral is not a glitch — it is **planned work with no
    address**.
  - `inv-handover-points` / `inv-gate-at-handover` govern moments work **changes hands**. This is
    work that never changes hands at all.

- **The cost this row must not create.** Named in the issue and load-bearing on the wording: an
  invariant phrased as *"file it somewhere"* pushes toward filing an issue for everything, and a
  tracker full of never-to-be-done issues is a **worse** graveyard — more expensive to prune than a
  markdown file. **"Drop it" is therefore first-class in the directive, not a footnote**, and the
  set entry states outright that dropping is not a failure. The measured failure was a ledger with
  no reader; the failure this row could *cause* is a tracker with no pruner, and `inv-prune-bias`
  already cuts that way.

## Decision

**1. `invariants-v1` gains `inv-no-orphan-followups`** (operating set, `trellis-design`,
*provisional*). A recorded next-step counts as recorded only if something will **re-present** it.
The entry carries the **address test** (a named consumer with a cadence — a queue, a tracker, a
failing or skipped test, a switched-on sweep; *"someone reading this file later"* is not one), the
**three legitimate outcomes** (do it now · drop it on the record · escalate), and the reason a
plausible ledger is worse than none. Surfacing rides SI-1's channel discipline and SI-2's existing
rituals; it shares SI-3's prune-bias and declares `inv-self-improvement` its neighbour.

**2. The catalog gains the matching entry** with **three honored / three violated** pairs —
process, code, ops — tag-aligned per `decision-0027` point 1. Coverage arithmetic moves
**15 → 16 assessable slugs**.

**3. A set amendment, not a catalog-only door.** Minting a catalog slug with no registry entry forks
annotation from annotated — `inv-graph-maintenance`'s own violated case. This follows the
`decision-0018` / `decision-0021` / `decision-0074` path.

**4. The wording is our synthesis** (naming guardrail): no external provenance implied. The evidence
is one measured in-house instance (math-quest, MQ-77) plus the two counter-instances this repo holds,
recorded in `profiles/trellis-self.md`.

**5. The release ships in this change, not after it.** `plugins/trellis/VERSION` 0.5.0 → 0.6.0, with
both plugin manifests. See Consequences for why this is not deferrable.

## Consequences

- **The derived chain regenerates in the same change** (`decision-0028`), both layers, as
  `decision-0074` mapped them: the **render** chain — catalog → `cli/assets/invariants.md` →
  `reference/` render → `rules.md` → both inline-block sandwiches → `checksums` → `version` stamp →
  `install.sh` bundle manifest → the invariant scorecard, plus the preset rows in
  `rules-a.toml` / `rules-b.toml` and this repo's own `.trellis/rules.toml` (**without a row the rule
  ships but is inactive**) — and the **contract** chain: `spec-0007`'s canonical inventory (`version`
  2 → 3), `spec-0002` §1 check 2 + AC1, `core/rubrics/artifact-contract.md`, the `corpus-reviewer`
  checklist, and `profiles/trellis-self.md`. **The catalog's own derivatives note named only three of
  them** and now names the row-set chain, because the note not naming them is the mechanism by which
  `decision-0074`'s author missed four.

- **VERSION bumps here because nothing will bump it later.** `d4a2c7b` had to bump 0.4.0 → 0.5.0 a
  day after `326ed16` shipped the 15th rule without one, leaving every consumer cached at 0.4.0
  *"governed by fourteen while the repo believed it shipped fifteen."* The guard written to stop
  that recurring — **trellis#245, "guard the payload→VERSION pair"** — is **still open**: there is no
  `release-guard.yml` in `.github/workflows/`, and `72f6a5e` is reachable only from
  `origin/decision-0028/payload-change-is-a-release`. So the obligation is discharged by hand, and
  **that open PR is itself an instance of this rule's `(ops)` violated case** — a designed consumer
  that was never switched on.

- **Row-set mismatch fails loudly on the two governed paths, and this was verified, not assumed.**
  `plugins/trellis/hooks/staleness.sh:600-601` refuses with `TRELLIS_RULES_NOT_LOADED … (missing:
  inv-no-orphan-followups)` and injects nothing; `plugins/trellis/hooks/codex-context.mjs:442` emits
  `invalid-rules` and delivers no context. Neither degrades to defaults, and the Claude side derives
  its expected slug set from the shipped `rules.md` rather than a literal, so it fires without an
  edit. The issue's stated risk — *"do not ship a row that silently disables the overlay in every
  downstream repo"* — does not hold on these paths: the overlay is disabled **loudly**, with a
  per-category repair remedy, which is the designed behaviour.

- **The curl path still lies, and this is the second row addition it bites.** `trellis#241` → **TRL-2**
  (open, Backlog): `staleness.sh`'s path C validates markers, stamp presence and stamp freshness, then
  emits *"That file and .trellis/rules.toml govern this session"* at line 445 — and every branch inside
  it exits before the row validation that begins at line 564. A curl consumer who re-runs `install.sh` gets a
  freshly stamped 16-rule rendered file beside their preserved 15-row `rules.toml`, and is told the
  config governs. Verified against current source, not inherited from `decision-0074`'s text.
  **Disclosed again rather than fixed**, on the same grounds 0074 gave — the fix is a hook behaviour
  change with its own tests, and TRL-2 is its address. Recorded here because a rule about deferrals
  with no consumer must show its own deferral's consumer.

- **Scope added deliberately: the repeated count literal is removed, not re-patched.**
  `decision-0074`'s self-check named its root cause as *"a sweep that matched only some of the shapes
  a count takes."* Re-patching `15 → 16` in six Go assertions would reproduce it on the 17th row. So
  `cli/payload_test.go`'s `assessableSlugs` is now the **one pin** and every count assertion in the
  suite derives from `len(assessableSlugs)` — including `cli/remove_skill_test.go`'s `16/16` needle,
  which now fails loudly on the 17th row instead of passing against an equally stale `SKILL.md`.
  `cli/rules_test.go`'s hardcoded `order[13]`/`order[14]` floor check is indexed from the end — that positional pair was a silent arithmetic edit on every row
  addition. `TestInvariantRulesCoverCatalog`'s four-slug spot-check became a full-set loop in the
  same pass: a count-only assertion passes when one slug is swapped for another. This is
  `inv-self-improvement` applied to the change that would otherwise repeat the friction, and it is
  flagged as scope rather than smuggled (`floor-transparency`).

- **Three pairs, not four.** A fourth *(research)* pair was drafted — an open question carried into
  an artifact's `## Open questions`, which its next consumer must read, versus one noted in a scratch
  file nothing depends on — and **dropped for readout budget**, the concern `decision-0074` flagged
  when it shipped four against `decision-0027` point 3's stated two. Recorded rather than filed:
  this record is the drop, not a deferral, and under this rule that is the correct disposal.

- **Renumbered 0077 → 0078 mid-flight, and the collision is worth naming.** This record was
  authored as `decision-0077`. While it was being written, a different session merged its own
  `decision-0077` (*"silence is not an adoption act"*, trellis#251) to `main`. **Decision ids are
  allocated by whoever merges first, and a branch cut before the race has no way to see it** — the
  `ratify-guard` check catches a draft left on `main`, and nothing catches two branches claiming one
  id. Found here only because CI never reported on the PR, which sent me to `main`'s log. Left as an
  observation rather than a fix: trellis#252 is currently open claiming an already-taken
  `decision-0076`, so this is not a one-off, and the fix is a check with its own tests. **Under this
  record's own rule that observation needs an address or a drop** — it is dropped, on this record,
  because no one has agreed to do it; if it recurs a third time, that is the trigger to file it.

- **The compression objection is recorded, not resolved.** `decision-0021`'s minting test is
  *mechanism*, and `inv-minimal-first` argues against the catalog growing. The maintainer ruled that
  combining moral cousins belongs to a periodic review of the set, not to this change — the same
  place `decision-0074` sent it (**trellis#239**). This row joins that review's corpus rather than
  pre-empting it.

- **Count change, not a pair-door change.** As with `decision-0074`: `decision-0027`'s pair door and
  `decision-0040`'s directive-extension door both preserve the count; this does not. It is an
  amendment, flagged as one.

- **The readout grows by one rule.** Unlike `decision-0074`, this content was **not** already in the
  readout in any form — no existing directive says anything about where a deferral goes. The cost is
  a genuine addition, not a re-render.

- **`profiles/trellis-self.md` gains a row at `confidence: inferred`.** `AGENTS.md` states the
  address test outright for ideas (*"An idea filed as an issue is a to-do nobody agreed to"*), which
  is real honoring evidence — but the repo holds two live counter-instances, both named in the row:
  trellis#245, and the catalog's own open question *"Owed to the Assess build (cluster 1)"*, whose
  named consumer does not exist yet. `verified` would overclaim.

## Open questions

- **Does the address test hold outside agent-generated ledgers?** The measured instance is an agent
  writing faster than a human reads. A human-scale backlog with the same write/read ratio would be
  the same defect, but no such instance is recorded yet.
- **Is "named consumer" checkable enough for Assess?** The signature offers a mechanizable tell
  (entries far outnumbering resolutions; an append step with no drain step) and a judgment tell (is
  the consumer real?). Whether the first is sufficient on its own is owed to the Assess build — and
  that consumer does not exist yet, which this record names rather than hides.
- **Does it survive the compression review?** trellis#239 takes `inv-self-improvement` /
  `inv-deliberate-succession` as its starting corpus; this row belongs in it. If the set carries an
  unnamed dimension that collapses several entries, this is a candidate, and `inv-prune-bias` cuts
  both ways.

## Self-check (gate)

Homed at the layer the change touches: a new slug is a set amendment, not a catalog-only door
(`decision-0018` / `0021` / `0074` path). The three rejected neighbours are recorded with their
kill-reasons **quoted from their own entries**, including the two clauses that come closest to
covering this (`inv-graph-maintenance`'s parked-item case, `inv-self-improvement`'s unowned-open-
question case) — because omitting the near-misses is how a merge argument gets won by silence. The
count change and the deliberate scope addition are both flagged, not smuggled (`floor-transparency`).
The dropped fourth pair is disposed of by this record rather than deferred, which is the rule
applied to itself. The iron-rule obligation is discharged in-artifact: three honored / three violated
pairs spanning process, code and ops.

**Claims verified against source, not inherited.** `decision-0074` states the curl-path false
all-clear and the loud-failure behaviour; both were re-read in current code before being repeated
here, with line numbers. The payload→VERSION guard was **assumed live and is not** — `git
merge-base --is-ancestor 72f6a5e HEAD` fails and `.github/workflows/` has no `release-guard.yml`.
Correcting that assumption is what moved the VERSION bump into this change.

**Not independently reviewed at the time of writing.** The author of this record also authored the
catalog entry, the code changes and the sweep. `inv-independent-judgment` is not discharged by a
self-check: the `corpus-reviewer` run and the PR review are the independent passes, and
`decision-0074`'s four missed obligations came from exactly there. Anything this section claims is
the author's own reading until then.

Flipped to `approved` on the maintainer's intent act relayed on TRL-24, not the author's judgement
(`decision-0046`); the record of that act is in the frontmatter.

# Self-repairing `.trellis/rules.toml` — design

**Date:** 2026-08-30
**Tracks:** TRL-20 (umbrella), TRL-2, TRL-27
**Supersedes in part:** `decision-0072` — its confirm-first row-repair remedy only
**Produced by:** `superpowers:brainstorming`, architectural path. This is a superpowers design
doc, not a resurrection of the retired `spec-*` artifact type (`decision-0079` moved planning
to the superpowers skills; `AGENTS.md` routes here).

## Problem

A `.trellis/rules.toml` whose row set does not match the slugs the installed payload ships
causes a **total delivery blackout**. `plugins/trellis/hooks/staleness.sh:612-613` fires on any
mismatch and injects **nothing** — one bad row costs all sixteen rules:

> `TRELLIS_RULES_NOT_LOADED — this project's .trellis/rules.toml does not match the rules the
> installed plugin ships ($slug_report). Nothing was injected, because a partial or unknown row
> set cannot be applied honestly.`

The repair is then manual, gated, and recurs every session until someone edits the file. Three
filed issues are three arms of this one defect:

| Issue | Arm |
|---|---|
| TRL-20 | A catalog-changing plugin upgrade leaves every existing project one row short, blacked out every session |
| TRL-2 | On the curl path, path C re-renders and then *claims* governance while a row is missing |
| TRL-27 | `unknown:` has two causes and opposite repairs; the live case would have deleted a ratified row to match a stale plugin |

**The maintainer's framing, which this design adopts:** correction is not the hazard — silence
is. All three mismatch kinds self-correct; the loudness is what gets engineered.

## Maintainer decisions (settled in brainstorming, 2026-08-30)

1. **The agent reconciles; the hook stays read-only.** `decision-0070` D4's *"The hook never
   writes"* (`decisions/0070-...:122`, pinned at `cli/plugin_hook_test.go:1516`) is **untouched**.
   What changes is the hook's *message*: advisory-and-gated becomes imperative-and-reported.
2. **`unknown:` rows are quarantined, never deleted** — commented out with dated provenance.
3. **`missing:` rows are added `active = true`.**
4. **Loudness = session report + in-file provenance.** No new artifact, no journal file.

## Resolution semantics

One table, applied identically by both hosts, in memory (for delivery) and on disk (for repair):

> **2026-08-30, after the branch shipped — "by both hosts" did not ship.** The table below is
> applied by `staleness.sh` only; `codex-context.mjs` still fails closed on any mismatch. See the
> note under *Codex parity* below, `decision-0083` §1, and the Linear follow-up (Trellis team).

| Kind | Resolution | Rationale |
|---|---|---|
| `missing:` | add row, `active = true` | Matches both shipped presets and `decision-0070` D3, where a project-scope plugin with no file at all governs at full strength. A newly ratified invariant behaves the same in a project installed today or two releases ago. |
| `unknown:` | quarantine — comment the row out, keep its value verbatim | The two causes (retired rule / stale plugin) are **indistinguishable at runtime** in config-only mode, which has no version stamp to compare. Quarantine is correct under both: inert clutter if retired, one uncomment away if the plugin was the stale side (TRL-27's live case). |
| `duplicate:` | keep the first occurrence, quarantine the extras | Same mechanism, no "which value did they intend" judgment call. |

Floor rows apply regardless of their value, unchanged.

**Why the parser needs no change:** the validator matches rows anchored at line start —
`staleness.sh:587`, `/^[[:space:]]*(inv|floor)-[a-z-]+[[:space:]]*=/` — so a commented row is
invisible to it. Quarantine satisfies validation with no grammar work, and a quarantined row is
**not** re-reported as `unknown:` next session.

### Written form

```toml
[rules]
inv-minimal-first         = { active = true }
# inv-no-orphan-followups = { active = true }  # quarantined
#   2026-08-30: not in payload@c9c2461c1fea. If a newer Trellis ships this
#   slug, run `claude plugin update trellis@kodhama` and uncomment.
floor-transparency        = { active = true }
```

**The comment instructs no deletion, deliberately.** `TestEveryDeletionInstructionIsGated`
(`cli/plugin_hook_test.go:579`) scans hook emit strings for deletion verbs and requires a
confirmation clause on each — so a template sentence saying to delete a retired block would
re-impose the gate this change removes. A quarantined row is inert; sweeping it stays a human
act the file does not need to instruct.

## The two halves

**Delivery half — no blackout.** On a mismatch the hook injects the full readout plus the
**reconciled** row block, labelled as reconciled. It must not `cat` the raw file in this branch:
delivered rows that disagree with the authority header would be worse than the blackout.

**Repair half — announced, not asked.** The emit instructs the agent to write the reconciliation
to `.trellis/rules.toml` and to **state what it changed, per row, before substantive work**.

**Why this may be ungated, against `decision-0072`'s finding #6** (*"retiring a confirm-gated
writer silently retires the gate"*): that gate guarded **destructive** writes to a
consumer-owned file — clobbered rows, deleted rows. Under quarantine semantics no prior value is
ever lost; every repair is additive or commenting, and fully reversible from the file itself
without git. The gate is not being dropped, it is no longer engaged. `floor-intent-gate` is
satisfied because nothing irreversible happens; `floor-transparency` is satisfied by the report.

## Codex parity

`plugins/trellis/hooks/codex-context.mjs:14-31` validates against a **hardcoded 16-slug array**,
while the Claude hook derives its set from the shipped `reference/rules.md`. **Nothing in CI
compares the two** (verified: the only `SLUGS` references are inside that file; the
`assessableSlugs` use at `cli/codex_hook_test.go:524` guards `block-codex.md`, a different
surface).

Consequence: on Codex a payload upgrade **cannot** repair the drift, and worse, an `unknown:`
can be produced by the stale array even when the payload is current — the agent would quarantine
a live row and report a reason that is false. Since this change's whole value is that the loud
message is trustworthy, the array goes.

**Change:** delete `SLUGS`/`SLUG_SET`; derive from `reference/rules.md` with the same match the
Claude hook uses. `codex-context.mjs:412` already loads that file. `parseRulesToml` currently
returns `null` → `invalid-rules` on any mismatch; it adopts the reconcile semantics instead.

> **2026-08-30, after the branch shipped — the second half of that sentence did not happen, and
> neither did §"One table, applied identically by both hosts" above.** What shipped is the slug
> derivation and a raised context cap; `parseRulesToml` still returns `null` on any mismatch and
> `codex-context.mjs` still calls `fail(PROJECT_CONFIG, "invalid-rules")`. Reconciliation is
> **Claude-only**. This note records the divergence rather than editing the plan the design
> actually made: adopting reconciliation in a parser is a restructure, not the reordering this
> section scoped, and it was deliberately deferred. See `decision-0083` §1 and its open questions;
> the follow-up is tracked in Linear (Trellis team).

**Unconfirmed, to be settled in implementation, not planned around:** `parseRulesToml` runs at
`:369` and the payload resolves at `:408`, so deriving the slug set needs that order swapped. It
reads like a reordering rather than a rewrite, but the uses in between were not traced.

## Tests first — the state matrix

`decision-0072` records **seven review rounds finding one defect class eleven times**: *"a
replacement remedy that was wrong about the reader's actual state."* That is the failure mode
this section exists to prevent. Enumerate the matrix as failing tests **before** any hook edit.

- **Mismatch kinds:** `missing:`, `unknown:`, `duplicate:`, and the **rename** case (missing +
  unknown simultaneously — the shape `staleness.sh:598-603` documents).
- **File shapes:** absent · complete · partial · `governed = false` · **already-quarantined**.
- **Hosts:** Claude and Codex, same semantics.

Two tests that carry most of the risk:

1. **Idempotency.** A second session over a repaired file reports nothing and changes nothing.
2. **Quarantine invisibility.** A quarantined row is not re-reported as `unknown:`.

**Three existing pins change, deliberately** — `TestRepairRemedyCoversEveryMismatchKind` (:620),
`TestRowMismatchRemedyIsNotDestructive` (:799), and the literal
`"duplicate:, delete the extra occurrences"` (:684). Rewriting a test that pins behaviour is
where a regression hides; each edit is justified in the decision record, not done silently.

## Artifact obligations

| Obligation | Why |
|---|---|
| Decision record, superseding `decision-0072` in part | The confirm-first remedy is replaced; `AGENTS.md` requires the forward pointer |
| `plugins/trellis/VERSION` bump | Payload change is a release; without it no cached consumer re-pulls (`cli/assets/invariants.md:65`). PR **#245** is adding the guard — do not re-file |
| Prose sweep: `README.md:204-215`, `plugins/trellis/README.md` | Quarantine is new user-visible state in a consumer-owned file |
| `corpus-reviewer` before merge | Repo-owned agent; required for `decisions/` changes |

`/trellis:remove` needs **no** change — it deletes `.trellis/` wholesale, quarantined rows included.

## Out of scope — named, not silently dropped

- **The producer-side sweep.** `cli/assets/invariants.md:55-67` makes a row-set change a
  ~15-surface manual obligation enforced by prose; its own note records that `decision-0074`'s
  review *"caught four obligations its author's sweep missed."* This is the root cause upstream
  of all three issues and **has no tracking issue** — to be filed separately.
- **The inline managed block's frozen row copy** (`cli/apply.go:207`, `README.md:210`) — the one
  derivative that genuinely goes stale. Not covered by any of the three issues.
- **Floor-row warning parity** — Codex warns when a floor row is set `active = false`
  (`codex-context.mjs:479-484`); Claude is silent.
- **Whether a `missing:` slug should ever deliver partially** — resolved here by reconciling in
  memory, so the question does not arise.

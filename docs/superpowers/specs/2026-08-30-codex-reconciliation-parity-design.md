# Codex reconciliation parity — design

**Date:** 2026-08-30
**Tracks:** TRL-30 (parity), TRL-29 (fail-closed budget cap, folded in)
**Builds on:** `decision-0083` — its resolution table is ratified and is not reopened here
**Produced by:** `superpowers:brainstorming`, architectural path

## Problem

The `rules.toml` reconciliation work (PR #258) removed the total-blackout-on-mismatch defect **on the
Claude hook only**. `plugins/trellis/hooks/codex-context.mjs` still refuses:

```js
const rows = parseRulesToml(rulesToml, slugs);
if (rows === null) {
  fail(PROJECT_CONFIG, "invalid-rules");
  process.exit(0);
}
```

The design behind that work claimed *"one table, applied identically by both hosts"*. That did not
ship; the claim is annotated as unmet in four places on that branch, and this design is the debt.

**Three defects, one change:**

| | |
|---|---|
| **TRL-30** | Codex fails closed on any slug-set mismatch — the blackout, still live on one host |
| **TRL-30, shape 1** | A file with no `strictness` is Codex-fatal but fine on Claude, so even a *repaired* file can be Codex-invalid. Reachable from the headline example: a reconciled empty file has no `strictness`; adding `strictness = "adaptive"` makes the identical row set load |
| **TRL-29** | Codex refuses when assembled context exceeds its own cap, where Codex itself would spill gracefully — a self-inflicted blackout at a different threshold |

Shipping TRL-30 without TRL-29 would rebuild the blackout at the budget boundary, since reconciliation
only adds bytes. They are one change.

## Maintainer decisions (settled in brainstorming, 2026-08-30)

1. **Codex becomes lenient on a missing `strictness`**, matching Claude — the divergence is fixed at
   its root rather than by having the reconciler write a key the user never chose.
2. **TRL-29 folds in.**
3. **Full symmetry: Codex mandates the write**, as Claude does.
4. **Two implementations, one conformance guard** (approach B, over a single shared implementation).
   Recorded with its expiry: *C becomes correct the moment a third host appears* — at three hosts the
   guard's cost grows and extraction wins. Revisit then, not now.

## The core split

`parseRulesToml` returns `null` for a dozen distinct reasons and the caller treats every one as fatal.
**Only three are reconcilable.** It stops being a gate and becomes a classifier:

| Condition | Disposition | Why |
|---|---|---|
| missing / unknown / duplicate slug | **Reconcile** per `decision-0083`'s table | The ratified semantics |
| missing `strictness` key | **Adaptive**, documented | Matches `staleness.sh:558-560`; decision 1 above |
| `strictness` present but not `firm`/`adaptive` | **Fail closed** | A typo must not silently pick a posture |
| malformed row, unknown top-level key, bad quoting | **Fail closed** | Reconciliation cannot repair a file it cannot parse; inventing structure here is the reconciler guessing |

**This split is the entire risk surface of the change.** Wrong in the permissive direction and a
corrupt file silently governs — the failure `decision-0070`'s *"fail loudly rather than govern
silently"* exists to prevent. Every case above gets a test.

## Components

### 1. `reconcileRows()` in `codex-context.mjs`

Mirrors the awk in `staleness.sh`: missing slugs appended `active = true`; unknown and duplicate rows
commented out with dated provenance; first occurrence of a duplicate kept; a `[rules]` header emitted
when absent, matched leniently (`^[[:space:]]*\[rules\]`) as the Claude side now does.

**Produces** the reconciled TOML text and a per-session count of added and quarantined rows — the
counts must be per-session, not cumulative, the defect already fixed once on the Claude side.

### 2. The conformance guard

One fixture table, **defined once**, run through both hooks, asserting **byte-identical** reconciled
output. This is what makes two implementations safe, and is the repo's own convention
(`decision-0028`: a guard per pair).

Fixtures, at minimum: rename (missing + unknown together) · indented `[rules]` · already-quarantined
(idempotency) · duplicate with differing values · empty file · no `strictness` · a file whose
`[rules]` table is absent entirely.

### 3. Budget degradation (TRL-29)

Assemble with provenance. If over cap, re-assemble **without the provenance comments** and append one
line naming the omission. **The file written by the mandate keeps full provenance** — the archive is
complete, only the working set is trimmed.

The cap stays under Codex's ~2,500-token spill threshold deliberately: spilling hands the model a
head-and-tail preview of the rule set, which is *partial governance* and worse than either full
delivery or an announced omission.

### 4. The mandate

The Claude mandate text, host-adjusted: same non-destructive quarantine, same requirement to report
per row before substantive work, and **no deletion verb**.

**A gap this change creates and must close.** `TestEveryDeletionInstructionIsGated` and
`TestEveryDestructiveInstructionIsGated` scan the *Claude* hook only. This change puts an agent-facing
instruction into the Codex payload for the first time, so both guards extend to cover it. Without
that, the safety argument — *the repair may be ungated because nothing destructive reaches the agent* —
is unenforced on the new host, which is exactly how it was nearly lost on the Claude side.

## Invariants preserved

- **The hook never writes.** `decision-0070` D4. Codex emits text; the agent writes.
- **Quarantine never deletes.** No row's value is lost on any path.
- **Every exit is `exit(0)`.** A hook must never fail the session.
- **Reconcile only the slug set.** Syntax faults still fail closed.

## Testing

Beyond the conformance table, the state matrix from `decision-0083` re-run against Codex: three
mismatch kinds plus rename; file shapes absent / complete / partial / `governed = false` /
already-quarantined; and the four classifier dispositions above.

**Mutation is the standard, not a green run.** The predecessor branch produced five tests that passed
for the wrong reason, every one caught only by breaking the code and confirming the test went red.
Each new assertion here is verified the same way.

## Out of scope

- **`block-codex.md`'s all-or-nothing prose** (`:10,22`) — the vendored-overlay bootstrap, which the
  project describes as *"carried rather than maintained"* (`plugins/trellis/README.md:83-85`). It is
  already named in `decision-0083`'s open questions. Revisit only if that path stops being carried.
- **Extracting a single shared reconciler** — decision 4 above, with its stated expiry.
- **The producer-side row-set sweep** — TRL-28, unaffected by this change.

## Artifact obligations

| Obligation | Why |
|---|---|
| A decision record | Supersedes `decision-0083`'s Claude-only scoping; the four "parity is owed" annotations become history and need their forward pointer |
| `plugins/trellis/VERSION` bump | Payload change is a release; without it no cached consumer re-pulls |
| `corpus-reviewer` before merge | Required for `decisions/` changes |
| Close out the four parity annotations | Four dated notes on the predecessor branch say parity is owed and name TRL-30. Once this lands they describe history, not an open debt. **Do not edit or delete them** — `decisions/` is append-only and the planning docs are historical records. Add a second dated line beside each saying the debt was paid and naming the new decision record, so a reader arriving at the "owed" note is not left believing it still stands |

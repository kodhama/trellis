---
id: decision-0026
type: decision
status: ratified
ratified: 2026-07-05
depends_on: [invariants-v1, spec-0003, decision-0020]
owner: gundi
date: 2026-07-05
---

# 0026 — The overlay always-loads the active rules; examples stay on-demand

## Context

Friction, caught by the maintainer: the M1 overlay auto-imports `trellis.md` → `profile.md`, while the
full catalog (`invariants.md`) is backticked (read on-demand). But `profile.md` listed the active
invariants **by name only** — each invariant's actual *rule* lived in the on-demand reference. So an
agent that never opened `invariants.md` had the names + the one B2 behavior, but **not the rules**. The
invariants were loaded *on demand*, not reliably *active*.

That is the same shape as `decision-0025`'s drift, one level in: a claim ("governed by these
invariants") without the substance present to back it — and a strictness gap in the overlay contract.

## Decision

Split the overlay by role:

- **Always-loaded (the enforcement floor).** `profile.md` carries each *active* invariant as a concise
  **rule** — its one-line `what`, **parsed from the bundled catalog** (single source; the rule can't
  drift from the reference). The governing rules are therefore in context every turn.
- **On-demand (the depth).** The full catalog — the *why*, the with/without examples, and the
  invariants *not* active here — stays in `invariants.md`, backticked. The agent pulls it for detail.

Rationale: loading ~300 lines of examples on *every* message pays context to hold recognition material
needed only occasionally; the **rules** are what must always be present.

**Honest limit (recorded, not hidden).** Always-loaded makes the rules *present*, not *enforced* — an
LLM can still ignore an in-context rule. A gate that actually **blocks** a violation is **supervisor
mode** (hooks on commit/PR), which is not built. M1 is now as strict as an *advisor overlay* can be;
hard enforcement is the supervisor slice.

## Consequences

- `renderProfile` (CLI) emits active rules parsed from the catalog; the plugin's `/trellis:setup` skill
  does the same; the landing distinguishes **rules: always** from **reference: on demand**.
- Rules never drift from the catalog — they are read from the one source, not restated.

## Open questions

- Should per-invariant `C1` (this one enforced, that one skippable) be visible in the always-loaded
  view, not just a single posture lean? Deferred until per-invariant strictness needs to be always-on.

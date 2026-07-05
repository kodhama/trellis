---
id: decision-0028
type: decision
status: draft
depends_on: [invariants-v1, decision-0025, decision-0026]
owner: gundi
date: 2026-07-05
---

# 0028 — Derived resources declare their source, and every pair gets a sync guard

## Context

Three sync failures in a row: docs↔release (`decision-0025`), the always-loaded rules (`decision-0026`),
and catalog↔page. Each time the invariant that *should* have fired is **B1 (graph-maintenance)** —
"dependents re-reviewed on upstream change" — yet it kept not firing, and the maintainer had to catch
it.

The diagnosis (maintainer's): the invariant wasn't **salient at the moment of action**. Our dependency
graph only points **backward** — a dependent declares `depends_on`, but a source has no idea what
derives from it. So editing the catalog surfaced nothing about the page or the bundled copies. An
abstract rule in context is not enough; the reminder has to appear where the edit happens. Invariants
are probabilistic (they can't be made deterministic), so the aim is to **raise the probability** the
right one fires — and put a deterministic floor under the part that can be checked.

## Decision

A general rule, made concrete so B1 actually fires:

1. **Forward-edges.** Every source artifact **names its derived resources**, co-located, so editing it
   shows the dependents right there. (The catalog now lists: rendered → `docs/invariants.html`; copied →
   `cli/assets/invariants.md`, `plugins/trellis/reference/invariants.md`.)
2. **A sync guard per source↔derivative pair** — a deterministic check wherever the relationship is
   expressible: byte-identical for copies, "the render contains the source's content" for renders.
   (`cli/sync_test.go`; docs↔release already has `cli/docs_consistency_test.go`.)
3. **The rule lives where it fires** — a line in `CLAUDE.md`'s operating method (always in the builder's
   instructions), phrased as a trigger, not an abstraction.

Forward-edges raise the odds B1 fires; the guard is the floor for when it doesn't.

## Consequences

- Sources gain a "derived resources" note; each source↔derivative pair gains a check.
- This is **B6 self-improvement**: a recurring friction becomes a standing rule (invariant 8 — add it
  where it fires, prefer retiring to adding).
- Honest limit (`decision-0026`): salience raises probability, it does not guarantee; the check is the
  deterministic backstop for the expressible part.

## Open questions

- A generic "derived-resource" manifest (source → [derivatives, guard]) instead of per-artifact prose,
  if the number of these grows. For now, prose + per-pair checks.

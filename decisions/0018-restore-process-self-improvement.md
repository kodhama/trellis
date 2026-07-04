---
id: decision-0018
type: decision
status: draft
depends_on: [invariants-v1, decision-0009, decision-0014, brief-§5]
owner: gundi
date: 2026-07-04
---

# 0018 — Restore process self-improvement: backport the trigger-driven loop from math-quest

**Raised by:** the maintainer — *merging `inv-self-improvement` (B6) into `inv-graph-maintenance`
(B1) lost one of the most important angles of the project: the **self-improvement of processes**, not
just the referential integrity of the artifact graph.*

## Context

In `invariants-v1`, B6 (`inv-self-improvement`) was **superseded_by** B1 (`inv-graph-maintenance`)
with the framing *"self-improvement is the rules facet of graph maintenance."* Two properties got
conflated:

- **Referential integrity** (B1's core) — keep the dependency graph *consistent and minimal*:
  propagate changes, repair upstreams, prune stale nodes. A **consistency** property.
- **Process self-improvement** (brief invariant 9, the *engine of Pillar II "fit + **evolve**"*,
  `brief-§5`) — the process **learns from friction and gets better over time.** A **growth** property.

You can maintain a graph perfectly and never improve the process. The merge kept the *prune-stale*
overlap and **treated self-improvement as finished** — losing the "evolve" thesis pillar as a
first-class, concrete concern.

**The evidence this matters (and how to fix it) already exists — in the source instance.** math-quest
(the project the invariants were extracted from, `decision-0009`) developed the concrete engine and
**kept it deliberately trellis-compatible** — its `CLAUDE.md` §Self-improvement, verbatim: *"slugs
kept stable and trellis-compatible … so the future supervisory layer imports these rather than
rewrites them."* It also carries the friction that justified hardening it from prose to invariants:
*"the loop was culture-prose and runs skipped it in practice — two builds shipped changes that
actioned parked items without touching them."* Trellis-core lost the **engine**, not merely the label.

## Decision

**Backport math-quest's trigger-driven self-improvement loop and its three hardened invariants into
Trellis-core**, re-elevating process self-improvement to a first-class, *concrete, enforced*
capability.

**The loop (operating guidance):** the operating rules evolve by **triggers, not vigilance** — a
trigger is a `condition → action` recorded **where its firer will trip over it** (never a passive
checklist). The loop **rides the change you already make**: when a change actions a trigger, retire it
in the same change, and record a successor *only* if a genuinely new boundary appeared. Always
**subordinate to product work** — it rides along, never a treadmill.

**The invariants (stable slugs, math-quest-compatible so they import rather than rewrite):**

- **`inv-propagation-surfaced` (SI-1)** — every change SHALL surface a **`## Propagation`** section
  that names each trigger / parked item / feedback disposition it fires or actions — with the
  retirement/update applied **in the same change** — or states **`none`**. *Surfacing is the floor;
  silence is the only violation* (D1). *(e.g. a PR that repairs an ADR must name the specs it
  re-triggers, or say `none`.)*
- **`inv-ride-existing-rituals` (SI-2)** — propagation SHALL ride artifacts the work already produces
  (the PR body, the files touched) — never a separate ceremony, sweep, or scheduled audit.
- **`inv-prune-bias` (SI-3)** — retiring SHALL be preferred over adding; the trigger set MUST NOT grow
  monotonically. *(Consolidates B1's existing "bias toward retiring over adding.")*

**Placement (the fork — maintainer's call, D2).** Recommended: **keep the B1 umbrella** (graph
maintenance) and attach SI-1/2/3 as **named, first-class self-improvement invariants under it** — the
framing is shared with math-quest and is fine; what was missing is the concrete engine. The
alternative — **un-merge B6** into a standalone self-improvement invariant — is available if the
"evolve" pillar warrants its own top-level slug. This ADR proposes the former; see Open questions.

## Consequences

*(Owed downstream, deferred until this decision ratifies — the invariant revision is the intent gate,
D2 / `decision-0014`: significant invariant changes are recorded as ADRs, then applied revise-in-place.)*

- **`invariants-v1` revised** — B1 gains the named self-improvement facet + SI-1/2/3, and the "evolve"
  framing (Pillar II) is restored in words, not just the prune mechanic.
- **The conformance check learns SI-1** — a change's record must carry a `## Propagation` section (or
  `none`); math-quest enforces this *substantively* (the check confirms a change didn't action a
  parked item it failed to name), not just structurally. Enforcement is a **dial** (surfacing floor;
  advisory red check, the maintainer holds the block) — consistent with C1/D1.
- **`CLAUDE.md` §Self-improvement gains the trigger-loop guidance** (it is currently a thin pointer).
- **Ties to `decision-0009`** — that loop improves *a project's process*; `decision-0009` covers how
  *Trellis-core* improves. This ADR restores the former as concrete core content the latter references.

## Open questions

- **The placement fork (D2):** first-class facets under B1 (recommended) vs. un-merging B6 as its own
  invariant. Does the "evolve" pillar warrant a top-level slug, or is slug-compatibility with
  math-quest worth more?
- **How much enforcement machinery is Trellis-core vs. instance-specific?** math-quest wires a
  PR-contract CI check + a `propagation-remediator` dispatcher; the *invariant* (SI-1) is core, but
  the CI wiring is arguably per-instance delivery — draw the line in the backport.
- **`inv-prune-bias` vs. B1's existing prune-bias** — consolidate to one statement, don't double-count.
- **Does this reopen the B6→B1 merge decision itself?** This ADR *refines* it (restores the concrete
  engine) rather than superseding it; confirm that's the right relationship at ratification.

## Supersedes / superseded by

— (none; refines the B6→B1 merge recorded in `invariants-v1` / `decision-0014`)

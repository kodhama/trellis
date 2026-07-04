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

Restore process self-improvement as a first-class, concrete capability by backporting the **intent**
of math-quest's loop — *the process learns from friction and gets better* — **not its mechanism.**
math-quest's specific *`## Propagation`-section-in-a-GitHub-PR* is **one instantiation**; Trellis
encodes the intent and lets each project choose the surface (below). Extracting intent over mechanism
is the whole discipline here (`brief-§7` iron rule; methodology-agnostic `decision-0003`).

**The engine (operating guidance):** the process evolves by **triggers, not vigilance** — a trigger is
a `condition → action` recorded **where its firer will trip over it** (never a passive checklist). The
loop **rides the work you already do** (never a separate ceremony), retires a trigger when a change
actions it, records a successor *only* if a genuinely new boundary appeared, and **prefers retiring to
adding**. Always **subordinate to product work**.

**What a trigger looks like — points of information for improvement** (few-shot, per the iron rule; a
project surfaces the ones its practice actually throws off):

- a **gate skipped** under deadline → was the gate wrong, or the deadline? ;
- **a human overriding Trellis and being right** — the gold signal, a labeled correction
  (`decision-0009`) ;
- **the same ambiguity clarified twice** → fix the upstream, don't re-clarify (B9) ;
- a **"manually verified" flag** → tracked debt: automate or accept it, don't let it rot ;
- a **rule that hasn't fired** in a while → candidate to retire (prune-bias) ;
- a **recurring divergence** from a reference framework → record it (B8) ;
- **friction at the same handover, repeatedly** → the gate or the artifact contract needs work.

**The invariants (intent-defined; slugs kept math-quest-compatible so they import, not rewrite):**

- **`inv-propagation-surfaced` (SI-1)** — the process-improvement signals a change or a work session
  throws off SHALL be **surfaced through the project's chosen channel**, never silently dropped —
  with any retirement/update applied **in the same change**. *Surfacing is the floor (an application
  of **D1**); silence is the only violation.* **The channel is not assumed — see below.**
- **`inv-ride-existing-rituals` (SI-2)** — surfacing SHALL ride the artifacts the work already
  produces — never a separate ceremony, sweep, or scheduled audit.
- **`inv-prune-bias` (SI-3)** — retiring SHALL be preferred over adding; the trigger set MUST NOT grow
  monotonically. *(Consolidates B1's existing "bias toward retiring over adding.")*

**The surfacing channel is a per-instance choice — asked or inferred, never assumed (the load-bearing
correction).** Trellis does **not** mandate GitHub issues or a `## Propagation` section. At
Assess/onboarding it either **asks** the team how they want improvement signals captured
(B9 `inv-clarify-before-commit`) — a PR section, an issue label, a running doc, a standup note, a
chat channel — or **infers** the channel from how the project already works (Assess), surfacing the
inference for confirmation (never silent). The chosen channel is a candidate **expression-profile
field** (per-instance, like the dials; `spec-0002`). This keeps it methodology-agnostic
(`decision-0003`) and augment-never-clobber (`spec-0001` §5). *math-quest's `## Propagation` section is
exactly SI-1 instantiated for a project that lives in GitHub PRs — concrete evidence the intent /
mechanism split is real, not hand-waving.*

**Placement (the fork — maintainer's call, D2).** Recommended: **keep the B1 umbrella** (graph
maintenance) and attach SI-1/2/3 as **named, first-class self-improvement invariants under it** — the
framing is shared with math-quest and is fine; what was missing is the concrete engine. The
alternative — **un-merge B6** into a standalone self-improvement invariant — is available if the
"evolve" pillar warrants its own top-level slug. This ADR proposes the former; see Open questions.

## Consequences

*(Owed downstream, deferred until this decision ratifies — the invariant revision is the intent gate,
D2 / `decision-0014`: significant invariant changes are recorded as ADRs, then applied revise-in-place.)*

- **`invariants-v1` revised** — B1 gains the named self-improvement facet + SI-1/2/3 (intent-defined),
  and the "evolve" framing (Pillar II) is restored in words, not just the prune mechanic.
- **The conformance check learns SI-1 — against the *declared channel*, not a fixed section.** It reads
  the project's chosen surfacing channel (from the profile/config) and checks that a change's
  improvement signals were surfaced *there* — never that a `## Propagation` heading exists. For a
  project that chose the PR-section channel, that heading is what it looks for; for another, its issue
  label or running doc. The check is **substantive** (did a change action a signal it failed to
  surface?), and enforcement is a **dial** (surfacing floor; advisory by default, the maintainer holds
  the block) — consistent with C1/D1.
- **The improvement channel becomes a candidate `expression-profile` field** (`spec-0002` extension) —
  Assess proposes it (asked or inferred), the human ratifies (D2).
- **`CLAUDE.md` §Self-improvement gains the trigger-loop guidance + the examples** (it is currently a
  thin pointer). Trellis-self declares its own channel while it's at it (dogfood).
- **Ties to `decision-0009`** — that loop improves *a project's process*; `decision-0009` covers how
  *Trellis-core* improves. This ADR restores the former as concrete core content the latter references.

## Open questions

- **The placement fork (D2):** first-class facets under B1 (recommended) vs. un-merging B6 as its own
  invariant. Does the "evolve" pillar warrant a top-level slug, or is slug-compatibility with
  math-quest worth more?
- **How much enforcement machinery is Trellis-core vs. instance-specific?** math-quest wires a
  PR-contract CI check + a `propagation-remediator` dispatcher; the *invariant* (SI-1) is core, but
  the CI wiring is arguably per-instance delivery — draw the line in the backport.
- **The surfacing channel:** where it lives (an `expression-profile` field vs. a separate config), and
  the **default when a project has no clear practice to infer** — does Trellis fall back to "surface
  in the change record," or refuse to assume and ask? (Leaning: ask, then a minimal default.)
- **Slug clarity:** keep `inv-propagation-surfaced` (math-quest lineage) or rename to
  `inv-improvement-surfaced` now that it's intent-defined and broader than change-propagation?
- **`inv-prune-bias` vs. B1's existing prune-bias** — consolidate to one statement, don't double-count.
- **Does this reopen the B6→B1 merge decision itself?** This ADR *refines* it (restores the concrete
  engine) rather than superseding it; confirm that's the right relationship at ratification.

## Supersedes / superseded by

— (none; refines the B6→B1 merge recorded in `invariants-v1` / `decision-0014`)

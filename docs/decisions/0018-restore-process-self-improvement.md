---
id: decision-0018
type: decision
status: ratified
ratified: 2026-07-04
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

**What a trigger looks like — concrete "points of information for improvement"** (few-shot per the iron
rule; some are GitHub-shaped — they show the *shape* of a signal, not a required tool, and each project
surfaces the ones its own practice throws off):

- a **PR that raises open questions but opens no follow-up** → the questions rot, unowned ;
- a **pipeline failing on the same step, run after run, with no corrective action** → the process
  isn't self-healing ;
- a change that **actioned a parked item without retiring it** → the exact friction that hardened this
  loop from prose to invariant (math-quest) ;
- a **direct push to `main` carrying unreviewed code** → a missing mechanical guard, not a scolding
  (math-quest wired a pre-push hook after one) ;
- a **"manually verified" flag** with no issue to automate it → tracked debt that resurfaces only by
  luck ;
- a **gate skipped** under deadline → was the gate wrong, or the deadline? ;
- **a human overriding Trellis and being right** — the gold signal, a labeled correction
  (`decision-0009`) ;
- **the same ambiguity clarified twice** → fix the upstream, don't re-clarify (B9) ;
- a **test that breaks on a valid refactor** (over-pins) → loosen or retire it ;
- a **rule / trigger that hasn't fired in a while** → candidate to retire (prune-bias) ;
- **repeated friction at the same handover** → the gate or the artifact contract needs work.

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

**The target behavior — a disposition, not only a floor.** SI-1 is the *checkable* floor (were the
signals surfaced through the chosen channel?). The real aspiration sits above it and mirrors **B3's two
faces** — a checkable conformance face plus an **intent face that lives in system prompts, weakly
checkable.** The target disposition: the agent **proactively notices** an improvement signal mid-work
and **proposes two things at once** — the immediate fix *and* a standing trigger so the class is caught
next time — **inferring the channel from the signal's own context**, and **asking**, never acting
silently (B9 / D2 consent, `decision-0009`).

*The north-star example (maintainer):* on a PR that raised open questions with nothing tracking them —

> "I noticed this PR has open questions but nothing to act on them. Want me to open GitHub issues for
> them now — and add a standing trigger so it's caught in future?"

Two things this adds beyond "surface the signal":

- **The loop bootstraps itself** — the agent doesn't only *follow* triggers, it *proposes creating*
  them ("…and in the future?"), subordinate to the human's yes.
- **Channel inference is contextual, per-occurrence** — "GitHub issues" is *inferred* here (the signal
  landed on a GitHub PR the agent could see), then offered — never globally assumed. On a non-GitHub
  project the same disposition infers a different surface.

**Honest limit (the maintainer flagged it; load-bearing, not a footnote).** This proactive-notice
disposition is **weakly checkable** — it lives in sub-agent design, system prompts, and few-shot
examples, not a mechanical gate. A meta-rule can set the surfacing floor (SI-1) and *shape* the
disposition with examples, but it cannot *guarantee* an agent reliably notices. Recorded as the
**target, not a claim**: *"not sure a meta-rule triggers this, but that's the aim."* (Same honesty as
B3's intent face and its positive-control open question.)

**Placement — DECIDED (2026-07-04): un-merge.** Restore self-improvement as a **first-class operating
invariant**, not a facet of B1. The deciding reason is a precedent in the set itself: **B3** is already
a first-class invariant with a *checkable conformance face **and** a weakly-checkable intent face* —
structurally identical to self-improvement's surfacing floor (SI-1) + proactive-notice disposition. If
B3 earns its own slot, so does this — which is *also* a whole pillar's engine (invariant 9 / Pillar II
"evolve"). And burying it as a facet already demonstrated the failure mode: it got treated as *done*
until the loss was caught. **Kept from the merge's real insight** (so this is a dial, not a clean
break): graph-maintenance and self-improvement stay documented **neighbors**, and **`inv-prune-bias`
remains the shared hinge** — retiring stale rules genuinely *is* maintenance, so it can live in B1,
while the surfacing floor + trigger loop + disposition constitute the standalone invariant. This ADR
therefore **refines** the B6→B1 merge (corrects an over-prune that lost signal), it does not supersede
graph-maintenance. Slug: reclaim `inv-self-improvement` or a clearer intent name — a small downstream
call (see Open questions).

## Consequences

*(Owed downstream, deferred until this decision ratifies — the invariant revision is the intent gate,
D2 / `decision-0014`: significant invariant changes are recorded as ADRs, then applied revise-in-place.)*

- **`invariants-v1` revised (un-merge)** — self-improvement restored as a **first-class operating
  invariant** (parallel to B3: surfacing floor SI-1 + the proactive-notice intent face), carrying
  SI-1/2/3 (intent-defined); the `inv-self-improvement` slug is un-superseded via the Identifiers
  registry (`decision-0013`); graph-maintenance (B1) noted as its neighbor, `inv-prune-bias` the shared
  hinge; the "evolve" framing (Pillar II) restored in words.
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

- ~~**The placement fork (D2).**~~ **Resolved 2026-07-04: un-merge** (first-class invariant, parallel
  to B3) — see Decision.
- **How much enforcement machinery is Trellis-core vs. instance-specific?** math-quest wires a
  PR-contract CI check + a `propagation-remediator` dispatcher; the *invariant* (SI-1) is core, but
  the CI wiring is arguably per-instance delivery — draw the line in the backport.
- **The surfacing channel:** where it lives (an `expression-profile` field vs. a separate config), and
  the **default when a project has no clear practice to infer** — does Trellis fall back to "surface
  in the change record," or refuse to assume and ask? (Leaning: ask, then a minimal default.)
- **Slug clarity:** keep `inv-propagation-surfaced` (math-quest lineage) or rename to
  `inv-improvement-surfaced` now that it's intent-defined and broader than change-propagation?
- **Instilling & verifying the disposition (the hard one).** The proactive-notice behavior is weakly
  checkable — how is it instilled (sub-agent prompt + few-shot from the trigger examples above), and
  can it be **positive-controlled** (a seeded signal the agent *should* catch)? Direct extension of the
  B3 positive-control open question in `invariants-v1`. The honest answer today: *this is a target we
  may only partly reach with rules.*
- **`inv-prune-bias` vs. B1's existing prune-bias** — consolidate to one statement, don't double-count.
- ~~**Does this reopen the B6→B1 merge?**~~ **Resolved: refines, not supersedes** — it corrects an
  over-prune (restores the invariant + concrete engine); B1/graph-maintenance stands.

## Supersedes / superseded by

— (none; refines the B6→B1 merge recorded in `invariants-v1` / `decision-0014`)

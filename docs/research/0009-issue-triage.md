---
id: research-0009
type: research-note
status: ratified
ratified: 2026-07-03
depends_on: [invariants-v1, spec-0001, decision-0009]
informed_by: [research-0005, research-0006, research-0007]
owner: gundi
---

# Research 0009 — Issue-backlog triage: five cluster-briefs (Task 1)

> **Method & honesty.** The open issues #22–#28 came from a *dry run* — pointing agents on other
> projects at Trellis to see what would help. This note groups them into five clusters and gives each
> a **problem + solution direction + next action**, using the vocabulary the research built
> (expression profile, signature catalog, supervisor/advisor, the two delivery axes). Ratified 2026-07-03 — accepting the *triage*, not a commitment to each proposed build. The
> solution directions are `inferred` design proposals, not decisions. Briefs are deliberately light —
> the point is to enable iteration, not to pre-decide builds.
>
> *Amended in place 2026-07-13 (`decision-0047` + `grove/adr-0011`; consumer-audit
> marking-class). WHAT: `research-0005`, `research-0006`, `research-0007` moved out of
> frontmatter `depends_on` into a new `informed_by` list — they supplied the vocabulary this
> triage uses without this note's own cluster-briefs being contingent on them; provenance,
> not coupling. No `version` counter on this artifact to bump. POINTER: `decision-0047`
> Consequence 4, `grove/adr-0011`.*

## The load-bearing observation (why the order below matters)

Four of the five clusters (1, 3, 4, 5) all revolve around **one shared object** the research named:
the **per-instance expression profile** (which invariants are active, at what C1 strength, gatekept
by whom, on which delivery axes) plus its product-level dictionary, the **invariant-signature
catalog** (genome annotation — what each invariant looks like when honored, #23 comment). Build those
two primitives **first**; then each cluster is a consumer of them rather than a bespoke solution.

## Cluster 1 — Application front-end (#23 assess→apply CLI · #24 RPI-Team instance)

- **Problem.** Conductor/author modes both presuppose you already know a project's shape. There is no
  front-end that *reads* a project, discovers which invariants it already honors *implicitly*, and
  makes them explicit **without clobbering**. #24 is the worked RPI-Team case (honors invariants
  1/2/4/5/6 implicitly); #23 generalizes it to a support-CLI.
- **Solution direction.** Two-phase, gated, producing an **expression profile**: **Assess**
  (read-only) detects the project's *implicit* profile against the **signature catalog**,
  confidence-tagged, loud-failure-biased ("assert-and-verify," never silently "assume honored") →
  **Apply** composes the *minimal* overlay at the human-ratified profile — **Model 1 (epigenetic
  overlay / supervisor)** by default (augment-never-clobber, `spec-0001` §5), with **advisor ×
  expressed-only** ([[research-0007]]) as the lighter, PR-reviewed alternative. Detection heuristics
  live in the catalog (`trellis-core`), not the CLI. Keep producer ≠ ratifier ≠ verifier separate.
- **Next action.** Spec the expression-profile artifact + the signature catalog first (shared with
  clusters 3/4/5), then the Assess/Apply flow. Lift #24 into `research/0004` (still owed) as the first
  instance report.

## Cluster 2 — Deferred spine machinery (#25 conformance-to-upstream · #26 self-improvement loop)

- **Problem.** `spec-0001` names but defers two mechanisms: **semantic conformance-to-upstream**
  (#25 — does an implementation match its *approved spec*, a judgment agent, distinct from the
  structural contract check) and the **friction→trigger→decision loop** (#26 — Pillar II "evolve").
  The conformance report *is* the friction substrate the loop consumes.
- **Solution direction.** #25: a read-only judgment sub-agent that derives its checklist from the
  approved upstream (never the producer), loud-failure on missing upstream. **Tie-in from
  [[research-0006]]:** for the *mechanizable* fragment it could verify against a **computed K**
  (supremal controllable sublanguage) rather than a hand-written checklist — the "compute the default
  gate-set" proposal. This is the B3 conformance face / the **observer** that makes bounded-context
  decisions checkable. #26: smallest checkable **trigger format** (condition→action, stored where it
  fires), graduation to an append-only decision, riding the PR ritual; keep the within-instance
  boundary clean vs. cluster 3.
- **Next action.** Build #25 first (it feeds #26 and materializes the deferred execution-layer
  `approved` state). Decide the `approved`-as-status vs. gate-outcome question then.

## Cluster 3 — Cross-instance validation (#28 · the N=1→N protocol)

- **Problem.** The invariants are validated on ~one project. No structured way to record what a *new*
  instance taught us, or to diff findings across instances.
- **Solution direction.** The **instance-report format = a filled-in expression profile + drift
  observations** — the object already exists ([[research-0005]]/[[research-0007]]); #28 is the
  *report schema + the cross-instance diff* on top. Diff profiles across instances (Math Quest,
  Trellis-self, RPI, the consultant-mode work usage), **weighted by instance diversity, not volume**
  (`decision-0009`). Which invariants recur across profiles = the measurable version of "genetic
  assimilation"; contested ones = candidates for revision.
- **Next action.** Define the report schema as an expression-profile + drift delta; seed it with the
  consultant-mode datapoint (memory) as entry #1. Keep the boundary with #26 (within- vs. cross-).

## Cluster 4 — In-situ tutoring (#27 per-gate explain-cards)

- **Problem.** Portable per-gate "explain cards" (why this gate, what PASS looks like, what to do when
  blocked) with expertise-adaptive suppression. Least-proven pillar; must not be written into the
  host's own files.
- **Solution direction.** Explain-cards are **derived artifacts generated from the signature catalog**
  (the shared dependency with #23 — this is the "second consumer" that graduates the catalog to a
  standalone artifact) **+ the host's expression profile**; they live Trellis-owned, track provenance,
  and **regenerate/supersede on process change** (never silently drift). **Delivery caveat
  ([[research-0007]]):** tutoring needs a runtime, so it is a **supervisor-mode** feature —
  advisor/pull mode can't tutor live. Scope as an experiment; validate it isn't noise.
- **Next action.** Defer until the catalog exists (cluster 1); then a small experiment on one host's
  gates. Interactive-only.

## Cluster 5 — Partial application / packaging (#22 "Trellis-lite")

- **Problem.** The behavioral subset (`inv-independent-judgment` intent-face, B9, D1-disposition) has
  no installable artifact independent of the pipeline — so it gets copied locally and drifts, or
  skipped.
- **Solution direction.** Specify "Trellis-lite" as a **named default expression profile** (a
  differentiated cell type, [[research-0005]]), *not* an ad-hoc rule list — the behavioral genes
  travel precisely because they need no machinery. It is the **advisor × expressed-only** cell
  ([[research-0007]]) — the lowest-friction wedge. #22's own "which granularity" ambiguity is
  dissolved by the two axes: *which* invariants are expressed (portable) vs. *how* each is
  instantiated (the heavy pipeline). This cluster is the seed that generated the whole model.
- **Next action.** Once the profile artifact exists, ship Trellis-lite as a shelf profile + a
  system-prompt fragment; use it as the `business-systems-roadmap` N≥2 probe #22 proposes.

## Recommended build order (the sequencing the shared primitive implies)

1. **Foundation:** the **expression-profile artifact** + the **invariant-signature catalog** (typed
   artifacts via a recorded decision; `research-0003` types are open). Unblocks 1, 3, 4, 5.
2. **Cluster 1** (Assess/Apply) — the first producer of profiles.
3. **Cluster 5** (Trellis-lite) — nearly free once the profile object exists; the wedge.
4. **Cluster 3** (instance reports) — profiles + drift; turns on the N=1→N loop.
5. **Cluster 2** (#25 then #26) — enforcement + evolution machinery; semi-parallel.
6. **Cluster 4** (tutoring) — experiment, last; supervisor-mode only.

## Acceptance criteria

*(research-note: Open questions + sources/confidence; no acceptance-criteria gate.)*

## Open questions

- Do the two shared primitives (expression profile, signature catalog) collapse into one typed
  artifact, or stay two (dictionary vs. per-instance readout)? ([[research-0005]] "two objects".)
- Per-cluster, which briefs graduate to specs/decisions first vs. stay parked? (The build order is a
  proposal, not a commitment.)
- Are any of these `trellis-core` (invariant/catalog changes — highest bar, cross-instance recurrence)
  vs. per-project? (`decision-0009` three-tier routing.)
- Sources: issues #22–#28 + their comments (the dry-run corpus); [[research-0005]]/[[research-0006]]/[[research-0007]]; `spec-0001`; `decision-0009`. `verified` (issue corpus); solution directions `inferred`.

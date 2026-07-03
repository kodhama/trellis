---
id: research-0007
type: research-note
status: ratified
ratified: 2026-07-03
depends_on: [invariants-v1, decision-0002, decision-0008, decision-0010, decision-0012, research-0005, research-0006]
owner: gundi
---

# Research 0007 — Delivery & activation control: the operating-model axes (deliverable 2.4)

> **Method & honesty.** This consolidates the delivery/operating-model thread developed with the
> maintainer (2026-07), which had overgrown [[research-0006]] Result 6. It is a **design synthesis
> with an unresolved intent-layer fork** (which on-ramp to build first) — so it is a `research-note`
> proposal, *not* a decision; when the fork is called it becomes a decision extending `decision-0012`.
> Tags: `verified` = grounded in a cited artifact/source; `inferred` = my/our design reasoning.
> Answers 2.4 ("possible delivery mechanism + how to control the activation dials").

## What 2.4 has to answer

`decision-0012` set delivery as *plugin marketplace → support CLI → git-copy*, all framed as
**activation strength** (how strongly an *installed* Trellis is wired into the host). Three refinements
since show that frame is one corner of a larger space: delivery is **two orthogonal axes**, and "how
the dials are controlled" has a concrete answer (an artifact + a gate flow). No runtime, throughout
(`decision-0010`).

## The two axes

- **Axis A — delivery *relationship*.** *Supervisor* (push): Trellis installed and live, constrains at
  runtime, stays current via the delivery channel, cleanly removable. *Advisor* (pull): Trellis not
  installed; the host's own agents consult it as an external reference; effect is baked in, nothing to
  review/secure/remove at runtime. (Extends `decision-0012`'s lowest "available+referenced" rung
  outward into a genuine pull relationship.) `inferred`.
- **Axis B — payload *depth* (what is transferred).** Three rungs:
  1. **Expressed genes only** — just the invariants active in this instance's profile.
  2. **+ Latent (inert) genes** — the rest of the gene set, dormant but present (expressible later).
  3. **+ the expression mechanism** — the regulatory apparatus (the dials + the means to change which
     genes express). **This rung unlocks self-regulation:** an instance carrying the mechanism can
     re-regulate its own genes over time — the adaptability/autonomy-ratchet property ([[research-0006]]
     §6). Adaptability is thus a *delivery* choice, governed by the dials + D2. `inferred`.

The axes are **orthogonal** (maintainer's key correction): payload is not locked to relationship.

## The grid (illustrative cells, all reachable)

| | **Expressed-only** | **+ mechanism (self-regulating)** |
|---|---|---|
| **Supervisor** (push/live) | fixed live enforcer of the chosen genes; no self-regulation | **full live regulator** — self-regulating, adaptable, removable (the "rich" supervisor) |
| **Advisor** (pull/one-shot) | **minimal graft → reviewable PRs** (maintainer's working default); team-reviewed, no runtime, no self-regulation | **full graft** — bakes the whole apparatus in as a one-shot; host self-regulates *without* Trellis present |

Two cells have real backing already: **advisor × expressed-only** is the modality the maintainer
already runs at work (`verified`, [[consultant-mode-work-usage]]); **supervisor × +mechanism** is
`decision-0012`'s implicit default (plugin auto-activation + the C dials). The off-diagonals are the
maintainer's point that neither is forced.

**Caveat — the axes converge at the top rung (Fable review).** Orthogonality is clean for
*expressed-only* and *+inert* payloads, but **breaks at +mechanism**: self-regulation needs
*something running*, and if it runs in the host, the host now **contains a supervisor**. So
"advisor × +mechanism" really means *a advisor that installs a host-native supervisor and
leaves* — the cell collapses toward Axis-A's supervisor pole. At +mechanism the live question is no
longer push-vs-pull but **who operates the runtime** (Trellis-vendor vs. host). Read the grid as two
axes for light payloads, converging to one at the self-regulating rung.

## Independence in advisor mode (resolving the B3 concern)

The earlier worry — "advisor mode has no runtime, so independent verification (B3/invariant-5)
isn't automatic" — is **answered for application-time** and *partly* for ongoing:

- **Application-time (partly resolved — softened per Fable review):** advisor changes land as
  **PRs the host team reviews**, and small **expressed-only diffs** are what make that review real.
  That review supplies *independent eyes* (the reviewers have no stake in Trellis being right) — but it
  checks **fit-for-this-project**, not **fidelity-to-the-invariants** (the reviewers aren't Trellis
  experts). So it satisfies the *spirit* of B3 (independence) but not necessarily the *content* of
  invariant-5 (conformance to the approved upstream). Genuine, but weaker than "the same property."
- **Ongoing (residual):** PR review checks the *change*, not the conformance of *future* work. Ongoing
  B3 requires the payload to graft a **host-native checker** (Axis B ≥ some mechanism, or a one-shot
  CI install) — else invariant-5 silently lapses after the graft. `inferred`.

## How the activation dials are *controlled* (the concrete answer)

The dials are not a UI knob; they are set through an artifact + a gate flow that already fall out of
the corpus:

- **The control surface is the expression profile** ([[research-0005]]) / the supervisor's control map
  ([[research-0006]]) — a per-instance declaration of *which invariants are active, at what C1 strength,
  gatekept by whom (C2), on which delivery axes (A/B)*.
- **The flow:** Assess (#23) *proposes* a profile from the project's environment → **the human ratifies
  it (D2)** — never silently maximal (`decision-0008`/`spec-0001` §5) → the delivery layer
  (`decision-0012`, no runtime `decision-0010`) *composes* exactly that profile into the host. Producer
  ≠ ratifier ≠ verifier stays intact (invariant 5).
- **The authoring interface can be natural.** Neto 2010's closing vision — specify the constraints in a
  natural form (temporal logic) that *compiles* to the supervisor (Lacerda & Lima 2008) — is a concrete
  model: the human states intent naturally; it compiles to the expression profile + enforcement.
  `verified` (thesis); the Trellis application `inferred`.

## The strategic fork (intent-layer — the maintainer's to call)

Which cell is the **first on-ramp / shipped default** is a product-strategy decision (D2, not mine):

- **Supervisor × +mechanism** — the richest, most on-thesis (live gates, self-regulation), but highest
  install/compliance friction — the enterprise blocker the brief §7 names.
- **Advisor × expressed-only** — lowest friction (referencing a repo sits *below* the compliance
  radar an installed dependency trips), reviewable, team-verified — and it already has a real
  (genealogically-tainted) usage datapoint. But it forgoes live enforcement and self-regulation, and
  goes stale.

Appetite for each is a genuine unknown. My read (`inferred`, for the maintainer to judge; **tempered
per Fable review**): advisor × expressed-only is a real *wedge for reach*, but its "low friction"
is **softer and more fragile** than it first looked — three discounts: (a) an external repo your
coding agents *read and apply* is a live **supply-chain / prompt-injection surface**, so "below the
compliance radar" means the risk is *unrecognized, not absent* — **arbitrage that closes** the moment
enterprise security notices agents pulling instructions from an outside repo; (b) host PR review
checks *fit*, not *invariant-fidelity* (above); (c) advisor grafts **go stale**, so
"land-advisor-then-expand-to-supervisor" inherits drift — the upsell migration may be *harder*,
not cleaner, than starting as a supervisor. Net: pursue advisor mode for **reach and
low-commitment trials**, but do **not** lean on "low friction" as durable or safe, and treat
wedge→depth as an open migration problem, not a clean upsell. The honest state is *two live options*,
not a settled sequence.

## Acceptance criteria

*(research-note: Open questions + sources/confidence; no acceptance-criteria gate.)*

## Sources

- `decision-0012` (delivery), `decision-0010` (no runtime), `decision-0008` (dials), `spec-0001` §5
  (activation/wiring, augment-not-clobber), `decision-0002` (adopt/adapt). `verified` (corpus).
- Neto 2010 (natural-spec→compiled supervisor; Lacerda & Lima 2008). `verified` (thesis, §7).
- Consultant-mode work datapoint: [[consultant-mode-work-usage]]. `verified` (user report; genealogically
  tainted — see the memory).

## Open questions

- **The fork:** which cell ships first (intent-layer, owed to the maintainer). "Land-advisor,
  expand-supervisor" is a hypothesis, not a decision.
- **Is Axis B three rungs or a continuum?** ("+ latent genes" without the mechanism may be a thin slice.)
- **Does advisor × expressed-only satisfy enough invariants to still be "Trellis,"** or is it
  Trellis-lite (#22) under a delivery name? (Ties the delivery fork to the partial-application question.)
- **Staleness protocol for advisor grafts** — how a pull-delivered result is re-synced when upstream
  Trellis evolves (the inverse of the supervisor's auto-currency).
- **Supply-chain / prompt-injection surface of advisor mode** (Fable review): an external repo
  that host agents read and apply is an attack surface; what integrity / pinning / signing story
  makes pull-delivery safe once it is on security's radar? Load-bearing for the wedge's durability.
- When the fork is called, this note graduates a **decision** extending `decision-0012` (append-only).

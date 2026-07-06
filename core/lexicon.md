---
id: lexicon-v1
type: lexicon
status: ratified
depends_on: [decision-0017, invariants-v1]
owner: gundi
scope: trellis-product
ratified: 2026-07-03
---

# Trellis lexicon — v1 (the Rosetta store)

> **What this is.** The one place the **equivalences** live. Trellis reasons through several lenses —
> the horticultural product identity, the **genetics** lens (`research-0005`), the **DES / supervisory-
> control** lens (`research-0006`), and the **delivery** lens (`research-0007`). Each names the same
> referents differently *on purpose*. This lexicon records, per concept: the **canonical** term Trellis
> uses in current-truth/product artifacts, its **lens synonyms**, a one-line definition, and where it is
> authoritative. Policy + type introduced by `decision-0017`.
>
> **Three registers that nest (`decision-0017`).** *The plant on the trellis expresses its genes, and
> its genes are its invariants* — the three vocabularies are layers, not rivals:
> - **garden** = identity/relationship (Trellis, host-as-plant, the delivery dial) — the product's face;
> - **gene** = mechanism + **official teaching register** (expression profile, active/latent genes,
>   catalog = genome annotation) — how the machinery is *conveyed* (gene expression reads easier than
>   "invariant"); promoted from "internal only" (`research-0008`);
> - **invariant** = the precise substrate — **canonical for what is enforced.**
> **Caveat:** gene does *not* go fully canonical — the analogy breaks at `floor-intent-gate`
> (`research-0005` §Limits:
> "no intent locus in a genome"), and the naming guardrail (`brief-§4`) forbids borrowing its authority;
> gene-talk stays *our-synthesis teaching metaphor*, never a provenance claim. Research notes keep their
> lens vocabulary; current-truth/product artifacts use the **canonical** column and link here.

> **Ratified 2026-07-03 (`floor-intent-gate`).** The maintainer confirmed the three-register model and the delivery
> dial names (`supervisor` / `advisor`); the `consultant → advisor` sweep is done.

## Canonical terms

| Canonical | Genetics lens | DES / control lens | Delivery lens | Definition · authoritative in |
|---|---|---|---|---|
| **the invariants** (invariant set) | gene library / gene set | specification *K* | — | the full rule set Trellis can express · `invariants-v1` |
| **invariant** | gene | (a legal-behavior constraint) | — | one rule, by stable slug · `invariants-v1` |
| **active** (expressed) | gene expression / expressed | enabled event | active | an invariant enforced in a given instance · `signature-catalog-v1`, profiles |
| **latent** | latent / inert gene | (disabled event) | latent payload | present but not active here · `research-0007` |
| **expression profile** | expression profile | control map / control policy | the profile | the per-instance readout (which invariants active × strength × gatekeeper × delivery) · `spec-0002` |
| **signature catalog** | genome annotation | — | — | the product-level dictionary of each invariant's "tells" · `signature-catalog-v1` |
| **enforcement strength** (`dial-enforcement-strength`) | (expression level) | (policy parameter) | activation level | how strictly a gate is enforced: expressed → default-on-but-skippable → enforced · `decision-0008` |
| **gatekeeper** (`dial-gatekeeper`) | — | modular / decentralized supervisor | — | who applies a gate: independent-agent · human · none · `decision-0008` |
| **floors** (`floor-transparency`, `floor-intent-gate`) | — | uncontrollable-event routing (the intent-gate floor) | — | the non-configurable minimums: transparency, the intent gate · `invariants-v1` |
| **supervisor** | (regulator) | supervisor *S* | supervisor (live/push end) | Trellis in its constraining role — enables/blocks the next steps · `research-0006` |
| **host** | organism | plant *G* | host | the project Trellis supervises · `research-0006` |
| **observer** | — | observer / state estimator | — | the estimate of project state a bounded-context op acts on = the `depends_on` graph · `research-0006` |
| **overlay** (Model 1) | epigenetic overlay | (external supervision) | — | apply Trellis *without editing the host's files* (augment-never-clobber) · `research-0005/0006` |
| **morph** (Model 2) | genome edit / genetic assimilation | modify the plant | — | apply Trellis by *rewriting the host's methodology* (baked in) · `research-0006` §R6 |
| **supervisor** (dial end) | — | — | supervisor (push/installed/live) | delivery relationship: Trellis installed and running live · `research-0007` |
| **advisor** *(formerly "consultant")* | — | — | (was consultant) | delivery relationship: Trellis *consulted* from outside, no live authority; the host internalizes its guidance and acts on its own — no runtime tie · `research-0007` |

*`advisor` (ratified 2026-07-03) is the role-noun parallel to `supervisor`; `consultant` is retired
(kept only in `decision-0017` as the historical record, and in the `consultant-mode-work-usage` memory
name). The two ends are asymmetric — live = active agent, pull = consulted source — so the pair is
role-register, not garden (`decision-0017`).*

## Notes on two overloads (why the mapping is what it is)

- **`supervisor` appears twice** (DES role · delivery dial end) *deliberately* — the live-delivery
  supervisor **is** Trellis performing the DES supervisor role. One metaphor, two zoom levels; kept as
  one word (`decision-0017`).
- **relationship ≠ application-model.** The *delivery relationship* (supervisor ↔ advisor: is Trellis
  present and running?) is distinct from the *application model* (overlay ↔ morph: does Trellis edit
  the host?). They correlate (an advisor tends toward morph) but are separate axes — do not conflate
  them (`research-0007` "the axes converge only at the +mechanism rung").
- **the dial is a role-noun pair, on purpose.** `supervisor` / `advisor` are matched agent-role nouns
  because the two ends are asymmetric — the live end is an active agent, the pull end is a consulted
  source. The garden/gene metaphors live in the identity + expression registers, not on this technical
  axis (`decision-0017`).

## Open questions

- **Payload-depth (Axis B)** terms (`expressed-only / +latent / +mechanism`) are not yet
  canonicalized — parked until they prove confusing.
- **Versioning:** is one revise-in-place `lexicon-v1` right, or should canonical-term changes be
  ADR-tracked like the invariant set (`decision-0014`)? Decide if the vocabulary starts churning.

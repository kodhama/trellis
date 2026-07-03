---
id: lexicon-v1
type: lexicon
status: draft
depends_on: [decision-0017, invariants-v1]
owner: gundi
scope: trellis-product
---

# Trellis lexicon — v1 (the Rosetta store)

> **What this is.** The one place the **equivalences** live. Trellis reasons through several lenses —
> the horticultural product identity, the **genetics** lens (`research-0005`), the **DES / supervisory-
> control** lens (`research-0006`), and the **delivery** lens (`research-0007`). Each names the same
> referents differently *on purpose*. This lexicon records, per concept: the **canonical** term Trellis
> uses in current-truth/product artifacts, its **lens synonyms**, a one-line definition, and where it is
> authoritative. Policy + type introduced by `decision-0017`.
>
> **Store, don't flatten (`decision-0017`).** Research notes keep their lens vocabulary — that is their
> analytical value. Current-truth/product artifacts (invariants, catalog, profile, specs, `CLAUDE.md`)
> use the **canonical** column and link here.

> **Status `draft` — awaiting ratification (D2).** The agent authored this; the builder does not
> self-ratify (B3). The one term still *under decision* is the pull-end dial name (see the ⚑ row and
> `decision-0017` open questions).

## Canonical terms

| Canonical | Genetics lens | DES / control lens | Delivery lens | Definition · authoritative in |
|---|---|---|---|---|
| **the invariants** (invariant set) | gene library / gene set | specification *K* | — | the full rule set Trellis can express · `invariants-v1` |
| **invariant** | gene | (a legal-behavior constraint) | — | one rule, by stable slug · `invariants-v1` |
| **active** (expressed) | gene expression / expressed | enabled event | active | an invariant enforced in a given instance · `signature-catalog-v1`, profiles |
| **latent** | latent / inert gene | (disabled event) | latent payload | present but not active here · `research-0007` |
| **expression profile** | expression profile | control map / control policy | the profile | the per-instance readout (which invariants active × C1 × C2 × delivery) · `spec-0002` |
| **signature catalog** | genome annotation | — | — | the product-level dictionary of each invariant's "tells" · `signature-catalog-v1` |
| **enforcement strength (C1)** | (expression level) | (policy parameter) | activation level | how strictly a gate is enforced: expressed → default-on-but-skippable → enforced · `decision-0008` |
| **gatekeeper (C2)** | — | modular / decentralized supervisor | — | who applies a gate: independent-agent · human · none · `decision-0008` |
| **floors (D1, D2)** | — | uncontrollable-event routing (D2) | — | the non-configurable minimums: transparency (D1), intent gate (D2) · `invariants-v1` |
| **supervisor** | (regulator) | supervisor *S* | supervisor (live/push end) | Trellis in its constraining role — enables/blocks the next steps · `research-0006` |
| **host** | organism | plant *G* | host | the project Trellis supervises · `research-0006` |
| **observer** | — | observer / state estimator | — | the estimate of project state a bounded-context op acts on = the `depends_on` graph · `research-0006` |
| **overlay** (Model 1) | epigenetic overlay | (external supervision) | — | apply Trellis *without editing the host's files* (augment-never-clobber) · `research-0005/0006` |
| **morph** (Model 2) | genome edit / genetic assimilation | modify the plant | — | apply Trellis by *rewriting the host's methodology* (baked in) · `research-0006` §R6 |
| **supervisor** (dial end) | — | — | supervisor (push/installed/live) | delivery relationship: Trellis installed and running live · `research-0007` |
| ⚑ **cutting** *(proposed — was "consultant")* | (a taken cutting) | — | consultant (pull/referenced) | delivery relationship: Trellis *referenced* from outside, its shape rooted in the host, grows independently, no runtime tie · `research-0007` |

*⚑ = the one term still under decision (`decision-0017`). Until ratified, the corpus still says
"consultant"; this lexicon records `consultant → cutting` as the proposed canonical.*

## Notes on two overloads (why the mapping is what it is)

- **`supervisor` appears twice** (DES role · delivery dial end) *deliberately* — the live-delivery
  supervisor **is** Trellis performing the DES supervisor role. One metaphor, two zoom levels; kept as
  one word (`decision-0017`).
- **relationship ≠ application-model.** The *delivery relationship* (supervisor ↔ cutting: is Trellis
  present?) is distinct from the *application model* (overlay ↔ morph: does Trellis edit the host?).
  They correlate (a cutting tends toward morph) but are separate axes — do not conflate them
  (`research-0007` "the axes converge only at the +mechanism rung").

## Open questions

- **The pull-end dial name** — `cutting` (proposed) vs `reference` (clear but collides with
  `inv-reference-relationship`/B8) vs `resident`/`reference` (renames both ends, fully de-overloads
  "supervisor") vs keep `consultant`. Owed to `decision-0017` ratification (D2).
- **Payload-depth (Axis B)** terms (`expressed-only / +latent / +mechanism`) are not yet
  canonicalized — parked until they prove confusing.
- **Versioning:** is one revise-in-place `lexicon-v1` right, or should canonical-term changes be
  ADR-tracked like the invariant set (`decision-0014`)? Decide if the vocabulary starts churning.

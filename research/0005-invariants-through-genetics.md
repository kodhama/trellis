---
id: research-0005
type: research-note
status: ratified
ratified: 2026-07-03
depends_on: [invariants-v1, decision-0002, decision-0008, decision-0012]
informed_by: [research-0003]
owner: gundi
---

# Research 0005 — The invariants through a genetics lens (partial application as gene expression)

> **Method & honesty (load-bearing).** This is an **analogy**, offered as *our synthesis*, not a
> claim of biological provenance for Trellis. Two things are kept strictly separate: **what the
> biology says** (externally sourced, confidence-tagged below) and **the mapping onto Trellis's
> invariants** (my inference — every mapping row is tagged `inferred` or `speculated`). The
> external biology was gate-tested by an adversarial-verification research pass (`w5t766osj`);
> several *stronger* biological claims were **refuted** and are deliberately excluded (see
> §Limits). An analogy that is oversold becomes the vaporware the iron rule (brief §7) warns
> against — so this note carries its own "where it breaks" section as a first-class part.
> Seeded by issue #22 (Trellis-lite). Companion: [[research-0006]] (the supervisor / DES lens).
>
> *Amended in place 2026-07-13 (`decision-0047` + `grove/adr-0011`; consumer-audit
> marking-class). WHAT: `research-0003` moved out of frontmatter `depends_on` into a new
> `informed_by` list — the artifact-type-taxonomy note informed this note's framing without
> this note's genetics analogy being contingent on it; provenance, not coupling. No `version`
> counter on this artifact to bump. POINTER: `decision-0047` Consequence 4,
> `grove/adr-0011`.*

## Why this lens at all

Issue #22 observed that a **subset** of the invariants (`inv-independent-judgment` intent-face,
`inv-clarify-before-commit` B9, `floor-transparency` D1-as-disposition) is meaningful **without
any pipeline machinery** — it installs as system-prompt behavior and travels independently. That
is not "Trellis minus features." It is a different *thing* built from the *same* set — which is
exactly what biology does with one genome. This note asks what the genome/expression/epigenetics
frame *buys* us, and (§Limits) where it costs us.

## The core analogy

| Biology | Trellis | Tag |
|---|---|---|
| **Gene library** — the full complement of *available* genes | The full **invariant set** (`invariants-v1`: A1–A4, B1–B9, C, D) — genes Trellis can express *in a host* | `inferred` |
| **Gene expression** — a gene being transcribed/translated *here, now* | An invariant being **active/enforced** in a given instance | `inferred` |
| **Expression profile** — *which* genes a given cell transcribes | The per-instance set of active invariants + their levels (see [[research-0006]] "control map") | `inferred` |
| **Cell differentiation** — one genome → many cell types, each expressing a different subset | One invariant set → many **instance shapes** (a 2-session non-code project vs. a regulated enterprise pipeline) expressing different subsets | `inferred` |
| **Trellis-lite (#22)** — the behavioral subset | A **differentiated cell type** that expresses only the housekeeping/behavioral genes | `inferred` |

The payoff of the frame is not the labels — it is the three structural results below, each of
which *does work* the current model does not.

**Which genome? (referent — pinned per Fable review.)** Two referents were being swapped (Trellis's
invariant set vs. the host's methodology). Pinned convention for this note: the **organism is the
host project**, its **genome is its own methodology/files**, and Trellis's **invariants are a gene
library expressed *in* that host**. So *cell differentiation* = different hosts express different
invariant-subsets; **Model 1 (overlay)** = the genes stay **episomal** (present and regulated, not
written into host DNA); **Model 2 (morph)** = the genes are **integrated into the host genome** (its
methodology is rewritten). One genome — the host's; Trellis supplies genes, not a second genome.

## Result 1 — the dial is coherent, not a category error (illustrates C1; does *not* ground it)

The sharpest empirically-grounded result answers a question Trellis's dials assume but never
justified: *is "how strongly enforced" even a coherent continuum, or a binary dressed up?*

Biology's answer is **both, and the resolution is instructive.** At the single-locus level a
promoter is **switch-like** — it toggles ON/OFF in stochastic "bursts" (the two-state / telegraph
model). The **graded** expression *level* emerges as the time- and population-average of those
discrete switch events. `verified` (Zhang 2024; Chen et al. 2023; resting on Golding/Raj/Suter).

More useful still: the graded level is set by **two separable, independently-tunable knobs mapped
to distinct regulatory machinery** — burst *frequency* (set primarily by **enhancers**) and burst
*size* (set primarily by **core promoters**). `verified` (Larsson 2019; Chen 2023).

- **Mapping (`inferred`) — illustration, not grounding (walked back, Fable review).** The defensible
  half: graded control *built from discrete on/off events* is a real, coherent thing in nature — so
  C1 (`expressed → default-on-but-skippable → enforced`) being a graded dial made of discrete
  gate-firings is **not a category error**. That is *illustration* (the shape is natural),
  **not grounding** (C1 stands or falls on its own design merits, not on biology). **Dropped as
  overclaim:** the earlier "C1 = burst-frequency, C2 = burst-size, so the dials are independent."
  Frequency and size are two knobs on the **same** dimension (*amount* of transcription); C1
  (strength) vs. C2 (*who* gatekeeps) is **amount vs. identity** — the orthogonality only rhymes, and
  biology cannot adjudicate whether Trellis's two dials are independent.

## Result 2 — epigenetics = "augment, never clobber" (grounds the overlay)

The overlay contract Trellis keeps stumbling on — `spec-0001` §5 and issues #23/#27 all insist the
application **augments, never clobbers** the host's own files — is, structurally, **epigenetic
regulation**: expression is changed *without editing the underlying sequence*.

- **The robust half (use it):** epigenetic marks (DNA methylation, histone modification, chromatin
  state) alter *which genes express and how strongly* **without changing the DNA**. That is
  precisely composing a Trellis overlay onto a host `CLAUDE.md`/methodology: you set the host's
  *expression profile* without rewriting its source. `verified` that regulation-without-substrate-
  change is the definitional property.
- **The contested half (do NOT lean on it):** the popular "epigenetics = *heritable*,
  *environment-responsive* memory" framing is **genuinely contested in the literature** —
  heritability is the central fault line, and several strong causal claims (the environment
  *directly reprograms* marks; robust transgenerational inheritance) were **adversarially refuted**
  in our research pass. `verified` that the definition is unsettled.
- **Mapping (`inferred`):** adopt "epigenetic" for the *mechanism* (regulate expression without
  editing the substrate = augment-not-clobber). Do **not** adopt it to argue overlays are
  "inherited" or "environmentally learned" — that borrows contested biology to manufacture
  authority, the exact move the naming guardrail forbids.
- **The mode this exposes — a config axis (maintainer, 2026-07).** Epigenetics is regulation
  *without* editing DNA — but organisms also do **genome editing**, and **genetic assimilation** (an
  environment-induced trait becoming hard-wired; `verified` def. in Sources). Those are Trellis's two
  application modes: **Model 1 — supervise / epigenetic overlay (default):** set the host's expression
  profile without touching its files (augment-never-clobber, `spec-0001` §5) — separable, reversible.
  **Model 2 — morph / genome edit (deferred option):** rewrite the host's methodology to *bake in* the
  invariants (= genetic assimilation of the overlay) — the host then honors them alone, but it is not
  cleanly reversible and entangles Trellis with host content. Trade: **separability (M1) vs.
  internalization (M2)**. Maintainer's call: **default M1; M2 an option, not the first focus.** Floors
  still bind M2 — a rewrite must be surfaced (D1) and human-ratified (D2), never silent. Distinct from
  adopt/adapt (B8 = *sourcing* the methodology; this = *whether Trellis edits the host*). Owed to
  2.3/2.4. `inferred`.

## Result 3 — it resolves #22's actual confusion (two axes, not one subset)

#22's own comment flags that "portable subset" is ambiguous between two cuts (epistemic-subset vs.
gate-machinery; and portable-*principles* vs. dev-*instantiation*) and calls it "really about
granularity." The genome frame names the two axes cleanly and dissolves the ambiguity:

- **Axis 1 — *which* invariants are expressed** (which genes are switched on). Portable. A
  short non-code project expresses the behavioral genes and leaves the pipeline genes silent.
- **Axis 2 — *how* each expressed invariant is instantiated** (expression *level* + the
  dev-flavored machinery that transcribes it). The "gate at every handover" *gene* travels
  (Axis 1); its heavy pipeline *instantiation* is an expression *level* (Axis 2), overhead for a
  2-session project.

#22's two "cuts" are these two axes conflated. `inferred`. **Consequence:** "Trellis-lite" should
be specified as *an expression profile* (a genotype: which genes on, at what level), not as a
hand-picked list of three rules — which makes it a special case of the same object #23/#28 need,
not a separate artifact.

## Two objects the frame separates (and Trellis already half-named)

- **The invariant-signature catalog** (your term, #23 comment) = **genome annotation**: for each
  gene, what it *is* and how to *recognize* it when expressed. One, product-level. `decision-0009`
  already lists "catalog" as `Trellis-core` content.
- **The per-instance expression profile** (this note's framing) = a **differentiated cell's
  expression readout**: which genes are on *here*, at what level. One per instance. This is what
  #23 detects, #24 fills in, #28 diffs, #22 minimizes.

Catalog : profile :: reference-genome-annotation : single-cell-expression-profile. `inferred`.

## Limits — where the analogy breaks (first-class, per the honesty rule)

- **No intent locus in a genome.** A3/D2 (humans own intent; the intent gate never opens) has
  *no* genetic analogue — genes have no "accountable owner," no place a *wrong* upstream is caught.
  This is exactly the gap the **supervisor lens fills** ([[research-0006]] §2): the genetics frame
  is silent on the single most load-bearing floor. Do not stretch it there.
- **Heritability is contested** — so the frame cannot ground cross-instance learning (#28). That
  is the DES/RL lens's job, not this one.
- **Genes optimize nothing.** Expression is regulated, not *chosen against a criterion*. The
  "which enforcement is right here" decision is a control/optimization question ([[research-0006]]),
  not a regulatory one.
- **Morphogen "selection" is contested mechanism.** The clean "environmental signal SELECTS the
  program via concentration thresholds" (French-flag) story is the *classical* model; the literal
  threshold mechanism is disputed. `verified` it is contested. Use it as illustration, not proof,
  that #23's Assess-from-environment has biological precedent.
- **Analogy ≠ identity.** The frame earns its keep on Results 1–3 (a decomposable dial, the
  overlay contract, and #22's resolution). Everything else is illustration.

## Sources

- Transcriptional bursting; graded level from ON/OFF averaging: PMC11437526 (Zhang 2024);
  arXiv:2304.08770 (Chen et al. 2023). `verified` (3-0 adversarial, two independent primaries).
- Separable knobs (frequency←enhancers, size←core-promoters): PMC11437526 (Larsson 2019 Nature);
  arXiv:2304.08770. `verified`. *Caveat: "primarily, not perfectly" orthogonal; two-state model.*
- Multimodal analog+digital control: PNAS 10.1073/pnas.1912500117 (Lammers 2020). `verified`
  (single developmental system — generalization is `inferred`).
- "Epigenetics" definition genuinely contested: Frontiers freae.2024.1434253 (2024); Heredity
  hdy201054 (2010); Bird 2007. `verified` (that it is contested).
- Epigenome plastic/environment-responsive (directional only): Heredity hdy201054. `verified`
  as directional; **direct environmental reprogramming and transgenerational inheritance were
  REFUTED** as over-strong — excluded.
- Morphogen/French-flag continuous→discrete; threshold mechanism contested: Royal Soc Open Biology
  12/220224 (2022); Development dev185967 (Sharpe 2019). `verified` (classical model; mechanism
  disputed).
- Genetic assimilation (env-induced phenotype → constitutive): PNAS 10.1073/pnas.2309760120 (2023).
  `verified` (definition); **concept itself contested** — flagged.

## Open questions

- Does the "C1 = burst frequency / C2 = burst size" decomposition (Result 1) actually hold — are
  the two dials empirically independent in use, or does gatekeeper choice drag enforcement
  strength? Testable against instance #2.
- Is "expression profile" worth promoting to a **typed artifact** (`research-0003` says types are
  open + a recorded decision) — or does it collapse into [[research-0006]]'s control map? (They
  may be one object under two lenses; see the naming/design decision, deferred.)
- Should "Trellis-lite" (#22) ship as a *named default profile* (a differentiated cell type on the
  shelf) rather than an ad-hoc rule list? Leaning yes; owed to the #22 triage.
- Where does the genetics frame stop being load-bearing and start being decoration? (The honesty
  discipline says name this boundary explicitly before adopting any of it into the product.)

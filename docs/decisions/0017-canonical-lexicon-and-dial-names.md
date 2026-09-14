---
id: decision-0017
type: decision
status: ratified
ratified: 2026-07-03
depends_on: [invariants-v1, research-0005, research-0006, research-0007, research-0008, decision-0015, brief-§4]
owner: gundi
date: 2026-07-03
---

# 0017 — A canonical lexicon + names for the delivery-relationship dial

**Raised by:** the maintainer — Trellis now carries several vocabularies for the *same* referents
(the horticultural product identity; the genetics lens `research-0005`; the DES/supervisor lens
`research-0006`; the delivery lens `research-0007`). The synonyms are *deliberate* (each lens is an
analytical tool), but the **equivalences are unstored**, and two delivery terms are unsatisfying:
"consultant" is only-ever-a-reference, and "supervisor" is **overloaded**.

## Context

A terminology survey of the corpus (2026-07-03) found: `supervisor` in 11 files (70 hits), `consultant`
in 6 (38), `gene`/genetics vocabulary in 15 files (102) — already woven into *current-truth* artifacts
(the catalog says "genome annotation," the profile says "gene"), not only the research notes. Two
problems:

1. **No store of equivalences.** `expression profile` = `control map`, `overlay` = `epigenetic`,
   `morph` = `genetic assimilation` — the corpus asserts these convergences in prose
   (`research-0006` §Convergence) but nowhere records them as a lookup.
2. **The delivery-relationship dial is badly named.** Its two ends are "supervisor" (installed/live)
   and "consultant" (referenced/pulled). "supervisor" *also* names Trellis's DES control role
   (`research-0006`); "consultant" was never a chosen term — the maintainer used it only as a
   rhetorical pointer to convey the pull idea. And the pair is **grammatically inconsistent** — an
   agent-role noun paired with a gerund/object.

The maintainer's steer: normalize around **Trellis's identity + the invariant/gene idea**, but
**keep the lenses** — store the equivalence, do not flatten.

## Decision

1. **Introduce a `lexicon` type** (`scope: trellis-product`, one shipped — `lexicon-v1`). The
   canonical-term registry: each concept → its canonical name + cross-lens synonyms + a one-line
   definition + where it is authoritative. Types are open, extended by a recorded decision
   (`research-0003`, as `decision-0016` did for catalog/profile). Required sections:
   `## Canonical terms`, `## Open questions`.

2. **Three registers, not one (store, don't flatten).** Trellis speaks in three registers that
   **nest**, not compete — *the plant on the trellis expresses its genes, and its genes are its
   invariants*:
   - **Identity / relationship — the garden register.** Trellis, host-as-plant, the delivery dial.
     The product's face.
   - **Mechanism / teaching — the gene register.** Expression profile, active vs. latent genes,
     catalog = genome annotation. **Promoted here from "internal vocabulary" (`research-0008`) to the
     official teaching register — external included** — because *gene expression conveys the machinery
     more easily than "invariant."** Refines `research-0008`'s "genetics stays internal."
   - **Substrate — the invariant register.** What a gene *is*, precisely, here. **`invariant` stays
     canonical** for what is enforced.
   **The load-bearing caveat (why gene does *not* go fully canonical):** the gene analogy **provably
   breaks at D2**, the most load-bearing floor — `research-0005` §Limits: *"No intent locus in a
   genome; A3/D2 has no genetic analogue."* Making genes canonical would seat the intent gate exactly
   where the metaphor is weakest (the DES lens covers that gap). Plus the naming guardrail
   (`brief-§4`): gene-talk *sounds* authoritative — keep it framed as *our synthesis / teaching
   metaphor*, sharper externally, never a provenance claim. **Policy:** current-truth/product
   artifacts use the **canonical** term + link the lexicon; research notes keep their lens vocabulary;
   gene-expression is the sanctioned way to *convey* it. Normalization ≠ monolingualism.

3. **Delivery-relationship dial ends (Axis A, `research-0007`) — a consistent role-noun pair.** The two
   ends are genuinely **asymmetric**: the live end *is* an active agent (Trellis running); the pull end
   is the *absence* of an ongoing agent (Trellis referenced, then just a source). So the pair is named
   in **role register** (matched agent-nouns), not garden register — a garden agent-noun for the pull
   end does not exist (a "cutting" is an object, not an actor; "trimmer" fits the grammar but means
   pruning, not pull-delivery).
   - **`supervisor`** — the installed/live/push end. **Kept.** Trellis with live authority, performing
     the DES supervisor role (`research-0006`).
   - **`advisor`** ← *was* "consultant"/"cutting" — the referenced/pulled end. Trellis **consulted**,
     no executive authority; the host internalizes the guidance and acts on its own — exactly pull
     mode. Grammatically twinned with `supervisor`; sidesteps the `reference`↔`inv-reference-
     relationship` (B8) collision and the garden-agent-noun gap. *(Garden-consistent alternative if
     the dial should carry garden flavor: `frame` / `cutting`, both garden-object nouns — retires
     "supervisor.")*

## Consequences

- **Rename sweep DONE** — `consultant → advisor` across research 0006–0009 + `spec-0002`. The
  `consultant-mode-work-usage` memory datapoint keeps its name (a proper-noun reference to a specific
  episode, not the generic term).
- **The conformance check learns the `lexicon` type** (rubric + `conformance-reviewer` updated).
- **`lexicon-v1` is the store** — **ratified 2026-07-03** alongside this decision (maintainer's D2
  confirm).
- **Ratified 2026-07-03** — the naming is intent-layer (D2, the naming guardrail `brief-§4`); the
  maintainer confirmed `advisor` and the three-register model.

## Open questions

- **Pull-end name — RESOLVED (2026-07-03): `advisor`.** Role-noun parallel with `supervisor`;
  precise (consulted, no live authority). Rejected: `cutting` (object vs. agent-noun), `frame`/`cutting`
  (would retire "supervisor"), `reference` (B8 collision), keep `consultant`.
- **Split `supervisor`?** Proposed: no — the delivery end and the DES role are the same metaphor.
  Revisit if the overload bites in practice.
- **Axis B (payload depth) terms** — `expressed-only / +latent / +mechanism` — left as-is for now;
  canonicalize if they prove confusing.

## Supersedes / superseded by

— (none)

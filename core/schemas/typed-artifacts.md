---
id: schema-typed-artifacts
type: schema
status: approved  # migrated verbatim from the approved spec-0002 by decision-0079; no clause changed
depends_on: [decision-0016, invariants-v1]
owner: gundi
scope: trellis-product
date: 2026-08-29
---

# Schema — the two typed artifacts

> **Provenance.** This is `spec-0002` §1–§3, migrated **unchanged in substance** when
> `decision-0079` retired the spec stage. `decision-0016` fixed the *existence and scope* of the
> two typed artifacts and explicitly deferred their **schema and lifecycle** to `spec-0002`, so
> deleting that spec without rehoming these sections would have left `core/catalog/` and
> `profiles/` with no authoring contract. Only cross-references were retargeted at their
> surviving homes; every field rule below is the ratified text.
>
> The **conformance checks** that read this schema live in `core/rubrics/artifact-contract.md`
> (checks 8–11) — they were `spec-0002` §4 and were already carried there in full.
>
> *Amended 2026-09-07 (TRL-64). WHAT: the `## Worked instances`, `## Acceptance criteria` and
> `## Open questions` sections below. WHY: `decision-0079` migrated §1–§3 and re-homed §4's checks,
> but `spec-0002` also carried `## 5. Worked fragments`, seven acceptance criteria and six open
> questions, and those three blocks were carried nowhere — deleted with the spec. Verified by
> diffing `c75eace^:specs/0002-expression-profile-and-signature-catalog.md` against this file.
> Seven of the thirteen items were already covered elsewhere and are recorded below as pointers
> rather than restored; the other six are stated here, because this file is where they bind. SCOPE:
> additive, plus three cross-references the migration left pointing at sections that do not exist
> here (`§4.5` and `§4` → their rubric checks; `this spec` → `this schema`) and one retired display
> code (`B4` → `inv-auditable-archive`, `decision-0038`). **No field rule changed.** POINTER:
> `decision-0079` Consequences, dated note.*

## 1. The signature-catalog schema (`trellis-product`, one, shipped)

Frontmatter (`spec-0001` §1 base + this type). Versioned + **revise-in-place** like `invariants-v1`
(it is `trellis-product`, not append-only): `id: signature-catalog-v1`, `type: signature-catalog`,
`scope: trellis-product`, `depends_on: [invariants-v1]` (it annotates that set).

Body: one **entry per assessable invariant** (A/B/D — not the C dials; see coverage note below),
keyed by its **stable slug** (`decision-0013`). Fields:

| Field | Req | Rule |
|---|---|---|
| `slug` | ✓ | a slug in the `invariants-v1` registry; **must resolve** (a superseded slug resolves through the registry) |
| `what` | ✓ | one line — what the invariant is (the dictionary voice; the reference + benefits page render this) |
| `directive` | ✓ | one line — the **imperative, host-agent-facing** instruction the always-loaded block renders; self-contained, no internal codes (`decision-0034`) |
| `why` | ✓ | the goal / benefit in one line, **agents-first** (`decision-0020`) — the benefits page renders this |
| `signature` | ✓ | the **observable tells** that a project honors it *implicitly* — the field Assess detects against (the genome annotation) |
| `honored` | ✓ | the **with** side of **≥2 matched pairs** — the fixed version; `honored[i]` pairs with `violated[i]` (same use case + layer tag, same order; `decision-0027`) |
| `violated` | ✓ | the **without** side of those pairs — the same situations shown broken; ≥2, spanning different layers, tagged |
| `class` | ✓ | the invariant's own class: `methodology` (A) · `trellis-design` (B) · `dial` (C) · `floor` (D) |
| `mechanizable` | ✓ | `true` for the SCT-computable fragment (`inv-directional-flow`, `inv-ratifiable-artifacts`, `inv-graph-maintenance` flow-facet, `inv-gate-at-handover`); `false` for the behavioral genes (`inv-independent-judgment`, `inv-clarify-before-commit`, `floor-transparency`) — `research-0006` §Limits partitions the set |
| `default_C1` | ✓ | default enforcement strength ∈ `{expressed, default-on-but-skippable, enforced}` (`decision-0008`) |
| `default_C2` | ✓ | default gatekeeper ∈ `{independent-agent, human, none}`; **never `none` at the intent locus** (`floor-intent-gate`/D2) |
| `intent_locus` | — | `true` on the intent-gate slugs (`inv-intent-locus`, `floor-intent-gate`) — marks entries a profile may never set to `C2: none` (rubric check 11, D2). Default `false`. |

**Coverage is a gate (AC1) — the *assessable* invariants, not the dials.** The catalog covers
every **assessable** invariant slug: the A structural set, the B operating set, and the D floors
(`inv-directional-flow` … `floor-intent-gate`; B8 collapsed into D1, `decision-0021`). It
**excludes the two C dials** (`dial-enforcement-strength`, `dial-gatekeeper`): a project does not
"honor a dial implicitly" — the dials are the *axes the catalog's entries are set along* (columns of a
profile, not rows). A missing *assessable* slug is a conformance failure; the two dials are correctly
absent.

**Examples are required, diverse, and stay in sync (the meta-rule, `decision-0020`).** Every entry
carries `why` + `honored` + `violated`, and **each of `honored`/`violated` carries ≥2 examples from
different layers** (CI / spec / research / code / UI / ops …). Diversity is the point: one example
reads as domain-specific; two-plus across layers show the principle *generalizes* (which is what lets
an agent recognize the invariant in a context it hasn't seen); a 3rd only when it teaches a genuinely
new layer, never padding. **A change that edits an invariant without updating its examples is a
conformance failure** (rubric check 8). Presence + count is the enforceable floor; not-left-stale is the
substantive check (weakly checkable, like SI-1). The iron rule + referential integrity applied to the
rule-set itself — and **the landing/benefits page derives from these fields**, so a page claim always
has a rule behind it.

## 2. The expression-profile schema (`core-methodology`, one per instance)

Frontmatter: `id: profile-<instance>` (e.g. `profile-rpi-team`), `type: expression-profile`,
`scope: core-methodology`, `depends_on: [signature-catalog-v1, invariants-v1]`, `owner`.

**Instance-level fields (the delivery choice — `research-0007`).** `decision-0016` colocated
"delivery axes A/B" under *each invariant*; this schema **sharpens** that (the one revision
`decision-0016` explicitly permits): Axis A and Axis B are **instance-level**, not per-gene —

| Field | Req | Rule |
|---|---|---|
| `delivery_relationship` | ✓ | Axis A: `supervisor` (push/installed/live) \| `advisor` (pull/referenced) |
| `payload_depth` | ✓ | Axis B: `expressed-only` \| `+latent` \| `+mechanism` (self-regulating) |
| `application_model` | ✓ | `M1-overlay` (default; augment-never-clobber) \| `M2-morph` (deferred option) — `research-0005/0006` |

**Per-invariant entry** (keyed by slug, each resolving to a catalog entry (§1)):

| Field | Req | Rule |
|---|---|---|
| `slug` | ✓ | **must resolve to a catalog entry** (else dangling — rubric check 9) |
| `active` | ✓ | `true` = the gene is expressed here; `false` = latent/silent |
| `C1` | ✓ if active | chosen strength ∈ `{expressed, default-on-but-skippable, enforced}` (may not exceed nothing, but is the *instance's* call) |
| `C2` | ✓ if active | `{independent-agent, human, none}`; **`none` forbidden when the catalog marks this an intent-locus gate** (D2) |
| `basis` | ✓ if active | `honored-implicitly` (Assess detected it) \| `to-be-added` (Apply will compose it) |
| `confidence` | ✓ if `honored-implicitly` | `verified` \| `inferred` \| `speculated` — Assess's certainty the project already honors it |
| `evidence` | ✓ if `honored-implicitly` | pointer to the concrete project tell that matched the catalog `signature` (path/quote) |

**Assert-and-verify, never silently "honored" (`research-0009`).** An `active: true` +
`basis: honored-implicitly` entry with **no `confidence` + `evidence`** is a conformance failure.
Assess is loud-failure-biased: it may claim a gene is honored only by pointing at the tell — the
iron rule applied to detection.

## 3. Lifecycle — the D2 gate, made concrete

- **Catalog** — `trellis-product`, revise-in-place, versioned. `draft → ratified` by the
  **maintainer**. A profile consumes only a **ratified** catalog (directional flow).
- **Profile** — `core-methodology`, per instance, produced by the **Assess** sub-agent as `draft`.
  The human **ratifies (D2)** — `draft → ratified`. **Apply consumes only the ratified profile.**
  This *is* `research-0007`'s flow made a lifecycle: *Assess proposes → human ratifies at D2 →
  delivery composes exactly that profile* — never silently maximal (`decision-0008`, `spec-0001`
  §5). **Producer ≠ ratifier ≠ verifier** (`inv-independent-judgment`): Assess produces, the human
  ratifies, the conformance check verifies, Apply consumes — four distinct roles.
- **Re-assessment** supersedes a prior profile (append forward pointer if the instance treats
  profiles as history; revise-in-place if it keeps one current-truth profile — the instance's
  `inv-auditable-archive` call, not fixed here).

## Worked instances (the iron rule — the schema exemplified, not just described)

`spec-0002` §5 carried three hand-written fragments. They are **not** restored: this repo now holds
live artifacts written against this schema, and a checked instance is a better example than a
fragment.

- **`core/catalog/signature-catalog-v1.md`** — §1 filled: an entry per assessable slug carrying
  `what` / `directive` / `why` / `signature`, matched `honored`/`violated` pairs and the default
  dials. Rubric check 8 grades it.
- **`profiles/trellis-self.md`** — §2 and §3 filled: the instance-level axes, per-gene `C1`/`C2`,
  and the `honored-implicitly` + `confidence` + `evidence` path §2 requires. Rubric checks 9–11
  grade it.

**The one fill with no live instance is `payload_depth: expressed-only`** — `trellis-self` is
`+mechanism`. That gap is why AC6 below is a criterion to be met rather than a property already
demonstrated.

## Acceptance criteria

`spec-0002` carried seven. Five are enforced elsewhere and are listed here as pointers; **AC6 and
AC7 grade the schema itself** rather than an artifact written against it, so no rubric check could
carry them and they are stated here. **The numbering is `spec-0002`'s and is deliberately kept** —
`profiles/trellis-self.md` cites `AC7` by that name.

| | Where it is enforced |
|---|---|
| **AC1** — the catalog covers every assessable invariant, with goal + examples | rubric check 8 |
| **AC2** — every profile gene resolves to a catalog entry | rubric check 9 |
| **AC3** — no silent "honored": `confidence` + `evidence`, or fail | rubric check 10 |
| **AC4** — no profile sets `C2: none` on an intent-locus gate | rubric check 11 |
| **AC5** — D2 ratification is real | §3 above states both clauses; rubric check 5's lifecycle proviso carries the structural half. `decision-0082` retired `status` for trellis-self and **left this product-layer lifecycle standing** |

- **AC6 — a partial application needs no bespoke artifact.** The behavioral subset is expressible
  purely as a profile — `payload_depth: expressed-only` with every pipeline gene `active: false` —
  so a lighter Trellis is a *fill* of this schema, not a new artifact type. **A proposal to mint a
  standalone "lite" rule-list fails this criterion**, and is the check to run when one is proposed.
- **AC7 — one object, four consumers.** The *same* profile schema is what Assess writes, the human
  ratifies, Apply reads, and a cross-instance diff compares: **no consumer may require a field the
  others do not produce.** A field added to serve one consumer alone breaks this. It is the
  standing test on every amendment to §2, and the round-trip test named in
  `profiles/trellis-self.md`'s own open questions.

## Open questions

**The consumer is this file.** It is the mandatory input to authoring a catalog or a profile and to
amending the schema, which is what `profiles/trellis-self.md` means by *"every artifact's
`## Open questions` rides a consumer that must read it."* Each item below is a constraint on the
next amendment of this schema — **none is a debt owed to an unbuilt component**, which is the shape
that same row names as the repo's live orphan.

- **Axis-B granularity vs. `decision-0016`.** §2 makes Axis A/B **instance-level**, against that
  decision's per-invariant wording — sharpened deliberately, and **flagged rather than silently
  reconciled**. If `+latent` genes ever need per-gene presence (active / latent / absent as three
  states), §2's `active` boolean must widen. Settle this before widening it.
- **Catalog vs. profile: still two types?** `research-0009` still carries the question of whether
  the dictionary and the per-instance readout collapse into one artifact. This schema keeps them
  two, and the `trellis-product` / `core-methodology` scope split is the reason. **Confirm at
  instance #2** — `profiles/` holds only `trellis-self.md`, so the confirming event has not happened.
- **Confidence scale.** §2's `confidence` enum (`verified` / `inferred` / `speculated`) borrows the
  research-note tag set. Whether a *detection* confidence is the same kind of thing as a *sourcing*
  confidence is unresolved; if it is not, the enum that changes is §2's.
- **Profile history model.** §3 leaves append-only-superseding versus revise-in-place to the
  instance. A cross-instance diff may force one convention; if it does, §3 stops being the
  instance's call.
- **The `C1` / `C2` field names.** `decision-0038` deferred renaming `default_C1` / `default_C2` to
  `default_strength` / `default_gatekeeper` *"when `spec-0002` is next opened"*. This file is that
  referent, so the deferral is recorded against a file that exists rather than a deleted one.

*Two of `spec-0002`'s six are deliberately not repeated here, because repeating them would fork a
live entry (`decision-0028`): the structure of detection heuristics is
`core/catalog/signature-catalog-v1.md`'s own open question ("Structured signatures"), and the
`approved` state is `decision-0082`'s ("Should `spec-0002`'s profile lifecycle follow?"), which
that record explicitly left standing as a separate product call.*

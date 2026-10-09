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
> The **conformance checks** that read this schema live in `docs/rubrics/artifact-contract.md`
> (checks 8–11) — they were `spec-0002` §4 and were already carried there in full.
>
> *Amended in place 2026-10-09 (TRL-64): the migration above carried nothing of `spec-0002`'s
> `## 5. Worked fragments`, `## Acceptance criteria` or `## Open questions`. The three sections at
> the end of this file are additions, not migrated text: they give each of those items a home, as
> a pointer where another artifact already carries it. Three references the migration left
> pointing at things this file is not (`§4`, `§4.5`, "this spec") now name their real targets,
> §2's note on `decision-0016` now says that record was revised to match, and §2 cites AC3 again.
> No field rule changed.*

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

**Instance-level fields (the delivery choice — `research-0007`).** `decision-0016` first colocated
"delivery axes A/B" under *each invariant*; this schema **sharpens** that (the one revision
`decision-0016` records `spec-0002` forcing, since folded into that record): Axis A and Axis B are **instance-level**, not per-gene —

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

**Assert-and-verify, never silently "honored" (AC3, `research-0009`).** An `active: true` +
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
  profiles as history; revise-in-place if it keeps one current-truth profile — the instance's B4
  call, not fixed here).

## Worked instances

`spec-0002` §5 carried three hand-written fragments. They are **not** restored: the repo now holds
two live artifacts written against this schema, and a checked instance is a better example than a
fragment.

- **`core/catalog/signature-catalog-v1.md`** fills §1. Rubric check 8 grades it.
- **`profiles/trellis-self.md`** fills §2 and §3, including the `honored-implicitly` +
  `confidence` + `evidence` path. Rubric checks 9–11 grade it.

**The one profile sits at the heavy end of every axis:** `supervisor` / `+mechanism` / `M2-morph`.
So `advisor`, `expressed-only`, `+latent` and `M1-overlay` are all undemonstrated, and the dropped
§5b fragment exercised three of them. That gap is why AC6 below is a criterion still to be met,
not a property already shown.

## Acceptance criteria

`spec-0002` carried seven. **The numbering is `spec-0002`'s and is kept**, because
`profiles/trellis-self.md` cites AC7 by that name and §1's `(AC1)` and §2's `(AC3)` resolve here.
Five live elsewhere and appear as pointers. AC6 and AC7 grade this schema itself rather than an
artifact written against it, so no rubric check can carry them, and they are stated here.

| | Where it lives now |
|---|---|
| **AC1** — the catalog covers every assessable invariant, with goal + examples | rubric check 8 |
| **AC2** — every profile gene resolves to a catalog entry | rubric check 9 |
| **AC3** — no silent "honored": `confidence` + `evidence`, or fail | rubric check 10 |
| **AC4** — no profile sets `C2: none` on an intent-locus gate | rubric check 11 |
| **AC5** — the human ratification gate is real: Apply consumes only a ratified profile, and a ratified profile depends only on a ratified catalog | §3 above states both clauses. Rubric check 5's proviso carries the second where a methodology declares a status lifecycle (no gated artifact `depends_on` a draft one). No corpus check exercises either today: this repo declares no lifecycle, and its artifacts' legacy `status:` lines are not read (`decision-0082`, which left this product-layer lifecycle standing) |

- **AC6 — Trellis-lite needs no bespoke artifact.** The behavioral subset is expressible purely as
  a profile with `payload_depth: expressed-only` and every pipeline gene `active: false`, proving a
  lighter Trellis is a special case of this schema, not a separate thing.
- **AC7 — one object, four consumers.** The *same* profile schema is what Assess writes, the human
  ratifies, Apply reads, and a cross-instance diff compares — no consumer needs a field the others
  don't produce. `profiles/trellis-self.md`'s open questions name its round-trip as the test that
  Assess works.

## Open questions

**These are open design questions, not deferred work.** Each constrains the next change that
authors a catalog or a profile, or amends this schema, and that change is written against this
file, so its author must read them. That is the address `profiles/trellis-self.md` counts as
honored for `inv-no-orphan-followups` — *"every artifact's `## Open questions` rides a consumer
that must read it"* — and the section rubric check 6 requires of every other typed artifact here.
It is not the *"someone reading this file later"* that the invariant rules out, because the reader
is required, not merely possible. None is a debt owed to a component that does not exist yet.

- **Axis-B granularity.** §2 makes Axis A/B instance-level, and `decision-0016` now states that form
  itself, so there is no open tension with it. What stays open: if `+latent` genes ever need
  per-gene presence (active / latent / absent), §2's `active` boolean must widen, and that re-opens
  the granularity `decision-0016` settled. Settle it deliberately rather than by widening a field.
- **Catalog vs. profile: still two types?** `research-0009` still asks whether the dictionary and
  the per-instance readout collapse into one artifact. This schema keeps them two, for the
  `trellis-product` / `core-methodology` scope split. **Confirm at instance #2**: `profiles/` holds
  only `trellis-self.md`, so that has not happened.
- **Confidence scale.** §2's `confidence` enum (`verified` / `inferred` / `speculated`) borrows the
  research-note tag set. Whether a detection confidence is the same kind of thing as a sourcing
  confidence is unresolved. If it is not, §2's enum is what changes.
- **Profile history model.** §3 leaves append-only superseding versus revise-in-place to the
  instance. A cross-instance diff may force one convention, and then §3 stops being the instance's
  call.
- **An `approved` state after `ratified`.** §3 stops at `draft → ratified`. A ratified profile that
  Apply has composed was `spec-0002`'s candidate first user of an execution-layer `approved`. This
  is related to, but not the same as, `decision-0082`'s open question *"Should `spec-0002`'s profile
  lifecycle follow?"*: retiring the lifecycle would moot it, and keeping the lifecycle brings it
  back.
- **The `C1` / `C2` field names.** Should `default_C1` / `default_C2` become `default_strength` /
  `default_gatekeeper`, with the profile's columns to match? `decision-0038` asked this and left it
  to *"when `spec-0002` is next opened"*; this file is where that question now lives. The rename
  would cascade through the catalog, the profile and the rubric, so it is not made here.

`spec-0002`'s fourth question, where the catalog's detection heuristics live, is not repeated here.
It is `core/catalog/signature-catalog-v1.md`'s own open question, *"Structured signatures"*, and a
second copy would fork a live entry (`decision-0028`). That entry is *"Owed to the Assess build
(cluster 1)"*, which `profiles/trellis-self.md` counts as this repo's live orphan; pointing at it
gives the question a home but not a consumer.

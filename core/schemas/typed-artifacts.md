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
| `intent_locus` | — | `true` on the intent-gate slugs (`inv-intent-locus`, `floor-intent-gate`) — marks entries a profile may never set to `C2: none` (§4.5, D2). Default `false`. |

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
conformance failure** (§4). Presence + count is the enforceable floor; not-left-stale is the
substantive check (weakly checkable, like SI-1). The iron rule + referential integrity applied to the
rule-set itself — and **the landing/benefits page derives from these fields**, so a page claim always
has a rule behind it.

## 2. The expression-profile schema (`core-methodology`, one per instance)

Frontmatter: `id: profile-<instance>` (e.g. `profile-rpi-team`), `type: expression-profile`,
`scope: core-methodology`, `depends_on: [signature-catalog-v1, invariants-v1]`, `owner`.

**Instance-level fields (the delivery choice — `research-0007`).** `decision-0016` colocated
"delivery axes A/B" under *each invariant*; this spec **sharpens** that (a spec-forces-revision the
decision explicitly permits): Axis A and Axis B are **instance-level**, not per-gene —

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
  profiles as history; revise-in-place if it keeps one current-truth profile — the instance's B4
  call, not fixed here).

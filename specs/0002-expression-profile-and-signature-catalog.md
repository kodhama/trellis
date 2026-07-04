---
id: spec-0002
type: spec
status: ratified
ratified: 2026-07-03
depends_on: [decision-0016, invariants-v1, spec-0001, decision-0008, decision-0009, research-0005, research-0007, research-0009]
owner: gundi
rubric: spec-quality
---

# Spec 0002 — The expression profile & the invariant-signature catalog: schema + lifecycle

> **Ratified 2026-07-03 (D2), together with `decision-0016` and `research-0005–0009`.** The
> condition this spec set for itself — that its gate and the underlying research ratify first — was
> met: they ratified in the same pass. The schema below is now current-truth; the first artifacts
> built against it are `signature-catalog-v1` and `profile-trellis-self` (both `draft`, pending
> their own D2 ratification — the builder does not self-ratify its output).

## Purpose

Give the two artifact `type`s that `decision-0016` introduced a **checkable schema + lifecycle**,
so they stop being a named idea and become producible/enforceable artifacts (the iron rule):

- **`signature-catalog`** (`trellis-product`, one shipped) — the genome annotation: per invariant,
  what it *is*, the observable **signature** that it is honored implicitly, and its **default
  dials**. Consumed by Assess (#23) and tutoring (#27).
- **`expression-profile`** (`core-methodology`, one per instance) — the per-instance readout:
  per invariant, is it **active**, at what **C1** strength, gatekept by whom (**C2**), and on what
  confidence/evidence. Produced by Assess, **ratified by the human (D2)**, consumed by Apply,
  diffed across instances (#28), minimized into "Trellis-lite" (#22).

`decision-0016` fixed their *existence, scope, and relationship*; this spec fixes their *fields,
required sections, cross-references, and lifecycle*, and teaches the `spec-0001` conformance check
to enforce them. No runtime, throughout (`decision-0010`): both are `.md` artifacts with
frontmatter; Assess/Apply and the check are sub-agents.

## Scope

**In scope:** the frontmatter + body schema for each type; the per-invariant entry fields; how a
profile references catalog slugs; how Assess populates and **confidence-tags** a profile; the
**Assess → ratify (D2) → Apply** lifecycle; the new conformance checks. Dogfooded with two worked
fragments (§5) — Trellis-lite (#22) and the RPI-Team instance (#24).

**Out of scope (later builds):** the *populated* catalog artifact itself (a `core/` build, per
`decision-0005`), the Assess/Apply sub-agents (cluster 1), the cross-instance **diff** report
(#28 — this spec fixes the per-instance object it diffs, not the diff), and tutoring's derived
explain-cards (#27). This spec is the shared primitive those consume (`research-0009` build order).

## 1. The signature-catalog schema (`trellis-product`, one, shipped)

Frontmatter (`spec-0001` §1 base + this type). Versioned + **revise-in-place** like `invariants-v1`
(it is `trellis-product`, not append-only): `id: signature-catalog-v1`, `type: signature-catalog`,
`scope: trellis-product`, `depends_on: [invariants-v1]` (it annotates that set).

Body: one **entry per assessable invariant** (A/B/D — not the C dials; see coverage note below),
keyed by its **stable slug** (`decision-0013`). Fields:

| Field | Req | Rule |
|---|---|---|
| `slug` | ✓ | a slug in the `invariants-v1` registry; **must resolve** (a superseded slug resolves through the registry) |
| `what` | ✓ | one line — what the invariant is |
| `why` | ✓ | the goal / benefit in one line, **agents-first** (`decision-0020`) — the benefits page renders this |
| `signature` | ✓ | the **observable tells** that a project honors it *implicitly* — the field Assess detects against (the genome annotation) |
| `honored` | ✓ | **≥2** concrete with-cases, **each drawn from a different layer** (CI / spec / research / code / UI / ops …) and tagged with it — so the principle reads as general, not domain-specific (`decision-0020`) |
| `violated` | ✓ | **≥2** concrete without-cases, likewise spanning different layers |
| `class` | ✓ | the invariant's own class: `methodology` (A) · `trellis-design` (B) · `dial` (C) · `floor` (D) |
| `mechanizable` | ✓ | `true` for the SCT-computable fragment (`inv-directional-flow`, `inv-ratifiable-artifacts`, `inv-graph-maintenance` flow-facet, `inv-gate-at-handover`); `false` for the behavioral genes (`inv-independent-judgment`, `inv-clarify-before-commit`, `floor-transparency`) — `research-0006` §Limits partitions the set |
| `default_C1` | ✓ | default enforcement strength ∈ `{expressed, default-on-but-skippable, enforced}` (`decision-0008`) |
| `default_C2` | ✓ | default gatekeeper ∈ `{independent-agent, human, none}`; **never `none` at the intent locus** (`floor-intent-gate`/D2) |
| `intent_locus` | — | `true` on the intent-gate slugs (`inv-intent-locus`, `floor-intent-gate`) — marks entries a profile may never set to `C2: none` (§4.5, D2). Default `false`. |

**Coverage is a gate (AC1) — the *assessable* invariants, not the dials.** The catalog covers
every **assessable** invariant slug: the A structural set, the B operating set, and the D floors
(14 slugs — `inv-directional-flow` … `floor-intent-gate`; B8 collapsed into D1, `decision-0021`). It
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

**Per-invariant entry** (keyed by slug, each resolving to a §1 catalog entry):

| Field | Req | Rule |
|---|---|---|
| `slug` | ✓ | **must resolve to a catalog entry** (else dangling — §4 check) |
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

## 4. What the conformance check must learn (extends `spec-0001` §3 / the rubric)

`decision-0016`: "the conformance check must learn the two new types + their required sections."
Added checks (they compose with `spec-0001`'s existing seven):

1. **Type registry.** `signature-catalog` (`scope: trellis-product`) and `expression-profile`
   (`scope: core-methodology`) are declared types with the required sections below.
2. **Catalog coverage + examples (AC1, `decision-0020`).** Every *assessable* `invariants-v1` slug
   (A/B/D — the 14, excluding the two C dials; a collapsed slug is covered by its successor) has a
   catalog entry carrying **all §1 required fields, including `why` / `honored` / `violated`**. *FAIL →
   name the uncovered assessable slug or the entry missing a field (a bare entry with no `why`/example
   is a fail).*
3. **Profile→catalog resolution (AC2).** Every profile `slug` resolves to a catalog entry — a
   profile gene with no catalog annotation is a **dangling reference** (`spec-0001` check 4). *FAIL
   → name the unresolved slug.*
4. **Evidence floor (AC3).** No `active: true` + `honored-implicitly` entry lacks `confidence` +
   `evidence`. *FAIL → name the bare "honored" claim.*
5. **Intent-gate floor (AC4, D2).** No profile sets `C2: none` on a gate the catalog marks
   intent-locus. *FAIL → name the offending gate.*
6. **Directional flow (already `spec-0001` check 5).** A `ratified` profile depends only on a
   `ratified` catalog — so a ratified profile can never rest on an unratified dictionary.

**Required body sections (per type — extends `spec-0001` §4):**
- `signature-catalog` → `## Entries` (the per-slug table/list), `## Acceptance criteria`,
  `## Open questions`.
- `expression-profile` → `## Delivery` (the instance-level axes), `## Profile` (the per-slug
  entries), `## Assessment notes` (confidence/evidence rationale), `## Open questions`.

## 5. Worked fragments (iron rule — the schema exemplified, not just described)

**5a. Catalog fragment** (three entries — one mechanizable A-gene, two behavioral genes):

```
- slug: inv-directional-flow            # A1
  what: one-way stages of decreasing ambiguity; downstream never consumes a draft
  signature: research/decision/spec dirs exist in order; no ratified file cites a draft upstream
  class: methodology   mechanizable: true
  default_C1: enforced   default_C2: independent-agent
- slug: inv-independent-judgment        # B3 (intent face)
  what: assessments track evidence, not the maintainer's preferences; builder never sole judge
  signature: reviews record dissent/risks; a verifier distinct from the producer; no reflexive assent
  class: trellis-design   mechanizable: false
  default_C1: default-on-but-skippable   default_C2: human
- slug: inv-clarify-before-commit       # B9
  what: ambiguity in an upstream is surfaced and resolved before downstream consumes it
  signature: open-questions sections; clarifying exchanges recorded before build starts
  class: trellis-design   mechanizable: false
  default_C1: default-on-but-skippable   default_C2: human
```

**5b. Trellis-lite as a profile (#22, `research-0005` "differentiated cell type").** The behavioral
subset — expressed, no pipeline machinery — is *a profile*, not a bespoke rule list:

```
delivery_relationship: advisor   payload_depth: expressed-only   application_model: M1-overlay
profile:
  - slug: inv-independent-judgment  active: true  C1: default-on-but-skippable  C2: human
    basis: to-be-added
  - slug: inv-clarify-before-commit active: true  C1: default-on-but-skippable  C2: human
    basis: to-be-added
  - slug: floor-transparency        active: true  C1: enforced                  C2: human
    basis: to-be-added
  - slug: inv-directional-flow      active: false     # pipeline gene left silent
```

**5c. RPI-Team, an assessed instance (#24 — "honors 1/2/4/5/6 implicitly").** Shows the
`honored-implicitly` + `confidence` + `evidence` path:

```
- slug: inv-directional-flow  active: true  C1: expressed  C2: independent-agent
  basis: honored-implicitly  confidence: verified
  evidence: "docs/ → design/ → src/ one-way; PR template blocks merge on open design Qs"
```

Trellis-lite and RPI-Team are the *same object* at different fills — which is the whole claim:
one schema serves #22 (minimize), #23/#24 (assess/apply), #28 (diff).

## Acceptance criteria

- **AC1 — catalog covers every assessable invariant, with goal + examples.** The catalog covers every
  A/B/D slug (**14** — `inv-self-improvement` restored `decision-0018`, `inv-reference-relationship`
  collapsed into D1 `decision-0021`), each with `what` / **`why`** / `signature` / **`honored`** /
  **`violated`** / `class` / `mechanizable` / `default_C1` / `default_C2`; a missing assessable slug or
  field (including a missing example) fails the check (§4.2). The two C dials are excluded by design.
- **AC2 — profiles resolve.** Every profile gene references a catalog slug; an unresolved slug is a
  named dangling reference (§4.3).
- **AC3 — no silent "honored".** Every `active + honored-implicitly` entry carries a `confidence`
  tag **and** an `evidence` pointer, or the check fails (§4.4). Assert-and-verify.
- **AC4 — intent gate holds.** No profile sets `C2: none` on an intent-locus gate; the check
  rejects one that does (§4.5, D2).
- **AC5 — D2 ratification is real.** A profile is consumed by Apply only in `ratified` state, and a
  ratified profile depends only on a ratified catalog (§3, §4.6). The human gate is a lifecycle
  transition, not a comment.
- **AC6 — Trellis-lite needs no bespoke artifact.** The behavioral subset is expressible purely as
  a profile (§5b) with `payload_depth: expressed-only` and every pipeline gene `active: false` —
  proving #22 is a special case of this schema, not a separate thing.
- **AC7 — one object, four consumers.** The *same* profile schema is what Assess writes, the human
  ratifies, Apply reads, and #28 diffs — no consumer needs a field the others don't produce.

## Open questions

- **Axis-B granularity vs. `decision-0016`.** This spec makes Axis A/B instance-level (§2), against
  the decision's per-invariant wording. If `+latent` genes ever need *per-gene* presence
  (active/latent/absent as three states), the per-invariant `active` boolean must widen — revisit
  when Apply is built. Flagged, not silently reconciled.
- **Catalog vs. profile: still two types?** `research-0005/0009` leave open whether the dictionary
  and the per-instance readout collapse into one artifact. This spec keeps them two (dictionary is
  `trellis-product`/shipped-once; profile is `core-methodology`/per-instance) — the scope split is
  the reason to keep them apart. Confirm at instance #2.
- **Confidence scale.** `verified/inferred/speculated` borrows the research-note tag set. Is a
  detection confidence the same kind of thing as a sourcing confidence, or does Assess need its own
  (e.g. `strong/weak/absent-signal`)? Decide when Assess is specified.
- **Where the catalog's detection heuristics live.** The `signature` field is prose here; Assess may
  need it structured (regex-ish tells vs. judgment tells, tracking `mechanizable`). Owed to cluster 1.
- **Profile history model (B4).** Per-instance: append-only superseding profiles, or one
  revise-in-place current-truth profile? Left to the instance in §3 — but #28's diff may force a
  convention. Revisit with the cross-instance report.
- **`approved` state.** `spec-0001` §2 defers an execution-layer `approved`. A ratified profile that
  Apply has *composed* is a candidate first user of it — but this spec stays on `draft → ratified`
  until that state is decided (still `trellis-product` scope, not guessed here).

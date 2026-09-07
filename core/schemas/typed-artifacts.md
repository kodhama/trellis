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
> **Six of the thirteen items were already covered elsewhere** and are recorded below as pointers
> rather than restored; the other **seven** are stated here, because this file is where they bind.
> SCOPE: additive, plus four repairs to prose the migration left inaccurate — two references
> pointing at sections that **do not exist in this file**, where a reader is stranded rather than
> redirected (`§4.5` and `§4` → their rubric checks); one stale self-description (`this spec` →
> `this schema`, since this file is not a spec); and §2's parenthetical *"(a spec-forces-revision
> the decision explicitly permits)"* → *"(the one revision `decision-0016` explicitly permits)"* —
> `decision-0016` supplies *"did force one revision"*, and "permits" is inherited from the migrated
> text rather than quoted from it. The dropped `AC3` inline marker is restored, so §1 and §2 again
> cite their criteria alike. Citations to `spec-0001`/`spec-0002` themselves are left alone:
> `decision-0079`'s retired-artifacts registry is what resolves those, and that record chose not to
> edit citing artifacts. **Retired display codes (`D2` ×7, `B4` ×1) are left as migrated**, against
> `decision-0038`'s instruction to *"migrate opportunistically as each file is next edited"* — on
> that record's own stated rationale, *"no sweep ceremony over revise-in-place files nothing else
> touches"*: an eight-code sweep is a second logical change, and doing one of the eight is worse
> than doing none. **No field rule changed.** POINTER: `decision-0079` Consequences, dated note.*

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

**The repo holds one profile, and it fills the heavy end of every axis.** `trellis-self` is
`supervisor` / `+mechanism` / `M2-morph`, so `advisor`, `expressed-only`, `+latent` and `M1-overlay`
are **all** undemonstrated — and the dropped §5b fragment exercised three of them at once. (The
catalog carries no axes at all; §2's three axis fields are the profile's.) That gap is why AC6
below is a criterion to be met rather than a property already shown, and it is the honest cost of
preferring checked instances to fragments: the fragments covered a corner the instances do not
reach.

## Acceptance criteria

`spec-0002` carried seven. Five are enforced elsewhere and are listed here as pointers; **AC6 and
AC7 grade the schema itself** rather than an artifact written against it, so no rubric check could
carry them and they are stated here. **The numbering is `spec-0002`'s and is deliberately kept** —
`profiles/trellis-self.md` cites `AC7` by that name, and §1's `(AC1)` and §2's `(AC3)` resolve into
this table.

| | Where it is enforced |
|---|---|
| **AC1** — the catalog covers every assessable invariant, with goal + examples | rubric check 8 |
| **AC2** — every profile gene resolves to a catalog entry | rubric check 9 |
| **AC3** — no silent "honored": `confidence` + `evidence`, or fail | rubric check 10 |
| **AC4** — no profile sets `C2: none` on an intent-locus gate | rubric check 11 |
| **AC5** — D2 ratification is real | §3 above states both clauses; rubric check 5's lifecycle proviso carries the structural half. `decision-0082` retired `status` for trellis-self and **left this product-layer lifecycle standing** |

- **AC6 — Trellis-lite needs no bespoke artifact.** The behavioral subset is expressible
  purely as a profile — `payload_depth: expressed-only` with every pipeline gene `active: false` —
  so a lighter Trellis is a *fill* of this schema, not a new artifact type. **A proposal to mint a
  standalone "lite" rule-list fails this criterion**, and is the check to run when one is proposed.
- **AC7 — one object, four consumers.** The *same* profile schema is what Assess writes, the human
  ratifies, Apply reads, and a cross-instance diff compares: **no consumer may require a field the
  others do not produce.** A field added to serve one consumer alone breaks this. It is the
  standing test on every amendment to §2, and the round-trip test named in
  `profiles/trellis-self.md`'s own open questions.

## Open questions

**The consumer is whoever next authors a catalog or a profile, or amends this schema** — reading
this file is not optional for them, which is the difference between an address and a shelf. The
catalog is blunt that *"someone reading this file later" is not a consumer*
(`inv-no-orphan-followups`), and that is the right test; it is met here by the reader being
*required*, not by the file being *available*. `profiles/trellis-self.md` states the same thing
positively — *"every artifact's `## Open questions` rides a consumer that must read it"* — and
`decision-0078`, the record that minted the invariant, drafted exactly this shape as an **honored**
example before dropping it *"for readout budget"*, not for being wrong.

Each item below is therefore a constraint on the next amendment of this schema. **None is a debt
owed to an unbuilt component** — that is the shape the same profile row names as the repo's one
live orphan, and adding more of them would make the reference instance worse at the invariant this
section reasons under.

- **Axis-B granularity.** §2 makes Axis A/B **instance-level**. That was a revision of
  `decision-0016`'s original per-invariant framing, and it is **not an open tension** — 0016 was
  amended to match *"before ratification"* and now states the instance-level form itself. What
  stays open is forward-looking: if `+latent` genes ever need per-gene presence (active / latent /
  absent as three states), §2's `active` boolean must widen, and widening it re-opens the
  granularity question 0016 settled. Settle it deliberately rather than by widening a field.
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
- **An `approved` state, below `ratified`.** §3 stops at `draft → ratified`. A ratified profile that
  Apply has *composed* was `spec-0002`'s candidate first user of an execution-layer `approved`.
  **This is dominated by, but not the same as, `decision-0082`'s open question** *"Should
  `spec-0002`'s profile lifecycle follow?"*: retiring the lifecycle would moot this one, but
  resolving that question as *keep* brings it straight back — so it is recorded here rather than
  left to return homeless.
- **The `C1` / `C2` field names.** `decision-0038` deferred renaming `default_C1` / `default_C2` to
  `default_strength` / `default_gatekeeper` *"when `spec-0002` is next opened"*. **That trigger has
  fired — this file is the opening — and the rename is deliberately not done here**, because
  `decision-0038` scopes it as a *"rubric/agent/catalog/profile cascade"*, which is its own change.
  Re-triggered, not silently re-parked: **the next change that already touches the catalog, the
  profile and the rubric together** should carry it.

*One of `spec-0002`'s six is deliberately not repeated here, because repeating it would fork a live
entry (`decision-0028`): the structure of detection heuristics is
`core/catalog/signature-catalog-v1.md`'s own open question ("Structured signatures"), which
`profiles/trellis-self.md` already counts against this repo as a live orphan.*

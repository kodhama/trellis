---
id: rubric-artifact-contract
type: rubric
status: ratified
ratified: 2026-07-03
depends_on: [spec-0001, spec-0002]
owner: gundi
scope: trellis-product
---

# Rubric — artifact-contract conformance

> The checkable gate the conformance sub-agent applies to a corpus of artifacts. Derived from
> `spec-0001` §3 (base contract) and `spec-0002` §4 (the two typed artifacts). Each item is
> **PASS / FAIL** with a *specific* reason (file + field + rule). **No vague failures, no false
> passes.**
>
> **Corpus:** `decisions/`, `specs/`, `research/`, `core/invariants/`, `core/rubrics/`,
> **`core/catalog/`**, **`core/lexicon.md`**, **`profiles/`**. Exclude `core/fixtures/` unless running
> the positive control.

## Checks

1. **Frontmatter present & required fields valid.** Every non-code `.md` artifact opens with
   YAML frontmatter carrying `id`, `type`, `status`, `depends_on`, `owner` — all present and
   well-typed (`depends_on` is a list; `status` a string; etc.). *FAIL → name the missing/
   malformed field.*
2. **`type` declared; `status` allowed.** `type` is a non-empty string carrying a `scope`
   (`core-methodology` / `trellis-product` / `trellis-meta`) and a rubric *(scope/rubric may be
   declared centrally, not per-file)*; `status` belongs to the methodology's **declared
   lifecycle** (`spec-0001` §2, `decision-0037` — for this repo:
   `{draft, ratified, superseded}`). Recognized typed artifacts include `signature-catalog`
   (`trellis-product`), `expression-profile` (`core-methodology`) — `spec-0002` — and `lexicon`
   (`trellis-product`) — `decision-0017`.
3. **`id` unique** across the corpus. *FAIL → name the colliding files.*
4. **`depends_on` resolves.** Each entry is an existing artifact `id`, a declared external-ref
   prefix (v0 allowlist: `brief-§…`), **or** a **retired id** in the invariant-set's Identifiers
   registry (mapping to a successor). *FAIL → name the dangling reference.*
5. **Directional flow (load-bearing — `inv-directional-flow`/`inv-graph-maintenance`).** No `ratified` artifact `depends_on` a `draft`
   artifact. *FAIL → name the ratified→draft edge.*
6. **Required body sections per type** (`spec-0001` §4, `spec-0002` §4): `spec`/`invariant-set` →
   Acceptance criteria + Open questions; `decision` → Context/Decision/Consequences;
   `research-note` → Open questions (+ sources); `signature-catalog` → Entries + Acceptance
   criteria + Open questions; `expression-profile` → Delivery + Profile + Assessment notes +
   Open questions; `lexicon` → Canonical terms + Open questions; `feedback` → exempt. *FAIL → name the missing section.*
7. **Supersede integrity.** A `superseded` artifact carries `superseded_by`; **revise-in-place**
   docs (specs, invariants, research, rubrics) re-point to the successor. *Exemption (`inv-auditable-archive`): an
   **append-only** `decision` may keep a dependency on the version current at its ratification
   (historical, not current-truth); a successor referencing its predecessor for diffing is also
   exempt.* *FAIL → name the offender.*

## Checks — the two typed artifacts (`spec-0002` §4)

*Apply only when a `signature-catalog` / `expression-profile` is in the corpus.*

8. **Catalog coverage + examples (`decision-0020`).** A `signature-catalog` has an entry for every
   **assessable** `invariants-v1` slug (structural + operating + floors — the 14, **excluding** the two dials; a collapsed
   slug is covered by its successor). Each entry carries `what` / **`directive`** / **`why`** /
   `signature` / **`honored`** / **`violated`** / `class` / `mechanizable` / `default_C1` / `default_C2`, and
   **`honored`/`violated` are ≥2 matched pairs** — `violated[i]` and `honored[i]` share a use-case tag,
   same order (`decision-0027`). *FAIL → name the uncovered assessable slug, an entry missing a field
   (a missing `why`/`honored`/`violated` is a fail), a `honored`/`violated` with fewer than 2, or a pair
   whose honored/violated layer tags don't align; a present dial entry is also a FAIL.*
9. **Profile → catalog resolution.** Every `expression-profile` gene `slug` resolves to a
   `signature-catalog` entry. *FAIL → name the unresolved slug (a dangling profile reference).*
10. **Evidence floor (assert-and-verify).** In a profile, every `active: true` +
    `basis: honored-implicitly` entry carries **both** a `confidence` tag and an `evidence`
    pointer. *FAIL → name the bare "honored" claim with no evidence.*
11. **Intent-gate floor (`floor-intent-gate`).** No profile sets `C2: none` on a gene whose catalog entry has
    `intent_locus: true` (`inv-intent-locus`, `floor-intent-gate`). *FAIL → name the offending
    gene.*

## Honesty clause (math-quest)

**Accurately listing the violations *is* success.** A run that hides drift to report "pass"
has failed this rubric. Missing/unparseable input → halt loudly (`floor-transparency`), never a partial pass.

## How it is graded

The conformance sub-agent emits one report: per-check PASS/FAIL, every FAIL naming the exact
file + field + rule. The check is **trusted only after it rejects the known-bad fixture**
(`core/fixtures/`, the positive control — `spec-0001` AC2).

## Acceptance criteria

- Every check above maps to a precise, file-level violation message (no vague output).
- The rubric is applied by an agent **with no runtime** (`0010`).

## Open questions

- Where `scope` and per-type `rubric` are declared (central registry vs per-file frontmatter)
  — check 2 currently allows either; resolve when the type registry is built.

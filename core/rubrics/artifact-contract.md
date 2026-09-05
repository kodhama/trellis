---
id: rubric-artifact-contract
type: rubric
status: ratified
ratified: 2026-07-03
depends_on: [invariants-v1, decision-0037, decision-0042, schema-typed-artifacts]
owner: gundi
scope: trellis-product
---

# Rubric — artifact-contract conformance

> The checkable gate the conformance sub-agent applies to a corpus of artifacts. This rubric is the
> **self-standing definition** of the artifact contract — it derived from `spec-0001` §3 and
> `spec-0002` §4 until `decision-0079` retired the spec stage, and every check below was already
> stated here in full. Each item is
> **PASS / FAIL** with a *specific* reason (file + field + rule). **No vague failures, no false
> passes.**
>
> **Corpus:** `decisions/`, `research/`, `core/invariants/`, `core/rubrics/`,
> **`core/catalog/`**, **`core/lexicon.md`**, **`profiles/`**. Exclude `core/fixtures/` unless running
> the positive control.

## Checks

1. **Frontmatter present & required fields valid.** Every non-code `.md` artifact opens with
   YAML frontmatter carrying `id`, `type`, `depends_on`, `owner` — all present and
   well-typed (`depends_on` is a list; etc.). **`status` is NOT required and NOT expected**
   (`decision-0082` retired it). *FAIL → name the missing/malformed field.*
2. **`type` declared.** `type` is a non-empty string carrying a `scope`
   (`core-methodology` / `trellis-product` / `trellis-meta`) and a rubric *(scope/rubric may be
   declared centrally, not per-file)*. **No status check** — `decision-0082` retired the field for
   trellis-self; merging is the acceptance. A legacy `status:` on any artifact predating
   `decision-0082` is **preserved history and never a violation**; do not flag it, and do not ask for it
   to be removed. Recognized typed artifacts include `signature-catalog`
   (`trellis-product`), `expression-profile` (`core-methodology`) — `schema-typed-artifacts` — and `lexicon`
   (`trellis-product`) — `decision-0017`.
3. **`id` unique** across the corpus. *FAIL → name the colliding files.*
4. **`depends_on` resolves.** Each entry is an existing artifact `id`, a declared external-ref
   form — `brief-§…`, **or** a qualified `<repo>/<id>` cross-repo reference whose `<repo>` is a
   member of the recognized registry (kodhama, trellis, grove, wisp, design-system,
   homebrew-tap, math-quest) (`decision-0044`; shape + registry-membership
   only — not verified against the referent's actual home corpus, same treatment as
   `brief-§…`) — **or** a **retired id** in the invariant-set's Identifiers registry (mapping to
   a successor) — **or** a **retired artifact id** in `decision-0079`'s retired-artifacts
   registry (`spec-0001`–`spec-0008`). That last clause is the same historical-reference
   exemption the Identifiers registry grants (`decision-0013`): a retirement does not reach back
   and edit the append-only records that cite it, so the registry — not the file's existence —
   is what makes those references resolve. A referent may carry a **`@version` pin** (shape only;
   semantics methodology-defined, `grove/adr-0010`); resolve it on **shape + the bare
   `id`/`<repo>/<id>`'s membership only** (v0, no-fetch) — the pin-vs-upstream-current *sync*
   comparison is **not** this check's (it is the operational chain's, grove `adr-0006`). *FAIL → name the dangling reference.*
5. **Directional flow (load-bearing — `inv-directional-flow`/`inv-graph-maintenance`).** For
   trellis-self the merge carries this (`decision-0082`): everything on `main` is settled, so the
   structural check is that every `depends_on` resolves within the corpus (check 4) — there is no
   status to compare. *(Where a methodology declares a status lifecycle, the original form applies:
   no gated/approved artifact `depends_on` a draft one.)* A decision's **`changes:`** relation
   (shape only) is a **forward-pointer of the `superseded_by` class,
   not a `depends_on`-class edge** — do **not** walk it as a flow edge; a spec both depending on
   its authorizing decision and named in that decision's `changes:` is a benign pair, not a cycle.
   *FAIL → name the edge.*
6. **Required body sections per type** (`schema-typed-artifacts`, `decision-0042`): `spec`/`invariant-set` →
   Acceptance criteria + Open questions; `decision` → Context/Decision/Consequences;
   `research-note` → Open questions (+ sources); `signature-catalog` → Entries + Acceptance
   criteria + Open questions; `expression-profile` → Delivery + Profile + Assessment notes +
   Open questions; `lexicon` → Canonical terms + Open questions; `feedback` → exempt. *FAIL → name the missing section.*
7. **Supersede integrity.** **Supersession is identified by the forward pointer** (`decision-0082`;
   formerly by `status: superseded`): an artifact carrying `superseded_by` is superseded, and its
   entries must resolve. **Revise-in-place** docs (invariants, research, rubrics, schemas) re-point
   to the successor. A **partially superseded** artifact stays current for its remainder and carries
   `superseded_in_part_by`, whose entries resolve like `depends_on` (`decision-0040`). *Exemption (`inv-auditable-archive`): an
   **append-only** `decision` may keep a dependency on the version current at its ratification
   (historical, not current-truth); a successor referencing its predecessor for diffing is also
   exempt.* *FAIL → name the offender.*

## Checks — the two typed artifacts (`schema-typed-artifacts`)

*Apply only when a `signature-catalog` / `expression-profile` is in the corpus.*

8. **Catalog coverage + examples (`decision-0020`).** A `signature-catalog` has an entry for every
   **assessable** `invariants-v1` slug (structural + operating + floors — **excluding** the two dials; a collapsed
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

## Check — version cross-check (retired)

12. *(Retired 2026-07-12, `grove/adr-0010` — the version cross-check is methodology semantics,
    re-homed to the operating model of the day. That model (grove-the-plugin) retired with
    `decision-0076`, so the check has no owner here and is not applied. Number retained so
    external references to "rubric check 12" resolve to this pointer rather than shifting; the
    typed checks 8–11 above are unaffected.)*

## Honesty clause (math-quest)

**Accurately listing the violations *is* success.** A run that hides drift to report "pass"
has failed this rubric. Missing/unparseable input → halt loudly (`floor-transparency`), never a partial pass.

## How it is graded

The conformance sub-agent emits one report: per-check PASS/FAIL, every FAIL naming the exact
file + field + rule. The check is **trusted only after it rejects the known-bad fixture**
(`core/fixtures/`, the positive control).

## Acceptance criteria

- Every check above maps to a precise, file-level violation message (no vague output).
- The rubric is applied by an agent **with no runtime** (`0010`).

## Open questions

- Where `scope` and per-type `rubric` are declared (central registry vs per-file frontmatter)
  — check 2 currently allows either; resolve when the type registry is built.

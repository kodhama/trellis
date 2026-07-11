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
   lifecycle** (`spec-0001` §2, `decision-0037`; adopted family-wide by
   `decision-0042` — for this repo: `{draft, gated, approved, superseded}`,
   with pre-0042 artifacts reading `ratified` = `approved`). Recognized typed artifacts include `signature-catalog`
   (`trellis-product`), `expression-profile` (`core-methodology`) — `spec-0002` — and `lexicon`
   (`trellis-product`) — `decision-0017`.
3. **`id` unique** across the corpus. *FAIL → name the colliding files.*
4. **`depends_on` resolves.** Each entry is an existing artifact `id`, a declared external-ref
   form — `brief-§…`, **or** a qualified `<repo>/<id>` cross-repo reference whose `<repo>` is a
   member of the recognized registry (kodhama, trellis, grove, wisp, design-system,
   homebrew-tap, math-quest) (`spec-0001` §1, `decision-0044`; shape + registry-membership
   only — not verified against the referent's actual home corpus, same treatment as
   `brief-§…`) — **or** a **retired id** in the invariant-set's Identifiers registry (mapping to
   a successor). A referent may carry a **`@version` pin** (`spec-0001` §1, `decision-0045`);
   resolve it on **shape + the bare `id`/`<repo>/<id>`'s membership only** (v0, no-fetch) — the
   pin-vs-upstream-current *sync* comparison is **not** this check's (it is grove#34 /
   `adr-0006`'s). *FAIL → name the dangling reference.*
5. **Directional flow (load-bearing — `inv-directional-flow`/`inv-graph-maintenance`).** No `gated`/`approved` (or legacy
   `ratified`) artifact `depends_on` a `draft` artifact. A decision's **`changes:`** relation
   (`spec-0001` §3, `decision-0045` item 7) is a **forward-pointer of the `superseded_by` class,
   not a `depends_on`-class edge** — do **not** walk it as a flow edge; a spec both depending on
   its authorizing decision and named in that decision's `changes:` is a benign pair, not a cycle.
   *FAIL → name the edge.*
6. **Required body sections per type** (`spec-0001` §4, `spec-0002` §4): `spec`/`invariant-set` →
   Acceptance criteria + Open questions; `decision` → Context/Decision/Consequences;
   `research-note` → Open questions (+ sources); `signature-catalog` → Entries + Acceptance
   criteria + Open questions; `expression-profile` → Delivery + Profile + Assessment notes +
   Open questions; `lexicon` → Canonical terms + Open questions; `feedback` → exempt. *FAIL → name the missing section.*
7. **Supersede integrity.** A `superseded` artifact carries `superseded_by`; **revise-in-place**
   docs (specs, invariants, research, rubrics) re-point to the successor. A **partially
   superseded** artifact keeps its pre-supersession status (`approved`, or legacy
   `ratified`) and carries `superseded_in_part_by`, whose
   entries resolve like `depends_on` (`spec-0001` §2, `decision-0040`). *Exemption (`inv-auditable-archive`): an
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

## Check — version cross-check (base contract, `spec-0001` §3 check 8, `decision-0045`)

*Placed here, not renumbered into the base checks 1–7, to avoid shifting the `spec-0002` typed
checks 8–11 above — it is a base-contract check (`spec-0001` §3), applied only when a
significant-change `decision` carries a `changes:` field.*

12. **Version cross-check (partial).** **Scope: behavioral / counter-versioned artifacts only**
    (the ordered `vN` form — `decision-0045` item 6; a content-hash has no ordering and cross-repo
    tags are the sync check's, both out of scope). Where a significant-change `decision` carries
    `changes: [X@vN]`, reconcile against `X`'s counter **record**, **not** `declared == current`:
    an append-only decision's `@vN` legitimately sits behind `X`'s current counter after a later
    bump. *FAIL → a **declared change that never landed**: `X`'s current counter is behind `vN`.* A
    bump in `X` with no accounting `changes:` decision is **not** a FAIL (`decision-0045`'s
    open question — "must every significant change flow from a decision?" — is unsettled). Bounded
    intra-repo frontmatter-vs-record audit; **distinct** from the pin-vs-current *sync* check
    (check 4), which is grove#34 / `adr-0006`'s.

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

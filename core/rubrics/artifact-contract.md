---
id: rubric-artifact-contract
type: rubric
status: draft
depends_on: [spec-0001]
owner: gundi
scope: trellis-product
---

# Rubric — artifact-contract conformance

> The checkable gate the conformance sub-agent applies to a corpus of artifacts. Derived from
> `spec-0001` §3. Each item is **PASS / FAIL** with a *specific* reason (file + field + rule).
> **No vague failures, no false passes.**

## Checks

1. **Frontmatter present & required fields valid.** Every non-code `.md` artifact opens with
   YAML frontmatter carrying `id`, `type`, `status`, `depends_on`, `owner` — all present and
   well-typed (`depends_on` is a list; `status` a string; etc.). *FAIL → name the missing/
   malformed field.*
2. **`type` declared; `status` allowed.** `type` is a non-empty string carrying a `scope`
   (`core-methodology` / `trellis-product` / `trellis-meta`) and a rubric *(scope/rubric may be
   declared centrally, not per-file)*; `status ∈ {draft, ratified, superseded}` (v0 — no
   `approved` yet, `spec-0001` §2).
3. **`id` unique** across the corpus. *FAIL → name the colliding files.*
4. **`depends_on` resolves.** Each entry is an existing artifact `id`, a declared external-ref
   prefix (v0 allowlist: `brief-§…`), **or** a **retired id** in the invariant-set's Identifiers
   registry (mapping to a successor). *FAIL → name the dangling reference.*
5. **Directional flow (load-bearing, A1/B1).** No `ratified` artifact `depends_on` a `draft`
   artifact. *FAIL → name the ratified→draft edge.*
6. **Required body sections per type** (`spec-0001` §4): `spec`/`invariant-set` →
   Acceptance criteria + Open questions; `decision` → Context/Decision/Consequences;
   `research-note` → Open questions (+ sources); `feedback` → exempt. *FAIL → name the
   missing section.*
7. **Supersede integrity.** A `superseded` artifact carries `superseded_by`; **revise-in-place**
   docs (specs, invariants, research, rubrics) re-point to the successor. *Exemption (B4): an
   **append-only** `decision` may keep a dependency on the version current at its ratification
   (historical, not current-truth); a successor referencing its predecessor for diffing is also
   exempt.* *FAIL → name the offender.*

## Honesty clause (math-quest)

**Accurately listing the violations *is* success.** A run that hides drift to report "pass"
has failed this rubric. Missing/unparseable input → halt loudly (D1), never a partial pass.

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

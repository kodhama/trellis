---
id: decision-0045
type: decision
status: approved  # ratified by PR #144 merge (2026-07-11); this commit is the post-merge bump
depends_on: [decision-0037, decision-0044, decision-0014, spec-0001]
owner: agent
superseded_in_part_by: [grove/adr-0010-versioning-is-operational]  # 2026-07-12 — Consequences 1–3 only (the execution HOME: grammar in spec-0001, cross-check in trellis's corpus-reviewer); the decision content (kinds, forms, counter, pins) stands as the origin record — semantics now evolve in the grove operating model
date: 2026-07-11
---

> Shaped interactively with the maintainer (2026-07-11), resolving
> kodhama/trellis#143 — the principles-layer piece of the
> artifact-conformance program (kodhama#31). Passed an independent
> `spec-adversary` round (NEEDS-REVISION → this revision): the spine held,
> several rationales were corrected to match their own sources. Operational
> application (the sync *check*, the decision→spec→tests→code chain) is
> grove#34, downstream.

# 0045 — Artifact-versioning kinds; the version form fits what "conform" means; pin-vs-current conformance

## Context

trellis is the type-agnostic contract layer: per `decision-0037`, types
(including "spec") are *methodology-defined* — trellis knows only
"artifacts" with a declared type/scope/rubric and a `depends_on`. This
decision extends the artifact contract so that "is a consumer still in
sync with the upstream it built against?" becomes a **derived, checkable**
question. It leans on `inv-ratifiable-artifacts` (which trellis itself
flags "strong, less settled" — a real cost, named).

**Precedents, read directly (three forms, not two).** The family already
versions artifacts three different ways, and the difference is instructive:
- trellis's **payload** — a **content-hash** (`payload@<12-hex>`,
  `decision-0043`). It's a vendored bundle where **byte-identity is
  correctness**; the hash is derived, automatic, bumps on any byte.
- design-system's **tokens** — **human-cut git tags** (`vX.Y.Z`;
  `design-system/README.md`: "Version = git tags … the tag *is* the
  version", "an untagged commit has no stable identity"). *Not* a
  content-hash (an earlier draft of this decision wrongly said so): a
  human-judgment-cut release marker, which sits closer to the behavioral
  model than to the payload's hash.
- (proposed here) **specs** — an agent-generated significance counter
  (below).

The lesson: **the version form is not a clean two-way function of "kind."
It fits what "conform" means for the artifact and how its significance can
be judged** — and that ranges over at least byte-identity (hash),
human-cut release (tag), and agent-judged behavioral significance
(counter).

## Decision

1. **Generalize "artifact."** Code, pages, mockups, designs are artifacts
   under the contract — each with a type/scope/rubric and a `depends_on`
   pointing upstream. "Code out of sync with its spec" is then the
   *existing* `depends_on`/directional-flow relationship extended to a new
   type, not a bespoke mechanism.

2. **Two artifact-versioning kinds — stated abstractly** (trellis names no
   concrete type):
   - **Append-only** — immutable; versions *implicitly* (the id already
     pins a unique state; supersession is the history). trellis's existing
     supersede discipline (`spec-0001` §2/§3.7).
   - **Versioned / revise-in-place** — mutable; carries an *explicit
     version stamp*, because the id alone doesn't identify *which* state a
     consumer built against.

3. **Pin-vs-current conformance.** A downstream `depends_on` pins a version
   of a versioned upstream; conformance = compare pinned vs. current —
   **derived, not a self-reported "in-sync" boolean.** The pin is the
   **versioned artifact's own version marker** (a counter, a git tag —
   whatever form fits the kind), **not a foreign decision-id**. *(The
   reason is granularity, corrected from an earlier draft: it is not that
   a past state is "unreconstructable" — git recovers a state whether you
   label it with a stamp or a decision's merge-commit. It is that the pin
   must be the artifact's **own** comparison-currency at the artifact's own
   granularity; a decision-id is a foreign, coarser marker shared across
   whatever it changed. design-system's precedent already pins a git tag —
   the artifact's own marker — not the decision that cut it.)*

4. **The concrete type→kind mapping is family-level and already exists** —
   not asserted here. Versioned ↔ specs (grove `adr-0004`, revise-in-place)
   and ↔ invariant-sets (`decision-0014`, a revise-in-place spec that is
   *not* a `spec` type); append-only ↔ decisions. This decision names the
   *kinds* and the *rule*; the family already says which of its types are
   which — and that classification is **incomplete** (see Open questions:
   several seed types, and dual-consumed artifacts, are unclassified).

5. **The version form fits what "conform" means for the artifact.** One
   primitive (stamp + pin + compare); the stamp's form is chosen by what
   "conform" means and how significance is judged:
   - **Byte-identity** (vendored bundles — trellis's `payload`): a
     **content-hash** (derived, automatic, can't lie, bumps on any byte).
     Unchanged; already so.
   - **Behavioral** (specs): an **agent-generated significance counter**
     (item 6).
   - **Human-cut release** (design-system's tokens today): **git tags**
     (`vX.Y.Z`). A distinct, human-judgment-gated form; whether it should
     migrate toward the agent-generated counter is out of scope here.
   - A content-hash on a behavioral contract answers the wrong question
     ("did any byte change?" incl. a typo fix) — so byte-identity and
     behavioral forms are genuinely different, even if the two-way "kind"
     framing is too crude for the full spectrum.

6. **The behavioral (spec) version is an agent-generated, review-bounded
   significance counter.** A plain monotonic counter (`v1, v2, …`) — an
   ordering, nothing more (major/minor semantics would force a fallible
   breaking-vs-additive call for zero sync benefit; a consumer re-verifies
   on any bump regardless). It is **agent-generated**, not human-curated
   (tedious, and the ADR-0030 self-report failure) — the agent bumps it
   when a change is behaviorally significant.
   - **Significance has a semi-mechanical anchor, not pure vibes:** because
     specs are GWT/EARS-structured (`adr-0004` / grove#21), *"did a
     testable clause — a scenario or an invariant — change?"* is the
     signal; a prose-only edit does not bump. Agent interpretation covers
     the fuzzy edges (a clause changed trivially; prose that carries real
     behavior).
   - **Honest epistemic status — bounded, not "can't lie."** Behavioral
     *significance is inherently a judgment*: no mechanical rule fully
     distinguishes a rewrite from a typo (the very reason a content-hash is
     wrong here). So this version is a **claim, bounded by independent
     review** (a reviewer sanity-checks that bumps track real
     significance) — *not* the "derived, can't-lie" property a hash has.
     Full derivation is not achievable for a behavioral version; bounded
     judgment is the ceiling. *(This corrects an earlier draft that claimed
     audit gave "the same can't-lie guarantee as derivation" — an audit
     detects drift after the fact; it does not prevent it.)*
   - **The change-without-a-decision residual (named, not hidden).** If a
     behaviorally-significant change is made with **no** decision *and* the
     agent fails to bump, it escapes detection. That is exactly grove#20's
     process-gap class; its enforcement is the artifact-gated dispatch +
     strict-TDD discipline of grove `adr-0005` and the sync check of
     grove#34 — **not** this decision. This decision does not guarantee
     completeness; it defines the version and names its dependency on that
     upstream discipline.
   - **Decision record as a *partial cross-check*, not the source.** Where
     a significant change *did* file a decision, the version and the record
     should agree — a `corpus-reviewer` reconciliation that catches
     decision-without-bump / bump-without-decision. It is a bounded audit,
     not the version's definition (not every significant change need flow
     from a decision — see Open questions).

7. **`changes:` is a distinct forward-pointer relation.** For the partial
   cross-check (item 6) a significant-change decision may declare which
   artifact(s) it changes. This is a **forward-pointer relation of the
   `superseded_by` class — never a reused `depends_on` edge.** A spec both
   `depends_on` its authorizing decision *and* being named in that
   decision's `changes:` is a benign two-relation pair only if `changes:`
   is not graph-typed as a `depends_on`-class edge; typed as forward-only,
   it raises no cycle against directional flow (`spec-0001` §3 check 5).

8. **Pin shape:** `repo/id@version` (and `id@version` local), extending
   `decision-0044`'s qualified `repo/id` form. Exact spelling (and the
   `@` delimiter's collision-safety, which `decision-0044` checked for
   `/`/`:`) is a `spec-0001`-amendment detail; this decision fixes shape.

## Considered and rejected

- **Pin = a decision-id** — rejected because the pin must be the
  artifact's *own* comparison-currency at its own granularity, not a
  foreign, coarser marker shared across everything a decision touched.
  *(Not* rejected for "unreconstructability" — an earlier draft's reason,
  false: git recovers a past state from either a stamp-label or a
  decision's commit.)
- **One uniform version form for all kinds** — rejected: it gets a kind
  wrong (a hash is noisy on a behavioral spec; a significance counter is
  meaningless for a byte-identical bundle). The real space is a spectrum
  (hash / release-tag / significance-counter), not one form.
- **A fully-derived spec version** (e.g. count of significant-change
  decisions) — rejected: it presumes every significant change flows from a
  decision, which is *not* established, and behavioral significance is a
  judgment no mechanism fully computes. Derivation is unachievable here;
  agent-generated + review-bounded is the honest ceiling.
- **Claiming "audit gives can't-lie"** — retracted as an overclaim: audit
  *bounds* drift (catches it at the next run), it does not *eliminate* the
  possibility of a wrong bump, because the underlying significance call is
  a judgment.
- **A self-reported "in-sync" boolean** — rejected at the root:
  conformance must be *derived by comparison* (pin vs. current), never a
  claim of the answer.
- **Semver for the behavioral version** — rejected for a plain counter:
  sync detection needs an ordering, not compatibility semantics; major/
  minor is a fallible agent judgment for zero sync benefit. (Revisitable
  if a real compatibility-signalling need ever emerges — not on spec.)

## Consequences

None executed by this decision — a follow-on `contract-author` pass amends
`spec-0001` once this is `approved`:

> *[Superseded in part 2026-07-12 by `grove/adr-0010-versioning-is-operational`
> — scope: Consequences 1–3 (per-clause markers below). The decision content
> above stands as the origin record; Consequence 4 stands untouched.]*

1. **`spec-0001` §1/§2 amendment** — the versioned kind's frontmatter gains
   a version marker; the `depends_on` grammar gains the `repo/id@version`
   pin form (extending `decision-0044`). The generated/vendored kind's
   content-hash and design-system's git tags stay as-is.
   *[Superseded 2026-07-12 by `grove/adr-0010` — the grammar written into
   `spec-0001` is reduced to shape-only clauses; the semantics' single home
   is the installed `.grove/versioning.md` companion.]*
2. **`changes:` / `affects:` forward-pointer relation** on significant-change
   decisions (for the partial cross-check), typed distinct from
   `depends_on`. Exact form settled in the amendment.
   *[Superseded 2026-07-12 by `grove/adr-0010` — the relation's shape (edge
   class) stays in `spec-0001` §1; its reconciliation semantics live in the
   companion.]*
3. **`corpus-reviewer` gains the partial version cross-check** — where a
   significant-change decision declares it changes a behavioral artifact,
   reconcile its declared version against the record; a mismatch is a
   finding. A *bounded* audit (this is the frontmatter-vs-record check,
   distinct from the consumer-vs-upstream *sync* check, which is grove#34).
   *[Superseded 2026-07-12 by `grove/adr-0010` — the cross-check is owned by
   the operating model's `corpus-reviewer` (grove#51), defined in the
   companion; trellis's spec-0001 §3 check 8 / rubric check 12 are retired
   to pointers.]*
4. **The concrete mapping is unchanged** — specs revise-in-place
   (`adr-0004`), decisions append-only; no repo restates it.

## Open questions

- **Must every significant spec change flow from a decision?** If yes, the
  partial cross-check strengthens toward derivation; if no, the
  agent-generated + review-bounded version stands on its own. Same shape as
  grove#20/#21 one level up. Open — not settled here.
- **Dual-consumed artifacts** (a **charter** is consumed *both* by
  vendoring/byte-identity *and* by conformance/behavioral — and grove#34
  acts on exactly this artifact): does it carry *two* markers (a byte-hash
  *and* a behavioral counter)? The two-kind partition doesn't answer this;
  0045's own "form fits what conform means" implies it might. Left to
  grove#34 / the amendment.
- **Unclassified seed types** — `plan`, `tasks`, `research-note`,
  `feedback`, `rubric` (`spec-0001` §1) have no declared kind; the
  family-level classification is incomplete. Not this decision's to
  complete (it's methodology-defined), but flagged.
- **The pin-vs-current *sync* check** — how a consumer's pin is compared to
  its upstream's current marker, and where it runs — is operational, parked
  to **grove#34**.
- **Execution-layer `approved`** (status vs. gate-outcome, `spec-0001` §2)
  — pairs with **trellis#142** + **trellis#25**; not this decision.
- **Exact spellings** (version field, `changes:`, `@version` delimiter) —
  the follow-on `spec-0001` amendment's to fix.

## Self-check (gate)

- **Frontmatter**: `id`/`type`/`status`/`depends_on`/`owner`/`date`
  present, well-typed. PASS.
- **`depends_on` resolution**: `decision-0037` (`ratified`), `decision-0014`
  (added this revision — the invariant-set-as-revise-in-place precedent,
  `ratified`), `decision-0044` (`approved`), `spec-0001` (`ratified`) — all
  resolve, none `draft`. PASS.
- **Directional flow**: `gated`; every dependency is ratified/approved,
  not draft. PASS.
- **Required body sections** (`spec-0001` §4): Context/Decision/Consequences
  present, plus Considered-and-rejected, Open questions, Self-check. PASS.
- **Append-only discipline**: new artifact; nothing edited in place; no
  ratified decision superseded (it *extends* the contract). N/A.
- **Adversary round**: `spec-adversary` returned NEEDS-REVISION; findings
  A (design-system form), B (can't-lie overclaim), C (change-without-
  decision hole), D (unreconstructable reasoning), E (`changes:` edge
  type), F (dual-consumed/unclassified) each corrected or named above.
  Verdict was "spine sound, rationales need to match sources" — done.
- **Approval mechanic**: ratified by the maintainer's approval, recorded
  the trellis-native way (`decision-0042`) — the **PR #144 merge is the
  ratification act**; this post-merge commit records `approved`. No in-PR
  self-approval (`ratify-guard`-compliant); no deviation from the written
  mechanic. PASS.

**Overall: internally sound, consumable, and `approved`** — self-checked,
survived two `spec-adversary` rounds, ratified by the maintainer
(PR #144 merge, 2026-07-11). Closes kodhama/trellis#143 and authorizes
the `spec-0001` amendment (Consequences).

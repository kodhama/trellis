---
id: decision-0045
type: decision
status: gated  # self-checked 2026-07-11, shaping converged — see Self-check
depends_on: [decision-0037, decision-0044, spec-0001]
owner: agent
date: 2026-07-11
---

> Shaped interactively with the maintainer (2026-07-11), resolving
> kodhama/trellis#143 — the principles-layer piece of the
> artifact-conformance program (kodhama#31). This decision fixes the
> type-agnostic primitives; their operational application (the
> decision→spec→tests→code sync chain, and the check that runs it) is
> grove#34, downstream.

# 0045 — Two artifact-versioning kinds (append-only/implicit vs. versioned/explicit); the version form follows the kind; pin-vs-current conformance

## Context

trellis is the type-agnostic contract layer: per `decision-0037`, types
(including "spec") are *methodology-defined* — trellis knows only
"artifacts" with a declared type/scope/rubric and a `depends_on`. This
decision extends the artifact contract so that "is a consumer still in
sync with the upstream it built against?" becomes a **derived, checkable**
question rather than a matter of trust — the structural prerequisite
`inv-ratifiable-artifacts` names (and which trellis itself flags "strong,
less settled" — a real cost, named here, not hidden).

The motivating gap (kodhama#31): an `approved` upstream whose downstream
(code, a vendored copy) hasn't caught up is a silent nonconformance. The
resolution is not new machinery but a generalization: treat the downstream
as *itself an artifact* whose `depends_on` pins a *version* of its
upstream, and make "out of sync" the comparison of pinned-vs-current.

**Precedent, not green-field.** design-system already does this for one
artifact kind — current-state assets + explicit git-tag versions
(`vX.Y.Z`), consumers pin the tag, "current?" = is the pin behind the
latest tag. And trellis already versions its own `payload` by content-hash
(`payload@<12-hex>`). Both are the same shape (stamp + pin + compare) on
different kinds; this decision names the shape and the two kinds.

## Decision

1. **Generalize "artifact."** Code, pages, mockups, designs are artifacts
   under the contract — each with a type/scope/rubric and a `depends_on`
   pointing upstream. "Code out of sync with its spec" is then the
   *existing* `depends_on`/directional-flow relationship extended to a new
   type, not a bespoke mechanism.

2. **Two artifact-versioning kinds — stated abstractly** (trellis names no
   concrete type):
   - **Append-only** — immutable; versions *implicitly* (the id already
     pins a unique state; supersession is the history). This is trellis's
     existing supersede discipline (`spec-0001` §2/§3.7).
   - **Versioned / revise-in-place** — mutable; carries an *explicit
     version stamp*, because the id alone doesn't identify *which* state a
     consumer built against.

3. **Pin-vs-current conformance.** A downstream `depends_on` pins a
   version of a versioned upstream; conformance = compare pinned vs.
   current — **derived, not a self-reported "in-sync" boolean** (which
   could lie; cf. wisp's ADR-0030). The pin is the versioned artifact's
   **own stamp, never a decision-id**: a revise-in-place artifact retains
   no deltas, so "its state as of decision X" is unreconstructable — only
   an explicit stamp (+ git) recovers a past state.

4. **The concrete type→kind mapping is family-level and already exists** —
   not asserted here. Versioned ↔ specs (grove `adr-0004`: specs are
   revise-in-place); append-only ↔ decisions (each repo's
   `decisions/README`). This decision names the *kinds* and the *rule*;
   the family already said which of its types are which.

5. **The version *form* follows the artifact kind.** One primitive
   (stamp + pin + compare); the stamp's form is chosen by what "conform"
   means for the kind:
   - **Byte-identity kinds** (generated / vendored — trellis's `payload`,
     design-system's tokens): a **content-hash**. Byte-identity *is*
     correctness there; derived, automatic, can't lie, bumps on any
     content change. (Unchanged — these already work this way.)
   - **Behavioral kinds** (specs): a **significant-change version** —
     bumps only on a *significant* (behavioral) change, which per
     `adr-0004` already gets a decision; editorial edits don't bump. So
     pin-vs-current tracks *behavioral* drift, not every byte. A
     content-hash on a behavioral contract answers the wrong question
     ("did any byte change?" incl. a typo fix) and cries wolf.

6. **The behavioral (spec) version is *stamped + audited*.** A static
   frontmatter field (`version: N`) — readable, pinnable — **reconciled by
   a `corpus-reviewer` check against the append-only decision record** (the
   significant-change decisions affecting the artifact). Drift is caught,
   so the stamp can't silently lie: static-field simplicity *and* the
   ADR-0030 "derived truth over self-report" property, via audit rather
   than pure derivation.
   - **Corollary (new requirement):** a **significant-change decision must
     declare which artifact(s) it changes**, so the audit has something to
     count against. The exact field is left to the amendment (below).

7. **Pin shape:** `repo/id@version` (and `id@version` for a local pin),
   extending `decision-0044`'s qualified `repo/id` cross-repo form. Exact
   spelling is a `spec-0001`-amendment detail; this decision fixes the
   shape.

## Considered and rejected

- **Pin = a decision-id** (pin the decision that last changed the
  upstream) — rejected: a revise-in-place artifact retains no deltas, so
  the decision can't reconstruct the pinned state. The stamp must be on
  the artifact.
- **One uniform version form for all kinds** — rejected: it gets one kind
  wrong. A content-hash is noisy on a behavioral spec (flags editorial
  edits); a significant-change version is meaningless for a vendored
  bundle where byte-identity *is* correctness.
- **Spec version stamped-only (no audit)** — rejected: a bare
  self-reported field can drift silently (decision filed, bump forgotten),
  the exact ADR-0030 failure. Audit closes it.
- **Spec version fully-derived (computed, no static field)** — rejected in
  favour of stamped+audited: derivation can't lie either, but it's not a
  static value a consumer can read off the file and pin against, and it's
  heavier. Stamped+audited keeps the static field and gets the same
  can't-lie guarantee via reconciliation.
- **A self-reported "in-sync" boolean** — rejected at the root: conformance
  must be *derived* (compare pin vs. current), never a claim.

## Consequences

None executed by this decision itself — a follow-on `contract-author` pass
amends `spec-0001` once this is `approved`:

1. **`spec-0001` §1/§2 amendment** — the versioned kind's frontmatter
   contract gains the version-stamp field; the `depends_on` grammar gains
   the `repo/id@version` pin form (extending `decision-0044`). The
   generated/vendored kind's content-hash stays as-is (already in use).
2. **Decisions declare what they change** — a `changes:` / `affects:`
   frontmatter list (or a reused relation) on significant-change
   decisions, so the version audit can reconcile. Exact form settled in
   the amendment.
3. **`corpus-reviewer` gains the version audit** — for a versioned
   behavioral artifact, its `version` reconciles against the
   significant-change decisions declaring they affect it; a mismatch is a
   finding. (This is the *frontmatter-vs-record* audit — distinct from the
   *pin-vs-current sync* check between a consumer and its upstream, which
   is operational and lives in grove#34.)
4. **The concrete mapping is unchanged** — specs already revise-in-place
   (`adr-0004`), decisions already append-only. No repo restates it.

## Open questions

- **The pin-vs-current *sync* check** — how a consumer's pin is compared
  to its upstream's current stamp, and where that check runs
  (`conformance-reviewer` / a mode of it / new) — is *operational*, parked
  to **grove#34**. This decision fixes only the primitive it acts on.
- **Does a passing sync check materialize the deferred execution-layer
  `approved`** (`spec-0001` §2 — a status, or a gate-outcome)? Pairs with
  **trellis#142** (the intent-act half) and **trellis#25**; not this
  decision.
- **Exact spellings** — the version field name, the `changes:`/`affects:`
  field, the `@version` delimiter — are the follow-on `spec-0001`
  amendment's to fix; this decision fixes shapes, not spellings.

## Self-check (gate)

- **Frontmatter**: `id`/`type`/`status`/`depends_on`/`owner`/`date`
  present, well-typed. PASS.
- **`depends_on` resolution**: `decision-0037` (`ratified`), `spec-0001`
  (`ratified`), `decision-0044` (`approved`) — all resolve in this repo,
  none `draft`. PASS.
- **Directional flow**: this artifact is `gated`; every dependency is
  ratified/approved, not draft. No `gated`→`draft` violation. PASS.
- **Required body sections** (`spec-0001` §4, `decision` →
  Context/Decision/Consequences): present, plus Considered-and-rejected,
  Open questions, this Self-check. PASS.
- **Append-only discipline**: new artifact; nothing edited in place; no
  ratified decision superseded (it *extends* the contract). N/A — no
  violation possible.
- **Approval mechanic**: left `gated`, **not** flipped to `approved`. The
  design is settled but ratification (the maintainer's act) is owed — this
  record does not pre-empt it. PASS.

**Overall: internally sound, consumable, and `gated`** — self-checked,
agent-consumable, awaiting the maintainer's approval, which closes
kodhama/trellis#143 and authorizes the `spec-0001` amendment (Consequences).

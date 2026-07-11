---
id: decision-0045
type: decision
status: draft
depends_on: [decision-0037, decision-0044, spec-0001]
owner: agent
date: 2026-07-11
---

> **Draft — shaping canvas for kodhama/trellis#143.** The principles-layer
> piece of the artifact-conformance program (kodhama#31). Being shaped
> interactively with the maintainer; the `## Decision state` below is the
> live record — Decided / Open / Parked.

# 0045 — Two artifact-versioning kinds: append-only (implicit) vs. versioned/revise-in-place (explicit), and pin-vs-current conformance

## Decision state

**Decided** (maintainer-affirmed in the kodhama#31 shaping conversation,
2026-07-11):
- **Generalize "artifact."** Code, pages, mockups, designs are all
  artifacts under the contract — each with a type/scope/rubric and a
  `depends_on` pointing upstream. Consistent with `decision-0037`'s open
  `type` field. "Code out of sync with its spec" is then the *existing*
  `depends_on`/directional-flow check extended, not new machinery.
- **Two versioning kinds — stated abstractly, trellis names no concrete
  type** (`decision-0037`: types are methodology-defined; trellis knows
  only "artifacts"):
  - **Append-only** — immutable; versions *implicitly* (the id already
    pins a unique state; supersession is the history). This is trellis's
    existing supersede discipline (`spec-0001` §2/§3.7).
  - **Versioned / revise-in-place** — mutable; needs an *explicit version
    stamp*, because the id alone doesn't identify *which* state a consumer
    built against.
- **Pin-vs-current conformance.** Downstream `depends_on` pins a version
  of a versioned artifact; conformance = compare pinned vs. current —
  **derived, not a self-reported "in-sync" boolean** (which could lie;
  cf. wisp's ADR-0030).
- **The concrete mapping is NOT this decision's** — it's family-level and
  already exists (versioned ↔ specs, per grove/adr-0004; append-only ↔
  decisions, per each `decisions/README`). This decision names the *kinds*
  and the *rule*, abstractly.
- **Precedent, not green-field:** design-system already does this for one
  artifact kind — current-state assets + explicit git-tag versions
  (`vX.Y.Z`) + consumers pin the tag. Generalize that to the artifact
  graph as a named kind.
- **The pin is the versioned artifact's own stamp, not a decision-id** —
  a revise-in-place artifact retains no deltas, so "its state as of
  decision X" is unreconstructable; only an explicit stamp (+ git)
  recovers a past state.
- **The version *form follows the artifact kind*** (2026-07-11) — the
  primitive is one (stamp + pin + pin-vs-current); the *form* is chosen
  by what "conform" means for that kind:
  - **Byte-identity kinds** (generated / vendored artifacts — trellis's
    `payload`, design-system's tokens): a **content-hash** (trellis
    already does exactly this, `payload@<12-hex>`). Byte-identity *is*
    correctness there; the hash is derived, automatic, can't lie, and
    bumps on any content change.
  - **Behavioral kinds** (specs): a **significant-change version** —
    bumps only when a *significant* (behavioral) change happens, which
    per `adr-0004` already gets a decision; editorial edits don't bump.
    So pin-vs-current tracks *behavioral* drift, not every byte, and the
    decision explains each bump. This avoids the content-hash's
    false-positive noise (flagging behavior-neutral edits) on the kind
    where behavior is what matters.
  - Rationale: a content-hash answers "did any byte change?"; for a
    behavioral contract the real question is "did anything I depend on
    change?" — and the family already draws that line (`adr-0004`
    significant vs. editorial). Forcing one form on both kinds gets one
    of them wrong.

**Open** (live design questions — the substance still to shape):
1. **Is the spec (behavioral) version *stamped* or *derived*?** A written
   frontmatter field (`version: 3`), bumped by hand when the
   significant-change decision is filed — can drift (decision filed,
   bump forgotten); vs. *derived* from the decision record (e.g. a
   function of the significant-change decisions affecting the artifact) —
   can't drift, but needs decisions to declare what they change and isn't
   a static value you read off the file. (The "can it lie?" question the
   `ADR-0030` principle cares about.)
2. **Stamp location + pin syntax** — proposed (to confirm): the stamp is a
   `spec-0001` frontmatter field on the versioned artifact; a pin extends
   `decision-0044`'s qualified form to `repo/id@version` (and `id@version`
   local). Low-fork unless the derive-vs-stamp answer (1) reshapes it.

**Parked** (moved out of this decision):
- **The check mechanism** — how the conformance check reads current-stamp
  vs. pinned and where it lands (corpus-reviewer / conformance-reviewer /
  new) — is *operational*, so it belongs in **grove#34**, not this
  principles decision. Recorded there.

**Parked** (out of scope for this decision):
- The *operational* application of these kinds — the decision→spec→tests→
  code sync chain — is grove#34's shaping run, downstream of this.
- Whether the conformance verdict materializes an execution-layer
  `approved` status vs. a gate-outcome (`spec-0001` §2's deferred
  question) — pairs with trellis#142 and #25; not this decision.

## Context

The principles piece of the conformance program (kodhama#31). trellis is
the type-agnostic contract layer; this decision extends its artifact
contract with (a) the recognition that any typed thing (code included) is
an artifact, and (b) the versioned-vs-append-only distinction that makes
"is a consumer in sync with its upstream" a *derived*, checkable question.
It leans on `inv-ratifiable-artifacts` (which trellis itself flags "strong,
less settled") — a real cost, named. Full derivation of the settled points
is in kodhama#31's own conversation trail and in kodhama/trellis#143.

## Decision

*(filled in as the Open questions above converge — not asserted ahead of
the maintainer's calls.)*

## Consequences

*(drafted on convergence — will include the `spec-0001` §1/§2 amendment
that adds the version-stamp + pin form, built by a follow-on
contract-author pass, and the check-side wiring per Open question 5.)*

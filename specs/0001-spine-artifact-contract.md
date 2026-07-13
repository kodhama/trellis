---
id: spec-0001
type: spec
status: ratified
depends_on: [invariants-v1, decision-0005, decision-0010, decision-0011, decision-0012, decision-0037, decision-0044, decision-0045, decision-0047, grove/adr-0010-versioning-is-operational]
informed_by: [research-0003]
owner: gundi
version: 1  # counter initialized 2026-07-12 with the adr-0010 de-reflection amendment — forward-only from materialization; prior states uncounted (.grove/versioning.md initialization rule)
rubric: rubric-artifact-contract
ratified: 2026-06-30
---

# Spec 0001 — The spine: artifact contract + lifecycle + conformance check

> **Ratified 2026-06-30 (A2 / D2)** — the gate is passed; the spine is being built against
> this. It is the first artifact of the spec stage (`decision-0011`), and the first user of
> that stage.
>
> *Amended in place 2026-07-12 (`grove/adr-0010`; layering: kodhama-0008 §3). WHAT: versioning
> content reduced to **shape only** — the §1 `version`/`changes` rows and the pin paragraph now
> defer semantics to the methodology (the open-field treatment `type` and `status` already get);
> §2's stamping note likewise; §3 check 8 (the version cross-check) retired, re-homed to the
> grove operating model's `corpus-reviewer` (rubric check 12 likewise). WHY: versioning is
> detection mechanics for the sync principle, not principle — its single home is the installed
> methodology companion (`.grove/versioning.md`; origin record `decision-0045`, superseded in
> part on its execution-home consequences only). SCOPE: §1 two rows + pin block, §2 one note,
> §3 checks 4/5 (citation repoints) + 8 (retired); rubric checks 4/5/12 in the same PR. POINTER:
> `grove/adr-0010`, kodhama/kodhama#35. VALUE: a contract reader gets the portable shape; no
> second home for semantics can drift. CONFIDENCE: verified (companion approved + installed).
> `status` unchanged; a significant change (testable checks retired), so the behavioral
> **`version` counter is initialized at `v1`** in this same edit, per the methodology's
> initialization rule.*
>
> *Amended in place 2026-07-13 (`decision-0047`; marking-class). WHAT: the §1 `depends_on` row
> gains a scope citation — the dependency edge denotes **genuine coupling** (a source the
> artifact's correctness is or was contingent on); **provenance** (informed construction without
> coupling) is a categorically distinct relationship, not a dependency, whose grammar is the
> methodology's (grove). Frontmatter `depends_on` gains `decision-0047`. WHY: keep the narrowing
> visible at its point of use, per this contract's per-term citation convention (`status`→`0037`,
> `version`→`adr-0010`). SCOPE: §1 `depends_on` row + frontmatter only. No §3 check changed —
> enforcement of the coupling narrowing is the grove operating model's (a future `corpus-reviewer`
> duty, grove#57), so this is **marking-class, not a testable-clause change: no `version` bump**.
> Check 7 needs no edit — under the coupling definition its frozen version-pin is a *frozen
> coupling* (a genuine dependency), coherent as written. POINTER: `decision-0047`, trellis#148/#154.
> CONFIDENCE: verified (decision approved + merged).*
>
> *Amended in place 2026-07-13 (`decision-0047` + `grove/adr-0011`; consumer-audit
> marking-class). WHAT: `research-0003` moved out of frontmatter `depends_on` into a new
> `informed_by` list — it informed this contract's construction (the artifact-type-taxonomy
> consolidation) without this spec's correctness being contingent on it; that is provenance,
> not coupling. WHY: `decision-0047` narrows `depends_on` to coupling-only; `grove/adr-0011`
> supplies the `informed_by` grammar this consumer audit applies. SCOPE: frontmatter only —
> the §1 schema (the `depends_on` row's text) is unchanged; `informed_by` is
> methodology-defined (`.grove/relations.md`), not a new schema row. No testable clause of
> this contract's own behavior changed, so this is **marking-class: no `version` bump**.
> POINTER: `decision-0047` Consequence 4, `grove/adr-0011`. CONFIDENCE: verified.*

## Purpose

Specify the **spine** — the smallest real machinery: a portable **artifact contract**, its
**lifecycle**, and an agentic **conformance check** that enforces them. It formalizes the
proto-contract we have been dogfooding across ~18 artifacts (every decision/invariant/research
file). Per `0010`, all of it is **agent instructions — no runtime, no script**: the check is a
sub-agent applying a rubric.

## Scope

**In scope (first build):** the frontmatter schema, the lifecycle states + transition rules,
the directional-flow rule, and the conformance sub-agent + its rubric, dogfooded on our own
corpus with a positive-control fixture.

**Named but build-deferred:** the **activation/wiring contract** (§5 — how the pack hooks into
a host's behavior; built in the delivery slice, `0012`). **Out of scope (later specs):**
conformance-*to-upstream* (does an implementation match its spec — a judgment agent); the
multi-surface CLI (`0012` v1); friction-export (`0009`) — though §3 notes the check's report
*is* the capture substrate.

## 1. The artifact contract (frontmatter schema)

Every non-code artifact opens with YAML frontmatter:

| Field | Req | Rule |
|---|---|---|
| `id` | ✓ | unique across the corpus; typed slug (`decision-0007`, `invariants-v1`, `spec-0001`) |
| `type` | ✓ | **open field — methodology-defined**, not a closed enum (`research-0003`); each type carries a `scope` (below) + a rubric |
| `status` | ✓ | **open field — methodology-defined**, like `type` (`decision-0037`); must belong to the methodology's declared lifecycle, which must have the §2 shape. Trellis default: `draft` → `ratified` (+ `superseded`) |
| `depends_on` | ✓ | list of `id`s and/or declared external refs; `[]` for a root. An edge denotes **genuine coupling** — a source the artifact's correctness is or was contingent on (`decision-0047`); **provenance** (a source that only *informed* construction, without coupling) is a categorically distinct relationship, not a dependency — its grammar is the methodology's (grove), not restated here |
| `owner` | ✓ | the accountable human (the `inv-intent-locus` role). The *role* is contract; the *field* is mappable — a methodology whose `owner` means something else declares which field/mechanism carries the accountable human (`decision-0037`) |
| `author` | — | optional: who wrote it (human or agent), distinct from accountability |
| `version` | — | **open field — methodology-defined**, like `type` and `status` (`grove/adr-0010`; origin record `decision-0045`): a **versioned (revise-in-place)** artifact's own version marker — present when downstreams pin it, **omitted** by append-only artifacts (which version *implicitly* via id + supersession). This contract states **shape only**; the forms, bump semantics, presence enforcement, and initialization rule live in the installed methodology companion (in a grove-managed install, `.grove/versioning.md`) — their single home, deliberately not restated here. |
| `changes` | — | on a **significant-change `decision`** only: the versioned artifact(s) it changes, each pinned (`id@version` or `<repo>/<id>@version`). **Shape at this layer:** a **forward-pointer relation of the `superseded_by` / `superseded_in_part_by` class — never a `depends_on`-class edge** (walked accordingly, §3 check 5); entries resolve like any `id`. Its reconciliation semantics are **methodology-defined** (`grove/adr-0010` — the operating model's `corpus-reviewer` owns the cross-check). |
| `date` / `ratified` / `supersedes` / `superseded_by` / `superseded_in_part_by` / `rubric` | — | optional |

**External refs:** a `depends_on` entry that is not an artifact `id` must match a declared
external-ref form. **v0 recognizes two:** `brief-§…` (an unverified section-cite into a
planning brief); and a qualified **`<repo>/<id>`** cross-repo reference (`decision-0044`) —
`<repo>` must be a member of the recognized registry (**kodhama, trellis, grove, wisp,
design-system, homebrew-tap, math-quest**) and `<id>` is the referenced artifact's own id
exactly as declared in its home corpus (e.g. `math-quest/adr-0030-espalier`,
`kodhama/kodhama-0007-one-render-many-copiers`). **Resolution depth (v0):** shape +
registry-membership only, matching `brief-§…`'s own non-verified treatment — no
fetch-and-confirm-the-referent-actually-exists mechanism. Anything else is a **dangling
reference** → fail.

**Version pins (`@version`) — shape only (`grove/adr-0010`).** A `depends_on` referent pinning a
versioned upstream may be qualified with the version it was built against: **`id@version`**
locally, **`<repo>/<id>@version`** cross-repo (extending `decision-0044`'s qualified form; `@` is
already the family delimiter — `decision-0043`'s `payload@<12-hex>`). Parse structurally: repo
names and `id`s contain no `@`, version markers no `/` or `@`, so **split on the first `/`, then
on `@`** — the same guarantee `decision-0044` established for `/`. v0's no-fetch resolution
strips `@version` and resolves the bare `id` on shape + registry/corpus membership only.
Everything past shape — which forms exist, what pinning an upstream means, pin-vs-current sync —
is **methodology-defined** (the installed companion; operationally the conformance chain's,
grove `adr-0006`).

**Types are open (`decision-0003`, `research-0003`).** Trellis does not impose a fixed type
set — a methodology brings its own (`spec`/`requirements`/`PRD`/`changes` are one function
under many names). Trellis ships a **soft seed spine** — `spec` · `plan` · `tasks` ·
`decision` · `research-note` · `feedback` · `rubric` · `invariant-set` — extensible by a
recorded decision. Each type carries a **`scope`**, so the layer split (`decision-0005`) is
enforceable at the type level:

- **`core-methodology`** — shipped to any supervised project: `decision`, `spec`, `plan`,
  `tasks`, `research-note`, `rubric`, object-level `feedback`.
- **`trellis-product`** — Trellis's own content, not per-project-instantiated: `invariant-set`;
  the contract + the type/rubric definitions.
- **`trellis-meta`** — specific to evolving Trellis: the `decision-0009` feedback-*on-Trellis*.

On install, **only `core-methodology` types ship.**

## 2. Lifecycle

**The concrete status enum is methodology-defined, like types (`decision-0037`).** The
contract requires a lifecycle **shape**, not names:

- a **working state** downstream may not consume;
- at least one **ratifiable state** — consumable, reachable only via **defined promotions**
  (the structural prerequisite `inv-ratifiable-artifacts` acts on);
- **the intent gate holds:** some ratified state is a human act — or a human-authorized,
  recorded ratchet — whatever the enum is called (B3 intent face / D2);
- **supersession is expressible**;
- the methodology **declares** its enum + promotion rules; the conformance check verifies
  `status` against that declaration. An undeclared status is a conformance failure; a
  lifecycle without this shape fails the admission gate loudly.

**Trellis's own lifecycle — the default / reference expression** (used by this repo, and
composed onto a host that brings none): `draft → ratified`; plus `ratified → superseded`
(via a successor with `supersedes`).

- **`draft`** — in progress. **Not consumable** by downstream.
- **`ratified`** — intent approved by the **human** (B3 intent face / D2). Consumable.
- **`superseded`** — replaced; must carry `superseded_by`; **never** consumed as current truth
  (B4). Decisions are append-only: supersede, never edit a ratified one.

**Supersession can be partial (`decision-0040`).** A decision can be outgrown in *part* while
its remainder stays live. The successor states what it supersedes in part; the old record
**keeps `status: ratified`** (the remainder is current) and gains
**`superseded_in_part_by: [successor…]`** — a **marking, not an edit-in-substance** (the same
class of permitted touch as the full-supersede status flip), so no reader lands on the
outgrown half without a forward link. Each entry must resolve like any `depends_on` id.

**Version stamping follows the artifact's kind (§1), and its semantics are
methodology-defined** (`grove/adr-0010`; origin record `decision-0045`) — versioned artifacts
carry the marker, append-only artifacts version implicitly via id + supersession; nothing more
is stated at this layer.

*(Worked instance of the open contract, `decision-0037`: math-quest's `draft → gated →
approved` — `gated` is rubric-self-checked and agent-consumable under a recorded ratchet,
`approved` is the human merge = ratified. Same shape, different names.)*

**Deferred — a *core* decision, not a v0 omission.** An execution-layer **`approved`** state
(B3 conformance face — implementation that passed independent conformance) is part of the
product's contract, but its model is undecided: *a third document status, or a gate-outcome
on a change rather than a status?* Evidence so far (`decision-0037`): math-quest's
conformance gate landed as a **PR gate-outcome**, not a status — while its `gated` shows a
third *document* status working for the intent layer. Because the lifecycle is
`trellis-product` scope we still do not guess Trellis's own answer here — it is decided when
the conformance-to-upstream slice is built. v0 has no execution-layer artifacts, so the
question is not yet live.

## 3. The conformance check (sub-agent + rubric — no script, `0010`)

A read-only sub-agent that takes the corpus (or one artifact + corpus) and applies the
**artifact-contract rubric**, emitting a **loud** pass/fail report (D1). It derives its
checklist from this spec, not from the producer (B3). Its checks:

1. Frontmatter present; all required fields present and well-typed.
2. `type` is declared (open field — must carry a `scope` + a rubric); `status` ∈ the
   methodology's **declared lifecycle** (here: `{draft, ratified, superseded}`;
   `decision-0037`).
3. `id` unique across the corpus.
4. Every `depends_on` resolves to an existing artifact `id`, a declared external ref, **or** a
   **retired id** in the invariant-set's Identifiers registry (mapping to its successor); no
   dangling references. A referent may carry a **`@version` pin** (§1 — shape only); resolve
   it on **shape + the bare `id`/`<repo>/<id>`'s membership only** (v0, no-fetch) — everything
   past shape is methodology-defined (`grove/adr-0010`); the pin-vs-current *sync* comparison is
   the operational chain's (grove `adr-0006`).
5. **Directional flow (load-bearing, A1/B1):** no `ratified` artifact `depends_on` a
   `draft` artifact. A decision's **`changes:`** relation (§1 — shape) is a
   **forward-pointer of the `superseded_by` class, not a `depends_on`-class dependency edge** — it
   is **not walked** as a flow edge. A spec both `depends_on`-ing its authorizing decision *and*
   named in that decision's `changes:` is a benign two-relation pair, **not a cycle** (the same way
   an append-only `decision`'s back-reference to its ratification-current upstream is exempt,
   check 7).
6. Required body sections present per type (§4).
7. **Supersede integrity:** a `superseded` artifact carries `superseded_by`; **revise-in-place**
   docs (specs, invariants, research, rubrics — B4 consolidated truth) re-point to the
   successor. A **partially superseded** artifact keeps `status: ratified` and carries
   `superseded_in_part_by`, whose entries must resolve (`decision-0040`). *Exemption (B4): an
   **append-only** `decision` may keep a dependency on the
   upstream version current at its ratification — a historical fact, not current-truth
   consumption.* A successor referencing its own predecessor (for diffing) is also exempt.
8. *(Retired 2026-07-12, `grove/adr-0010` — the version cross-check is methodology semantics,
   re-homed to the operating model: `.grove/versioning.md` §"The `changes:` relation and its
   cross-check" defines it; the operating model's `corpus-reviewer` owns it. Number retained so
   external references to "§3 check 8" resolve to this pointer rather than shifting.)*

**Honesty clause (math-quest):** *accurately listing the violations is success.* A check that
hides drift to report "pass" has failed this spec. The report is also the raw **friction
capture** substrate for `0009`.

## 4. Required body sections (per type)

- `spec` → `## Acceptance criteria`, `## Open questions`.
- `invariant-set` → the set, `## Acceptance criteria`, `## Open questions`.
- `decision` → `## Context`, `## Decision`, `## Consequences` (no acceptance criteria —
  ratification *is* a decision's acceptance).
- `research-note` → `## Open questions` (+ sources & confidence tags); **no** acceptance-
  criteria gate.
- `feedback` → exempt; an advisory rubric, never a gate (math-quest pattern).
- *Other (methodology-defined) types* declare their required sections via their rubric.

*(Surfacing our own drift is expected — e.g. decisions that predate this rule, or informal
`brief-§…` refs. The check must report them, not paper over them. See AC6.)*

## 5. Activation / wiring contract (specified here; built in the delivery slice, `0012`)

Named per `0012`, because *resources present ≠ resources used* (availability vs activation —
expressed-vs-enforced at the delivery level). The spine must define how its resources bind to
a host's behavior, even though the binding is built when delivery is:

- **Mechanism (v0, Claude plugin):** the conformance check fires via **hooks** (on the host's
  commit/PR/Write events), skills are **model-invoked**, and an optional **default agent** can
  shape the host's behavior.
- **Composition (load-bearing):** Trellis **augments, never clobbers** the host's existing
  `CLAUDE.md`/instructions — coexist, and record any change to them as a surfaced decision.
- **Activation level = the C1 dial, surfaced** (`0008`): *available + referenced* → *hooks
  fire* → *default agent*, chosen by the user, never silently maximal.
- **Acceptance (deferred to the delivery build):** installing at a chosen dial level produces
  *exactly* that degree of binding, surfaced; the host's prior instructions are preserved;
  uninstall is clean.

## Acceptance criteria

- **AC1 — no false pass / no vague fail.** On our corpus, every artifact either passes or
  yields a *specific, accurate* violation (exact field/rule/id), never a vague or absent one.
- **AC2 — positive control (B3 open question).** Given a known-bad fixture exhibiting each
  violation class (missing field; bad `status`; dangling `depends_on`; **ratified-depends-on-
  draft**; missing required section; superseded-but-consumed), the check **rejects it and
  names the exact violation**. The check is not trusted until it fails this fixture.
- **AC3 — loud, never degraded.** An unparseable/missing input halts with a visible error; no
  partial "pass" is emitted (D1).
- **AC4 — directional flow always caught.** Any `ratified`/`approved` artifact depending on a
  `draft` is always flagged (no exceptions).
- **AC5 — no runtime.** The check runs as a sub-agent + rubric on the agentic surface, with
  **no Python/Node/other runtime** (`0010`).
- **AC6 — finds real drift.** Run on the current corpus, it surfaces the *known* existing
  inconsistencies (decisions lacking the §4 sections; informal external refs), proving it
  detects, not rubber-stamps.

## Open questions

- **Spec granularity (`0011`):** does every change need a spec, or only non-trivial ones
  (minimal-first threshold)? This spec assumes the latter.
- **Two consumable states or one?** Is the `ratified`/`approved` split worth it at v0, or
  collapse to `draft → ratified`? (Keeps the B3 two-faces distinction; may be premature.)
- **External-ref mechanism — extended, not replaced (`decision-0044`):** refs multiplied (a
  2026-07-10 family-wide consistency sweep found four concrete dangling-reference instances
  across kodhama/trellis/wisp/grove) and the resolution kept the allowlist mechanism rather
  than moving to a registry *artifact* — a second recognized form (`<repo>/<id>`, §1) extends
  the existing `brief-§…` pattern instead. The **registry of recognized repo names** is inlined
  directly in §1 for v0 (duplicated here, not a pointer at a separate canonical source) — revisit
  if the registry itself starts drifting across repos, or the list keeps growing enough to
  justify externalizing it into its own artifact.
- **`core/` placement (`0005`):** the built resources (rubric, sub-agent) are Layer-A product
  → `core/`; this spec moves there in the `0005` reorg.
- **Activation/wiring (§5, `0012`):** which hooks/skills/default-agent per dial level — owed
  by the delivery slice, not this build.

## Rubric check

**First rubric-check pass applied to `spec-0001` itself.** Specs `0002`–`0004` predate the
self-check convention and carry no such section; `0005` is the first spec authored under it.
This spec's situation differs from a fresh `0005`-style authoring: it is not moving through a
lifecycle stage here, it is an already-`ratified` (family-enum equivalent: `approved`) artifact
receiving an **in-place amendment** — the same class of touch `decision-0037` and `decision-0040`
made to this same spec previously (`spec-0001` is revise-in-place current-truth,
`decision-0014`/`decision-0037` pattern). So the scope of this check is **the amendment only** —
the new external-ref form added to §1, the Open Questions update, and the frontmatter
`depends_on` addition — not a retroactive re-audit of the spec's entire pre-existing body.

Self-checked against `core/rubrics/artifact-contract.md`, per the `contract-author` agent's own
§Method item 4 (trellis has no dedicated spec-quality rubric).

| Check | Result | Note |
|---|---|---|
| 1. Frontmatter present & required fields valid | PASS | `id/type/status/depends_on/owner` shape unchanged; `depends_on` gained one well-typed entry, `decision-0044`. |
| 2. `type`/`status` declared | PASS | `type: spec`, `status: ratified` (pre-`decision-0042` spelling of the family enum's `approved`) — left untouched by this amendment; bumping/relabeling `status` is explicitly out of scope for this task, done as a separate step. |
| 3. `id` unique | PASS | `spec-0001` — no change. |
| 4. `depends_on` resolves | PASS | New entry `decision-0044` — read directly this run: `status: approved`. |
| 5. Directional flow (no `ratified`/`approved` depends on `draft`) | PASS | `decision-0044` is `approved`, not `draft` — no violation. |
| 6. Required body sections per type (spec → Acceptance criteria + Open questions) | PASS | Both present; structure untouched by this amendment. |
| 7. Supersede integrity | N/A | Not a supersession — an in-place amendment, the established precedent for this spec. |
| Honesty clause | Self-assessed honest | This section states plainly that it checks the amendment's own conformance, not a fresh full audit of `spec-0001`'s pre-existing content. |

No promotion statement follows. The `draft → gated → approved` mechanic in the `contract-author`
charter governs *new* artifacts moving through the lifecycle; this is an in-place amendment to
an already-`approved`/`ratified` artifact, matching the `decision-0037`/`decision-0040`
precedent — `status` is not touched here, per this task's explicit scope.

### Rubric check — `decision-0045` versioning-grammar amendment (2026-07-11)

A **second in-place amendment**, the same class as the `decision-0044` one above (`spec-0001` is
revise-in-place current-truth, `decision-0014`/`decision-0037` pattern — not a supersession).
**Scope of this check: this amendment only** — the new `version` and `changes` frontmatter rows
(§1), the `@version` pin grammar + `@` collision-safety + no-fetch resolution note (§1), the §2
version-stamping note, the §3 check 4/5 extensions + new check 8, and the frontmatter `depends_on`
addition of `decision-0045`. **Not** a re-audit of the spec's pre-existing body. Self-checked
against `core/rubrics/artifact-contract.md`.

| Check | Result | Note |
|---|---|---|
| 1. Frontmatter present & required fields valid | PASS | Required shape unchanged; `depends_on` gained one well-typed entry, `decision-0045`. The added `version`/`changes` rows are **optional** (`Req: —`) fields, correctly typed. |
| 2. `type`/`status` declared | PASS | `type: spec`; `status: ratified` left **untouched** — bumping/relabeling `status` is explicitly out of scope for this amendment (same posture as the `decision-0044` amendment above). |
| 3. `id` unique | PASS | `spec-0001` — no change. |
| 4. `depends_on` resolves | PASS | New entry `decision-0045` — read directly this run: `status: approved` (ratified via PR #144). |
| 5. Directional flow (no `ratified`/`approved` depends on `draft`) | PASS | `decision-0045` is `approved`, not `draft` — no violation. |
| 6. Required body sections per type (spec → Acceptance criteria + Open questions) | PASS | Both present; structure untouched by this amendment. |
| 7. Supersede integrity | N/A | An in-place amendment, not a supersession — the established precedent for this spec. |
| Honesty clause | Self-assessed honest | This entry checks only the amendment's own conformance; the rubric-sync gap (below) is stated openly, not passed over. |

**Rubric sync (`core/rubrics/artifact-contract.md`).** The rubric **duplicates** §3's checklist
(its checks 1–7 mirror §3 checks 1–7), so it needs matching edits — all **made in this pass**:
- **check 4** (`@version` no-fetch resolution) and **check 5** (`changes:` is forward-only, not a
  flow edge) — small mechanical mirrors.
- **§3 check 8 (partial version cross-check)** — wired in as rubric **check 12** under its own
  `## Check — version cross-check` heading, **not** renumbered into the base checks: the rubric's
  slots 8–11 are already `spec-0002`'s typed checks (cited by `decision-0020`/`decision-0027`), so
  appending under a labelled heading avoids a renumber while still delivering `decision-0045`
  Consequences item 3 (the `corpus-reviewer` *gains* the check in the operative gate, not only in
  spec prose). The rubric's numbering is already not 1:1 with §3 past check 7 (its 8–11 have no §3
  counterpart), so the §3-check-8 ↔ rubric-check-12 mapping is consistent with that.

An earlier draft of this amendment deferred the rubric wiring of check 8; an independent
adversary pass (`spec-adversary`, 2026-07-11) noted that (a) the check dropped `decision-0045`'s
explicit **behavioral-artifact** scoping — its "behind" test is undefined for the unordered
content-hash form the same amendment admits — and (b) a check living only in spec prose, not the
operative rubric, does not actually deliver Consequences item 3. Both are fixed above: check 8 is
now **scoped to the behavioral / counter-versioned form** and **wired into the rubric**.

**Status unchanged.** As with the `decision-0044` amendment, `status` stays `ratified`; no
promotion statement follows — the `draft → gated → approved` mechanic governs *new* artifacts, not
an in-place amendment to an already-ratified one.

### Rubric check — `grove/adr-0010` de-reflection amendment (2026-07-12)

Scope: this amendment only (§1 two rows + pin paragraph, §2 stamping note, §3 checks 4/5
repoints + check 8 retired; `version: 1` initialized) — not a re-audit of the pre-existing body.
Self-checked by the amending agent; an independent conformance review of the same diff runs on
the amending PR (its verdict is recorded there).

| Check | Verdict | Evidence |
|---|---|---|
| 1. Frontmatter valid | PASS | `version: 1` added (optional field, initialization rule applied at first significant change); `depends_on` gained `grove/adr-0010-versioning-is-operational`, well-typed per `decision-0044`'s qualified form. |
| 4. `depends_on` resolves | PASS | `grove/adr-0010` read directly this run: `status: approved` (grove#50). |
| 5. Directional flow | PASS | The new upstream is `approved`, not `draft`. |
| 7. Supersede integrity | PASS | Nothing here superseded; `decision-0045`'s partial marks land in the same PR with resolving entries. |
| 8. Version cross-check | N/A | Retired by this amendment (pointer retained); the re-homed check is the operating model's. `adr-0010` carries no `changes:` field — recorded honestly: the bump is decision-backed but not `changes:`-declared, the soft direction the semantics permit. |

Status stays `ratified` (in-place amendment, the repo's precedent). The amendment note at the
head is the delta record; POINTER `grove/adr-0010` / kodhama/kodhama#35.

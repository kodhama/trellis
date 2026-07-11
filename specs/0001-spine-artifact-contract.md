---
id: spec-0001
type: spec
status: ratified
depends_on: [invariants-v1, decision-0005, decision-0010, decision-0011, decision-0012, decision-0037, decision-0044, decision-0045, research-0003]
owner: gundi
rubric: rubric-artifact-contract
ratified: 2026-06-30
---

# Spec 0001 — The spine: artifact contract + lifecycle + conformance check

> **Ratified 2026-06-30 (A2 / D2)** — the gate is passed; the spine is being built against
> this. It is the first artifact of the spec stage (`decision-0011`), and the first user of
> that stage.

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
| `depends_on` | ✓ | list of `id`s and/or declared external refs; `[]` for a root |
| `owner` | ✓ | the accountable human (the `inv-intent-locus` role). The *role* is contract; the *field* is mappable — a methodology whose `owner` means something else declares which field/mechanism carries the accountable human (`decision-0037`) |
| `author` | — | optional: who wrote it (human or agent), distinct from accountability |
| `version` | — | the **versioned (revise-in-place) artifact's own version marker**, at its own granularity — *not* a foreign decision-id (`decision-0045` item 3). **Required** on a versioned artifact that downstreams pin; **omitted** by append-only artifacts (decisions), which version *implicitly* via id + supersession (item 2). **Form fits the kind** (item 5 — a spectrum, not a two-way function): a **behavioral spec** → a plain monotonic counter (`v1`, `v2`, …), agent-generated, a review-bounded significance *ordering* (**not** semver; a testable-clause — scenario/invariant — change bumps it, a prose-only edit does not — item 6); a **vendored / byte-identical bundle** (trellis's `payload`) → a content-hash (`payload@<12-hex>`, `decision-0043` — unchanged); a **human-cut release** (design-system tokens) → a git tag (`vX.Y.Z` — unchanged). The behavioral counter is a **claim bounded by review, not a "can't-lie" derivation** (item 6). |
| `changes` | — | on a **significant-change `decision`** only: the versioned artifact(s) it changes, each pinned to the version it set (`id@version` or `<repo>/<id>@version`). A **forward-pointer relation of the `superseded_by` / `superseded_in_part_by` class — never a `depends_on`-class edge** (`decision-0045` item 7); entries resolve like any `id`. Feeds the §3 partial version cross-check. |
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

**Version pins (`@version`, `decision-0045`).** A `depends_on` entry pinning a **versioned**
upstream (one that carries a `version` marker, §1/§2) may qualify the referent with the version it
was built against: **`id@version`** locally (e.g. `spec-mastery-engine@v3`), or
**`<repo>/<id>@version`** cross-repo (e.g. `math-quest/spec-slice-01-first-loop@v3`) — extending
`decision-0044`'s qualified `<repo>/<id>` form. **`@` is already the family's version delimiter:**
`decision-0043`'s `payload@<12-hex>` content-hash stamp already uses it, so this amendment
*generalizes* that existing delimiter to all versioned pins — it does not invent one. The
`<version>` is whatever form fits the upstream's kind (a counter `vN`, a git tag `vX.Y.Z`, a hex
hash — the §1 `version` row).

**`@` collision-safety (checked the way `decision-0044` checked `/` and `:`).**
`<repo>/<id>@<version>` parses unambiguously: repo names (the registry — **kodhama, trellis,
grove, wisp, design-system, homebrew-tap, math-quest**) and artifact `id`s (kebab slugs) contain
no `@`; version markers (`vN`, `vX.Y.Z`, a hex hash) contain no `/` or `@`. So **split on the
first `/`, then split on `@`** recovers `<repo>`, `<id>`, and `<version>` with no ambiguity — the
same *structural* (not heuristic) guarantee `decision-0044` established for the `/` delimiter.

**Resolution depth (v0, no-fetch — `decision-0044`).** A `@version` pin is checked on **shape +
the bare `id`/`<repo>/<id>`'s registry/corpus membership only**; v0 does **not** fetch the upstream
to compare the pinned version against its current one. That **pin-vs-current *sync* comparison is
the operational check owned by grove#34 / grove `adr-0006`**, not this spec's §3 conformance check
— the same non-verified treatment `brief-§…` and a bare `<repo>/<id>` already get.

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

**Version stamping is a property of *kind*, not lifecycle state (`decision-0045`).** A
**versioned / revise-in-place** artifact (a spec — grove `adr-0004`; a `decision-0014`
invariant-set) carries an **explicit `version` stamp** (§1): its `id` alone does not identify
*which* state a downstream built against, so the stamp is that pin currency. An **append-only**
artifact (a `decision`) needs no stamp — it versions *implicitly*: the `id` already pins a unique
immutable state and supersession is its history (`decision-0045` item 2).

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
   dangling references. A referent may carry a **`@version` pin** (`decision-0045`, §1); resolve
   it on **shape + the bare `id`/`<repo>/<id>`'s membership only** (v0, no-fetch) — the
   pinned-version-vs-upstream's-current *sync* comparison is **not** this check's; it is grove#34 /
   grove `adr-0006`'s operational check.
5. **Directional flow (load-bearing, A1/B1):** no `ratified` artifact `depends_on` a
   `draft` artifact. A decision's **`changes:`** relation (`decision-0045` item 7) is a
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
8. **Version cross-check (partial, `decision-0045` Consequences item 3).** Where a
   significant-change `decision` carries `changes: [X@vN]`, reconcile it against `X`'s `version`
   **record** — **not** a naive `declared == current` equality. A `decision` is append-only, so
   its declared `@vN` is a *historical* fact that legitimately sits **behind** `X`'s current
   `version` once `X` bumps again (a decision that set `X@v3` is not wrong because `X` later
   reached `v4`). The **sound finding is a declared change that never landed** — `changes:
   [X@vN]` where `X`'s current `version` is *behind* `vN` (`X` never reached the version the
   decision claims to have set). The reverse direction — a bump in `X` with **no** accounting
   `changes:` decision — is **softer, never a hard FAIL**: `decision-0045`'s own open question
   leaves *"must every significant change flow from a decision?"* unsettled, so an unaccounted
   bump is at most a prompt to look, not a violation. A **bounded, intra-repo
   frontmatter-vs-record audit**, owned by the conformance check / `corpus-reviewer` — **distinct**
   from the consumer-vs-upstream *sync* check (check 4), which is grove#34 / grove `adr-0006`'s.

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
(its checks 1–7 mirror §3 checks 1–7), so it needs matching edits. The **check 4** (`@version`
no-fetch resolution) and **check 5** (`changes:` is forward-only, not a flow edge) additions are
small, mechanical mirrors — **made in the same pass**. The **new §3 check 8 (partial version
cross-check)** is *not* mirrored into the rubric here: it is `decision-0045` Consequences item 3
(the `corpus-reviewer` *gains* it), a distinct named deliverable, and it collides with the
rubric's existing numbering (checks 8–11 are already the `spec-0002` typed-artifact checks —
inserting it would force a renumber touching `decision-0020`/`decision-0027` citations).
**Flagged as a substantive follow-on**, not guessed here.

**Status unchanged.** As with the `decision-0044` amendment, `status` stays `ratified`; no
promotion statement follows — the `draft → gated → approved` mechanic governs *new* artifacts, not
an in-place amendment to an already-ratified one.

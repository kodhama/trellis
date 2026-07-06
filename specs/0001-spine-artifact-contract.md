---
id: spec-0001
type: spec
status: ratified
depends_on: [invariants-v1, decision-0005, decision-0010, decision-0011, decision-0012, decision-0037, research-0003]
owner: gundi
rubric: spec-quality
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
| `date` / `ratified` / `supersedes` / `superseded_by` / `rubric` | — | optional |

**External refs:** a `depends_on` entry that is not an artifact `id` must match a declared
external-ref prefix (v0 allowlist: `brief-§…`). Anything else is a **dangling reference** →
fail.

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
   dangling references.
5. **Directional flow (load-bearing, A1/B1):** no `ratified` artifact `depends_on` a
   `draft` artifact.
6. Required body sections present per type (§4).
7. **Supersede integrity:** a `superseded` artifact carries `superseded_by`; **revise-in-place**
   docs (specs, invariants, research, rubrics — B4 consolidated truth) re-point to the
   successor. *Exemption (B4): an **append-only** `decision` may keep a dependency on the
   upstream version current at its ratification — a historical fact, not current-truth
   consumption.* A successor referencing its own predecessor (for diffing) is also exempt.

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
- **External-ref mechanism:** an allowlist prefix (v0) vs a registry artifact — revisit when
  refs multiply.
- **`core/` placement (`0005`):** the built resources (rubric, sub-agent) are Layer-A product
  → `core/`; this spec moves there in the `0005` reorg.
- **Activation/wiring (§5, `0012`):** which hooks/skills/default-agent per dial level — owed
  by the delivery slice, not this build.

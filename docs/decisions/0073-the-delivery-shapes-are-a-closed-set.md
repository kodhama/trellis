---
id: decision-0073
type: decision
status: approved  # maintainer intent act 2026-08-03, in session, at the governed run's batched intent gate: "Approve both, in order (0072 then 0073)" — recorded AFTER decision-0072's flip, satisfying this record's own sequencing consequence. An in-PR flip recording the act; author (agent) != approver (maintainer). Decision-adversary SOUND at revision 4 (record: decision-adversary.d0073.toml). Originally: drafted by agent; revision 4 after three decision-adversary rounds (6+6+3 findings, all folded; the revision record locates every one); awaiting fresh adversary convergence, then the maintainer intent act at this run's intent gate
depends_on: [decision-0068, decision-0070, decision-0072]
superseded_in_part_by: [decision-0077]  # 2026-08-28 — D3's second bullet only, and only its "one ignored prompt re-governs at 14/14" half, which restated decision-0070 D4 faithfully and inherited its error. The surviving half — deleting the recorded decline re-arms the announcement — is verified true. Everything else in this record stands: the closed set, the transaction ordering, the morph preflight, the predicate discipline
informed_by: [spec-0006]
owner: agent
date: 2026-08-03
provenance: maintainer intent act 2026-08-03, in session, opening a governed run — "Fix the P1s under governance yes, and don't forget the reviewers have to converge before we take this to the PR reviewers." The two P1s come from the 2026-08-03 state-coverage audit of the decision-0072 retirement branch, which found twenty instances of one defect class across seven review rounds and traced the class to this decision's subject.
---

# 0073 — the delivery shapes are a closed set, enumerated once

## Context

**Twenty review findings on trellis#227 were one defect**: a remedy, recipe or
guard that was wrong about the state the reader was actually in. The audit that
followed found the root cause: **at least five components each hold their own
private, partial enumeration of how Trellis delivery can be laid out in a
project**, and every one of them misses states the others handle. This
decision's own first draft then repeated the defect — its "closed set of five"
omitted the product's primary delivery mode (see the revision record below) —
which is the strongest available evidence that the enumeration must be
verified, not recalled.

Measured against `main`+#227, 2026-08-03 (fixtures and a live-session probe):

- **`plugins/trellis/hooks/staleness.sh`** has **no code for the inline
  managed-block shape** — it never reads an instructions file (verified by
  full read). Three consequences, each confirmed by code-walk and fixture:
  double delivery (an inline project falls through to path B, which injects
  the full payload — 7,646 bytes over 29 live rule lines already in
  `CLAUDE.md`); a silent `governed = false` (the disregard branch's condition
  tests three paths, none of them an instructions file); and a blind
  coexistence check (inline-plus-rendered draws the quiet stand-down, which
  asserts the rules are not in context twice while they are). **These hold for
  real installs**: a headless control session in a properly-installed project
  received the full hook-delivered readout, so the hook genuinely fires and
  its blind spots genuinely reach users.
- **`plugins/trellis/skills/remove/SKILL.md`** counts three installed shapes
  in its no-op predicate and never mentions the vendored bundle at
  `.claude/skills/trellis/` — the artifact `decision-0070` D1/D3 name the
  project-scope **adoption act**. Its reviews under this run added two more
  omissions: step 4 deletes `.trellis/` wholesale, which destroys **the
  recorded `governed = false` decline** (0070 D4's durable opt-out) with no
  surfacing category, and destroys **`.trellis/rollback`**, the M2 morph's
  rollback pointer, before the procedure ever reaches the section that
  depends on it.
- **A live measurement narrowed the bundle finding, and corrected this run's
  own audit.** The audit's fixture invoked the hook manually with
  `CLAUDE_PLUGIN_ROOT` pointed at the bundle — constructing the premise it
  claimed to test. Measured properly (headless sessions, trusted fixture,
  with a positive control): **a curl-vendored bundle's hook does not fire**;
  the bundle's **hook** is delivery-inert on the measured surface. Stated
  precisely: 0068's open question 5 asks whether the vendored **skills** load
  at all, and skills discovery and hook registration are different host
  mechanisms — this measurement answers only the hook-firing sub-question,
  negatively, headless; the skills half of OQ5 remains open in full. So a bundle left
  behind by `/trellis:remove` does **not** silently re-govern the project;
  what remains true is that the skill leaves the adoption-act artifact and
  the product's skills on disk, unreported, in a project it just declared
  clean — and that a properly-installed project-scope plugin (which does
  fire) is outside the skill's world model entirely.
- **`install.sh`** recognises four static shapes including the inline block
  (column-0 `<!-- trellis:begin` probe over `CLAUDE.md`/`AGENTS.md`); its own
  comment records that the two-file scope is a deliberate narrowing against
  the remove skill's five-file set — the per-component relevance D1 now makes
  explicit.

The class survived seven review rounds because each fix repaired the instance
shown and left the enumeration private and partial. **The fix for the class is
that the enumeration stops being private**: one closed, named set, and every
component provably covering it.

## Decision state

**Decided** (maintainer direction 2026-08-03; drafted by agent; revised on
adversary round 1):

- D1–D5 below.
- Formerly Open 1 (bundle deletion): **decided in D3** — remove deletes the
  bundle behind the existing confirmation gate, with its transaction position
  pinned. The measurement lowered the stakes (the bundle is delivery-inert on
  the measured surface) and the adversary's ordering finding raised the
  interruption cost of leaving it last; both point the same way.
- Formerly Open 2 (spec-0006 pointer): **decided — no pointer.** The
  adversary verified independently: spec-0006 scopes to this repository's own
  entrypoints, and D1 adds and retires none.

**Open** (0).

**Parked** (3, deliberately not this run):

- The remaining P2/P3 audit findings — filed as issues per `kodhama-0027`.
- The guard's unclaimed-type over-owing on SKILL.md files (grove#200) — its
  fail-closed behaviour left intact here; the typing question (a deliberate
  `skill` lane vs the full set forever) is grove's to decide.
- Whether interactive (non-headless) sessions fire a vendored bundle's
  **hook** — a question this run's measurement raised and did not answer. It
  is **not** 0068's open question 5, which asks whether the vendored
  **skills** load and remains open in full; the hook sub-question is measured
  negative headless and unmeasured interactive, and D3's report wording is
  chosen to be true under any answer.

## Decision

**1. The delivery states are a closed set, and this record is its single
normative home.** A project consuming Trellis is in exactly one of these
states, or a conflicting combination that every reader must name rather than
absorb:

| # | state | on-disk signature |
|---|---|---|
| S0 | plugin-native / none | **no** static artifact — delivery, when any, comes from an installed plugin's hook; with `.trellis/rules.toml` present this is the product's primary mode ("config only"), and with nothing present it is the unadopted state |
| S1 | rendered file (curl path) | `.claude/rules/trellis.md` |
| S2 | internal overlay | `.trellis/internal/` + managed block in an instructions file |
| S3 | legacy flat overlay | `.trellis/trellis.md` (+ `.trellis/version`) + managed block |
| S4 | inline managed block | `<!-- trellis:begin` at column 0 of **a documented instruction file** (the remove skill's five-file set: `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.github/copilot-instructions.md`, `.clinerules`), rules body embedded **or** a dangling import whose overlay was deleted |
| S5 | project-scope plugin bundle | `.claude/skills/trellis/` — the adoption-act artifact (`decision-0070` D3), **except when the project root is `$HOME`** (0070 D6 carve-out); measured delivery-inert on the headless surface, adoption semantics unchanged |
| S6 | M2 morph | `.trellis/rollback` and/or the `trellis-pre-morph` tag (`spec-0004` §2's first-class markers) — delivery is the project's **own rewritten instruction files**; no block, overlay, rendered file or bundle need exist. **The tag can outlive the state** (a rollback or completed removal leaves it; `git tag -d trellis-pre-morph` clears it) — see AC1's surfacing rule |

Two closure clauses, both load-bearing: **the none-state is a member** — a
test that enumerates S1–S5 and forgets S0 repeats draft 1's defect; and
**per-component relevance is explicit** — a component may deliberately scope
to a subset (install.sh's two-file probe vs the skill's five files) only by
saying so where it does it, with a pointer here. Adding a state is a decision
superseding this one in part — never a code change alone. A genuinely novel
layout no one has enumerated is invisible to any suite; what D4 guarantees is
that every *named* state stays covered.

**2. `staleness.sh` handles every state, and its inline-shape refusal is
honest about what the probe can know.** The hook gains the column-0
`<!-- trellis:begin` probe **over `CLAUDE.md` and `AGENTS.md` only — a
deliberate S4 subset, stated in the probe's own comment with a pointer to
this record, which is where D1's relevance clause requires it (the build owes
that comment; this record is the pointer's target, not a substitute for
it)**: those are the files the Claude host loads, and refusing delivery
over a block in `GEMINI.md` or `.clinerules` — files this host never reads —
would ungovern a Claude session for content that was never in it, the exact
wrong-about-the-reader's-state class this record exists to end. The probe
feeds the hook's three decision sites: the coexistence check, the
`governed = false` disregard message, and a refusal before path B.
The probe cannot distinguish an embedded block (rules live in context) from a
dangling import whose overlay was deleted (no rules in context, silently) —
so the refusal message is written for both: it names the block, says which of
the two states the project may be in and how to tell, and never asserts
"loaded twice" as fact when the block may be a dangling import. Both fixtures
land in AC2.

**3. `/trellis:remove` inventories, reports, and removes every state.**
Specifically, beyond its current three:

- **The bundle (S5)** is inventoried, reported, and deleted behind the
  skill's existing explicit-confirmation gate. Transaction position pinned:
  the bundle is handled **with or before** the `.trellis/` deletion, never
  after it, so no interruption window leaves the adoption-act artifact as the
  sole survivor. The report's retained-wording states what is true under
  either OQ5 answer: the bundle is the adoption-act artifact and the
  product's skills; delivery from it was measured inert headless.
- **The decline artifact**: a `.trellis/rules.toml` carrying
  `governed = false` is surfaced by name before any deletion, with the
  consequence stated — on a machine with a user-scope plugin, deleting the
  recorded decline re-arms the adoption announcement, and one ignored prompt
  re-governs at 14/14. Deleting it remains legitimate (removal is removal);
  doing it unnamed is not.
- **Morph detection moves to preflight**, before any write — step 4's
  `.trellis/` deletion destroys `.trellis/rollback`, and the current
  procedure reaches the morph section only after the destruction. The no-op
  predicate distinguishes **S0-unadopted** (the one true "already absent")
  from every removable state: S1–S6 and S0-config-only, whose bare
  `rules.toml` is removable state — today's predicate would call a
  fully-governed plugin-native project "already absent".

**4. The closed set is executable, with the assertion mode matched to the
component class.** For the two shell components, behavioural fixtures: every
state in D1 has a fixture, and each component either handles it or refuses it
by name. For the skill — model-interpreted prose — the honest assertion is
textual: a `cli/` test asserts SKILL.md's inventory, no-op predicate and
report enumerate every D1 state by name (the existing docs-guard pattern).
Prose tests pin what the text claims; only the shell fixtures pin behaviour —
D4 states this split rather than implying uniform strength.

**5. Acceptance criteria for this run's build stage (regression-test-first):**

- **AC1**: the bundle-only fixture — every SKILL.md deletion performed, the
  bundle surviving — draws a report that names the bundle as retained (or a
  removal that deletes it on confirmation); the no-op predicate reports
  "already absent" **only in S0-unadopted** — never while any removable D1
  artifact (S1–S6, or S0's config file) is present. The `governed = false`
  and morph-marker fixtures draw their named surfacing before any write —
  and the morph surfacing presents the marker **as a marker**, never the
  morph as asserted fact: it may be stale after a rollback, so the message
  says how to tell (does the rewritten content still stand?) and how to
  clear a stale tag.
- **AC2**: the embedded-inline fixture draws the loud refusal from path B,
  never double delivery; the dangling-import fixture draws the same refusal
  with its either-state wording, never a false "loaded twice" claim;
  `governed = false` beside an inline block emits the disregard message
  naming S4; inline-plus-rendered emits the coexistence alarm naming both.
- **AC3**: every new guard goes red against the pre-fix code (mutation), and
  every fixture provably contains the state it names — including the
  count discipline the first draft slipped on (the delivered payload is 12
  `inv-` rules plus 2 `floor-` rules; 14 slugs total, stated once).

## Consequences

- `staleness.sh` and `skills/remove/SKILL.md` change under this record's
  build stage. `install.sh` gets exactly one touch: its existing two-file
  narrowing comment gains the pointer to this record that D1's relevance
  clause requires (it currently cites 0068 D7 only); its behaviour does not
  change.
- The build must also reconcile `README.md`'s manual inline recipe, which
  currently instructs copying `.trellis/internal/` *and* appending the inline
  block — a layout D1 classifies as S2-plus-S4 conflicting. Producing clean
  S4 requires more than a recipe edit: the shipped inline payload itself
  points readers at `.trellis/internal/invariants.md`
  (`block-inline-tail.md`), so the generated block text changes in
  `cli/apply.go` with the payload re-rendered — the `decision-0072`
  generator-boundary rule — or the recipe keeps shipping the conflict.
- The OQ5 headless measurement is recorded here. It answers only the
  hook-firing sub-question (negative, headless surface); 0068's question as
  written — whether the vendored **skills** load — remains open in full, and
  the interactive hook surface stays unmeasured. 0068 is not edited.
- **Sequencing at the gate (adversary F4):** `decision-0072` is `gated` and
  rides #227; this record depends on it. The intent act on 0073 must land
  together with or after 0072's — the gate owner rules the order as part of
  the act, and this record is not consumable on `main` before 0072 is.
- `decision-0072`'s "remains the clean exit" claim becomes true; today it is
  aspirational.

## Revision record

**Revision 3 → revision 4, on decision-adversary round 3 NEEDS-REVISION
(2026-08-03), one blocking item and two tightenings, all folded:** A the
Parked section still carried the hook-for-skills substitution the N4 fix
corrected two sections above it — the parked question is now stated as the
interactive *hook* surface, explicitly not OQ5; B the S6 tag can outlive the
morph state — the row says so and AC1's surfacing rule presents the marker as
a marker, never the morph as fact; C D2 no longer reads as if the
decision-side record satisfies D1's say-it-where-you-do-it clause — the build
owes the probe comment, with this record as the pointer's target.

**Revision 2 → revision 3, on decision-adversary round 2 NEEDS-REVISION
(2026-08-03), all six items folded:** N1 the morph state D3 named three times
was absent from D1's "closed" table — now S6, signed by spec-0004 §2's
markers; N2 the hook probe's file scope was an absorbed enumeration — now
stated in D2 as a deliberate two-file subset with its reason; N3 S0's
membership made AC1 literally unsatisfiable — the quantifier is now "any
removable D1 artifact", and D3's predicate splits S0-unadopted from
S0-config-only; N4 the OQ5 narrowing had substituted the hook for the skills —
both sentences now scope the measurement to the hook-firing sub-question;
N5 "clean S4" now names the generated-payload change it requires; N6 the
install.sh pointer is now an explicit single touch. The adversary's Part-4
objection (Open(0) hiding two silently-decided questions) is met by deciding
both in the text above rather than reopening them.

**Draft 1 → revision 2, on decision-adversary NEEDS-REVISION (2026-08-03),
all six findings folded:** F1 the set omitted the none/config-only state —
the product's primary delivery mode — and now carries S0 plus the closure
clause; F2 S4's two-file signature contradicted the documented five-file
manual path — widened, with per-component relevance made explicit; F3 the
probe cannot distinguish embedded from dangling-import — D2's refusal
rewritten for both states, fixture added; F4 the gated-0072 dependency is now
a named sequencing consequence at the gate; F5 D4 overclaimed executability
for prose — assertion modes split by component class; F6 bundle deletion
ordering pinned. Its three minor notes (the `$HOME` carve-out, the 12+2
count discipline, the README inline recipe) are in D1/AC3/Consequences. The
SKILL.md decision-adversary's decline-artifact finding (F3 there) is D3's
second bullet. The OQ5 measurement corrected this run's own audit fixture,
and the correction is stated in Context rather than absorbed.

## Self-check (revision 4)

Sections present; every load-bearing code claim either verified by the round-1
adversary against source or re-measured this session (the OQ5 probe, with a
positive control and the constructed-premise error owned in Context);
`depends_on` unchanged and the gated-0072 bind now an explicit gate-sequencing
consequence rather than silent settled-ground; Open count is zero with both
former items decided and recorded; parked items carry their routing.
Round-2 items were folded as edits to the exact clauses the adversary named,
with the revision record locating each; Open stays 0 because both
silently-decided enumerations are now explicitly decided in the text, which is
the remedy the adversary's Part 4 asked for. Promoted to `gated` on this
self-check; `approved` remains the maintainer's intent act, after the fresh
adversary pass on this revision.

---

> **Superseded in part (2026-08-28, append-only pointer).**
> `decision-0077` supersedes **half of D3's second bullet**. *"One ignored prompt
> re-governs at 14/14"* was a faithful restatement of `decision-0070` D4, which the
> hook never implemented; an unanswered announcement leaves the project **ungoverned**
> and recurs next session. The other half — *"deleting the recorded decline re-arms
> the adoption announcement"* — is **verified true** and stands: the branch is a bare
> file-existence test with no persisted already-announced state.
>
> The surfacing obligation D3 creates is unchanged; only the consequence it states
> changes. Deleting a decline returns the project to the **unadopted** state — the
> announcement returns, and the project is governed only if someone then accepts.
> `skills/remove/SKILL.md` carries the corrected wording, as the live surface.
>
> **Everything else in this record stands**, including the closed set itself. The
> global sweep on this record's own run flagged the clause as L3 and routed it to the
> maintainer; it is closed here rather than left riding.

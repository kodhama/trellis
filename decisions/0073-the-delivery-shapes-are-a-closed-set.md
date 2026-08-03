---
id: decision-0073
type: decision
status: gated  # drafted by agent; revision 2 after decision-adversary NEEDS-REVISION (six findings, all folded); self-check re-run and recorded below; awaiting fresh adversary convergence, then the maintainer intent act at this run's intent gate
depends_on: [decision-0068, decision-0070, decision-0072]
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
  the bundle is delivery-inert on the measured surface. This partially
  answers `decision-0068` open question 5 in the negative. So a bundle left
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
- Whether interactive (non-headless) sessions load a vendored bundle's hook —
  the measurement covered the headless surface with a positive control;
  0068's open question 5 stays open for the interactive surface, and D3's
  report wording is chosen to be true under either answer.

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
`<!-- trellis:begin` probe, feeding its three decision sites: the coexistence
check, the `governed = false` disregard message, and a refusal before path B.
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
  predicate counts every state above, morph state included.

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
  removal that deletes it on confirmation); the no-op predicate no longer
  reports "already absent" while any D1 state is present. The
  `governed = false` and morph-marker fixtures draw their named surfacing
  before any write.
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
  build stage; `install.sh` is touched only if its enumeration drifts from
  D1's per-component relevance clause.
- The build must also reconcile `README.md`'s manual inline recipe, which
  currently instructs copying `.trellis/internal/` *and* appending the inline
  block — a layout D1 classifies as S2-plus-S4 conflicting; the recipe should
  produce clean S4.
- The OQ5 headless measurement is recorded here; `decision-0068`'s open
  question 5 narrows to the interactive surface. 0068 is not edited.
- **Sequencing at the gate (adversary F4):** `decision-0072` is `gated` and
  rides #227; this record depends on it. The intent act on 0073 must land
  together with or after 0072's — the gate owner rules the order as part of
  the act, and this record is not consumable on `main` before 0072 is.
- `decision-0072`'s "remains the clean exit" claim becomes true; today it is
  aspirational.

## Revision record

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

## Self-check (revision 2)

Sections present; every load-bearing code claim either verified by the round-1
adversary against source or re-measured this session (the OQ5 probe, with a
positive control and the constructed-premise error owned in Context);
`depends_on` unchanged and the gated-0072 bind now an explicit gate-sequencing
consequence rather than silent settled-ground; Open count is zero with both
former items decided and recorded; parked items carry their routing. Promoted
to `gated` on this self-check; `approved` remains the maintainer's intent act,
after the fresh adversary pass on this revision.

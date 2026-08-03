---
id: decision-0073
type: decision
status: gated  # drafted by agent; self-check recorded in the lifecycle note below; awaiting decision-adversary convergence, then the maintainer intent act at this run's intent gate
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
project**, and every one of them misses shapes the others handle.

Measured against `main`+#227, 2026-08-03 (fixtures, not reasoning):

- **`install.sh`** recognises four static shapes (rendered file, internal
  overlay, legacy flat overlay, inline managed block) — the inline probe was
  deleted on the #227 branch on the claim the other tests covered it, and
  restored when that proved false.
- **`plugins/trellis/hooks/staleness.sh`** recognises three. It has **no code
  for the inline managed block at all** — it never reads an instructions file.
  Three consequences, each reproduced:
  - **Double delivery.** An inline-block project falls through to path B and
    the hook injects the full payload on top of the block already in context
    (fixture: 7,646 injected bytes over 29 live rule lines in `CLAUDE.md`).
  - **The opt-out goes silent while rules are live.** `governed = false`
    beside an inline block emits nothing — the exact defect the disregard
    branch was built to prevent, on a third transport it does not check.
  - **The coexistence check is blind to it.** Inline block plus a rendered
    file emits the quiet stand-down — asserting the rules are not in context
    twice while they are.
- **`plugins/trellis/skills/remove/SKILL.md`** counts three installed shapes
  in its no-op predicate. There are more, and the missing one **governs**:
  the project-scope vendored bundle under `.claude/skills/trellis/` is the
  adoption act (`decision-0070` D3), the hook delivers from it every session,
  and the skill neither inventories, removes, nor reports it. Fixture: perform
  every deletion SKILL.md names, run the hook — **12 `inv-` rules injected,
  14/14 with floors, in a project `/trellis:remove` just called clean**.
- **The docs** taught a two-branch posture recipe until review round six found
  the third config-file shape (`governed = false`), and a guard per spelling
  until rounds four, five and seven found `drop` vs `delete`, bare `setup` vs
  `/trellis:setup`, and a hand-written surface list.

The class survived seven review rounds because each fix repaired the instance
shown and left the enumeration private and partial. **The fix for the class is
that the enumeration stops being private**: one closed, named set, and every
component provably covering it.

## Decision state

**Decided** (maintainer direction 2026-08-03; drafted by agent):

- D1–D5 below.

**Open** (2):

1. **Should `/trellis:remove` delete the project-scope bundle itself, or
   report it with the removal command?** The skill executes *from* that bundle
   at project scope (`.claude/skills/trellis/skills/remove/`), so it would be
   deleting its own home mid-run — operationally safe once loaded into
   context, but worth deciding deliberately rather than by default. Draft
   answer in D3: delete it, behind the same explicit confirmation every
   destructive step already carries, because a "clean exit" that leaves the
   adoption act in place is not an exit (the fixture above is the proof).
2. **Does `spec-0006` want a forward pointer?** It defines the shared
   entrypoint contract the inline block instantiates; this decision adds no
   entrypoint and retires none, so the draft answer is no pointer — D1 names
   the shapes, it does not redefine entrypoints. The contract-author should
   confirm on contact.

**Parked** (2, deliberately not this run):

- The remaining P2/P3 audit findings (state-dependent installer messages, doc
  contradictions, guard-coverage gaps outside these two subjects) — filed as
  issues per `kodhama-0027`, not folded into this run's scope.
- The guard's unclaimed-type over-owing on SKILL.md files (grove#200) — a
  grove defect, already filed, its fail-closed behaviour left intact here.

## Decision

**1. The delivery shapes are a closed set of five, and this record is its
single normative home.** A project consuming Trellis is in exactly one or a
conflicting combination of:

| # | shape | on-disk signature |
|---|---|---|
| S1 | rendered file (curl path) | `.claude/rules/trellis.md` |
| S2 | internal overlay | `.trellis/internal/` + managed block in the instructions file |
| S3 | legacy flat overlay | `.trellis/trellis.md` (+ `.trellis/version`) + managed block |
| S4 | inline managed block | `<!-- trellis:begin` at column 0 of `CLAUDE.md`/`AGENTS.md`, rules body embedded, no overlay directory |
| S5 | project-scope plugin bundle | `.claude/skills/trellis/` — **also the adoption act** (`decision-0070` D3) |

Adding a shape is a decision superseding this one in part — never a code
change alone.

**2. Every component that branches on delivery state handles all five
shapes.** Concretely for the P1 subjects: `staleness.sh` gains the same
column-0 `<!-- trellis:begin` probe `install.sh` already carries, feeding all
three of its decision sites — the coexistence check, the `governed = false`
disregard message, and the pre-path-B refusal. A shape a component cannot
handle is handled loudly (a named refusal), never silently.

**3. `/trellis:remove` inventories, reports, and removes all five shapes.**
Its inventory step includes `.claude/skills/trellis/`; its no-op predicate
counts five shapes, not three; and its report states plainly that a bundle
left in place means the project is still governed at the shipped defaults.
Bundle deletion rides the skill's existing explicit-confirmation gate, and the
skill states the self-deletion property out loud (Open 1).

**4. The closed set is executable, not prose.** A test in `cli/` enumerates
the five signatures and asserts, per component, either handling or a named
refusal — so a sixth shape, or a component dropping one, fails the suite
rather than waiting for review round eight. The twenty-findings history is the
argument that prose enumeration does not hold.

**5. Acceptance criteria for this run's build stage (regression-test-first,
per the bug pipeline):**

- **AC1**: the remove-then-hook fixture — perform every step SKILL.md names,
  then run the hook — injects **zero** rules and the skill's report names the
  bundle. (Currently: 12 rules, bundle unreported.)
- **AC2**: the inline fixture draws a loud refusal from path B, never double
  delivery; `governed = false` beside an inline block emits the disregard
  message naming S4; inline-plus-rendered emits the coexistence alarm naming
  both.
- **AC3**: every new guard goes red against the pre-fix code (mutation), and
  every fixture provably contains the state it names.

## Consequences

- `staleness.sh` and `skills/remove/SKILL.md` change under this record's
  build stage; `install.sh` is touched only if its enumeration drifts from D1.
- The D4 test is new coverage; no existing test is weakened or deleted.
- `decision-0070` is untouched — D3 already made the bundle the adoption act;
  this record makes its removal reachable.
- `decision-0072`'s "remains the clean exit" claim about `/trellis:remove`
  becomes true; today it is aspirational.

## Self-check

Sections present (Context, Decision state, Decision, Consequences); evidence
is reproduced fixtures, quoted with their measured outputs, not recall;
`depends_on` names only records this decision's correctness rests on
(`0068` approved, `0070` approved, `0072` gated), and the research trail is
`informed_by`/provenance; both Open questions carry draft answers flagged as
drafts; scope is the two P1s plus the class mechanism, with the audit's
remaining findings explicitly parked to issues. Promoted `draft → gated` on
this self-check; `approved` remains the maintainer's intent act at the run's
`intent` gate, after the decision-adversary converges.

---
id: decision-0072
type: decision
status: approved  # maintainer intent act 2026-08-03, in session, at the governed run's batched intent gate: "Approve both, in order (0072 then 0073)" — an in-PR flip recording that act per lifecycle; author (agent) != approver (maintainer). The #227 merge remains the separate ship act. Originally: drafted by agent; awaiting the maintainer's intent act. Requested 2026-08-01 ("retire trellis:setup"), evaluated in #219, unblocked by decision-0071 removing the last overlay.
depends_on: [decision-0053, decision-0065, decision-0070, decision-0071]
superseded_in_part_by: [decision-0083, decision-0084]  # 2026-08-30 decision-0083 — the confirm-first row-repair remedy only: point 2's three mismatch shapes and its reseed gate, insofar as they say how a row mismatch is repaired, AND point 2's "a hand-written partial file leaves the project ungoverned" premise — measured on the current tree, an empty rules.toml reconciles to all sixteen rows active at the adaptive posture, so the preset copy is no longer mandatory in order to be governed at all (it remains the advice wherever the posture or the row set matters). A mismatch now reconciles (missing rows added active = true, unknown and duplicate rows quarantined, never deleted), and the repair is announced rather than confirm-gated because quarantine loses nothing. ALL OF THIS IS THE CLAUDE HOOK ONLY: codex-context.mjs still fails closed on any mismatch (invalid-rules, nothing injected), so on Codex point 2's ungoverned-partial-file premise still holds and the preset copy is still mandatory. Parity is owed, not shipped — decision-0083 section 1 and its open questions. What stands: the retirement of /trellis:setup (point 1), what was lost and its mitigations (point 3), the discoverability call (point 4), the rejected /trellis:migrate (point 5), and finding #6's own rule — a confirm-gated writer's gate does not retire with it  # 2026-08-30 decision-0084 — the HOST SCOPING of the pointer above only: "ALL OF THIS IS THE CLAUDE HOOK ONLY … on Codex point 2's ungoverned-partial-file premise still holds and the preset copy is still mandatory. Parity is owed, not shipped" is no longer true. codex-context.mjs now reconciles a mismatched or incomplete row set exactly as staleness.sh does, so point 2's premise is retired on BOTH hosts and the preset copy is mandatory on neither (it remains the advice wherever the posture or the row set matters). Nothing else in this pointer or in decision-0072 changes
owner: agent
date: 2026-08-02
---

# 0072 — retire `/trellis:setup`

## Context

**After `decision-0070`, nothing needs this skill to make a project governed.**
`install.sh` seeds `.trellis/rules.toml` on the curl path; a project-scoped
plugin applies the shipped defaults with no file at all. Both land at 14/14 on
the adaptive posture.

**What is left is a file copy and a one-line edit.** The two presets differ in
exactly two string values:

```
- seeded_from = "conductor"          + seeded_from = "author-adapt"
- strictness  = "firm"               + strictness  = "adaptive"
```

Every rule row is `active = true` in both. So §1, §2 and §5 — most of a 193-line
skill — exist to copy one of two shipped files and change `strictness` in it. The
copy is not optional (see D2); what is optional is the 193 lines around it.

**The one non-trivial part lost its population.** §3 migrated a vendored overlay,
and `/trellis:remove` could not substitute because it *"deletes the whole of
`.trellis/`, including the `rules.toml` you just wrote"*. That mattered while an
overlay existed. `decision-0071` removed the last one — this repository's own —
so §3 now serves nobody.

## Decision

**1. `/trellis:setup` is retired.** `plugins/trellis/skills/setup/` is deleted,
with its manifest entry and registrations.

**2. The replacement has two shapes, and which one applies depends on the
project's state.** With **no `.trellis/rules.toml`**, copy a complete preset —
`reference/rules-a.toml` for firm, `rules-b.toml` for adaptive — then
`active = false` on any row to turn a rule off. With a **file already present**,
edit `strictness` in place and never copy a preset over it: both presets set
every row active, so a posture-only change made by copying discards every row
the consumer disabled. And with the **one-line `governed = false` opt-out** (`decision-0070` D5),
re-enabling is a replace rather than an edit: editing `strictness` beside the
opt-out leaves it in force and the hook stays silent, while deleting the line
alone leaves the partial file above. Confirm the intent, then write a complete
preset over it.

That file is the consumer's and is already the authority (`decision-0053`), and
every surface that pointed at the skill now points at these three shapes.

**The copy step is mandatory, and an earlier draft of this record got it wrong.**
It said the replacement was *"one sentence: edit `.trellis/rules.toml`"*, which
is true of the presets and false as an instruction. The hook validates the row
set against what the plugin ships and injects **nothing** when a slug is missing,
so a project-scope install that is governed 14/14 becomes **ungoverned** the
moment someone hand-writes a file containing `strictness = "firm"` and nothing
else. Measured against the hook, not reasoned: every one of the fourteen slugs
came back as `missing:` and no rule was injected. The retired skill's §1 did not
merely set `strictness`; it copied a whole preset first, and that is the part of
it that had to survive.

**3. What is genuinely lost, named rather than discovered later.**

| lost | mitigation |
|---|---|
| diff-and-confirm before replacing existing rows | git is the guard **if the file is committed** — the installer suggests the `git add`, it never runs one |
| slug validation against the shipped payload | **already duplicated** — the hook fails loudly on a mismatch and is not going away |
| a discoverable entry point for a newcomer | real, and weighed below |

The middle row is the reason this is safe rather than merely cheap: the check
that mattered was never unique to the skill.

**4. Discoverability is accepted as a cost, on the maintainer's judgement that
there are no newcomers yet** — one user, on Claude Code. A first-time-user review
had already found the page presenting `/trellis:setup` as a grey comment rather
than a step, so what is being given up is smaller than its line count suggests.
If that changes, the replacement is a documented file format, not a resurrected
skill.

**5. NOT proposed: a `/trellis:migrate` skill.** Floated and rejected in the same
session. Splitting a skill nobody needs into two skills nobody needs is not an
improvement.

## Consequences

- `decision-0065`'s *"setup writes exactly one file … and nothing else, ever"*
  becomes vacuous rather than false; recorded as superseded in part for accuracy.
- `/trellis:remove` is unaffected and remains the clean exit.
- The hook's staleness nudge no longer points at a skill that exists; its overlay
  branches are now unreachable in practice and are left in place deliberately —
  a consumer on an old checkout may still hit them.

## What the sweep found

Retiring the skill touched **31 references across 12 live files** — the hook's
migration nudges, the installer's next-step lines, four READMEs, the landing page
and its source, `cli/main.go`'s usage text, and the remove skill.

Three of those files were **not covered by `TestDocsClaimOnlyRealCommands`**,
the guard whose entire job is to stop the docs advertising a skill the plugin
lacks: `plugins/trellis/README.md` (5 references), `skills/remove/SKILL.md` (2),
and `cli/README.md` (1). Fixing only the four covered files would have turned the
guard green with three user-facing surfaces still teaching `/trellis:setup`.
This is the second time that list has been found short — the note above
`docs/lp-content.md` records the first, on 2026-07-31 — so **the failure mode is
the list, not either omission**, and `docSurfaces` becomes a walk of the tree
rather than a third hand-written extension. The walk immediately covered a
surface no version of the list ever had: `hooks/staleness.sh`, which emits slash
commands straight into the consumer's session and is the one surface a user
cannot skim past.

Every message that named the skill as the remedy for a stale overlay now carries
the manual steps instead: delete the overlay, keep `.trellis/rules.toml`. A nudge
that reports a problem without a way out of it is worse than the skill it lost.

**Seven rounds of review found ONE defect class eleven times.** Every finding on this
change was the same thing: a replacement remedy that was wrong about the reader's
actual state, or that dropped a gate the skill had carried. Listed together
because the pattern is the finding, not any single repair:

| # | where | wrong about | found by |
|---|---|---|---|
| 1 | `staleness.sh` coexistence nudge | overlay shape — flat overlays have no `internal/` | cold review |
| 2 | `install.sh` refusal | overlay shape — three values, one remedy; the inline shape has no overlay directory at all | Claude |
| 3 | `staleness.sh` row mismatch | posture — told the agent to clobber a firm project's rows, ungated | Claude |
| 4 | the documented posture recipe | file completeness — a hand-written partial file leaves the project **ungoverned** | Codex |
| 5 | the corrected recipe | file presence — copying a preset over an existing file re-enables every disabled row | Codex |
| 6 | every deletion the hook instructs | authority — the retired skill confirmed first; the replacements did not | Codex |
| 7 | the guard written for #6 | its own coverage — it matched the literal word `delete`, so a remedy saying "drop the unknown ones" walked past it ungated | Codex |
| 8 | the repair remedy | mismatch kind — `$slug_report` emits `missing:`, `unknown:` **and** `duplicate:`; the remedy explained two, so a duplicate had no working repair | Codex |
| 9 | the corrected recipe, again | file shape — `governed = false` is a legal one-line file, and "edit `strictness` in place" leaves the opt-out in force and the hook **silent** | Codex |
| 10 | the docs guard, again | spelling — it matched `/trellis:setup`, so bare-word claims ("Setup installs no receipt", "setup reports bootstrap-only degradation") stayed invisible | Codex |
| 11 | the two shipped READMEs | agreement — one said Codex is not supported yet, the other kept a "Local Codex support" section saying "the same plugin supports Codex" | Codex |

**Three of the six are the same mechanism: retiring a confirm-gated writer
silently retires the gate.** `/trellis:setup` diffed and asked before replacing
rows or deleting an overlay. Each remedy that replaced it inherited the action
and not the confirmation, and none of that was noticed while writing them —
because the thing being removed was a skill, and the gate lived inside it.

Fixing #6 with a class-wide guard rather than three edits surfaced **four more
ungated deletion instructions** in the rendered-file branches that predate this
change entirely and that no reviewer flagged. They are gated here too.

**Finding #7 is the most useful of the seven, because it is about the guard.**
The class-wide check matched the literal string `delete`, and one remedy said
*"drop the unknown ones"* — a consumer-owned row removed with no confirmation,
two clauses away from a reseed path that WAS gated for exactly that risk. A guard
that recognises one verb is a guard against one verb. It now scans a verb list,
excludes slash-command names (`/trellis:remove` is a gated skill, not an
instruction), and asserts a floor on how many messages it matches — so narrowing
the filter fails the test instead of quietly passing an empty set. Widening it
found an eleventh message the narrow version had missed.

**The count is the point.** Nine findings, one class, and the last two arrived
*after* the class had been named and tabled here. Enumerating the reader's
possible states is not something this change did once and got right; it is
something six review rounds did incrementally, each finding a state the previous
round's fix had not considered. Three file shapes (absent, present,
`governed = false`) and three mismatch kinds (`missing:`, `unknown:`,
`duplicate:`) are covered now because they were finally **enumerated from the
code that produces them** — `$slug_report`'s three branches, the `governed`
sentinel, the presence test — instead of from what the author could recall.

The third is the worst of them and is not a wording problem. The rewrite told
the agent to *"copy `rules-b.toml` over `.trellis/rules.toml`"* — the adaptive
preset, every row active. That silently converts a firm project and re-enables
every rule the consumer turned off, with no diff and no confirmation, on a
branch that fires whenever the shipped catalog gains a rule. The retired skill
diff-and-confirmed before a write of that kind, per `floor-intent-gate`.
**Retiring a confirm-gated writer is this decision; replacing its remedy with an
unconditional clobber was not.** The remedy now repairs rows in place and gates
any reseed behind an explicit confirmation, naming `rules-a.toml` for firm.

**The first version of the coexistence rewrite was wrong too.** The
coexistence nudge hard-coded `.trellis/internal/` as the thing to delete, while
the branch it sits in fires for two shapes — a flat pre-`decision-0051` overlay
has no `internal/` directory. A flat-layout project was told to delete something
it does not have; following the advice removed nothing, and because the branch
keys on file existence, the same alarm would fire every session forever. The
remedy is now built from the shape that is present. The generic "names a
deletion and names `rules.toml`" assertion could not catch this — both substrings
were satisfied by the wrong message — so the flat-shape subtest asserts the
specific path instead.

## Open questions

1. **Do the hook's overlay branches (paths A and the legacy stamp) still earn
   their keep?** No project in the family has an overlay after `decision-0071`,
   but an outside consumer might. Retiring them is a separate question with a
   separate blast radius, and is not answered here.

**Finding #10 is the third guard-coverage failure on this change, and the three
rhyme.** `TestDocsClaimOnlyRealCommands` matched `/trellis:setup` and missed
`setup`. `TestEveryDeletionInstructionIsGated` matched `delete` and missed
`drop`. `docSurfaces` was a hand-written list and missed three files. Each guard
was written against **the spelling in front of me at the time**, and each was
then described in this record as covering a class.

`TestNoUnqualifiedSetupClaims` is the answer to #10 and, like the other two
widenings, it found more than the reviewer did: **11 more references across four
files**, one of them in `reference/block-codex.md` — the payload injected into
every governed Codex session, still telling agents about a `setup-verified`
overlay. That file is generated, so the fix belongs in `cli/apply.go` and the
payload is re-rendered; editing the artifact directly broke its own checksum and
the suite said so, which is the generator boundary working.

The exemption list in that guard names *why* each allowed form is allowed — the
retired v0 `setup` CLI is a different artifact from the retired setup skill, and
the remove skill legitimately cleans up what a past setup left behind. An
exemption list without reasons is a mute button.

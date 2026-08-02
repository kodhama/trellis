---
id: decision-0072
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. Requested 2026-08-01 ("retire trellis:setup"), evaluated in #219, unblocked by decision-0071 removing the last overlay.
depends_on: [decision-0053, decision-0065, decision-0070, decision-0071]
owner: agent
date: 2026-08-02
---

# 0072 — retire `/trellis:setup`

## Context

**After `decision-0070`, nothing needs this skill to make a project governed.**
`install.sh` seeds `.trellis/rules.toml` on the curl path; a project-scoped
plugin applies the shipped defaults with no file at all. Both land at 14/14 on
the adaptive posture.

**What is left is a one-line edit.** The two presets differ in exactly two string
values:

```
- seeded_from = "conductor"          + seeded_from = "author-adapt"
- strictness  = "firm"               + strictness  = "adaptive"
```

Every rule row is `active = true` in both. So §1, §2 and §5 — most of a 193-line
skill — exist to change `strictness` in a file the user owns and can edit.

**The one non-trivial part lost its population.** §3 migrated a vendored overlay,
and `/trellis:remove` could not substitute because it *"deletes the whole of
`.trellis/`, including the `rules.toml` you just wrote"*. That mattered while an
overlay existed. `decision-0071` removed the last one — this repository's own —
so §3 now serves nobody.

## Decision

**1. `/trellis:setup` is retired.** `plugins/trellis/skills/setup/` is deleted,
with its manifest entry and registrations.

**2. The replacement is documented in one sentence: edit `.trellis/rules.toml`.**
`strictness = "firm"` for the by-the-book posture; `active = false` on a row to
turn a rule off. That file is the consumer's, it is already the authority
(`decision-0053`), and every surface that pointed at the skill now points at it.

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

**Two rounds of review found the same defect class three times, in three
places.** A remedy is only useful if it names the shape the reader actually has,
and every rewritten message here was first written against one shape:

| where | wrong for | found by |
|---|---|---|
| `staleness.sh` coexistence nudge | flat overlays (no `internal/`) | cold review |
| `install.sh` refusal | flat overlays **and** the inline managed-block shape (no overlay directory at all) | Claude, on the PR |
| `staleness.sh` row mismatch | firm-posture projects and any hand-disabled row | Claude, on the PR |

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

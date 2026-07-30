---
id: decision-0067
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. The direction was given in-session 2026-07-29 ("rethink how the support not-always-on files could be reimagined as skills... even for the plugin path"); the shape below is the agent's, and the ratification is not.
depends_on: [decision-0035, decision-0051, decision-0053, decision-0065, decision-0066]
owner: agent
date: 2026-07-30
---

# 0067 — what must apply is injected; what is consulted is a skill

## Context

`decision-0065` moved rule delivery to the plugin's own SessionStart hooks. That
settled *how* the rules arrive. It did not settle **what else** the payload
carries, and the leftovers are now the expensive part.

Measured on the installed `0.3.0` payload, and on the injected context of a live
session:

| file | bytes | who reads it | when |
|---|---|---|---|
| `reference/rules.md` | 5,370 | both hooks | injected every session |
| `reference/trellis-a.md` | 670 | both hooks | injected every session (firm) |
| `reference/trellis-b.md` | 683 | both hooks | injected every session (adaptive) |
| `reference/invariants.md` | **23,614** | **nobody** | **never read — only pointed at** |
| `reference/rules-<p>.toml` | — | `skills/setup` only | during `/trellis:setup` |

### The pointer is the defect

`invariants.md` is the largest file in the payload and **no code ever opens it.**
`staleness.sh:197,204` computes an absolute path into the version-keyed plugin
cache and splices that *string* into the injected prose. The session this record
was drafted in received:

> read its entry in
> `/Users/…/plugins/cache/kodhama/trellis/0.2.0/reference/invariants.md`

while the machine's user-scope install was already `0.3.0`. So the pointer was
**wrong inside the very session that emitted it**. It is also meaningless on any
other machine, in CI, in a headless runner, and on the install path, where no
plugin cache exists at all.

The substitution that produces it is a `gsub` matching a **backticked string
literal** (`staleness.sh:204`, rewriting `` `.trellis/internal/invariants.md` ``).
Reword that sentence in the source file and the substitution silently no-ops,
shipping a dead path to every session with nothing red.

### The posture headers are a template, not a document

`trellis-a.md` and `trellis-b.md` are **byte-identical except line 5** — nine
lines each, one of which differs (firm vs adaptive). Of the nine:

- lines 1–3 are the preamble that frames the rules,
- **line 5 is the only per-project content**, and it is selected from
  `strictness` in `rules.toml`, which both hooks already parse
  (`staleness.sh:135-143`, `codex-context.mjs:339`),
- line 7 is `@rules.md`, an import the hook resolves itself,
- line 9 is the invariants pointer above.

Two of nine lines exist only so a template can pretend to be a document.

## Decision

**1. The boundary: what must APPLY is injected; what is CONSULTED is a skill.**

A skill's body loads when the model judges it relevant. Rules must hold whether
or not it does. So `rules.md` and the posture prose stay injected and **must not
become a skill** — making them conditional would put `floor-transparency` and
`floor-intent-gate` behind the model's own judgement, which is the failure those
floors exist to prevent. Reference material that a session reads *when a rule is
ambiguous* is the opposite case, and belongs behind on-demand loading.

This is the rule for future payload additions, not only for today's files.

**2. `invariants.md` becomes a skill, on both delivery paths.**

The skill is addressed as `${CLAUDE_SKILL_DIR}/invariants.md` — documented as
"the directory containing the skill's `SKILL.md`… regardless of the current
working directory", and resolving correctly whether the skill arrived inside the
plugin or was copied into a project's `.claude/skills/`. The absolute-path
computation is **deleted, not relocated**.

`SKILL.md` stays a short navigator; the 23.6 KB reference sits beside it and is
read only when the skill is actually used. Always-on cost falls from a 23.6 KB
pointer that can be wrong to one capped description line.

**3. The injected prose refers to the skill by description, never by command
name.** Plugin skills are invoked `/trellis:invariants`; a project-installed copy
is `/trellis-invariants`. Naming either in the always-on text would re-create the
path-fragility one layer up. Model invocation from the description is the
mechanism.

**4. `trellis-a.md` and `trellis-b.md` are deleted.** The preamble moves to the
head of `rules.md`; the strictness sentence becomes a two-branch string chosen
from the `strictness` value both hooks already read; the `@rules.md` self-import
and the invariants pointer cease to exist. **Two payload files and two of the
hook's three runtime string edits go with them.** The third — splicing the live
rows from `rules.toml` — stays, because it is genuinely per-project.

**5. `reference/rules-<p>.toml` moves into `skills/setup/`.** Its only reader is
that skill. Supporting files beside a `SKILL.md` are the documented home for
exactly this.

**6. The refactor is output-preserving, and that is the acceptance criterion.**

`decision-0053` binds this: *"The tested wording is the shipped wording… what the
data validated is the artifact, not a paraphrase of it."* The
`does-trellis-help` experiment validated a specific authority header and layout.
This record moves **where those bytes are assembled from**, and changes **no
byte a session receives** — with one deliberate exception, the invariants pointer
line, which is removed because the skill replaces it.

So: run the hook before and after, diff the emitted `additionalContext`. Expected
difference is exactly the removed pointer line. Anything else is a defect in this
change, not a new decision.

## Consequences

- The payload loses two files outright and gains a skill directory. Net file
  count falls; net always-on context falls by the pointer line.
- **A class of bug is removed rather than fixed**: no absolute path into a
  version-keyed cache is ever computed, so it cannot be stale, wrong on another
  machine, or absent on the install path.
- `staleness.sh` and `codex-context.mjs` both simplify, because both consume the
  same posture files today.
- The install path gets the invariants by copying a directory instead of
  resolving a path — which is the shape the installer rework needs anyway.
- **Cost, named rather than smuggled:** reference material behind a skill is read
  *if the model elects to*. Today's pointer is not read either — it is a string —
  so this is not a coverage reduction against reality. It is a reduction against
  the intent of the current design, and that is the honest way to state it.

## Open questions

1. **What does Codex get?** Skills are a Claude mechanism. `codex-context.mjs`
   computes the same paths and inherits the same fragility. Inlining 23.6 KB per
   session is not acceptable; leaving Codex on a path pointer keeps the defect on
   one host. Not resolved here.
2. **Does the skill listing cost anything measurable?** The listing budget is 1%
   of the context window and descriptions are truncated at 1,536 characters. One
   more skill should be free; `/context` reports the real figure and should be
   checked rather than assumed.
3. **Does the `does-trellis-help` result survive the assembly change?** D6 says
   the bytes are identical, so it should — but the experiment's own harness reads
   the payload files, and moving them may touch it.
4. **Should `rules.md`'s preamble be per-posture too?** Today only line 5 varies.
   Folding the preamble in may make that harder to keep true.

## Self-check (gate)

Per `charters/lifecycle.md`. Every row re-derived from the tree, not from the
drafting conversation.

| # | check | result |
|---|---|---|
| 1 | Every measurement is from the installed payload or a live session | **PASS** — sizes from `0.3.0`; the stale `0.2.0` pointer is quoted from this session's own injected context |
| 2 | The `trellis-a`/`trellis-b` claim is a diff, not a reading | **PASS** — `diff` reports one changed line of nine |
| 3 | "Nobody reads `invariants.md`" is a grep, not an inference | **PASS** — both hooks and both skills grepped; only a path string is built |
| 4 | Upstream constraints checked before proposing | **PASS** — `decision-0053`'s tested-wording clause found and made D6 rather than discovered later |
| 5 | Both hosts checked, not just Claude | **PASS** — `codex-context.mjs:339` uses the same posture files; recorded as Open 1 |
| 6 | The skills mechanism is quoted from its documentation, not assumed | **PASS** — `${CLAUDE_SKILL_DIR}` semantics and supporting-file loading |
| 7 | Acceptance criteria | **PRESENT** — D6 is mechanically checkable by diffing emitted context |
| 8 | Nothing is deleted whose absence is unguarded | **PARTIAL** — D4 deletes two payload files pinned by `TestVendoredPayloadIsCurrent` and `TestRepoOverlayIsCurrent`; both must be advanced in the same change, and neither is advanced by this record |
| 9 | `status: gated` earned | **PASS** — self-check run; its one partial is recorded, not hidden |

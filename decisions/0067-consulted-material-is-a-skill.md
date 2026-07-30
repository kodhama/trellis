---
id: decision-0067
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. The direction was given in-session 2026-07-29 ("rethink how the support not-always-on files could be reimagined as skills... even for the plugin path"); the shape below is the agent's, and the ratification is not.
depends_on: [decision-0035, decision-0051, decision-0053, decision-0065, decision-0066]
owner: agent
date: 2026-07-30
---

# 0067 — what must apply is injected; what is consulted is a skill

> **SCOPE NARROWED after independent adversary review, 2026-07-30 —
> NEEDS-REVISION.** The review verified five factual errors in the original
> draft and one design objection that cuts at its premise. **D4 (deleting the
> posture files) is withdrawn to its own future record**, and this record is now
> **plugin-path-only**. What survives is D1's boundary and a narrowed D2. The
> errors are listed in §Corrections rather than edited away, because the record
> claimed measurement discipline it did not exercise.

## Corrections — what the first draft got wrong

Each verified against the tree, not accepted on the reviewer's word.

1. **"`codex-context.mjs` computes the same paths and inherits the same
   fragility"** — false. `grep -c invariants` on that file returns **0**. Codex
   never rewrites the pointer; it ships `.trellis/internal/invariants.md`
   verbatim into sessions where that path does not exist. That is a *different*
   defect, arguably worse, and the record mis-diagnosed it while grading the row
   PASS.
2. **D4 is blocked by Codex, not merely unresolved for it.**
   `codex-context.mjs:382-385` hard-fails the entire injection unless the prose
   carries **exactly one** `@rules.md`. D4 deletes that placeholder. The record
   blocked its own decision while Open 1 said "not resolved here".
3. **The preamble and pointer live in six and five payload files**, not two —
   `trellis-a/b.md` plus the `block-inline-*` fragments, and
   `cli/payload_test.go` pins the pointer in them. The measured table covered 5
   of 15 files and self-check row 1 graded it PASS.
4. **`decision-0053` was mis-attributed** to the `does-trellis-help` experiment.
   0053 never mentions it; its evidence is `research-0012` /
   `annotation-vs-absence`. Row 4 ("upstream constraints checked") was PASS on a
   citation to the wrong experiment.
5. **"The pointer is meaningless … on the install path"** — false. On a vendored
   overlay the pointer is *relative* and resolves. The absolute-path defect is
   **plugin-path-only**, which is also why the narrowing below is coherent.
6. **The eval harness copies the two files D4 deletes.**
   `eval/experiments/does-trellis-help/run.sh:79-81` copies `invariants.md` and
   `trellis-a.md` into a vendored `.trellis/internal/`. Open 3's "moving them
   *may* touch it" is a verified certainty.

## The design objection, which is better than the record's own reasoning

D1 sorts payload into *must apply* (injected) and *consulted* (a skill).
**`invariants.md` does not sit cleanly in either bucket: it is a deviation
gate.** Its own trigger reads "read its entry … **before deviating**". A gate
that loads only when the model elects to load it is the same conditional-floor
failure D1 rejects for `rules.md` — and the model about to deviate is precisely
the one whose judgment about consulting the reference is least trustworthy.

The record half-conceded this in Consequences ("a reduction against the intent of
the current design") while D1 still asserted the boundary as clean. **A two-way
sort is insufficient; a third bucket is needed** — *reference you may want*
versus *reference you must read before deviating*. That is unresolved and is now
Open 1.

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

**2. `invariants.md` becomes a skill on the PLUGIN path only.**

*(Narrowed from "on both delivery paths".)* The install path keeps a vendored
`invariants.md` and a working relative pointer — correction 5 shows that path is
not broken. The absolute-path computation this replaces exists only in
`staleness.sh`. **`codex-context.mjs`, the `block-inline-*` fragments and the
vendored/inline shapes are explicitly out of scope**, which is what makes this
compatible with `decision-0068`'s sequencing rather than atomic with it.

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

**4. WITHDRAWN — the posture-file collapse moves to its own record.**

~~`trellis-a.md` and `trellis-b.md` are deleted; the preamble moves to `rules.md`;
the strictness sentence becomes a two-branch string.~~ Struck, not deleted,
because the reasoning holds and only the scope was wrong. It cannot land here:

- it is **blocked by Codex** (correction 2);
- it touches **six files, not two** (correction 3);
- it changes the **tested configuration** of a live experiment (correction 6);
- and it **cannot satisfy D6**. Measured on the emitted stream: the `---`
  separator at line 42 exists only to introduce the pointer footer. Delete it and
  the diff is two lines, failing D6's "exactly one difference"; keep it and the
  payload ships an orphaned rule. The acceptance criterion rejects both available
  implementations.

A further defect the first draft missed: posture today is a **file selection**
(`staleness.sh:141-144`), not a string edit. Folding the preamble into `rules.md`
forces the strictness sentence to be injected *inside* that file — requiring a
placeholder or a `gsub` on literal text, which is **the exact mechanism this
record condemns**. So "two of the hook's three runtime string edits go with them"
was wrong: it goes 3 -> 2, and the survivor is the same failure class.

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

1. **Is the injected/consulted boundary sufficient?** See §The design objection —
   `invariants.md` is a deviation gate, and a gate the model elects to load is
   the failure D1 rejects elsewhere. **This cuts at D1's premise and is not
   resolved here.** A third bucket may be needed. Until it is answered, D2 buys
   a correct pointer but does not prove the reference is *read*.
2. **Is the model-invocation assumption measurable?** The whole record turns on
   the model invoking the skill when a rule is ambiguous, and **that is the one
   thing not measured** — in a repo with a working eval harness. `decision-0053`
   set the precedent that this class of question is settled by experiment.
3. **What does Codex get?** Corrections 1-2 show the Codex hook has a different,
   unaddressed defect and hard-fails on D4's deletion. Out of scope here by D2's
   narrowing; it does not go away.
4. **Four Claude Code mechanism claims carry no source and no measurement** —
   `${CLAUDE_SKILL_DIR}` semantics, the skill-listing budget figures, the
   invocation naming, and whether a skill nested inside a vendored bundle is
   discovered at all. Self-check row 6 graded them PASS with no source named.
   They should be cited or measured before implementation.

## Self-check (gate)

Re-derived after independent adversary review. **Four rows the first draft graded
PASS were false**; they are corrected, not restated.

| # | check | result |
|---|---|---|
| 1 | The payload inventory is complete | **WAS FALSE** — measured 5 of 15 files; preamble and pointer live in 6 and 5. Now scoped so the inventory it relies on is the one it measured |
| 2 | Diff-based claims are arithmetic, not estimate | **WAS FALSE** — D6's "exactly one difference" rejects both available implementations. D4 withdrawn |
| 3 | Both hosts checked | **WAS FALSE** — the Codex claim was the opposite of true (`grep -c invariants` = 0), and Codex hard-fails on the deletion |
| 4 | Upstream constraints checked | **WAS FALSE** — `decision-0053` attributed to the wrong experiment; the real harness copies the files D4 deletes |
| 5 | The install-path claim | **WAS FALSE** — the vendored pointer is relative and resolves; the defect is plugin-path-only. This is also what makes the narrowing coherent |
| 6 | Mechanism claims are cited or measured | **STILL FALSE** — four claims carry neither. Open 4 |
| 7 | Guard inventory complete | **WAS FALSE** — row 8 named 2 of 8 pins, and the payload is generated, so "two files are deleted" was really "the generator changes" |
| 8 | The premise survives review | **NO** — Open 1 records an objection that cuts at D1's own boundary and is unresolved |
| 9 | `status: gated` earned | **NO.** Scope narrowed, D4 withdrawn, premise contested. This record is **not ready for an intent act**; it is kept `gated` and open so the corrections are not lost |

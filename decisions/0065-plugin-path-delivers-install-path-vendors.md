---
id: decision-0065
type: decision
status: approved  # maintainer intent act 2026-07-27: "Yes, I approve decision 65 that overrides decision 57. The other was done in a context that is no longer applicable."
depends_on: [decision-0010, decision-0012, decision-0035, decision-0039, decision-0043, decision-0049, decision-0050, decision-0051, decision-0053, decision-0057, decision-0058]
informed_by: []  # research-0013 is an unmerged draft; see §Standing on research-0013
owner: agent
date: 2026-07-27
---

# Decision: the plugin path delivers, the install path vendors

## Decision state

### Decided

- **`/trellis:setup` writes exactly one file — `.trellis/rules.toml` — and nothing
  else, ever.** It creates that file from the chosen preset when absent, and
  replaces its configuration with the chosen preset when present. It vendors no
  payload, patches no instructions file, stamps no version, and verifies no
  checksum.
- **The rules reach the model at session start**, injected by the plugin's own
  `SessionStart` hook from the plugin's payload. The consumer holds no copy, so
  no copy can drift.
- **The two delivery paths separate completely.** The plugin path is
  configuration and never vendors. `install.sh` is vendoring and never
  configures — it copies the plugin *bundle* into `.claude/skills/trellis/` for
  harnesses with no plugin system, and continues to never touch `.trellis/`.
- **A vendored overlay still governs where one exists.** The hook detects
  `.trellis/internal/` and runs its staleness comparison instead of injecting, so
  nothing is delivered twice and no existing consumer breaks.
- **The plugin path offers no delivery-model choice.** It never rewrites, weaves
  into, or otherwise touches the project's own instruction files. There is no
  M1/M2 fork to pick on this path — the question does not arise. See §One model
  on the plugin path.
- **Both hosts deliver.** Trellis already ships a Codex `SessionStart` hook
  (`hooks/codex-context.mjs`, declared by `hooks/codex-hooks.json`). It read the
  vendored overlay as required input, so removing vendoring would have broken it;
  it now reads the plugin payload when no overlay is present, exactly as the
  Claude hook does.

### Open

- None.

### Parked

- Whether the **install path** delivers through a managed block, an inlined
  block, or an unremovable inline payload. It vendors either way; the shape is
  deliberately left open.
- The M2 morph (`decision-0050`) — [issue
  #200](https://github.com/kodhama/trellis/issues/200). Removed from setup and
  **not returning to it**. Speculative, with no established use case.
- Verifying the Codex hook against a live Codex session — [issue
  #199](https://github.com/kodhama/trellis/issues/199). The change is verified by
  direct invocation across all four project shapes, not yet by a real session.

## Context

Two things were true at once and pulled against each other. `/trellis:setup`
vendored a 373-line overlay into every consumer, and the overlay was the only way
the rules reached a model. Keeping the copies current was therefore load-bearing,
which is what `decision-0043`'s staleness stamp and `decision-0035`'s sync-guard
exist for.

**The staleness surface was never delivering.** `hooks/staleness.sh` emitted a
bare top-level `{"additionalContext": "..."}`; Claude Code accepts that JSON and
silently discards it. Measured 2026-07-27, flat versus nested, with file tools
disabled so injected context was the only possible source: the nested codeword
came back, the flat one was absent. The contract test decoded the flat shape, so
it passed while nothing was delivered. `kodhama/stewards` sat at
`payload@b2395d518ec8` against a shipped `payload@0760a802ccd1` — 383 lines of
`invariants.md` against the current 303 — and was never nudged.

So the cost of vendoring was being paid without the mechanism that made it safe.

`research-0013` explored plugin-native delivery and was parked by maintainer
direction with the framing *"model 1 stays, add model 2"*. It also doubted that
headless `claude -p` was covered by a hook. **That doubt is now falsified** — the
hook fires under `claude -p` and its injection reaches the model, measured. What
survives of the objection is narrower: a bare subagent is not a session, and
plugin provisioning in an ephemeral container is a separate problem.

`decision-0051:192-195` records that *"the per-session hook variant was considered
and rejected"* because *"no reader exists"*. A reader exists now — this hook is
it.

## Decision

### Setup's one job

`/trellis:setup` selects a preset and writes `.trellis/rules.toml`. A preset is a
whole `rules.toml`: `strictness`, `seeded_from` provenance, and one row per rule.
The payload ships two — `a·conductor` and `b·author-adapt`. Setup copies the
preset file rather than composing it, so a future preset that deactivates rows
needs no change to the skill.

Writing over an existing `.trellis/rules.toml` destroys consumer-owned rows. The
skill SHALL `diff` first, show what changes, name any row returning from
`active = false` to `active = true`, and require explicit confirmation. With no
human available it SHALL report and stop, never write (`floor-intent-gate`).

### What the hook delivers

For a project holding `.trellis/rules.toml` and no `.trellis/internal/`, the hook
injects the always-loaded chain from the plugin payload: the posture header with
its `@rules.md` import resolved, the rule bodies, and the project's live rows.
The tested wording stays the shipped wording (`decision-0053`); the one edit
repoints the invariants pointer at the plugin's own copy, which is where the file
is in this mode and which therefore cannot go stale.

`.trellis/rules.toml` is the opt-in signal. The plugin may be installed
user-wide; a project that never adopted Trellis is never governed by surprise.

### One model on the plugin path

The old skill framed itself as the **M1 "alongside" overlay** and hosted the
**M2 morph** as an alternative. That fork is removed, not relocated: on the
plugin path there is exactly one model and no choice to make.

**The vocabulary no longer fits and is retired here rather than reused.** M1 was
defined as a `.trellis/` bundle *plus a managed block in the instructions file*.
The plugin path now has neither — one config file and a runtime injection. Saying
"the plugin path is always M1" would import a definition that has stopped being
true, so the plugin path is described by what it does instead: **it writes
`.trellis/rules.toml` and never touches a file the project authored.**

That property is the point, and it is worth more than the naming: whatever the
project's own `CLAUDE.md`, `AGENTS.md` or README say, Trellis on the plugin path
cannot have edited them.

**Morph is speculative.** It rewrites the consumer's own prose to interleave the
rules, and it has no established use case — the maintainer's own assessment on
2026-07-27 was that it "was just a concept". If it ever returns it is a separate,
explicitly-invoked skill, plausibly one that weaves the rules in *and removes the
plugin*, which is a different product from this one. It does not return to setup.

The install path's model stays open: it vendors, and whether it delivers through
a managed block, an inlined block, or something with no removal affordance at all
is deliberately not decided here.

### Both hooks, one rule

`hooks/staleness.sh` (Claude) and `hooks/codex-context.mjs` (Codex) take the same
two paths and use the same discriminator: **the `.trellis/internal/` directory**,
not any file inside it.

Directory present → vendored mode: every payload file within is required, and a
missing one is a broken overlay that fails loudly. Directory absent → the plugin's
payload. A *file* discriminator would silently convert a half-deleted overlay into
a config-only project, which is a mode switch nobody asked for; the existing Codex
failure-vocabulary test caught exactly that.

### The envelope

Every `SessionStart` emission SHALL use the nested envelope
`{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"..."}}`.
The flat form is silently discarded. Contract tests SHALL decode only the nested
shape, so a regression fails rather than passing quietly.

## This decision operates under `decision-0058` phase 4

**An earlier draft of this record did not cite `decision-0058` at all.** That was
the sharpest defect in it: 0058 is `approved` (maintainer intent act 2026-07-24)
and its phase table governs precisely this act. Both an independent adversary and
the repository's own Codex reviewer raised it separately.

0058 does not forbid this change. Its phase 4 (`:123`) **conditionally authorises**
it:

> | 4 | Other hosts, including any Claude hook replacement | A host-native proof at
> least as strong as Phase 1; remove or disable the old transport in the same
> change so rules still arrive once |

and `:188-189` sets the boundary condition:

> The Claude import may eventually be replaced by a verified Claude hook, but
> **never while both paths would inject the full rule payload.**

This decision claims phase 4 and must show it clears both bars:

**Host-native proof.** Phase 1's bar is a live startup positive control plus
production contract and reversibility tests. Measured against a real Claude
session with `Read`/`Bash`/`Glob`/`Grep` disabled, so injected context was the
only possible source: the model quoted a rule's failure example verbatim,
reported the posture, and stated that the floor rules hold regardless of their
rows. Contract tests cover both hooks across config-only, vendored, absent and
partial-overlay projects. Reversibility is §Migrating a vendored overlay.

**Rules arrive exactly once.** Both hooks discriminate on the `.trellis/internal/`
directory. A project that still has an overlay keeps its import transport and the
hooks inject nothing; a project without one has no import block and receives the
injection. There is no state in which both paths carry the payload — that is
tested, not asserted.

**What 0058 loses.** Its rule 2 — *"Claude's existing import transport stands.
Setup and refresh retain the managed import block in `CLAUDE.md`"* — is
superseded, which is exactly the "remove or disable the old transport" phase 4
contemplates. Its rule 1's *"from `.trellis/internal/`"* narrows to vendored
projects. Its `:182-183` claim that `.trellis/internal/` "remains useful rather
than vestigial" holds only for those projects.

**What 0058 keeps.** Phases 1–3 stand untouched: this decision claims nothing
about Codex resume, clear, compact, subagents, desktop, IDE, headless or cloud,
and `:186-187`'s warning that those are "named next experiments, not fine print"
applies to the Claude hook exactly as written. `startup` and `resume` are
measured; nothing else is.

## Standing on `research-0013`

An earlier draft listed `research-0013` under `informed_by`. **It is not in this
repository** — it exists only on an unmerged branch, and its own title records it
as parked. Building on a draft is what `inv-directional-flow` forbids, so the
dependency is removed rather than dressed up.

The relationship is the other way around: this work **tested** two of that
draft's premises and found one false. It doubted that headless `claude -p` was
covered by a hook; measured, the hook fires and its injection reaches the model.
Nothing here is built on it, and if it lands later it should be updated to record
that result.

## Supersession

**`decision-0051` rule 1** — `.trellis/internal/` as trellis-authoritative vendored
state. Nothing is copied there on the plugin path. The file *layout* remains
correct for projects that already have one, and `install.sh` still vendors the
bundle, so rule 1 survives for the install path.

**`decision-0053` rule 2** — delivery of the rows through the managed block. The
rows still govern at read time, unchanged; only their carrier moves from an
import to an injection.

**`decision-0043` rules 2–3** — the copied version stamp and the file-to-file
staleness compare. Both remain live for vendored projects and become inert where
nothing is vendored: with no copy there is nothing to be stale.

**`decision-0035` rule 2** — the repo running setup, committing `.trellis/` plus
the block, with a CI sync-guard. The guard's subject no longer exists on the
plugin path.

**`decision-0057` rules 3–4** — the managed Trellis block remaining in
`CLAUDE.md`, byte-identical to `block-claude.md`, plus its CI guard. **This is
the sharpest supersession here**, because `0057:17-19` records that the block is
present by *explicit maintainer requirement*.

**The maintainer overrode it directly on 2026-07-27**, naming it: *"I approve
decision 65 that overrides decision 57. The other was done in a context that is
no longer applicable."* The context that changed: when 0057 was written the
managed block was the only way rules reached a Claude model, and the staleness
hook that should have flagged drift was emitting into a shape the host discards
— so the block was load-bearing and its failure mode was invisible. A working
hook removes the first condition and fixes the second.

Trellis's own `CLAUDE.md` keeps its block: the repo is a vendored consumer and
stays one until it migrates, so 0057's self-application guard still has a
subject.

**`decision-0049`** — offering to hide `.trellis/` from consumer linters. Its
subject evaporates when only `rules.toml` remains.

**`decision-0050`** — the M2 morph's cold-isolated rewrite. Not reworked, and not
merely relocated: the plugin path offers no morph at all, so 0050's subject has no
home on it. 0050 is not wrong and is not superseded on its merits — a future
opt-in morph skill would still want its isolation contract.

**`decision-0010` is amended, not superseded**, and the amendment is wider than a
first reading suggests. 0010 permits a support CLI but holds that *"the
methodology runs without it"*. Two things narrow that here:

- On the plugin path the hook is the sole carrier, so the methodology no longer
  runs without *something* executing. The Claude hook is bash — no node, no
  compiled artifact — which keeps the spirit.
- **The Codex hook has always been Node** (`node "${PLUGIN_ROOT}/hooks/codex-context.mjs"`).
  That predates this decision and is not introduced by it, but this record is the
  first to say so plainly: 0010's runtime-free claim has not been literally true
  on the Codex path for some time.

The install path remains the genuinely runtime-free route, which is what keeps
0010's intent intact where it still holds.

`specs/0004` (`/trellis:remove`) is **not** superseded and is left as it stands:
vendored overlays still exist in the wild and removing them is exactly its job.

## The gap this decision accepts

**An earlier draft of this record claimed Codex would be left ungoverned. That was
wrong** — it rested on a survey finding that `hooks/codex-context.mjs` did not
exist. It does, it ships, and it is declared by `hooks/codex-hooks.json`. The real
risk was subtler and is fixed here: the Codex hook required the vendored overlay,
so removing vendoring would have broken working delivery rather than leaving an
absence. Recorded rather than quietly amended, because the corrected version is
the weaker claim and the graph should show which one this decision acted on.

What genuinely remains uncovered:

- **The install path delivers nothing after this change** —
  [#201](https://github.com/kodhama/trellis/issues/201). `install.sh` copies
  `hooks/hooks.json` and `staleness.sh` into `.claude/skills/trellis/` but
  **registers no hook** — no settings write, no plugin manifest — so a vendored
  skill install has no injection, and setup no longer writes a managed block.
  Its delivery has to be a file the model already loads, because that path exists
  for hosts that do not run our code.
- **Bare subagent workers**, which are not sessions and so never fire
  `SessionStart`, on either host.
- **Plugin provisioning in ephemeral containers**, a family delivery problem
  rather than a trellis one.
- **`compact`, `clear`, `fork`, cloud and CI-runner surfaces** are unverified.
  `startup` and `resume` are measured, including under headless `claude -p`.

## Consequences

- A new consumer's Trellis footprint is one file, ~20 lines, instead of 373.
- The overlay cannot go stale, because there is no overlay.
- Rule changes take effect at the next session, not at the next `setup` run.
- The install path delivers no rules at all until it gets its own mechanism.
- `install.sh` is now the only vendoring route, and the only route for harnesses
  without plugins.

## Acceptance criteria

- `skills/setup/SKILL.md` issues no write command whose target is anything but
  `.trellis/rules.toml`, enforced by test rather than by review.
- The setup skill requires explicit confirmation before overwriting an existing
  `.trellis/rules.toml`, and refuses to write without a human.
- `hooks/staleness.sh` emits only the nested envelope, and its contract test
  decodes only that shape.
- A project holding only `.trellis/rules.toml` receives the posture header, the
  rule bodies and its live rows — verified against a real session with file tools
  disabled, so injected context is the only possible source.
- A project holding a vendored `.trellis/internal/` receives the staleness nudge
  and **not** the rule bodies.
- `install.sh` never reads or writes `.trellis/`, and its bundle manifest
  advances in the same commit as any payload change.
- No surface on the plugin path offers a delivery-model choice, and no skill on
  it writes to a file the project authored.
- Both hooks deliver from the plugin payload for a config-only project, and both
  fail loudly rather than switching modes on a partial overlay.
- The install path's delivery gap is recorded in an issue before ratification —
  done, [#201](https://github.com/kodhama/trellis/issues/201).

## Lifecycle record

Authored 2026-07-27 from the maintainer's direction that the plugin path carries
no vendoring and setup's only job is writing the chosen preset to
`.trellis/rules.toml`.

The direction was put twice. The first time I raised that `decision-0057` and
`research-0013`'s parking were the maintainer's own recorded direction *against*
it; the direction was then restated explicitly, which is what this record acts on.

**`approved` 2026-07-27** by the maintainer's intent act quoted in the frontmatter,
after two independent reviews — an adversary pass and the repository's Codex
reviewer — which agreed on four defects without seeing each other's work. All are
fixed on this branch; the review record is in PR #198.

Nothing is deleted from a consumer repo by this change: a project that already
carries a vendored overlay keeps it, and keeps being governed by it, until it
chooses the migration.

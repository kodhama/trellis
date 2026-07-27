---
id: decision-0065
type: decision
status: gated
depends_on: [decision-0010, decision-0012, decision-0035, decision-0039, decision-0043, decision-0049, decision-0050, decision-0051, decision-0053, decision-0057]
informed_by: [research-0013]
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
- **Codex loses rule delivery entirely** until a Codex hook ships. Named as a
  regression, not a scope note — see §The gap this decision accepts.

### Open

- None.

### Parked

- Whether the **install path** delivers through a managed block, an inlined
  block, or an unremovable inline payload. It vendors either way; the shape is
  deliberately left open.
- The M2 morph (`decision-0050`) — [issue
  #200](https://github.com/kodhama/trellis/issues/200). Removed from setup and
  **not returning to it**. Speculative, with no established use case.
- Codex delivery — [issue #199](https://github.com/kodhama/trellis/issues/199). The
  prototype at `eval/experiments/codex-hook-delivery/` measured working on Codex
  CLI 0.145.0.

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

### The envelope

Every `SessionStart` emission SHALL use the nested envelope
`{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"..."}}`.
The flat form is silently discarded. Contract tests SHALL decode only the nested
shape, so a regression fails rather than passing quietly.

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
present by *explicit maintainer requirement*. It is overridden by the maintainer's
later direction of 2026-07-27, in the same voice: setup vendors nothing, the
plugin path carries no files. `0057`'s scoping of itself as *"a repo shared-entrypoint
change, not a new Trellis delivery contract"* is what makes this a bounded
override rather than a reversal.

**`decision-0049`** — offering to hide `.trellis/` from consumer linters. Its
subject evaporates when only `rules.toml` remains.

**`decision-0050`** — the M2 morph's cold-isolated rewrite. Not reworked, and not
merely relocated: the plugin path offers no morph at all, so 0050's subject has no
home on it. 0050 is not wrong and is not superseded on its merits — a future
opt-in morph skill would still want its isolation contract.

**`decision-0010` is amended, not superseded.** It permits a support CLI but
holds that *"the methodology runs without it"*. With the hook as sole carrier on
the plugin path, that clause narrows: the methodology runs without a *binary*,
and the hook is bash — no node, no compiled artifact. The install path remains
the runtime-free route, which is what keeps 0010's intent intact.

`specs/0004` (`/trellis:remove`) is **not** superseded and is left as it stands:
vendored overlays still exist in the wild and removing them is exactly its job.

## The gap this decision accepts

**Codex receives no Trellis rules at all.** Setup's vendored Codex block is gone
and no Codex `SessionStart` hook ships. Until one does, a Codex-only consumer that
runs the new setup is configured but ungoverned.

This is a regression, stated as one. Three things bound it: the maintainer's
direction was explicitly *"let's focus on Claude, which is what we're running at
the moment"*; the prototype at `eval/experiments/codex-hook-delivery/` already
measured working on Codex CLI 0.145.0, so closing the gap is a port rather than a
design; and an existing consumer keeps its vendored overlay until it chooses to
remove it.

**Also uncovered:** bare subagent workers, which are not sessions and so never
fire `SessionStart`; and plugin provisioning in ephemeral containers, which is a
family delivery problem rather than a trellis one.

## Consequences

- A new consumer's Trellis footprint is one file, ~20 lines, instead of 373.
- The overlay cannot go stale, because there is no overlay.
- Rule changes take effect at the next session, not at the next `setup` run.
- Codex is ungoverned on the plugin path until its hook ships.
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
- The Codex gap is recorded in an issue before this decision is ratified — done,
  [#199](https://github.com/kodhama/trellis/issues/199); morph's disposition is
  [#200](https://github.com/kodhama/trellis/issues/200).

## Lifecycle record

Authored 2026-07-27 from the maintainer's direction that the plugin path carries
no vendoring and setup's only job is writing the chosen preset to
`.trellis/rules.toml`.

The direction was put twice. The first time I raised that `decision-0057` and
`research-0013`'s parking were the maintainer's own recorded direction *against*
it; the direction was then restated explicitly, which is what this record acts on.

`gated`. Awaiting an independent adversary pass and the maintainer's intent act.
The implementation is on the same branch and does not depend on this record being
ratified to be reverted — nothing is deleted from a consumer repo by this change.

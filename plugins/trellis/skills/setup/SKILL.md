---
name: setup
description: Choose a Trellis preset for this project and write it to .trellis/rules.toml — the only file this skill ever writes. It vendors nothing: the rules themselves are delivered at session start by the plugin's SessionStart hook. Use when the user asks to set up, add, install, configure, refresh, or change the posture of Trellis in their repo.
---

# Set up Trellis in this project

**This skill writes exactly one file: `.trellis/rules.toml`.** It creates that file from the
chosen preset when it is absent, and replaces its configuration with the chosen preset when it is
present. It writes nothing else, anywhere, ever.

**Since `decision-0070`, running this is no longer how a project becomes governed.** `install.sh`
seeds `rules.toml` on the curl path, and a project-scoped plugin applies the shipped defaults
without any file at all — both at posture B, every rule active. So the two remaining reasons to run
this skill are: to choose the **firm** posture instead of the default adaptive one, and to migrate a
project off a pre-`decision-0065` vendored overlay (§3). Everything else it does, the default
already does.

Worth knowing before you invoke it: the two presets differ in exactly two string values,
`strictness` and `seeded_from`. Every rule row is `active = true` in both. A user who wants the firm
posture, or wants one rule off, can edit that line themselves — this skill is a convenience over a
one-line change, and `kodhama/trellis#219` tracks whether it should exist at all.

That is the whole job. If you find yourself copying a payload file, patching an instructions file,
stamping a version, or touching anything under `.trellis/internal/`, you are working from an older
version of this skill's intent — stop and re-read this file.

## Why there is nothing to copy

The rules reach the model at session start, injected by the plugin's own `SessionStart` hook
(`hooks/staleness.sh`) straight from the plugin's payload. The project holds no copy, so there is
no copy to install, refresh, verify, or let drift.

`.trellis/rules.toml` is the one thing that is genuinely per-project — it says which rules are
active here — so it is the one thing that lives in the repo.

**Two paths, cleanly separated:**

- **Plugin path (this skill)** — configuration only. No vendoring.
- **Install-script path (`install.sh`)** — vendoring only, for harnesses with no plugin system. It
  vendors the plugin *bundle* into `.claude/skills/trellis/` and, since `decision-0070` D2, seeds
  `.trellis/rules.toml` from the shipped preset when none exists — one config file, never
  overwritten. It configures nothing else.

Do not blur them. This skill never vendors.

**If `.trellis/rules.toml` declares `governed = false`, STOP and ask before writing anything.**
That line is a recorded decision not to be governed in this project (`decision-0070` D5), and
replacing the file with a preset silently revokes it. Confirm the user wants governance switched
back on, and tell them that is what you are doing.

## The presets

The payload ships two, and no others. `seed` and `custom` stay parked (`decision-0033`,
`decision-0051` rule 7) — do not offer them.

- **A · conductor** — hold the rules firmly, by-the-book.
- **B · author-adapt** — same rules, follow by default and adapt out loud. **Default** when the
  user is unsure.

A preset is a whole `rules.toml`: its `strictness`, its `seeded_from` provenance, and one row per
rule. Today the two differ only in `strictness` and `seeded_from` — every row is `active = true` in
both. If a future preset turns rows off, this skill needs no change, because it copies the preset
file rather than composing it.

## 0. Two things to resolve first

**The plugin root.** Every path below is relative to it. The host exports it:
Claude sets `CLAUDE_PLUGIN_ROOT`, Codex sets `PLUGIN_ROOT`. Use
`${CLAUDE_PLUGIN_ROOT:-$PLUGIN_ROOT}` and, if both are empty, **stop and say so**
— without it the copy below would read from `/reference/`, which is not a path
this plugin owns.

**The project root.** Run from the repository root, so `.trellis/rules.toml`
lands at the top of the project rather than in whatever subdirectory the session
happens to be in. If the working directory is not the root, change to it first.

## 1. Pick the preset

- The user named one (`conductor`/`a`, `author-adapt`/`b`) → use it.
- `.trellis/rules.toml` exists and the user asked only to "refresh" → read its `strictness`
  (`"firm"` → A, `"adaptive"` → B) and re-apply that same preset.
- Otherwise → **ask.** Present exactly the two above. Never assume a default when a human is
  present; if none is (an autonomous run) and no preset can be read, **stop and say so** rather
  than picking one.

## 2. Write it

`<p>` is `a` for conductor, `b` for author-adapt.

**If `.trellis/rules.toml` does not exist** — write it:

```sh
root="${CLAUDE_PLUGIN_ROOT:-$PLUGIN_ROOT}"
mkdir -p .trellis
cp "$root/reference/rules-<p>.toml" .trellis/rules.toml
```

**If it already exists** — this overwrites the project's own configuration, including any rows they
turned off by hand. That is a destructive write on a consumer-owned file, so:

1. `diff` the current file against `$root/reference/rules-<p>.toml`.
2. If they are identical, say so and write nothing.
3. Otherwise **show the diff and ask for explicit confirmation**, calling out by name any row that
   is `active = false` today and would return to `active = true`. Applying a preset is the point of
   this skill; doing it silently over someone's edits is not.
4. On confirmation, copy as above. If declined, change nothing and say so.

If no human is available to confirm, **do not write** — report what would have changed and stop
(`floor-intent-gate`).

## 3. Migrate a vendored overlay, if there is one

A project set up before plugin delivery carries `.trellis/internal/` and a
managed block in its instructions file. **Those still govern it** — the hook
detects the overlay and steps aside — so such a project is running on a copy the
plugin no longer refreshes, and its staleness nudge now points here.

This is the one place the skill removes files. Removing a retired overlay is not
vendoring, and without it a vendored project has no way forward: `/trellis:remove`
deletes the whole of `.trellis/`, including the `rules.toml` you just wrote.

If `.trellis/internal/` exists, **offer** the migration — never impose it:

1. Say what will be removed: the `.trellis/internal/` directory, and the block
   between the `<!-- trellis:begin` and `<!-- trellis:end -->` markers in the
   instructions file that carries them (`CLAUDE.md`, or whichever file holds the
   markers). Say what is kept: `.trellis/rules.toml`, with their rows intact.
2. Say what changes: the rules will arrive from the plugin at session start
   instead of by import, and **the switch takes effect next session**.
3. On confirmation, delete the directory and remove exactly the marked block,
   leaving every other byte of that file untouched. If more than one file carries
   the markers, ask which — never guess.
4. If declined, change nothing and say the project stays on its vendored copy.

With no human available, do not migrate; report that the project has a vendored
overlay and stop.

## 4. Check what you wrote

Read the file back and confirm it parses: `strictness` is `"firm"` or `"adaptive"`, and there is one
`[rules]` row per rule the payload ships. If it does not, say so plainly — do not repair it by hand.
A hand-repaired config is a second writer, and second writers drift (kodhama/trellis#112).

Note for the report, without changing anything: a floor row (`floor-transparency`,
`floor-intent-gate`) set to `active = false` **is not honored**. The floors hold regardless of their
row.

## 5. Report

Say exactly:

- which preset was applied, and whether the file was created, replaced, or left
  alone — or, on a refresh, that nothing was written at all;
- **every path touched.** Normally that is `.trellis/rules.toml` and nothing
  else. **If step 3 ran, name the deleted `.trellis/internal/` directory and the
  instructions file the managed block was removed from** — the user authorised
  destructive edits and the handoff must show them, not bury them under a
  sentence about one config file;
- that the rules themselves arrive at session start from the plugin, so **the change takes effect in
  the next session, not this one**;
- any floor row set `active = false`, named, as overridden-by-floor.

## 6. Hand back

Perform no git operations and impose no landing workflow (`decision-0048`). Committing
`.trellis/rules.toml` is the user's call.

## What this skill no longer does

Recorded so a reader of an older overlay knows where each job went, rather than finding it simply
gone:

| Was | Now |
|---|---|
| Copy `invariants.md`, `rules.md`, `trellis-<p>.md` into `.trellis/internal/` | Injected at session start by `hooks/staleness.sh` |
| Stamp `.trellis/internal/version` | No vendored copy exists, so nothing can be stale |
| Patch the `trellis:begin`/`trellis:end` managed block | The hook delivers the same chain; the block is no longer written |
| Verify copies against the checksum manifest | Nothing is copied |
| Offer to add `.trellis/` to lint-ignore files (`decision-0049`) | Only `rules.toml` remains, and a project lints its own config |
| Migrate a pre-`decision-0051` flat overlay | Step 3 above, on confirmation |
| M2 morph, a model-driven rewrite of the project's own instructions (`decision-0050`) | Removed, and not returning — see below |

**This skill never touches a file the project authored.** There is no "alongside
versus morph" choice to offer on the plugin path — the question does not arise
here. Whatever the project's own `CLAUDE.md`, `AGENTS.md` or README say, Trellis
did not write it. If a morph ever ships it will be a separate skill the user
invokes on purpose, not a mode of this one.

**A project that still has a vendored `.trellis/internal/` keeps working.** The hook detects it and
runs its staleness comparison instead of injecting, so nothing is delivered twice. Until that
directory is removed, the project is governed by its vendored copy rather than by the plugin's
payload — removing it is `/trellis:remove`'s job, not this skill's.

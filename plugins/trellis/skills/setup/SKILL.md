---
name: setup
description: Choose a Trellis preset for this project and write it to .trellis/rules.toml — the only file this skill ever writes. It vendors nothing: the rules themselves are delivered at session start by the plugin's SessionStart hook. Use when the user asks to set up, add, install, configure, refresh, or change the posture of Trellis in their repo.
---

# Set up Trellis in this project

**This skill writes exactly one file: `.trellis/rules.toml`.** It creates that file from the
chosen preset when it is absent, and replaces its configuration with the chosen preset when it is
present. It writes nothing else, anywhere, ever.

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
  vendors the plugin *bundle* into `.claude/skills/trellis/` and never touches `.trellis/`.

Do not blur them. This skill never vendors; that script never configures.

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
mkdir -p .trellis
cp "${TRELLIS_PLUGIN_ROOT}/reference/rules-<p>.toml" .trellis/rules.toml
```

**If it already exists** — this overwrites the project's own configuration, including any rows they
turned off by hand. That is a destructive write on a consumer-owned file, so:

1. `diff` the current file against `${TRELLIS_PLUGIN_ROOT}/reference/rules-<p>.toml`.
2. If they are identical, say so and write nothing.
3. Otherwise **show the diff and ask for explicit confirmation**, calling out by name any row that
   is `active = false` today and would return to `active = true`. Applying a preset is the point of
   this skill; doing it silently over someone's edits is not.
4. On confirmation, copy as above. If declined, change nothing and say so.

If no human is available to confirm, **do not write** — report what would have changed and stop
(`floor-intent-gate`).

## 3. Check what you wrote

Read the file back and confirm it parses: `strictness` is `"firm"` or `"adaptive"`, and there is one
`[rules]` row per rule the payload ships. If it does not, say so plainly — do not repair it by hand.
A hand-repaired config is a second writer, and second writers drift (kodhama/trellis#112).

Note for the report, without changing anything: a floor row (`floor-transparency`,
`floor-intent-gate`) set to `active = false` **is not honored**. The floors hold regardless of their
row.

## 4. Report

Say exactly:

- which preset was applied, and whether the file was created, replaced, or left alone;
- the one path written — `.trellis/rules.toml` — and that nothing else was touched;
- that the rules themselves arrive at session start from the plugin, so **the change takes effect in
  the next session, not this one**;
- any floor row set `active = false`, named, as overridden-by-floor.

## 5. Hand back

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
| Migrate a pre-`decision-0051` flat overlay | `/trellis:remove`'s job — see below |
| M2 morph, a model-driven rewrite of the project's own instructions (`decision-0050`) | Removed from this skill; tracked separately |

**A project that still has a vendored `.trellis/internal/` keeps working.** The hook detects it and
runs its staleness comparison instead of injecting, so nothing is delivered twice. Until that
directory is removed, the project is governed by its vendored copy rather than by the plugin's
payload — removing it is `/trellis:remove`'s job, not this skill's.

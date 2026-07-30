---
id: decision-0068
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. Direction given in-session 2026-07-30 ("let's focus on claude only for the install path, potentially with the symlink alternative if you think it has merit"). The measurements are the agent's; the ratification is not.
depends_on: [decision-0043, decision-0053, decision-0058, decision-0065, decision-0066]
owner: agent
date: 2026-07-30
---

# 0068 — the install path delivers rules through `.claude/rules/`, Claude only

## Context

**The install path delivers no rules at all. This is measured, not inferred.**

`decision-0065` moved rule delivery to the plugin's SessionStart hooks. The
install path was left vendoring the whole bundle into `.claude/skills/trellis/`
and registering nothing — which issue #201 reported and which this record
confirms by experiment.

### The measurement, including the reading that was wrong

Four headless runs in a scratch git repo, asking whether Trellis rules were in
the always-loaded context:

| # | setup | user scope | result |
|---|---|---|---|
| 1 | bundle copied to `.claude/skills/trellis/` — exactly what `install.sh` produces | **included** | TRELLIS=YES |
| 2 | same | **excluded** | **TRELLIS=NO** |
| 3 | nothing installed — baseline | **excluded** | TRELLIS=NO |
| 4 | `.claude/rules/trellis.md` | **excluded** | **TRELLIS=YES** |

**Run 1 was read as "the install path works" and that was wrong.** The
user-scope Trellis plugin, installed on the test machine earlier the same day,
governs any repo carrying a `.trellis/rules.toml` — including the scratch one. A
control (run 3) exposed it. The honest result is run 2 against run 3: **the
vendored bundle delivers nothing**, even though it contains
`.claude-plugin/plugin.json` and `hooks/hooks.json`.

The documented skills-directory-plugin path requires accepting the workspace
trust dialog, which a headless run cannot grant — and headless CI is exactly the
case the install path exists to serve.

### What `.claude/rules/` gives that a hook does not

- Loaded at launch "with the same priority as `.claude/CLAUDE.md`" when the rule
  carries no `paths` frontmatter. No hook, no `settings.json` write, no trust
  dialog, no `.claude-plugin/` manifest.
- **Discovery is symlink-aware, and the link resolves first**: an `@` import
  inside a symlinked rules file resolves relative to the **real file's**
  directory, not the link's. Measured both ways:

  | file at `.claude/rules/trellis.md` | import written | result |
  |---|---|---|
  | symlink → `.trellis/trellis.md` | `@rules.toml` | **loaded** |
  | same symlink | `@../../.trellis/rules.toml` | **NONE** |

- **A `.toml` placed directly in `.claude/rules/` is never loaded.** Only `.md`
  is discovered. Measured: the sentinel was absent across both runs above.

The wrong import form produces **no error and no content**. That is the third
silent-no-op this format has produced during this work and it must be pinned by
a test rather than by care.

## Decision

**1. The install path delivers rules by rendering one file and linking to it.**

```
.trellis/trellis.md   rendered: posture prose + rules body + `@rules.toml`
.trellis/rules.toml   the project's rows — seeded if absent, never overwritten
.claude/rules/trellis.md → symlink to ../../.trellis/trellis.md
```

`install.sh` renders one file and makes one link. It registers no hook and edits
no settings file.

**2. `.trellis/` is the home; `.claude/rules/` is only the discovery point.**

No harness lets a project nominate its own always-on directory — every one fixes
its own location, and files in arbitrary subdirectories load lazily or not at
all. Keeping the real file in `.trellis/` means one file to render and to keep
current, with a thin pointer per harness when others are added (#209).

**3. The rows stay live. The rendered file is not a snapshot.**

Because `@rules.toml` is imported rather than inlined, editing `.trellis/rules.toml`
takes effect in the next session with no re-run of the installer. This matches
the plugin path's behaviour, which re-splices the rows every session.

**4. The import form is load-bearing and is pinned by a test.**

The rendered file carries `@rules.toml` — a sibling reference, valid because the
symlink resolves to `.trellis/`. A test asserts the emitted form, and asserts
that the `../../` form is *not* emitted. Without it, a future reword ships a file
that loads nothing and passes every other check.

**5. Claude only, deliberately, and stated rather than implied.**

`.claude/rules/` is a Claude Code mechanism. Codex, OpenCode, Devin and VS Code
have no equivalent that inlines a non-markdown file; the research is recorded in
**#209**. The install path therefore claims **Claude Code only**, and the
existing prose authority sentence at `reference/rules.md:1` remains the fallback
for the rows anywhere that cannot import them.

**6. Platform boundary: macOS, Linux, WSL — recorded, not discovered later.**

`install.sh` is `#!/bin/sh` with POSIX dependencies, so Windows was already
outside the install path before symlinks arose. WSL is unaffected — symlinks are
native there. Only the Git Bash slice narrows further, and **a copy fallback is
not a drop-in**: a copied file cannot use `@rules.toml`, because the import
resolves from the real file's directory and a copy's differs. It would have to
inline the rows and accept the snapshot. Tracked in **#210**.

**7. The plugin hook stands down when the install artifact is present. This is
required, not optional — the alternative was measured and it double-delivers.**

With the user-scope plugin installed *and* `.claude/rules/trellis.md` present,
the rules arrive **twice**: once in the project-instructions block, once in the
hook's `additionalContext`. Measured, not predicted.

`staleness.sh` already has this shape — it detects `.trellis/internal/` and
declines to inject over a vendored overlay, so that "a project that vendored the
overlay never receives the rules twice". The install path needs the same branch
keyed on its own artifact: **if `.trellis/trellis.md` exists, the hook injects
nothing** and says so, exactly as it does for the overlay today.

This widens the change beyond the installer, and that is stated rather than
discovered at implementation: **`decision-0065`'s hook gains a third path.**

**8. Out of scope, named rather than silently retained: whether the bundle
vendoring stays.** `install.sh` also vendors the whole `plugins/trellis/` tree so
`/trellis:setup` and `/trellis:remove` have a home (`spec-0005` §1). Whether
those skills actually load from that location is **unmeasured** — this record
measured rule delivery only, and does not touch the bundle.

## Consequences

- **A path that shipped nothing now ships something.** That is the whole value;
  everything else is simplification.
- `install.sh` no longer needs to reason about hooks or host manifests for rule
  delivery. The hook files remain bundle bytes under `spec-0005` §1 until D8 is
  answered — and D7 keeps the plugin's copy of them from firing here.
- **The install path vendors, and that is correct here.** `decision-0065`'s split
  is "plugin configures, install.sh vendors". Staleness is inherent to vendoring
  and is what the version stamp exists for. It was only the *plugin* path where
  writing into a consumer repo was the error.
- A committed symlink is a new artifact class for this repo. It is portable on
  every platform the install path claims.
- **Cost:** two delivery mechanisms now exist for Claude — hook injection via the
  plugin, file discovery via the install. D7 keeps them mutually exclusive; without
  it a repo with both receives the rules twice, which is measured, not feared.
  `.trellis/rules.toml` cannot be the discriminator because both paths require it,
  so the install artifact itself is the signal.

## Open questions

1. **Do the vendored skills load at all?** D8 defers it; the same trust-dialog
   question that defeated hook registration may defeat them.
3. **Should `/trellis:remove` remove the symlink and the rendered file?** It
   deletes `.trellis/` today; the link would dangle.

## Self-check (gate)

| # | check | result |
|---|---|---|
| 1 | The central claim is measured, not read from docs | **PASS** — four runs, tabulated, with the user-scope confound named |
| 2 | A wrong reading is recorded rather than quietly corrected | **PASS** — run 1 and the control that overturned it |
| 3 | Both import forms tested, not just the working one | **PASS** — the failing form is recorded because it fails silently |
| 4 | Non-Claude harnesses checked before claiming Claude-only | **PASS** — #209, four harnesses, sources cited there |
| 5 | The platform boundary is stated with its real cause | **PASS** — POSIX `sh`, not the symlink; #210 |
| 6 | Scope creep resisted and named | **PASS** — D8 leaves the bundle untouched and says so |
| 7 | Acceptance criteria | **DEFERRED to the spec amendment**, which this record pairs with |
| 8 | Double-delivery risk surfaced | **PASS** — measured rather than deferred: both paths live delivers the rules twice, so it became D7 (the hook stands down) instead of an open question. Found because the risk was written down and then tested |
| 9 | `status: gated` earned | **PASS** — self-check run; no partials remain, because the one that existed was resolved by measurement rather than downgraded |

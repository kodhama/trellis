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

**1. The install path delivers rules by rendering exactly one file.**

```
.claude/rules/trellis.md   rendered: posture prose + rules body + `@../../.trellis/rules.toml`
```

`install.sh` writes that file and nothing else new. It registers no hook, edits
no settings file, creates no symlink, and **never touches `.trellis/`**.

**2. `install.sh` owns `.claude/`; `/trellis:setup` owns `.trellis/`.**

An earlier draft of this record put the real file at `.trellis/trellis.md` and
symlinked it into `.claude/rules/`. **Withdrawn**, for three reasons:

- **Windows.** A symlink narrows the Git Bash slice on top of the POSIX-`sh`
  barrier, and a copy fallback there could not use the sibling import form —
  it would need a different one and would fail silently if got wrong. Without
  the link, #210 is purely about the shell again.
- **Ownership.** `/trellis:setup` writes `.trellis/rules.toml` and nothing else,
  ever. A symlink would place an install-owned object inside setup's directory.
- **Count.** One filesystem object rather than two, for the same delivered bytes.

The cost is a longer import path, and it is cosmetic: **whichever form is
correct, the other is a silent no-op**, so the test in D4 is mandatory either
way. Once it exists, `../../` being uglier costs nothing.

**3. The rows stay live. The rendered file is not a snapshot.**

Because `@rules.toml` is imported rather than inlined, editing `.trellis/rules.toml`
takes effect in the next session with no re-run of the installer. This matches
the plugin path's behaviour, which re-splices the rows every session.

**4. The import form is load-bearing and is pinned by a test.**

The rendered file carries `@../../.trellis/rules.toml`, correct because the file
is real and lives at `.claude/rules/`. A test asserts that exact emitted form and
asserts the sibling form `@rules.toml` is *not* emitted. Both were measured; each
is correct in one location and silently loads nothing in the other. Without the
test a future reword ships a file that loads nothing and passes every other
check.

**5. Claude only, deliberately, and stated rather than implied.**

`.claude/rules/` is a Claude Code mechanism. Codex, OpenCode, Devin and VS Code
have no equivalent that inlines a non-markdown file; the research is recorded in
**#209**. The install path therefore claims **Claude Code only**, and the
existing prose authority sentence at `reference/rules.md:1` remains the fallback
for the rows anywhere that cannot import them.

**6. Platform boundary: macOS, Linux, WSL — and this record adds nothing to it.**

`install.sh` is `#!/bin/sh` with POSIX dependencies, so Windows is already
outside the install path. Because D2 drops the symlink, **this change narrows
nothing further** — #210 stays purely about the shell, and a future PowerShell or
Go installer inherits no symlink problem.

**7. With no `.trellis/rules.toml`, the install is inert except the floors — and
that needs no seeding.**

`install.sh` makes no posture choice and writes no `rules.toml`; that is
`/trellis:setup`'s single file. The rules text already defines this state: the
two `floor-` rows *"apply regardless of their row value"*, and every other rule
applies **only** where its row says `active = true`. So a bare install yields
exactly `floor-transparency` and `floor-intent-gate` — a safe, well-defined
default that requires no seed and no decision from the script. The rendered file
directs the reader to `/trellis:setup` to activate the rest.

This is why `decision-0065`'s clause that setup writes "exactly one file …
and nothing else, ever" **needs no amendment**: this record does not widen it.

**8. The plugin hook stands down when the install artifact is present. This is
required, not optional — the alternative was measured and it double-delivers.**

With the user-scope plugin installed *and* `.claude/rules/trellis.md` present,
the rules arrive **twice**: once in the project-instructions block, once in the
hook's `additionalContext`. Measured, not predicted.

`staleness.sh` already has this shape — it detects `.trellis/internal/` and
declines to inject over a vendored overlay, so that "a project that vendored the
overlay never receives the rules twice". The install path needs the same branch
keyed on its own artifact: **if `.claude/rules/trellis.md` exists, the hook
injects nothing** and says so, exactly as it does for the overlay today.

An earlier draft claimed this came free by reusing the `.trellis/internal/`
path-A discriminator. **That was wrong** — it held only while invariants were
also placed there, which this record no longer does. Path A never fires under
this design, so the branch is new code, roughly three lines, and explicit.

This widens the change beyond the installer, and that is stated rather than
discovered at implementation: **`decision-0065`'s hook gains a third path.**

**9. Out of scope, named rather than silently retained: whether the bundle
vendoring stays.** `install.sh` also vendors the whole `plugins/trellis/` tree so
`/trellis:setup` and `/trellis:remove` have a home (`spec-0005` §1). Whether
those skills actually load from that location is **unmeasured** — this record
measured rule delivery only, and does not touch the bundle.

## Consequences

- **A path that shipped nothing now ships something.** That is the whole value;
  everything else is simplification.
- `install.sh` no longer needs to reason about hooks or host manifests for rule
  delivery. The hook files remain bundle bytes under `spec-0005` §1 until D9 is
  answered — and D8 keeps the plugin's copy of them from firing here.
- **The install path vendors, and that is correct here.** `decision-0065`'s split
  is "plugin configures, install.sh vendors". Staleness is inherent to vendoring
  and is what the version stamp exists for. It was only the *plugin* path where
  writing into a consumer repo was the error.
- **No new artifact class.** An earlier draft introduced a committed symlink; D2
  withdrew it, so the install path writes only ordinary files.
- **Invariants stay a vendored file** at `.claude/skills/trellis/reference/invariants.md`,
  and the rendered pointer names that path. That is already better than today,
  where the shipped prose names `.trellis/internal/invariants.md` and nothing on
  the install path creates it. `decision-0067` replaces the pointer with a skill
  on **both** paths afterwards; between the two records the install path is
  correct and the plugin path keeps its runtime substitution.
- **Cost:** two delivery mechanisms now exist for Claude — hook injection via the
  plugin, file discovery via the install. D8 keeps them mutually exclusive; without
  it a repo with both receives the rules twice, which is measured, not feared.
  `.trellis/rules.toml` cannot be the discriminator because both paths require it,
  so the install artifact itself is the signal.

## Open questions

1. **Do the vendored skills load at all?** D9 defers it; the same trust-dialog
   question that defeated hook registration may defeat them.
3. **Should `/trellis:remove` delete `.claude/rules/trellis.md`?** It removes
   `.trellis/` today, which would leave the rendered file importing a path that
   no longer exists — inert, but present.

## Self-check (gate)

| # | check | result |
|---|---|---|
| 1 | The central claim is measured, not read from docs | **PASS** — four runs, tabulated, with the user-scope confound named |
| 2 | A wrong reading is recorded rather than quietly corrected | **PASS** — run 1 and the control that overturned it |
| 3 | Both import forms tested, not just the working one | **PASS** — the failing form is recorded because it fails silently |
| 4 | Non-Claude harnesses checked before claiming Claude-only | **PASS** — #209, four harnesses, sources cited there |
| 5 | The platform boundary is stated with its real cause | **PASS** — POSIX `sh`, not the symlink; #210 |
| 6 | Scope creep resisted and named | **PASS** — D9 leaves the bundle untouched and says so |
| 7 | Acceptance criteria | **DEFERRED to the spec amendment**, which this record pairs with |
| 8 | Double-delivery risk surfaced | **PASS** — measured rather than deferred: both paths live delivers the rules twice, so it became D8 (the hook stands down) instead of an open question. Found because the risk was written down and then tested |
| 9 | A withdrawn design is struck, not deleted | **PASS** — D2 records the symlink and the three reasons it lost; D8 records that its "free mutual exclusion" claim was conditional and wrong |
| 10 | `status: gated` earned | **PASS** — self-check run; no partials remain, because the one that existed was resolved by measurement rather than downgraded |

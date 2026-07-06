---
name: setup
description: Install Trellis governance onto this project — pick a posture and compose the invariants alongside the project's existing instructions (the M1 overlay, augment-never-clobber). Use when the user asks to set up, add, install, or apply Trellis in their repo.
---

# Set up Trellis in this project

You are composing the **Trellis** governance layer onto the user's project as the **M1 "alongside"
overlay**: add a small bundle plus one import line, and **never touch anything else** (augment,
never clobber). This mirrors `trellis setup --mode m1` but needs no binary — you do it directly.

## 1. Pick a posture

Ask which posture fits; default to **B · author-adapt** if the user is unsure.

- **A · conductor** — adopt strictly. All invariants active at high enforcement (`enforced`).
- **B · author-adapt** — evolve as you go. All active, lighter (`default-on-but-skippable`), with
  self-improvement kept high.
- **seed** — start minimal: only the structural core (`inv-directional-flow`, `inv-handover-points`,
  `inv-ratifiable-artifacts`) plus the two floors (`floor-transparency`, `floor-intent-gate`),
  enforcement `expressed`, ratcheting up over time.

## 2. Write the `.trellis/` bundle (create the files; never overwrite the user's own)

- **`.trellis/invariants.md`** — copy it from `${CLAUDE_PLUGIN_ROOT}/reference/invariants.md` (the full
  invariant reference; the user reads it on demand).
- **`.trellis/profile.md`** — the readout **plus the active rules** (this file is auto-loaded, so it
  must carry the rules, not just names — `decision-0026`): the chosen posture, the enforcement lean, the
  line *"gatekeeper: detected from this project, not preset,"* and then **each active invariant as a
  concise rule** — its one-line description taken from the `what:` field of that invariant in
  `${CLAUDE_PLUGIN_ROOT}/reference/invariants.md`. The agent must always see what each active invariant
  *requires*, not just its name.
- **`.trellis/version`** — the staleness stamp (`decision-0035`/`decision-0039`): one line naming
  what generated this overlay. Run `git -C "${CLAUDE_PLUGIN_ROOT}" rev-parse --short HEAD` and write
  `plugin@<sha>` (plugin versions are commit SHAs, `decision-0036`). If that fails (not a git
  checkout), write `plugin@unknown` — an honest stamp beats none; an unstamped overlay is invisible
  to every staleness check (`trellis status` and any future update surface).
- **`.trellis/trellis.md`** — the header the project's agents read. It must contain, in this order:
  1. a one-line intro (this project is governed by Trellis);
  2. **the key behavior**, verbatim in spirit: *"Surface any human-gated handover performed without
     its human approval (`inv-gate-at-handover`). Agent-gated handovers proceed silently. Gatekeepers are
     whatever this project already declares — respected, not imposed."*
  3. `@profile.md` (imports the profile);
  4. a pointer: *"The active rules are in the profile above (always loaded). The full why + with/without
     examples for each — and the invariants not active here — are in `.trellis/invariants.md`; read it
     for the detail behind a rule."* (backtick the path so it is **not** auto-imported).

## 3. Compose the import into `CLAUDE.md` (augment, never clobber)

Append this block to the project's `CLAUDE.md` (create the file if it does not exist). Touch **nothing
else** in the file:

```
<!-- trellis:begin (managed by trellis — edit .trellis/, not this block) -->
This project is governed by **Trellis** (see the `.trellis/` folder). Its rules are imported here:
@.trellis/trellis.md
<!-- trellis:end -->
```

If a `trellis:begin … trellis:end` block already exists, **replace only what is between the markers**
(idempotent — never a second block, never a change outside the markers).

## 4. Confirm

Tell the user exactly what you wrote (`.trellis/{trellis,profile,invariants}.md`, the
`.trellis/version` stamp, and the `CLAUDE.md` import block), and that they can remove it any time by
deleting `.trellis/` and the block — or with `trellis remove` if they have the CLI. Do **not** attempt a morph (M2) here; that is the CLI's job.

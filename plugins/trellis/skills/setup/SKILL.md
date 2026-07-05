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
- **`.trellis/profile.md`** — the readout: the chosen posture, the active invariants, the enforcement
  lean, and the line: *"gatekeeper (C2): detected from this project, not preset."*
- **`.trellis/trellis.md`** — the header the project's agents read. It must contain, in this order:
  1. a one-line intro (this project is governed by Trellis);
  2. **the key behavior**, verbatim in spirit: *"Surface any human-gated handover performed without
     its human approval (invariant B2). Agent-gated handovers proceed silently. Gatekeepers are
     whatever this project already declares — respected, not imposed."*
  3. `@profile.md` (imports the profile);
  4. a pointer: *"Full invariant reference: `.trellis/invariants.md` — read it when you need the
     detail behind a rule."* (in backticks, so it is not auto-imported).

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

Tell the user exactly what you wrote (`.trellis/{trellis,profile,invariants}.md` and the `CLAUDE.md`
import block), and that they can remove it any time by deleting `.trellis/` and the block — or with
`trellis remove` if they have the CLI. Do **not** attempt a morph (M2) here; that is the CLI's job.

---
id: spec-0004
type: spec
status: ratified
ratified: 2026-07-05
depends_on: [spec-0003, decision-0023, invariants-v1]
owner: gundi
rubric: spec-quality
---

# Spec 0004 — Clean exits: uninstall (the CLI) and remove (from a project)

> **Every install has a reverse.** `install.sh` puts the binary down; `trellis setup` composes Trellis
> onto a project. This spec adds the two symmetric exits — **uninstall** the binary and **remove**
> Trellis from a project — and is honest about the one case (a morph) it cannot cleanly reverse.

## Purpose

Give a user a trustworthy way out. A governance tool that is hard to remove is not trustworthy; a
clean, predictable exit is part of the product. Draft; ratified at merge (`decision-0022`).

## Scope

**In:** `trellis uninstall` (remove the installed binary) and `trellis remove` (undo `setup` on a
project) for both install modes (M1 overlay, M2 morph). **Out:** removing supervisor-mode hooks — that
mode isn't built yet.

## 1. `trellis uninstall` — remove the binary

- Removes the running binary at its own path (`os.Executable`) after a confirmation (or `--yes`), and
  prints the path removed. On Unix, unlinking a running binary is safe (the inode frees on exit).
- **Homebrew-managed installs are deferred, not deleted (`decision-0032`).** If the binary resolves
  under a Homebrew `Cellar` (directly or via the `bin` symlink), `uninstall` does **not** remove it —
  that would leave brew's records inconsistent — and instead points at `brew uninstall trellis`.
- `install.sh` gains a mirrored `--uninstall` for parity with the curl install.
- It removes **only the binary** — never a project's `.trellis/` (that's `remove`'s job); it says so.

## 2. `trellis remove [--dir]` — undo setup on a project

Detect what `setup` left, and reverse it **per mode**.

- **M1 (overlay) — deterministic, clean.** Delete the `.trellis/` directory and strip the delimited
  Trellis block from **whichever instruction file(s) setup attached to** — `CLAUDE.md`, `AGENTS.md`,
  `GEMINI.md`, … (`decision-0029`); the block markers are the same in each — leaving everything else
  **byte-for-byte as it was** — augment-never-clobber, in reverse. Idempotent; if nothing is there,
  say so and no-op.
- **M2 (morph) — git rollback, warned.** A morph rewrote the project's **own** files; Trellis cannot
  cleanly reverse that. So `remove` **does not mutate history**: it prints a **loud destructive
  warning**, reports the **rollback commit reference saved at apply time**, and gives the git command
  to roll back (`git reset --hard <ref>` on the morph branch, or `git revert`). It states plainly that
  this is the limit of what it can do — the honest floor (D1), not a silent best-effort.

To make M2 reversible-by-reference, **`setup --mode m2` records the pre-morph commit before the
rewrite** — a `.trellis/rollback` note (the SHA + the git command) and, git-native, a `trellis-pre-morph`
tag. `remove` reads it; if it is missing, `remove` says it cannot locate a rollback point rather than
guess.

## 3. Honest limits (D1)

`remove` never deletes host content **outside** the managed block, never auto-rewrites git history, and
**surfaces exactly what it did — or couldn't**. A no-op is reported, not silent.

## Acceptance criteria

- **AC1 — uninstall.** `trellis uninstall` removes the binary (after confirm / `--yes`) and prints the
  path; a second run reports it is already gone. It does not touch any project's `.trellis/`.
- **AC2 — M1 remove.** `.trellis/` is deleted and the `CLAUDE.md` block stripped; **all other
  `CLAUDE.md` content is byte-preserved**; a re-run is a clean no-op.
- **AC3 — M2 remove.** Never auto-reverses; prints a **valid rollback ref + the git command + a loud
  warning**; if no rollback marker exists, it says so instead of guessing.
- **AC4 — detection.** `remove` picks the path from what is present (bundle + block → M1; rollback
  marker → M2; neither → a clear no-op message). A project that had both is handled marker-by-marker.
- **AC5 — reverse augment-never-clobber.** Stripping the block preserves everything outside the markers
  exactly; if `CLAUDE.md` is left empty, it is removed only if `setup` created it (else left in place).

## Open questions

- **Confirmation UX** — interactive prompt vs `--yes`; default to interactive, `--yes` for scripts.
- **`.trellis/rollback` note vs the git tag alone** — the tag is git-native and survives a moved
  `.trellis/`; the note is a human breadcrumb. Ship both; pick the authoritative one in the build.
- **Provenance of the install mode** — M1 is self-evident from the bundle; M2's marker is the rollback
  record. A hand-edited project with neither is treated as "nothing to remove."

---
name: remove
description: Remove Trellis from this project — delete the .trellis/ overlay, strip the managed block from CLAUDE.md, and strip any .trellis/ lint-ignore entry setup added, touching nothing else. Use when the user asks to remove, uninstall, undo, or take out Trellis from their repo.
---

# Remove Trellis from this project

Cleanly reverse the shared **M1 overlay** from either host. This is a **product-wide** remove, not a
per-host disable: it handles both the Claude block in `CLAUDE.md` and the Codex receipt/fallback in
`AGENTS.md`, then removes `.trellis/`. Preserve every surrounding user byte.

## 1. Preflight every target before any edit

Snapshot the complete project paths this operation may touch. Inspect **every documented instruction file**
— `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.github/copilot-instructions.md`, and
`.clinerules` — for **all recognized managed blocks**:

- `<!-- trellis:begin … -->` through `<!-- trellis:end -->`, including import and
  legacy/manual inline/full-rule forms; and
- `<!-- trellis:codex-bootstrap:begin … -->` through
  `<!-- trellis:codex-bootstrap:end -->`.

Also inspect, without writing:

- `.trellis/`, including consumer-owned `rules.toml` and any legacy `expression.md`;
- every recognized lint/format ignore target that may contain a setup-added `.trellis/` line.

For each marker family, require either no marker or exactly one nonnested paired region. Duplicate,
unpaired, nested, overlapping, or otherwise ambiguous markers stop the entire operation **before**
any block or overlay change. Resolve every required consent before writing: in particular, surface
hand-written `expression.md` content and any ignore entry whose ownership is ambiguous. If consent
is unavailable, stop with the whole-project snapshot unchanged.

## 2. Stage byte-safe instruction-file removals

Prepare the resulting bytes for all five documented instruction files before changing any one:

- Remove only each recognized managed region, including the one separator newline setup added.
- Preserve all bytes before and after the region exactly.
- Delete an instruction file only when it becomes empty because Trellis created it; otherwise keep
  it, even when the remainder is whitespace.

Do not treat an absent block as an error. A recognized, valid block is removed wherever setup or a
documented manual path placed it; ambiguous placement or markers were already a preflight failure.

## 3. Stage consented ignore cleanup

For ESLint, Prettier, Biome, and markdownlint targets detected by setup, remove only a `.trellis/`
entry known to have been added by Trellis. If an ignore file created by setup then becomes empty,
it may be removed. Preserve all other patterns byte-for-byte. An ambiguous entry requires consent
in step 1; never guess.

## 4. Apply the complete product-wide transaction

Only after every preflight and consent succeeds:

1. write or delete every staged documented instruction-file result;
2. apply the staged, consented ignore cleanup; and
3. delete the shared `.trellis/` overlay last.

Verify surrounding instruction-file and ignore-file bytes against the snapshots. If a preflight
failed, verify that every block and the overlay remain unchanged.

## 5. Confirm

Report every recognized item as removed, retained, ambiguous, or absent: the Claude block, Codex
bootstrap, shared overlay, legacy consumer content, and ignore entries. If no managed block or
overlay was present, make no change and say Trellis is **already absent**. A second remove is this
same reported no-op.

## Reversing an M2 morph

If this project was changed by the **M2 morph** (`/trellis:setup`'s model-driven rewrite of the
project's own files, on the `trellis/morph` branch), there is no overlay to strip — the reversal is
**git's**, using the rollback point the morph recorded: the `trellis-pre-morph` tag, or the SHA in
`.trellis/rollback`. Show the user the options (`git reset --hard trellis-pre-morph`, `git revert`,
or deleting the unmerged `trellis/morph` branch) and let them run the destructive step — never
attempt to reverse a morph by editing files back by hand.

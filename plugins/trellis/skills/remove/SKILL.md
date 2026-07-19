---
name: remove
description: Remove Trellis from this project — delete the .trellis/ overlay, strip the managed block from CLAUDE.md, and strip any .trellis/ lint-ignore entry setup added, touching nothing else. Use when the user asks to remove, uninstall, undo, or take out Trellis from their repo.
---

# Remove Trellis from this project

Cleanly reverse the **M1 overlay** that `/trellis:setup` (or the manual copy path, or the retired
CLI's `setup --mode m1`) installed. This is augment-never-clobber **in reverse**: remove *only* what
Trellis added, and preserve everything else byte-for-byte.

## 1. Delete the bundle

Delete the entire `.trellis/` directory, if it exists — the generated `internal/` half and the
consumer-owned root (`rules.toml`, `expression.md`) alike, and any legacy flat-layout files from an
overlay installed before `decision-0051`. If `expression.md` carries a hand-written body the user
may want to keep, say so before deleting — it is theirs, and this is the one step that destroys it.

## 2. Strip the import from `CLAUDE.md`

In the project's `CLAUDE.md`, remove the managed block **between and including** these markers:

```
<!-- trellis:begin … -->
   … (the @import lines) …
<!-- trellis:end -->
```

Remove **only** that block (and the single blank line that preceded it). **Touch nothing else** —
every other line of the user's `CLAUDE.md` must stay exactly as it was. If, after removal, `CLAUDE.md`
contains only whitespace (it held nothing but the Trellis block, i.e. Trellis created the file), delete
it; otherwise leave it in place.

## 3. Strip any `.trellis/` ignore entry setup added (`decision-0049`)

`/trellis:setup` may have offered to add `.trellis/` to the project's linters/formatters (setup's
step 9). Reverse that too — augment-never-clobber, the same as the `CLAUDE.md` block above:

- **Detect the same tools setup did** — ESLint (`.eslintrc*` / `eslint.config.*` / `eslintConfig` in
  `package.json`), Prettier (`.prettierrc*` / `.prettierignore`), Biome (`biome.json`), markdownlint
  (`.markdownlint*` / `.markdownlintignore`) — and look for a `.trellis/` entry in each one's ignore.
- **Remove only the `.trellis/` entry, and only that line** — touch no other ignore pattern. If an
  ignore file that setup *created* now holds nothing but that entry (it exists only because of
  Trellis), delete the file; otherwise leave it, minus the one line — the same rule as the `CLAUDE.md`
  block.
- **Only what Trellis added.** You cannot always be certain the user did not add `.trellis/`
  themselves. If it is ambiguous — the entry sits among the user's own patterns, or the file is
  clearly theirs — **surface it and ask** before removing; never strip a line you are not sure
  Trellis put there.
- If no `.trellis/` ignore entry is found, do nothing here and say so — **do not invent changes**.

## 4. Confirm

Tell the user exactly what you removed (`.trellis/`, the `CLAUDE.md` block, and any `.trellis/`
lint-ignore entry). If none was present, say so plainly — **do not invent changes**.

## Reversing an M2 morph

If this project was changed by the **M2 morph** (`/trellis:setup`'s model-driven rewrite of the
project's own files, on the `trellis/morph` branch), there is no overlay to strip — the reversal is
**git's**, using the rollback point the morph recorded: the `trellis-pre-morph` tag, or the SHA in
`.trellis/rollback`. Show the user the options (`git reset --hard trellis-pre-morph`, `git revert`,
or deleting the unmerged `trellis/morph` branch) and let them run the destructive step — never
attempt to reverse a morph by editing files back by hand.

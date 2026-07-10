---
name: remove
description: Remove Trellis from this project — delete the .trellis/ overlay and strip the managed block from CLAUDE.md, touching nothing else. Use when the user asks to remove, uninstall, undo, or take out Trellis from their repo.
---

# Remove Trellis from this project

Cleanly reverse the **M1 overlay** that `/trellis:setup` (or the manual copy path, or the retired
CLI's `setup --mode m1`) installed. This is augment-never-clobber **in reverse**: remove *only* what
Trellis added, and preserve everything else byte-for-byte.

## 1. Delete the bundle

Delete the entire `.trellis/` directory, if it exists.

## 2. Strip the import from `CLAUDE.md`

In the project's `CLAUDE.md`, remove the managed block **between and including** these markers:

```
<!-- trellis:begin … -->
   … (the @import line) …
<!-- trellis:end -->
```

Remove **only** that block (and the single blank line that preceded it). **Touch nothing else** —
every other line of the user's `CLAUDE.md` must stay exactly as it was. If, after removal, `CLAUDE.md`
contains only whitespace (it held nothing but the Trellis block, i.e. Trellis created the file), delete
it; otherwise leave it in place.

## 3. Confirm

Tell the user exactly what you removed (`.trellis/` and the `CLAUDE.md` block). If neither was present,
say so plainly — **do not invent changes**.

## Reversing an M2 morph

If this project was changed by the **M2 morph** (`/trellis:setup`'s model-driven rewrite of the
project's own files, on the `trellis/morph` branch), there is no overlay to strip — the reversal is
**git's**, using the rollback point the morph recorded: the `trellis-pre-morph` tag, or the SHA in
`.trellis/rollback`. Show the user the options (`git reset --hard trellis-pre-morph`, `git revert`,
or deleting the unmerged `trellis/morph` branch) and let them run the destructive step — never
attempt to reverse a morph by editing files back by hand.

---
name: remove
description: Remove Trellis from this project — delete the .trellis/ overlay, strip the managed block from CLAUDE.md, and strip any .trellis/ lint-ignore entry setup added, touching nothing else. Use when the user asks to remove, uninstall, undo, or take out Trellis from their repo.
---

# Remove Trellis from this project

Cleanly reverse the shared **M1 overlay** from either host. This is a **product-wide** remove, not a
per-host disable: it handles both the Claude block in `CLAUDE.md` and the Codex receipt/fallback in
`AGENTS.md`, removes the curl install path's rendered
`.claude/rules/trellis.md`, and **then** removes `.trellis/`. Preserve every
surrounding user byte.

## 1. Preflight every target before any edit

Snapshot the complete project paths this operation may touch. Inspect **every documented instruction file**
— `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.github/copilot-instructions.md`, and
`.clinerules` — for **all recognized managed blocks**:

- `<!-- trellis:begin … -->` through `<!-- trellis:end -->`, including import and
  legacy/manual inline/full-rule forms; and
- `<!-- trellis:codex-bootstrap:begin … -->` through
  `<!-- trellis:codex-bootstrap:end -->`.

Also inspect, without writing:

- `.claude/rules/trellis.md` — rendered by `install.sh` on the curl install path
  (`decision-0068` D1). It is a Trellis-authored instructions file loaded into
  every session, and it imports `.trellis/rules.toml`;
- `.trellis/`, including consumer-owned `rules.toml` and any legacy `expression.md`;
- every recognized lint/format ignore target that may contain a setup-added `.trellis/` line.

**Order matters between those two.** Remove `.claude/rules/trellis.md` *before*
`.trellis/`. An interrupted removal must never leave the governing file present
with its rows already gone: a posture header and a full rules body in
always-loaded context, importing a file that no longer exists, with nothing left
to activate any rule. That is a worse state than either end of the operation.

Remove only `.claude/rules/trellis.md`. `.claude/rules/` is a shared directory a
project may fill with unrelated rules of its own, and the directory itself is not
Trellis's to delete — the same file-not-directory boundary the hook's path C
uses.

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

1. delete the rendered `.claude/rules/trellis.md` if present;
2. write or delete every staged documented instruction-file result;
3. apply the staged, consented ignore cleanup; and
4. delete the shared `.trellis/` overlay last.

**Why the rendered file goes FIRST, not third.** A project mid-migration can hold
both delivery paths at once — a managed block plus `.trellis/internal/`, and a
rendered file. Removing the block (step 2) before the rendered file leaves an
interruption window in which *neither* path governs: the block is gone, the
rendered file is gone, and the surviving `.trellis/internal/` makes the plugin
hook's path A exit without injecting. The session that follows is ungoverned
while the removal looks half-done in both directions.

Ordering it first inverts that: every interruption window leaves **at least one**
delivery path intact, and the last thing removed is the overlay, which is what
the guarantee at the end of this section already promises. An interrupted removal
may leave the rendered file gone and the block still importing rows — governed,
if redundantly — which is the safe direction.

Delete only the file. `.claude/rules/` is a shared directory a project may fill
with unrelated rules of its own, and the directory itself is not Trellis's to
remove.

Verify surrounding instruction-file and ignore-file bytes against the snapshots. If a preflight
failed, verify that every block and the overlay remain unchanged.

## 5. Confirm

Report every recognized item as removed, retained, ambiguous, or absent: the Claude block, Codex
bootstrap, shared overlay, the rendered `.claude/rules/trellis.md`, legacy consumer content, and
ignore entries.

**The no-op predicate counts all three installed shapes, not two.** Say Trellis is
**already absent** — and make no change — only when there is no managed block, no overlay, **and no
`.claude/rules/trellis.md`**. A curl install has the rendered
file and neither of the other two: reporting that project "already absent" would leave an
always-loaded governing file on disk while telling the user Trellis was gone. A second remove is
this same reported no-op, once all three are genuinely absent.

## Reversing an M2 morph

If this project was changed by the **M2 morph** (the retired setup skill's model-driven rewrite of the
project's own files, on the `trellis/morph` branch), there is no overlay to strip — the reversal is
**git's**, using the rollback point the morph recorded: the `trellis-pre-morph` tag, or the SHA in
`.trellis/rollback`. Show the user the options (`git reset --hard trellis-pre-morph`, `git revert`,
or deleting the unmerged `trellis/morph` branch) and let them run the destructive step — never
attempt to reverse a morph by editing files back by hand.

---
name: remove
description: Remove Trellis from this project across every delivery state — managed blocks in the five documented instruction files, the vendored .trellis/ overlay and config, the rendered .claude/rules/trellis.md, the project-scope plugin bundle at .claude/skills/trellis/, and Trellis-added lint-ignore entries — and point a morphed project at its git rollback. Use when the user asks to remove, uninstall, undo, or take out Trellis from their repo.
---

# Remove Trellis from this project

Cleanly reverse Trellis from **every delivery state in `decision-0073` D1's closed set** (S0–S6).
This is a **product-wide** remove, not a per-host disable: it handles the managed blocks in the
documented instruction files, the vendored overlay (internal and legacy flat), the curl install
path's rendered `.claude/rules/trellis.md`, the project-scope plugin bundle, and **then**
removes `.trellis/`. Preserve every surrounding user byte. No state below can be assumed to arrive alone —
a fresh curl install, for example, carries **two** shapes at once: the rendered file *and* the
seeded `.trellis/rules.toml` (`decision-0070` D2).

## 1. Preflight every target before any edit

Snapshot the complete project paths this operation may touch. Inspect **every documented instruction file**
— `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.github/copilot-instructions.md`, and
`.clinerules` — for **all recognized managed blocks**:

- `<!-- trellis:begin … -->` through `<!-- trellis:end -->`, including import and
  legacy/manual inline/full-rule forms; and
- `<!-- trellis:codex-bootstrap:begin … -->` through
  `<!-- trellis:codex-bootstrap:end -->`.

Marker cardinality is **per file, per family**: in each instruction file, require either no marker
or exactly one nonnested paired region of each family. A block in `CLAUDE.md` *and* another in
`AGENTS.md` is a legitimate multi-file state — remove each; it is not a duplicate. Duplicate,
unpaired, nested, overlapping, or otherwise ambiguous markers **within one file** stop the entire
operation **before** any block or overlay change.

Also inspect, without writing:

- `.claude/rules/trellis.md` — rendered by `install.sh` on the curl install path
  (`decision-0068` D1). It is a Trellis-authored instructions file loaded into
  every session, and it imports `.trellis/rules.toml`;
- `.trellis/`, including consumer-owned `rules.toml` and any legacy `expression.md`. Two things
  inside it need surfacing **by name** before any deletion:
  - **the decline artifact** — a `rules.toml` carrying top-level `governed = false` is the
    project's recorded decline (`decision-0070` D4). State the consequence out loud: on a machine
    with a user-scope plugin, deleting the recorded decline re-arms the adoption announcement, and
    one ignored prompt re-governs at 14/14. Deleting it is legitimate — removal is removal — but
    doing it unnamed is not;
  - **the consumer's own rows** — `rules.toml` is the consumer's file, and deleting `.trellis/`
    deletes every row they edited. Show them the rows that will go, as part of the consent below;
- **the vendored plugin bundle** at `.claude/skills/trellis/` — the project-scope **adoption act**
  (`decision-0070` D3), **except when the project root is `$HOME`**: that path is then the
  user-scope install location (`decision-0070` D6), not this project's bundle, and it is not this
  operation's to touch;
- **the morph markers, before any write**: `.trellis/rollback` and the `trellis-pre-morph` git
  tag. Detection belongs here in preflight — step 5 of the transaction deletes `.trellis/`
  wholesale, which destroys `.trellis/rollback`, so a morph discovered only after the transaction
  reads a rollback pointer this operation already destroyed. If either marker is present, read the
  morph section at the end **now**, before staging anything;
- every recognized lint/format ignore target that may contain a `.trellis/` line added by a
  past setup (the skill that wrote these is retired; the artifacts it left are not).

Provenance cannot be proven from disk alone: no receipt exists for who added an ignore entry or an
unmarked line. Where a predicate would have to claim provenance to proceed, treat the item as
**ambiguous and consent-gated** — never decide it from the bytes. Resolve every required consent
before writing: hand-written `expression.md` content, any ignore entry without evidence of Trellis
ownership, the decline artifact and rows above, and the bundle deletion. If consent is
unavailable, stop with the whole-project snapshot unchanged.

## 2. Stage byte-safe instruction-file removals

Prepare the resulting bytes for all five documented instruction files before changing any one:

- Remove only each recognized managed region. A block appended by a past setup carries one
  separator newline before it — remove that with the region; a **manually pasted** block may
  carry no separator at all, so remove exactly the region's own bytes and never a byte the user
  wrote.
- Preserve all bytes before and after the region exactly.
- Delete an instruction file only when it becomes empty because Trellis created it; otherwise keep
  it, even when the remainder is whitespace.

Do not treat an absent block as an error. A recognized, valid block is removed wherever a past setup or a
documented manual path placed it; ambiguous placement or markers were already a preflight failure.

## 3. Stage consented ignore cleanup

For ESLint, Prettier, Biome, and markdownlint targets, remove a `.trellis/` entry only with
**evidence** it is Trellis's — the user's confirmation from step 1, or surrounding content a past
setup demonstrably wrote. If an ignore file a past setup created then becomes empty, it may be
removed. Preserve all other patterns byte-for-byte. An entry without evidence is ambiguous and
required consent in step 1; never guess.

## 4. Apply the complete product-wide transaction

Only after every preflight and consent succeeds:

1. delete the rendered `.claude/rules/trellis.md` if present;
2. write or delete every staged documented instruction-file result;
3. apply the staged, consented ignore cleanup;
4. delete the vendored plugin bundle — **only** the `.claude/skills/trellis/` directory,
   never `.claude/skills/` itself, which is a shared directory this project may fill with
   unrelated skills — behind the explicit confirmation resolved in step 1; and
5. delete the shared `.trellis/` overlay and config last.

The bundle rides **with or before** the `.trellis/` deletion, never after it (`decision-0073` D3):
an interruption between the two must never leave the adoption-act artifact as the sole survivor of
a removal that already reported the config gone.

**Why the rendered file goes FIRST, not third.** A project mid-migration can hold
both delivery paths at once — a managed block plus `.trellis/internal/`, and a
rendered file. The property this ordering guarantees is narrow and load-bearing: **no
interruption window strands an always-loaded governing file whose rows are already gone**
(`decision-0068` D11). Removing `.trellis/` before the rendered file could do exactly that — a
posture header and a full rules body in always-loaded context, importing a file that no longer
exists. Removing the rendered file first cannot. Not every window is governed — after step 2, a
project that carried `.trellis/internal/` sits ungoverned until step 5 completes, in whichever
order these steps run — but an interruption in the early steps leaves the block still importing
live rows: governed, if redundantly, which is the safe direction.

Delete only the file. `.claude/rules/` is a shared directory a project may fill
with unrelated rules of its own, and the directory itself is not Trellis's to
remove.

Verify surrounding instruction-file and ignore-file bytes against the snapshots. If a preflight
failed, verify that every block and the overlay remain unchanged.

## 5. Confirm

Report **every state in the closed set** (`decision-0073` D1) as removed, retained, ambiguous, or
absent, by artifact name: the Claude block, the Codex bootstrap, inline managed blocks in the
other documented instruction files, the `.trellis/internal/` overlay, the legacy flat
`.trellis/trellis.md` overlay, the rendered `.claude/rules/trellis.md`, the vendored bundle at
`.claude/skills/trellis/`, `.trellis/rules.toml` (the rows — and the decline artifact, when it
was one), legacy consumer content, ignore entries, and the morph markers (`.trellis/rollback`,
the `trellis-pre-morph` tag).

When the bundle is **retained** — consent withheld, or found only after the fact — say what is
true under any answer to `decision-0068`'s open question 5: the bundle is the project-scope
**adoption-act artifact** and the product's skills on disk; delivery from its hook was
**measured inert** on the **headless** surface. Retaining it keeps the adoption signal in the repo, so the
report must name it rather than declare the project clean.

**The no-op predicate counts every removable state, not three.** Say Trellis is
**already absent** — and make no change — only when there is no managed block in any documented
instruction file, no overlay (`.trellis/internal/` or legacy flat `.trellis/trellis.md`), no
rendered `.claude/rules/trellis.md`, no vendored bundle at `.claude/skills/trellis/`, no
`.trellis/` config (a bare `rules.toml` is the plugin-native mode's removable state, not
absence), and no morph marker (`.trellis/rollback` or the `trellis-pre-morph` tag). That is
`decision-0073` D1's **S0-unadopted**, the one true already-absent state; every other named state
— S1 through S6, and config-only S0 — leaves this operation something to remove or surface. A
second remove is this same reported no-op, once everything above is genuinely absent.

## Reversing an M2 morph

Preflight (step 1) already looked for the two markers — `.trellis/rollback` and the
`trellis-pre-morph` tag — **before any write**, because step 5 of the transaction destroys the
first of them. If this project was changed by the **M2 morph** (the retired setup skill's
model-driven rewrite of the project's own files, on the `trellis/morph` branch), the reversal is
**git's**, using the rollback point the morph recorded — never a hand-edit back.

**A marker is a marker, not the morph as fact.** The tag can outlive the state: a rollback or a
completed removal leaves it behind (`decision-0073` D1, S6). To tell, read the project's own
instruction files — does the rewritten content still stand? If the morph was already reversed,
the leftover tag is stale; offer to clear it with `git tag -d trellis-pre-morph`, behind the same
explicit confirmation as every other change here.

If the morph stands, show the user the options (`git reset --hard trellis-pre-morph`,
`git revert`, or deleting the unmerged `trellis/morph` branch) and let them run the destructive
step. If **both markers are absent** — no `.trellis/rollback` and no `trellis-pre-morph` tag —
say you **cannot locate a rollback point** rather than guess (`spec-0004` §2): name what was
searched, and leave the reversal to the user's own git history.

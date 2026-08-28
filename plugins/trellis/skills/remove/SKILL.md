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
- `.trellis/`, including consumer-owned `rules.toml` and any legacy `expression.md`. Three things
  inside it need surfacing **by name** before any deletion:
  - **the decline artifact** — a `rules.toml` carrying top-level `governed = false` is the
    project's recorded decline (`decision-0070` D4). State the consequence out loud: on a machine
    with a user-scope plugin, deleting the recorded decline re-arms the adoption announcement, so
    the project returns to **unadopted** and every future session asks again until someone answers.
    Answering yes governs it at 15/15. Ignoring the prompt does not:
    **silence is not an adoption act** (`decision-0076`), so the project stays ungoverned and is
    asked again next session. Deleting it is legitimate — removal is removal — but doing it
    unnamed is not;
  - **the consumer's own rows** — `rules.toml` is the consumer's file, and deleting `.trellis/`
    deletes every row they edited. Show them the rows that will go, as part of the consent below;
  - **anything unrecognized** — a file in `.trellis/` this document does not name (not
    `rules.toml`, not overlay content, not `expression.md`, not `rollback`) is presumptively the
    user's. The transaction's last step deletes the directory wholesale, so name each such file
    in the `.trellis/` consent: "preserve every surrounding user byte" is kept by consent there,
    never by assumption;
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

Provenance cannot be proven from disk alone: no receipt exists for who added an ignore entry, an
unmarked line, or a file. Where a predicate would have to claim provenance to proceed, treat the
item as **ambiguous and consent-gated** — never decide it from the bytes.

**The consent model, stated once.** Two failure classes, two behaviors:

- **Integrity problems stop everything.** Ambiguous marker structure, any other preflight
  failure, or a §4 verification mismatch: stop with the whole-project snapshot unchanged (or,
  mid-§4, with no further step applied), and still produce the §5 report — every untouched item
  **retained**, save the stopping artifact itself, where the stop has one (a nobody-to-ask stop
  has none): an ambiguity or preflight stop's artifact is reported **ambiguous**; a §4
  mismatch's target is reported **verification-failed**. The stopping reason is named either
  way.
- **Denied item-scoped consents narrow the transaction; they never abort it.** Each of these is
  asked in this preflight, and a **denied** answer removes exactly that item from the staged
  transaction while everything else proceeds: the bundle deletion; the `.trellis/` deletion (one
  consent covering the directory and everything in it — the rows, a decline artifact,
  `expression.md`, and every unrecognized file named above); each ignore entry; and the deletion
  of any instruction file or ignore file that becomes empty once its Trellis content is removed.
  A denied item is **retained**, and §5 reports it as retained with its reason.

Consent that cannot be obtained at all — nobody to ask — is not a denial: stop with the
whole-project snapshot unchanged, and report per §5.

## 2. Stage byte-safe instruction-file removals

Prepare the resulting bytes for all five documented instruction files before changing any one:

- Remove only each recognized managed region. A block appended by a past setup carries one
  separator newline before it — remove that with the region; a **manually pasted** block may
  carry no separator at all, so remove exactly the region's own bytes and never a byte the user
  wrote.
- Preserve all bytes before and after the region exactly.
- Delete an instruction file only when it becomes empty once its managed regions are removed
  **and** its deletion was consented in the preflight (§1) — whether Trellis created the file
  cannot be proven from its bytes, so emptiness alone never authorizes a deletion. Without that
  consent, keep the file, even when the remainder is whitespace.

Do not treat an absent block as an error. A recognized, valid block is removed wherever a past setup or a
documented manual path placed it; ambiguous placement or markers were already a preflight failure.

## 3. Stage consented ignore cleanup

For ESLint, Prettier, Biome, and markdownlint targets, remove a `.trellis/` entry only with the
user's confirmation from the preflight (§1): who added an entry cannot be proven from the bytes,
so every entry is consent-gated — never guess. Preserve all other patterns byte-for-byte. If
removing the entry leaves the ignore file empty, delete the file exactly when its deletion was
consented in the preflight (§1); otherwise keep the empty file.

## 4. Apply the complete product-wide transaction

Only after every preflight and consent resolves (a denial narrows the transaction — §1 — it
never bars entry here):

1. delete the rendered `.claude/rules/trellis.md` if present;
2. write or delete every staged documented instruction-file result;
3. apply the staged, consented ignore cleanup;
4. delete the vendored plugin bundle — **only** the `.claude/skills/trellis/` directory,
   never `.claude/skills/` itself, which is a shared directory this project may fill with
   unrelated skills — behind the explicit confirmation resolved in the preflight (§1); and
5. delete the shared `.trellis/` overlay and config last.

The bundle rides **with or before** the `.trellis/` deletion, never after it (`decision-0073` D3):
an interruption between the two must never leave the adoption-act artifact as the sole survivor of
a removal that already reported the config gone. At project scope this very skill file ships
inside the bundle that step deletes — the remaining steps and the report are already in your
context, so carry them through to the end.

A step whose consent was denied in the preflight (§1) is skipped, not a failure: the transaction
narrows to the consented steps and §5 reports the skipped artifact as retained.

**Why the rendered file goes FIRST, not third.** A project mid-migration can hold
both delivery paths at once — a managed block plus `.trellis/internal/`, and a
rendered file. The rendered file has one shape — `install.sh`'s render: it embeds the posture
header and the full rules body, and **imports** its activation rows from `.trellis/rules.toml`.
The property this ordering guarantees is narrow and load-bearing: **no
interruption window strands an always-loaded governing file whose imported activation rows are
already gone** (`decision-0068` D11). Removing `.trellis/` before the rendered file could do
exactly that — the embedded rules body still loaded every session, importing an activation file
that no longer exists. Removing the rendered file first cannot. Not every window is governed — after step 2, a
project that carried `.trellis/internal/` sits ungoverned until step 5 completes, in whichever
order these steps run — but an interruption in the early steps leaves the block still importing
live rows: governed, if redundantly, which is the safe direction.

Delete only the file. `.claude/rules/` is a shared directory a project may fill
with unrelated rules of its own, and the directory itself is not Trellis's to
remove.

Each target — written **or deleted** — is verified against its §1 snapshot **immediately before
that target's own write or delete**, never once per numbered step: steps 2 and 3 each carry
several targets, and verifying them together at the step's start reopens the window this
paragraph exists to close. What is compared is the target's own bytes, the surrounding
instruction-file and ignore-file bytes around each managed region included. Verifying only
after the fact would make the mismatch branch below dead code, and would destroy a concurrent
user edit and then report clean. **A mismatch
stops the transaction where it stands**: apply no further step, leave every not-yet-applied
staging unwritten, and report per §5 — the mismatched path in the **verification-failed**
category, completed steps as removed, remaining steps as retained, and the fact that the tree
changed under the operation said in so many words. If a preflight failed, verify that every
block and the overlay remain unchanged.

## 5. Confirm

Report **every state in the closed set** (`decision-0073` D1) as removed, retained, ambiguous,
absent — or, after a §4 mismatch, **verification-failed** — by artifact name: the Claude block,
the Codex bootstrap, inline managed blocks in the
other documented instruction files, the `.trellis/internal/` overlay, the legacy flat
`.trellis/trellis.md` overlay, the rendered `.claude/rules/trellis.md`, the vendored bundle at
`.claude/skills/trellis/`, `.trellis/rules.toml` (the rows — and the decline artifact, when it
was one), legacy consumer content, ignore entries, and the morph markers (`.trellis/rollback`,
the `trellis-pre-morph` tag).

This report is owed on **every** exit — a completed transaction, one narrowed by denied
consents, a preflight stop, a verification failure, the already-absent no-op (the predicate
sentence below is itself that exit's report), or a morph hand-off (a **standing morph** only:
showing the rollback options,
or saying none can be located, **ends this operation** — the reversal itself is the user's, and a
completed reversal re-enters through a fresh preflight; a stale tag alone is no hand-off, and
that removal continues — see the morph section). A stop before any step is applied reports
**every untouched item retained** — save the stopping artifact itself, where the stop has one: it is reported
**ambiguous** after an ambiguity or preflight stop, **verification-failed** after a
first-target mismatch — with the
stopping reason named; a narrowed transaction reports each denied item
as retained by consent.

When the bundle is **retained** — consent withheld, or found only after the fact — say what is
true under any answer to `decision-0068`'s open question 5: the bundle is the project-scope
**adoption-act artifact** and the product's skills on disk; delivery from its hook was
**measured inert** on the **headless** surface. Retaining it keeps the adoption signal in the repo, so the
report must name it rather than declare the project clean. (At a `$HOME` project root there is
no project bundle to retain: `.claude/skills/trellis/` there is the user-scope install, outside
this operation's scope per the preflight's carve-out — report it as that, never as a retained
project artifact.)

**The no-op predicate counts every removable state, not three.** Say Trellis is
**already absent** — and make no change — only when there is no managed block in any documented
instruction file, no overlay (`.trellis/internal/` or legacy flat `.trellis/trellis.md`), no
rendered `.claude/rules/trellis.md`, no vendored bundle at `.claude/skills/trellis/` (with one
exception: at a `$HOME` project root that path is the user-scope install — `decision-0070` D6 —
whose presence does not make *this project* non-absent), no
`.trellis/` config (a bare `rules.toml` is the plugin-native mode's removable state, not
absence), no `.trellis/` ignore entry in any recognized lint/format ignore target (a leftover
line in `.prettierignore` is removable residue, not absence), and no morph marker
(`.trellis/rollback` or the `trellis-pre-morph` tag). That is
`decision-0073` D1's **S0-unadopted**, the one true already-absent state; every other named state
— S1 through S6, and config-only S0 — leaves this operation something to remove or surface. A
second remove is this same reported no-op, once everything above is genuinely absent.

## Reversing an M2 morph

Enter this section when either morph marker is present — or when **the user says this project
was morphed** even though neither marker is (the markers can be lost; the user's assertion is
this section's other trigger). The preflight (§1) already looked for the two markers —
`.trellis/rollback` and the
`trellis-pre-morph` tag — **before any write**, because step 5 of the transaction destroys the
first of them. If this project was changed by the **M2 morph** (the retired setup skill's
model-driven rewrite of the project's own files, on the `trellis/morph` branch), the reversal is
**git's**, using the rollback point the morph recorded — never a hand-edit back.

**A marker is a marker, not the morph as fact.** The tag can outlive the state: a rollback or a
completed removal leaves it behind (`decision-0073` D1, S6). To tell, read the project's own
instruction files — does the rewritten content still stand? If the morph was already reversed,
the leftover tag is stale; offer to clear it with `git tag -d trellis-pre-morph`, behind the same
explicit confirmation as every other change here. Clearing — or declining to clear — a stale tag
is **not** the morph hand-off: with the morph already reversed there is no morph to hand off, so
the removal **continues from §2** as an ordinary removal. Only a standing morph ends the
operation.

If the morph stands, show the user the options (`git reset --hard trellis-pre-morph`,
`git revert`, or deleting the unmerged `trellis/morph` branch) and let them run the destructive
step. If **both markers are absent** — no `.trellis/rollback` and no `trellis-pre-morph` tag,
reachable here on the user's say-so — say you **cannot locate a rollback point** rather than
guess (`spec-0004` §2): name what was searched, and leave the reversal to the user's own git
history.

After any reversal runs, **start over from the preflight (§1)**: the reversal rewrote the
working tree, so every earlier snapshot, staging, and consent is stale.

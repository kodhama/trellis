---
name: setup
description: Install or refresh Trellis governance on this project — read the posture, copy the pre-rendered payload from the plugin into .trellis/, patch the managed block in the instructions file, and verify against the shipped checksum manifest. Use when the user asks to set up, add, install, refresh, or apply Trellis in their repo.
---

# Set up Trellis in this project

You are installing the **Trellis** governance layer onto the user's project as the **M1 "alongside"
overlay**: a small `.trellis/` bundle plus one managed block in the instructions file, and **never
touch anything outside the bundle or the markers** (augment, never clobber).

You are a **mechanical copier** (`kodhama-0007`, "one render, many copiers"). Every byte of bundle
content was rendered at release time into `${CLAUDE_PLUGIN_ROOT}/reference/` — the payload. Your job
is exactly three verbs: **copy** payload files into `.trellis/`, **paste** one payload block between
the `trellis:begin`/`trellis:end` markers, and **verify** against the shipped checksum manifest. You
never compose, re-derive, paraphrase, or "fix up" bundle content — a prose re-derivation is a second
writer, and second writers drift (that is the incident class behind kodhama/trellis#112). There is
also **no binary to delegate to**: an earlier version of this skill preferred the `trellis` CLI as
the canonical writer (`kodhama-0005` rule 2); `kodhama-0007` rule 6 superseded that in part — the
deterministic thing is now the **artifact** (the payload + its manifest), not a privileged writer,
so this skill copies directly.

## The payload

In `${CLAUDE_PLUGIN_ROOT}/reference/`, where `<p>` is the posture key (`a` or `b`, step 1):

| payload file | goes to | notes |
|---|---|---|
| `invariants.md` | `.trellis/invariants.md` | full catalog; posture-independent |
| `profile-<p>.md` | `.trellis/profile.md` | the active-rules readout: each rule as a directive plus the ✗ failure it prevents, ending with the "(Generated from your profile …)" line — no posture/gatekeeper header. Auto-loaded, so it carries the rules themselves (`decision-0026`). Byte-identical across postures today |
| `trellis-<p>.md` | `.trellis/trellis.md` | the header agents read; the strictness line is the only per-posture difference |
| `block-claude.md` | the managed block in `CLAUDE.md` | import style: one line + `@.trellis/trellis.md` |
| `block-inline-<p>.md` | the managed block in a no-`@import` instructions file | inline style: the whole overlay, self-contained |
| `version` | *(not installed)* | the payload's render stamp (`payload@…`); the project's `.trellis/version` is stamped at install time instead (step 4) |
| `checksums` | *(not installed)* | `shasum -a 256` manifest over the other files — the verify oracle (step 6) |

## 1. Determine the posture — from `.trellis/expression.md`, never a guess

`.trellis/expression.md` is the project's one **hand-owned** bundle file (`kodhama-0007` rule 4):
its YAML frontmatter carries the machine-read posture; its body is the project's own hand-authored
expression of the invariants. Setup writes it **once** (the seeding below) and never again.

- **It exists and its frontmatter parses** (`profile: a` or `profile: b`): use that posture and ask
  nothing — this is the refresh / declarative-install path (a project may pre-commit the file so
  setup runs with no questions).
- **It exists but the frontmatter is missing or unparseable**: ask the user which posture applies —
  and offer to fix the frontmatter so the next run reads it. If no human is available to answer (an
  autonomous run), **stop and fail loudly**, naming the file and what is wrong with it. Never
  assume a default.
- **It does not exist** (first run — including refreshing an overlay installed before this file
  existed): ask the user to pick a posture, then seed the file.

Ask exactly this choice (the payload carries these two variants and no others; `seed` and `custom`
are parked per `decision-0033` — do not offer them):

- **A · conductor** — hold the rules firmly, by-the-book (strictness: "treat these as hard
  requirements").
- **B · author-adapt** — same rules, follow by default and adapt out loud (**default** if the user
  is unsure).

Then seed `.trellis/expression.md` with exactly this skeleton, filling in only `<p>` (the answered
key, `a` or `b`) and `<project>` (the project's name):

```markdown
---
profile: <p>
---

# <project> — Trellis expression

<!-- This file is yours (hand-owned; kodhama-0007 rule 4). Setup seeded it
once and will never rewrite it; it is excluded from the checksum manifest.
The `profile:` key above (a = conductor · b = author-adapt) is the only
machine-read line — a refresh reads it and asks nothing. Record below how
this project expresses the invariants: dials, mappings, gate tables.
Agents and humans read the body; machinery never parses it. -->
```

## 2. Guard hand-authored content before overwriting (the #112 backstop)

The three generated files in step 3 are **pure generated snapshots** (`decision-0035`): a re-run
rewrites each one whole, with no markers to protect additions. People have hand-appended content to
`profile.md` anyway, and lost it — twice for real (kodhama/trellis#106 → #111 on this repo's own
overlay; kodhama/trellis#112 downstream). So, **on a refresh, before overwriting**:

- If an existing `.trellis/profile.md` has anything after its closing "(Generated from your
  profile …)" line — or any of the three files differs from every payload variant in a way that
  looks hand-authored rather than merely stale — **stop and show the user the content**. Offer to
  move it into the body of `.trellis/expression.md` (its hand-owned home) before continuing.
- Never silently overwrite it. The ownership rule: every bundle file is 100% generated or 100%
  hand-owned, never mixed — `expression.md` is the hand-owned one; the three below are generated.

## 3. Copy the bundle — byte-for-byte, no edits

```sh
mkdir -p .trellis
cp "${CLAUDE_PLUGIN_ROOT}/reference/invariants.md"  .trellis/invariants.md
cp "${CLAUDE_PLUGIN_ROOT}/reference/profile-<p>.md" .trellis/profile.md
cp "${CLAUDE_PLUGIN_ROOT}/reference/trellis-<p>.md" .trellis/trellis.md
```

Copy with `cp`, not by retyping content. Do not reword, reformat, trim, or annotate any of these
files — step 6 checks them byte-for-byte against the manifest, and any "improvement" is a
verification failure you will have to undo.

## 4. Stamp `.trellis/version`

Run `git -C "${CLAUDE_PLUGIN_ROOT}" rev-parse --short HEAD` and write `plugin@<sha>` as the single
line of `.trellis/version` (plugin versions are commit SHAs, `decision-0036`). If that fails (the
plugin is not a git checkout), write `plugin@unknown` — an honest stamp beats none; an unstamped
overlay is invisible to the staleness hook (`decision-0039`). Do **not** copy the payload's own
`version` file here: `payload@…` is the render stamp, `plugin@<sha>` is the install stamp.

## 5. Patch the instructions file (augment, never clobber)

**Re-detect the target and style first.** Search the known instruction files — `CLAUDE.md`,
`AGENTS.md`, `GEMINI.md`, `.github/copilot-instructions.md`, `.clinerules` — for an existing
`<!-- trellis:begin` marker:

- **Exactly one file carries the block** → refresh it in place, keeping its style: a block
  containing `@.trellis/trellis.md` is **import** style (paste `block-claude.md`); a block carrying
  the rules directly is **inline** style (paste `block-inline-<p>.md`).
- **No file carries the block** → fresh install: target `CLAUDE.md` with **import** style
  (`block-claude.md`) — this skill runs inside Claude Code, where `@import` works. Create
  `CLAUDE.md` if it does not exist. (Non-Claude harnesses install the pre-rendered
  `block-inline-<p>.md` via the manual copy path documented in the repo README — not this skill's
  decision to make.)
- **More than one file carries the block, or you cannot classify an existing block's style** →
  ambiguous: **ask the user**; never guess.

**Before editing, save a pre-edit copy** of the target file (to your temp directory — you will need
it for verification), then paste:

- If the block exists: replace everything **from the first `<!-- trellis:begin` line through the
  first `<!-- trellis:end -->` line, inclusive**, with the payload block file's content. Touch
  nothing else — not even whitespace outside the markers.
- If it does not: append one blank separator line, then the block, then a trailing newline.
- Never write a second block.

## 6. Verify — data, not trust (`kodhama-0007` rule 3)

Run **all four** checks from the project root. Substitute `<p>`, `<target>` (the instructions file)
and `<block-file>` (`block-claude.md` or `block-inline-<p>.md`) for what you actually used.

**(a) Copied files match the shipped manifest:**

```sh
sed -n \
  -e 's|  invariants\.md$|  .trellis/invariants.md|p' \
  -e 's|  profile-<p>\.md$|  .trellis/profile.md|p' \
  -e 's|  trellis-<p>\.md$|  .trellis/trellis.md|p' \
  "${CLAUDE_PLUGIN_ROOT}/reference/checksums" | shasum -a 256 -c -
```

All three lines must print `OK`. (`.trellis/version` and `.trellis/expression.md` are deliberately
outside the manifest: the stamp is per-install, and `expression.md` is hand-owned.)

**(b) Exactly one begin and one end marker** in the target:

```sh
grep -c 'trellis:begin' <target>   # must print 1
grep -c 'trellis:end' <target>     # must print 1
```

**(c) The block is byte-identical to the payload:**

```sh
sed -n '/<!-- trellis:begin/,/<!-- trellis:end -->/p' <target> \
  | diff - <(cat "${CLAUDE_PLUGIN_ROOT}/reference/<block-file>"; echo)
```

Empty output = pass. (The `echo` supplies the trailing newline the block's last line gains inside
the target file; the payload block files end without one.)

**(d) Nothing outside the markers changed:**

```sh
diff <(sed '/<!-- trellis:begin/,/<!-- trellis:end -->/d' <pre-edit copy>) \
     <(sed '/<!-- trellis:begin/,/<!-- trellis:end -->/d' <target>)
```

On a refresh this must be empty; on a fresh append the only difference is the one added separator
blank line (and if you created the file, the pre-edit copy is empty and the post-edit remainder is
empty too).

**On any failure:** fix it mechanically — redo the copy (step 3) or the paste (step 5) for the
failing file — and re-run the checks. If it still fails, **report loudly**: name the exact check
that failed and what differed, leave the working tree as evidence, and stop. Never report success
on a failed or skipped check, and never hand-adjust file content to make a checksum pass — a loud
failure beats a plausible-looking install.

## 7. Confirm

Tell the user: which posture was used and whether it was **read** from `expression.md` or **asked
and seeded**; exactly what was written (`.trellis/{invariants,profile,trellis}.md`, the
`.trellis/version` stamp, `expression.md` seeded or left untouched); which instructions file was
patched and in which style; and the result of each verification check. They can remove it all any
time with `/trellis:remove`, or by deleting `.trellis/` and the managed block.

## Not this skill's job (yet)

Do **not** attempt a morph (**M2** — the model-driven rewrite of the project's own instruction
files) here. M2 hosting moves to this skill when the binary's M2 path retires (`kodhama-0007`
rule 5, slice 4 — kodhama/trellis#120); until that lands, the CLI remains its only home.

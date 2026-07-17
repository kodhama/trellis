---
name: setup
description: Install or refresh Trellis governance on this project — read the posture, copy the pre-rendered payload from the plugin into .trellis/, patch the managed block in the instructions file, and verify against the shipped checksum manifest. Also hosts the optional M2 morph (a model-driven rewrite of the project's own instructions, on a git branch) when the user explicitly asks for it. Use when the user asks to set up, add, install, refresh, apply, or morph Trellis in their repo.
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
the canonical writer (`kodhama-0005` rule 2); `kodhama-0007` rule 6 superseded `kodhama-0005` in
part, retiring rule 2 outright — the deterministic thing is now the **artifact** (the payload + its
manifest), not a privileged writer, so this skill copies directly.

(The one exception to "no model-driven work" is the **M2 morph** at the end of this file — and its
scope is the project's *own* files, never bundle content.)

## The payload

In `${CLAUDE_PLUGIN_ROOT}/reference/`, where `<p>` is the posture key (`a` or `b`, step 1):

| payload file | goes to | notes |
|---|---|---|
| `invariants.md` | `.trellis/invariants.md` | full catalog; posture-independent |
| `profile-<p>.md` | `.trellis/profile.md` | the active-rules readout: each rule as a directive plus the ✗ failure it prevents, ending with the "(Generated from your profile …)" line — no posture/gatekeeper header. Auto-loaded, so it carries the rules themselves (`decision-0026`). Byte-identical across postures today |
| `trellis-<p>.md` | `.trellis/trellis.md` | the header agents read; the strictness line is the only per-posture difference |
| `expression-<p>.md` | `.trellis/expression.md` — **first run only** | the hand-owned declaration file's seed: frontmatter pre-filled (`profile: <p>`), body a commented stub. Copied only when the file is absent (step 1); a refresh never touches an existing one (`kodhama-0007` rule 4) |
| `block-claude.md` | the managed block in `CLAUDE.md` | import style: one line + `@.trellis/trellis.md` |
| `block-inline-<p>.md` | the managed block in a no-`@import` instructions file | inline style: the whole overlay, self-contained |
| `version` | `.trellis/version` | the payload's render stamp (`payload@…`), copied like any other payload file (step 4) — the staleness hook compares the two files (`decision-0043`) |
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

Then seed it by copying the payload's pre-filled skeleton for the answered posture — a copy, not a
composition (the skeleton is payload content like every other bundle file, so it has no second home
in this skill's prose and nothing is left to fill in):

```sh
cp "${CLAUDE_PLUGIN_ROOT}/reference/expression-<p>.md" .trellis/expression.md
```

From that moment the file is the project's own: they may retitle it and write the body freely, and
no later run of this skill touches it.

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

## 4. Stamp `.trellis/version` — a copy, like everything else

```sh
cp "${CLAUDE_PLUGIN_ROOT}/reference/version" .trellis/version
```

The payload's content-derived render stamp (`payload@…`) **is** the install stamp
(`decision-0043`, superseding `decision-0039` rule 2's `plugin@<sha>` format): the bundled
`SessionStart` staleness hook compares this file against the installed plugin's
`reference/version` — file to file, no git, no binary — and nudges a refresh when they differ.
An unstamped overlay is invisible to that check, so never skip this step.

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
  -e 's|  version$|  .trellis/version|p' \
  "${CLAUDE_PLUGIN_ROOT}/reference/checksums" | shasum -a 256 -c -
```

All four lines must print `OK`. (`.trellis/expression.md` is deliberately outside install-time
verification: it is hand-owned from the moment it is seeded — the payload's `expression-<p>.md`
skeletons are manifest-covered like any payload file, but the installed copy is the project's to
change.)

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
`.trellis/version` payload stamp, `expression.md` seeded or left untouched); which instructions
file was patched and in which style; and the result of each verification check. They can remove it
all any time with `/trellis:remove`, or by deleting `.trellis/` and the managed block.

## 8. Hand back — setup performs no git, and imposes no landing workflow (`decision-0048`)

The overlay is now written and **uncommitted**. How it gets committed or landed is **this
project's decision, made by this project's own conventions** — not setup's. So this skill
**performs no git of its own** (no `add`/`commit`, no branch, no push, no PR) and — just as
important — **injects no landing opinion** into the session it runs in: it does *not* recommend a
PR, a branch, or committing anywhere.

The reason for the restraint, not merely "hand off to a PR": setup runs **inline in the consumer's
own conversation**, so any git-workflow instruction here would bias how *that* session handles git
for its own unrelated work — importing trellis's preferences into a project that has its own. A
recommendation would be the very contamination this step exists to prevent (`decision-0048`); a
neutral hand-back avoids it.

So, once the files are written and verified:

- **Surface the state plainly** — that the overlay is written and uncommitted, and which files
  changed — and **let the user decide how to land it**, following whatever this repo normally does.
  Do **not** act on git yourself, and do **not** commit onto the current branch (least of all a
  default branch such as `main`/`master`).
- **Nothing to land** (an idempotent refresh whose `git diff` is empty, or a project that is not a
  git repo): say so and stop.
- **No human to answer** (an autonomous run): leave the change in the working tree, report exactly
  what is uncommitted, and stop — never land it unasked (`floor-intent-gate`).

## M2 — morph (model-driven, only on explicit request)

Everything above is the default, **M1** flow. **M2** rewrites the project's *own* instruction
files to bake the invariants in — hosted here since the binary's M2 path retired with the binary
channel (`kodhama-0007` rule 5, kodhama/trellis#120; the flow below ports `applym2.go`'s
contract). Run it **only when the user explicitly asks** for a morph/rewrite — never as a default,
and never combined silently with an M1 install.

**The boundary stays absolute.** M2 is the one place in this skill where model-driven writing is
sanctioned, and its scope is the project's own files (`CLAUDE.md`, rule/convention files). It is
**not** bundle composition: `.trellis/` files and the managed block still come only from the
payload via steps 3–6 above — a morph never writes, rewords, or "adapts" bundle content.

1. **Refuse without git.** The rewrite must be reviewable and revertable. If the project is not a
   git repository, stop: suggest `git init` first, or the M1 overlay instead.
2. **Determine the posture** exactly as in step 1 above (read `expression.md`, or ask).
3. **Record the rollback point, then branch.** Note the current commit
   (`git rev-parse HEAD`), create and switch to a fresh branch `trellis/morph`
   (`git checkout -b trellis/morph`), write the pre-morph SHA as the single line of
   `.trellis/rollback`, and set a tag that survives a reset:
   `git tag -f trellis-pre-morph <sha>`. Never morph the working branch in place.
4. **Perform the rewrite yourself** (you are the model the binary used to shell out to). Rewrite
   the project's instruction files to bake in the active invariants, **in the project's own voice
   and structure**. Preserve the project's existing behaviors unless they directly conflict. The
   single most important behavior to encode: **surface any human-gated handover performed without
   its human approval**; agent-gated handovers proceed silently. Respect whatever gatekeeping the
   project already declares — detect it, do not impose it. Keep the edits direct and reviewable.
5. **Stop and hand the diff to the human.** Summarize what changed, point at the branch, and let
   *them* review the diff and open/merge a PR — the merge is theirs, never yours. Reversal is
   git's: `git reset --hard trellis-pre-morph` (or the SHA in `.trellis/rollback`), or simply
   delete the branch.

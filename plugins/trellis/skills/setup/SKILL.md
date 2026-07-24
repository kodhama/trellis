---
name: setup
description: Install or refresh Trellis governance on this project — read and validate the config in .trellis/rules.toml (seeding it from a posture question, or migrating a legacy profile-key overlay, when absent), copy the pre-rendered payload from the plugin into .trellis/internal/ (including the complete rules readout — rows govern rule activation at read time), patch the managed block in the instructions file, and verify against the shipped checksum manifest. Also hosts the optional M2 morph (a model-driven rewrite of the project's own instructions, on a git branch) when the user explicitly asks for it. Use when the user asks to set up, add, install, refresh, apply, or morph Trellis in their repo.
---

# Set up Trellis in this project

You are installing the **Trellis** governance layer onto the user's project as the **M1 "alongside"
overlay**: a small `.trellis/` bundle plus one managed block in the instructions file, and **never
touch anything outside the bundle or the markers** (augment, never clobber).

You are a **mechanical copier** (`kodhama-0007`, "one render, many copiers"). Every byte of bundle
content was rendered at release time into `${TRELLIS_PLUGIN_ROOT}/reference/` — the payload. Your job
is exactly three verbs: **copy** payload files into `.trellis/` (the readout ships complete —
per-row assembly retired with `decision-0053`; the `rules.toml` rows govern which rules apply at
read time), **paste** one payload block between the `trellis:begin`/`trellis:end` markers, and
**verify** against the shipped checksum manifest. You never compose, re-derive, paraphrase, or
"fix up" bundle content — a prose re-derivation is a second writer, and second writers drift (that
is the incident class behind kodhama/trellis#112). There is also **no binary to delegate to**: the
deterministic thing is the **artifact** (the payload + its manifest), not a privileged writer
(`kodhama-0007` rule 6), so this skill copies directly.

(The one exception to "no model-driven work" is the **M2 morph** at the end of this file — and its
scope is the project's *own* files, never bundle content.)

## 0. Detect the host and preflight the whole transaction

This preflight is mandatory and finishes **before the first project write**. Staging in a temporary
directory is allowed; `.trellis/`, `AGENTS.md`, and `CLAUDE.md` must remain byte-for-byte unchanged
until every check below passes.

Apply this exact host precedence:

1. If `PLUGIN_ROOT` is set, it must name an absolute existing directory whose
   `.codex-plugin/plugin.json` parses as JSON and has `name` exactly `"trellis"`. A valid root
   selects **Codex**, even when `CLAUDE_PLUGIN_ROOT` is also set. An invalid Codex root stops; it
   never falls through to Claude.
2. Only when `PLUGIN_ROOT` is absent may an absolute existing `CLAUDE_PLUGIN_ROOT` select
   **Claude**, and only when `.claude-plugin/plugin.json` parses with `name` exactly `"trellis"`.
3. A missing, relative, nonexistent, malformed, or wrong-plugin root stops visibly.

Set `TRELLIS_PLUGIN_ROOT` to the validated selected root for every command below. For Codex,
preflight `node`: major version **Node.js 20** or newer enables the native hook. If it is unavailable
or older, continue with a visible **bootstrap-only** degradation: install the overlay and fallback,
but do not claim native delivery. Node is a local plugin-hook prerequisite, never a project runtime
or daemon.

Then preflight, without writing. The documented instruction-file inventory is exactly
`CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.github/copilot-instructions.md`, and `.clinerules`;
scan **all five**, not only the selected target, for every legacy/manual
`trellis:begin`/`trellis:end` block (including an inline/full-rule block) and every
`trellis:codex-bootstrap:begin`/`trellis:codex-bootstrap:end` block.

Classify one **canonical opposite-host block** as an explicit exemption:

- **When Codex is selected**, exactly one valid Claude import block in `CLAUDE.md`, byte-identical
  to `block-claude.md`, is allowed and byte-preserved.
- **When Claude is selected**, exactly one valid Codex bootstrap in `AGENTS.md`, byte-identical to
  `block-codex.md`, is allowed and byte-preserved.

This exemption applies only at those canonical paths with exact payload bytes. It is not a
migration input and requires no removal consent. A mismatched, misplaced, duplicate, nested, or
otherwise ambiguous opposite-host block is not exempt.

Every other existing block is a migration input, never permission to add a second transport:

- In the selected target, show the byte-exact proposed replacement and obtain explicit consent to
  migrate the old block to the host-native block in the same transaction.
- A non-exempt recognized block outside the selected target must make setup stop, unless the user gives
  explicit consent to remove that exact block in the same transaction.
- Never add a Codex bootstrap while any inline/full-rule block remains in any documented
  instruction file. Never leave both a migrated block and its replacement.

Continue the preflight:

- the complete payload checksum manifest and every source file used below, including
  `block-claude.md` and `block-codex.md`;
- the existing `.trellis/rules.toml`, or the user's resolved posture choice when a seed is needed;
- every generated-file clobber guard and legacy migration/consent decision;
- the selected target: `CLAUDE.md` for Claude, `AGENTS.md` for Codex, plus every consented
  cross-file removal above;
- exactly zero or one nonnested, paired recognized marker set per file; duplicate, unpaired,
  nested, overlapping, or ambiguous markers anywhere stop;
- a byte snapshot of both instruction files and every project path this transaction may touch.

Do not start steps 2–7 until all choices and checks above resolve. This scan and consent gate
finishes before the first project write. On any preflight failure, report it and verify the
whole-project snapshot is unchanged. The skill directs these checks; it does not claim that
best-effort model execution turns them into a transactional filesystem primitive.

## The overlay's two halves — placement by authority (`decision-0051` rule 1, as amended)

- **`.trellis/` root — consumer-authoritative.** **`rules.toml` alone** (the machine-read config:
  which rules apply, how strictly — its rows govern rule activation **at read time**,
  `decision-0053`: a row edit takes effect at the next host context-loading boundary, no refresh needed). Seeded **once**; a
  refresh never clobbers it; excluded from manifest verification — the consumer owns it. There is
  **no** `expression.md`: it is retired from the bundle (`decision-0051` amendment) — a project's
  governance prose belongs in its **own instructions file**, which every harness already loads;
  trellis reserves no home for it.
- **`.trellis/internal/` — trellis-authoritative.** `trellis.md`, `rules.md` (the complete
  readout), `invariants.md`, `version`. Copied verbatim on **every** refresh; hand-edits are
  overwritten; manifest-verified byte-for-byte.

## The payload

In `${TRELLIS_PLUGIN_ROOT}/reference/`, where `<p>` is the posture key (`a` or `b`, step 1):

| payload file | goes to | notes |
|---|---|---|
| `invariants.md` | `.trellis/internal/invariants.md` | full catalog; posture-independent |
| `trellis-<p>.md` | `.trellis/internal/trellis.md` | the header agents read; the strictness line is the only per-posture difference. It imports its sibling `rules.md` (imports resolve relative to the importing file) |
| `rules.md` | `.trellis/internal/rules.md` — **every** install | the complete readout: all 14 rules, each ending with its row's slug, opened by the authority header — rules apply only where the `rules.toml` row is `active = true`; the two `floor-` rows apply regardless (`decision-0053`) |
| `rules-<p>.toml` | `.trellis/rules.toml` — **first run only** | the posture seed: explicit rows, all active, `seeded_from` + `strictness` pre-filled. Consumer-owned from the moment it is seeded — editing rows *is* the configuration act, and row edits take effect at the next supported boundary; a refresh reads rows, validates them, and asks nothing |
| `block-claude.md` | the managed block in `CLAUDE.md` | import style: two block-level imports — `@.trellis/internal/trellis.md` and `@.trellis/rules.toml` (the rows load below the rules) |
| `block-codex.md` | the managed receipt/fallback in `AGENTS.md` | Codex only: best-effort installed-file bootstrap, never rule prose or row values |
| `block-inline-<p>-head.md` · `block-inline-tail.md` | dormant shared payload for the documented manual no-import path | not registered or installed by either Phase 1 host branch |
| `block-inline-<p>.md` | the manual no-import seed-state sandwich | not a claimed native Phase 1 transport |
| `version` | `.trellis/internal/version` | the payload's render stamp (`payload@…`) — the staleness hook compares the two files (`decision-0043`, path per `decision-0051`) |
| `checksums` | *(not installed)* | `shasum -a 256` manifest over the other files — the verify oracle (step 8) |

## 1. Read the config — `.trellis/rules.toml`, never a guess

`.trellis/rules.toml` is the project's machine-read configuration (`decision-0051` rule 2:
posture-as-seed, rows-as-truth). Its `strictness` key selects the posture variant of the header;
its `[rules]` rows govern which rules apply **at read time** (`decision-0053` — the readout ships
complete; its authority header defers to these rows). `seeded_from` is provenance only — **the
rows win if they diverge**. Every refresh **validates** the rows (`decision-0053` point 3 —
validation survives assembly's retirement):

- **It exists and parses** (`strictness` is `"firm"` or `"adaptive"`, and `[rules]` rows are
  present): use it and ask nothing — this is the refresh / declarative-install path. `strictness
  = "firm"` → `<p>` = `a`; `"adaptive"` → `<p>` = `b`.
- **It exists but `strictness` is missing/unparseable, a row names a slug the readout has no
  rule for (unknown slug), or a readout rule's slug has no row (missing row)**: **ask the user**
  — the seed writes every row, so any of these means a hand-edit went wrong; offer to fix the
  file. If no human is available (an autonomous run), **stop and fail loudly**, naming the file
  and what is wrong. Never assume a default, and never silently drop or invent a row.
- **A floor row (`floor-transparency`, `floor-intent-gate`) is set `active = false`**: the row is
  **never silently honored** — the floors are not rows a consumer can turn off; the readout's
  authority header holds them (`decision-0051` rule 3, held per `decision-0053`). Continue the
  install, but **surface it loudly in step 10**, naming the row as **overridden-by-floor**: the
  floor rule applies regardless of the row's value.
- **The file's *comments* carry retired pre-`decision-0053` wording** (a top comment
  claiming rows apply only after a refresh, or floor rows annotated with the old
  assembly note): **offer** — never impose (`decision-0049` consent pattern, per the
  `decision-0053` amendment) — to update the comments to the current seed's live-rows
  wording, **touching comments only, never a row**. If declined, note it and move on;
  if no human is available, leave the file untouched and **report the stale comments
  loudly** — they contradict the authority header this context now carries.
- **It does not exist, but a legacy `.trellis/expression.md` has a `profile:` frontmatter key**
  (`a` or `b` — an overlay from before `decision-0051`): **migrate.** Seed the config from that
  posture in the transaction staging area. Do not copy it into the project until the complete
  preflight passes. The eventual write is
  `cp "${TRELLIS_PLUGIN_ROOT}/reference/rules-<p>.toml" .trellis/rules.toml`. The
  legacy file itself is handled in step 6 (preserved, never silently deleted).
- **Neither exists** (first run): ask the user to pick a posture, then seed.

Ask exactly this choice (the payload carries these two variants and no others; `seed` and `custom`
stay parked per `decision-0033`/`decision-0051` rule 7 — do not offer them):

- **A · conductor** — hold the rules firmly, by-the-book (strictness: "treat these as hard
  requirements").
- **B · author-adapt** — same rules, follow by default and adapt out loud (**default** if the user
  is unsure).

After the complete step-0 preflight succeeds, seed the one consumer-owned file by copying — a copy, not a composition (the seed is payload
content like every other bundle file, so it has no second home in this skill's prose and nothing
is left to fill in):

```sh
mkdir -p .trellis
cp "${TRELLIS_PLUGIN_ROOT}/reference/rules-<p>.toml" .trellis/rules.toml
```

From that moment the file is the project's own: they may edit rows freely, and no later run of
this skill rewrites it (seed-once, never-clobber).

## 2. Guard hand-authored content before overwriting (the #112 backstop)

The generated files (everything under `.trellis/internal/`, and the flat-layout files step 6
migrates) are **pure generated snapshots**: a re-run rewrites each one whole, with no markers to
protect additions. People have hand-appended content to the generated readout anyway, and lost it
— twice for real (kodhama/trellis#106 → #111; #112 downstream). So, **on a refresh, before
overwriting**:

- If any generated file differs from every payload variant in a way that looks hand-authored
  rather than merely stale, **stop and show the user the content**. For **legacy** files the old
  closing-line sentinel still works: a flat `.trellis/profile.md`, or a pre-`decision-0053`
  `rules.md`, closes with a "(Generated from your …" line — anything after it is hand-authored.
  (The current readout has no closing sentinel — that footer retired with `decision-0053`; the
  detector for current-era files is the whole-file comparison against the payload.) Offer to
  move the content into the project's **own instructions file**, outside the managed block (the
  amendment's home for a project's governance prose — trellis reserves none), before continuing.
- Never silently overwrite it. The ownership rule holds at directory granularity now
  (`decision-0051` rule 6): everything under `internal/` is 100% generated; `rules.toml` is 100%
  consumer-owned.
- `rules.toml` needs no rescue: its rows are the consumer's and are **never clobbered** — the
  guard's target stays the generated files.

## 3. Copy the generated files — byte-for-byte, no edits

```sh
mkdir -p .trellis/internal
cp "${TRELLIS_PLUGIN_ROOT}/reference/invariants.md"  .trellis/internal/invariants.md
cp "${TRELLIS_PLUGIN_ROOT}/reference/trellis-<p>.md" .trellis/internal/trellis.md
```

Copy with `cp`, not by retyping content. Do not reword, reformat, trim, or annotate any of these
files — step 8 checks them byte-for-byte against the manifest, and any "improvement" is a
verification failure you will have to undo.

## 4. Copy the complete readout (`decision-0053` — assembly retired)

Every install carries the complete readout, whatever the rows say — a plain copy, the same as
every other generated file:

```sh
cp "${TRELLIS_PLUGIN_ROOT}/reference/rules.md" .trellis/internal/rules.md
```

Which rules **apply** is not decided here: the readout opens with an **authority header**
instructing agents to apply each rule only where its `rules.toml` row is `active = true`; the two
`floor-` rows apply regardless of their row value. So **row edits take effect at the next supported
host context-loading boundary** — Claude imports current rows each session and the Codex startup
hook reads them directly. A refresh updates the payload copies and **validates** the rows
(step 1); it is not what activates them. (Per-row subset assembly, and the floor-held assembly
override, retired with `decision-0053` — the floors are now held by the header's floor sentence
plus the step-1 validation.)

## 5. Stamp `.trellis/internal/version` — a copy, like everything else

```sh
cp "${TRELLIS_PLUGIN_ROOT}/reference/version" .trellis/internal/version
```

The payload's content-derived render stamp (`payload@…`) **is** the install stamp
(`decision-0043`): the bundled `SessionStart` staleness hook compares this file against the
installed plugin's `reference/version` — file to file, no git, no binary — and nudges a refresh
when they differ. An unstamped overlay is invisible to that check, so never skip this step.

## 6. Migrate an earlier overlay (pre-`decision-0051`, or pre-amendment)

**Flat layout.** Overlays installed before the authority split keep the generated files directly
in `.trellis/`. After steps 3–5 have written the new layout, **delete the old-path copies** so
the two layouts never sit side by side:

```sh
rm -f .trellis/trellis.md .trellis/profile.md .trellis/invariants.md .trellis/version
```

(Step 2 already rescued anything hand-authored in them; their live content is under `internal/`
now — `profile.md`'s readout lives on as `rules.md`, renamed by `decision-0051` rule 5. The
legacy `profile:` frontmatter key was migrated in step 1.) Report every deleted file in step 10.

**A leftover `.trellis/expression.md`** (seeded before the `decision-0051` amendment retired it
from the bundle) is **preserved — never silently deleted**. Its disposition is the user's:

- **It carries a hand-written body** (anything beyond the seed stub's comment and heading):
  **surface the content** and *offer* to move it into the project's **own instructions file**,
  outside the managed block — the amendment's home for governance prose — or to leave the file in
  place untouched. Either answer is fine (a left-in-place file harms nothing; nothing reads it);
  **write or delete nothing without a yes** (`floor-intent-gate`).
- **It is only the seed stub** (no hand-written content): you may *offer* to delete it as part of
  the migration; if declined, leave it.

Report the disposition — moved, left in place, or deleted — in step 10.

## 7. Patch the instructions file (augment, never clobber)

Patch exactly the selected host's target from step 0:

| selected host | target | payload block | default other-host result |
|---|---|---|---|
| Claude | `CLAUDE.md` | `block-claude.md` between `trellis:begin` / `trellis:end` | `AGENTS.md` remains byte-for-byte identical |
| Codex | `AGENTS.md` | `block-codex.md` between `trellis:codex-bootstrap:begin` / `trellis:codex-bootstrap:end` | `CLAUDE.md` remains byte-for-byte identical |

The Codex bootstrap is a receipt and best-effort fallback for a **fresh trusted local Codex
startup**. It contains no rule prose, generated readout, row values, or posture content. The native
hook reads current rows at startup. Phase 1 supports a fresh trusted local Codex startup. An
ordinary refresh is not a preset reset and must never overwrite an existing valid `rules.toml`.

Use the all-instruction-file marker analysis and migration consents completed in step 0:

- With no pair, append one separator newline, the selected payload block verbatim, and one final
  newline. Create only the selected target if it was absent.
- With one valid selected-target pair, replace from its begin marker through paired end marker
  inclusive with the selected payload block after the explicit migration consent. Preserve every
  byte outside it.
- Remove each specifically consented legacy/manual or outside-selected-target block as part of this
  same transaction, preserving every surrounding byte.
- Any duplicate, unpaired, nested, overlapping, unconsented cross-file, or otherwise ambiguous
  marker stops before any overlay or instruction-file write.

Save the pre-edit target bytes for verification. Running Claude setup then Codex setup, or the
reverse, may leave exactly one transport in each canonical file only when the all-file scan found
no old inline/full-rule copy; setup must never claim co-installation is safe while another effective
rule copy remains.

## 8. Verify — data, not trust (`kodhama-0007` rule 3)

Run **all five** checks from the project root. Substitute `<p>`, the selected `<target>`, marker
names, and payload block for the current host.

**(a) Copied files match the shipped manifest:**

```sh
sed -n \
  -e 's|  invariants\.md$|  .trellis/internal/invariants.md|p' \
  -e 's|  trellis-<p>\.md$|  .trellis/internal/trellis.md|p' \
  -e 's|  version$|  .trellis/internal/version|p' \
  "${TRELLIS_PLUGIN_ROOT}/reference/checksums" | shasum -a 256 -c -
```

All three lines must print `OK`. (`.trellis/rules.toml` is deliberately outside install-time
verification: it is consumer-owned from the moment it is seeded — the payload's `rules-<p>.toml`
seeds are manifest-covered like any payload file, but the installed copy is the project's to
change, `decision-0051` rule 1.)

**(b) The installed readout is byte-identical to the shipped complete readout** (`decision-0053`
— the readout is a static payload file on every install, so the oracle is a plain byte-compare):

```sh
diff .trellis/internal/rules.md "${TRELLIS_PLUGIN_ROOT}/reference/rules.md"
```

Empty output = pass.

**(c) Exactly one selected-host begin and one end marker** in the target. When step 0 found no
consented cross-file migration, every other documented instruction file is byte-for-byte identical
to its snapshot. Otherwise, compare each changed file to the exact snapshot-minus-consented-block
result:

```sh
grep -c '<selected-begin-marker>' <target>   # must print 1
grep -c '<selected-end-marker>' <target>     # must print 1
cmp <unchanged-instruction-snapshot> <unchanged-instruction-file>
cmp <staged-snapshot-minus-consented-block> <consented-migration-file>
```

**(d) The block is byte-identical to its host oracle**:

````sh
# Claude:
sed -n '/<!-- trellis:begin/,/<!-- trellis:end -->/p' <target> \
  | diff - <(cat "${TRELLIS_PLUGIN_ROOT}/reference/block-claude.md"; echo)

# Codex:
sed -n '/<!-- trellis:codex-bootstrap:begin/,/<!-- trellis:codex-bootstrap:end -->/p' <target> \
  | diff - <(cat "${TRELLIS_PLUGIN_ROOT}/reference/block-codex.md"; echo)
````

Empty output = pass. (The `echo` supplies the trailing newline the block's last line gains inside
the target file; payload blocks end without one.)

**(e) Nothing outside the markers changed:**

```sh
diff <(sed '/<selected-begin-marker>/,/<selected-end-marker>/d' <pre-edit copy>) \
     <(sed '/<selected-begin-marker>/,/<selected-end-marker>/d' <target>)
```

On a refresh this must be empty; on a fresh append the only difference is the one added separator
blank line (and if you created the file, the pre-edit copy is empty and the post-edit remainder is
empty too).

**On any failure:** fix it mechanically — redo the copy (steps 3–4) or the paste (step 7) for
the failing file — and re-run the checks. If it still fails, **report loudly**:
name the exact check that failed and what differed, leave the working tree as evidence, and stop.
Never report success on a failed or skipped check, and never hand-adjust file content to make a
checksum pass — a loud failure beats a plausible-looking install.

## 9. Offer to hide `.trellis/` from the consumer's own tooling (`decision-0049`)

`.trellis/` is **vendored trellis territory, not consumer source** — but the consumer's linters and
formatters don't know that. A Prettier or markdownlint pass that reformats the generated files
under `.trellis/internal/` would change their bytes and **break the step-8 checksum verify on the
next refresh** — the consumer's tidy-up silently corrupting the install. So, **offer** (never
impose) to keep `.trellis/` out of their tooling:

- **Detect**, best-effort, by config presence — ESLint (`.eslintrc*` / `eslint.config.*` /
  `eslintConfig` in `package.json`), Prettier (`.prettierrc*` / `.prettierignore`), Biome
  (`biome.json`), markdownlint (`.markdownlint*` / `.markdownlintignore`). If none is found, say so
  plainly and skip — never invent a tool that isn't there.
- **Offer, don't impose.** For each tool found, ask whether to add `.trellis/` to its ignore.
  **Write nothing without a yes** (`floor-intent-gate`); if declined, note it and move on.
- **Augment-never-clobber.** Add `.trellis/` to that tool's own ignore mechanism (`.prettierignore`,
  `.eslintignore` or `ignorePatterns`, `.markdownlintignore`, …) — skip if it is already there,
  touch no other line, and create an ignore file only if the tool needs one and none exists. This is
  the **one** place setup may touch a consumer file **outside `.trellis/` and the managed block**,
  and only with consent.
- **Report exactly what you touched** — each ignore file and the line added — in step 10.

(Target the whole `.trellis/` directory, matching the namespace boundary — `decision-0049`'s scope
is unchanged by the split. `rules.toml` is the consumer's own file and needs no checksum
protection — a formatter over it is harmless — but the load-bearing ignore target is `internal/`,
and ignoring the whole directory is simpler and costs nothing.)

## 10. Confirm

Tell the user:

- **Where the config came from**: read from `.trellis/rules.toml`'s rows, migrated from a legacy
  `profile:` frontmatter key, or asked and seeded.
- **Row semantics — live at the next boundary** (`decision-0053`, clarified by
  `decision-0058`): row edits take effect at the next supported host context-loading boundary
  without setup or refresh; they do not mutate a model context already in flight. This ordinary
  refresh validated the rows and updated generated payload copies but did not perform a preset
  reset.
- **Floor overrides, loudly** (`decision-0051` rule 3, held per `decision-0053`): if
  `floor-transparency` or `floor-intent-gate` was set `active = false`, name the row as
  **overridden-by-floor** — the floors are not rows a consumer can turn off; the floor rule
  applies regardless (the readout's authority header says so), and the row was **never silently
  honored**.
- **Which rows are inactive**: the readout carries all 14 rules on every install; name each row
  set `active = false` — those rules do not apply at read time (and for a floor row, that it is
  overridden as above).
- **Exactly what was written**: `.trellis/internal/{invariants,trellis,rules}.md`, the
  `.trellis/internal/version` stamp, `rules.toml` seeded or left untouched.
- **Any flat-layout files deleted** by the step-6 migration, and **the disposition of a leftover
  legacy `expression.md`** (content moved into the instructions file, left in place, or deleted
  with consent).
- Which host was selected, which instructions file was patched, and whether Codex native delivery
  passed the Node.js 20 preflight or is **bootstrap-only**; **any lint/format ignore entry added
  (which file, which line), or that the offer was declined or no tooling was found**; and the
  result of each verification check.

They can remove it all any time with `/trellis:remove`, or by deleting `.trellis/` and the managed
block.

## 11. Hand back — setup performs no git, and imposes no landing workflow (`decision-0048`)

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
payload via steps 3–8 above — a morph never writes, rewords, or "adapts" bundle content.

1. **Refuse without git.** The rewrite must be reviewable and revertable. If the project is not a
   git repository, stop: suggest `git init` first, or the M1 overlay instead.
2. **Determine the posture** exactly as in step 1 above (read `rules.toml`, migrate a legacy
   overlay, or ask).
3. **Record the rollback point, then branch.** Note the current commit
   (`git rev-parse HEAD`), create and switch to a fresh branch `trellis/morph`
   (`git checkout -b trellis/morph`), write the pre-morph SHA as the single line of
   `.trellis/rollback`, and set a tag that survives a reset:
   `git tag -f trellis-pre-morph <sha>`. Never morph the working branch in place.
4. **Dispatch a cold sub-agent to perform the rewrite — never do it in this warm session
   (`decision-0050`).** The rewrite is the one *generative* step, and a generative step carries
   whatever context the running session holds; performing it here would let this session's ambient
   context bleed into the consumer's own files. So **dispatch a sub-agent** (the Task/Agent tool)
   whose prompt carries **only its declared inputs** — the posture (step 2), the specific instruction
   files to rewrite, and the invariants to bake in (`.trellis/internal/invariants.md`). That
   sub-agent:
   - **reads only those inputs, never this conversation** — the isolation (`inv-bounded-context`) is
     the whole point, and it is what M1's opinion-removal could not achieve for a *generative* step;
   - rewrites the project's instruction files **in the project's own voice and structure**, preserving
     existing behaviors unless they directly conflict. The single most important behavior to encode:
     **surface any human-gated handover performed without its human approval**; agent-gated handovers
     proceed silently. Respect whatever gatekeeping the project already declares — detect it, do not
     impose it; keep the edits direct and reviewable;
   - writes the rewritten files **on the `trellis/morph` branch** (step 3), makes **no** git decision,
     and asks the human **nothing** — it returns a summary of what it changed.
   You (the warm session) do not rewrite anything in-context; you dispatch, then go to step 5.
5. **Stop and hand the diff to the human.** Summarize what changed, point at the branch, and let
   *them* review the diff and open/merge a PR — the merge is theirs, never yours. Reversal is
   git's: `git reset --hard trellis-pre-morph` (or the SHA in `.trellis/rollback`), or simply
   delete the branch.

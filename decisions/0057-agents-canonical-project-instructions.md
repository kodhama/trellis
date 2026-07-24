---
id: decision-0057
type: decision
status: approved  # maintainer intent act 2026-07-23: AGENTS is canonical shared prose; CLAUDE retains @AGENTS.md plus the Trellis import block; independent SOUND pass preceded this record
superseded_in_part_by: [decision-0058]
depends_on: [decision-0005, decision-0028, decision-0035, decision-0053]
informed_by: [research-0010]
owner: agent
date: 2026-07-23
---

# 0057 — `AGENTS.md` is Trellis's canonical shared project-instruction entrypoint

> **Human direction (2026-07-23).** The maintainer asked Trellis to adopt the
> same shared-entrypoint pattern being prepared in Grove, then supplied the
> exact working-copy reference. They clarified that the change must tell Claude
> to edit shared rules in `AGENTS.md`, not merely make both hosts load the same
> bytes. They then explicitly required the existing Trellis managed import
> block to remain in `CLAUDE.md`. This direction is the intent act recorded by
> the `approved` status after the revised decision received an independent
> `SOUND` pass.

## Context

Trellis's Layer-B build methodology and its self-applied Trellis import block
currently live together in `CLAUDE.md`. Codex discovers `AGENTS.md`, so opening
the repo in both hosts has produced a second, untracked copy. Identical copies
load the same initial text but create two apparent edit locations and can drift
immediately.

Claude Code can import `AGENTS.md` from `CLAUDE.md`; Codex discovers
`AGENTS.md` directly. The shared Layer-B method can therefore have one native
source. Trellis's managed block is different: it is the existing Claude import
channel from `decision-0053`, and the maintainer explicitly requires that
live-row delivery to remain in `CLAUDE.md`. Codex delivery of the Trellis
overlay is separate plugin-hook work, not something this file split can claim
to solve.

## Decision

**1. `AGENTS.md` is the one shared authority.** Trellis's Layer-B operating
method and Grove routing block live in root `AGENTS.md`. Shared project
instructions are added to unmarked prose there. They are never copied into
`CLAUDE.md`.

**2. Claude is an explicit adapter and editor routing is load-bearing.**
Root `CLAUDE.md` starts with exactly one active:

```md
@AGENTS.md
```

and retains the existing Trellis managed import block below it. It contains no
duplicate Layer-B/Grove shared prose.

The imported `AGENTS.md` tells Claude and Codex that shared rules are edited in
`AGENTS.md`; genuinely Claude-only rules belong in `.claude/rules/`; project
choices owned by Grove or Trellis remain in their named configuration files.
This instruction is part of the guarded canonical state, not optional
documentation.

**3. Trellis self-application keeps the Claude import channel.** The managed
Trellis block remains in `CLAUDE.md`, byte-identical to
`block-claude.md`: it imports `.trellis/internal/trellis.md` and the live
`.trellis/rules.toml`. It does not move to `AGENTS.md`, and its live-row
semantics do not change. Codex receives the shared Layer-B method from
`AGENTS.md`; delivery of the separate Trellis overlay to Codex remains
out of scope for this decision.

**4. Self-application drift fails in CI.** The repo sync guard asserts all of:

- `CLAUDE.md` contains exactly one active `@AGENTS.md` adapter import;
- `AGENTS.md` contains the editor-routing instruction;
- no Trellis managed block exists in `AGENTS.md`;
- exactly one Trellis managed block remains in `CLAUDE.md`; and
- that block is byte-identical to the import-channel payload.

Instruction-file and `.trellis/` changes trigger that guard in `cli-ci`.

**5. This is a repo shared-entrypoint change, not a new Trellis delivery
contract.** Setup/refresh continues to find and refresh the existing import
block in `CLAUDE.md`. General migration of arbitrary consumer pairs and
plugin-native Codex delivery remain separate product work; this decision does
not silently broaden either contract.

**6. Earlier location claims are superseded in part.** `decision-0005`'s
Layer-B classification, `decision-0028`'s where-it-fires derivative rule, and
`decision-0035`'s Build/Govern/Method split stand. Their claims that the
physical Layer-B Method and derived-resource instruction home is `CLAUDE.md`
are superseded by points 1–2 and 4. `decision-0035`'s Trellis
self-application/import-block location in `CLAUDE.md` stands. All three
records receive append-only forward pointers that state this exact boundary.

## Consequences

- Claude and Codex start from the same shared Layer-B project method, with one
  authoritative place to edit it.
- `CLAUDE.md` becomes a host adapter plus Trellis's existing Claude-specific
  import delivery point, rather than a second shared-policy surface.
- Claude's live `.trellis/rules.toml` semantics remain unchanged.
- This PR does not claim that Codex receives the Trellis overlay; that remains
  the separate plugin-hook delivery line.
- Current-truth references and Grove consumer configuration move from
  `CLAUDE.md` to `AGENTS.md`; decisions 0005, 0028, and 0035 retain their
  original wording with forward pointers for the superseded physical location.
- Consumer-wide entrypoint migration remains visible follow-up work rather than
  being smuggled into a repo-local cleanup.

## Open questions

- Should the later consumer-wide contract reuse Grove's deterministic
  two-entrypoint helper, or keep Trellis's no-runtime, agent-instruction
  composition model (`decision-0010`)?

## Self-check

The decision records the maintainer's stated target and exact reference rather
than inferring intent from an untracked copy. It preserves the managed
Trellis-import block exactly where the maintainer required and refuses to
pretend that an `AGENTS.md` split delivers its nested imports to Codex. The
scope line keeps a repo-local fix from silently changing every consumer's
install contract. The three older decisions are narrowed only at the shared
Layer-B physical-entrypoint claim; their architecture stands.

---

> **Superseded in part (2026-07-24, append-only pointer).**
> `decision-0058` adds one distinct, small Trellis-managed Codex
> receipt/fallback block to `AGENTS.md`; the prohibition on putting the full
> Trellis rules or Claude import block there stands. `AGENTS.md` remains the
> shared-prose authority, and Claude's existing import block remains in
> `CLAUDE.md`.

# One gateway for every payload read — design

**Date:** 2026-09-03
**Tracks:** TRL-33 (the umbrella and its preferred remedy), TRL-34 (folds in)
**Produced by:** `superpowers:brainstorming`, architectural path. A superpowers design doc,
retained per `decision-0085`; not the retired `spec-*` artifact type (`decision-0079`).

## Problem

Across the work that produced `decision-0083` and `decision-0084`, **eleven defects shared one
shape**: an absent, empty, truncated or unreadable *payload* input reached downstream logic and
the session ran ungoverned at `exit 0` with nothing signalling a problem. Each was fixed
individually, where it was found. Two of the eleven were the **inverse** — a guard that
over-corrected and refused a *healthy* payload (a CRLF-terminated `rules.md`; an unreadable
`reference/rules-b.toml` reported as payload incoherence).

**Almost none was found by the test suite or by reading the code.** Every one was found by a
reviewer running the hook against a deliberately broken input. That recurrence is the finding.

Four instances are still live on `main`. All four were reproduced by hand before this design was
written, with the harness at `scratchpad/repro.sh` (a vendored-bundle project plus a real payload,
running the real `staleness.sh`):

| Instance | Break | Measured today |
|---|---|---|
| **TRL-33** | `reference/rules-b.toml` absent, vendored-defaults path | `exit 0`, **0 bytes stdout**, 0 bytes stderr — a completely ungoverned session with zero signal |
| **TRL-34** | `reference/version` mode 000, vendored-overlay path (A) | `exit 0`, **0 bytes stdout** — the staleness warning is withheld with no explanation |
| new | `.trellis/internal/version` zero bytes, path A | `exit 0`, **0 bytes stdout** — while a *missing* stamp on the same path refuses loudly |
| new | `.trellis/internal/rules.md` mode 000, path A | `exit 0`, `"Trellis overlay may be stale… Until then this session is governed by the vendored copy."` — **false**: the host's import of that file fails, so the session is governed by nothing. The `-s` guard beside it catches missing-or-empty and not unreadable. |

The inconsistency TRL-33 names is visible in the first row's sibling: `reference/rules-b.toml` at
mode 000 **is** caught, loudly — but by the row validator two hundred lines downstream, whose
message says *"this project's `.trellis/rules.toml`… exists but could not be read"* to a project
that **has no `.trellis/rules.toml`**. Right marker, wrong file named.

## What "systematic" has to mean

Not a twelfth patch. The deliverable is **one mechanism every payload read goes through**, such
that:

1. a read that fails, returns nothing, or returns unusable content produces a **loud refusal**;
2. adding a *new* payload read later is **guarded by where you write it**, not by remembering;
3. the guard **discriminates** — a healthy payload, including its legitimate variants, must
   never be refused.

Requirement 3 is not decoration. Refusing a healthy input is as bad for a consumer as governing
a broken one: they see `TRELLIS_RULES_NOT_LOADED` with nothing to fix. It has happened twice.

## Scope: what is a *payload* read

**Payload** = a file that ships in the Trellis bundle, whether it is read from the installed
plugin or from a vendored copy inside the consuming repository. Its absence or corruption is
always a broken install, never a legitimate project state.

**Project** = a file the consuming repository owns. Absence and emptiness are legitimate states
with defined meanings, and the hook must keep treating them as such.

The two classes need opposite defaults, which is why one gateway cannot serve both.

### Every payload read, enumerated

`plugins/trellis/hooks/staleness.sh` — line numbers as of `3f44620`:

| # | Path | Line | Today |
|---|---|---|---|
| S1 | `$plugin/reference/version` | 79 | `head -n1 … 2>/dev/null` → `""`; three consumers exit silently (327, 338) or skip a comparison (446) |
| S2 | `$plugin/reference/rules-b.toml` *as the rows*, defaults path | 529 | `[ -f "$toml" ] \|\| exit 0` — **TRL-33** |
| S3 | `$plugin/reference/trellis-{a,b}.md` (posture header) | 545-550, 1024 | `[ -f ]` → loud; `cat … 2>/dev/null` → loud on empty |
| S4 | `$plugin/reference/rules.md` | 550, 594, 622, 855, 1063 | `[ -f ]` → loud, then **four further opens** of the same path (two positional awk operands, two `getline` redirections) |
| S5 | `$plugin/reference/rules-b.toml` *as the coherence comparison* | 767 | `[ -f ]` → skip silently; deliberate, and pinned by a regression test |
| S6 | `$root/.trellis/internal/version` | 310, 325 | absent → loud; **empty → silent** |
| S7 | `$root/.trellis/internal/{trellis.md,rules.md}` | 317 | `[ ! -s ]` → loud; **unreadable passes** |
| S8 | `$root/.trellis/version` (legacy flat stamp) | 336 | **empty → silent** |

`plugins/trellis/hooks/codex-context.mjs`:

| # | Path | Line | Today |
|---|---|---|---|
| C1 | `<PLUGIN_ROOT>/.codex-plugin/plugin.json` | 66 | `try/catch` → `invalid-plugin-root`. **Loud already.** |
| C2 | prose / rules / version, plugin-native or vendored | 730-737 | `readRequired` → tagged `missing-file` / `unreadable-file` / `context-over-budget`. **Loud already.** |
| C3 | `<projectRoot>/.trellis/rules.toml` | 658 | `readRequired`; empty is legitimate here |

**Codex is already close to right, and `readRequired` is the model this design copies.** Its one
structural gap: a zero-byte file returns `{ value: "" }` — a *success*. Emptiness is caught only
by post-checks each caller remembers to write (`empty-prose` at 744/748, the version regex at
751). That is requirement 2 unmet: the next payload read added to that file inherits no guard.

Two message defects sit beside it: `fail(".trellis/internal/trellis.md", …)` at 756 and
`fail(".trellis/internal/rules.md", …)` at 770 hardcode the *vendored* path and so name the wrong
file on the plugin-native path.

## Design

### 1. The gateway — `staleness.sh`

One function, the only place in the file that opens a payload file. It **classifies; it does not
judge** — because two of the eleven defects were guards that judged healthy input.

```sh
payload_status=""   # ok | missing | unreadable | empty
payload_why=""      # the clause every refusal below splices in
payload_text=""     # the content — set only when status is ok

payload_read() { ... }
```

- `missing` — the path does not exist (or its symlink target is gone).
- `unreadable` — it exists but could not be opened: a permission mode, a stale ACL, or a
  directory where a file must be.
- `empty` — it opened and yielded nothing.
- `ok` — content is in `payload_text`.

`missing` and `unreadable` are reported *separately* because their remedies differ (reinstall vs.
fix the mode) — but neither is silent anywhere, which is exactly TRL-33's finding: the two were
handled differently for no reason at all.

`payload_read` returns 0 only for `ok`, so the shortest thing a caller can write —
`payload_read "$f" || { emit "…"; exit 0; }` — is also the safe thing.

### 2. Dispositions, per call site

| # | After |
|---|---|
| S1 | Gateway. On failure `current=""` and `stamp_defect` carries the reason. The stamp is then **shape-checked** (`payload@` + exactly 12 lowercase hex, matching `codex-context.mjs:751`): a *truncated* stamp is not a different version, it is an unreadable one, and comparing it reports a healthy overlay as stale. |
| S2 | Gateway, **loud refusal** — `TRELLIS_RULES_NOT_LOADED`, naming `rules-b.toml` and why. **TRL-33.** |
| S3 | Gateway; existing wording kept, gaining the missing/unreadable/empty distinction. |
| S4 | Gateway once. The four later opens keep their present form: the gateway has already proven the path readable and non-empty, so the fatal-and-silent shape can no longer arise there. |
| S5 | Gateway; **still skipped silently on any failure** — unchanged, and pinned by the regression test that exists because this check once over-corrected. |
| S6, S7 | Gateway, all three overlay files, uniformly. Missing, empty *and* unreadable now draw the same loud refusal that missing alone drew. |
| S8 | Gateway; an unusable legacy stamp still draws the migration nudge, saying the stamp could not be read instead of vanishing. |

**S1's three consumers**, per TRL-34's *"produces the warning it exists to produce, or says why
it cannot"*:

- Path A and the legacy path: a new marker **`TRELLIS_STALENESS_UNKNOWN`** — the plugin's own
  stamp is unusable, so drift cannot be checked. The session **is still governed** by the
  overlay, and the message says so. Not `TRELLIS_RULES_NOT_LOADED`: nothing is ungoverned, and
  reusing the blackout marker would be the over-correction this design is guarding against.
- Path C: the existing stand-down message gains a variant literal saying staleness could not be
  checked.
- Path B: delivery is unaffected — the rules are fine. The trailer states that the stamp could
  not be read, and the quarantine note substitutes a stable phrase for the empty stamp instead
  of writing `not in .` into the consumer's file.

**Known divergence, recorded rather than closed:** on the plugin-native path Codex *refuses* a
malformed `reference/version` (`invalid-version`), where Claude will now govern and annotate.
Claude's is the behaviour TRL-34 asks for; Codex is the stricter side. Left as an open question
in the decision record rather than changed here — `codex-context.mjs` has a queued change behind
this one.

### 3. `codex-context.mjs` — make the default safe, change no behaviour

`readRequired(projectRoot, relativePath, options)` gains a third, **required-by-convention**
argument. After a successful read of zero bytes it returns `{ error: options.emptyError }`, or
`{ value }` when `options.emptyIsValid` is set. Call sites:

| Call | Option | Resulting class |
|---|---|---|
| prose | `{ emptyError: "empty-prose" }` | unchanged |
| rules | `{ emptyError: "empty-prose" }` | unchanged |
| version | `{ emptyError: "invalid-version" }` | unchanged |
| `.trellis/rules.toml` | `{ emptyIsValid: true }` | unchanged — empty is the supported hand-written-partial shape |
| a future read that passes nothing | *default* | `empty-file`, loud |

**Every existing failure class is preserved byte-for-byte**, so this is a structural change with
no observable behaviour change. What changes is the *default*: emptiness is now refused unless a
call site says otherwise, in its own source, where a reader sees it.

Plus the two hardcoded labels at 756 and 770 become `sources.prose` / `sources.rules`. On the
vendored fixture those resolve to the same strings the tests assert.

### 4. What makes it stick — two guards, one behavioural and one structural

**The behavioural guard is the primary deliverable.** `TestBrokenPayloadIsNeverSilent`: for every
file in `plugins/trellis/reference/` — **enumerated from the directory, so a payload file added
later joins the matrix without anyone remembering** — crossed with every break shape
(`absent`, `zero-byte`, `mode-000`, `truncated-to-half`) and both delivering project shapes
(config-only, vendored-defaults), assert the hook's output is **either a complete governed
injection or a message carrying a `TRELLIS_` marker** — never empty, never an activation list
with no rules under it.

Its healthy half runs in the same table: a `healthy` row and a `crlf` row (every payload file
converted to CRLF) must produce a complete injection **with no marker**. That is requirement 3,
measured, on exactly the shape that produced one of the two over-corrections.

**The structural guard** is `TestNoPayloadReadBypassesTheGateway`: no `$plugin/`-derived path may
be a file operand outside `payload_read`, and every payload path variable must reach
`payload_read`. Its Codex half asserts every `readRequired(` call site passes a third argument.
Source-scanning guards have precedent here — the destructive-verb scans in
`cli/plugin_hook_test.go` work the same way.

### 5. Tests that change, and why

| Test | Change |
|---|---|
| `TestStalenessHook`/"unreadable plugin reference is silent" (`:240`) | Inverted. It **pins TRL-34's defect** — it asserts the silence the issue exists to end. Rewritten to require `TRELLIS_STALENESS_UNKNOWN`. |
| `TestStalenessHook`/"empty stamp is silent" (`:232`) | Split. The `.trellis/internal/version` half is inverted (an empty overlay stamp is a broken overlay, and its two sibling files already refuse loudly for exactly this). The legacy-flat half becomes "still nudges, without a stamp value". |
| `TestEveryLegitimateShapeStillGoverns` corrupted-defaults cases (`:3956`) | Gains an `absent` state — TRL-33's shape, beside the three already there. Wording asserted is unchanged. |

Every other existing assertion is expected to survive untouched. Any that does not is a
behaviour change to justify or back out, not a test to edit.

## Non-goals

- **The hook still never writes** (`decision-0070` D4). No path here writes anything.
- **Quarantine still never deletes.** No row's value is lost on any path.
- **Every exit stays zero.**
- Project reads keep their present semantics. `.trellis/rules.toml`, `CLAUDE.md`, `AGENTS.md`,
  `.claude/rules/trellis.md` are not payload; an absent one is a legitimate project state.
  The one exception is `.trellis/rules.toml` on the defaults path, where `$toml` *is* the
  payload — the dual nature that hid TRL-33 — and there it goes through the gateway.
- No change to the reconciler, the coherence check's semantics, or the injection budget.

## Risks

- **Over-refusal is the failure mode to watch**, and it has landed twice before. Mitigated by
  the matrix's healthy and CRLF rows, and by leaving S5's silent skip exactly as it is.
- **A new marker is a contract change.** `TRELLIS_STALENESS_UNKNOWN` is agent-facing text; it is
  recorded in the decision, not introduced quietly.
- **`codex-context.mjs` has a queued change behind this one.** Its diff is kept to the smallest
  shape that satisfies requirement 2 — one function signature, four call sites, two labels.

---
id: decision-0086
type: decision
depends_on: [decision-0070, decision-0083, decision-0084]
informed_by: [decision-0010, decision-0028, decision-0035, decision-0051, decision-0073, decision-0081, decision-0082, decision-0085]
owner: agent
date: 2026-09-03
---

> **Provenance.** Directed by the maintainer as **TRL-33**, with **TRL-34** folded in. Designed
> through `superpowers:brainstorming` and executed through `superpowers:executing-plans`; the
> design and plan are retained under `docs/superpowers/` per `decision-0085`.

# 0086 — one gateway for every payload read; a defect class closed by construction, not by patch twelve

## Context

`decision-0083` and `decision-0084` shipped eleven fixes that shared one shape:

> **An absent, empty, truncated or unreadable *payload* input reaches downstream logic, and the
> session runs ungoverned at `exit 0` with nothing signalling a problem.**

Two of the eleven were the **inverse** — a guard that over-corrected and refused a *healthy*
payload. A CRLF-terminated `rules.md` was reported as truncated; an unreadable
`reference/rules-b.toml` was reported as payload incoherence while `rules.md` and the project's
rows were both perfectly well. A consumer who sees `TRELLIS_RULES_NOT_LOADED` with nothing wrong
to fix is served no better than one governed by a broken payload.

**The recurrence is the finding, and how each was found is the argument.** Almost none was found
by the test suite or by reading the code. Every one was found by a reviewer *running* the hook
against a deliberately broken input, one file at a time, as they thought of it. Two more were
filed rather than fixed (`TRL-33`, `TRL-34`); this change found two further live instances while
enumerating, so the count at the point of writing was **fifteen**, not eleven.

### What was still live, measured before anything was changed

Reproduced against `main` with a vendored-bundle project, a real payload, and the real
`plugins/trellis/hooks/staleness.sh`:

| Break | Measured |
|---|---|
| `reference/rules-b.toml` **absent**, vendored-defaults path | `exit 0`, **0 bytes stdout**, 0 bytes stderr. Ungoverned, zero signal. (`TRL-33`) |
| `reference/version` **mode 000**, vendored-overlay path | `exit 0`, **0 bytes stdout**. The staleness warning withheld with no explanation. (`TRL-34`) |
| `.trellis/internal/version` **zero bytes** | `exit 0`, **0 bytes stdout** — while a *missing* stamp on the same path refused loudly |
| `.trellis/internal/rules.md` **mode 000** | `exit 0`, *"Trellis overlay may be stale… Until then this session is governed by the vendored copy."* — **false.** The host's import of that file fails, so nothing governs |

The inconsistency `TRL-33` names is sharpest in the first row's sibling: the same file at mode 000
**is** caught, loudly — but two hundred lines downstream, by a message that tells a project with
**no `.trellis/rules.toml`** that *"this project's `.trellis/rules.toml`… could not be read"*.
Right marker, wrong file named, and only reachable by accident.

## Decision

**1. One gateway, and it is the only thing that opens a payload file.** `staleness.sh` gains
`payload_read`, which classifies a read as `ok` / `missing` / `unreadable` / `empty` and returns
zero only for `ok` — so the shortest thing a caller can write, `payload_read "$f" || { emit …;
exit 0; }`, is also the safe thing. Every plugin-side and vendored-overlay read goes through it.

**2. It classifies; it does not judge.** The gateway reports *why* a read failed and the call site
decides what that costs. This is written the way it is *because* two of the fifteen were the
inverse defect: a gateway that decided for its callers would have to pick one severity for every
payload file, and the severities genuinely differ.

**3. A payload file and a project file are different classes, and only the payload class is
guarded here.** A payload file ships in the bundle; its absence is always a broken install. A
project file (`.trellis/rules.toml`, `CLAUDE.md`, `.claude/rules/trellis.md`) is the consumer's,
and absent or empty are legitimate states with defined meanings that this change does not touch.
The one file that is *both* — `.trellis/rules.toml` when `rows_are_default=yes` repoints `$toml`
at the payload's own preset — is where `TRL-33` hid, and it goes through the gateway.

**4. `missing` and `unreadable` are told apart in the message and treated alike in the
disposition.** Their remedies differ (reinstall vs. fix the mode), so a reader is told which. What
neither is, anywhere, is silent — which is `TRL-33`'s whole finding: nothing ever chose to handle
them differently.

**5. A withheld *warning* is not a blackout, and gets its own marker.** `TRELLIS_STALENESS_UNKNOWN`
is new agent-facing contract. It fires when the plugin's own `reference/version` cannot be read or
is malformed: the session is still governed — by the vendored overlay, or by the rendered file —
and the hook says only that it could not check for drift. Reusing `TRELLIS_RULES_NOT_LOADED` there
would have been the over-correction this record is half about.

**6. The plugin's version stamp is shape-checked**, `payload@` plus exactly twelve lowercase hex,
matching what `codex-context.mjs` has always required of the same file. A *truncated* stamp is not
a different version — it is an unreadable one — and comparing it reported a healthy overlay as
stale.

**7. On Codex, the default becomes loud without any behaviour changing.** `readRequired` — which
has been the model for all of this, reporting `missing-file` and `unreadable-file` where the shell
hook swallowed both — returned `{ value: "" }` for a zero-byte file, a *success*, with emptiness
caught only by post-checks each caller remembered to write. It now takes a third argument saying
what an empty read means. Every existing failure class is preserved byte for byte; **no Codex
assertion in the suite needed editing, which is the evidence for that claim rather than the claim
itself.**

**8. Two guards hold it, and the behavioural one is primary.**

- `TestBrokenPayloadIsNeverSilent` crosses **every file in `plugins/trellis/reference/`, read from
  the directory rather than hardcoded**, with `absent` × `zero-byte` × `unreadable` × `truncated`,
  on both delivering paths — 180 cells. The invariant is *either a complete governed injection or a
  loud `TRELLIS_` marker*: never silence, and never an activation list with no rules under it.
  Because the file list comes from the bundle, **a payload file added later joins the matrix
  without anyone remembering.** Its `healthy` and `crlf` rows run in the same table, so the
  over-refusal direction is measured by the same guard.
- `TestNoPayloadReadBypassesTheGateway` and `TestEveryReadRequiredStatesWhatEmptyMeans` are the
  structural half: nothing opens a payload path behind the gateway's back, and no `readRequired`
  call site stays silent about emptiness. Source-scanning guards have precedent here — the
  destructive-verb scans work the same way.

**9. Both were mutation-proved, in both directions.** Reintroducing `[ -f "$toml" ] || exit 0`
makes the matrix report `SILENT` on `rules-b.toml/absent/vendored-defaults`; removing the
`sub(/\r$/)` from the sentinel gate makes `rules.md/crlf` fail as an over-correction. A new
unguarded payload variable, an inline literal read, and a `readRequired` call with no empty
disposition each turn the covering structural guard red.

## Consequences

- **`TRL-33` and `TRL-34` close together**, as `TRL-34` predicted they would if the systematic
  remedy landed.
- **Two subtests were inverted because they pinned the defect.** `TestStalenessHook`/"unreadable
  plugin reference is silent" asserted `TRL-34`'s silence outright; the `.trellis/internal/version`
  half of "empty stamp is silent" asserted the third row of the table above. Both carry a comment
  saying what they used to assert and why that was wrong. **Nothing else in the suite needed an
  edit** — the full run went from 2 failures to green with no other assertion touched, which is
  the containment evidence for a change this wide.
- **`TRELLIS_STALENESS_UNKNOWN` is a new marker in the agent-facing vocabulary.** Recorded here
  rather than introduced quietly, because a marker is contract.
- **A `VERSION` bump ships with it** (0.8.0 → 0.9.0) so cached consumers re-pull, and the baked
  bundle manifest in `install.sh` advances for the four files that changed (`decision-0028`: a
  source and its derivative move together).
- **What is not claimed:** that this is the last instance. The claim is narrower and checkable —
  that the *next* one is caught by a guard rather than by whoever thinks to try it, and that a
  payload file added later is covered without anyone deciding to cover it.

## Open questions

- **The two hosts now disagree about a malformed `reference/version`, and neither is obviously
  wrong.** On the plugin-native path `codex-context.mjs` refuses outright (`invalid-version`);
  `staleness.sh` now governs and annotates. `TRL-34`'s framing — *"the session stays governed, only
  a warning is withheld"* — is the reason for the Claude behaviour, and by that reading Codex is
  the over-corrected side. Not changed here: `codex-context.mjs` has a queued change behind this
  one, and aligning the two is a decision about which reading is right rather than a bug fix.
  Named per `decision-0078`: the consumer that will re-present it is the next Codex change.
- **The vendored overlay's `trellis.md` and `rules.md` are checked for readability by the hook but
  imported by the *host*.** The hook can now tell the reader the import will fail; it still cannot
  tell whether the host succeeded. That gap is inherent to the transport, not to this change.

## Self-check

- **The measurements in Context are reproductions, not inferences.** Each row was produced by
  running the shipped hook against a broken fixture before any code was changed, and each was
  re-run afterwards to confirm the new behaviour.
- **The claim in Decision 7 — "no behaviour change" — is the one most likely to be wrong**, since
  it is a claim about a file this change edits. It rests on a negative that the suite can check:
  the Codex failure-vocabulary test asserts exact stdout bytes for eighteen classes and required no
  edit. If a class had moved, that test would have said so.
- **This record was written by the agent that made the change**, and the guard against that is
  Decision 9: every assertion added here was broken deliberately and observed to fail with the
  expected symptom. A guard that has never failed is not known to work.

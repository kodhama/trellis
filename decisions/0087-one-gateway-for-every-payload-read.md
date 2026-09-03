---
id: decision-0087
type: decision
depends_on: [decision-0043, decision-0051, decision-0070, decision-0083, decision-0084]
changes: [decision-0043]
informed_by: [decision-0010, decision-0028, decision-0035, decision-0065, decision-0073, decision-0081, decision-0082, decision-0085]
owner: agent
date: 2026-09-03
---

> **Provenance.** Directed by the maintainer as **TRL-33**, with **TRL-34** folded in. Designed
> through `superpowers:brainstorming` and executed through `superpowers:executing-plans`; the
> design and plan are retained under `docs/superpowers/` per `decision-0085`.
>
> **Numbered 0087, not 0086.** This record was drafted as `0086` — `0085` was the highest on
> `main` when the branch opened, verified against `origin/main` rather than asserted — and
> renumbered before merge because `kodhama/trellis#263` (TRL-29) claims `0086` and lands first.
> `decisions/` numbers are allocated by whichever branch merges first, not by whichever drafts
> first; nothing else in this record moved.
>
> **`decision-0078` recorded this same mechanism and armed a trigger on it**
> (`decisions/0078-no-orphan-followups.md:146-155`): *"nothing catches two branches claiming one
> id… if it recurs a third time, that is the trigger to file it."* The two it counted were
> `0077` and trellis#252's `0076`. **This is the third**, so the trigger has fired and the
> obligation is discharged rather than noted: filed as
> [TRL-40](https://linear.app/kodhama/issue/TRL-40). Found by a `corpus-reviewer` pass, not by the
> author who was standing in the collision.

# 0087 — one gateway for every payload read; a defect class closed by construction, not by patch twelve

## Context

`decision-0083` and `decision-0084` shipped a run of fixes that shared one shape:

> **An absent, empty, truncated or unreadable *payload* input reaches downstream logic, and the
> session runs ungoverned at `exit 0` with nothing signalling a problem.**

**The count, stated as this record's own measurement on 2026-09-03 rather than attributed to
either predecessor**, because neither states a total and a count in a record is a measurement with
a date (`decision-0083`'s own discipline). What they say is: *"That branch turned out to be **one
of seven**, not the only one. Review of this branch, before merge, found six more"*
(`decisions/0083-rules-toml-reconciles-itself.md:87-88`), then *"An **eighth** instance of the same
class is fixed but is NOT in the table"* (`:127-128`) — *line numbers as of this record; `0083`'s
frontmatter has grown since it was written, and an earlier draft of this paragraph cited the
pre-growth positions. The quoted text is the citation; the numbers are a convenience that decays.*
`decision-0084` adds further instances in
its §3 and §5 without totalling them. Adding the two filed-but-unfixed (`TRL-33`, `TRL-34`) and the
two this change found while enumerating gives **fifteen at the point of writing**. Any reader who
counts differently should trust their own count over this sentence — the number is not what the
argument rests on.

**The recurrence is the finding, and how each was found is the argument.** Almost none was found
by the test suite or by reading the code. Every one was found by a reviewer *running* the hook
against a deliberately broken input, one file at a time, as they thought of it.

**Some were the inverse** — a guard that over-corrected and refused a *healthy* payload. An
unreadable `reference/rules-b.toml` was reported as payload incoherence while `rules.md` and the
project's rows were both perfectly well (`decision-0083` records the general property at
`:115-116`; the specific fix is in the hook). A CRLF-terminated `rules.md` was reported as
truncated — **that one is recorded in the hook rather than in either decision**, at
`plugins/trellis/hooks/staleness.sh:796-804` (*"an exact ASCII comparison fails — reporting
`not-last` and blacking out a COMPLETE, CORRECT payload"*), and pinned by
`TestTruncatedRulesMdIsRefusedByItsOwnTerminator`/"a healthy CRLF payload governs and is not
refused". Cited to its actual source, because a corpus review of an earlier draft of this record
correctly found it attributed to two records that do not say it. A consumer who sees
`TRELLIS_RULES_NOT_LOADED` with nothing wrong to fix is served no better than one governed by a
broken payload.

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
stale. This narrows `decision-0043` rule 2's `payload@<content-hash>` without contradicting it:
every stamp the payload generator has ever written is twelve hex.

**6a. This changes `decision-0043` rule 3, and the forward pointer says exactly which clause.**
That rule reads, verbatim at `decisions/0043-generator-only-cli-and-payload-stamp-staleness.md:54-56`:

> **Staleness is a file-to-file compare; `trellis status` retires.** `hooks/staleness.sh` compares
> `.trellis/version` against the installed plugin's `reference/version`: warn on mismatch,
> **no-op when either side is missing or empty.**

**The no-op half is what this record changes**, on all **three** branches that carried it — missing
and empty are inside "cannot be read", and each now says something instead of nothing:

| Branch | Stamp | Was | Is |
|---|---|---|---|
| Path A, the plugin's stamp | `$plugin/reference/version` | silent | `TRELLIS_STALENESS_UNKNOWN` — still governed |
| Path C, the plugin's stamp | `$plugin/reference/version` | silent | `TRELLIS_STALENESS_UNKNOWN` — still governed |
| Path A, the project's stamp | `.trellis/internal/version` | silent when empty | `TRELLIS_RULES_NOT_LOADED` — the overlay is broken |
| **The legacy flat path** | **`.trellis/version`** | **silent when either stamp was unreadable** | **the migration nudge, saying which stamp it could not read — no `TRELLIS_` marker, because nothing is ungoverned and nothing is stale-by-comparison; the LAYOUT is what is stale, and that is true without either stamp** |

**The fourth row is the one a reader most needs, and an earlier draft of this record omitted it** —
found by a corpus review, which also named why it matters: **`.trellis/version` is the path
`decision-0043` rule 3 literally names.** The comparison moved to `.trellis/internal/version` under
`decision-0051`, whose own record carries no supersession pointer at `0043`, so every statement
about "rule 3's stamp" since then has silently meant the post-`0051` path. This record does too,
and now says so rather than leaving the reader to notice.

The compare-and-warn half stands untouched, and so does the rest of `0043` — **except the parts
already reached by `decision-0059`, `decision-0061` and `decision-0065`**, which are visible in the
same frontmatter field. `decision-0065` in particular already narrowed rules 2-3 to vendored
projects. An earlier draft said "everything else in `0043`" flat, which read as a claim that
nothing else had ever been changed; it had.

Two honesties about the scope:

- **Half of the departure predates this record.** A *missing* `.trellis/internal/version` has
  refused loudly on `main` since before this branch — `decision-0043` rule 3's no-op had already
  been left behind on that side, with no pointer. This record marks the whole departure rather
  than only its own half.
- **The `TRL-34` case is not covered by rule 3 at all.** Mode 000 is neither missing nor empty, so
  a stamp that exists and cannot be opened is a gap `0043` never addressed, not a clause this
  record overrides.

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
  **It scopes to two of `decision-0073` D1's delivery states, and says so where it does it**
  (D1's per-component relevance clause): config-only and vendored-defaults are the two states on
  which this hook *delivers*, so only there is "silence is wrong" unconditional — a current
  vendored overlay is legitimately silent, and path A is covered by named subtests instead.
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
- **A `VERSION` bump ships with it** (0.8.0 → 0.9.0) so cached consumers re-pull, with **both
  plugin manifests** (`.claude-plugin/plugin.json`, `.codex-plugin/plugin.json`) and `install.sh`'s
  baked bundle manifest advanced in the same commit (`decision-0028`: a source and its derivative
  move together).
- **`decision-0043` gains `superseded_in_part_by: [decision-0087]`**, scoped in its trailing
  comment to rule 3's no-op clause and to nothing else. Under `decision-0082` the forward pointer
  is the only mark there is.
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
  **Named per `decision-0078`, whose test is a consumer with a cadence: filed as
  [TRL-39](https://linear.app/kodhama/issue/TRL-39/the-two-hosts-disagree-about-a-malformed-referenceversion-codex).**
  An earlier draft of this record named *"the next Codex change"*, which `decision-0078` explicitly
  rules out — corrected after a corpus review said so.
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
- **The `corpus-reviewer` returned FAIL on this record's first draft, and it was right on every
  count that mattered.** It found an undeclared change to a live clause of `decision-0043` (now
  Decision 6a and a forward pointer), `decision-0051` demoted to `informed_by` where
  `decision-0083` — the immediate predecessor, on the same paths — has it as `depends_on` (now
  moved), a count of "eleven" attributed to two records that state no total (now this record's own
  measurement, with what they actually say quoted), a CRLF instance attributed to records that do
  not contain it (now cited to the hook and its test), and an open question with no consumer that
  meets `decision-0078`'s own test (now TRL-39). Each finding was re-verified against the cited
  source before being acted on; none was accepted on the reviewer's say-so. Recorded here because
  a record that hid its own review would be arguing against its own thesis.
- **The re-review of the corrected draft found three more, and they are the interesting ones.**
  Two line citations to `decision-0083` were off by one and two lines (`0083`'s frontmatter had
  grown since the numbers were taken); the `changes:` scope named two dispositions where the code
  has **four**, omitting the legacy flat path — *the path rule 3 literally names*; and the
  corrected CRLF attribution had **not reached the two derivatives this record itself points at**,
  so `staleness.sh` and `cli/plugin_hook_test.go` still carried the misattribution the record had
  just fixed. That last one is a `decision-0028` pair failing inside the change that cites
  `decision-0028` — a source and its derivative disagreeing, caught by a reviewer rather than by
  the author who wrote both.
- **A code review on the pull request found two more, and the second is the sharpest finding on
  this branch.** A P2 showed the version-stamp read was `head -n1 | tr -d '[:space:]'`, so a valid
  stamp followed by garbage, and a stamp with a space inside it, were both accepted as
  authoritative at zero bytes of output — the defect class this record is about, reached through a
  door this record's own author left open. And the structural guard's payload-path regex was
  anchored at line start, so `firm) header="$plugin/…"` — a real payload path in a `case` arm — was
  invisible to it, while its count floor passed anyway because an unrelated variable filled the
  slot. **The mutation used to prove that guard added a variable at line start: the shape the
  regex already handled.** A guard is known to work only against the mutations actually tried,
  which is this record's own thesis arriving one level up and landing on it.

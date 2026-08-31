---
id: decision-0084
type: decision
depends_on: [decision-0072, decision-0083]
informed_by: [decision-0028, decision-0070, decision-0074, decision-0078, decision-0081, decision-0082]
owner: agent
date: 2026-08-30
---

> **Provenance.** Directed by the maintainer as **TRL-30**, the open Codex arm `decision-0083` left
> behind, with **TRL-29** (the fail-closed byte cap) folded in. Designed through
> `superpowers:brainstorming` and executed through `superpowers:executing-plans` on
> `feature/trl-30-codex-reconciliation-parity`, stacked on the TRL-20 branch (PR #258).
>
> **This record is written after the code shipped and reports what shipped, not what was planned.**
> Three of its load-bearing sections exist only because execution found them: the
> `!rulesSectionSeen` correction (the plan said keep it fatal, and the plan was wrong), the
> empty-slug-set guard that was a *repeat* of the Claude side's own Critical defect, and the
> host-neutral provenance ruling. Where execution overturned the design, the record follows
> execution and says so.

# 0084 — Codex reaches reconciliation parity; `parseRulesToml` becomes a classifier

## Context

**`decision-0083` retired the blackout on one host and said so.** Its §1 recorded that the design's
parity claim — *"one table, applied identically by both hosts"* — did not ship, and its open
questions named the remainder:

> **Host parity is owed, and the Codex arm of TRL-20 is therefore still open — `TRL-30`.** … A
> Codex project with a mismatched row set still gets `invalid-rules` and no rules — the exact
> blackout this record retires on Claude — so the defect TRL-20 names is half-closed, not closed.

The deferral was deliberate and its reason was structural: **Codex refuses in a parser, not in a
delivery branch.** `codex-context.mjs`'s `parseRulesToml` returned `null` for a missing row, an
unknown row and a duplicate row exactly as it did for a malformed one, and the caller turned any
`null` into `fail(PROJECT_CONFIG, "invalid-rules")` with nothing injected. There was no branch to
change; the gate and the syntax check were the same `return null`.

**`decision-0083` also named the sharpest instance, reachable from its own headline example.** The
empty file it measured — sixteen rows added under a `[rules]` table, no `strictness` — is a file
**Claude governs from and Codex rejects**. So Claude's repair could produce a file that blacked out
the other host, and the mandate's promise that *"the file matches what governs"* was false for a
Codex reader.

## Decision

### 1. `parseRulesToml` becomes a classifier, not a gate — and the split is stated exactly

It now returns `{ rows, mismatch }` for any file it can make sense of, and `null` **only** for a
genuine syntax fault. The caller distinguishes the two: `mismatch !== null` reconciles, `null` still
fails closed.

| Condition | Before | Now |
|---|---|---|
| a slug in the payload with no row in the file (`missing:`) | `null` → `invalid-rules` | **reconciled** — row added `active = true` |
| a row whose slug the payload does not ship (`unknown:`) | `null` → `invalid-rules` | **reconciled** — quarantined |
| the same slug twice (`duplicate:`) | `null` → `invalid-rules` | **reconciled** — first kept, extras quarantined |
| no `[rules]` table at all | `null` → `invalid-rules` | **reconciled** — see §2 |
| `strictness` absent | `null` → `invalid-rules` | **not fatal** — the caller defaults it. Posture selection never depended on this parser at all: it tests the raw file for a `strictness = "firm"` line and falls to adaptive otherwise, exactly as `staleness.sh`'s `case "$strictness" in firm) … *) … esac` does. The gate this closes was in the validation, not in the posture |
| `strictness` present but neither `firm` nor `adaptive` | `null` | **still `null`** — a typo must not silently pick a posture |
| a malformed row — including one not shaped `(inv\|floor)-…` | `null` | **still `null`** |
| an unknown or duplicate top-level key, or a duplicate/foreign section | `null` | **still `null`** |
| `governed` with a non-boolean value | `null` | **still `null`** |
| the payload's own derived slug set is empty | reconciled (defect) | **`no-slugs-in-payload`, loud** — §3 |

**The risk runs in the permissive direction, and this record names it rather than leaving it to be
inferred.** Every row moved from the first column to the second is a file Codex used to refuse and
now governs from. Wrong in that direction, **a corrupt file silently governs** — the session runs,
the readout looks normal, and the corruption is expressed as rules rather than as a refusal. That is
why the four fatal rows above stayed fatal and were not swept along with the rest: the classifier
splits on *"is this file's slug set wrong"* versus *"is this file not the schema"*, and only the
first is reconcilable. A slug set can be repaired against the payload; a grammar fault cannot be
repaired against anything.

**One consequence of getting the boundary exactly right was measured, not reasoned.** The row regex
used to accept `[a-z][a-z-]*`, wider than `reconcileRows`' own `(?:inv|floor)-` row detection. A
line like `bogus-rule = { active = true }` was therefore classified `unknown` — triggering
reconciliation — and then not recognised as a row to quarantine, so it passed through uncommented,
the mismatch never cleared, and the hook re-reconciled it every session to no effect. The classifier
was narrowed to the reconciler's own grammar rather than the reconciler widened to the classifier's,
because a wider `rowLead` would also match `strictness` itself, with no state tracking to
disambiguate. Such a line is now a malformed row and fails closed.

### 2. `!rulesSectionSeen` became reconcilable, and the plan that said otherwise was wrong

The task brief said to keep a missing `[rules]` table fatal. **That was the controller's error, not
a reading of the code**, and it was corrected during execution.

**A missing `[rules]` section is not a syntax fault. It is a slug set that is entirely absent** —
which is `missing:` for every slug, reconcilable by definition. Keeping it fatal would have left
Codex failing closed on **exactly the hand-written partial file the Claude hook repairs**: a file
carrying only `strictness = "firm"` is the canonical shape `decision-0083` §5 measured, and
`staleness.sh` turns it into a full `[rules]` table plus all sixteen rows. A parity change that
refused that file would have shipped parity in name.

No special case was needed to get there. With no `[rules]` seen, the loop never enters the
row-matching branch, `rows` is empty, and every slug falls out as `missing:` on its own — which is
also what makes `reconcileRows`' own `if (!hasRules)` table insertion reachable at all.

### 3. The empty-slug-set guard — and that it was a *repeat*, which is the part worth recording

`slugsFromRules` derives the payload's slug set from `reference/rules.md` (`decision-0083` §6). A
payload that carries the sentinel but **no slug tags** yields an empty set. Fed to the reconciler,
every row in the project's file classifies as `unknown` and gets quarantined — **an ungoverned
session delivered at exit 0**, silently inverting the fail-loud rule this file otherwise holds.
Measured before the guard existed: a tag-less `rules.md` produced sixteen quarantine notes, context
delivered, exit 0.

The guard now runs **before** `parseRulesToml`, because the parser cannot tell *"this project's file
is wrong"* from *"the payload gave me nothing to check it against"*. The path is
`no-slugs-in-payload`, and it names the plugin, not the consumer's rows.

**This is the same defect the Claude hook had, and its own review caught it there first.**
`decision-0083` records it as that branch's one Critical finding — *"the branch's one Critical defect
(`no-slugs-in-payload` entering the reconciler with an empty want set, quarantining every legitimate
row and running ungoverned at exit 0)"* — and `staleness.sh:642`/`:679-682` grew a guard for it.

**The transferable lesson is not the bug; it is how the bug recurred.** This task ported a
*semantic* (reconcile instead of refuse) from one host to the other and did not port the *guard*
that semantic had grown. The reconciler was copied faithfully; the precondition that makes
reconciling safe was not, because it lives somewhere else in the file and is invisible from the
function being mirrored. **A semantic and its guards are one unit. Porting half of it reproduces the
original defect exactly, and the second occurrence is harder to find than the first** — it now looks
like a deliberate difference between two implementations rather than an oversight.

### 4. The provenance remedy goes host-neutral in **both** hooks

The quarantine comment written into `.trellis/rules.toml` used to read:

> `  # quarantined <date>: not in <stamp>. If a newer Trellis ships this slug, run
> ``claude plugin update trellis@kodhama`` and uncomment.`

It now reads, **identically on both hosts**:

> `  # quarantined <date>: not in <stamp>. If a newer Trellis release ships this slug, update the
> Trellis plugin and uncomment this row.`

**The reason is not the contradiction it resolved.** The Codex payload did contradict itself for one
review round — the mandate said Codex has no such command while the row comment told the reader to
run one — but that is a symptom. **The reason is that `.trellis/rules.toml` is one file read by both
hosts.** A row comment naming a Claude-only command is wrong for a Codex user opening that file, and
it is wrong *whichever hook wrote it*: a Claude session repairs a file a Codex session then reads.
The rule this settles, stated so it generalises past this string:

> **A host's own mandate text may name its own command. A comment written into the shared file may
> not.**

Both hooks changed together, in the same commit, because byte-parity between them is now enforced
(§5) and a one-sided fix would go red. The two surviving `claude plugin update trellis@kodhama`
mentions are both Claude speaking about itself in its own hook message — its loud
`no-slugs-in-payload` emit and its repair mandate — verified by grep across both hooks, not assumed.
Neither is a comment written into `.trellis/rules.toml`.

**This supersedes `decision-0083`'s recorded provenance wording**, and `decision-0083` carries a
forward pointer for it. That record's remedy prose named the Claude-only command as the repair
instruction; the string it described is no longer the string either hook writes.

### 5. Two implementations, one conformance guard — and the guard's expiry

`decision-0028` requires **"a sync guard per source↔derivative pair"**. `staleness.sh`'s awk
reconciler and `codex-context.mjs`'s `reconcileRows` are now such a pair: same semantics, two
languages, one output file. `TestBothHostsReconcileIdentically` runs both real hooks against the
same fixture and compares the reconciled row block **byte for byte** across seven fixtures — five
that genuinely reconcile (rename, indented `[rules]` plus a missing row, duplicate with a differing
value, no `[rules]` table at all, empty file) and two pass-through (already-quarantined,
missing `strictness`).

**The guard earned itself on its first run.** It found an **empty-file divergence**:
`"".split(/\r?\n/)` yields `[""]` — one phantom line — where awk reads a 0-byte file as zero
records, so the JS pushed a spurious leading blank line ahead of `[rules]` that `staleness.sh` did
not. Fixed on the Codex side and now held by its own fixture; reverting the fix goes red.

**A second divergence was found earlier on the same branch, and its provenance is worth being exact
about.** The **CRLF** case was caught by a *reviewer* running the awk against CRLF input by hand,
before this guard existed — not by the guard. With `RS = "\n"` a CRLF line arrives with its `\r`
still on `$0`, so `print "# " $0 note` emitted a bare CR **mid-line**, before the note; the JS's
`/\r?\n/` consumes the pair. **Here the awk was the wrong one**, and it was fixed to match the JS:
`staleness.sh` gained `{ sub(/\r$/, "") }` as its first rule, ahead of every other rule that reads
`$0`, pinned by `TestReconciliationStripsCRFromCRLFInput`. **The reference implementation is not
automatically the correct one** — each divergence was decided on the merits, not by seniority.

Neither was findable by reading either implementation alone: both are properties of the *pair*.
That is the case for the guard — and the CRLF one is also the case for its coverage, since it took
a human comparison to find what the fixture set did not yet reach. **That coverage gap is now
closed**, in the direction of more guard rather than a softer claim: `TestBothHostsReconcileIdentically`
carries a `CRLF line endings, plus a rename` fixture, built on the rename precisely so a row is
actually *quarantined* — `print "# " $0 note` is the only line the stray CR reaches, so a CRLF
fixture with nothing to quarantine would have been vacuous. Deleting `{ sub(/\r$/, "") }` from
`staleness.sh:876` turns that subtest red and leaves every LF subtest green. `decision-0083`'s
*"for LF and CRLF input"* therefore now names a guard that has a CRLF case inside it.

**The byte-identity claim is scoped, not absolute — one divergence class survives, and it is now
covered.** This qualifies both the sentence above and the same claim in `decision-0083`'s
`superseded_in_part_by` comment: byte identity holds **on the full-provenance path**, for LF and
CRLF input. **CR-only** input (classic-Mac line endings) is the exception. Both reconcilers read
such a file as a single line — awk's `RS` is `"\n"`, the JS splits on `/\r?\n/` — so both find no
rows, both classify all sixteen slugs as missing, and both append all sixteen; **both hosts deliver
and govern**. They then differ in two measured ways: `staleness.sh:876`'s `sub(/\r$/, "")` strips
the record's trailing CR while the JS splitter keeps it, and the sixteen-row append assembles to
9481 B, over `MAX_CONTEXT_BYTES`, so Codex silently takes §6's provenance-free path and omits the
`# added 16 row(s) below on <date>` header Claude writes. What diverges is the *text of the repair*,
not the governing set.

`TestCROnlyLineEndingsAreTheOneKnownDivergence` pins both differences, so closing either is a
deliberate act with a red test to update rather than a silent change. **A second, sharper guard was
added with it:** `codexReconciledRows` now asserts the Codex side stayed on the full-provenance path
before any byte comparison. Without it the first fixture to cross 9500 B — the `rename` fixture
already assembles to 8939 B, 561 B of headroom — would have failed with *"the two hosts reconciled
the same file differently"*, sending the next reader after a parity bug that is not there.

**The expiry is stated now rather than discovered later** (`decision-0074`). Two implementations
held in step by a byte-identity test is the right shape at two hosts: each hook is idiomatic in its
own runtime, and the guard is cheap and total. **At a third host it stops being right** — three
implementations pinned pairwise is where extraction into one shared implementation (with thin
per-host adapters) becomes the correct call, and this record is the place a future author should
find that said. Until then, the duplication is deliberate and guarded, not drift.

### 6. TRL-29 folded in: over budget, Codex degrades instead of refusing

**Shipping reconciliation alone would have rebuilt the blackout at the byte boundary.**
Reconciliation only *adds* bytes — added rows, quarantine notes, a repair mandate — and
`codex-context.mjs` refused outright above `MAX_CONTEXT_BYTES`, emitting `context-over-budget` and
injecting nothing. A reconciled Codex session on a badly mismatched file would have delivered
nothing, which is the exact failure this work exists to remove. The two could not ship separately.

`decision-0083` already established that the cap has no external provenance and that **Codex spills
rather than rejects** (<https://learn.chatgpt.com/docs/hooks>): over its own threshold it saves the
full text and hands the model a head-and-tail preview with the path. **The old refusal was therefore
strictly worse than the host's own behaviour** — a self-inflicted total loss replacing the host's
graceful degradation.

Over budget, the hook now re-reconciles **without provenance** and injects that instead. What
degrades is the *injected context*, never the repair: the added/quarantined decisions are identical
either way, so governance is unchanged, and the mandate switches to *"write the full-provenance
version of these rows, not the abbreviated ones shown above"* — **the file keeps the provenance the
session's context gave up**. That property is the whole point of the branch and is mutation-tested
directly: forcing the flag back to `false` makes the mandate say *"write exactly the rows shown
above"* and the test goes red.

A hard refusal survives underneath. **An earlier draft of this record called it a runaway guard
that could only be reached by a pathologically large file, with "nothing left to degrade." That was
false, and it concealed the limitation stated below.**

**The degradation is one-shot, because it is gated on `mismatch !== null`.** It runs only in a
session that had something to reconcile. The session *after* the repair has no mismatch — the file
the mandate asked for already carries every row plus the persisted quarantine comments — so the
branch is skipped and the refusal fires instead. Nothing about that file changes again, so the
refusal is **permanent**.

Measured against the real firm payload (`rules-a.toml` plus N foreign rows), reproduced
independently while fixing this record:

| N foreign rows | session 1 | the file the mandate produces | session 2 |
|---|---|---|---|
| 5–8 | degrades, delivers | 2.1–2.6 KB | 8806–9364 B — delivered |
| **≥ 9** | degrades, delivers (9129 B) | **2816 B** | **refuses, `context-over-budget`, nothing injected** |

At N = 9 `staleness.sh` delivers 9833 B of context from the identical 2816 B file and governs
normally. **So applying this branch's own mandate can black Codex out for good on a project Claude
still governs** — and a 2.8 KB file Trellis itself told the agent to write is not pathological. The
persisted provenance comments in it are exactly what would be left to degrade; the gate, not a
shortage of material, is what stops it.

**The counterfactual, stated so the sentence above is not read as a regression — this is inference,
not measurement.** It follows from the degrade flag plus the gate, not from re-running the old hook.
Session 1 *degrades* at N = 9, and degradation fires only when the full-provenance assembly already
exceeded the cap — so **the predecessor state refused session 1 too**, with `context-over-budget`
and zero governed sessions; before TRL-30 entirely it refused earlier still, at `invalid-rules` on
the mismatch alone. **This branch buys the project one governed session where it previously had
none. It does not create the wall; it moves the project up to it and then stops.** The limitation
is real and worth fixing — it is not a loss against what shipped before.

**Not fixed here, and tracked rather than left open.** Degrading on the no-mismatch path is a
behaviour change that needs its own tests and its own reviewable diff — the same argument that kept
`block-codex.md` out of this commit. [TRL-29](https://linear.app/kodhama/issue/TRL-29) is **reopened
with this measurement** and is the named consumer that will re-present it (`decision-0078`). The
claim is corrected here, before merge, rather than superseded afterwards, because `decisions/` is
append-only and this record is not yet on `main`.

### 7. What this supersedes

**`decision-0083`, in two respects and no others:**

1. **§1's Claude-only scoping**, and the parity open question that carried it. `codex-context.mjs`
   now reconciles; the table in §1 applies to both hosts, which is what the design claimed and what
   only now shipped.
2. **Its recorded provenance wording** (§4 above) — the quarantine comment is host-neutral in both
   hooks.

**Standing, untouched:** the resolution table itself (missing → add `active = true`; unknown →
quarantine; duplicate → keep the first, quarantine the extras), the ungated-write argument and its
premise that quarantine is non-destructive, the quarantine semantics, **`decision-0083` §5's
partial-file finding**, `decision-0083` §6's payload-derived slug set, the byte-cap investigation,
and the `decision-0070` D4 re-reading. `decision-0072`'s forward pointer is extended for the same
reason: it said the preset copy remains mandatory *on Codex*, which is no longer true.

**Why the partial-file finding is named explicitly.** This list is a derivative of the same
enumeration in `decision-0083`'s `superseded_in_part_by` comment, and it was a strict subset of it —
seven items there, six here, with §5 the one missing (`decision-0028`: a source and its derivative
disagree, and the derivative is the stale side). It stands: `decision-0083` §5 measured the
partial-file shape on Claude, and §1 above **widens** that measurement to Codex rather than
contradicting it. Corrected here rather than in a later record because neither record is on `main`
yet, so append-only does not bite.

**`decision-0070` D4's first two sentences still stand and are still pinned by test on both hosts.**
*"The hook never writes"* is now enforced on Codex behaviourally as well as by construction —
`codexReconciledRows` re-reads `.trellis/rules.toml` after every reconciling run and fails if a byte
changed.

## Consequences

- **`plugins/trellis/VERSION` 0.7.0 → 0.8.0**, with both plugin manifests and `install.sh`'s baked
  bundle manifest advanced in the same commit (`decision-0028`). A payload change that does not move
  `VERSION` never reaches an installed copy.
- **`parseRulesToml`'s contract changed for every caller.** It returns `{rows, mismatch}` or `null`,
  and `null` no longer means "mismatch". Any future call site that treats a non-`null` return as
  "the file was fine" is wrong; the mismatch is inside the success value.
- **A Codex project's `.trellis/rules.toml` can now acquire commented rows**, exactly as a Claude
  project's can. Both READMEs are corrected: they scoped reconciliation to Claude and told Codex
  users to copy a complete preset, which is no longer necessary in order to be governed at all.
- **Four fatal conditions are the remaining fail-closed surface on Codex**, enumerated in §1's
  table. They are the file-is-not-the-schema cases, and they are the guard against the permissive
  direction this change moved in.
- **Four parity annotations on the predecessor branch's plan and design documents are closed out,
  not edited.** Each keeps its original "owed" note and gains a second dated line naming this
  record, so a reader arriving at the debt is not left believing it stands. A fifth (the design's
  *Hosts* matrix bullet) and the plan's now-superseded exact provenance string were closed the same
  way — both were the same class, and leaving them would have reproduced the failure the close-out
  exists to prevent.
- **`plugins/trellis/reference/block-codex.md` still teaches the retired all-or-nothing predicate —
  deferred to [TRL-31](https://linear.app/kodhama/issue/TRL-31), not left open.** The defect stated
  plainly: on the vendored-overlay fallback path that bootstrap tells a Codex agent activation TOML
  is complete only when *"every canonical slug below occurs exactly once, no unknown or duplicate
  slug occurs"*, and to say exactly *"Trellis was not loaded"* when it is not. **An agent following
  it refuses a file this hook now reconciles.** The instruction and the hook disagree about the same
  file, and the instruction is the stale side.

  **Why it is not fixed here.** `block-codex.md` is a **generated payload artifact**. Changing it
  moves `reference/checksums`, the `reference/version` payload stamp and `install.sh`'s baked
  bundle manifest with it — release-shaped work, and wrong inside a governance-only commit whose
  whole value is a diff a reviewer can read end to end.

  **Why it kept being scoped out — and why that argument only ever covered half the question.**
  `decision-0083` left it alone on the ground that this path is *carried rather than maintained*,
  and this record was about to repeat that. **That ground justifies deferring the work; it never
  justified leaving the deferral untracked.** Two consecutive records raising one question with
  nothing to re-present it is exactly the shape `inv-no-orphan-followups` was minted against
  (`decision-0078`) — whose own analysis names this case in `inv-self-improvement`'s violated
  example, *"a PR raises the same open question every time, with no follow-up, and it rots
  unowned"* (`decisions/0078:47`), while ruling that a deferral is not a glitch but **planned work
  with no address**. It now has one: **TRL-31** (Medium, Bug, related to TRL-30 and TRL-20) is the
  named consumer that will re-present it.

- **The over-budget degradation is one-shot and can strand a Codex project permanently** —
  §6 states the measurement and the mechanism. Reopened as
  [TRL-29](https://linear.app/kodhama/issue/TRL-29) rather than carried as an open question, for the
  same reason `block-codex.md` became TRL-31: a deferral with no address is the failure
  `inv-no-orphan-followups` exists to catch.

## What execution found that the design did not

### A recurring defect class: fixtures that cannot produce the condition they name

This is recorded as **method**, not as a bug list. Across this branch and its predecessor, one class
kept reappearing: **a test that passes for the wrong reason because its fixture never reaches the
code under test.** Four instances on this branch alone:

1. **A literal string-replace that no-ops.** `strings.Replace(base, "strictness  = \"firm\"\n", "", 1)`
   — the real line carries a trailing comment, so that exact substring never occurs, `Replace`
   returns `base` unchanged, and the "missing strictness" fixture was the complete file. Replaced
   with a `stripTOMLLine` helper that strips the logical line by key and **fails loudly if it
   removed nothing**.
2. **A fixture gated out before the code under test.** An indented `[rules]` alone never reaches the
   reconciler: `parseRulesToml` trims every line, so `  [rules]` parses as the identical table, the
   mismatch is `null`, and reconciliation is gated on `mismatch !== null`. Proven by mutation —
   deleting `[ \t]*` from `reconcileRows`' own header regex left the indent-only fixture **green**.
   Fixed by pairing the indent with a removed row, which forces a real mismatch. The same shape
   appeared in a strictness subtest whose posture assertion could not fire, because the helper that
   built the project also created `.trellis/internal/`, taking the vendored branch and bypassing
   posture selection entirely.
3. **Slugs the parser rejects.** An `inv-foreign-rule-00` fixture written to exercise the over-budget
   path returned `invalid-rules` and never reached reconciliation at all — the digits fail the row
   grammar. The fixture named a condition it could not produce.
4. **A mutation masked by Go's test cache** — see below.

**Every one was caught by running, never by reading.** The instrument is mutation: break the thing
the test claims to cover and require the test to go red. A fixture that stays green under the
mutation is not testing what its name says.

`decision-0083` recorded five instances of the same class on the predecessor branch and called it
structural rather than incidental. **Two branches, nine instances: the recurrence is the finding.**
A test written alongside the code it covers inherits that code's blind spots, and a fixture written
from a *description* of a condition inherits the description's errors. This is product research — a
candidate obligation for Trellis's own review contract, not a habit that happened to work here.

### Go's test cache makes mutation testing lie about external files

**Specific and mechanical, and it invalidates results silently.** Both hooks are **external files
read at runtime** — `go test` does not track them as cache inputs. A mutation applied to
`staleness.sh` or `codex-context.mjs` and then "verified" with a plain `go test` can replay a
**cached PASS from before the mutation**. Observed directly: the first byte-parity mutation of this
branch showed a stale pass; re-run with `-count=1` it went correctly red.

**Consequence: any mutation verified without `-count=1` is not a result.** It may show the mutation
caught when it was not, or harmless when it was not. Every mutation from that point on this branch
was re-run with `-count=1`.

**The flag moved out of the reviewer's memory and into the project's own instructions.**
`.github/workflows/cli-ci.yml`'s test step and `AGENTS.md`'s documented command both carry
`-count=1`, and `TestCliCIProvidesNode20BeforeGoTests` asserts the workflow still does — so anyone
following the project's own instructions is safe by default, and dropping the flag is a red test.
CI itself was never exposed (`cache: false`, and a fresh runner has nothing to restore); the hazard
is entirely local, which is exactly why the *documented* command is the right place to close it.
The durable fix — making the hook files cache inputs so the toolchain enforces this rather than the
command line — is still **not implemented here**.

## Open questions

- **Should the hook files be made `go test` cache inputs?** The hazard above is now mitigated by
  the *default command* rather than by memory (`-count=1` in the workflow and in `AGENTS.md`), which
  is better than discipline but still not the toolchain enforcing it — a run typed by hand without
  the flag is still a lie. A `//go:embed` of the hook sources, or a checksum fixture the tests read,
  would close it properly. Not attempted here; no consumer named yet, so under `decision-0078` this
  is a question, not a follow-up.
- **When does the third host arrive, and does anything watch for it?** §5's expiry is stated but
  nothing enforces it — the guard will keep passing at three implementations while quietly being
  the wrong shape.
- **`block-codex.md`'s retired predicate is *not* an open question here.** `decision-0083` carried
  it as one, with no consumer. It is filed as **TRL-31** and recorded as a tracked deferral under
  *Consequences* above — deliberately moved out of this list so a third record does not inherit it.
- **Should the byte proxy be replaced by a token estimate?** Inherited unchanged from
  `decision-0083`. Degradation makes the consequence of guessing wrong smaller, not absent.
- **Should the Claude hook warn on a false floor row, as Codex does?** Inherited unchanged from
  `decision-0083`; still out of scope, still asymmetric.

## Self-check (gate)

**The supersession is partial, marked, and scoped.** `decision-0083` keeps its resolution table, its
ungated-write argument, its quarantine semantics and its byte-cap investigation; the forward pointer
in its frontmatter names the two clauses this record reaches and nothing else. Its prose is
unedited (`inv-auditable-archive`, `decision-0082`).

**The invariant this most plausibly violates is confronted rather than left for a reviewer.** This
change moves a parser in the **permissive** direction on a host that used to fail closed, and the
failure mode of getting that wrong — a corrupt file that silently governs — is worse than the
blackout it replaces. §1 states the risk in those terms and enumerates the four conditions that
stayed fatal, rather than reporting only what was gained.

**The boundary with what came before is drawn in both directions** (`decision-0074`): §7 enumerates
what `decision-0083` loses and what it keeps, §5 states the guard's expiry before anyone hits it,
and the surface this change does **not** fix (`block-codex.md`) is **filed as TRL-31** rather than
carried forward as a second record's unowned open question.

**A false claim this record itself made is corrected in place, before merge, and named as false.**
§6's original *"nothing left to degrade"* was measurably wrong and concealed a permanent-blackout
path. `decisions/` is append-only once merged, so the cheap moment to fix it was while this record
was still unmerged; the correction says what the earlier draft claimed rather than quietly
substituting better prose, and the limitation now has a tracked consumer (TRL-29) instead of a
reassuring sentence.

**Where the plan and execution disagree, execution is recorded.** The brief said keep
`!rulesSectionSeen` fatal; §2 records that this was wrong and why. The brief's own fixtures produced
three of the four defects in the method section above, and each is recorded as the controller's
error rather than smoothed away.

**The two supersession acts were taken in the work rather than raised as blocking questions, and the
ground for that is named.** Both are frontmatter lines — `decision-0083`'s forward pointer and the
extension of `decision-0072`'s — cheap to undo, quick to notice, and low-stakes while they stand,
with the prior rationale surfaced and the pointers intact. That is the shape `decision-0081`
describes (*"supersession is weighed by cost of reversal rather than by the seniority of the
record"*), which `AGENTS.md` routes to for exactly this call. **Cited honestly: `decision-0081` is a
proposal, not a taken decision** — its own heading says so — so it is the reasoning relied on, not
an authority claimed. Either pointer can be reverted in one line if the maintainer reads the scoping
differently.

**Author is an agent.** The maintainer directed the work and ruled on host-neutral provenance; he has
not reviewed this record. **His merge is the acceptance** (`decision-0082`) — there is nothing here
to flip, and no agent may merge on his behalf (`floor-intent-gate`). The judgment calls made without
him — narrowing the row grammar rather than widening the reconciler, fixing the awk rather than the
JS on the CRLF divergence, closing two annotations beyond the four named — are each stated above and
each reversible.

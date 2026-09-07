---
id: decision-0095
type: decision
depends_on: [decision-0093, decision-0094]  # coupling, not provenance (decision-0047). decision-0093: rule 1 is the constraint this record works inside -- a consulted reference may not fail a session closed -- and rule 2 is what leaves this cell reachable at all, by making the plugin-copy half of the fallback condition disqualifying rather than substitutive. decision-0094: this record corrects one Consequences clause of it and closes the cell that record named and deliberately left; without 0094's D4 the cell would be two shapes rather than six, and without its disclosure there would be nothing here to correct
changes: [decision-0094]  # ONE clause, in "What is not decided here", and no part of its ruling: "The vendored branch where both copies are unusable is still a silently dead pointer." True when written and false after this change. decision-0094's own superseded_in_part_by carries the scope. Declared as convention, not requirement -- decision-0045:143-144 is permissive ("MAY declare which artifact(s) it changes") -- and following the majority precedent decision-0094 itself measured: of the eleven most recent successors that partially supersede a predecessor, seven declare `changes:` naming it
superseded_in_part_by: [decision-0096]  # 2026-09-07 decision-0096 — TWO clauses, and they sit in different places: the FIRST is in "What is not decided here" and is no part of the ruling; the SECOND is a clause of D6 and IS inside the ruling, scoped at the end of this comment. (An earlier version of this annotation opened "ONE clause ... and no part of the ruling" and then marked the D6 clause further down -- the same blanket-standing-beside-a-named-exception the rest of this comment records catching twice already, surviving one clause earlier than the version that was repaired. Round 3 of the corpus check fixed the mid-field blanket; round 8 found the opening one.) THE FIRST CLAUSE: "The vendored branch where the overlay's OWN copy is present but unusable — zero-byte, newline-only, NUL-filled or mode-0000 — still keeps its token and still says nothing." True when written; false after decision-0096, which closes that cell by reporting it on systemMessage while leaving the pointer exactly where this record left it. The clause named TRL-73 as the consumer that would come back for it (decision-0078), and it did. Marked here so a reader of 0095 alone does not meet it as current truth. THE RULING IS TOUCHED IN EXACTLY ONE CLAUSE, OF D6, SCOPED AT THE END OF THIS COMMENT -- an earlier version of this annotation said flatly "THE RULING IS UNTOUCHED" and then marked that clause three sentences later, which is a contradiction inside one field and the corpus check caught it. D3 in particular is what decision-0096 stands on rather than corrects: "This widens no eligibility check ... What classifies is the REPORT" is the sentence that lets the substitution question be answered NO and the reporting question YES in the same change, and D3's factual claim — that TRL-73's four shapes satisfy existingFile and never reach THIS record's arm — stays true, because decision-0096:1 leaves that check where it was. What did move is the NAME of the guard D3 cites: TestCodexSaysNothingWhenTheOverlayCarriesItsOwnInvariants is now TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting, keeping its fixture, its shape table and its healthy-row silence assertion — a name asserting silence would be false for four of its five rows once the report exists. D1, D2, D4, D5 and D7 are unrelated and are relied on. A SECOND CLAUSE IS MARKED, and unlike the first it sits INSIDE the ruling: D6's "so Claude never delivers an invariants pointer into a vendored project and has nothing to report about". Measured by decision-0096, that is false and was false WHEN WRITTEN rather than falsified by any later change -- the HOOK does not deliver the pointer there, but the DELIVERY does: reference/block-claude.md imports @.trellis/internal/trellis.md, and reference/trellis-a.md and reference/trellis-b.md -- the two posture renders, either of which a vendored overlay copies to that path -- each carry the invariants pointer exactly once, on a line the two files share. (An earlier version of this annotation said trellis-a.md was the ONLY payload file that does; review measured that false. The correction STRENGTHENS the point rather than weakening it: the pointer reaches a vendored Claude session under EITHER posture, not just one.) THE REST OF D6 STANDS, and the distinction is the whole scope of this mark: staleness.sh's path A really does exit at :529 before the repoint at :1403, the cell really is Codex-only AS FAR AS THE HOOK IS CONCERNED, and D6's reason for guarding it in TestCodexDoesNotFallBackOnAnUnusablePluginCopy rather than in the pair guard is untouched and correct. What overreaches is the step from "the hook does not deliver it" to "Claude does not". TRL-76 carries the fix; this mark is what stops a reader of 0095 alone relying on the clause meanwhile.
informed_by: [decision-0028, decision-0078, decision-0087]  # provenance, not coupling (decision-0047): this record would still hold if each said something else. decision-0028 supplies the guard-per-pair reading applied to the shared phrase table and the payload-change-is-a-release rule; decision-0087 is why the four classifications exist to be mirrored at all, but the ruling here is about WHICH artifact a report blames, which does not turn on the gateway; decision-0078 is why the deferred clause in decision-0094 named a consumer, which is the mechanism that brought this cell back rather than a premise of the decision
owner: agent
date: 2026-09-07
---

> **Provenance.** `TRL-71`, found by **both reviewers** of `#294` and filed rather than fixed
> inside it, on a stated cost. `decision-0094` inherited that cost estimate and deferred the cell
> again. Written by a later session, which authored none of `#284`, `#294`, `#295` or `#296`, and
> which measured the estimate before accepting it.

# A vendored delivery names its own dead pointer, and blames the overlay

## Context

Four merged changes took the invariants pointer from *"Codex names a file nothing installs"* to
*"both hosts classify a broken consulted reference identically"*. One cell survived all four: a
**vendored** project whose overlay carries no `invariants.md`, beside a plugin copy that is
unusable. `decision-0093`:2's fallback cannot fire: its second conjunct fails, and there is no
second address to move to — so the token survives, the model is told to read
`` `.trellis/internal/invariants.md` ``, and **nothing is said**.

`decision-0094` disclosed the cell rather than closing it, and named the reason:

> That is `TRL-71`, untouched: a report there would fire on ~28 fixtures and perturb the
> byte-budget tests, which is a change with its own argument to make.

**Measured on this tree, that cost is not what it reads as.** Wiring a bare report onto the arm
and running `go test -count=1 ./...`:

| | measured |
|---|---|
| hook invocations that fire the report | **58** |
| top-level tests that fail | **5** (4 named subtests) |
| failures that are byte-budget overruns | **0** |
| fixtures needing a `reference/invariants.md` | **0** |

**Four** of the five fail on `got.SystemMessage != ""` — silence standing in for *"the hook
did not fail closed"*. The fifth, `TestCodexHookFalseFloorRowsWarnButSucceed`, asserts equality
against the floor warning, which the report displaces by prefixing it; it is the same breakage
for the same reason, but it is not a silence proxy and an earlier draft called all five that.
The 58 firings are an artifact of the suite's own shorthand, not of a noisy
report: `writeCodexPluginRoot` (`cli/codex_hook_test.go:54`) builds a plugin root holding only
`.codex-plugin/plugin.json`, with **no `reference/` at all**, and `writeValidCodexOverlay` writes
no `invariants.md`. Both addresses are dead on every vendored fixture in the suite. A real install
always ships `reference/invariants.md`, so in production the fallback fires and this report stays
silent.

The clause measured false is the Linear issue's conditional — that the fixtures would have to
*grow* a `reference/invariants.md`, "which would move them onto the `decision-0093` fallback path
and perturb the byte-budget tests". They do not have to. `decision-0094`'s own phrasing is
literally borne out, since `TestReconciledCodexPayloadFitsContextBudget` does fail; **it fails on
its silence proxy, not on a byte count**, and the context it measures never moves.

**The one argument that is not about cost** is `#294`'s, and it is still in the hook:

> Scoped to this branch deliberately. Here the plugin's `reference/` IS the delivery source this
> run just read three payload files from, so a missing fourth is this delivery's own defect and it
> names the install to repair. On the vendored branch it is not the source: a plugin root with no
> `reference/` at all is a legitimate shape there […]

That is sound, and it is an argument about **what to say**, not about **whether to say anything**.

## Decision

**A delivery that has proved its own pointer dead says so, on every branch. Which artifact it
blames is what the branch decides.**

1. **The vendored branch reports too.** The hook reaches this arm having already computed both
   answers: the overlay has no copy (`existingFile`) and the plugin's is unusable
   (`payloadDefect`). It then hands the model an address it has just proved nothing is at. Silence
   there is the same defect `TRL-52` named, surviving on the last arm that never had to answer for
   it. The severity argument in `TRL-71` — that `.trellis/internal/` at least **exists** here, so
   the pointer cannot mislead a model into creating an overlay that flips the delivery mode — is
   accepted and is why this is a `systemMessage`, not a `fail()`. It is not a reason for silence.

2. **The report blames the overlay, and that is `#294`'s reason applied rather than overridden.**
   On this branch the plugin's `reference/` is not the delivery source, so *"this plugin payload
   has no readable …"* would name the wrong artifact and send the reader to reinstall an install
   that may be fine. What the reader must repair is the **overlay's** missing file, and the remedy
   sentence says so. The plugin's copy is named second, as the substitute that was unavailable —
   because *which* fault it is decides whether reinstalling would even have helped.

3. **Both halves carry a classification, which applies `decision-0094`'s ruling rather than
   extending it.** Its Decision headline (`0094`:68, above the numbered points) is *"A host that
   cannot say **which** fault it found has not reported the fault."* A report naming the plugin's
   fault precisely and the overlay's not at all would leave the file the reader must actually
   repair as the one described least well — and that is not
   hypothetical: `existingFile` asks `statSync(…).isFile()`, so a *directory* at the overlay path
   is not a present file and reaches this arm, where a bare *"restore invariants.md"* would be
   incomplete until the directory is gone. The shipped remedy is worded for both shapes —
   *"putting a readable invariants.md at that overlay path"* — and the classification beside it
   is what tells the reader a directory is what is in the way.
   **This widens no eligibility check.** `decision-0094`:5 reserves the question of whether an
   overlay's *unusable* copy may be substituted for, and that check is the `existingFile` call,
   untouched. What classifies is the **report**. `TRL-73`'s four shapes still satisfy
   `existingFile` and so never reach this arm at all — pinned by
   `TestCodexSaysNothingWhenTheOverlayCarriesItsOwnInvariants`.

4. **The vocabulary is the shared one, unchanged.** `no readable <absolute path>`, `(it <defect>)`
   in `payload_read`'s four phrases verbatim, and *"yields nothing to read"* rather than
   *"cannot be read"* — false for an empty file, which reads fine (`#295`). Both paths the
   sentence names carry the lead phrase. An operator who meets this report and then the
   plugin-native one must not have to reconcile two vocabularies for one class of fault
   (`decision-0028`'s
   guard-per-pair, applied to a string table that lives in two files in two languages).

5. **`decision-0093`:1 is obeyed, and on this host it is obeyed structurally.** The report rides
   `systemMessage`, outside the `MAX_CONTEXT_BYTES` bound on `context`. Measured on a vendored
   fixture across every shape `payload_read` classifies, the delivered context is **7908 bytes on
   all six broken shapes** — and that one IS a constant, because the broken path keeps the 32-byte
   token rather than an absolute path. Independently reproduced at 7908 by a second session on its
   own fixture. The healthy row is *larger*, so the report's arm never costs bytes at all; the
   healthy figure and the report's own size are deliberately **not** quoted as constants, because
   both embed the plugin root's absolute path and track its length — the error
   `decision-0094`:148-151 recorded paying for in this same series. The report cannot grow the
   context, so it cannot tip a governed session into a refusal. That is the trade
   `staleness.sh:1497` records paying for on the other host and fixed
   by ordering its report after the budget check; here the ordering is not needed because the
   channel is separate.

6. **This cell is Codex-only, and that is now a test rather than a reading.** `staleness.sh`'s
   path A opens at `if [ -d "$internal" ]` (`:482`) and every route out of it exits,
   unconditionally at `:529`, long before the repoint at `:1403`, so Claude
   never delivers an invariants pointer into a vendored project and has nothing to report about.
   **This is why the guard is not `TestBothHostsReportAMissingInvariantsTarget`**: that guard's
   subject is host symmetry on a *config-only* project, and every row of it asserts both hosts
   deliver a governing context. A vendored row would have to assert Claude delivers **nothing**,
   contradicting the guard's own assertion. The cell is guarded where it already had a fixture —
   `TestCodexDoesNotFallBackOnAnUnusablePluginCopy`, built by `#296` from the same fault table —
   and the asymmetry is pinned there as a check, so "Codex-only" cannot quietly stop being true.

7. **The five assertions that used the hook's silence are narrowed; the fixtures are not
   healed.** Giving
   `writeCodexPluginRoot` a `reference/invariants.md` is the one-line alternative and is rejected:
   it would move all 58 invocations onto the `decision-0093` fallback path, silently changing what
   every vendored test covers and the pointer each one delivers. `warningsBesideVendoredInvariants`
   strips this report and nothing else, so each test keeps its own subject and still asserts *"and
   nothing else was said"*.

**What is not decided here.** The vendored branch where the overlay's **own** copy is present but
unusable — zero-byte, newline-only, NUL-filled or mode-0000 — still keeps its token and still
says nothing. `decision-0094`:5 declined to widen the presence check on its own authority and this
record does not either: that is a ruling about whether a project's own unusable file may be
substituted for, and `TRL-73` carries it (`decision-0078`). This record's arm is reached only when
`existingFile(overlayInvariants)` is **false**, so it cannot absorb that shape by accident.

## Consequences

- A vendored session whose pointer is dead at both addresses now says which file is missing, which
  copy could not stand in, and why. Measured on the fixture above, the report is 832–922 bytes
  depending on the classification and the path lengths, and **none of it reaches `context`**.
- **`decision-0094`'s disclosure of this cell is corrected, and its ruling is not touched.** The
  clause *"The vendored branch where both copies are unusable is still a silently dead pointer"*
  sits in *"What is not decided here"*, not in the five-point Decision, and it named `TRL-71` as
  the consumer that would come back for it. It has. `decision-0094` now carries `decision-0095` in
  `superseded_in_part_by` for that clause alone, so a reader of `0094` alone does not meet it as
  current truth. Its cost estimate is left as written and answered in Context rather than edited
  out: what a record believed when it deferred something is part of why it deferred it.
- **`decision-0093` needs no further marking.** Its first Consequences bullet is already marked by
  `decision-0094`, and this change moves no pointer: in this cell the address stays the overlay's,
  exactly as `0094`:4 left it. Only what is *said* about it changed.
- `TestCodexDoesNotFallBackOnAnUnusablePluginCopy` gains the report assertions and the
  Claude-silence check; `TestCodexRepointsWhenAVendoredOverlayLacksInvariants` gains the
  discrimination control, because it is the fixture where the fallback actually fires and a hook
  that reported unconditionally would otherwise pass everything.
- `claudeContextFor` is split, keeping its contract and lifting out `claudeStdoutFor`. The vendored
  fixtures need staleness.sh's raw stdout because that hook legitimately emits **nothing** there,
  and `nudgeContext` fails on an empty string — right for a caller expecting a delivery, wrong for
  one asserting the absence of one.
- A payload change, so a release: `VERSION` 0.20.0 → 0.21.0, both `plugin.json` manifests, and
  `install.sh`'s baked `TRELLIS_BUNDLE_MANIFEST` (`decision-0028`). Only the four files this change
  touches moved in the manifest; `reference/checksums` and `reference/version` hash `reference/`
  alone and did not.
- `decision-0053` is not engaged: the shipped *prose* is unchanged. What changed is what one
  channel says **about** a file it points at.

## Self-check (gate)

- **Smaller thing that works?** Yes, and two smaller ones were rejected with reasons. *Leave the
  cell silent and record that* is the smallest, and it is what `decision-0094` provisionally did;
  it rests on a cost this record measured and did not find. *Heal `writeCodexPluginRoot` so the
  question never arises* is one line, and is answered in D7 — it buys a smaller diff by silently
  changing what 58 existing invocations cover.
- **Grounded in an artifact?** The ruling is one `else if`, one warning, and assertions driven by
  the fault table that already existed. Every claim in Context is a measurement re-runnable from
  the suite; the byte figures in D5 are from a fixture, not an estimate. Reverting the arm fails
  `TestCodexDoesNotFallBackOnAnUnusablePluginCopy` on all twelve subtests — two overlay shapes
  against six plugin faults — which is the red this change was written against. Twelve on a
  normal runner: the two mode-0000 rows `t.Skip` as root, where that fixture cannot be built.
- **Reversible?** One `else if` and one `warnings.push`. Deleting both restores today's behaviour
  exactly; the test helper would then strip a report nobody emits, which its own `t.Fatalf` reports
  rather than hides.
- **Intent gate.** Authored by an agent; the merge is the maintainer's act.

---
id: decision-0086
type: decision
depends_on: [decision-0083, decision-0084]
informed_by: [decision-0028, decision-0070, decision-0078, decision-0081, decision-0082]
owner: agent
date: 2026-09-03
---

> **Provenance.** Directed by the maintainer as **TRL-29**, reopened on 2026-09-03 after PR #259's
> merge auto-closed it while only half of it had shipped. Designed through
> `superpowers:brainstorming`, planned through `superpowers:writing-plans`
> (`docs/superpowers/plans/2026-09-03-codex-degrades-on-the-no-mismatch-path.md`) and executed
> through `superpowers:executing-plans` on
> `feature/trl-29-codex-degrades-on-the-no-mismatch-path`.
>
> **Written after the code shipped, reporting what shipped.** Two of its claims exist only because
> execution contradicted the plan: the mismatch-path strip turned out to be a *refusal* rather than
> a merely smaller degradation, and the runaway guard's new threshold is thirty rows where the plan
> asserted a derived thirty-seven. Both are corrected below against measurement.
>
> **Amended 2026-09-03, on review of PR #263** (thread `PRRT_kwDOTIeCVc6eu78z`): the announcement
> the degraded path appends was itself unbudgeted, and §4's claim about when the guard is reached
> was false by the announcement's own length. The dated note under §4's table records what was
> found, what changed, and the re-measured threshold. The original sentences stand above it.
>
> **Amended 2026-09-04, second review pass on PR #263** (thread `discussion_r3931029360`): two of
> §4's byte figures were one fixture's, stated as general — a quarantined row costs 42 B / 192 B
> when its slug is a character longer than the one the tests use. Two dated notes in §4 correct
> that, one under the table and one under the 2026-09-03 note; the second also records the rename
> of the test that note cites. The
> original sentences stand above each, and no runtime behaviour changed in that pass.

# 0086 — The injected copy degrades on any over-budget session, not only a reconciling one

## Context

**`decision-0084` §6 removed the Codex blackout for one session and rebuilt it for the next.** Over
`MAX_CONTEXT_BYTES` the hook stopped refusing and began injecting the rows without provenance,
announcing the omission and mandating a **full-provenance** write. That much shipped and works.

The branch was gated on `mismatch !== null`. So it ran only in a session that had something to
reconcile — and the session *after* the repair has nothing to reconcile, because the file the
mandate asked for already carries every row plus its persisted `# quarantined …` / `# added …`
comments. The gate was skipped, the hard refusal fired, and nothing about that file ever changes
again. **Applying Trellis's own repair was what caused it.**

`decision-0084` records this, and records that an earlier draft of the same section had called the
hard refusal a state reachable only by a file "pathologically large" with "nothing left to degrade."
That claim was corrected before merge because it was measurably false: the file in question is
2.8 KB, Trellis wrote it, and its quarantine comments are exactly what was left to give up.

**Measured, on `3f44620`, before any change** — real firm payload, `rules-a.toml` plus N foreign
rows, `go test -count=1`:

| N foreign rows | session 1 | file the mandate produces | session 2 |
|---|---|---|---|
| 5 | degrades, delivers 9010 B | 2097 B | 8831 B delivered |
| 8 | degrades, delivers 9133 B | 2670 B | 9404 B delivered |
| **9** | degrades, delivers 9174 B | **2861 B** | **refuses, `context-over-budget`, nothing injected** |
| 12 | degrades, delivers 9299 B | 3434 B | refuses |

`staleness.sh` delivers 9833 B from that identical 2861 B file. **A project was governed on one host
and silently ungoverned on the other, from one file, permanently.**

## Decision

### 1. The trigger is the budget, not the reconciliation

The over-budget branch no longer asks whether a reconciliation ran. It asks what it was always
asking — whether the assembly fits — and drops Trellis's provenance from the **injected copy** in
both directions:

- provenance this session would **generate** (`reconcileRows`' own notes, left off by
  `withProvenance = false` — unchanged from `decision-0084`);
- provenance the file **already carries** from an earlier repair (`stripPersistedProvenance`, new).

The `mismatch !== null` test survives only to choose which **announcement** the session gets.

**The invariant this rests on is unchanged and is the reason the change is safe: the file is the
archive, the injection is the working set.** Nothing on disk is touched (`decision-0070` D4), no
row's value is lost on any path, and quarantine still never deletes — a stripped quarantine keeps
its commented-out line and loses only the note appended to it, which is byte-for-byte the shape
`reconcileRows` already produces with `withProvenance = false`.

### 2. The reader is derived from the writer, not written beside it

`quarantineNote` and `addedHeader` become templates with `{date}` / `{stamp}` / `{count}`
placeholders, and both the writer and `stripPersistedProvenance`'s two patterns are derived from
them.

**A pattern written by hand next to a string built by hand is two statements of one text, and they
drift silently** — the writer's wording moves, the reader keeps matching yesterday's, and the
degradation stops degrading with no signal. That is TRL-29 again, reached from the other side.
Deriving both from one template makes the drift impossible rather than merely tested for.

It is also cross-host: `staleness.sh:862` and `:933` write the identical text, one
`.trellis/rules.toml` is read by both, so a file repaired on Claude must strip on Codex.
`TestBothHostsReconcileIdentically` keeps the two **writers** in step; the templates keep this
host's **reader** in step with this host's writer. Mutating only this host's template makes the pin
go red against a Claude-written file, which is the drift that matters.

### 3. The no-mismatch path announces, and instructs nothing

`provenanceOmittedNotice` is `repairMandate`'s counterpart where there is no repair. It shares the
mandate's marker sentence *verbatim* — "Provenance was omitted above to fit the context budget" —
because the abbreviation is one fact however the session reached it, and the test helper
`codexDegradedMarker` matches exactly that sentence to tell a degraded response from a full one.
Two wordings would leave that helper half-blind, and the host-parity comparison would then blame
parity for a degradation.

**It carries no write instruction, and must not grow one.** The mandate has to say *"the
full-provenance version, not the abbreviated ones shown above"* because it is asking for a write at
all. Here nothing asked. The only thing an instruction could achieve is the exact outcome the
branch exists to prevent — an agent helpfully rewriting `.trellis/rules.toml` from the abbreviated
copy and losing the provenance for good.

### 4. The hard refusal survives, and its comment says what reaching it means

It is the runaway guard and nothing more. It is reached only when a context with **no Trellis
provenance left in it at all** — neither generated nor persisted — is still over the cap. At that
point what remains to abbreviate is the **consumer's own content**, their comments and their row
set, and that is not a call this hook may make on its own. So it stops, loudly, at `exit 0`.

**The corrected claim is not reintroduced.** The guard is not described as a state with "nothing
left to degrade": that wording was false once and would be false again, since what is left is the
consumer's file rather than nothing.

**Measured after the change**, same payload, same slug family as the baseline above:

| | before | after |
|---|---|---|
| cost of a quarantined row to the injected copy | 192 B | **42 B** |
| session 2 at N = 9 | refused, 0 B | **delivered, 8659 B** |
| first refusal | **N ≥ 9** | **N ≥ 30** |

Thirty is a measured order of magnitude, not a threshold: it is a byte budget, and a longer slug
reaches it with fewer rows. No test asserts the number.

> **Dated note, 2026-09-04 — second review pass on PR #263, thread `discussion_r3931029360`.**
> The table's first row states a *fixture-specific* measurement as a general one. A quarantined
> row's cost is not a constant: the injected copy emits `` `# ${line}${note}` ``
> (`codex-context.mjs:515`), so the row's own text is part of the figure and the figure moves with
> the slug's length. **42 B / 192 B is what a slug one character longer costs** — any 19-character
> slug, of which a two-letter suffix is one instance. The 18-character slug this branch's own tests
> use — `inv-foreign-rule-a` — costs **41 B stripped / 191 B with its note**. No committed fixture
> uses a longer one, so 42/192 was reconstructed rather than measured from anything in the repo.
>
> **This record's own baseline table proves it, in both columns.** In the `3f44620` table under
> *Context*, session 2 delivers 8831 B at N = 5 and 9404 B at N = 8: 573 B for three quarantined
> rows carrying their notes, which is **191 B each**, not 192. The stripped column agrees — session
> 1 delivers 9010 B at N = 5, 9133 B at N = 8 and 9174 B at N = 9: 123 B for three rows and 41 B for
> one, which is **41 B each**, not 42. (The N = 12 step is 125 B rather than 123 because the mandate
> spells the counts, and "9" became "12" in two places.)
>
> What does *not* vary is the note itself: `QUARANTINE_NOTE_TEMPLATE` filled is **150 B** whatever
> the row it hangs off, which is why "stripping frees 150 B per quarantined row" in the note below
> holds generally where 42/192 does not. The `N ≥ 30` first-refusal figure in the table is untouched
> by this, and so is the `N ≥ 37` it was re-measured to — but that one was reproduced in a review
> pass on 2026-09-03, not by anything committed here. No fixture in this repo reaches 37 rows, and
> none asserts either number, so a reader cannot check them from the tests. Deliberately: the hook's
> own comment says to treat thirty-seven as a measured order of magnitude, not a threshold to test
> against.
>
> **One instance of the retired figure is left uncorrected**, in the runaway guard's comment
> (`codex-context.mjs:1113`), which also says 42 B / 192 B. Its next sentence already tells the
> reader the figures move with slug length — "It is a byte budget, not a row count — a longer slug
> reaches it sooner" — so what is stale there is the bare number, not the reasoning around it. It is
> left as it stands, and **nothing tracks it** — said plainly rather than parked, per
> `decision-0078`. Correcting it means changing a byte of the hook, which fails
> `TestInstallScriptBundleManifestIsCurrent` until the whole bundle is re-rendered; that price is not
> worth one figure whose reasoning is already sound. The next change that edits the hook for its own
> reasons passes within a line of it.
>
> **Two dated notes sit under this table now.** The provenance block's and the self-check's phrase
> "the dated note under §4's table" was written on 2026-09-03 and means the note *below* this one —
> the one about the unbudgeted announcement. Both locators are left as written.

> **Dated note, 2026-09-03 — review of PR #263, thread `PRRT_kwDOTIeCVc6eu78z`.** The paragraph
> above says the guard is reached "only when a context with no Trellis provenance left in it at all
> is still over the cap." That was not what shipped. The degraded path appended its announcement —
> `provenanceOmittedNotice` on one side, the degraded `repairMandate` on the other —
> *unconditionally*, and the guard measured the sum. So a body that fit on its own was refused for
> the announcement's bytes, which are Trellis's own text and not the consumer's content.
>
> **Measured on the branch as it stood:** the full notice costs **414 B** and the degraded mandate
> **156 B more** than the full mandate it replaces, while stripping frees **150 B** per quarantined
> row. At one or two persisted notes the "degraded" assembly was therefore *larger* than the full
> one it stood in for and could rescue nothing; above that, a 414 B (or 156 B) band of fitting
> bodies still refused. The reviewer's fixture — firm preset, one persisted note, a valid 1450 B
> project comment — assembled to 9517 B in full, 9367 B stripped, and was refused at 9781 B.
>
> **What changed.** Each path now lists its announcements from most to least informative — the full
> form, then a one-paragraph compact form (**129 B** for the notice, **571 B** for the mandate) —
> and the first assembly that fits ships. The guard is reached only when the stripped body will not
> fit alongside the *shortest* honest announcement. The residual is that line's own length and is
> **left open on purpose**: closing it to zero would mean injecting an abbreviated copy with no word
> that it was abbreviated, or a reconciliation with no mandate to write it back — a quiet failure
> traded for a loud one. Both compact forms live *inside* the functions the destructive-verb guards
> already scan, so no registration step can be forgotten.
>
> **Re-measured, same payload, same slug family.** The per-row figures hold — a quarantined row
> still costs the injected copy **42 B** stripped against **192 B** with its note. The threshold
> moved: the reviewer's fixture now delivers at **9496 B**; the mismatch path delivers where a
> single foreign row used to refuse outright; and the first refusal appears at **N ≥ 37** rows
> where the table above says thirty. Thirty-seven is, by coincidence, the number the plan derived
> before the notice was ever budgeted — a derivation that ignored the announcement's cost was right
> about a design that no longer pays it. It remains a byte budget, and no test asserts the number.
>
> **Pinned** by `TestCodexAnnouncementNeverTipsAFittingBodyOverBudget` (both paths, sized against a
> measured baseline rather than a hardcoded total) and by "the full form must be preferred when it
> fits" assertions in the two existing degradation tests. Eight mutations, each killed:
> the unconditional append restored (both subtests refuse); compact made the default (both
> preference pins red); the compact notice without its marker; a deletion verb in the compact
> notice (both verb guards, and the window test itself — the one-note fixture leaves the line 11 B
> of growth, which is the pin for the reviewer's case); the extractor without the compact stems;
> the compact mandate asking for the abbreviated rows; the guard measuring the body without its
> announcement; a deletion verb in the compact mandate.

> **Dated note, 2026-09-04 — second review pass on PR #263, thread `discussion_r3931029360`.**
> The "Re-measured" paragraph above restates the table's per-row figure — "**42 B** stripped against
> **192 B** with its note" — and inherits its error. Those are the numbers for a slug one character
> longer than any the tests use; the cost varies with the slug's length, and the 18-character
> `inv-foreign-rule-a` costs 41 B / 191 B. See
> the note under the table for the derivation. Nothing else in the paragraph changes, including
> `N ≥ 37` — reproduced in review on 2026-09-03 rather than by any committed fixture, as the note
> under the table records.
>
> **The test named in the *Pinned* paragraph above was renamed in the same pass.**
> `TestCodexAnnouncementNeverTipsAFittingBodyOverBudget` is now
> `TestCodexBudgetsTheAnnouncementAlongsideTheBody`. "Never" claimed a guarantee the code does not
> make and says it does not make: a body that fits alone but not alongside even the *compact*
> announcement is still refused, deliberately, because injecting an abbreviated copy with no word
> that it was abbreviated is the worse failure. Three subtests were added — the residual window on
> each path, and the cap's own boundary, where a candidate of exactly `MAX_CONTEXT_BYTES` must be
> accepted. Two further mutations, each killed: `<` for `<=` in the announcement loop (the boundary
> subtest refuses at exactly the cap), and an empty last-resort announcement that lets the guard
> inject unannounced (both residual-window subtests). No runtime behaviour changed.

## Consequences

**A project repaired by Trellis on Codex stays governed on Codex.** That is the whole of what TRL-29
was reopened for, and it closes.

**The mismatch path changed too, and the issue did not ask for it.** `decision-0081` puts this
inside an agent's authority — it is cheap to reverse, changes no full-provenance output, and the
alternative was worse. The reason is not symmetry. A file repaired once that drifts again arrives
carrying *both* kinds of provenance, and reconciling it from the raw source leaves every persisted
note in the injected copy. Execution measured what that costs: **at nine persisted rows plus one new
foreign row it is a refusal, not a smaller degradation** — the same permanent blackout one step to
the left. The plan predicted a cosmetic difference and was wrong; the test that pins it
(`TestCodexDegradesPersistedProvenanceOnTheMismatchPathToo`) exists because the first mutation
attempt survived, proving the branch was unguarded.

**The counts a repair reports are unaffected.** `reconcileRows` classifies from uncommented
`(inv|floor)-… =` rows only, so removing comment text changes no quarantine or addition decision
and neither counter. The pinning test asserts `added 0 row(s); quarantined 1 row(s)` on a
nine-persisted-note file, which is this session's work and not a running total — the defect
`staleness.sh` fixed for itself and this must not reacquire.

**A new agent-facing text channel is inside the destructive-instruction guards, not beside them.**
`codexPayloadAssembly` now names both `repairMandate` and `provenanceOmittedNotice`. Proven
load-bearing: a deletion verb planted in the notice is caught, and lands **undetected** if the scan
is narrowed back to `repairMandate` alone. `decision-0028`'s "a guard per pair" applied to a text
channel rather than a file pair.

**`decision-0084` §6 is superseded in part**, in the two sentences that describe the degradation as
one-shot and gated on `mismatch !== null`. Everything else in that record — the parity table, the
byte-identity guard, the corrected claim about the hard refusal — stands unchanged and is depended
on here.

**What did not change.** The cap itself (9500 B) and its units. `decision-0083` established that the
cap has no external provenance and that Codex *spills* rather than rejects; this record does not
revisit either. The `readRequired` size gate that refuses a `.trellis/rules.toml` larger than
`MAX_CONTEXT_BYTES` before it is ever parsed is also untouched — it is a separate limit on the file
rather than on the assembly, and it does not bind in the measured range (a 7 KB file assembles
fine).

## Open questions

**Should the cap be expressed in tokens, or removed in favour of Codex's own spill?**
`decision-0083` recorded that Codex measures tokens and degrades by spilling, and that any byte cap
is an approximation that will drift. This record keeps the byte cap because the degradation now
makes it survivable, not because the unit is right. **The consumer that would re-present this**
(`decision-0078`): the first Codex payload change that pushes the *undegraded* assembly over 9500 B
for an unmodified preset — at that point the cap is shaping the healthy path, not the runaway one,
and its units become load-bearing.

**Nothing else is deferred.** No follow-up issue is filed, and the absence is deliberate: an idea
with no trigger is a to-do nobody agreed to.

## Self-check (gate)

- **Iron rule.** Grounds out in shipped artifacts, not description: `codex-context.mjs`'s branch, and
  the tests named above. Every claim in this record is either quoted from source or measured with
  `go test -count=1` on this branch.
- **`inv-auditable-archive`.** Append-only. `decision-0084` gains a forward pointer and no edit.
- **`inv-no-orphan-followups`.** The one open question names the consumer that re-presents it
  (`decision-0078`); nothing else is parked.
- **`floor-intent-gate`.** No deletion verb reaches the agent through either text channel, enforced
  by `TestEveryDeletionInstructionIsGated` / `TestEveryDestructiveInstructionIsGated` over both
  functions. The hook still never writes.
- **Measurement discipline.** Two numbers in the plan were derived rather than measured and were
  wrong (thirty-seven rows; a cosmetic mismatch-path difference). Both were re-measured and
  corrected before this record was written, and the hook comment carries the measured figures.
  A third claim — §4's "reached only when … is still over the cap" — was falsified in review and is
  corrected by the dated note under §4's table, with the threshold re-measured rather than
  re-derived; the original wording is kept above the note.

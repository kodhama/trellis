---
id: decision-0090
type: decision
depends_on: [decision-0037, decision-0047, decision-0068, decision-0070, decision-0078, decision-0079, decision-0088]  # coupling, not provenance (decision-0047). decision-0088: this record changes ONE clause of it -- the closing sentence of the Consequences bullet under "What is not claimed here" -- and leaves D5 standing; if that text said something else there would be nothing here to correct. decision-0068: D12 is where the budget was created as an amendment to spec-0005 AC2, and D5.2 below leaves its "next simplification" sentence standing on a distinction that is false if D12 says something else. decision-0079: the whole ruling rests on its retirement of the spec stage -- if spec-0005 were still a ratified document, AC2 would OUTRANK a test comment and the count could not live there. decision-0070: D5 is the live decision the third read implements -- D3 below, and the whole ground for letting that read stand without a record of its own, are false if D5 is not on main and current. decision-0078: D1's licence for the test rests on its address test for a deferral's consumer -- a different test for what counts as a named consumer that re-presents a constraint could make the test an unacceptable home. decision-0037 and decision-0047 for the same reason decision-0088 declares them: this record's `owner: agent` and this frontmatter's shape are contingent on their rulings, not informed by them
changes: [decision-0088]
informed_by: [decision-0010, decision-0028, decision-0040, decision-0081, decision-0082, decision-0085]  # decision-0040 added after an independent corpus review noted it undeclared: this record uses the partial-supersession instrument it created and its Self-check names 0040's clause-level scope requirement as the check that caught this branch's one defect. Placed in informed_by, not depends_on, following decision-0082 -- which partially supersedes 0040 and still lists it here -- and matching decision-0071 and decision-0088, both of which add partial pointers without declaring it at all
owner: agent
date: 2026-09-04
---

> **Provenance.** Filed as `TRL-42` by the agent session that merged `#267`, which named this
> and deliberately did not act on it: *"If you want that widening in the corpus the way the second
> read got `0088` D5, that is a small record and your call — I did not draft one speculatively."*
> Written by a different session, which did not author `#267`.

# 0090 — the read budget's count is the test's; a read that *decides* is the corpus's

## Context

`install.sh` may read a project file's contents only a counted number of times. The count was set
once and has moved twice since, and the three statements do not sit in the same place:

| move | read added | where it was argued |
|---|---|---|
| → **one** | the managed-block opening marker | `decision-0068` D12, as an amendment to `spec-0005` AC2 |
| → **two** | `strictness` in `.trellis/rules.toml` | `decision-0088` D5 **and** the test, named as both |
| → **three** | the top-level `governed` key of that same file | `#267` (`74d9d25`), in `cli/install_script_test.go` |

`decision-0088` required the third to *"come and argue itself"* here. It came, and argued somewhere
else. `TRL-42` asks whether that was wrong.

**Where that requirement actually sits is load-bearing, and both `TRL-42` and an earlier draft of
this record got it wrong.** It is not in D5. It is one sentence in the Consequences bullet under
*What is not claimed here* — its only statement in the record. D5 says something different, and is
quoted here in full because the half closest to `TRL-42`'s reading is its first (emphasis original):

> **5. The content-read budget goes from one to two, and the second is named** rather than absorbed.
> The test that pins the count said a second read "has to come here and argue itself"; this record
> and that test are the argument.

Its first sentence is the half that most supports `TRL-42`'s reading — the second read was *named*
in the corpus rather than absorbed — and D2 below concedes it: that read **selects**, so it was owed
this corpus and got it. But the second sentence is the one that decides the question. The demand
`0088` was discharging came **from the test**, and `0088` counts **the test** as one of the two
things that answered it. The record whose sentence `TRL-42` reads as reserving the budget to the
corpus had already, in its operative ruling, treated a test as a place where a widening is argued.

### The behaviour is not what is in question

Refusing to render over a `governed = false` opt-out **implements `decision-0070` D5**, which is
already on `main`: the plugin hook reads that key before every delivery path and injects nothing
on it, while `install.sh` rendered the full rules body into always-loaded context for the same
project. Implementing a live decision needs no record. What `TRL-42` reopens is only whether the
**count's widening** does.

### The three reads are not the same kind of thing

- **The marker read** and **the `governed` read** only ever **REFUSE**. They select nothing,
  patch nothing, write nothing, and branch to exactly one outcome: render, or decline loudly.
- **The `strictness` read** **SELECTS** — it decides which of two shipped payload files becomes
  the rendered header. That is decision logic in a script whose stated boundary is that it makes
  one decision (scope).

This asymmetry is why `decision-0088` had to exist and `#267` did not need a twin. `0088` was not
written to raise a count. It was written because taking the header from `strictness` **reversed
`decision-0068` D5's constant-header ruling** — a decision on `main`, falsified by the change.
The budget sentence in its D5 is a rider on a record that had to exist for another reason
entirely. Read 3 falsified nothing; it discharged something.

### `decision-0088`'s own Context already made the argument for the test

The ground it stood on to widen the budget at all is `decision-0079`:

> `decision-0079` retired the spec stage: `spec-0005`'s only surviving statement is
> `cli/install_script_test.go`. The constraint is now a **test**, which is changed by argument in
> the same commit, not a ratified document that outranks the change.

That is `0088` saying the budget's constraint lives in a test — and then, in its Consequences,
requiring the next change to it to be argued in the corpus instead. Both halves cannot govern the
same kind of change. `TRL-42` is the bill for that.

### Only one instrument can fail

A record cannot notice a fourth read. `cli/install_script_test.go`'s exactly-*N* subtest fails,
loudly, in the commit that adds one, in front of the person adding it — classifying each `grep` by
its **operand** so that aliasing a path through a variable does not launder a read past it. It
already carries the count's whole history in one place, which the corpus does not: `0068` D12 holds
the *one*, `0088` D5 the *two*, and neither holds the *three*. Under `decision-0078` that is a
consumer with a name and a firing condition; a record whose reader is nobody-in-particular is not.

### Why "just a note on `0088`" was not available

`TRL-42`'s second answer suggests a dated note in `decision-0085` §5's shape. That shape is not
reachable here. `AGENTS.md` forbids editing a decision's body — supersession is by forward pointer
— and the artifact contract's check 7 requires a pointer's entries to **resolve** to an artifact.
So correcting `0088`'s third-read sentence needs a record to point *at*, whichever way the ruling
goes. **Both of `TRL-42`'s answers land in a record.** The live question is only what it says: the
instance, or the class.

## Decision

**1. The count is the test's.** The content-read budget for `install.sh` is stated and enforced by
`cli/install_script_test.go`'s exactly-*N* subtest. A change to the **count** is argued in the
comment above that subtest, in the same commit that makes it, and needs no decision record. This
is `decision-0079`'s ruling applied to the artifact it already named: a constraint whose only
surviving statement is a test is changed by argument in the same commit.

**2. A read that *decides* is still the corpus's.** A read comes here — before it lands — when it
does anything beyond refusing: **selects** an output, **patches** a file, **writes** state, or
**removes** an existing refusal. The test governs how many; this corpus governs what the installer
is allowed to decide. Read 2 crossed that line and got `decision-0088` D5, correctly. Read 3 did
not.

**3. The third read is named here, once**, so the budget's corpus statement is not left
mid-sentence: `install.sh` reads the top-level `governed` key of `.trellis/rules.toml`, first and
before every other branch, with the hook's own matcher copied byte for byte and pinned by
`TestInstallScriptGovernedParserMatchesHook`, guarded regular-and-readable, and refuses the render
when it says `false`. It only refuses. It needs no record of its own; this ruling is what it
needed.

**4. `decision-0088`'s third-read sentence is corrected as to *where*, not as to *whether*.** The
sentence is *"A third read still has to come and argue itself"*, closing its Consequences bullet
under *What is not claimed here* — **not** in D5, where `TRL-42` places it. It stands for a read
under D2 above and is wrong only for a read that merely refuses. `#267` is precedent for a read that only refuses, and **not** for one that decides;
the next read is owed that question — *does it only refuse, and does it implement a decision already
on `main`?* — rather than a bare count to cite.

**5. `decision-0068` D12 is *not* corrected.** Its *"the next simplification has to come here and
argue itself"* survives untouched, and the distinction — *simplification*, D12's own word — is
load-bearing rather than a courtesy: a simplification **removes a refusal**, which is D2's fourth
clause. D12's own history is the
evidence — deleting the marker grep achieved a genuine existence-only script and regressed inline
consumers into silent double delivery. Lowering the count is not the mirror of raising it.

## Consequences

- **`decision-0088` is changed in part, not superseded.** What is corrected is one sentence in one
  place: *"A third read still has to come and argue itself"*, closing the Consequences bullet under
  *What is not claimed here*. The rest of that bullet stands — no general licence to read
  `.trellis/`, no branching on other project state, the widening still *"one key, for one purpose,
  counted and guarded"*. **D5 is not corrected**, and saying so is not a courtesy: its closing
  sentence is about the second read and names the test as one of the two arguments that answered the
  demand, which under D1 above makes it supportive rather than merely untouched. D1–D4 stand in
  full. The forward pointer on `0088` records the scope.
- **What is *not* claimed here:** that `install.sh` may read `.trellis/` generally; that a fourth
  read is pre-approved; or — the widest misreading available and the one worth naming — that any
  constraint whose statement happens to be a test escapes the corpus. This ruling is about **one
  budget, one script, one count**. `decision-0010` keeps artifact conformance agent-applied for the
  same reason in reverse: the instrument follows the constraint, not the other way round.
- **The same misattribution stood in the test comment, and is corrected in the same change**
  (`decision-0028` — a source and its derivative are fixed together, never one without the other).
  `cli/install_script_test.go` also attributed the third-read sentence to `0088` D5. Leaving it would
  have left the artifact D1 elevates as the count's home carrying the very error this record
  corrects, which is the defect `inv-graph-maintenance` names. An automated reviewer on `#271` found
  it after the corpus half was already fixed. That comment now also carries D1 and D2 themselves, so
  the rule is legible at the point where it fires rather than only in `decisions/`.
- **The pair-guards are a separate and untouched obligation** (`decision-0028`). Both copied
  parsers are pinned to the hook's by name; a read that copies hook logic still owes a guard,
  whether or not it owes a record.
- **The recurrence this ends** is a record per read whose entire content is that the number went up
  by one. That is a decision record with no decision in it, and maintaining one beside a test that
  already fails is the duplication `decision-0079` retired the spec stage to stop.

## Self-check

- **Written by a session that did not author `#267`**, on the ticket that session filed against its
  own omission. That is the one structural advantage this record has over `decision-0088`, whose
  Self-check records that it was authored by the agent that wrote the code it justifies
  (`decision-0037` — authorship, not accountability). It is not independent review of the ruling
  itself, which is still owed (`inv-independent-judgment`).
- **Two reviewers found the same defect in two places, and both findings were material.** The count
  matters, so it is stated here rather than left to be reconstructed: `corpus-reviewer` found the
  first, an automated reviewer on `#271` the second, and neither was found by the agent that wrote
  the draft. **First**: the draft — following `TRL-42`'s own wording — placed the corrected clause in
  `decision-0088` D5, where it is not. It is in that record's Consequences, once. The check that caught it is
  `decision-0040`'s clause-level scope requirement, and the correction improved the argument rather
  than only its citation: D5's real closing sentence attributes the demand to the **test** and counts
  the test as one of the two things that answered it, which is D1's case made by the very record
  `TRL-42` reads against it. A ruling argued from a misquoted source would have deserved rejection
  regardless of whether it reached the right answer. **Second**: after that fix, the identical
  misattribution was still sitting in `cli/install_script_test.go` — the artifact D1 elevates — and
  is corrected in the same change under the Consequences bullet above. Fixing a citation in the
  corpus while leaving it in the derivative is the failure this record's own D1 makes most costly,
  since it would have made the count's designated home the least accurate statement of it. The one
  violation the review reports elsewhere (`decision-0044`'s bare `kodhama-0004-uniform-lifecycle`)
  predates this change and is left alone, as `decision-0088` also left it: that record preserves it
  deliberately.
- **The strongest argument against this ruling, stated rather than left for the reader.** The
  budget is a proxy for `spec-0005` AC2's claim that `install.sh` carries no decision logic, and
  erosion is cumulative: every individual read looks justified in isolation, and a test comment is
  written by the same agent making the change, so the standard and the change are authored
  together. The corpus is where a standard outlives its author. This record answers that with D2
  rather than by denying it — the *class* of read that erodes the boundary still comes here, and it
  is the class, not the count, that the argument is actually about. A reader who thinks the count
  itself is the boundary should reject D1 and take `TRL-42`'s first answer.
- **`TRL-42` listed "write the record" first and it was not taken as written.** What is delivered is
  a record, so its "Done when" is met either way; what is refused is the reading under which every
  future read owes one.
- **`decision-0081`'s cost-of-reversal framing applies and is cited at its own stated weight** —
  that record calls itself *"(proposal, not a decision)"*. Under it this is cheap to reverse: the
  ruling has no code, and retiring it costs one forward pointer and restores `0088`'s third-read
  sentence to full force.
- **This record is an instance of the thing it rules on, and that is not hidden.** It exists partly
  because the corpus sentence had to be corrected *somewhere* and the contract left a record as the
  only target. A ruling that the corpus should hold less, which itself required a corpus artifact to
  land, is worth reading with that in mind.
- **The maintainer has not ruled on this.** It is drafted for his act; the merge is the acceptance
  (`decision-0082`, `floor-intent-gate`). No agent may merge it.

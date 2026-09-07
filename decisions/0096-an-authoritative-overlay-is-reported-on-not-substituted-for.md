---
id: decision-0096
type: decision
depends_on: [decision-0093, decision-0094, decision-0095]  # coupling, not provenance (decision-0047). decision-0093: rule 1 is the constraint this record works inside -- a consulted reference may not fail a session closed -- and rule 2's overlay half is the check this record deliberately does NOT move. decision-0094: its D5 is the question this record answers, and answers in the direction D5 leaned; without D5's reservation there would be nothing reserved to rule on. decision-0095: its DECISION HEADLINE is the primary authority here -- "A delivery that has proved its own pointer dead says so, on every branch" -- and its D3 draws the line this record stands on, that eligibility and report are separable questions; its "What is not decided here" named this cell and this consumer. An earlier draft cited D3 alone and under-argued the case; review found the headline, which this repo treats as the ruling (decision-0095:3 does exactly that to decision-0094:68)
changes: [decision-0094, decision-0095]  # FOUR clauses across two records. TWO sit in a "What is not decided here" section and are no part of a ruling; TWO are inside rulings -- decision-0095's D6 and decision-0094's D5 -- and both are named at the end of this comment. The count reached four over four corpus-check rounds, each catching the same shape in a new place: a blanket certification standing beside a named exception. Round 2 found decision-0094 unmarked; round 3 found decision-0095's mark saying "THE RULING IS UNTOUCHED" while marking a D6 clause false; round 4 found decision-0094's mark overtaking a D5 clause while this record still said no clause of 0094 was marked; round 5 found THIS FIELD still saying THREE while the Consequences below said four. Reconciled here, and the earlier drafts' reasoning is not preserved in this field because it was superseded rather than merely refined. decision-0095: "The vendored branch where the overlay's OWN copy is present but unusable ... still keeps its token and still says nothing." decision-0094: the SECOND bullet of its own such list -- "keeps the token, names its own empty file, and says nothing" -- together with that paragraph's opening count, "Two shapes stay silent", which after decision-0095 closed the first bullet and this record the second is now ZERO. Each predecessor's own superseded_in_part_by carries its scope. NAMING 0094 HERE IS A CORRECTION CAUGHT IN REVIEW: an earlier draft named only 0095 and excused the omission on the ground that decision-0094:5 "reserved a question rather than asserting a claim". That is true of D5's RESERVATION, and irrelevant to the bullet -- the bullet is a different clause in a different section, and it DOES assert a claim that this change falsifies. It is also NOT true of D5's CLOSING SENTENCE, which asserts a claim of its own and is now false; the FOURTH clause below marks it. An earlier draft of this field endorsed "true of D5" unqualified, which was the same blanket-beside-an-exception shape it exists to record. And decision-0095 had already marked 0094 for bullet ONE of the very same list, on the very same reasoning, so leaving bullet two unmarked was an inconsistency rather than a judgement. decision-0093 records the repo paying for exactly this shape before: "The claim was wrong in two places and marked in neither." THE THIRD AND FOURTH CLAUSES are the ones inside rulings. THIRD, decision-0095's D6 asserts "so Claude never delivers an invariants pointer into a vendored project and has nothing to report about", which this record's Consequences measure false -- and false WHEN WRITTEN, not falsified here, since the static import chain predates it; TRL-76 carries the fix. FOURTH, decision-0094's D5 closes "Whether an overlay's unusable copy should be substituted for is a different ruling, AND NOBODY HAS MADE IT" -- and D1 below makes it, so that closing is overtaken. The reservation itself is answered rather than falsified, but the sentence asserting nobody had answered it IS a claim and it is now false. The rest of D5 and D6 stands and is relied on, and each predecessor's own superseded_in_part_by carries the precise scope. Declared as convention, not requirement -- decision-0045:144-145 is permissive ("MAY declare which artifact(s) it changes"; the clause heading sits at :143, and the TWO earlier records in this chain that cite the span -- decision-0094 and decision-0095 -- both cite it one line short. Two, not three: decision-0093 cites decision-0045 nowhere and has no changes: field at all, as decision-0094's own precedent measurement records) -- and following the majority precedent decision-0094 measured and decision-0095 followed
informed_by: [decision-0028, decision-0078, decision-0087]  # provenance, not coupling (decision-0047): this record would still hold if each said something else. decision-0028 supplies the guard-per-pair reading applied to the shared phrase table and the payload-change-is-a-release rule; decision-0087 is why four classifications exist to be mirrored at all, but the ruling here turns on WHICH artifact may be blamed and which may be substituted, neither of which the gateway decides; decision-0078 is why decision-0094:5 and decision-0095 both named a consumer for this cell, which is the mechanism that brought it back
owner: agent
date: 2026-09-07
---

> **Provenance.** `TRL-73`, named as the consumer of this cell by **both** `decision-0094`'s
> *"What is not decided here"* — in its second bullet, **not** in D5, which reserves the
> substitution question without naming a consumer — and `decision-0095`'s section of the same name
> (`decision-0078`). The prior art was re-derived from the records rather than carried over, and
> the cost was **measured before choosing** rather than inherited: the discipline `decision-0095`
> established after finding `decision-0094`'s deferral resting on an estimate measurement did not
> support.
>
> *No claim of session independence is made here, and an earlier draft's — "written by a later
> session, which authored none of `#284`, `#294`, `#295`, `#296` or `#297`" — was removed in review
> as unverifiable: `#297`'s commit carries the same `Claude-Session` trailer this one does, so the
> assertion is contradicted by the only record that could check it. `decision-0095` makes the same
> claim with the same contradiction; that is `decision-0095`'s own to correct, and `TRL-79` carries it rather than this record fixing it
> here. What can honestly be said is the sentence above — what this record re-derived and what it
> re-measured — which is the property the claim was reaching for anyway.*

# An authoritative overlay is reported on, never substituted for

## Context

Five merged changes took the invariants pointer from *"Codex names a file nothing installs"* to
*"both hosts classify a broken consulted reference"*. One cell survived all five: a **vendored**
project whose **own** `.trellis/internal/invariants.md` is present but unusable — zero-byte,
newline-only, NUL-filled, or mode 0000.

`existingFile(overlayInvariants)` is a bare `statSync().isFile()`, which needs no read permission
and succeeds on every one of those. So the fallback arm does not fire, the pointer stays on the
overlay's own dead file, and **nothing is reported**. `decision-0095`:3 pinned that silence
deliberately, in `TestCodexSaysNothingWhenTheOverlayCarriesItsOwnInvariants`, across all four
shapes.

The vendored state table, before this record:

| the overlay's own copy | the plugin's copy | pointer | report |
|---|---|---|---|
| healthy | any | the overlay's own | none (correct) |
| absent | healthy | the plugin's | none (correct) |
| absent | unusable | the overlay's own | reports (`decision-0095`) |
| **unusable** | **healthy** | the overlay's own | **none — this record** |
| **unusable** | **unusable** | the overlay's own | **none — this record** |

**Two readings are defensible and the tension is real.** `decision-0094`:5 reserved the question
in terms that lean toward silence — *"a zero-byte file at the authoritative path is still the
project's own file"* — and `plugins/trellis/README.md`:27 holds that where a vendored
`.trellis/internal/` exists **it remains authoritative**, with *"the plugin's `reference/` files
stay installation sources rather than runtime **substitutes**"*. On that reading, remarking on the
overlay's copy is commentary on the project's own content rather than on a broken install.

**Measured on this tree before choosing.** Wiring a bare report onto the arm and running
`go test -count=1 ./...`:

| | measured | `decision-0095`'s cell, for contrast |
|---|---|---|
| hook invocations that fire the report | **4** | 58 |
| top-level tests that fail | **1** | 5 |
| named subtests that fail | **4** | 4 |
| failures that are byte-budget overruns | **0** | 0 |
| delivered `context`, every overlay shape | **7908 bytes, constant** | 7908 |

The blast radius is small for a reason worth stating rather than discovering later:
`writeValidCodexOverlay` writes no `invariants.md`, so **no existing fixture is in this cell** —
the only four firings are inside the guard `#297` built for it. A first run also showed 27
install/vendor failures; every one is a bundle-checksum artifact that clears once
`TRELLIS_BUNDLE_MANIFEST` is regenerated, which any payload change requires anyway.

*Every figure above is a measurement of **the tree as it stood before this change**, which is what
a cost weighed before choosing has to be. Two of them do not survive the change that follows, and
review was right to check: the rewritten guard runs each shape **twice** — against a healthy plugin
copy and a missing one — so the same instrumentation returns **8** firings on the merged tree, and
the failing-test count is meaningless once the test it counts has been rewritten. What is stable is
the shape of the answer: every firing is inside this cell's own guard, and no other fixture reaches
it.*

## Decision

**Reporting on a file and substituting for it are two questions. The answer to the first is yes,
to the second no, and `decision-0095` already drew the line between them.**

1. **The reserved substitution question is answered NO, and the eligibility check does not move.**
   `decision-0094`:5 reserved *"whether an overlay's unusable copy should be substituted for"*.
   This record answers it, and the ground is the one `decision-0094`:5 already named: **the
   project has its own authoritative location for this file, and it is occupied.** Substitution
   presupposes something to substitute *for* (`decision-0093`:3); here there is. Repointing would
   move a vendored project off the address it chose onto one the plugin controls — the runtime
   substitution `README.md`:27 forbids — and it would do so **invisibly**, where the dead pointer
   it replaces is at least visible to whoever opens it. `existingFile` stays a presence test, at
   the same call site, unchanged. The question is now answered for the four shapes `TRL-73` names
   rather than reserved indefinitely — **and for those four only**, which review forced this record
   to say out loud: a dangling symlink, a directory or a FIFO at the overlay path all fail
   `existingFile`, take the fallback, and are substituted for silently today. Those are named in
   *"What is not decided here"* rather than swept into the word "closed".

   *The sharpest instance is the mode-0000 shape, and it is stated as an instance rather than as
   the whole reason, because review found an earlier draft resting the entire ruling on it.* There
   the file has contents the hook cannot read, so it **may hold a different version of the
   invariants** and substituting would silently change which invariants govern. That argument does
   **not** reach the zero-byte, newline-only and NUL-filled shapes — those hold nothing, and the
   hook has just proved it — which is exactly why the ruling rests on the address rather than on
   the bytes. An overlay that owns the location owns it whether or not this session found anything
   at it.

2. **The report fires, it blames the overlay alone, and it offers no reinstall.** In
   `decision-0095`'s cell the plugin's copy was an **eligible substitute that happened to be
   unavailable**, so naming it and classifying it told the reader whether reinstalling would even
   have helped. Here it is **ineligible whatever state it is in** — by (1), not by availability —
   so offering the reinstall would send the reader to a remedy that *cannot work* — reinstalling
   moves nothing, because the pointer will still name the overlay's address afterwards. **That half
   is forced.** Withholding the plugin copy's *address* is the other half and is a **judgement, not
   a consequence** — an earlier draft called the whole difference "forced" and review was right to
   reject that. An operator copying that file into the overlay would be performing an explicit,
   visible act, the opposite of the invisible substitution D1 refuses. It is withheld because the
   copy may be a different version from the one the project vendored, so naming it as a source
   invites the version skew D1 exists to prevent — a judgement about advice, and reversible if it
   proves wrong. The guard states the forced half as a check rather than as prose: **each shape
   runs the hook twice, against a healthy plugin copy and a missing one, and the report must be
   byte-identical both times.**

3. **`decision-0095`'s own Decision headline had already ruled this, and the objection does not
   survive what the predicate can see.** The headline, above its numbered points, is *"A delivery
   that has proved its own pointer dead says so, **on every branch**. Which artifact it blames is
   what the branch decides."* This repo treats a headline as the ruling — `decision-0095`:3 does
   exactly that to `decision-0094`:68, and the hook's own comment repeats it (*"the headline, not
   point 1"*). By that precedent `decision-0095` had already decided the question and then left
   this cell alone in its *"What is not decided here"*, which is a deferral rather than a
   counter-ruling. *An earlier draft leaned on D3 alone and under-argued its own case; review
   found the headline.* The objection itself — that remarking on the overlay's copy is commentary
   on the project's content — does not survive what the predicate can see. `payloadDefect` reads no content. It answers one question — does this path yield a
   byte that is not a newline or a NUL — so it cannot distinguish invariants the project meant to
   write from a file that yields nothing. A zero-byte `invariants.md` is not an editorial choice
   about invariants; and the hook has **just told the model to read it**, in the last line of the
   context it assembled this run. That makes the report a statement about **this delivery**, which
   is the hook's own, not about the project's prose.

4. **The decisive shape is the adjacency, and it is what tips two defensible readings.** `absent`
   beside an unusable plugin copy reports; `zero-byte` beside the same plugin copy was silent.
   Both hand the model the same dead token, both yield nothing to read, and both have the same
   **primary** remedy — put a readable `invariants.md` at the overlay path. The only thing between
   them is `statSync().isFile()`, which asks whether a **regular file is at that path** and is
   silent about **diagnosis** — whether the address yields anything. Reporting on one and not the
   other makes what an operator is told turn on a distinction that does not reach them.

   *Two precisions review forced, both of which narrow this argument rather than widening it.* The
   remedies are not identical: `decision-0095`'s report offers a second repair (*"or reinstalling
   the Trellis plugin so its copy can stand in"*) which D2 rules out here, so the cells share the
   repair that matters and differ on the one that only works there. And `isFile()` is **not** a
   test of who owns the address — an earlier draft said it was. It cannot be: a dangling symlink, a
   directory or a FIFO at the overlay path all fail it and take the fallback, so a project that
   symlinks its `invariants.md` into a submodule is substituted for today, silently, exactly as D1
   says it should not be. That behaviour is pre-existing and this record does not touch it; see
   *"What is not decided here"*. The worst row is `unusable` × **healthy** plugin copy: the
   pointer is dead, a usable copy sits one directory away, (1) correctly forbids using it, and the
   only person who can repair it was told nothing.

5. **The vocabulary is the shared one and the guard is the existing one, renamed.** The lead
   `Trellis warning: this project's vendored overlay has no readable <absolute path>`, the
   `(it <defect>)` classification in `payload_read`'s four phrases verbatim, and *"yields nothing
   to read"* rather than *"cannot be read"* — false for an empty file, which reads fine (`#295`).
   Both vendored reports blame the overlay, so both open the same way (`decision-0028`'s
   guard-per-pair, applied to a string table living in two files in two languages). No parallel
   guard is written: `TestCodexSaysNothingWhenTheOverlayCarriesItsOwnInvariants` keeps its
   fixture, its shape table and its subject and becomes
   `TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting`, because a name asserting
   silence would now be false for four of its five rows.

6. **`decision-0093`:1 is obeyed, and on this host structurally.** The report rides
   `systemMessage`, outside the `MAX_CONTEXT_BYTES` bound on `context`. Measured across all five
   overlay shapes, the delivered context is **7908 bytes — the same on the healthy row as on the
   four broken ones**, because the pointer keeps the 32-byte token rather than an absolute path.
   That is the figure `decision-0095`:5 measured on its own cell and it reproduces here, which is
   expected: neither cell moves the pointer. The report cannot grow the context, so it cannot tip
   a governed session into a refusal, and a session that meets no ambiguous rule is unaffected
   either way.

**What is not decided here.** Three things, each filed rather than left in this paragraph
(`decision-0078`).

- **The Claude divergence this change opens**, described in Consequences below: Claude's static
  chain delivers the same pointer and says nothing about it. Different host, different transport,
  its own argument to make. `TRL-76` carries it, and notes that `decision-0095`:6's version of the
  false claim sits in its **Decision** rather than in Consequences, so correcting it is a
  supersession question rather than a bookkeeping one.
- **The shapes `existingFile` reads as absent while the project arguably owns the address** — a
  dangling symlink, a directory, a FIFO at `.trellis/internal/invariants.md`. All three take
  `decision-0093`:2's fallback and are substituted for silently when the plugin's copy is healthy,
  which is the harm D1 refuses for the four shapes it does reach. The behaviour is pre-existing and
  untouched here; deciding it means revisiting what `existingFile` is asking, which is D1's
  question again on a wider set. `TRL-77` carries it, and records the real case for the current
  reading — those three yield nothing whatever they are called, and `decision-0093` holds a
  resolving pointer beats a dead one — so it may close as *working as intended, now stated*.
- **The pointer-presence gap on the two older report arms.** This record's arm now requires the
  invariants token to be in the prose about to ship before it claims one was delivered;
  `decision-0095`'s arm and the plugin-native one make the same claim without the same check.
  Pre-existing, same one-line shape, not fixed here because each is a different arm with its own
  fixture. `TRL-78` carries it, and separates the two: the vendored arm reads the **project's**
  prose and is genuinely exposed, while the plugin-native arm reads the payload generator's own
  output, where the token is present by construction and the check would guard a regression rather
  than fix a defect.

Whether the vendored-overlay path should exist at all is still
`decision-0072`'s open question and `TRL-31`'s, untouched — this record makes one more cell of
that path honest, not permanent.

**`staleness.sh` is untouched, and the reason an earlier draft gave for that was false.** It said
*"Claude never delivers an invariants pointer into a vendored project and has nothing to report
about"*, inheriting the claim from `decision-0095`:6. Measured, the **hook** does not; the
**delivery** does. `reference/block-claude.md` imports `@.trellis/internal/trellis.md`, and
`reference/trellis-a.md` **and** `reference/trellis-b.md` — the two posture renders, either of which
a vendored overlay copies to that path — each carry the invariants pointer exactly once, on a line
the two files share. *An earlier draft called `trellis-a.md` the only payload file that does;
review measured that false, and the correction strengthens the point: the pointer reaches a
vendored Claude session under **either** posture.* So a Claude session in a vendored project
*is* handed the pointer, through the static import chain, while `staleness.sh` reads only
`version`, `trellis.md` and `rules.md` on that branch and exits.

**So this change opens a real host divergence rather than leaving none:** on a vendored project
whose own `invariants.md` is empty, Codex now warns and Claude stays silent about a pointer its own
delivery carried. `TestCodexDoesNotFallBackOnAnUnusablePluginCopy` cannot see it — it inspects the
hook's stdout, not the delivered session. That is not closed here, because it is a different host,
a different transport and a change with its own argument to make; it is **filed rather than left in
this paragraph** (`decision-0078`) as `TRL-76`, and the scope conclusion is unchanged while its
reason is now the true one.

## Consequences

- **The vendored state table has no silent dead-pointer cell left.** Every row either delivers a
  pointer that resolves or says which file does not and what would repair it.
- **`decision-0095` is corrected in two places: one disclosure and one clause inside its ruling.**
  *An earlier draft added "unlike its predecessors in this series", which is false in this record's
  own document — it marks a ruling clause on `decision-0094` too — and review caught it.* What is
  new to this series is marking a **ruling** clause at all: every earlier mark in the chain
  (`0094` on `0093`, `0095` on `0094`) touched only a Consequences or *"What is not decided here"*
  clause. The disclosure clause
  *"The vendored branch where the overlay's own copy is present but unusable — zero-byte,
  newline-only, NUL-filled or mode-0000 — still keeps its token and still says nothing"* sits in
  *"What is not decided here"*, not in the seven-point Decision, and it named `TRL-73` as the
  consumer that would come back for it. It has. The second is **D6**'s *"so Claude never delivers
  an invariants pointer into a vendored project and has nothing to report about"*, measured false
  below — and false **when written**, since the static import chain predates `decision-0095`, so it
  is an error discovered rather than a claim this change falsified. The instrument is the same
  either way, and the reason is the one `decision-0095`'s own mark gives: *"so a reader of 0095
  alone does not meet it as current truth"*. `decision-0095` now carries `decision-0096` in
  `superseded_in_part_by` for those two clauses, scoped there. **The rest of D6 stands and is
  relied on** — `staleness.sh`'s path A does exit before the repoint, the cell is Codex-only as far
  as the *hook* is concerned, and D6's reason for guarding it where it did is correct. What
  overreaches is the step from *"the hook does not deliver it"* to *"Claude does not"*.
  *An earlier draft of this record said `decision-0095`'s ruling was untouched while its own
  Consequences measured D6 false; the corpus check caught the contradiction.* **Its D3 is not corrected and is what this record
  applies:** *"This widens no eligibility check … What classifies is the report."* That sentence
  is the whole ground for answering the two questions in opposite directions, and its factual
  claim — that `TRL-73`'s four shapes satisfy `existingFile` and never reach `decision-0095`'s arm
  — remains true, because (1) leaves that check where it was. Only the test **name** D3 cites has
  moved, and D5 above states what it moved to.
- **`decision-0094`:5 is answered, and its closing clause is marked.** The two are not the same
  act. The *reservation* — "this is a different ruling" — is answered rather than falsified, and a
  reservation answered is not a claim that became false. But D5 **closes** with *"and nobody has
  made it"*, and D1 above makes it, so that sentence is a claim and it is now false: a reader of
  `decision-0094` alone would otherwise carry away that the substitution question is still open.
  `decision-0094` carries `decision-0096` in `superseded_in_part_by` for that clause alongside its
  *"What is not decided here"* bullet. *An earlier draft of this record said no clause of
  `decision-0094`:5 was marked while the mark on `decision-0094` itself said one was — the third
  blanket-certification-beside-a-named-exception the corpus check caught in this change, and the
  reason the `changes:` count is four rather than three.* D5's substantive prediction for the
  eligibility half stands and this record agrees with it.
- **The report is `decision-0095`'s sentence in its first half and deliberately not in its second.**
  The pair is behaviourally pinned rather than textually: `assertOverlayOwnCopyReport` requires the
  shared lead, the fixture's own classification, the shared claim and the absence of the wording
  `#295` corrected, and separately requires that the plugin root appear **nowhere** in the report
  and that no reinstall be offered. **No byte constant is claimed for the report**, deliberately:
  it embeds the overlay's absolute path, so its size tracks that path's length and a figure
  measured under `t.TempDir()` is a property of the fixture rather than of the report — the error
  `decision-0095`:5 refused to make and `decision-0094`:148-151 recorded paying for. What **is** reproducible is its
  shape: the report is exactly **519 bytes + the overlay path + the classification**, and the
  classification contributes either 8 bytes (`is empty`) or 98 (the unreadable phrase). *An earlier
  draft quoted "667–759 bytes across the classifications", which review showed is arithmetically
  impossible for any single fixture — the classification span is exactly 90 bytes, so a 92-byte
  range had silently mixed in two different temp-path lengths. The formula above reproduces all
  four measurements.* What is claimed is only that **none of it reaches `context`**.
- **The overlay's copy is now read on every vendored session, where before it was only stat'd.**
  `payloadDefect` scans a chunk at a time and stops at the first byte that is not a newline, so a
  healthy `invariants.md` costs one 4 KB read and nothing is retained — the same discipline the
  predicate already keeps for the plugin's copy. Learning that a file is non-empty cannot be done
  without reading a byte of it.
- **Two stats, and the report is guarded on the two defects that leave the file present.**
  `existingFile` and `payloadDefect` can disagree if the file changes between them, and
  `payloadDefect` has **four** non-empty answers rather than the two an earlier draft enumerated.
  Repaired between the calls yields `""` and nothing is reported, which is right because the
  pointer now resolves. `is missing` or `is not a readable file` means the file was deleted or
  replaced by a directory in the window — the condition that now holds is the arm *above* this one,
  which was not taken — and reporting anyway would ship a second sentence that is **false** in
  exactly those states, since with no file at the overlay path `decision-0093`:2's fallback does
  make the plugin's copy eligible and the withheld reinstall would be a real remedy. So the report
  fires only on `is empty` and `exists but could not be read`. *Review found this; an earlier draft
  called both directions correct after checking only the lead sentence.* Nothing can render the
  empty `(it )` that `decision-0095` fixed on the other arm. **This half is deliberately unpinned,
  and that is inherent rather than an omission**: the states it excludes are reachable only by
  mutating the filesystem between two `statSync` calls on one path, so no fixture can build one and
  relaxing the condition to `!== ""` survives the suite. Recorded because a guard that cannot exist
  is worth naming rather than leaving for a reader to find missing.
- **The report will not claim a pointer the delivery never carried.** It asserts as fact that *"the
  invariants pointer in the context just injected names a file that yields nothing to read"*, and
  on this branch the prose is the **project's own** `trellis.md`, which nothing validates beyond
  its `@rules.md` placeholder count. A vendored `trellis.md` with the pointer edited out would have
  drawn a report about a pointer nothing delivered — the right-diagnosis-wrong-artifact class this
  thread has spent five changes closing, reintroduced by the change that closes its last cell. The
  arm now also requires the token to be present in the prose about to ship, pinned by
  `TestCodexSaysNothingWhenTheVendoredProseCarriesNoPointer`. **The same gap is live on
  `decision-0095`'s arm and on the plugin-native one**, where it is pre-existing; `TRL-78` carries
  it, not fixed here.
- **The renamed guard's assertions are strictly stronger than the silence they replace.** Four
  rows stopped asserting `SystemMessage == ""` — a proxy for *"the eligibility half was not
  widened"* — and now assert that property **directly**: the delivered pointer must still name the
  overlay's own address. A mutation widening `existingFile` to `payloadDefect` moves that address
  to the plugin's copy and fails there. The **healthy** row keeps asserting silence untouched: it
  is `decision-0095`:3's real over-correction guard, added after a mutant firing the report on
  every vendored project survived the whole suite, and nothing else covers it.
- **`warningsBesideVendoredInvariants` needs no widening.** This report shares the sibling's lead
  and its trailing `is the likely fix.`, so the strip removes either; the two can never both fire,
  sitting on mutually exclusive arms; and no fixture the helper serves is in this cell, since
  `writeValidCodexOverlay` writes no `invariants.md`. Its comment is updated to say so rather than
  left to be rediscovered.
- A payload change, so a release: `VERSION` 0.21.0 → 0.22.0, both `plugin.json` manifests, and
  `install.sh`'s baked `TRELLIS_BUNDLE_MANIFEST` (`decision-0028`), regenerated from the actual
  files. **Five** files moved in the manifest — `VERSION`, both `plugin.json`s,
  `hooks/codex-context.mjs`, and `README.md`, which the review round added as the derivative
  `decision-0028` requires and an earlier version of this bullet failed to count.
  `reference/checksums` and `reference/version` hash `reference/` alone and did not move.
- `decision-0053` is not engaged: the shipped *prose* is unchanged. What changed is what one
  channel says **about** a file it points at.

## Self-check (gate)

- **Smaller thing that works?** Yes, and two smaller ones were rejected with reasons. *Leave the
  cell silent and record why* is the smallest, and it is a legitimate outcome — it is rejected on
  D4, because it makes what an operator is told turn on a distinction that never reaches them, and
  because the cost that made deferral attractive measured at 4 firings and 1 failing test. *Widen
  `existingFile` to `payloadDefect` and let the fallback fire* is one call site and is rejected on
  D1: it would silently change which invariants govern a project that vendored its own.
- **Grounded in an artifact?** The ruling is one `else if`-chain arm, one `warnings.push`, and a
  guard driven by the shape table that already existed. Every figure in Context is re-runnable
  against **the tree it names** — the pre-change one — and Context says which two do not survive
  the change, because review measured 8 firings where the table records 4. The byte figures in D6
  are from a fixture, not an estimate, and no constant is claimed for the report itself. Reverting the arm
  fails `TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting` on its four unusable
  shapes and leaves the healthy one green — the red this change was written against.
- **Reversible?** One arm and one `warnings.push`. Deleting both restores today's behaviour
  exactly; the guard would then fail loudly on the four rows rather than decaying quietly.
- **Intent gate.** Authored by an agent; the merge is the maintainer's act.

---
id: decision-0096
type: decision
depends_on: [decision-0093, decision-0094, decision-0095]  # coupling, not provenance (decision-0047). decision-0093: rule 1 is the constraint this record works inside -- a consulted reference may not fail a session closed -- and rule 2's overlay half is the check this record deliberately does NOT move. decision-0094: its D5 is the question this record answers, and answers in the direction D5 leaned; without D5's reservation there would be nothing reserved to rule on. decision-0095: its D3 drew the line this record stands on -- eligibility and report are separable questions -- and its "What is not decided here" named this cell and this consumer
changes: [decision-0095]  # ONE clause, in "What is not decided here", and no part of its ruling: "The vendored branch where the overlay's OWN copy is present but unusable ... still keeps its token and still says nothing." True when written and false after this change. decision-0095's own superseded_in_part_by carries the scope. Declared as convention, not requirement -- decision-0045:143-144 is permissive ("MAY declare which artifact(s) it changes") -- and following the majority precedent decision-0094 measured and decision-0095 followed
informed_by: [decision-0028, decision-0078, decision-0087]  # provenance, not coupling (decision-0047): this record would still hold if each said something else. decision-0028 supplies the guard-per-pair reading applied to the shared phrase table and the payload-change-is-a-release rule; decision-0087 is why four classifications exist to be mirrored at all, but the ruling here turns on WHICH artifact may be blamed and which may be substituted, neither of which the gateway decides; decision-0078 is why decision-0094:5 and decision-0095 both named a consumer for this cell, which is the mechanism that brought it back
owner: agent
date: 2026-09-07
---

> **Provenance.** `TRL-73`, named as the consumer of this cell by **both** `decision-0094`:5 and
> `decision-0095`'s *"What is not decided here"* (`decision-0078`). Written by a later session,
> which authored none of `#284`, `#294`, `#295`, `#296` or `#297`, and which measured the cost
> before choosing — the discipline `decision-0095` established after finding `decision-0094`'s
> deferral resting on an estimate measurement did not support.

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

## Decision

**Reporting on a file and substituting for it are two questions. The answer to the first is yes,
to the second no, and `decision-0095` already drew the line between them.**

1. **The reserved substitution question is answered NO, and the eligibility check does not move.**
   `decision-0094`:5 reserved *"whether an overlay's unusable copy should be substituted for"*.
   This record answers it, and answers it for a reason stronger than the one that deferred it: a
   vendored overlay is authoritative **precisely because it may hold a different version of the
   invariants from the plugin's**. Substituting the plugin's copy for a zero-byte overlay copy
   would silently change *which invariants govern* — the runtime substitution `README.md`:27
   forbids, and strictly worse than a dead pointer because it is invisible where a dead pointer is
   not. `existingFile` stays a presence test, at the same call site, unchanged. The question is
   now closed rather than reserved indefinitely.

2. **The report fires, it blames the overlay alone, and it offers no reinstall.** In
   `decision-0095`'s cell the plugin's copy was an **eligible substitute that happened to be
   unavailable**, so naming it and classifying it told the reader whether reinstalling would even
   have helped. Here it is **ineligible whatever state it is in** — by (1), not by availability —
   so naming it would be noise and offering the reinstall would send the reader to a remedy that
   *cannot work*. There is exactly one repair on this arm and it is the project's own file. The
   guard states this as a check rather than as prose: **each shape runs the hook twice, against a
   healthy plugin copy and a missing one, and the report must be byte-identical both times.**

3. **The "commentary on the project's own content" objection does not survive what the predicate
   can see.** `payloadDefect` reads no content. It answers one question — does this path yield a
   byte that is not a newline or a NUL — so it cannot distinguish invariants the project meant to
   write from a file that yields nothing. A zero-byte `invariants.md` is not an editorial choice
   about invariants; and the hook has **just told the model to read it**, in the last line of the
   context it assembled this run. That makes the report a statement about **this delivery**, which
   is the hook's own, not about the project's prose.

4. **The decisive shape is the adjacency, and it is what tips two defensible readings.** `absent`
   beside an unusable plugin copy reports; `zero-byte` beside the same plugin copy was silent.
   Both hand the model the same dead token, both yield nothing to read, and both have the same
   remedy. The only thing between them is `statSync().isFile()` — load-bearing for **eligibility**
   (does this project own the address?) and silent about **diagnosis** (does the address yield
   anything?). Reporting on one and not the other makes what an operator is told turn on a
   distinction that does not reach them. The worst row is `unusable` × **healthy** plugin copy: the
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

**What is not decided here.** Whether the vendored-overlay path should exist at all is still
`decision-0072`'s open question and `TRL-31`'s, untouched — this record makes one more cell of
that path honest, not permanent. `staleness.sh` is likewise untouched and needs no change: its
path A exits at the vendored branch long before the repoint, so Claude never delivers an
invariants pointer into a vendored project and has nothing to report about. That asymmetry is
already a check rather than a reading, in `TestCodexDoesNotFallBackOnAnUnusablePluginCopy`.

## Consequences

- **The vendored state table has no silent dead-pointer cell left.** Every row either delivers a
  pointer that resolves or says which file does not and what would repair it.
- **`decision-0095`'s disclosure of this cell is corrected; its ruling is untouched.** The clause
  *"The vendored branch where the overlay's own copy is present but unusable — zero-byte,
  newline-only, NUL-filled or mode-0000 — still keeps its token and still says nothing"* sits in
  *"What is not decided here"*, not in the seven-point Decision, and it named `TRL-73` as the
  consumer that would come back for it. It has. `decision-0095` now carries `decision-0096` in
  `superseded_in_part_by` for that clause alone. **Its D3 is not corrected and is what this record
  applies:** *"This widens no eligibility check … What classifies is the report."* That sentence
  is the whole ground for answering the two questions in opposite directions, and its factual
  claim — that `TRL-73`'s four shapes satisfy `existingFile` and never reach `decision-0095`'s arm
  — remains true, because (1) leaves that check where it was. Only the test **name** D3 cites has
  moved, and D5 above states what it moved to.
- **`decision-0094`:5 is answered rather than superseded.** It reserved a question and declined to
  decide it *"on its own authority"*; a reservation that is later answered is not a claim that
  became false, so no clause of it is marked. Its D5 predicted the outcome for the eligibility
  half and this record agrees with it.
- **The report is `decision-0095`'s sentence in its first half and deliberately not in its second.**
  The pair is behaviourally pinned rather than textually: `assertOverlayOwnCopyReport` requires the
  shared lead, the fixture's own classification, the shared claim and the absence of the wording
  `#295` corrected, and separately requires that the plugin root appear **nowhere** in the report
  and that no reinstall be offered. Measured on a fixture, the report is **667–759 bytes**
  depending on the classification and the overlay path's length, and **none of it reaches
  `context`**.
- **The overlay's copy is now read on every vendored session, where before it was only stat'd.**
  `payloadDefect` scans a chunk at a time and stops at the first byte that is not a newline, so a
  healthy `invariants.md` costs one 4 KB read and nothing is retained — the same discipline the
  predicate already keeps for the plugin's copy. Learning that a file is non-empty cannot be done
  without reading a byte of it.
- **Two stats, and the guard is on the defect rather than on the arm — the same correction the
  sibling arm already carries.** `existingFile` and `payloadDefect` can disagree if the file
  changes between them. Both directions resolve correctly: repaired between the calls yields `""`
  and nothing is reported, which is right because the pointer now resolves; deleted between them
  yields `is missing`, a true sentence about the address the model was handed. Neither can render
  the empty `(it )` that `decision-0095` fixed on the other arm.
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
  files. Only the four files this change touches moved in the manifest; `reference/checksums` and
  `reference/version` hash `reference/` alone and did not.
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
  from the suite; the byte figures in D6 are from a fixture, not an estimate. Reverting the arm
  fails `TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting` on its four unusable
  shapes and leaves the healthy one green — the red this change was written against.
- **Reversible?** One arm and one `warnings.push`. Deleting both restores today's behaviour
  exactly; the guard would then fail loudly on the four rows rather than decaying quietly.
- **Intent gate.** Authored by an agent; the merge is the maintainer's act.

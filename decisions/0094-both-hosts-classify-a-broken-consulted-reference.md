---
id: decision-0094
type: decision
depends_on: [decision-0093, decision-0087]  # coupling, not provenance (decision-0047). decision-0093: this record corrects one Consequences bullet of it and leans on its rule 1 and rule 2 for everything it does NOT change -- if 0093 said something else there would be nothing here to correct and no ground for keeping the pointer alive over a broken target. decision-0087: the gateway is the whole reason "relax the Claude host" is not available as an option -- if payload_read were optional on the invariants read, the cheaper symmetry would be reachable and this record would decide the other way
informed_by: [decision-0028, decision-0040, decision-0065, decision-0082]
owner: agent
date: 2026-09-06
---

> **Provenance.** `TRL-72`, filed by the session that merged `#295` from its own adversarial
> review of that change, rather than fixed inside it: different host, different file, and `#295`
> was already one logical change on the Claude arm. Written by a later session, which authored
> neither `#294` nor `#295`.

# Both hosts classify a broken consulted reference, not merely detect one

## Context

The invariants pointer defect was closed twice in one day, once per host, and the two fixes did
not land in the same place.

`#294` (`TRL-69`) taught `codex-context.mjs` to report a plugin payload whose
`reference/invariants.md` is not there. It asks `existingFile`, the file-shaped sibling of
`existingDirectory`:

```js
function existingFile(value) {
  if (typeof value !== "string" || !path.isAbsolute(value)) return false;
  try {
    return fs.statSync(value).isFile();
  } catch {
    return false;
  }
}
```

`#295` (`TRL-70`) closed the same gap on `staleness.sh`, which cannot ask a bare predicate:
`TestNoPayloadReadBypassesTheGateway` names `inv` among the `$plugin/`-rooted payload paths that
must go through `payload_read` (`decision-0087`). That gateway **opens** the file, and classifies
what it finds four ways.

**`statSync` needs no read permission**, so it succeeds on a zero-byte file and on one at mode
0000. Measured on the same fixtures, before this record:

| the plugin's `reference/invariants.md` | `staleness.sh` | `codex-context.mjs` |
|---|---|---|
| missing | `(it is missing)` | reports, unclassified |
| a directory at the path | `(it is not a readable file …)` | reports, unclassified |
| zero-byte | `(it is empty)` | **silent** |
| mode 0000 | `(it exists but could not be read …)` | **silent** |
| newlines only | `(it is empty)` | **silent** |

So Claude became the stricter host on a broken install, with
`TestBothHostsRepointTheInvariantsPointerIdentically` green throughout — that guard pins the
**address** the two hosts hand the model, never the **diagnosis**.

**One sentence appears to forbid closing this, and it is why the issue was filed rather than
fixed.** `decision-0093`'s Consequences record `existingFile`'s flattening as deliberate:

> `existingFile` joins `existingDirectory` with a deliberately identical contract: absolute paths
> only, every stat failure reading as absent. **No caller here can act on the difference between
> missing and unreadable.**

## Decision

**A host that cannot say *which* fault it found has not reported the fault.**

1. **Codex classifies; Claude does not relax.** The symmetry could be reached from either end, and
   only one end is available. Relaxing Claude cannot mean *stop opening the file* —
   `decision-0087`'s gateway is structural and `TestNoPayloadReadBypassesTheGateway` enforces it —
   so it can only mean opening the file, learning which of four faults it is, and **discarding
   that** to print a sentence that fits the host which never knew. That is a strictly worse
   product for no structural gain. The classification is the part an operator acts on:
   `payload_read`'s own reason for existing is that *"missing and unreadable are told apart
   because their remedies differ"*, and Codex's single remedy — reinstall — is the wrong advice
   for a file at mode 0000.

2. **The two hosts share one vocabulary, byte for byte.** The four phrases now live in two files
   in two languages, so `decision-0028`'s guard-per-pair applies and the pair is pinned
   **behaviourally**: `TestBothHostsReportAMissingInvariantsTarget` drives both hooks over every
   shape `payload_read` classifies and requires the same phrase from each, and no other. There is
   no line to diff between an `awk` substitution and a JS one; what can be compared is what they
   say.

3. **The bullet quoted above is corrected; `decision-0093`'s ruling is not touched.** It sits in
   **Consequences**, not in the four-point Decision, and its factual half was overtaken by `#294`
   in the same week it was written — that change added the first caller that *acts* on the
   difference by **reporting it to a person**, and its own comment says so
   (*"This is the first caller to report that result to a person, so it must not turn a
   permissions fault into a claim that the file is gone"*). `decision-0093` now carries
   `decision-0094` in `superseded_in_part_by` for that clause alone. **Rule 1 is untouched and is
   what this change obeys**: a consulted reference is not a delivered payload, so nothing here
   fails a session closed — both hosts still deliver a complete governing context on every fault
   in the table.

4. **The plugin-copy half of rule 2 moves to the classifying predicate, and that is rule 2
   *applied*, not amended.** `decision-0093` states that half's purpose exactly: it *"stops the
   fallback replacing one dead pointer with another"*. Measured, `existingFile` did not deliver
   it — a zero-byte or mode-0000 plugin copy satisfied the check, the fallback fired, and a
   vendored project's pointer was moved off its own authoritative address onto a file that yields
   nothing to read. Asking whether the copy is **usable** is what that sentence already asks for.

5. **The presence question keeps `existingFile`, and this record deliberately does not widen it.**
   The other half of rule 2's condition asks whether the project has its own authoritative copy of
   the file. That is presence, not usability: a zero-byte file at the authoritative path is still
   **the project's own file**, and `README.md:27` with `decision-0093`:3 keep the plugin's
   `reference/` an installation source rather than a runtime substitute for it. Whether an
   overlay's *unusable* copy should be substituted for is a different ruling, and nobody has made
   it.

**What is not decided here:** the vendored branch where **both** copies are unusable is still a
silently dead pointer. That is `TRL-71`, untouched — a report there would fire on ~28 fixtures and
perturb the byte-budget tests, which is a change with its own argument to make.

## Consequences

- Both hosts now name the same fault in the same words. `codex-context.mjs` gains `payloadDefect`,
  which mirrors `payload_read`'s four answers, and its warning reads
  `no readable <path> (it is empty)` where it used to say nothing at all.
- **Emptiness means what the other host means by it.** `payload_read` reaches *empty* through
  `$(cat …)` and `[ -z … ]`, and command substitution strips **trailing newlines** — so a file of
  nothing but newlines is empty, and a file holding one space is not. The obvious JS mirror,
  `readFileSync(…).length === 0`, agrees on the zero-byte file and diverges on the newline-only
  one; the guard carries that case for exactly this reason.
- **The report still costs the injected context nothing on Codex.** Measured across all six
  states — healthy plus five faults — `additionalContext` is **7971 bytes each time**, against the
  9500-byte bound: the warning rides `systemMessage`, outside it. That is the one way a report
  about a *consulted* file could fail a session closed (rule 1), and the pair guard now pins it
  structurally by requiring the broken delivery to be byte-identical to the healthy one, rather
  than waiting for a payload large enough to notice.
- **`TRL-71`'s population widens slightly, and this is stated rather than discovered later.** With
  the plugin copy now checked for usability, a vendored overlay lacking `invariants.md` beside a
  plugin copy that is *unusable* no longer takes the fallback, so it joins the both-missing case
  as a silent dead pointer. Each such session previously got a pointer to a file that yields
  nothing; it now keeps the overlay's own address, which on that branch is not a lie — the project
  really does have an overlay. Neither is good, which is what `TRL-71` is for.
- `existingFile` keeps one caller and its contract stays true for that caller. The predicate is
  not deleted, and `existingDirectory` is untouched.
- A payload change, so a release: `VERSION` 0.19.0 → 0.20.0, both `plugin.json` manifests, and
  `install.sh`'s baked `TRELLIS_BUNDLE_MANIFEST` (`decision-0028`). Only the four files this
  change touches moved in the manifest; `reference/checksums` and `reference/version` hash
  `reference/` alone and did not.
- `decision-0053` is not engaged: the shipped *prose* is unchanged. What changed is what one
  channel says **about** a file it points at.

## Self-check (gate)

- **Smaller thing that works?** Yes, and two smaller ones were rejected with reasons. *Relax
  `staleness.sh`* is the smaller diff and is answered in D1 — it buys symmetry by throwing away
  the answer the gateway already computed. *Classify only at the report and leave the vendored
  eligibility check on `existingFile`* is smaller still, and leaves a call site whose stated
  purpose in `decision-0093`:2 is defeated by the same two shapes this record exists to close.
- **Grounded in an artifact?** The ruling is a predicate plus a pair guard driven by a shared
  fault table, not prose about a rule. Every row of the Context table is a test case.
- **Reversible?** Two call sites and one function. Restoring `existingFile` at both restores
  today's behaviour exactly.
- **Intent gate.** Authored by an agent; the merge is the maintainer's act.

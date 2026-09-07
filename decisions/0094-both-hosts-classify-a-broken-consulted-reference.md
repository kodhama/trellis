---
id: decision-0094
type: decision
depends_on: [decision-0093, decision-0087]  # coupling, not provenance (decision-0047). decision-0093: this record corrects one Consequences bullet of it and leans on its rule 1 and rule 2 for everything it does NOT change -- if 0093 said something else there would be nothing here to correct and no ground for keeping the pointer alive over a broken target. decision-0087: the gateway is the whole reason "relax the Claude host" is not available as an option -- if payload_read were optional on the invariants read, the cheaper symmetry would be reachable and this record would decide the other way
changes: [decision-0093]  # TWO of its Consequences clauses and no part of its ruling — the closing sentence of the existingFile bullet, and the first bullet's "handed a pointer that resolves"; decision-0093's own superseded_in_part_by carries the full scope. Declared as convention, not requirement: decision-0045:143-144 is permissive ("MAY declare which artifact(s) it changes"). On the precedent, measured over every record on main: of the eleven most recent successors that partially supersede a predecessor, SEVEN declare `changes:` naming it (0085/0079, 0087/0043, 0088/0068, 0089/0078, 0090/0088, 0091/0068, 0092/0089) and FOUR do not (0083, 0084, 0086, and 0093 itself). An earlier version of this comment cited 0079/0011 and 0082/0080 as examples: both are FULL supersessions — 0011 and 0080 carry `superseded_by` — so neither exemplifies the pattern, and 0082's own `changes:` names 0080 rather than any of its five partial predecessors. Review caught that much; re-measuring caught the rest (0091 declares and was missing from the list, 0083 and 0093 do not and were missing from the exceptions)
informed_by: [decision-0028, decision-0040, decision-0065, decision-0082]
superseded_in_part_by: [decision-0095, decision-0096]  # 2026-09-07 decision-0095 — ONE clause, in "What is not decided here", and no part of the ruling: "The vendored branch where both copies are unusable is still a silently dead pointer." True when written; false after decision-0095, which closes that cell by reporting it on systemMessage. The clause named TRL-71 as the consumer that would come back for it (decision-0078), and it did. Marked here so a reader of 0094 alone does not meet it as current truth. The cost estimate in the same sentence — "a report there would fire on ~28 fixtures and perturb the byte-budget tests" — is deliberately NOT edited: decision-0095's Context measures it (58 firings, 5 failing tests, ZERO byte-budget overruns, zero fixtures needing a reference/invariants.md) and answers it there, because what this record believed when it deferred the cell is part of why it deferred it. THE RULING IS UNTOUCHED: D1-D3 are unrelated, D4's widening of this cell is what made it reachable in six shapes rather than two and is relied on, and D5's refusal to widen the presence check is obeyed — decision-0095's arm runs only where the overlay has no copy at all, so TRL-73's shape stays untouched. // 2026-09-07 decision-0096 — the SECOND clause of the same "What is not decided here" list, and again no part of the ruling: "The vendored branch where the overlay's own copy is present but unusable — zero-byte, newline-only, NUL-filled or mode-0000 — keeps the token, names its own empty file, and says nothing." The closing "and says nothing" is false after decision-0096, and so is the paragraph's opening count, "Two shapes stay silent" — after decision-0095 closed the first bullet and decision-0096 the second, ZERO stay silent. The clause named TRL-73 as the consumer that would come back for it (decision-0078), and it did. Marked here for the same reason decision-0095 marked its sibling bullet: so a reader of 0094 alone does not meet it as current truth. The rest of the bullet stands as history — the behaviour WAS bit-for-bit unchanged by 0094, and D5 WAS why 0094 did not widen the presence check. THE RULING IS NOT OVERTURNED, AND ONE CLAUSE OF D5 IS OVERTAKEN -- its closing "and nobody has made it", said of the substitution ruling, is false once decision-0096 makes it. Stated precisely rather than as a flat "untouched", for the same reason the corpus check rejected that phrasing on decision-0095: a blanket certification beside a named exception is a contradiction in one field. D5 is otherwise obeyed rather than overturned: decision-0096 answers the question D5 reserved by REFUSING the substitution D5 declined to authorise, and leaves the existingFile presence test at that call site byte-identical. What decision-0096 changes is only what is SAID about the file, which D5 does not govern.
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

**What is not decided here, and where each piece lives.** Two shapes stay silent, and neither is a
deferral this record creates — both are pre-existing behaviour it leaves exactly as it found it.

- The vendored branch where **both** copies are unusable is still a silently dead pointer. That is
  `TRL-71`, untouched: a report there would fire on ~28 fixtures and perturb the byte-budget
  tests, which is a change with its own argument to make.
- The vendored branch where the **overlay's own** copy is present but unusable — zero-byte,
  newline-only, NUL-filled or mode-0000 — keeps the token, names its own empty file, and says
  nothing. D5 is why this record does not widen the presence check on its own authority, and the
  behaviour is **bit-for-bit what it was before this change**, since that half keeps
  `existingFile`. `TRL-73` carries it, so the shape has a consumer rather than living in this
  paragraph (`decision-0078`).

## Consequences

- Both hosts now name the same fault in the same words. `codex-context.mjs` gains `payloadDefect`,
  which mirrors `payload_read`'s four answers, and its warning reads
  `no readable <path> (it is empty)` where it used to say nothing at all.
- **Emptiness means what the other host means by it, and that is two properties of `$( )`, not
  one.** `payload_read` reaches *empty* through `$(cat …)` and `[ -z … ]`, and command
  substitution strips **trailing newlines** *and* **NUL bytes** — the second because a shell
  variable holds a C string. So the file is empty over there exactly when every byte is a newline
  or a NUL. Measured on bash (the sibling hook's shebang), `sh` and `dash`: an all-NUL file is
  empty on all three; zsh alone disagrees and runs neither hook. The obvious JS mirror,
  `readFileSync(…).length === 0`, agrees on the zero-byte file and diverges on **both** the
  newline-only and the NUL-filled one — the latter the classic post-crash zero-fill. The guard
  carries both cases. The **mixed** shapes (`a\0b`, `\0a`) are deliberately left unpinned: they
  agree on every shell measured, but POSIX leaves NUL handling in `$( )` unspecified, and pinning
  them would assert a portability property nobody has checked.
- **The report costs the injected context nothing on Codex.** The warning rides `systemMessage`,
  outside the `MAX_CONTEXT_BYTES` bound on `context` — which is the one way a report about a
  *consulted* file could fail a session closed (rule 1). The pair guard pins that structurally:
  on each fixture the broken delivery must be **byte-identical to the healthy delivery on that
  same fixture**, which fails the moment the warning moves into the context. *No absolute byte
  count is claimed, deliberately.* The injected context embeds the plugin root's absolute path, so
  its size tracks the length of that path — on one local fixture every state measured 7971 bytes,
  and under the suite's own `t.TempDir()` roots the same states measure 8025–8027, differing by
  exactly the path length. A cross-state constant would have been the same class of error as this
  series' earlier 32573-characters-quoted-as-bytes.
- **`TRL-71`'s population widens slightly, and this is stated rather than discovered later.** With
  the plugin copy now checked for usability, a vendored overlay lacking `invariants.md` beside a
  plugin copy that is *unusable* no longer takes the fallback, so it joins the both-missing case
  as a silent dead pointer. Each such session previously got a pointer to a file that yields
  nothing; it now keeps the overlay's own address, which on that branch is not a lie — the project
  really does have an overlay. Neither is good, which is what `TRL-71` is for. Measured, the
  vendored state table moves in **exactly four cells** — overlay absent × plugin copy
  {zero-byte, newline-only, NUL-filled, mode-0000} — and every overlay-present row is untouched,
  as is `absent × missing` and `absent × a single space`. *An earlier draft of this bullet said
  **two**, counting only the two shapes the probe behind it happened to build; the other two
  unusable shapes are named in the Consequences bullet directly above and in `TRL-73`, so the
  record contradicted itself. Corrected in review, and the count is now the one
  `TestCodexDoesNotFallBackOnAnUnusablePluginCopy` fails on when D4 is reverted: four subtests.*
- **D4 has its own guard, because the existing vendored fixture cannot see it.**
  `TestCodexRepointsWhenAVendoredOverlayLacksInvariants` builds a *healthy* plugin root, so the
  second half of the fallback condition is satisfied in every case it runs and is never exercised;
  reverting D4's one call to `existingFile` survives that test and the rest of the suite.
  `TestCodexDoesNotFallBackOnAnUnusablePluginCopy` is the fixture that kills it, driven by the same
  fault table. Recorded because the repo has paid for this shape before: `TRL-52` shipped a broken
  pointer on a branch no fixture reached.
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
- **Grounded in an artifact?** The ruling is a predicate plus two guards driven by one shared
  fault table, not prose about a rule. Every row of the Context table is a case in
  `TestBothHostsReportAMissingInvariantsTarget`, run against **both** hosts. D4 is the row that
  table cannot reach, and review caught it ungrounded: reverting D4 alone survived the whole
  suite, because the only fixture on that arm builds a healthy plugin root.
  `TestCodexDoesNotFallBackOnAnUnusablePluginCopy` was added for it and kills that mutant.
- **Reversible?** Two call sites and one function. Restoring `existingFile` at both restores
  today's behaviour exactly.
- **Intent gate.** Authored by an agent; the merge is the maintainer's act.

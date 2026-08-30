---
id: decision-0083
type: decision
depends_on: [decision-0051, decision-0070, decision-0072, decision-0074]
informed_by: [decision-0008, decision-0028, decision-0053, decision-0065, decision-0078, decision-0081, decision-0082]
owner: agent
date: 2026-08-30
---

> **Provenance.** Directed by the maintainer as **TRL-20** (umbrella) with **TRL-2** and **TRL-27**
> as its other two arms, designed through `superpowers:brainstorming` and executed through
> `superpowers:executing-plans` on `feature/trl-20-rules-toml-self-repair`. His framing, adopted
> verbatim by the design: **correction is not the hazard — silence is.**
>
> **This record is written after the code shipped, and reports what shipped rather than what was
> planned.** Two of its load-bearing sections (the invented byte cap, the five tests that passed for
> the wrong reason) exist only because execution found them; neither was in the design. Where
> execution overturned the plan, the record follows execution and says so.

# 0083 — `.trellis/rules.toml` reconciles itself; the repair is reported, not gated

## Context

**A single unmatched row cost all sixteen rules, every session, until a human edited the file.**
The hook validated the project's row set against the slugs the installed payload ships, and on any
mismatch injected **nothing** — `plugins/trellis/hooks/staleness.sh`, the message this change
retires:

> `TRELLIS_RULES_NOT_LOADED — this project's .trellis/rules.toml does not match the rules the
> installed plugin ships ($slug_report). Nothing was injected, because a partial or unknown row set
> cannot be applied honestly. … for missing:, add those slugs; for unknown:, remove those rows; for
> duplicate:, delete the extra occurrences and keep the one whose value they intend.`

Three filed issues are three arms of that one defect:

| Issue | Arm |
|---|---|
| **TRL-20** | A catalog-changing plugin upgrade leaves every existing project one row short — blacked out every session until someone edits the file. |
| **TRL-2** | On the curl path, path C re-renders and then *claims* governance while a row is missing. Delivery and the row set could disagree. |
| **TRL-27** | `unknown:` has two causes and opposite repairs, and the remedy named only the destructive one. |

**TRL-27's live case is the one that decides the shape.** `decision-0078` minted
`inv-no-orphan-followups` and released it in `VERSION` 0.5.0 → 0.6.0. A project that took that row
while still running a 0.5.0 plugin gets `unknown: inv-no-orphan-followups` — and the retired remedy
above says *"for unknown:, remove those rows"*. Following it **deletes a ratified row to match a
stale plugin**, and the blackout returns on the next plugin update. The two causes — the rule was
retired, or the installed plugin is behind the project's config — are **indistinguishable at
runtime** in config-only mode, which carries no version stamp to compare against.

## Decision

### 1. Reconciliation replaces refusal on the **Claude** hook, in memory, for delivery

On a mismatch the hook reconciles the project's rows against the shipped set and injects the
**reconciled** block, labelled as reconciled, with the readout above it. It never `cat`s the raw
file in this branch: delivered rows that disagree with the authority header would be worse than the
blackout they replace.

**This is `plugins/trellis/hooks/staleness.sh` and only that.** The design brief said of Codex's
`parseRulesToml` that *"it adopts the reconcile semantics instead"*, and described the reconciled
table as *"one table, applied identically by both hosts"*. **Neither shipped.**
`codex-context.mjs` still returns `null` from `parseRulesToml` on any mismatch and calls
`fail(PROJECT_CONFIG, "invalid-rules")`, injecting nothing — for a missing row, an unknown row and
a duplicate alike. What the Codex hook gained in this change is §6 (its slug set derived from the
payload rather than hardcoded) and a raised context cap; **the blackout this record retires is
retired on Claude only.** Stated here rather than left to be inferred from §6's narrower title,
because a reader who took the design's parity claim as shipped would believe a Codex project is
covered when it is not. The gap is carried as an open question below, not as a silent remainder.

| Kind | Resolution |
|---|---|
| `missing:` | add the row, `active = true` — matching both shipped presets and `decision-0070` D3, where a project-scope plugin with **no file at all** governs at full strength. A newly ratified invariant then behaves the same in a project installed today and one installed two releases ago. |
| `unknown:` | **quarantine** — comment the row out, keeping its value verbatim, with the date and the payload stamp. |
| `duplicate:` | keep the first occurrence, quarantine the extras — no "which value did they intend" judgment call. |

Floor rows apply regardless of their value, unchanged. **The parser needed no grammar work:** the
validator anchors rows at line start, so a commented row is invisible to it — which is also why a
repaired file draws no second notice next session.

**One verdict is explicitly *not* reconciled.** `no-slugs-in-payload` means the validator found
nothing to check rows against — the *plugin's own* `rules.md` is unreadable or malformed. Feeding it
to the reconciler would quarantine every legitimate row and run the session ungoverned at exit 0,
inverting the fail-loud rule stated forty lines above the reconciler in the same file. It gets its
own loud branch, and names the plugin rather than the consumer's rows as the thing to fix.

### 2. Quarantine, not deletion — and that is the whole safety argument

**Quarantine is correct under both readings of `unknown:`**: inert clutter if the rule was retired,
one uncomment away if the plugin was the stale side. Nothing the consumer chose is ever lost, and
every repair is reversible **from the file itself, without git**.

That is not a nicety. It is the premise the next section rests on.

### 3. The gate is not being dropped; it is no longer engaged

`decision-0072` recorded the mechanism this change is most at risk of repeating:

> **Three of the six are the same mechanism: retiring a confirm-gated writer silently retires the
> gate.** `/trellis:setup` diffed and asked before replacing rows or deleting an overlay. Each
> remedy that replaced it inherited the action and not the confirmation.

**That gate guarded destructive writes** — clobbered rows, deleted rows, a preset copied over a
consumer's choices. Under quarantine semantics no prior value is ever lost: every repair is an
addition or a comment. `floor-intent-gate` is satisfied because nothing irreversible happens;
`floor-transparency` is satisfied because the agent must state what it reconciled, row by row,
before substantive work. So the repair is **announced, not asked**.

**The clause this most nearly touches, quoted rather than paraphrased.** `decision-0070` D4 reads:

> **The hook never writes.** It has one channel, injected context, and keeps it. Writes stay
> agent-mediated and human-consented, which is what makes this comply with `decision-0008` rather
> than merely resemble compliance.

The first two sentences are untouched and still pinned by test — the hook reconciles in memory and
instructs; the *agent* writes. **The third is materially re-read, and is marked rather than
narrowed in prose.** Consent for a reconciliation write moves from a per-write confirmation to
**the adoption act plus the session announcement**, because the write is non-destructive and
self-reversible. A reader of `decision-0070` who was not told that would believe per-write consent
still governs row repair — the silent divergence `inv-graph-maintenance` and `inv-auditable-archive`
exist to prevent — and under `decision-0082` the forward pointer is the only mark there is. So
`decision-0070` carries `superseded_in_part_by: [decision-0077, decision-0083]`, scoped in its
frontmatter comment to that clause **as it applies to row repair only**; *"the hook never writes"*,
the announcement, the decline bullet and D1–D3, D5–D7 all stand.

**This was the maintainer's call to make, not the author's** (`decision-0081` — supersession
authority scales with cost of reversal; a frontmatter line is cheap to reverse, so it is his and it
is reversible). The agent escalated it rather than deciding it. If a repair class ever becomes
destructive again, D4's clause re-engages in full and so does the confirm gate: the safety argument
is a property of quarantine, not a general licence.

**The argument is enforced by construction, not by care.** `TestEveryDeletionInstructionIsGated`
and `TestEveryDestructiveInstructionIsGated` scan hook messages for destructive verbs and require a
confirmation clause on each. Both matched only `^\s*emit "` — and the repair mandate is written
through the payload `printf` block, the one output channel neither guard saw. Both guards were
widened to scan it, so an ungated deletion verb entering the mandate fails the suite rather than
shipping. The widening was proven by mutation (inject one; watch it fail), not by reading.

### 4. What this supersedes — and what it does not

**Superseded: `decision-0072`'s point 2, in two respects and no others** — its confirm-first
row-repair remedy (the three mismatch shapes and the reseed gate, insofar as they describe how a row
mismatch is repaired), and its *"a hand-written partial file leaves the project **ungoverned**"*
premise, which is what made the preset copy mandatory. §5 below states the second in full, because
retiring a hazard silently is worse than the hazard. `decision-0072` carries
`superseded_in_part_by: [decision-0083]`.

**Standing, untouched:** its retirement of `/trellis:setup` (point 1), its accounting of what was
lost (point 3), the discoverability call (point 4) and the rejected `/trellis:migrate` (point 5).
Its finding #6 also stands as a rule — a confirm-gated writer's gate does not retire with it; what
this record argues is that the gate is not engaged here, not that the rule is wrong. Also standing:
`decision-0070` D4's *"the hook never writes"*, `decision-0053`'s live-rows authority, and
`decision-0065`'s vendorless plugin path.

### 5. What a partial file now means — `decision-0072`'s hazard is retired, and inverted

`decision-0072` finding #4 cost that record its first draft: *"a hand-written partial file leaves
the project **ungoverned**"*. That is why its point 2 made the preset copy **mandatory** rather than
advisory, and it was measured rather than reasoned — every one of the fourteen slugs came back
`missing:` and no rule was injected.

**Reconciliation retires that hazard by making a partial file self-complete.** Measured the same
way on the current tree: an empty `.trellis/rules.toml` yields `RECONCILED … added 16 row(s);
quarantined 0 row(s)`, with all sixteen rows delivered `active = true`.

**The practical effect is the opposite of the old advice, and this record states it rather than
leaving it to be inferred.** The old failure mode was **silent under-governance** — a stub file, and
no rules. The new one is **immediate full governance** at the **adaptive** posture (an empty file
carries no `strictness`, so the header falls to `trellis-b.md`) from a file its author may have
meant as a stub.

**The new mode is preferable, and not marginally.** Both are surprises; only one is silent.
Under-governance fails closed on the user's expectations and open on the rules — a project that
believes it is governed is not, and nothing says so. Over-governance is announced in the session it
happens, names every row it added, and leaves those rows **visible in the file**, where any of them
can be set `active = false`. A surprise you can see and edit is not the same class of defect as one
you cannot detect.

**What survives from point 2 is the advice, not the hazard.** Copy a complete preset when the
posture or the row set matters — reconciliation defaults, it does not read minds. What does not
survive is the reason: a partial file is no longer a governance blackout, so the copy is no longer
mandatory in order to be governed at all. Both READMEs' `governed = false` bullets are corrected to
match; they had kept the retired reason after the surrounding prose was updated.

### 6. Codex derives its slug set from the payload

`codex-context.mjs` validated against a **hardcoded sixteen-slug array**, while the Claude hook
derived its set from the shipped `reference/rules.md`. Nothing in CI compared the two. The
consequence is worse than duplication: on Codex a payload upgrade **could not** repair the drift,
and a stale array could manufacture an `unknown:` for a row the payload actually ships — the agent
would quarantine a live row and cite a reason that is false. Since this change's entire value is
that the loud message is trustworthy, the array is gone. Both `parseRulesToml`'s row validation and
the floor-warning list now derive from `reference/rules.md`, using the identical trailing-backtick
match the Claude hook already uses. The derived array is de-duplicated at its source: membership was
checked through a `Set` but completeness through `length`, so a `rules.md` that ever tagged one slug
twice would make **every** Codex project read `invalid-rules` while Claude governed normally from
the same file — the same blame-the-consumer mislabel this change exists to close.

The payload's own well-formedness check moved **ahead** of the derivation it feeds, for the same
reason: deriving first meant a broken plugin payload produced an empty slug set, the project's file
then failed to parse, and the reported label blamed the project's config for a defect in the
plugin.

### 7. The three retired test pins, named

The design predicted three; execution retired exactly those three, one task earlier than planned
(the blackout emit and its assertions had to move together or the task would have ended red).

| Retired | Why it is safe |
|---|---|
| `TestRepairRemedyCoversEveryMismatchKind` → *"a renamed slug reports BOTH categories, not the first"* | It asserted the **text** of a remedy that no longer exists. The replacement subtest, *"a rename is both kinds at once and both are reconciled"*, asserts the strictly stronger property: both kinds are **resolved**, not merely explained. |
| `TestRepairRemedyCoversEveryMismatchKind` → *"a duplicated slug is reported AND its repair is explained"* | Same shape, same replacement: *"a duplicate keeps the first occurrence and quarantines the extra."* Explaining a repair is weaker than performing one. The unrelated `governed = false` subtest in that test stays. |
| `TestDocumentedPostureRecipeActuallyGoverns` → *"hand-written partial file governs nothing"* | Its premise retired: a partial file **is** a `missing:` mismatch, and now reconciles. Retiring it left a real hole — the redirect pointed at a one-missing-row test, weaker than the all-sixteen case `decision-0072` actually described — so a positive replacement pins the recipe end to end: a bare `strictness = "firm"` file reconciles to all sixteen rows with the posture preserved. |

`TestEveryDestructiveInstructionIsGated`'s floor moved 13 → 12 for the same reason: the retired
blackout message was one of the gated messages, so removing it removed **something to gate, not a
gate**. The emit count itself is unchanged at nineteen — the blackout emit went and the
`no-slugs-in-payload` branch's loud emit took its place — and the guards report **zero ungated
destructive messages** across both channels. Counted, not argued.

## Consequences

- **`plugins/trellis/VERSION` 0.6.0 → 0.7.0**, with both plugin manifests and `install.sh`'s baked
  bundle manifest advanced in the same commit (`decision-0028`). A payload change that does not
  move `VERSION` never reaches an installed copy — the catalog's own row-set obligation note
  records that this obligation is still **unguarded** (`trellis#245`).
- **`decision-0070` D4 is untouched and still pinned** — `cli/plugin_hook_test.go:1704` fails with
  *"the hook wrote `.trellis/rules.toml` — 'the hook never writes' is the half of `decision-0070` D4
  that stands"*.
- **A commented row is new user-visible state in a consumer-owned file.** Both READMEs now say what
  it is: inert, safe to leave, one uncomment away if a newer release ships that slug.
- **The row-set obligation list is one surface shorter.** Its note in
  `core/catalog/signature-catalog-v1.md` (copied verbatim to `cli/assets/invariants.md`) named
  *"`plugins/trellis/hooks/codex-context.mjs`'s hardcoded `SLUGS`"* among the surfaces a row-set
  change touches. That hardcode is gone, so the entry is marked retired **in this same change**
  rather than left to send a future author hunting for a constant that no longer exists
  (`decision-0028`, `inv-deliberate-succession`). The list itself — ~15 surfaces enforced only by
  prose, the root cause upstream of all three issues — is **TRL-28**, which is the consumer that
  will re-present it (`decision-0078`).
- **`/trellis:remove` needs no change.** It deletes `.trellis/` wholesale, quarantined rows
  included.
- **Cross-host regression coverage is new.** `TestReconciledRowsParseForCodexToo` runs Claude's hook
  to produce the real reconciled text, writes it to `.trellis/rules.toml`, and runs Codex's actual
  `codex-context.mjs` against the identical file. It exists because the first reconciler appended
  missing rows at EOF with no awareness of whether a `[rules]` table had been opened — on the
  hand-written-partial shape that put sixteen rows at top level, which Claude governs from normally
  and Codex reads as `invalid-rules`. **It carries a second case, from the final whole-branch
  review, for the same class through the opposite mistake:** the table detection was anchored at
  column 0 while `parseRulesToml` trims each line first, so an **indented** `[rules]` was the table
  for Codex and not for Claude, and the repair appended a second one — a duplicate section, which
  is exactly what that parser rejects. Not a regression (such a file was already Codex-invalid) but
  the mandate promises the written file *matches what governs*, and a file Codex refuses does not.
  The detection now matches Codex's own leniency, as the row match beside it already did.
- **The repair summary counts this session, not every session before it.** The counts were read
  back out of the reconciled text with `grep`; quarantine notes and the `# added N row(s)` header
  are **persisted** provenance, so a partially repaired file made the summary cumulative — measured
  at *"added 2 row(s); quarantined 1 row(s)"* for a session that added 1 and quarantined 0. The
  in-file provenance was right either way, which is what made it easy to miss: only the **spoken**
  summary inflated, and that summary is the whole loudness channel this record rests on. The
  reconciler now states its own counts on a trailer line the shell strips before delivery.

## What execution found that the design did not

### The Codex byte cap was invented, and the refusal it caused was self-inflicted

`codex-context.mjs` refused to inject anything over `MAX_CONTEXT_BYTES = 8000`, emitting
`context-over-budget`. Reconciliation's per-row provenance comments crossed it, so a reconciled
Codex session would have injected **nothing** — reintroducing on Codex exactly the blackout this
change removes on Claude. The obvious reading was that 8000 is an external Codex limit and the
provenance must be compressed to fit under it.

**It is not an external limit. It has no provenance at all.** The string `8000` appears nowhere in
`decisions/`, `research/` or `core/`; the constant entered in commit `3490555` with no rationale
recorded anywhere. Codex's own documented behaviour, primary source
(<https://learn.chatgpt.com/docs/hooks>):

> By default, Codex limits each model-visible hook-output message to roughly **2,500 tokens** …
> [over the threshold it] saves the full text under `<temp_dir>/hook_outputs/<session_id>/<uuid>.txt`
> and gives the model a head-and-tail preview with the saved-file path.

Configurable per handler via `additionalContextLimit`; the installed `codex-cli` 0.149.0 binary
contains the string `ignoring additionalContextLimit for`, corroborating that the setting exists.

So: **Codex measures tokens; Trellis measured bytes.** At ~4 bytes/token, 8000 B is ~2000 tokens,
comfortably *under* Codex's own default — and Codex does not reject over its limit, it **spills
gracefully** and tells the model where the full text went. Trellis's `context-over-budget` refusal
was therefore **strictly worse than doing nothing**: a self-inflicted blackout guarding a limit
nobody imposed, replacing graceful degradation with total loss.

The cap is now **9500** (~2375 tokens — still under Codex's default, so this hook never triggers
Codex's spill path either), with the units mismatch and the provenance documented on the constant
itself. The provenance compression found along the way was kept — one header comment above the
appended block instead of one per row is better regardless — but it is no longer the thing holding
the budget together.

**The generalisable finding is not the number.** It is that an undocumented magic constant had been
enforcing a refusal for so long that a design brief, a controller ruling and an implementer all
treated it as an external constraint to design around. The only thing that broke the frame was
checking whether the constant had a source. It did not.

### Five tests passed for the wrong reason, and only mutation testing found them

This is recorded as a finding about the **method**, not as a list of fixed bugs. Across this branch,
five separate assertions were green while proving nothing:

- **Two** asserted hook message strings that the same commit's parent had already **deleted**
  (`"does not match the rules the installed plugin ships"`, `"Nothing was injected"`). A
  `Contains` check against a string that no longer exists can only ever pass.
- **One** slug check scraped the hook's **whole output**, including the rules prose, rather than the
  delivered TOML rows — so it stayed true even with every row quarantined. This is precisely why
  the branch's one Critical defect (`no-slugs-in-payload` entering the reconciler with an empty want
  set, quarantining every legitimate row and running ungoverned at exit 0) got past its own test.
- **One** guard matched the verb `delete` **case-sensitively**, so a capitalised *"Delete the
  unknown rows…"* walked through both destructive-instruction guards untouched.
- **One** subtest written to prove a guard was *"actually enforced"* recomputed the scanned message
  set **independently** — it exercised the guard's helpers rather than the guard's own wiring.
  Proven by mutation: replacing `msgs = append(msgs, payloadMsgs...)` with `_ = payloadMsgs` left
  both the outer test and the new subtest green, with only a log line falling from 28 messages
  to 19.

**Every one was found by breaking the code and confirming the test failed** — never by reading the
test, and never by a green run. Three were found by the reviewer after the implementer reported the
work done and green.

**This is the same shape `decision-0072` recorded**, where seven review rounds found one defect
class eleven times, and where finding #7 — *"a guard that recognises one verb is a guard against one
verb"* — was called the most useful of them because it was about the guard rather than the code.
The recurrence says the class is structural, not incidental: **a test written alongside the code it
covers inherits that code's blind spots**, and the only cheap instrument that reliably exposes it is
mutation. Every fix on this branch was accepted only after its author reverted the fix and
reproduced the exact symptom. That practice — not any single repair — is what this section is
recording, and it is product research: it is a candidate obligation for Trellis's own review
contract, not merely a habit that worked once here.

## Open questions

- **Should the byte proxy be replaced by a token estimate outright?** 9500 B is a calibrated guess
  at Codex's ~2500-token limit. If the ratio for this content is worse than 4:1, Codex spills —
  which loses nothing and names the file. The safe direction is known; whether to keep measuring
  the wrong unit at all is not answered here.
- **Should the Claude hook warn on a false floor row, as Codex does?** Codex warns when a `floor-`
  row is `active = false`; Claude is silent. Named out of scope by the design and still open.
- **Host parity is owed, and the Codex arm of TRL-20 is therefore still open.** The design
  claimed *"one table, applied identically by both hosts"* and that Codex's `parseRulesToml`
  *"adopts the reconcile semantics instead"*; §1 records that neither shipped. A Codex project
  with a mismatched row set still gets `invalid-rules` and no rules — the exact blackout this
  record retires on Claude — so the defect TRL-20 names is half-closed, not closed. Deferred
  deliberately rather than half-implemented under a design that had not been re-thought for the
  host: Codex refuses in a **parser**, not a delivery branch, so adopting reconciliation there is
  a restructure, not an anchor change. **Tracked in Linear (Trellis team)**
  (`inv-no-orphan-followups` — the named consumer that will re-present it, `decision-0078`).
- **A file with no `strictness` key stays Codex-invalid after a repair, so the reconciler can
  produce a file it has just promised is correct.** `parseRulesToml` requires `strictness` to be
  exactly `"firm"` or `"adaptive"` and returns `null` otherwise; the reconciler adds rows and
  quarantines rows, and never adds a `strictness` line. The shape is reachable from this record's
  own §5 measurement: an empty `.trellis/rules.toml` reconciles to sixteen rows under a `[rules]`
  table with no `strictness`, which Claude governs from at the adaptive posture and Codex rejects
  outright. **Not changed here** — the repair is a strict improvement over the blackout either way
  and adding a posture the consumer never wrote is a different act from adding a row the payload
  ships, which is a call worth making deliberately. Named beside the parity question because the
  same work closes both, and because a repaired file that one host still refuses is exactly what
  the mandate's *"so the file matches what governs"* promises it is not.
- **Two live surfaces still teach the retired behaviour, and neither is fixed here.** The inline
  managed block's frozen row copy (`cli/apply.go`, `README.md`) is the derivative that genuinely
  goes stale, and reconciliation does not reach it. Beside it,
  `plugins/trellis/reference/block-codex.md:10,22` still states the all-or-nothing activation
  predicate — *"every canonical slug below occurs exactly once, no unknown or duplicate slug
  occurs"* — and instructs the agent to say exactly *"Trellis was not loaded"* when it fails. That
  is the vendored-overlay fallback path, which `plugins/trellis/README.md:83-85` records as
  **carried rather than maintained**, so it is deliberately left alone rather than half-updated.
  Named here so the record does not claim one stale surface when there are two. Neither is covered
  by any of the three issues; both owed.
- **Do the hook's overlay branches still earn their keep?** Inherited unanswered from
  `decision-0072`'s open questions; this change does not touch them.

## Self-check (gate)

**The supersession is partial and marked, not implied.** `decision-0072` keeps every clause this
change does not reach, and its prose is unedited — the forward pointer is in frontmatter
(`inv-auditable-archive`, `decision-0082`).

**The invariant this most plausibly violates is confronted rather than left for a reviewer.**
Ungating a write to a consumer-owned file is the shape `floor-intent-gate` exists for, and
`decision-0072` finding #6 is the exact precedent against it. The argument that it does not apply
rests entirely on quarantine being non-destructive — stated as a load-bearing premise above, not
as a passing remark — and it is enforced by two widened guards rather than by intent.

**The boundary with what came before is drawn in both directions** (`decision-0074`): what
`decision-0072` loses is enumerated in §4, what it keeps is enumerated beside it, and the one
obligation this change retires from the catalog's row-set list is migrated in the same commit
rather than left to be discovered.

**Where the plan and execution disagree, execution is recorded.** The byte cap was to be designed
around; it was investigated instead, and the ruling that told this task to compress was withdrawn
by the maintainer's own instruction to investigate rather than choose. Three test pins retired, not
the two the plan named. Both are stated as changes, not smoothed into the plan's shape.

**Author is an agent.** The maintainer directed the work and ruled on the byte cap; he has not
reviewed this record. **His merge is the acceptance** (`decision-0082`) — there is nothing here for
him to flip, and no agent may merge on his behalf (`floor-intent-gate`). The judgment calls made
without him — retiring the third pin, correcting the catalog's obligation list here rather than in
TRL-28, keeping the provenance compression after the cap moved — are named above and each is
reversible.

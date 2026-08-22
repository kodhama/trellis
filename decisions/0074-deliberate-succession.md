---
id: decision-0074
type: decision
status: approved  # maintainer's intent act 2026-08-22, in-conversation ("I already decided it's a 15th one"), taken through a separate session first — this flip records it (decision-0046, decision-0022). Author (agent) != approver (maintainer). Scope: the mint. The contract-layer corrections below landed after an independent code + corpus review and are the maintainer's to accept at merge
depends_on: [invariants-v1, signature-catalog-v1, decision-0052]
informed_by: [decision-0018, decision-0021, decision-0027, decision-0028, decision-0040]
owner: agent
date: 2026-08-19
---

> **Provenance.** Shaped in-session with the maintainer (2026-08-19), from a live incident
> in `kodhama/math-quest`: four recurrences in one session of the *backward* direction of
> `decision-0052`'s entropy lean, with the forward rendering loaded in context throughout.
> The maintainer's framing — *"being explicit about shape inheritance has merit on its own"*
> and *"that will make you author specific examples about what to do with the previous
> version of the knowledge or artifact base"* — is what this decision formalizes. The
> evidence instance is recorded on trellis#166; **this decision is the content slice, and
> does not schedule that experiment.**

# 0074 — `inv-deliberate-succession`: the boundary with what came before is decided, not drifted through

> **Decided: mint it.** The maintainer settled this on 2026-08-22, having taken it through a
> separate session first: *"I already decided it's a 15th one."*
>
> **Recorded against it, because the objection is real and should not vanish.**
> `decision-0021`'s minting test is *mechanism* — *"keep it when it introduces a mechanism the
> others don't"* — and by this record's own Context, `inv-self-improvement`'s dispositional face
> already covers succession generically. This buys **worked examples**, not a new mechanism, and
> `inv-minimal-first` argues against the catalog growing. The counter that carried the day is the
> iron rule: a rule you can't exemplify is vaporware, and a disposition covering succession
> generically cannot host *"what do you do with the previous version of this knowledge base?"*
> without becoming a rule about two things. Two instances back it, one of them four recurrences in
> a single session **with the existing forward rendering loaded throughout**.
>
> **Where the objection goes next.** Compression is a set-wide question, not this record's:
> **trellis#239** asks whether the set carries an unnamed dimension that would let several
> entries collapse, and takes this pair as its starting corpus. That audit — not a veto here —
> is where "should this be fifteen?" gets answered properly, and `inv-prune-bias` cuts both
> ways when it runs.

## Context

- **The incident (backward direction).** A new mastery model in `math-quest` needed a value
  for the lower of two thresholds. The constant from the model it replaces — a 4-in-a-row
  streak — filled the gap, described in the architecture as *"the existing 4-in-a-row with
  strict reset"*. It had been calibrated for massed practice inside a topic lock **that same
  model retires**, and it is a *streak* where the other threshold is *coverage*, silently
  making "two thresholds on one fold" two different measures. The maintainer flagged it four
  times across the session, the last explicitly as a pattern: *"I keep having to say we owe
  no obligation to what is implemented."*

- **Why `decision-0052`'s rendering did not catch it.** 0052 extended
  `inv-self-improvement`'s directive with the *forward* case: *"when you introduce a new
  pattern … the existing stock now sitting outside it is a signal to surface."* Every clause
  presupposes a new pattern exists and old stock is now non-conforming. The backward case has
  no clause: **nothing changed and nothing conflicted — a gap was filled.** It arrives looking
  like continuity, so no directive text applies and no salient event exists to attach to.

- **Two homings tried and withdrawn in-session, recorded so neither is re-litigated.**
  1. *A 15th slug justified on "the lean does not fit `inv-self-improvement`'s `why`"* —
     withdrawn. That argument read the **catalog** (annotation) alone; `invariants-v1:184-195`
     states *"One principle, two faces"* outright, the dispositional face being direction-
     agnostic. This is the same annotation-vs-annotated error 0052's own adversarial pass
     killed in its third homing.
  2. *Extracting the lean from `inv-self-improvement`* — withdrawn on inspection: **the lean is
     not in `invariants-v1` at all.** 0052 states *"no `invariants-v1` amendment is needed or
     made"*, because the dispositional face covers it generically. There is nothing at the set
     layer to extract; a new slug is a set **amendment**, not a re-homing.

- **The surviving justification is the iron rule, not a homing error.** `AGENTS.md`: *"every
  invariant or abstract instruction carries ≥1 concrete example (few-shot) — a rule you can't
  exemplify is probably vaporware."* The dispositional face **covers** succession generically
  and therefore cannot carry worked examples of it: a rule about friction cannot host *"what
  do you do with the previous version of this knowledge base?"* without becoming a rule about
  two things. A slug can, and this one must — it ships with four honored and four violated
  pairs spanning structure, design, docs and metrics.

- **Why a set amendment is the honest layer.** Minting a catalog slug with no set entry forks
  annotation from annotated — `inv-graph-maintenance`'s own violated case, and the failure 0052
  steered around. Identity-level changes go through the set (`decision-0018` restored a slug;
  `decision-0021` collapsed one). This follows that path, and says so rather than smuggling a
  slug in at the annotation layer.

## Decision

**1. `invariants-v1` gains `inv-deliberate-succession`** (operating set, `trellis-design`,
*provisional*), stating one relationship in two directions — forward (what the new leaves
outside it) and backward (what the old still supplies to a gap) — with the prior version
framed as *evidence to weigh, never gravity to drift with, never debris to step over*.
Surfacing rides SI-1's channel discipline and SI-2's existing rituals; it declares
`inv-self-improvement` its neighbor.

**2. `inv-self-improvement` reverts to its reactive face** — in the catalog, which is the only
layer that carried the lean (the Context above records why: 0052 made no `invariants-v1`
amendment, so there is nothing at the set layer to revert). Its directive drops
the appended entropy-lean sentence (`decision-0052` point 1); its signature clause
(0052 point 2) and its `(structure)` honored/violated pair (0052 point 3) move to the new
entry, with a pointer left in place. **`decision-0052` is superseded in part, and carries the
forward marking this contract requires**: `spec-0001` §2 and `decision-0040` point 5 class
`superseded_in_part_by` as *"a marking, not an edit-in-substance"* and require it on the
predecessor, so *not* marking it would leave a reader on the outgrown half with no forward
link — the failure this record's own new signature names. An earlier draft of this decision
claimed the opposite (that leaving 0052 untouched was the append-only-correct move); that was
a misreading, caught at the independent review, and the mark lands in this change per
`decision-0040`'s own worked instance (`decision-0013` → `decision-0038`). Its substance is
untouched; SI-1 continues to govern how any such signal reaches the channel.

**3. The catalog gains the matching entry** with the two-direction directive and **four
honored / four violated pairs** — the iron-rule obligation this decision rests on. Coverage
arithmetic moves **14 → 15 assessable slugs**.

**4. The wording is our synthesis** (naming guardrail): no external provenance implied. The
evidence is two in-house instances — trellis PR #165 (forward) and math-quest's phase-1
architecture (backward) — both named above.

## Consequences

- **The derived chain regenerates in the same change** (`decision-0028`), and it has two layers:
  the **render** chain — catalog → the plugin's `reference/` render → `rules.md` → both
  inline-block sandwiches → `checksums` → `version` stamp → `install.sh` bundle manifest → the
  invariant scorecard, plus the preset rows in `rules-a.toml` / `rules-b.toml` (without a row the
  rule ships but is inactive) — and the **contract** chain: `spec-0007`'s canonical slug inventory
  and its activation-row predicate, `spec-0002` §1 check 2 + AC1, `core/rubrics/artifact-contract.md`,
  and the `corpus-reviewer` checklist. Miss the second and the gate instructs its reader to expect
  fourteen. `spec-0007` takes a `version` bump: its inventory is a testable clause.
- **An existing curl install upgrades into a false all-clear** (trellis#241, found by the review
  on this record's PR). `install.sh` never overwrites an existing `.trellis/rules.toml`, so a
  re-run pairs a freshly rendered 15-rule file with a preserved 14-row config; the staleness
  hook's curl-path branch checks markers, stamp presence and stamp freshness — all of which pass —
  and then reports that the config governs. The rule is delivered, inactive, and *claimed as
  governing*. The mechanism predates this change (that branch never validated rows); this is the
  first slug added since the curl path existed, which is what makes it bite. Disclosed here rather
  than fixed in this record's change: the fix is a hook behaviour change with its own tests.
- **`decision-0052` gains `superseded_in_part_by: [decision-0074]`.** Required by `spec-0001` §2
  and `decision-0040` point 5, which class the marking as *"a marking, not an edit-in-substance"*.
  Its substance is untouched; SI-1 still governs how any such signal reaches the channel.
- **This retires the live worked instance of `decision-0027`'s amendment.** That amendment permits
  the first `violated` bullet to carry an appended clause from another pair's use case, and cites
  exactly one example: `inv-self-improvement`'s *(CI)* bullet carrying the *(structure)* clause.
  Point 2 deletes that clause, so the amendment keeps its rule and loses its example. It is not
  edited — recorded here so the dangling example is known, not discovered.
- **Four pairs exceeds `decision-0027` point 3, which says two.** The enforced gates say `≥2`, and
  the catalog already ships three- and four-pair entries, so four is conformant and precedented.
  Named because point 3's rationale — *"uniform, scannable cards"* — is a readout-budget concern
  this entry doubles.
- **Count change, not a pair-door change.** `decision-0027`'s pair door and `decision-0040`'s
  directive-extension door both preserve the count; this does not. It is an amendment, flagged as one.
- **The readout grows by one rule** — `decision-0052`'s open question applies with more force here.
  The counter-argument: the content was already in the readout as an unexemplified disposition, and
  this buys examples for the same budget plus one heading.
- **`inv-clarify-before-commit` and `inv-graph-maintenance` stay untouched** — 0052's Context records
  why both were rejected as homes.
- **trellis#166 is unaffected and unscheduled.** The backward direction is plausibly the harder moment
  to trap — forward has a change to attach to, backward has none — which is evidence *for* the
  trigger-format hypothesis, recorded there.
- **`profiles/trellis-self.md`** gains a row, at `confidence: inferred` — see its own note for why
  `verified` would overclaim.

## Open questions

- **Does it pull its weight?** Two instances justify the slug; if the backward case does not
  recur in projects that carry the rendered rule, the honest move is to collapse it back into
  `inv-self-improvement`'s disposition — `inv-prune-bias` cuts both ways.
- **Is "succession" the right frame, or is this really about *versions*?** The examples span
  conventions, constants, docs and schemas; if a later instance is about none of those, the
  `what` may be drawn too narrowly.
- **Is "succession" load-bearing across instances, or is this a math-quest shape?** Both
  recorded instances are kodhama-family. A third from an unrelated project would settle it.

## Self-check (gate)

Homed at the layer the change touches: a new slug is a set amendment (`decision-0018`/`0021`
path), not a catalog-only door. Two rejected homings recorded with their kill-reasons, one of
them the author's own withdrawn argument (`inv-auditable-archive`). `decision-0052` carries its
`superseded_in_part_by` mark, substance untouched. The count change and the `decision-0027`
point-3 excess are both flagged, not smuggled (`floor-transparency`). The iron-rule obligation is
discharged in-artifact: four honored / four violated pairs, tag-aligned per `decision-0027`
point 1.

**Not self-caught.** Four defects came from the independent review, not the author: the missing
`superseded_in_part_by` mark; the contract-layer count desync; `spec-0007`'s inventory left at
fourteen while its prose was made count-free; and a guard that asserted the entropy lean had
*moved* while searching the whole catalog, so it passed with the move un-made. Three of the four
are the same root cause — a sweep that matched only some of the shapes a count takes. Recorded
because a rule about noticing succession, whose own introduction missed four succession
obligations, has earned the `provisional` tag it carries and the `inferred` confidence in the
profile.

Flipped to `approved` on the maintainer's intent act, not the author's judgement
(`decision-0046`); the record of that act is in the frontmatter.

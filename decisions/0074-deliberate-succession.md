---
id: decision-0074
type: decision
status: draft  # the author does not grade its own decision (decision-0046); the gate and any flip are the maintainer's intent act
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
  pairs spanning structure, design, docs, data and metrics.

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

**2. `inv-self-improvement` reverts to its reactive face** at both layers. Its directive drops
the appended entropy-lean sentence (`decision-0052` point 1); its signature clause
(0052 point 2) and its `(structure)` honored/violated pair (0052 point 3) move to the new
entry, with a pointer left in place. **`decision-0052` is not edited** — it is superseded in
part by this record, per the append-only rule; SI-1 continues to govern how any such signal
reaches the channel.

**3. The catalog gains the matching entry** with the two-direction directive and **four
honored / four violated pairs** — the iron-rule obligation this decision rests on. Coverage
arithmetic moves **14 → 15 assessable slugs**.

**4. The wording is our synthesis** (naming guardrail): no external provenance implied. The
evidence is two in-house instances — trellis PR #165 (forward) and math-quest's phase-1
architecture (backward) — both named above.

## Consequences

- **The derived chain regenerates in the same change** (`decision-0028`): catalog → the
  plugin's `reference/` render → `rules.md` assembly → both inline-block sandwiches →
  `checksums` → `version` stamp → `install.sh` bundle manifest (Go-test-guarded) → the
  invariant scorecard. A new slug also needs its **row in the preset `rules-a.toml` /
  `rules-b.toml`**, and in consuming repos' `.trellis/rules.toml` on their next refresh —
  without it the rule is delivered but inactive.
- **Count change, not a pair-door change.** `decision-0027`'s pair door and `decision-0040`'s
  directive-extension door both preserve the count; this does not. It is an amendment and is
  flagged as one.
- **The readout grows by one rule.** `decision-0052`'s open question — *"does the extended
  wording pull its weight in the readout?"* — applies with more force here, and
  `inv-minimal-first` argues against its own catalog growing. The counter-argument is the iron
  rule: the content was already in the readout as an unexemplified disposition, and this
  buys examples for the same budget plus one heading.
- **`inv-clarify-before-commit` and `inv-graph-maintenance` stay untouched** — 0052's Context
  records why both were rejected as homes, and nothing here disturbs that.
- **trellis#166 is unaffected and unscheduled.** The backward direction is plausibly the harder
  moment to trap — forward has a change to attach to, backward has none — which is evidence
  *for* the trigger-format hypothesis, recorded there. Whether this rule needs a `fires when:`
  line is that decision's to make, not this one's.
- **`profiles/trellis-self.md`** gains a row for the new slug; the self-application overlay and
  managed block regenerate with the payload.

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

Homed at the layer the change actually touches: a new slug is a set amendment
(`decision-0018`/`0021` path), not a catalog-only door, and the decision says so instead of
minting at the annotation layer. Two rejected homings recorded with their kill-reasons, one of
them my own withdrawn argument and why it was wrong (`inv-auditable-archive`). `decision-0052`
is superseded in part, never edited (append-only). The count change is flagged, not smuggled
(`floor-transparency`). The iron-rule obligation is discharged in-artifact — four honored and
four violated pairs, spanning five domains. Grove-routing and readout-budget hazards surfaced
in-artifact rather than discovered later. Left at `draft`: the author does not grade its own
decision, and the `approved` flip is the maintainer's intent act (`decision-0046`), ideally
after an independent pass (`inv-independent-judgment`).

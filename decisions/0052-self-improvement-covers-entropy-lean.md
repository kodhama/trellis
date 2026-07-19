---
id: decision-0052
type: decision
status: draft
depends_on: [invariants-v1, signature-catalog-v1, decision-0040]
informed_by: [decision-0018, decision-0028, decision-0051]
owner: agent
date: 2026-07-19
---

> **Provenance.** Shaped in-session with the maintainer (2026-07-19), from a live
> incident during PR #165's review. Three homings were tried; two died under review —
> the trail is recorded in Context, because the corpus's own history turns out to
> explain why both wrong homes felt close. The maintainer's framing — "a lean towards
> less entropy… a pull towards consistency… not too different from self-improvement…
> it's not a graph" — is what the final homing formalizes. The broader trigger-format
> question is parked separately (trellis#166).

# 0052 — `inv-self-improvement` covers the entropy lean: a pattern-introducing change surfaces its retrofit signal

## Context

- **The incident (the evidence instance).** PR #165 introduced the `eval/experiments/`
  convention (one self-contained directory per experiment) while the original framework
  harness stayed loose at the `eval/` root. The change threw off a signal — old stock
  now sat outside the new pattern — and the agent **silently dropped it**, resolving
  the open migrate-or-exempt question in confident boundary prose ("those belong to the
  framework A/B suite…"). The maintainer had to flag it: "we'll have two conventions
  inside the same folder, which is kind of confusing."
- **The homing trail (two rejected, recorded so neither is re-litigated).**
  1. *`inv-graph-maintenance`, via a graph metaphor* ("orphaned stock is blast
     radius") — killed by the maintainer: *folder structure is not graph maintenance;
     the dependencies could be identical while the files live somewhere else.*
  2. *`inv-clarify-before-commit`* ("a structural choice is the same ambiguity") —
     killed as tail-not-head: clarify covers *don't guess at a noticed fork*; the
     missing behavior was the disposition that makes the fork salient at all. The set
     itself already wires the tail to clarify from inside self-improvement (SI-1 cites
     it by name, below), so nothing is lost by withdrawing this homing.
  3. *`inv-graph-maintenance`, via a coherence-identity re-scope* (widen its `what`) —
     killed **UNSOUND** by an adversarial pass: the reading cherry-picked the catalog
     (the annotation layer) while `invariants-v1` — the annotated layer — states the
     graph identity four independent times ("directional-graph maintenance"; "The
     **dependency graph** … is kept consistent and minimal"; "the four operations of
     keeping the graph true"; "graph-maintenance keeps the graph *true* (referential
     integrity)"). A catalog-only re-scope would fork annotation from annotated — this
     invariant's own violated case — and no decision has ever edited a `what`
     (identity-level changes went through the set: `decision-0018`, `decision-0021`).
- **The right home, from the set's own text.** `invariants-v1`'s
  `inv-self-improvement` has *two faces*, and only one is currently rendered in the
  shipped rule. Its **checkable floor SI-1** is the incident nearly verbatim: "the
  improvement signals **a change or a work session throws off** are **surfaced through
  the project's chosen channel** — asked (`inv-clarify-before-commit`) or inferred
  (Assess), **never assumed** — never silently dropped." Its **dispositional face**:
  "the agent **proactively notices** a signal and proposes *both* the fix *and* a
  standing trigger… **asking not acting**." The shipped directive ("When something
  breaks or causes friction, fix the root cause…") renders only the *reactive* face —
  the proactive entropy-lean the set already owns was never expressed in the readout.
  That is the actual gap this decision closes. SI-2 (`inv-ride-existing-rituals`) even
  supplies the delivery: the retrofit question rides the change that creates it — the
  PR you are already writing — no new ceremony.
- **Why both wrong homes felt close.** The set's history already litigated this
  boundary: `inv-self-improvement` was once merged *into* `inv-graph-maintenance`, and
  `decision-0018` restored it because the merge "kept the prune mechanic but lost the
  *evolve* pillar." They remain declared neighbors sharing the prune-bias hinge
  (SI-3). The maintainer's dual intuition — graph-adjacent, improvement-adjacent, and
  neither exactly — recapitulates that split.
- **Why the fix is a payload change, not a repo rule.** A CLAUDE.md line fixes
  trellis; the overlay's rule fragment fixes every project the payload reaches — the
  product's entire premise.
- **What this is not.** The trigger-format question (rules don't encode their firing
  moments) stays parked at trellis#166, sequenced after the annotation-vs-absence
  experiment. This decision is the content slice only.

## Decision

**1. The catalog directive extends** (`core/catalog/signature-catalog-v1.md`,
`inv-self-improvement` — the single home; everything below derives from it,
`decision-0028`), rendering the dispositional face the set already owns. Appended to
the existing directive text:

> And notice the friction you are about to create: when you introduce a new pattern —
> a convention, a naming scheme, a format — the existing stock now sitting outside it
> is a signal to surface, riding the same change: migrate it, or name the exemption
> and ask — never resolve it silently in prose.

This is a directive change to the always-loaded block, **beyond the pair door's
letter** — flagged exactly as `decision-0040` point 4 flagged its one directive
extension ("folds into the entry (directive extended, signature clause, a pair) rather
than minting a slug"). It expresses set-layer content (`invariants-v1`'s dispositional
face + SI-1) that the catalog never rendered; the annotation follows the annotated —
**no `invariants-v1` amendment is needed or made, and no `what` is edited.**

**2. The signature gains the clause**: *the entropy lean as proactive notice — a
pattern-introducing change throws off a retrofit-scope signal (what now sits outside
the pattern?); migrate-or-exempt is surfaced through the channel, asked not assumed,
never resolved in boundary prose (SI-1 applied at change-introduction time).*

**3. One honored/violated pair is added via the signature-pair door** (originated by
`decision-0027`; `decision-0040` is the worked precedent — pairs on existing
invariants; **the count stays 14**):

- honored: *(structure)* a change that introduces a pattern ships with its retrofit
  question — "the old suite now sits outside this convention: migrate or exempt?" —
  and the human rules on it.
- violated: *(structure)* a new directory convention lands while the old stock stays
  loose beside it, exempted by confident boundary prose nobody approved — two
  conventions in one tree.

**4. The rendered rule's ✗ line extends in the same breath** (still one ✗ bullet, the
readout format unchanged): "…and everyone just re-runs it, forever — or a new
convention lands and the old stock stays loose beside it, exempted by prose nobody
approved."

**5. The wording is our synthesis** (naming guardrail): no external provenance
implied; the evidence is one in-house instance plus the set's own pre-existing
dispositional-face text, both named above.

## Consequences

- **The derived chain regenerates in the same change** (`decision-0028`; the
  `decision-0051` machinery makes this cheap): catalog edit → generator render →
  `plugins/trellis/reference/rules/inv-self-improvement.md` fragment → `rules.md`
  assembly → both inline-block sandwiches → `checksums` → `version` stamp →
  `install.sh` bundle manifest (Go-test-guarded) → the invariant scorecard
  (`gen-invariant-scorecard.py` renders directive + violated lines; its sync-guard CI
  fires) → this repo's own `.trellis/internal/` overlay + managed block. Family repos
  and consumers pick the extended rule up on their next refresh.
- **Sequencing hazard (surfaced, not discovered later):** the `annotation-vs-absence`
  experiment (`research-0012`) has not yet run, and this rule's directive rides in
  **all three arms'** readouts — a mid-run change would confound every arm. **This
  decision's build lands before the experiment's first batch or after the run
  completes — never between batches**; `runs/provenance`'s payload stamp records which
  text was live either way.
- **`inv-graph-maintenance` and `inv-clarify-before-commit` are deliberately
  untouched** — the Context trail records why, with the adversarial verdict, so
  neither homing is re-litigated.
- **`profiles/trellis-self.md` row for `inv-self-improvement`** (line 52, `verified`):
  named for a re-glance at build time — the dispositional face was already set-layer
  content, so the row's evidence should still cover its claim, but the re-reviewer
  confirms that rather than this decision assuming it.
- **Catalog flags unchanged**: `mechanizable: false` already names "the
  proactive-notice disposition" as the uncheckable part — this decision renders that
  disposition into the readout without changing its checkability claim. Coverage
  arithmetic unchanged (14 assessable slugs).
- **trellis#166 unaffected** — if fires-when lines land later, this rule's moment
  ("you are introducing a pattern") is written by that decision, not this one.

## Open questions

- **Does the extended wording pull its weight in the readout?** It lengthens one rule
  bullet in every session's context. If annotation-vs-absence and #166 show rule-text
  form dominates content, revisit length there.
- **Second evidence instance.** One incident plus the set's pre-existing face justify
  the door; if the silent-drop failure recurs *despite* the rendered disposition, that
  is #166's trigger-format hypothesis gaining evidence — note it there.

## Self-check (gate)

Homed by the set's own text, not a metaphor: the directive extension renders
`invariants-v1` content (SI-1 + the dispositional face) the catalog never expressed —
annotation follows annotated, no layer forked, no `what` edited, no set amendment
(the adversarial pass's F4 boundary respected). Uses the signature-pair door plus the
`decision-0040`-precedented directive-extension exception, flagged in the decision
body, not smuggled (`floor-transparency`). Count stays 14; nothing minted; two
rejected homings recorded with their kill-reasons (`inv-auditable-archive`'s
why-is-it-this-way). The experiment-sequencing hazard is surfaced in-artifact.
`depends_on` carries genuine coupling (the set whose face is rendered, the catalog
being amended, the door's precedent); `decision-0018`/`0028`/`0051` are context
(`informed_by`, `decision-0047`). Left at `draft`: the author does not grade its own
decision — the gate and any `approved` flip are the maintainer's intent act
(`decision-0046`), ideally after an independent pass (`inv-independent-judgment`).

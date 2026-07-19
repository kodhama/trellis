---
id: decision-0052
type: decision
status: draft
depends_on: [invariants-v1, signature-catalog-v1, decision-0040]
informed_by: [decision-0028, decision-0051]
owner: agent
date: 2026-07-19
---

> **Provenance.** Shaped in-session with the maintainer (2026-07-19), from a live
> incident during PR #165's review: the maintainer had to flag a two-conventions
> inconsistency the agent had *rationalized* rather than surfaced. The maintainer's
> direction: the behavior belongs in the shipped invariant set — "the point was really
> how do we take this behavior to other projects... that's the whole idea of Trellis" —
> with the broader trigger-format question parked separately (trellis#166, sequenced
> after the annotation-vs-absence experiment). A first draft homed this on
> `inv-graph-maintenance`; the maintainer's challenge re-homed it (see Context).

# 0052 — `inv-clarify-before-commit` covers structural choices: a new pattern's retrofit scope is a question, not a default

## Context

- **The incident (the evidence instance).** PR #165 introduced the `eval/experiments/`
  convention (one self-contained directory per experiment) while the original framework
  harness stayed loose at the `eval/` root. The ask ("structure the experiment machinery
  in a way that leaves room for future experiments") left the old harness's fate
  genuinely open; the agent *silently picked one reading* — exempt — and documented it
  as design ("those belong to the framework A/B suite…"). The maintainer had to flag
  it: "we'll have two conventions inside the same folder, which is kind of confusing."
  The failure mode was not blindness but **rationalization** — boundary prose that is a
  question wearing an answer's costume.
- **The right home — and the rejected one.** A first draft widened
  `inv-graph-maintenance` ("orphaned stock is blast radius"). The maintainer's
  challenge killed it: *folder structure is not graph maintenance — the dependencies
  could be identical while the files live somewhere else.* Verified against the
  catalog: graph-maintenance's `what` is the **dependency graph** ("kept consistent and
  minimal, information flowing one way"), and relocation leaves every edge untouched;
  meanwhile the resulting *state* (two homes for one kind of thing) is already
  speakable in its existing signature ("**one home per kind of information**, placed by
  which consumer must trip over it", `decision-0040`) — no widening needed there. What
  the incident actually violated is `inv-clarify-before-commit`, near-verbatim: its
  `what` is "ambiguity in an upstream is actively surfaced and resolved… **never
  silently guessed**", and its violated pair — "an agent silently picks one reading of
  a vague spec, builds it" — is the incident with "spec" generalized to **structural
  scope**. The behavior to ship is *surface the open choice*; that is clarify's
  territory.
- **Why the fix is a payload change, not a repo rule.** A CLAUDE.md line fixes trellis;
  the overlay's rule fragment fixes every project the payload reaches — the product's
  entire premise. A session-memory note (kept) nudges only this maintainer's agent;
  neither ships.
- **What this is not.** The broader observation — the shipped rules are standing
  obligations that do not encode their *firing moments* (invariant 8's "triggers, not
  vigilance", unapplied to the readout itself) — is a format question with an
  experiment logged and sequenced (trellis#166). This decision is the **content slice
  only**: one rule's wording widens; the format stays.

## Decision

**1. The catalog directive widens** (`core/catalog/signature-catalog-v1.md`,
`inv-clarify-before-commit` — the single home; everything below derives from it,
`decision-0028`). Appended to the existing directive text:

> A structural choice your ask leaves open — like whether existing stock migrates into
> a new pattern you are introducing, or stays exempt — is the same ambiguity: surface
> it and ask, never grandfather silently.

**2. The signature gains the clause**: *a change that introduces a pattern surfaces its
retrofit question — "X migrates; Y stays, because …" — for the human to rule on,
instead of resolving it in confident boundary prose.*

**3. One honored/violated pair is added via the signature-pair door** (originated by
`decision-0027`; `decision-0040` is the worked precedent this replays — pairs on
existing invariants; **the count stays 14**):

- honored: *(structure)* a change that introduces a pattern ships with its retrofit
  question — "the old suite now sits outside this convention: migrate or exempt?" —
  and the human rules on it.
- violated: *(structure)* a new directory convention lands while the old stock stays
  loose beside it, exempted by confident boundary prose nobody approved — two
  conventions in one tree.

**4. The rendered rule's ✗ line extends in the same breath** (still one ✗ bullet, the
readout format unchanged): "…and it's the wrong one — or a new convention lands and the
old stock stays loose beside it, exempted by prose nobody approved."

**5. The wording is our synthesis** (naming guardrail): no external provenance is
implied; the evidence is one in-house instance, named above, and the pair door exists
precisely for refinements at this evidence level.

## Consequences

- **The derived chain regenerates in the same change** (`decision-0028`; the
  `decision-0051` machinery makes this cheap): the catalog edit flows to the
  generator's render → `plugins/trellis/reference/rules/inv-clarify-before-commit.md`
  fragment → `rules.md` assembly → both inline-block sandwiches → `checksums` →
  `version` stamp → `install.sh` bundle manifest (Go-test-guarded) → this repo's own
  `.trellis/internal/` overlay + managed block. Family repos and consumers pick the
  widened rule up on their next refresh.
- **`inv-graph-maintenance` is deliberately untouched** — the maintainer's challenge,
  recorded above, is the reason; future readers asking "why isn't retrofit scope under
  graph maintenance?" find the answer here rather than re-litigating it.
- **One experiment-facing note**: `inv-clarify-before-commit` is the deactivated rule
  in the running `annotation-vs-absence` experiment (`research-0012`). Its fragment
  text changing mid-experiment would alter the readout under test — so **this
  decision's build must not land between that experiment's batches**: land it before
  the first batch or after the run completes, and the experiment's `runs/provenance`
  payload stamp will show which text was live either way. Flagged here so the
  sequencing is a choice, not an accident.
- **The catalog is amended, not superseded**: `signature-catalog-v1` remains the
  current-truth set; this decision is the recorded authority for the edit. Coverage
  arithmetic unchanged (14 assessable slugs).
- **The repo-local mechanisms stand down where redundant**: the offered CLAUDE.md
  method line is *not* added (one home per kind — the overlay carries the behavior here
  too); the maintainer's-agent session memory remains as a personal nudge outside the
  repo.
- **trellis#166 is unaffected** — if the trigger-format experiment later lands
  fires-when lines, this rule's moment ("your ask leaves a structural choice open") is
  written by that decision, not this one.

## Open questions

- **Does the widened wording pull its weight in the readout?** It lengthens one rule
  bullet in every session's context. If the annotation-vs-absence and #166 results show
  rule-text form dominates rule-text content, revisit the wording's length there.
- **Second evidence instance.** One incident justifies the pair door, not more; if the
  silent-structural-guess failure recurs *despite* the widened rule, that is #166's
  trigger-format hypothesis gaining evidence — note it there.

## Self-check (gate)

Uses the signature-pair door as designed (refinement on an existing invariant; count
stays 14 — `inv-minimal-first` holds; no new invariant, no new mechanism — and the
re-homing *shrank* the change: one entry widened, none overloaded). The single-home
rule is respected: one catalog edit, everything else derived (`decision-0028`), chain
named, and the experiment-sequencing hazard surfaced rather than discovered later
(`floor-transparency`). The rejected alternative is recorded with the maintainer's
reasoning, not silently replaced (`inv-auditable-archive`'s why-is-it-this-way).
`depends_on` carries genuine coupling; `decision-0028`/`decision-0051` are context
(`informed_by`, `decision-0047`). Left at `draft`: the author does not grade its own
decision — the gate and any `approved` flip are the maintainer's intent act
(`decision-0046`), ideally after an independent pass (`inv-independent-judgment`).

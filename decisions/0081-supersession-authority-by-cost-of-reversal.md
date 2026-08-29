---
id: decision-0081
type: decision
status: approved  # maintainer's intent act 2026-08-29, in-conversation on the PR ("So I accept the draft") after reading the drafted record — this flip records it (decision-0046, decision-0022). Author (agent) != approver (maintainer). Scope of the acceptance: this record's answers and recommendations as the basis for the catalog change, which is a SEPARATE change still owed. The in-PR approved flip is legitimate here because it records a human act (decision-0046 point 2; 0042's no-in-diff-approved rule is superseded in part by 0046)
depends_on: [invariants-v1, signature-catalog-v1, decision-0022, decision-0046, decision-0074]
informed_by: [decision-0008, decision-0021, decision-0024, decision-0027, decision-0040, decision-0052]
owner: agent
date: 2026-08-28
---

> **Provenance.** Produced by an unattended Cyrus session for **TRL-23**, from a forge session in
> `kodhama/math-quest` (MQ-72, 2026-08-28). The consuming half is tracked there (MQ-75, MQ-76);
> this is the portable half. **The session was scope-limited to a proposal**: answer the four open
> questions with evidence, draft the wording, change no catalog file, bump no version, ask nothing.
> It did none of the excluded acts — `core/catalog/signature-catalog-v1.md`, both rendered copies,
> every `rules*.toml`, and `plugins/trellis/VERSION` are untouched in this change.

# 0081 — Supersession authority scaled by cost of reversal *(proposal, not a decision)*

> **Evidence is as of 2026-08-28; line cites do not resolve against `main`.** This record was
> authored on a branch cut before `decision-0079` (which deleted `specs/` and trimmed `AGENTS.md`
> from 163 lines to 95) and before `decision-0082` (which retired the `status` field). Its block
> quotes of `AGENTS.md` — the *Gates* bullet, *"in conversation, review, or by merging"*,
> *"Downstream consumes only gated/approved upstream, never drafts"* — quoted the file **as it
> stood that day**; several of those sentences no longer exist. Its `AGENTS.md:NN`,
> `signature-catalog-v1.md:NN` and `trellis-invariants-v1.md:NN` line numbers are likewise
> pre-trim and are off against `main`. The record is **not edited to chase them**
> (`inv-auditable-archive`); read its quotes against the tree at `9d0675a`. Its `spec-0001` /
> `spec-0002` citations resolve through `decision-0079`'s retired-artifacts registry, and the
> forward edits it *recommends* to those files now land on `core/rubrics/artifact-contract.md`
> and `core/schemas/typed-artifacts.md`.

> **What this is — and what the maintainer accepted.** Written as a proposal for the maintainer to
> rule on; **accepted in conversation on 2026-08-29** (*"So I accept the draft"*), which is the
> intent act the frontmatter flip records (`decision-0046`).
>
> **The acceptance is of the argument, not of a catalog change.** No catalog file, row set or
> `VERSION` is touched here — that edit is a **separate change, still owed**, and it carries the
> version sequencing in `## Consequences`. The `## Decision` section states a *recommendation* plus
> the options it was chosen from; **Recommendation 1 is explicitly held at moderate confidence and
> keeps option B live**, so acceptance of this record does not settle the extend-vs-mint question —
> `## Open questions` is the live list.

## Context

### The gap is real, and the issue quotes it accurately

`inv-deliberate-succession`'s `what` names decisions explicitly, and its `directive` gives an agent
exactly two moves —
`core/catalog/signature-catalog-v1.md:189-193`:

> what: wherever new work meets an existing base — code, conventions, decisions, data, docs — the
> boundary is **decided out loud in both directions** … The prior version is evidence to weigh,
> never gravity to drift with and never debris to step over.
>
> directive: When you introduce a new pattern, say what the existing stock outside it should
> become — **migrate it, or name the exemption and ask**; never resolve it silently in prose. …

Verified: **neither move is "retire the prior decision yourself."** The entry's signature does name
supersession — *"the superseded version reachable and marked superseded, never edited in substance or
silently dropped"* (`:201-202`) — but that specifies supersession's **form**, never its **authority**.

The absence is corpus-wide, not local to this entry. `spec-0001:153` says *"Decisions are append-only:
supersede, never edit a ratified one"* and `AGENTS.md:49` says *"You supersede (with a forward pointer),
never edit, a ratified decision"* — both impersonal. **No ratified artifact in this repo qualifies who
may perform the supersede act.** By contrast the *approval* act is qualified absolutely
(`decision-0046` point 2: *"An agent writing `approved` with no human act is forbidden"*). Authority is
specified for approval and unspecified for supersession.

### The evidence, at its actual size

The measured signal from math-quest (2026-08-28, ~99 answered elicitations across 15 sessions) is
strong but **should not be quoted at its headline number for this fix**:

| Bucket | Share | What the fix does to it |
|---|---|---|
| Answers that merely confirm the agent's own recommendation | ~80% (~79) | Unmeasured overlap — some of these are supersession asks, some are not |
| Answers that change something, of the *"taken too literally"* shape | ~half of the remaining ~20% (**~10**) | **Directly targeted** |

So the directly-evidenced target is **~10% of elicitations**, not 80%. That is still a real, recurring,
maintainer-flagged failure — and it is the failure this invariant exists to prevent, occurring in a
project that has the invariant active — but the 80% figure measures a broader over-escalation problem
than this proposal addresses. **Recorded at its true size so the fix is not oversold.**
*These numbers are the issue's; they were not independently reproducible from this repo.*

### Both cited sources were read; both need a correction

**The Bezos attribution is CONFIRMED from the primary source** — the letter was fetched and read, not
summarized. Amazon 2015 Letter to Shareholders, section **"Invention Machine"**:

> Some decisions are consequential and irreversible or nearly irreversible – **one-way doors** – and
> these decisions must be made methodically, carefully, slowly, with great deliberation and
> consultation. … We can call these Type 1 decisions. But most decisions aren't like that – they are
> changeable, reversible – they're **two-way doors**. … Type 2 decisions can and should be made
> quickly by high judgment individuals or small groups.

Two corrections this forces on the issue's framing:

1. **The test has two conjoined prongs, and the issue keeps only one.** The letter says *"consequential
   **and** irreversible."* The issue's operative test — *"how hard would it be to change?"* — is
   reversibility alone. A decision can be cheap to reverse and still consequential while it stands
   (a security default, a data-retention setting). Dropping the consequentiality prong widens agent
   latitude beyond the source. **The proposed wording below restores it as a third explicit clause
   ("cheap to live with meanwhile").** An earlier draft claimed "hard to detect" already covered it;
   the independent review showed that is wrong — detectability and consequence magnitude are
   different axes, and a cheap-to-reverse, immediately visible change (a security default, a
   retention setting) can still do serious harm while it stands. The prong is now carried explicitly
   rather than folded into detectability.
2. **The letter's footnote is the sharpest argument *against* a loosely-specified version of this
   proposal**, and it is why the default matters:

   > Any companies that habitually use the light-weight Type 2 decision-making process to make Type 1
   > decisions go extinct before they get large.

   The error is **asymmetric**. Misclassifying one-way as two-way is the fatal direction — which is
   decisive when the party doing the classifying is the same agent that benefits from the latitude.

**The second quoted source was located and is weaker than its use implies.** The verbatim quotes
(*"If this turned out to be wrong, how hard would it be to change?"*, *"a week of refactoring"* /
*"change a config value"*, *"immutable during implementation"*) come from Sagar Mandal, *"Agentic
Engineering, Part 8: Decision Classification"* (sagarmandal.com, 2026-03-15) — a **personal blog,
prescriptive opinion, no data or citations**. Its actual claim is narrower than the issue's use of it:
*"If it finds a better logging pattern during implementation, it can adjust."* That is **latitude
within a decision's scope during implementation**, not **authority to retire the decision record**.
Those are different acts, and the extension from one to the other is the issue's inference, not the
source's claim.

### The constraint that decides the shape: only the catalog is portable

The issue wants this to *"apply to every project he works in."* Verified against the shipped payload:
`plugins/trellis/reference/` contains `invariants.md`, `rules.md`, `rules-a/b.toml`, the block
sandwiches, `checksums`, `version` — **no spec, no rubric, no artifact contract**. `spec-0001` and
`core/rubrics/artifact-contract.md` do not ship.

**Therefore any contract-layer or frontmatter-layer answer is unreachable by consumers.** A portable
change must land in the catalog, as directive/signature/example text.

### This repo carries the same trap the issue found in math-quest

The issue's aggravating factor — a consumer's own `AGENTS.md` defeating the overlay — is **not
math-quest-specific.** `AGENTS.md:51`, this repo:

> - **Gates.** Human approval at the **intent** layer (vision, decisions, the invariant set).

Decisions are declared intent-layer here too, and an agent reads project law before plugin overlay.

**Scope of this finding, narrowed by the second independent review:** it bites the issue's **literal**
framing (agent supersedes with no human act) — there, `AGENTS.md:51` defeats the clause exactly as
math-quest's does. It does **not** bite the recommended framing, because `AGENTS.md:55` already counts
approval *"in conversation, review, or by merging"*, which is the gate the recommendation preserves.
An earlier draft stated the trap unconditionally; that was too strong. See `## Consequences`.

### The blocker that reframes the whole proposal

Read literally — *"an agent may supersede a two-way-door decision **on its own**"* — the proposal
collides with a ratified floor. In this repo supersession is performed by minting a **successor
decision**, and a successor reaching `approved` requires a human intent act (`decision-0046` point 2,
absolute). `floor-intent-gate` adds *"Unsure whether a human must approve? Assume yes."*

**But the collision narrows sharply once the real problem is named.** The measured failure is not that
agents lack authority — it is that they **stop mid-session and block** on a question the ratification
gate would have answered anyway. `decision-0022` already makes **merge a ratification act**. So the
proposed fix asks for **no new authority**: for cheap-to-reverse decisions the agent stops *asking* and
starts *drafting the supersession into the PR*, where the human ratifies it at the gate they were
already going to use.

That is strictly more conservative than the issue proposes, it is built from mechanisms this repo has
already ratified, and it removes the blocking wait — which is the thing the evidence actually measures.

**Two honesty notes on this reframing, both from the independent review** (§ Self-check):

- *`decision-0046` point 5 does not support it as squarely as an earlier draft claimed.* Point 5
  governs ambiguity about whether **an existing human instruction already was** the approval act —
  *"an agent must not infer the gate has opened, nor stall a real approval by failing to recognize
  it."* It does **not** convert a *possible future merge* into approval already given. It is an
  **analogous posture** (the framework does treat stalling as a failure mode, not only inferring),
  not authority for this reframing. Cited at that weaker weight.
- *The reframing leaves one genuine seam open* — the agent-drafted successor is `status: draft` at
  the moment the same PR implements against it, and `AGENTS.md:36` says *"Downstream consumes only
  gated/approved upstream, never drafts."* Existing records dodge this by flipping to
  `gated`/`approved` **in the PR**, recording a human act that has already happened — but an
  agent-initiated supersession has no such act at drafting time. **Unresolved; carried to Open
  question 7 rather than papered over.**

## Decision

**Proposed — not taken.** Four recommendations, each with the options it was chosen from.

### Recommendation 1 (issue Q1) — Extend `inv-deliberate-succession`; do **not** mint a 16th slug

The repo's own minting rubric, quoted verbatim at `decisions/0040-reverse-ports-from-instance-1.md:27-29`:

> the set stays at 14 (minimal-first, and the same rubric that retired `inv-reference-relationship`:
> **no new mechanism → no new invariant**).

`decision-0040` point 4 is the closest precedent — a sharpening that folded in rather than minting:

> This sharpens the intent face's mechanism: it is *how* an agent avoids flattery in the moment, not a
> new principle — so it folds into the entry (**directive extended**, signature clause, a *(collab)*
> pair) rather than minting a slug.

Supporting reasons: `inv-deliberate-succession` is **`provisional`** and carries its own open question
*"Does it pull its weight?"* (`decision-0074:161`); **TRL-4** is a live audit (Backlog, verified
2026-08-28) asking whether the set should *shrink*, taking this very entry as its corpus; and — the
strongest one — **an extension changes no slug, so no consumer's row set breaks** (Recommendation 3).

**This recommendation is held at lower confidence than the other three, for two reasons that must not
be buried.**

*First, the "no new mechanism" argument is contestable, and the independent review contested it
successfully.* An earlier draft asserted flatly that agent-vs-human authority already exists as
`dial-gatekeeper`, so this only applies an existing dial at a new granularity. That does not survive
contact with `decision-0024` point 1: *"The host project owns its gate declarations … Trellis **reads
and respects** them; it does not build, choose, or impose a per-gate map"* — with per-gate
configuration deferred to v2 (point 5). **An agent-run, per-collision classifier is arguably a new
routing mechanism, and a finer one than the map `0024` already deferred.** If that reading holds,
option B does *not* fail `decision-0021`'s test and the choice is genuinely open.

*Second, the maintainer has pre-registered a warning that lands on this very recommendation.* **TRL-4**:

> **Why the agentic pushback is weak evidence.** Every agent that reviewed `decision-0074` argued
> against growth — but they all read the same rule set, one shipping `inv-minimal-first` and
> prune-bias. Four verdicts are one opinion with four voices.

This proposal was written by an agent reading that same rule set, and it argues against growth.
**Discount it accordingly.** *(One partial counter-observation: the independent reviewer here was a
different model family and argued the opposite — that option B is viable — so the anti-growth verdict
is at least not unanimous across readers of this rule set.)*

| Option | Assessment |
|---|---|
| **(A) Extend the entry** (directive + signature clause + one pair) | **Recommended, at moderate confidence.** Precedented (`0040` p4), count-preserving, portable, breaks no consumer |
| (B) Mint a 16th slug | **Not dismissible.** Costs a row-set change (breaks the six 14-row repos further) and adds beside a `provisional` entry mid-audit — but `decision-0024` supports reading this as a genuinely new mechanism, which is exactly what `0021`'s test licenses minting for |
| (C) Contract-layer only (`spec-0001` supersession clause) | Correct layer for *frontmatter mechanics* (`0040` p5 precedent) — but **`spec-0001` does not ship**, so it never reaches a consumer. Fails the portability requirement |

### Recommendation 2 (issue Q2) — Apply the test **at collision**, with "unclassified ⇒ one-way"

**Recommended: at the moment a later agent collides with the decision**, defaulting to one-way when
undetermined.

This is forced by the same portability finding: a `cost_of_reversal:` frontmatter key would be a
trellis-local convention that consumers never receive. It is also a granularity this repo has
**explicitly deferred twice** — `decision-0024` point 5: *"Per-gate configuration by Trellis → v2.
Enumerating and choosing per gate is too cumbersome for v0."*

And the default **dissolves the retroactive-classification cost the issue worried about**: with
"unclassified ⇒ one-way," none of the **69 live decisions** needs backfilling. Nothing changes for any
existing record until an agent actually collides with one and judges it cheap to reverse. The default
is also already the framework's stated posture (`floor-intent-gate`: *"Unsure whether a human must
approve? Assume yes"*) and the correct side of Bezos's asymmetry footnote.

**The limitation, stated where the recommendation is made rather than only in the open questions:**
"unclassified ⇒ one-way" protects only **recognized** uncertainty. It does nothing against **confident
misclassification** — an agent that is sure a decision is cheap to reverse and is wrong. Since the
Context calls that asymmetry *decisive*, this mitigation is **partial, and known to be partial.**
Whether it is sufficient is a judgment this session cannot make (Open question 3).

*Recording at decision time remains available later as a **local** enrichment* — a `spec-0001`
amendment in `decision-0040` point 5's mould. Two notes if it is ever wanted: the frontmatter field set
is **open in practice** (`informed_by` was added with no schema row and no version bump,
`spec-0001:47-56`), and **28 of 73 decisions already carry unstructured governance prose in a `status:`
comment** — so per-decision governance metadata is de-facto existing practice, just unparseable.

### Recommendation 3 (issue Q3) — The trap is real and currently live; the extension avoids it

**Yes, the row-set trap applies — and it is not hypothetical.** Measured on this machine today
(read-only inspection of `.trellis/rules.toml` and `~/.claude/plugins/installed_plugins.json`):

| Repo | Rows | Has `inv-deliberate-succession` |
|---|---|---|
| math-quest · trellis (`Projects/`) · this worktree | **15** | yes |
| **grove · stewards · kodhama · wisp · design-system · spore** | **14** | **NO** |

**Six consumer repos are still on the 14-row set** — i.e. the previous slug addition
(`decision-0074`, 14 → 15) has *not* propagated. That record carries `date: 2026-08-19` and a
maintainer intent act of 2026-08-22, so the un-propagated window is **six to nine days**, not the
three weeks an earlier draft claimed — that interval belongs to the separate 0.4.0 caching window in
the `d4a2c7b` quote below, and was carried onto the wrong clock. Whatever this proposal does, the
drift is a pre-existing open wound worth its own issue.

*Verification note: the six above were counted directly and are firm. A seventh 14-row copy under
the plugin marketplace cache was reported by a search agent but **could not be independently
confirmed** — the path was not readable from this session, so it is excluded from the count rather
than asserted. The two `math-quest-*` directories are eval fixtures, also excluded.*

The failure is **symmetric inside a validator** — `staleness.sh:586-596` aggregates `missing:`,
`unknown:` and `duplicate:` into one report, and any of them blocks injection at `:599-601` — but
**asymmetric across delivery paths**:

- **Plugin-native:** hard `TRELLIS_RULES_NOT_LOADED`, nothing injected.
- **Curl install:** **still unvalidated.** The curl branch `emit`s *"That file and `.trellis/rules.toml`
  govern this session"* at `staleness.sh:444` and `exit 0`s at `:445`; the row validation lives at
  `staleness.sh:558-602`, which that path never reaches. **`decision-0074:126-133` disclosed this and
  deferred the fix; it is tracked as TRL-2 and still Backlog (verified 2026-08-28).** A 16th slug would
  silently deliver an inactive rule while claiming to govern.
- **Preset rows:** *"without a row the rule ships but is inactive"* (`decision-0074:121`).

**The recommended extension sidesteps all of it.** Row sets key on **slugs**, and a directive extension
adds none — every consumer's `rules.toml` stays valid. The cost is not zero, but it is a different
class: the payload content hash changes, so vendored consumers get a **staleness nudge (stale-but-valid)**
rather than a **hard fail or a silent-inactive rule**.

**One sequencing requirement, from a mistake already made.** Commit `d4a2c7b` records that
`decision-0074` shipped the 15th rule **without moving `VERSION` off 0.4.0**, so *"any consumer who
pulled in the three weeks since is cached at 0.4.0 and would never see the 15th rule — governed by
fourteen while the repo believed it shipped fifteen."* **A directive change is invisible to consumers
unless `plugins/trellis/VERSION` moves.** Current state: `VERSION` = `0.5.0`, payload stamp =
`payload@457ab23911d9`. *(Not bumped in this change — out of scope.)*

### Recommendation 4 (issue Q4) — Provenance discharged; see the two corrections above

Attribution **confirmed against the primary source**, both phrases literal, section "Invention Machine."
The letter does **not** use *"cost of reversal"* — that phrasing is ours, and under `decision-0017` /
`core/lexicon.md` it likely needs a lexicon entry if adopted. Per the naming guardrail, the framing is
**attributed to Amazon (2015)**, and the *application to agent supersession authority* is **our
synthesis** — the source applies it to human organizational decision-making, not to agents.

### The proposed wording

*Drafted for review. **Not applied to any catalog file.** Written without trellis or math-quest nouns —
`decision-0040`'s portability admission test.*

**Directive** — appended to `inv-deliberate-succession`'s existing directive:

> When the prior version is a recorded decision, weigh it by what being wrong would cost: if it would
> be cheap to undo, quick to notice, and cheap to live with meanwhile, **draft its replacement into
> the work** — out loud, with the original rationale surfaced and a forward pointer left behind —
> instead of stopping to ask; it still goes through the project's ordinary approval gate, just not as
> a blocking question. If undoing it would be expensive, a mistake would be slow to surface, or the
> damage while it stands would be serious, **raise it before you act**. When you cannot tell, raise it.

**`why`** — appended:

> …and **a recorded decision is evidence, not a wall** — an agent that cannot retire a cheap one
> reports the obvious fix as impossible and blocks, and the human spends a wait confirming what the
> agent already knew.

**Signature clause** — appended:

> supersession is weighed by cost of reversal rather than by the seniority of the record; a decision
> that is cheap to reverse, quick to notice **and low-stakes while it stands** is superseded *in the
> work* with its prior rationale surfaced and its forward pointer intact, not raised as a blocking
> question; one that is expensive to reverse, slow to surface **or damaging while it stands** is
> raised before acting, as is an unclassified one.

**Matched pair** (`decision-0027` point 1 — one use case, failing then fixed, same layer tag):

> - honored *(decision)*: a recorded log-format choice blocks the obvious fix; the agent supersedes it
>   in the same change — quoting the original rationale and why it no longer holds, leaving the forward
>   pointer — and the reviewer ratifies it at the gate they were already using.
> - violated *(decision)*: the same log-format record is treated as immovable; the agent reports the
>   obvious fix as impossible and waits, and the human replies that the record could simply be
>   superseded.

**Dials.** The entry's own dials are **unchanged**: `default_C1: default-on-but-skippable` ·
`default_C2: human`. What the clause adds is a *per-collision* reading of `dial-gatekeeper` for the
decision being collided with — `human` when it is expensive to reverse, slow to surface, damaging
while it stands, or unclassified; otherwise the project's ordinary review gate. **`none` is not proposed.** It is unused across all 15
entries today (7 `human`, 8 `independent-agent`), and choosing it here would put the agent's own
supersession beyond any check — which `inv-independent-judgment` forbids. *Open sub-question below.*

## Consequences

*(These would follow **if** the recommendation is accepted. None has been enacted.)*

- **The derived chain must regenerate in the same change** (`decision-0028`) — catalog → the plugin's
  `reference/` render → `rules.md` → both inline-block sandwiches → `checksums` → `version` stamp →
  `install.sh` bundle manifest → the invariant scorecard → `docs/invariants.html` → `cli/assets/`. A
  directive change lands in the **always-loaded block**, so it touches every rendered surface.
  **Attribution, since an earlier draft blurred it:** `decision-0074:118-125` enumerates the render
  chain through the scorecard *and* names a second, **contract** layer this list must not drop —
  `spec-0007`'s slug inventory and activation-row predicate, `spec-0002` §1 check 2 + AC1,
  `core/rubrics/artifact-contract.md`, and the `corpus-reviewer` checklist. `docs/invariants.html`
  and `cli/assets/` come from `decision-0028:33-34`, not from 0074. `decision-0040:78-80` is the
  narrower worked instance, cited only for the point that a directive change reaches the
  always-loaded block. *(A directive-only change plausibly leaves the contract layer alone — it
  changes no count — but that must be checked at the edit, not assumed here.)*
- **`plugins/trellis/VERSION` must move**, or no plugin consumer sees it (the `d4a2c7b` lesson).
- **No `rules*.toml` row changes**, and no consumer `rules.toml` needs editing — the count stays at 15.
- **`AGENTS.md:51` most likely needs no amendment** — an earlier draft asserted it did, and the
  independent review showed that claim is self-contradictory. `AGENTS.md:51-58` already states that
  a human's approval *"in conversation, review, or **by merging**"* is the ratification act, which is
  exactly the workflow the recommendation preserves. **The Context's "trellis carries the same trap"
  finding therefore applies to the issue's *literal* framing, not to the recommended one** — and if
  the maintainer chooses the literal framing, amending `AGENTS.md:51` becomes required *and* "no new
  authority" stops being true. Recorded as the fork it is, not as an unconditional dependent.
- **The readout grows by roughly three sentences** on an entry `decision-0074:148` already flagged as
  a readout-budget concern.
- **The pair count reaches five**, exceeding `decision-0027` point 3's two. Precedented — the enforced
  gates say `≥2` and this entry already carries four — and a *(decision)* pair is *"a genuinely new
  layer"* against existing structure/design/docs/metrics tags (`spec-0002`'s ceiling rule, per
  `decision-0040:83-85`). **Flagged, not smuggled**; it compounds a budget concern 0074 already raised.
- **No `decision-0046` amendment is needed** under the recommended reading **as it stands today** —
  the human performs the intent act at the merge gate. *Caveat, per the second independent review:*
  Open question 2 leaves live the alternative that "the ordinary gate" means an `independent-agent`
  check. Under **that** branch a human does not perform the act, and `floor-intent-gate`'s ratchet
  clause (*"a human, or, by ratchet, an independent check the human authorized"*) is what would carry
  it — permitted by the floor, but it is a different authority story and this bullet does not cover
  it. If the maintainer instead wants the issue's literal framing
  (agent supersedes with *no* human act), that **does** require amending `decision-0046` point 2 and
  confronting `floor-intent-gate` directly. Recorded so the choice is visible.
- **TRL-2 stays open and gets worse in expectation** — every future catalog change with a row
  delta hits an unvalidated curl path. Not this proposal's to fix; named because it is load-bearing.
- **`inv-minimal-first` cuts against even the extension.** The honest counter is that this buys a
  *directive clause and one worked pair* on an entry that already owns the subject, at no slug cost —
  the cheapest shape available that is still portable.

## Open questions

1. **Does the maintainer accept the reframing?** The recommendation deliberately does *not* grant
   agents unilateral supersession — it re-routes a blocking mid-session question to the existing merge
   gate. That is more conservative than the issue asked for. **If the intent really is authority
   without a human act, say so** — it is achievable, but it costs a `decision-0046` amendment and a
   direct argument against `floor-intent-gate`, and this session had no mandate to make either.
2. **Cheap-to-reverse ratifier — is the ordinary review gate the right one?** Recommended above as
   the project's ordinary gate. **`none` is not an available option and an earlier draft was wrong to
   offer it:** `dial-gatekeeper` permits `none` *"never at the intent gate"*
   (`core/invariants/trellis-invariants-v1.md:259-260`), and decisions are intent-layer under
   `AGENTS.md:51`. Caught by the independent review; corrected rather than left as a choice the floor
   forbids. What remains genuinely open is whether "the ordinary gate" should mean human review or an
   `independent-agent` check where a project has one.
3. **Who classifies, and is that safe?** The classifying agent is the one that benefits from the
   latitude. "Unclassified ⇒ one-way" is the proposed mitigation, and Bezos's footnote is the reason
   it must not be relaxed. Whether that is *sufficient* is a judgment this session cannot make.
4. **Should the 14-row consumer fleet be fixed first?** Six repos never received the 15th rule.
   Shipping a 16th change onto that base compounds a known drift. Plausibly a prerequisite, not a
   follow-up — but sequencing is the maintainer's call.
5. **Does `cost of reversal` need a `core/lexicon.md` entry** (`decision-0017`)? The phrase is ours;
   the door metaphor is Amazon's.
6. **Is this the unnamed dimension TRL-4 is looking for?** TRL-4 asks whether the set carries a
   dimension along which entries could collapse. "Authority scaled by reversibility" is a candidate
   answer, and TRL-4 already takes `inv-deliberate-succession` as its starting corpus. If so, this
   proposal should be folded into that audit rather than landed ahead of it.
7. ~~**How does an agent-drafted supersession avoid consuming its own draft?**~~ **Dissolved by
   `decision-0082`** (same change as this record's landing): there is no `status` field, so there
   is no draft to consume and nothing to wait for. Retained below as the reasoning that led
   there. The successor is
   `status: draft` while the same PR implements against it, and `AGENTS.md:36` forbids downstream
   consumption of drafts. Existing records dodge this by flipping in-PR to record a human act that
   already happened; an agent-initiated supersession has none at drafting time. Candidate answers:
   the implementation waits for the flip (reintroducing a wait, though a review-time one rather than
   a mid-session one); or `gated` is accepted as sufficient for same-PR consumption. **Not resolved
   here — it is an intent-layer question.**

## Self-check (gate)

Scope limit honored: no catalog file, no row set, no `VERSION`, no math-quest file touched; nothing
asked; no elicitation posted. Both cited sources were **read at the primary source**, not accepted from
search summaries — which produced two corrections *against* the issue's framing (the dropped
consequentiality prong; the blog's narrower claim) and one structural blocker the issue did not
anticipate (`decision-0046` point 2). The headline 80% statistic is restated at the size the fix
actually addresses (~10%) rather than quoted at its most favorable. The recommendation argues for the
*smallest* portable shape and records `inv-minimal-first`'s objection to even that.

**BMAD was not used: this repo has no BMAD installation** (no `.bmad*` or `bmad*` at the root; the
issue anticipated this). Ordinary careful work was done instead — not a fabricated BMAD run.

**Independently reviewed twice, and the review changed the substance — it did not ratify it.**
The repo's `corpus-reviewer` checked contract conformance (PASS; one off-by-two line cite fixed).
A **different model family** then ran an adversarial pass at high effort and produced twelve findings.
**Ten were accepted and are fixed above**, including four that weaken the author's own case:

1. The proposed directive implied *unilateral* authority, contradicting this record's own Consequences
   — **rewritten** to keep the approval gate explicit.
2. `decision-0046` point 5 was **over-read** as authority for the reframing — **downgraded** to an
   analogy.
3. *"Hard to detect" does not restore Amazon's consequentiality prong* — correct; the prong is now a
   **third explicit clause**. This also contradicts the issue's claim that blast radius and
   detectability collapse into one question.
4. **The "no new mechanism" argument does not survive `decision-0024`** — Recommendation 1 is
   downgraded from a confident call to a moderate-confidence one with option B kept live.

Also fixed: `none` offered where the floor forbids it; an `AGENTS.md` amendment claimed as needed when
it is not; the seven-vs-six consumer count; dead GitHub pointers (#239/#241) replaced with the live
TRL-4/TRL-2, whose Backlog status was verified rather than assumed; two `staleness.sh` line numbers;
one overclaimed citation. **Two findings were not adopted as stated** — the draft-consumption seam and
the residual misclassification risk are real but are intent-layer questions this session may not
settle, so they are carried as Open questions 7 and 3 rather than resolved.

**A second contract review then caught four more defects in the revision itself**, which is the
honest reason this record should not be read as settled:

- **The consequentiality prong had been restored in the directive only** — the `signature` clause and
  the dials paragraph still carried the two-prong test. Since the signature is the assessable text
  the catalog gates read, the *weaker* version sat in the *more* load-bearing place. Fixed in all
  three.
- **`AGENTS.md:44` was wrong twice** (the quoted sentence is at `:36`) — an off-by-eight that survived
  the first review, whose own self-check had claimed line cites were cleared.
- **The derived chain over-attributed `decision-0074`** — `docs/invariants.html` and `cli/assets/` are
  `decision-0028`'s, and 0074's second *contract* layer had been dropped. Both corrected.
- **"Three weeks on" was wrong** — `decision-0074` dates to 2026-08-19/08-22, so the un-propagated
  window is six to nine days; the three-week figure belongs to a different clock.

Two edge classifications were also corrected: `decision-0022` and `decision-0046` moved from
`informed_by` to `depends_on`, since this record would be *wrong* if either changed — that is
coupling, not provenance (`decision-0047`).

**Read that pattern as a caution.** Two review rounds, and the second still found four defects in
work the first had passed — including one in the very correction the record advertises most loudly.
The author is an agent; the defect rate here is the argument for the maintainer reading the
load-bearing quotes against their sources rather than taking this record's word for them.

**The remaining epistemic weakness is disclosed in Recommendation 1 itself:** TRL-4 pre-registers that
an agent arguing against catalog growth is weak evidence, because it is reading a rule set that ships
`inv-minimal-first`. This proposal is exactly that. **Discount Recommendation 1 accordingly** — it is
the one call here that most needs the maintainer's independent judgment rather than deference.

The math-quest elicitation statistics were taken from the issue and could **not** be verified from this
repo — flagged in-line rather than presented as established. The author is an agent and does not grade
its own work (`inv-independent-judgment`). *(Authored intent: stay `draft` on a draft PR so the gate
is the maintainer's. Superseded by events — the maintainer accepted it in conversation on
2026-08-29, which the frontmatter records, and `decision-0082` retired the `status` field
altogether in the change that lands this record.)*

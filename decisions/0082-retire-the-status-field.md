---
id: decision-0082
type: decision
depends_on: [decision-0022, decision-0037, decision-0042, decision-0046]
informed_by: [decision-0011, decision-0040, decision-0054, decision-0081]
changes: [decision-0080]
owner: agent
date: 2026-08-29
---

> **Provenance.** Directed by the maintainer in session on 2026-08-29, on PR #252: *"This draft
> stuff never worked and it's only bureaucracy… I want to remove that as a status"*, then, asked
> whether he meant `draft` alone or the ladder: *"I mean drop the whole status field, tbh… Or make
> it usable, don't make me have to go back and do PRs to approve something that merged already."*
>
> **This record is the first artifact to carry no `status` field.** That is the change, applied to
> itself.

# 0082 — retire the `status` field; merging is the acceptance

## Context

- **The maintainer's test, in his words.** *"Those status work if I never have to think about them
  once, they are for your guidance, and if you need to tell me 'I merged but it's gated' or 'it's
  still in draft so I can't PR' after I just said to PR, forcing me to an explicit approval, then
  it's just unnecessary bureaucracy."* The requirement is not "fewer states." It is: **the field
  must never generate work for the human, and never override an instruction he just gave.**

- **The field failed that test in the session that produced this record.** PR #252 was opened as a
  *draft PR* carrying a `status: draft` decision, and the maintainer was told it was a proposal not
  to merge — because `AGENTS.md:59-60` instructs exactly that. The ceremony bought nothing: he read
  the record and accepted it. The instruction generated the friction, not a lapse in following it.

- **Empirically the working state was never used.** Before `decision-0081`, **not one** of the 73
  decisions, 8 specs or 11 research notes carried `status: draft`. The corpus had 38 `ratified`,
  31 `approved`, 4 `superseded`, and — until this session — **zero drafts**. *(Counts as measured
  2026-08-28. At landing the corpus is 80 decisions, **no** specs — `decision-0079` deleted them —
  and 11 research notes: 53 `ratified`, 38 `approved`, 5 `superseded`, still **zero drafts**. The
  argument is unchanged and strengthened; the printed numbers are the as-of figures.)* `ratify-guard` has
  been guarding a state that does not occur. That is the strongest evidence for the maintainer's
  read, and it is measured, not asserted.

- **Half the reported pain came from a rule that is already dead.** *"I merged but it's gated"* is
  `decision-0042` D2's post-merge-bump mechanic — **superseded in part by `decision-0046`** on
  2026-07-11, which permits the in-PR flip. Agents have been performing a retired ritual. This
  matters for the diagnosis: part of the problem was never the design, it was drift between the
  corpus and what agents actually do. Retiring the field removes the surface that drift lives on.

- **The field is not consumer-visible.** `plugins/trellis/reference/invariants.md` carries **no
  frontmatter at all** (`decision-0054`). No file in the shipped payload has a `status:` key. So
  this change needs **no version bump, touches no activation row, and cannot break a consumer** —
  the failure mode `decision-0081` § Recommendation 3 is about does not apply here.

- **What the field actually did, and what else can do it.** Two jobs. (1) *Mark unsettled work so
  downstream does not consume it* — but "unsettled" and "not merged" are the same set here, and the
  PR's own draft/ready state already carries it, at zero cost to the human. (2) *Mark supersession*
  — but `superseded_by` / `superseded_in_part_by` are separate fields that already carry it; the
  `status: superseded` value is redundant with the presence of a forward pointer.

## Decision

**1. `status` is retired from the trellis artifact contract.** It is removed from the required
field set — now `core/rubrics/artifact-contract.md` check 1, `spec-0001` having retired with
`specs/` under `decision-0079`. New artifacts do not carry it. The required set becomes
`id / type / depends_on / owner`.

**2. Merging to `main` is the acceptance, and the only acceptance.** An artifact on `main` is
current truth and downstream may consume it. An artifact not yet merged is not. There is no
intermediate recorded state, no flip to perform, and nothing for a human to do after a merge.
This is `decision-0022`'s core — *merge is ratification* — kept and made the whole mechanic rather
than one option beside a ladder.

**3. Supersession is marked by the forward pointer, not by a status value.** An artifact carrying
`superseded_by` is superseded; one carrying `superseded_in_part_by` is partially superseded and its
remainder is live. Both fields already exist and already resolve (`decision-0040` D5). Conformance
check 7 keys on the pointer's presence instead of on a status string.

**4. `ratify-guard` is deleted.** Its only executable line greps for `^status: draft`; with no such
status it has no job. This is the check that produced *"it's still in draft so I can't PR."*

**5. The intent gate does not move.** `floor-intent-gate` is untouched: a human still approves
before something lands, and an agent still may not merge on the human's behalf without his act.
What is retired is the **record-keeping ceremony around** that act, not the act. The gate is now
exactly where the maintainer already exercises it — the merge — instead of being mirrored into a
field he must also maintain.

**6. Existing artifacts keep their `status:` lines, untouched.** All 97 of them (99 as counted 2026-08-28, before `decision-0079` deleted `specs/`). **34 carry a
trailing comment recording a specific maintainer intent act** (`decisions/0046`: *"the first
decision ratified under its own rule"*; `research/0012`: *"ok to merge!"*). Those comments are the
audit trail `inv-auditable-archive` exists to protect. Stripping them would destroy real history to
tidy a field — the field is retired **going forward**, not erased backward. Historical values read
as "accepted"; the reader needs no lifecycle to interpret them.

**7. Trellis diverges from the family enum, and says so.** The enum is
`kodhama-0004-uniform-lifecycle`, adopted by trellis in `decision-0042` and — per that record —
*"identical across the family"* (grove, kodhama, wisp, design-system, stewards). **This record
exits that adoption for trellis only.** Per `inv-deliberate-succession` the exemption is named
rather than resolved in prose: **the family question is filed, not silently answered.** Whether the
family follows is the family's call, and no other repo is touched by this change.

## Consequences

- **`decision-0080` is superseded by this record.** It landed on `main` at 17:31 on 2026-08-29 —
  after this record was drafted and while the authoring machine had no network — and reached the
  opposite conclusion: *"the approval signal is the review, not the merge"*, keeping the `status`
  field and adding `.github/workflows/ratify-flip.yml` to flip `gated → approved` automatically.
  Neither session could see the other. **This record wins on the maintainer's own stated ground:**
  he named the gate machinery as friction he wanted gone, and `0080`'s answer was a better sensor
  where this record's is no sensor to need. **Both `ratify-guard.yml` and `ratify-flip.yml` are
  deleted here** — with no `status` field both are dead code, and `ratify-flip.yml` was the more
  urgent of the two: it is `pull_request_target` with `contents: write`, and its stamp *writes
  back* a `status:` line, so leaving it would have let a future `gated` record reintroduce the
  retired field by automation. `0080`'s one surviving finding is recorded rather than
  lost: **an agent merging with the maintainer's token is indistinguishable from the maintainer**,
  so "the merge is the human act" holds only while the maintainer's account is his alone. That is
  now the load-bearing assumption of this decision, and the reason point 2 keeps
  `floor-intent-gate`'s bar on an agent merging on his behalf.


- **Superseded in part** — each gets a forward marking, substance untouched. The clause-level
  precision below was corrected after an independent review found the first pass named the wrong
  points in two records:
  - `decision-0042` — D1 (the family enum), D2's remaining flip mechanic, D4's bootstrap. Its
    merge-is-ratification core stands, and D3 ("history stands") is *reinforced* by point 6.
  - `decision-0037` — D2 (trellis's `draft → ratified` default) entirely, and within D1 exactly one
    bullet: *"an undeclared status is a conformance failure"*, which would now fail every trellis
    artifact. **D1's actual principle — the enum is methodology-defined — stands, and this record is
    its strongest instance:** a methodology declaring *no* enum rather than a different one. D3's
    `owner:` mapping untouched.
  - `decision-0022` — D2 entirely (the in-PR flip convention and the no-draft-on-main clause) **and
    D3** (who proposes the flip), plus the Consequence answering `spec-0001`'s *"two consumable
    states or one?"* with *keep two*, which this record reverses to *keep none*. **D1 stands and is
    vindicated:** the ratified *state* is core, the workflow expressing it is instance-specific —
    trellis changed the workflow and kept the state.
  - `decision-0046` — D1–D4, the flip-records-the-act machinery and the kept draft-landing guard.
    **D2's substance is kept by point 5, not lost:** an agent still may not manufacture acceptance
    without a human act. **D5 survives** as general agent behavior.
  - `decision-0040` — D5's *"keeps `status: ratified`"* clause only; partial supersession as a
    concept and the field it introduced both stand.
- **Contract surface edited in this change:** `core/rubrics/artifact-contract.md` (checks 1, 2,
  5, 7), `.claude/agents/corpus-reviewer.md` (same four), `AGENTS.md`, and
  `core/fixtures/known-bad.md` + its README — the fixture's check-2 violation is an invalid status,
  which stops being a violation class, so the fixture is re-cut rather than left asserting a check
  that no longer exists. *(This record was authored against `spec-0001` §1/§2/§3 and
  `.grove/config.toml`; both retired underneath it — `decision-0079` deleted `specs/` and
  `decision-0076` retired grove — so the same edits land on the surviving homes.)*
- **The typed-artifact schema's `draft → ratified` profile/catalog lifecycle**
  (`core/schemas/typed-artifacts.md` §3, formerly `spec-0002` §3) is product-layer, not
  trellis-self, and is **left standing in this change.** It describes the Assess→Apply gate for a
  consumer's profile, which is a different object from this repo's own decisions. Flagged rather
  than swept: if it should follow, that is a separate product decision.
- **`inv-ratifiable-artifacts` is honored, not violated — and this is the load-bearing
  reconciliation.** Its signature names *"a `status` lifecycle"*, but `decision-0037` makes statuses
  methodology-defined, and the invariant's substance is that **upstream can reach a ratifiable state
  downstream consumes**. Here that state is *merged on `main`*, reached by a human act, and
  consumable. The lifecycle moved from a frontmatter field to VCS state; it did not disappear.
  **This is a real instance of `invariants-v1`'s own open question** — *"Does
  `inv-ratifiable-artifacts` over-constrain … methodologies that have no explicit 'approved'
  state?"* — and the answer this change supplies is: the invariant is fine, its *signature* is
  written too concretely. **Recorded at that open question in
  `core/invariants/trellis-invariants-v1.md` in this change.** The catalog's `signature:` line is
  **deliberately not edited** — it rewrites the shipped payload, and that is not what this change is
  about. *(An earlier draft of this record claimed the evidence was recorded when it had not been;
  the independent review caught the unfulfilled claim and the edit was then actually made.)*
- **`inv-directional-flow` still has teeth**, but they are the merge rather than a string: an
  unmerged artifact is not consumable. What is genuinely lost is the ability to land a
  known-unsettled artifact on `main` and have tooling refuse to consume it. **That capability is
  given up deliberately** — the corpus shows it was never used.
- **Not consumer-visible; no payload change, no `VERSION` bump, no row-set delta.**
- **The two shipped `signature:` lines that mention `status`** (`inv-directional-flow`,
  `inv-ratifiable-artifacts`, in the catalog and its renders) describe the tell *a consumer's own
  project* exhibits, not a trellis-imposed field. **Left untouched** — editing them would change the
  payload hash for no behavioral reason. Named so the omission is deliberate.
- **`README.md:316` and the landing copy** claim a `draft → ratified` lifecycle for trellis
  artifacts. Corrected in this change where it describes *this repo*; the brief
  (`agentic-dev-meta-layer-brief.md`) is historical and left alone.
- **Eval fixtures under `eval/experiments/.../05-build-against-draft/`** deliberately carry
  `status: draft` to test whether an agent notices *another* project's draft marker. **Left as-is**
  — they simulate a third-party project, not trellis's corpus.

## Open questions

- **Does the family follow?** `kodhama-0004-uniform-lifecycle` still defines the enum for five other
  repos. Trellis is now the outlier. Filed as its own item rather than decided here.
- **Does a revise-in-place artifact need an unsettled marker after all?** Specs are revised in
  place; a spec being reworked across several PRs has no way to say "in flux" now. The bet is that
  the PR carries it. If that bites, the honest fix is a narrow marker for specs, not the ladder back.
- **Should `spec-0002`'s profile lifecycle follow?** See Consequences.

## Self-check (gate)

The change is applied to itself: this record carries no `status`. Every decision whose clauses it
outgrows gets a forward marking rather than an edit (`inv-auditable-archive`), and the 34 intent-act
comments are preserved rather than tidied away — the point of retiring a field is not to erase what
it recorded. The family divergence is named and filed rather than resolved in prose
(`inv-deliberate-succession`). The one invariant this plausibly violates is confronted directly
rather than left for a reviewer to find.

**Author is an agent; the maintainer directed the change and has not reviewed this record.** The
judgment calls he did not make explicitly — retiring the field going forward rather than stripping
99 files, keeping `spec-0002`'s product lifecycle, diverging trellis-only instead of changing family
law — are recorded above as calls, and each is reversible.

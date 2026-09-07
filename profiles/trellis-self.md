---
id: profile-trellis-self
type: expression-profile
status: ratified
depends_on: [signature-catalog-v1, invariants-v1]
owner: gundi
scope: core-methodology
ratified: 2026-07-04
---

# Expression profile — Trellis-self (the self-hosting instance)

> **Assessment dated 2026-07-04; the repairs named below dated 2026-09-07.** Read each verdict as of
> the first date. As of the second: every `AGENTS.md` quotation in this file, and every pointer in
> the rows repaired below, was opened against its source. The pre-migration GitHub citations (`#28`,
> `#165`, `#166`) were **not** re-audited. The verdicts are not re-run here; a *pointer* is repaired
> in place when what it names is retired or has moved, and where no pointer repair is possible the
> evidence is re-argued from durable material instead. Either way the repair names the change that
> broke the pointer.
>
> `decision-0079` retired the spec stage and deleted `specs/`. The `inv-directional-flow` evidence
> was re-staged at `decision-0082`'s change (`310894c`), and the surviving `spec-0002`
> citations resolve through `decision-0079`'s retired-artifacts registry.
>
> The `conformance-reviewer` this profile cited in four places went in the two steps
> `core/README.md` records: the vendored `.claude/agents/` copy was dropped at `ac63e7c`, and the
> role ran on as the plugin-carried `grove:conformance-reviewer` until `decision-0076` retired
> grove. All four
> citations — one in Delivery, three in the rows — are repaired here. The `independent-agent`
> gatekeeper column, which at ratification named that repo-authored agent, is filled by the
> repo-owned `corpus-reviewer` and `decision-0007`'s PR-review workflow.
>
> **A third staleness reached three more rows, and is repaired here too.** They quoted two bullets
> that have since left this repo's shared instructions. *"Loud failure"* went from the Claude adapter
> at `61b6618` (`decision-0035`, 2026-07-06). *"Gates"* — quoted by two rows — did **not**: that same
> commit says the specific-method bullets *"stay"*, `AGENTS.md` still carried Gates when this file's
> citations were re-pointed to it at `31bb938`, and the bullet left at `c75eace`, `decision-0079`'s
> own change. So `inv-handover-points`, `floor-transparency` and `floor-intent-gate` are re-pointed
> at live `AGENTS.md` text, or re-argued from durable material where the live text does not carry
> the claim. `AGENTS.md` carries no "Gates" section and no "Loud failure" today.
>
> *An earlier form of this note said the assessment was "kept unedited". It was not, and had not been
> when that was written: `git log -- profiles/trellis-self.md` shows rows rewritten and two added
> (`inv-deliberate-succession`, `inv-no-orphan-followups`) after 2026-07-04. Withdrawn rather than
> quietly deleted — a snapshot that claims to be frozen while it is edited is the same defect this
> change repairs.*

> **Ratified via merge (`decision-0022`).** This is an **assessment** the agent produced; the
> maintainer's **merge of this PR is the ratification** (`floor-intent-gate`) — the producer proposes, the merge accepts.
> It is the **first worked instance** of the
> `expression-profile` schema (`spec-0002`), authored by hand (Assess does not exist yet, cluster 1).

> **Honest discount (load-bearing, not hidden).** Trellis-self is the **reference organism** — the
> repo is *built to* honor its own invariants (`AGENTS.md`: "We build Trellis with Trellis"). That it
> expresses the full genome is therefore **expected, and is not independent validation** of the
> invariants — it is genealogically N=1 (`decision-0009`, `research-0006` §Result 5 discount). A
> *different* project's profile (RPI, the consultant-mode work usage, Math Quest) is what would test
> generalization. Read this as a schema demonstration + a self-audit, not evidence the set travels.

## Delivery

- **delivery_relationship:** `supervisor` — the checks run *live* on this repo (CI review
  `decision-0007`; the repo-owned `corpus-reviewer` sub-agent, `.claude/agents/`, which took over
  the artifact-contract half of the `conformance-reviewer` named here at ratification, retired with
  grove by `decision-0076`), not as an external consult.
- **payload_depth:** `+mechanism` — the instance carries the full regulatory apparatus (it *is* the
  apparatus) and self-regulates (`decision-0009` improvement loop).
- **application_model:** `M2-morph` — the **degenerate self-hosting case**: the host's own
  methodology *is* Trellis's, so the invariants are integrated natively, not overlaid on a separate
  host. (A real external instance would default to `M1-overlay`, augment-never-clobber.)

## Profile

*All assessable genes are active and honored natively — unsurprising for the reference organism
(see the discount above; `inv-reference-relationship` collapsed into `floor-transparency`, `decision-0021`). Each
`evidence` points at a real artifact in this repo.*

| slug | active | C1 | C2 | basis | confidence | evidence |
|---|---|---|---|---|---|---|
| `inv-directional-flow` | true | enforced | independent-agent | honored-implicitly | verified | `research/ → decisions/` staging (`specs/` retired by `decision-0079`); the merge carries the flow since `decision-0082` (everything on `main` is settled), so the check is that every `depends_on` resolves in-corpus — conformance run confirms it does |
| `inv-handover-points` | true | enforced | independent-agent | honored-implicitly | verified | `AGENTS.md` states the handover live — *"one logical change per PR; descriptive, linear history; diffs small enough to review on a phone"* — and its `## Checks and review` section names what must run before a merge, which is what makes the PR a point rather than a continuum. **Re-pointed:** the *"Gates"* bullet quoted here left `AGENTS.md` at `c75eace` (`decision-0079`), having survived the `31bb938` re-point intact |
| `inv-intent-locus` | true | enforced | human | honored-implicitly | verified | `owner:` on every artifact; ratification is a recorded human act (this session) |
| `inv-ratifiable-artifacts` | true | enforced | independent-agent | honored-implicitly | verified | the ratifiable state is **merged on `main`**, reached by the maintainer's merge (`decision-0082` retired the status field; the lifecycle moved to VCS, it did not disappear); `core/rubrics/artifact-contract.md` carries `## Acceptance criteria` |
| `inv-graph-maintenance` | true | enforced | independent-agent | honored-implicitly | verified | `depends_on` graph; `invariants-v1` supersede registry; v0 retirement resolved this session |
| `inv-gate-at-handover` | true | default-on-but-skippable | independent-agent | honored-implicitly | verified | automated PR review (`decision-0007`) still runs — `.github/workflows/claude-code-review.yml`, on every `opened`/`synchronize` — and `AGENTS.md` requires the repo-owned `corpus-reviewer` *"before merging a change to `decisions/`, `research/` or `core/`"*. One fires mechanically; the other is agent-applied with no runtime (`decision-0010`), which is what `default-on-but-skippable` describes rather than a gap in it. Both resolve today. **Re-pointed:** the `conformance-reviewer` named here at ratification retired with grove (`decision-0076`), and between `decision-0079` and the delivery slice the handover is checked more narrowly than in 2026-07 — `core/README.md` records conformance of code to its authorizing decision as *"currently uncovered"* until `decision-0012` packages an agent for it. That a gate fires at the handover is what this row claims, and it does |
| `inv-independent-judgment` | true | default-on-but-skippable | independent-agent | honored-implicitly | verified | the repo-owned `corpus-reviewer` is read-only **by construction** — `tools: Read, Grep, Glob` in its frontmatter, no write tool — and independent of the producer by charter: *"Derive your checklist yourself … Do not accept a checklist from whoever produced the artifacts."* That it *operates*, and is not merely chartered, is recorded where an author had every reason to smooth it away: `decision-0076` counts its own defects at *"None of the sixteen was self-caught — three from the corpus review, nine from the diff review, four from the automated one. The author caught two arithmetic errors and no substantive defect."* The automated one is `decision-0007`'s, so no single seat carries the claim, and that workflow still writes records no author wrote — PR #298 carries seven `claude[bot]` review comments, each naming a defect at a file it read. **The intent face** — the catalog gives this gene two faces, *"the builder does not grade itself … the agent does not flatter the human"* — is shown in this table rather than asserted: two of the rows below sit at `inferred` with their counter-instances named, against the reference organism's own interest, and `decision-0076` keeps its defect count *"in rather than being smoothed away"*. **Re-argued, not re-pointed:** the original evidence named `conformance-reviewer` (retired with grove, `decision-0076`) and rested on *"ran independently this session"* — a claim that could not be re-pointed, because `corpus-reviewer` was not created until 2026-07-08, four days after this snapshot |
| `inv-auditable-archive` | true | enforced | independent-agent | honored-implicitly | verified | `decisions/` append-only; `decision-0014` splits current-truth from change-history |
| `inv-bounded-context` | true | default-on-but-skippable | independent-agent | honored-implicitly | verified | the one repo-owned sub-agent, `corpus-reviewer`, is scoped to declared inputs and shows both sides of that scoping in one readable file: a **Default corpus** paragraph naming the paths it may read and the one it must exclude (`core/fixtures/`), and a three-tool allowlist in its frontmatter, `tools: Read, Grep, Glob`. It holds the **artifact-contract** half of the role the retired `conformance-reviewer` held (grove's retirement is `decision-0076`); `core/README.md` records both that succession and that the other half has no home in this repo right now |
| `inv-self-improvement` | true | default-on-but-skippable | human | honored-implicitly | verified | `decision-0018` restored it after friction (the merge into `inv-graph-maintenance` lost "evolve"); the conformance check caught *this row's own absence* and it was added in the same change |
| `inv-deliberate-succession` | true | default-on-but-skippable | human | honored-implicitly | inferred | PR #165 is a real forward instance (the retrofit question surfaced and ruled on) and math-quest's phase-1 architecture the backward one (`#166`). **Deliberately not `verified`:** the entry is `*provisional*` in the set, and the change that minted it failed this rule four times — three count sweeps that each matched only some of the shapes a succession leaves behind, and the `superseded_in_part_by` mark omitted on `decision-0052` — all caught by independent review, not by the author. The repo holds this one with help, not natively |
| `inv-no-orphan-followups` | true | default-on-but-skippable | human | honored-implicitly | inferred | `AGENTS.md` states the address test outright — *"Ideas are a document, not issues — the team's Ideas doc, each entry carrying the trigger that would promote it. An idea filed as an issue is a to-do nobody agreed to"* — and every artifact's `## Open questions` rides a consumer that must read it. `decision-0074` deferred the curl-upgrade false all-clear and gave it a real address (trellis#241 → TRL-2, which the backlog carried and closed on 2026-09-02): the honored *(process)* shape. **Deliberately not `verified`:** the entry is `*provisional*`, and the repo held two counter-instances, one still live — the catalog's own open question *"Owed to the Assess build (cluster 1)"* names a consumer that does not exist yet, which is the live one; the payload→VERSION guard was a designed consumer never switched on for thirteen days, so `decision-0078`'s own release obligation had to be discharged by hand, until trellis#245 merged `release-guard.yml` on 2026-09-05. One live orphan still holds this row below `verified` |
| `inv-minimal-first` | true | expressed | human | honored-implicitly | verified | `AGENTS.md`: "a deliberately tiny instance of the seed operating method" |
| `inv-clarify-before-commit` | true | default-on-but-skippable | human | honored-implicitly | verified | `## Open questions` in every artifact; the delivery-axis + dial-coverage frictions were surfaced, not guessed |
| `floor-transparency` | true | enforced | human | honored-implicitly | verified | the floor is named by slug in both honesty clauses that gate this corpus, and both resolve today: `core/rubrics/artifact-contract.md` — *"Missing/unparseable input → halt loudly (`floor-transparency`), never a partial pass"* — and the `corpus-reviewer` charter, *"Accurately listing the violations **is** success. Never hide drift to report PASS."* Practised, not merely written: `decision-0076` keeps its own retractions in the record rather than deleting them, and this profile's *"kept unedited"* claim is withdrawn in place above rather than quietly removed. **Re-argued:** the *"Loud failure"* bullet quoted here left the Claude adapter at `61b6618` (`decision-0035`), and what remained was a session anecdote true only when written |
| `floor-intent-gate` | true | enforced | human | honored-implicitly | verified | `AGENTS.md` states the gate live — *"An agent still may not merge on his behalf without his act (`floor-intent-gate`) — the gate did not move; only the bookkeeping around it went away"* — and this profile is ratified by the maintainer's merge (`decision-0022`): the gate exercised, not asserted. **Re-pointed:** the *"Gates: Human approval at the intent layer"* bullet quoted here left `AGENTS.md` at `c75eace` (`decision-0079`) |

*(The two dials are not rows here — they are the `C1`/`C2` columns above (the schema field names,
`spec-0002`). Catalog excludes them by
design, `signature-catalog-v1`.)*

## Assessment notes

- **Confidence is `verified` on every row but two** because each tell is a real, citable artifact in this
  repo — not because the invariants are proven in general. The evidence is strong *for this instance*;
  the N=1 caveat above governs any wider claim. **Both exceptions are the newly minted, `*provisional*`
  entries.** `inv-deliberate-succession` (`inferred`, `decision-0074`): the change that minted it
  violated it four times before independent review caught them. `inv-no-orphan-followups` (`inferred`,
  `decision-0078`): the repo states the address test in `AGENTS.md` but holds a live orphan,
  cited in the row. A rule the reference organism needs help to hold is not one it honors natively,
  and saying otherwise would be the sycophancy the floors forbid.
- **The behavioral genes are the hardest to evidence.** `inv-clarify-before-commit` still rides its
  2026-07-04 grounding — `## Open questions` in every artifact, plus frictions surfaced rather than
  guessed in the authoring session — which is the strongest honest evidence short of a longitudinal
  audit, and is dated accordingly. **`inv-independent-judgment` and `floor-transparency` no longer
  ride it:** a session instance is true only at the moment of writing, so both rows now cite durable
  material a reader can open today — the reviewer's construction and records of it catching what the
  author missed, and the two honesty clauses that name the transparency floor by slug.
- **`floor-intent-gate` is the live demonstration:** the producer proposed this profile and the
  **maintainer's merge ratifies it** (`decision-0022`, merge=ratify) — the gate is exercised, not
  asserted.
- **This profile is the seed for the cross-instance diff (#28):** entry #1 in the eventual N=1→N
  table. Its value is as a *baseline to diff against*, not as corroboration.

## Open questions

- **Is self-hosting `M2-morph` or a category of its own?** The host = the product, so overlay-vs-morph
  may not apply cleanly to the reference instance. Revisit when a real external `M1` profile exists.
- **Do any genes deserve `C1: enforced` that are only `default-on-but-skippable` here** (`inv-gate-at-handover`, `inv-independent-judgment`, `inv-bounded-context`)?
  This instance runs them near-strict; a lighter instance would dial down — which is the point of the
  profile. The right defaults are the catalog's open question, not this profile's.
- **When Assess is built (cluster 1), does it reproduce this hand-authored profile** from the same
  evidence? That round-trip (`spec-0002` AC7) is the test that Assess works — this profile is its
  target output.

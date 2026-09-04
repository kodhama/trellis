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

> **Snapshot, 2026-07-04 — read as dated.** This is the assessment as it stood that day, kept
> unedited. Two later changes are *not* reflected in the rows below: `decision-0079` retired the
> spec stage and deleted `specs/` (so the `inv-directional-flow` evidence "`research/ →
> decisions/ → specs/` staging" names a stage that is gone, and the `spec-0001`/`spec-0002`
> citations point at retired records preserved in git history), and `decision-0076` retired
> grove (so the `independent-agent` gatekeeper column names roles no longer installed here). The
> assessment is not re-run; it is marked.

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
  `decision-0007`; the `conformance-reviewer` sub-agent), not as an external consult.
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
| `inv-handover-points` | true | enforced | independent-agent | honored-implicitly | verified | one-change-per-PR; `AGENTS.md` "Gates" (intent approval + execution verification) |
| `inv-intent-locus` | true | enforced | human | honored-implicitly | verified | `owner:` on every artifact; ratification is a recorded human act (this session) |
| `inv-ratifiable-artifacts` | true | enforced | independent-agent | honored-implicitly | verified | the ratifiable state is **merged on `main`**, reached by the maintainer's merge (`decision-0082` retired the status field; the lifecycle moved to VCS, it did not disappear); `core/rubrics/artifact-contract.md` carries `## Acceptance criteria` |
| `inv-graph-maintenance` | true | enforced | independent-agent | honored-implicitly | verified | `depends_on` graph; `invariants-v1` supersede registry; v0 retirement resolved this session |
| `inv-gate-at-handover` | true | default-on-but-skippable | independent-agent | honored-implicitly | verified | automated PR review (`decision-0007`) + `conformance-reviewer` fire at the PR handover |
| `inv-independent-judgment` | true | default-on-but-skippable | independent-agent | honored-implicitly | verified | `conformance-reviewer` is read-only + distinct from producer; ran independently this session |
| `inv-auditable-archive` | true | enforced | independent-agent | honored-implicitly | verified | `decisions/` append-only; `decision-0014` splits current-truth from change-history |
| `inv-bounded-context` | true | default-on-but-skippable | independent-agent | honored-implicitly | verified | sub-agents scoped to declared inputs (conformance-reviewer corpus; narrow tool sets) |
| `inv-self-improvement` | true | default-on-but-skippable | human | honored-implicitly | verified | `decision-0018` restored it after friction (the merge into `inv-graph-maintenance` lost "evolve"); the conformance check caught *this row's own absence* and it was added in the same change |
| `inv-deliberate-succession` | true | default-on-but-skippable | human | honored-implicitly | inferred | PR #165 is a real forward instance (the retrofit question surfaced and ruled on) and math-quest's phase-1 architecture the backward one (`#166`). **Deliberately not `verified`:** the entry is `*provisional*` in the set, and the change that minted it failed this rule four times — three count sweeps that each matched only some of the shapes a succession leaves behind, and the `superseded_in_part_by` mark omitted on `decision-0052` — all caught by independent review, not by the author. The repo holds this one with help, not natively |
| `inv-no-orphan-followups` | true | default-on-but-skippable | human | honored-implicitly | inferred | `AGENTS.md` states the address test outright — *"Ideas are a document, not issues — one long-form Linear doc, each entry carrying the trigger that would promote it. An idea filed as an issue is a to-do nobody agreed to"* — and every artifact's `## Open questions` rides a consumer that must read it. `decision-0074` deferred the curl-upgrade false all-clear and gave it a real address (trellis#241 → TRL-2, live in the Linear backlog): the honored *(process)* shape. **Deliberately not `verified`:** the entry is `*provisional*`, and the repo holds two live counter-instances — the catalog's own open question *"Owed to the Assess build (cluster 1)"* names a consumer that does not exist yet, and the payload→VERSION guard (trellis#245, open since 2026-08-23) is a designed consumer never switched on, so `decision-0078`'s own release obligation had to be discharged by hand |
| `inv-minimal-first` | true | expressed | human | honored-implicitly | verified | `AGENTS.md`: "a deliberately tiny instance of the seed operating method" |
| `inv-clarify-before-commit` | true | default-on-but-skippable | human | honored-implicitly | verified | `## Open questions` in every artifact; the delivery-axis + dial-coverage frictions were surfaced, not guessed |
| `floor-transparency` | true | enforced | human | honored-implicitly | verified | `AGENTS.md` "Loud failure"; this session surfaced the merge conflict + catalog friction rather than papering over |
| `floor-intent-gate` | true | enforced | human | honored-implicitly | verified | `AGENTS.md` "Gates: Human approval at the intent layer"; this profile is ratified by the maintainer's merge — the intent gate, exercised (`decision-0022`) |

*(The two dials are not rows here — they are the `C1`/`C2` columns above (the schema field names,
`spec-0002`). Catalog excludes them by
design, `signature-catalog-v1`.)*

## Assessment notes

- **Confidence is `verified` on every row but two** because each tell is a real, citable artifact in this
  repo — not because the invariants are proven in general. The evidence is strong *for this instance*;
  the N=1 caveat above governs any wider claim. **Both exceptions are the newly minted, `*provisional*`
  entries.** `inv-deliberate-succession` (`inferred`, `decision-0074`): the change that minted it
  violated it four times before independent review caught them. `inv-no-orphan-followups` (`inferred`,
  `decision-0078`): the repo states the address test in `AGENTS.md` but holds two live orphans, both
  cited in the row. A rule the reference organism needs help to hold is not one it honors natively,
  and saying otherwise would be the sycophancy the floors forbid.
- **The behavioral genes** (`inv-independent-judgment` intent face, `inv-clarify-before-commit`,
  `floor-transparency`) are the hardest to evidence — I ground them in the `AGENTS.md` rule **plus a
  demonstrated instance from this very session** (surfacing frictions, and this profile authored as a
  proposal for the maintainer's gate), which is the strongest honest evidence short of a longitudinal
  audit.
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

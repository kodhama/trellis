---
id: profile-trellis-self
type: expression-profile
status: draft
depends_on: [signature-catalog-v1, invariants-v1]
owner: gundi
scope: core-methodology
---

# Expression profile — Trellis-self (the self-hosting instance)

> **Status `draft` — awaiting your ratification (D2).** This is an **assessment**, and I (the agent)
> produced it. Per `inv-independent-judgment` / D2 the producer does not ratify its own assessment —
> it is `draft` pending the maintainer's gate. It is the **first worked instance** of the
> `expression-profile` schema (`spec-0002`), authored by hand (Assess does not exist yet, cluster 1).

> **Honest discount (load-bearing, not hidden).** Trellis-self is the **reference organism** — the
> repo is *built to* honor its own invariants (`CLAUDE.md`: "We build Trellis with Trellis"). That it
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

*All 15 assessable genes are active and honored natively — unsurprising for the reference organism
(see the discount above). Each `evidence` points at a real artifact in this repo.*

| slug | active | C1 | C2 | basis | confidence | evidence |
|---|---|---|---|---|---|---|
| `inv-directional-flow` | true | enforced | independent-agent | honored-implicitly | verified | `research/ → decisions/ → specs/` staging; conformance run this session confirms no ratified→draft edge |
| `inv-handover-points` | true | enforced | independent-agent | honored-implicitly | verified | one-change-per-PR; `CLAUDE.md` "Gates" (intent approval + execution verification) |
| `inv-intent-locus` | true | enforced | human | honored-implicitly | verified | `owner:` on every artifact; ratification is a recorded human act (this session) |
| `inv-ratifiable-artifacts` | true | enforced | independent-agent | honored-implicitly | verified | `status: draft→ratified` lifecycle; `spec-0001/0002` carry `## Acceptance criteria` |
| `inv-graph-maintenance` | true | enforced | independent-agent | honored-implicitly | verified | `depends_on` graph; `invariants-v1` supersede registry; v0 retirement resolved this session |
| `inv-gate-at-handover` | true | default-on-but-skippable | independent-agent | honored-implicitly | verified | automated PR review (`decision-0007`) + `conformance-reviewer` fire at the PR handover |
| `inv-independent-judgment` | true | default-on-but-skippable | independent-agent | honored-implicitly | verified | `conformance-reviewer` is read-only + distinct from producer; ran independently this session |
| `inv-auditable-archive` | true | enforced | independent-agent | honored-implicitly | verified | `decisions/` append-only; `decision-0014` splits current-truth from change-history |
| `inv-bounded-context` | true | default-on-but-skippable | independent-agent | honored-implicitly | verified | sub-agents scoped to declared inputs (conformance-reviewer corpus; narrow tool sets) |
| `inv-self-improvement` | true | default-on-but-skippable | human | honored-implicitly | verified | `decision-0018` restored it after friction (the B6→B1 merge lost "evolve"); the conformance check caught *this row's own absence* and it was added in the same change |
| `inv-minimal-first` | true | expressed | human | honored-implicitly | verified | `CLAUDE.md`: "a deliberately tiny instance of the seed operating method" |
| `inv-reference-relationship` | true | default-on-but-skippable | human | honored-implicitly | verified | `decision-0002` adopt/adapt as a dial; framework analysis recorded (`research-0002`) |
| `inv-clarify-before-commit` | true | default-on-but-skippable | human | honored-implicitly | verified | `## Open questions` in every artifact; the delivery-axis + dial-coverage frictions were surfaced, not guessed |
| `floor-transparency` | true | enforced | human | honored-implicitly | verified | `CLAUDE.md` "Loud failure"; this session surfaced the merge conflict + catalog friction rather than papering over |
| `floor-intent-gate` | true | enforced | human | honored-implicitly | verified | `CLAUDE.md` "Gates: Human approval at the intent layer"; this profile is left `draft` for exactly that gate |

*(The two C dials are not rows here — they are the `C1`/`C2` columns above. Catalog excludes them by
design, `signature-catalog-v1`.)*

## Assessment notes

- **Confidence is `verified` across the board** because each tell is a real, citable artifact in this
  repo — not because the invariants are proven in general. The evidence is strong *for this instance*;
  the N=1 caveat above governs any wider claim.
- **The behavioral genes** (`inv-independent-judgment` intent face, `inv-clarify-before-commit`,
  `floor-transparency`) are the hardest to evidence — I ground them in the `CLAUDE.md` rule **plus a
  demonstrated instance from this very session** (surfacing frictions, leaving this profile draft),
  which is the strongest honest evidence available short of a longitudinal audit.
- **`floor-intent-gate` is the live demonstration:** this profile sits `draft` precisely so a human
  ratifies it — the gate is exercised, not asserted.
- **This profile is the seed for the cross-instance diff (#28):** entry #1 in the eventual N=1→N
  table. Its value is as a *baseline to diff against*, not as corroboration.

## Open questions

- **Is self-hosting `M2-morph` or a category of its own?** The host = the product, so overlay-vs-morph
  may not apply cleanly to the reference instance. Revisit when a real external `M1` profile exists.
- **Do any genes deserve `C1: enforced` that are only `default-on-but-skippable` here** (B2, B3, B5)?
  This instance runs them near-strict; a lighter instance would dial down — which is the point of the
  profile. The right defaults are the catalog's open question, not this profile's.
- **When Assess is built (cluster 1), does it reproduce this hand-authored profile** from the same
  evidence? That round-trip (`spec-0002` AC7) is the test that Assess works — this profile is its
  target output.

---
id: research-0006
type: research-note
status: draft
depends_on: [invariants-v1, decision-0002, decision-0008, decision-0009, decision-0012, research-0005]
owner: gundi
---

# Research 0006 — Trellis as a supervisor (the Discrete Event Systems / supervisory-control lens)

> **Method & honesty (load-bearing).** Same discipline as [[research-0005]]: **what the theory
> says** (externally sourced, tagged) is kept separate from **the mapping onto Trellis** (my
> inference, tagged). The DES/supervisory-control theory is standard and well-sourced (primary
> Ramadge–Wonham papers + an open-access SCT+RL paper, verbatim). **Caveat I will not hide:** I
> did **not** read Cassandras & Lafortune's textbook directly (book, not fetchable) — statements
> are grounded in the primary RW papers and the open literature the maintainer can cross-check.
> The maintainer's own PhD thesis — **Neto 2010** (IST/UTL; advisor P. Lima), *Planning, Learning
> and Control Under Uncertainty Based on Discrete Event Systems and Reinforcement Learning* — **has
> now been read**, and §6 folds in its actual stance; it bears on Trellis more than the "fun
> reference" it was offered as (see §6). This lens is
> stronger than the genetics one on the load-bearing floors (esp. D2); it is the primary frame.

## Why this lens

Trellis's own `CLAUDE.md` already calls it "a layer that **supervises** an agentic software-
development process" and "the layer above the steps." Ramadge–Wonham **supervisory control theory
(SCT)** is the formal theory of *exactly that* — a supervisor that constrains a system's behavior
to a specification by enabling/disabling events, while leaving maximal freedom otherwise. The fit
is close enough that several invariants turn out to have a control-theoretic *rationale* in the
frame, not just a restatement — most importantly D2 (with the honest caveat in §Limits that the D2
derivation rests on an analogical modeling step, not a proof).

## The core mapping

| SCT (Ramadge–Wonham) | Trellis | Tag |
|---|---|---|
| **Plant / generator G** — produces a formal language over an event alphabet Σ | The agentic dev **process**: all sequences of agent actions, artifact transitions, gate crossings | `inferred` |
| **Events Σ** (strings/traces) | Actions: promote `draft→ratified`, dispatch an agent, consume an upstream, cross a handover | `inferred` |
| **Specification K ⊆ L(G)** — the admissible sublanguage | The **invariants** as the set of *legal* process behaviors (illegal string: "downstream consumes a `draft`") | `inferred` |
| **Supervisor S** — maps the observed string to the set of **enabled controllable events** | **Trellis** — after each step, permits/blocks the next controllable actions | `inferred` |
| **Marked states / nonblocking** — desired complete behaviors remain reachable | A change reaches a **ratified/approved** terminal state; no dead ends | `inferred` |

## Result 1 — D2 has a control-theoretic *rationale*, not just a stipulation (the strongest payoff)

SCT partitions events into **controllable** (Σc — the supervisor can disable) and
**uncontrollable** (Σuc — "the supervisor **cannot** prevent [them] from occurring"). `verified`
(PMC11500187, verbatim). A specification is only *enforceable* if it never requires disabling an
uncontrollable event — the **controllability condition**: *"if s is a prefix of a string in L(H),
and e is an uncontrollable event physically possible after s, then se must also be a prefix of a
string in L(H)."* `verified` (verbatim).

- **Mapping (`inferred`, but tight):** "the upstream is itself **wrong**" (a ratified intent that
  shouldn't have been) is an **uncontrollable event** — no automated supervisor can *disable*
  wrongness, because catching it requires judgment outside the controller's authority. By the
  controllability condition, a specification that must catch it **cannot be enforced by the
  supervisor alone** — it must be routed to a gatekeeper who *can* act on it: the human. That is
  **D2 (`floor-intent-gate`, "the intent gate never fully opens")** — *motivated* by the frame, not
  merely stipulated. Precisely: **conditional on modeling "wrong intent" as an uncontrollable
  event**, the controllability condition *forces* D2 — so the load-bearing claim is that modeling
  choice (defended in §Limits), not a bare theorem. The reason C2 can never be `none` at the intent
  locus is then structural: *there is a class of faults the automated supervisor cannot disable.*
- This is the exact gap the genetics lens **could not** cover ([[research-0005]] §Limits). The two
  lenses are complementary: genetics explains *partial expression*; SCT explains *the floor*.

## Result 2 — bounded context is partial observation (grounds B5)

SCT under **partial observation**: the supervisor sees only a **natural projection** of the
plant's events (some events unobservable), which strictly weakens what it can enforce
(observability condition). `verified` (standard RW result).

- **Mapping (`inferred`):** **B5 (`inv-bounded-context`)** — "each operation reads only its
  declared inputs, never the whole archive" — *is* a partial-observation supervisor by
  construction. Consequence worth surfacing: SCT proves partial observation **weakens control**,
  so bounded context is a real *cost*, not a free scaling win. Trellis pays it deliberately (scale),
  and the theory says exactly what it buys and what it forfeits. **The thesis sharpens this:** Neto
2010's technical core *is* control under partial observability — it shows you must build an explicit
**observer** (a state-estimator over the declared inputs) for the controller to decide or learn at
all, and derives when learning still converges under it. Read into Trellis: a bounded-context
operation acts on a *projection* of project state, so its decision quality rides on an explicit
observer — the `depends_on` graph + provenance *is* that estimator; under-specify it and the
operation decides on a bad state estimate. `inferred`.

## Result 3 — the maximally-permissive supervisor grounds C1 *defaults* (not, cleanly, B7/D1)

For any spec K there is a unique **supremal controllable sublanguage** — the *largest* controllable
behavior inside K — realized by the **maximally permissive** supervisor: *"minimum restriction and
maximum admissible control."* `verified` (SIAM 10.1137/0325036; PMC11500187, verbatim).

- **Mapping (`inferred`) — sharpened per Fable review.** The earlier "minimal-first = D1 =
  maximally permissive supervisor" fused **three distinct minimalities**; separate them: (i) **B7
  minimal-first** keeps *K itself* small (fewer invariants/steps) — a choice about the *spec*, not
  the supervisor; (ii) the **maximally-permissive supervisor** is the least-restrictive *enforcement
  of a fixed K* (the supremal controllable sublanguage); (iii) **D1 "surface, don't enforce"**
  disables *nothing even when K would be violated* — that is the **null** supervisor, *more*
  permissive than (ii), not the same object. The spirit ("restrict as little as possible") is shared;
  the technical identity is not. **The real payoff sits on (ii) and grounds C1 defaults, not B7/D1:**
  given the invariants as K, the *default* enforcement Trellis should ship is the **supremal
  controllable sublanguage** — computably the least Trellis *must* block. See the promoted Proposal
  below.

## Proposal — compute the default gate-set (promoted from an open question, Fable review)

The single most concrete, iron-rule-friendly output of this lens: **formalize the *mechanizable*
fragment of the invariants — A1 (directional flow), A4 (ratifiable artifacts), B1-flow (no ratified
consumes a draft), B2 (a gate at each handover) — as a specification K over the artifact-event
alphabet, and *compute* Trellis's default gate-set as the supremal controllable sublanguage of K.**
That turns "what does Trellis block by default" from a per-instance judgment into a **computed,
checkable artifact** — exactly what the iron rule demands — and it plugs straight into the conformance
sub-agent (`spec-0001`): the sub-agent checks conformance to a *computed* K, not a hand-written
checklist. Scope caveat (§Limits): this covers only the mechanizable fragment; the behavioral genes
(B3/B9/D1) stay outside K. **Owed to the spine's enforcement machinery** — candidate for its own
note/decision. `inferred` (design proposal).

## Result 4 — modular/decentralized supervision = per-gate gatekeepers (grounds C2, pillars)

SCT composes **modular/decentralized** supervisors — several running concurrently, each enforcing
one conjunct of the spec. `verified` (Modular SC, Wonham–Ramadge). **Mapping (`inferred`):** the
per-gate **C2 gatekeeper** choice (`independent-agent | human | none`) and the four-pillar split
are modular supervision — one supervisor per gate/concern, conjoined. Known SCT caveat that
transfers as a real warning: modular supervisors can be individually correct yet **jointly
blocking** (conflict) — i.e., independently-fine gates can deadlock a project. Trellis should expect
and check for gate-conflict, a failure mode the monolithic view hides.

## Result 5 — the RL frontier = the autonomy ratchet + self-improvement (your thesis area)

This is where the maintainer's thesis lives, and the published literature already builds the
structure Trellis needs. The pattern (multiple works): **synthesize a supervisor by SCT to bound
the safe envelope (disable unsafe controllable event sequences), convert the controlled automaton
to an MDP, then let RL optimize *within* that confined space.** `verified` (PMC11500187 / Sci.
Reports 2024, verbatim: SCT confines the action space to supervisor-permitted paths; Neurocomputing
2024; Yamasaki & Ushio 2005 for the decentralized/RL case). The stated motivation is exactly
Trellis's: **SCT alone cannot optimize an implicit objective** ("get to A fastest" — read: "ship
well"), **RL alone is unsafe/slow** — so you want the safety envelope *and* learned optimization
within it. The safety community's **"shielding"** is the same idea: a shield synthesized from a
safety spec either **preemptively** hands the agent "a list of safe actions" or **post-posed**
corrects an unsafe action. `verified` (Alshiekh 2018, arXiv:1708.08611).

- **Mapping (`inferred`):** the invariants are the **shield / supervisor** (the safe envelope);
  the process runs agents (RL-like) that optimize "good software" *within* it. The **autonomy
  ratchet** (invariant-4 note: "the boundary moves, ratcheting toward autonomy on a track record")
  is **policy improvement**: as evidence accrues, more of the envelope is safely opened — the
  dials (C) are the **policy parameters**, and `decision-0009`'s cross-instance loop is the
  learning signal (the "human overrides Trellis and is right" gold signal = a labeled reward). And
  Trellis's whole conversational UX ("the next step is X — and you may skip it," `decision-0008`) is
  **preemptive shielding**: surface the safe-action set, let the agent choose.
- **What the thesis actually says (now read).** Neto 2010's cornerstone is *"supervisory control of
  discrete event systems to provide loosely specified planning options over which a reinforcement
  learning based controller can optimize"* — the supervisor specifies *"the behaviors the agent is
  allowed to have, planning constraints rather than fixed pre-programmed plans,"* RL optimizing
  *"within the bounds defined by the supervisor."* `verified` (thesis, verbatim). Two consequences:
  **(a)** its supervisor is **designer-given and fixed** (offline), learning only *inside* it —
  evidence *for* the D2 stance (humans own the constraints; agents optimize within), *against*
  "learn the invariants themselves." **(b)** it wants the constraints specified *naturally*, pointing
  at temporal-logic→automata compilation (Lacerda & Lima 2008) — a direct input to the delivery
  question (owed 2.4): human states intent in a natural form, it compiles to enforcement.
- **Honest discount (anti-sycophancy).** Neto and the Trellis maintainer are the **same person**, so
  this resonance is shared architectural taste, **not independent corroboration** — it is
  *genealogically* N=1, the very risk `decision-0009` flags for math-quest. Real as lineage and
  vocabulary; **not** evidence Trellis is right. The Trellis mapping stays `inferred`.

## Result 6 — supervise vs. modify the plant (a switch classical SCT forecloses)

Classical SCT (and Neto 2010) hold the **plant G fixed** — the supervisor only *disables events*, never
rewrites the system. But Trellis's "plant" is **soft** (the host's instructions are editable text), so a
move SCT forbids is *available*: modify the plant itself. Staying a pure supervisor is a **designer's
choice**, not a necessity (maintainer, 2026-07). **Model 1 — supervise (default):** constrain from
outside, plant unchanged (= the epigenetic overlay, [[research-0005]]). **Model 2 — morph (deferred
option):** *compile the supervisor into a new plant* — ship the controlled behavior L(S/G) as the host's
own methodology, so it honors the invariants without Trellis present. Floors still bind M2 (surfaced D1 +
ratified D2). That both lenses **and** the maintainer's
own SCT framing land on this same switch is strong evidence the axis is real. `inferred`.

> The delivery model below (supervisor/consultant + the payload axis + how the dials are set) is
> **consolidated and extended in [[research-0007]] (deliverable 2.4)** — kept here as the thread's
> origin; 0007 is canonical.

**Refinement (maintainer, 2026-07) — it's really a *relationship* axis: supervisor vs. consultant.**
Overlay-vs-morph is the surface of a deeper delivery fork:
- **Supervisor** — Trellis *installed and live* (push): constrains at runtime, stays current via the
  delivery channel (`decision-0012`), **cleanly removable**.
- **Consultant** — Trellis *not installed* (pull): the host's own agents **consult** Trellis as an
  external reference and internalize what applies — no runtime dependency to review, secure, or remove;
  the effect is baked in (its kinship with M2).
This **reframes 2.4**: delivery is not only "which C1 activation level" (0012's push spectrum) but a
top-level **supervisor(push/installed) vs. consultant(pull/referenced)** fork — extending 0012's lowest
"available + referenced" rung into a genuine pull relationship where Trellis is never a host component.
**Datapoint (`verified`, user report; genealogically tainted):** the maintainer already runs consultant
mode at work — pointing opencode agents at the Trellis repo (via authenticated `gh`) to mine ideas for a
*company* project, **without** installing a dependency or clearing the compliance question an install
triggers. *Transferable insight:* **referencing a repo sits below the compliance/vendor-review radar an
installed dependency trips** — so consultant may be the lower-friction enterprise on-ramp (§7, new
angle). *Taint not hidden:* the frictionless-*trust* half is same-person; the *compliance-radar* half is
structural. One person, one datapoint — but a real N≥2-ish signal (`decision-0009`/#28; pattern
shareable, specifics consent-gated). **Caveats consultant mode must design for:** (a) iron rule — advice
must still yield *concrete host artifacts* or it's a philosophy deck (§7); (b) B3 — no runtime ⇒
independent verification isn't automatic; the advice must *install* a checker or invariant-5 lapses; (c)
staleness — a consultant result can't auto-update, so it forks from upstream Trellis as Trellis evolves.
Net: M2/consultant is **elevated from a deferred option to an open strategic fork with a datapoint** —
appetite unknown, but no longer clearly second. Owed to 2.4.

## Limits — where the analogy breaks (first-class)

- **The plant is not a fixed automaton.** RW assumes a *known, fixed* event alphabet and generator.
  An agentic process's events are **open and evolving** (new agents, new artifact types) — closer
  to SCT over an *unknown/parametric* plant, which is harder and less settled. Do not imply the
  clean finite-state guarantees transfer wholesale. `inferred` limit.
- **The invariants are not (yet) a regular language.** The mapping treats "no ratified-consumes-
  draft" as a language constraint; most invariants (B3 intent-face, B9 clarify) are **not
  mechanically checkable** (v1 says so). SCT's synthesis theorems apply only to the mechanically-
  specifiable fragment (roughly A1/A4/B1-flow/B2). The behavioral genes ([[research-0005]]) sit
  *outside* the formal frame — the two lenses partition the set.
- **"Wrongness" as an event is analogical.** D2's derivation (Result 1) treats wrong-intent as an
  uncontrollable event; it is really a *property of a state*, not a literal symbol in Σ. The
  structural conclusion holds (a fault class the controller can't disable); the formalization is a
  gloss, not a proof. `inferred`.
- **Nonblocking ≠ "good."** SCT guarantees a marked state is reachable, not that the product is
  worth building — the intent gate (human) still owns "worth it." The theory bounds *safety/
  liveness*, never *value*.

## Convergence with the genetics lens

Both lenses point at **one object**: the per-instance **control map / expression profile** —
"which invariants are active here, at what enforcement level, gatekept by whom." SCT calls it the
supervisor's control policy; genetics calls it the expression profile. That two independent
theories converge on the same missing artifact is the strongest evidence it is real — and it is
exactly the primitive four issues circle (#22/#23/#24/#28). The delivery question (how the map is
*set* — Assess proposes, human ratifies at D2, delivery layer consumes) is deferred to the
design/delivery decision (owed: 2.3/2.4).

## Sources

- Controllable/uncontrollable partition; supervisor disables controllable, cannot prevent
  uncontrollable; controllability condition; supremal controllable sublanguage = "minimum
  restriction and maximum admissible control"; SCT→MDP + RL-confined-to-permitted-paths; motivation
  (SCT can't optimize implicit specs, RL alone unsafe/slow): PMC11500187 (Sci. Reports 2024),
  verbatim. `verified`.
- Supremal controllable sublanguage existence/computation: SIAM J. Control Optim. 10.1137/0325036
  (Wonham & Ramadge 1987); RW 10.1137/0325013 (1987). `verified`.
- Safe RL via shielding (preemptive safe-action set / post-posed correction): arXiv:1708.08611
  (Alshiekh et al., AAAI 2018). `verified`.
- Decentralized SC via RL: Yamasaki & Ushio (2005); "Integrating RL and SCT for optimal directed
  control of DES," Neurocomputing 2024 (ScienceDirect PII S0925231224014917). `verified` (works
  exist and match the pattern; internal details `inferred`, not read in full).
- **Cassandras & Lafortune, *Introduction to Discrete Event Systems* — canonical reference named by
  the maintainer; NOT read directly** (but cited as the observer-construction source *within* Neto
  2010 — secondary grounding). `inferred` that the above is faithful to its treatment.
- G. Neto, *Planning, Learning and Control Under Uncertainty Based on Discrete Event Systems and
  Reinforcement Learning*, PhD thesis, Instituto Superior Técnico / UTL, Nov 2010 (advisor P. Lima).
  **Read** (abstract, §1.4 contributions, §3.3, ch. 7). `verified`.

## Open questions

- **Compute the default gate-set** — *promoted from this open question to the named Proposal above*
  (Fable review). Remaining sub-question: is the artifact-event alphabet finite/stable enough for the
  supremal-controllable-sublanguage computation to be tractable, given the "plant is open/evolving"
  limit (§Limits)? Owed to the spine's enforcement machinery.
- **Learned vs. ratified ratchet.** Neto 2010 sides with a *fixed, designer-given* supervisor and
  learning only *within* it (supports D2). Open for Trellis: does *any* constraint safely become
  learned/relaxed across instances (the ratchet), or is the floor strictly human-ratified? The
  thesis keeps the *safety envelope* human-given; any ratchet lives in what's optimized inside it.
- **Gate-conflict (Result 4).** Do Trellis's modular gates ever jointly block? Add a conflict check?
- **Does the control-map = expression-profile identity hold** ([[research-0005]]), making them one
  typed artifact — or are they two views that should stay separate? Decide with 2.3/2.4.

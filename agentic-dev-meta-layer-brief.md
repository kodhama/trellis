# Meta-process layer for agentic-assisted development — discovery brief

> **What this is.** A bootstrap context document to start a *separate* project. It
> captures an idea — a process-agnostic "meta layer" that guides, fits, teaches, and
> guards an agentic software-development process — and an honest assessment of whether
> it's worth building. Much of the thinking is *extracted from a real project*
> ("Math Quest", referenced throughout as the **source project**), which turns out to
> be an unintentional working prototype of the idea. Name for the thing:
> **Trellis** — keep a process *minimal, deliberately pruned, and shaped to fit, yet alive
> and adapting*: the art of trellis (start small, prune relentlessly, shape to the
> specimen, never let it sprawl). The self-improvement loop's "prefer retiring" is
> pruning; "start minimal, add on friction" is cultivation. Not a spec; a brief to
> argue with.
>
> **Bootstrapping this project cold?** Read **§10 (verdict)**, then **§11 (start here)**
> and **§12 (seed operating method)** — those three are enough to take the first real
> steps. §1–§9 are the depth behind them. §13 points at an optional reference.

---

## 1. The idea in one sentence

A pack for Claude (or any agentic surface) that is **not a development process**, but a
**supervisor/tutor for development processes** — you point it at a methodology (BMAD,
spec-kit, your own, or nothing) and it either *conducts* that process or *authors a
project-fitted one*, while always enforcing a small set of invariants, teaching the
team in situ, adapting as the project changes, and producing an audit trail.

## 2. The distinction that makes it interesting

Almost everything in this market sells a **process**:
- The "500 prompts to 10× your workflow" packs (unstructured, disposable).
- The structured frameworks — BMAD, spec-kit, etc. (opinionated, fixed pipelines).

Both sell *the steps*. The gap: **nobody sells the layer above the steps** — the thing
that decides *which* steps a given project needs, *teaches* them, *adapts* them as the
project and the tooling change, and *guarantees* the load-bearing properties hold no
matter which steps you picked. Call it the **meta-process**. That's the wedge.

Why the meta-process is the durable place to stand:
- **Fixed processes go stale and get adapted.** Orgs *will* modify any process you give
  them. Unmanaged, that adaptation loses rationale and drifts into chaos. A layer whose
  *entire job* is disciplined adaptation turns the inevitable into a feature.
- **You can't make people learn a moving target.** In a field changing this fast,
  betting on humans reading and retaining a fixed manual is a losing bet. A system that
  *onboards and guides in situ* — explaining the step you're in *as* you hit it — keeps
  working even as the steps change underneath.
- **Different companies at different stages need different processes.** A 4-person
  startup and a regulated bank need different gates. One fixed process fits neither
  well. A fitter does.

## 3. Why it might be real (and where the market pain is)

Enterprises adopting "AI-first development" have concerns the prompt-pack market
ignores entirely:
- *Can we trust AI-written code / decisions?* (verification, not vibes)
- *Can we prove the process was followed?* (audit, traceability, compliance)
- *Who is accountable for what an agent did?* (authority boundaries)
- *How do we adopt this without a 6-month retraining program?* (onboarding)
- *How do we keep it from rotting as models and tools change every quarter?* (adaptation)

A meta-layer built around **verification, provenance, authority, onboarding, and
adaptation** speaks directly to those — and is a *less crowded* position than "here is
my opinionated pipeline." That's the case for validity.

## 4. The foundation: are there real invariants? (yes — this is the crux)

The whole idea collapses if "agentic-assisted development" is pure churn with nothing
durable to encode. It isn't. The invariants below are properties of *reducing
uncertainty with semi-autonomous agents under human accountability* — they don't change
when the model or the tool changes. This list is the intellectual core of the product;
everything else is machinery around it.

1. **Directional handover.** Information flows one way through stages of *decreasing
   ambiguity*: research → decisions → contracts → implementation → validation.
   Downstream consumes only *ratified* upstream, never drafts. (The user named this; it
   is just the shape of uncertainty reduction, independent of stage names.)
2. **A gate at every handover.** Each stage transition has a verification gate. The
   gate's *nature* varies (human judgment vs automated check); its *existence* does not.
3. **Provenance / traceability.** Every artifact links to its upstream and its rationale
   (sources, confidence level, dependency edges). This is what makes the output
   auditable and is the answer to the enterprise "prove it" concern.
4. **Authority split: humans own intent, agents own conformance.** Humans are mandatory
   where *values/intent* are set; agents run autonomously where work merely *conforms*
   to an approved upstream. The *boundary moves* (it should ratchet toward autonomy on a
   track record), but the *principle* is invariant. Critically: the intent gate never
   fully opens — it's the only place an upstream that is itself *wrong* gets caught.
5. **Independent verification — "the builder does not grade itself."** Whatever produces
   work is never its sole judge. Some independent check (an agent, a panel, a human)
   verifies against the approved upstream with a checklist it derives itself.
6. **Bounded context.** Each operation reads only what it needs — its contract, the code,
   and the *live* decisions it cites — never the whole archive. A scaling necessity, not
   a preference.
7. **Decision history is immutable; current truth is consolidated.** Decisions append
   (you supersede, never edit, a ratified decision — it records why a thing was decided
   *then*); current truth lives in revised-in-place consolidated docs. Separating "why we
   decided" from "what is true now" keeps the archive navigable as it grows.
8. **Loud failure over silent degradation.** When a required capability is missing (a
   tool, a source, a verification), the system fails *visibly* rather than producing
   plausible-but-unverified output that a human will mistake for the real thing.
9. **Self-improvement is a disciplined loop, not vigilance.** The process evolves by
   *triggers* (condition → action, recorded where the actor will trip over them), the
   evolution *rides an existing ritual* rather than adding ceremony, and it is biased
   toward *retiring* rules and staying *subordinate to the actual work* — so it adapts
   without bloating or becoming a meta-treadmill.
10. **Minimal-first, reference-not-adoption.** Start with the smallest process that
    works; add a step only when *friction reveals the boundary*; treat existing
    frameworks (BMAD, spec-kit) as a *parts catalog* to borrow from, recorded as an
    explicit decision — not a wholesale identity to inherit.

If these are genuinely invariant — and I think 1–8 clearly are, 9–10 are strong — then
Trellis has bedrock. The product is: **encode these, make everything else adaptable.**

> **A naming lever for later (use it honestly).** A memorable, authoritative-sounding
> name for *this set* — the way "Conway's Law" or "Postel's Law" travels — is a real
> adoption move; eponymous framing borrows a citability that "ten properties of agentic
> dev" never gets. Treat it as a *deliberate later decision*, distinct from the product
> name (Trellis), and made only once the set's durability is proven across instances.
> **Hard guardrail** (this project's own research discipline — sources + confidence tags,
> loud-failure over silent degradation): if you name them authoritatively, attribute the
> set clearly as *our synthesis*, never implying pre-existing provenance. An
> authoritative tone manufactures false credibility — the author of this brief briefly
> mistook an invented codename ("Keel's invariants") for an established concept, which is
> the canary. For now they are exactly: **"Trellis's invariants — our synthesis, v0."**

## 5. The four pillars

The invariants organize into four capabilities. A real product needs all four; the
source project has 1, 2, and 4 working and 3 only as a seed.

- **I. Spine (the invariants).** The non-negotiable properties above, encoded as
  always-loaded guardrails + checkable rubrics. This is what "guards the current process
  to be respected" means concretely.
- **II. Adaptation (fit + evolve).** *Fit:* given a project's stage/scale/risk,
  instantiate the minimal set of stages, gates, and artifacts — using BMAD/spec-kit/etc.
  as a parts catalog ("you don't need a story-map yet; you do need a conformance gate").
  *Evolve:* the trigger-driven self-improvement loop (invariant 9), so the fitted process
  changes safely over time, with rationale preserved.
- **III. Onboarding (teach in situ).** The system explains the step you're in *as you
  reach it* — why this gate exists, what "done" looks like, what to do when blocked —
  instead of requiring prior study of a manual. This is the answer to "don't force people
  to learn a moving target." (Least proven; possibly the hardest and the biggest
  differentiator.)
- **IV. Enforcement & audit (guard + trace).** Independent verification (invariant 5),
  the artifact-graph + decision-log as an audit trail (invariants 3, 7), and loud-failure
  (invariant 8). This is the enterprise-trust pillar.

**Two operating modes** (the "point it at BMAD or spec-kit" idea, made concrete):
- *Conductor mode:* you adopt an existing methodology; Trellis runs it, enforces its gates,
  teaches it, and flags when a step is being skipped or has gone stale.
- *Author mode:* you have no fixed methodology; Trellis proposes a minimal fitted process
  from the parts catalog, omitting steps you don't need and *suggesting new steps when
  friction warrants* — exactly how the source project built its own operating model.

## 6. Existence proof: it is already running (the source project)

The strongest evidence that this isn't vaporware is that a solo, real project built
*itself* with this discipline and the meta-layer was implicit the whole time. Concrete
assets to harvest (each is a process-agnostic embodiment of an invariant):

| Source-project asset | Invariant it embodies | Harvest as |
|---|---|---|
| **Artifact contract** — YAML frontmatter (`id/type/status/depends_on/rubric/owner`); `draft → gated → approved` lifecycle; "never consume a draft"; mandatory `Acceptance criteria` + `Open questions` | 1, 3 (directional flow, provenance) | The core data model of Trellis — a portable artifact schema + lifecycle |
| **Stage taxonomy** — stages defined by *nature of work + how output is verified*, explicitly marked "provisional, evolves with friction" | 1, 2, 10 | A *template* (take the shape, not the specific five stages) |
| **Gate-authority model** — human at intent, agent-autonomous at execution under conformance gates, *ratcheting* on track record, intent gate "never opens" | 4 | The authority-boundary engine + the ratchet mechanism |
| **`conformance-reviewer`** — independent, read-only, derives its own checklist from the approved upstream, "builder doesn't grade itself," honesty clause | 5 | A reusable verification sub-agent template |
| **Research discipline** — every load-bearing claim needs a linked source + confidence tag (`verified/inferred/speculated`); abort-loudly preflight | 3, 8 | A research-quality gate + the loud-failure pattern |
| **Self-improvement loop** — triggers-not-vigilance, live where they fire, ride the existing PR field, prefer-retiring, subordinate-to-work | 9 | The adaptation engine (Pillar II's "evolve") |
| **ADR discipline** — append-not-edit, supersede with forward pointers, "bound context in the consolidated layer" | 6, 7 | Decision-log + current-truth separation |
| **Rubric mechanism** — quality gates as checkable rubrics; "listing failures accurately *is* success" | 2 | The gate-design kit |
| **"References & inspiration" stance** — *"we deliberately do not adopt full frameworks like BMAD, but treat them as battle-tested reference points… borrowing their gate designs is always on the table… record any change of course as a decision"* | 10 | This paragraph is *literally the product thesis already written down* — the parts-catalog philosophy |

That last row matters: the source project already articulated reference-not-adoption in
its own operating instructions. The meta-layer isn't a hypothesis bolted onto the
project — it's the *implicit method the project was already using*, waiting to be lifted
out and generalized.

What **not** to harvest (project-specific): the domain (a math tutor), the tech stack,
the persona, the specific 5 stages, the GitHub-runner specifics. (One transferable idea
buried in there: the **venue split** — async fire-and-forget dispatch vs. hands-on
iterative work — is a generalizable pattern for *where* agent work runs.)

## 7. Honest risks and open questions

I'd be doing you a disservice to only sell it. The real hazards:

- **Abstraction → vaporware.** "Process-agnostic meta-layer" can float free of anything
  concrete and become a philosophy deck. **Mitigation, and the single most important
  design rule:** Trellis must *always ground out* in concrete, project-specific artifacts (a
  generated instructions file, real gates, real sub-agents). If it ever just *describes*
  process instead of *producing and enforcing* it, it has failed. The source project is
  proof it can be concrete — keep it that way.
- **N=1 generalization.** The source project is one solo maintainer, one domain. It
  validates the *mechanics* of the invariants and loop; it does **not** validate that
  they generalize across team sizes, domains, and regulatory contexts. That leap is the
  central unproven claim. Early work should be *finding a second and third instance*, not
  polishing the first.
- **Product vs. consulting.** "Guide and audit companies adopting AI-first dev" can be a
  *product* (a pack they run) or a *consulting practice* (you run it, the pack is your
  delivery tool / lead-gen). These imply very different builds. Decide deliberately;
  don't drift.
- **Onboarding is the hardest pillar and the least proven here.** In-situ tutoring that's
  genuinely good (not annoying, not condescending, adaptive to expertise) is a real
  research/UX problem. It's also the biggest differentiator. High risk, high reward —
  scope it explicitly.
- **Moat.** What stops Anthropic or the framework authors from doing this? The defensible
  parts are (a) a *well-curated* invariant set + adaptation discipline (hard-won, like
  this project's was), (b) onboarding quality, (c) enterprise trust/audit depth. It's a
  real but not enormous moat; being early and opinionated-about-invariants is the edge.
- **Adoption friction.** Selling a *meta*-process is harder than selling a concrete one —
  "give me the steps" is an easier buy than "give me the thing that decides the steps."
  Conductor-mode (run *their* existing framework) may be the easier on-ramp than
  Author-mode.

## 8. A possible MVP shape (so the conversation has something to push on)

Smallest thing that tests the core bet without boiling the ocean:

1. **The portable artifact contract + lifecycle** (Pillar I) — schema, statuses,
   directional-flow enforcement. Harvest near-directly from the source project.
2. **One verification sub-agent** (Pillar IV) — the `conformance-reviewer`, generalized:
   "verify any artifact/change against its approved upstream + a self-derived checklist."
3. **The trigger-driven self-improvement loop** (Pillar II / invariant 9) — the smallest,
   most distinctive, most demo-able piece.
4. **Author-mode v0** — point it at *one* reference (start with spec-kit *or* BMAD), and
   have it produce a *fitted, minimal* instructions file + gates for a sample project,
   explaining each inclusion/omission. This is the headline demo.
5. *Defer* full in-situ onboarding (Pillar III) and the audit/compliance depth to v2 —
   name them as the known frontier.

Then the validating move is not more building — it's **running the MVP on a second real
project in a different domain** and seeing whether the invariants survive contact.

## 9. The strategic forks to decide first (your call, not mine)

1. **Product or consulting** (§7) — determines everything downstream.
2. **Conductor-first or Author-first** — which mode is the on-ramp?
3. **Which reference framework to support first** (spec-kit's lighter weight vs BMAD's
   richer structure) as the parts catalog seed.
4. **Target buyer** — startup speed-and-safety, or enterprise trust-and-audit? Different
   pillars lead.

## 10. Verdict

Not off. The thesis (sell the meta-process, not the process) is genuinely
under-served, it rests on real invariants rather than abstraction, and you have a
working — if N=1 — existence proof to harvest from. The two things most likely to kill
it are **staying too abstract** (mitigated by the iron rule: always produce concrete
artifacts) and **over-trusting the single instance** (mitigated by getting to a second
project fast). The onboarding pillar is the high-risk, high-differentiation frontier.
Worth a real exploration. Start minimal, harvest the source project, find instance #2.

---

## 11. Start here — first moves (for an agent or person picking this up cold)

You can begin **without** resolving the strategic forks (§9) — by making "surface the
forks" the first act and building the fork-independent spine first.

1. **Read for the *why*:** §1–§5 (idea + invariants) and §10 (verdict). Then §12 for
   *how this project works*.
2. **First human-facing act — surface the forks (§9) and get decisions.** Direction is
   the human's to set (invariant 4: humans own intent). Do not build past these blind.
   Record the answers as the project's first decision records (per §12). If the human is
   unavailable, proceed to step 3 — it is fork-independent.
3. **Build the spine first (always safe, fork-independent):** the **portable artifact
   contract + lifecycle** (MVP item §8.1) — the schema, statuses, and directional-flow
   enforcement. Every version of this product needs it regardless of any fork, so it is
   the correct first build under any uncertainty.
4. **Dogfood from commit one:** adopt §12 as *this project's own* operating method
   immediately. The first real validation of Trellis is whether building Trellis with Trellis's
   invariants feels right; friction you hit *is* product research, so record it (§12's
   self-improvement rule).
5. **Then proceed in MVP order** (§8): independent-verification sub-agent → the
   self-improvement loop → the author-mode demo — surfacing each fork as it becomes
   decision-relevant rather than all at once.

The headline demo to aim at (proves the core bet): **point Trellis at one reference
framework (spec-kit or BMAD) and have it produce a *fitted, minimal* operating file +
gates for a sample project, explaining every inclusion and omission.**

## 12. Seed operating method (deliberately tiny — grow it by the loop)

This is the new project's *own* bootstrap process — and writing it small is on-thesis
(invariant 10: start minimal, add only on friction). It is the seed that §13's
`CLAUDE.md` is a mature instance of.

- **Artifacts.** Every non-code document opens with frontmatter
  (`id / type / status / depends_on / owner`); statuses `draft → ratified → approved`;
  **downstream consumes only ratified/approved upstream**, never drafts. Every artifact
  carries `## Acceptance criteria` and `## Open questions`.
- **Decisions.** Significant choices get an **append-only** decision record; you
  *supersede*, never edit, a ratified one. **The four forks (§9) are the first records.**
- **Gates.** Human approval at the **intent** layer (vision, decisions); **independent
  verification** at the **execution** layer (a conformance check against the approved
  upstream before merge — "the builder does not grade itself").
- **Work.** One logical change per PR; descriptive, linear history; diffs small enough
  to review on a phone.
- **Self-improvement.** Triggers, not vigilance (invariant 9): when friction reveals a
  missing rule, add it *where it fires*, prefer retiring to adding, keep it subordinate
  to the work. **This file is the first trigger home.**
- **Loud failure.** If a required tool, source, or verification is missing, stop
  *loudly*; never emit plausible-but-unverified output.

Everything else is added only when friction reveals the boundary. That restraint is the
product demonstrating itself.

## 13. Source-project reference (optional — a parts catalog, not a dependency)

The source project is **Math Quest** — **`https://github.com/gundisalwa/math-quest`**
(private; you'll need access). The brief carries enough to start **without** it, but the
repo has *working* versions of every harvested asset, and working files beat prose when
you build the real thing. If you have access, read these repo-relative paths as
**adaptable templates** — clone-and-harvest, do not link or depend:

- `CLAUDE.md` — a mature instance of §12's operating method (what §12 grows into).
- `decisions/adr-0009-governance.md` — the gate-authority + autonomy-ratchet model
  (invariant 4).
- `decisions/adr-0008-spec-stage.md` — behavioral-contract authoring; plan-inline-by-
  default (invariants 1, 6).
- `rubrics/spec-quality.md` + any spec's `## Rubric check` — the gate-as-rubric mechanism
  + the honesty clause (invariant 2).
- `.claude/agents/conformance-reviewer.md` — the independent-verification sub-agent
  (invariant 5).
- the *Artifact contract* and *Self-improvement* sections of `CLAUDE.md` — invariants
  1/3/9 made concrete.

**Caveat (itself on-thesis).** Treat the source project exactly as the product treats
BMAD — a parts catalog to harvest from (invariant 10), not an identity to inherit and
not a runtime dependency. If you find yourself unable to proceed without it, that is a
signal the brief is missing something — fix the brief, don't deepen the coupling.

---

— *Prepared as conversation-bootstrap context; argue with all of it.*

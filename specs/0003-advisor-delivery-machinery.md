---
id: spec-0003
type: spec
status: ratified
ratified: 2026-07-05
depends_on: [spec-0001, spec-0002, signature-catalog-v1, decision-0010, decision-0012, decision-0019, decision-0023, invariants-v1]
owner: gundi
rubric: spec-quality
---

# Spec 0003 — The v0 delivery machinery: advisor mode as a Claude Code plugin

> **The first real machinery.** Everything so far is intent + a checkable rule-set; this is the slice
> that puts Trellis *into a host project*. `decision-0019` fixed **advisor mode via a host-invoked CLI**
> as the v0 on-ramp; `decision-0012` fixed **v0 delivery = a self-hostable Claude-Code plugin**;
> `decision-0010` fixed **no runtime — it ships as instructions.** This spec assembles those into a
> concrete plugin + flow, and executes the **activation/wiring contract** named-but-deferred in
> `spec-0001` §5.

## Purpose

Specify the smallest thing that lets a developer point their coding agents at Trellis and get value:
**onboard → assess → a ratified expression profile the host's agents consult.** No new runtime — every
part is a **skill, a sub-agent, or a reference doc** (`decision-0010`). Draft; ratified at merge
(`decision-0022`).

## Scope

**In scope (v0 advisor):** the plugin's shape (skills, the **Assess** sub-agent, the bundled
invariants + catalog as reference); the advisor flow (onboard with a starting posture, Assess → a draft
profile, human ratifies, agents consult); the **activation/wiring** contract (compose onto the host,
augment-never-clobber); the onboarding **postures** (A conductor · B author-adapt · seed · Custom).

**Out of scope (later specs):** **supervisor mode** (installed, live gates firing on commit/PR events
— the paid/heavier tier); **hosted services** (conformance-in-CI, cross-instance analytics); the
**non-Claude-host CLI** (git-copy always works; a thin cross-host CLI is later). This spec is *pull /
consult*, not *push / enforce*.

## 1. Delivery form — a Claude-Code plugin, no runtime (`decision-0010/0012`)

v0 ships as a **self-hostable Claude-Code plugin** (a git repo with `marketplace.json`, no Anthropic
approval — `decision-0012`). It bundles: the **skills** (onboard, assess), the **Assess** and existing
**conformance-reviewer** sub-agents, and the **invariants + signature catalog** as reference docs. "The
CLI" the maintainer invokes is **the host's own** (Claude Code) — Trellis adds advisor skills to it;
there is no separate Trellis binary in the host's critical path. Git copy-in also works (availability
without activation).

## 2. The advisor flow

1. **Onboard** *(skill)* — the maintainer picks a starting **posture**, which seeds an
   `expression-profile` (`spec-0002`): **A · conductor** (adopt one framework, strictly — high
   enforcement), **B · author-adapt** (synthesize/evolve — moderate, self-improvement high), **seed**
   (grow from scratch — minimal, ratcheting), or **Custom**. These map to `decision-0009`'s three
   scenarios; the profiles are shipped presets (owed alongside this build).
2. **Assess** *(read-only sub-agent)* — reads the host project and detects which invariants it already
   honors **implicitly**, matching each against the catalog's `signature` tells, and drafts an
   `expression-profile`: per honored gene, `basis: honored-implicitly` + `confidence` + a concrete
   `evidence` pointer (`spec-0002` evidence floor; **assert-and-verify, never silently "honored"**).
   **Producer ≠ ratifier:** Assess emits a **draft**.
3. **Ratify** *(D2)* — the maintainer reviews and ratifies the profile (merge = ratify, `decision-0022`,
   or an explicit approval). Trellis **suggests, never self-applies** (B3/D2).
4. **Consult** *(advisor mode)* — the ratified profile + the active invariants + their catalog `why`/
   examples become the reference the host's agents consult; the profile's dials say *which* invariants
   at *what* strength. Composed onto the host (below), the host's agents follow them — by reading
   instructions, no runtime dependency on Trellis.

**Gatekeeping is detect-and-respect (`decision-0024`).** Assess reads which handovers a project treats
as human- vs agent-gated and **respects them** — it never imposes a per-gate map (that is v2). Trellis's
contribution is invariant **B2** made real: **surface a human-gated handover that runs without its human
approval** — one-directional (agent gates proceed silently, and human approval on an agent gate is
never sought). The profile's `C2` is exactly this detected readout, not a Trellis-imposed choice.

## 2b. The setup CLI — the interactive front door (`decision-0023`)

The advisor flow (§2) is driven by a **Go setup CLI** (single binary, `curl … | sh`, no package
manager — `decision-0023`), run once. It is **setup tooling, not a runtime**. Its interactive flow:

1. **Pick the install mode first (`decision-0029`)** — the mode decides what the rest of setup even
   needs to detect:
   - **M1 — alongside (overlay, default, v0):** Trellis installs *next to* the project's existing
     instructions and is **called from them** — augment-never-clobber (`spec-0001` §5). It is a
     **deterministic file overlay and needs no harness binary**; requiring one to overlay was friction
     we hit and removed (`decision-0029`).
   - **M2 — rewrite (morph, the v0-next stacked follow-up):** rewrite the project's own machinery to
     **bake in** the profile's invariants. **Always on a fresh git branch, opened as a PR for the
     maintainer** — never in place; if **no git repo** is detected, the CLI **hard-warns / refuses**.
     Carries a **"keep embedded behaviors" dial** — *keep* the host's existing behaviors or *replace*
     them with the profile's (default **keep**).
2. **Detect what the mode needs.**
   - **M1** targets an **instruction file** (v0: `CLAUDE.md`; detecting/choosing among instruction
     files — `AGENTS.md`, etc. — is a stacked follow-up). No binary; if there is nothing to attach to,
     offer to create `CLAUDE.md`.
   - **M2** detects the **harness binary** that drives the rewrite (v0: **Claude Code** — the `claude`
     binary; `.claude/` / `CLAUDE.md` corroborate). *v0 assumes CLI harnesses; Claude-only.* If **none**
     is found, the CLI **exits with a clear message** rather than guessing. Multiple/other harnesses →
     deferred, Claude favored.
3. **Pick the expression profile** — the posture **A conductor · B author-adapt · seed · Custom**
   (§2 step 1), which seeds a profile. M2 also picks a reasoning **model** (M1 is deterministic — none).

**Edge cases the build must handle:** an **existing `CLAUDE.md`** (M1 appends a Trellis section — never
silent overwrite); **idempotent re-run** (no double-apply); **Custom** profile (edit-after, or
interactive dials); **clean uninstall** (AC4). **Supervisor mode** (installed, live gates): when the
harness + files point to Claude, the **plugin route** (`decision-0012`) is favored; other harness
surfaces deferred.

## 3. Activation / wiring (executes `spec-0001` §5)

- **Composition — augment, never clobber.** Trellis composes onto the host's `CLAUDE.md`/instructions;
  it never overwrites them, and any change it proposes to host instructions is a **surfaced decision**
  (D1) the maintainer rules on. Uninstall is clean.
- **Activation level = the C1 dial, surfaced** (`decision-0008`): *available + referenced* → *skills
  fire* → *default agent*, chosen by the maintainer at onboarding, **never silently maximal**. The
  posture (A/B/seed) sets the initial dials.
- **Model 1 overlay by default** (`research-0005/0006`): set the host's expression profile without
  editing its files; morph (M2) is the **v0-next** follow-up (§2b), not dropped.

## 4. No runtime (`decision-0010`)

Every component is agent instructions: **skills** (model-invoked), **sub-agents** (Assess drafts a
profile; conformance-reviewer checks the corpus), the **catalog/invariants** (reference). No Python/
Node process sits in the host's critical path. The plugin *may* offer opt-in **hooks** (e.g.
conformance-on-PR), but that is the *supervisor* posture's wiring — v0 advisor is **consult-only**.

## Acceptance criteria

- **AC1 — a real plugin, no runtime.** Installing it into a Claude-Code host adds the onboard + assess
  skills and the sub-agents; nothing requires a Trellis runtime process (`decision-0010`).
- **AC2 — onboarding seeds a profile.** Choosing A / B / seed / Custom produces a starting
  `expression-profile` conforming to `spec-0002`, with the dials the posture implies.
- **AC3 — Assess is grounded and honest.** Assess emits a **draft** profile whose every
  `honored-implicitly` gene carries `confidence` + a real `evidence` tell from the project (iron rule);
  it never claims a gene honored without evidence, and fails loudly on an unreadable project (D1).
- **AC4 — augment-not-clobber.** Applying the overlay preserves the host's prior instructions; any
  change to them is a surfaced decision; uninstall leaves the host as it was.
- **AC5 — producer ≠ ratifier.** Assess proposes; the human ratifies the profile (D2). Trellis never
  self-applies a profile to the host.
- **AC6 — dogfood.** The plugin, run on *this* repo, reproduces (within confidence) the hand-authored
  `profile-trellis-self` — the round-trip that proves Assess works (`spec-0002` AC7).

## Open questions

- **Skill/command surface** — the exact names + invocation (`/trellis onboard`, `/trellis assess`?),
  and how much is one skill vs several.
- **Preset dial precision** — the per-invariant `C1`/`C2` for the A / B / seed profiles (owed with the
  preset artifacts; `decision-0020` flagged it).
- **Assess detection, generically** — how the sub-agent reads an arbitrary project (languages, layouts)
  and matches the catalog `signature` tells; the heuristics live in the catalog (`trellis-product`),
  not the skill. First real test is the dogfood (AC6), then a genuinely external instance (the N=1 risk).
- **The consult substrate** — does "compose onto the host" mean a generated `CLAUDE.md` section, a
  referenced file, or a default agent? Likely the C1 dial chooses; pin it in the build.
- **Non-Claude hosts** — git-copy works today; a thin cross-host CLI is a later slice.
- **Supervisor mode** — installed, live gates (hooks on commit/PR), the paid/heavier tier — the next
  delivery spec, not this one.

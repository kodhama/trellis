---
id: research-0011
type: research-note
status: ratified
depends_on: [invariants-v1, decision-0001]
owner: gundi
ratified: 2026-07-06
---

# 0011 — Eval harness: does applying Trellis measurably change agent behavior?

The whole project rests on an asserted claim — that laying Trellis over a project makes an agent
*behave better according to that project's own rules and the invariants*. This note designs an
**empirical, repeatable** way to test it, using real, auto-installable frameworks as instances (which
also chips at the **N=1** risk, `decision-0001`).

## Frameworks (auto-installable + declared rules)

Survey confidence high (installs verified against project docs, 2026-07). Two gates: a **prompt-free
setup command** and **rules concrete enough to score adherence**.

| Framework | Non-interactive setup | Declares checkable rules |
|---|---|---|
| **GitHub Spec Kit** (primary, scaffold verified) | `uvx --from git+https://github.com/github/spec-kit.git specify init NAME --integration claude --script sh --ignore-agent-tools` | strong: `constitution.md` + ordered `spec → plan → tasks` |
| **OpenSpec** (Fission-AI, 58.8k★) | `npm i -g @fission-ai/openspec@latest && openspec init . --tools claude --force` | change folders (`proposal/specs/design/tasks`) + `openspec validate` + `--json` |
| **cc-sdd** (gotalab, best-fit new) | `npx cc-sdd@latest --claude-skills` | EARS requirements → approval gate → per-task TDD review |
| **BMAD-METHOD** | `npx bmad-method install --yes --tools claude-code --modules bmm` | phased role-agent workflow; PRD/architecture/stories |
| **Agent OS** | `curl -sSL .../setup/project.sh \| bash -s -- --no-base --claude-code` | `standards/` + `instructions/` |
| **spec-workflow-mcp** (GPL-3.0) | `npx -y @pimzino/spec-workflow-mcp@latest <path>` (MCP server) | Req→Design→Tasks + per-stage approval — enforced via **MCP tool state**, not files |

**Blocked — `spec-swarm`** (MartyBonacci): good rules (constitution + a 0–100 ship gate, default 80), but
install is **Claude-Code-plugin-only via interactive slash commands** (`/plugin install` + `/ss:init`) —
no scriptable setup, so it fails the harness's one hard gate. Include only if we can drive Claude Code's
plugin install out-of-band (unverified). **Tessl** is excluded too (closed beta, non-deterministic).

**Aider is the harness *driver*, not a framework to A/B** — it runs headless (`aider --message … --yes`)
but imposes no process. `AGENTS.md` / Cursor rules are *targets* (where instructions land), not
installable scaffolds. Ranked fit for A/B (install × checkable rules × reproducibility): **cc-sdd ·
OpenSpec · Spec Kit** lead; the MCP-shaped `spec-workflow-mcp` needs a scorer that reads tool state, not
files. (Full survey: the two research subagent reports, folded here.)

## Method

**Blind A/B, same worker, same task; the only difference is the Trellis overlay.**

- **Arms.** *A (baseline)* = framework X scaffolded, its instructions only. *B (+Trellis)* = the same,
  plus `trellis setup` (the overlay). Everything else identical.
- **Worker.** A coding agent (headless Claude by default; Aider a pluggable alternative) runs a dev
  **task** in the project and produces a transcript of what it did and said.
- **Two scorecards.** A reviewer scores the transcript against **(a)** the framework's *own* rules
  (`eval/scorecards/<framework>.md` — "does Trellis help you follow *your* method?") and **(b)** the
  **invariants** (`eval/scorecards/invariants.md`, auto-derived from the catalog — the ✗-failure
  rubric). Each item: *followed / violated / n-a* + an evidence quote.
- **Independent, blind reviewer.** The reviewer sees the **task + the worker's transcript, not the
  instructions the worker was given** — so it can't infer the arm and scores *behavior*, not the presence
  of Trellis. A **judge panel** (≥3, majority) absorbs stochasticity (`inv-independent-judgment`,
  dogfooded).
- **Outcome.** Δ = B − A on each scorecard, over repeats × tasks. **Win condition:** +Trellis clearly
  lifts the invariant score **without hurting** framework-adherence (Trellis *respects* your method).

**Tasks** are chosen to *create* invariant-testing moments (`eval/tasks/`): an underspecified feature
(clarify vs guess), a breaking rename (propagate + gate vs silent edit), a "finalize and ship" (halt for
human sign-off vs auto-complete).

## Open questions (assumptions this note makes — flag to revisit)

- **Worker agent = headless Claude** by default (Aider pluggable). A different worker could shift the
  baseline; the *delta* should be more robust than the absolute.
- **Blinding is imperfect** — the overlay is *in* the worker's instructions; we blind the reviewer to
  instructions, not the world. Scoring behavior (not instructions) is the mitigation, not a cure.
- **Sample size** — first runs are a proof-of-concept (few tasks × few repeats), not a powered study;
  the harness is built to scale the counts up.
- **Effect size ∝ task subtlety × baseline weakness** (learned from a confounded task-03 run,
  `eval/runs/spec-kit-lite/03-finalize-and-ship/NOTES.md`): the worker is Claude, whose *default* judgment
  is already strong, so Trellis's measurable Δ concentrates where a strong agent **still slips** (task 01's
  self-certification of a plausibly-complete feature) — not on obvious violations (task 03's nothing-to-
  merge, which the baseline already catches). Design tasks for *subtle* invariant moments; consider a
  weaker/faster worker on some arms to widen the gap.
- **Framework set** — Spec Kit first (richest checkable artifacts, official/MIT); BMAD/Agent OS next.
- **Harness shape** — a shell orchestrator (scaffold → apply → `claude -p` worker → `claude -p`
  reviewer → aggregate) vs a Trellis Workflow script; the former is opt-in-free and CI-friendly.
- **The invariant scorecard is a derived resource** of the catalog (`decision-0028`) — it has a
  generator; a CI sync-guard is owed.

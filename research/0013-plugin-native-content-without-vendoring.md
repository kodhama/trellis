---
id: research-0013
type: research-note
status: draft
depends_on: [decision-0051, decision-0053]
informed_by: [decision-0054, decision-0055, research-0010]
owner: agent
date: 2026-07-21
---

# 0013 — Can plugin-native delivery replace vendoring? (parked, not pursued)

> **Provenance.** The maintainer observed grove migrating agent/companion files toward
> being pure "plugin files" — never copied into a consumer's repo — and asked whether
> trellis's own vendored `.trellis/internal/*` could do the same, "without having to edit
> ... CLAUDE.md," with a human-facing README kept for reference only. Explicitly **not**
> assumed to be the identical class of problem grove has ("this is likely not the same
> issue... although it might be around the same class" — maintainer, 2026-07-21) — captured
> here as its own open question, not folded into grove's. Directed to be parked: **"let's
> stick to what we have but consider it later"** (maintainer, 2026-07-21). This note records
> the finding so the question doesn't have to be re-researched from zero when it comes back.

## Question

Can trellis deliver `invariants.md` / `rules.md` / `trellis.md` content to a consumer
session as **always-on context, with zero copies committed into the consumer's repo** —
so a refresh is "set `rules.toml` rows," never "rewrite files" — the same simplification
grove is reportedly moving toward for its own agent/companion files?

## Findings (researched 2026-07-20/21, via a `claude-code-guide` dispatch against current
Claude Code docs and issue tracker)

- **`@import` cannot reach a plugin's own files.** It does not support variable
  substitution — `${CLAUDE_PLUGIN_ROOT}` inside an `@import` line is a confirmed *open*
  bug (`anthropics/claude-code#9354`: works in JSON hook/MCP configs, fails in command
  markdown files). Even setting the bug aside, there is no *stable* absolute path to
  hardcode — a plugin's install location varies by user and machine, which is the entire
  reason `${CLAUDE_PLUGIN_ROOT}` exists. So a one-line `@import` straight into the
  plugin's own directory is not achievable today.
- **The mechanism that *does* work: a `SessionStart` hook returning `additionalContext`.**
  Docs confirm this output "is injected into Claude's context window as a system reminder
  and persists across multiple prompts in the same session" — genuine standing
  instructions, not a one-shot nudge. `${CLAUDE_PLUGIN_ROOT}` resolves correctly inside a
  hook's shell execution (the asymmetry with `@import` is real and specific to the import
  parser, not to plugins generally). Trellis already ships a `SessionStart` hook
  (`decision-0039`) — today it only compares the overlay's `plugin@<sha>` stamp and emits a
  staleness nudge; extending it to emit the rule content itself is a small step from what
  already exists, not new machinery.
- **Plugin-provided `agents/` need no vendoring at all** — they appear in `@`-mention
  typeahead as `plugin-name:agent-name` once the plugin is enabled, no per-project copy.
  This is the part of the maintainer's framing closest to what grove is reportedly doing.
  It does not, on its own, solve grove's actual reason for vendoring agent files (per-
  project placeholder resolution, e.g. `<TEST_CMD>`) — a separate problem.
- **Skills are auto-discovered without vendoring, but load on demand only** — not
  always-on, so they don't fit the "standing instructions" requirement.
- **No real-world precedent was found** for a plugin shipping purely-static, always-on
  reference content with zero vendoring. Existing patterns are either "vendor into the
  repo" or "hook injects dynamic, computed content." If trellis built this, it would be
  relatively novel — not necessarily wrong, but without an existing implementation to
  crib gotchas from.

## A concrete counter-example already in this repo

`eval/experiments/does-trellis-help/run.sh`'s `+trellis` arm construction
(`run_arm()`) does exactly the vendoring this note asks whether trellis can avoid — it
`cp`s `.trellis/internal/{invariants,rules,trellis}.md`, `version`, and `rules.toml` into
a scaffolded fixture directory, then inlines the block into `AGENTS.md`. The comment
directly on that call names why:

> "+Trellis arm only: apply the overlay, inlined into `AGENTS.md` so both subagent and
> `claude -p` workers see the directives (an `@import` wouldn't resolve for a bare
> subagent worker)."

That's independent, already-discovered evidence for one of this note's own findings: even
*within* Claude Code, headless `claude -p` invocations and bare subagent workers are not
guaranteed to have the plugin installed/enabled or the hook firing in that ephemeral
context — so a hook-only delivery mechanism would not, on current evidence, cover the
surface trellis's own eval harness already has to support by mechanical file copy. Any
future design has to either accept a narrower guarantee (interactive sessions only) or
keep a fallback vendoring path for headless/subagent invocations — it can't cleanly
replace vendoring end-to-end, only for one class of session.

## Scope, pragmatically (maintainer, 2026-07-21)

Trellis today has exactly one user — the maintainer, on Claude Code. They "wouldn't be
shocked" to narrow support to Claude Code alone for now, dropping the other harnesses
`research-0010` catalogued (`AGENTS.md`, `GEMINI.md`, `.clinerules`,
`.github/copilot-instructions.md`) and reintroducing them later. The stated reason this
is acceptable *now* but not permanently: there's a separate, not-yet-started effort (a
landing page) aimed at eventually making trellis available to other people — multi-harness
portability matters again once that audience exists, not before. Recorded here so a future
session doesn't mistake single-harness scoping for a permanent decision to drop
`research-0010`'s registry — it's a pragmatic sequencing call, reversible, tied to a
concrete future trigger (the landing-page effort), not a retraction of the portability
goal.

## Target shape: two delivery models, not a third abstraction (maintainer, 2026-07-21)

Given the above, the maintainer is "okay with just supporting two models already from the
start":

1. **Vendored, via installation** ("quite classical") — what trellis does today:
   `.trellis/internal/*` copied into the consumer's repo at setup/refresh. Stays as the
   portable fallback — works for any harness, including ones with no native plugin
   system at all.
2. **Plugin-native** — delivery through a harness's *own* plugin system, e.g. Claude
   Code's `SessionStart` hook + `additionalContext` (this note's main finding above).

**Explicitly not a third option**: a trellis-owned, generic, cross-harness plugin
abstraction that tries to unify however each harness's plugin system works underneath.
Reasoning given directly: "every vendor will have its own [plugin system], I guess" —
each harness's plugin system is already its own investment to target correctly; layering
trellis's own abstraction on top would be a second thing to maintain in sync with N
underlying systems, not a simplification. Instead: target each harness's *actual* native
plugin system directly, harness by harness, keeping model 1 as the universal fallback
for anything not yet targeted natively.

This reframes `decision-0051`'s vendoring apparatus as **model 1, staying**, not
something plugin-native delivery replaces — the open work is adding model 2 for Claude
Code specifically, not retiring model 1.

## Per-harness plugin system survey (started 2026-07-21)

The maintainer named a second concrete harness worth checking directly, rather than
assuming Claude Code's `SessionStart`/`additionalContext` shape generalizes:

- **Claude Code** — covered above. Plugin-native, always-on delivery is possible via a
  `SessionStart` hook returning `additionalContext`.
- **OpenCode** — verified via its own docs (`opencode.ai/docs/plugins/`, fetched
  2026-07-21), confirming the maintainer's recollection: plugins are npm/Bun-based.
  Declared in `opencode.json`'s `"plugin"` array (npm packages, installed automatically
  via `bun install`, cached under `~/.cache/opencode/node_modules/`) or loaded locally
  from `.opencode/plugins/`; a plugin needing external deps carries its own
  `package.json`. **But its hook surface is tool/session/file/message-scoped** —
  `tool.execute.before/after`, `session.created/compacted/updated`, `file.edited`,
  `message.updated`, permission and LSP events — **with no hook for injecting a
  persistent system prompt or always-on instructions**. The closest is
  `experimental.session.compacting`, which only fires at compaction time, not
  continuously. So OpenCode having *a* native plugin system does not automatically mean
  it can deliver model 2's actual goal (always-on, zero-vendor instructions) — that
  needs checking per harness, not assumed from "has a plugin system."

**Correction/addition (2026-07-21) — OpenCode has a separate, non-plugin mechanism that
does deliver the goal.** The maintainer flagged this directly: OpenCode's config
(`opencode.json`/`opencode.jsonc`, project *or* global `~/.config/opencode/opencode.json`)
has an `instructions` field — a static array, not a hook, not part of the plugin system.
Verified against the actual resolution code
(`packages/opencode/src/session/instruction.ts`, `anomalyco/opencode`, fetched 2026-07-21),
not just the docs page (which two independent fetches gave inconsistent answers on — see
below):

```js
const urls = (config.instructions ?? []).filter(
  (item) => item.startsWith("https://") || item.startsWith("http://"),
)
```

`instructions` accepts local paths, glob patterns (`packages/*/AGENTS.md`), and — per
this source, not just the docs — genuine **remote `http(s)://` URLs**, fetched live with
a 5-second timeout, all combined automatically with `AGENTS.md`. Set in the *global*
config, this needs zero per-project file commits at all: a single line, once, per user —
which fits the current single-user reality (Scope section above) even better than a
hook would. One reliability caveat found in the same search: a real bug report
(`anomalyco/opencode#4758`, "custom instruction files in `opencode.jsonc` not being
loaded") describes exactly this field silently not firing — closed as completed
2026-03-26, so presumed fixed, but it shows the field has had at least one real rough
edge, not a purely theoretical mechanism.

This doesn't fit cleanly into either of the two models above — it's neither vendored
copy nor plugin-hook delivery, but a third shape: **static, harness-native config
declaring where to always-load instructions from, with the harness itself doing the
fetch.** Worth naming as its own category rather than stretching "model 2" to cover it.
It's also a useful asymmetry to note against Claude Code directly: Claude Code has the
*hook* mechanism OpenCode's plugin system lacks, but OpenCode's `instructions` field can
do something Claude Code's own `@import` cannot — fetch a remote URL live. Neither
harness's native mechanism is a strict superset of the other's.

*(Methodological note: the two doc-page fetches earlier in this survey disagreed on
remote-URL support — one said yes, one said no. Source code, not docs, settled it. Kept
here as a reminder that this survey's doc-only findings — including the Claude Code
`SessionStart`/`additionalContext` finding above — are one level less certain than this
OpenCode finding, which is source-verified.)*

This is the shape a fuller survey would take: for each harness in `research-0010`'s
registry (once the landing-page effort makes that registry relevant again), check not
just "does it have a plugin system" but specifically "can a plugin inject persistent,
always-on instructions" — the two are independent questions, as OpenCode's case shows.

## Tradeoffs (not just the simplification)

- **Auditability drops.** Today `.trellis/internal/rules.md` is a committed file —
  inspectable in a GitHub browse or `git diff` with zero Claude Code session running.
  Hook-injected content exists only at runtime; a teammate reading the repo on GitHub
  would see nothing.
- **Claude-Code-only.** `research-0010` found the portable target is `AGENTS.md`,
  read natively by Codex, Devin/Cascade, Copilot, and Windsurf, none of which have any
  hook-equivalent. A hook-based path covers Claude Code alone — every other harness this
  project already targets would still need the vendored/inlined file, so this would *add*
  a second delivery path rather than replace the existing one.
- **Pinning disappears.** The committed `.trellis/internal/version` stamp today freezes
  exactly which payload a consumer is running until they refresh. Live hook injection
  means the effective rules change the moment the plugin updates, with no commit in the
  consumer's own history marking that it happened.
- **No loud-failure guarantee exists yet for this path.** The checksum/sync-guard
  apparatus (`decision-0028`, `TestBundledCatalogInSync`) exists specifically to make a
  broken or stale vendored copy visible. A hook has no equivalent today — if it fails
  silently, the agent gets zero injected rules with no visible sign anything is wrong.
- **`rules.toml` stays vendored regardless.** It is the one genuinely project-specific
  artifact (which rows are active) — nothing about plugin-native delivery removes the
  need for a consumer-owned config file; it only changes what happens to the *content*
  the config governs.
- **"Has a plugin system" ≠ "can deliver always-on instructions."** OpenCode has a real,
  well-developed native plugin system (npm/Bun-based) but no hook for persistent system
  prompts — confirmed above. Targeting each harness's native plugin system (model 2) is
  therefore a per-harness capability check, not a single pattern that transfers once
  proven on Claude Code.

## What this would touch, if pursued

Most of the vendoring/checksum/authority-split apparatus decisions `0051`, `0053`,
`0054`, and `0055` built exists specifically because vendoring was assumed necessary. A
hook-based redesign would be a genuine fork of that apparatus, not an incremental change
to it — sized like its own decision (or decision set) with an adversarial pass, not a
same-session extension of the current line of work.

## Codex survey update (2026-07-23 — bounded local prototype)

Codex is no longer wholly unsurveyed. Current official Codex documentation and
an isolated local experiment establish the following:

- **Plugin-bundled standing context is supported in the Codex hook contract.**
  Plugins can bundle lifecycle hooks; `SessionStart` and `SubagentStart` accept
  `additionalContext`, and plugin hooks receive `PLUGIN_ROOT` (plus
  `CLAUDE_PLUGIN_ROOT` compatibility). **Verified** against the current official
  Codex hooks/plugin documentation.
- **The current Trellis payload fits a compact hook message.** The prototype
  composed the real `rules.md`, project `rules.toml`, and payload version into
  6,519 characters. It returned byte-exact context for `startup`, `resume`,
  `clear`, `compact`, and `SubagentStart`; a missing config produced a visible
  `TRELLIS_HOOK_FAILURE` stop instruction. **Verified** by
  `eval/experiments/codex-hook-delivery/`.
- **Codex accepted the actual plugin packaging.** A temporary
  `.codex-plugin/plugin.json` was installed and enabled through a temporary
  marketplace and isolated Codex home; the installed-cache hook copy passed the
  same contract tests. **Verified** locally with Codex CLI `0.145.0`.
- **Live startup delivery is verified locally.** `codex debug prompt-input`
  renders static model input but does not fire `SessionStart`, so the
  maintainer ran one isolated, ephemeral, read-only `codex exec` probe against
  the installed prototype. With file and tool access forbidden by the prompt,
  it returned `TRELLIS_HOOK_CONTEXT payload@0760a802ccd1` and the requested
  injected rule slug `inv-handover-points`. This proves end-to-end local CLI
  startup injection. Live `resume`, `clear`, `compact`, `SubagentStart`,
  cloud/headless, and IDE behavior remain unverified.
- **A static bootstrap is still required for loud absence.** A hook can report
  its own read failure, but cannot report that it was disabled, untrusted, or
  never executed. The prototype therefore keeps a minimal `AGENTS.md`
  instruction requiring a separate injected block and stopping if it is
  absent. **Inferred** from the verified failure boundary; the model's
  compliance with that bootstrap is not yet measured.
- **Codex named agents remain a separate gap.** Current plugin documentation
  lists skills, hooks, MCP/apps, and assets but no plugin-carried custom-agent
  field; Codex custom agents live in `.codex/agents/*.toml` or the user-level
  equivalent. Rule delivery can be plugin-native without proving that Grove's
  named role fleet can be. **Verified** at the documented schema level; an
  undocumented compatibility path was not tested.

This changes the Codex row from “unsurveyed” to **contract-, packaging-, and
local-startup-verified; broader lifecycle/surfaces unverified**. It does not
justify replacing model 1 yet.

## Headless/cloud provisioning boundary (2026-07-24)

The missing headless/cloud install mechanism is a **family delivery
prerequisite, not a Trellis transport implementation**. Trellis owns its
product-side plugin package, installed overlay, host hook contract, fallback,
and the live proof required before it claims a surface. Stewards owns the thin
cross-product mechanism that puts the selected family plugins into ephemeral
Claude/Codex homes before either agent starts. The proposed authoring skill,
per-job GitHub Actions provisioner, container setup/maintenance path, trust
boundary, and Cloud experiment now live in
[kodhama/stewards#13](https://github.com/kodhama/stewards/issues/13).

One clean-consumer probe found a concrete prerequisite behind that boundary:
the public Stewards Claude marketplace exposes Trellis, while its native Codex
catalog currently exposes only Grove. A fresh isolated Codex home could add the
Stewards marketplace but could not install `trellis@kodhama`. A generated local
Codex marketplace pointing at the production `kodhama/trellis`,
`plugins/trellis` source then installed Trellis's manifest and hook
successfully. **Verified** locally with Codex CLI `0.145.0`; this isolates the
current failure to catalog/provisioning rather than Trellis packaging.

Accordingly, a portable headless or Cloud promotion requires either the
Stewards provisioner in issue 13 or an equivalent independently verified
pre-provisioned environment. Trellis may test its hook contract inside that
environment, but it does not own workflow discovery, multi-plugin profiles,
Claude/Codex home creation, marketplace updates, or container bootstrap. Codex
Cloud remains unverified even after installation: provisioning a separate CLI
does not prove that the hosted agent loads the same plugin state or trusts its
hook.

## Open questions

- **Does `@import` genuinely resolve absolute paths outside the project?** Docs suggest
  yes; not empirically tested by the research this note is based on. Needs a real check
  before any design leans on it (unlikely to be load-bearing given the `${CLAUDE_PLUGIN_ROOT}`
  finding above, but not yet ruled out as a partial mechanism).
- **Is there a simpler built-in mechanism this research missed?** Not found, but the
  search was one dispatch, not exhaustive.
- **How does `${CLAUDE_PLUGIN_ROOT}` behave in non-JSON, non-hook contexts** — e.g. a
  `SKILL.md` body — as opposed to the confirmed-broken command-markdown case? Unconfirmed
  either way.
- **Would a hook-based rewrite keep or drop `decision-0053`'s live-rows model?** A hook
  could compute the row-filtered readout mechanically at session start, instead of
  shipping the complete readout and relying on the model to respect an authority header.
  `research-0012`'s result (zero measured leak at n=20) means this wouldn't be *fixing* a
  known problem — it would only be a cleaner mechanism, not a correctness gain the
  existing data calls for.
- **Does the headless/subagent gap (see counter-example above) mean this is only ever a
  partial replacement**, with vendoring/inlining kept permanently for non-interactive
  invocations? Genuinely unresolved — the eval harness is existing, concrete evidence
  that the gap is real today, not hypothetical. Note this gap is specific to *model 2*
  (plugin-native) — model 1 (vendored) has no such gap, which is part of why model 1
  stays as the fallback rather than being retired.
- **Which other harnesses' native mechanisms support always-on instruction delivery
  without vendoring?** Surveyed so far: Claude Code — yes, via plugin + `SessionStart`
  hook (doc-sourced, not source-verified); OpenCode — its plugin system, no; its
  `instructions` config field, yes, including remote-URL fetch (source-verified);
  Codex — plugin hook contract, packaging, and one local live startup yes;
  broader lifecycle and surfaces remain unverified (bounded prototype above).
  The rest of `research-0010`'s registry
  (Copilot, Gemini CLI, Cline, Devin/Cascade, Windsurf, Cursor, Continue.dev, Aider)
  is unsurveyed — deferred until multi-harness support is active again (see Scope
  section above). Worth checking each harness for both shapes found so far
  (hook-based and static-config-with-remote-fetch), not just one.
- **Does Claude Code have an OpenCode-`instructions`-equivalent this survey missed** —
  a static config field with native remote-URL fetch, checked against source rather than
  docs? The doc-vs-source discrepancy found for OpenCode is reason enough to re-verify
  the Claude Code `@import`/`${CLAUDE_PLUGIN_ROOT}` finding (`anthropics/claude-code#9354`)
  against source before this note's Claude Code side is treated as equally solid as its
  OpenCode side.
- **Is OpenCode's `instructions` field (global-config form) actually the better near-term
  fit** given the current single-user reality — one config line, no per-project
  vendoring, no hook/plugin machinery at all — worth prototyping before any Claude-Code-
  specific hook work, if OpenCode ever becomes a harness trellis targets. Not decided
  here; the maintainer's near-term focus is stated as Claude Code only (Scope section).

## Sources & confidence

- Claude Code docs (`SessionStart` hooks, `additionalContext` behavior, plugin
  `agents`/`skills` discovery) — **High** (current official docs, via `claude-code-guide`).
- Claude Code live hook execution — **unverified**: the empirical probe is
  deferred until the maintainer's Claude usage limit resets; no negative result
  should be inferred from the absence of a run.
- `anthropics/claude-code#9354` (`${CLAUDE_PLUGIN_ROOT}` not substituted in `@import`) —
  **High** (confirmed open issue, not a workaround-documented closed one).
- Absolute-path `@import` resolution outside the project — **Low**, not empirically
  tested (named above as an open question).
- The `does-trellis-help` counter-example — **High** (in-repo, read directly:
  `eval/experiments/does-trellis-help/run.sh` lines 72–85).
- OpenCode plugin system (npm/Bun distribution, hook surface) — **High** (official docs,
  `opencode.ai/docs/plugins/`, fetched directly 2026-07-21; corroborates the maintainer's
  own recollection rather than resting on it alone).
- OpenCode `instructions` config field, including remote-URL fetch — **High**, source-
  verified (`packages/opencode/src/session/instruction.ts`,
  `anomalyco/opencode@dev`, fetched 2026-07-21) after two doc-page fetches
  (`opencode.ai/docs/rules/` vs `opencode.ai/docs/config/`) gave contradictory answers on
  remote-URL support — source settled it in favor of "yes."
- `anomalyco/opencode#4758` (`instructions` field silently not loading) — **High**
  (GitHub issue, confirmed closed/completed 2026-03-26 via `gh issue view`) — a real,
  now-apparently-fixed reliability caveat on the mechanism above, not a live blocker.
- Codex hook/plugin contract — **High** (current official docs,
  `developers.openai.com/codex/hooks` and `/codex/plugins/build`, checked
  2026-07-23).
- Codex local prototype — **High** for plugin install and direct hook-contract
  behavior; **medium-high** for maintainer-observed live startup delivery;
  **unverified** for other live lifecycle events and surfaces
  (`eval/experiments/codex-hook-delivery/`, Codex CLI `0.145.0`).
- Headless/cloud provisioning boundary and proposed shared primitive —
  **High for the observed catalog/install behavior; design proposal otherwise**
  ([kodhama/stewards#13](https://github.com/kodhama/stewards/issues/13)).

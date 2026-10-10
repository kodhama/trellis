---
title: Trellis delivers its rules as committed files - Plan
type: refactor
date: 2026-10-09
topic: trellis-delivery-direction
artifact_contract: ce-unified-plan/v1
product_contract_source: ce-brainstorm
execution: code
---

# Trellis delivers its rules as committed files - Plan

## Goal Capsule

- **Objective:** Every session in a repo that adopted Trellis starts with the same rules, wherever it runs: the maintainer's laptop, claude.ai/code cloud, Codex or Droid. A repo adopts, updates and leaves Trellis through one explicit command whose result is a diff someone reviews.
- **Means:** A Trellis CLI, run as `npx github:kodhama/trellis#v<version> install|update|check|remove`, writes the rules into files the repo commits. The Claude plugin, both session-start hooks, `install.sh` and decision-0073's closed set of seven delivery states retire after one migration round.
- **Product authority:** the maintainer (`gundisalwa`), through the Simplify Trellis project, issue TRL-107. The requirements were confirmed in dialogue on 2026-10-09, and merging this file is the approval. The Kodhama family install convention, the machine and cloud setup for skill-type tools, and ogham's U8 change are not active scope (see How This Work Fits Together).
- **Open blockers:** none for planning.

---

## Product Contract

### Summary

Trellis stops delivering rules through a plugin and a session-start hook. A repo adopts Trellis by running the Trellis CLI once. The CLI writes the rules, minus the repo's opt-outs, into committed files that every host reads at session start. Updating means running the CLI again and reviewing the diff. Existing consumers move over in one round of PRs, and the plugin then retires.

### Problem Frame

Trellis reaches a session today through a plugin whose hook decides at session start what to inject. Four costs follow from that.

**The rules a session sees depend on the machine, not the repo.** On 2026-10-09 the repo's HEAD was plugin 0.24.2, while this machine's installed copies were 0.22.0 or 0.23.1 (TRL-5 triage comment). `installed_plugins.json` holds 113 project-scope Trellis records at 0.22.0 and 31 at 0.11.0.

**Cloud and Codex sessions are second-class.** A claude.ai/code session installs no plugin a repo declares in `enabledPlugins`, so it gets Trellis only when an environment setup script installs it (`plugins/trellis/README.md:62-71`). Codex support is "not supported yet" (`README.md:135-142`). A committed `.claude/rules/trellis.md`, by contrast, loaded on turn one in cloud (`README.md:71-72`).

**Every component re-derives the delivery state.** decision-0073 found five components each holding "a private, partial enumeration" of the seven states S0–S6 and made itself their single home. ogham's plan unit U8 was about to become a sixth copy: as planned, it probes S0–S6 itself and writes `.trellis/rules.toml` and `.claude/settings.json` entries (kodhama/ogham PR #1, KTD6).

**The machinery outweighs the rules.** `staleness.sh` is 1,526 lines, `codex-context.mjs` is 1,366 lines and `install.sh` is 1,136 lines. About 11.6k test lines pin them, and about 18 decision records own them. The payload they deliver, `reference/rules.md`, is 6,070 bytes.

The trigger in the Ideas entry "Simplify Trellis delivery" fired when ogham's plan was approved (D17, 2026-10-08). It names the comparison directly: ogham chose a vendored kit with no hook so that "a cloud session sees exactly what the laptop sees" (KOD-5).

### Key Decisions

- **Vendored delivery: the repo holds the rules, and updating means re-running the CLI and reviewing the diff.** The same files reach every host. The cost is that a repo stays behind until someone updates it. (session-settled: user-approved — chosen over plugin-only delivery with automatic updates: that keeps the machine lag and the cloud gap.) Governs R2, R5, R8, R11.
- **Trellis keeps no plugin; the CLI is its only delivery path.** Rules must be in context from turn one, and a skill loads only when an agent calls it, so a plugin would add a channel without adding a capability. (session-settled: user-approved — chosen over a hookless plugin whose commands wrap the CLI, and over the full impeccable.style model where a plugin also delivers rules live.) Governs R11, R16.
- **Tools split by payload type.** Rule-type tools (Trellis, ogham) are vendored per repo by their own CLI. Skill-type tools (compound-engineering, grove, impeccable's skill) are installed once per machine at user scope and once in the cloud Default environment, and each repo also declares them in its committed `.claude/settings.json` so the repo is self-sufficient in its tooling. The declaration alone installs nothing in cloud, so it supplements the machine and cloud installs and does not replace them. This plan covers only Trellis's side of that split. (session-settled: user-approved — chosen over a family "stack bootstrap" product that runs every tool's installer: a cloud session never installs repo-declared plugins, and committed ids rot unnoticed, as design-system's `main` still enables `grove@kodhama`, which the kodhama marketplace no longer lists. The per-repo declaration is user-directed, 2026-10-10, relayed by majordomo: "I think the plugin should be declared in the repo, so the repo is self-sufficient in its tooling.")
- **The family install convention lives as a doc in the kodhama repo,** and Trellis conforms to it. (session-settled: user-directed — chosen over a Kodhama Linear doc and over stating it in this plan.)
- **One migration round, with named exemptions.** (session-settled: user-approved — chosen over migrating each repo when it is next touched, which would keep two delivery paths alive for an open-ended time.) Governs R14, R15, R16.
- **`check` guards integrity, not freshness.** Failing CI on every new release would break each adopted repo's main branch on every release; there were 26 VERSION bumps between 2026-07-24 and 2026-10-09. (session-settled: user-approved — confirmed with the scoping synthesis.) Governs R12, R13, R21.
- **`governed = false` remains the one recorded decline.** Without a record, a nudge or another tool's bootstrap cannot tell a declined repo from one never asked. (session-settled: user-approved — confirmed with the scoping synthesis.) Governs R3.
- **`remove` leaves a decline behind.** A repo that removed Trellis said no, and decision-0077 holds that silence is not adoption. Without the record, the next tool that calls `install` would re-adopt the repo. ogham asked for this ("never re-seeds over a removal"). Governs R6.
- **The CLI is the only code that knows Trellis's delivery states.** Every other caller asks the CLI and does not enumerate states itself, which closes the root cause decision-0073 named. Governs R3, R18.
- **The invocation names the GitHub repo, not an npm package.** The npm name `trellis` belongs to an unrelated publisher (turtle.tech, an "Agentic State Engine"), and `@kodhama/trellis` does not exist; checked 2026-10-09. The `github:` form is the one ogham uses. Governs R1.

### Requirements

**The CLI and adoption**

- R1. One CLI, invoked as `npx github:kodhama/trellis#v<version> <verb>`, provides `install`, `update`, `check` and `remove`. Each verb runs with no prompt and ends with one stdout line and an exit code that tell apart installed, already present, declined, migrated and error.
- R2. In a repo with no Trellis state, `install` writes the rules into committed files and records the installed version in a lock. It also says what it adopted. Running `install` is the adoption act; decision-0070 D2 already holds that "running an installer inside a repository is an unambiguous adoption act".
- R3. `install` changes nothing in a repo that already carries any Trellis state or a recorded decline, and it reports which one it found. The recorded declines are `governed = false` in `.trellis/rules.toml` and `"trellis@kodhama": false` in `enabledPlugins`.
- R4. `install` does not refuse because other files in the working tree are untracked or uncommitted.
- R5. `update` re-renders the rules from the requested version, keeps the repo's opt-outs, and leaves the change as a diff for review.
- R6. `remove` deletes every file and section Trellis wrote and records a decline, so no later `install` call re-adopts the repo silently.
- R7. The rendered rules leave out every rule the repo switched off, under the opt-out semantics of `docs/decisions/2026-09-15-rule-rows-only-switch-rules-off.md`. A change to the opt-outs takes effect at the next `update`.

**What a session sees**

- R8. Rules reach Claude Code (laptop and cloud), Codex and Droid from the repo's committed files alone, from the first turn, with nothing installed on the machine or in the cloud environment.
- R9. Each host loads the rules exactly once. A repo whose `CLAUDE.md` imports `AGENTS.md` does not also get a separate Claude copy, and a repo nested inside another adopted repo, such as the repos under `~/hq`, does not inherit a second copy from its parent.
- R10. For Codex and Droid, which cannot import a file from `AGENTS.md`, the rules are inlined in one marked section of `AGENTS.md`. That section sits beside other tools' marked sections and never wraps them. It stays near today's inline block (7,065 bytes) so it fits within Codex's 32 KiB `AGENTS.md` limit alongside ogham's inline core (ogham D20).
- R11. No hook or startup script runs in any session for Trellis.

**Integrity**

- R12. `check` runs offline. It fails when the committed Trellis files differ from what the lock records as written, as a hand edit or a partial write would cause, and when the opt-outs in `.trellis/rules.toml` differ from the opt-outs the lock records as rendered, naming the rows. The existence of a newer release never fails it.
- R13. In this repo, the committed Trellis files match what HEAD's own payload renders, and CI fails when they do not. This closes TRL-5.
- R21. Release drift stays visible without failing CI: each release reaches every adopted repo as an update PR opened by the update routine. That PR replaces the session-start staleness nudge as the place where drift is made visible (floor-transparency).

**Migration and retirement**

- R14. An explicit migration mode, separate from plain `install`, converts a repo in any of decision-0073's states S0–S6 to the vendored form. It keeps the opt-out rows and removes the old overlay, inline block, rendered file or bundle, as well as the `trellis@kodhama` plugin entry.
- R15. One round of PRs migrates every live consumer in the inventory below. wisp and review-kit are named exemptions: TRL-104's resume brief records that both are being retired, and they keep their current state until they go. Each migration PR also wires `check` into the repo's CI where the repo has CI, and rewrites the `rules.toml` header comment that says rows govern rule activation live.
- R16. After the round, and once the update routine (R21) and the nudge line exist, the plugin ships one last release and then retires. That last release injects nothing in a repo that carries the Trellis lock; anywhere else it says the plugin is retired and gives the CLI command. Retiring it removes the marketplace entry, `staleness.sh`, `codex-context.mjs`, `install.sh` and their tests. A decision record written by the build replaces decision-0073's closed set.
- R17. A consumer outside the inventory that still runs the plugin keeps working until the plugin retires. A migrated repo does not load the rules twice on a machine whose install runs R16's last release; on a machine with an older install, R19's uninstall is what prevents it.
- R19. The round also ends machine-level delivery on each machine the inventory covers: it uninstalls `trellis@kodhama` at user scope from Claude Code and from Codex, and removes the Trellis lines from cloud environment setup scripts. Until then, a user-scope install keeps injecting rules and advice into migrated repos.

**Callers**

- R18. ogham's `init` and any other tool adopt Trellis only by calling `install` pinned to a version tag, and they read its outcome from R1's line and exit code.
- R20. Every Trellis release creates the tag `v<version>` on its merge commit on `main`, and CI fails a release whose tag is missing. `v*` tags cannot be created, moved or deleted outside that path, so a pin always names reviewed, merged code.

Consumer inventory for R15, found by scanning `~/hq` and `~/projects` on 2026-10-09:

| Repo | State today | R15 |
|---|---|---|
| design-system | S2 overlay + S4 block in `CLAUDE.md`, plugin declared | migrate |
| spore | S2 overlay + S4 blocks in `CLAUDE.md` and `AGENTS.md` | migrate |
| mq-source | S3 flat overlay + S4 block in `CLAUDE.md`, no `rules.toml` | migrate |
| kodhama, ogham, stewards, trellis (this repo), math-quest, math-quest-bmad, math-quest-openspec, math-quest-narrative, math-quest-supervision | S0: `rules.toml` + plugin declared | migrate |
| `~/hq`, machine-skills | `rules.toml` only | migrate |
| wisp, review-kit | S0, being retired | exempt |

The scan covered one machine. Consumer repos elsewhere, such as a consultant-mode company repo, are not listed and fall under R17.

### Acceptance Examples

- AE1. **Covers R1, R2.** **Given** a repo with no Trellis state, **when** `install` runs, **then** the rules files and the lock are written, the stdout line says installed, and the exit code is success.
- AE2. **Covers R3.** **Given** a repo whose `rules.toml` says `governed = false`, **when** `install` runs, **then** no file changes and the stdout line says declined.
- AE3. **Covers R3, R14.** **Given** spore (S2 overlay plus S4 blocks), **when** plain `install` runs, **then** nothing changes and it reports the states it found. **When** the migration mode runs instead, **then** spore ends in the vendored form, its opt-out rows are unchanged, and the overlay, both blocks and the plugin entry are gone.
- AE4. **Covers R4, R18.** **Given** ogham's `init` has just written untracked `.ogham/` files, **when** it calls `install`, **then** Trellis installs and ogham reads "installed" from the stdout line.
- AE5. **Covers R5, R12.** **Given** a repo installed at v1 when v2 is released, **when** `check` runs, **then** it passes. **When** `update` runs to v2, **then** the change appears as a diff and `check` passes against the new lock.
- AE6. **Covers R12.** **Given** someone hand-edits the rendered rules, **when** `check` runs, **then** it fails and names the file.
- AE7. **Covers R8.** **Given** a migrated repo and a cloud environment with no setup script, **when** a claude.ai/code session starts, **then** the rules are in context on turn one.
- AE8. **Covers R9.** **Given** a repo whose `CLAUDE.md` is exactly `@AGENTS.md`, **when** a Claude session starts, **then** the rules appear once.
- AE9. **Covers R6, R3.** **Given** `remove` has run in a repo, **when** ogham's `init` later calls `install`, **then** the stdout line says declined and nothing is re-adopted.
- AE10. **Covers R12.** **Given** someone commits `<slug> = { active = false }` to `rules.toml` without running `update`, **when** `check` runs, **then** it fails and names that row.
- AE11. **Covers R19, R9.** **Given** a migrated repo on a machine where the round ran, **when** a Codex session starts, **then** no Trellis hook runs and the rules appear once, from `AGENTS.md`.

<!-- ce-section: work-relationships -->
### How This Work Fits Together

This plan covers Trellis's own delivery. The breakdown below is the current understanding, not a committed roadmap.

- **Kodhama family install convention** (a doc in the kodhama repo; majordomo routes it): rule-type tools vendor per repo through their CLI, and skill-type tools install per machine and in the cloud Default environment and are declared in each repo's committed settings. *Shares* the CLI shape with this plan; this plan conforms to it.
  - **Machine and cloud setup for skill-type tools** (compound-engineering, grove, impeccable's skill). *Can proceed independently of* this plan.
  - **The nudge**: one line in the maintainer's global agent instructions that offers adoption in a repo without Trellis or ogham, asked through the question dialog and never acted on headless. *Depends on* R1 and R3. *Enables* R16: it replaces the plugin's in-session adoption offer, so the plugin does not retire before it exists.
  - **An update routine** that runs `update` across adopted repos and opens one PR per repo. *Depends on* R1 and R5. *Enables* R16 and R21.
- **ogham U8** (ogham D19, replacing KTD6's own probe): *depends on* this plan's CLI being tagged.
- **TRL-5** (this repo does not dogfood HEAD): closed by R13.
- **Idea 5, "One delivery path"** (listed in TRL-104's resume brief; `docs/ideation/2026-09-14-repo-simplification-ideation.html`): this plan takes its place. TRL-104's other ideas are untouched. It keeps that idea's migration round and its constraint that adoption stays explicit, and it replaces "plugin-native only" with vendored files.

### Scope Boundaries

- The family convention doc, the skill-tool setup script, the nudge line and the update routine (see How This Work Fits Together). R16 waits on the last two, but building them is not this plan's work.
- ogham's U8 change and the superseding entry ogham records against its D16 and KTD6.
- A decline marker for a client repo that should not commit any Trellis file. Deferred: `governed = false` is committed, and no client repo is in the inventory.
- Building the CLI. This plan stops at requirements.

### Dependencies / Assumptions

- **Assumption, measured once:** a committed `.claude/rules/*.md` file loads on turn one in a claude.ai/code session (`README.md:71-72`). R8 rests on it.
- **Assumption, not measured:** Codex and Droid follow rules inlined in `AGENTS.md` about as well as Claude follows its rules file.
- **A decision record is due with the build, not with this plan.** Retiring decision-0073's closed set, the plugin path of decision-0065 and decision-0071's project-scope plugin declaration meets the record test, because a wrong call could silently stop rules reaching sessions.

### Outstanding Questions

**Deferred to Planning**

- What the CLI is built in and how `npx github:` runs it without a Go toolchain on the caller's machine.
- Where the Claude copy lives: a rules file, or only the `AGENTS.md` section when `CLAUDE.md` imports `AGENTS.md` (R9). Hooks already shipped constrain the choice until R19 runs: `staleness.sh` refuses on a column-0 `trellis:begin` block and advises deleting it (`:754-755`), and it checks a rendered rules file's markers (`:626-710`).
- The lock's format, and the exact mapping from outcomes to exit codes (R1).
- Whether migration is a flag on `install` or a verb of its own (R14).
- How to measure Codex adherence to the inlined section before claiming parity.

### Sources / Research

- `docs/decisions/0073-the-delivery-shapes-are-a-closed-set.md`: the S0–S6 table (lines 108-130) and the "private, partial enumeration" diagnosis (lines 17-25).
- `docs/decisions/0070-adoption-is-the-consent-act-not-installation.md` D1 and D2 (running an installer in the repo is the adoption act); `0071` (this repo self-applies through the released plugin); `0077` (silence is not an adoption act).
- `plugins/trellis/README.md:62-72` (cloud and plugin reach), `:135-148` (Codex), `:75` (the hook stands down for a rendered file).
- `docs/ideation/2026-09-14-repo-simplification-ideation.html:295-348` (Idea 5).
- kodhama/ogham PR #1, `docs/plans/2026-10-08-0641-feat-ogham-coding-standards-companion-plan.md`: KTD4 (`npx github:` invocation), KTD6 and U8.
- Linear: TRL-107, TRL-5 (triage comment of 2026-10-09), TRL-104, KOD-5, and the Trellis Ideas entry "Simplify Trellis delivery".
- impeccable.style (`pbakaus/impeccable` at `d631a88`): one source compiled per harness, with a CLI as the main install path. Its use case is skills rather than always-on rules, which is why only its CLI shape carries over.
- Claude Code cloud environments docs: "A cloud session doesn't install the plugins a repository turns on under `enabledPlugins`."
- OpenAI Codex `AGENTS.md` docs: no include or import mechanism; 32 KiB `project_doc_max_bytes`.

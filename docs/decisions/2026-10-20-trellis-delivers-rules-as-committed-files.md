---
id: decision-2026-10-20-trellis-delivers-rules-as-committed-files
type: decision
depends_on: [decision-0060, decision-0061, decision-0069, decision-0070, decision-0073, decision-0077, decision-2026-09-15-rule-rows-only-switch-rules-off]  # coupling under decision-0047's test. 0073 is superseded in part by point 5, whose set keeps 0073 D1's table as the definition of S0–S6. 0060 §2 requires a new project-local decision for any release choice, which point 11 is. 0061 §1 makes plugins/trellis/VERSION the version authority that point 11's single version line keeps until the plugin retires, and point 11 lifts the parts of 0061 it names. 0069 D1 requires a record that prunes the manual copy path to supersede it, which point 13's last bullet answers. 0070 D2 is why running install is the adoption act (point 1). 0077 is why remove leaves a decline behind (point 9). The rule-rows record is the opt-out semantics the rendered rules follow (point 9)
changes: [decision-0073]
informed_by: [decision-0028, decision-0065, decision-0071]
owner: agent
date: 2026-10-20
---

> **Provenance.** TRL-122 writes the record that unit U4 of the vendored-delivery build plan calls
> for (TRL-113; `docs/plans/2026-10-09-1945-refactor-trellis-delivery-direction-plan.md`, merged in
> PR #339). The maintainer confirmed the plan's requirements in dialogue on 2026-10-09. On
> 2026-10-18 he ruled that the CLI is committed, zero-dependency Node ESM, that migration refuses a
> morphed repo, and that the plugin ends with "notice, then stop". On 2026-10-19 he ruled that release
> tags start at the CLI's first release, and that the guarantee of no Trellis hook covers only the
> repos the migration round migrated. Those rulings reached the author as relayed by the supervising
> session `trl-cli-lead`, and are recorded here as relayed, not quoted. Unit names (U1–U21) are the
> plan's.

# Trellis delivers its rules as committed files

## Context

Trellis reaches a session today through a plugin. Its session-start hooks,
`plugins/trellis/hooks/staleness.sh` on Claude Code and `plugins/trellis/hooks/codex-context.mjs` on
Codex, decide at each session start which rules to inject. `install.sh` vendors a copy for repos the
plugin does not cover, and `decision-0073` names the seven delivery states (S0–S6) every component
must handle. The plan measured four costs on 2026-10-09:

- **The rules a session sees depend on the machine, not the repo.** The repo's HEAD was plugin 0.24.2,
  while the machine's installed copies were 0.22.0 or 0.23.1, and `installed_plugins.json` held 113
  project-scope Trellis records at 0.22.0 and 31 at 0.11.0.
- **Cloud and Codex sessions are second-class.** A claude.ai/code session installs no plugin that a
  repository enables, so it gets Trellis only through an environment setup script, and Codex support
  was "not supported yet". A committed `.claude/rules/` file, by contrast, loaded on turn one in a
  cloud session.
- **Every component re-derives the delivery state.** `decision-0073` found that "at least five
  components each hold their own private, partial enumeration" of the states, and ogham's plan was
  about to add a sixth.
- **The machinery outweighs the rules.** `staleness.sh`, `codex-context.mjs` and `install.sh` ran to
  1,526, 1,366 and 1,136 lines, pinned by about 11.6k test lines, to deliver a 6,070-byte
  `reference/rules.md`.

Three records ask for a record like this one. `decision-0073` D1: "Adding a state is a decision
superseding this one in part — never a code change alone." `decision-0060` §2: "A future Trellis
package-version or release choice, if one becomes necessary, requires a new project-local decision on
its own evidence." `decision-0069` D1, of the manual copy path and `staleness.sh`'s path A: "a future
prune that deletes either must supersede this record".

## Decision

**1. Vendored delivery: the repo holds the rules, and the Trellis CLI is the only way they get
there.** A repo adopts, updates and leaves Trellis with
`npx -y github:kodhama/trellis#v<version> <verb>`, whose verbs are `install`, `update`, `check` and
`remove`. None of them prompts. `install` writes the rules into files the repo commits, leaving out the
rules the repo switched off. Claude Code (laptop and cloud), Codex and Droid read the rules from those
files alone, from the first turn, with nothing installed on the machine or in the cloud environment.
Running `install` is the adoption act, as `decision-0070` D2 holds for any installer run in a repo.

- **Updating means running `update` and reviewing the diff.** A repo stays on its version until someone
  updates it. That cost is accepted, and the per-repo update PRs of the Kodhama update routine (KOD-7)
  are where drift becomes visible.
- **Trellis keeps no plugin.** The rules must be in context from turn one, and a skill loads only when
  an agent calls it, so a plugin would add a channel without adding a capability.
- **The invocation names the GitHub repository**, because the npm name `trellis` belongs to another
  publisher and `@kodhama/trellis` does not exist (checked 2026-10-09).

**2. The CLI is committed, zero-dependency Node ESM, and Go stays a build-time generator.**
`bin/trellis.mjs` and `lib/*.mjs` use Node built-ins only and need no build step. The root
`package.json` is the CLI's package: it stays `"private": true`, has no `prepare` script, and lists
only `bin/`, `lib/` and `payload/` in `files`, so `npx` packs the CLI and its payload rather than the
whole tree. The Go `payload` command keeps rendering every piece of prose into
a top-level `payload/`, so the JavaScript filters, assembles and writes, and words nothing of its own.
Rejected: Go binaries fetched from release assets by a JavaScript shim, and TypeScript built by
`prepare`, which npm runs on every first git install.

**3. The files Trellis owns.**

- **`AGENTS.md` carries the rules in one section** between `<!-- trellis:vendored:begin -->` and
  `<!-- trellis:vendored:end -->`, beside other tools' marked sections. It never wraps them and never
  sits inside them. The rules are inlined because Codex and Droid cannot import a file from
  `AGENTS.md`. `install`, `update` and `check` warn on stderr when an `AGENTS.md` chain Codex reads
  passes its 32 KiB cap.
- **The Claude copy**, where one is needed, is `.claude/rules/trellis-rules.md` (point 4).
- **`.trellis/invariants.md`** holds the catalog entries the rules point to for detail. It is committed
  and read on demand, never loaded at session start.
- **`.trellis/rules.toml`** holds the opt-outs. `install` creates it when absent.
- **`.trellis/lock.json`** is the lock (point 7).

The marker and the filename keep an old plugin install from advising their deletion: `staleness.sh`
advises deleting a block that opens `<!-- trellis:begin` at column 0, and a
`.claude/rules/trellis.md` that fails its landmark walk. Such an install sees an ordinary S0 repo and
injects the rules a second time. That is accepted until the machine's install is removed or runs the
plugin's last delivering release (point 13).

**4. Each host loads the rules once.** Where the Claude copy goes is decided again at every `install`
and `update`, and the lock records which case applied:

1. `CLAUDE.md` imports `AGENTS.md`, through a line that is exactly `@AGENTS.md` or `@./AGENTS.md`
   outside a code fence, or by being a symlink to `AGENTS.md`: no Claude copy.
2. There is no `CLAUDE.md`: the CLI creates it as the single line `@AGENTS.md`, as ogham does.
   `update` keeps this case, recorded as `created-claude-md`, while `CLAUDE.md` is still exactly
   `@AGENTS.md`, so `remove` can still delete it (point 9).
3. A `CLAUDE.md` exists without the import: the CLI writes `.claude/rules/trellis-rules.md`, which
   loads on turn one in cloud sessions.
4. `install --no-claude-copy` writes no Claude copy, and `update` keeps that choice. It is for a repo
   that contains other adopted repos and whose own sessions get the rules another way, as `~/hq`'s
   do: majordomo reads `AGENTS.md` there by instruction. Claude Code loads a parent directory's
   `CLAUDE.md` and `.claude/rules/` into a session started in a nested repo (measured on Claude Code
   2.1.302, 2026-10-18), so any Claude copy in such a repo would reach every repo under it.

`CLAUDE.local.md` and `.claude/CLAUDE.md` never count as the import. `check` fails when the recorded
case no longer matches the files, as when someone adds `@AGENTS.md` beside an existing rules file.

**5. The CLI holds the one enumeration of delivery states, and this record states the set.** This
supersedes `decision-0073` D1 in part. The parts the probe looks for are of two kinds:

- **Old parts:** S1's `.claude/rules/trellis.md`, S2's `.trellis/internal/`, S3's
  `.trellis/trellis.md` or `.trellis/version`, an S4 block, S5's `.claude/skills/trellis/`, the
  codex-bootstrap block, and a Trellis plugin entry, `true` or `false`. The codex-bootstrap block is
  the `AGENTS.md` block opening `<!-- trellis:codex-bootstrap:begin`, which `cli/apply.go` renders
  into the payload and D1's table omits.
- **New parts:** a new artifact (the vendored section, `.claude/rules/trellis-rules.md` or
  `.trellis/invariants.md`), a lock, and the `governed = false` line.

`.trellis/rules.toml` is neither: alone it is D1's config-only S0, and beside a new part it belongs to
the vendored or declined form. The probe gives a repo the first of these states that matches, so every
repo is in exactly one:

1. **S6** when any S6 marker is present, with every other part it finds named as evidence, so no verb
   removes `.trellis/rollback`.
2. **mixed** when an old part sits beside a new part, which is what a migration that crashed before its
   removal step leaves. This is one of the "conflicting combination[s]" D1 requires every reader to
   name, and the probe names both parts.
3. **S0–S5 or codex-bootstrap**, as D1's table defines S0–S5, when old parts or a bare `rules.toml`
   are present and no new part is. The probe names each state whose parts it finds.
4. **`declined`** when `.trellis/rules.toml` holds `governed = false`, whatever other new parts
   remain. D1's table reads such a repo as S0.
5. **`vendored`** when a lock is present and every file it names exists.
6. **`vendored-incomplete`** when a new artifact is present without a lock, or a lock names a missing
   file.
7. **`none`** when nothing Trellis wrote is present, which is D1's S0-unadopted state.

The probe is a pure function in the CLI's `lib/states.mjs`. It takes the repo's files and whether the
`trellis-pre-morph` tag exists, and returns a state plus the evidence that decided it. The tag's
presence is passed in as an input, so the function stays pure. A tag with no `.trellis/rollback`
reads as S6, because the probe cannot tell a stale tag from a standing morph whose rollback file is
gone. **No new component enumerates delivery states**: every new caller, ogham's
`install` included, asks the CLI. The plugin's components keep their own scoped enumerations, under
D1's relevance clause, until each retires (point 13). The last delivering release adds only the lock
to them, so their hooks stand down in a vendored repo. D1's table stays the definition of S0–S6 and the
input migration reads until the plugin package retires. Adding a state is a new record
superseding this one in part, never a code change alone.

**6. The outcome contract, which callers pin.** Every run prints exactly one stdout line,
`trellis: <outcome> <detail>`, where `<outcome>` is one of `installed`, `already-present`,
`declined`, `migrated`, `pass`, `fail`, `done` or `error`. Every outcome exits 0 except `fail` and
`error`, which exit 1. A usage error prints `error` and exits 2. Warnings go to stderr, so stdout stays
one line.

- **`install` by state.** In `none` it reports `installed` and writes every file; only this case
  writes. In `vendored` it reports `already-present`, and in S0–S6 or codex-bootstrap
  `already-present <state>`. In `declined` it reports `declined`. In `vendored-incomplete` or mixed it
  reports `error` with "run install --migrate", so a caller never reads success from a half-written
  repo. The exception is a mixed repo whose new parts include `governed = false`, which reports
  `declined`.
- `install` never overrides a decline. Its `declined` line says how a person reverses one: delete the
  `governed = false` line, then run `install --migrate`.
- **`update` by state.** In `vendored` it reports `done` and re-renders the rules. In `none` it reports
  `error` with "run install", in `declined` `error` with "declined; see install", and in
  `vendored-incomplete` `error` with "run install --migrate".
- **`remove` by state.** In `vendored`, `vendored-incomplete` and `declined` it reports `done` and
  deletes what Trellis wrote (point 9). In `none` it reports `done`, with a note, and changes nothing.
- **S6 refuses every verb that would write.** On S6, `install --migrate`, `update` and `remove` change
  nothing and report `error`, telling the person to roll back the morph first.
- **Old state goes through migration.** In S0–S5, codex-bootstrap or a mixed state, `update` and
  `remove` change nothing and report `error` with "run install --migrate". In a mixed state the old
  part decides every verb but two: plain `install`, as above, and `install --migrate` when the new
  parts include `governed = false`, which reports `declined` (below). Otherwise `install --migrate`
  completes the conversion.
- Migrating a repo whose plugin entry is `false`, or a mixed repo whose new parts include
  `governed = false`, reports `declined`, not `migrated`, because nothing was adopted.

**Once the CLI's first release is tagged, this contract changes only through a new decision record and
a notice to ogham's lead**, because ogham's `install` pins a Trellis tag and parses the line.

**7. The lock, and how the CLI writes.**

- **The lock** is `.trellis/lock.json`: JSON with a 2-space indent and a trailing newline. Its fields
  are `schema` (1), `version`, `claude_copy` (`import`, `rules-file`, `created-claude-md` or `none`),
  `opt_outs` (the switched-off slugs as rendered), `files` (each wholly owned file mapped to its
  sha256) and `sections` (`AGENTS.md` mapped to the sha256 of its section, markers included). Hashes
  run over text with any BOM stripped and CRLF turned into LF. The lock records no commit, because a
  `v*` tag is never moved once created (point 11), so the version names the code. A CLI that meets a
  newer `schema` reports `error` and names the version that wrote the lock.
- **The working tree only.** The CLI reads file content from the working tree, never from HEAD or the
  git index, and it never changes the index. It never refuses because a file is untracked or
  uncommitted, `AGENTS.md` included, and it keeps other tools' bytes exactly. It finds the root with
  `git rev-parse --show-toplevel` and reports `error` outside a git repo.
- **Atomic writes in a fixed order.** Each file goes to a temporary file beside it and is renamed into
  place, in this order: `invariants.md`, the Claude copy, the `AGENTS.md` section, `rules.toml`, and
  the lock last. A crash therefore leaves no lock, a state the probe names and `install --migrate`
  completes. The run that completes it must repeat the original flags, `--no-claude-copy` included,
  because the lock that would have recorded them is written last.
- **Migration's order.** Migration writes the new artifacts first, then removes the old ones, then
  writes the lock. It makes one exception to the order above: where an old artifact holds opt-outs or a
  decline that no `rules.toml` holds yet (S3's overlay, an S4 block in a repo with no `rules.toml`, or a
  `false` plugin entry), it writes `rules.toml` first, before the section replaces any old block. So no
  old artifact goes before `rules.toml` carries what it held. When two runs overlap, the last writer
  wins.
- **Hand edits and broken markers.** A missing end marker, a duplicated section, or a section inside
  another tool's markers stops every verb with `error`, and nothing is written. `update` and `remove`
  act on Trellis-owned content even when it was hand-edited, warning on stderr and naming each file,
  so the reviewed diff shows the reverted edit.

**8. `check` guards integrity, not freshness.** `check` runs offline. It fails when a committed
Trellis file or section differs from what the lock records, naming each, and when the opt-outs in
`rules.toml` differ from those the lock records as rendered, naming the rows. A newer release never
fails it: there were 26 version bumps between 2026-07-24 and 2026-10-09, and failing on each would
break every adopted repo's `main`. Each migrated repo that has CI runs `check` there at the version
its lock records, so `update` never edits CI. With no lock, the CI step passes only when
`rules.toml` holds the top-level `governed = false` line, and otherwise fails with "Trellis lock
missing: run install --migrate or remove". A migration that ends `declined` wires no step.

**9. Opt-outs and the decline.**

- **Switched-off rules are left out.** The rendered rules omit every rule the repo switched off, under
  `decision-2026-09-15-rule-rows-only-switch-rules-off`'s semantics, and a change to `rules.toml` takes
  effect at the next `update`, not at session start. The rendered header says both, so an agent reads
  a missing rule as switched off.
- **`governed = false` in `.trellis/rules.toml` stays the one recorded decline.** A
  `"trellis@kodhama": false` plugin entry is not a decline: no shipped code reads it, and it goes with
  the plugin. The CLI writes `governed = false` as the single top-level line above the first table,
  keeping `seeded_from`, `strictness` and every row, the only form the old hook parsers read as a
  decline.
- **`remove` leaves a decline behind.** In a `vendored`, `vendored-incomplete` or `declined` repo, it
  deletes every file and section Trellis wrote except `rules.toml`, which it leaves holding
  `governed = false`. In `none` it reports `done` and changes nothing, and in S0–S6, codex-bootstrap
  or a mixed state it changes nothing and reports `error` (point 6). A `CLAUDE.md` that
  Trellis created is deleted only while it is still exactly `@AGENTS.md` and `AGENTS.md` is gone, so
  another tool's use of the import survives. A repo that removed Trellis said no, and `decision-0077`
  holds that silence is not adoption: without the record, the next tool that calls `install` would
  adopt the repo again.

**10. Migration is `install --migrate`, run once across the consumers.** It converts a repo in S0–S5
or codex-bootstrap to the vendored form. It keeps the opt-out rows, removes the old overlay, block,
rendered file or bundle, and removes every Trellis plugin entry: `trellis@kodhama` in
`.claude/settings.json` and `trellis@stewards` in a committed `.factory/settings.json`. A `false`
entry keeps its no: the run writes `governed = false`, removes the old artifacts and the entry, writes
no vendored file and no lock, and reports `declined`.

- **On S6 it changes nothing and reports `error`, telling the person to roll back the morph first.**
  The line adds that a morph already reversed leaves a stale tag, cleared with
  `git tag -d trellis-pre-morph`. The maintainer narrowed migration to S0–S5 on 2026-10-18.
- **Plain `install` never migrates.** The flag keeps the CLI at four verbs.
- **One round of PRs migrates every live consumer** in the plan's inventory, and each PR also wires
  `check` into the repo's CI where the repo has CI. wisp and review-kit are exempt because they are
  being retired, and keep their current state until they go. A consumer outside the inventory keeps
  working until the plugin retires.

**11. The release choice.** This is the new project-local release decision `decision-0060` §2 asks
for.

- **One version line.** `package.json`'s `version` and `plugins/trellis/VERSION` hold one number, and a
  parity test enforces it until the plugin package retires. The CLI's first release (U9) is v0.25.0,
  or the next unused minor if plugin releases land first.
- **Every CLI release is tagged `v<version>`, from the CLI's first release on. The plugin-only releases
  before it get no tag**, as the maintainer ruled on 2026-10-19, because no pin targets them. A
  workflow on pushes to `main` creates the tag on the commit that set the version, when that version has
  no tag, and a second job fails when any version `main` has carried has no tag. **Both skip every
  version below a floor committed in the workflow file.** U1 commits the floor unset, so neither job
  acts on any version, and U9 sets it to its own version.
- **Tags are protected and never moved.** The tag workflow only creates tags. A tag ruleset on
  `refs/tags/v*` lets only GitHub Actions create, move or delete them, and workflows change only
  through reviewed PRs to `main`, so a pin always names reviewed, merged code. The ruleset is a
  repository setting the maintainer applies. If GitHub Actions cannot be the ruleset's bypass actor,
  or if its bypass also lets a workflow on an unmerged branch create a tag, the bypass becomes a
  GitHub App whose key only `main` can use.
- **No reused names.** New tags never reuse a name from `v0.1.0`–`v0.2.30`, which name the retired v0
  CLI and are not a supported pin.
- **The release guard widens at the first release.** From U9, `release-guard` also counts a change to
  `bin/`, `lib/`, `payload/`, or a `package.json` field that changes what `npx github:` runs, as a
  release.

This record decides that parts of `decision-0061` stop applying to Trellis's own releases. The lift
takes effect with U1, which adds the tag workflow and the version parity test and notes that record.
The parts are the
"synchronized bump" and "tag workflow" of its Decision state's "No family release, synchronized bump,
tag workflow, release record, generated support table, history, approval ledger, or shared validator
returns"; the tag in §1's "It does not restore a release channel, auto-update mechanism, or tag"; and
the "immutable tag", "release workflow" and "synchronized version or bump policy" of §5's "Trellis
adds no" list. The rest of those lists stays ruled out, a family release, a release record and a
GitHub Release among them.

**12. This repo self-applies through the CLI.** Once migration exists (U10), this repo migrates itself
(U11) and carries the vendored files. Its CI runs `node bin/trellis.mjs check` from the checkout,
because HEAD's version has no tag until it merges, and a test fails when the committed files differ
from what HEAD's own payload renders. That closes TRL-5.

**13. The retirement order: notice, then stop.**

- **The round ends machine-level delivery too.** On each machine the inventory covers, it uninstalls
  `trellis@kodhama` at user scope from Claude Code and Codex, uninstalls `trellis@stewards` at user
  scope from Droid and removes it from untracked project `.factory/settings.json` files, and removes
  the Trellis lines from cloud environment setup scripts.
- **Then the plugin's last delivering release (U14)**, and only once the adoption nudge (KOD-6) and the
  update routine (KOD-7) exist. It injects nothing in a repo that carries the lock. Anywhere else it
  delivers the rules as before, says the plugin is retiring, and gives the CLI command. This is the
  maintainer's reading of 2026-10-18.
- **Then the releases that stop delivery.** They retire, one per release, `install.sh` (U16),
  `staleness.sh` (U15), which ends delivery on Claude Code, and `codex-context.mjs` (U18), after
  which no Trellis hook runs in any install that has updated to U18. U19 then deletes the plugin
  package, `VERSION` included, and the `trellis` entry leaves the kodhama marketplace last, in
  kodhama/stewards (U17).
- **On a machine where the round ran, no hook or startup script runs for Trellis in any session in a
  repo the round migrated.** The maintainer narrowed this guarantee to those repos on 2026-10-19.
  wisp and review-kit keep their hooks until they retire or a removal release takes the hooks away.
- **The manual copy path retires with `plugins/trellis/reference/` in U19.** This is the statement
  `decision-0069` D1 requires of a record that prunes that path.

**14. No claim of Codex or Droid parity.** Whether Codex and Droid follow rules inlined in `AGENTS.md`
about as well as Claude follows its rules file is unmeasured, and the build claims no parity. U9 adds
a Trellis Ideas entry whose trigger is "a reported Codex or Droid miss of an inlined rule, or any plan
to claim parity". That entry is what brings the measurement back.

## Consequences

- **This record is the successor for the delivery machinery's records.** Each unit that makes a point
  of a record untrue adds a dated note naming this record, and the unit that makes a record's last live
  point untrue adds `superseded_by` naming it. This change notes `decision-0073` D1. The plan's Record
  schedule assigns the rest to their units.
- **Two `AGENTS.md` rules become untrue during the build** and are rewritten where they do: that the
  invariants are "delivered live by the Trellis plugin at session start" (U11), and that a payload
  change is a release with its `plugins/trellis/VERSION` bump (U1, U9, U19).
- **ogham waits on the first tag.** Its `install` calls Trellis's pinned to a tag and parses point 6's
  line.
- **Some sessions get the rules twice during the round.** On a machine with an old install, a migrated
  repo's Claude sessions get the hook's injection as well as the committed copy, and its Codex sessions
  do too, because `codex-context.mjs` keys vendored mode on `.trellis/internal/` only. The round's
  machine-level step ends this per machine, and the last delivering release ends it on any machine
  whose install runs that release.
- **A consumer the round missed loses delivery at the retirement releases.** It stops receiving the
  rules from the plugin at U15 on Claude Code and at U18 on Codex, and the last delivering release's
  notice is its only warning. The notice reaches only an install that updates while a release that
  still ships the host's hook is the marketplace head: U14 or U16 on Claude Code, and U14, U16 or
  U15 on Codex.
- **Sessions see less advice.** A repo's sessions stop receiving the hook's per-session activation
  section and advice. Switched-off rules are absent instead of being named as inactive, and the
  rendered header says so.
- **Tag protection rests on a setting no PR carries.** Until the maintainer applies the ruleset, a
  hand-made `v*` tag could name unreviewed code. U1's done signal includes the ruleset.

## Open questions

- Whether GitHub Actions can be a tag ruleset's bypass actor at all, and if it can, whether its bypass
  lets a workflow on an unmerged branch create a `v*` tag. U1 (TRL-119) checks both, and point 11 says
  what follows in each case.
- Whether Claude Code still loads a parent's `CLAUDE.md` and `.claude/rules/` into a nested repo when
  `~/hq` migrates. Point 4's `~/hq` case rests on one measurement, and U12 measures again before
  migrating `~/hq`.
- Whether Codex and Droid follow inlined rules as well as Claude follows its rules file. Point 14's
  Ideas entry brings it back.

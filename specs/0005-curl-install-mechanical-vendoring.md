---
id: spec-0005
type: spec
status: gated
depends_on: [kodhama/kodhama-0007-one-render-many-copiers, decision-0043]
superseded_in_part_by: [decision-0066, decision-0068]  # 2026-07-29 decision-0066 — §1's bundle table at two sites: the surfaces.json row, and the version-parity clause in the VERSION row. 2026-07-30 decision-0068, in TWO amendments: (a) AC2's blanket "never … writes any instructions file", narrowed to permit the one rendered rules file; and (b) after review, AC2's blanket "never reads … .trellis/" AND its managed-block-detection prohibition, both narrowed to permit an EXISTENCE-ONLY read of static-delivery state (.trellis/internal/, the legacy flat .trellis/trellis.md, and a PAIRED trellis:begin/end region in CLAUDE.md/AGENTS.md) whose only outcome is render-or-refuse-loudly. What still stands untouched: the "zero decision logic" heading, the posture-prompt and marker-PATCHING prohibitions, and "never WRITES to .trellis/". §Scope and §1's scope note are amended in the same act.
owner: agent
rubric: rubric-artifact-contract
date: 2026-07-10
---

# Spec 0005 — The curl install path: `install.sh` as a mechanical vendoring copier

> **Forward pointer.** [`decision-0066`](../decisions/0066-retire-the-surface-matrix.md)
> retires `plugins/trellis/surfaces.json` and supersedes this spec in part, at
> **two sites in §1's bundle table**, both amended in place rather than left
> stale — §1 declares the table normative (*"AC1 depends on this table's left
> column being exhaustive for the actual `plugins/trellis/` tree at build
> time"*), so a row naming a file that no longer exists is a live spec defect,
> not archive.
>
> **What moved:** the `surfaces.json` row (a bundle-membership claim), and the
> version-parity clause inside the `VERSION` row, which read *"both host
> manifests **and `surfaces.json`** must match it"*. The second is a
> version-parity claim rather than a bundle-membership one and would be missed
> by a scope worded around the bundle alone. **What stands:** everything else,
> including AC1's exhaustiveness requirement itself — the bundle is now 25 files,
> and `TestInstallScriptBundleManifestIsCurrent` enforces that the manifest and
> the tree agree in the same commit.

> **This spec corrects kodhama/trellis#124's issue text; the issue should be edited to match
> after this spec lands.** The issue as currently written describes a "thin copier" that
> reimplements the setup skill's own mechanical steps in shell — target/style detection, posture
> resolution (prompt-or-`expression.md`), managed-block marker patching (detection is narrowly
> permitted per `decision-0068`'s second amendment; patching is not), and the four-check
> payload verify. A first implementation of that reading (kodhama/trellis#128, `feat/124-curl-writer-script`)
> was built and closed unmerged after maintainer review: reimplementing those steps in a second
> language is itself a **second writer** of decision logic that only `/trellis:setup` should own —
> exactly the drift class `kodhama-0007` exists to kill, just relocated from *content* (rule 2) to
> *control flow*. This spec specifies the corrected design: `install.sh`'s only job is getting the
> **plugin bundle** onto disk somewhere Claude Code can find it; every decision about postures,
> target files, and markers stays inside `/trellis:setup`, run unmodified, afterward.

> **Provenance note (this contract-author run).** This spec was authored without access to a
> `Bash`/`gh`/`WebFetch` tool in this session; kodhama-0007 and PR #128's actual (rejected) body
> were read directly from pre-fetched local copies, and the corrected-design bullets below were
> supplied by the dispatching task brief rather than independently re-read from #124's live issue
> thread or PR #128's closing comment. Where a claim rests on that brief rather than a document
> this run opened directly, it is marked. This does not change once this run gains normal tool
> access — it is recorded here because the spec must say so, not because it is expected to persist.

> **Revision note (post-adversarial-review, this run).** `spec-adversary` reviewed this spec
> (verdict: NEEDS-REVISION) and found the `## Test coverage` table named no scenario for AC2
> (zero decision logic — this spec's central guarantee, the exact thing the rejected #128 got
> wrong), AC9 (no git mutation), or AC10 (exact post-write guidance) — leaving all three
> unverifiable by the test suite this spec itself requires, with AC2 in particular resting on
> nothing but a prose grep heuristic. This revision adds four table entries (two new rows for
> AC2/AC9; two existing rows extended with an AC10 assertion) and nothing else — see the amended
> `## Test coverage` and `## Rubric check` below. Cross-checked against kodhama/trellis PR #129
> (an independently-built implementation of this corrected design, not built from this spec) as
> an existence proof where its test suite happens to cover the same ground; noted inline where it
> doesn't.

> **Retrofit note (post-`kodhama/trellis#138`, this run).** `decision-0044` (approved, merged)
> and its implementation in this repo's own `specs/0001` §1 (PR #137) established a qualified
> `<repo>/<id>` form for cross-repo `depends_on` references; this spec's frontmatter cited
> `kodhama-0007-one-render-many-copiers` unqualified. Retrofitted to
> `kodhama/kodhama-0007-one-render-many-copiers` — `kodhama` is a member of the recognized
> registry, `<id>` unchanged. See the new re-check appended to `## Rubric check` below: it
> replaces the original check 4's flawed real-world-status reasoning with the declared
> shape + registry-membership test `specs/0001` §1 actually specifies.

## Purpose

Give every non-Claude-Code-marketplace user (and any Claude Code user who prefers not to use
`/plugin install`) a `curl | sh` path onto the same governed artifact the plugin marketplace
installs: the **plugin bundle**, unmodified, placed somewhere Claude Code's skills-directory
mechanism discovers it. `install.sh` is a **sibling writer under kodhama-0007**, not a rebuild of
`/trellis:setup` in a second language — it stops the moment the bundle is on disk.

## Scope

**In scope:** fetching, pin-verifying, and writing the plugin bundle (`.claude-plugin/plugin.json`,
`hooks/`, `reference/`, `skills/`) to a Claude-Code-discoverable skills directory; scope (personal
vs. project) selection; fail-closed integrity verification; post-write guidance text.

**Out of scope (stays `/trellis:setup`'s job, run afterward, unmodified by this spec):** posture
resolution, `.trellis/` bundle composition, managed-block **patching/verification** (detection is
narrowly in scope per `decision-0068`'s second amendment — existence-only, to refuse rendering into a
project that already delivers the rules statically; the script still patches nothing and verifies nothing), the
`#112` hand-authored-content backstop, and anything else SKILL.md (`plugins/trellis/skills/setup/SKILL.md`)
already owns. **Also out of scope:** any git mutation (`add`/`commit`); any GitHub release
mechanism (retired, `decision-0043` §4) — the fetch source is a pinned commit's raw content, not a
release asset.

## 1. What `install.sh` vendors

The **whole plugin bundle** — every file under `plugins/trellis/` in this repo — not a subset and
not just the M1 overlay payload:

| Path | Why it must come along |
|---|---|
| `VERSION` | The Trellis plugin package's canonical SemVer (`decision-0061`); both host manifests must match it. (Amended by `decision-0066`: `surfaces.json` was named here too, and is retired.) |
| `.claude-plugin/plugin.json`, `.codex-plugin/plugin.json` | Identify the vendored directory as the `trellis` plugin to Claude Code and Codex respectively. The Codex declaration does not expand this Claude-oriented installer into a Codex installer. |
| `skills/setup/SKILL.md`, `skills/remove/SKILL.md` | The actual skills — `install.sh` gets you *to* these, it does not replace them. Both, not just `setup`: a vendored install with no `/trellis:remove` is a governance tool with no clean exit, which spec-0004 already treats as a trust defect. |
| `hooks/hooks.json`, `hooks/staleness.sh`, `hooks/codex-hooks.json`, `hooks/codex-context.mjs` | The complete host-isolated hook set. Claude's files provide the `decision-0043` staleness surface; Codex's files provide the trusted-local fresh-start context boundary from `decision-0058` and `spec-0007@v1`. They remain bundle bytes even though this installer exposes only the Claude host. |
| `reference/*` (all 15 files, including `checksums`) | The pre-rendered payload `/trellis:setup` copies from — unchanged, verbatim. |
| `README.md` (plugin's own) | Documentation only and not load-bearing to any skill's function, but included and manifest-verified for byte parity with a marketplace install. |

**AC1 depends on this table's left column being exhaustive for the actual `plugins/trellis/` tree
at build time** — if the tree grows a new top-level file class between this spec and
implementation, the executor extends the table (and the manifest, §3) rather than silently
dropping it from the bundle.

## 2. Target resolution — scope and path

| Scope | Target directory | Selection |
|---|---|---|
| **project (default)** | `<repo-root>/.claude/skills/trellis/`, where `<repo-root>` is the output of `git rev-parse --show-toplevel` — **never `$PWD`** | Default when no scope is otherwise resolved and a git repo is detected from the current directory |
| **personal** | `~/.claude/skills/trellis/` | Opt-in only: `--scope personal` flag or `TRELLIS_SKILLS_SCOPE=personal` env var |

Rules:

- **Never `$PWD`-based path construction for project scope.** The script resolves the repo root
  via `git rev-parse --show-toplevel` and vendors there regardless of which subdirectory it was
  invoked from.
- **Ambiguous case: not inside a git repo, and no `--scope`/env override given.** Project scope has
  no target. If a controlling tty exists (`/dev/tty` readable), prompt once (offer personal scope,
  or abort). If no tty exists, fail closed: non-zero exit, name the missing input, no write —
  **never hang, never silently fall back to personal or to `$PWD`.**
- `--scope`/`TRELLIS_SKILLS_SCOPE` always wins over the ambiguous-case prompt; the prompt only
  fires when neither is given and detection is ambiguous.

## 3. Bundle integrity — pin, fetch, verify, then write

Adapts the fetch/pin/verify pattern already built and validated in kodhama/trellis#128's
`install.sh` (its "THE PIN" section and steps 1/7) — that half of #128 was not the mistake; only
the decision-logic half (§§2–4, 6) was. Reused as prior art, not reinvented:

1. **Fetch into a staging directory first.** Nothing under the target scope's real path is touched
   until every fetched byte is verified.
2. **Pin by content, not by release tag or a bare `main` ref** (per `decision-0043`: release-tag
   and release-asset machinery retired with the binary channel; `main` is not self-verifying).
   The script bakes in a pinned manifest hash and a bundle stamp, advanced only by a same-commit
   regenerate-and-diff CI guard — the same mechanism kodhama/trellis#128 built
   (`TestInstallScriptPinIsCurrent`), generalized to the bundle-wide manifest instead of the
   M1-payload-only one.
3. **Verify the fetched manifest against the pin, then every fetched file against the
   (pin-verified) manifest** — both checks, in that order, before any write.
4. **Fail closed.** Any mismatch — pin vs. fetched manifest, or manifest vs. a fetched file —
   exits non-zero, names the exact failing check and file, and leaves the target scope's directory
   exactly as it was before the run (no partial tree).

**AC (new deliverable — no such manifest exists today):** `plugins/trellis/reference/checksums`
covers only the 11-file M1 overlay payload subset (verified against its current contents:
`block-claude.md`, `block-inline-a.md`, `block-inline-b.md`, `expression-a.md`, `expression-b.md`,
`invariants.md`, `profile-a.md`, `profile-b.md`, `trellis-a.md`, `trellis-b.md`, `version` — no
entry for `plugin.json`, any `SKILL.md`, or anything under `hooks/`). A checksum manifest covering
every file in §1's table does not exist anywhere in the repo today. **Its existence, generated at
release time alongside the existing payload render (kodhama-0007 rule 1), is itself an acceptance
criterion of this spec (AC6) — not an assumption that one will appear.** Where it lives and exactly
how it's generated (extend the existing generator vs. a second small release step) is an
implementation choice for the executor; this spec fixes only that it must exist, cover the full
bundle, and move atomically with the bundle it describes.

## 4. Post-write output — guidance, not action

On success, in order, `install.sh` prints (and performs none of b–e itself):

1. **What was written** — scope, resolved target path, bundle stamp.
2. **The trust-dialog note (project scope only)** — Claude Code shows a folder-trust prompt the
   first time it loads a project directory with unfamiliar `.claude/` content; this script does
   not and cannot bypass that prompt, and the note says so rather than leaving a first-time user to
   wonder why Claude Code is asking.
3. **The no-walk-up caveat** — a project-scope skill is discovered only when Claude Code is
   launched with the vendored directory as its actual root; launching from a subdirectory (or any
   other directory) will not surface it, regardless of git repo membership. Personal-scope skills
   have no such restriction.
4. **The commit suggestion (project scope only)** — a one-line, non-imperative suggestion that
   committing the vendored bundle is the human's decision to make (e.g. "review, then `git add
   .claude/skills/trellis && git commit` if you want collaborators to get it on clone"). The script
   never runs this itself.
5. **The suggested next step, verbatim:** run `/trellis:setup`.

`install.sh` never invokes `/trellis:setup` itself, never touches `.trellis/`, never touches any
**pre-existing** instructions file, and never runs a git-mutating command (`add`, `commit`, or
otherwise) — the only git invocation permitted anywhere in this script is the read-only
`git rev-parse --show-toplevel` of §2.

> **Amended by `decision-0068` (2026-07-30), matching AC2.** The one exception is
> `.claude/rules/trellis.md`, a file the script wholly owns and writes fresh. It edits no file it
> did not create, and `.trellis/` remains untouched — so item 5's "run `/trellis:setup`" is still
> the step that activates the rules beyond the two floors.

Output item 1 ("what was written") names the rendered rules file alongside the bundle. Item 5's
suggestion gains its reason: **until `/trellis:setup` writes `.trellis/rules.toml`, only
`floor-transparency` and `floor-intent-gate` apply**, because every other rule is gated on a row
that does not yet exist. Committing the vendored bundle (project scope) is the human's decision; the script suggests
it, never does it.

## Acceptance criteria

- **AC1 — vendors the whole bundle, nothing else.** A fresh run writes every file in §1's table to
  the resolved target and nothing outside it; the vendored tree is byte-identical, file-for-file,
  to the source `plugins/trellis/` tree at the pinned commit.
- **AC2 — zero decision logic.** The script contains no posture prompt, no target/style detection
  for any instructions file, no managed-block marker handling, and never reads or writes
  `.trellis/` or any *pre-existing* instructions file (`CLAUDE.md`, `AGENTS.md`, etc.). (Checkable
  by absence: grep the script source for `trellis:begin`, `expression.md`, `profile-`, `CLAUDE.md`
  outside of comments/help text — none should drive a write decision.)

  > **Amended by `decision-0068` (2026-07-30) — one clause, narrowly.** The script **does** now
  > write exactly one instructions file it wholly owns: `.claude/rules/trellis.md`, rendered from
  > bundle bytes. Measured basis: with user scope excluded, the vendored bundle alone delivers
  > **no rules at all** (issue #201), while this file delivers them.
  >
  > **What is amended:** the blanket "never … writes any instructions file", **and** the blanket
  > "never reads … `.trellis/`" — the second one widened after review, see below.
  > **What is NOT:** the "zero decision logic" heading and everything under it. The script still
  > makes no posture choice — it writes no `rules.toml`, reads none to select prose, patches no
  > marker, and detects no style. It still **never WRITES to `.trellis/`**, so `/trellis:setup`
  > keeps sole ownership of `rules.toml` and `decision-0065`'s "exactly one file … and nothing
  > else, ever" needs no amendment.
  >
  > **Second amendment, 2026-07-30, after an independent reviewer refused the first as
  > insufficient.** The script now **reads** whether the project already delivers the rules
  > statically — `.trellis/internal/`, the legacy flat `.trellis/trellis.md`, or a
  > `trellis:begin` managed block in `CLAUDE.md`/`AGENTS.md` — and refuses to render when any is
  > present.
  >
  > This is a genuine widening and is recorded as one rather than smuggled in a test. The reviewer's
  > words: *"either obtain the intent-level amendment or preserve the mechanical, state-independent
  > installer contract."* The contract cannot be preserved: **both static chains are loaded by the
  > host before any hook runs**, so there is no runtime fix, and rendering blindly would ship
  > guaranteed double delivery. The read is existence-only, of files this script never writes, and
  > it selects **nothing** — its only outcome is render-or-refuse-loudly.
  >
  > **This widening postdates what the maintainer was shown when D1 was ruled** and is flagged for
  > the same intent act rather than assumed under it.
  >
  > With no `.trellis/rules.toml` present the install is inert but for the two `floor-` rows, which
  > "apply regardless of their row value" — a defined default requiring no seed.

- **AC2a — the rendered rules file, and the import form is the assertion.** A fresh run writes
  `.claude/rules/trellis.md` containing the posture prose, the rules body, and the exact import
  line `@../../.trellis/rules.toml`. The test asserts **that** form and asserts the sibling form
  `@rules.toml` is absent. Both were measured: each is correct in one location and **silently loads
  nothing** in the other — no error, no content. A reword that swaps them ships a file that governs
  nothing and passes every other check in this spec.

- **AC2b — the plugin hook stands down when this file exists.** Measured: with the plugin installed
  *and* `.claude/rules/trellis.md` present, the rules arrive **twice** — once in the
  project-instructions block, once in the hook's `additionalContext`. `hooks/staleness.sh` gains a
  branch: when `.claude/rules/trellis.md` exists it injects nothing and says so, exactly as it does
  today for a vendored `.trellis/internal/` overlay. Red-first: a fixture with both present must
  show one delivery, not two.
- **AC3 — project scope resolves via git root, never `$PWD`.** Running the script from the repo
  root and from an arbitrary subdirectory of the same repo (same scope, no explicit target
  override) produce a byte-identical vendored tree at the same absolute path.
- **AC4 — personal scope is opt-in, project is default.** With no scope flag/env and a resolvable
  git repo, the script vendors to project scope. `--scope personal` / `TRELLIS_SKILLS_SCOPE=personal`
  vendors to `~/.claude/skills/trellis/` instead.
- **AC5 — never hangs; fails closed when scope is ambiguous and unresolvable.** Outside a git repo
  with no scope flag/env: prompts once if `/dev/tty` is available, otherwise exits non-zero
  immediately with an actionable message. No code path blocks indefinitely waiting on input that
  cannot arrive.
- **AC6 — a bundle-wide checksum manifest exists and gates every write.** Per §3: the manifest
  covers every file in §1's table (a strict superset of the current `reference/checksums`), and no
  file is written to the target scope until both the manifest-vs-pin and file-vs-manifest checks
  pass.
- **AC7 — fail-closed integrity, no partial writes.** A tampered/corrupted fetch (manifest hash
  mismatch, or any file not matching the pin-verified manifest) exits non-zero, names the specific
  failing check and file, and leaves the target scope's directory in exactly its pre-run state —
  verified by a before/after snapshot diff of the target directory being empty.
- **AC8 — idempotent re-run.** Running the script again with the same scope over an
  already-current vendored tree exits 0 and leaves the tree byte-identical (no duplication, no
  drift, no error on "already exists").
- **AC9 — no git mutation, ever.** The only git invocation in the script is the read-only
  `git rev-parse --show-toplevel`; no `git add`/`commit`/anything-else appears anywhere in the
  script, including on the success path.
- **AC10 — post-write guidance is exactly §4's five items, in order, and nothing more.** No
  automatic invocation of `/trellis:setup`, no automatic git commit (item 4 is a printed suggestion
  only — the boundary this AC actually protects, not the line count), no additional prompts beyond
  scope selection (§2) and the pin-verify fail path (§3).

## Test coverage (required; each maps to an AC above)

| Scenario | Proves | AC |
|---|---|---|
| Personal fresh vendor (clean `~/.claude/skills/`) | Full bundle lands at the personal path; stdout carries the next-step pointer to `/trellis:setup` and does **not** carry the project-only trust-dialog note | AC1, AC4, AC10 |
| **Project fresh vendor renders the rules file** | `.claude/rules/trellis.md` exists and carries: the posture prose byte-identical to `reference/trellis-b.md`, the rules body byte-identical to `reference/rules.md`, the exact import line `@../../.trellis/rules.toml` (and **not** the sibling `@rules.toml`, **not** the project-root `@.trellis/rules.toml`, **not** an unresolved `@rules.md`), the invariants pointer naming a path this install creates, no absolute path from the installing machine, and the `trellis:rendered-from` payload stamp. Each wrong form was measured to load **nothing, silently** | **AC2a** |
| **Personal scope renders nothing, and says so** | No rules file at `~/.claude/rules/` or in the working directory; stdout names "no rules file" | **AC2a** |
| **Render failure fails closed and loudly** | A directory at the target path exits non-zero with the script's own `FAIL:` message and no claim of having rendered; a read-only regular file is replaced (the install owns that name, and re-running is idempotent) | **AC2a**, AC6, AC7 |
| **A vendored `.trellis/internal/` overlay refuses the render** | Nothing is written to `.claude/rules/`; stdout names the overlay as the reason and names the migration route; the bundle still vendors. Both static chains would otherwise load before any hook runs, which no runtime fix can undo | **AC2a** |
| **The hook stands down exactly once** | With the plugin installed and a complete rendered file present, the emitted context carries the stand-down note naming `.claude/rules/trellis.md` and **not** the rule bodies. Composition, not fixtures: the real installer writes the file, the real hook reads it | **AC2b** |
| **Incomplete or stale rendered files never silence the hook** | Zero-byte, one-byte, cut-after-sentinel, import-missing and stamp-missing shapes each either deliver the rules or emit `TRELLIS_RULES_NOT_LOADED`; a stale stamp emits a nudge naming both stamps; a current one stands down quietly. Eleven shapes audited | **AC2b** |
| **Path C sits after path A** | A project holding both a stale overlay and a rendered file still receives the overlay staleness nudge — `decision-0035`'s floor, which path C would otherwise silence for exactly the consumers mid-migration | **AC2b** |
| Project fresh vendor, run from repo root | Full bundle lands at `<repo-root>/.claude/skills/trellis/`; stdout contains, verbatim, the trust-dialog note, the no-walk-up caveat, the commit suggestion (present, and not a command the script itself ran — the target directory's git status shows no staged/committed change from the script), and the next-step pointer to `/trellis:setup` | AC1, AC3, AC4, AC9, AC10 |
| Project fresh vendor, run from a subdirectory | Git-root resolution, not `$PWD` — byte-identical to the repo-root case | AC3 |
| Re-run over an already-vendored tree | Idempotency, no duplication/drift | AC8 |
| Tampered fetch (corrupted file or mismatched manifest hash) | Fails closed: non-zero exit, named check, empty/untouched target directory | AC6, AC7 |
| No controlling tty, scope ambiguous (no git repo, no flag/env) | Exits non-zero with an actionable message; never hangs | AC5 |
| Two otherwise-identical fixture repos, vendored to the same scope: one carrying a pre-existing `CLAUDE.md` with hand-authored prose plus a `.trellis/expression.md` declaring a posture (the managed-block markers moved to the refusal scenario, since `decision-0068`'s second amendment makes a PAIRED block a legitimate branch input — leaving them here would assert the opposite of the contract); the other carrying an `AGENTS.md` instead and no `.trellis/` at all | The vendored bundle tree and stdout are byte-identical between the two runs (only the absolute target path differs — a scope-resolution input, not a decision-logic one); the first fixture repo's `CLAUDE.md` and `.trellis/expression.md` are byte-identical before and after the run (untouched — not read-and-rewritten, not read-and-left-alone-by-luck). A script that branched on target-file presence, content, or posture — under any name — fails at least one of these assertions | AC2 |
| Every scope/error path from the rows above (personal, project-from-root, project-from-subdirectory, ambiguous-no-tty fail-closed, tampered-fetch fail-closed) plus an invalid-`--scope`-value run and a re-run, each executed with a logging `git` shim shadowing `PATH` | The shim's invocation log, read back across every path, contains only `rev-parse --show-toplevel` calls — zero `add`/`commit`/or-any-other-subcommand invocations anywhere, on the success paths or the failure paths alike | AC9 |

**Existence-proof cross-check against PR #129 (informational, not authoritative — the rows above
stand regardless):** its `TestVendorNeverRunsGitAdd` is a same-shape but single-path precedent
for the AC9 row (happy-path project-scope only, via a post-run `git status --porcelain` diff); it
does not extend the git-invocation check to any failure/edge path, so the row above's cross-path
shim is a strictly stronger requirement, not a restatement of an existing test. PR #129's stdout
assertions (`TestVendorPersonalScopeFreshInstall`, `TestVendorProjectScopeFromSubdirectoryResolvesToRoot`)
cover only AC10 item 1 (the `scope: ...` line) — none of its Go tests assert on items 2–5
(trust-dialog note, no-walk-up caveat, commit suggestion, next-step pointer); that content is established in PR #129
only by static reading of `install.sh` and by manual, non-automated transcripts in the PR body, so
the AC10 table rows above are also new requirements, not echoes of an existing automated test. PR
#129 has no test at all resembling the AC2 row: none of its scenarios vary instructions-file or
`.trellis/` content and assert output invariance — its nearest tests
(`TestVendorDefaultFallsBackToPersonalOutsideGitRepo`, `TestVendorExplicitProjectScopeOutsideRepoFailsLoudly`)
vary git-repo *presence*, not instructions-file content, which is a different axis. Where PR #129's
script *behavior* would in fact satisfy the AC2/AC9/AC10 rows above (it does — see its `install.sh`
header's own "makes exactly ONE decision" claim and the fact that it never invokes `.trellis/` or
any instructions-file path), this is cited as evidence the rows are achievable, not as evidence
they were already tested.

## Open questions

- ~~Load-bearing, unverified by this run: does Claude Code's skills-directory discovery treat a
  vendored tree identically to a marketplace-installed plugin~~ — **resolved post-authoring** by
  the dispatching session, live against `code.claude.com/docs/en/plugins-reference` (fetched fresh,
  not from a cached copy). Evidence: "Any folder under a skills directory that contains a
  `.claude-plugin/plugin.json` manifest is loaded as a plugin named `<name>@skills-dir`... **Unlike
  a marketplace install, the plugin is discovered in place rather than copied into the plugin
  cache**" (§Skills-directory plugins) — a *location* difference, not a *capability* one; "A plugin
  `foo@skills-dir`, which **can bundle its own skills, agents, hooks, and more**" (component table,
  same section); `${CLAUDE_PLUGIN_ROOT}` is defined exactly once, non-conditionally, as "the
  absolute path to your plugin's installation directory" (§Environment variables) — no
  marketplace-only carve-out anywhere in the reference. The only scope-conditional restrictions
  documented anywhere are project-scope trust-gating of MCP servers, LSP servers, and background
  monitors (§Skills-directory plugins) — keyed on **scope** (personal vs. project), not on
  **discovery mechanism** (marketplace vs. skills-dir), and hooks are not in that restricted list.
  Confidence: high by consistent silence + explicit hook/skill support, not a single sentence
  reading "identical for skills-dir" verbatim — treat §1/§4 as settled; the invocation surface stays
  `/trellis:setup` since a skills-dir-loaded skill is invoked the same as any other loaded skill
  (no distinct invocation path is documented for either mechanism).
- **Exact manifest mechanism for §3/AC6** — extend the existing Go generator (`cli/`, generator-only
  since `decision-0043`) to emit a second, bundle-wide manifest, or add a small separate release
  step. Either satisfies this spec; left to the executor.
- **`install.sh`'s own location and name** — this spec assumes root `install.sh` (the conventional
  `curl | sh` home, shortest URL), matching kodhama/trellis#128's placement; not re-litigated here
  since it was never the part of #128 that was rejected.
- **Whether the folder-trust dialog and no-walk-up behavior (§4 items 2–3) are still accurately
  worded against current Claude Code behavior** — same unverified-by-this-run caveat as above;
  these are advisory strings, not enforced behavior, so a wording drift is low-severity but should
  still be checked against current docs at execution time.
- **This spec does not resolve #124's own text.** The issue itself still describes the rejected
  framing (thin copier reimplementing setup's mechanical steps) as of this spec's writing — a
  human or a follow-up pass should edit #124 to point at this spec rather than carry the stale
  description forward for future readers.

## Rubric check

Self-checked against `core/rubrics/artifact-contract.md` (trellis has no dedicated spec-quality
rubric — noted in trellis's own `contract-author` agent instructions; this is the closest real
gate, per that agent's own §Method item 4).

| Check | Result | Note |
|---|---|---|
| 1. Frontmatter present & required fields valid | PASS | `id/type/status/depends_on/owner` all present and well-typed; `depends_on` is a list. |
| 2. `type` declared; `status` in the declared lifecycle | PASS | `type: spec`; `status: gated` — the family enum (`decision-0042`), applicable since this is a forward artifact authored after that decision. |
| 3. `id` unique across the corpus | PASS (assumed against the read corpus) | `spec-0005` — no existing spec above `0004` was found in the read tree; this run could not run a live corpus-wide `grep` against the actual current `kodhama/trellis` main (no fetch tool), so this is asserted from the most current local mirror available, not a fresh remote check. |
| 4. `depends_on` resolves | PASS | `kodhama-0007-one-render-many-copiers` — read directly, `status: approved`. `decision-0043` — read directly, `status: gated` (a `gated`/agent-consumable upstream is a legitimate dependency for a spec that is itself not yet `approved`; both this spec and `decision-0043` await human merge together, consistent with `decision-0042`'s mechanic). |
| 5. Directional flow — no `gated`/`approved` artifact depends on a `draft` | PASS | Both dependencies are `gated` or `approved`, never `draft`; this spec is itself `gated`, not `approved`, so it does not violate the rule in the other direction either. |
| 6. Required body sections per type (spec → Acceptance criteria + Open questions) | PASS | Both present, non-empty. |
| 7. Supersede integrity | N/A | This spec supersedes nothing; it is new. |
| Checks 8–11 (typed-artifact checks) | N/A | Not a `signature-catalog` or `expression-profile`. |
| Check 12 (version cross-check, `decision-0045`) | N/A | Not a `changes:`-bearing `decision`; added to the rubric 2026-07-11. |
| Honesty clause | Self-assessed honest | The two flagged gaps above (check 3's corpus freshness, and the load-bearing Open Question on skills-directory plugin mechanics) are reported as gaps, not hidden to force a clean pass. |

### Re-check (post-adversarial-review revision, this run)

`core/rubrics/artifact-contract.md` was re-read fresh this run (available locally; not
re-fetched from the live repo — same corpus-freshness caveat as check 3 above) rather than
assumed unchanged from the first pass. The edit was scoped to the `## Test coverage` table plus
this section and the two provenance/revision blockquotes — no frontmatter, AC wording, or Open
Questions content touched — so checks 1–7 above stand as previously assessed; re-verified
directly relevant to this edit:

| Check | Result | Note |
|---|---|---|
| 6. Required body sections per type | PASS | `## Acceptance criteria` and `## Open questions` both still present and non-empty; the edit added content to `## Test coverage`, which is not itself a rubric-required section but is the artifact this spec's own "each maps to an AC above" header commits to — now true for AC2, AC9, AC10 as well as the other seven. |
| Honesty clause | Self-assessed honest | The existence-proof paragraph appended to the test-coverage table reports, without softening, that PR #129 has *no* automated equivalent for the AC2 or AC10 rows (only a narrower, single-path one for AC9) — cited as an achievability signal, not misrepresented as prior test coverage. A second, out-of-scope divergence between this spec's AC5 and PR #129's actual not-in-a-git-repo fallback behavior was also found during this run; it is not fixed here (targeted revision, not a rewrite) and is flagged instead in this run's report to the dispatching session. |

### Re-check (post-`kodhama/trellis#138` retrofit, this run)

`specs/0001` §1 was re-read fresh this run (the external-refs paragraph specifically) to confirm
the recognized cross-repo form before re-grading — not assumed from memory of the first pass.

| Check | Result | Note |
|---|---|---|
| 4. `depends_on` resolves | PASS | Re-graded against the mechanism `specs/0001` §1 actually declares — shape + registry-membership — not the referent's real-world status (the prior pass's error, see below). `kodhama/kodhama-0007-one-render-many-copiers`: matches the qualified `<repo>/<id>` shape; `kodhama` is a member of the recognized registry (`kodhama, trellis, grove, wisp, design-system, homebrew-tap, math-quest`); `<id>` (`kodhama-0007-one-render-many-copiers`) is cited exactly as declared in its home corpus — this is in fact `specs/0001` §1's own worked example, verbatim. Per that section's stated "Resolution depth (v0): shape + registry-membership only... no fetch-and-confirm-the-referent-actually-exists mechanism," the check stops there; that the referent also happens to carry `status: approved` in the kodhama repo (confirmed by direct read this run, for this run's own confidence only — not because the rubric requires it) is evidence beyond what this check tests. `decision-0043`: same-corpus id, not an external ref — resolves by ordinary corpus lookup, `status: gated`, unchanged from the original check. |

**What was wrong with the original check 4 (as first written).** It graded the dependency's
resolution PASS by reading the referent directly and reporting its real-world status
(`approved`) — accurate as a fact, but not the test `specs/0001` §1 defines for an external
ref, and it did not flag that the reference was unqualified (missing the `kodhama/` prefix) at
the time, which is the actual defect `kodhama/trellis#138` filed. The re-check above applies
the declared allowlist mechanism instead of the referent's live status.

**Promotion: `draft → gated`.** `approved` happens only by human PR merge (`decision-0042`) —
not set here. This retrofit does not change that: no frontmatter field other than the
`depends_on` entry's qualification was touched, and the fix is corrective, not a new promotion.

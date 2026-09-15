# Batch 3: decision-0051 to decision-0068

Sorted 2026-09-15 against worktree `simplify-1b` at `6f292aa` (TRL-97, PR #316 merged). All paths are repo-relative. "Verified" means I read the cited file and it carries the behaviour.

**Missing numbers.**
- **0056 was never drafted.** `decision-0055` parked it (0055:23, :79, :105) as the rewrite of the inline decision citations in the catalog. No record exists. Its only remaining consumer is a comment at `cli/apply_test.go:11` ("exactly what decision-0056 will eventually do").
- **0067 was drafted but never merged.** Two commits sit outside `main`: `02253f3` ("decide: consulted material is a skill; what must apply stays injected (gated)") and `7744768` ("revise(decision-0067): narrow to plugin-path-only…"). `decision-0089:203` records "#208 → `0067`, which is a genuine gap on `main`". `decision-0068:463` and `0068:346-353` record that 0068 dropped its reliance on that unmerged draft.

---

### decision-0051 — overlay authority split and `rules.toml`
- cluster: rule-model
- disposition: mixed
- in force:
  - Rule 1 (root half): `.trellis/rules.toml` belongs to the consumer, and no Trellis writer ever rewrites an existing file. Home: `install.sh:975-1000` seeds only when the file is absent, `install.sh:674` says "this installer never overwrites it", and neither hook has a write path (`plugins/trellis/hooks/staleness.sh`, `plugins/trellis/hooks/codex-context.mjs`, no write calls). Verified.
  - Rule 1 (legacy half): an existing `.trellis/internal/` overlay is recognised as Trellis-generated legacy state. Home: `staleness.sh:442-529` (Claude path A) and `codex-context.mjs:46-48` (Codex reads it). Also `plugins/trellis/README.md:175-190`. Verified.
  - Rule 2 (residue): editing `rules.toml` is the configuration act, and no refresh step is needed. The row semantics are now opt-out. Home: `plugins/trellis/README.md:150-157` and `staleness.sh:1017-1035`. Verified.
  - Rule 3: floor rules cannot be switched off, and a `false` floor row is surfaced rather than honoured silently. Home: `plugins/trellis/reference/rules.md:1`, `staleness.sh:1224`, `:1296-1297`. Verified. The mechanism changed; see retired.
  - Rule 5 (de-collision): the `profile:` key and `profile.md` are gone. Verified: no `expression.md` or `profile.md` in the payload; `plugins/trellis/README.md:207-212` gives the manual migration for legacy files. The name "profile" is reserved for the `expression-profile` artifact. Stated only in the record: `docs/rubrics/artifact-contract.md:54,98` uses the type but never states the reservation.
  - Rule 6: Trellis diverges from `kodhama-0007` rule 4's carrier half. Its config is a dedicated hand-owned TOML file, and every installed file is wholly generated or wholly hand-owned. The stewards amendment stays parked. Stated only in the record.
  - Rule 7 (residue): per-rule C1/C2 dials stay out of `rules.toml`, and no unenforced column is added. Home: `staleness.sh:1179`, whose row grammar accepts only `active`; any other shape is a malformed row. Also `rules.md:1`: "Nothing else in `.trellis/rules.toml` changes which rules apply". Verified.
  - Amendment (2026-07-19): `expression.md` is retired from the bundle, and a project's governance prose belongs in its own instructions file. Home: `plugins/trellis/README.md:185-187`, `cli/selfapply_test.go:64` (`TestRepoOverlayCarriesNoExpressionFile`), and `plugins/trellis/skills/remove/SKILL.md:38` (legacy handling). Verified.
- retired:
  - Rule 1 (internal half), vendoring `.trellis/internal/` on setup and refresh: retired by decision-0065 (the plugin path never vendors) and decision-0072 (setup retired). Nothing writes the directory any more.
  - Rule 2's posture-as-seed, its normative shape ("one row per assessable catalog slug", `seeded_from`) and its refresh-time semantics: retired by the dated note of PR #316 and by decision-0053.
  - Rule 3's "assembly ignores `active = false`… says so loudly in the confirm step" (the mechanism): retired by decision-0053. Floors are also dropped under `governed = false` (decision-0070 D5; `staleness.sh:285-290`).
  - Rule 4 (fragment assembly, "no per-session reader"): retired by decision-0053.
  - Rule 5's closing-line spec: retired by decision-0053.
  - Rule 5's `expression.md` seed cross-link: lapsed with the amendment.
  - Rule 7's "`strictness` is one instance-level key": retired by the dated note of PR #316.
  - Rule 7's "seed/custom presets stay parked": moot, because no preset ships at all (2026-09-15 record, point 6).
  - Amendment's "migration step preserves/offers to move expression.md": retired by decision-0072. The step is now manual (`plugins/trellis/README.md:210-212`).
  - 2026-07-24 pointer's confirmed preset reset: retired by the dated note of PR #316.
  - Rule 1's rejection of a per-session hook ("no reader exists"): retired by decision-0065 (forward pointer).
- spent: the `/trellis:setup` rewrite, the family overlay migration and the forward pointer onto 0033. None has a live subject: setup was retired by 0072, and this repo has no `.trellis/internal/` (checked: absent).
- reason: The consumer-owned `rules.toml`, floor protection, no per-rule dials and the retirement of `expression.md` still hold. The overlay layout, presets, strictness and assembly are gone.
- confidence: medium. Whether rule 6 (the family divergence) and the "profile" name reservation count as standing rules is a judgment call; both are stated only here.
- flags:
  - This repo's own `.trellis/rules.toml` still carries `seeded_from`, `strictness` and a comment claiming "one row per assessable catalog slug". These are inert under TRL-97, which leaves them in place deliberately (2026-09-15 record, Consequences).
  - `plugins/trellis/README.md:177-178` still describes rows as "one row per rule, `active = true|false`", but only inside the legacy-overlay section, which is scoped to older installs.

### decision-0052 — `inv-self-improvement` covers the entropy lean
- cluster: rule-model
- disposition: repealed
- in force: none. The SI-1 channel discipline that the frontmatter comment says "still governs" is `invariants-v1` content, which predates this record. The catalog carries it at `core/catalog/signature-catalog-v1.md:195`.
- retired:
  - Point 1 (the appended directive sentence), point 2 (the signature clause) and point 3 (the *(structure)* honored/violated pair): retired by decision-0074. Tree: `core/catalog/signature-catalog-v1.md:195` says "the entropy lean it briefly carried now homes in `inv-deliberate-succession`".
  - Point 4 (the rendered ✗ line extension): also retired by decision-0074. Tree: `plugins/trellis/reference/rules.md:17-18` gives `inv-self-improvement`'s ✗ line without the extension, and the extension's text now sits on `inv-deliberate-succession` at `rules.md:19-20`.
- spent:
  - The experiment-sequencing hazard: batch 1 ran and 0053 recorded it.
  - Regenerating the derived chain: done.
  - Point 5, the naming-guardrail attribution: a one-time statement. The general rule lives in `AGENTS.md:116-120`.
- reason: Everything this record added to the catalog moved to `inv-deliberate-succession` under decision-0074. Nothing of its own is left to carry.
- confidence: high
- flags: The `superseded_in_part_by` comment names "points 1-3" only. Point 4 (the ✗ line) is also gone from the rule, so the comment is incomplete. This is a record-graph defect, not code drift.

### decision-0053 — live rows: the readout ships complete, rows govern at read time
- cluster: rule-model
- disposition: mixed
- in force:
  - Point 1: the complete readout ships. There is no per-row subset assembly, and per-rule fragment files do not ship. Home: `cli/payload.go:35,47` (`renderRulesReadout()`), `cli/apply.go:379-406`, and `plugins/trellis/reference/` (no `rules/` directory). Verified.
  - Point 2 (residue): rows govern which rules apply when the payload is read, not when it is written; a `false` row's rule does not apply; the floor rules apply regardless of their row while the project is governed. Home: `plugins/trellis/reference/rules.md:1` and `staleness.sh:1017-1366`. Verified.
  - Point 3: a row edit takes effect at the next context-loading boundary with no refresh (as clarified by 0058). Home: `plugins/trellis/README.md:52-53`, `README.md:147-149`, and `staleness.sh`, which reads `.trellis/rules.toml` on every `SessionStart`. Verified.
  - Point 4: no shipped text may claim refresh-time semantics for rows. Tree complies: a grep for "refresh" across `plugins/trellis/reference/*.md` returns nothing. Stated only in the record; no guard test exists.
  - The rows are imported rather than inlined on the install path. Carried forward by 0068 D3 (`install.sh` render footer; pinned at `cli/install_script_test.go:1036-1040`). Verified.
- retired:
  - Point 2's authority-header wording ("**only** where `rules.toml` marks the row `active = true`"): retired by the dated note of PR #316.
  - Point 2's import channel through the managed block on the plugin path: retired by decision-0065. The hook injects instead.
  - Point 2's inline channel ("the block inlines the rows below the rules"): retired by the dated note of PR #316. `plugins/trellis/reference/block-inline.md` ends at the tail and carries no rows section.
  - The "floors apply regardless" clause is narrowed by decision-0070 D5: under `governed = false` the floors go too.
  - Point 3's "`/trellis:setup` refresh … validates `rules.toml` on every run": retired by decision-0072. Validation now happens per entry, with warnings, in the hooks (2026-09-15 record point 3; `staleness.sh:1288-1355`).
  - The forward pointer's and Context's "the tested wording is the shipped wording": retired on the plugin-native path by the dated note of PR #316. The reworded sentence is untested (2026-09-15 record, Open questions).
  - Amendment (setup offers to repair legacy seed comments): retired by decision-0072, and moot under TRL-97.
- spent: point 5 (supersession bookkeeping on 0051); machinery deletions (SKILL.md steps, the two-oracle verify); unparking trellis#166; the `research-0012` appendix.
- reason: The complete readout, rows deciding at read time, floor protection and "edits apply next session" still hold. The tested authority wording, the inline rows and setup validation are gone.
- confidence: medium. Records 0093-0096 still say "`decision-0053` is not engaged", treating the tested-wording clause as a standing rule. After TRL-97 that clause holds on no path with shipped new wording, so a spec author must decide whether "ship tested wording" survives as a principle.
- flags:
  - **DRIFT (comment):** `plugins/trellis/hooks/staleness.sh:27` still says "The tested wording stays the shipped wording (decision-0053)". The 0053 dated note and the 2026-09-15 record say that is no longer true.
  - Point 4 has no guard.

### decision-0054 — vendored `invariants.md` ships without frontmatter
- cluster: delivery
- disposition: mixed
- in force:
  - Point 2: the source catalog keeps its frontmatter unconditionally. Home: `core/catalog/signature-catalog-v1.md:1-9`. Verified.
  - Point 3 (surviving half): `cli/assets/invariants.md` stays byte-identical to the catalog source. Home: `cli/sync_test.go:27-34` (`TestBundledCatalogInSync`) and `cli/apply.go:21-22` (`go:generate cp`). Verified.
  - No shipped payload file carries artifact-contract frontmatter. `decision-0082:57` relies on this. Home: the payload files under `plugins/trellis/reference/` (verified: none opens with `---`). No dedicated guard beyond the extraction test.
- retired: Point 1 (strip frontmatter only) and point 3's payload assertion `stripFrontmatter(source)`: retired by decision-0055 point 1, which replaced the strip at the same call site (`cli/apply.go:40`, "This REPLACES decision-0054's narrower stripFrontmatter"). No pointer exists on 0054.
- spent:
  - Point 4 (our synthesis, not a claim on grove's behalf): one-time. Grove itself retired (decision-0076).
  - Refreshing family repos: done or moot.
  - This repo's `.trellis/internal/invariants.md`: that directory no longer exists (decision-0071).
  - The open `.toml` and extensionless gap in grove's check: moot since decision-0076.
- reason: The source keeps its metadata and the payload ships none. The narrow strip itself was replaced a day later by 0055's wider extraction.
- confidence: high
- flags: 0054 carries no `superseded_in_part_by` or dated note, even though 0055 replaced its point 1 at the same site.

### decision-0055 — vendored `invariants.md` carries only the entries body
- cluster: delivery
- disposition: in-force
- in force:
  - Point 1: the payload `invariants.md` is the text from `## Entries` (inclusive) to `## Acceptance criteria` (exclusive), and the build fails loudly if either heading is missing. Home: `cli/apply.go:32-91` (`extractEntriesSection`), `cli/payload.go`, `cli/sync_test.go:36-45`, and `cli/apply_test.go:40-83`. Catalog headings at `core/catalog/signature-catalog-v1.md:82` and `:452`. Verified.
  - Point 2: the source catalog keeps its preamble and tail, and the embedded `invariantsRef` stays untouched. Home: `core/catalog/signature-catalog-v1.md` and `cli/apply.go:21-22`. Verified.
  - Point 3: inline decision citations inside entries ship exactly as they read in the source; no build-time transform touches them. Home: `plugins/trellis/reference/invariants.md` (citations at lines 84, 88, 114, 123, 135, 172, 233, 259, 324, 334). Verified.
  - Point 4: `site/invariants.html` and the eval scorecard read the source catalog, never the payload copy. Home: `cli/sync_test.go:50` (`TestInvariantsPageMatchesCatalog`) and `eval/experiments/does-trellis-help/gen-invariant-scorecard.py:12`. Verified.
- retired: none
- spent: the build sequencing after PR #173; regenerating the derived chain; this repo's `.trellis/internal/invariants.md` (gone under decision-0071).
- reason: Every decision point is visible in the generator and its test today. Only the one-off build sequencing is done.
- confidence: high
- flags: **Orphan follow-up.** Point 3 parks the citation rewrite for `decision-0056`, which was never drafted. The inline citations grew from seven lines to ten (0074 and 0078 added more). The only consumer is `cli/apply_test.go:11`, which decision-0078's no-orphan rule would not accept. A spec author must either carry "citations ship as-is" as a deliberate rule or name an owner for the rewrite.

### decision-0057 — `AGENTS.md` is the canonical shared project-instruction entrypoint
- cluster: method
- disposition: mixed
- in force:
  - Rule 1: root `AGENTS.md` is the one shared authority for this repo's project instructions, and shared prose is never copied into `CLAUDE.md`. Home: `AGENTS.md:173-179` and `cli/selfapply_test.go:75-142` (`TestSharedProjectInstructionEntrypoints`). Verified.
  - Rule 2: root `CLAUDE.md` is the `@AGENTS.md` adapter. Since 0071 it is exactly that line (`CLAUDE.md` holds `@AGENTS.md`; `cli/selfapply_test.go:97-98`). `AGENTS.md` carries the routing statement: shared rules are edited in `AGENTS.md`, Claude-only rules go in `.claude/rules/`, Trellis choices stay in `.trellis/`. Home: `AGENTS.md:175-179` and `cli/selfapply_test.go:111-119`. Verified.
  - Rule 4 (surviving bullets): CI asserts exactly one `@AGENTS.md` adapter in `CLAUDE.md`, the routing instruction in `AGENTS.md`, and no Trellis managed block in `AGENTS.md`. Home: `cli/selfapply_test.go:75-182`; `:175` also forbids the Codex bootstrap, per 0071. The test runs in `cli-ci` `build-test` on every PR (`AGENTS.md`, Checks). Verified.
- retired:
  - Rule 1's "Grove routing block live[s] in root `AGENTS.md`": retired by decision-0076 (`cli/selfapply_test.go:142` forbids retired Grove content).
  - Rule 2's "retains the existing Trellis managed import block below it": retired by decision-0065 and decision-0071.
  - Rule 3 (the managed block stays in `CLAUDE.md`, byte-identical to `block-claude.md`): retired by decision-0065 and decision-0071.
  - Rule 4 bullets 4-5 (exactly one managed block in `CLAUDE.md`, byte-identical to the payload): retired by decision-0065 and decision-0071 (`TestRepoOverlayIsCurrent` deleted, per 0071 Consequences).
  - The 2026-07-24 pointer (0058's Codex receipt block in `AGENTS.md`): retired for this repo by decision-0071.
- spent: rule 6 (forward pointers on 0005, 0028 and 0035); moving the current-truth references from `CLAUDE.md` to `AGENTS.md` (done).
- reason: `AGENTS.md` as the single edited home, with `CLAUDE.md` as a one-line adapter under a CI test, still holds. The Trellis import block and the Grove block are gone.
- confidence: high
- flags: **Scope contradiction.** Rule 5 says this is "a repo shared-entrypoint change, not a new Trellis delivery contract". Yet `plugins/trellis/hooks/staleness.sh:374-391`, `install.sh:583` and decision-0091 (:33, :141) now treat rule 2's "one standalone `@AGENTS.md` line" as the consumer-facing adapter contract that decides whether Claude reads a consumer's `AGENTS.md`. The spec author must state that consumer-side rule in delivery, not only in method.

### decision-0058 — phase host-native live-rule delivery
- cluster: delivery
- disposition: mixed
- in force:
  - Point 1: the stable contract is the rule payload plus the current rows, not a particular transport. A row edit is seen at the next host context-loading boundary, never mid-context. Home: `plugins/trellis/README.md:52-53` and `README.md:147-149`. Verified. The loaded-context sentinel `<!-- trellis:rules-loaded -->` is at `plugins/trellis/reference/rules.md:39`, `staleness.sh:825-859` and `codex-context.mjs:12`. The `payload@…` stamp serves as a diagnostic only, not an authority (`staleness.sh:1397-1400`). Verified.
  - Point 3 (residue): Codex has a host-gated `SessionStart(startup)` hook emitting `hookSpecificOutput.additionalContext`, and the two hosts are isolated (the Claude registration carries no Codex transport). Home: `plugins/trellis/hooks/codex-hooks.json` (matcher `startup`) and `cli/codex_hook_test.go:200-212`. Verified. There is no per-host disable, and `/trellis:remove` is product-wide. Home: `plugins/trellis/skills/remove/SKILL.md:8-12` and `plugins/trellis/README.md:118`. Verified.
  - Point 4 (content): the Codex bootstrap block uses loaded context when present, otherwise reads the installed files, and fails visibly only when inputs are unusable. Home: `plugins/trellis/reference/block-codex.md`, rewritten for opt-out under TRL-97. It reaches consumers only by manual copy (decision-0069; `README.md:162-165`), and `/trellis:remove` strips it (`SKILL.md:24-25`). Verified as shipped text.
  - Point 5: host and surface expansion is evidence-gated, and README support wording advances only with named evidence. Home: `plugins/trellis/README.md:55-62` ("Not verified on any surface: `compact`, `clear`, `fork`, subagent…") pinned by `cli/plugin_readme_test.go:55` (`TestPluginReadmeStatesHostSupportClaim`), and `README.md:157-158`. Verified.
  - Point 6 (residue): nothing overwrites an existing consumer `rules.toml`. Home: `install.sh:674`, `:975-1000`. Verified.
  - Point 7 (residue): failure is visible without being needlessly fatal. A hook that cannot deliver says so and tells the agent to inform the user; a missing hook run is not itself a failure; with no valid rules the agent must not claim governed execution. Home: `staleness.sh` `TRELLIS_RULES_NOT_LOADED` messages ("Tell the user before doing substantive work"), `codex-context.mjs:63-69` (systemMessage), and `block-codex.md`'s last paragraph ("Trellis was not loaded"). The version must have the shape `payload@<12 hex>` (`staleness.sh:256-273`). Verified.
- retired:
  - Point 1's source "from `.trellis/internal/`": narrowed to vendored projects by decision-0065. The plugin payload is the source otherwise.
  - Point 2 (Claude's import transport and managed block stand): retired by decision-0065.
  - Point 3's "reads the installed project's `.trellis/internal/*`" as required input: retired by decision-0065 (`codex-context.mjs` reads the plugin payload when no overlay exists).
  - Point 3's Codex setup/refresh branch: retired by decision-0072.
  - Point 3's "Phase 1's supported claim": retired by the tree. `README.md:130-135` says "Codex is not supported yet" (citing upstream `kodhama-0028` and `kodhama-0021` §2), and `plugins/trellis/README.md:83-85,135-142` says "Trellis claims no support". No local dated note records it.
  - Point 4's "Setup adds … bootstrap to `AGENTS.md`": retired by decision-0072 (no writer). For this repo it is retired by decision-0071 (`cli/selfapply_test.go:175`).
  - Point 6's explicit "apply preset" operation: retired by the dated note of PR #316.
  - Point 7's "strictness, complete-known-row, and floor validation": retired by the dated note of PR #316.
  - Point 7's "All four hook inputs are required": retired by decision-0065 and TRL-97. With no overlay the inputs are plugin files, and `rules.toml` may be absent under a vendored bundle (decision-0070 D3).
- spent: point 8 (pointer bookkeeping); phase 4 was claimed and discharged by decision-0065 and decision-0068 D10; the Phase 1 build (PR #181).
- reason: Payload plus live rows, next-boundary effect, host isolation, evidence-gated claims and loud-but-not-fatal failure still hold. The Claude import block, presets, strict validation and the Codex "supported" claim do not.
- confidence: medium. The "supported" claim was retired by upstream family records and README wording, not by a Trellis record. It is also unclear whether `block-codex.md` counts as a live delivery shape or as manual-copy payload.
- flags:
  - **DRIFT:** the support claim in point 3 ("Phase 1's supported claim") contradicts `README.md:132` and `plugins/trellis/README.md:83,135`, with no dated note on 0058.
  - **DRIFT (message):** `plugins/trellis/hooks/codex-context.mjs:63-68`'s failure message tells the agent "The AGENTS.md bootstrap must attempt the installed overlay". No install writes that bootstrap any more (0072; 0071; the 2026-09-15 record says "no install writes that block today").
  - The 0071 frontmatter scopes its retirement of point 4 to "the trellis repo ONLY … while Codex has no install channel".

### decision-0059 — adopt the family plugin release and surface contract
- cluster: delivery
- disposition: repealed
- in force: none
- retired: all points (§1-§6), by decision-0060 (`superseded_by: [decision-0060]`, and 0060 §2: "No partial requirement from `decision-0059` … remains current"). Tree agrees: no `trellis-v*` tag exists (`git tag -l` is empty), no release workflow exists, and `plugins/trellis/surfaces.json` is gone.
- spent: none
- reason: The family release certification was adopted and retired the same day, and nothing from it was built.
- confidence: high
- flags: Package SemVer in `VERSION` exists today, but it comes from decision-0061, not from this record. Do not credit 0059 for it.

### decision-0060 — retire family release certification; keep local live-rule delivery
- cluster: positioning
- disposition: mixed
- in force:
  - §2 (negative): Trellis does not implement family release certification, an immutable release-tag contract, release metadata or inventory, generated support derivatives, release history or approval records, a pre-tag/tag/release workflow, or a shared validator runtime. Nothing from 0059/spec-0008 stays current merely because it was described there. Stated only in the record. Tree complies: no tags, and `.github/workflows/` has no tag or release workflow (`release-guard.yml` is a PR check).
  - §4: any future headless test depends on Stewards only for a narrow CI provisioning step, which must not certify releases, define Trellis versioning, validate behaviour or promote support. Trellis owns the test and its evidence. Stated only in the record. The closest tree artifact is the cloud setup script at `plugins/trellis/README.md:93-116` (TRL-35), which uses `claude plugin marketplace add kodhama/stewards` directly, not a Stewards skill.
- retired: §2's "canonical shared SemVer authority or cross-host version-equality contract". Retired in its Trellis-local form by decision-0061 §1, which 0060 itself invited ("requires a new project-local decision"). Tree: `plugins/trellis/VERSION` and `cli/plugin_package_test.go:218-257`.
- spent:
  - §1 (0058 and spec-0007 remain authoritative): a statement at the time. spec-0007 was since retired (decision-0079).
  - §3 (0059's narrowing of 0043 withdrawn): later re-narrowed by 0061.
  - §5 (pointer propagation onto 0059 and spec-0008): done.
- reason: The refusal to build family release machinery still holds and the tree has none. The version-equality part was re-adopted locally by 0061.
- confidence: medium. §2's negatives and §4's conditional Stewards boundary are stated nowhere else. It is a judgment whether they are standing rules or historical scope-setting.
- flags: none

### decision-0061 — version the dual-host plugin independently
- cluster: delivery
- disposition: mixed
- in force:
  - §1: `plugins/trellis/VERSION` is the sole plugin-package SemVer authority, and both host manifests declare `name: "trellis"` with a version exactly equal to it. Home: `plugins/trellis/VERSION` (`0.24.0`), `plugins/trellis/.claude-plugin/plugin.json`, `plugins/trellis/.codex-plugin/plugin.json`, `cli/plugin_package_test.go:218-252`, and `plugins/trellis/README.md:9-10`. Verified. That no other product's version is read or compared is stated only in the record; the test reads no other product.
  - §2: package SemVer and the content-derived `payload@<12 hex>` stamp are separate identities, and the staleness comparison is file to file. Home: `plugins/trellis/reference/version`, `staleness.sh:228-276`, `:525-527`, `:665-668`, and `plugins/trellis/README.md:11-13`. Verified.
  - §4 (surviving three bullets): canonical SemVer, cross-manifest identity and version equality, and existence of manifest-declared paths. Home: `cli/plugin_package_test.go:239` (SemVer), `:251` (equality), `:254-257` (paths). Verified.
  - §5: Stewards owns marketplace exposure. Trellis adds no tag, GitHub Release, release metadata or workflow, generated support doc, shared family schema, or support inference from package or marketplace presence. Home: `plugins/trellis/README.md:85-91` (listing ≠ shown to work) and `:122-129` (install from `kodhama/stewards`). The negatives are stated only in the record; the tree complies.
- retired:
  - §3 (`surfaces.json` and its closed row shape): retired by decision-0066.
  - §4 bullets 3-5 (the surfaces version, surface ids and states, marketplace observations): retired by decision-0066.
  - §2's "Setup continues to copy that value to `.trellis/internal/version`": retired by decision-0065 and decision-0072.
- spent: the initial `0.2.0` value; the supersession of 0036's commit-SHA rule and 0043 point 4's "plugin versions are commits"; the two Consequences bullets that 0066's pointer marks false.
- reason: One `VERSION` file matched by both manifests, kept separate from the payload stamp and guarded by a Go test, still holds. The surface matrix is gone.
- confidence: medium. §5's "no … synchronized version or bump policy" is in tension with the bump rule that now exists (see flags).
- flags: **Tension.** §5 lists "synchronized version or bump policy" among things Trellis adds none of, and Consequences (0061:229-230) says the decision "does not automate or centrally dictate that judgment". `.github/workflows/release-guard.yml` now enforces a `VERSION` bump on every `plugins/trellis/` change, and `AGENTS.md` (Checks) says "A payload change is a release". No dated note on 0061 records the change. If "synchronized" qualifies both nouns, this is not a contradiction, but the spec must state the bump rule from `release-guard.yml` and `AGENTS.md`, not from 0061.

### decision-0062 — receive the Stewards adoption-posture strategy
- cluster: positioning
- disposition: spent
- in force: none
- retired: none
- spent: recording receipt of `kodhama-0021` via `kodhama-0022`, as communication only. Its parked follow-up ("any future Trellis adoption-posture choice is a separate local decision") was taken by decision-0063. Visible as 0063's `depends_on: [decision-0062]`.
- reason: A receipt memo that authorized nothing, and its one follow-up was resolved by 0063.
- confidence: high
- flags: Frontmatter uses `updated:` rather than `date:`, as 0063, 0064 and 0066 also do. This is irrelevant to the sort but may matter to a corpus parser.

### decision-0063 — permit Codex preview adoption through the Stewards catalog
- cluster: positioning
- disposition: mixed
- in force:
  - Standing permission: Trellis's Codex package may be catalog-listed in the Stewards marketplace for explicit opt-in preview use, provided the listing points at `kodhama/trellis` `plugins/trellis`, makes no support claim of its own and defers to Trellis's docs. The catalog entry and its copy live in `kodhama/stewards` and cannot be checked from this repo. `plugins/trellis/README.md:136-137` confirms "a catalog entry" exists. Otherwise stated only in the record.
  - Catalog availability changes no Trellis behaviour, version or support claim. Home: `plugins/trellis/README.md:85-91`, pinned by the marketplace-hedge needle at `cli/plugin_readme_test.go:62`. Verified.
  - Rollback, part 1: `/trellis:remove` cleans a project, including both host-managed blocks, and is not described as uninstalling the plugin. Home: `plugins/trellis/skills/remove/SKILL.md:8-25,121-127`. Verified. Part 2, `codex plugin remove trellis@kodhama`, is stated only in the record (grep finds it only at `0063:103`). Removing the shared marketplace registration is optional and must not be suggested while other family plugins use it; also stated only in the record.
- retired:
  - §Preview's co-authority of `surfaces.json`: retired by decision-0066.
  - "The existing `codex-cli-local-startup` behavior claim remains supported": retired by the tree. `README.md:130-135` says "Codex is not supported yet" and `plugins/trellis/README.md:83-85,135-142` says "Trellis claims no support". The upstream is `kodhama-0028` / `kodhama-0021` §2, with no local dated note.
- spent: authorizing the one Stewards catalog-admission change (a catalog entry exists per `plugins/trellis/README.md:137`); the parked GitHub Actions item #182, which has moved to Linear tracking (decision-0075).
- reason: Preview listing without a support claim, and the project-cleanup-then-uninstall rollback, still hold. The "supported" wording and the `surfaces.json` co-authority do not.
- confidence: medium. The catalog state is in another repo, and the retirement of "supported" happened upstream, not in a Trellis record.
- flags:
  - **DRIFT:** "remains supported" (`0063:90-92`) contradicts `README.md:132` and `plugins/trellis/README.md:83,135`.
  - `codex plugin remove trellis@kodhama` appears in no user-facing doc, so a Codex preview user's uninstall step lives only in this record.

### decision-0064 — receive the Stewards surface grammar
- cluster: positioning
- disposition: spent
- in force: none
- retired: the received upstream `kodhama-0023` is superseded in full by `kodhama-0025` (the pointer at 0064:13-26).
- spent:
  - The receipt act.
  - Its one live clause, "A future product decision must define and authorize any migration from `behavior_state`", was discharged by decision-0066 (0066:21-23 names itself as that decision). Tree: `plugins/trellis/surfaces.json` is gone, and `cli/surface_matrix_guard_test.go:85,286` forbids its return.
- reason: A receipt of a family strategy that no longer exists; its only live clause was answered by 0066.
- confidence: medium. It could equally be sorted as repealed, since its upstream was retired. Either way nothing moves to a spec.
- flags: none

### decision-0065 — the plugin path delivers, the install path vendors
- cluster: delivery
- disposition: mixed
- in force:
  - Decided 2 / §What the hook delivers: on the plugin path the plugin's own `SessionStart` hook injects the rules from the plugin's payload; the consumer holds no copy; the invariants pointer is repointed at the plugin's copy. Home: `staleness.sh:22-33`, `:799-1009` and `codex-context.mjs:899-974`. Verified.
  - Decided 3 (surviving half): the plugin path never vendors, and `install.sh` copies the plugin bundle into `.claude/skills/trellis/`. Home: `install.sh:26-35,280-282` and `plugins/trellis/README.md:24`. Verified.
  - Decided 4 (Claude half): where a `.trellis/internal/` overlay exists, the Claude hook injects nothing and runs the staleness comparison instead. Home: `staleness.sh:442-529` and `plugins/trellis/README.md:27-29`. Verified.
  - Decided 5 / §One model: the plugin path offers no delivery-model choice and never writes a file the project authored. Home: both hooks have no write path (`codex-context.mjs` has no write calls; `staleness.sh` only emits, and its remedies say "this hook advises, it never authorises a deletion"), and `plugins/trellis/README.md:24`. Verified.
  - Decided 6: both hosts carry a `SessionStart` hook that reads the plugin payload when no overlay is present. Home: `plugins/trellis/hooks/hooks.json`, `codex-hooks.json` and `codex-context.mjs`. Verified as behaviour, but see flags on support.
  - §What the hook delivers (residue): a project that never adopted Trellis is never governed by surprise. Home: the user-scope announcement at `staleness.sh:770-796`. Verified.
  - §Both hooks: the `.trellis/internal/` directory, not a file inside it, decides vendored mode, and a partial overlay fails loudly rather than switching modes. Home: `staleness.sh:442-446`, `:481-515` and `codex-context.mjs:794`. Verified.
  - §The envelope: every `SessionStart` emission uses the nested `hookSpecificOutput` envelope, and contract tests decode only that shape. Home: `staleness.sh:72-79,118-121` and `cli/plugin_hook_test.go:114-130`. Verified.
  - The measured boundaries are `startup` and `resume` on Claude; nothing else is claimed. Home: `plugins/trellis/hooks/hooks.json` (matcher `startup|resume`) and `plugins/trellis/README.md:57-62`. Verified.
  - Acceptance criteria: the install bundle manifest advances in the same commit as any payload change. Home: `cli/install_script_test.go:233` (`TestInstallScriptBundleManifestIsCurrent`). Verified.
  - §Supersession on 0010: the Codex hook needs Node, and the install path is the runtime-free route. Home: `plugins/trellis/README.md:50-51` (Node.js 20) and `staleness.sh:65-66`. Verified.
- retired:
  - Decided 1 / §Setup's one job (setup writes exactly one file, presets, diff-before-overwrite): vacuous under decision-0072. Presets are also retired by the dated note of PR #316.
  - Decided 3's "`install.sh` … never configures … never touch[es] `.trellis/`": retired by decision-0070 D2 (`install.sh` seeds `.trellis/rules.toml`, `install.sh:975-1000`) and by decision-0068 D12's reads.
  - Decided 4's "the hook … instead of injecting" on Codex, and §Rules arrive exactly once's "A project that still has an overlay … the hooks inject nothing": retired by decision-0093. Codex reads the overlay and injects it (`plugins/trellis/README.md:29-31`).
  - §What the hook delivers' "`.trellis/rules.toml` is the opt-in signal … never governed": retired in part by decision-0070 D3 and D4 and decision-0077. A vendored bundle adopts without a file (`staleness.sh:770-777`); "by surprise" survives.
  - §What the hook delivers' "the posture header": retired by the dated note of PR #316; one header ships.
  - §Both hooks' "take the same two paths": retired by decision-0068 (Claude path C and the coexistence branch).
  - §Both hooks' "every payload file within is required": narrowed by decision-0093 rule 1. `invariants.md` is consulted and repointed, not required (`staleness.sh:504` requires only `version`, `trellis.md` and `rules.md`).
  - §Rules arrive exactly once's "There is no state in which both paths carry the payload": retired by decision-0068 (the coexistence branch at `staleness.sh:449-479`).
  - §Supersession's "Trellis's own `CLAUDE.md` keeps its block": retired by decision-0071 (`CLAUDE.md` is `@AGENTS.md`).
  - §The gap's "install path delivers nothing": resolved by decision-0068.
  - Acceptance "`install.sh` never reads or writes `.trellis/`": retired by decision-0068 D12 and decision-0070 D2.
  - Acceptance "setup issues no write … enforced by test" and "setup requires confirmation": vacuous under decision-0072.
- spent:
  - Parked items: the install-path shape was decided by 0068; M2 morph issue #200; Codex live verification #199.
  - §Standing on `research-0013`.
  - Supersession pointers onto 0010, 0035, 0043, 0049, 0050, 0051, 0053 and 0057.
  - The staleness-envelope fix itself (done).
  - M1/M2 vocabulary retired for the plugin path; morph never returns to setup (setup gone).
  - The `specs/0004` note (specs retired, decision-0079).
- reason: Hook injection from the plugin payload, no vendoring and no edits to project files on the plugin path, the directory discriminator and the nested envelope all still hold. Setup, the strict install/config split and host symmetry do not.
- confidence: medium. The record is long, with many clauses retired piecemeal by 0068, 0070, 0072, 0093 and TRL-97. Each was checked against the tree, but a clause-level miss is possible.
- flags:
  - Decided 6 ("Both hosts deliver") is true as behaviour, but `README.md:130-135` says Codex is "carried, not supported". A spec must separate "the Codex hook exists and delivers" from any support claim.
  - `plugins/trellis/README.md:19-20` still says the hook "applies the shipped defaults when it is absent at project scope". Since TRL-97 no default rows ship, and the behaviour is "every rule applies" (`staleness.sh:770-777`). This is a wording lag in a file outside this batch.

### decision-0066 — retire Trellis's surface matrix
- cluster: delivery (secondary: checks)
- disposition: mixed
- in force:
  - §1 (standing half): Trellis carries no surface-matrix file, no exact surface rows, no `behavior_state` and no marketplace-observation record. Home: `cli/surface_matrix_guard_test.go:85` (`TestNoSurfaceMatrixFile`) and `:286` (`TestNoMatrixFieldsInCode`). Verified.
  - §2 / AC4: `plugins/trellis/README.md` carries one paragraph naming the hosts Trellis is known to work on, the check that establishes that, that support is not claimed, and the marketplace hedge, all under a test. Home: `plugins/trellis/README.md:55-91` and `cli/plugin_readme_test.go:55-70` (`TestPluginReadmeStatesHostSupportClaim`). Verified.
  - AC5: repo-relative links in the plugin README must resolve. Home: `cli/plugin_readme_test.go:81` (`TestPluginReadmeLinksResolve`). Verified.
  - §3 (general): a bundle file and its `install.sh` manifest entry change in the same commit, because `curl | sh` fetches from a moving `main`. Home: `cli/install_script_test.go:233` and the manifest at `install.sh:334-361`. Verified.
  - AC2 / §4: JSON UTF-8 and surrogate hardening stay tested against the plugin-manifest fixture. Home: `cli/plugin_package_test.go:50,115-123,175` (`TestPackageValidatorsRejectMalformedMetadata`). Verified that the test and validators exist; the fixture was not mutation-checked.
- retired:
  - §5 "`VERSION` stays `0.2.0`" (a point-in-time judgment): overtaken, now `0.24.0` under the `release-guard.yml` practice.
  - Consequences' "Trellis diverges from Grove, deliberately": moot since decision-0076 retired Grove.
  - §Supersession's "the `codex-cli-local-startup` behavior claim remains supported": contradicted by the tree (`README.md:132`, `plugins/trellis/README.md:83`) and by this record's own AC4 ("support is not claimed").
- spent: deleting `surfaces.json` and its manifest line atomically; removing the matrix Go machinery; the supersession pointers onto 0061, 0063 and spec-0005; the README `:9-11` link fix.
- reason: The guards against a surface matrix and the tested host-support paragraph still hold. The deletion itself is done.
- confidence: high
- flags: **Internal contradiction.** The supersession prose (0066:176-178) says the Codex claim "remains supported", while AC4 (0066:270-274), which the tree implements, requires "that support is not claimed". The README follows AC4.

### decision-0068 — the install path delivers rules through `.claude/rules/`, Claude only
- cluster: delivery
- disposition: mixed
- in force:
  - D1: `install.sh` renders exactly one rules file, `<repo-root>/.claude/rules/trellis.md`, in project scope only. `--scope personal` renders none and says so. It registers no hook, edits no settings file and creates no symlink. Home: `install.sh:419-429`, `:456`, `:1040-1044` and `plugins/trellis/README.md:73-80`. Verified.
  - D2 (residue): the rendered file is a real file, not a symlink. Home: `install.sh:861-900` (temp file plus move). Verified.
  - D3: rows stay live because the rendered file imports `@../../.trellis/rules.toml` rather than inlining a snapshot. Home: the `install.sh` render footer and `staleness.sh:626`. Verified.
  - D4: the import form is pinned by a test, including the negative for the sibling `@rules.toml` form. Home: `cli/install_script_test.go:1036-1040`. Verified.
  - D7: the install path claims Claude Code only, and other hosts get nothing from it. Home: `install.sh:1042-1044`, `plugins/trellis/README.md:76-80` and `cli/plugin_readme_test.go:174` (`TestPluginReadmeInstallPathClaimIsCurrent`). Verified. The prose authority sentence at `plugins/trellis/reference/rules.md:1` remains the fallback where rows cannot be imported (reworded under TRL-97).
  - D8: `install.sh` runs on macOS, Linux and WSL only (POSIX `sh`). Home: `plugins/trellis/README.md:80-83`. Verified.
  - D10: the Claude hook stands down when `.claude/rules/trellis.md` exists; a both-shapes project is reported, not double-delivered. Home: `staleness.sh:548-671` and `:449-479`. Verified.
  - D11: `/trellis:remove` enumerates the rendered file and deletes it before `.trellis/`. Home: `plugins/trellis/skills/remove/SKILL.md:121,138-149` and `cli/plugin_readme_test.go:112` (`TestRemoveSkillEnumeratesTheRenderedRulesFile`). Verified.
  - D12 (residue): `install.sh` refuses to render when the project already delivers statically: a `.trellis/internal/` overlay, a legacy flat `.trellis/trellis.md`, or a column-0 managed block in `CLAUDE.md` (or in an `AGENTS.md` that `CLAUDE.md` imports). It patches no marker. It hard-refuses a non-regular `.claude/rules/trellis.md` before the move. Home: `install.sh:525-529`, `:642-644`, `:861-863`. Verified.
  - D12 (budget principle): project-file content reads are counted and pinned by a test, so a change to the count has to argue itself. Home: `cli/install_script_test.go:1607` (now exactly three content reads: the marker, the `governed` key and the `@AGENTS.md` gate) and decision-0090. Verified.
  - Open question 4 (closed): the rendered file's last line is `<!-- trellis:rendered-from payload@<stamp> -->`, and the hook compares it to the installed plugin's stamp and nudges on mismatch. Home: `staleness.sh:627`, `:647-668`. Verified.
  - D13 / Consequences: the bundle is still vendored, and the rendered pointer names `.claude/skills/trellis/reference/invariants.md`. Home: `install.sh:280` and `:886`. Verified.
  - Consequences: the install artifact itself, not `rules.toml`, is the discriminator between the two Claude delivery mechanisms. Home: `staleness.sh:555-558`. Verified.
- retired:
  - D1's "`install.sh` writes that file and nothing else new … never touches `.trellis/`": retired by decision-0070 D2 (seeds `rules.toml`, `install.sh:975-1000`).
  - D2's "`/trellis:setup` owns `.trellis/`": retired by decision-0072.
  - D5's constant-header ruling: retired by decision-0088, which was then overtaken by TRL-97 (one header).
  - D5's added sentence making `strictness` authoritative: retired by the dated note of PR #316.
  - D9's "a bare install yields exactly `floor-transparency` and `floor-intent-gate`": retired by the dated note of PR #316.
  - D9's "writes no `rules.toml`": retired by decision-0070 D2.
  - D9's "makes no posture choice": reversed by decision-0088, then moot under TRL-97.
  - D9's "directs the reader to `/trellis:setup`": retired by decision-0072.
  - D12's opening paragraph, the `AGENTS.md` half of the `:293` row, and `:320-321`'s "is refused" for an `AGENTS.md` block that `CLAUDE.md` does not import (that shape is now rendered over): retired by decision-0091.
  - D12's `:294` existence-read row and the `:298-300` "exactly one reads a file's contents … selects no posture": retired by decision-0088, TRL-38, TRL-44 and TRL-97. The count is now three (`cli/install_script_test.go:1607`).
  - D12's "nothing under `.trellis/` is written": retired by decision-0070 D2.
  - D12's "`/trellis:setup` does not [clean an inline block]": vacuous under decision-0072.
- spent:
  - D6 (argument that 0058 phase 4 is satisfied by D10).
  - The six measurement runs.
  - D2's withdrawal of the symlink design.
  - Open questions 1-3 (ruled into D1, D5 and D11).
  - Open question 6's `0.3.0` → `0.4.0` bump.
  - The supersession pointer onto 0065.
- reason: The Claude-only, project-scope rendered rules file, with live imported rows, test pins, hook stand-down, remove ordering, static-delivery refusals and a staleness stamp, still holds. The posture and strictness wording, the floors-only bare install and "never touches `.trellis/`" do not.
- confidence: medium. D12's read inventory was changed by Linear-issue work (TRL-37, TRL-38, TRL-44, TRL-97) as well as by 0088, 0090 and 0091. I sorted from the current test and `install.sh`, not from a single retiring record.
- flags:
  - **Open question 5 is still open with no named consumer.** Whether the vendored skills load at all is "open in full" per `decision-0073:56,101-103`, and `plugins/trellis/skills/remove/SKILL.md:192-199` is written to be true under any answer. decision-0078 would want an owner.
  - **Tension:** Open question 6's "establishes no policy and binds no future bump" (0068:441-442) versus `.github/workflows/release-guard.yml` and `AGENTS.md`'s "A payload change is a release".
  - After TRL-97 the rendered header is again a single shipped constant (`plugins/trellis/reference/trellis.md:5`, "By default"). D5's outcome is effectively current again, on a different ground, so a spec should state the TRL-97 ground, not D5's.

---

## Tally

| Disposition | Count | Records |
|---|---|---|
| in-force | 1 | 0055 |
| mixed | 11 | 0051, 0053, 0054, 0057, 0058, 0060, 0061, 0063, 0065, 0066, 0068 |
| spent | 2 | 0062, 0064 |
| repealed | 2 | 0052, 0059 |
| **total** | **16** | 0056 and 0067 do not exist (explained at top) |

| Cluster | Count | Records |
|---|---|---|
| rule-model | 3 | 0051, 0052, 0053 |
| delivery | 8 | 0054, 0055, 0058, 0059, 0061, 0065, 0066, 0068 |
| method | 1 | 0057 |
| checks | 0 | (0066 is secondary checks) |
| positioning | 4 | 0060, 0062, 0063, 0064 |

## Medium and low confidence calls

All eight are medium; there are no low calls.

- **0051:** whether rule 6 (the `kodhama-0007` rule-4 divergence) and the "profile" name reservation are standing rules; both are stated only in the record.
- **0053:** whether "the tested wording is the shipped wording" survives as a principle after TRL-97, given that 0093-0096 still cite it.
- **0058:** the Codex "supported" claim was retired by upstream records and README wording, not by a Trellis record; the status of `block-codex.md` (manual-copy payload or live shape) is unclear.
- **0060:** whether the §2 negatives and the §4 Stewards boundary are standing rules or historical scope.
- **0061:** the §5 "no … bump policy" versus the current `release-guard.yml` bump rule.
- **0063:** the catalog state lives in `kodhama/stewards`; the retirement of "supported" happened upstream.
- **0064:** spent versus repealed; nothing moves either way.
- **0065 and 0068:** long records retired clause by clause by 0068/0070/0072/0088/0090/0091/0093 and TRL issues; a clause-level miss is possible.

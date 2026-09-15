# Sort — batch 4: decision-0069 … decision-0084

Verified against the worktree `simplify-1b` (branch `feature/trl-103-decisions-become-history`, tree after #316), 2026-09-15. "verified" means the named file was read and does what the point says. Line numbers are as of that tree.

Convention used for disposition: **in-force** when every Decision point still holds and the only spent items are bookkeeping consequences (pointer marks, VERSION bumps, a README edit); **mixed** when at least one Decision point is retired or was purely a one-time act.

---

### decision-0069 — the manual copy path is for uncovered harnesses
- cluster: delivery
- disposition: in-force
- in force:
  - D1 — the manual copy path is retained and scoped to harnesses the plugin does not cover. It is not legacy, and `staleness.sh`'s path A (overlay check) is not cleanup debt — home: `README.md:161-167`, `README.md:263-265` ("decision-0069 retains … it is not legacy"), `plugins/trellis/hooks/staleness.sh:8-20` and `:442-546` (path A present) (verified). "A prune that deletes either must be a decision" is stated only in the record.
  - D2 — on Claude Code the manual path is superseded, and the README says so where it gives the instruction. On Codex CLI it is not superseded: it is Codex's only bootstrap until a real channel exists — home: `README.md:161-165`, `plugins/trellis/README.md:256-260`, `README.md:132` (Codex unsupported) (verified)
  - D3 — the curl path and the manual path are mutually exclusive: `install.sh` refuses to render over a hand-built overlay or a managed block — home: `README.md:165-167`, `install.sh:564-644`, `cli/install_script_test.go:1415` `TestVendorRefusesForEveryStaticDeliveryShape` (verified)
  - D4 — the hook's overlay migration nudge stays — home: `staleness.sh:526`, `:541-544` (verified). Its text now carries manual steps instead of `/trellis:setup` (decision-0072).
- retired: none
- spent: the README scope sentence and exclusivity note (`README.md:161-167`)
- reason: Every rule here still holds in the README and the hook. The open questions it names are still unanswered.
- confidence: high
- flags:
  - Open question 1 is still unresolved. The manual path has no documented staleness-compare procedure: `README.md:169-233` only verifies the copy's checksums. So decision-0035's staleness floor is claimed but unmet on uncovered harnesses, and this is stated only in the record.
  - Open question 2 is also unresolved: a covered and an uncovered harness sharing one hand-built overlay. Accepting the migration on the covered host deletes the uncovered host's only delivery. Stated only in the record.

### decision-0070 — adoption is the consent act, not installation
- cluster: delivery (secondary: rule-model, for D5)
- disposition: mixed
- in force:
  - D1 — the unit of consent is adoption, and every path has an adoption act:
    - curl: running `install.sh` in the repo
    - project-scope plugin: the bundle at `<repo>/.claude/skills/trellis/`
    - user-scope plugin: the install itself, announced per project
    - home: `staleness.sh:43-56`, `:727-797`; `install.sh:53-56`; decision-2026-09-15 point 8 (verified)
  - D2 (surviving half) — `install.sh` seeds `.trellis/rules.toml` when none exists, which amends decision-0065's "install.sh never touches `.trellis/`" — home: `install.sh:53-56`, `:470`. The content is now the two-line file (decision-2026-09-15 point 7), and the seeded bytes are pinned to what the hook's announcement quotes: `cli/plugin_hook_test.go:166`, `cli/docs_consistency_test.go:523` (verified)
  - D3 — a project-scope bundle with no rules file counts as adopted, so every rule applies, and the hook writes nothing — home: `staleness.sh:770-777`; `cli/plugin_hook_test.go:1678` `TestProjectScopedPluginGovernsWithoutRulesToml`, `:3083-3091` (verified)
  - D4 (surviving) — a user-scope plugin in an unadopted project announces once per session:
    - it names the project and injects no rules that turn
    - decline → a file holding only `governed = false`
    - accept → the agent writes the file
    - "the hook never writes"
    - home: `staleness.sh:794`; `cli/plugin_hook_test.go:1812` (verified)
  - D5 — `governed = false` is a top-level key, read before every path. Not governed means no rule applies, the floors included, and this holds on both hosts. Where a static file already loaded the rules, the hook emits one DISREGARD override that names the fix — home: `staleness.sh:405-437`; `plugins/trellis/hooks/codex-context.mjs:745-784`; `install.sh:476-499`, `:656-673`; tests `cli/codex_hook_test.go:1043`, `cli/plugin_hook_test.go:190`, `cli/install_script_test.go:2634` (verified)
  - D5 principle — floors bind configuration, not adoption — home: `staleness.sh:278-293` comment, and the TRELLIS_NOT_GOVERNING text at `:435` (verified)
  - D6 — project scope has exactly one location: the plugin root resolves under `<project>/.claude/skills/` and the project root is not `$HOME`. When the hook cannot tell, it treats the project as unadopted and announces — home: `staleness.sh:734-768`; `plugins/trellis/skills/remove/SKILL.md:55-57` (verified)
  - D7 — D3 and D4 apply to Claude only. On Codex the config file is the adoption signal (`nearestOverlay` walks up for `.trellis/rules.toml`). D5 holds on both hosts — home: `codex-context.mjs:238-253`, `:729-733`; `README.md:140-142`; `plugins/trellis/README.md:144-147` (verified)
  - Open question 1 ruling — a decline persists in the repository and is committed, because it is a project fact reviewed in the diff, never a machine-local list — home: `staleness.sh:794` tells the agent to write it into `.trellis/rules.toml` (verified). The reasoning is stated only in the record.
- retired:
  - D2 "seeds from `reference/rules-b.toml` … 14/14 at posture B" — by dated note PR #316 (decision-2026-09-15 point 7)
  - D4 "accept, or no objection → seed", "silence never reads as refusal", and the opt-OUT self-description — by decision-0077. The rules-b seeding on accept went with dated note PR #316.
  - D4 "writes stay agent-mediated and human-consented", as re-read for row repair — first by decision-0083 (frontmatter). That re-reading now has no subject: dated note PR #316, since nothing reconciles.
  - D3 "all-14, posture B" (the posture half) — tree: `staleness.sh:799-800` ("strictness selects nothing any more"). Not named in the dated note; see flags.
  - Consequence "`/trellis:setup` becomes the posture/rows tool" — by decision-0072 (skill retired)
- spent: the `superseded_in_part_by` marks on decision-0065 (two clauses); the landing-page `# then run /trellis:setup` demotion
- reason: The adoption model, the opt-out, scope detection and Codex's file-as-signal rule all still govern the hooks. The preset seeding and silence-adopts halves were retired.
- confidence: high
- flags:
  - The D3 posture half ("posture B") is false in the tree, but the PR #316 dated note names only D2 and D4.
  - Open question 2 is now load-bearing. The eval never tested "no rows at all", yet since decision-2026-09-15 a missing row means the rule applies in every project. That record's own open question confirms the new wording is unmeasured.

### decision-0071 — Trellis self-applies through the plugin, not an overlay
- cluster: method
- disposition: in-force
- in force:
  - D1 — this repo holds no vendored overlay: no `.trellis/internal/`, no legacy flat overlay or stamp, no CLAUDE.md managed block, no AGENTS.md Codex bootstrap. `.trellis/rules.toml` stays — home: `cli/selfapply_test.go:97-102`, `:165-176`; `cli/selfapply_test.go:26-53` `TestRepoDeclaresRulesConfig`; the root `.trellis/` holds only `rules.toml` (verified)
  - D2 — self-application goes through the plugin path, declared at project scope. The accepted cost is that the repo dogfoods the last released plugin, not HEAD — home: `.claude/settings.json:3` (`trellis@kodhama`); `README.md:308-324`; AGENTS.md "delivered live by the Trellis plugin at session start" (verified)
  - D3 — the guards that remain: vendored payload == generator render, and `install.sh`'s manifest pins what a consumer receives — home: `cli/payload_test.go:520` `TestVendoredPayloadIsCurrent`; `cli/install_script_test.go:233` `TestInstallScriptBundleManifestIsCurrent`; rationale at `cli/selfapply_test.go:11-25` (verified)
  - D4 — CLAUDE.md is exactly `@AGENTS.md` — home: `CLAUDE.md`; `cli/selfapply_test.go:97` (verified)
  - D5 — this repo is not governed on Codex until Codex has an install channel. The gap is accepted — home: `README.md:313-316` (verified)
  - D5 clause — decision-0058 point 4 (Codex fallback retained) lapses for this repo only — stated only in the record
- retired: none
- spent: deleting `.trellis/internal/`, the CLAUDE.md block and the AGENTS.md bootstrap; deleting `TestRepoOverlayIsCurrent` (noted at `cli/selfapply_test.go:11`); the pointer on decision-0035; unblocking #219
- reason: The repo still self-applies only through the plugin, and tests pin the absence of every overlay shape.
- confidence: high
- flags:
  - Open question 1 is partly answered in the tree. `TestRepoDeclaresRulesConfig` (`cli/selfapply_test.go:26-53`) now pins that this repo switches no rule off, which is the nearest thing to CI proof that the repo runs its own rules.
  - "#220" in D5 is a dead GitHub pointer. Codex support is tracked in Linear (`README.md:135`).

### decision-0072 — retire `/trellis:setup`
- cluster: delivery (secondary: checks)
- disposition: mixed
- in force:
  - point 1 — `/trellis:setup` is retired, and only the remove skill ships — home: `plugins/trellis/skills/` (only `remove/`); `cli/payload_test.go:436` `TestOnlyTheRemoveSkillShips` (verified)
  - point 3, rows 1 and 3 — `/trellis:remove` is the clean exit, and the installer never runs git — home: `plugins/trellis/skills/remove/SKILL.md`; `install.sh:67` (verified)
  - point 4 — discoverability is an accepted cost. If newcomers appear, the answer is a documented file format, not a resurrected skill. Stated only in the record; the format is documented at `README.md:118-128`.
  - point 5 — no `/trellis:migrate` skill — stated only in the record (the tree agrees)
  - finding #6 rule — retiring a confirm-gated writer must not silently retire its gate. Every destructive instruction a hook emits must carry a confirmation clause. The guard scans a verb list and asserts a floor on how many messages it matched — home: `cli/plugin_hook_test.go:588-602` (`destructiveVerbs`), `:765` `TestEveryDestructiveInstructionIsGated`, `:969` `TestEveryDeletionInstructionIsGated`, floor at `:840-851` (verified)
  - sweep — the docs guard walks the whole tree rather than a hand list, so it covers hook output — home: `cli/docs_consistency_test.go:34-75` (`docSurfaces`, `WalkDir`), `:163` `TestDocsClaimOnlyRealCommands` (verified)
  - sweep — no bare-word "setup" claims, and every allowed form in the exemption list names its reason — home: `cli/docs_consistency_test.go:207-226` `TestNoUnqualifiedSetupClaims` (verified)
  - sweep — the generator boundary: a fix to a generated payload file goes into `cli/apply.go` and is re-rendered, never hand-edited — enforced by `cli/payload_test.go:520` (verified). The rule's wording is stated only in the record.
  - sweep — the overlay nudges carry the manual steps (delete the overlay, keep `.trellis/rules.toml`), and the overlay branches stay for old checkouts — home: `staleness.sh:17-18`, `:526`, `:544` (verified)
- retired:
  - point 2 — the three replacement shapes, "the copy step is mandatory", and the frontmatter's "remains the advice wherever the posture or the row set matters" — by decision-0083 §5 and decision-0084 (a partial file governs on both hosts), then dated note PR #316 (no presets, `strictness` ignored, re-enabling writes the two-line file)
  - point 3, row 2 "the hook fails loudly on a mismatch" — by decision-0083, then decision-2026-09-15 points 3 and 5 (a bad entry is ignored with a warning)
  - the sweep's "the remedy repairs rows in place and gates any reseed" — by decision-0083, then decision-2026-09-15 point 5
- spent: `plugins/trellis/skills/setup/` deleted (absent); the 31-reference sweep; the decision-0065 "vacuous" marking
- reason: The skill stays retired, and its guard lessons live on as tests. Its preset-copy replacement recipe is gone with the presets.
- confidence: high
- flags:
  - The guard-coverage lesson ("a guard written against the spelling in front of you covers one spelling; assert a floor") is carried by test floors but stated in prose only in the record.
  - Open question 1 (do the overlay branches earn their keep?) is still open.

### decision-0073 — the delivery shapes are a closed set
- cluster: delivery (secondary: checks)
- disposition: mixed
- in force:
  - D1 — delivery states form a closed set:
    - S0: unadopted or config-only
    - S1: rendered file
    - S2: internal overlay
    - S3: legacy flat overlay
    - S4: inline block in the five documented files
    - S5: project-scope bundle, except at `$HOME`
    - S6: M2 morph markers
    - The none-state is a member. A component that covers only a subset must say so where it does it, with a pointer. Adding a state takes a decision, never a code change alone.
    - home: `plugins/trellis/skills/remove/SKILL.md:8`; `cli/remove_skill_test.go:71-140` `TestRemoveSkillCoversEveryDeliveryState`; subset comments at `staleness.sh:333-352` and `install.sh:564-615` (verified). "Adding a state needs a decision" is stated only in the record.
  - D2 — `staleness.sh` probes for a column-0 `<!-- trellis:begin`. The probe feeds the `governed = false` disregard, the coexistence alarm and a refusal before path B, and the refusal is worded for both the embedded and the dangling-import state — home: `staleness.sh:333-403`, `:421-437`, `:469-477`, `:683-694`; `cli/plugin_hook_test.go:1836` `TestStalenessHookHandlesInlineManagedBlock` (verified)
  - D3 (surviving) — `/trellis:remove` inventories, reports and removes every state:
    - the bundle is deleted behind confirmation, with or before `.trellis/`
    - the decline artifact is surfaced by name, with its consequence
    - morph detection happens in preflight, before any write
    - the no-op predicate reports "already absent" only in S0-unadopted
    - home: `SKILL.md:40-63`, `:124-129`, `:170-212`; `cli/remove_skill_test.go:148`, `:177`, `:219` (verified)
  - D4 — each component class gets its own assertion mode: shell components get a behavioural fixture per state, and the skill gets textual assertions (prose tests pin text only) — home: `cli/remove_skill_test.go`; `cli/install_script_test.go:1415`; `cli/plugin_hook_test.go:1836` (verified)
- retired:
  - D3 bullet 2, the "one ignored prompt re-governs at 14/14" half — by decision-0077 (the corrected wording is at `SKILL.md:40-45`)
- spent: D5's acceptance criteria AC1–AC3 (discharged by the tests above); the `install.sh` pointer comment (`install.sh:564-566`); the README inline-recipe fix (`README.md:184-208`, "exactly one delivery branch"); the headless hook-firing measurement for a vendored bundle; the revision record and self-check (history)
- reason: The closed set, the hook probe and the remove skill's coverage are all live and pinned by tests. Only the silence-adopts half of one bullet was retired.
- confidence: high
- flags:
  - The tree narrows D2: AGENTS.md is probed only when CLAUDE.md is exactly the `@AGENTS.md` adapter (`staleness.sh:390-398`, comment at `:371-380`). The spec must state the narrowed rule, not the record's two-file wording.
  - Unmeasured and parked: whether an interactive session fires a vendored bundle's hook. The skills half of decision-0068 open question 5 is also still open.
  - S6 stays in the set although the M2 morph is retired (`README.md:265`).

### decision-0074 — `inv-deliberate-succession`
- cluster: rule-model
- disposition: mixed
- in force:
  - point 1 — `invariants-v1` carries `inv-deliberate-succession` (operating set, `trellis-design`, provisional). It covers both directions and names `inv-self-improvement` as its neighbour — home: `core/invariants/trellis-invariants-v1.md:206-226` (verified)
  - point 2 — `inv-self-improvement` is back to its reactive face, with a pointer left where the lean used to be — home: `core/catalog/signature-catalog-v1.md:195-204` (verified)
  - point 3 — the catalog entry has a two-direction directive and four honored/four violated pairs — home: `core/catalog/signature-catalog-v1.md:216-251`; `plugins/trellis/reference/rules.md:19` (verified)
  - point 4 — the wording is our synthesis — home: AGENTS.md "Naming guardrail" (verified, general)
  - consequence (standing) — a new slug is a set amendment, not a catalog-only door, and the derived chain follows in the same change — home: `cli/row_set_guard_test.go:80-123` ("a slug is a set amendment"); `core/catalog/signature-catalog-v1.md:55-80` (verified)
- retired:
  - Consequences render chain "preset rows in `rules-a.toml`/`rules-b.toml` (without a row the rule ships but is inactive)", and the next bullet's "false all-clear" on a re-run — by dated note PR #316. A rule with no row applies: `plugins/trellis/reference/rules.md:1` (verified).
  - contract-chain members `spec-0007`, `spec-0002` §1 and the `corpus-reviewer` checklist — by decision-0079 (specs deleted) and decision-0098 (corpus-reviewer retired)
- spent: the `superseded_in_part_by` mark on decision-0052; the 14 → 15 count move; the `profiles/trellis-self.md` row; the `spec-0007` version bump
- reason: The invariant and its catalog entry ship today. Only the preset-row plumbing it described is gone.
- confidence: high
- flags:
  - The consequence "decision-0027's amendment loses its only worked example" still stands as a known dangling example.
  - The open question "does it pull its weight?" was routed to the compression audit (trellis#239, then TRL-4). Its state in Linear could not be verified.

### decision-0075 — Linear tracks the work
- cluster: method
- disposition: mixed
- in force:
  - point 1 — Linear (Kodhama workspace, team Trellis, `TRL-*`) tracks the work; GitHub keeps pull requests, CI and code — home: AGENTS.md "Where work lives" (verified)
  - point 2 — the Linear taxonomy:
    - type is the Bug/Feature/Improvement labels, and stage is Linear's workflow states
    - `facing:` is dropped
    - severity maps blocker→High, broken-feature→Medium, papercut→Low; `Critical` is only for outage, data loss or security exposure
    - the `kodhama:issues` convention does not apply here
    - stated only in the record
  - point 3 — ideas live in a document, not in issues, and each entry carries its promotion trigger — home: AGENTS.md "Where work lives" (verified)
  - point 5 — a Linear issue carries its own content (summary plus link). The day a repo artifact exists behind it, the issue body must become a pointer — stated only in the record
  - consequence — resolve a Linear team by id, never by name — home: AGENTS.md "Where work lives" (verified)
  - consequence — research issues carry no type label — stated only in the record
  - consequence — trellis is a declared exception to `kodhama-0026-issue-taxonomy` until the family moves — stated only in the record
  - consequence — old GitHub issue numbers must keep resolving; no record citation gets rewritten — stated only in the record
- retired: none
- spent:
  - point 4 — the migration banners on closed GitHub issues (`<!-- trellis:linear-migration -->`), on GitHub and not in the tree
  - the backlog triage (27 → 19 ported)
  - `stage:` labels not ported
- reason: Linear is still the tracker and AGENTS.md carries the core rules, but the taxonomy and the issue-content rule exist only in this record.
- confidence: medium — the banners and the Linear-side taxonomy live outside the repository and could not be checked.
- flags:
  - Points 2 and 5 and the no-type-label rule have no home in the tree. A spec must state them or they vanish with the record.
  - "The Kodhama team is owed an archive" names no consumer (a possible orphan under decision-0078).

### decision-0076 — retire grove
- cluster: method
- disposition: mixed
- in force:
  - point 1 — grove-the-plugin is retired. No instruction routes to `grove:<role>` or `/grove:*`; `.claude/settings.json` must not enable `grove@kodhama`; `.grove/` and the AGENTS.md grove block must not return — home: `cli/selfapply_test.go:135-144`, `:196-198`, `:210-219`; `.claude/settings.json`; AGENTS.md "Grove is retired" (verified)
  - point 2 — grove-the-repo citations (`grove/adr-*`) stay live, kept valid by registry membership, not by the repo's liveness — home: AGENTS.md "Grove is retired"; `docs/rubrics/artifact-contract.md:62-63` (registry lists grove) (verified)
  - point 3 harvest — the test, typecheck and lint commands and the category-scoped branch-naming rule live in AGENTS.md — home: AGENTS.md "Checks and review" (verified)
- retired:
  - point 6's standing rule "invoke `corpus-reviewer` before merging corpus changes", and its "no CI check" ground — by decision-0098 (frontmatter; the conformance tests run in CI per AGENTS.md)
  - point 3's "four corpus tokens need no home" bullet and point 5's repair of `.claude/agents/corpus-reviewer.md` — by decision-0098 (the agent file is deleted; `.claude/` holds only `settings.json`)
  - point 4, the `spec-0006` AC1 partial supersession — mooted by decision-0079 (`specs/` deleted)
- spent:
  - `.grove/` deleted (absent; asserted at `selfapply_test.go:196`)
  - the AGENTS.md block and the settings entry removed
  - the `core/README.md` correction (`core/README.md:13`)
  - point 5's owed carrier-pointer repair, since done: the rubric no longer says "plugin-carried", and `artifact-contract.md:78` cites grove's relations charter (verified)
- reason: Grove's retirement and the two-kinds-of-grove split still hold and are test-pinned. The corpus-reviewer rule it introduced was retired by CI conformance.
- confidence: high
- flags:
  - The open question "what stops the next operating model arriving as a chore commit?" has no guard. AGENTS.md's record test would now catch it only by judgment.
  - The self-check lesson "where else does this claim appear?" is method guidance stated only in the record.

### decision-0077 — silence is not an adoption act
- cluster: delivery
- disposition: in-force
- in force:
  - point 1 — silence never adopts. Under a user-scope install with no rules file, only an explicit act that writes the file governs. An unanswered announcement leaves the project ungoverned and recurs next session — home: `staleness.sh:779-795`; `plugins/trellis/README.md:23`; `cli/plugin_hook_test.go:1738` `TestSilenceNeverAdoptsAfterTheDeclineIsDeleted`; decision-2026-09-15 point 8 (verified)
  - point 2 — decision-0070 D4 keeps its announcement, the project naming, no rules that turn, the decline bullet, and "the hook never writes" — home: `staleness.sh:794`; `cli/plugin_hook_test.go:1812` (verified)
  - point 3 — the mechanism is announce-then-accept: the announcement is the disclosure, the file is the consent — home: `staleness.sh:779-789` (verified)
  - point 4 — deleting a recorded decline returns the project to unadopted, and the announcement comes back — home: `plugins/trellis/skills/remove/SKILL.md:40-45`; `cli/remove_skill_test.go:177-207` (verified)
  - point 5 — pinned by behaviour: two unanswered runs inject zero slugs, repeat the announcement and write no file — home: `cli/plugin_hook_test.go:1738` (verified)
- retired: none
- spent: the pointer marks on decision-0070 and decision-0073; the `staleness.sh` comment correction and manifest rehash
- reason: The hook, the remove skill and a behavioural test all carry this rule today.
- confidence: high
- flags:
  - The consequence "decision-0015's open question deliberately not settled" has since been answered by AGENTS.md (a dated note is the only text added to a record).
  - Codex has no announcement branch: adoption is the file (`codex-context.mjs:729-733`).

### decision-0078 — `inv-no-orphan-followups`
- cluster: rule-model
- disposition: mixed
- in force:
  - point 1 — `invariants-v1` carries `inv-no-orphan-followups` (provisional), with the address test, the three outcomes and "a plausible ledger is worse than none" — home: `core/invariants/trellis-invariants-v1.md:227` (verified)
  - point 2 — the catalog entry has three honored/three violated pairs (process, code, ops) — home: `core/catalog/signature-catalog-v1.md:253-287`; `plugins/trellis/reference/rules.md:21` (verified)
  - point 3 — a new slug is a set amendment — home: `cli/row_set_guard_test.go:102-105` (verified)
  - point 4 — the wording is our synthesis — home: AGENTS.md "Naming guardrail"
  - consequence — `assessableSlugs` is the one count pin. Every count assertion derives from `len(assessableSlugs)`, and the floor check is indexed from the end — home: `cli/payload_test.go:63-67`, `cli/rules_test.go:12-30`, `cli/remove_skill_test.go:187` (verified)
  - the repo applies the rule to itself ("record a next step — name the consumer … or drop it") — home: AGENTS.md table (verified)
- retired:
  - Consequences render chain "preset rows … without a row the rule ships but is inactive" — by dated note PR #316
  - the renumbering observation ("ids allocated by whoever merges first") — by decision-0089 (frontmatter). decision-0089 was itself later retired in full, and ids are now date plus slug (AGENTS.md naming rule); `.github/scripts/decision-id-guard.sh` is absent (verified).
  - "row-set mismatch fails loudly on both paths" — by decision-0083 and decision-0084, then decision-2026-09-15 points 3 and 5 (a bad entry is ignored with a warning)
  - "the curl path still lies (TRL-2)" — the hazard is gone: a rule with no row applies (`reference/rules.md:1`)
  - the contract-chain members `spec-0007`, `spec-0002` and `corpus-reviewer` — by decision-0079 and decision-0098
- spent: point 5 (VERSION 0.5.0 → 0.6.0; now `0.24.0`); the `profiles/trellis-self.md` row; the catalog derivatives note rewrite; the dropped fourth pair
- reason: The invariant, its catalog entry and the one-pin count discipline are live. The row-set and id-collision mechanics it described are gone.
- confidence: high
- flags:
  - **Drift:** the live catalog note at `core/catalog/signature-catalog-v1.md:77-78` still says `plugins/trellis/VERSION` is "unguarded — trellis#245 is still open". `.github/workflows/release-guard.yml:63` fails a payload change that has no bump, and AGENTS.md says so. This record's Consequences make the same (then-true) claim.
  - The inline 2026-09-03 note names `decision-id-guard.sh`/`.yml` and `cli/decision_id_guard_test.go` as the live guard. All three were deleted by #315, and this record has no dated note for that; only the decision-0089 pointer reaches it.
  - The open questions (the Assess consumer, the compression review) are still open.

### decision-0079 — retire the spec stage
- cluster: method
- disposition: mixed
- in force:
  - point 1 (surviving) — there is no spec stage, and `specs/` does not exist — home: AGENTS.md table ("the spec stage retired in decision-0079, and specs/ with it"); tree has no `specs/` (verified)
  - point 2 — `docs/decisions/` and `docs/research/` keep their `spec-*` citations as history — home: the AGENTS.md append-only rule; `docs/rubrics/artifact-contract.md:65-72` (verified)
  - point 3 — the artifact contract is self-standing in the rubric, and the typed-artifact schema lives at `core/schemas/typed-artifacts.md` (`schema-typed-artifacts`) — home: `artifact-contract.md:14-15`; `core/schemas/typed-artifacts.md:4-18` (verified)
  - point 3a — the retired-artifacts registry: `spec-0001`…`spec-0008` resolve as retired ids under rubric check 4 — home: `artifact-contract.md:65-67`; **the table itself is read out of this record** by `cli/corpus_conformance_test.go:652`, `:712-717` (verified). The "where its content lives now" mapping is stated only in the record.
  - point 5 — `spec` stays a recognized artifact type — home: `artifact-contract.md:106` (verified)
  - point 6 — `grove/adr-*` citations stay untouched — home: AGENTS.md (verified)
  - cost accepted — the `spec-0005`/`spec-0007` requirement ids survive as markers, and the tests are the executable contract — home: `install.sh:4`, `codex-context.mjs:4`, `cli/install_script_test.go`, `cli/codex_hook_test.go` (verified)
- retired:
  - point 1 "planning … with the superpowers skills" — by decision-0097 (Compound Engineering; AGENTS.md table)
  - Consequence "plan retention … dropped" — by decision-0085, now carried by decision-0097 point 7 (`docs/plans/` exists)
  - Consequence "`ratify-guard` no longer globs `specs/*.md`" — `ratify-guard` was deleted by decision-0082
- spent: `specs/` deleted; decision-0011 marked `superseded_by: [decision-0079]` (verified); current-truth citations re-pointed; the schema migration
- reason: The spec stage stays retired and its registry still resolves old ids. Its superpowers routing was replaced by Compound Engineering.
- confidence: high
- flags:
  - **Consolidation hazard:** rubric check 4's retired-id form (4e) is implemented by parsing decision-0079's 3a table out of the record file (`cli/corpus_conformance_test.go:712-717`). Turning 0079 into history, or moving that table, breaks check 4 unless the test moves with it.

### decision-0080 — the approval signal is the review, not the merge
- cluster: method
- disposition: repealed
- in force: none
- retired:
  - the whole record — by decision-0082 (`superseded_by`). The tree agrees: no `ratify-flip.yml` in `.github/workflows/` (verified).
- spent: `ratify-flip.yml` was built and then deleted with the record's retirement
- reason: decision-0082 retired the status field and both ratify workflows, so nothing here governs.
- confidence: high
- flags:
  - Its one surviving finding lives only in records: an agent merging with the maintainer's token is indistinguishable from him, so "the merge is the human act" holds only while the account is his alone. decision-0082 carries it forward; AGENTS.md states the bar (no agent merge without his act) but not the assumption.

### decision-0081 — supersession authority scaled by cost of reversal *(proposal)*
- cluster: method (secondary: rule-model)
- disposition: mixed
- in force:
  - accepted framing — who retires a recorded decision scales with its cost of reversal:
    - when being wrong would be cheap to undo, quick to notice and cheap to live with meanwhile, the agent drafts the replacement into the work, with the prior rationale surfaced and a forward mark left, instead of stopping to ask
    - the replacement still passes the ordinary gate (the merge)
    - when undoing would be expensive, a mistake slow to surface, or the damage while it stands serious, or when the agent cannot tell, it raises the question first
    - home: AGENTS.md:96 routes here by id only (verified). The rule text is stated only in the record.
  - the three-prong test — home: the AGENTS.md "Operating method" record-admission test (verified). AGENTS.md uses the prongs for *whether a change gets a record*, which the #313 plan says is a different question (`docs/plans/2026-09-14-1851-docs-decision-records-become-rare-plan.md:85`, `:165`). The plan leaves this record's "who may retire" claim unchanged.
  - no unilateral agent supersession: the human act at the merge gate stays, with no decision-0046 or `floor-intent-gate` amendment — home: AGENTS.md "Merging to `main` is the acceptance … An agent still may not merge on his behalf" (verified)
  - provenance — the one-way/two-way door framing is attributed to Amazon's 2015 shareholder letter, and its application to agent supersession is our synthesis — stated only in the record
- retired:
  - Context "supersession is performed by minting a successor decision" — by the AGENTS.md "Operating method" (TRL-98 PR #313, TRL-99 PR #315): partial retirement, and full retirement with no successor, now go by dated note
  - open question 7 (the draft-consumption seam) — dissolved by decision-0082
  - Recommendation 3's row-set trap analysis (the 14-row fleet, the unvalidated curl path, "a 16th slug breaks rows") — decision-2026-09-15 point 1: a rule with no row applies, and unknown rows are ignored
  - Recommendation 2's "`spec-0001` amendment as a local enrichment" — `spec-0001` was retired by decision-0079
- spent: none (a proposal that changed no file)
- reason: Its cost-of-reversal principle is what AGENTS.md cites for retiring records, but the rule text lives only here, and its proposed catalog wording was never applied.
- confidence: medium — the record calls itself a proposal whose accepted scope was the argument, and later records disagree about what it grants.
- flags:
  - **Accepted but never applied (possible orphan):** Recommendation 1 and the proposed directive, `why`, signature clause and *(decision)* pair for `inv-deliberate-succession`. The catalog directive (`core/catalog/signature-catalog-v1.md:221`) has no cost-of-reversal clause. The record says "still owed" and names no consumer in the repo.
  - **Contradictory readings in later records:**
    - decision-0083 §3 cites 0081 to say a cheap frontmatter pointer "is his", meaning the maintainer's call.
    - decision-0092:149 says supersession "is the act decision-0081 reserves for the maintainer", and decision-0098:267 follows decision-0092.
    - decision-0084's self-check reads 0081 the other way: cheap pointers are taken in the work by the agent. That matches this record's own recommendation.
    - The spec must pick one.
  - Its quotes and line cites predate the AGENTS.md trims; its own header note says so.

### decision-0082 — retire the `status` field; merging is the acceptance
- cluster: method
- disposition: mixed
- in force:
  - point 1 — `status` is not a contract field. The required set is `id / type / depends_on / owner`, and new artifacts carry no status — home: `docs/rubrics/artifact-contract.md:44-51`; AGENTS.md "There is no `status` field" (verified)
  - point 2 — merging to `main` is the only acceptance; unmerged artifacts may not be consumed — home: AGENTS.md "Operating method"; `README.md:326-327` (verified)
  - point 3 (surviving) — full retirement by a successor record uses `superseded_by`, and the conformance check keys on the pointer's presence — home: `artifact-contract.md:122-130`; AGENTS.md (verified)
  - point 5 — the intent gate does not move: a human approves, and an agent may not merge on his behalf without his act — home: AGENTS.md "Never tell the maintainer a change is blocked…" paragraph (verified)
  - point 6 — existing `status:` lines stay byte-for-byte as history and read as accepted; never add the field, never strip it — home: AGENTS.md "Operating method" (verified)
  - point 7 — trellis alone leaves the family enum `kodhama-0004-uniform-lifecycle`; the family question is filed — stated only in the record
  - consequence (load-bearing assumption, carried from decision-0080) — "the merge is the human act" holds only while the maintainer's account is his alone — stated only in the record
  - consequence — the typed-artifact schema's `draft → ratified` catalog and profile lifecycle is product-layer and stands — home: `core/schemas/typed-artifacts.md:100-104` (verified)
  - consequence — shipped `signature:` lines that mention `status` describe a consumer's own tell and stay unedited; the eval fixtures carrying `status: draft` stay — stated only in the record
- retired:
  - point 3 "partial supersession is marked by `superseded_in_part_by`" — by dated note PR #313 (TRL-98)
  - point 3 "supersession is marked by the forward pointer", for a full retirement with no successor — by dated note PR #315 (TRL-99)
- spent: point 4 (`ratify-guard.yml` and `ratify-flip.yml` deleted; both absent); the contract surface edits (rubric, known-bad fixture, `corpus-reviewer.md`, since deleted); the README correction; `decision-0080`'s `superseded_by`
- reason: No-status and merge-as-acceptance are the working rules in AGENTS.md and the rubric. Only its forward-pointer mechanics were narrowed by the dated-note method.
- confidence: high
- flags:
  - Point 7 and the account-is-his-alone assumption have no home in the tree.
  - The partial supersessions this record applied (decision-0022 D2/D3, decision-0037 D2, decision-0040 D5, decision-0042 D1/D2/D4, decision-0046 D1–D4) have surviving halves the spec must take from those records: decision-0046 D5, decision-0037's D1 principle, decision-0022 D1.

### decision-0083 — `.trellis/rules.toml` reconciles itself
- cluster: delivery (secondary: rule-model)
- disposition: mixed
- in force:
  - §6 — both hooks derive the payload's slug set from `reference/rules.md` (a trailing backticked slug), de-duplicated at source. An empty set is refused loudly as a plugin fault, never blamed on the project — home: `codex-context.mjs:684-687`, `:1190-1205`; `staleness.sh:861-883` (verified)
  - §1 (the surviving class rule, see flags) — a broken plugin payload is refused loudly and never delivered at exit 0 as a governed session. That covers an unreadable or empty rules file or header, a missing terminator, no slugs, and a failed `@rules.md` import — home: `staleness.sh:804-820`, `:846-856`, `:876-883`, `:901-907`, `:1012` (verified)
  - §1 eighth instance — paths reach awk through `ENVIRON`, never `-v`, so a plugin root containing a backslash works — home: `staleness.sh:922-990` (`ENVIRON` at `:988-989`) (verified)
  - the byte-cap finding:
    - Codex's own hook-output limit is ~2,500 tokens, and Codex spills rather than rejects
    - Trellis's cap is 9500 B, under that limit, with the provenance documented on the constant
    - the caps stay as runaway guards
    - home: `codex-context.mjs:17-30`; decision-2026-09-15 point 5 (verified)
  - §3 (surviving clause) — the Claude hook never writes `.trellis/rules.toml` — home: `cli/plugin_hook_test.go:1812`, `:3090` (verified)
  - §5 (surviving) — an empty `.trellis/rules.toml` governs every rule — home: `codex-context.mjs:735-744`; `plugins/trellis/reference/rules.md:1` (verified)
- retired:
  - §1 (the resolution table and reconciliation; the entry-point table *as reconciler guards*), §2 (quarantine), §3 ("announced, not asked", and the re-reading of decision-0070 D4 consent), §5 ("added 16 row(s)"; the adaptive-posture header) — by dated note PR #316. The one header is decision-2026-09-15 point 6.
  - §1's Claude-only scoping and the quarantine provenance wording — by decision-0084 (frontmatter), then PR #316
  - §1's seventh guard (`rules.md` vs `rules-b.toml` self-consistency) — tree: no `rules-b` reference left in `staleness.sh` (verified)
  - §4 (supersession of decision-0072 point 2) — moot under decision-0072's own dated note
  - Consequences: the commented-row state in consumer files, `TestReconciledRowsParseForCodexToo`, the repair-summary trailer — by PR #316
  - open question "should Claude warn on a false floor row?" — resolved: `staleness.sh:1296-1297` warns; decision-2026-09-15 point 3
  - open question "host parity owed (TRL-30)" — resolved by decision-0084, then retired by PR #316
  - open question "`block-codex.md` still teaches all-or-nothing" — resolved in the tree: `plugins/trellis/reference/block-codex.md:9` ("no entry in it makes the file invalid") (verified)
- spent: §7 (the three retired test pins; the emit count); VERSION 0.6.0 → 0.7.0; the catalog row-set list entry for the hardcoded `SLUGS` (`signature-catalog-v1.md:79-80`); the README corrections
- reason: Reconciliation and quarantine are gone. The payload-derived slug set, the loud refusal of a broken payload, the byte-cap finding and "the hook never writes" still govern.
- confidence: medium — the dated note retires §1 wholesale while the tree keeps §1's broken-payload refusals, so how much of §1 survives is a judgment call.
- flags:
  - **Drift between note and tree:** the PR #316 note says §1 "no longer holds", but the loud broken-payload refusals from §1's entry-point table survive in `staleness.sh:804-907`, now as delivery guards. A spec author reading only the note would drop a live rule.
  - **Drift:** the Consequences claim the payload→VERSION pair is unguarded (trellis#245), and the live catalog note repeats it at `core/catalog/signature-catalog-v1.md:77-78`. `.github/workflows/release-guard.yml:63` now guards the pair.
  - The method findings live only in records: a guard is proven only by breaking the code under it, and a count in a record is a measurement with a date. `docs/solutions/`, which AGENTS.md names as the home for past-problem learnings, does not exist in the tree.
  - The open question "should the byte proxy become a token estimate?" is still open.

### decision-0084 — Codex reaches reconciliation parity
- cluster: delivery (secondary: checks)
- disposition: mixed
- in force:
  - §3 — Codex refuses a payload whose derived slug set is empty (`no-slugs-in-payload`), naming the plugin. Stands per the PR #316 note — home: `codex-context.mjs:1190-1205` (verified). The lesson "a semantic and its guards are one unit when porting" is stated only in the record.
  - Go test-cache hazard — the hook tests exec external files, so `go test` must run with `-count=1`. CI and AGENTS.md carry the flag, and a test pins the workflow's flag — home: AGENTS.md "Checks and review"; `.github/workflows/cli-ci.yml:67-78`; `cli/codex_hook_test.go:1012-1032` `TestCliCIProvidesNode20BeforeGoTests` (verified)
  - §5 expiry principle — two host implementations held in step by a parity test is right at two hosts. At a third host, extract one shared implementation with thin adapters. Divergences are decided on the merits, not by which implementation is the reference. Stated only in the record; the current pair is `cli/rules_rows_parity_test.go:587` `TestBothHostsClassifyRulesRowsIdentically`.
  - §7 (surviving) — "the hook never writes" holds on Codex. Holds by construction only: `codex-context.mjs` contains no write call (verified). See flags.
- retired:
  - §1 (the classifier/reconcile split and the four fatal conditions), §2 (`!rulesSectionSeen` reconcilable), §4 (the host-neutral quarantine comment), §6 (over-budget degradation) — by dated note PR #316. decision-2026-09-15 point 3 now ignores those entries with warnings.
  - §5's byte-identity guard `TestBothHostsReconcileIdentically` and the CR-only divergence pin — by dated note PR #316. The CR-only case is now filed as TRL-100 per decision-2026-09-15's Consequences.
  - §6's one-shot gating — earlier by decision-0086 (itself retired in full by decision-2026-09-15 point 5)
  - §7's behavioural pin "`codexReconciledRows` re-reads `.trellis/rules.toml` after every run" — tree: the helper is gone (no match in `cli/`) (verified)
  - Consequences: the `parseRulesToml` contract; Codex commented rows; TRL-29 (retired with degradation); `block-codex.md` TRL-31 (resolved: `block-codex.md:9`)
- spent: VERSION 0.7.0 → 0.8.0; the plan/design annotation close-outs; the README corrections; §7's pointer extensions on decision-0083 and decision-0072
- reason: Codex reconciliation is gone with the opt-out model. The empty-slug guard, the `-count=1` discipline and the third-host rule are what remain.
- confidence: medium — whether §5's third-host expiry survives the retirement of the guard it described is a judgment call.
- flags:
  - **Drift (minor):** §7 says "the hook never writes" is enforced on Codex behaviourally. It no longer is: no Codex test re-reads the file, and the rule holds only by construction. The PR #316 note does not name §7.
  - §4's general rule ("a comment written into the shared file may not name a host-only command") lost its subject, but today's seeded comment is host-neutral (`install.sh:470`, `staleness.sh:794`). Worth carrying as a rule.
  - The open question "make the hook files `go test` cache inputs" is still open, with no consumer.
  - The "fixtures that cannot produce the condition they name" method finding is stated only in the record.

---

## Tally

**By disposition (16 records):**

| disposition | count | records |
|---|---|---|
| in-force | 3 | 0069, 0071, 0077 |
| mixed | 12 | 0070, 0072, 0073, 0074, 0075, 0076, 0078, 0079, 0081, 0082, 0083, 0084 |
| spent | 0 | — |
| repealed | 1 | 0080 |

**By primary cluster:**

| cluster | count | records |
|---|---|---|
| delivery | 7 | 0069, 0070, 0072, 0073, 0077, 0083, 0084 |
| method | 7 | 0071, 0075, 0076, 0079, 0080, 0081, 0082 |
| rule-model | 2 | 0074, 0078 |
| checks | 0 | (secondary on 0072, 0073, 0084) |
| positioning | 0 | — |

## Medium- and low-confidence calls

- **decision-0075 (medium)** — the migration banners and the Linear taxonomy live outside the repo and could not be checked.
- **decision-0081 (medium)** — a proposal whose accepted scope is the argument. Later records (0083, 0092, 0098 vs 0084) read its grant in opposite directions, and its catalog wording was never applied.
- **decision-0083 (medium)** — the PR #316 note retires §1 wholesale, but §1's broken-payload refusals survive in `staleness.sh`.
- **decision-0084 (medium)** — whether §5's third-host expiry principle outlives the byte-identity guard it described.
- Low: none.

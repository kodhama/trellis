# Batch 2 — decision-0026 … decision-0050

Sorted against the worktree at `main` at `6f292aa`
(branch `feature/trl-103-decisions-become-history`, head `6f292aa`). "Verified" means I opened the
home file and it states or implements the point. Paths are repo-relative.

### decision-0026 — The overlay always-loads the active rules; examples stay on-demand
- cluster: delivery
- disposition: mixed
- in force:
  - Decision, bullet 1 — every rule is in context each session as one concise line rendered from the single catalog source, never restated by hand — home: `plugins/trellis/reference/rules.md:3-38` (rendered by `cli/apply.go` `renderRulesReadout`/`ruleFragment`, pinned by `TestVendoredPayloadIsCurrent` at `cli/payload_test.go:520`; injected by `plugins/trellis/hooks/staleness.sh` path B) (verified)
  - Decision, bullet 2 — the depth (why, with/without pairs, inactive rules) stays on demand behind a pointer — home: `plugins/trellis/reference/trellis.md` last paragraph ("read its entry in `.trellis/internal/invariants.md` … before deviating") (verified)
- retired:
  - bullet 1's "its one-line `what`" as the rendered text — by decision-0034 point 1 (tree: `cli/apply.go:418-445` renders `directive`)
  - `profile.md` / `renderProfile` as carrier, and "the plugin's `/trellis:setup` skill does the same" — by decision-0053 (readout), decision-0072 (skill retired); tree: no `renderProfile` in `cli/`
  - Open question (per-invariant C1 visible in the block) — moot: `strictness` ignored, decision-2026-09-15-rule-rows-only-switch-rules-off point 1
- spent: the landing's "rules: always / reference: on demand" copy — not present in `site/index.html` or `site/lp-content.md` today
- reason: The split it made (rules always loaded, examples on demand, rendered from the catalog) is still how the payload works; only the carrier files and the rendered field changed.
- confidence: high
- flags: `cli/rules_test.go:8` `TestInvariantRulesCoverCatalog` says it "guards decision-0026", but it tests `invariantRules()` (`cli/apply.go:495-518`), which parses the `what` field and has no caller except that test. The live directive render is pinned by `TestInvariantDirectivesCoverCatalog` instead. The guard pins a retired render path. The "honest limit" (present ≠ enforced) is framing, not a rule.

### decision-0027 — Examples are matched without→with pairs, rendered as contrastive cards
- cluster: rule-model
- disposition: mixed
- in force:
  - point 1 — each entry's examples are matched pairs: `violated[i]` and `honored[i]` share a use case, layer tag and order — home: `core/catalog/signature-catalog-v1.md:27-31` and AC at `:455-460`; `docs/rubrics/artifact-contract.md:138-145` (check 8); `core/schemas/typed-artifacts.md:39`; `cli/corpus_conformance_typed_test.go:246` (verified)
  - point 2 — pairs render as contrastive cards (without over with) on the invariants page, and as with/without compare-pairs on the landing — home: `site/invariants.html` (16 `card` / `pairs` blocks, 39 `case` blocks), `site/lp-content.md:96-116`; content pinned by `cli/sync_test.go` `TestInvariantsPageMatchesCatalog` (verified; the card layout itself is not test-pinned)
  - Amendment 2026-07-19 — the first `violated` bullet doubles as the always-loaded ✗ line and may carry an appended clause from another pair's use case; the pair guarantee is bullet-level — home: the ✗ extraction is `cli/apply.go:453-458`; the composition permission is stated only in the record (decision-0074:139-143 confirms "the amendment keeps its rule and loses its example")
- retired:
  - point 3 "two pairs per invariant" — by decision-0074:144-149, decision-0078:173-175, decision-0081:388 (explicit count-change clauses); tree: rubric check 8 says `≥2`, and the catalog ships 3–5-pair entries
  - the amendment's worked instance (`inv-self-improvement` CI bullet) — by decision-0074 point 2 (:139-143)
- spent: catalog examples re-authored into aligned pairs; bundled copies regenerated (visible in catalog and `cli/assets/invariants.md`)
- reason: Matched pairs and card rendering are live and test-guarded; the fixed count of two was amended to "at least two" by later records.
- confidence: high
- flags: No dated note or pointer on 0027 marks point 3's retirement; 0074, 0078 and 0081 amended it only in their own bodies. A spec must not carry "two pairs". The amendment's permission has no home outside the record.

### decision-0028 — Derived resources declare their source, and every pair gets a sync guard
- cluster: checks
- disposition: mixed
- in force:
  - point 1 — a source artifact names its derived resources where it is edited (forward edges) — home: `core/catalog/signature-catalog-v1.md:49-53`, `docs/rubrics/artifact-contract.md:27-36`, catalog `inv-graph-maintenance` signature `:164-166` (verified)
  - point 2 — a deterministic sync guard per source↔derivative pair: byte-identical for copies, contains-source for renders — home: `cli/sync_test.go:10-70`, `cli/row_set_guard_test.go:8`, `cli/rules_rows_parity_test.go:8`, `cli/artifact_contract_guard_test.go:15`, `cli/docs_consistency_test.go:515,529`, `cli/invariants_pointer_test.go:562`, `.github/workflows/release-guard.yml:3`, `.github/workflows/eval-scorecard.yml:2` (verified)
  - point 3 — the rule sits where it fires, as a trigger line in the project instructions — home: `AGENTS.md:97` ("update derivatives in the same change; a guard per pair"), applied to the payload at `AGENTS.md:142` (verified)
- retired: point 3's home `CLAUDE.md` — by decision-0057 (pointer in the record); now `AGENTS.md`
- spent: none (the Open question's manifest was never built and has no consumer)
- reason: All three points are live and heavily cited by tests, workflows and AGENTS.md; only the instruction file that carries point 3 moved.
- confidence: high
- flags: DRIFT (minor) `core/catalog/signature-catalog-v1.md:50-51` says the catalog is "copied verbatim" to `plugins/trellis/reference/invariants.md`, but `cli/sync_test.go:15-20,38-45` checks an entries-section-only extract (decision-0055). The forward-edge note is itself stale. Also, idea 4 of the merged ideation doc (`docs/ideation/2026-09-14-repo-simplification-ideation.html:275`) proposes replacing "a guard per pair" with "guards follow their subject / no committed copy". That idea is not decided, but a spec author should expect this rule to be contested.

### decision-0029 — Setup asks the mode first; detection is per-mode
- cluster: delivery
- disposition: repealed
- in force: none
- retired: whole record — by `superseded_by: [decision-0043]` (rule 1); tree: `cli/main.go:52-58` has only `payload` (+ `version`/`help`); no `setup.go`/`harness.go`
- spent: none
- reason: The interactive CLI setup it ordered was deleted, along with its M1/M2 detection logic.
- confidence: high
- flags: none (`research-0010` keeps it as `informed_by`, which is provenance and is fine)

### decision-0030 — The interactive setup takes one dependency (golang.org/x/term)
- cluster: checks
- disposition: mixed
- in force:
  - the note's surviving principle — a dependency question is answered with the smallest dependency that does the job, or none — home: `inv-minimal-first` in `core/catalog/signature-catalog-v1.md` and `plugins/trellis/reference/rules.md:31` (verified in substance, not CLI-specific); `cli/go.mod` has no `require` (verified), but no check forbids adding one (`scripts/check-go-mod.sh` checks tidiness only), so "the CLI stays dependency-free" is stated only in the record and in decision-0043:39
- retired: `x/term` dependency, hand-rolled selector, accent colour, TTY-only flow, Go 1.22 pin rationale — by decision-0043 rule 1 (mooted note in the record); tree: `cli/go.mod` has no require, no `tui.go`
- spent: none
- reason: The TUI and its dependency are gone; only a restatement of minimal-first for future dependency choices survives.
- confidence: medium — the surviving point may be just `inv-minimal-first` restated, in which case the record is repealed
- flags: none

### decision-0031 — Always-load one primary failure example per active rule
- cluster: delivery
- disposition: in-force
- in force:
  - Decision — under each always-loaded rule, one ✗ line: that rule's `violated[0]`, pulled from the catalog — home: `plugins/trellis/reference/rules.md:5-38`, `cli/apply.go:377-395` (`ruleFragment`) and `:453-458` (`invariantPrimaryFailure`), guard `cli/rules_test.go:25` `TestInvariantPrimaryFailureCoverCatalog` (verified)
  - bullet 3 — curation is by ordering (the example to load goes first); no new catalog syntax — home: `cli/apply.go:455-458` comment (verified)
  - Consequences — full pairs stay on demand — home: `plugins/trellis/reference/trellis.md` pointer (verified)
- retired: none of substance (the function names `renderProfile`/`activeRuleLines` and the `.trellis/invariants.md` path changed with later delivery records; the behaviour did not)
- spent: none
- reason: Every always-loaded rule still carries exactly one ✗ failure line taken from the first violated example, and a test pins it.
- confidence: high
- flags: The "one, not a pair" rationale and the "second example is a later knob" note exist only in the record.

### decision-0032 — Homebrew as a second install channel
- cluster: positioning
- disposition: repealed
- in force: none
- retired: whole record — by `superseded_by: [decision-0041]`; the channel itself by decision-0043 rule 4; tree: no `scripts/gen-formula.sh`, no `.github/workflows/update-formula.yml`, `site/lp-content.md:33` ("the Homebrew tab retired")
- spent: tap repo created (external, not visible here)
- reason: Trellis no longer ships a binary, so there is no formula to keep in sync.
- confidence: high
- flags: 0032's own top pointer still says its formula-sync mechanics "still apply verbatim against the new tap". That has been false since decision-0043, and only 0041's note records it.

### decision-0033 — Offer two postures; park seed and custom
- cluster: rule-model
- disposition: repealed
- in force: none
- retired:
  - point 1 (two postures, default B; "A vs B is a stated stance") — by dated note TRL-97 (PR #316) → decision-2026-09-15-rule-rows-only-switch-rules-off point 6; tree: `plugins/trellis/reference/trellis.md` carries only "By default", `cli/apply.go:148-151`
  - point 2 (`seed`/`custom` parked as presets) — its trigger was met by decision-0051 (active-subset lever returned as rows); no preset ships at all after decision-2026-09-15 point 6, so nothing remains parked
  - point 3 (inert `Profile.Active` mechanism) — by decision-0051 rule 4 (pointer in record), then decision-0053; tree: no `Profile.Active` in `cli/`
  - Consequences' `--profile a|b` flag — by decision-0043 (generator-only CLI)
- spent: none
- reason: Postures, presets and the in-code subset mechanism are all gone; switching rules off is now done with `active = false` rows.
- confidence: medium — point 2's retirement is inferred from "no preset ships"; the dated note speaks only to point 1 and the pointer says point 2 "remain[s] parked"
- flags: The 0051 pointer on 0033 ("the rest of this record stands: seed and custom remain parked") contradicts the tree after TRL-97, and the TRL-97 note does not correct that sentence for point 2.

### decision-0034 — The always-loaded block speaks to the host agent
- cluster: delivery
- disposition: mixed
- in force:
  - point 1 — every invariant carries an imperative, self-contained, code-free `directive`, and the always-loaded text renders it; `what` stays for the reference — home: `core/schemas/typed-artifacts.md:36`, `docs/rubrics/artifact-contract.md:140` (check 8), `cli/apply.go:418-445`, `plugins/trellis/reference/rules.md` (verified)
  - point 2 — imperative header ("You are working in a project that follows Trellis … Follow the rules below as you work here") — home: `cli/apply.go:153-160` `governanceHeader`, `plugins/trellis/reference/trellis.md:1-3` (verified)
  - point 3 (narrowed) — enforcement strength is stated in plain language, never as the dial word — home: `cli/apply.go:148-151` `defaultPostureLine`, `reference/trellis.md:5` (verified)
  - point 4 (narrowed) — no Trellis-internal codes in directives, enforced by test — home: `cli/rules_test.go:137-156` `TestInvariantDirectivesCoverCatalog` (verified); a rule's own slug suffix is allowed (`cli/apply.go:381-387`, citing a maintainer addendum to decision-0051)
- retired:
  - point 3's per-profile variation (firmly / by default / as guidance) — by decision-2026-09-15-rule-rows-only-switch-rules-off point 6; tree: `cli/apply.go:150` ("The 'Firmly' and 'As guidance' lines retired")
  - point 4's absolute ban, narrowed to allow slugs — by decision-0051's slug addendum (per `cli/apply.go:381-387`; I did not open 0051 to confirm the wording)
- spent: the cold-read re-test; the catalog gaining the `directive` field (visible)
- reason: Directive field, imperative header, plain-language strength and the no-codes test are all live; only the three-way strength wording retired.
- confidence: high
- flags: MISSING DATED NOTE: decision-2026-09-15 retired point 3's three strength phrasings, but 0034 is not among the thirteen records that PR noted (`2026-09-15…md:159-160`).

### decision-0035 — Trellis self-applies through its own install boundary
- cluster: method
- disposition: mixed
- in force:
  - rule 1 — producing Trellis (`core/`, plugin, CLI) is separated from consuming it; Build/Govern/Method roles; the invariants reach this repo only through delivery and are never hand-restated, while the method is hand-authored — home: `AGENTS.md:1-15` (layer statement), `AGENTS.md:112-114` ("delivered live by the Trellis plugin at session start, not hand-written here … belongs in the catalog") (verified)
  - rule 2 (as re-homed) — the repo self-applies through the same delivery a consumer gets — home: `.claude/settings.json:3` (`trellis@kodhama`), `.trellis/rules.toml`, `cli/selfapply_test.go:20-61`, `README.md:308-324` (verified)
  - rule 3 floor — drift is made visible, not silent — home: `plugins/trellis/hooks/staleness.sh:8-20,526-544` (vendored overlays), `install.sh:912` (verified); on the plugin path there is no copy to drift (decision-0065)
  - rule 4 — the dual role is stated out loud — home: `README.md:308`, `AGENTS.md:1-15` (verified)
- retired:
  - rule 2's mechanics (run setup, commit `.trellis/` plus block, CI sync-guard `TestRepoOverlayIsCurrent`) — by decision-0065 (forward pointer) and decision-0071 (frontmatter comment); tree: no `.trellis/internal/`, `CLAUDE.md` is `@AGENTS.md` only
  - rule 3's surface (`trellis status`, "update the tool, re-run setup") and the user-side version stamp as universal — by decision-0043 (note), narrowed to vendored projects by decision-0065
  - "Method hand-authored in `CLAUDE.md`" — by decision-0057 (note) → `AGENTS.md`
- spent: `CLAUDE.md` stripped of invariant echoes (visible: `AGENTS.md:112-114`); the plugin-version open question answered by decision-0036 and then decision-0061
- reason: The layer split, self-application and the visible-drift floor still hold; the overlay, setup and sync-guard it used to achieve them are gone.
- confidence: high
- flags: DRIFT (comment) `cli/main.go:6-8` still says the package tests keep "the repo's own overlay" in sync, which has been false since decision-0071. The spec must carry 0071's stated reductions: this repo runs the last released plugin, not HEAD, and has no Codex governance until #220. It must not carry 0035's "drift is impossible" claim.

### decision-0036 — The plugin versions by commit, not a frozen number
- cluster: delivery
- disposition: repealed
- in force: none
- retired: whole record — by `superseded_by: [decision-0059, decision-0061]` (0060 retired 0059; 0061 is the successor); tree: `plugins/trellis/.claude-plugin/plugin.json:3` `"version": "0.24.0"`, `plugins/trellis/VERSION` = `0.24.0`, `AGENTS.md:142-148`, `.github/workflows/release-guard.yml`
- spent: none
- reason: The plugin now carries an explicit SemVer version that every payload change bumps, the opposite of "omit the version".
- confidence: high
- flags: none (the "auto-update is off by default for third-party marketplaces" observation is historical host behaviour, not a rule)

### decision-0037 — Statuses are methodology-defined; the contract requires a ratifiable shape
- cluster: method
- disposition: mixed
- in force:
  - D3 — each artifact has an accountable-human role; `owner:` carries it by default, a methodology may map fields when the mapping is declared, and this repo's mapping is "`owner: agent` carries authorship; accountability stays with the maintainer" — home: `AGENTS.md:94` (verified); `owner` required by rubric check 1 (`docs/rubrics/artifact-contract.md:44-47`) (verified)
  - D1 principle — the lifecycle vocabulary is the methodology's own; the invariants need a ratifiable shape, not names — home: `README.md:326-327` ("its lifecycle is the repository's own: merging to `main` is the acceptance"), `docs/rubrics/artifact-contract.md:82-86` (check 5's "where a methodology declares a status lifecycle" clause), `core/invariants/trellis-invariants-v1.md:344-353` (verified). The shape bullets as a contract requirement on any methodology (working state not consumable, defined promotions, intent gate holds, supersession expressible) are stated only in the record: the contract is repository-internal (decision-0100 point 1), and nothing shipped checks a consumer's lifecycle.
- retired:
  - D2 (trellis's `draft → ratified` default) — by decision-0082 (frontmatter comment)
  - D1's fifth bullet (declared enum checked; an undeclared status fails conformance; a shapeless lifecycle fails loudly) — by decision-0082 (frontmatter comment); tree: rubric check 2 "No status check"
  - the earlier partial supersession by decision-0042 (trellis adopts the family enum) — itself retired by decision-0082 point 7
- spent: the `spec-0001` amendments (spec retired by decision-0079); `conformance-reviewer` check 2 (agent deleted by decision-0098)
- reason: The owner/authorship mapping and "the methodology owns its lifecycle" are live in AGENTS.md and README; the enum and status checks retired with the status field.
- confidence: medium — whether D1's shape requirement is still a product rule (for consumer methodologies via Assess) or died with the repository-internal contract is a judgment call
- flags: DRIFT (stale pointer) `AGENTS.md:94` sends artifact writers to "`decision-0042` (family lifecycle)", which decision-0082 exited for this repo. A consequence was never built: capturing a host's lifecycle declaration in the expression profile (a "candidate spec-0002 field"). `core/schemas/typed-artifacts.md` has no such field, and no consumer is named (the decision-0078 no-orphan rule).

### decision-0038 — Retire the display codes; slugs are the only name
- cluster: rule-model
- disposition: mixed
- in force:
  - point 1 — a slug is an invariant's only name, for reference and display; letter-number codes and group letters are not used — home: `core/invariants/trellis-invariants-v1.md:69-78`, `plugins/trellis/reference/rules.md` (slug-suffixed, no codes), `site/invariants.html` (no `[ABCD][1-9]` tokens), `eval/experiments/does-trellis-help/gen-invariant-scorecard.py:14` (verified by grep)
  - point 2 — the legacy code map lives only in the `invariants-v1` Identifiers registry, titled "Legacy code (retired)" — home: `core/invariants/trellis-invariants-v1.md:80-100` (verified)
  - point 3 — append-only artifacts are not rewritten to chase the retirement — home: `core/invariants/trellis-invariants-v1.md:75-76`, `AGENTS.md` append-only rule (verified)
  - Consequences — retired codes stay frozen and are never reassigned — stated only in the record (no match for reassign/frozen in `core/`, `AGENTS.md` or the rubric; decision-2026-09-15 point 11 covers retired slugs, not codes)
  - residual exemptions — schema field names `C1`/`C2`/`default_C1`/`default_C2` are identifiers, not display; Go-internal names and past eval runs are exempt — home: field names still in rubric check 8 (`docs/rubrics/artifact-contract.md:140-141`) and catalog entries (verified); the exemption rationale is stated only in the record
- retired: "migrate `spec-0001`–`spec-0004` opportunistically" — moot (specs deleted by decision-0079); the swept derivatives `.trellis/invariants.md` and `conformance-reviewer` — gone (decision-0071, decision-0098)
- spent: the slug-first sweep of catalog, copies, site, README, profile, scorecard generator (visible)
- reason: Slug-only naming and the legacy map are live; the "never reassign a retired code" rule has no home outside the record.
- confidence: high
- flags: possible DRIFT: `core/catalog/signature-catalog-v1.md:434` (`floor-intent-gate` `what`, shipped in the on-demand reference) says "`C2` can never be `none`". That is prose using a retired code, arguably covered by the field-name exemption. The Open question on renaming the schema fields has no named consumer.

### decision-0039 — The staleness surface is a SessionStart hook; the plugin stamps `.trellis/version` too
- cluster: delivery
- disposition: mixed
- in force:
  - rule 1 — any agent-facing staleness surface is a `SessionStart` hook emitting `additionalContext`, never `InstructionsLoaded` — home: `plugins/trellis/hooks/hooks.json:3-8` (`SessionStart`, `startup|resume`), `plugins/trellis/hooks/staleness.sh:2,73,119` (nested `hookSpecificOutput` envelope), `plugins/trellis/hooks/codex-hooks.json` (`SessionStart`, `startup`) (verified)
- retired:
  - rule 2 (setup skill writes `plugin@<short-sha>` to `.trellis/version`) — by decision-0043 (note in record); tree: `plugins/trellis/reference/version` = `payload@63217cf5580b`; setup skill retired by decision-0072
  - Consequences/Open questions about `trellis status` — by decision-0043
- spent: none
- reason: Staleness and rule delivery still run through SessionStart hooks; the commit-SHA stamp and the status command are gone.
- confidence: high
- flags: Rule 1's "(or plain stdout)" and bare `additionalContext` wording is incomplete: decision-0065 found the flat JSON form is silently discarded and requires the nested envelope (`staleness.sh:73`). The spec should state the envelope. The marketplace "how far behind" question is still open, with no consumer.

### decision-0040 — Five reverse-ports from instance #1
- cluster: rule-model
- disposition: mixed
- in force:
  - point 1 — one home per kind of information, placed by which consumer must trip over it; copies point, never carry — home: `core/catalog/signature-catalog-v1.md:167-169,175-177,185-187` (verified)
  - point 2 — tests name their upstream; a test↔spec conflict is resolved deliberately; a regression test is never weakened to fit a reading — home: catalog `:169-170,178-181,189-190` (verified)
  - point 3 — bounded runs checkpoint and resume; hitting the bound is loud — home: catalog `:413-415,420-422` (`floor-transparency`) (verified)
  - point 4 — verify against the source before calling a thing right or wrong — home: catalog `:308,314` and the collab pair; `plugins/trellis/reference/rules.md:25` (verified)
  - point 5 remainder — a record can be retired in part while its remainder stays current, and existing `superseded_in_part_by` entries still resolve — home: `docs/rubrics/artifact-contract.md:121-131` (check 7), `AGENTS.md` operating method (dated notes), `cli/corpus_conformance_test.go:857`; catalog `inv-auditable-archive` signature `:337-340` (verified)
  - Context — admission test for a reverse-port: its catalog text carries no source-project nouns, and it enters through an existing entry (signature clause + matched pair) unless it adds a mechanism — stated only in the record (decision-0081:195,320 cites it as live)
- retired:
  - point 5 "the marked record keeps `status: ratified`" — by decision-0082 (frontmatter comment)
  - point 5's new `superseded_in_part_by` as the marking for a partial retirement — by dated note TRL-98 (PR #313); tree: `AGENTS.md` operating method, rubric check 7 `:124-126`
- spent: worked instance on decision-0013; catalog pairs added and copies regenerated; the `spec-0001` amendment (spec retired)
- reason: Four catalog sharpenings are live in the shipped catalog; the partial-supersession field was replaced by dated notes, though existing entries still resolve.
- confidence: high
- flags: The shipped catalog clause (`signature-catalog-v1.md:337-340`, copied to `plugins/trellis/reference/invariants.md:256-259`) still says "a superseded-in-part pointer marks the outgrown half". It describes a consumer's tell, and a dated note is also a forward link, so this is probably not drift, but rewording it is a payload release. The instance-#2 falsification question has no consumer.

### decision-0041 — The tap is the family's, not trellis's own
- cluster: positioning
- disposition: repealed
- in force: none
- retired:
  - tap moves to `kodhama/homebrew-tap`; `brew install kodhama/tap/trellis` — by decision-0043 rule 4 / kodhama-0007 rule 5 (note in record); tree: no Trellis `brew install` in `README.md` (the only brew line is ShellCheck, `:338`), `site/lp-content.md:33`
  - formula-sync mechanics and `TAP_DISPATCH_TOKEN` — by decision-0043 (note)
- spent: 0032 marked superseded (visible in 0032's frontmatter); release workflow and docs repointed in PR #103 (since removed)
- reason: Trellis ships no binary and no formula, so where a tap lives is no longer a Trellis question.
- confidence: medium — the note says "the family-tap shape … stands for any future kodhama product that ships a real binary"
- flags: That surviving "one org tap, each product pushes its own formula" rule belongs to `kodhama-0001-family-delivery`, not to Trellis. It should not enter a Trellis spec. `homebrew-tap` stays a registry member in rubric check 4, from decision-0044 and unaffected.

### decision-0042 — Adopt the family lifecycle for trellis-self
- cluster: method
- disposition: mixed
- in force:
  - core — merging is the ratification act (now the only acceptance) — home: `AGENTS.md:78-80` ("Merging to `main` is the acceptance"), `README.md:326-327` (verified)
  - D1 remainder — an artifact not yet accepted is not consumed — home: `AGENTS.md:79-80` ("one not yet merged may not [be consumed]") (verified)
  - D3 — history stands: older artifacts keep their `status:` lines and are never relabelled — home: `AGENTS.md:84-87`, `docs/rubrics/artifact-contract.md:50-53` (verified)
- retired:
  - D1 family enum `draft → gated → approved` for this repo — by decision-0082 point 7 (frontmatter comment)
  - D2 PR carries `draft → gated`, a post-merge bump records `approved` — by decision-0046 (post-merge bump / no in-diff approved) and decision-0082 (remaining flip)
  - D4 bootstrap — by decision-0082 (comment); the act itself is spent (its `status: approved` line)
  - Consequences' `ratify-guard` — deleted by decision-0082 point 4; tree: no `ratify-guard.yml` in `.github/workflows/`
- spent: the `.trellis/profile.md` lifecycle-mapping section (file gone); CLAUDE.md statuses line rewritten
- reason: Merge-as-acceptance and "keep old status lines" are live in AGENTS.md; the family enum and every flip mechanic retired with the status field.
- confidence: high
- flags: DRIFT (stale pointers): `AGENTS.md:94` lists "`decision-0042` (family lifecycle)" as required reading before writing an artifact, and `docs/rubrics/artifact-contract.md:6,93` cite decision-0042 for per-type body sections. 0042 defines neither body sections nor, after decision-0082, this repo's lifecycle. The live content is 0022's and 0082's merge-is-acceptance, so the spec should cite that instead.

### decision-0043 — Generator-only CLI; the overlay stamp is the payload stamp, compared file-to-file
- cluster: delivery
- disposition: mixed
- in force:
  - rule 1 — the Go CLI is generator-only: `payload` (+ `version`/`help`) renders the bundle and checksum manifest into `plugins/trellis/reference/`; no end-user commands; its tests are the CI guards (regenerate-and-diff, docs-consistency, hook contract) — home: `cli/main.go:1-12,52-58`, `cli/main_test.go:44-50` `TestGeneratorOnlyCommandSurface`, `cli/payload_test.go:510-520`, `cli/docs_consistency_test.go:159`, `cli/plugin_hook_test.go:17` (verified)
  - rule 2 — the overlay stamp is the payload's content-derived `version` file (`payload@<content-hash>`), copied verbatim by an installing writer — home: `plugins/trellis/reference/version`, `README.md:194` (manual copy path writes `.trellis/internal/version`) (verified)
  - rule 3 — staleness is a file-to-file compare of the project stamp against the installed plugin's `reference/version`: warn on mismatch, no git, binary or network; legacy stamps (flat path, `plugin@sha`, semver) draw a migration nudge — home: `plugins/trellis/hooks/staleness.sh:8-20,526,541-544`, `plugins/trellis/hooks/codex-context.mjs:48` (verified); scope is vendored projects only (decision-0065)
  - rule 4 — no end-user binary release channel (no auto-release or release workflow, no tap formula); the root `install.sh` is a plugin vendor script, a different artifact class — home: `.github/workflows/` (no release workflows), `install.sh:8`, `README.md:286,298-300`, `cli/install_script_test.go:21` (verified)
- retired:
  - rule 1's live homes `/trellis:setup` (incl. M2) and the repo-overlay sync-guard — by decision-0065 and decision-0072 (setup and morph), decision-0071 (overlay guard)
  - rule 2's "setup skill" writer and the repo-overlay no-stamp exception — by decision-0072 and decision-0071; universality narrowed to vendored projects by decision-0065; stamp path moved to `.trellis/internal/version` by decision-0051 (unmarked on 0043, disclosed in 0087's comment)
  - rule 3's "no-op when either side is missing or empty" — by decision-0087 (frontmatter comment); tree: `staleness.sh:506-513` `TRELLIS_RULES_NOT_LOADED`
  - rule 4's "shipped artifact is the payload at HEAD (plugin versions are commits)" — by decision-0061 (Approval record section); tree: `plugins/trellis/VERSION`, `release-guard.yml`
  - rule 4 note's "every product decision stays in `setup/SKILL.md`; `install.sh` vends the plugin, never the overlay" — contradicted by decision-0072 (skill gone) and by decision-0068/0070 (`install.sh` renders `.claude/rules/trellis.md` and seeds `.trellis/rules.toml`); neither marks 0043
- spent: deletion of the TUI, `status`/`remove`/`uninstall`, harness detection, release workflows; forward annotations; eval harness moved to mechanical copy
- reason: The generator-only CLI, content-hash stamp, file compare for vendored overlays and the no-binary channel are live; setup, the repo overlay and commit-versioning pieces retired.
- confidence: medium — several partial scopes stack up here, and one of them (the install.sh note) is retired in the tree without a pointer
- flags: DRIFT (comment) `cli/main.go:5-8` still says "pre-built M1 payload" and that the tests keep "the repo's own overlay" in sync. Unmarked retirement: the rule 4 note's claim about `install.sh`'s scope, contradicted by decision-0068/0070. decision-0072's open question 1 still stands: staleness path A and the legacy-stamp branches serve an unknown outside population and may be retired.

### decision-0044 — Cross-repo `depends_on`: a qualified `repo/id` form
- cluster: method
- disposition: mixed
- in force:
  - Decision — a cross-repo reference is `<repo>/<id>`, where `<id>` is spelled exactly as in its home corpus and the delimiter is `/` — home: `docs/rubrics/artifact-contract.md:61-64` (check 4), read by `cli/corpus_conformance_test.go:109,199,251-270` and pinned by `cli/artifact_contract_guard_test.go:338` (verified)
  - registry — kodhama, trellis, grove, wisp, design-system, homebrew-tap, math-quest — home: `docs/rubrics/artifact-contract.md:62`; the Go test reads the list from that line (`corpus_conformance_test.go:251`) (verified)
  - depth — shape and registry membership only; the referent is not fetched — home: `docs/rubrics/artifact-contract.md:63-64` (verified)
  - retrofit, not grandfather — home: decision-0044:5 now qualified, pinned by `cli/corpus_conformance_test.go:1352-1353` (verified)
- retired:
  - "amending `spec-0001` alone suffices family-wide; composes onto sibling repos" — by decision-0100 point 1 (the contract is repository-internal and nothing shipped reads it); `spec-0001` retired by decision-0079
  - `corpus-reviewer` check 4 as the executor — deleted by decision-0098; now the Go test
- spent: the follow-on `spec-0001` amendment (PR #137); retrofit of `specs/0005` moot (specs deleted); kodhama/grove retrofits (outside this repo, not verifiable here); wisp's code-comment channel left to wisp#13
- reason: The qualified form, registry and shape-only check are live in the rubric and Go test; the family-wide reach ended when the contract became repository-internal.
- confidence: high
- flags: decision-0100 ended the family-wide claim without a note on 0044.

### decision-0045 — Artifact-versioning kinds; pin-vs-current conformance
- cluster: method
- disposition: mixed
- in force:
  - point 7 — `changes:` is a forward pointer of the `superseded_by` class, never a `depends_on`-class edge, and is not walked as flow — home: `docs/rubrics/artifact-contract.md:86-89` (check 5), `cli/corpus_conformance_test.go:95,857`; used by ten records, e.g. decision-0100:5 (verified)
  - point 8 — a pin has shape `id@version` / `repo/id@version`; shape-only acceptance — home: `docs/rubrics/artifact-contract.md:72-75`, `cli/corpus_conformance_test.go:112` (4f), `:1062` `bareRef` (verified)
  - point 2 — two kinds: append-only (decisions; supersession is the history) and revise-in-place (invariants, research, rubrics, schemas; re-point to the successor) — home: `AGENTS.md` ("`docs/decisions/` stays append-only"), rubric check 7 `:123-124` (verified). The kind's explicit version stamp is not in force: no corpus artifact carries `version:` (grep).
  - point 5 (byte-identity case) — a vendored bundle versions by content hash — home: `plugins/trellis/reference/version` (verified; originally decision-0043)
- retired:
  - point 3 (derived pin-vs-current conformance), point 6 (agent-generated significance counter; decision/version cross-check) and Consequences 1–3 — re-homed by `grove/adr-0010` (frontmatter comment); grove retired from this repo by decision-0076; rubric check 12 "has no owner here and is not applied" (`docs/rubrics/artifact-contract.md:155-161`); no spec exists (decision-0079) and no artifact carries a counter
  - point 1 (code, pages and mockups are contract artifacts with `depends_on`) — never executed in this repo, since the rubric covers non-code `.md` only (check 1); the operational application was grove#34, gone with grove
- spent: point 4 (type→kind mapping is family-level, descriptive); the adversary rounds
- reason: The `changes:` relation, `@version` pin shape and append-only/revise-in-place split are live in the rubric; the versioning semantics went to grove and left with it.
- confidence: medium — points 1, 3 and 6 are "no subject here" rather than explicitly repealed
- flags: `AGENTS.md:169-171`: the `superseded_in_part_by: grove/adr-0010` edge on 0045 is "live and load-bearing — do not clean up". No file in this repo states spec-versioning semantics any more. A spec that wants a revise-in-place version marker must state it anew.

### decision-0046 — The approval act is a human intent act, not the merge
- cluster: method
- disposition: mixed
- in force:
  - D1 principle — approval is a human intent act; the mechanism that records or performs it is secondary — home: `core/catalog/signature-catalog-v1.md:443` ("the approval recorded as a human intent act, whatever mechanism performs it") (verified)
  - D2 substance — an agent may not manufacture acceptance, or merge on the human's behalf, without his act — home: `AGENTS.md:89-91` (verified), catalog `floor-intent-gate` directive `:437` (verified), kept by decision-0082 point 5
  - D5 — when it is unclear whether a human instruction is the approval act, clarify before treating it as one, and do not stall a real approval by failing to recognise it — stated only in the record (decision-0082's comment says it survives "as general agent behavior"; `floor-intent-gate`'s "Unsure whether a human must approve? Assume yes" and `inv-clarify-before-commit` answer different questions; `AGENTS.md:88` "If he has asked for a PR, open it" covers only part of the don't-stall half)
- retired:
  - D1's "the merge is one way, not the only way" as this repo's mechanic, and the `status: approved` flip that records the act — by decision-0082 points 2 and 6 (frontmatter comment)
  - D3 in-PR `gated → approved` flips — by decision-0082
  - D4 slimmed `ratify-guard` (draft-landing check kept) — by decision-0082 point 4; tree: no guard workflow
  - Consequence 2 family-wide guard (grove#38) — grove retired (decision-0076); decision-0080's `ratify-flip.yml` deleted by decision-0082
- spent: `ratify-guard` slimmed (later deleted); CLAUDE.md and `conformance-reviewer` charter edits
- reason: "Only a human approves, and agents never self-approve" is live; the flip-and-guard machinery retired; the clarify-when-ambiguous rule survives with no home file.
- confidence: high
- flags: D5 must be newly stated in a spec: decision-0082 keeps it, but no current file does.

### decision-0047 — Provenance is not a dependency
- cluster: method
- disposition: mixed
- in force:
  - Decision — `depends_on` means genuine coupling (the artifact's correctness is or was contingent on the source); provenance is a distinct relation; carrying provenance in `depends_on`, or relabelling a coupling as `informed_by`, is non-conformant for new and live artifacts — home: `AGENTS.md:153-158`, `docs/rubrics/artifact-contract.md:76-78,90-92`, `cli/corpus_conformance_test.go:121` (5d, judgment) (verified)
  - coupling is live (drift-bearing) or frozen (an append-only record pinned at ratification), and both are dependencies — home: rubric check 7 exemption `:129-131` (verified in substance); the live/frozen terms are stated only in the record
  - already-frozen append-only artifacts keep their existing edges and are not migrated — home: `AGENTS.md` append-only rule (verified in substance); stated explicitly only in the record
  - the choice is on cost grounds, not a deduction — framing, stated only in the record
- retired:
  - "mechanism-free: names no relation grammar; the field and grammar are grove's" — grove retired (decision-0076); this repo now defines and checks `informed_by` itself (`docs/rubrics/artifact-contract.md:76-81`, `cli/corpus_conformance_test.go:114-115,1011-1020`); no note on 0047
  - Consequence 1 (marking on `spec-0001` §1) — spec retired by decision-0079
- spent: Consequence 4 consumer audit (visible: "Amended in place 2026-07-13 (`decision-0047` …)" in `docs/research/0002,0003,0005–0009`; `docs/research/0010` on 2026-09-06, TRL-57)
- reason: Coupling-only dependencies and a separate provenance edge are enforced by the rubric, the Go test and AGENTS.md; only the "grove owns the grammar" layering fell away.
- confidence: high
- flags: DRIFT (stale attribution) `docs/rubrics/artifact-contract.md:77-78` says "the field name and the fuller taxonomy are grove's relations charter", though grove is retired and this repo's test owns the field. Consequence 2 (a catalog line saying the dependency edge is coupling) was never done; the catalog has no such line, and the item has no consumer.

### decision-0048 — Setup performs no git and injects no landing workflow
- cluster: delivery
- disposition: repealed
- in force: none
- retired: every point constrains `/trellis:setup` — by decision-0072 (frontmatter comment: "VACUOUS rather than false"); tree: `plugins/trellis/skills/` holds only `remove/`
- spent: the `SKILL.md` step 8 rewrite and manifest advance (gone with the skill); the M2 open question closed by decision-0050, then the morph removed by decision-0065
- reason: The skill it constrained no longer exists.
- confidence: medium — the subject is gone, but the principle is reusable
- flags: The underlying principle is stated only here. A shipped skill's prose runs inside the consumer's session, so it must carry no git-workflow opinion (no "open a PR", no committing onto the current branch): surface the uncommitted state and defer landing to the project's own conventions. It is not restated for `plugins/trellis/skills/remove/SKILL.md` (which today leaves destructive git steps to the user, `:235-240`), for hook messages, or for `install.sh`. decision-0072's table notes the installer "suggests the `git add`". A spec author should decide whether to carry it.

### decision-0049 — Setup offers to hide `.trellis/` from consumer tooling
- cluster: delivery
- disposition: repealed
- in force: none
- retired: detect linters/formatters and offer to add `.trellis/` to their ignores — by decision-0065 ("its subject evaporates when only `rules.toml` remains"; forward pointer) and decision-0072 (skill retired; frontmatter comment)
- spent: the "symmetric removal" follow-up was built: `plugins/trellis/skills/remove/SKILL.md:64-65,108-114` removes consent-gated `.trellis/` ignore entries a past setup added
- reason: Nothing writes to consumer ignore files any more; only the remove skill's cleanup of old entries remains, and that belongs to the remove skill.
- confidence: high
- flags: The spec should cite the remove skill's ignore-entry cleanup from the remove skill itself (and decision-0073), not from 0049.

### decision-0050 — The M2 morph's rewrite runs in a cold, isolated sub-agent
- cluster: delivery
- disposition: repealed
- in force: none
- retired: `/trellis:setup` M2 §4 dispatches a cold sub-agent for the rewrite — by decision-0065 (morph removed from setup, "not returning to it"; forward pointer) and decision-0072 (setup retired; frontmatter comment)
- spent: the `SKILL.md` M2 §4 edit and manifest advance (gone with the skill)
- reason: There is no morph to isolate: it was removed as speculative, and setup itself is retired.
- confidence: medium — decision-0065:262-265 says 0050 "is not superseded on its merits — a future opt-in morph skill would still want its isolation contract"
- flags: If any spec keeps a "future morph" note, the contract is stated only in this record: the rewrite reads only the target instruction files, the posture and the invariants, never the invoking conversation; interactive bookends stay warm. The remove skill still reverses legacy morphs (`plugins/trellis/skills/remove/SKILL.md:215-243`).

---

## Tally

**By disposition (25):**
- in-force: 1 (0031)
- mixed: 16 (0026, 0027, 0028, 0030, 0034, 0035, 0037, 0038, 0039, 0040, 0042, 0043, 0044, 0045, 0046, 0047)
- repealed: 8 (0029, 0032, 0033, 0036, 0041, 0048, 0049, 0050)
- spent: 0

**By cluster (25):**
- delivery: 10 (0026, 0029, 0031, 0034, 0036, 0039, 0043, 0048, 0049, 0050)
- method: 7 (0035, 0037, 0042, 0044, 0045, 0046, 0047)
- rule-model: 4 (0027, 0033, 0038, 0040)
- checks: 2 (0028, 0030)
- positioning: 2 (0032, 0041)

## Medium / low confidence calls

- decision-0030 (mixed): the surviving "smallest dependency, or none" may just be `inv-minimal-first`; if so the record is repealed.
- decision-0033 (repealed): point 2's retirement is inferred from "no preset ships"; no note retires it, and the 0051 pointer says it "remains parked".
- decision-0037 (mixed): whether D1's lifecycle-shape requirement is still a product rule for consumer methodologies or died with the repository-internal contract.
- decision-0041 (repealed): its note keeps a family-tap rule "for any future kodhama product"; I judged that kodhama's, not Trellis's.
- decision-0043 (mixed): many stacked partial scopes; the `install.sh` scope note is contradicted by 0068/0070 with no pointer.
- decision-0045 (mixed): points 1, 3 and 6 have no subject here (grove retired) rather than an explicit repeal.
- decision-0048 (repealed): vacuous by subject, but its "inline skills carry no git-workflow opinion" principle is stated nowhere else.
- decision-0050 (repealed): decision-0065 says it stays the right contract for any future opt-in morph.

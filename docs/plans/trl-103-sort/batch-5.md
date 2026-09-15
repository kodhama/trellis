# Batch 5 — decision-0085 to decision-0100, plus decision-2026-09-15-rule-rows-only-switch-rules-off

Sorted against the worktree `simplify-1b` at `6f292aa` (2026-09-15). Every "verified" home was opened or grepped in this tree. Line numbers are as of that commit.

### decision-0085 — superpowers planning artifacts are retained
- cluster: method
- disposition: repealed
- in force: none
- retired: whole record — by decision-0097 (`superseded_by`; 0097 point 7). Point 2 (retention of superpowers specs, plans and SDD workspace) and point 3 (`.superpowers/` ignored, `docs/superpowers/` tracked) retired with superpowers. Point 4 (docs-consistency exemption for `superpowers`/`.superpowers`) — replaced by 0097 point 4's top-level `docs/` path skip; tree: cli/docs_consistency_test.go:63-72 says the name skips "retired with superpowers". Point 5 (planning records are history, corrected by dated note) — restated as 0097 point 5. Point 1 only restated decision-0079.
- spent: correction of 0079's Consequences sentence on plan retention (carried forward by 0097 point 7); rewrite of the exemption comment in cli/docs_consistency_test.go (since replaced).
- reason: Superseded in full by 0097, which restates the two points worth keeping on its own ground.
- confidence: high
- flags: `.superpowers/` is still git-ignored (.gitignore:2-4), but as 0097's legacy guard, not under this record. `docs/superpowers/{specs,plans}` still exist (0097 point 6).

### decision-0086 — the injected copy degrades on any over-budget session
- cluster: delivery
- disposition: repealed
- in force: none
- retired: whole record — by decision-2026-09-15-rule-rows-only-switch-rules-off (`superseded_by`; that record's point 5: "decision-0086 is retired in full"). §1–§3 (strip persisted provenance, templates shared by writer and reader, `provenanceOmittedNotice`) — tree: none of `stripPersistedProvenance`, `provenanceOmittedNotice`, `repairMandate`, `QUARANTINE_NOTE` remains in plugins/trellis/hooks/codex-context.mjs; TestCodexBudgetsTheAnnouncementAlongsideTheBody, TestCodexDegradesPersistedProvenanceOnTheMismatchPathToo and TestBothHostsReconcileIdentically are gone from cli/. §4's hard refusal — restated by the 2026-09-15 record point 5 as a runaway guard only (codex-context.mjs:30, :1239). Open question (cap in bytes vs tokens) — lapses with the record.
- spent: the measurement tables and dated notes (history); the partial supersession of decision-0084 §6.
- reason: The rule-rows record retired it in full; its one surviving idea, the 9500 B runaway guard, is restated there.
- confidence: high
- flags: The 9500 B cap now stands on the 2026-09-15 record, not this one. The bytes-versus-tokens open question has no home any more; a spec must raise it again if it still matters. The record's "`readRequired` refuses a rules.toml larger than MAX_CONTEXT_BYTES" is also stale in the tree (the bound is now `MAX_PROJECT_CONFIG_BYTES` = 1 MiB, codex-context.mjs:39), covered by the full retirement.

### decision-0087 — one gateway for every payload read
- cluster: delivery
- disposition: mixed
- in force:
  - point 1 — every plugin-side and vendored-overlay payload read in `staleness.sh` goes through `payload_read`, which returns 0 only for `ok` — home: plugins/trellis/hooks/staleness.sh:183-215; cli/plugin_hook_test.go TestNoPayloadReadBypassesTheGateway (verified)
  - point 2 — the gateway classifies (missing / unreadable / empty / ok) and each call site decides severity — home: staleness.sh:183-215 and its callers (:505, :815, :819, :983) (verified)
  - point 3 (surviving part) — payload files and project files are different classes; absent or empty project files keep their own meanings and are not gateway-guarded — home: the mechanism is verified in staleness.sh; the class rule itself is stated only in the record
  - point 4 — missing and unreadable get different messages (different remedies) and the same disposition; neither is ever silent — home: staleness.sh:185-201 (verified)
  - point 5 — `TRELLIS_STALENESS_UNKNOWN`: an unreadable or malformed plugin `reference/version` withholds only the drift check, and the session stays governed — home: staleness.sh:522, :662 (verified)
  - point 6 — the plugin's version stamp is shape-checked: the whole file must be `payload@` plus exactly 12 lowercase hex — home: staleness.sh:231-275; codex-context.mjs:852 (verified)
  - point 6a — decision-0043 rule 3's "no-op when either stamp is missing or empty" is replaced on all four branches, including the legacy flat `.trellis/version` nudge that names the unreadable stamp and carries no `TRELLIS_` marker — home: staleness.sh:513, :522, :539-544, :662 (verified)
  - point 7 — every Codex `readRequired` call states what an empty read means (`emptyError` / `emptyIsValid`); the default is loud — home: codex-context.mjs:255-277; TestEveryReadRequiredStatesWhatEmptyMeans (verified)
  - point 8 — TestBrokenPayloadIsNeverSilent crosses every file read from `reference/` (not a hardcoded list) with absent / zero-byte / unreadable / truncated on the delivering paths; the invariant is a complete governed injection or a loud `TRELLIS_` marker — home: cli/plugin_hook_test.go:3128-3134 (verified)
- retired:
  - point 3's "one file that is *both*" (`.trellis/rules.toml` repointed at the preset when `rows_are_default=yes`) — by dated note PR #316
  - the provenance's "whose CI guard now fails the higher-numbered claimant at the PR" — by dated note PR #315
- spent: renumbering 0086 → 0087; filing TRL-40 (0078's trigger discharged); the VERSION bump 0.9.0 → 0.10.0 with manifests; inverting two subtests; the pointer on decision-0043 (not re-checked).
- reason: The payload-read gateway, the staleness-unknown marker, the stamp shape check and loud-empty `readRequired` all still ship and are tested; only the preset "both" file and the id-guard aside retired.
- confidence: high
- flags: Open question TRL-39 is still a live host divergence, as recorded, not drift: Codex fails `invalid-version` on a malformed plugin stamp (codex-context.mjs:852-853), while staleness.sh governs and annotates (:522, :662). The "fifteen" count and "180 cells" are dated measurements; do not carry them as current.

### decision-0088 — the install render follows the project's own strictness
- cluster: delivery
- disposition: mixed
- in force:
  - D3 (in part) — `install.sh` renders over an unreadable `.trellis/rules.toml`, does not abort, and says on its own line that no opt-out could be honoured and every rule applies; this is a recorded divergence from the hook, which emits `TRELLIS_RULES_NOT_LOADED` — home: install.sh:821-842, :1007-1009; staleness.sh:1266 (verified); restated by decision-2026-09-15 point 9
  - Consequences, "What is *not* claimed" (the half decision-0090 left standing) — `install.sh` may not read `.trellis/` generally or branch on other project state; each content read is counted and guarded — home: install.sh:742-788; cli/install_script_test.go:1607, :1855 (exactly-three subtest) (verified)
- retired:
  - D1 (header from `strictness`), D2 (fall-through to `trellis-b.md` with `rules-b.toml` seed), D4 (copied strictness parser and pair guard), D5 (second content read) — by dated note PR #316; tree: install.sh:813-819 reads no `strictness`; `reference/` holds no trellis-a/b or rules-a/b
  - D3's "the adaptive header is rendered" wording — by dated note PR #316
  - the Consequences sentence "A third read still has to come and argue itself" — by decision-0090 (`superseded_in_part_by`)
  - Context/Consequences' "the footer must go on saying the rows are authoritative" and "the footer sentence stands" — tree: cli/install_script_test.go:1083 pins a footer of marker, heading, import and stamp only; decision-2026-09-15 point 9. The rows-govern statement now lives in the rendered `rules.md` body (plugins/trellis/reference/rules.md:1).
  - "Not fixed: a `governed = false` project still gets a rendered rules file" (TRL-38) — tree: fixed, install.sh:650-668 refuses the render (decision-0090 D3)
  - "`posture_note` still says the plugin hook falls back the same way" — tree: fixed, install.sh:1009 says "The plugin hook does NOT do this"
- spent: the partial change to decision-0068 (pointer); filing TRL-38.
- reason: TRL-97 retired the strictness-driven header, so only rendering over an unreadable file and the bounded-read boundary remain live.
- confidence: high
- flags: The frontmatter `superseded_in_part_by` comment still says D1, D2 and D4 "STAND in full"; the dated note overrides it, so the comment is stale. The footer-sentence retirement is not named in the dated note, only implied by 2026-09-15 point 9. Both defects this record lists as open (TRL-38, the `posture_note` wording) are fixed in the tree; do not carry them as open.

### decision-0089 — a decision id is claimed at the pull request
- cluster: checks
- disposition: repealed
- in force: none
- retired: whole decision — by dated note PR #315; tree: `.github/scripts/` is absent, `.github/workflows/` has no `decision-id-guard.yml`, and `cli/decision_id_guard_test.go` is absent. Earlier partial retirements: points 1 and 4 by decision-0092; the Consequences conformance clause by decision-0098. The "`main` carries no branch protection" claim is out of date (cli-ci.yml:18-22: `build-test` is a required check).
- spent: point 6 — closing decision-0078's parked observation (pointer and dated note on 0078); discharging TRL-40.
- reason: The id guard it defined was deleted when records moved to date-and-slug names, and AGENTS.md now says no check guards the name.
- confidence: high
- flags: none. Its dated-note-beside-the-claim practice now lives in AGENTS.md's dated-note rule, and its no-runtime ground in decision-0010's fourth bullet.

### decision-0090 — the read budget's count is the test's
- cluster: delivery (D2 overlaps method)
- disposition: mixed
- in force:
  - D1 — `install.sh`'s content-read count is stated and enforced by cli/install_script_test.go's exactly-N subtest, and a change to the count is argued in the comment above it in the same commit, with no record — home: cli/install_script_test.go:1578-1588, :1607, :1855; install.sh:786-788 (verified; the count is now three: marker, `governed`, `@AGENTS.md` gate)
  - D2 — a read that selects, patches, writes or removes a refusal is argued in the corpus before it lands — home: install.sh:776-788; cli/install_script_test.go:1580-1584 (verified that both comments state it); see the conflict flag
  - D3 — `install.sh` reads the top-level `governed` key first, with the hook's matcher copied byte for byte and pinned by TestInstallScriptGovernedParserMatchesHook, guarded regular-and-readable, and refuses the render on `false` — home: install.sh:482-501, :650-668; cli/install_script_test.go (verified)
  - D5 — lowering the count by removing a refusal is not the mirror of raising it; decision-0068 D12's "the next simplification has to come here" survives — home: install.sh:776-782 (verified as D2's fourth clause); the asymmetry argument is stated only in the record
  - Consequences — a read that copies hook logic owes a pair guard whether or not it owes a record — home: TestInstallScriptGovernedParserMatchesHook, TestInstallScriptAgentsImportGateMatchesHook (verified)
- retired: Context and D2's treatment of the `strictness` read as live ("Read 2 … got decision-0088 D5") — tree: install.sh:783-786 (retired with TRL-97); decision-2026-09-15 point 9. No dated note on 0090 records this.
- spent: D4 — correcting decision-0088's third-read sentence as to where (pointer on 0088); correcting the misattribution in the test comment.
- reason: The test still owns the read count, and the governed-key read and its pair guard still ship; the strictness read it discussed is gone.
- confidence: medium — D2 is a record-writing rule that AGENTS.md's newer record test did not reconcile.
- flags: **CONFLICT:** D2, and decision-0068 D12's "next simplification" clause it upholds, require a decision record for any install.sh read that decides. AGENTS.md *Operating method* (TRL-98, PR #313) now writes a record only when a wrong call is expensive to undo, slow to notice or seriously damaging; #313 put notes only on 0014, 0040 and 0082. One must yield. **DRIFT in D1's own home:** cli/install_script_test.go:1599-1605 still says "The render refusal still runs off the ungated $static_conflict, so an un-imported AGENTS.md block refuses exactly as before … Gating the refusal on it WOULD remove one … deliberately left open". That has been false since decision-0091 / TRL-48 gated the refusal (install.sh:645, :734; TestVendorAgentsBlockRefusalFollowsTheImportGate); install.sh:776-788 is correct. No TRL-97 note sits on 0090, and the 2026-09-15 record's list of thirteen noted records omits it.

### decision-0091 — a static-delivery conflict is a conflict for a reader
- cluster: delivery
- disposition: mixed
- in force:
  - D1 — `install.sh`'s render refusal over a managed block in `AGENTS.md` fires only when `CLAUDE.md` carries a standalone `@AGENTS.md` line, read with the hook's anchored, BOM-tolerant matcher (pinned by TestInstallScriptAgentsImportGateMatchesHook). The ground is the adapter contract (decision-0057 rule 2), not how the host resolves imports — home: install.sh:636-647, :734; cli/install_script_test.go TestVendorAgentsBlockRefusalFollowsTheImportGate (verified)
  - D2 — rendering over an un-imported `AGENTS.md` block prints a NOTE about the block and this host only: Codex reads the block, deleting it ungoverns Codex, adding the import line makes both load, nested chains go unseen. An explicit `opted_out = no` test keeps it exclusive of the opt-out NOTE — home: install.sh:1053-1073 (verified)
  - D3 — a block in `CLAUDE.md` is refused regardless of any import line; the decision-0070 D5 opt-out branch is untouched; no other refusal is gated on a guess about the reader — home: install.sh:640-646 (`static_host_reads=no` only for `AGENTS.md`), :650 (verified)
  - D4 — the `@AGENTS.md` gate is classed as a read that removes a refusal, and the read count is unchanged — home: install.sh:771-782 (verified)
  - Consequences — the hook catches a late-arriving `@AGENTS.md` line with `TRELLIS_STATIC_SHAPES_CONFLICT`; an installer re-run warns that the rendered file "ALREADY EXISTS" — home: staleness.sh:477; install.sh:797-803 (verified)
- retired: none by note or later record. (The `block-inline-b.md` fixture named in Context retired with TRL-97; `reference/` now has only `block-inline.md`. That is history, not a rule.)
- spent: D5 — changing decision-0068 D12 in part, scoped by paragraph (pointer on 0068); filing TRL-49.
- reason: The import-gated render refusal, its NOTE and the unchanged CLAUDE.md refusal all ship in install.sh with tests.
- confidence: high
- flags: The known, accepted regression on nested `@`-import chains is still open (TRL-49; install.sh:602). The test comment at cli/install_script_test.go:1599-1605 contradicts D1 (see 0090). The provenance's "code lands separately" is now done (install.sh:645).

### decision-0092 — a claim is a new record path
- cluster: checks
- disposition: repealed
- in force: none
- retired: whole decision — by dated note PR #315; tree: the guard script, workflow and test are deleted. The claim path in point 1 and point 4's exit-2 path were already changed by decision-0099.
- spent: updating the AGENTS.md derivative line (since rewritten); discharging TRL-46; the pointer on decision-0089.
- reason: It restated the claim rules of an id guard that no longer exists.
- confidence: high
- flags: none. The maintainer's "factual corrections get dated notes; rule divergences get successors" was deliberately not minted as a rule here, and AGENTS.md now handles partial retirement by dated note (PR #313).

### decision-0093 — a consulted pointer falls back; a delivered payload does not
- cluster: delivery
- disposition: mixed
- in force:
  - D1 — delivered payload and consulted reference follow different rules: the overlay's three delivered files are required and never fall through to the plugin (the directory is the discriminator), while `invariants.md` is consulted and never fails a session closed — home: codex-context.mjs:45 (`VENDORED_PAYLOAD`), :799-803, :976-1000 (verified)
  - D2 — on the vendored branch, when the overlay has no `invariants.md` and the plugin's copy is usable, only the consulted pointer moves to the plugin's copy; every source stays the overlay's — home: codex-context.mjs:976-1000; TestCodexRepointsWhenAVendoredOverlayLacksInvariants, TestCodexLeavesAVendoredInvariantsPointerAlone (verified; "exists" is read as "usable" per decision-0094 D4)
  - D3 — the hosts differ on a vendored overlay (Claude's hook injects nothing, Codex reads the overlay), and `reference/` stays an installation source, not a runtime substitute — home: plugins/trellis/README.md:27-37 (verified)
  - Consequences, the `existingFile` bullet (surviving part) — `existingFile` keeps a flattened presence contract for the overlay-presence half — home: codex-context.mjs:96-115, :976 (verified)
- retired: Consequences bullet 1, "handed a pointer that resolves", in four cells — by decision-0094; the `existingFile` bullet's "No caller here can act on the difference between missing and unreadable" — by decision-0094.
- spent: D4 — marking decision-0065:191-193 (pointer on 0065); the VERSION bump 0.14.0 → 0.15.0 with manifests.
- reason: The consulted-versus-delivered rule and the pointer-only fallback still ship; two Consequences clauses were overtaken by 0094.
- confidence: high
- flags: "README.md:27" in this record means plugins/trellis/README.md, not the root README. decision-0096 found the "runtime substitutes" clause cited at :27 when it sits a few lines lower (it is at :32-33 today). "install.sh:529" is now :525. Whether the vendored path should exist at all is still decision-0072's open question (TRL-31).

### decision-0094 — both hosts classify a broken consulted reference
- cluster: delivery
- disposition: mixed
- in force:
  - headline + D1 — a host that cannot say which fault it found has not reported it; Codex classifies the fault (`payloadDefect`) rather than Claude relaxing — home: codex-context.mjs:160-200, :1284 (verified)
  - D2 — both hosts share one byte-identical four-phrase fault vocabulary, pinned behaviourally by TestBothHostsReportAMissingInvariantsTarget — home: staleness.sh:185-208, :1461; codex-context.mjs:110-115, :1284; cli/invariants_pointer_test.go (verified)
  - D4 — the fallback's plugin-copy condition asks whether the copy is usable (`payloadDefect`), not merely present — home: codex-context.mjs:985-986; TestCodexDoesNotFallBackOnAnUnusablePluginCopy (verified)
  - D5 (except its closing clause) — the overlay-presence half keeps `existingFile`, a presence test — home: codex-context.mjs:976, :1055 (verified)
  - Consequences — "empty" means every byte is a newline or a NUL, matching `$( )`; mixed shapes are deliberately unpinned — home: codex-context.mjs:140-156, :186-193 (verified)
  - Consequences — the report rides `systemMessage`, outside the `MAX_CONTEXT_BYTES` bound; a broken delivery is byte-identical to the healthy one on the same fixture; no absolute byte count is claimed — home: codex-context.mjs:1304, :1364 (verified)
- retired: D5's closing "and nobody has made it" — by decision-0096; "What is not decided here" bullet 1 (both copies unusable, silent) — by decision-0095; bullet 2 and the count "Two shapes stay silent" — by decision-0096.
- spent: D3 — marking decision-0093's clause; the VERSION bump 0.19.0 → 0.20.0; filing TRL-71 and TRL-73.
- reason: Shared fault classification across hosts and the usability test on the fallback still ship; its "not decided" list was closed by 0095 and 0096.
- confidence: high
- flags: none. The precedent counts in its `changes:` comment are a dated inventory, not a rule.

### decision-0095 — a vendored delivery names its own dead pointer
- cluster: delivery
- disposition: mixed
- in force:
  - headline + D1 — a delivery that has proved its own pointer dead says so on every branch; the vendored arm (no overlay copy, unusable plugin copy) reports on `systemMessage`, never through `fail()` — home: codex-context.mjs:1010-1034, :1289-1321 (verified)
  - D2 — that report blames the overlay, names the plugin copy second as an unavailable substitute, and words the remedy for both repairs — home: codex-context.mjs:1318-1321 (verified)
  - D3 — both halves carry a classification, and no eligibility check widens (the report classifies; the `existingFile` test does not change) — home: codex-context.mjs:976, :1023-1032 (verified)
  - D4 — the shared vocabulary: "no readable <path>", the four phrases, "yields nothing to read" — home: codex-context.mjs:1318; staleness.sh:1461 (verified)
  - D5 (ruling only) — the report stays outside the context bound, so it cannot tip a governed session into refusal — home: codex-context.mjs:1304 (verified)
  - D6 (surviving part) — the cell is Codex-only as far as the hook goes: staleness.sh exits path A before any repoint, so the guard lives in TestCodexDoesNotFallBackOnAnUnusablePluginCopy, not the pair guard — home: plugins/trellis/README.md:45-48; cli/invariants_pointer_test.go (verified)
  - D7 — silence-asserting tests are narrowed with `warningsBesideVendoredInvariants`; `writeCodexPluginRoot` is not given a `reference/` — home: cli/invariants_pointer_test.go:990; cli/codex_hook_test.go:53-63 (verified)
- retired: "What is not decided here" (the overlay's own unusable copy says nothing) — by decision-0096; D6's "so Claude never delivers an invariants pointer into a vendored project and has nothing to report about" — by decision-0096 (false when written).
- spent: marking decision-0094; splitting `claudeContextFor` / `claudeStdoutFor`; the VERSION bump 0.20.0 → 0.21.0.
- reason: The vendored dead-pointer report and its overlay-blaming wording still ship; D6's Claude claim was measured false, and the open cell was closed by 0096.
- confidence: high
- flags: D5's "32-byte token" is wrong (33 bytes), flagged by 0096 and left for TRL-79; it is a figure, not a rule. D3 cites TestCodexSaysNothingWhenTheOverlayCarriesItsOwnInvariants, which is now TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting (the old name is absent from cli/). The provenance's session-independence claim is contradicted by commit trailers, per 0096 (TRL-79). The Claude-side gap (TRL-76) is still open, as plugins/trellis/README.md:45-48 says.

### decision-0096 — an authoritative overlay is reported on, not substituted for
- cluster: delivery
- disposition: mixed
- in force:
  - D1 — an overlay's own present-but-unusable `invariants.md` (zero-byte, newline-only, NUL-filled, mode 0000) is never substituted for, and `existingFile` stays a presence test — home: codex-context.mjs:976, :1038-1060; TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting; plugins/trellis/README.md:40-44 (verified)
  - D2 — the report blames the overlay alone, never names the plugin path, offers no reinstall, and is byte-identical whether the plugin copy is healthy or missing — home: codex-context.mjs:1351-1354; the same test (verified)
  - D3 / D4 — decision-0095's headline covers this cell, and the adjacency argument is why — home: stated in codex-context.mjs comments around :1040-1080 (verified in part); the argument is otherwise stated only in the record
  - D5 — the shared lead and classification, with the guard renamed rather than duplicated — home: codex-context.mjs:1351; cli/invariants_pointer_test.go (verified)
  - D6 — the report rides `systemMessage` outside the bound, and the delivered context is unchanged — home: codex-context.mjs:1344 (verified)
  - Consequences — the report fires only on "is empty" and "exists but could not be read" (a race guard, deliberately unpinned) — home: codex-context.mjs:1113-1128 (verified)
  - Consequences — the report requires the invariants token to be present in either delivery half (prose or rules body) — home: codex-context.mjs:1129-1135; TestCodexSaysNothingWhenTheVendoredProseCarriesNoPointer, TestCodexReportsWhenTheVendoredPointerArrivesThroughTheRulesHalf (verified)
  - Consequences — the overlay's own copy is read (a single 4 KB chunk) on every vendored session — home: codex-context.mjs:175-193, :1099 (verified)
- retired: none.
- spent: marking decision-0094 and decision-0095; renaming the guard; the VERSION bump 0.21.0 → 0.22.0 with manifests.
- reason: The no-substitution ruling and the overlay-only report still ship with guards; its open items are filed, not decided.
- confidence: high
- flags: Four filed items are still open and named nowhere in the tree except the Claude gap in README: TRL-76 (Claude's static chain carries the pointer silently; plugins/trellis/README.md:45-48), TRL-77 (a symlink, directory or FIFO at the overlay path is substituted silently), TRL-78 (0095's arm and the plugin-native arm lack the pointer-presence check), TRL-81 (`String.replace` `$` hazard). Context's "reference/trellis-a.md and trellis-b.md carry the pointer" is outdated: both retired with TRL-97, `reference/trellis.md:9` carries it now, and frozen overlays keep their old copy (cli/codex_hook_test.go `frozenFirmOverlayHeader`).

### decision-0097 — Compound Engineering replaces superpowers
- cluster: method
- disposition: mixed
- in force:
  - point 1 — CE is enabled at project scope and superpowers is set `false` in `.claude/settings.json`; cloud sessions need the environment's setup script (TRL-91) — home: .claude/settings.json:4-5, :14-19 (verified); the cloud half is stated only in the record
  - point 2 — work between a decision and a merge uses CE's skills: `ce-brainstorm`/`ce-plan`, `ce-work`, `ce-code-review`, `ce-resolve-pr-feedback`, `ce-commit-push-pr`; the maintainer merges — home: AGENTS.md:99 "plan a build" row (verified; that row omits `ce-resolve-pr-feedback`)
  - point 3 — CE artifacts are committed, each on CE's own lifecycle: `docs/plans/` and `docs/ideation/` are kept as records; `docs/solutions/` is maintained by `ce-compound-refresh` (update or delete, git is the archive); `CONCEPTS.md` at the root is live; `.context/compound-engineering/` is scratch and never committed — home: .gitignore:13; AGENTS.md:100 (solutions row) (verified in part); the plans / ideation / CONCEPTS.md lifecycle is stated only in the record
  - point 4 — the top-level `docs/` is exempt from TestDocsClaimOnlyRealCommands' walk, by location; `marketplaceCommands` still reads `docs/`; `CONCEPTS.md` gets no exemption — home: cli/docs_consistency_test.go:63-72, :687-705 (verified)
  - point 5 — planning records are history: a wrong claim is corrected by a dated note beside it, never an edit; rename sweeps (decision-0015:117-121) still apply; learnings are not history — home: AGENTS.md:70-73 (a plan is never the place a note names, and a citation in a plan is not live) (partial); the correction-by-dated-note rule for plans is stated only in the record
  - point 6 — the superpowers specs and plans in `docs/superpowers/` stay as they are — home: docs/superpowers/{specs,plans} present (verified); the rule is stated only in the record
  - point 7 (legacy guard) — `.superpowers/` stays git-ignored — home: .gitignore:2-4 (verified)
  - Consequences — the guard walks skip `.context/compound-engineering/` at the walk root only — home: cli/docs_consistency_test.go:107-113 (verified)
- retired: point 4's "`cli-ci`'s path filter keeps `docs/**` for it" — by dated note PR #312; tree: cli-ci.yml:13-16 has no filter.
- spent: point 7 — superseding decision-0085 in full and deleting this checkout's `.superpowers/` (verified absent); the forward pointers on 0079 and 0085; the AGENTS.md row rewrite; replacing the docs-consistency name skips.
- reason: CE is still the enabled workflow, AGENTS.md routes work through it, and the docs/ exemption ships; only the path-filter aside retired.
- confidence: medium — point 3's plans/ideation lifecycle and point 5 rest partly on inference the record itself labels, and AGENTS.md states only fragments of them.
- flags: `CONCEPTS.md` does not exist yet, so its rule is latent. "Superpowers stays installed in `review-kit`" is about another repository and was not verified.

### decision-0098 — artifact conformance is a CI check
- cluster: checks
- disposition: mixed
- in force:
  - point 1 — a Go test applies the artifact contract to the corpus: cli/corpus_conformance_test.go (checks 1–7, TestCorpusConformsToArtifactContract, and the positive control TestCorpusConformanceRejectsKnownBadFixture over `cli/testdata/known-bad/`); cli/corpus_conformance_typed_test.go (checks 8–11); a missing or unparseable input halts by name; cli/artifact_contract_guard_test.go pins the check to the rubric; it runs in `build-test`; `corpus-reviewer` is retired — home: those files; AGENTS.md:148-160; `.claude/agents/corpus-reviewer.md` absent (verified)
  - point 2 — every rubric check part has a recorded outcome in `contractOutcomeTable` / `contractClauseOutcomes`; 2c, 2d, 5b and the charter's "derive your checklist" are dropped; 5d and 10b stay review judgment, written in AGENTS.md — home: cli/corpus_conformance_test.go; cli/artifact_contract_guard_test.go:41; AGENTS.md:154-160 (verified)
  - point 6 — the check carries no exemption for decision-0044's `kodhama/…` entry or research-0010's `informed_by` edge — home: cli/corpus_conformance_test.go (no match for either; verified)
  - point 7 — in profiles/trellis-self.md, where C2 is `independent-agent` the gatekeeper is decision-0007's review workflow with `ce-code-review`, and the Go check is supporting evidence only — home: profiles/trellis-self.md:32, :57, :88-91 (verified)
- retired:
  - point 1's "that touches a corpus path, the rubric, the fixtures or the check itself"; all of point 5 (not required, does not block); the Consequences' "`cli/ci_paths_guard_test.go` forces the filter…", "the `cli-ci` filter mirror it, and guards pin both" and "TRL-92's other paths … stay open" — by dated note PR #312; tree: cli-ci.yml:13-22 runs `build-test` on every PR as a required check; cli/ci_paths_guard_test.go is absent
  - "`decision-id-guard` remains a separate check" and "Both conclusions still hold, because `decision-id-guard` is not a required check" — by dated note PR #315
  - point 4's "until TRL-95 lands, the rubric sits at `core/rubrics/`" — carried out by decision-0100; tree: docs/rubrics/artifact-contract.md
- spent: point 3 — clause-scoped pointers on 0010, 0076, 0089 and 0099; point 4 — the ruling that the rubric leaves `core/` (done by 0100); deleting corpus-reviewer.md; the same-PR file edits listed in Consequences.
- reason: The CI conformance test, its outcome table, the two judgment rules in AGENTS.md and the profile's gatekeeper framing all hold; the trigger scope and "not required" statements were retired by #312 and #315.
- confidence: high
- flags: none. (Checked: #312 is the TRL-96 PR that removed the path filter, so the dated notes' PR number is correct; #314 is TRL-96's second PR.)

### decision-0099 — the governance corpus lives under docs/
- cluster: method
- disposition: mixed
- in force:
  - point 1 — the corpus lives at `docs/decisions/` and `docs/research/`, and `eval/` stays at the repository root — home: those directories (verified); cli/corpus_conformance_test.go:45 `corpusRoots`; AGENTS.md "Which layer is this?" paragraph (verified); `eval/` at the root (verified); the ground "docs/ shouldn't hold runnable code" is stated only in the record
  - point 3 (tripwire half) — a root `decisions/` or `research/` directory fails the Go suite — home: cli/selfapply_test.go:200-209 (verified)
  - point 7 — decision-0005 stands: a corpus under `docs/` is outside `core/` and still the build methodology — home: AGENTS.md "Which layer is this?" paragraph (verified)
  - Consequences — `docSurfacesIn` no longer skips directories named `decisions` or `research`, because the top-level `docs/` path skip covers the corpus — home: cli/docs_consistency_test.go:55-72 (verified)
- retired:
  - point 2 (claim rule), point 4's "once this merges … the guard sees it", point 6's "`decision-0089` is not superseded" — by dated note PR #315
  - point 3's filter clause, point 5, and the TRL-92 item under *Not decided here* — by dated note PR #312
  - the Consequences bullet on the corpus paragraph and `corpus-reviewer`'s charter naming the new paths together — by decision-0098 (`superseded_in_part_by`)
- spent: the `git mv` of both folders; point 4's one-off hand check of `decision-0099`; point 6's pointer on 0092; point 8's rename sweep; sweeping the path in the codex-context.mjs comment (a release).
- reason: The docs/ location, the root tripwire and the layer statement hold; the guard and path-filter clauses retired with the guard and the filter.
- confidence: high
- flags: Broken outside links to `blob/main/decisions/…` were accepted and are not guarded; links from Linear were never checked.

### decision-0100 — the artifact contract is repository-internal
- cluster: method
- disposition: mixed
- in force:
  - point 1 — the artifact contract is repository-internal: the rubric sits at `docs/rubrics/artifact-contract.md` with `scope: trellis-meta`; its positive control is in `cli/testdata/known-bad/`; the typed-artifacts schema stays in `core/` because it defines product types — home: docs/rubrics/artifact-contract.md:1-9; cli/corpus_conformance_test.go:53-56; core/schemas/typed-artifacts.md (verified)
  - point 3 — a product artifact does not name a repository-internal id; invariants-v1 no longer names `rubric-artifact-contract` — home: core/invariants/trellis-invariants-v1.md:148-150 (verified); the general rule is stated only in the record
  - point 4 — the `docs/` exemption covers `docs/rubrics/` — home: cli/docs_consistency_test.go:63-71 (verified)
  - Consequences — the README repo map lists `core/` without the rubric — home: README.md:297 (verified)
  - Consequences — nothing fails if `core/rubrics/` or `core/fixtures/` reappear; this was accepted, not guarded — home: stated only in the record and its dated note
- retired: the Consequences' "`cli-ci` no longer selects them" — by dated note PR #312.
- spent: the `git mv` of the rubric and fixtures; point 2's pointer on decision-0010 (bullet 1's parenthetical, and bullet 2 as it describes the product); the path-token sweep over eleven records and four plans.
- reason: The rubric's repository-internal home and the docs/ exemption for it hold; the moves and pointers are done.
- confidence: high
- flags: Whoever sorts decision-0010 must read point 2: with 0098, bullet 2 is superseded whole, and Open question 2 has no product subject left.

### decision-2026-09-15-rule-rows-only-switch-rules-off — rule rows only switch rules off
- cluster: rule-model
- disposition: in-force
- in force:
  - point 1 — only an `active = false` row has effect; a rule with no row applies; an empty file means every rule applies; `active = true` rows, repeated rows, `strictness` and `seeded_from` are inert; `governed = false` keeps its meaning (exactly one uncommented top-level line, above every table header, read first) — home: plugins/trellis/reference/rules.md:1; plugins/trellis/reference/block-codex.md:9; install.sh:482-501; staleness.sh:323-330; plugins/trellis/README.md:152, :167 (verified)
  - point 2 — the activation rule is stated once, in `rules.md` — home: plugins/trellis/reference/rules.md:1 (verified; block-inline.md:8 embeds the same sentence)
  - point 3 — a bad entry costs that entry, never the session: each listed kind is ignored with a warning naming the line number and slug or key, never the raw line; five warnings are named, then a count; a file with a NUL byte delivers every rule, is not shown, and draws one warning; a project file that is non-regular, unreadable or over 1 MiB is refused loudly on both hosts — home: codex-context.mjs:39, :484, :487-677; staleness.sh:723, :1040, :1051-1060, :1085, :1088-1092, :1266, :1291; cli/rules_rows_parity_test.go TestBothHostsClassifyRulesRowsIdentically (verified)
  - point 4 — both hooks frame the file identically: `## Project rule activation`, a computed sentence naming the rules switched off, the file shown unchanged only when it and its warnings fit 1800 B (on Codex an invalid UTF-8 byte counts as 3), otherwise a too-large line, then the warnings. A symlinked file is classified but never shown. The Codex bootstrap fallback does not read a linked file. A hook's activation section counts as loaded. Codex also puts warnings on `systemMessage`, and its vendored branch uses the same framing — home: staleness.sh:1050, :1358; codex-context.mjs:625, :639-677; block-codex.md:9, :15 (verified)
  - point 5 — nothing reconciles or rewrites; the caps of 9500 B (Codex) and 32768 B (Claude) remain only as runaway guards; the plugin path stays within 400 B, pinned by the parity test — home: codex-context.mjs:30, :1239; staleness.sh:1410; cli/rules_rows_parity_test.go:67 (verified; no reconciliation code remains)
  - point 6 — one header ships (`trellis.md`, "By default"), with no presets, and the shipped inline block carries no rows — home: plugins/trellis/reference/trellis.md:5; `reference/` has no rules-a/b, trellis-a/b or per-posture blocks; block-inline.md has no rows section (verified)
  - point 7 — a new file is exactly the two lines; `install.sh` seeds them only when no file exists; the hook's accept instruction quotes the same bytes, and a test pins the two; an existing file is never touched — home: install.sh:470; staleness.sh:541, :723; cli/plugin_hook_test.go:166, :2470; cli/docs_consistency_test.go:519-524 (verified)
  - point 8 — adoption is unchanged: on user scope and on Codex the file is the adoption act; a bundle vendored under `.claude/skills/` with no file applies every rule — home: plugins/trellis/README.md:18-25 (verified)
  - point 9 — the installer renders the one header whether or not it can read the file; the footer is marker, heading and rows import; over an unreadable file it says no opt-out could be honoured and every rule applies — home: install.sh:813-842, :1007-1009; cli/install_script_test.go:1083 (verified)
  - point 10 — frozen-text shapes (vendored overlay, inline block, a file rendered by an older installer) keep a full row for every rule until they migrate — home: plugins/trellis/README.md:171-172, :181-182; README.md:48 (verified)
  - point 11 — a retired slug is never reused — home: stated only in the record (not found in core/, cli/, plugins/, README.md or AGENTS.md; TestRowSetDerivativesFollowThePin checks set membership, not reuse)
- retired: none
- spent: retiring rules-a.toml, rules-b.toml, trellis-a.md, trellis-b.md and the per-posture inline blocks (verified absent); retiring decision-0086 in full; the thirteen dated notes.
- reason: Merged today and not retired; every point except the retired-slug rule is carried by the hooks, payload, installer, README and tests.
- confidence: high
- flags: Point 11 has no home and no guard outside this record, so the spec must state it newly. Open question: neither new wording (the activation sentence or the computed switched-off sentence) has been measured; research-0012 measured the old one. The TRL-100 host divergences in Consequences are known and unfixed. Its list of thirteen noted records omits decision-0090, whose Context and D2 still treat the strictness read as live.

---

## Tally

**By disposition (17 records)**
- in-force: 1 (2026-09-15)
- mixed: 12 (0087, 0088, 0090, 0091, 0093, 0094, 0095, 0096, 0097, 0098, 0099, 0100)
- spent: 0
- repealed: 4 (0085, 0086, 0089, 0092)

**By cluster**
- delivery: 9 (0086, 0087, 0088, 0090, 0091, 0093, 0094, 0095, 0096)
- method: 4 (0085, 0097, 0099, 0100)
- checks: 3 (0089, 0092, 0098)
- rule-model: 1 (2026-09-15)
- positioning: 0

## Medium / low confidence calls

- **decision-0090 (medium):** D2's rule that an install.sh read which decides needs a decision record conflicts with AGENTS.md's newer record test (PR #313), and nothing reconciles the two.
- **decision-0097 (medium):** point 3's plans/ideation lifecycle and point 5 rest partly on inference the record labels itself, and AGENTS.md states only fragments of them.

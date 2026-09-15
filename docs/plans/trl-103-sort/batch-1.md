# Batch 1: decision-0001 to decision-0025

Sorted 2026-09-15 against the worktree `simplify-1b` at `6f292aa` (TRL-97 merged). Line numbers are for that tree. "verified" means I opened the home file and it states or does the point. "stated only in the record" means no current file states it. The required-checks list was read from the GitHub rules API (`repos/kodhama/trellis/rules/branches/main`) on 2026-09-15.

### decision-0001 — Product form: a portable pack, validated by dogfooding
- cluster: positioning
- disposition: mixed
- in force:
  - Decision — Trellis is a self-serve, shippable, portable pack (instructions, gates, sub-agents), not a consulting service — home: README.md:5-8, README.md:16-22 (plugin, curl and manual copy paths) (verified)
  - Decision — Trellis is validated by dogfooding this repository, not through a consulting practice — home: AGENTS.md:9-11, README.md:276 (verified)
  - Consequence 2 — N=1 is accepted for now, and instance #2 is an open need — home: core/invariants/trellis-invariants-v1.md:355-356, README.md:289-290 (verified)
- retired:
  - Consequence 3 "distribution / go-to-market is explicitly deferred" — overtaken by decision-0012, decision-0019 and decision-0043, and by the tree (README.md:16-22). It was a deferral, not a rule, and has no dated note.
- spent: none
- reason: The product-form stance still holds and README states it. Only the "distribution deferred" placeholder was overtaken by later delivery records.
- confidence: high
- flags: none

### decision-0002 — Adaptation is a user-controlled dial, not a mode choice
- cluster: rule-model
- disposition: mixed
- in force:
  - Decision + Refinement — how far a project adopts or adapts a reference methodology is the user's choice, not a mode Trellis sets. The working stance is two coarse modes (adopt one framework, or adapt from several), deliberately not formalized into the invariant model — home: core/invariants/trellis-invariants-v1.md:95, :259-266, :370-373 (verified that the set names "the `decision-0002` dial"; it only points back here, so the substance is stated only in the record)
  - Refinement, via decision-0021 — the dial now carries the adopt/adapt role that the retired `inv-reference-relationship` used to front — home: trellis-invariants-v1.md:264-265 (verified)
- retired:
  - Decision "Trellis must support sitting anywhere on it" and Consequence "support both extremes from one mechanism", as a shipped mechanism — the only shipped form, the conductor (A) and author-adapt (B) postures, was retired by decision-2026-09-15-rule-rows-only-switch-rules-off point 6. decision-0033's dated note (TRL-97, PR #316) records this. This record has no note.
  - Consequence 3 "enough seeds to spawn a coherent methodology" as a catalog design target — `seed` was parked by decision-0033 point 2 and never came back. The tree has no seed.
- spent: none
- reason: The adopt/adapt stance now survives only as a pointer in the invariant set. The postures that expressed it in the product were retired on 2026-09-15.
- confidence: medium — no record says whether "support anywhere on the dial" is still a product commitment or was dropped when TRL-97 removed the postures.
- flags: **DRIFT candidate.** trellis-invariants-v1.md:264-265 still says "the adopt/adapt choice stays `decision-0002`'s dial". Since TRL-97 no shipped surface lets a project choose adopt or adapt: plugins/trellis/reference/trellis.md:5 gives every project one "By default" sentence, and `.trellis/rules.toml:3-4` `seeded_from`/`strictness` do nothing (README.md:35-36). TRL-97 added dated notes to 13 records, but not to 0002 or 0020. The spec author must decide whether the dial is still a commitment.

### decision-0003 — Methodology-agnostic: inspect or be told, then supervise
- cluster: positioning
- disposition: in-force
- in force:
  - Decision — Trellis is methodology-agnostic by construction. No framework is privileged, and a built-in default reference is optional — home: README.md:5-8 (verified)
  - Decision — Trellis finds the methodology by inspecting the project or reading a standardized instruction file, and supervises only a methodology that clears the structural invariants — home: core/invariants/trellis-invariants-v1.md:44-46, :316-317 describe the ingestion check as the model. **Not implemented:** signature-catalog-v1.md:471-473 says "Fold in when the ingestion check (`decision-0003`) is built", and nothing under plugins/ or in install.sh inspects a methodology. This point is stated in the record and invariant-set prose only.
  - Consequence 1 — the first machinery owed is a standardized way to describe a methodology to Trellis — stated only in the record (never built)
  - Consequence 3 — spec-kit and BMAD are reference points to harvest from, not adoption targets — home: README.md:5-8 (verified)
- retired: none
- spent: none
- reason: The agnostic stance is live in README and the invariant set. The inspect-or-be-told machinery it calls for was never built, and nothing retired it.
- confidence: medium — I call the ingestion point in force because the invariant set still describes it. The maintainer may consider it dropped, since the setup flow retired (decision-0043, decision-0072).
- flags: Unbuilt commitment. The shipped plugin delivers rules to any project without working out its methodology or applying an admission gate. decision-0005's Layer-B clause also waits on this format "once it exists".

### decision-0004 — Buyer-neutral by construction; validate the invariants first
- cluster: positioning
- disposition: mixed
- in force:
  - Decision — build buyer-neutral: startup and enterprise differences live in gates, steps, artifacts and dial settings layered on the shared invariants — home: core/invariants/trellis-invariants-v1.md:49-50, :282-283, :320-321; README.md:245-248 (verified)
- retired: none
- spent:
  - Decision / Consequence 1 — validate the invariant set first, as the project's first deliverable and gate — done: invariants-v1 ratified 2026-06-29 (trellis-invariants-v1.md:1-12)
  - Consequence 2 — "the meta of relaxing invariants" is out of scope for now — a scoping deferral with no standing rule
- reason: Buyer-neutrality is still the model the invariant set states. The "validate the invariants first" step is done.
- confidence: high
- flags: Consequence 2 is overtaken in practice: per-rule opt-out (`<slug> = { active = false }`, README.md:33-38; decision-0051, decision-2026-09-15-rule-rows-only-switch-rules-off) now ships relaxation as a feature. No record says so.

### decision-0005 — Trellis self-hosts: separate Trellis-core from the build methodology
- cluster: method
- disposition: mixed
- in force:
  - Decision 1 — Trellis-core (Layer A, the product) and the build methodology (Layer B) stay separate and must not leak into each other; the product never ships build cruft — home: AGENTS.md:12-16; core/README.md:3-6; cli/payload.go:24-30 (the payload excludes artifact-contract metadata and catalog-maintainer governance prose, decision-0055). decision-0100 point 1 applies the split to the rubric (verified)
  - Decision 2, as amended by decision-0057 — the root instructions file is Layer B, instance #1, not the product's agent instructions. The canonical file is AGENTS.md and CLAUDE.md is the adapter — home: AGENTS.md:15-16, AGENTS.md:173-180; CLAUDE.md contains `@AGENTS.md` (verified)
  - Decision 3 — Trellis-core lives in its own namespace, `core/` — home: AGENTS.md:13-14, core/README.md. decision-0099 point 7 and decision-0100 point 1 both say "decision-0005 stands" (verified)
- retired:
  - Decision 2 "`CLAUDE.md` is Layer B", as the physical location — by decision-0057 point 6 (superseded_in_part_by; the comment is accurate)
- spent:
  - Consequence 3 — the physical reorg into core/ — done (core/invariants, core/catalog, core/schemas, core/lexicon.md)
  - Consequences 1-2 and Open question 2 — instance #1 as the first test of the admission gate — done as profiles/trellis-self.md
  - Open question 1 — the namespace name — settled as `core/`
- reason: The Layer A/B split and the core/ namespace are live and restated in AGENTS.md. Only the CLAUDE.md location moved (0057), and the reorg is done.
- confidence: high
- flags:
  1. core/README.md is stale in its home. Lines 3-4 describe core/ as "rubrics, sub-agents, conventions", and line 15 says the conformance sub-agent's "product home is `core/agents/`, which the delivery slice (`0012`) will package". Against decision-0098 and decision-0100 there is no rubric or sub-agent in the product. Its contents list also omits catalog/ and lexicon.md.
  2. Decision 2's "to be expressed in the standard instruction-file format (decision 0003) once it exists" waits on a format that was never built (see 0003).

### decision-0006 — Research before ratifying the invariants
- cluster: positioning
- disposition: spent
- in force: none
- retired: none
- spent:
  - Decision — the invariant set stayed draft and was validated in three research steps. Step 0 is docs/research/0001-target-landscape.md, Step 1 is docs/research/0002-gate-test-tier1.md, and Step 2 produced invariants-v1, ratified 2026-06-29 (core/invariants/trellis-invariants-v1.md:1-22)
  - Consequence 2 — findings feed back as revise-in-place revisions — became decision-0014's rule and is in force there
- reason: The research-before-ratification sequence ran and ended with invariants-v1 ratified. Nothing in the record is a standing rule.
- confidence: high
- flags: The general habit of validating invariant claims with evidence rather than assertion survives only implicitly, in invariants-v1's durability tags and open questions (:62-65, :333-376). A spec that wants it as a principle must state it fresh.

### decision-0007 — Automated PR review
- cluster: checks
- disposition: mixed
- in force:
  - Decision — a workflow runs the official `code-review` plugin on every PR (`opened`, `synchronize`) — home: .github/workflows/claude-code-review.yml:12-13, :45 (verified)
  - Decision — it authenticates with `CLAUDE_CODE_OAUTH_TOKEN`, and the `ANTHROPIC_API_KEY` swap is documented in the workflow — home: claude-code-review.yml:5-7, :43 (verified)
  - Decision — it is an assistive reviewer, not a merge gate — home: the required checks on `main` are Analyze (go), Analyze (javascript), build-test and release-guard, with no `review` (GitHub rules API, 2026-09-15). The human merge gate is AGENTS.md:87-90 (verified)
  - Consequence 3 — minimal-first: only the review workflow, no `@claude` mention workflow — home: .github/workflows/ has no mention workflow (verified). As a rule it is stated only in the record.
  - Downstream — this workflow is the named independent-agent gatekeeper in profiles/trellis-self.md:113-114 (decision-0098 point 7) (verified)
- retired: none
- spent:
  - Prerequisites — install the Claude GitHub App and add the secret, a one-time admin act, recorded in claude-code-review.yml:3-8. The tree cannot show whether the secret is set.
  - Consequence 2 — the workflow stays inert and fails loudly until configured — a one-time state
- reason: The review workflow still runs on every PR as an advisory check and is not a required check, as decided.
- confidence: high
- flags: The workflow has grown a prompt override (re-review each push, no waiting on background agents, scoped ranges), `show_full_output` (:112-118) and an allowed-tools list. Their reasons live only in workflow comments, and a checks spec should carry them. A green `review` check does not prove a review ran (:112-117).

### decision-0008 — Enforcement is a configurable layer; Trellis surfaces, the org configures
- cluster: rule-model
- disposition: mixed
- in force:
  - D1 — enforcement strength is a dial (`expressed` → `default-on-but-skippable` → `enforced`), and strictness is opt-in — home: core/invariants/trellis-invariants-v1.md:280-293; the catalog's `default_C1` on each entry (e.g. signature-catalog-v1.md:102); core/schemas/typed-artifacts.md:43, :87 (verified as a concept; no shipped surface sets it, see flags)
  - D2 — what cannot be negotiated away is surfacing, not enforcement: a skip must be a visible choice — home: trellis-invariants-v1.md:297-306 (`floor-transparency`); the delivered directives for `floor-transparency` and `inv-gate-at-handover` in plugins/trellis/reference/rules.md (verified)
  - D5 — the gatekeeper per gate is `independent-agent | human | none`, and `none` is allowed only when the skip is surfaced — home: trellis-invariants-v1.md:288-290; typed-artifacts.md:44, :88; the profiles/trellis-self.md rows (verified as a concept)
  - Open question 3, answered — the intent gate is the floor that can never be set to `none` — home: trellis-invariants-v1.md:307-310; signature-catalog-v1.md:450 (verified)
  - D3 — skipping reuses the underlying framework's own skip machinery where it exists — stated only in the record
  - D4 — the target UX is conversational ("the next step is X — you may skip it") — stated only in the record (never built)
  - D5 — Trellis may recommend a gatekeeper, but the choice belongs to the org — the recommendation half is the catalog's `default_C2`; "the choice is the org's" is stated only in the record
- retired: none by record or note (see drift flag)
- spent:
  - Open question 1 — surfacing elevated to its own floor, `floor-transparency`, in invariants-v1 (though trellis-invariants-v1.md:342-343 still keeps "its own floor, or loud-failure?" open)
- reason: The dial-plus-floor model proposed here is the live invariant structure, but since `strictness` stopped doing anything neither dial has a surface a consumer can set.
- confidence: medium — invariants-v1, the catalog and the schema state the dials as live, yet TRL-97 made `strictness` inert and nothing sets a gatekeeper. "The org configures" may be a live rule the product no longer implements.
- flags:
  - **DRIFT candidate.** D1 and D5 say the gatekeeper is configurable, the choice is the org's, and strictness is opt-in. In the tree, README.md:35-36 says "`strictness` and `seeded_from` do nothing", plugins/trellis/reference/trellis.md:5 gives every project one posture, and decision-2026-09-15 point 1 agrees. The dials survive only as catalog defaults and profile columns. TRL-97 added no dated note here.
  - core/lexicon.md:46 still maps decision-0012's "activation level" onto this dial.
  - Incidental stale citation: README.md:246 points to `trellis-invariants-v1.md:234` for the dial values, which now sit at :285.
  - decision-0070 and decision-0083 (and decision-0077 alongside `floor-intent-gate`) cite D2's surfacing floor as authority, so D2 is load-bearing.

### decision-0009 — How Trellis improves itself (validation scenarios + feedback loop)
- cluster: positioning
- disposition: mixed
- in force:
  - A, honest limit — scenarios run on projects we shaped test mechanism, not generalization. math-quest is genealogically N=1, and an independent instance #2 is still owed — home: profiles/trellis-self.md:81; README.md:289-290; trellis-invariants-v1.md:355-356 (verified)
  - B, human gate — Trellis suggests changes to itself and never merges them; core changes are never auto-applied — home: AGENTS.md:87-90 (an agent may not merge without the maintainer's act, `floor-intent-gate`) (verified, general form)
  - B, consent — a project's process feedback is recorded only with the user's explicit permission, never exfiltrated silently — stated only in the record (no feedback channel exists)
  - B, triage — weight signal by instance diversity, not raw volume; the gold signal is a human overriding Trellis and being right — stated only in the record
  - Consequence 3 — improvements route to the project, the methodology or Trellis-core, and the core tier needs cross-instance recurrence — stated only in the record (decision-0040:17-25 applies it to math-quest reverse-ports)
- retired: none
- spent: none
- reason: The N=1 honesty and "Trellis suggests, never merges" still hold. The consent-based feedback loop designed here was never built, and nothing retired it.
- confidence: medium — whether the unbuilt loop (issue type, triage agent, critical-mass threshold) is still a commitment or a stale proposal is a judgment call.
- flags:
  - The GitHub-issue feedback channel does not exist. .github/ISSUE_TEMPLATE/config.yml routes work tracking to Linear (decision-0075) and has no feedback issue type. decision-0075 says nothing about consumer feedback.
  - A's three validation scenarios (empty project, math-quest, a spec-kit project) have no run artifact: eval/experiments/ holds only annotation-vs-absence and does-trellis-help. They are an unrun plan, not a rule. I can't confirm whether Linear tracks them.
  - Consequence 1 ("the spine must capture surfaced friction as structured, exportable records") was never built, and the spine and spec stage are gone.

### decision-0010 — Trellis imposes no runtime; it ships as agent instructions
- cluster: delivery
- disposition: mixed
- in force:
  - Bullet 1, minus its parenthetical — Trellis's resources are agent instructions that need no runtime, interpreted by and composed into the host's agent surface alongside what is already there — home: README.md:229-233, :153-154; plugins/trellis/reference/trellis.md:3 ("They add guardrails; they don't replace this project's own instructions"); install.sh:506-526 detects a conflicting static delivery shape (verified)
  - Bullet 1, as amended by decision-0065 (note under the title) — on the plugin path a hook is the only carrier. The Claude hook is bash (plugins/trellis/hooks/staleness.sh:65), the Codex hook is Node (plugins/trellis/hooks/codex-context.mjs; README.md:152 requires Node 20), and the install path stays runtime-free (verified)
  - Bullet 3 — a support CLI or installer is allowed for install, scaffolding and ops, as support only and never a runtime dependency — home: cli/main.go:53-60 (generator-only `payload`), cli/README.md:14, install.sh (verified)
  - Bullet 4 — any deterministic helper for CI gating is written in the target project's own stack, never in a runtime Trellis imposes — home: cli/corpus_conformance_test.go, which decision-0098 point 1 rests on this bullet; AGENTS.md:136-156 (verified)
  - Open question 1 — which agentic surfaces come first — decision-0100's comment keeps it standing. In practice README.md:16-22, :130-159 answer it: Claude Code is primary, Codex is carried but not supported.
- retired:
  - Bullet 1's parenthetical "(including the artifact contract and its conformance check)" — by decision-0100 point 2 (the rubric is at docs/rubrics/artifact-contract.md, scope trellis-meta) (verified)
  - Bullet 2 "the validator is a conformance sub-agent applying a rubric, not a program", in full — by decision-0098 point 3 for this repository and decision-0100 point 2 for the product (verified: `.claude/agents/corpus-reviewer.md` is gone)
  - Bullet 3's "the methodology runs without it" — narrowed on the plugin path by decision-0065
  - Open question 2 (how the sub-agent runs in CI) — answered for this repository by decision-0098; decision-0100's Consequences say it has no product subject
- spent:
  - Consequence 1 — "the Python vs Node stack question is closed" — a one-time framing, since overtaken by decision-0023 (Go) and the Node Codex hook
- reason: The instructions-first, no-runtime stance and the own-stack rule for CI helpers still hold. The conformance-agent clauses retired in 0098 and 0100, and 0065 narrowed the runtime-free claim for hooks.
- confidence: high
- flags: The frontmatter comments match decision-0065, decision-0098 and decision-0100 (checked). "Requires no runtime" is literally false on Codex, where the hook needs Node, as decision-0065 says, so the spec should state the narrowed claim rather than bullet 1's wording. The core/README.md staleness from 0005's flag applies here too.

### decision-0011 — Add a spec stage between decisions and build
- cluster: method
- disposition: repealed
- in force: none
- retired: the whole record — by decision-0079 (`superseded_by`). Verified: there is no `specs/` directory, and planning uses the Compound Engineering skills (AGENTS.md:99; decision-0097).
- spent: none
- reason: decision-0079 retired the spec stage in full, and specs/ no longer exists.
- confidence: high
- flags: The pointer note in 0011 still says planning "moves to the superpowers skills". decision-0097 replaced superpowers, as 0079's comment records. The idea that tests derive from acceptance criteria survives only as AC#/R##/S## markers in cli/install_script_test.go and cli/codex_hook_test.go (decision-0079 Consequences). `spec` stays a recognized type under decision-0079 point 5, which is 0079's rule, not this one.

### decision-0012 — Delivery: own marketplace (v0) → support CLI (v1) → git copy always
- cluster: delivery
- disposition: mixed
- in force:
  - Decision 3 — git copy-in always works, because the resources are instructions — home: README.md:161-227 (manual copy path), plugins/trellis/reference/ (verified)
  - Open question 2 / Consequence 2 — coexist with the host's existing instructions file, never overwrite it — home: plugins/trellis/reference/trellis.md:3; README.md:184-227 (pick exactly one delivery branch; managed blocks between markers); install.sh:506-526 (verified)
  - Context / Decision 1 — a plugin means activation, not mere availability: its hooks fire on events — home: plugins/trellis/hooks/hooks.json; README.md:52-53 (the rules are injected at session start by the plugin's hook) (verified)
  - Decision 5 — a `claude-community` submission is a later growth move — stated only in the record (no mention anywhere in the tree)
- retired:
  - Decision 1 "host our own Claude Code plugin marketplace (`marketplace.json` in the Trellis repo)" — by the tree: there is no marketplace.json in the repo, README.md:16-22 installs from the family marketplace `kodhama/stewards`, and decision-0066:62-64 records Trellis listed there. No dated note.
  - Decision 1 "the plugin repo is the update channel (push → `/plugin marketplace update`)" — by AGENTS.md:142-148 (a payload change is a release; installs re-pull only when plugins/trellis/VERSION moves; `release-guard`) and decision-0061's SemVer package authority (per decision-0043's approval record)
  - Decision 2 — a v1 support CLI that wires the pack into non-plugin surfaces and carries ops — by decision-0043 (generator-only CLI; cli/main.go:53-60)
  - Decision 4 — install offers activation levels (available + referenced → hooks fire → default agent) as the surfaced enforcement choice — by the tree: README.md:24-26 (installing at project scope is the adoption act and every rule applies; decision-0070); plugins/trellis/.claude-plugin/plugin.json sets no default agent (verified). No dated note.
- spent:
  - Consequence 1 — delivery runs in parallel to the spine — historical
  - Consequence 2 — the spine spec must specify the activation and wiring contract — the spec stage retired (decision-0079); the wiring now lives in decisions 0065, 0068 and 0070 and their tests
  - Open question 1 — hooks and skills per dial level — moot now Decision 4 is retired
- reason: Git copy-in and coexisting with host instructions still hold. The self-hosted marketplace, the v1 CLI and the activation-level dial were all overtaken without dated notes.
- confidence: medium — I read the marketplace and activation-level retirements from the tree and from decision-0066 and decision-0070, not from a note naming this record. I can't confirm which kodhama record moved the marketplace to `kodhama/stewards`.
- flags: Three points were retired silently, with no dated note and no `superseded_in_part_by`: the own marketplace, the update channel and the activation-level dial. core/lexicon.md:46 still lists "activation level" as a delivery-lens synonym, a term with no live mechanism.

### decision-0013 — Invariants carry stable slugs; references use slugs
- cluster: rule-model
- disposition: mixed
- in force:
  - Decision 1 — every invariant has a stable slug, its canonical id, recorded in the invariants-v1 Identifiers registry — home: core/invariants/trellis-invariants-v1.md:69-100 (verified)
  - Decision 3 — references use slugs. A merged or retired slug is `superseded_by` its survivor in the registry, so any old reference, including a legacy code, still resolves — home: trellis-invariants-v1.md:75-78, :95; rubric check 4, docs/rubrics/artifact-contract.md:64-71 (verified)
  - Decision 3 — no citing artifact is edited to chase a retirement or merge (the historical-reference exemption) — home: trellis-invariants-v1.md:75-76, :102-103; artifact-contract.md:69-71. decision-0079 point 3a reuses it for spec ids (verified)
  - Strengthened later — retired codes are never reassigned (decision-0038), and a retired slug is never reused (decision-2026-09-15-rule-rows-only-switch-rules-off point 11) — stated in those records only
- retired:
  - Decision 2 — the A/B/C/D ordinals kept as frozen display labels — by decision-0038 (verified: the registry column now reads "Legacy code (retired)", trellis-invariants-v1.md:80-100)
- spent:
  - Consequence 3's deferred items — slugs in invariant headings, and moving machinery references off ordinals — done by decision-0038's sweep (headings are slug-first at trellis-invariants-v1.md:117-310)
- reason: Stable slugs and registry resolution are live in the invariant set and the rubric. decision-0038 retired only the old codes' display role.
- confidence: high
- flags:
  - **Contradiction candidate.** This record says no citing artifact is edited to chase a rename. decision-0015's correction (0015:117-121) and AGENTS.md:57-58 say a rename sweep rewrites the renamed token everywhere, append-only records included. The two cover different cases (a slug supersession versus a global rename), but no current text draws that line.
  - The schema field names `default_C1`/`default_C2` still carry the retired codes (decision-0038's open question; typed-artifacts.md:43-44).

### decision-0014 — Invariant set: a compiled spec plus change records
- cluster: rule-model
- disposition: mixed
- in force:
  - Decision 1 — the invariant set is a compiled, revise-in-place, current-truth spec — home: core/invariants/trellis-invariants-v1.md:18-22, :327-331 (verified)
  - Decision 2, second half — minor or editorial edits are plain spec edits with no record — home: AGENTS.md's record test (*Operating method*) and trellis-invariants-v1.md:328-331 (verified; now folded into the record test)
  - Decision 3 — the spec carries its rationale by reference, with no inline change narrative — home: trellis-invariants-v1.md:19-22, :327-331 (verified)
- retired:
  - Decision 2 "a change to an invariant's meaning or structure gets a record" and Consequence 3 "an ADR per significant invariant change going forward" — by the dated note TRL-98 (PR #313), pointing to AGENTS.md's record test (verified: trellis-invariants-v1.md:21 and :329-330 state the record test)
  - Decision 1 "it sits at `ratified` and is not re-ratified per change" — the status field was retired by decision-0082; merging is the acceptance (AGENTS.md:77-80)
- spent:
  - Decision 4 — invariants-v0 superseded, file retirement deferred — done: the file is deleted and the id resolves through the registry (trellis-invariants-v1.md:102-107)
  - Consequence 1 — the shipped file becomes a clean compiled spec — done
  - The migrated change-history table (0014:47-72) — history, not a rule
- reason: Revise-in-place with rationale by reference still governs the invariant set. TRL-98's note retired the "record per significant change" rule.
- confidence: high
- flags: The change-history table is the only home of the v0→v1 rationale. It is history and should stay in the record rather than move into a spec. core/lexicon.md:80-81 asks whether the lexicon should be change-tracked "like the invariant set (`decision-0014`)", which now points at a retired rule.

### decision-0015 — Rename the product Bonsai → Trellis
- cluster: method
- disposition: mixed
- in force:
  - Correction block (0015:117-121) — a rename sweep rewrites the renamed token wherever it appears (heading, filename, slug, body prose), in every artifact class including append-only records. A quotation containing a renamed token follows its source — home: AGENTS.md:57-58 permits "a rename sweep (`decision-0015`)" only by pointer. decision-0099 point 8 applies the rule and adds exemptions, and decision-0097:109 says it "still applies" (verified, short form only)
  - Amendment rule 1 — a closed issue is frozen and is not swept — stated only in the record
  - Amendment — issue comments are exempt as dated utterances. The genuine historical set is not swept: this record, docs/research/0008 Part 2, merged PR titles #1/#10/#31, and git history — stated only in the record
  - Decision — the invariant framing stays "our synthesis, v1"; only the possessive changed — home: AGENTS.md:116-121 (naming guardrail); trellis-invariants-v1.md:12-16 (verified)
- retired:
  - Amendment rule 2 as first written ("a quotation of an append-only record pins") — by this record's own correction block (0015:104-123)
  - Correction block's open question (whether an appended amendment is an edit, or needs a superseding record) — answered by AGENTS.md *Operating method*: a dated note is the sanctioned addition (decision-0082 as noted by TRL-98 and TRL-99)
- spent:
  - Decision — the rename of repo content, three files and the GitHub repo — done (README.md:1; the repo is now kodhama/trellis, README.md:85)
  - Follow-up (a) — rename the working directory — done (Projects/trellis)
  - Follow-up (c) — sweep "bonsai" from open issues #22–#28 — owed per the amendment; I can't confirm from the tree whether it happened
  - Follow-up (b) — auto-memory files — outside the repository
- reason: The rename is done. What stays live is the rename-sweep rule and its exemptions, which AGENTS.md cites only by pointer.
- confidence: medium — follow-up (c) cannot be checked without reading GitHub issues, and the "closed issue frozen" rule and the exemption list have no home outside this record.
- flags: Several records cite this record by line range (`decision-0015:117-121`: 0097, 0098, 0099, 0100), so a spec must carry the rule's full text, including rule 1 and the exemption list; AGENTS.md says only "a rename sweep". It is in tension with decision-0013 (see there). Planned work moved to Linear (decision-0075), so "a closed issue" now spans two trackers, which nothing addresses.

### decision-0016 — Two typed artifacts: the signature catalog and the expression profile
- cluster: rule-model
- disposition: mixed
- in force:
  - Decision, first type — `signature-catalog`, scope `trellis-product`, one shipped. For each slug it gives what the invariant is, its observable signature, its class, and default C1/C2 — home: core/catalog/signature-catalog-v1.md:1-9, :82-450; core/schemas/typed-artifacts.md:23-53; rubric check 8 (docs/rubrics/artifact-contract.md:138-145); cli/corpus_conformance_typed_test.go:18 (verified)
  - Decision, second type — `expression-profile`, scope `core-methodology`, one per instance: per slug `active`, C1 and C2, plus instance-level delivery axes — home: profiles/trellis-self.md:1-8; typed-artifacts.md:66-96; rubric checks 9-10 (verified)
  - Relationship — the catalog is the dictionary and a profile is one instance's readout against it — home: typed-artifacts.md:85 (a profile slug must resolve to a catalog entry); rubric check 9 (verified)
- retired:
  - "Delivery axes A/B per invariant" — revised to instance-level before ratification, as the record itself says (typed-artifacts.md:71-79)
  - "Ratified by the human (D2)" as a `draft → ratified` status flip — by decision-0082 (no status; merge is the acceptance). The home schema was not updated (see drift flag).
- spent:
  - Consequence 2 — spec-0002 owed — written, then migrated to core/schemas/typed-artifacts.md by decision-0079 point 3
  - Consequence 3 — the conformance check learns both types — done: rubric checks 8-11 and cli/corpus_conformance_typed_test.go
  - Consequence 1 — unblocks the research-0009 build order — historical
- reason: Both typed artifacts exist, have their schema in core/schemas and are checked in CI. The Assess and Apply roles and the ratify flip they assumed are gone.
- confidence: medium — the profile's roles "produced by Assess, consumed by Apply, diffed across instances (#28), minimized into Trellis-lite (#22)" have no implementation. I treat them as unbuilt, not retired, because no record retires them.
- flags: **DRIFT in the home file.** core/schemas/typed-artifacts.md:100-107 still states a `draft → ratified` lifecycle, an Assess sub-agent that produces profiles, and an Apply that consumes them. decision-0082 retired status, and nothing under plugins/ mentions Assess or Apply (only core/ does). Delivery reads `.trellis/rules.toml`, not a profile (README.md:24-50). Cross-instance diffing and Trellis-lite are unbuilt.

### decision-0017 — A canonical lexicon, and names for the delivery-relationship dial
- cluster: positioning
- disposition: mixed
- in force:
  - Decision 1 — a `lexicon` type, scope `trellis-product`, one shipped (`lexicon-v1`), with required sections Canonical terms and Open questions — home: core/lexicon.md:1-9, :36, :76; docs/rubrics/artifact-contract.md:53-55, :101; cli/corpus_conformance_test.go:47 reads core/lexicon.md (verified)
  - Decision 2 — three nested registers: garden (identity), gene (the official teaching register), invariant (the substrate). `invariant` stays canonical for what is enforced, and gene talk is our teaching metaphor, never a provenance claim — home: core/lexicon.md:20-31 (verified)
  - Decision 2 policy — current-truth and product artifacts use the canonical term and link the lexicon; research notes keep their lens vocabulary — home: core/lexicon.md:30-31 (verified; no rubric check or test enforces it)
  - Decision 3 — the delivery-relationship dial's ends are `supervisor` (installed, live) and `advisor` (referenced, pulled); `consultant` is retired — home: core/lexicon.md:49, :54-60; core/schemas/typed-artifacts.md:77; README.md:258-272 (verified)
  - Open question 2 — `supervisor` stays one word for both the control role and the dial end — home: lexicon.md:64-66 (verified)
- retired: none
- spent:
  - Consequence 1 — the `consultant → advisor` sweep — done (lexicon.md:33-34)
  - Consequence 2 — the conformance check learns the lexicon type — done (rubric :53-55, :101); since decision-0098 a Go test applies the rubric instead of the reviewer agent
  - Open question 1 — the pull-end name — resolved as `advisor`
- reason: The lexicon type, the three-register policy and the supervisor/advisor names all live in core/lexicon.md and README, and the one-time sweep is done.
- confidence: medium — the cluster is a judgment (vocabulary could sit under rule-model), and one reading below may be drift.
- flags: **Possible naming DRIFT.** README.md:260-266 calls the plugin path "Advisor (open, no runtime — shipped)". Yet the plugin is installed and its SessionStart hook injects the rules live each session (README.md:52-53), which lexicon.md:54-55 defines as the supervisor end ("installed and running live"), not advisor ("consulted from outside… no runtime tie"). Current truth has not settled whether the plugin is advisor or supervisor.

### decision-0018 — Restore process self-improvement (trigger-driven loop)
- cluster: rule-model
- disposition: mixed
- in force:
  - Placement — `inv-self-improvement` is a first-class operating invariant, a neighbour of `inv-graph-maintenance`, with `inv-prune-bias` shared between them — home: core/invariants/trellis-invariants-v1.md:88, :93, :185-205; signature-catalog-v1.md:195-214; the delivered directive in plugins/trellis/reference/rules.md (verified)
  - SI-1 `inv-propagation-surfaced` — improvement signals are surfaced through the project's chosen channel, never silently dropped, and a retirement lands in the same change — home: trellis-invariants-v1.md:189-193; signature-catalog-v1.md:200-204 (verified)
  - SI-2 `inv-ride-existing-rituals` and SI-3 `inv-prune-bias` — home: trellis-invariants-v1.md:198-202 (verified)
  - Channel — asked or inferred per instance, never assumed; Trellis does not mandate GitHub issues or a `## Propagation` section — home: trellis-invariants-v1.md:190-192 (verified)
  - Disposition — proactively notice a signal, propose both the fix and a standing trigger, ask rather than act; weakly checkable — home: trellis-invariants-v1.md:194-197 (verified)
  - Engine — triggers, not vigilance: a `condition → action` recorded where its firer will trip over it. It rides existing work, retires a trigger when actioned, prefers retiring to adding, and stays subordinate to product work — home: signature-catalog-v1.md:200-202 and AGENTS.md:105-106 (short form, verified). The eleven trigger examples are stated only in the record.
  - Port intent, not mechanism, from math-quest — stated only in records (applied by decision-0040:17-19)
  - Consequence 2 — the conformance check learns SI-1, checked against the declared channel — stated only in the record (unbuilt: no SI-1 or channel rule in docs/rubrics/ or cli/*.go)
  - Consequence 3 — the improvement channel as an `expression-profile` field — stated only in the record (not in typed-artifacts.md §2)
  - Consequence 4 — Trellis-self declares its own channel — stated only in the record (AGENTS.md names none, though its decision-0078 row routes next steps to a named consumer)
- retired: none
- spent:
  - Consequence 1 — invariants-v1 revised to un-merge self-improvement — done (trellis-invariants-v1.md:185-205)
  - Open questions 1 (placement) and 7 (does this reopen the merge) — resolved inside the record
- reason: Self-improvement is back as a first-class invariant, with its three facets stated in the invariant set. The checker, profile field and self-declared channel this record promised were never built.
- confidence: medium — the unbuilt consequences are owed work rather than rules. I list them as in force so the spec author decides whether to carry or drop them.
- flags: Open questions still unanswered anywhere: the SI-1 slug name, the default channel when none can be inferred, and merging `inv-prune-bias` with graph-maintenance's own prune bias. That bias is still stated twice (trellis-invariants-v1.md:156-157 and :199-200), the double count the record warned about. The SI sub-slugs have no rows in the Identifiers registry (:82-100).

### decision-0019 — Free & open, MIT-licensed; advisor CLI as the v0 on-ramp
- cluster: positioning
- disposition: mixed
- in force:
  - Decision 1 — Trellis is free to use, with no monetization now — home: README.md:372-377; site/index.html:604-605; site/lp-content.md:201-206 (verified)
  - Decision 2 — MIT license, with the maintainer's handle as copyright holder — home: LICENSE:1-3 ("Copyright (c) 2026 gundisalwa") (verified)
  - Decision 3 — keep the Apache-2.0 upgrade path open — home: README.md:375-376 (verified)
  - Decision 3 — add a DCO (`Signed-off-by`) the day external contributions start — stated only in the record. Not triggered yet: `git shortlog -sne HEAD` shows only the maintainer, Claude and dependabot.
  - Decision 4 — any future monetization is services on top, never a paywall on the core — home: README.md:376-377; site/index.html:605 (verified)
- retired:
  - Decision 5 — delivery v0 is advisor mode through a CLI the host invokes, and the CLI-advisor path is the v0 deliverable — by decision-0043 (the end-user CLI channel retired) and the tree (README.md:16-22: the plugin is the primary path). The advisor-first idea survives in README.md:260-266, carried by the plugin, not a CLI. No dated note.
  - Consequence 2 — decision-0012 refined to the CLI-advisor target — retired with Decision 5
- spent:
  - Consequence 1 — the landing page drops paid tiers — done (site/index.html:604)
  - Consequence 3 — the README license line reads MIT — done (README.md:374)
- reason: Free use, MIT, services-only monetization and the Apache path are all live in README, LICENSE and the site. The CLI-as-v0-delivery point died with the binary channel.
- confidence: high
- flags: decision-0043's supersedes list (0043:112-115) does not name 0019, so Decision 5 has no pointer. The DCO trigger has nothing watching for it; under `inv-no-orphan-followups` it needs a named consumer or a recorded drop.

### decision-0020 — Goals and examples live in the rules; the landing page derives from them
- cluster: rule-model
- disposition: mixed
- in force:
  - Decision 1 — each catalog entry carries `why` (one line, agents first) and contrastive `honored`/`violated` examples — home: signature-catalog-v1.md entries and :456-460; core/schemas/typed-artifacts.md:37-40 (verified). decision-0027 refines this to at least two matched pairs.
  - Decision 2, the meta-rule — every invariant must carry why, honored and violated. Presence and pairing are checked by rubric check 8 (docs/rubrics/artifact-contract.md:138-145) and cli/corpus_conformance_typed_test.go:18, :244-254 (verified). The other half, that editing an invariant without updating its examples is a failure, is stated at signature-catalog-v1.md:44-47 and typed-artifacts.md:60-62 but has no mechanical check (weakly checkable, as the record says).
  - Decision 3 — the invariants and benefits page is a derived, simplified view of the catalog, with no claim on the page lacking a rule behind it — home: site/invariants.html:530; cli/sync_test.go:55-68 (every example appears on the page); cli/row_set_guard_test.go:114-116 (cards); signature-catalog-v1.md:49-53 (verified)
  - Decision 4 — "no silent drift" is a cross-cutting theme (`floor-transparency` generalized), not a new invariant — home: core/invariants/trellis-invariants-v1.md:53-60 (verified)
- retired:
  - Consequence 3 — the A·conductor and B·author-adapt preset profiles ship on the shelf, C·seed is the starting point, and the advisor CLI offers A/B/C-seed/Custom onboarding — seed and custom were parked by decision-0033 point 2, the A/B presets retired by decision-2026-09-15-rule-rows-only-switch-rules-off point 6, and the setup CLI and skill retired by decision-0043 and decision-0072. No dated note on this record.
  - Open question 2 (the presets' exact dial values) — moot now the presets are retired
- spent:
  - Consequence 1 — the catalog schema gains why/honored/violated, and the check learns completeness — done
  - Consequence 4 — folded into the un-merge pass — historical
- reason: Goals and paired examples live in the catalog and are checked in CI, and a test keeps the invariants page from drifting away from it. The preset profiles announced here are gone.
- confidence: high
- flags: Consequence 2 says the landing page's "how it works" section also derives from the catalog. Only site/invariants.html is guarded (cli/sync_test.go); site/index.html is not, so "no claim without a rule" holds there by discipline only. On Open question 1, no render step exists (nothing in scripts/): the page is rendered by hand and checked by a test. The presets were retired with no dated note here.

### decision-0021 — Collapse `inv-reference-relationship` (B8): no distinct mechanism
- cluster: rule-model
- disposition: mixed
- in force:
  - Context test — collapse an invariant when it adds no distinct mechanism (an existing floor or rule applied to one object); keep it, or mint a new one, when it introduces a mechanism — stated only in records (applied as the live test by decision-0040:30-32, decision-0074:28, decision-0078:167 and decision-0081:216; trellis-invariants-v1.md:59-60 gives the reason for this collapse, not the test itself)
  - Decision 1 — `inv-reference-relationship` is a retired slug that resolves to `floor-transparency`. Its framework-divergence case is a `floor-transparency` example, and the adopt/adapt choice stays with decision-0002's dial — home: trellis-invariants-v1.md:95, :99, :259-266; signature-catalog-v1.md:405-424 (verified)
  - Decision 2 — a significant change to the invariant set triggers a prune and collapse review of the set; the trigger is change, not time — stated only in the record (no hit in AGENTS.md, core/, docs/rubrics/ or plugins/)
- retired: none
- spent:
  - Consequences — invariants-v1 revised, the catalog entry dropped, the 15→14 cascade carried out in increment 2 — done. The set has since grown to 16 (signature-catalog-v1.md:454).
  - Open question 1 — review through a non-Claude model, "owed when the tooling exists" — the tooling now exists and is used (decision-0100's self-check: ce-code-review with a Codex cross-model pass). No rule in AGENTS.md requires it.
- reason: The retirement is done and resolves through the registry. The collapse test and the review-the-set-when-it-changes trigger are live rules that exist only in records.
- confidence: medium — Open question 1 could be read as a standing rule (cross-family review of invariant changes) rather than done; I found no text stating it as a rule.
- flags: The mechanism test is the de facto rule for minting or collapsing any invariant, and no spec-like file states it. The adopt/adapt dial this record hands B8's role to lost its shipped form in TRL-97 (see 0002).

### decision-0022 — Merge = ratify: the ratified state is core, the workflow is per instance
- cluster: method
- disposition: mixed
- in force:
  - Decision 1 — the ratified state is core (an approvable state, human acceptance at the intent gate, producer ≠ ratifier), and how acceptance happens is each instance's call — home: trellis-invariants-v1.md:127-133 (`inv-ratifiable-artifacts`) and :344-352 (trellis-self's ratifiable state is "merged on `main`"); the principle in decision-0037 Decision 1. decision-0082's comment on this record says Decision 1 stands (verified)
  - The core of Decision 2, kept and promoted by decision-0082 — in this repo, merging to `main` is the acceptance, riding the merge rather than a separate ceremony — home: AGENTS.md:77-80, :87-90 (verified)
- retired:
  - Decision 2's in-PR `status: draft → ratified` flip and "no draft left un-ratified on main" — by decision-0082, and earlier by decision-0042's `gated` flip, which 0082 also retired (verified: AGENTS.md:77 "There is no `status` field")
  - Decision 3 — the agent proposes the flip in the PR, and the merge needs no automation — by decision-0082
  - Consequence 3 — spec-0001's "two consumable states or one?" answered "keep two" — reversed by decision-0082 to none
  - Consequence 1 — the convention goes in CLAUDE.md under Gates — the home moved to AGENTS.md (decision-0057), and decision-0082 replaced the content
- spent:
  - Consequence 2 — the draft limbo ends: signature-catalog-v1 and profile-trellis-self are ratified at their next merge — done (signature-catalog-v1.md:4; profiles/trellis-self.md:4)
  - Open question 1 (an automated flip on merge) — moot now there is no status field
- reason: The core of this record is live in AGENTS.md: merging is the acceptance, the state belongs to core, and the workflow to each instance. decision-0082 retired every status-flip mechanic.
- confidence: high
- flags: Open question 2 (does merge = ratify extend to a supervised project's own artifacts?) is answered only by decision-0037 Decision 1's principle that each methodology defines its lifecycle. The shipped `inv-ratifiable-artifacts` signature still expects a status lifecycle (trellis-invariants-v1.md:348-352 records this as known evidence). The legacy note at 0022:12 cites decision-0042's `gated` flip, which is itself retired.

### decision-0023 — Trellis's first code: Go, no package manager; the dev cycle
- cluster: checks
- disposition: mixed
- in force:
  - Decision 1 — Trellis's code is written in Go — home: cli/go.mod (`module github.com/kodhama/trellis/cli`, `go 1.22`) (verified)
  - Decision 3 — the code lives in its own module, `cli/`. Build, test and lint run in GitHub Actions on every PR, and the job is required — home: .github/workflows/cli-ci.yml:14-17, :20-23 and its steps (`npm run quality` including staticcheck, `go build`, `go vet`, `go test -count=1`, a 90% coverage floor); `build-test` is a required check on `main` (GitHub rules API); AGENTS.md:134-156 (verified)
  - Decision 3 — test-first for non-trivial logic — stated only in the record (no hit in AGENTS.md, README.md, cli/README.md or docs/solutions/)
  - Decisions 3-4 — code rides the same PR ritual (the review workflow, and merging as acceptance), and the CLI is support tooling that no agent needs at run time — home: cli/main.go:53-60; README.md:229-233 (verified)
  - Context — no package manager on the user's install path (the enterprise npm wall) — home: README.md:16-22, :66-86 (plugin marketplace, curl of a POSIX script, no npm) (verified in effect; the reason itself is stated only in the record)
- retired:
  - Decision 2 — distribution through GitHub Releases plus a `curl … | sh` binary installer — by decision-0043 (verified: no auto-release.yml or release.yml in .github/workflows/)
  - Consequence 1 — Releases and a curl endpoint for the binary — by decision-0043
  - Decision 1 "a single static cross-platform binary for the user" — no binary ships any more (decision-0043; README.md:231-233)
- spent:
  - Consequence 2 — the cli/ module, its CI, and a dev-cycle note in the instructions file — done (AGENTS.md "Checks and review")
  - Consequence 3 — spec-0003 assumes this stack — the spec retired (decision-0079)
  - Open questions — the curl endpoint host (moot), the prompt library (the TUI was deleted, decision-0043), the module layout (settled: `cli/`)
- reason: Go in cli/, with CI build, vet and test required on every PR, still holds. decision-0043 retired the binary-release distribution.
- confidence: medium — the header note at 0023:14-27 says point 1 "survives verbatim". I split it, because nothing ships a single static binary to users any more.
- flags: **SUPERSESSION NOTE PARTLY WRONG.** The note at 0023:14-27 says point 1 (Go, a single static binary) "survives verbatim". Go survives; the user-facing static binary does not (README.md:231-233). The repo's quality gate now also needs Node tooling (package.json; cli-ci `npm ci` and `npm run quality`), and the Codex hook is Node. That sits outside this record's "no package manager" framing, which was about end users.

### decision-0024 — Gatekeeping is detect-and-respect; Trellis surfaces silent human-gate bypasses
- cluster: rule-model
- disposition: mixed
- in force:
  - Decision 1 — the host project owns its gate declarations. Trellis reads and respects them, and never builds, chooses or imposes a per-gate map — stated only in the record (decision-0081:212-216 treats it as binding)
  - Decision 2 — Trellis's job is surfacing: `inv-gate-at-handover` at the dial's strength, scoped by the gatekeeper dial and `floor-intent-gate`, and made loud by `floor-transparency`. It is a composite, not a new invariant — home: trellis-invariants-v1.md:160-163; the delivered `inv-gate-at-handover` and `floor-intent-gate` directives in plugins/trellis/reference/rules.md (verified)
  - Decision 3 — one direction only: a human-gated handover run without its human approval is surfaced, agent-gated handovers proceed, and Trellis never flags human approval on agent gates — stated only in the record
  - Decision 4 — where a project declares no gate, recommend the posture default only, never gate by gate — stated only in the record
  - Decision 5 — per-gate configuration by Trellis is deferred to v2 — stated only in the record (decision-0081:242-243 cites it as still deferred)
  - Consequence 1 — no per-gate structure in the profile — home: core/schemas/typed-artifacts.md:81-91 has none (verified)
- retired:
  - Consequence 2's carrier "M1 overlay / M2 morph" — the M2 morph is retired (README.md:62-65, :265), and rules now reach sessions through the plugin hook or the curl-rendered file. The surfacing behaviour itself survives (see Decision 2).
- spent:
  - Consequence 4 — simplifies the setup CLI — the CLI and the setup skill are retired (decision-0043, decision-0072)
- reason: Detect-and-respect, surfacing silent human-gate bypasses, is still the gate model, but only the surfacing half is written anywhere outside this record.
- confidence: medium — Decisions 1, 3 and 4 have no implementing text. The delivered `floor-intent-gate` directive, "Unsure whether a human must approve? Assume yes" (rules.md), sets a default for undeclared gates that Decision 4 does not describe.
- flags:
  - Consequence 1's "Assess populates it from the project" relies on an Assess step that does not exist under plugins/ (see 0016).
  - Consequence 3's testable target (a human-gated spec→plan run without approval gets surfaced) has no test or eval: "human-gated" has no hit in eval/, cli/ or plugins/.
  - Tension to settle: the `inv-gate-at-handover` directive surfaces any skipped review, agent-gated or not, while Decision 3 says agent-gated handovers proceed silently. They concern different things (skipping a check versus missing a human approval), but no current text separates them.

### decision-0025 — Keep the landing page, docs and release in sync automatically
- cluster: checks
- disposition: mixed
- in force:
  - Guard 1 — a CI test fails when a user-facing doc names a `trellis <command>` or `/trellis:<skill>` that the CLI or plugin lacks — home: cli/docs_consistency_test.go:159-196 (`TestDocsClaimOnlyRealCommands`), with surfaces discovered at :15-33 and :51-80 and the governance corpus under docs/ excluded (verified)
  - Consequence 2 — the CLI's command set has a single source, the `commands` map in main.go, which the doc check reads — home: cli/main.go:53-60 (verified)
  - Context principle — a fix must be a checkable guard, not the maintainer's memory — home: AGENTS.md:18-25 (the iron rule) and AGENTS.md:97 (decision-0028: a guard per derived pair) (verified)
- retired:
  - Guard 2 — auto-release on merge to main (auto-release.yml, release.yml, a patch bump) — by decision-0043 point 4 (verified: neither workflow exists)
  - Guard 1's trigger, "runs whenever cli/, plugins/, install.sh, site/ or README.md change" — broadened by TRL-96 (PR #314): cli-ci has no path filter and runs on every PR (.github/workflows/cli-ci.yml:14-23). No dated note on this record.
  - Guard 1's fixed scan list "the landing, README, and install.sh" — widened in the tree: `docSurfaces` walks every .md, .html, .sh and .mjs doc surface (cli/docs_consistency_test.go:15-33, :51-80)
  - Consequences 1 and 3 — every shippable change auto-publishes, one patch release each — retired with guard 2. The successor rule is that a payload change bumps plugins/trellis/VERSION, guarded by release-guard (AGENTS.md:142-148; .github/workflows/release-guard.yml).
  - Open question 1 (the version scheme) — moot now that auto-bump is gone (decision-0061, SemVer VERSION)
- spent: none
- reason: The guard that docs may name only real commands is live, and broader than decided. The auto-release guard retired with the binary channel in 0043.
- confidence: high
- flags: The header note (0025:14-18) says guard 1 "stands, unchanged", but both its scope and its trigger have changed since. Open question 2 still holds: the check covers names, not described behaviour. Later tests pin specific pairs (`TestReadmesQuoteTheNewRulesFile` :515-521, `TestMarketplaceAddNamesTheRepoThatServesIt` :529-551, `TestInstallTerminalIsConsistentAcrossSourceRenderAndScript` :295), with no general behaviour-claim check.

---

## Tally

**By disposition (25 records)**

| Disposition | Count | Records |
|---|---|---|
| in-force | 1 | 0003 |
| mixed | 22 | 0001, 0002, 0004, 0005, 0007, 0008, 0009, 0010, 0012, 0013, 0014, 0015, 0016, 0017, 0018, 0019, 0020, 0021, 0022, 0023, 0024, 0025 |
| spent | 1 | 0006 |
| repealed | 1 | 0011 |

**By cluster**

| Cluster | Count | Records |
|---|---|---|
| rule-model | 9 | 0002, 0008, 0013, 0014, 0016, 0018, 0020, 0021, 0024 |
| positioning | 7 | 0001, 0003, 0004, 0006, 0009, 0017, 0019 |
| method | 4 | 0005, 0011, 0015, 0022 |
| checks | 3 | 0007, 0023, 0025 |
| delivery | 2 | 0010, 0012 |

## Medium or low confidence calls

All twelve are medium; none is low.

- **0002:** it is unclear whether "support anywhere on the adopt/adapt dial" is still a commitment after TRL-97 retired the postures.
- **0003:** the inspect-or-be-told ingestion was never built. I call it in force only because the invariant set still describes it.
- **0008:** the dials are stated as live, but no consumer can set strictness or gatekeeper any more.
- **0009:** it is unclear whether the unbuilt consent-based feedback loop is a commitment or a stale proposal.
- **0012:** the own-marketplace and activation-level retirements come from the tree and later records, not from a note naming 0012.
- **0015:** issue sweep (c) can't be verified from the tree, and the exemption rules have no home outside the record.
- **0016:** Assess, Apply, cross-instance diffing and Trellis-lite are read as unbuilt, not retired.
- **0017:** the cluster is a judgment call, and README's "Advisor" label may contradict the lexicon.
- **0018:** its unbuilt consequences (SI-1 check, profile channel field, self-declared channel) are listed as in force for the spec author to rule on.
- **0021:** it is unclear whether open question 1 (cross-family review) is done or a standing rule.
- **0023:** I split point 1 against its supersession note: Go survives, the static binary does not.
- **0024:** three of the five decision points have no implementing text, and the delivered "assume yes" default goes beyond Decision 4.

---
id: decision-0066
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. No code changes under this record until it is `approved`.
depends_on: [decision-0061, decision-0063, decision-0064, kodhama/kodhama-0025-retire-the-surface-matrix]
owner: agent
updated: 2026-07-28
---

# 0066 — retire Trellis's surface matrix

## Context

Stewards
[`kodhama-0025`](https://github.com/kodhama/stewards/blob/main/decisions/0025-retire-the-surface-matrix.md)
is `approved` and retires the surface matrix as a **family contract**. It does
not and cannot retire `plugins/trellis/surfaces.json`, which is Trellis's own
artifact authorized by `decision-0061` §3. `decision-0064`'s Consequences
reserved that explicitly: *"A future product decision must define and authorize
any migration from `behavior_state`."* This is that decision.

**What the file is.** 28 lines, two rows — `claude-interactive` and
`codex-cli-local-startup` — each carrying `behavior_state: "supported"`, two
`behavior_contracts` pointing at `decision-0058` and `spec-0007@v1`, an **empty**
`marketplace_test_observations` array, and a `disclosure` string.

**What reads it.** Two things, and neither reads a *value* out of it for any
product purpose:

- `cli/plugin_package_test.go` — validates the file's shape against itself, and
  checks `surfaces.json.version` equals `VERSION`.
- `install.sh:288` — a sha256 line in the bundle manifest. `install.sh` never
  parses the JSON: it takes the file list from `awk '{print $2}' "$stage/manifest"`
  and then runs `shasum -c`. It verifies bytes it never interprets.

No hook, no skill, no workflow, and no CLI command touches it. Verified by grep
across Go, shell, markdown, `.github/workflows/`, `plugins/trellis/hooks/` and
`plugins/trellis/skills/` for `surfaces.json`, `behavior_state`,
`behavior_contracts`, `surface_id`, `disclosure`, and
`marketplace_test_observations`.

**Why the claim is better made elsewhere, and where it actually is.**
`decision-0063` §Preview already names the co-authority, at `:73-78`:

> The existing `codex-cli-local-startup` behavior claim remains supported
> exactly as recorded by decision 0058, spec-0007, and
> `plugins/trellis/surfaces.json`. … the product-owned package **README** and
> `surfaces.json` remain authoritative for it.

So the README is already declared authoritative alongside the matrix by an
approved decision two days old, and `0063:74` carries the string
`codex-cli-local-startup` in its own prose. The matrix's only unique payload —
the literal surface identifier — survives its deletion inside the very decision
that cites it.

**The honest counter-evidence, stated because it is real.** The README does
**not** carry everything the matrix carries. Both matrix rows disclose *"marketplace
registration for 0.2.0 is not yet evidenced"*; `grep -n "evidenc" plugins/trellis/README.md`
returns nothing. And Trellis is now genuinely catalog-listed — in `kodhama/stewards`
it appears in **both** `.claude-plugin/marketplace.json` and
`.agents/plugins/marketplace.json`, the latter with `decision-0063`'s preview
copy telling readers to *"consult Trellis product documentation for exact host
and surface boundaries."* That documentation must therefore say more than it says
today. This decision requires the gap to be closed in the same change, not waived.

**A drifted hand-maintained row is the argument, not an aside.** Both rows say
marketplace registration is unevidenced; the package is now listed in two
catalogs. That is exactly `kodhama-0025`'s case: a hand-maintained row answers
*"does this work on my host?"* worse than a check that ran this morning.

## Decision

### 1. `plugins/trellis/surfaces.json` is retired

The file, its `install.sh` bundle-manifest entry, and the matrix-specific half of
its Go contract tests are removed. Trellis carries no exact surface rows, no
`behavior_state`, and no marketplace-observation record.

### 2. The shipped README carries the claim, under a test

`plugins/trellis/README.md` gains — or has pinned — one paragraph naming the
hosts Trellis is known to work on, **which check establishes that**, and **that
support is not claimed**, per `kodhama-0025` §2. The first two are largely
present at `:28-36` (*"Measured against a real session with file tools disabled"*
… *"Not verified on any surface: `compact`, `clear`, `fork`…"*); the third is
absent from the file entirely, as is the marketplace hedge.

**This paragraph is unguarded today.** The repo's only doc-consistency test is
`cli/docs_consistency_test.go:35 TestDocsClaimOnlyRealCommands`, whose subject
list at `:15-20` is `../README.md`, `../docs/index.html`, `../docs/invariants.html`,
`../install.sh` — **`plugins/trellis/README.md` is not in it** — and whose only
assertions are that docs name no nonexistent `trellis <cmd>` or `/trellis:<skill>`.
It would not fail if the entire host-support section were deleted. The assertion
is a **deliverable of this decision**, not an existing guarantee.

### 3. Atomicity is a correctness requirement, not a preference

`install.sh` fetches each manifest-listed path from a moving `main` and then
verifies checksums. Removing the file while the manifest still lists it makes
`curl -fsSL` 404 and `fail` for every user, immediately, before checksums are
even reached. The reverse — dropping the manifest line while the file remains —
fails `install_script_test.go:256` plus seven vendor integration tests, because
`assertBundleVendored` (`:195-199`) walks the bundle and compares file sets.

Both directions fail closed, which is why one commit is the requirement rather
than the tidy option. The manifest also hashes **every** bundle file, so editing
`plugins/trellis/README.md` in this change makes its manifest line at
`install.sh:265` stale — `install_script_test.go:260` reports exactly that. The
commit therefore removes `install.sh:288` **and** re-hashes `install.sh:265`.

### 4. The two JSON-hardening subtests are re-homed, not deleted

`TestPackageValidatorsRejectMalformedMetadata` (`cli/plugin_package_test.go:331`)
has six subtests. Three are matrix-specific (`:353`, `:377`, `:401`) and go. Two
are **not**: `:408` *"raw JSON must be valid UTF-8"* and `:415` *"JSON escapes
require valid surrogate pairs"*. They merely *use* the observation fixture; what
they exercise is `validateJSONUnicodeEscapes` (`:136`) and the UTF-8 check inside
`decodeJSON` (`:90`) — and `decodeJSON` stays live for **both host manifests** at
`:443`. They are re-pointed at a plugin-manifest fixture. `TestPluginPackageParity`
(`:429`) likewise survives; only its `surfaces.json` block at `:458-491` goes.

### 5. `VERSION` stays `0.2.0`, and that is a judgment

`decision-0061:200-201`: *"A package change now requires an intentional
Trellis-local SemVer update; this decision does not automate or centrally dictate
that judgment."* Removing a file from the shipped bundle and editing the shipped
README **is** a package change, so two different byte-sets would ship as `0.2.0`.
The judgment made here: no bump, because nothing a consumer can invoke changes —
no skill, hook, command, or rule payload is touched, and the removed file was
never read at runtime. Named rather than left silent, so it can be overruled.

## Supersession

**`decision-0061` §3 and §4 — superseded in part.** §3 (`:97-150`) authorized
`surfaces.json` and its closed row shape. §4 (`:150-160`) lists what the parity
guard validates, including *"`surfaces.json.version` equality and its closed row
shape"* and *"every present marketplace observation's closed structure and row
match"* — those two bullets go. **§1, §2 and §5 stand unchanged**: independent
package SemVer, the two host manifests, and their parity are untouched.
`decisions/0061-independent-dual-host-plugin-package.md` gains
`superseded_in_part_by: [decision-0066]`.

**`decision-0063` — superseded in part**, at `:73-78` only, and narrowly: the
`codex-cli-local-startup` behavior claim **remains supported**; what changes is
that the README alone is authoritative for it, not the README *and*
`surfaces.json`. `0063` already names the README as co-authoritative, so this
removes a co-author rather than an authority. Nothing about preview adoption,
the catalog copy, or the rollback path changes.
`decisions/0063-permit-codex-preview-adoption.md` gains
`superseded_in_part_by: [decision-0066]`.

**`spec-0005` — amended, at two sites.** `:84` is the bundle-table row for
`surfaces.json`. `:82` is the **`VERSION`** row — *"both host manifests **and
`surfaces.json`** must match it"* — which is a version-parity claim, not a
bundle-membership one, and would be missed by a scope worded around the bundle.
Both go. The table is normative: `:88-90` says *"AC1 depends on this table's left
column being exhaustive for the actual `plugins/trellis/` tree at build time"*,
so a stale row is a live spec defect. `specs/0005-curl-install-mechanical-vendoring.md`
gains `superseded_in_part_by: [decision-0066]`.

**`decision-0059` and `spec-0008` need nothing.** Both are `status: superseded`
with `superseded_by: [decision-0060]`, and `decision-0061:185-187` already
declares *"Superseded `decision-0059` and `spec-0008` remain historical and
authorize nothing."* Their many `surfaces.json` references are inert archive.

**`decision-0064` already carries its forward pointer to `kodhama-0025`**, landed
separately as pure graph maintenance under that approved record's own acceptance
criterion. This decision adds nothing to it.

**Full sweep result.** Greps for `surfaces.json`, `behavior_state`, `surface_id`
and `marketplace_test_observations` across all of `decisions/` and `specs/` hit
only `0059`, `0061`, `0063`, `0064`, `spec-0005`, `spec-0008`, and this record.
The list above is exhaustive.

## Consequences

**A machine-readable statement of where Trellis works goes away.** Nothing
consumed it and no consumer asked for one. It is a capability withdrawn before
use, not taken away. If one is ever needed, the honest version is generated from
checks that ran.

**The claim becomes user-visible and enforced instead of structured and
unenforced.** It ships inside the package, so it survives outside the
marketplace, and — unlike today — a test fails when it drifts.

**Two catalogs now point at documentation that must be complete.** Trellis is
listed on both hosts in `kodhama/stewards`, with copy deferring to Trellis's own
docs for host and surface boundaries. After this change the README is the only
place that deferral can land.

**Trellis diverges from Grove, deliberately.** Grove keeps its fields because it
has a runtime consumer — a lifecycle gate deciding whether to write into a
consumer repository. Trellis has none. `kodhama-0025` §1 permits exactly this
asymmetry; a future reader should not read Grove's copy as a family standard.

**The install bundle shrinks by one file.** Consumers who vendored via
`install.sh` before this change keep a `surfaces.json` nothing will update. It is
inert, and `/trellis:remove` still removes the overlay.

## The risk this decision accepts

**Meeting every mechanical criterion can still ship a false statement.** A full
simulation of the coordinated deletion — file removed, manifest line dropped, Go
surface machinery excised, README untouched — produced `go vet` clean and
`go test ./...` `ok`, with all eight vendor integration tests green, **while
`plugins/trellis/README.md:9-11` still contained a live relative markdown link
into the deleted file**:

> `[`VERSION`](VERSION) is the sole plugin-package SemVer authority. Both host
> manifests and [`surfaces.json`](surfaces.json) carry that same value, guarded
> by the repository's `TestPluginPackageParity` Go test.`

That sentence ships to every consumer's disk. Because of the gap in §2, nothing
in the suite catches it. AC4 and AC5 exist for that reason and are the two most
load-bearing criteria here.

## Acceptance criteria

**AC1 — atomic removal.** `plugins/trellis/surfaces.json`, the `install.sh:288`
manifest entry, and the matrix-specific Go machinery are removed **in one
commit**, and that commit re-hashes `install.sh:265` for the edited
`plugins/trellis/README.md`. `install_script_test.go` is green; no intermediate
commit exists in which the manifest and the tree disagree.

**AC2 — surviving coverage.** `cli/plugin_package_test.go:408` and `:415` are
re-pointed at a plugin-manifest fixture and still fail on invalid UTF-8 and on a
lone surrogate. Verified by mutation, not by assertion:
`validateJSONUnicodeEscapes` and `decodeJSON`'s UTF-8 check each still have a
test that goes red when broken.

**AC3 — parity intact.** `TestPluginPackageParity` still enforces `VERSION` as
canonical SemVer and exact version equality across both host manifests.

**AC4 — the README states all three things.** One paragraph in
`plugins/trellis/README.md` names the hosts Trellis is known to work on, which
check establishes that, and **that support is not claimed** — and carries the
marketplace hedge the retired rows carried, or states plainly why it no longer
applies.

**AC5 — that paragraph is guarded, and `:9-11` is corrected.** A new assertion
covers `plugins/trellis/README.md` (the existing `docSurfaces` list does not).
The `surfaces.json` clause and its relative link at `:9-11` are gone. Both
verified by mutation: deleting the support sentence, and restoring the link, each
turn the suite red.

**AC6 — the reintroduction guard is two-part.** A single content grep is **not**
implementable: `surfaces.json` appears today in `decisions/0059`, `0061`, `0063`,
`specs/0005` and `specs/0008` — the append-only archive, which must keep the
name. Such a guard is red on day one and degenerates into an exclusion list. The
implementable shape, both halves verified to return empty against the
post-deletion tree:

1. a repo walk excluding `.git` that fails on **any file whose basename is
   `surfaces.json`** (case-insensitive); and
2. a **code-scoped** content assertion over `cli/`, `plugins/` and `install.sh`
   only, for `behavior_state|marketplace_test_observations|surface_id`.

**AC7 — the graph is closed.** `decision-0061`, `decision-0063` and `specs/0005`
each carry `superseded_in_part_by: [decision-0066]` with a scope note, and their
prose forward pointers name which clauses moved.

**AC8 — suites green.** `go vet ./...` clean and `go test ./...` `ok`, including
all eight vendor integration tests.

## Open questions

- None blocking. `VERSION` staying at `0.2.0` (§5) is a judgment offered for
  overrule, not an open question.

## Lifecycle record

Drafted 2026-07-28 under `kodhama-0025`, which the maintainer approved
2026-07-27 with an acceptance criterion naming this repository's atomicity
requirement directly.

**This is the second draft.** The first was reviewed by an independent
`decision-adversary` and returned **NEEDS-REVISION** with twelve defects, three
blocking. The blocking ones were: it claimed the README paragraph was *"already
under a doc-consistency test"* when no test covers that file at all (disproved by
simulation — the full deletion passes `go test ./...` with the README's dead link
intact); it never named `plugins/trellis/README.md:9-11`, so the change would
have shipped a live link into a deleted file; and it described the `install.sh`
change as one manifest entry when it is one removal plus one re-hash. It also
omitted `## Decision` and `## Consequences`, which
`core/rubrics/artifact-contract.md:53-57` requires of every `decision`; omitted
the `superseded_in_part_by` mechanism on all three targets; scoped the `spec-0005`
amendment to one of its two sites; and specified a reintroduction guard that
would have been red on its first run against this repo's own archive.

The first draft was also destroyed before review completed — staged, never
committed, and lost to a concurrent `git reset` in the same clone. Recorded
because it is why the artifact under review no longer matched the artifact on
disk, not because it changed the analysis.

**`gated`.** No code changes under this record until the maintainer's intent act
moves it to `approved`.

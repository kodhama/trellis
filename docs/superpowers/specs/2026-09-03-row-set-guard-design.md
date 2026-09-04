# Row-set guard — design

**Ticket:** TRL-28 · **Decision:** recorded in the ticket's blockquote (maintainer, 2026-09-03):
*shrink the surface first, then guard the rest.* This spec records the mechanism, which the
decision left to the implementing session.

## Problem

Minting or retiring an invariant slug obliges edits across many surfaces, and the only thing
enforcing the sweep is a prose note in the catalog (`core/catalog/signature-catalog-v1.md:55-67`)
that has been found short twice (`decision-0074`, `decision-0078`). The consumer-side symptoms
were fixed by reconciliation (`decision-0083`, `decision-0084`); this is the producer-side cause.

## What was measured (2026-09-03, on `main` at `3362bd6`)

**Prose count sites — 22 sites in 12 files.** Every one states the current row count as a fact
about the system today, in one of four shapes: digits (`16 rules`), spelled out (`sixteen rows`),
`N/N` (`16/16`), or a class breakdown (`the four structural, the ten remaining operating, the two
floors`).

| File | Sites | Shapes |
|---|---|---|
| `README.md` | 25, 44, 70 | spelled ×2, `N/N` |
| `plugins/trellis/README.md` | 116, 148 | spelled ×2 |
| `install.sh` | 765 (`say` — runtime output), 783 (comment) | spelled ×2 |
| `plugins/trellis/hooks/staleness.sh` | 543 (`emit` — runtime output to the agent) | digits |
| `docs/index.html` | 446, 584, 596 | digits, spelled ×2 |
| `docs/lp-content.md` | 50, 178, 196 | digits, spelled ×2 |
| `docs/invariants.html` | 7 (meta), 167 (h1) | spelled ×2 |
| `profiles/trellis-self.md` | 45 | digits |
| `core/catalog/signature-catalog-v1.md` | 37 (Coverage), 444 (AC1 + class breakdown) | digits, breakdown |
| `core/rubrics/artifact-contract.md` | 81 | spelled |
| `.claude/agents/corpus-reviewer.md` | 75 | spelled |
| `plugins/trellis/skills/remove/SKILL.md` | 44 | `N/N` (already derived, `remove_skill_test.go:179`) |

Both prior counts were wrong in the same direction. The ticket said "six files"; the controlling
session's own sweep found two sites. **The undercount was a grep artifact:** on this platform
`git grep -E '\bsixteen\b'` matches nothing (`\b` is unsupported in git grep's ERE here — verified;
`git grep -w sixteen` finds four hits in the two READMEs alone). The first pass of this session's
sweep then lost two more sites to line-level exclusion filters (`decision-00`, `px`). A sweep is not
a guard; that is the finding, and it is why the mechanism below is a precise per-site table rather
than a pattern.

Excluded, with reason: `install.sh:78` "14 rendered M1 payload" is a *file* count, a different
quantity; the ~20 code comments in `staleness.sh` / `codex-context.mjs` are maintainer comments,
several deliberately historical ("the hook reported `quarantined 14 row(s)`"); `eval/` scores are
frozen runs; `decisions/`, `research/`, `docs/superpowers/` are records.

**Structural surfaces — most were already derived or guarded.**

| Surface | State on `main` |
|---|---|
| `reference/rules-a.toml`, `rules-b.toml` | **Rendered** from `catalogSlugOrder()` (`cli/apply.go:341`); pinned forward by `TestPayloadRulesTomlSeeds`, byte-pinned by `TestVendoredPayloadIsCurrent`. Not independent. |
| catalog entries ↔ pin | `rules_test.go` — count + every slug present. |
| `.trellis/rules.toml` | Forward only (`TestRepoDeclaresRulesConfig`) — a stale row after a retire is not caught. |
| `docs/invariants.html` | Forward only, via examples (`TestInvariantsPageMatchesCatalog`) — a stale card after a retire is not caught. |
| `core/invariants/trellis-invariants-v1.md` | **No guard.** |
| `profiles/trellis-self.md` | **No guard.** |
| `plugins/trellis/hooks/*` | Derive from `reference/rules.md` at runtime (`decision-0083/0084`). No surface. |
| `plugins/trellis/VERSION` | PR #245. Out of scope here. |

So the unguarded structural remainder is smaller than the ticket said, and the prose half is
larger than anyone measured. The decision stands as written; the weight within it moves.

## Mechanism

**One new test file, `cli/row_set_guard_test.go`, two tests. Both read `assessableSlugs` /
`len(assessableSlugs)` (`cli/payload_test.go`, "the ONE pin for the row set") as the single
source.** Nothing else in this change is generated or rewritten; the tests *read* the surfaces.

### 1. `TestRowCountProseSitesFollowThePin` — the shrink

A table with one row per site above: `{path, template}` where the template encodes *that site's
shape* — `"all %s rows active"` with the count spelled out, `"governed at %d/%d"`, `"all %d rules
active"`, and for the AC1 breakdown the class counts derived from the catalog's `- class:` lines.
For each row the test renders the template from the pin and asserts the phrase is present,
failing **per site** with the path and the exact expected text. `install.sh:783`'s "two rules
out of sixteen" is pinned as `"out of %s"` spelled out.

A bounded negative sweep runs over the pure-doc files (both READMEs, `docs/index.html`,
`docs/lp-content.md`, `docs/invariants.html`, the profile, the contract, the corpus-reviewer
charter, the catalog): any number from ten to twenty-nine — digits or spelled out — followed
within a short window by `rules|rows|invariants|genes|slugs`, and any `N/N` with `N ≠ count`,
fails unless the number is the pin's. `install.sh` and `staleness.sh` are excluded from the
negative sweep (they carry legitimate historical comments) and get the positive check only.

Why this and not a generated fragment: there is no templating pipeline for `README.md`,
`docs/*.html`, `install.sh` or a hook's `emit` string, and building one would edit twelve files
including hooks and `install.sh`, which concurrent PRs own. The test is the least invasive
checkable mechanism, and it follows the repo's existing exemplar (`remove_skill_test.go:179`,
`decision-0078`). It does not make the edit automatic; it makes the *omission* impossible, which
is the ticket's "done when".

### 2. `TestRowSetDerivativesFollowThePin` — the guard

Set equality — missing **and** extra — against the pin, one sub-check per derivative, each
naming the slugs that differ:

| Derivative | Extraction |
|---|---|
| catalog entries | `catalogSlugOrder()` |
| `core/invariants/trellis-invariants-v1.md` | `` ^- **`((inv|floor)-[a-z-]+)` — `` — live entries; the collapsed `inv-reference-relationship` heads with `→` and is excluded by shape, dials by prefix |
| `profiles/trellis-self.md` | `` ^\| `([a-z][a-z-]*)` \| (true|false) \| `` |
| `docs/invariants.html` | `<span class="code">([a-z-]+)</span>` |
| `.trellis/rules.toml` | `^([a-z][a-z-]*)\s*= \{` |
| `reference/rules-a.toml`, `rules-b.toml` | same row regex, on the vendored files |

The vendored TOMLs and the catalog are already pinned elsewhere; they are included so that this
one test's failure output is the *complete* map of what has not followed. The now-redundant
forward loop in `TestRepoDeclaresRulesConfig` (`cli/selfapply_test.go`) retires with a pointer;
its strictness check stays.

### 3. The catalog note points at the guard

`signature-catalog-v1.md:55-67` is rewritten: a row-set change is guarded by
`cli/row_set_guard_test.go`, whose failures name every derivative and every prose site; the two
obligations the guard cannot check are named — a new `docs/invariants.html` *card* needs its
examples (the set guard catches the missing code; `TestInvariantsPageMatchesCatalog` catches the
missing examples) and `plugins/trellis/VERSION` (#245). The `SLUGS` retirement parenthetical
collapses to a pointer at `decision-0083`. `cli/assets/invariants.md` regenerates (`go generate`).
The note is preamble, so `reference/invariants.md` and `reference/version` do not move; that is
verified, not assumed.

## Proofs required before the PR is ready

- Both tests pass on `main` as it stands.
- **Mutation, structural:** add `inv-zzz-fake` to `assessableSlugs` only → the set test names
  every derivative; the prose test names every site (16 → 17). Remove a real slug from
  `rules-a.toml` only → red, naming `rules-a.toml`. Restore.
- **Mutation, prose:** change `README.md:70` `16/16` → `15/15` → red naming `README.md` (positive
  check) and the `15/15` (negative sweep). Restore.
- **Not over-constrained:** edit an `- *(honored)*` example line in the catalog → both new tests
  still pass (`-run TestRowSet`). `decision-0081` prefers extending an entry to minting a slug;
  that path stays frictionless.
- `git diff --stat -- plugins/trellis/reference/` shows nothing after the catalog edit.

## Out of scope

- Editing any of the 22 prose sites; editing `install.sh` (its bundle hashes included),
  `plugins/trellis/hooks/*`, `VERSION` or `plugin.json` at all (owned by #261/#262/#263).
- A generator for `docs/*.html` or the READMEs.
- Hook code comments that narrate historical counts.
- A decision record. This is `decision-0028` applied ("a guard per pair"), with `decision-0078`'s
  derived needle as the exemplar, and it changes no maintenance workflow — same ground PR #245
  stood on. Flagged on the ticket; the maintainer can ask for `0089`.

---

## 2026-09-04 — reviewed and shrunk (record, not a rewrite)

This spec's design shipped as written and was then cut. Keeping it as recorded (`decision-0085`
§5: these are records, not contracts) so the correction is legible:

The design above measured 22 prose count sites and proposed pinning all 22. Review (a local
Codex pass, then the maintainer: *"I would say shrink the guard yeah"*) found the table was the
same mistake one layer down — a list of surfaces, in a test file instead of a note. The ruling
applied the ticket's own order — *shrink the surface first, then guard the rest* — to the
remainder this spec had chosen to guard.

What changed:

- **The numeral was deleted wherever the sentence still means what it meant** ("all sixteen rules"
  → "all rules"). That took the prose surface from 22 sites to 4 — plus 2 in
  `plugins/trellis/README.md` deferred, because deleting them re-bakes `install.sh`'s bundle
  manifest, off-limits while #262 is open.
- **The table survives only where the number is the claim**: `install.sh`'s closing `say`,
  `staleness.sh`'s `emit`, `README.md`'s `16/16`, the catalog's AC1 class breakdown. The
  whole-file exclusion of `install.sh` and the hooks was replaced by positive checks on exactly
  their runtime strings.
- **The bounded negative sweep was dropped.** It was wrong in both directions: it would fail
  legitimate future prose (a subset count, a date, a version ratio) and still miss noun-first
  shapes ("the invariant count is 17").
- **A value check was added** where membership was not the contract: `profiles/trellis-self.md`
  claims every gene is active, and the row pattern accepted `| false |`, so a flip passed both
  tests. It now matches `| true |` only.

---
id: decision-0054
type: decision
status: approved  # maintainer's intent act 2026-07-21, in-conversation ("Approved, carry on.") — this flip records it (decision-0046); pre-gate: corpus-reviewer FAIL->fix->PASS (one confirmed inaccuracy on the decision-0053 rows-section characterization, corrected before the gate)
depends_on: [decision-0043, decision-0028]
informed_by: [decision-0044, decision-0053]
owner: agent
date: 2026-07-21
---

> **Provenance.** Found mid-rollout (2026-07-20/21): merging decision-0051–0053's
> live-rows refresh into grove and math-quest (both grove-managed, both running
> grove's `review-bookkeeping` CI) turned up a real, reproducible finding —
> `.trellis/internal/invariants.md`'s vendored frontmatter (`id`, `type`,
> `depends_on`, `owner`, `status`, `ratified`) causes two distinct CI failures in
> any grove-managed consumer, because trellis's own artifact-contract metadata is
> byte-copied into a place a third party's tooling reads it as claims about
> *its own* corpus. A first proposed fix (qualify the ids with the `repo/id` form,
> `decision-0044`) was drafted, then withdrawn on re-reading `0044`'s actual text —
> that mechanism is for one repo citing *another* repo's artifact, not a repo
> self-qualifying its own local ids; applying it to `core/catalog/signature-catalog-v1.md`
> would have been a misuse of the convention to paper over a symptom, and risked
> trellis's own internal graph tooling. The maintainer's question — "does the
> payload need this frontmatter at all?" — found the real fix.

# 0054 — the vendored `invariants.md` ships without frontmatter; the source keeps it

## Context

- **The two failures, precisely** (`grove/specs/0002-review-bookkeeping-check.md`,
  read directly, not from memory): (1) `invariants.md` carries `type:
  signature-catalog` frontmatter, which grove's bookkeeping check classifies as
  a real, reviewable artifact type — unclaimed in a consumer's own policy, it
  "owes the full [review] set" (§C.2 step 1; §B). (2) Independently, and
  regardless of review status ("for every changed artifact of any type —
  whether or not it owes any review"), §C.7's graph resolution walks the
  file's `depends_on: [invariants-v1, spec-0002]` and finds both ids
  unresolvable in the consumer's own corpus — they are trellis-internal,
  bare, and meaningless outside trellis's own repo.
- **Why qualifying the ids (the first idea) was wrong.** `decision-0044`'s
  qualified `<repo>/<id>` form is explicitly for cross-repo citation — "the
  referenced artifact's own id exactly as declared in **its home corpus**"
  (0044's own text). Within `core/catalog/signature-catalog-v1.md` itself,
  `invariants-v1` and `spec-0002` **are** trellis's own local artifacts —
  bare is correct there, and is what trellis's own sync-guards
  (`decision-0028`) expect. Qualifying them in the source would have answered
  a question decision-0044 was never asked, and risked trellis's own internal
  graph resolution to fix a *consumer's* misreading of a *vendored copy*.
- **Why the frontmatter has no consumer value at all.** `id`, `status`,
  `owner`, `ratified`, `depends_on`, `scope` — every field describes
  `invariants.md`'s life *inside trellis's own repo*. Nothing downstream of
  the vendored copy reads any of them (verified, not assumed): `docs/invariants.html`
  regenerates from `core/catalog/signature-catalog-v1.md` directly, never the
  payload copy (`cli/sync_test.go` `TestInvariantsPageMatchesCatalog`); the
  eval scorecard generator reads the same source file directly
  (`eval/experiments/does-trellis-help/gen-invariant-scorecard.py`); nothing
  in `cli/` parses the *payload* copy's frontmatter fields at all — the
  generator uses the embedded catalog text (`invariantsRef`) only to extract
  each rule's prose (why/examples), never its own frontmatter. This is the
  same class of dead weight `decision-0052`'s session already eliminated from
  six consumer repos' `expression.md` files — trellis's own internal
  bookkeeping, permanently loaded into every session, for zero operational
  value to the reader.
- **Why this is not a new architectural risk.** `rules.md` is *already*
  generator-composed, not a byte-identical copy (`decision-0053`:
  `renderRulesReadout()` builds it from the authority header + a loop of
  per-rule fragments — the rows themselves are a *separate* mechanism, not
  part of `rules.md`: a block-level `@.trellis/rules.toml` import on the
  import channel, folded into the distinct `block-inline-<p>.md` on the
  inline channel). The "mechanical copier, never
  re-derive" principle (`kodhama-0007`) governs the **install-time skill**
  (never hand-authoring or paraphrasing at install), not the **release-time
  generator**, which is explicitly the one deterministic place composition
  happens (`decision-0043`, "generator-only CLI"). Stripping frontmatter at
  write time is the same class of operation already proven safe for
  `rules.md`.

## Decision

**1. The payload's `invariants.md` ships without frontmatter.** The
generator's write step for the payload file changes from writing the
embedded catalog text (`invariantsRef`) verbatim to writing it with the
leading `---\n...\n---\n` block stripped — one new function, applied only
at the payload-write site (`cli/payload.go`'s `invariants.md` entry). The
embedded `invariantsRef` variable itself, and every other place that reads
it (rule-fragment extraction, `apply.go`), is **untouched** — the strip is
a display-time transform on the final write, not a change to what the
generator has access to.

**2. The source keeps its frontmatter, unconditionally.**
`core/catalog/signature-catalog-v1.md` is unaffected — trellis's own
`depends_on`/sync-guard machinery (`decision-0028`) needs it, and
`docs/invariants.html` / the eval scorecard both already read the source
directly, not the payload copy.

**3. `TestBundledCatalogInSync`'s assertion changes** from three-way byte
identity (`assets/invariants.md` == `plugins/trellis/reference/invariants.md`
== catalog source) to: `assets/invariants.md` stays byte-identical to the
catalog source (unchanged — the `//go:generate cp` step is untouched); the
**payload** copy is asserted equal to the source **with frontmatter
stripped** — a new, precise, still-mechanical check, not a loosened one.

**4. This is our synthesis, not a claim on grove's behalf** (naming
guardrail, same posture as every cross-family observation this session):
trellis is not asserting what grove's bookkeeping check *should* accept —
only that trellis's own vendored file carries content with no reason to
exist outside trellis's own repo, and removing it is correct on trellis's
own terms regardless of what any consumer's CI does with the result.
Whether this alone is sufficient for a consumer's own review-bookkeeping
policy (the `non_behavioral_allowlist` path, `spec-0002` §C.2 step 2) is
that consumer's own config to set — not built or asserted here.

## Consequences

- **Derived chain regenerates in the same change** (`decision-0028`):
  the new strip function (`cli/apply.go`, called from `cli/payload.go`'s
  write site); `plugins/trellis/reference/invariants.md`
  (frontmatter gone); `checksums`; `version` stamp; `install.sh` manifest;
  this repo's own `.trellis/internal/invariants.md` (self-application,
  `decision-0035`). `docs/invariants.html` and the eval scorecard are
  **unaffected** — confirmed above, both read the source, not the payload.
- **Family repos pick this up on their next refresh** — no urgency; the six
  repos just refreshed (grove, kodhama, wisp, design-system, stewards,
  math-quest) carry the frontmatter-bearing copy until their next
  `/trellis:setup` run. Two of those six (grove, math-quest) merged with a
  known-red `review-bookkeeping` check specifically because of this issue
  (maintainer's own call, both repos, 2026-07-20/21) — this decision doesn't
  retroactively fix already-merged state; the next refresh does.
- **Consumer-side CI policy work (the originally-scoped "fix B" —
  `non_behavioral_allowlist` entries for the three orientation-prose payload
  files, plus whatever `.trellis/rules.toml`/`.trellis/internal/version`
  still need) is simplified but not eliminated by this decision**: once
  `invariants.md` carries no frontmatter, it becomes plain prose eligible for
  the same simple allowlist path `rules.md`/`trellis.md`/`CLAUDE.md` already
  need — no `reviewless_types` declaration required at all. The `.toml`/
  extensionless-file gap (`.trellis/rules.toml`, `.trellis/internal/version`)
  is **unaffected by this decision** and remains open — `spec-0002`'s prose
  predicate excludes them by design regardless of frontmatter. That gap
  stays a disclosed, accepted limitation, not solved here.

## Open questions

- **Should the strip generalize?** Today `invariants.md` is the only
  vendored file carrying artifact-contract frontmatter (verified: `rules.md`,
  `trellis-<p>.md`, `rules-<p>.toml`, `block-claude.md` carry none). If a
  future payload file is ever rendered from a raw artifact-corpus file
  again, revisit whether the same strip applies by default rather than
  per-file.
- **The `.toml`/extensionless gap** (`.trellis/rules.toml`,
  `.trellis/internal/version`) has no fix under grove's current
  `spec-0002` — named here as the standing open item this decision does
  *not* close; real upstream work for grove, not trellis's to solve.

## Self-check (gate)

Corrects a withdrawn first attempt in the open, with the reasoning for the
withdrawal recorded (`floor-transparency`; `inv-auditable-archive`'s
why-is-it-this-way) — the wrong idea is not silently replaced. Verified,
not assumed, that nothing downstream reads the payload copy's frontmatter
before proposing removal (`inv-independent-judgment`). Reuses the
generator-composition precedent already proven by `decision-0053` rather
than inventing new machinery (`inv-minimal-first`). Source (`core/catalog/`)
and payload (`plugins/trellis/reference/`) stay cleanly separated — the
fix touches only the render boundary, never trellis's own artifact-contract
truth. `depends_on` carries genuine coupling (the generator being changed,
`decision-0028`'s derived-sync obligation); `decision-0044`/`0053` are
context for the withdrawn alternative and the generator-composition
precedent (`informed_by`, `decision-0047`). Left at `draft`: the author
does not grade its own decision — the gate and the `approved` flip are the
maintainer's intent act (`decision-0046`), ideally after an independent
pass (`inv-independent-judgment`).

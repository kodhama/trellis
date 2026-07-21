---
id: decision-0055
type: decision
status: draft
depends_on: [decision-0054, decision-0028]
informed_by: [decision-0043]
owner: agent
date: 2026-07-21
---

> **Provenance.** Follow-on to `decision-0054`, same day: the maintainer's own
> observation — "invariants.md ... has a bunch of governance prose about which
> decision originated what that is not useful in a consumer context" — pointed
> at content `0054`'s frontmatter-strip left untouched. Sizing the catalog
> before proposing anything (`core/catalog/signature-catalog-v1.md`, 383
> lines) found the governance prose splits into three tiers, not one: a
> heading-bounded preamble (lines 1–52) and a heading-bounded tail (lines
> 356–383) — both mechanically excisable, the exact same class of transform
> `0054` already proved safe — and seven lines of decision citations woven
> *inside* individual invariant entries (53–355), which are not block-
> removable and need real judgment. The maintainer chose to ship the
> mechanical tier now and give the seven-line judgment-call tier its own
> decision, later, done properly (parked as `decision-0056` — not drafted
> here).

# 0055 — the vendored `invariants.md` carries only the entries body; the catalog's preamble and tail stay source-only

## Context

- **The preamble** (`core/catalog/signature-catalog-v1.md` lines 1–52, after
  `decision-0054`'s frontmatter strip removes lines 1–9): a ratification
  blockquote ("Ratified via merge (`decision-0022`)... Amended in place
  2026-07-12 (kodhama-0008 Lane A...)") and four meta-notes ("What this is,"
  "Coverage," "On `mechanizable`," "Derived resources — sync them on any
  change") — all addressed to whoever *maintains* the catalog, none of it
  needed to *apply* an invariant.
- **The tail** (lines 356–383): `## Acceptance criteria` (the catalog's own
  self-check rubric — "every entry carries `what` · `directive` · ...") and
  `## Open questions` (unresolved design items — "Owed to the Assess build"
  is verbatim; another item reads "Fold in when the ingestion check
  (`decision-0003`) is built"). Both are the catalog's own governance
  apparatus, not invariant content.
- **The entries body** (lines 53–355, the three `###` subsections — Structural,
  Operating, Floors): the actual 14 invariant descriptions. Verified clean
  almost throughout — the exception is seven lines carrying an inline
  `` `decision-NNNN` `` citation woven into a `signature`/`why` field (four
  distinct decisions: `0018`, `0021`, `0028`, `0040`). That's `decision-0056`'s
  scope, not this one's — small enough to deserve its own careful,
  independently-verified rewrite rather than being rushed into a mechanical
  change.
- **Why this is the same class of transform as `0054`, not a new risk.**
  Both the preamble and the tail are bounded by literal, syntactic markers
  (`## Entries`, `## Acceptance criteria`) — no semantic judgment is needed
  to find the boundary, exactly like frontmatter's `---...---` delimiters.
  The strip composes naturally with `0054`'s: once frontmatter is gone, "keep
  only the entries body" is a single slice between two heading matches,
  subsuming the narrower frontmatter-only cut rather than stacking a second
  regex on top of it.

## Decision

**1. The payload's `invariants.md` carries only the entries body.** The
generator's write step for `invariants.md` changes from stripping frontmatter
alone to extracting the text strictly between the `## Entries` heading
(inclusive) and the `## Acceptance criteria` heading (exclusive) — one
function, replacing `decision-0054`'s narrower `stripFrontmatter` call at
the same site (`cli/payload.go`'s `"invariants.md"` entry), since the wider
slice already excludes the frontmatter that sat before `## Entries` too.

**2. The source is unaffected**, unconditionally — `core/catalog/signature-catalog-v1.md`
keeps its preamble and tail exactly as-is; trellis's own `decision-0028`
sync-guards, the `## Acceptance criteria` self-check, and the `## Open
questions` log all need them. `invariantsRef` (the embedded catalog text
used by the generator for rule-fragment extraction) stays untouched, same
as `decision-0054`'s own constraint.

**3. Nothing here touches the seven inline decision-citation lines.** They
ship in the payload exactly as they read in the source today — parked for
`decision-0056`, a separate, carefully-authored, independently-verified
rewrite, not a build-time transform.

**4. `docs/invariants.html` and the eval scorecard are unaffected**, same
guarantee as `decision-0054` established and the same mechanism confirms it
by (both read `core/catalog/signature-catalog-v1.md` directly, never the
payload copy).

## Consequences

- **Derived chain regenerates in the same change** (`decision-0028`):
  the extraction function (`cli/apply.go`, called from `cli/payload.go`'s
  write site, replacing `decision-0054`'s narrower strip at that one call
  site); `plugins/trellis/reference/invariants.md` (preamble + tail gone,
  ~80 fewer lines); `checksums`; `version` stamp; `install.sh` manifest; this
  repo's own `.trellis/internal/invariants.md` (self-application,
  `decision-0035`).
- **`TestBundledCatalogInSync`'s payload-copy assertion widens again** —
  from `stripFrontmatter(source)` to the new extraction, on the same test,
  same discipline: red observed against the pre-fix payload before the fix
  lands.
- **Build sequencing, named plainly**: this decision's build should land
  *after* `decision-0054`'s build (PR #173) is merged to `main`, not stacked
  on its still-open branch — the same stacked-PR risk this repo's own
  history has already been burned by once. The decision *record* has no
  such dependency (pure content, no code); only the build does.
- **`decision-0056` is the standing follow-up**, explicitly not built here:
  the seven-line inline-citation rewrite, one-time authored and adversarially
  verified against the source, becoming its own checked-in artifact the
  generator copies from thereafter — never a judgment call re-made at build
  time (the `kodhama-0007` "second writer, and second writers drift" risk
  this whole session has protected against consistently).

## Open questions

- **Does `decision-0056` rewrite the catalog's own prose in place, or
  produce a separate consumer-facing rendering?** Genuinely undecided —
  rewriting in place changes what trellis's own internal readers see too;
  a separate rendering keeps the source's citations for trellis's own
  audit trail but adds a second file to keep in sync. Left for `0056`'s own
  shaping.
- **Should future catalog entries avoid inline citations going forward**,
  so this class of problem doesn't recur? Worth a line in `0056` once its
  shape is settled — not decided here.

## Self-check (gate)

Reuses `decision-0054`'s exact mechanism and test discipline rather than
inventing a new one (`inv-minimal-first`); the boundary is syntactic
(heading markers), never semantic, so no LLM judgment enters the build
pipeline (the same "mechanical copier, never a second writer" discipline
`decision-0053`/`0054` already established). The harder, genuinely
judgment-requiring piece is named and explicitly deferred, not silently
folded in or silently dropped (`floor-transparency`). Source and payload
stay cleanly separated — trellis's own sync-guards, self-check, and
open-questions log are untouched. `depends_on` carries genuine coupling
(`decision-0054`, the transform being widened; `decision-0028`, the
derived-resource sync obligation); `decision-0043` is context for the
generator-only-CLI precedent (`informed_by`, `decision-0047`). Left at
`draft`: the author does not grade its own decision — the gate and the
`approved` flip are the maintainer's intent act (`decision-0046`), ideally
after an independent pass (`inv-independent-judgment`).

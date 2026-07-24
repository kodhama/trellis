---
id: decision-0061
type: decision
status: draft
depends_on: [decision-0043, decision-0058, decision-0060]
owner: agent
date: 2026-07-24
---

# 0061 — version the Trellis dual-host plugin independently and carry lean surface metadata

## Decision state

**Decided** (maintainer, 2026-07-24):

- Trellis independently adopts the small package pattern: one SemVer
  `VERSION`, matching Claude and Codex manifests, version-bound product-owned
  surface metadata, and one lightweight local parity check.
- Trellis starts this package line at `0.2.0`; it does not reuse the `0.1.0`
  value previously seen by both host ecosystems.
- Package SemVer and the content-derived `payload@…` install stamp remain
  separate identities.
- Surface metadata records Trellis behavior and marketplace-test observations
  separately; neither is inferred from the other.
- No family release, synchronized bump, tag workflow, release record,
  generated support table, history, approval ledger, or shared validator
  returns.

**Open** (0):

- None.

**Parked** (0):

- None. Broader Codex lifecycle and hosted-surface tests remain the phased
  work already owned by `decision-0058`, not a package-version question.

## Context

Trellis already ships one source tree with native declarations for Claude and
Codex. `decision-0058` and `spec-0007@v1` bound its behavioral claim to
Claude's existing import path plus trusted local Codex setup and fresh
startup. They deliberately leave Codex resume, clear, compact, subagents,
desktop, IDE, headless, automation, and cloud evidence-gated.

The package identity is nevertheless inconsistent. The Claude manifest omits
`version` under historical `decision-0036`, while the Codex manifest says
`0.1.0` without a current product-owned authority. The rendered payload has a
different and correct identity:
`plugins/trellis/reference/version` is `payload@<content-hash>`, copied to the
consumer overlay and compared file-to-file under `decision-0043`.

Trellis briefly adopted a family release-certification system in
`decision-0059`/`spec-0008`, then retired it completely through approved
`decision-0060`. That reset explicitly requires a new project-local decision
if package versioning later becomes useful. The maintainer has now selected a
smaller local pattern: the version carriers and surface metadata are useful;
the retired release machinery is not.

Reusing `0.1.0` would repeat the concrete stale-update problem recorded by
`decision-0036`: Claude previously cached a Trellis manifest at that value,
and the current Codex manifest also carries it. `0.2.0` is therefore the first
coherent dual-host package value. It is a Trellis-local version and has no
relationship to Grove, Stewards, Wisp, or the retired Go binary's version.

## Decision

### 1. One Trellis-local package version

`plugins/trellis/VERSION` is the sole Trellis plugin-package SemVer authority.
It contains exactly `0.2.0` plus one terminal newline for the initial adoption.

Both host manifests:

- declare `name: "trellis"`; and
- carry a `version` exactly equal to `VERSION`.

The host manifests may otherwise differ where their native formats or
interfaces differ. No other product's version is read, compared, coordinated,
or constrained.

This supersedes `decision-0036`'s remaining commit-SHA package-version rule
and `decision-0043` point 4 only where it says plugin versions are commits.
It does not restore a release channel, auto-update mechanism, or tag.

### 2. Package version and payload identity remain separate

`plugins/trellis/reference/version` remains the rendered payload's
content-derived `payload@<12-hex>` identity. Setup continues to copy that
value to `.trellis/internal/version`, and the staleness hooks continue their
file-to-file comparison.

The package SemVer answers “which plugin package is this?” The payload stamp
answers “which generated rule bytes are installed?” Neither derives from or
replaces the other. A consumer-owned `rules.toml` edit changes neither value.

### 3. Lean, product-owned surface metadata

`plugins/trellis/surfaces.json` uses one closed product-local shape:

```json
{
  "schema_version": 1,
  "version": "0.2.0",
  "rows": [
    {
      "surface_id": "codex-cli-local-startup",
      "host": "codex",
      "behavior_state": "supported",
      "behavior_contracts": ["decision-0058", "spec-0007@v1"],
      "marketplace_test_observations": [],
      "disclosure": "Supported only for the trusted local fresh-start boundary; marketplace registration for 0.2.0 is not yet evidenced."
    }
  ]
}
```

The top-level object and every row are closed. `version` equals `VERSION`;
`surface_id` is unique and matches
`^[a-z0-9][a-z0-9._/-]{0,127}$`; `host` is `claude` or `codex`;
`behavior_state` is `supported`, `candidate`, or `unsupported`;
`behavior_contracts` contains artifact identifiers; and `disclosure` is
nonblank.

`marketplace_test_observations` contains normalized repository-relative JSON
paths with no `..` segment. A referenced record must structurally satisfy
Stewards marketplace-observation schema version 1 and match the row's host
and surface identifier.

The initial file contains exactly the two already supported Phase-1
boundaries:

- `claude-interactive`; and
- `codex-cli-local-startup`.

Both point to `decision-0058` and `spec-0007@v1`. Their marketplace observation
arrays start empty and their disclosures say marketplace registration for
`0.2.0` is unverified. No other surface is listed or implied.

Behavior state is Trellis's product claim. A marketplace observation says
only that one exact marketplace checkout was registered in one run. Empty
observations do not withdraw an independently established behavior claim;
present observations do not create one.

### 4. One lightweight local parity guard

The existing Go test suite gains one package-parity test. It validates:

- `VERSION` as canonical SemVer;
- shared `trellis` identity and exact version equality across both manifests;
- `surfaces.json.version` equality and its closed row shape;
- unique, valid surface identifiers and allowed states;
- every present marketplace observation's closed structure and row match; and
- existence of paths declared by each host manifest.

The guard reads no other product version and decides no bump, tag, release,
catalog admission, marketplace availability, or support transition.

Because the package tree is vendored whole, the same change updates
`install.sh`'s existing exhaustive checksum manifest and spec-0005's bundle
inventory. `plugins/trellis/README.md` names the carrier derivation and
corrects obsolete host/package wording without expanding support claims.

### 5. Distribution and release remain outside this decision

Stewards owns whether a host marketplace exposes the Trellis package and the
narrow CI authoring skill that can register that marketplace before a job.
Catalog admission remains host-specific and evidence-gated. Neither catalog
is changed by this decision.

Trellis adds no:

- immutable tag or GitHub Release;
- release metadata or inventory;
- release workflow, history, or approval record;
- generated support document;
- shared family schema/runtime/validator;
- synchronized version or bump policy; or
- support inference from package or marketplace presence.

Superseded `decision-0059` and `spec-0008` remain historical and authorize
nothing. This decision implements the small local choice that
`decision-0060` left open; it does not revive their family machinery.

## Consequences

- Claude and Codex declarations identify one coherent Trellis package version.
- The package version can advance intentionally without disturbing the
  payload-hash staleness contract or live consumer rows.
- Surface facts have one small product-owned home, while marketplace
  availability stays independently evidenced.
- The existing plugin-vendor script carries the new package files because it
  already promises to carry the complete plugin tree.
- Future surface tests can add bounded rows or observations without requiring
  a family release system.
- A package change now requires an intentional Trellis-local SemVer update;
  this decision does not automate or centrally dictate that judgment.

## Open questions

None.

## Self-check

The decision resolves the exact local choice invited by `decision-0060`,
preserves `decision-0058`'s behavioral boundary, and changes only the
commit-version sentence in `decision-0043`. It distinguishes package,
payload, behavior, and marketplace facts rather than deriving one from
another. The initial version avoids a verified stale-cache value. The
implementation surface is bounded to four carriers, one existing-language
test, and existing checksum/document propagation; none of the retired family
release machinery returns.

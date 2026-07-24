---
id: decision-0059
type: decision
status: superseded  # decision-0060 records the maintainer's 2026-07-24 scope reset; historical approval remains documented below
superseded_by: [decision-0060]
depends_on: [decision-0036, decision-0043, decision-0058, kodhama/kodhama-0015-family-plugin-release-and-surface-contract, kodhama/kodhama-0016-distribution-availability-and-effective-support]
owner: agent
date: 2026-07-24
---

# 0059 — adopt the family plugin release and surface contract without conflating package, payload, or distribution state

> **Human direction (2026-07-24).** The maintainer chose a common family
> SemVer and distribution strategy, asked that each plugin adopt it in the way
> that fits that product, and authorized the end-to-end rollout across
> Stewards, Grove, Trellis, and Wisp. This record applies that direction to
> Trellis; the independent soundness gate precedes the approval record.

## Context

Trellis currently carries two incompatible package identities. Its Claude
manifest deliberately omits `version` under `decision-0036`, so Claude treats
the repository commit as the plugin version. Its Codex manifest says
`0.1.0`, but no product-owned release authority or release tag binds that
value. Separately, the installed rule payload correctly identifies its own
rendered bytes as `payload@<content-hash>` under `decision-0043`.

The family contract now separates those concerns:

- Stewards decision `kodhama-0015` requires a product-owned SemVer package
  release, immutable `<plugin>-v<version>` tag, source commit, and
  version-bound product surface contract.
- Stewards decision `kodhama-0016` keeps catalog/provisioner availability
  separate from product support and derives effective support only when all
  independently owned facts agree.

Trellis also has a deliberately narrow host claim. `decision-0058` supports
Claude's import path and local Codex fresh-start delivery, while keeping
resume, clear, compact, subagents, desktop, IDE, headless, and cloud
evidence-gated. Package versioning must not promote any of those candidates,
and distribution success must not stand in for live-rule behavior.

The repository's old `v0.x.y` tags identify the retired Go binary channel.
Reusing that namespace for plugin packages would make two unrelated release
lines look identical.

## Decision

### 1. Trellis adopts product-owned SemVer for the plugin package

`plugins/trellis/VERSION` becomes the one canonical package-version
authority. Both host manifests carry exactly that value, and release
validation fails if either differs.

The first coherent dual-host package release is `0.1.0`, matching the existing
Codex manifest. This is not a claim that the historical binary `v0.1.0` and
the plugin package are the same artifact. Plugin release tags use the
collision-safe family form:

`trellis-v<version>`

The initial tag is therefore `trellis-v0.1.0`. It points to the exact commit
whose package tree, manifests, surface contract, vendoring manifest, and
validation results all carry `0.1.0`.

### 2. Package version, release ref, source commit, and payload identity remain separate

Trellis records four identities without substitution:

| Identity | Trellis form |
|---|---|
| package version | `plugins/trellis/VERSION`, for example `0.1.0` |
| immutable release ref | `trellis-v0.1.0` |
| source identity | the full commit named by that tag |
| generated payload identity | `plugins/trellis/reference/version`, for example `payload@0760a802ccd1` |

`plugins/trellis/reference/version` remains the content-derived payload stamp.
Setup copies it to `.trellis/internal/version`, and staleness remains the
file-to-file comparison defined by `decision-0043`. A SemVer value never
replaces that stamp.

Once this package is distributed, a payload-byte change requires a package
release because the payload is consumer-observable managed content. A
package-only metadata or evidence change may change SemVer without changing
the payload stamp. Release validation records both values so either case is
visible.

### 3. Trellis owns one version-bound surface contract

`plugins/trellis/surfaces.json` becomes the product-owned surface authority.
It conforms to the versioned Stewards family schema and registry, carries the
same package version as `VERSION`, and may add Trellis-namespaced fields for:

- host-native transport;
- context-loading boundary;
- installed-file fallback;
- live-row reload behavior;
- duplicate-context prevention;
- setup/refresh/remove evidence; and
- payload-stamp diagnostics.

Every canonical surface row is `supported`, `candidate`, or `unsupported`.
Only `supported` is a product behavior claim, and it requires retained
exact-surface evidence plus a support record. Candidate and unsupported rows
name the missing capability and user-visible disclosure. Evidence never
flows from one lifecycle event, host, local/cloud boundary, or interactive/
headless surface to another.

The surface contract and generated public support table must preserve
`decision-0058`'s current boundary. This decision does not promote resume,
clear, compact, subagents, desktop, IDE, headless, automation, or cloud.

### 4. Stewards owns availability; Trellis owns behavior and setup

Stewards owns Claude/Codex catalog state, generic pre-agent marketplace
provisioning, clean-install evidence, and the effective-support rendering.
Trellis owns its package, manifests, rule delivery, hook and fallback
behavior, setup/refresh/remove semantics, and behavioral evidence.

The existing mutable Claude catalog entry is disclosed legacy published stock
until `trellis-v0.1.0` and the version-bound surface contract exist. It is not
distribution-verified and cannot contribute to effective support. The first
family implementation wave cannot close while the entry remains legacy.
After Trellis publishes the conforming release, Stewards either adopts or
replaces the entry as a normal published record or delists it. Catalog
`verified` requires an immutable selector; provisioner `verified` separately
requires exact-release acquisition. Neither verification route substitutes
for resolving the legacy entry itself.

A successful Stewards provisioner run proves acquisition and pre-agent
installation only. It does not prove Trellis loaded rules, reloaded current
rows, avoided duplicate context, or completed product setup.

### 5. Release validation and tagging are product-owned

Trellis adds product-local release validation that checks, before tagging:

- canonical SemVer syntax and equality across `VERSION`, both manifests, and
  `surfaces.json`;
- exact canonical surface IDs and allowed support states against the pinned
  Stewards contract version;
- required exact-surface evidence and support records for every supported row;
- generated support documentation parity;
- the existing payload generation/checksum/self-application contracts;
- setup/remove reversibility and host-isolated hook behavior;
- whole-plugin `install.sh` inventory and checksums, including the new release
  and surface metadata; and
- release-tree cleanliness and tag/version equality.

The release workflow creates `trellis-v<VERSION>` only after those checks pass
on the merged release commit. It is idempotent for the same commit and fails
loudly if the tag exists elsewhere. Existing binary-era `v0.x.y` tags and
frozen binaries remain historical and are never rewritten.

The family bump rules apply to Trellis's public plugin contract:

- patch for a backward-compatible fix or evidence/provenance correction;
- minor for a backward-compatible capability/supported-surface addition and,
  by family convention below `1.0.0`, a breaking public-contract change; and
- major for a breaking public-contract change at or after `1.0.0`.

False support is withdrawn immediately; version convenience never delays the
correction.

### 6. Existing decisions narrow at exact points

- `decision-0036` is superseded for plugin package identity and update
  detection: the package no longer versions by commit or omits its manifest
  version. Its observation that third-party marketplace auto-update is
  host-controlled remains historical context, not a release mechanism.
- `decision-0043` point 4 is superseded only where it says the shipped plugin
  artifact is HEAD because plugin versions are commits. Its generator-only
  CLI, payload stamp, file-to-file staleness, vendor-script distinction, and
  retired binary-release channel stand.
- `decision-0058` is refined, not reversed: its payload/live-row authority,
  support phases, host transports, fallback, and no-double-injection rules
  remain product truth. The surface contract records those states, while
  Stewards records distribution availability separately.

On approval, `decision-0036` and `decision-0043` receive append-only
forward-pointer annotations limited to these boundaries. `decision-0058` is a
dependency whose delivery contract is extended without supersession, so it
receives no misleading supersession pointer. The implementation spec derives
only after this record reaches its intent gate.

## Consequences

- Claude and Codex see one coherent Trellis package version without making
  their behavioral support claims identical.
- Consumers and diagnostics can distinguish “which plugin release,” “which
  source commit,” and “which rendered payload” instead of overloading one
  stamp.
- Payload changes now require an explicit plugin release; ordinary row edits
  in consumer-owned `.trellis/rules.toml` still take effect at the next
  supported context-loading boundary without a package bump or refresh.
- Grove's already working release machinery remains a reference, not copied
  product code. Wisp and Trellis may choose different canonical authority
  files while conforming to the same family invariant.
- Headless and cloud provisioning can proceed in Stewards without promoting
  Trellis behavior until issues `stewards#14`/`#15` and
  `trellis#182`/`#183` produce exact-surface evidence.
- The initial release adds no public-directory submission, GitHub Release
  object, or synchronized family release train.

## Open questions

- None for adoption. The exact Stewards schema revision and generated support
  table format belong to the implementation spec and must conform to the
  merged family contract.

## Self-check

The decision consumes settled `decision-0036` plus approved `decision-0043`,
`decision-0058`, and Stewards contracts. It reverses
the exact commit-version clause that conflicts with family SemVer, preserves
the payload stamp as a different identity, and does not infer support from
packaging or installation. The initial `0.1.0` value is grounded in the
existing Codex manifest, while the prefixed tag avoids collision with the
retired binary history. Release ownership stays in Trellis; catalog and
generic provisioning ownership stay in Stewards. No unsupported surface is
promoted. The independent decision-adversary returned SOUND at `c41b2ba`;
the `approved` frontmatter records the maintainer's already stated rollout
intent after that condition was satisfied.

## Supersession record

`decision-0060` supersedes this adoption in full. Trellis keeps the independent
local live-rule delivery defined by `decision-0058`/`spec-0007@v1`, but does
not implement this record's family release certification, shared SemVer/tag,
release metadata/inventory/surface, history/approval/workflow, or validator-
runtime machinery. The original approval record above remains historical.

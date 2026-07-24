---
id: spec-0008
type: spec
status: approved  # v3 gates passed; records the maintainer's standing family-wave merge authorization
version: 3
depends_on: [decision-0059, decision-0058, spec-0007@v1, kodhama/kodhama-spec-0001-family-plugin-release-and-distribution-metadata@v2]
implements: decision-0059
owner: agent
rubric: rubric-artifact-contract
date: 2026-07-24
---

# Spec 0008 — Trellis plugin release and surface contract

> **Amended 2026-07-24 — acyclic bundle binding and validated history staging.**
> **WHAT:** Replaced the recursive raw checksum between `install.sh` and
> `release-inventory.json` with one exact normalized inventory preimage plus
> independent raw-install-script field checks, and reordered the post-tag flow
> so the candidate or already-recorded history row is present before family
> release validation and is committed only after that validation passes.
> **WHY:** V2 required each file's raw digest to contain the other's raw digest
> and invoked `--phase release` before the family-required appended row
> existed; neither contract admitted a conforming execution.
> **SCOPE:** Whole-plugin bundle generation/checking, post-tag release/history
> staging, retry/conflict semantics, S5–S6, S8, S12–S13, and R9, R18, R20, R24,
> R33–R36; package and payload identities, support rows, validator runtime,
> product/distribution ownership, and all other v2 behavior remain unchanged.
> **POINTER:** `decision-0059` point 5 and approved
> `kodhama/kodhama-spec-0001-family-plugin-release-and-distribution-metadata@v2`
> S16/S22/S26 and R29/R39/R45.
> **VALUE:** Trellis maintainers can generate one finite release tree and prove
> the exact history row before publishing its history-only change.
> **CONFIDENCE:** verified.

> **Amended 2026-07-24 — family v2 validator-runtime protocol.**
> **WHAT:** Bound the Trellis extension validator to the family v2 request/result
> protocol and immutable runtime digest, and made an explicit POSIX
> `--runtime-store` argument mandatory for both product-validation phases.
> **WHY:** The approved family v2 contract replaced v1's underdetermined
> extension-validator launch boundary with an executable, fail-closed
> runtime-store and subprocess contract.
> **SCOPE:** Release metadata, product extension validation, pre-tag/release
> invocations, S10–S11, and R28–R32; the six decision-0058 support rows,
> package/payload identities, and product/distribution ownership remain
> unchanged.
> **POINTER:** Stewards merge
> `fe95bb93e59e4e24faaabe5ddfe1a6c8e8b9215c`, approved
> `kodhama-spec-0001-family-plugin-release-and-distribution-metadata@v2`.
> **VALUE:** Trellis maintainers can validate the product extension through one
> deterministic family interface without transferring Trellis behavior
> ownership to Stewards.
> **CONFIDENCE:** verified.

## Purpose

Define the smallest product-owned release contract that lets Trellis publish
one dual-host SemVer package, retain its distinct payload stamp, preserve the
support boundary in `decision-0058`, conform to the family v2 validator-runtime
protocol, and supply the exact inputs Stewards needs to retire Trellis's legacy
Claude catalog exception.

## Scope

This spec covers Trellis package metadata, product validation, support
derivatives, release tagging/history, the product side of the family extension-
validator protocol, and the Stewards adoption handoff.

It does not:

- implement or specify the Stewards provisioner;
- provision, discover, or choose the caller/Stewards-owned immutable runtime
  store consumed by family validation;
- change Trellis setup, refresh, remove, hook, fallback, or live-row behavior
  defined by `spec-0007@v1`;
- make catalog publication or clean installation a product-support claim;
- add a GitHub Release object, public-directory submission, or synchronized
  family release train; or
- promote a lifecycle event, client, headless, automation, or cloud boundary
  that `decision-0058` leaves evidence-gated.

## Authorities and paths

All paths are relative to the Trellis repository root. Product validation uses
the repository root as `--package-root`; the distributed package remains
`plugins/trellis/`.

| Path | Role | Mutability |
|---|---|---|
| `plugins/trellis/VERSION` | Sole package-version authority | Release source |
| `plugins/trellis/.claude-plugin/plugin.json` | Claude host manifest and version carrier | Release source |
| `plugins/trellis/.codex-plugin/plugin.json` | Codex host manifest and version carrier | Release source |
| `plugins/trellis/release.json` | Family `release-metadata.v1` document | Release source |
| `plugins/trellis/release-inventory.json` | Family `release-inventory.v1` document | Generated, checked in |
| `plugins/trellis/surfaces.json` | Family `surface-contract.v1` plus Trellis extension | Release source |
| `plugins/trellis/bin/release-contract.mjs` | Product inventory provider, extension validator, derivative generator, and adoption emitter | Product validator |
| `plugins/trellis/SUPPORT.md` | Public support table | Generated, checked in |
| `README.md` | Root support block | Generated region |
| `plugins/trellis/README.md` | Package support block | Generated region |
| `release/trellis/history.json` | Family `release-history.v1` ledger | Append-only source; not distributed in the plugin |
| `release/trellis/approvals/<version>.json` | Product-human release approval bound to the family approval projection | Immutable once referenced |
| `.github/workflows/trellis-plugin-release.yml` | Product-owned validation and tag workflow | Release mechanism |
| `install.sh` | Whole-plugin vendoring manifest | Derived checksum carrier |

`plugins/trellis/release.json`, `release-inventory.json`, `surfaces.json`,
`SUPPORT.md`, and `bin/release-contract.mjs` are part of the plugin tree and
therefore part of `install.sh`'s complete file inventory and checksum guard.
Every plugin file except `release-inventory.json` uses its raw SHA-256 there;
the inventory uses the acyclic normalized preimage below and independently
binds the raw `install.sh` bytes. The post-tag history ledger is outside that
tree so appending the row permitted by the family contract does not change the
released package bytes.

## Package and payload identity

### Canonical version and carriers

`plugins/trellis/VERSION` is UTF-8 without BOM and contains exactly one SemVer
2.0.0 value plus one terminal LF. Its initial value is `0.1.0`.

`plugins/trellis/release.json` declares exactly this authority and these two
host-manifest carriers:

| Field | Path | Extractor |
|---|---|---|
| `version_authority` | `plugins/trellis/VERSION` | `plain-text` |
| Claude carrier | `plugins/trellis/.claude-plugin/plugin.json` | `json`, selector `/version`, role `host-manifest` |
| Codex carrier | `plugins/trellis/.codex-plugin/plugin.json` | `json`, selector `/version`, role `host-manifest` |

Both manifests shall carry `"version": "0.1.0"` for the initial release.
The inventory provider shall discover both manifests, emit each exactly once
in `host_manifests`, and emit the same extractors exactly once in
`version_carriers`. Discovery of another host manifest fails until it is added
to both sets; a declared carrier with no discovered manifest also fails.

The expected package tag is computed only as
`trellis-v<plugins/trellis/VERSION>`. Existing unprefixed `v0.x.y` binary tags
are outside this package-tag namespace and shall neither enter plugin history
nor be moved, deleted, or rewritten.

### Distinct payload identity

`plugins/trellis/reference/version` remains the sole Trellis payload identity.
It shall match `^payload@[0-9a-f]{12}$` after removing exactly one optional
terminal LF, and shall not contain or derive from the package SemVer.

The inventory declares it once as:

- `payload_id: trellis.rendered-payload`;
- `kind: version-stamp`;
- extractor `text-line` at
  `plugins/trellis/reference/version`, line `0`; and
- consumer-acted `true`, because setup copies it and staleness compares it.

The same exact kind/value shall appear in `release.json.payload_identity`, the
current release-inventory payload array, and the release-history row. Because
the payload is consumer-acted, the inventory shall also carry its required
`consumed-output-protocol` public-contract item. A package release may change
without changing this value; any change to the value or rendered payload bytes
requires a package release.

## Release metadata and inventory

`plugins/trellis/release.json` conforms to family contract version `1` and
contains these product-fixed values:

| Field | Value |
|---|---|
| `schema_version` | `1` |
| `family_contract_version` | `1` |
| `plugin_id` | `trellis` |
| `surface_contract` | `plugins/trellis/surfaces.json` |
| `release_inventory` | `plugins/trellis/release-inventory.json` |
| `release_history` | `release/trellis/history.json` |
| `inventory_provider` | `plugins/trellis/bin/release-contract.mjs` |
| `extensions.trellis.validator` | `plugins/trellis/bin/release-contract.mjs` |
| `extensions.trellis.validator_runtime_sha256` | Lowercase 64-hex digest of the selected canonical `extension-validator-runtime.v1` manifest |
| `extensions.trellis.family_validator` | Family `repo-path` stable reference to `kodhama/kodhama:distribution/manage` |

The metadata also carries the authority, carriers, payload identity, and the
current immutable `release_approval` reference in release phase. No copied
Stewards schema becomes a Trellis authority. The two reserved validator fields
are present together. The runtime digest is product-selected release input; it
identifies the immutable runtime object but does not make its store or
provisioning a Trellis-owned product surface.

The `schema_version` and `family_contract_version` values remain `1` because
Stewards v2 completes the version-1 family schemas and interfaces rather than
renaming them. The `@v2` dependency pin identifies the approved behavioral
state that governs those version-1 documents.

The executable inventory-provider interface is exactly:

```text
plugins/trellis/bin/release-contract.mjs \
  --package-root <repository-root> \
  --emit-release-inventory
```

It writes canonical `release-inventory.v1` JSON to stdout and no other stdout
bytes, writes no file, performs no network access, and produces identical bytes
on two runs in a clean checkout. Those bytes shall equal the checked-in
`plugins/trellis/release-inventory.json`.

The provider shall inventory every family-required host manifest, payload
identity, public-contract item, support derivative, and host-manifest support
claim. The initial `public_contract_items` set is exactly this table; every
extractor is an additional-properties-forbidden family extractor:

| `contract_id` | Category | Path | Extractor |
|---|---|---|---|
| `trellis.configuration.posture-a` | `configuration` | `plugins/trellis/reference/rules-a.toml` | `file-bytes` |
| `trellis.configuration.posture-b` | `configuration` | `plugins/trellis/reference/rules-b.toml` | `file-bytes` |
| `trellis.documentation.package-install` | `installation-input` | `plugins/trellis/README.md` | `file-bytes` |
| `trellis.documentation.root-install` | `installation-input` | `README.md` | `file-bytes` |
| `trellis.entrypoint.remove` | `host-visible-entrypoint` | `plugins/trellis/skills/remove/SKILL.md` | `file-bytes` |
| `trellis.entrypoint.setup` | `host-visible-entrypoint` | `plugins/trellis/skills/setup/SKILL.md` | `file-bytes` |
| `trellis.hook.claude.registration` | `host-visible-entrypoint` | `plugins/trellis/hooks/hooks.json` | `file-bytes` |
| `trellis.hook.claude.staleness` | `consumed-output-protocol` | `plugins/trellis/hooks/staleness.sh` | `file-bytes` |
| `trellis.hook.codex.context` | `consumed-output-protocol` | `plugins/trellis/hooks/codex-context.mjs` | `file-bytes` |
| `trellis.hook.codex.registration` | `host-visible-entrypoint` | `plugins/trellis/hooks/codex-hooks.json` | `file-bytes` |
| `trellis.install.claude-coordinate` | `installation-coordinate` | `plugins/trellis/.claude-plugin/plugin.json` | `file-bytes` |
| `trellis.install.codex-coordinate` | `installation-coordinate` | `plugins/trellis/.codex-plugin/plugin.json` | `file-bytes` |
| `trellis.install.vendor-script` | `installation-input` | `install.sh` | `file-bytes` |
| `trellis.managed.block-claude` | `managed-state` | `plugins/trellis/reference/block-claude.md` | `file-bytes` |
| `trellis.managed.block-codex` | `managed-state` | `plugins/trellis/reference/block-codex.md` | `file-bytes` |
| `trellis.managed.block-inline-a` | `managed-state` | `plugins/trellis/reference/block-inline-a.md` | `file-bytes` |
| `trellis.managed.block-inline-a-head` | `managed-state` | `plugins/trellis/reference/block-inline-a-head.md` | `file-bytes` |
| `trellis.managed.block-inline-b` | `managed-state` | `plugins/trellis/reference/block-inline-b.md` | `file-bytes` |
| `trellis.managed.block-inline-b-head` | `managed-state` | `plugins/trellis/reference/block-inline-b-head.md` | `file-bytes` |
| `trellis.managed.block-inline-tail` | `managed-state` | `plugins/trellis/reference/block-inline-tail.md` | `file-bytes` |
| `trellis.managed.invariants` | `managed-state` | `plugins/trellis/reference/invariants.md` | `file-bytes` |
| `trellis.managed.rules` | `managed-state` | `plugins/trellis/reference/rules.md` | `file-bytes` |
| `trellis.managed.trellis-a` | `managed-state` | `plugins/trellis/reference/trellis-a.md` | `file-bytes` |
| `trellis.managed.trellis-b` | `managed-state` | `plugins/trellis/reference/trellis-b.md` | `file-bytes` |
| `trellis.protocol.payload-checksums` | `consumed-output-protocol` | `plugins/trellis/reference/checksums` | `file-bytes` |
| `trellis.protocol.payload-stamp` | `consumed-output-protocol` | `plugins/trellis/reference/version` | `text-line`, line `0` |
| `trellis.protocol.release-metadata` | `consumed-output-protocol` | `plugins/trellis/release.json` | `file-bytes` |
| `trellis.protocol.release-validator` | `consumed-output-protocol` | `plugins/trellis/bin/release-contract.mjs` | `file-bytes` |
| `trellis.protocol.version-authority` | `consumed-output-protocol` | `plugins/trellis/VERSION` | `text-line`, line `0` |
| `trellis.support.public-table` | `surface-support` | `plugins/trellis/SUPPORT.md` | `file-bytes` |
| `trellis.support.surface-contract` | `surface-support` | `plugins/trellis/surfaces.json` | `file-bytes` |

The provider compares `git ls-files plugins/trellis` plus `install.sh` and
`README.md` with the table paths. The only permitted unmatched path is the
generated self-manifest `plugins/trellis/release-inventory.json`; its digest is
bound separately by family release history. The approval and history files,
workflow, tests, and fixtures are outside the distributed package/public
contract and are bound by their family-specific release roles. Any other
tracked path fails inventory generation until this spec's current-truth map is
revised; no implementation-only allowlist exists.

### Acyclic whole-plugin bundle binding

The `trellis.install.vendor-script` public-contract item remains a
`file-bytes` extractor over raw `install.sh`. Its `source.sha256` is the
SHA-256 of those raw bytes. Its `fingerprint` is the family-required SHA-256
of these canonical JSON bytes with no terminal LF, where `<install-sha256>` is
that same raw digest:

```json
{"category":"installation-input","contract_id":"trellis.install.vendor-script","extracted_sha256":"<install-sha256>"}
```

`install.sh`'s bundle manifest contains every tracked regular file beneath
`plugins/trellis/` exactly once, sorted by normalized relative path. Every path
except `release-inventory.json` carries the SHA-256 of its exact raw bytes.
The `release-inventory.json` entry carries the SHA-256 of this one normalized
preimage:

1. Require the checked-in inventory to be valid canonical
   `release-inventory.v1` JSON followed by exactly one LF and equal to the
   provider's stdout.
2. Locate exactly one `public_contract_items` row whose `contract_id` is
   `trellis.install.vendor-script`; require its category, source path, and
   extractor to be exactly the raw `install.sh` contract above.
3. In an in-memory copy, replace only that row's `source.sha256` and
   `fingerprint` values with exactly 64 ASCII `0` characters each.
4. Canonicalize the complete copied object with the family-v2 canonical JSON
   grammar, append exactly one LF, and SHA-256 those bytes.

The manifest path selects this normalization; no other path may use a
normalized or excluded-byte digest. Bundle verification shall recompute that
normalized digest before installation. Release generation, inventory
validation, and check mode shall separately recompute the raw repository
`install.sh` digest, the row's exact `source.sha256`, and its exact family
fingerprint above. A missing/duplicate row, a different zeroed field, an
unexpected raw-checksum exception, or any mismatch fails the applicable
bundle or release check.

Release preparation is ordered and converges without a fixed point while
preserving `--generate`'s closed write set:

1. run product generation once, forming the complete inventory from the
   current raw `install.sh`;
2. run the existing bundle-manifest regeneration over that inventory's
   normalized preimage and every other plugin path's raw bytes;
3. run product generation again so the final inventory records the resulting
   raw `install.sh` source digest and family fingerprint; and
4. rerun both generators in check mode and fail unless neither direction
   changes.

This is whole-plugin coverage, not an omitted self-manifest: every other
plugin byte is raw-hashed by `install.sh`; every inventory byte outside the two
normalized values affects its bundle digest; the two normalized values are
recomputed from raw `install.sh` by release validation; and the complete raw
inventory is additionally bound by the family release-history digest.
Generation and check mode shall reject a change to any one of those bytes or
bindings.

### Family v2 runtime-store and extension-validator interface

The family engine, not the Trellis entrypoint, owns platform detection,
runtime-store resolution and verification, the immutable runtime projection,
the process-tree filesystem/network sandbox, repeated execution, and common
request/result validation. It shall apply the exact canonical JSON, POSIX
platform vocabulary, content-addressed store layout, execution order, process
contract, limits, audit flags, and diagnostics in the approved family spec
`@v2`.

Every `validate-product` invocation shall receive one explicit
`--runtime-store <absolute-posix-path>` argument. The value shall be a
`/`-rooted normalized POSIX path whose existing components are non-symlink
directories. Neither the Trellis workflow nor the product validator shall
substitute an environment variable, default directory, package-relative path,
upward search, fetched object, or alternate runtime when the argument or its
digest object is absent or invalid. Runtime-store provisioning and selection
remain caller/Stewards infrastructure; Trellis consumes only the explicit path
and the product-selected digest bound in `release.json`.

On a family-v2 supported POSIX launcher, the family engine shall detect only
the four Linux/macOS `x86_64`/`aarch64` tuples through its direct `uname`
mapping, then resolve only:

```text
<runtime-store>/sha256/<first-2-digest-hex>/<remaining-62-digest-hex>/manifest.json
<runtime-store>/sha256/<first-2-digest-hex>/<remaining-62-digest-hex>/image/
```

Windows, failed/unknown `uname` results, or another platform shall fail with
`error: extension-runtime-platform-unsupported\n` before store lookup.
Unavailable, malformed, platform-mismatched, drifted, or unenforceable
runtimes shall fail before validator spawn with the exact distinct family-v2
diagnostic carrying the declared digest. Trellis defines no fallback from
those failures.

For the single `extensions.trellis` declaration, the common engine invokes the
resolved non-symlink executable exactly twice per phase as:

```text
<resolved-absolute-release-contract.mjs> --stewards-extension-validator-v1
```

It uses the resolved repository root supplied as `--package-root` as the
working directory, passes no other argument, writes one canonical
`extension-validator-request.v1` object plus one LF to stdin, and requires one
canonical `extension-validator-result.v1` object plus one LF on stdout. The
family-v2 environment, file-descriptor, timeout, size, immutable
runtime/package visibility, private `/tmp`, `/dev/null`, no-network, audit, and
no-persistent-mutation rules apply without a Trellis exception.

The Trellis entrypoint accepts only a schema-version `1`
`product-extension` request whose `namespace` and `plugin_id` are both
`trellis`, whose paths name the release metadata, surface contract, inventory,
and history already validated by the family engine, and whose
`validator_runtime_sha256` equals the declaration. It validates the complete
request `extension` value and the product surface extension against the exact
Trellis contract in this spec, including every decision-0058 support state,
transport, boundary, exclusion, fallback, live-row, duplicate-context,
setup/refresh/remove, and payload-diagnostic value.

With no Trellis finding, the entrypoint emits
`{schema_version: 1, request_sha256: <sha256-of-exact-stdin-with-LF>, outcome:
"pass", findings: []}` and exits `0`. With one or more product findings, it
emits `outcome: "fail"` with sorted, duplicate-free family-v2 finding objects
at the applicable request instance pointers and exits `1`. It writes no
stderr for either protocol exit. Malformed request/result bytes, another
identity or phase binding, a request-hash mismatch, differing repeated
results, another exit, stderr, timeout, audit flag, or persistent mutation
fails common validation. Product finding codes and messages remain
Trellis-owned diagnostics; Stewards shall neither reinterpret them as family
support facts nor treat a passing process envelope as product evidence.

## Product surface contract

### Canonical rows

`plugins/trellis/surfaces.json` has `schema_version: 1`,
`family_contract_version: 1`, and `version` equal to `VERSION`. It contains
each family registry-v1 row exactly once, sorted by `surface_id`:

| Surface | Initial state | `missing_capability` / `disclosure` when candidate |
|---|---|---|
| `claude-code.ci.headless` | `candidate` | `No retained exact-surface evidence proves Claude Code CI/headless loads the installed payload and current rows exactly once.` / `Candidate only: Trellis does not support Claude Code CI/headless.` |
| `claude-code.cloud-container.headless` | `candidate` | `No retained exact-surface evidence proves Claude Code cloud-container/headless installation and live rule delivery.` / `Candidate only: Trellis does not support Claude Code cloud-container/headless.` |
| `claude-code.local.interactive` | `supported` | forbidden / forbidden |
| `codex.ci.headless` | `candidate` | `No retained exact-surface evidence proves Codex CI/headless installation, startup hook delivery, and current-row loading.` / `Candidate only: Trellis does not support Codex CI/headless or automation.` |
| `codex.cloud-container.headless` | `candidate` | `No retained exact-surface evidence proves Codex cloud-container/headless trust, installation, hook delivery, and current-row loading.` / `Candidate only: Trellis does not support Codex cloud-container/headless.` |
| `codex.local.interactive` | `supported` | forbidden / forbidden |

Each supported row carries exact-version evidence, a load path, a support
record, and a required post-install setup contract. Each candidate row carries
the missing capability and a user-visible disclosure that the row is not
supported. Evidence, support records, and setup references use the family
stable-reference types and bind the exact plugin version and surface.

### Trellis extension

`extensions.trellis` is exactly
`{schema_version: 1, surfaces: {<surface_id>: <row-extension>}}`; `surfaces`
has the same six keys as the common surface array. The object and every row
extension forbid unknown fields. Each row extension records the decision-0058
dimensions that the family registry does not split:

| Field | Contract |
|---|---|
| `transport_id` | One exact transport id below, or `null` for a non-supported row |
| `supported_boundary_ids` | Sorted duplicate-free array of the exact boundary ids below |
| `excluded_claims` | Sorted duplicate-free objects with `claim_id`, `state: candidate`, exact `missing_capability`, and exact `disclosure` |
| `installed_file_fallback` | Object with `state`, `load_path`, `missing_capability`, `disclosure`, and `evidence`; inapplicable fields are `null` |
| `live_row_reload` | Object with `state`, `boundary_id`, and exact evidence stable references |
| `duplicate_context_prevention` | Object with `state` and exact evidence stable references |
| `setup_refresh_remove` | Object with `state` and exact evidence stable references |
| `payload_stamp_diagnostics` | Fixed path `plugins/trellis/reference/version`, pattern `^payload@[0-9a-f]{12}$`, and role `diagnostic-only` |

The two supported rows carry these exact values:

| Surface | `transport_id` | `supported_boundary_ids` | Family `load_path` |
|---|---|---|---|
| `claude-code.local.interactive` | `claude-managed-import` | `["claude-code.session-load"]` | `kind: host-discovery`, `locator: CLAUDE.md:trellis-managed-import`, `invocation: claude-code.session-load` |
| `codex.local.interactive` | `codex-session-start-startup-hook` | `["codex.session-start.startup"]` | `kind: hook`, `locator: plugins/trellis/hooks/codex-hooks.json#SessionStart[startup]`, `invocation: node "${PLUGIN_ROOT}/hooks/codex-context.mjs"` |

For both supported rows, `live_row_reload`,
`duplicate_context_prevention`, and `setup_refresh_remove` have
`state: supported` and non-empty exact-release stable-reference evidence.
`live_row_reload.boundary_id` is `claude-code.session-load` for Claude and
`codex.session-start.startup` for Codex.
Claude's `installed_file_fallback` has `state: not-applicable` and all other
fields `null`, because installed-file imports are its supported transport.
Codex's value is exactly:

| Field | Value |
|---|---|
| `state` | `candidate` |
| `load_path` | `kind: host-discovery`, `locator: AGENTS.md:trellis:codex-bootstrap`, `invocation: read-installed-overlay-before-substantive-work` |
| `missing_capability` | `Model-directed fallback execution is not deterministic host enforcement.` |
| `disclosure` | `Best-effort fallback only: absence of native-hook delivery is not itself support proof.` |
| `evidence` | `[]` |

All `excluded_claims` arrays are projections of this exact sorted registry.
The Claude local row selects `claude-code.native-hook-replacement`; the Codex
local row selects `codex.clear`, `codex.compact`, `codex.desktop`, `codex.ide`,
`codex.resume`, and `codex.subagent`:

| `claim_id` | `missing_capability` | `disclosure` |
|---|---|---|
| `claude-code.ci.headless` | `No retained exact-surface evidence proves Claude Code CI/headless loads the installed payload and current rows exactly once.` | `Candidate only: Trellis does not support Claude Code CI/headless.` |
| `claude-code.cloud-container.headless` | `No retained exact-surface evidence proves Claude Code cloud-container/headless installation and live rule delivery.` | `Candidate only: Trellis does not support Claude Code cloud-container/headless.` |
| `claude-code.native-hook-replacement` | `No live positive and negative control proves a Claude native hook replaces managed imports without duplicate context.` | `Not supported: Claude Code continues to use managed imports; no native-hook replacement is claimed.` |
| `codex.automation` | `No exact automation proof establishes plugin trust, startup delivery, and current-row loading.` | `Not supported: Codex automation is outside Phase 1.` |
| `codex.ci.headless` | `No retained exact-surface evidence proves Codex CI/headless installation, startup hook delivery, and current-row loading.` | `Candidate only: Trellis does not support Codex CI/headless or automation.` |
| `codex.clear` | `No live positive and negative control proves the Codex clear boundary reloads current rows exactly once.` | `Not supported: Codex clear is outside Phase 1.` |
| `codex.cloud` | `No exact cloud proof establishes plugin trust, installation, startup delivery, and current-row loading.` | `Not supported: Codex cloud is outside Phase 1.` |
| `codex.cloud-container.headless` | `No retained exact-surface evidence proves Codex cloud-container/headless trust, installation, hook delivery, and current-row loading.` | `Candidate only: Trellis does not support Codex cloud-container/headless.` |
| `codex.compact` | `No live positive and negative control proves the Codex compact boundary preserves or reloads current rows exactly once.` | `Not supported: Codex compact is outside Phase 1.` |
| `codex.desktop` | `No exact Codex desktop proof establishes plugin root, trust boundary, context shape, and current-row loading.` | `Not supported: Codex desktop is outside Phase 1.` |
| `codex.headless` | `No exact headless proof establishes plugin trust, startup delivery, and current-row loading.` | `Not supported: Codex headless is outside Phase 1.` |
| `codex.ide` | `No exact Codex IDE proof establishes plugin root, trust boundary, context shape, and current-row loading.` | `Not supported: Codex IDE is outside Phase 1.` |
| `codex.resume` | `No live positive and negative control proves the Codex resume boundary reloads current rows exactly once.` | `Not supported: Codex resume is outside Phase 1.` |
| `codex.subagent` | `No live positive and negative control proves a Codex subagent receives one complete Trellis context.` | `Not supported: Codex subagents are outside Phase 1.` |

The headless and cloud rows have no supported context boundary. Their
`transport_id` is `null`, `supported_boundary_ids` is `[]`, and
`excluded_claims` selects their exact surface-id claim plus:
`codex.ci.headless` selects `codex.automation` and `codex.headless`;
`codex.cloud-container.headless` selects `codex.cloud` and `codex.headless`.
Their `installed_file_fallback` objects have `state: candidate`,
`load_path: null`, `evidence: []`, and these exact literals:

| Surface | `missing_capability` | `disclosure` |
|---|---|---|
| `claude-code.ci.headless` | `No retained exact-surface evidence proves Claude Code CI/headless reads the installed Trellis overlay when managed-import delivery is absent.` | `Candidate only: installed-file fallback is not supported on Claude Code CI/headless.` |
| `claude-code.cloud-container.headless` | `No retained exact-surface evidence proves Claude Code cloud-container/headless reads the installed Trellis overlay when managed-import delivery is absent.` | `Candidate only: installed-file fallback is not supported on Claude Code cloud-container/headless.` |
| `codex.ci.headless` | `No retained exact-surface evidence proves Codex CI/headless reads the installed Trellis overlay when startup-hook delivery is absent.` | `Candidate only: installed-file fallback is not supported on Codex CI/headless.` |
| `codex.cloud-container.headless` | `No retained exact-surface evidence proves Codex cloud-container/headless reads the installed Trellis overlay when startup-hook delivery is absent.` | `Candidate only: installed-file fallback is not supported on Codex cloud-container/headless.` |

For those same rows, `live_row_reload`,
`duplicate_context_prevention`, and `setup_refresh_remove` each have
`state: candidate` and `evidence: []`; every `boundary_id` is `null`. No
fixture, catalog record, provisioner result, local-host evidence, or
package-version change may change one of these states to `supported`.

The family engine invokes the product extension validator only through the
family-v2 process protocol above. It returns a product finding for any boundary
mismatch, unknown extension field, unsupported promotion, cross-surface
evidence reuse, or divergence from the exact decision-0058 exclusions above.
No standalone `--validate-surface-extension` subprocess is part of the public
release contract.

## Generated support and bundle derivatives

The same product entrypoint supplies:

```text
plugins/trellis/bin/release-contract.mjs --package-root <repository-root> --generate
plugins/trellis/bin/release-contract.mjs --package-root <repository-root> --check
```

`--generate` deterministically writes only:

- `plugins/trellis/release-inventory.json`;
- `plugins/trellis/SUPPORT.md`; and
- one named managed support block in `README.md` and
  `plugins/trellis/README.md`.

The complete support-claim corpus is closed to these five locations:

| Location | Structured claim carrier | Initial state |
|---|---|---|
| `plugins/trellis/SUPPORT.md` | exact managed Markdown block | present |
| `README.md` | exact managed Markdown block | present |
| `plugins/trellis/README.md` | exact managed Markdown block | present |
| `plugins/trellis/.claude-plugin/plugin.json` | optional JSON Pointer `/support` | absent |
| `plugins/trellis/.codex-plugin/plugin.json` | optional JSON Pointer `/support` | absent |

The Markdown block begins with
`<!-- trellis:support-claims:v1:begin -->` and ends with
`<!-- trellis:support-claims:v1:end -->`; each marker occurs exactly once in
each Markdown location. Bytes between and including the markers are identical
in all three files. The block contains, in this exact order:

1. `Package: trellis@<VERSION>`;
2. a table with header
   `| surface_id | status | transport_id | supported_boundary_ids | missing_capability | disclosure |`;
3. separator `|---|---|---|---|---|---|`, then one row per canonical surface
   sorted by `surface_id`;
4. a table with header
   `| claim_id | state | missing_capability | disclosure |`; and
5. separator `|---|---|---|---|`, then every excluded-claim registry row
   sorted by `claim_id`.

Table cells are exactly `null`, one canonical JSON string, or one compact
canonical JSON string array; raw pipes/newlines and Markdown escapes are
forbidden. The surface/claim values are exact projections of
`surfaces.json`. The block is UTF-8 without BOM, LF-only, with one terminal LF.

If a host manifest later carries `/support`, it is an
additional-properties-forbidden object with exactly:
`family_contract_version: 1`, `surface_contract:
"plugins/trellis/surfaces.json"`, `version: <VERSION>`, and
`supported: [{surface_id, transport_id, supported_boundary_ids}]`. The array
contains only supported rows, uses the exact extension values, and sorts by
`surface_id`; each nested object also forbids additional properties.

The initial `support_derivatives` set is exactly:

| `derivative_id` | Kind | Path / extractor |
|---|---|---|
| `trellis.support.package-readme` | `public-support-table` | `plugins/trellis/README.md`, `file-bytes` |
| `trellis.support.package-table` | `public-support-table` | `plugins/trellis/SUPPORT.md`, `file-bytes` |
| `trellis.support.root-readme` | `public-support-table` | `README.md`, `file-bytes` |

A future manifest `/support` adds exactly one `host-manifest-claim` derivative
for that manifest with a `json-pointer` extractor at `/support`. No other path,
free-text sentence, manifest description, test fixture, or provisioner output
is a release-valid product support claim. `--check` inspects only this closed
corpus, writes nothing, and names every missing/duplicate marker, malformed
carrier, stale projection, or undeclared structured carrier.

After generation, the existing payload regenerate/checksum/self-application
guards still pass. The whole-plugin `install.sh` manifest is regenerated from
the complete `plugins/trellis/` file set using the one normalized
`release-inventory.json` entry and raw checksums for every other path. It fails
on any missing, extra, stale, multiply listed, or wrongly normalized digest,
including all new release and surface files. The generated inventory's raw
install-script source digest and fingerprint must also match the final script.
Package SemVer is never written into `plugins/trellis/reference/version` or
`.trellis/internal/version`.

## Validation, approval, release, and history

### Product release check

The workflow checks out
`release.json.extensions.trellis.family_validator.source_commit` from
`kodhama/kodhama` into `$RUNNER_TEMP/kodhama`, then requires
`distribution/manage` path and SHA-256 to match that `repo-path` stable
reference. In a clean full-commit checkout with all tags fetched, the pre-tag
check is exactly:

```sh
"$RUNNER_TEMP/kodhama/distribution/manage" validate-product \
  --phase pre-tag \
  --package-root "$GITHUB_WORKSPACE" \
  --release-metadata "plugins/trellis/release.json" \
  --runtime-store "<caller-supplied-absolute-posix-path>"
"$GITHUB_WORKSPACE/plugins/trellis/bin/release-contract.mjs" \
  --package-root "$GITHUB_WORKSPACE" --check
(cd "$GITHUB_WORKSPACE/cli" && go test ./...)
git -C "$GITHUB_WORKSPACE" diff --exit-code -- .
test -z "$(git -C "$GITHUB_WORKSPACE" status --porcelain=v1)"
```

The angle-bracket runtime-store token denotes a required invocation input, not
a literal path or a default. The release caller supplies it explicitly; the
workflow does not provision, search for, or substitute the store.

`--check` includes the same Trellis surface-extension rules without launching
a second family protocol, two inventory emissions, support-corpus validation,
bundle-manifest validation, and all product derivatives. The check emits the
package version, expected tag, payload identity, surface/inventory digests,
and no caller-authored release identity.

The family validator owns common schema, history, compatibility, tag,
canonical request/result, runtime-store, and sandbox semantics. The Trellis
validator owns only the product extension, complete Trellis inventory, product
derivatives, existing Trellis guards, and product finding vocabulary.

### Human approval

The human release gate approves the family-defined canonical comparison and
proposed bump. `release/trellis/approvals/<version>.json` carries that approval
and `release.json.release_approval` references it. `decision-0059` is the
adoption decision, not a substitute for this release approval. Bump,
false-claim, and non-self-reference semantics are exactly the family contract.

### Tag workflow

`.github/workflows/trellis-plugin-release.yml` is manually dispatched with one
full merged commit and no caller-supplied version or tag. It checks out that
commit and all tags, requires the commit to be reachable from `origin/main`,
runs the complete pre-tag check, and creates/pushes only the computed
`trellis-v<VERSION>`.

It never creates or rewrites `v0.x.y`, never force-pushes a tag, and never
chooses or edits a version. After the tag push, the same workflow's
`post-tag-release` job owns candidate-row staging, release phase, and the
history-only follow-up. It first fetches `origin/main`, resolves
`HISTORY_BASE_COMMIT` from `refs/remotes/origin/main`, and creates one clean
dedicated history staging worktree at that exact commit. The tag's peeled
release commit shall be an ancestor of the base, and every release-controlled
path other than `release/trellis/history.json` shall be byte-identical to its
tagged retained bytes. A normal invocation uses that fresh worktree. Only
recovery after an interrupted invocation may reuse its existing worktree, and
only when `HEAD` still equals `HISTORY_BASE_COMMIT`, every non-history tracked
path still equals `HEAD`, and the history destination passes the exact
post-rename recovery classification below. In either mode the exact order is:

```sh
"$HISTORY_WORKTREE/plugins/trellis/bin/release-contract.mjs" \
  --package-root "$HISTORY_WORKTREE" \
  --append-release-history \
  --authoritative-history-commit "$HISTORY_BASE_COMMIT"
"$RUNNER_TEMP/kodhama/distribution/manage" validate-product \
  --phase release \
  --package-root "$HISTORY_WORKTREE" \
  --release-metadata "plugins/trellis/release.json" \
  --runtime-store "<caller-supplied-absolute-posix-path>"
"$HISTORY_WORKTREE/plugins/trellis/bin/release-contract.mjs" \
  --package-root "$HISTORY_WORKTREE" --check
```

The job runs only after the tag step succeeds or reports the same tag/commit.
Release phase resolves only `refs/tags/trellis-v<VERSION>` and therefore
validates the exact row already present in the staging worktree, whether the
append command just materialized it or authoritative main already contained
it. Product check runs against that same row. A failed append, release
validation, or product check leaves the destination ledger either at the
captured authoritative bytes or at the complete atomically materialized
candidate, removes the disposable worktree, and creates or updates no branch,
commit, or PR.

On an `appended` outcome after both validations pass, the job re-fetches
`origin/main` and fails with `conflict` if it no longer equals
`HISTORY_BASE_COMMIT`. Otherwise it creates or updates only branch
`release/trellis-v<VERSION>-history` from that exact base commit, commits only
`release/trellis/history.json`, and opens or updates its history-only PR. An
existing branch is reusable only when its history bytes equal the same
canonical next ledger and its other tree bytes equal the captured base; its
observed ref may be updated only by compare-and-swap. Divergent branch bytes,
an unexpected branch-tip change, or another changed path is `conflict`. The
job never pushes `main`. The repository's human ship gate owns merging that
PR.

On `already-recorded`, the clean worktree created from authoritative main
already contains the exact row. The job still runs both validations against
it, then creates no commit or PR and does not touch an old history branch.
Before-merge retries reproduce and revalidate the same canonical candidate;
after-merge retries validate the row from authoritative main and return
`already-recorded`. No retry uses the immutable tag's pre-row ledger to infer
whether the row has merged.

The workflow declares one repository-wide concurrency group
`trellis-release-history` with `cancel-in-progress: false`; every package
version shares it. Therefore only one Trellis post-tag history job can prepare
or refresh a history PR at a time.

### Append-only history

`release/trellis/history.json` follows family `release-history.v1`. Before a
tag, it contains all and only prior `trellis-v<SemVer>` releases. The immutable
tagged checkout is evidence for that pre-tag state, never the authoritative
post-tag ledger. After the tag, in the dedicated staging worktree created from
the captured authoritative main commit,
`plugins/trellis/bin/release-contract.mjs --package-root <repository-root>
--append-release-history --authoritative-history-commit <40-hex>` reads the
ledger only from
`<40-hex>:release/trellis/history.json`, where the workflow has resolved the
full commit from `refs/remotes/origin/main`. It changes no tracked path except
the working copy of `release/trellis/history.json`.

The command verifies that the authoritative commit exists, still equals the
captured `origin/main`, descends from the tag's peeled release commit, and
contains a family-valid ledger. The staging worktree's `HEAD` must equal that
commit and every tracked path other than `release/trellis/history.json` must
equal `HEAD`.

While holding the exclusive lock, the command classifies the destination from
bytes rather than a caller assertion:

- **fresh staging:** the destination is byte-identical to the authoritative
  ledger;
- **post-rename recovery:** authoritative main has no current or later row,
  and the destination is byte-identical to the independently derived complete
  canonical-next ledger; or
- **conflict:** every other destination state.

When authoritative main has no current row, the tagged checkout's ledger must
also be byte-identical to the authoritative ledger before either fresh append
or recovery succeeds. The tagged checkout is never used to decide whether an
already-merged row exists.

The append operation has exactly these outcomes:

| Starting state | Result |
|---|---|
| Authoritative main ledger has no current-version row or later-version row; the clean staging destination and tagged pre-row ledger equal its bytes; prior ledger/tag/reference checks pass | Materialize one complete appended row atomically in the staging worktree; stdout one canonical JSON line with `outcome: appended`; exit `0` |
| Authoritative main ledger has no current-version row or later-version row; the staging destination already equals the independently derived canonical-next bytes after a post-rename interruption; tagged prior and all bindings remain valid | Validate the existing complete destination, perform no destination rewrite, remove only safe stale temporary files as defined below, stdout the same canonical `outcome: appended` line; exit `0` |
| Authoritative main ledger and clean staging destination contain one byte-identical current-version row, all following rows remain family-valid, and the tag still peels to its recorded commit | No write; stdout one canonical JSON line with `outcome: already-recorded`; exit `0`, even though the immutable tagged checkout still has the pre-row ledger |
| Current-version row differs, duplicates exist, a later row exists without the current row, authoritative/tagged prior bytes differ before append, the main ref moved during preparation, tag identity differs, cleanup ownership differs, or a required platform primitive is unavailable | No destination write; only safe stale temps removed before a later conflict may remain removed; nonzero exit with `conflict`; never repair or replace |

Success stdout is exactly one LF-terminated line:
`{"outcome":"<appended|already-recorded>","package_version":"<version>","release_tag":"<tag>","source_commit":"<40-hex>"}`.
No failure emits a success object.

History is single-writer data. Only this command may write its working-copy
path, and every invocation must first open
`<git-common-dir>/trellis-history.append.lock` (where `<git-common-dir>` is the
absolute result of `git rev-parse --path-format=absolute --git-common-dir`),
acquire one nonblocking OS-enforced exclusive lock on that open descriptor,
and hold it through temporary-file cleanup and rename. The stable lock file is
Git metadata, not a worktree path. Lock acquisition failure is `conflict`; the
lock is released by the OS when the process exits. Together with the workflow
concurrency group, this is the required single-writer precondition: no
conforming writer can change the destination between comparison and rename.

Before cleanup or destination write, the platform must provide the numeric
effective UID, directory-relative no-follow list/stat/unlink operations,
stable device/inode identity, exclusive-create/no-follow, the nonblocking
exclusive lock, and same-filesystem atomic rename semantics required below. An
unsupported or unenforceable primitive is `conflict`; the command removes no
temporary file, does not change the destination, and emits no success object.

After acquiring the lock and passing that platform check, the command may
remove stale matching unique temporary files only through one already-open
no-follow directory descriptor for the exact `release/trellis` parent. The
descriptor's numeric owner UID must equal the running process's numeric
effective UID before cleanup or append continues. A cleanup candidate's
basename must match exactly
`^\.history\.json\.trellis-append\.[0-9a-f]{32}\.tmp$`.
Directory-relative no-follow metadata immediately before unlink must prove
that the candidate:

- is a regular file, not a symbolic link;
- has numeric owner UID equal to both the running process's numeric effective
  UID and the parent directory's numeric owner UID; and
- has the same device, inode, type, and owner UID as its first no-follow
  observation under the lock.

Comparison is by numeric UID only; username, group, mode-derived writability,
and temporary-file contents never establish ownership or authority. Cleanup
first enumerates basenames in bytewise ascending order and validates the
complete candidate set without unlinking. Any list or first-pass predicate
failure is `conflict` and removes no candidate. It then revalidates each
candidate's device, inode, type, and UID against its first observation and
uses directory-relative unlink against that same open parent descriptor. A
second-pass mismatch, disappearance, replacement, or unlink error is
`conflict`; a previously unlinked safe candidate remains removed, the failing
candidate is never treated as authority, and the destination ledger remains
unchanged.

After safe cleanup, the command re-reads authoritative main and independently
derives and validates the complete canonical-next ledger bytes. If main
contains the exact row, it returns `already-recorded`. If main is still
pre-row:

- from fresh staging, it writes canonical-next to a unique sibling path
  `release/trellis/.history.json.trellis-append.<32-lowercase-hex>.tmp` created
  as a new regular file with exclusive-create/no-follow semantics, re-reads and
  re-hashes the destination, and performs one atomic same-filesystem rename
  only while the destination still equals the authoritative/tagged starting
  bytes; or
- from post-rename recovery, it verifies the existing destination equals the
  independently derived canonical-next bytes, creates no new temporary file,
  performs no destination rename or rewrite, and returns `appended`.

It never uses a shared temporary path. Any other digest is `conflict`. A
pre-rename interruption leaves the destination byte-identical to authoritative
main and may leave only a safely cleanable unique temporary file; a
post-rename interruption leaves the complete candidate and is the sole
recovery prestate. The same candidate is release-validated before the history
PR is opened or updated. Thus neither a crash nor the immutable tag's pre-row
ledger can recreate a merged history PR.

The append commit may follow the tagged release commit, changes no plugin byte,
and must exist at a stable reference before Stewards publication.

The initial `0.1.0` adoption has no prior `trellis-v*` package release. Its
history seed ignores unprefixed binary tags and records the complete initial
public contract, payload identity, and human approval required by the family
contract.

## Stewards adoption and legacy-catalog handoff

After release phase and the history append, this command emits canonical JSON
to stdout, emits no other stdout bytes, and writes nothing:

```text
plugins/trellis/bin/release-contract.mjs \
  --package-root <repository-root> \
  --emit-stewards-adoption
```

The top-level object and every nested object have
`additionalProperties: false` and these exact properties:

| Property | Exact contract |
|---|---|
| `schema_version` | Integer `1` |
| `subject` | Object with exactly `plugin_id: "trellis"`, `repository: "kodhama/trellis"`, authority `package_version`, computed `release_tag`, peeled `source_commit`, and `family_contract_version: 1` |
| `product_adoption` | Exact complete Stewards product-adoption row below |
| `release_metadata_reference` | `repo-path` stable reference to `plugins/trellis/release.json` at `subject.source_commit` |
| `surface_contract_reference` | `repo-path` stable reference to `plugins/trellis/surfaces.json` at `subject.source_commit` |
| `release_history_reference` | `repo-path` stable reference to the post-tag `release/trellis/history.json`; its final row binds the exact subject |
| `release_approval_reference` | The exact family stable reference in `release.json.release_approval` |
| `resolved_legacy_elements` | Exact sorted array below |

`product_adoption` is directly insertable as the Trellis row in Stewards
`distribution/product-adoptions.json` and contains exactly:

| Property | Exact value |
|---|---|
| `plugin_id` | `trellis` |
| `repository` | `kodhama/trellis` |
| `state` | `complete` |
| `standing_decisions_to_reconcile` | `["decision-0036","decision-0043"]` |
| `ownership_changes` | `["behavioral-evidence:trellis-owned","hook-fallback-live-rule-refresh:trellis-owned","package-release-identity:trellis-semver-and-prefixed-tag","payload-identity:trellis-content-stamp-retained","setup-refresh-remove:trellis-owned","surface-support:trellis-owned"]` |
| `adoption_decision` | `repo-path` stable reference to approved `decisions/0059-family-plugin-release-and-surface-contract.md` at `subject.source_commit` |

Every `repo-path` reference contains exactly `kind`, `repository`,
`source_commit`, `path`, and raw-file `sha256` per the family stable-reference
grammar. The four repository references use `repository: "kodhama/trellis"`;
the first three release-tree references use the peeled tag commit, while the
history reference uses the merged history commit. A path, digest, commit,
approval, version, tag, or final-history-row mismatch fails emission.

`resolved_legacy_elements` is exactly
`["canonical-semver-authority","immutable-release-tag","product-adoption-decision","version-bound-surface-contract"]`.
The emitter fails until all four are proven by the referenced bytes.

Canonical output uses NFC strings, Unicode-code-point object-key order,
lexicographically sorted set arrays, no insignificant whitespace, the family
JSON escaping rules, and exactly one terminal LF. Two invocations at the same
repository state shall emit byte-identical output.

This output is a handoff, not a Trellis claim that the catalog is published or
verified. Stewards alone adopts or delists the legacy Claude row, records
catalog availability, and removes its transition-stock row. Trellis validation
shall accept product-support claims only from the closed structured corpus
above; the adoption output, legacy mutable catalog entry, provisioner result,
and clean-install evidence are not members of that corpus.

## Acceptance criteria

### GWT scenarios

**S1 — one authority and two carriers**

- **Given** a clean initial release tree,
- **When** pre-tag validation runs,
- **Then** it extracts `0.1.0` from `VERSION`, both host manifests, and
  `surfaces.json`, and fails if any one is missing or differs.

**S2 — distinct payload identity**

- **Given** matching package metadata and a valid content-derived payload stamp,
- **When** inventory validation runs,
- **Then** it records SemVer, expected tag, commit source, and payload stamp as
  separate identities and rejects substitution between them.

**S3 — current support boundary**

- **Given** the initial six canonical surface rows,
- **When** surface and extension validation run,
- **Then** only Claude local session-load and trusted local Codex fresh startup
  are supported, while every decision-0058 excluded claim remains non-support
  with a missing capability and disclosure.

**S4 — no evidence promotion**

- **Given** a fixture, catalog record, provisioner result, or proof for another
  lifecycle event, host, environment, mode, client, or release,
- **When** a non-supported Trellis claim is validated,
- **Then** validation rejects promotion and does not reuse the proof.

**S5 — deterministic inventory and derivatives**

- **Given** a valid release tree,
- **When** inventory/generation run twice and check mode follows,
- **Then** every tracked package/public path matches the exhaustive contract
  map, the three Markdown carriers and any manifest `/support` match the closed
  grammar, the normalized inventory/raw-install-script bindings converge to
  identical generated bytes, and changing one source causes check mode to name
  every stale derivative without writing.

**S6 — immutable tag**

- **Given** a merged, approved, clean release commit,
- **When** the release workflow runs,
- **Then** it creates `trellis-v<VERSION>` at that commit, reruns idempotently
  for the same commit, fails if the tag peels elsewhere, and its
  `post-tag-release` job materializes or observes the exact row before the
  family release-phase invocation and changes the history branch only after
  that invocation and product check pass.

**S7 — historical namespace isolation**

- **Given** existing unprefixed binary-era `v0.x.y` tags,
- **When** plugin tag and history validation run,
- **Then** those refs remain untouched and do not satisfy or collide with a
  `trellis-v<version>` release.

**S8 — append-only history**

- **Given** a valid Trellis package tag, its immutable pre-row ledger, and the
  captured authoritative `origin/main` ledger,
- **When** history staging and release validation run, are interrupted at
  either side of the atomic rename, are repeated before PR merge, or are
  repeated after PR merge,
- **Then** it yields exactly one complete family-valid row or the exact
  release-validated `already-recorded` no-op from authoritative main,
  treats a canonical-next recovery destination as an `appended` no-rewrite,
  preserves prior bytes/tag bindings under the exclusive single-writer and
  numeric-ownership cleanup protocol, fails closed when those primitives are
  unavailable, never publishes an unvalidated history branch, never recreates
  a merged history PR, and never mutates the plugin package.

**S9 — bump, approval, and adoption**

- **Given** the prior/current public-contract inventories and surface rows,
- **When** release validation and the post-tag adoption emission run,
- **Then** it derives the family change set, rejects an insufficient SemVer
  bump, requires the projection-bound product-human approval, and emits the
  closed canonical adoption object with verified stable references without
  making an availability or support claim.

**S10 — explicit POSIX runtime and product protocol**

- **Given** a supported POSIX launcher, release metadata bound to a matching
  immutable runtime digest, and an explicit valid runtime-store root,
- **When** either family product-validation phase runs,
- **Then** the engine resolves only that digest object, invokes the Trellis
  product-extension identity twice through the exact v2 request/result
  protocol, and accepts it only when both executions return the same canonical
  passing result without an audit or side-effect failure.

**S11 — runtime boundary fails closed**

- **Given** a missing runtime-store argument, an absent or invalid digest
  object, a platform mismatch or drift, unavailable enforcement, Windows, or
  another unsupported launcher result,
- **When** family product validation starts,
- **Then** it rejects the missing argument through the command parser or emits
  the exact applicable family-v2 runtime diagnostic before validator spawn,
  and does not search, fetch, default, substitute, or continue.

**S12 — acyclic whole-plugin checksum**

- **Given** an inventory whose raw install-script source digest and fingerprint
  match the final raw `install.sh`,
- **When** bundle generation, bundle verification, and release check compute
  their respective bindings,
- **Then** it zero-normalizes exactly those two fields, raw-hashes every other
  plugin path, release-checks both raw install-script fields, and fails for any
  omitted path, extra normalization, or changed byte without requiring a
  recursive SHA-256 fixed point.

**S13 — row-before-release validation and retries**

- **Given** a valid package tag and a captured authoritative main ledger that
  either lacks or already contains the exact current row,
- **When** the post-tag job stages, validates, and is retried before or after
  the history PR merges,
- **Then** family release validation always sees the exact appended row, only a
  validated `appended` result may update the history branch, an exact merged row
  validates and returns `already-recorded` without a PR, an exact post-rename
  recovery candidate returns `appended` without a second rename, and any main,
  tag, candidate, ownership, platform, or branch divergence returns
  `conflict`.

### EARS requirements

- **R1:** Trellis shall use `plugins/trellis/VERSION` as its sole package
  SemVer authority.
- **R2:** Both host manifests shall carry and extract exactly the authority
  value.
- **R3:** The inventory provider shall discover every host manifest and shall
  reject a manifest/carrier completeness mismatch.
- **R4:** Trellis shall compute package tags only as
  `trellis-v<VERSION>`.
- **R5:** Trellis shall keep package version, package tag, peeled source
  commit, and payload stamp separate.
- **R6:** The payload stamp shall remain content-derived and shall continue to
  drive setup-copy and file-to-file staleness behavior.
- **R7:** When rendered payload bytes or their stamp change, Trellis shall
  require a package release.
- **R8:** Release metadata, inventory, history, surfaces, and extensions shall
  conform to family contract version `1`.
- **R9:** The product inventory provider shall be deterministic,
  network-disabled, non-mutating, equal to the exhaustive contract
  id/path/extractor map, and compatible with the exact acyclic bundle preimage.
- **R10:** Trellis shall carry each family registry-v1 surface exactly once.
- **R11:** Only Claude local session-load and trusted local Codex fresh startup
  shall be supported in the initial surface contract.
- **R12:** Candidate/excluded claims shall name their missing capability and
  user-visible non-support disclosure.
- **R13:** Trellis shall not infer product support from catalog publication,
  provisioner success, clean installation, or package version.
- **R14:** Evidence and support records shall bind exactly one package release
  and surface and shall not flow across decision-0058 boundaries.
- **R15:** The product extension validator shall require the exact transport,
  boundary, excluded-claim, missing-capability, and disclosure literals and
  shall fail any unsupported promotion or schema divergence.
- **R16:** Only the five closed structured support carriers shall make
  release-valid product support claims, and each shall derive from the exact
  version-bound surface contract.
- **R17:** Check mode shall name every stale derivative and shall write
  nothing.
- **R18:** The whole-plugin vendoring manifest shall cover every shipped
  release/surface file exactly once through raw SHA-256 except for the one
  normalized inventory entry and shall fail on file-set, raw-field, normalized-
  field, or digest drift.
- **R19:** Product validation shall preserve existing payload,
  self-application, setup/remove, and host-isolation guards.
- **R20:** The release workflow shall use the exact pre-tag and release
  invocations, and `post-tag-release` shall atomically stage or observe the
  history row before release-phase validation and shall own the validated
  history-only branch/PR.
- **R21:** The release workflow shall be idempotent at the same commit and
  shall fail rather than move, reuse, or force-push a tag.
- **R22:** Trellis shall leave every unprefixed binary-era tag and frozen
  binary untouched.
- **R23:** The product-human release gate shall approve the family comparison
  and a sufficient bump before tagging.
- **R24:** Release history shall preserve prior rows/tag bindings and shall
  derive repeat state from the captured authoritative main ledger and shall
  use the clean-staging/exclusive-single-writer/unique-temp/atomic-rename
  protocol to return exactly `appended`, `already-recorded`, or `conflict`
  without mutating the tagged plugin tree or publishing an unvalidated row;
  cleanup shall require numeric effective-UID/parent/candidate equality plus
  stable no-follow device/inode identity and shall fail closed when any
  required platform primitive is unavailable.
- **R25:** The adoption emitter shall emit the closed canonical object and
  shall fail until every required stable reference and legacy element resolves.
- **R26:** The adoption emitter shall preserve Trellis ownership of product
  behavior and shall make no Stewards-owned availability claim.
- **R27:** The implementation shall not add, specify, or absorb Stewards
  provisioner behavior.
- **R28:** `release.json` shall declare
  `extensions.trellis.validator` and its lowercase 64-hex
  `validator_runtime_sha256` together and shall bind that digest into the
  versioned public release contract.
- **R29:** Every family `validate-product` invocation shall supply exactly one
  explicit normalized absolute POSIX `--runtime-store` path and shall not use
  environment, default, package-relative, upward-search, fetch, or fallback
  discovery.
- **R30:** Before validator spawn, the family engine shall apply its exact v2
  POSIX platform mapping, content-addressed runtime-object verification, and
  distinct unsupported/unavailable/malformed/platform-mismatch/drift/
  enforcement-unavailable diagnostics.
- **R31:** The common engine shall invoke the Trellis product-extension
  identity twice through the exact family-v2 argv, working-directory,
  canonical request/result, environment, sandbox, audit, limit, exit, and
  no-side-effect contract.
- **R32:** The Trellis validator shall accept only its exact
  `product-extension` request, shall preserve every decision-0058 support row
  and exclusion, and shall emit only a deterministic family-v2 pass or
  product-owned fail result without making or absorbing a distribution-
  availability claim.
- **R33:** The bundle checksum for `release-inventory.json` shall be SHA-256 of
  the family-canonical inventory plus LF after replacing only the
  `trellis.install.vendor-script` row's `source.sha256` and `fingerprint` with
  64 ASCII zeroes each.
- **R34:** Bundle verification shall recompute the normalized inventory digest;
  release generation, inventory validation, and check mode shall independently
  recompute the raw `install.sh` SHA-256, require it as that row's source
  digest, derive the exact family fingerprint from it, and reject every
  unlisted path, other normalized field, or mismatch.
- **R35:** After tag creation, the workflow shall create a clean worktree from
  captured authoritative main, materialize or observe the exact current row,
  and run family release validation and product check against that row before
  creating, updating, or reusing a history branch or PR.
- **R36:** Before-merge retries shall reproduce and revalidate the same
  candidate; after-merge retries shall validate authoritative main's exact row
  and return `already-recorded`; when main remains pre-row, a fresh destination
  shall be atomically renamed once and an exact canonical-next recovery
  destination shall return `appended` without rewrite; any changed main, tag,
  candidate, owner identity, branch tip, or non-history tree path shall fail
  with `conflict`.

## Open questions

None.

## Rubric check

Self-checked against `core/rubrics/artifact-contract.md`, the configured
closest rubric; Trellis has no dedicated spec-quality rubric.

| Check | Result | Evidence |
|---|---|---|
| 1. Required frontmatter | PASS | `id`, `type`, `status`, `version`, `depends_on`, `implements`, `owner`, and `rubric` are present and typed. |
| 2. Type and lifecycle | PASS | `type: spec`; `status: approved` records the maintainer's standing authorization to merge this family wave after v3's independent spec-adversary and conformance gates passed. |
| 3. Unique id | PASS | Repository scan finds no other `spec-0008`. |
| 4. Dependencies resolve | PASS | Append-only decisions are unpinned; local `spec-0007@v1` and approved external family spec `@v2` use the version-pin grammar and resolve. |
| 5. Directional flow | PASS | Direct upstreams are approved or gated, never draft. |
| 6. Required sections and grammars | PASS | `## Acceptance criteria` and `## Open questions` exist; S1–S13 are GWT and R1–R36 are EARS `shall` statements. |
| 7. Supersede integrity | N/A | This revise-in-place spec has no successor or supersession marker. |
| 8–11. Typed catalog/profile checks | N/A | This artifact is neither a signature catalog nor an expression profile. |
| 12. Version semantics | PASS | Testable bundle and release/history clauses changed, so the behavioral counter advances from v2 to v3 and the new whole-spec delta note identifies scope and provenance while retaining v2's note. |
| Decision boundary | PASS | Requirements derive from decision 0059, decision 0058/spec 0007 current behavior, and approved family spec v2; the acyclic digest is a product-local way to satisfy both whole-plugin obligations, the staged row satisfies the family release prerequisite, and runtime provisioning, provisioner behavior, and unsupported promotions remain excluded. |
| Adversary exactness | PASS | Literal support/fallback values, exhaustive public-contract map, exact two-field normalization plus raw recomputation, explicit runtime-store and validator protocol, row-before-release ordering, byte-derived fresh/recovery states, numeric effective-UID/parent/inode cleanup with unsupported-platform failure, enforced single writer, closed adoption object, and bounded support corpus have executable pass/fail boundaries. |

**Result: PASS.**

## Gate record

On 2026-07-24 the maintainer approved v1 after the spec-adversary returned
`APPROVE-READY` and the conformance reviewer returned `PASS`. The maintainer
then authorized this family wave to merge when its independent gates passed.
V2 received spec-adversary `APPROVE-READY`, conformance `PASS`, and a
change-scoped corpus `PASS`; recording `approved` here records that maintainer
act for the family validator-runtime amendment rather than reusing v1's
approval. V3's first spec-adversary pass returned
`NEEDS-REVISION` on `afbf022` because fresh staging contradicted post-rename
recovery and stale-temp ownership/platform behavior was undefined; this
revision makes those states and predicates explicit. The second
spec-adversary pass returned `APPROVE-READY` on `5a9362c`, and the independent
conformance review returned `PASS` against decision 0059 and the approved
family contract. Recording `approved` now records the maintainer's standing
authorization to merge this family wave once its independent gates pass.

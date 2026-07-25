---
id: decision-0063
type: decision
status: gated
depends_on: [decision-0058, decision-0061, decision-0062, kodhama/kodhama-0021-separate-adoption-posture-from-support]
owner: agent
updated: 2026-07-26
---

# 0063 — permit Codex preview adoption through the Stewards catalog

## Decision state

**Decided** (agent shaping from maintainer direction, 2026-07-26):

- Trellis permits explicit preview adoption of its Codex plugin through the
  Stewards marketplace.
- The catalog disclosure says plainly that the preview listing makes no
  support claim and sends readers to Trellis's product-owned boundaries.
- Catalog availability changes no Trellis behavior, package version, surface
  metadata, or support claim.

**Open** (0):

- None.

**Parked** (2):

- GitHub Actions behavior and support evidence remain in Trellis issue #182
  until the exact acquisition route and live host credentials are available.
- Resume, compact, subagent, desktop, IDE, automation, and cloud promotion
  remain under decision-0058's evidence gates.

## Context

Decision 0062 received the approved Stewards adoption-posture strategy while
deliberately making no Trellis product choice. The concrete need has now
arisen: Stewards can register its Codex marketplace in ephemeral or local
homes, but its Codex catalog does not expose Trellis.

Trellis already owns a structurally valid dual-host package at version 0.2.0.
Decision 0058 and spec-0007 support only the trusted local Codex fresh-start
boundary, while decision 0061 records that marketplace registration for that
package is not yet evidenced. Broader Codex lifecycle and hosted surfaces
remain unproven.

The Stewards strategy defines preview as explicit opt-in use by a real
consumer that accepts instability. It also permits a host-valid preview
package to be catalog-listed before it is supported everywhere, provided the
listing points to the product-owned source and discloses the limitation.
Catalog presence and installation do not themselves establish support.

## Decision

Trellis permits its current Codex package — version 0.2.0 when this decision is
approved — to enter the Stewards Codex catalog for **preview** adoption.

The catalog entry points directly to the product-owned
`kodhama/trellis` repository and `plugins/trellis` package path. Its
description must disclose both sides of the boundary:

> Preview — Trellis governance with live project rules. This catalog listing
> makes no support claim; consult Trellis product documentation for exact host
> and surface boundaries.

Preview is the adoption posture for this catalog route, not a package version,
release tier, surface-schema value, or support promotion. The existing
`codex-cli-local-startup` behavior claim remains supported exactly as recorded
by decision 0058, spec-0007, and `plugins/trellis/surfaces.json`. No other
Codex surface gains a claim. The catalog copy deliberately does not restate
that separate product claim; the product-owned package README and
`surfaces.json` remain authoritative for it.

Preview use has a two-part practical rollback at the current Codex host
boundary:

1. in every project where setup was applied, `/trellis:remove` removes the
   project overlay and both host-managed instruction blocks under the existing
   product-wide remove contract; and
2. `codex plugin remove trellis@kodhama` removes the installed plugin from
   Codex local configuration and cache.

The second step is the host-native uninstall operation; the first is not
represented as uninstalling the plugin. Removing the shared `kodhama`
marketplace registration is optional and must not be suggested when other
installed family plugins still use it.

This decision authorizes one separate Stewards catalog-admission change and
its catalog validation. It does not change the Trellis package, `VERSION`,
host manifests, installed payload, setup/refresh/remove behavior, or
`surfaces.json`; none of those facts changed. It also does not authorize a
tag, release, synchronized version, support derivation, or hosted behavioral
test.

Trellis issue #182 remains the product-owned GitHub Actions validation. It may
consume the catalog route after admission, but it is not a prerequisite for
preview listing and cannot infer support merely from successful acquisition.

Claude delivery and the existing Stewards Claude catalog entry are unchanged.

## Alternatives considered

- **Dogfood-only catalog use:** rejected because the intended route is
  explicitly available to opt-in consumers outside family-maintenance
  repositories. Trellis may still dogfood that same route in its own
  repository without changing the catalog's broader preview disclosure.
- **Wait for GitHub Actions support evidence:** rejected because the Stewards
  strategy deliberately separates honest preview distribution from support
  promotion. Issue #182 remains necessary before Trellis claims that hosted
  surface.
- **Describe the whole Codex plugin as unsupported:** rejected because it
  would contradict the product-owned, already-evidenced local fresh-start
  claim. The catalog disclosure instead makes no support claim of its own and
  points to the separate product-owned boundaries.

## Consequences

- Codex users can opt into the real Trellis package through the shared
  marketplace with an explicit preview warning.
- Preview consumers have an explicit project-cleanup plus plugin-uninstall
  rollback rather than a misleading claim that either step does both.
- Trellis can exercise the same acquisition route in later dogfood or hosted
  tests without misrepresenting catalog availability as behavior evidence.
- The public support boundary stays narrow and product-owned.
- Stewards owns the thin catalog edit; Trellis retains every behavior,
  evidence, and promotion decision.

## Self-check

The artifact has unique decision identity and all required frontmatter and
body sections. Every `depends_on` edge resolves to an approved local artifact
or the registered `kodhama/` cross-repository namespace; no draft upstream is
consumed. The choice resolves decision 0062's deliberately parked local
follow-up without copying or redefining the Stewards strategy. It authorizes
one catalog edge, preserves decision 0058's exact support boundary, and
creates no package, release, schema, or hosted-behavior claim.

## Lifecycle record

The maintainer reviewed the new family strategy and directed the agreed task
list to proceed on 2026-07-26. That direction authorized shaping this canvas;
it did not ratify unseen wording. The shaper resolved the dogfood-versus-preview
choice in favor of preview from the intended external opt-in audience, recorded
the rejected alternatives, found no open item, self-checked the artifact
against the corpus rubric, and moved it to `gated`. The exact decision still
requires independent soundness review and the human intent act required by
Trellis's gate profile. The first independent review returned
`NEEDS-REVISION`: the proposed preview copy made a positive support promise,
and project cleanup alone did not uninstall the plugin. This revision makes
the listing itself claim no support and separates product cleanup from the
verified host-native plugin-removal command.

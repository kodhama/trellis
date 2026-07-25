---
id: decision-0063
type: decision
status: draft
depends_on: [decision-0058, decision-0061, decision-0062, kodhama/kodhama-0021-separate-adoption-posture-from-support]
owner: agent
updated: 2026-07-26
---

# 0063 — permit Codex preview adoption through the Stewards catalog

## Decision state

**Decided** (agent shaping from maintainer direction, 2026-07-26):

- Trellis permits explicit preview adoption of its Codex plugin through the
  Stewards marketplace.
- The catalog disclosure names the exact already-supported local boundary and
  says that broader Codex surfaces are not supported.
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

Trellis permits its version-0.2.0 Codex package to enter the Stewards Codex
catalog for **preview** adoption.

The catalog entry points directly to the product-owned
`kodhama/trellis` repository and `plugins/trellis` package path. Its
description must disclose both sides of the boundary:

> Preview — Trellis governance with live project rules. Support is limited to
> trusted local Codex fresh starts; resume, compact, subagents, IDE,
> automation, and cloud are not supported.

Preview is the adoption posture for this catalog route, not a package version,
release tier, surface-schema value, or support promotion. The existing
`codex-cli-local-startup` behavior claim remains supported exactly as recorded
by decision 0058, spec-0007, and `plugins/trellis/surfaces.json`. No other
Codex surface gains a claim.

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

## Consequences

- Codex users can opt into the real Trellis package through the shared
  marketplace with an explicit preview warning.
- Trellis can exercise the same acquisition route in later dogfood or hosted
  tests without misrepresenting catalog availability as behavior evidence.
- The public support boundary stays narrow and product-owned.
- Stewards owns the thin catalog edit; Trellis retains every behavior,
  evidence, and promotion decision.

## Lifecycle record

The maintainer reviewed the new family strategy and directed the agreed task
list to proceed on 2026-07-26. That direction authorized shaping this canvas;
it did not ratify unseen wording. The exact draft still requires independent
soundness review and the human intent act required by Trellis's gate profile.

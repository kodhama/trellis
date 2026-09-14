---
id: decision-0019
type: decision
status: ratified
depends_on: [decision-0001, decision-0004, decision-0010, decision-0012, research-0007]
owner: gundi
date: 2026-07-04
ratified: 2026-07-04
---

# 0019 — Free & open, MIT-licensed; advisor-CLI as the v0 on-ramp

**Raised by:** the maintainer — *for the landing page to point somewhere, licensing and the delivery
default have to be settled. Lean: just free to use.*

## Context

- Trellis **ships as instructions** (`decision-0010`), so the core is trivially copyable — you cannot
  effectively paywall the text. Monetization, if ever, is **services around it**, never gating it.
- At **N=1** (`decision-0009`), the scarce resource is **adoption** (more instances → validation of
  the invariants), not revenue. Free maximizes exactly what the project most needs.
- The **advisor / reference** on-ramp is free by nature (pointing agents at an open repo).
- **Peer norm is uniform:** the closest comparables — spec-driven / agentic-dev methodology packs that
  ship as instructions + a CLI, like Trellis — are all **MIT**: BMAD-METHOD, GitHub Spec Kit, OpenSpec.
- **Relicensing is cheap before adoption:** as the sole copyright holder, the maintainer can release
  future versions under any license (MIT→Apache and back). The only permanent thing is that an
  already-published version stays available under its license — moot pre-adoption. Freedom erodes only
  once external contributions land without a sign-off.

## Decision

1. **Free & open. No monetization now.** Trellis is free to use.
2. **License: MIT.** Matches the niche norm, maximally simple, ~nil patent risk for an
   instructions+CLI pack. A `LICENSE` (MIT) lands with this decision. *(Copyright line uses the
   maintainer's handle for now; swap for a legal name/entity if one is formed.)*
3. **Keep the Apache-2.0 upgrade path open.** If an enterprise / open-core future makes Apache-2.0's
   **patent grant + trademark clause** worth having, upgrade then — it is cheap while sole-owner. Add
   a **DCO** (one-line `Signed-off-by`) the day external contributions start, to preserve that freedom.
4. **Monetization, if ever, is open-core *services*** — a managed/hosted supervisor, hosted
   conformance, cross-instance analytics, compliance — never gating the core text. Consistent with
   buyer-neutral (`decision-0004`).
5. **Delivery v0 = advisor mode via a host-invoked CLI.** Going free resolves the `research-0007`
   fork (which was half-driven by monetization): the lowest-friction **advisor / reference** path,
   triggered through a **CLI the host invokes**, is the first on-ramp; the installed **supervisor** is
   the heavier, later thing (and the natural home for any future paid services). Sharpens
   `decision-0012` (plugin / CLI / git-copy → the **CLI-advisor path is the v0 deliverable**).

## Consequences

- **The landing page** drops the paid tiers → *free & open*; licensing points to **MIT**.
- **`decision-0012` refined** — CLI-triggered advisor is the v0 delivery target; supervisor deferred.
- **README** already states intended open-core; update the license line to **MIT (this decision)**.
- Monetization stays a *future, services-only* option — never a paywall on the invariants.

## Open questions

- **Copyright holder form** — handle now; a legal name or entity later (for a formal LICENSE / any
  future commercial arm)?
- **The Apache-upgrade trigger** — what concrete enterprise signal flips it (a paying design partner,
  a compliance requirement)?
- **DCO vs CLA** when contributions start — DCO is lighter and usually enough; revisit if a commercial
  arm needs stronger relicensing rights.

## Supersedes / superseded by

— (none; refines `decision-0012` delivery)

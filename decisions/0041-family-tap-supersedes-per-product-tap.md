---
id: decision-0041
type: decision
status: ratified
depends_on: [decision-0032]
owner: gundi
ratified: 2026-07-07
---

# 0041 — the tap is the family's, not trellis's own; supersedes decision-0032's tap shape

> **Superseded in part by `kodhama-0007` rule 5 / `decision-0043` (2026-07-10, #120; text below
> preserved as written).** Trellis's brew channel retired: the `kodhama/homebrew-tap` formula is
> deprecated (pinned at `v0.2.29`, pointing at the plugin + manual copy paths) and this repo's
> release/tap-dispatch machinery is removed. The family-tap *shape* this decision established
> stands for any future kodhama product that ships a real binary — only trellis's own formula and
> its formula-sync mechanics are retired.

## Context

`decision-0032` stood up `gundisalwa/homebrew-trellis` as trellis's own Homebrew tap — a
reasonable call when trellis was the only kodhama-family product shipping a binary. Since then
the family formalized how it delivers as a whole: `kodhama-0001-family-delivery` (ratified in
`kodhama/kodhama`, the org meta repo, migrated from a math-quest discovery pass) found that
**one shared org-level tap serving many product repos is established practice** (F1: Charm's
`homebrew-tap`, 15+ formulas; F2: comparable-scale tool families run this way) and decided the
kodhama family follows it: `kodhama/homebrew-tap`, one tap, N formulas, each product's own
release pipeline pushes its own formula — no product owns its own tap repo.

`decision-0032`'s formula-sync mechanics (regenerated on release, never hand-edited, dispatch
via a fine-grained PAT, `curl … | sh` as the no-Homebrew fallback) are unaffected and still hold
— only the tap's *address* and *ownership* changes.

## Decision

- **The tap moves**: `gundisalwa/homebrew-trellis` → `kodhama/homebrew-tap`. Install becomes
  `brew install kodhama/tap/trellis` (was `brew install trellis/trellis`).
- **`decision-0032` is superseded**, not deleted or rewritten — its formula-sync reasoning
  (the `decision-0028` derived-resource-with-a-sync-guard pattern, the dispatch-token mechanics,
  the `curl` fallback) still applies verbatim against the new tap address. Only the "the tap is
  `gundisalwa/homebrew-trellis`, a repo this product owns" premise is retired.
- This repo's release workflow and docs already point at the new address as of PR #103
  (`chore/family-tap`, merged 2026-07-07) — this decision is the retroactive record for that
  change, not a spec it still has to satisfy.

## Consequences

- `decisions/0032-homebrew-distribution.md` gets `status: superseded` and a forward pointer to
  this decision — its content is untouched (append-only; the reasoning it recorded was correct
  for its time).
- Any future kodhama-family product adding a tap formula reads `kodhama-0001-family-delivery`
  + this decision, not `decision-0032`, for where the tap lives.
- `TAP_DISPATCH_TOKEN` (decision-0032's PAT) needs Contents:write on `kodhama/homebrew-tap`
  instead of the old `gundisalwa/homebrew-trellis` — an operational follow-up, not decided here.

## Open questions

- Does the existing `TAP_DISPATCH_TOKEN` secret already target the new repo, or does it need
  reissuing? (Operational check, not a decision.)
- `gundisalwa/homebrew-trellis` itself: archive, redirect, or leave as-is? Not decided here —
  a small follow-up once this decision is ratified.

## Supersedes / superseded by

Supersedes `decision-0032`'s tap-ownership premise (its formula-sync mechanics carry forward
unchanged). Not superseded by anything as of this writing.

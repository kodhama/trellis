---
id: decision-0034
type: decision
status: ratified
depends_on: [decision-0026, decision-0031, spec-0002]
owner: gundi
ratified: 2026-07-05
---

# 0034 — The always-loaded block speaks to the host agent: imperative directives, no internal codes

## Context

The always-loaded block (`profile.md` / the inline AGENTS.md block) is Trellis's core lever — its whole
job is to change the host agent's behavior. But it was rendered from the catalog's `what` field, which is
a **dictionary definition** written in Trellis's own vocabulary. A cold-read agent given the block with
**no Trellis context** scored it **2.5/5** and called it *"a values statement transplanted from a spec,
not agent instructions"*: high readability, low actionability. Concretely it flagged:

- **Internal codes it can't resolve** — `A1/A2/A3/B2/C2/D1`, `decision-0024` (implying `0001–0023`
  exist), `intent locus`, `by ratchet` — which *"invite paralysis or performative compliance."*
- **The header describes rather than commands**, and the strength was the jargon phrase `C1 lean:
  default-on-but-skippable`.
- **Not self-contained** — it inlined the rules "so I wouldn't need `.trellis/`," then left the decoder
  ring (definitions/codes) in the file it replaced.

## Decision

1. **Every invariant carries an agent-facing `directive`** — one imperative, self-contained, code-free
   instruction (a verb the agent can execute) — and the block renders **that**, not `what`. The `what`
   dictionary text stays for the on-demand reference (`invariants.md`) and the benefits page.
2. **Imperative header** — "You are working in a project that follows Trellis … **follow the rules below
   as you work here**" — no implementation leakage (`@import`), no "governed by".
3. **The profile's `C1` lean is translated to plain-language strength** — *firmly* / *by default* / *as
   guidance* — not the raw dial word.
4. **No Trellis-internal codes in the block.** A test (`TestInvariantDirectivesCoverCatalog`) fails if a
   directive leaks `A1/B2/C2/decision-0…`.

## Consequences

- The catalog gains a `directive` field per invariant; `spec-0002` §1 and the rubric require it; the
  bundled reference is regenerated. `renderProfile`/`renderInlineBlock`/`renderHeader` rewritten to the
  imperative framing + strength; the block is now self-contained.
- Verified by re-running the same cold-read test on the new block.

## Open questions

- The on-demand `invariants.md` is still dictionary-voiced with codes — clean it if agents actually open
  it (the block no longer depends on it).
- Per-profile strength phrasing could be tuned further once `C1` is actually enforced (supervisor mode).

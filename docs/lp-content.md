# trellis — landing page content

This is trellis's own copy for its generated landing page (`docs/index.html`),
per `kodhama/design-system`'s LP generator contract (`lp-generator.md`). The
design system supplies no copy — everything below is trellis's, extracted
verbatim from the hand-built page this repo already shipped
(`docs/index.html` as of commit `0e3b6df`, the last content edit before this
retrofit).

trellis is a special case among the family's LP derivatives: this page *is*
the design system's source of truth. `kodhama/design-system`'s `tokens.css`
and `patterns.md` were both extracted verbatim from this exact page (see
their own file headers). So this generation is a retrofit, not a fresh
build — composing the DS's tokens/patterns back against trellis's own
content should reproduce the original almost exactly, and it does; see the
parity note in `docs/index.html`'s own top-of-file comment.

## Eyebrow

Governance for agentic development

## Hero

**Title:** The structure your agents **grow along.** (the last two words
carry the `.em` accent-ink emphasis)

**Subtitle:** Trellis is a governance layer for agentic software
development. It fits whatever methodology your project already uses,
teaches it to your coding agents, and enforces a small set of invariants —
so a process glitch never has to happen twice.

**Install block** (terminal pattern, three tabs — Claude Code / curl / manual copy;
the Homebrew tab retired with the end-user binary channel, `kodhama-0007` rule 5 /
kodhama/trellis#120, and the family marketplace is the canonical front door per
`kodhama-0002`. The curl tab returned in kodhama/trellis#124 as a **plugin vendor
script** — a different, much smaller artifact class than the retired binary
installer: it makes exactly one decision, scope, and composes nothing else):

- `cc` (Claude Code, default/active tab):
  ```
  > /plugin marketplace add kodhama/kodhama
  > /plugin install trellis@kodhama
  > /trellis:setup    # the plugin covers the overlay natively
  ```
- `curl` (same plugin, no marketplace — kodhama/trellis#124):
  ```
  $ curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
  $ # vends .claude/skills/trellis (project scope, default) or ~/.claude/skills/trellis
  $ # (--scope personal); then run /trellis:setup as above
  ```
- `manual` (any other harness):
  ```
  $ git clone --depth 1 https://github.com/kodhama/trellis
  $ cp trellis/plugins/trellis/reference/... .trellis/    # copy, paste, shasum -c — see the README
  ```

**Note under the terminal:** No binary, no runtime — the bundle is
pre-rendered plain files with a checksum manifest; the plugin (or you)
just copies and verifies them. Clean exits, always: `/trellis:remove`
clears it from a project, and a bundled session hook tells you when the
overlay is behind the installed plugin.

**CTAs:**
- Primary → `invariants.html` — "Explore the invariants →"
- Ghost → `https://github.com/kodhama/trellis` — "View on GitHub →"

## Section: The problem

**Eyebrow:** The problem
**Heading:** Agents move fast. Without structure, they lose the thread.
**Lede:** Trellis holds the load-bearing rules so your agents can move
quickly without building on shifting ground.

Three cards:

1. **Referential integrity** — Every artifact — research, decisions,
   specs, code — points to the settled ground it depends on. Agents build
   on ratified truth, never a draft that's still moving.
2. **Knowledge flows back** — When a downstream discovery contradicts an
   upstream doc, the doc gets updated — not just the code. Learnings
   propagate instead of forking.
3. **A glitch, once** — Friction becomes a rule where it fires. The same
   process failure doesn't recur every few weeks because nothing captured
   the lesson.

## Section: With vs without

**Eyebrow:** With vs without
**Heading:** The same project, guarded.

Three compare-pairs (`.compare-pairs` pattern — a case label, a "without"
row, a "with" row):

1. **directional flow**
   - Without: an agent codes against a spec that's still being edited; it
     shifts, and the work is built on a version that no longer exists.
   - With: implementation reads only ratified specs; downstream never
     consumes a draft.
2. **the intent gate**
   - Without: a human-gated decision gets merged with no approval —
     silently.
   - With: a human-gated handover performed without its approval is
     **surfaced**, loudly.
3. **self-improvement**
   - Without: the same flaky step fails every week and everyone just
     re-runs it.
   - With: the recurring failure becomes a checkable rule that rides the
     PR you already write.

## Section: How it works

**Eyebrow:** How it works
**Heading:** One command. It reads your project, you choose the fit.
**Lede:** Trellis rides your existing harness (Claude Code today). It
asks, copies, and — only with your go-ahead — verifies itself onto
your project. No runtime, no lock-in.

Four-step flow (`01` – `04`):

1. **01 · install — Add the plugin.** From the kodhama family
   marketplace — or copy the pre-rendered bundle into any harness.
2. **02 · posture — Pick a posture.** Conductor or author-adapt — seeded
   as explicit rows in your `rules.toml`: how strict, and what's active.
   A refresh reads the rows and asks nothing.
3. **03 · mode — Alongside or rewrite.** Overlay next to your rules, or
   — on request — morph them in on a branch.
4. **04 · verify — You approve.** Augment-never-clobber, checked against
   a shipped checksum manifest. Trellis proposes; the merge is yours.

**Repo footprint** (rendered as a small code block, not the terminal
pattern — this is a file-tree illustration, not a shell session):

```
CLAUDE.md          # + a small managed block importing the header + your rules.toml
.trellis/
  rules.toml       # which rules are active, how strictly — yours to edit
  internal/        # generated, refreshed verbatim:
    trellis.md     #   the header your agents read
    rules.md       #   the rules readout — always loaded; your rules.toml rows say which apply
    invariants.md  #   the full why + examples — on demand
```

Label above it: "What it leaves in your repo — small, single-source, and
yours to remove:"

## Section: The core (alt background)

**Eyebrow:** The core
**Heading:** A small set of invariants, expressed at your strength.
**Lede:** Not a process — the layer above it. A handful of load-bearing
invariants (directional flow, ratifiable artifacts, gate-at-handover,
independent judgment, transparency…), each set along two dials: how
strictly it's enforced, and who gates it. Everything else, Trellis
respects.

Two cards:

1. **It grounds out in real artifacts** — Trellis never just *describes*
   process. It produces and enforces concrete, project-specific artifacts
   — a real instructions file, real gates, a real conformance check. If it
   can't check it, it doesn't claim it.
2. **It fits, it doesn't dictate** — Gatekeepers are whatever your project
   already declares — detected and respected, not imposed. Trellis
   enforces the invariants and gets out of the way of your methodology.

Secondary CTA below the cards: ghost → `invariants.html` — "See all
fourteen, with why + examples →"

## Section: Free & open

**Eyebrow:** Free & open
**Heading:** Free, and open. That's the whole pricing page.
**Lede:** MIT licensed — read it, fork it, run it, keep it. If paid
services ever show up (a managed supervisor, hosted conformance), they'll
be services on top. Never a paywall on the rules.

CTA: primary → `https://github.com/kodhama/trellis` — "Get Trellis →"

## Footer

- Left: "Trellis — our synthesis of the invariants, v1. Built with
  Trellis."
- Right: `github.com/kodhama/trellis` (linked) · MIT

## Header / nav (not a named lp-content section elsewhere, noted for
completeness)

- Brand: trellis mark (posts + laths — identical path data to
  `kodhama/design-system`'s `icons/trellis.svg`, since that mark was
  extracted from this page) + wordmark "Trellis"
- Nav links: `#how` ("How"), `invariants.html` ("Invariants"), `#open`
  ("Free & open"), `https://github.com/kodhama/trellis` ("GitHub"), plus
  the theme-toggle button

## Behavior (not copy, but load-bearing — carried over unchanged)

- Theme toggle: flips `data-theme` on `<html>`, persisted to
  `localStorage` under the key `trellis-theme` (already product-namespaced
  per `patterns.md`'s own note on that pattern).
- Terminal tabs: switches the active install-method panel; copy button
  copies the active panel's commands to the clipboard.
- Climbing-plant hero animation: decorative, `prefers-reduced-motion`
  aware — DS `patterns.md`'s "Climbing-plant animation" pattern, used
  as-is (this page is that pattern's origin).

## Out of scope for this retrofit

`docs/invariants.html` is a separate page (the invariants detail page
linked from this one) and is untouched by this lane — only
`docs/index.html` is a DS derivative as of this change.

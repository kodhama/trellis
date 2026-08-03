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
teaches it to your coding agents, and governs a small set of invariants —
so a process glitch never has to happen twice.

**Install block** (terminal pattern, three tabs — Claude Code / curl / manual copy;
the Homebrew tab retired with the end-user binary channel, `kodhama-0007` rule 5 /
kodhama/trellis#120, and the family marketplace is the canonical front door per
`kodhama-0002`. The curl tab returned in kodhama/trellis#124 as a **plugin vendor
script** — a different, much smaller artifact class than the retired binary
installer. It vends the plugin bundle and, on project scope, renders one rules
file it wholly owns; `decision-0068` amended the earlier "composes nothing else"
framing, which is why this copy changed):

- `cc` (Claude Code, default/active tab):
  ```
  > /plugin marketplace add kodhama/kodhama
  > /plugin install trellis@kodhama
  ```
- `curl` (same plugin, no marketplace — kodhama/trellis#124):
  ```
  $ curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
  # that's it — all 14 rules active, adaptive posture
  ```
- `manual` (harnesses the plugin does not cover — **not** Claude Code,
  which the tabs above serve; `decision-0069`):
  ```
  $ git clone --depth 1 https://github.com/kodhama/trellis
  # then follow the copy recipe in the README — this path is
  # for harnesses the plugin does not cover.
  ```

**Note under the terminal:** The curl path is **Claude Code, project
scope**: it vends the plugin *and* renders the rules to
`.claude/rules/`. `--scope personal` vends the plugin but delivers no
rules. No
binary, no runtime — the bundle is pre-rendered plain files with a
checksum manifest, verified before anything is written. Clean exits: `/trellis:remove` clears the rules and config
from a project — on the curl path the vendored bundle under
`.claude/skills/trellis/` is yours to delete — and a bundled session hook delivers the rules on the
plugin path, stands down when the curl path has already delivered them,
and warns if it ever finds both.

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
**Lede:** Trellis rides your existing harness — Claude Code today. The
rules land as plain instructions your agents read, and one small config
file you own says how strictly they apply. No runtime, no lock-in.

*(Codex CLI is deliberately not named. The plugin supports it and the
hook is real, but there is no way to install it there — `trellis#220`.
Naming a host a visitor cannot reach is worse than not naming it.)*

Four-step flow (`01` – `04`):

1. **01 · install — Add the plugin.** From the kodhama family
   marketplace — or, for a harness the plugin does not cover, copy the
   pre-rendered bundle by hand.
2. **02 · posture — A posture you can change.** Every rule active,
   adaptive posture — the shipped default, seeded as explicit rows in
   your `rules.toml`. No file yet? Copy a complete preset —
   `reference/rules-a.toml` for firm, `rules-b.toml` for adaptive.
   Already have one? Edit `strictness` in place; copying a preset over it
   re-enables every row you turned off. Rows govern at read time, and the
   set has to stay whole.
3. **03 · deliver — The rules arrive on their own.** On the plugin path a
   session hook injects them; on the curl path they are rendered into
   `.claude/rules/`. Never both — the hook stands down when it finds the
   rendered file, and says so if it ever sees both.
4. **04 · verify — You approve.** The curl path checks every byte against
   a shipped checksum manifest before it writes one, and never
   overwrites rows you already have — `.trellis/rules.toml` is seeded
   only when it is absent. Trellis proposes; the merge is yours.

**Repo footprint** (rendered as a small code block, not the terminal
pattern — this is a file-tree illustration, not a shell session):

```
.trellis/
  rules.toml       # which rules are active, how strictly — yours to edit.
                   #   Seeded by the installer, all 14 on (decision-0070)
.claude/           # curl path only — the plugin path writes neither:
  rules/
    trellis.md     #   the rules readout, loaded every session — Trellis owns
                   #   this file and replaces it wholesale on re-install
  skills/
    trellis/       #   the vendored plugin bundle
```

Label above it: "What it leaves in your repo — small, yours to edit
where it says so, and yours to remove:" *(dropped "single-source": the
curl path leaves two trees under `.claude/`, so the old label was no
longer true.)*

## Section: The core (alt background)

**Eyebrow:** The core
**Heading:** A small set of invariants, expressed at your strength.
**Lede:** Not a process — the layer above it. Fourteen load-bearing
invariants (directional flow, ratifiable artifacts, gate-at-handover,
independent judgment, transparency…), each set along two dials: how
strictly it applies, and who gates it. Everything else, Trellis
respects.

Two cards:

1. **It grounds out in real artifacts** — Trellis never just *describes*
   process. It grounds out in things you can point at — an instructions
   file your agents actually load, rules you switch on and off by name,
   and a conformance check it runs on itself. If it can't check it, it
   doesn't claim it.
2. **It fits, it doesn't dictate** — Gatekeepers are whatever your project
   already declares — respected, not imposed. Trellis guides your agents
   on the invariants and gets out of the way of your methodology.

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

---
id: research-0010
type: research-note
status: ratified
depends_on: [invariants-v1]
informed_by: [decision-0029]
owner: gundi
ratified: 2026-07-05
---

# 0010 — Agent-instruction file landscape (which files M1 can target)

> *Amended in place 2026-09-06 (`decision-0047`; the 2026-07-13 consumer audit that amended
> seven sibling research notes and `invariants-v1` did not reach this one).
> **WHAT:** `decision-0029` moved out of `depends_on` into `informed_by`. It **commissioned**
> this survey — its own "follow-up (B): M1 detects instruction files and lets you pick the
> target" — and is named below as the consumer of the findings; but what it *ruled* is the
> setup CLI's mode-first ordering of harness detection, and no row of the table and no
> conclusion here would be different or wrong had that ruling gone the other way. Provenance,
> not coupling. It is deliberately **not** re-pointed to `decision-0043`: that record retires
> `decision-0029`'s subject matter rather than carrying it, and a provenance edge records what
> informed this note in July 2026, so it cannot name a record written five days later.
> `invariants-v1` added to `depends_on` as the coupling this note carried undeclared — the
> inline-over-`@import` recommendation rests on D1 (`floor-transparency`), which is what makes
> a silently unresolved import a violation. **POINTER:** `decision-0047`; TRL-57.*

Which file each agentic coding tool auto-loads as its project instructions, so the M1 overlay
(`decision-0029`) knows what to target. The load-bearing distinction is **single instruction file**
(our create-or-append-a-block model works) vs **rules *directory*** (needs a dedicated file written,
not a block appended — a different apply path).

## Findings (researched 2026-07-05)

| Tool | Canonical file / path | Shape | In-file imports? | Conf. |
|---|---|---|---|---|
| Claude Code | `CLAUDE.md` | single file | **yes** (`@path`) | high |
| OpenAI Codex CLI | `AGENTS.md` | single file (dir-hierarchical concat) | no | high |
| GitHub Copilot | `.github/copilot-instructions.md` (also reads `AGENTS.md`) | single file | no (treat inline) | high |
| Google Gemini CLI | `GEMINI.md` | single file | yes (`@file.md`) | high |
| Cline | `.clinerules` **or** `.clinerules/` | single file **or** dir | no | high |
| Devin CLI / Cascade | `AGENTS.md` (also `.devin/rules/`, legacy `.windsurf/rules/`) | single file + dir | no | high |
| Windsurf | `AGENTS.md` + `.windsurf/rules/` (legacy; folded into Devin) | single file + dir | no | high |
| Cursor | `.cursor/rules/*.mdc` (legacy `.cursorrules` deprecated) | **directory** | yes (`@file`) | high |
| Continue.dev | `.continue/rules/` (`AGENTS.md` is only a proposal) | **directory** | no | high |
| Aider | `CONVENTIONS.md` — **not auto-loaded** | single file, opt-in | via config `read:` | high |

**`AGENTS.md` is the cross-tool standard** — auto-read by Codex, Devin/Cascade, Copilot, Windsurf, and
listed compatible by others. A single `AGENTS.md` covers four+ tools; it is our primary portable target.

## Recommendation (what to add now)

Add as **single-file, inline** targets (append/create a block; high confidence):

- `GEMINI.md` — Gemini CLI. *(Gemini supports `@import`, but we inline anyway: inline can't silently
  fail to resolve — D1 — so only `CLAUDE.md` keeps `Imports: true`.)*
- `.github/copilot-instructions.md` — GitHub Copilot (note the `.github/` parent dir).
- `.clinerules` — Cline's single-file form.

Registry after this: `CLAUDE.md` (import) · `AGENTS.md` · `GEMINI.md` · `.github/copilot-instructions.md`
· `.clinerules` (the last four inline).

## Open questions

- **Directory-based tools** (Cursor `.cursor/rules/*.mdc`, Continue `.continue/rules/`, and the richer
  `.clinerules/` / `.devin/rules/` forms) need a **"write a dedicated rule file"** apply path, not a
  block appended to an existing file — a separate follow-up, out of scope for the append-a-block model.
- **Aider** needs both the file *and* a config entry (`read: CONVENTIONS.md` in `.aider.conf.yml`);
  a bare file does nothing. Deferred until the config-writing path exists.
- Whether to later exploit Gemini's / Cursor's native `@import` instead of inlining (sync vs
  robustness trade-off; today robustness wins).

## Sources

- Codex `AGENTS.md`: https://developers.openai.com/codex/guides/agents-md
- Copilot custom instructions: https://docs.github.com/en/copilot/reference/custom-instructions-support
- Gemini `GEMINI.md`: https://geminicli.com/docs/cli/gemini-md/
- Cline rules: https://docs.cline.bot/customization/cline-rules
- Devin CLI rules / `AGENTS.md`: https://docs.devin.ai/cli/extensibility/rules · https://docs.devin.ai/onboard-devin/agents-md
- Cursor rules: https://docs.cursor.com/context/rules
- Continue rules: https://docs.continue.dev/customize/deep-dives/rules
- Aider conventions: https://aider.chat/docs/usage/conventions.html

---
id: decision-2026-09-15-rule-rows-only-switch-rules-off
type: decision
depends_on: [decision-0028, decision-0070, decision-0077]  # coupling under decision-0047's test. "A rule with no row applies" is safe only where adoption is evidenced otherwise, which 0070 D3, D4 and D7 and 0077 rule; 0028 is the pair rule the new-file pin between install.sh and the hook rests on
informed_by: [research-0012, decision-0033, decision-0051, decision-0053, decision-0083, decision-0084, decision-0086, decision-0088]
owner: agent
date: 2026-09-15
---

> **Provenance.** TRL-97 carries out idea 2 of the Simplify Trellis project. The maintainer approved
> the requirements on 2026-09-14 and answered the questions through the project's orchestrator session.
> He chose sparse rows over a new `disabled` list, the "By default" sentence for every project, kept
> `governed = false`, and asked that new files hold only what has effect. On 2026-09-15 he kept the
> sentence each hook adds naming the rules a file switches off, told that its wording is untested and
> that it sits above rows shown unchanged. The same day he chose: any `active = false` row switches its
> rule off; the file is shown only when it and its warnings fit one shared limit; five warnings are
> named in total; the shipped inline block carries no rows; and a symlinked rules file is classified
> but never shown. Those answers reached the author as relayed, and are recorded here as relayed, not
> quoted. The plan is `docs/plans/2026-09-14-1858-refactor-rules-opt-out-list-plan.md`, and the review
> that led to the last four answers is recorded on the TRL-97 pull request.

# Rule rows only switch rules off

## Context

`.trellis/rules.toml` carried three things, and only the rows decided which rules applied. A valid
`strictness` selected one sentence of the delivered text. An invalid one made the Codex hook refuse
the whole file, while the Claude hook read it as adaptive. `decision-0033` records the lean as
"descriptive, not enforced", and the two shipped presets were byte-identical in all sixteen rows.
`seeded_from` had no reader.

The rows had to list every rule the plugin shipped, so every catalog change put every existing file
out of step. `decision-0083` records the cost: "A single unmatched row cost all sixteen rules, every
session, until a human edited the file." The repair was reconciliation, quarantine and a write-back
mandate in the Claude hook. Guards followed for the seven ways a broken payload could reach that
repair, then Codex parity (`decision-0084`) and a byte-budget degradation of the injected copy
(`decision-0086`). On 2026-09-14 nine `kodhama/*` repositories carried the file. Every row in every
one was active, and four were still on the fourteen-row set, reconciling every session.

Switching a rule off is what works. `research-0012` measured it on one rule and one task: flipping one
row from `true` to `false` moved the rate from 95% to 0%.

## Decision

1. **Only a row set `active = false` has effect.** A rule is off when any valid row for it under
   `[rules]` says `active = false`. A rule with no row applies, and an empty file means every rule
   applies. An `active = true` row, a repeated row, `strictness` and `seeded_from` change nothing and
   draw no comment. Switching a rule back on means deleting its false row. `governed = false` keeps its
   meaning (`decision-0070` D5): exactly one uncommented top-level `governed` line, above every table
   header, read before anything else in the file.
2. **The activation rule is stated once, in `rules.md`.** It now reads: apply each rule unless a row
   for it says `active = false`; a rule with no row applies; a rule whose row is `active = false` does
   not apply in this project — do not follow it; the two floor rules always apply while the project is
   governed; nothing else in the file changes which rules apply. The sentence `research-0012` tested
   said to apply a rule "ONLY if its row says `active = true`". Its second sentence, which carries a
   switched-off rule's measured effect, is kept word for word. The clause that retires assumed a row
   for every rule, which a file no longer needs to carry.
3. **A bad entry costs that entry, never the session.** Each of these is ignored with a warning: a
   `false` row for a floor rule, a `false` row naming a slug the plugin does not ship, a row outside
   `[rules]`, a table other than `[rules]`, a malformed row, an unknown top-level key, and a `governed`
   line that does not opt out. A warning names the line number and the slug, key or kind of entry, never
   the raw line. Five warnings are named in total. Entries that try to switch a shipped rule off come
   first, and the rest are counted in one line saying how many of them try to switch a rule off. A
   file containing a NUL byte cannot be parsed at all: it delivers every rule, is not shown, and draws
   one warning. A project file that is not a regular file, cannot be read, or is larger than 1 MiB is
   refused loudly on both hosts, because an opt-out the hook cannot read cannot be ruled out.
4. **Both hooks frame the file identically, and say which rules it switches off.** On the
   plugin-native path, below the rules, the context carries `## Project rule activation` and then one
   sentence the hook computes: "The project file .trellis/rules.toml switches these rules off: …
   Every other rule above applies.", or that it switches no rule off. The file's own rows follow,
   shown unchanged, when the file and its warnings together fit 1800 bytes. The file counts as the
   bytes that would enter the context: its bytes as read, and on Codex, which decodes it as UTF-8, a
   byte that is not valid UTF-8 counts as the three bytes of the replacement character that stands in
   for it. Otherwise one line says they are too large to show. Then come the warnings. A symlinked `.trellis/rules.toml`, or one under
   a symlinked `.trellis`, is read and classified by the hooks, so its opt-outs apply and the sentence
   names them. The Codex bootstrap's no-hook fallback does not read a linked file at all, because it
   cannot apply the file's rows without reading them into the agent's context: there the file
   switches no rule off, every rule applies, and the agent says the file is a link. A hook's
   activation section counts as loaded, so the fallback never reads a file a hook withheld.
   But none of its contents is shown: one line says so, and its warnings are a count with no text taken
   from the file. The computed sentence is untested wording; `research-0012` measured the rows
   themselves, which the session still sees. The sentence exists so the session gets the effective
   result even when the file is not shown. Codex also puts the warnings on `systemMessage`, and its
   vendored-overlay branch uses the same framing. Where the host imports or embeds the file (a vendored
   overlay, a rendered `.claude/rules/trellis.md`, an inline managed block), no hook reads it.
5. **Nothing reconciles, so nothing is rewritten.** Neither hook reconciles, quarantines, writes
   provenance or tells the agent to rewrite `.trellis/rules.toml`. A file cannot fall out of step with
   the plugin: a missing row applies and an unknown one is ignored. `decision-0086` is retired in full,
   because the provenance its degradation removed no longer exists. The byte caps, 9500 B of context on
   Codex and 32768 B on Claude, stay as guards against a runaway payload. No project file within the
   1 MiB read bound can reach them, because the shared limit in point 4 and the warning cap in point 3
   bound what the file adds. On Codex that holds while the plugin's own path, which the context names
   once, stays within 400 bytes, a bound the parity test pins; the installed Codex plugin path measured
   56 bytes.
6. **One header ships, and no preset.** `trellis.md` carries the "By default" sentence for every
   project. `rules-a.toml`, `rules-b.toml`, `trellis-a.md`, `trellis-b.md` and the per-posture inline
   blocks retire. The shipped inline block carries no rows section; a project with opt-outs builds the
   block with its own `false` rows.
7. **A new file holds only what has effect.** It is exactly two lines:
   `# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }` and
   `[rules]`. `install.sh` seeds those bytes when no file exists, the hook's accept instruction quotes
   the same bytes, and a test pins the two to each other. An existing file is never touched.
8. **Adoption is unchanged.** Where the plugin's files live outside the repository, and on Codex, a
   project adopts by having the file (`decision-0070` D4 and D7, `decision-0077`). With the bundle
   vendored under the project's `.claude/skills/` and no file, every rule applies and the hook reads no
   rows (`decision-0070` D3).
9. **The installer renders one header whether or not it can read the file.** Its footer carries only
   the marker, the heading and the rows import. Over an unreadable file it still renders and says no
   opt-out in the file could be honoured and every rule applies (`decision-0088` D3).
10. **Frozen-text shapes keep full rows.** A vendored overlay, an inline managed block and a rendered
    file from an older installer carry text that applies a rule only when its row says
    `active = true`. A project on one of those shapes keeps a row for every rule until it migrates.
11. **A retired slug is never reused.** A stale `active = false` row for a retired slug would
    otherwise switch off a new rule that took its name.

## Consequences

- Every existing `.trellis/rules.toml` stays valid, and nothing asks a project to edit it. Inert rows
  and keys remain in the Kodhama files, including this repository's `strictness = "firm"`, whose
  sessions now receive the "By default" sentence. On those files, "add a row" works: an added
  `active = false` row below an existing `active = true` row switches the rule off.
- **Removing rows from an existing file is safe only once every install that opens the repository
  runs this release.** Installed plugin versions are pinned per checkout. On 2026-09-14 two files were
  measured, one holding only `[rules]` and one holding a single `false` row. The cached 0.6.0
  `staleness.sh` refuses such a file and loads no rules. The hooks of 0.7.0 to 0.23.x keep its `false`
  rows but reconcile the file and ask for the rows to be written back.
- On the Codex bootstrap's no-hook fallback, a project that declined Trellis through a symlinked rules
  file is still governed: the fallback does not read a linked file, so it never sees the
  `governed = false` inside it. Reaching this takes an older or hand-copied fallback block, since no
  install writes that block today, together with a linked rules file and a decline. A filtered read for
  that one line was judged not worth another instruction.
- A renamed rule's opt-out lapses: the old slug's row draws the unknown-slug warning, and the new slug
  applies until the project adds a row for it.
- The Codex context, measured with this change, is 7021 B for an empty file and 8164 B for this
  repository's sixteen-row file (1142 B), against the 9500 B cap. The shared limit in point 4 is set
  from the worst case: every non-floor rule switched off, with a file and warnings exactly at the
  limit. At 1800 B that case measured 9153 B, leaving room for a longer plugin path, and the
  too-large branch peaked at 8869 B. At the 2.5 KB first estimated it measured 9853 B, over the cap,
  so a project could have lost every rule on Codex. Known consumer files, 1.1 to 1.2 KB, are still
  shown in full.
- `json_escape` in `staleness.sh` runs in the C locale. In a UTF-8 locale, a byte that is not valid
  UTF-8 made it stop and silently cut the rest of the context. The classifier's move to the C locale
  had turned that byte from a loud refusal into silent truncation.
- A rendered `.claude/rules/trellis.md` imports `.trellis/rules.toml` through the host, and no hook
  reads that path, so a symlinked rules file on the curl-install path is still loaded as the host
  resolves it. This predates the change.
- `TestBothHostsClassifyRulesRowsIdentically` (`cli/rules_rows_parity_test.go`) is the executable
  statement of points 1, 3 and 4 on both hosts. The divergences it does not cover predate this change
  and are filed as TRL-100:
  - CR-only line endings
  - invalid UTF-8 bytes in a shown file
  - C0 control bytes
  - a second malformed `governed` line
  - `governed=false#comment` with no space before the comment
- `eval/experiments/annotation-vs-absence/run.sh` reads its payload from `2ec7da8`, the commit its
  recorded run names. `eval/experiments/does-trellis-help/run.sh` reads the shipped payload. Its
  recorded results used the posture-a header, rows and inline block, so a new run is not comparable
  with them.
- Thirteen records carry a dated note pointing here: `decision-0033`, `0051`, `0053`, `0058`, `0068`,
  `0070`, `0072`, `0074`, `0078`, `0083`, `0084`, `0087` and `0088`.

## Open questions

- Neither wording this change ships has been measured: the reworded activation sentence and the
  computed sentence naming the switched-off rules are both untested. `research-0012`'s effect was
  measured on a row flip under the old sentence. That experiment's runner still runs at its recorded
  commit, so a new run could measure the new payload the same way.

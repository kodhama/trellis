---
title: Shell reads of a project file must see the bytes and lines Node sees
date: 2026-10-09
category: conventions
module: plugins/trellis/hooks
problem_type: convention
component: tooling
severity: high
applies_when:
  - "Adding or changing a read of .trellis/rules.toml (or any project file) in staleness.sh or install.sh"
  - "Changing how codex-context.mjs splits or matches that file"
  - A parity row passes on Linux CI but the hosts still disagree on a Mac
tags: [host-parity, locale, lc-all, nul-byte, utf-8, bsd-sed, awk, rules-toml]
---

# Shell reads of a project file must see the bytes and lines Node sees

## Context

The Claude Code hook (`plugins/trellis/hooks/staleness.sh`, bash 3.2, BSD or GNU tools) and the Codex hook (`plugins/trellis/hooks/codex-context.mjs`, Node) read the same `.trellis/rules.toml` and must reach the same answer. Before TRL-100 (PR #324) they disagreed on six shapes, and each disagreement came from the shell tools seeing a different text than Node did. A disagreement here is not cosmetic. When one host honours `governed = false` and the other does not, one host silently drops every rule.

The fixes live at each read site, so a reader of any one site sees only its own comment. The C-locale pin alone was added three separate times: in `json_escape`, in the TRL-97 classifier awk, and in the TRL-100 governed head read. The next new read will need the same treatment.

## Guidance

Treat every shell read of project text as needing four things, and pin each new shape in the parity table.

1. **`LC_ALL=C` on every `sed`, `grep`, `awk` and `tr` that touches project text.** It goes on every stage of a pipeline, not just the first. Under a UTF-8 locale, macOS `sed` matches NBSP and U+2028 as `[[:space:]]`, while Node's `[ \t\v\f\r]` does not. A byte that is not valid UTF-8 makes BSD `sed` stop with `RE error: illegal byte sequence`, which silently empties the read: an opt-out beside a Latin-1 comment then governed. See `staleness.sh:335` and `install.sh:500`, which `TestInstallScriptGovernedParserMatchesHook` pins identical.
2. **Map NUL before a command substitution.** `$(…)` drops NUL bytes, so `governed = false\0` read as an opt-out in bash while Node, which keeps the byte, governed. The governed read starts with `LC_ALL=C tr '\000' '\001' < file`, so the NUL survives as a non-space byte (`staleness.sh:335`).
3. **On the Node side, split on `"\n"` only.** A regex with the `m` flag ends lines at CR, U+2028 and U+2029, which `sed` and `awk` never do. CR-only files, and opt-outs followed by U+2028, diverged that way. Use `.split("\n")` as `codex-context.mjs:788` and `:409` do, and anchor per line without `m`.
4. **Validate or transform one long line with one anchored match, not `gsub`.** A `gsub` that deletes each valid UTF-8 sequence took about 6 s on a 1 MiB single-line file under macOS awk. One `$0 ~ /^(…)*$/` per line is linear (`staleness.sh:1269`). Review checked the same regex against Node on macOS awk, gawk 5.3.1, mawk 1.3.4 and onetrue-awk.

When the hosts must agree on whether text is valid UTF-8, the Node test is `Buffer.from(text, "utf8").equals(bytes)`, in `readRequired`. Decoding replaces each bad byte with U+FFFD, so the round trip matches only valid input.

## Why This Matters

Each divergence was invisible from the host it favoured. The Claude hook printed a clean opt-out, the Codex hook a governed session, and nothing compared the two outside the parity test. Linux CI also hides the macOS-only half. GNU `sed` under a UTF-8 locale does not abort on Latin-1 and does not treat NBSP as space, so a missing `LC_ALL=C` passes CI and breaks on the maintainer's Mac.

## When to Apply

Any edit that makes either hook, or `install.sh`, read project text in a new way. Add the shape as a row in `TestBothHostsClassifyRulesRowsIdentically` (`cli/rules_rows_parity_test.go:595`, rows in `cli/rules_rows_parity_cases_test.go`). Set `utf8Locale: true` (`cli/rules_rows_parity_test.go:252`) for a shape that only a UTF-8 locale exposes. Such a row runs both hooks under `LC_ALL=en_US.UTF-8`, and it skips by name where that locale is missing. It catches the macOS regression only on a macOS run, so run the suite locally before pushing.

## Examples

Before (TRL-97): the governed read ran in the caller's locale, and NUL was dropped by `$(…)`:

```sh
governed_head="$(sed "1s/^$bom//" "$root/.trellis/rules.toml" 2>/dev/null | sed -n '/^[[:space:]]*\[/q;p')"
```

After (TRL-100):

```sh
governed_head="$(LC_ALL=C tr '\000' '\001' < "$root/.trellis/rules.toml" 2>/dev/null | LC_ALL=C sed "1s/^$bom//" 2>/dev/null | LC_ALL=C sed -n '/^[[:space:]]*\[/q;p')"
```

The redirect stays before `2>/dev/null`. `TestVendorGuardsAddedByReviewAreActuallyPinned` cuts each line at ` 2>` and would no longer see the read.

## Related

- TRL-100, PR #324 (unmerged as of this writing): the six divergences and their parity rows.
- `decision-2026-09-15-rule-rows-only-switch-rules-off`, point 4 and its 2026-10-09 dated note.

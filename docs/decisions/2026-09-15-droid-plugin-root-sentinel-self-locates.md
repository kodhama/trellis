---
id: decision-2026-09-15-droid-plugin-root-sentinel-self-locates
type: decision
depends_on: [decision-0028, decision-0070, decision-0087]
informed_by: [decision-0039]
changes: [decision-0070]
owner: agent
date: 2026-09-15
---

> **Provenance.** TRL-90 reports this from Droid 0.219.0 with Trellis 0.24.1. A fresh
> Math Quest session supplied `CLAUDE_PLUGIN_ROOT=/PLUGIN_ROOT_NOT_EXPANDED_ERROR` while still
> executing the installed hook. The intact payload delivered 8,222 bytes, the
> `trellis:rules-loaded` terminator and 16 active rows when the same hook was run with its real
> plugin root. Revalidated on current `main`: the sentinel run emitted `TRELLIS_RULES_NOT_LOADED`;
> the real-root run emitted 8,235 bytes, the terminator and 16 rows.

# Droid's plugin-root sentinel self-locates from the running hook

## Context

`hooks/hooks.json` launches `"$CLAUDE_PLUGIN_ROOT"/hooks/staleness.sh`. Droid resolves that command
far enough to execute the real installed script, but the environment visible inside the script
still carries `/PLUGIN_ROOT_NOT_EXPANDED_ERROR`. `staleness.sh` accepted every non-empty value as
the plugin root, derived `reference/rules.md` from the sentinel, and let the payload gateway
correctly report that invented path as missing. The result was a governance blackout even though
the host had already proved the installed hook's real location by running it.

Falling back whenever a payload file is absent would fix this case by hiding a different one. For
example, an explicit `CLAUDE_PLUGIN_ROOT=/tmp/broken-plugin` must still reach the payload gateway
and report the broken install, not read a second copy beside whichever script happened to run.

## Decision

1. When `CLAUDE_PLUGIN_ROOT` is exactly `/PLUGIN_ROOT_NOT_EXPANDED_ERROR`, `staleness.sh` takes
   the directory one level above `${BASH_SOURCE[0]}`'s directory and canonicalizes it with
   `pwd -P` as the effective plugin root.
2. No other value self-heals. An unset root and an arbitrary bad path retain the existing loud
   payload failure. If resolving the known sentinel from the hook location fails, the sentinel
   also remains in place and the payload gateway reports it.
3. `hooks.json` stays unchanged. Droid already resolves its command to the installed script; the
   defect is the different value exposed to that script after launch.

## Consequences

- `TestPluginRootSentinelFallsBackToHookLocation` executes a copied hook beside a copied payload,
  with the sentinel in its environment. It must load that copy, classify the project row, and
  point at that copy's invariants file. Existing broken-payload coverage keeps arbitrary invalid
  roots loud.
- The plugin release advances from 0.24.1 to 0.24.2, including both manifests and `install.sh`'s
  baked bundle manifest. The reference payload and its content stamp do not change.
- Droid's host expansion bug remains visible at the host boundary, but it no longer disables
  Trellis governance while the host can execute the installed hook.

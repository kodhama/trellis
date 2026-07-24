---
id: eval-codex-hook-delivery
type: research-note
status: draft
depends_on: []
informed_by: [research-0013]
owner: agent
---

# Codex plugin-hook delivery prototype

This experiment checks whether a Codex plugin can deliver Trellis rules as
always-on developer context without copying the generated rule text into the
consumer repository.

This directory is **research apparatus, not a shipping implementation**. It
tests an unsettled delivery question and implements no product spec. The
prototype is informed by draft `research-0013`; it does not build on that note
as an approved contract.

It deliberately does **not** start a nested Codex model session. The runner:

1. exercises the hook command directly for `SessionStart` and `SubagentStart`;
2. installs a temporary `.codex-plugin` through a temporary marketplace and
   temporary Codex home;
3. asks `codex debug prompt-input` to render the model-visible static prompt;
4. verifies the repository bootstrap and records whether that debug surface
   executes lifecycle hooks.

The plugin payload is staged from the current Trellis payload at run time, so
the size check uses the real `rules.md`, `rules.toml`, and payload version.

`codex debug prompt-input` does not start a session and, as this experiment
discovered, does not execute `SessionStart` hooks. That makes it suitable for
verifying the committed bootstrap but not for an end-to-end hook assertion.
The runner reports this explicitly instead of treating absent hook context as
a product failure or silently claiming an end-to-end pass.

Run from the repository root:

```sh
eval/experiments/codex-hook-delivery/run.sh
```

## Results

Run on 2026-07-23 with Codex CLI `0.145.0`:

- **Verified:** the source hook and the plugin-cache copy both returned valid
  Codex hook JSON for `SessionStart` sources `startup`, `resume`, `clear`, and
  `compact`.
- **Verified:** `SubagentStart` returned the same standing rule context.
- **Verified:** missing `.trellis/rules.toml` returned both a visible
  `systemMessage` and `TRELLIS_HOOK_FAILURE` developer context instructing the
  agent to stop.
- **Verified:** the real Trellis rule payload plus project rows was 6,519
  characters, below this experiment's conservative 8,000-character envelope.
- **Verified:** Codex installed and enabled the temporary plugin from an
  isolated marketplace and temporary Codex home.
- **Verified:** `codex debug prompt-input` loaded the static `AGENTS.md`
  bootstrap.
- **Verified:** the hook-contract assertion is byte-exact, and adversarial
  prompt fixtures reject rule-shaped metadata and user-role text while
  accepting the same marker and rules in a separate developer-role message.
- **Verified in a maintainer-run live probe:** one ephemeral, read-only
  `codex exec` startup returned `HOOK=TRELLIS_HOOK_CONTEXT payload@0760a802ccd1`
  and `RULE=inv-handover-points` when instructed not to call tools or read
  files. This verifies live `SessionStart` delivery at startup through the
  installed plugin.
- **Not verified live:** `resume`, `clear`, `compact`, `SubagentStart`, Codex
  cloud/headless, and IDE surfaces. Their direct hook-command contracts pass,
  but the live probe exercised startup only.

**Finding:** the mechanism is viable at the plugin-install and hook-contract
layers and is proven end to end for one local Codex CLI startup. Broader
lifecycle and surface reliability remain open.

## Sources and confidence

- [Codex hooks](https://developers.openai.com/codex/hooks) — **high**:
  official event, output, plugin-root, and trust behavior.
- [Build Codex plugins](https://developers.openai.com/codex/plugins/build) —
  **high**: official manifest and marketplace structure.
- `codex plugin marketplace add`, `codex plugin add`, and
  `codex debug prompt-input` from local Codex CLI `0.145.0` — **high** for the
  observed local behavior recorded above.
- Live startup delivery — **medium-high**: maintainer-run ephemeral Codex CLI
  probe, model-reported marker and rule slug, with file/tool access explicitly
  forbidden. Hook trust UX was bypassed for the already-vetted prototype hook.

## Open questions

- Does the same hook output load in Codex cloud/headless execution?
- Does hook trust remain valid across a plugin upgrade, or does every changed
  hook hash require renewed trust?
- Should a production version inject all rules plus activation rows, or render
  only active rules before injection?
- What is the correct static fallback for Codex IDE sessions, where plugins are
  unavailable?
- Should a shipping hook avoid the prototype's Node.js runtime dependency, or
  can the Codex plugin contract guarantee that runtime on every target surface?

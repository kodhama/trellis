#!/usr/bin/env bash
set -euo pipefail

experiment_root="$(cd "$(dirname "$0")" && pwd)"
repo_root="$(cd "$experiment_root/../../.." && pwd)"
prototype_tmp="$(mktemp -d "${TMPDIR:-/tmp}/trellis-codex-hook.XXXXXX")"
marketplace_root="$prototype_tmp/marketplace"
plugin_root="$marketplace_root/plugins/trellis-hook-probe"
fixture_root="$prototype_tmp/project"
prototype_codex_home="$prototype_tmp/codex-home"
prompt_output="$prototype_tmp/prompt-input.json"
install_output="$prototype_tmp/plugin-install.json"

mkdir -p \
  "$marketplace_root/.agents/plugins" \
  "$marketplace_root/plugins" \
  "$plugin_root/reference" \
  "$fixture_root/.trellis" \
  "$prototype_codex_home"

cp -R "$experiment_root/plugin-template/." "$plugin_root/"
cp "$repo_root/plugins/trellis/reference/rules.md" "$plugin_root/reference/rules.md"
cp "$repo_root/plugins/trellis/reference/version" "$plugin_root/reference/version"
cp "$repo_root/.trellis/rules.toml" "$fixture_root/.trellis/rules.toml"
cp "$experiment_root/marketplace.json" \
  "$marketplace_root/.agents/plugins/marketplace.json"
cp "$experiment_root/bootstrap-AGENTS.md" "$fixture_root/AGENTS.md"

node "$experiment_root/verify-hook.mjs" \
  "$plugin_root/hooks/inject-rules.mjs" \
  "$plugin_root" \
  "$fixture_root"

env CODEX_HOME="$prototype_codex_home" \
  codex plugin marketplace add "$marketplace_root" --json
env CODEX_HOME="$prototype_codex_home" \
  codex plugin add trellis-hook-probe@trellis-hook-prototype --json \
  > "$install_output"

installed_plugin_root="$(
  node -e \
    'const fs = require("node:fs"); const value = JSON.parse(fs.readFileSync(process.argv[1], "utf8")); process.stdout.write(value.installedPath);' \
    "$install_output"
)"

node "$experiment_root/verify-hook.mjs" \
  "$installed_plugin_root/hooks/inject-rules.mjs" \
  "$installed_plugin_root" \
  "$fixture_root"

(
  cd "$fixture_root"
  env CODEX_HOME="$prototype_codex_home" \
    codex --dangerously-bypass-hook-trust \
    debug prompt-input "Confirm the loaded Trellis delivery markers." \
    > "$prompt_output"
)

node "$experiment_root/verify-prompt.mjs" "$prompt_output" absent
node "$experiment_root/verify-prompt.mjs" \
  "$experiment_root/metadata-only-negative-control.json" \
  absent
node "$experiment_root/verify-prompt.mjs" \
  "$experiment_root/developer-hook-positive-control.json" \
  present

printf 'Prototype artifacts: %s\n' "$prototype_tmp"

import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const pluginRoot = path.join(repoRoot, "plugins", "trellis");
const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), "trellis-plugin-smoke-"));
const rulesDir = path.join(projectRoot, ".trellis");

function contextFrom(response, host) {
  const context = response?.hookSpecificOutput?.additionalContext;
  assert.equal(typeof context, "string", `${host} did not return SessionStart context`);
  assert.match(context, /<!-- trellis:rules-loaded -->/u, `${host} omitted the rules sentinel`);
  assert.doesNotMatch(
    context,
    /TRELLIS_RULES_NOT_LOADED/u,
    `${host} reported a broken payload during the healthy smoke path`,
  );
}

try {
  fs.mkdirSync(path.join(projectRoot, ".git"));
  fs.mkdirSync(rulesDir);
  fs.copyFileSync(
    path.join(pluginRoot, "reference", "rules-b.toml"),
    path.join(rulesDir, "rules.toml"),
  );

  const claudeResponse = JSON.parse(
    execFileSync(path.join(pluginRoot, "hooks", "staleness.sh"), {
      cwd: projectRoot,
      encoding: "utf8",
      env: {
        ...process.env,
        CLAUDE_PLUGIN_ROOT: pluginRoot,
        CLAUDE_PROJECT_DIR: projectRoot,
      },
    }),
  );
  contextFrom(claudeResponse, "Claude");

  const codexResponse = JSON.parse(
    execFileSync(process.execPath, [path.join(pluginRoot, "hooks", "codex-context.mjs")], {
      cwd: projectRoot,
      encoding: "utf8",
      env: { ...process.env, PLUGIN_ROOT: pluginRoot },
      input: JSON.stringify({
        cwd: projectRoot,
        hook_event_name: "SessionStart",
        source: "startup",
      }),
    }),
  );
  contextFrom(codexResponse, "Codex");

  console.log("plugin smoke: both SessionStart hooks delivered the complete rule payload");
} finally {
  fs.rmSync(projectRoot, { force: true, recursive: true });
}

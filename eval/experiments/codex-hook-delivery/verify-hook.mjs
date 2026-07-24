import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";

const [hookPath, pluginRoot, projectRoot] = process.argv.slice(2);
assert.ok(hookPath && pluginRoot && projectRoot, "hook, plugin, and project paths are required");

const rules = fs
  .readFileSync(path.join(pluginRoot, "reference", "rules.md"), "utf8")
  .trim();
const version = fs
  .readFileSync(path.join(pluginRoot, "reference", "version"), "utf8")
  .trim();
const rows = fs
  .readFileSync(path.join(projectRoot, ".trellis", "rules.toml"), "utf8")
  .trim();
const expectedContext = [
  `TRELLIS_HOOK_CONTEXT ${version}`,
  "The Trellis rules below are standing project instructions. Apply a rule only when its rules.toml row is active; the two floor rules always apply.",
  "",
  rules,
  "",
  "## Project rule activation",
  "",
  rows,
].join("\n");

const runHook = (event) => {
  const result = spawnSync(process.execPath, [hookPath], {
    input: JSON.stringify(event),
    encoding: "utf8",
    env: {
      ...process.env,
      PLUGIN_ROOT: pluginRoot,
    },
  });
  assert.equal(result.status, 0, result.stderr);
  return JSON.parse(result.stdout);
};

for (const source of ["startup", "resume", "clear", "compact"]) {
  const output = runHook({
    hook_event_name: "SessionStart",
    source,
    cwd: projectRoot,
  });
  const context = output.hookSpecificOutput.additionalContext;
  assert.equal(context, expectedContext);
  assert.equal(output.hookSpecificOutput.hookEventName, "SessionStart");
  assert.ok(
    context.length < 8000,
    `context is ${context.length} characters; expected a conservative sub-2,500-token envelope`,
  );
}

const subagentOutput = runHook({
  hook_event_name: "SubagentStart",
  agent_type: "explorer",
  cwd: projectRoot,
});
assert.equal(
  subagentOutput.hookSpecificOutput.hookEventName,
  "SubagentStart",
);
assert.equal(
  subagentOutput.hookSpecificOutput.additionalContext,
  expectedContext,
);

const missingProject = path.join(projectRoot, ".missing-config-fixture");
assert.equal(fs.existsSync(missingProject), false);
const failureOutput = runHook({
  hook_event_name: "SessionStart",
  source: "startup",
  cwd: missingProject,
});
assert.match(failureOutput.systemMessage, /^TRELLIS_HOOK_FAILURE /);
assert.match(
  failureOutput.hookSpecificOutput.additionalContext,
  /Stop substantive work/,
);

process.stdout.write(
  `${JSON.stringify({
    direct_hook_contract: "PASS",
    session_sources: ["startup", "resume", "clear", "compact"],
    subagent_start: "PASS",
    missing_config_failure: "PASS",
    context_characters: expectedContext.length,
  })}\n`,
);

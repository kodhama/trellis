import fs from "node:fs";
import path from "node:path";

const rawInput = await new Promise((resolve, reject) => {
  let input = "";
  process.stdin.setEncoding("utf8");
  process.stdin.on("data", (chunk) => {
    input += chunk;
  });
  process.stdin.on("end", () => resolve(input));
  process.stdin.on("error", reject);
});

let event;
try {
  event = JSON.parse(rawInput || "{}");
} catch {
  emitFailure("Trellis hook received invalid event JSON.");
}

const eventName = event.hook_event_name;
if (eventName !== "SessionStart" && eventName !== "SubagentStart") {
  emitFailure(`Trellis hook received unsupported event: ${eventName || "unknown"}.`);
}

const pluginRoot =
  process.env.PLUGIN_ROOT || process.env.CLAUDE_PLUGIN_ROOT || "";
const projectRoot = event.cwd || process.cwd();
const rulesPath = path.join(pluginRoot, "reference", "rules.md");
const versionPath = path.join(pluginRoot, "reference", "version");
const rowsPath = path.join(projectRoot, ".trellis", "rules.toml");

const missing = [rulesPath, versionPath, rowsPath].filter(
  (candidate) => !fs.existsSync(candidate),
);
if (missing.length > 0) {
  emitFailure(`Trellis hook cannot read: ${missing.join(", ")}.`);
}

const rules = readRequired(rulesPath, "rules").trim();
const version = readRequired(versionPath, "version").trim();
const rows = readRequired(rowsPath, "activation rows").trim();

if (!rules || !version || !rows) {
  emitFailure("Trellis hook found an empty rules, version, or rows file.");
}

const additionalContext = [
  `TRELLIS_HOOK_CONTEXT ${version}`,
  "The Trellis rules below are standing project instructions. Apply a rule only when its rules.toml row is active; the two floor rules always apply.",
  "",
  rules,
  "",
  "## Project rule activation",
  "",
  rows,
].join("\n");

process.stdout.write(
  `${JSON.stringify({
    hookSpecificOutput: {
      hookEventName: eventName,
      additionalContext,
    },
  })}\n`,
);

function readRequired(filePath, label) {
  try {
    return fs.readFileSync(filePath, "utf8");
  } catch (error) {
    emitFailure(
      `Trellis hook could not read ${label} at ${filePath}: ${error.message}.`,
    );
  }
}

function emitFailure(reason) {
  const eventNameForOutput =
    event?.hook_event_name === "SubagentStart" ? "SubagentStart" : "SessionStart";
  const message = `TRELLIS_HOOK_FAILURE ${reason} Stop substantive work and tell the user that Trellis rules were not loaded.`;
  process.stdout.write(
    `${JSON.stringify({
      systemMessage: message,
      hookSpecificOutput: {
        hookEventName: eventNameForOutput,
        additionalContext: message,
      },
    })}\n`,
  );
  process.exit(0);
}

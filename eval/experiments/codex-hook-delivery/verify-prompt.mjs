import assert from "node:assert/strict";
import fs from "node:fs";

const promptPath = process.argv[2];
const expectedHookState = process.argv[3] || "absent";
assert.ok(promptPath, "prompt JSON path is required");
assert.ok(
  expectedHookState === "present" || expectedHookState === "absent",
  "expected hook state must be present or absent",
);

const parsed = JSON.parse(fs.readFileSync(promptPath, "utf8"));
assert.ok(Array.isArray(parsed), "prompt input must be a message array");

const messages = parsed.filter(
  (item) => item?.type === "message" && Array.isArray(item.content),
);
const textEntries = messages.flatMap((message) =>
  message.content
    .filter((item) => item?.type === "input_text" && typeof item.text === "string")
    .map((item) => ({
      role: message.role,
      text: item.text,
    })),
);

const bootstrapEntries = textEntries.filter(
  ({ text }) =>
    text.startsWith("# AGENTS.md instructions for ") &&
    text.includes("`TRELLIS_BOOTSTRAP`"),
);
assert.equal(
  bootstrapEntries.length,
  1,
  "expected exactly one structured AGENTS.md bootstrap entry",
);

const hookEntries = textEntries.filter(
  ({ role, text }) =>
    role === "developer" &&
    !text.startsWith("# AGENTS.md instructions for ") &&
    /^TRELLIS_HOOK_CONTEXT payload@/m.test(text) &&
    text.includes("inv-directional-flow") &&
    text.includes("[rules]"),
);
const hookContextVisible = hookEntries.length > 0;
assert.equal(
  hookContextVisible,
  expectedHookState === "present",
  `expected hook context to be ${expectedHookState}`,
);

process.stdout.write(
  `${JSON.stringify({
    codex_static_prompt_render: "PASS",
    bootstrap_visible: true,
    hook_context_visible: hookContextVisible,
    hook_end_to_end:
      hookContextVisible
        ? "PASS"
        : "NOT_RUN_BY_DEBUG_PROMPT_INPUT",
  })}\n`,
);

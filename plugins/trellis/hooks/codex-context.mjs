#!/usr/bin/env node

// spec-0007@v1: trusted local Codex SessionStart(startup) transport.
// The installed project overlay is the sole authority. This handler validates
// every input before writing its one JSON response to stdout.

import fs from "node:fs";
import path from "node:path";

const SENTINEL = "<!-- trellis:rules-loaded -->";
const MAX_CONTEXT_BYTES = 8000;
const SLUGS = [
  "inv-directional-flow",
  "inv-handover-points",
  "inv-intent-locus",
  "inv-ratifiable-artifacts",
  "inv-graph-maintenance",
  "inv-self-improvement",
  "inv-gate-at-handover",
  "inv-independent-judgment",
  "inv-auditable-archive",
  "inv-bounded-context",
  "inv-minimal-first",
  "inv-clarify-before-commit",
  "floor-transparency",
  "floor-intent-gate",
];
const SLUG_SET = new Set(SLUGS);
// The project always owns its rows. The three payload files come from the
// vendored overlay when one exists, and from the plugin's own payload when it
// does not (decision-0065: the plugin path vendors nothing). Vendored projects
// keep reading their own copies, so nothing about them changes.
const PROJECT_CONFIG = ".trellis/rules.toml";
const VENDORED_PAYLOAD = {
  prose: ".trellis/internal/trellis.md",
  rules: ".trellis/internal/rules.md",
  version: ".trellis/internal/version",
};

function emit(value) {
  process.stdout.write(`${JSON.stringify(value)}\n`);
}

function fail(label, validationClass) {
  emit({
    systemMessage:
      `Trellis hook did not load rules: ${label}: ${validationClass}. ` +
      "The AGENTS.md bootstrap must attempt the installed overlay.",
  });
}

function existingDirectory(value) {
  if (typeof value !== "string" || !path.isAbsolute(value)) return false;
  try {
    return fs.statSync(value).isDirectory();
  } catch {
    return false;
  }
}

function validPluginRoot(root) {
  if (!existingDirectory(root)) return false;
  try {
    const manifest = JSON.parse(
      fs.readFileSync(path.join(root, ".codex-plugin", "plugin.json"), "utf8"),
    );
    return manifest !== null && manifest.name === "trellis";
  } catch {
    return false;
  }
}

function nearestGitBoundary(cwd) {
  let current = cwd;
  for (;;) {
    const marker = path.join(current, ".git");
    try {
      const stat = fs.statSync(marker);
      if (stat.isDirectory() || stat.isFile()) return current;
    } catch {
      // Keep walking only to the filesystem root.
    }
    const parent = path.dirname(current);
    if (parent === current) return null;
    current = parent;
  }
}

function nearestOverlay(cwd, boundary) {
  let current = cwd;
  for (;;) {
    try {
      if (fs.statSync(path.join(current, ".trellis", "rules.toml")).isFile()) {
        return current;
      }
    } catch {
      // This directory has no candidate overlay.
    }
    if (current === boundary) return null;
    const parent = path.dirname(current);
    if (parent === current) return null;
    current = parent;
  }
}

function readRequired(projectRoot, relativePath) {
  const absolute = path.join(projectRoot, relativePath);
  let stat;
  try {
    stat = fs.statSync(absolute);
  } catch (error) {
    if (error?.code === "ENOENT") return { error: "missing-file" };
    return { error: "unreadable-file" };
  }
  if (!stat.isFile()) return { error: "unreadable-file" };
  if (stat.size > MAX_CONTEXT_BYTES) {
    return { label: "assembled-context", error: "context-over-budget" };
  }
  let descriptor;
  try {
    fs.accessSync(absolute, fs.constants.R_OK);
    descriptor = fs.openSync(absolute, "r");
    const openedStat = fs.fstatSync(descriptor);
    if (!openedStat.isFile()) return { error: "unreadable-file" };
    if (openedStat.size > MAX_CONTEXT_BYTES) {
      return { label: "assembled-context", error: "context-over-budget" };
    }

    const buffer = Buffer.alloc(MAX_CONTEXT_BYTES + 1);
    let total = 0;
    while (total < buffer.length) {
      const count = fs.readSync(
        descriptor,
        buffer,
        total,
        buffer.length - total,
        null,
      );
      if (count === 0) break;
      total += count;
    }
    if (total > MAX_CONTEXT_BYTES) {
      return { label: "assembled-context", error: "context-over-budget" };
    }
    return { value: buffer.subarray(0, total).toString("utf8") };
  } catch {
    return { error: "unreadable-file" };
  } finally {
    if (descriptor !== undefined) {
      try {
        fs.closeSync(descriptor);
      } catch {
        // The read result already captures the only protocol-visible outcome.
      }
    }
  }
}

function parseQuotedTomlString(source) {
  if (source.startsWith("'")) {
    const end = source.indexOf("'", 1);
    if (end < 0 || !/^[ \t]*(?:#.*)?$/u.test(source.slice(end + 1))) return null;
    const value = source.slice(1, end);
    for (const character of value) {
      const codePoint = character.codePointAt(0);
      if ((codePoint < 0x20 && codePoint !== 0x09) || codePoint === 0x7f) {
        return null;
      }
    }
    return value;
  }
  if (!source.startsWith('"')) return null;

  let value = "";
  for (let index = 1; index < source.length; index += 1) {
    const character = source[index];
    if (character === '"') {
      return /^[ \t]*(?:#.*)?$/u.test(source.slice(index + 1)) ? value : null;
    }
    if (character !== "\\") {
      const codePoint = character.codePointAt(0);
      if ((codePoint < 0x20 && codePoint !== 0x09) || codePoint === 0x7f) {
        return null;
      }
      value += character;
      continue;
    }

    index += 1;
    if (index >= source.length) return null;
    const escape = source[index];
    const simpleEscapes = {
      b: "\b",
      t: "\t",
      n: "\n",
      f: "\f",
      r: "\r",
      '"': '"',
      "\\": "\\",
    };
    if (Object.hasOwn(simpleEscapes, escape)) {
      value += simpleEscapes[escape];
      continue;
    }
    if (escape !== "u" && escape !== "U") return null;

    const digits = escape === "u" ? 4 : 8;
    const hex = source.slice(index + 1, index + 1 + digits);
    if (hex.length !== digits || !/^[0-9A-Fa-f]+$/u.test(hex)) return null;
    const codePoint = Number.parseInt(hex, 16);
    if (
      codePoint > 0x10ffff ||
      (codePoint >= 0xd800 && codePoint <= 0xdfff)
    ) {
      return null;
    }
    value += String.fromCodePoint(codePoint);
    index += digits;
  }
  return null;
}

// This deliberately parses only Trellis's declared rules.toml schema, not an
// approximation that silently accepts unknown TOML. It supports the two TOML
// string forms used by consumer edits (basic and literal), and rejects duplicate
// keys, sections, and rows deterministically.
function parseRulesToml(source) {
  const topLevel = new Map();
  const rows = new Map();
  let rulesSectionSeen = false;
  let inRules = false;

  for (const rawLine of source.split(/\r?\n/u)) {
    const line = rawLine.replace(/^[ \t]+|[ \t]+$/gu, "");
    if (line === "" || line.startsWith("#")) continue;

    const section = line.match(/^\[([^\]]+)\][ \t]*(?:#.*)?$/u);
    if (section) {
      if (section[1] !== "rules" || rulesSectionSeen) return null;
      rulesSectionSeen = true;
      inRules = true;
      continue;
    }

    if (!inRules) {
      const assignment = line.match(/^([A-Za-z_][A-Za-z0-9_-]*)[ \t]*=[ \t]*(.*)$/u);
      if (
        !assignment ||
        (assignment[1] !== "seeded_from" && assignment[1] !== "strictness") ||
        topLevel.has(assignment[1])
      ) {
        return null;
      }
      const value = parseQuotedTomlString(assignment[2]);
      if (value === null) return null;
      topLevel.set(assignment[1], value);
      continue;
    }

    const row = line.match(
      /^([a-z][a-z-]*)[ \t]*=[ \t]*\{[ \t]*active[ \t]*=[ \t]*(true|false)[ \t]*\}(?:[ \t]*#.*)?$/u,
    );
    if (!row || rows.has(row[1]) || !SLUG_SET.has(row[1])) return null;
    rows.set(row[1], row[2] === "true");
  }

  const strictness = topLevel.get("strictness");
  if (
    !rulesSectionSeen ||
    (strictness !== "firm" && strictness !== "adaptive") ||
    rows.size !== SLUGS.length ||
    SLUGS.some((slug) => !rows.has(slug))
  ) {
    return null;
  }
  return rows;
}

let input;
try {
  input = JSON.parse(fs.readFileSync(0, "utf8"));
} catch {
  fail("stdin", "invalid-json");
  process.exit(0);
}

if (input === null || typeof input !== "object" || Array.isArray(input)) {
  fail("stdin", "invalid-json");
  process.exit(0);
}
if (input.hook_event_name !== "SessionStart") {
  fail("hook_event_name", "wrong-event");
  process.exit(0);
}
if (input.source !== "startup") {
  fail("source", "wrong-event");
  process.exit(0);
}
if (!existingDirectory(input.cwd)) {
  fail("cwd", "invalid-cwd");
  process.exit(0);
}

const pluginRoot = process.env.PLUGIN_ROOT;
if (!validPluginRoot(pluginRoot)) {
  fail("PLUGIN_ROOT", "invalid-plugin-root");
  process.exit(0);
}

const gitBoundary = nearestGitBoundary(input.cwd);
if (gitBoundary === null) {
  fail("project-root", "project-root-not-found");
  process.exit(0);
}
const projectRoot = nearestOverlay(input.cwd, gitBoundary);
if (projectRoot === null) {
  fail("project-root", "project-root-not-found");
  process.exit(0);
}

// The rows are read first: they carry the posture that selects which prose
// variant the plugin payload should supply.
const configResult = readRequired(projectRoot, PROJECT_CONFIG);
// decision-0070 D5. A project that declares `governed = false` is not governed —
// on EITHER host. Checked here, before the rules are parsed or assembled, for the
// same reason the Claude hook checks it before every delivery path: an opt-out
// that only one host honours is not an opt-out. Matched on the raw text rather
// than through the row parser, because the parser deliberately understands only
// the declared rules schema and would reject an unknown top-level key.
if (
  !configResult.error &&
  /^[ \t]*governed[ \t]*=[ \t]*false/m.test(configResult.value ?? "")
) {
  process.exit(0);
}
if (configResult.error) {
  fail(configResult.label ?? PROJECT_CONFIG, configResult.error);
  process.exit(0);
}
const rulesToml = configResult.value;

// The `.trellis/internal/` DIRECTORY decides the mode, not any file inside it.
// Present -> vendored: every file within is required, and a missing one is a
// broken overlay that must fail loudly rather than silently falling through to
// the plugin's payload. Absent -> plugin-native. Using a file as the
// discriminator would turn a half-deleted overlay into a silent mode switch.
const vendored = existingDirectory(path.join(projectRoot, ".trellis", "internal"));
// Both TOML string forms. parseRulesToml below accepts literal strings, so
// matching only the basic form served a firm project the adaptive posture
// without saying so.
const posture = /^\s*strictness\s*=\s*(?:"firm"|'firm')\s*(?:#.*)?$/mu.test(rulesToml)
  ? "a"
  : "b";
const sources = vendored
  ? { root: projectRoot, ...VENDORED_PAYLOAD }
  : {
      root: pluginRoot,
      prose: `reference/trellis-${posture}.md`,
      rules: "reference/rules.md",
      version: "reference/version",
    };

const payload = {};
for (const key of ["prose", "rules", "version"]) {
  const result = readRequired(sources.root, sources[key]);
  if (result.error) {
    fail(result.label ?? sources[key], result.error);
    process.exit(0);
  }
  payload[key] = result.value;
}

const trellis = payload.prose;
const rules = payload.rules;
const version = payload.version;

if (trellis.length === 0) {
  fail(sources.prose, "empty-prose");
  process.exit(0);
}
if (rules.length === 0) {
  fail(sources.rules, "empty-prose");
  process.exit(0);
}
if (!/^payload@[0-9a-f]{12}\n?$/u.test(version)) {
  fail(sources.version, "invalid-version");
  process.exit(0);
}
const rows = parseRulesToml(rulesToml);
if (rows === null) {
  fail(PROJECT_CONFIG, "invalid-rules");
  process.exit(0);
}
if (trellis.split("@rules.md").length - 1 !== 1) {
  fail(".trellis/internal/trellis.md", "invalid-placeholder-count");
  process.exit(0);
}
if (
  rules.split(SENTINEL).length - 1 !== 1 ||
  !rules.endsWith(`${SENTINEL}\n`)
) {
  fail(".trellis/internal/rules.md", "invalid-rules");
  process.exit(0);
}

const stamp = version.endsWith("\n") ? version.slice(0, -1) : version;
const context =
  trellis.replace("@rules.md", rules) +
  "\n" +
  rulesToml +
  (rulesToml.endsWith("\n") ? "" : "\n") +
  `Trellis hook loaded installed overlay: ${stamp}\n`;

if (Buffer.byteLength(context, "utf8") > MAX_CONTEXT_BYTES) {
  fail("assembled-context", "context-over-budget");
  process.exit(0);
}

const response = {
  hookSpecificOutput: {
    hookEventName: "SessionStart",
    additionalContext: context,
  },
};
const falseFloors = ["floor-intent-gate", "floor-transparency"]
  .filter((slug) => rows.get(slug) === false)
  .sort();
if (falseFloors.length > 0) {
  response.systemMessage =
    "Trellis warning: floor rows set active = false are overridden-by-floor and remain active: " +
    `${falseFloors.join(", ")}.`;
}
emit(response);

#!/usr/bin/env node

// Trusted local Codex SessionStart(startup) transport (`decision-0058`;
// contract recorded as spec-0007@v1, retired with `specs/` by decision-0079 —
// cli/codex_hook_test.go is now its executable statement).
// The installed project overlay is the sole authority. This handler validates
// every input before writing its one JSON response to stdout.

import fs from "node:fs";
import path from "node:path";

const SENTINEL = "<!-- trellis:rules-loaded -->";
// 8000 (this constant's prior value) had no recorded rationale — it entered in
// commit 3490555 with none, and "8000" appears nowhere in decisions/, research/
// or core/. Investigated for Ruling 6 (TRL-20 task 3, fix round 1): Codex's own
// default per-hook-message limit is documented at
// https://learn.chatgpt.com/docs/hooks as roughly 2,500 TOKENS, not bytes, and
// Codex does not reject over that limit — it spills gracefully, saving the full
// text under `<temp_dir>/hook_outputs/<session_id>/<uuid>.txt` and giving the
// model a head-and-tail preview plus the saved-file path (the setting is
// configurable per handler via `additionalContextLimit`, and the installed
// codex-cli binary corroborates the setting exists). So the OLD 8000-byte cap
// measured the wrong unit against a limit that does not even fail closed: at
// ~4 bytes/token, 8000 B is ~2000 tokens, comfortably under Codex's 2500 —
// this hook's own "context-over-budget" refusal was a SELF-INFLICTED blackout,
// strictly worse than what Codex would have done on its own (spill and point at
// the file, not lose the rules). 9500 B is ~2375 tokens: still under Codex's
// ~10,000-byte-equivalent default, so this hook still never triggers Codex's
// own spill path either — it only stops refusing at a limit nobody imposed.
const MAX_CONTEXT_BYTES = 9500;
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
//
// slugs is the set the payload actually ships (see slugsFromRules below), passed
// in rather than closed over. A hardcoded list here could not be repaired by a
// plugin upgrade, and a stale one made a quarantine reason false: the agent would
// quarantine a live row and cite a payload that ships it.
//
// Returns `{ rows, mismatch }` on any file this parser can make sense of, or
// `null` only for a genuine syntax fault — a malformed row, an unknown top-level
// key, a duplicate top-level key or section, or a `strictness` that is present
// but neither "firm" nor "adaptive". `mismatch` is null when every row's slug
// matched exactly once; otherwise it names the three ways a row can fail to
// match the slug set — missing, unknown, duplicate — and the caller reconciles
// rather than refusing (staleness.sh's TRL-20 fix, now mirrored here). This used
// to be a pass/fail gate (`return null` on any of those three); the classifier
// split is what lets a single bad row stop costing the other fifteen.
function parseRulesToml(source, slugs) {
  const slugSet = new Set(slugs);
  const topLevel = new Map();
  const rows = new Map();
  const unknown = [];
  const duplicate = [];
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
      // `governed` is skipped, not parsed. decision-0070 D5 gave it meaning, and
      // the opt-out is handled far earlier by a raw-text match that never reaches
      // this parser. But leaving it OFF the accepted set made `governed = true` —
      // the natural way to reverse an opt-out without deleting the line — a fatal
      // `invalid-rules` on Codex while Claude governed normally: measured, 12
      // rules vs 0. A key one host acts on and the other rejects is worse than a
      // key neither knows.
      if (assignment && assignment[1] === "governed") {
        // Skipped, but only for the two values it can legally hold. `governed =
        // falsehood` must reach the reject path below and surface as
        // invalid-rules — skipping every value would make a typo look like a
        // deliberate setting, which is exactly how `falsehood` came to silently
        // disable every rule before the raw-text match was anchored.
        if (/^(true|false)[ \t]*(#.*)?$/.test(assignment[2].trim())) {
          continue;
        }
        return null;
      }
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

    // A malformed row is still a genuine syntax fault and stays fatal — only
    // the SLUG-SET checks (duplicate, unknown) move from `return null` to
    // collection, so the scan continues and every row still gets classified.
    const row = line.match(
      /^([a-z][a-z-]*)[ \t]*=[ \t]*\{[ \t]*active[ \t]*=[ \t]*(true|false)[ \t]*\}(?:[ \t]*#.*)?$/u,
    );
    if (!row) return null;
    if (rows.has(row[1])) {
      duplicate.push(row[1]);
      continue;
    }
    if (!slugSet.has(row[1])) {
      unknown.push(row[1]);
      continue;
    }
    rows.set(row[1], row[2] === "true");
  }

  // A missing strictness is no longer fatal (it used to fold into the same
  // `rows.size !== slugs.length` style all-or-nothing check this function
  // replaces): left unset here, it is the caller's job to default it — which
  // codex-context.mjs's posture selection already does (`strictness !== "firm"`
  // falls to adaptive), matching staleness.sh:558-560's `case "$strictness" in
  // firm) ... ; *) ... ;; esac`. A strictness that IS present but invalid still
  // fails closed: a typo must not silently pick a posture.
  const strictness = topLevel.get("strictness");
  if (
    !rulesSectionSeen ||
    (strictness !== undefined && strictness !== "firm" && strictness !== "adaptive")
  ) {
    return null;
  }

  const missing = slugs.filter((slug) => !rows.has(slug));
  const mismatch =
    missing.length === 0 && unknown.length === 0 && duplicate.length === 0
      ? null
      : { missing, unknown, duplicate };
  return { rows, mismatch };
}

// reconcileRows mirrors staleness.sh's reconciliation awk block byte-for-byte in
// its provenance strings (both hosts govern from the same rules.toml, so an agent
// reading the repair notice must see identical wording regardless of which host
// wrote it). Quarantine, never delete: an unknown or duplicate row is commented
// out with a dated note rather than dropped, so nothing a project chose is ever
// lost, and a payload upgrade that later re-recognises the slug is a one-line
// uncomment. Missing rows are appended, defaulted to `active = true`, under one
// shared header comment rather than one note per row (Ruling 6, TRL-20 task 3 —
// per-row notes on a firm, all-sixteen-missing file blew Codex's own
// MAX_CONTEXT_BYTES and reintroduced the blackout this exists to remove).
//
// `stamp` is the installed payload's own version stamp (`payload@<hash>`, no
// trailing newline) — what the note calls "not in <stamp>" / "missing from
// <stamp>". `today` is the caller's `YYYY-MM-DD` for the same reason
// staleness.sh takes one `date +%Y-%m-%d` call up front rather than one per row:
// every note in a single reconciliation shares one date.
function reconcileRows(source, slugs, stamp, today) {
  const want = new Set(slugs);
  const note =
    `  # quarantined ${today}: not in ${stamp}. If a newer Trellis` +
    " ships this slug, run `claude plugin update trellis@kodhama` and uncomment.";

  // Mirrors parseRulesToml's own newline handling: `\r?\n` consumes a CRLF pair
  // as one delimiter, so a raw line here never carries a trailing `\r`. Both
  // functions read the same source string and must agree on what a "line" is.
  const hadTrailingNewline = /\r?\n$/u.test(source);
  const rawLines = source.split(/\r?\n/u);
  if (hadTrailingNewline) rawLines.pop(); // a terminal newline is not an extra blank record — matches awk's own line semantics

  const rulesHeader = /^[ \t]*\[rules\][ \t]*(?:#.*)?$/u;
  const rowLead = /^[ \t]*(?:inv|floor)-[a-z-]+[ \t]*=/u;

  const seen = new Set();
  let hasRules = false;
  let quarantined = 0;
  const out = [];

  for (const line of rawLines) {
    if (rulesHeader.test(line)) hasRules = true;

    if (rowLead.test(line)) {
      // The slug is the row's first whitespace-delimited field, trimmed to its
      // leading [a-z-]+ run — mirrors the awk's `row = $1; sub(/[^a-z-].*$/, ""
      // , row)` exactly, so a row with no space before `=` still classifies
      // correctly.
      const field = line.replace(/^[ \t]+/u, "").match(/^\S+/u)?.[0] ?? "";
      const slug = field.match(/^[a-z-]+/u)?.[0] ?? field;
      if (!want.has(slug) || seen.has(slug)) {
        out.push(`# ${line}${note}`);
        quarantined += 1;
        continue;
      }
      seen.add(slug);
      out.push(line);
      continue;
    }

    out.push(line);
  }

  const missing = slugs.filter((slug) => !seen.has(slug));
  if (missing.length > 0) {
    if (!hasRules) {
      out.push("[rules]");
      hasRules = true;
    }
    out.push(`# added ${missing.length} row(s) below on ${today} (missing from ${stamp})`);
    for (const slug of missing) out.push(`${slug} = { active = true }`);
  }

  return { text: `${out.join("\n")}\n`, added: missing.length, quarantined };
}

// The slugs the payload actually ships, read from the same rules.md the Claude
// hook validates against (staleness.sh's own `want[]` scan uses the identical
// trailing-backtick anchor). A hardcoded list here could not be repaired by a
// plugin upgrade, and a stale one made a quarantine reason false.
function slugsFromRules(rulesMd) {
  const found = [];
  for (const line of rulesMd.split(/\r?\n/u)) {
    const m = line.match(/`((?:inv|floor)-[a-z-]+)`[ \t]*$/u);
    if (m) found.push(m[1]);
  }
  return found;
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
// Read the file directly rather than reusing configResult: that path is gated on
// MAX_CONTEXT_BYTES, so an oversized rules.toml made this check unreachable and
// Codex then told the model to go load the overlay — in a project that had
// declared itself ungoverned. An opt-out must not have a size limit.
//
// The BOM strip and the whitespace class are matched to the shell hook
// deliberately. They disagreed on four input classes (\v, \f, NBSP, \u2028) and
// on a leading BOM, and every disagreement meant the two hosts differed about
// whether a project was governed. A BOM in particular failed toward GOVERNING a
// project that had refused.
try {
  // Only the region BEFORE the first table header. `governed = false` appended
  // under `[rules]` is not a top-level key and must not opt out — D5 defines it
  // as top-level, and a raw multiline match honoured it anywhere in the file, so
  // a misplaced line silently disabled all sixteen rules instead of reaching
  // parseRulesToml and surfacing as invalid-rules.
  const raw = fs
    .readFileSync(path.join(projectRoot, PROJECT_CONFIG), "utf8")
    .replace(/^\uFEFF/, "")
    .split(/^[ \t\v\f]*\[/m)[0];
  // The value must be the COMPLETE token. Unanchored, `governed = falsehood`
  // read as an opt-out on both hosts and silently disabled every rule —
  // a typo is supposed to fail loudly as invalid-rules, not govern nothing.
  // ASCII horizontal whitespace only, matching the shell's `[[:space:]]` under
  // the C locale. `\s` and `[^\S...]` accept NBSP; POSIX `[[:space:]]` under
  // C.UTF-8 does not — so an NBSP-indented opt-out was honoured on Codex and not
  // on Claude, and the decision claimed both matched the same inputs. That is the
  // third time these two have diverged (classes, then BOM, now locale), which is
  // why the class is now written out rather than borrowed from a shorthand.
  const gov = /^[ \t\v\f\r]*governed[ \t\v\f\r]*=[ \t\v\f\r]*(\S*)[ \t\v\f\r]*(#[^\r\n]*)?$/gm;
  const values = [...raw.matchAll(gov)].map((m) => m[1]);
  if (values.length === 1 && values[0] === "false") {
    process.exit(0);
  }
} catch {
  // Unreadable here is not decisive; the normal error path below reports it.
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
if (trellis.split("@rules.md").length - 1 !== 1) {
  fail(".trellis/internal/trellis.md", "invalid-placeholder-count");
  process.exit(0);
}
// The rules payload's own well-formedness (the sentinel gate) must be checked
// BEFORE it is trusted enough to derive a slug set from it — moved ahead of
// that derivation for exactly this reason. Deriving first and validating after
// meant a broken rules.md (no sentinel, a truncated one, or a doubled one)
// yielded an empty or malformed slug set, parseRulesToml then failed on the
// PROJECT's .trellis/rules.toml, and the reported label blamed the project's
// config for a defect that was actually in the plugin's own payload.
if (
  rules.split(SENTINEL).length - 1 !== 1 ||
  !rules.endsWith(`${SENTINEL}\n`)
) {
  fail(".trellis/internal/rules.md", "invalid-rules");
  process.exit(0);
}
// Derived from the payload actually resolved above (vendored or plugin-native,
// whichever `sources` picked), not a hardcoded list — this is the row set a
// payload upgrade CAN repair, and it is what the row-count/row-membership
// checks inside parseRulesToml validate against.
//
// De-duplicated: parseRulesToml checks membership through a Set (slugSet) but
// checks completeness against slugs.length/slugs.some, so a rules.md that ever
// tagged one slug twice would make rows.size !== slugs.length permanently true
// — every Codex project would read .trellis/rules.toml: invalid-rules while
// Claude (whose want[] is already a set) kept governing normally from the same
// file. Not reachable with the current payload (every tag occurs exactly
// once), but it is the same blame-the-consumer mislabel this task exists to
// close, so the array is deduplicated at the source rather than trusted to
// stay duplicate-free forever.
const slugs = [...new Set(slugsFromRules(rules))];
// An EMPTY derived set is the Codex twin of staleness.sh's
// `no-slugs-in-payload` refusal, and it needs its own branch for the same
// reason: nothing downstream can tell it apart from a satisfied one. The
// sentinel gate above proves rules.md is well-formed, not that it TAGS any
// slug — a payload whose rule lines lost their trailing backticked slug keeps
// its sentinel and yields []. parseRulesToml then ACCEPTS a config carrying
// `strictness` plus an empty `[rules]` table, because both completeness checks
// pass vacuously (rows.size 0 === slugs.length 0, and slugs.some() over an
// empty array is false), and this hook emits a successful "loaded installed
// overlay" response with no activation rows in it at all: a silent governance
// blackout at exit 0 — the fail-loud invariant inverted, on the host where the
// blackout is hardest to notice because the response looks like success.
// Rejected here, before anything consumes `slugs` — including the floor-row
// warning below, which would also have nothing to filter.
if (slugs.length === 0) {
  fail(sources.rules, "no-slugs-in-payload");
  process.exit(0);
}
const parsed = parseRulesToml(rulesToml, slugs);
if (parsed === null) {
  fail(PROJECT_CONFIG, "invalid-rules");
  process.exit(0);
}
const { rows, mismatch } = parsed;

const stamp = version.endsWith("\n") ? version.slice(0, -1) : version;
// Reconcile rather than refuse (TRL-20, mirroring staleness.sh): a missing,
// unknown or duplicate row used to fail the whole file closed, so one bad row
// cost all sixteen rules every session until a human edited it by hand. The
// rows the payload ships are still the authority; what changes is that an
// unmatched row is quarantined instead of blocking delivery. `rows` (used below
// for the false-floor check) is unaffected either way: it already carries only
// the recognised, first-occurrence rows parseRulesToml collected.
let effectiveRulesToml = rulesToml;
if (mismatch !== null) {
  const today = new Date().toISOString().slice(0, 10);
  effectiveRulesToml = reconcileRows(rulesToml, slugs, stamp, today).text;
}
const context =
  trellis.replace("@rules.md", rules) +
  "\n" +
  effectiveRulesToml +
  (effectiveRulesToml.endsWith("\n") ? "" : "\n") +
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
// Floors are the `floor-` half of the same derived slug set, not a second
// hardcoded pair — the prefix is the product's own classification (matched the
// same way everywhere else this file and staleness.sh distinguish inv- from
// floor-), so a payload that ever ships a third floor picks it up here too.
const falseFloors = slugs
  .filter((slug) => slug.startsWith("floor-") && rows.get(slug) === false)
  .sort();
if (falseFloors.length > 0) {
  response.systemMessage =
    "Trellis warning: floor rows set active = false are overridden-by-floor and remain active: " +
    `${falseFloors.join(", ")}.`;
}
emit(response);

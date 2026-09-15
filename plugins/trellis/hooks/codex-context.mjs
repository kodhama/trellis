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
// commit 3490555 with none, and "8000" appears nowhere in docs/decisions/, docs/research/
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
// The read bound for the PROJECT's .trellis/rules.toml alone (TRL-97). Every
// payload read stays bounded by MAX_CONTEXT_BYTES because each payload file is
// delivered whole. The project file is echoed only while it and its warnings fit
// RULES_ECHO_MAX_BYTES and is otherwise only classified, so the context budget is
// the wrong bound for reading it: held at 9500 B, a file between that and
// staleness.sh's own 32768 B budget was refused here and governed there. One MiB
// is a runaway guard, roughly nine hundred times the largest consumer file known
// on 2026-09-14.
const MAX_PROJECT_CONFIG_BYTES = 1024 * 1024;
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
// The invariants pointer as the posture prose SHIPS it -- a placeholder each
// delivery channel resolves for the mode it is running, never an address in
// itself. Backticks included: the substitution below replaces the whole
// delivered span, code fence and all, so the model is never handed a half-
// rewritten path. Byte-identical to the token staleness.sh matches (its awk
// `tok`) and to install.sh's own sed pattern; the three are pinned to each
// other through the pointer they deliver, not through this literal.
const INVARIANTS_TOKEN = "`.trellis/internal/invariants.md`";

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

// The file-shaped sibling, and a PRESENCE test: absolute paths only, and any
// stat failure -- missing, unreadable, a dangling symlink -- reads as absent.
//
// TRL-72 narrowed what this may be asked. decision-0093 recorded the flattening
// as deliberate on the ground that "no caller here can act on the difference
// between missing and unreadable", and that stopped being true the moment #294
// added a caller that REPORTS the answer to a person: which fault it is decides
// which remedy the operator needs. Both callers that ask whether the invariants
// pointer resolves to something USABLE now go through payloadDefect below.
//
// What is left here is the one question presence actually answers: does this
// project have its own authoritative copy of the file. A zero-byte file at the
// authoritative path is still the project's own file, and not the plugin's to
// substitute for (README.md:27, decision-0093:3) -- so widening THAT arm would
// be a different ruling from the one this change makes, not the same one
// applied consistently.
function existingFile(value) {
  if (typeof value !== "string" || !path.isAbsolute(value)) return false;
  try {
    return fs.statSync(value).isFile();
  } catch {
    return false;
  }
}

// The four classifications staleness.sh's payload_read hands its own report,
// mirrored here BYTE FOR BYTE. Two hosts, two languages, one vocabulary: an
// operator who moves between them must not have to work out that "is empty" and
// some Codex-only paraphrase are the same finding, and the pair guard
// (TestBothHostsReportAMissingInvariantsTarget) can only compare reports that
// are stated in the same words. Change one of these and change the other.
const DEFECT_MISSING = "is missing";
const DEFECT_NOT_A_FILE = "is not a readable file — a directory or a device sits at that path";
const DEFECT_UNREADABLE =
  "exists but could not be read — a permission mode, a stale ACL, or a symlink whose target is gone";
const DEFECT_EMPTY = "is empty";

// payloadDefect answers "does this path yield something to read", and says
// WHICH way it does not. "" means it does. It is this hook's half of
// staleness.sh's payload_read gateway (decision-0087), reaching the same four
// answers on the same shapes -- deliberately, because the two hosts report the
// same broken install to the same person.
//
// existingFile cannot answer this and TRL-72 is the measurement of that:
// statSync needs no read permission, so it succeeds on a zero-byte copy and on
// one at mode 0000, and those are the two shapes Codex passed over in silence
// while Claude named them.
//
// READS, BUT NEVER HOLDS THE FILE. The bytes are scanned a chunk at a time and
// the scan stops at the first byte that is not a newline, so a healthy
// invariants.md costs one 4 KB read and nothing is retained.
// TestCodexHookBoundsAuthoritativeFileReads forbids pulling an AUTHORITATIVE
// file wholly into memory before its byte bound is enforced; this file is
// consulted rather than delivered, so it has no such bound to enforce, and the
// same discipline is kept anyway because there is no reason to read an
// arbitrarily large file to learn whether it is empty.
//
// EMPTY MEANS WHAT THE OTHER HOST MEANS BY IT, and that is TWO properties of
// command substitution rather than one. payload_read reaches "empty" through
// `payload_text="$(cat "$1")"` followed by `[ -z "$payload_text" ]`, and `$( )`
// strips both TRAILING NEWLINES and NUL BYTES -- the second because a shell
// variable holds a C string, which cannot carry a NUL at all. So the file is
// empty over there exactly when every byte is a newline or a NUL, which is what
// the scan below tests. A file holding one SPACE is not empty on either host.
//
// Both halves were measured on the shell the sibling hook actually runs
// (`#!/usr/bin/env bash`), and on sh and dash for the Linux CI runner: an
// all-NUL file is EMPTY on all three. zsh alone disagrees, and does not run
// either hook. Writing `readFileSync(...).length === 0` here would agree on the
// zero-byte file and diverge on both the newline-only and the NUL-filled one --
// the latter being the classic post-crash zero-fill, exactly the kind of
// half-written payload this report exists to name.
//
// What is deliberately NOT claimed: the MIXED cases. `a\0b` and `\0a` are
// non-empty on bash, sh and dash alike and agree with the scan below, but POSIX
// leaves NUL handling in `$( )` unspecified, so a shell that TRUNCATED at the
// first NUL rather than dropping it would call `\0a` empty and diverge. No such
// shell is in play; the guard pins the all-NUL case, which every shell measured
// agrees on, and leaves the mixed ones unpinned rather than asserting a
// portability property nobody has checked.
function payloadDefect(value) {
  if (typeof value !== "string" || !path.isAbsolute(value)) {
    return DEFECT_MISSING;
  }
  let stat;
  try {
    stat = fs.statSync(value);
  } catch {
    // Every stat failure, exactly as existingFile reads them and as `[ ! -e ]`
    // does on the other host: statSync follows symlinks, so a link whose target
    // is gone is missing here and missing there.
    return DEFECT_MISSING;
  }
  if (!stat.isFile()) return DEFECT_NOT_A_FILE;
  let fd;
  try {
    fd = fs.openSync(value, "r");
  } catch {
    return DEFECT_UNREADABLE;
  }
  try {
    const chunk = Buffer.alloc(4096);
    for (;;) {
      let read;
      try {
        read = fs.readSync(fd, chunk, 0, chunk.length, null);
      } catch {
        // A file can open and still fail to read -- a stale network mount, a
        // device that refuses. `cat` fails the same way, and payload_read reads
        // that as unreadable rather than empty.
        return DEFECT_UNREADABLE;
      }
      if (read === 0) return DEFECT_EMPTY;
      for (let i = 0; i < read; i += 1) {
        // 0x0a and 0x00 are the two byte values `$( )` discards, so a byte
        // outside that pair is one the other host would still be holding when
        // it asks `[ -z ... ]`.
        if (chunk[i] !== 0x0a && chunk[i] !== 0x00) return "";
      }
    }
  } finally {
    try {
      fs.closeSync(fd);
    } catch {
      // The answer is already decided; a descriptor that will not close cannot
      // change it, and this hook must never throw on the way to its one write.
    }
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

// The third argument is the point of this signature, and TRL-33 is why it
// exists. readRequired has always been the model the Claude hook is now built
// to match — it reports missing-file and unreadable-file where staleness.sh
// used to swallow both — but a ZERO-BYTE file came back as { value: "" }, a
// SUCCESS. Emptiness was caught only by post-checks each caller remembered to
// write (empty-prose for the prose and the rules, the version regex for the
// stamp), and the project config had none at all. That is guarded by
// remembering, which is precisely the failure mode decision-0083 and
// decision-0084 kept shipping fixes for, one instance at a time.
//
// So the DEFAULT is now loud: a call that says nothing about emptiness gets
// empty-file. A call site where an empty read is legitimate says so in its own
// source, where a reader sees it, rather than by omission.
//
//   options.emptyError   the failure class to report for a zero-byte read
//   options.emptyIsValid true when empty is a supported state for this file
//   options.maxBytes     the read bound, MAX_CONTEXT_BYTES unless the caller
//                        states its own in its own source
//
// Every existing failure class is preserved byte for byte: the callers pass the
// classes their post-checks already produced, so this is a structural change
// with no behaviour change. TestEveryReadRequiredStatesWhatEmptyMeans holds it.
function readRequired(projectRoot, relativePath, options = {}) {
  const absolute = path.join(projectRoot, relativePath);
  const maxBytes = options.maxBytes ?? MAX_CONTEXT_BYTES;
  // A payload file over MAX_CONTEXT_BYTES could never fit the context it is
  // read into, so the refusal names the context. A read with its own bound is
  // bounded for its own sake, and the refusal names the file that crossed it.
  const overBound = {
    label: options.maxBytes === undefined ? "assembled-context" : relativePath,
    error: "context-over-budget",
  };
  let stat;
  try {
    stat = fs.statSync(absolute);
  } catch (error) {
    if (error?.code === "ENOENT") return { error: "missing-file" };
    return { error: "unreadable-file" };
  }
  if (!stat.isFile()) return { error: "unreadable-file" };
  if (stat.size > maxBytes) return overBound;
  let descriptor;
  try {
    fs.accessSync(absolute, fs.constants.R_OK);
    descriptor = fs.openSync(absolute, "r");
    const openedStat = fs.fstatSync(descriptor);
    if (!openedStat.isFile()) return { error: "unreadable-file" };
    if (openedStat.size > maxBytes) return overBound;

    const buffer = Buffer.alloc(maxBytes + 1);
    let total = 0;
    while (total < buffer.length) {
      const count = fs.readSync(descriptor, buffer, total, buffer.length - total, null);
      if (count === 0) break;
      total += count;
    }
    if (total > maxBytes) return overBound;
    const value = buffer.subarray(0, total).toString("utf8");
    if (value.length === 0 && options.emptyIsValid !== true) {
      return { error: options.emptyError ?? "empty-file" };
    }
    return { value };
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

// KTD3's one row-classification contract (TRL-97). staleness.sh implements the
// same contract in awk, and TestBothHostsClassifyRulesRowsIdentically runs both
// hooks on one table, so a line class the two read differently is a red test
// rather than two hosts governing one project differently.
//
// Only a row set `active = false` has any effect, and one is enough: a rule is
// off when ANY row for it under [rules] says false, whatever its other rows say,
// so a repeated row is never an error and draws no warning. Nothing here refuses
// the file: a bad entry costs that entry and is reported, never the session,
// because over-governance the session announces beats under-governance nobody
// sees (decision-0083 §5). The classes, in the order they are tested:
//
//   blank, or `#` after trimming   nothing
//   `[rules]` header               opens the table; a second one continues it, warned
//   any other `[` header           warned once; every line under it is ignored silently
//   above every header             a row lead is a row outside [rules], warned;
//                                  strictness and seeded_from are silent; governed is
//                                  silent only as a first `governed = true`; any other
//                                  key or line is warned
//   under [rules]                  a false row for a shipped, non-floor slug switches
//                                  that rule off; a true row is silent; a false row
//                                  for a floor or an unknown slug is warned; any other
//                                  line is warned
//
// A warning that names a slug (a floor row set false, an unknown slug set false, a
// row above [rules]) is noted once per kind and slug, at the first line that draws
// it, so a repeated row adds none. An entry is an ATTEMPT when a well-formed false
// row for a shipped slug sits in it and takes no effect: a floor row set false, or
// such a row above [rules]. ruleWarnings names attempts first.
//
// Whitespace is ASCII space and tab only, as on the other host under LC_ALL=C,
// so an NBSP-indented row is malformed on both. Lines split on LF with one
// trailing CR stripped and a line-1 BOM removed; a CR-only file is one line.
//
// `slugs` is the set the payload ships (slugsFromRules below), passed in rather
// than hardcoded so a plugin upgrade can repair it; floors are its `floor-` half.
const ROW_PATTERN =
  /^((?:inv|floor)-[a-z-]+)[ \t]*=[ \t]*\{[ \t]*active[ \t]*=[ \t]*(true|false)[ \t]*\}(?:[ \t]*#[\s\S]*)?$/u;
const ROW_LEAD_PATTERN = /^((?:inv|floor)-[a-z-]+)[ \t]*=/u;
const BARE_KEY_PATTERN = /^([A-Za-z0-9_-]+)[ \t]*=[ \t]*([\s\S]*)$/u;
const RULES_HEADER_PATTERN = /^\[[ \t]*rules[ \t]*\](?:[ \t]*#[\s\S]*)?$/u;
// One token and an optional comment. Anything else a `governed` line holds is
// not `true`, so it is warned rather than read.
const GOVERNED_VALUE_PATTERN = /^([^ \t]*)[ \t]*(?:#[\s\S]*)?$/u;
// The warning kinds that name a slug, each noted once per slug.
const SLUG_WARNING_KINDS = new Set(["floor-row", "unknown-slug", "row-above-rules"]);

function classifyRules(source, slugs) {
  const shipped = new Set(slugs);
  const offSet = new Set();
  const bySlug = new Map();
  const entries = [];
  let section = "top";
  let rulesOpened = false;
  let governedSeen = false;

  // A later row of a slug-naming kind adds no entry; it can only make the first
  // one an attempt, which is what orders the warnings (KTD4).
  const note = (line, kind, name, attempt = false) => {
    if (SLUG_WARNING_KINDS.has(kind)) {
      const first = bySlug.get(`${kind} ${name}`);
      if (first !== undefined) {
        if (attempt) first.attempt = true;
        return;
      }
      const entry = { line, kind, name, attempt };
      bySlug.set(`${kind} ${name}`, entry);
      entries.push(entry);
      return;
    }
    entries.push({ line, kind, name, attempt });
  };

  source
    .replace(/^\uFEFF/u, "")
    .split("\n")
    .forEach((raw, index) => {
      const lineNo = index + 1;
      const line = raw.replace(/\r$/u, "").replace(/^[ \t]+|[ \t]+$/gu, "");
      if (line === "" || line.startsWith("#")) return;

      if (line.startsWith("[")) {
        if (RULES_HEADER_PATTERN.test(line)) {
          if (rulesOpened) note(lineNo, "rules-again");
          rulesOpened = true;
          section = "rules";
        } else {
          section = "other";
          note(lineNo, "other-table");
        }
        return;
      }
      if (section === "other") return;

      const row = line.match(ROW_PATTERN);
      const key = line.match(BARE_KEY_PATTERN);
      const shippedFalse = row !== null && row[2] === "false" && shipped.has(row[1]);

      if (section === "top") {
        const lead = line.match(ROW_LEAD_PATTERN);
        if (lead !== null) {
          note(lineNo, "row-above-rules", lead[1], shippedFalse);
        } else if (key === null) {
          note(lineNo, "top-level-line");
        } else if (key[1] === "governed") {
          // The opt-out read above already exited on the one shape that opts
          // out, so any `governed` line reaching here did not, unless it is the
          // first and says `true`.
          const value = key[2].match(GOVERNED_VALUE_PATTERN);
          if (governedSeen || value === null || value[1] !== "true") note(lineNo, "governed");
          governedSeen = true;
        } else if (key[1] !== "strictness" && key[1] !== "seeded_from") {
          note(lineNo, "unknown-key", key[1]);
        }
        return;
      }

      if (row === null) {
        if (key !== null && key[1] === "governed") note(lineNo, "governed");
        else note(lineNo, "malformed-row", key === null ? undefined : key[1]);
        return;
      }
      const [, slug, value] = row;
      if (value === "true") return;
      if (!shipped.has(slug)) note(lineNo, "unknown-slug", slug);
      else if (slug.startsWith("floor-")) note(lineNo, "floor-row", slug, true);
      else offSet.add(slug);
    });

  return { off: slugs.filter((slug) => offSet.has(slug)), entries };
}

// The warnings both hooks deliver, and their bound (KTD4). staleness.sh carries
// the same texts as printf format literals inside its payload block, so none of
// them may hold an apostrophe, a backslash, or a percent sign beyond the value
// it substitutes. None quotes the raw line: a warning names the line number and
// a slug, a key, or only the kind of entry, so non-ASCII bytes in a project file
// cannot make the two hosts word one differently.
//
// No text here asks the agent to change .trellis/rules.toml. Nothing reconciles
// the file any more, so nothing needs writing back, and the destructive- and
// deletion-instruction guards scan this function's literals
// (codexPayloadFunctions, cli/plugin_hook_test.go).
//
// WARNINGS_NAMED warnings are named in total: the attempts first, in file line
// order, then every other warning in file line order. Past that, one count line
// says how many more entries were ignored and how many of those are attempts. A
// name taken from the file is cut at WARNING_NAME_MAX bytes, so a runaway file
// cannot turn its warnings into the context.
//
// A file reached through a symbolic link is named nowhere. Its target can be any
// file the user can read, so no key, slug or line number read from it is
// repeated: one count line stands in for every warning, and a file with nothing
// ignored draws none.
const WARNINGS_NAMED = 5;
const WARNING_NAME_MAX = 40;

function ruleWarnings(entries, symlink) {
  const attempts = entries.filter((entry) => entry.attempt).length;
  if (symlink) {
    const total = entries.length;
    const linked = "Trellis warning: .trellis/rules.toml is a symbolic link, so its ";
    if (total === 0) return [];
    if (total === 1) {
      return [
        attempts === 1
          ? `${linked}1 ignored entry is counted but not named; it tries to switch a rule off.`
          : `${linked}1 ignored entry is counted but not named; it does not try to switch a rule off.`,
      ];
    }
    if (attempts === 0) {
      return [
        `${linked}${total} ignored entries are counted but not named; none of them tries to switch a rule off.`,
      ];
    }
    if (attempts === 1) {
      return [
        `${linked}${total} ignored entries are counted but not named; 1 of them tries to switch a rule off.`,
      ];
    }
    return [
      `${linked}${total} ignored entries are counted but not named; ${attempts} of them try to switch a rule off.`,
    ];
  }
  const ordered = [
    ...entries.filter((entry) => entry.attempt),
    ...entries.filter((entry) => !entry.attempt),
  ];
  const warnings = [];
  for (const entry of ordered.slice(0, WARNINGS_NAMED)) {
    const name =
      entry.name !== undefined && entry.name.length > WARNING_NAME_MAX
        ? `${entry.name.slice(0, WARNING_NAME_MAX)}...`
        : entry.name;
    const at = `Trellis warning: line ${entry.line} of .trellis/rules.toml `;
    switch (entry.kind) {
      case "nul-byte":
        warnings.push(
          "Trellis warning: .trellis/rules.toml contains a NUL byte, so none of its rows take effect and it is not shown; every rule applies.",
        );
        break;
      case "unknown-slug":
        warnings.push(
          `${at}sets ${name} to active = false, but this plugin ships no rule by that name, so the row is ignored; another plugin version may ship it.`,
        );
        break;
      case "floor-row":
        warnings.push(
          `${at}sets the floor rule ${name} to active = false, but floor rules cannot be switched off, so the row is ignored and the rule applies.`,
        );
        break;
      case "row-above-rules":
        warnings.push(
          `${at}is a row for ${name} above the [rules] table, so it is ignored; rows count only under [rules].`,
        );
        break;
      case "rules-again":
        warnings.push(`${at}opens [rules] a second time; the rows under it still count.`);
        break;
      case "other-table":
        warnings.push(`${at}opens a table other than [rules], so every line under it is ignored.`);
        break;
      case "governed":
        warnings.push(
          `${at}sets governed but does not opt out; only one governed = false line, above every table, opts a project out, so this project stays governed.`,
        );
        break;
      case "unknown-key":
        warnings.push(
          `${at}sets ${name}, which is not a key this file defines, so it is ignored and switches no rule off; a row under [rules] such as slug = { active = false } switches a rule off.`,
        );
        break;
      case "top-level-line":
        warnings.push(
          `${at}is not an entry this file defines, so it is ignored and switches no rule off.`,
        );
        break;
      default:
        warnings.push(
          name === undefined
            ? `${at}is a malformed row, so it is ignored and switches no rule off.`
            : `${at}is a malformed row naming ${name}, so it is ignored and switches no rule off; any rule it names still applies.`,
        );
    }
  }
  const more = ordered.length - WARNINGS_NAMED;
  if (more > 0) {
    // The attempts are named first, so an attempt is counted only once every
    // named slot holds one.
    const tries = attempts - Math.min(attempts, WARNINGS_NAMED);
    const counted = "Trellis warning: .trellis/rules.toml has ";
    if (more === 1) {
      warnings.push(
        tries === 1
          ? `${counted}1 more ignored entry, and it tries to switch a rule off.`
          : `${counted}1 more ignored entry, and it does not try to switch a rule off.`,
      );
    } else if (tries === 0) {
      warnings.push(
        `${counted}${more} more ignored entries, none of which tries to switch a rule off.`,
      );
    } else if (tries === 1) {
      warnings.push(
        `${counted}${more} more ignored entries, 1 of which tries to switch a rule off.`,
      );
    } else {
      warnings.push(
        `${counted}${more} more ignored entries, ${tries} of which try to switch a rule off.`,
      );
    }
  }
  return warnings;
}

// B, the bound on what the project file may add to the context (TRL-97). The file
// is echoed verbatim under the framing only while its decoded bytes plus the bytes
// of its warning block (every warning line with its newline, the count line
// included) fit this; otherwise one line stands in for it, and the computed
// sentence still states which rules it switches off. staleness.sh shares the
// value, and the parity table pins both hosts to it.
//
// MEASURED, NOT CHOSEN, because what it protects is this hook's context budget:
// no project file may ever cost a Codex session its rules. The largest section
// the bound admits is every non-floor rule switched off (the longest sentence)
// beside a file and warnings of exactly this many bytes, the file with no
// trailing newline so one more is supplied. Measured on 2026-09-15 on a 74-byte
// plugin root, that context is 9153 bytes, 347 under MAX_CONTEXT_BYTES; the
// too-large branch peaks at 8869 (the too-large line beside five of the longest
// warning at seven-digit line numbers and a count line). 1900 left 247, under
// the 250 bytes held back for a longer plugin root path, which appears once in
// the prose, so 1800 is the largest round value that keeps that margin. It is
// below the 2.5 KB first approved, and every consumer file known on 2026-09-14
// (1.1-1.2 KB) is still shown beside a few warnings.
// TestBothHostsClassifyRulesRowsIdentically builds both largest sections and
// fails when either leaves less than 250 bytes.
const RULES_ECHO_MAX_BYTES = 1800;

// activationSection builds what follows the rules prose on both branches (KTD1,
// KTD7): the heading, the computed sentence, the file or the line that stands in
// for it, and the warnings, one blank line apart. The sentence is what tells the
// model the effective result, so it stays right where the file still shows a row
// this hook ignored, and under a vendored overlay whose frozen text applies a rule
// only when its row says `active = true`.
//
// A file holding a NUL byte is not classified (KTD5). It is not text both hosts
// split into the same lines, so none of its rows takes effect, it is not echoed,
// and one warning says why. The `governed = false` read runs before this, so an
// opted-out file with a NUL byte still loads nothing.
//
// A file reached through a symbolic link (`symlink`) is classified like any other,
// so its opt-outs apply and the sentence names them from the payload's own slugs,
// but nothing read from it is shown, at any size: one line stands in for it, and
// its warnings are one count (ruleWarnings). A NUL byte keeps its own handling.
//
// The bound charges the file what it would add to the context: its bytes once
// decoded, not as read. The two differ only for a byte that is not valid UTF-8,
// which decodes to U+FFFD, three bytes; charging the bytes read let a file of
// such bytes fit the bound and still push the context past MAX_CONTEXT_BYTES,
// costing the session every rule. staleness.sh charges the bytes read, and its
// budget is never reached, so the hosts differ only for such a file (TRL-100).
function activationSection(rulesToml, slugs, symlink) {
  const nulByte = rulesToml.includes("\u0000");
  const { off, entries } = nulByte
    ? { off: [], entries: [{ kind: "nul-byte", attempt: false }] }
    : classifyRules(rulesToml, slugs);
  const sentence =
    off.length === 0
      ? "The project file .trellis/rules.toml switches no rule off, so every rule above applies."
      : `The project file .trellis/rules.toml switches these rules off: ${off.join(", ")}. Every other rule above applies.`;
  const warnings = ruleWarnings(entries, symlink && !nulByte);
  const warningBytes =
    warnings.length === 0 ? 0 : Buffer.byteLength(`${warnings.join("\n")}\n`, "utf8");
  let segment = "";
  if (nulByte) {
    segment = "";
  } else if (symlink) {
    segment =
      "The project file is a symbolic link, so its contents are not shown here; the sentence above names every rule it switches off.\n";
  } else if (Buffer.byteLength(rulesToml, "utf8") + warningBytes > RULES_ECHO_MAX_BYTES) {
    segment =
      "The project file and its warnings are too large to show here; the sentence above names every rule it switches off.\n";
  } else if (rulesToml !== "") {
    segment = rulesToml.endsWith("\n") ? rulesToml : `${rulesToml}\n`;
  }
  const parts = [`${sentence}\n`];
  if (segment !== "") parts.push(segment);
  if (warnings.length > 0) parts.push(`${warnings.join("\n")}\n`);
  return { text: `## Project rule activation\n\n${parts.join("\n")}`, warnings };
}

// The slugs the payload actually ships, read from the same rules.md the Claude
// hook validates against (staleness.sh's own `want[]` scan uses the identical
// trailing-backtick anchor). A hardcoded list here could not be repaired by a
// plugin upgrade.
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

// emptyIsValid, and this is the ONE read in this file that gets it. A
// zero-byte .trellis/rules.toml is a PROJECT file, not payload, and it means
// every rule applies (TRL-97). Refusing it would be the over-correction — the
// same failure direction as the CRLF and unreadable-preset blackouts the Claude
// hook shipped and had to withdraw. maxBytes, and this is also the one read
// with its own bound: see MAX_PROJECT_CONFIG_BYTES.
const configResult = readRequired(projectRoot, PROJECT_CONFIG, {
  emptyIsValid: true,
  maxBytes: MAX_PROJECT_CONFIG_BYTES,
});
// decision-0070 D5. A project that declares `governed = false` is not governed —
// on EITHER host. Checked here, before the rules are parsed or assembled, for the
// same reason the Claude hook checks it before every delivery path: an opt-out
// that only one host honours is not an opt-out. Matched on the raw text rather
// than through the classifier, because the opt-out has to be settled before
// anything else reads the file, whatever else the file holds.
// Read the file directly rather than reusing configResult: that path is bounded
// (it was once MAX_CONTEXT_BYTES), so an oversized rules.toml made this check unreachable and
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
  // the classifier, which warns that it does not opt out.
  const raw = fs
    .readFileSync(path.join(projectRoot, PROJECT_CONFIG), "utf8")
    .replace(/^\uFEFF/, "")
    .split(/^[ \t\v\f]*\[/m)[0];
  // The value must be the COMPLETE token. Unanchored, `governed = falsehood`
  // read as an opt-out on both hosts and silently disabled every rule —
  // a typo is supposed to be reported as one, not govern nothing.
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
// One header ships on the plugin-native branch (TRL-97): `strictness` selects
// nothing, and every project receives the By default posture sentence.
const sources = vendored
  ? { root: projectRoot, ...VENDORED_PAYLOAD }
  : {
      root: pluginRoot,
      prose: "reference/trellis.md",
      rules: "reference/rules.md",
      version: "reference/version",
    };

// Each payload read states what an empty file means, in the class the
// post-check below it already produced. Nothing observable changes; what
// changes is that the statement is now at the READ, so a fourth key added to
// this loop without one is refused as empty-file rather than delivered as "".
const PAYLOAD_EMPTY_CLASS = {
  prose: "empty-prose",
  rules: "empty-prose",
  version: "invalid-version",
};

const payload = {};
for (const key of ["prose", "rules", "version"]) {
  const result = readRequired(sources.root, sources[key], {
    emptyError: PAYLOAD_EMPTY_CLASS[key] ?? "empty-file",
  });
  if (result.error) {
    fail(result.label ?? sources[key], result.error);
    process.exit(0);
  }
  payload[key] = result.value;
}

// `let`, uniquely among the three: the invariants repoint below rewrites it
// in place on the plugin-native branch (TRL-52). The rules body and the
// version stamp are delivered exactly as read and stay `const`.
let trellis = payload.prose;
const rules = payload.rules;
const version = payload.version;

// The three length/shape checks that follow are now the SECOND lock on the
// same door — readRequired refuses a zero-byte file before they run. They stay
// because they catch what emptiness cannot: a version stamp that is present
// and malformed, and prose that is non-empty and truncated.
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
// sources.prose, not a hardcoded path. This named .trellis/internal/trellis.md
// on BOTH branches, so on the plugin-native path it reported a failure against
// a file that was never read — the file actually read is
// reference/trellis.md. Same class of defect as every message this
// change set corrects: right diagnosis, wrong file named. On the vendored
// branch sources.prose IS that path, so nothing the tests assert moves.
if (trellis.split("@rules.md").length - 1 !== 1) {
  fail(sources.prose, "invalid-placeholder-count");
  process.exit(0);
}
// TRL-52. decision-0065:106-111 defines the plugin-native delivery as the
// resolved @rules.md import PLUS one more edit: "the one edit repoints the
// invariants pointer at the plugin's own copy, which is where the file is in
// this mode and which therefore cannot go stale." That rule is scoped to a
// SHAPE -- rules.toml and no .trellis/internal/ -- not to a host, and this
// hook delivers that shape on Codex. It performed the import edit above and
// not this one, so every plugin-native Codex session was handed
// `.trellis/internal/invariants.md`: a directory this mode is DEFINED by not
// having. staleness.sh:977-1005 makes the same substitution for Claude, at
// the same target; TestBothHostsRepointTheInvariantsPointerIdentically runs
// both hooks on one project and pins them to each other, because the two
// implementations share no line to diff (awk there, JS here) and only the
// address they hand the model can be compared.
//
// Gated on the prose's ACTUAL origin rather than on `vendored`, so a third
// source added to `sources` cannot inherit this edit by default. The gate is
// load-bearing in the vendored direction, not incidental: a vendored overlay
// really does carry .trellis/internal/invariants.md -- the retired setup skill
// copied it in beside trellis.md and rules.md, and decision-0051 specifies it
// as part of the trellis-authoritative half -- and
// README.md:27 rules that where such an overlay exists "the plugin's
// reference/ files stay installation sources rather than runtime
// substitutes". Repointing there would override a real, authoritative file
// with the plugin's possibly-newer copy, which is that substitution exactly.
// So the token survives on the vendored branch, which is also what Claude
// does there by never injecting into a vendored project at all.
//
// split/join, not replace(): a bare string replace() substitutes only the
// FIRST occurrence, and `$&`-style patterns in the replacement are
// interpreted. Neither can bite today (the prose carries one pointer and the
// path holds no `$`), but the pointer is payload and the plugin root is a
// user path -- the same two assumptions staleness.sh's own escaping comment
// records having been wrong about.
const pluginInvariants = path.join(pluginRoot, "reference", "invariants.md");
const repointedInvariants = `\`${pluginInvariants}\``;
// TRL-69. The two arms below ask DIFFERENT questions, which is why the same
// answer is put to different use in each, and why making them textually
// symmetric would be a regression rather than a tidy-up. The vendored arm's
// check is fallback ELIGIBILITY: that project has its own authoritative
// location for this file, so the plugin's copy may stand in for it only when
// the plugin's copy is really there (decision-0093). This arm has no competing
// location to substitute FOR -- the repointed path is the only address there
// is -- so the same check here would not validate the target, it would abandon
// it. Leaving
// the raw token ships `.trellis/internal/invariants.md` to the one mode
// DEFINED by not having that directory: both pointers are dead when the copy
// is missing, and only that one ALSO tells the model the project has an
// overlay it does not have. So the pointer moves either way.
//
// Dropping the sentence was rejected on a different ground. The token sits
// mid-sentence in shipped prose, so removing "the sentence" means guessing
// sentence boundaries in a payload this hook is otherwise forbidden to
// rewrite -- and it would leave the model with no idea the invariants exist
// at all. A dead pointer is a broken lead; no pointer is a governance loss.
//
// What a missing copy earns instead is a REPORT, carried at the end of this
// file on the systemMessage channel the rules.toml warnings also use. Not a
// fail(): invariants.md is consulted on demand, never injected, and
// decision-0093:1 rules that failing a session closed over such a file trades
// a dead pointer for no governance at all. The context still delivers whole.
//
// Scoped to this branch deliberately. Here the plugin's reference/ IS the
// delivery source this run just read three payload files from, so a missing
// fourth is this delivery's own defect and it names the install to repair. On
// the vendored branch it is not the source: a plugin root with no reference/
// at all is a legitimate shape there -- most of the suite's own vendored
// fixtures are exactly that (writeCodexPluginRoot), and
// TestCodexHookValidStartupAndLiveRows pins them silent.
//
// TRL-72. Both arms ask payloadDefect rather than existingFile, and the second
// is not a drive-by: decision-0093:2 states the plugin-copy half of the
// vendored condition as the thing that "stops the fallback replacing one dead
// pointer with another", and a bare stat does not deliver that. Measured, a
// zero-byte or mode-0000 plugin copy satisfied existingFile, the fallback
// fired, and a vendored project's pointer was moved off its own authoritative
// address onto a file that yields nothing -- the exact substitution that half
// of the condition exists to prevent. Asking whether the copy is USABLE is what
// 0093:2 already says it wants, not a widening of it.
//
// Where the two arms still part is what they do with the answer. Here a defect
// is REPORTED and the pointer moves anyway, because there is no second address
// to fall back to. There a defect is disqualifying, because there is: the
// overlay's own path, which stays named.
//
// TRL-71 is the third arm below, and it is the cell the two above leave. When
// the overlay has no invariants.md AND the plugin's copy is unusable, NEITHER
// half of the fallback can help: there is no second address to move to, so the
// token survives naming a file that is not there -- and until this change the
// hook knew that and said nothing. What it earns is the same REPORT the
// plugin-native arm earns, on the same channel, with one difference that
// #294's scoping argument dictates rather than contradicts. That argument --
// "on the vendored branch it is not the source: a plugin root with no
// reference/ at all is a legitimate shape there" -- is sound about WORDING and
// silent about SILENCE. It says the plugin install is the wrong thing to blame
// here, not that nothing is wrong: the defect the reader must repair is the
// OVERLAY's missing file, and the plugin's copy is named only as the
// substitute that was unavailable.
let pluginInvariantsDefect = "";
// The overlay's own authoritative address. Hoisted because two arms need it
// and it is a pure path join -- no read is moved earlier by naming it here.
const overlayInvariants = path.join(projectRoot, ".trellis", "internal", "invariants.md");
let vendoredInvariantsOverlay = "";
let vendoredInvariantsOverlayDefect = "";
let vendoredInvariantsPluginDefect = "";
// TRL-73's own, kept separate from the two above because it is a DIFFERENT
// report with a different remedy -- see the fourth arm and its warning below.
let ownOverlayInvariantsDefect = "";
if (sources.root === pluginRoot) {
  trellis = trellis.split(INVARIANTS_TOKEN).join(repointedInvariants);
  pluginInvariantsDefect = payloadDefect(pluginInvariants);
} else if (!existingFile(overlayInvariants)) {
  // ONE read, shared by the two arms below, and that is a correctness property
  // rather than a saving. Asking payloadDefect twice let the answer CHANGE
  // between the eligibility test and the report: a plugin copy restored between
  // the two calls yielded a defect of "" on the second, and the report -- guarded
  // on the overlay path rather than on the defect -- still fired, rendering
  // `there is no readable <path> either (it ).` Reading once cannot disagree with
  // itself. The read still happens only when the overlay has no copy of its own,
  // because that is what this branch tests.
  const pluginCopyDefect = payloadDefect(pluginInvariants);
  if (pluginCopyDefect === "") {
    // TRL-58: the vendored overlay that has no invariants.md. The paragraph above
    // argues the token must SURVIVE on the vendored branch, and it holds wherever
    // the file is there -- but the overlay's writer was the retired setup skill:
    // model-executed instructions rather than a script, so an overlay whose
    // session skipped the `cp` is a shape nothing rules out. There the surviving
    // token is TRL-52's defect verbatim: a last line telling the model to read a
    // file that is not there.
    //
    // This is not the silent fall-through :1025-1030 forbids, on two counts. That
    // rule governs the three DELIVERED files, whose absence makes the injected
    // chain itself wrong; invariants.md is CONSULTED -- read on demand when a rule
    // seems ambiguous, never injected -- so a session that meets no ambiguous rule
    // never opens it, and failing the session closed over it would trade a dead
    // pointer for no governance at all. And nothing switches mode: prose, rules and
    // version still come from `sources`, which is untouched. Only the pointer moves.
    //
    // README.md:27 is not in tension either. "Installation sources rather than
    // runtime substitutes" presupposes something to substitute FOR; this arm runs
    // only when there is not. The plugin's copy is checked for USABILITY too
    // (decision-0094:4), so the fallback cannot replace one dead pointer with
    // another.
    trellis = trellis.split(INVARIANTS_TOKEN).join(repointedInvariants);
  } else {
    // TRL-71. The plugin's copy is unusable and the overlay has none, so both
    // addresses are dead and the pointer stays at the overlay's -- which on this
    // branch is not a lie: the project really does have a `.trellis/internal/`.
    // decision-0094:4 widened this cell and said so ("TRL-71's population widens
    // slightly, and this is stated rather than discovered later"), naming TRL-71
    // as its consumer.
    //
    // BOTH halves are classified, which applies decision-0094's ruling rather
    // than extending it. Its Decision HEADLINE, above the numbered points, is
    // "A host that cannot say which fault it found has not reported the fault"
    // -- the headline, not point 1, which is about which host changes.
    // Classifying only the plugin's copy would leave the file the
    // reader must actually repair as the one described least precisely -- and it
    // is not hypothetical, because existingFile asks statSync(...).isFile(): a
    // directory at the overlay path is not a present FILE and so reaches this
    // arm, where "restore invariants.md" is the wrong instruction until the
    // directory is gone.
    //
    // This does NOT widen the eligibility check decision-0094:5 reserves. That
    // check is the `existingFile(overlayInvariants)` above and it is untouched;
    // what is classified is the REPORT. The distinction matters for TRL-73: an
    // overlay whose own copy is present but unusable never reaches this arm at
    // all, because existingFile succeeds on it.
    vendoredInvariantsOverlay = overlayInvariants;
    vendoredInvariantsOverlayDefect = payloadDefect(overlayInvariants);
    vendoredInvariantsPluginDefect = pluginCopyDefect;
  }
} else {
  // TRL-73, the last cell of the table decision-0093 opened: the overlay's OWN
  // copy is present but unusable -- zero-byte, newline-only, NUL-filled, or
  // mode 0000. existingFile succeeds on every one of those, so the arm above
  // is not reached, the pointer stays on the overlay's own address, and until
  // decision-0096 nothing was said about it.
  //
  // THE ELIGIBILITY CHECK ABOVE IS UNTOUCHED, and that is decision-0096:1
  // rather than an omission. decision-0094:5 reserved "whether an overlay's
  // unusable copy should be substituted for" and decision-0096 answers it NO,
  // on the ground decision-0094:5 itself named: THE PROJECT HAS ITS OWN
  // AUTHORITATIVE LOCATION FOR THIS FILE, AND IT IS OCCUPIED. Substitution
  // presupposes something to substitute FOR (decision-0093:3); here there is.
  // Repointing would move a vendored project off the address it chose onto one
  // the plugin controls -- the runtime substitution README.md:32 forbids (:27
  // carries the authoritativeness claim; :32 the prohibition, and the three
  // older citations in this file conflate them) -- and
  // would do it INVISIBLY, where the dead pointer it replaces is at least
  // visible to whoever opens it. So `existingFile` above stays a presence test
  // and this arm moves nothing.
  //
  // THE DIFFERENT-VERSION ARGUMENT IS AN INSTANCE, NOT THE GROUND, and this
  // comment said otherwise until review caught it against the record it cites.
  // It reaches the mode-0000 shape, where the file has contents this hook cannot
  // read and so MAY hold a different version of the invariants. It does not
  // reach the zero-byte, newline-only or NUL-filled shapes -- those hold
  // nothing, and the hook has just proved it -- which is exactly why
  // decision-0096:1 rests the ruling on the ADDRESS rather than on the bytes.
  // An overlay that owns the location owns it whether or not this session found
  // anything at it.
  //
  // WHAT IS DECIDED IS THE REPORT, which is the line decision-0095:3 already
  // drew on the sibling cell ("This widens no eligibility check ... What
  // classifies is the REPORT"). The two questions are independent, and this
  // record answers them in opposite directions.
  //
  // The objection this arm has to answer is that the overlay's copy is the
  // PROJECT's file, so remarking on it is commentary on the project's own
  // content. It does not survive what payloadDefect can actually see: it reads
  // no content, only whether the path yields a byte that is not a newline or a
  // NUL. A zero-byte invariants.md is not an editorial choice about invariants,
  // and the hook has JUST TOLD THE MODEL TO READ IT. That makes the report a
  // statement about this delivery, not about the project's prose.
  //
  // The decisive shape is the adjacency. `absent` + an unusable plugin copy
  // reports (decision-0095); `zero-byte` + the same plugin copy was silent.
  // Both hand the model the same dead token and both have the same remedy; the
  // only thing between them is statSync().isFile(), which is load-bearing for
  // ELIGIBILITY and says nothing about whether the address yields anything.
  //
  // ONE READ, AND THE GUARD IS ON THE DEFECT rather than on the arm -- the same
  // correction the arm above carries. existingFile above and payloadDefect here
  // are two stats and can disagree if the file changes between them, and
  // payloadDefect has FOUR non-empty answers rather than the two an earlier
  // version of this paragraph enumerated. Repaired between the calls gives ""
  // and nothing is reported, which is right because the pointer now resolves.
  // The other two disagreements -- "is missing" and "is not a readable file" --
  // are handled by condition ONE below rather than tolerated here: that earlier
  // version called them "a true sentence about the address the model was
  // handed", which is true of the LEAD and false of the second sentence, and
  // the guard was narrowed for exactly that reason. Nothing can render the
  // empty `(it )` the sibling arm was fixed for.
  const ownCopyDefect = payloadDefect(overlayInvariants);
  // TWO CONDITIONS, both found by review, and each closes a way this report
  // could say something untrue.
  //
  // A NOTE ON THE UNREADABLE PHRASE, since review asked and the answer is "no
  // change": DEFECT_UNREADABLE offers three causes, and one of them -- "a
  // symlink whose target is gone" -- CANNOT apply on this arm, because such a
  // link fails statSync and so fails the existingFile above. The phrase is not
  // narrowed for that, deliberately: these four strings are the vocabulary
  // shared byte for byte with staleness.sh payload_read (decision-0094:2,
  // decision-0028 guard-per-pair), and forking one of them per arm is the exact
  // divergence that pair guard exists to prevent. A slightly over-broad cause
  // list costs a reader one discarded hypothesis; two hosts describing one fault
  // differently costs them the whole diagnosis.
  //
  // ONE: the defect must be one of the two that leave the file PRESENT. This
  // arm was entered because existingFile said a regular file is at that path,
  // and payloadDefect is a SECOND stat that can disagree if the file changes in
  // between. `is missing` or `is not a readable file` coming back here means it
  // did: the file was deleted or replaced by a directory, and the condition that
  // now holds is the arm ABOVE this one, which was not taken. Reporting anyway
  // would ship a second sentence -- "A vendored overlay is authoritative, so the
  // plugin's own copy does not stand in for it WHATEVER STATE THAT COPY IS IN"
  // (the tail is load-bearing and an earlier version of this comment dropped it:
  // it is the clause that makes the copy ineligible by rule rather than merely
  // unavailable, and the whole sentence is what the guard pins) -- that is FALSE in exactly
  // those states, because with no file at the overlay path decision-0093:2's
  // fallback does make the plugin's copy eligible and the reinstall this report
  // withholds would be a real remedy. Silence through a microsecond race is the
  // right answer: the pointer still names the overlay's own address, and the
  // next session classifies it correctly.
  //
  // TWO: the pointer must actually BE in the context about to be delivered. This
  // report asserts, as fact, that "the invariants pointer in the context just
  // injected names a file that yields nothing to read" -- and on this branch the
  // delivery is assembled from the PROJECT's own files, which nothing here
  // validates beyond trellis.md's @rules.md placeholder count and rules.md's
  // sentinel. A vendored overlay with the pointer edited out would otherwise draw
  // a report about a pointer that was never delivered, handing the reader a
  // repair for a file nothing names: the right-diagnosis-wrong-artifact class
  // this whole thread has been closing.
  //
  // BOTH HALVES, because the delivery is both. The context is
  // `trellis.replace("@rules.md", rules)` below, so the pointer ships if EITHER
  // file carries it, and an overlay is free to put it in its rules half. Review
  // found this check written against `trellis` alone: a project that moved the
  // pointer into .trellis/internal/rules.md got silence while the delivered
  // context really did carry a dead pointer -- the same defect the check exists
  // to prevent, missed by half. The disjunction is exact for any rules body
  // holding no `$`, which is every one this payload ships: the rules body is
  // substituted INTO the prose at the placeholder, so the assembled text
  // contains the token exactly when one of the two files does.
  //
  // NOT claimed as exact without that qualifier, which review corrected. The
  // substitution below is `trellis.replace("@rules.md", rules)`, and
  // String.replace interprets `$&`, `` $` `` and `$'` in the REPLACEMENT --
  // which `rules` is. A `$`-pattern adjacent to the token could therefore make
  // the assembled text and this test disagree in either direction. That hazard
  // is pre-existing and untouched here, and it is the same one the split/join
  // comment above this file's repoint records avoiding (cited without a line
  // span: an earlier version said :1127-1132 and the rationale runs one line
  // later at both ends); the assembly was never given
  // the same treatment. Filed rather than fixed in this change, which is about
  // a report rather than about how the payload is joined.
  // TestCodexReportsWhenTheVendoredPointerArrivesThroughTheRulesHalf pins the
  // rules half and TestCodexSaysNothingWhenTheVendoredProseCarriesNoPointer the
  // absence of both, so neither direction can be widened or narrowed unobserved.
  if (
    (ownCopyDefect === DEFECT_EMPTY || ownCopyDefect === DEFECT_UNREADABLE) &&
    (trellis.includes(INVARIANTS_TOKEN) || rules.includes(INVARIANTS_TOKEN))
  ) {
    ownOverlayInvariantsDefect = ownCopyDefect;
  }
}
// The rules payload's own well-formedness (the sentinel gate) must be checked
// BEFORE it is trusted enough to derive a slug set from it — moved ahead of
// that derivation for exactly this reason. Deriving first and validating after
// meant a broken rules.md (no sentinel, a truncated one, or a doubled one)
// yielded an empty or malformed slug set, the PROJECT's .trellis/rules.toml was
// then judged against it, and the report blamed the project's config for a
// defect that was actually in the plugin's own payload.
if (rules.split(SENTINEL).length - 1 !== 1 || !rules.endsWith(`${SENTINEL}\n`)) {
  // sources.rules, for the same reason as sources.prose above: hardcoded, this
  // named the vendored path on the plugin-native branch too, where the file
  // actually read is reference/rules.md.
  fail(sources.rules, "invalid-rules");
  process.exit(0);
}
// Derived from the payload actually resolved above (vendored or plugin-native,
// whichever `sources` picked), not a hardcoded list — this is the slug set a
// payload upgrade CAN repair, and it is what the classifier finds unknown slugs
// and floors against. De-duplicated at the source, so a rules.md that ever
// tagged one slug twice cannot make the computed sentence name a rule twice.
const slugs = [...new Set(slugsFromRules(rules))];
// An EMPTY derived set is the Codex twin of staleness.sh's
// `no-slugs-in-payload` refusal, and it needs its own branch for the same
// reason: nothing downstream can tell it apart from a satisfied one. The
// sentinel gate above proves rules.md is well-formed, not that it TAGS any
// slug — a payload whose rule lines lost their trailing backticked slug keeps
// its sentinel and yields []. Against an empty set every row names an unknown
// slug, nothing is switched off, and this hook would emit a successful "loaded
// installed overlay" response about rules the model cannot identify: a silent
// governance blackout at exit 0, on the host where the blackout is hardest to
// notice because the response looks like success. Rejected here, before
// anything consumes `slugs`.
if (slugs.length === 0) {
  fail(sources.rules, "no-slugs-in-payload");
  process.exit(0);
}

const stamp = version.endsWith("\n") ? version.slice(0, -1) : version;
// A rules file reached through a symbolic link is classified and never shown
// (TRL-97): a repository can commit .trellis/rules.toml, or the directory
// holding it, as a link to any file the user can read, a credentials file
// included, and the echo would put that file into the model's context. Only
// those two components are tested, with lstat, and never the project root, whose
// own path often runs through a link (a temp directory, a home directory). An
// lstat that fails on a path just read is a race and reads as no link, as
// `[ -L ]` does on the other host.
let rulesSymlink = false;
for (const component of [".trellis", PROJECT_CONFIG]) {
  try {
    if (fs.lstatSync(path.join(projectRoot, component)).isSymbolicLink()) rulesSymlink = true;
  } catch {
    // See above: not a link.
  }
}
// One assembly for both branches (KTD7). The prose ends with the header's
// end marker and its newline, so the blank line before the heading is the one
// staleness.sh prints too.
const section = activationSection(rulesToml, slugs, rulesSymlink);
const context =
  `${trellis.replace("@rules.md", rules)}\n${section.text}\n` +
  `Trellis hook loaded installed overlay: ${stamp}\n`;

// The runaway guard, and nothing more. The file is echoed only while it and its
// warnings fit RULES_ECHO_MAX_BYTES, a bound measured so the largest section it
// admits still fits this budget, and the warnings are bounded, so what reaches
// this is a payload that has outgrown the budget on its own. TRL-29's degradation
// retired with the provenance it degraded (TRL-97).
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
// One systemMessage, so the warnings that can fire on a SUCCESSFUL delivery
// collect here rather than each assigning the field — the second writer would
// silently drop the first, and these two are independent conditions that can
// hold at once. Joined with a space: with one warning the string is
// byte-identical to what that warning shipped alone.
const warnings = [];
// TRL-69, decided at the repoint above. The pointer was rewritten to a file
// that cannot be read, which makes this a half-installed plugin payload. Said
// on the channel fail() and the rules.toml warnings already use, rather than left
// for whoever opens the pointer to meet as an unexplained missing read.
//
// "no readable", not "no": this must never turn a permissions fault into a
// claim that the file is gone. #294 wrote it that way because existingFile
// could not tell the two apart and the lead sentence had to survive either;
// TRL-72 gave the reader the rest, so the lead now carries a classification
// behind it rather than covering for the absence of one.
//
// `(it ${defect})`, in staleness.sh's own words and punctuation. The two hosts
// report the same broken install, and an operator who reads both must not have
// to reconcile two vocabularies for one fault -- which remedy applies is the
// whole reason payload_read classifies at all ("missing and unreadable are told
// apart because their remedies differ", staleness.sh:164).
//
// "YIELDS NOTHING TO READ", not "cannot be read", and that sentence had to move
// together with the classification rather than after it. #295 corrected the
// identical wording on the Claude side because it is false for exactly one of
// the four classifications -- an empty file reads fine, it just yields
// nothing -- and it left this host alone on the stated ground that "Claude is
// the only host that fires on `empty`". This change is what makes that premise
// false. Keeping the old wording would have shipped `(it is empty), so ... names
// a file that cannot be read`: a sentence contradicting its own parenthetical,
// and a fresh host divergence opened by the change that exists to close one.
if (pluginInvariantsDefect !== "") {
  warnings.push(
    `Trellis warning: this plugin payload has no readable ${pluginInvariants} (it ${pluginInvariantsDefect}), so the invariants pointer in the context just injected names a file that yields nothing to read. ` +
      "The rules themselves were delivered and govern this session normally; the reference is consulted on demand, so only a rule that turns out ambiguous needs it. " +
      "Reinstalling or updating the Trellis plugin is the likely fix.",
  );
}
// TRL-71, decided at the third arm above. The overlay is what the reader must
// repair, so the overlay is what this sentence blames -- naming the plugin
// install here would name the wrong artifact on a branch where the plugin's
// reference/ is not the delivery source (#294's own scoping reason, applied
// rather than overridden).
//
// The shared vocabulary is kept verbatim on BOTH paths it names: the
// `no readable <abs path>` lead that TestBothHostsReportAMissingInvariantsTarget
// pins on the other branch, the `(it <defect>)` classification in
// staleness.sh's own words and punctuation, and "yields nothing to read" rather
// than "cannot be read" -- false for an empty file, which reads fine (#295).
// An operator who meets one host's report and then the other's must not have to
// reconcile two vocabularies for one fault.
//
// Not a fail(): decision-0093:1 rules that a CONSULTED reference may not fail a
// session closed, and this rides systemMessage, outside the MAX_CONTEXT_BYTES
// bound on `context`. The delivery is untouched -- prose, rules and version
// still come from the overlay.
//
// GUARDED ON THE DEFECT, not on the path, and that is the same correction the
// arm above already carries -- review found it applied to one half only. The
// overlay's eligibility is decided by existingFile and its classification by a
// SECOND stat, so the two can disagree if the file appears between them: the
// defect comes back "" and a path-guarded push renders `has no readable <path>
// (it ), so ...`. Guarding on the defect makes the race resolve the right way
// rather than merely quietly -- if the overlay's own file now exists, the
// pointer this report is about RESOLVES, and there is nothing to report.
if (vendoredInvariantsOverlayDefect !== "") {
  warnings.push(
    `Trellis warning: this project's vendored overlay has no readable ${vendoredInvariantsOverlay} (it ${vendoredInvariantsOverlayDefect}), so the invariants pointer in the context just injected names a file that yields nothing to read. ` +
      `The plugin's own copy could not stand in for it: there is no readable ${pluginInvariants} either (it ${vendoredInvariantsPluginDefect}). ` +
      "The rules themselves were delivered and govern this session normally; the reference is consulted on demand, so only a rule that turns out ambiguous needs it. " +
      "Putting a readable invariants.md at that overlay path, or reinstalling the Trellis plugin so its copy can stand in, is the likely fix.",
  );
}
// TRL-73, decided at the fourth arm above. Same lead, same classification
// vocabulary, same claim as the report directly above -- both blame the
// overlay, and an operator who meets one and then the other must not have to
// reconcile two vocabularies for one class of fault (decision-0028's
// guard-per-pair, applied to a string table that lives in two files in two
// languages).
//
// WHERE IT DELIBERATELY DIFFERS IS THE SECOND SENTENCE AND THE REMEDY, and that
// difference is forced rather than stylistic (decision-0096:2). In the cell
// above, the plugin's copy is an ELIGIBLE substitute that happened to be
// unavailable, so naming it and classifying it tells the reader whether
// reinstalling would even have helped. Here it is ineligible WHATEVER STATE IT
// IS IN, so naming it would be noise and offering the reinstall would send the
// reader to a remedy that cannot work -- there is exactly one repair on this
// arm, and it is the project's own file. That is pinned by
// TestCodexReportsTheOverlaysOwnDeadInvariantsWithoutSubstituting, which runs
// each shape against a healthy plugin copy and a missing one and requires the
// same answer.
//
// Not a fail(): decision-0093:1 rules that a CONSULTED reference may not fail a
// session closed, and this rides systemMessage, outside the MAX_CONTEXT_BYTES
// bound on `context`. Measured across all five overlay shapes, the delivered
// context is 7908 bytes -- the SAME on the healthy row as on the four broken
// ones, because the pointer keeps the 33-byte token rather than an absolute
// path. The report costs the context nothing at all.
if (ownOverlayInvariantsDefect !== "") {
  warnings.push(
    `Trellis warning: this project's vendored overlay has no readable ${overlayInvariants} (it ${ownOverlayInvariantsDefect}), so the invariants pointer in the context just injected names a file that yields nothing to read. ` +
      "A vendored overlay is authoritative, so the plugin's own copy does not stand in for it whatever state that copy is in. " +
      "The rules themselves were delivered and govern this session normally; the reference is consulted on demand, so only a rule that turns out ambiguous needs it. " +
      "Putting a readable invariants.md at that overlay path is the likely fix.",
  );
}
// The rules.toml warnings, mirrored from the context (KTD4). They are already
// inside `context`, where the budget measures them; they are repeated here
// because nothing in this repository establishes whether Codex puts
// systemMessage in front of the model, and the context is what it certainly
// reads. The floor-row warning that used to be assembled here is one of them.
warnings.push(...section.warnings);
if (warnings.length > 0) {
  response.systemMessage = warnings.join(" ");
}
emit(response);

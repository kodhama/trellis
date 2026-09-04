# Codex Degrades on the No-Mismatch Path Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

> **Executed 2026-09-03. Three predictions below were wrong; `decision-0086` follows execution, not this plan.** Kept unedited as the planning record (`decision-0085`), with the corrections stated here rather than silently patched into the steps:
>
> 1. **Task 1, Step 5** predicted that moving writer *and* reader together would leave the pin green. It goes **red** — the file under test is written by the *Claude* host, so a Codex-only template change breaks the strip against a `staleness.sh`-repaired file. A better outcome than predicted: the pin catches cross-host drift too.
> 2. **Task 2, Step 7, row 3** predicted that reconciling the mismatch branch from the raw source would trip *"the two degraded paths disagree"*. **The mutation survived** — the fixture there carries no *persisted* provenance, so that branch was unguarded. `TestCodexDegradesPersistedProvenanceOnTheMismatchPathToo` was added for it, and it showed the branch was a **refusal**, not the cosmetic difference this plan assumed.
> 3. **The runaway guard's new threshold is thirty quarantined rows, not the ~37 derived here.** Measured, same payload and slug family as the baseline: 42 B per row stripped against 192 B unstripped, first refusal at N = 30 where it was N = 9.
>
> Also unlisted here and required in practice: the `VERSION` bump drags both plugin manifests with it (`TestPluginPackageParity` pins them to it), so `install.sh`'s manifest advances for four files, not one.
>
> 4. **Found in review of PR #263 (thread `PRRT_kwDOTIeCVc6eu78z`), not by this plan:** the announcement the degraded path appends was itself unbudgeted. One quarantine note frees 150 B; the full notice costs 414 B and the degraded mandate 156 B more than the full one — so at one or two notes the "degraded" assembly was *larger* than the full one, and above that a notice-sized band of fitting bodies still refused. Fixed by tiering each announcement (full, then a 129 B / 571 B compact form) and taking the first that fits; the residual is the compact line's own length and is left open rather than going silent. Threshold re-measured at **N = 37** (from 30) — the number correction 3 above retired, now true for a different reason: the derivation that ignored the announcement's cost describes a design that no longer pays it. `decision-0086` carries the dated note.

**Goal:** A Codex session whose `.trellis/rules.toml` already carries Trellis's own persisted provenance comments still governs when the assembled context would exceed `MAX_CONTEXT_BYTES` — instead of refusing outright, permanently, on a file Trellis itself told the agent to write.

**Architecture:** The degradation `decision-0084` §6 built is gated on `mismatch !== null`, so it only ever fires in the session that has something to reconcile. The session *after* the repair has no mismatch, nothing to degrade, and hits the hard refusal. This plan moves the trigger from "a reconciliation ran" to "the assembly is over budget", which is what the branch was always about. The mechanism is the one already there — provenance comes off the *injected* copy while the file on disk keeps it — extended to reach provenance the file already carries. Both provenance strings become templates with named placeholders so the writer (`quarantineNote` / `addedHeader`) and a new reader (`stripPersistedProvenance`) are derived from one source and cannot drift.

**Tech Stack:** Node ESM (`plugins/trellis/hooks/codex-context.mjs`), Go tests (`cli/`), POSIX shell + awk (`plugins/trellis/hooks/staleness.sh`, read-only reference).

**Spec:** [TRL-29](https://linear.app/kodhama/issue/TRL-29) — "The Codex hook fails closed on a self-imposed byte cap where Codex itself would degrade gracefully", reopened 2026-09-03 for this half. `decisions/0084-codex-reaches-reconciliation-parity.md` §6 carries the measurement and the corrected claim.

## Global Constraints

- **The hook never writes `.trellis/rules.toml`.** `decision-0070` D4, pinned by `codexReconciledRowsAllowingDegraded` in `cli/codex_hook_test.go`. The hook emits text; the *agent* writes. No task adds a filesystem write to `codex-context.mjs`.
- **Quarantine never deletes.** No row's value is lost on any path, degraded included. Stripping a quarantine *note* leaves the commented-out row line it annotates.
- **The file is the archive; the injection is the working set.** The agent is never told to write an abbreviated file. On the no-mismatch path it is not told to write the file at all.
- **Every exit is `process.exit(0)`.**
- **No deletion verb in any agent-facing hook output string.** `destructiveVerbs` in `cli/plugin_hook_test.go:513` is the list: `delete drop remove overwrite replace reset discard "rm "`.
- **`TestBothHostsReconcileIdentically` must keep passing and must not become vacuous.**
- **Every test run uses `-count=1`.** `cd cli && go test -count=1 ./...`. The hooks are external files Go's test cache does not track; a cached PASS has already produced a false result on this work (`AGENTS.md`, *Checks and review*).
- **Typecheck:** `cd cli && go build ./... && go vet ./...`.
- **Every guard is mutation-proven.** Break it, watch the covering test go red with the expected symptom, restore.
- **`install.sh` carries baked bundle-manifest hashes** — advance them for any hook edited (`cli/install_script_test.go` `TestInstallScriptBundleManifestIsCurrent` regenerates and diffs).
- **A payload change needs a `plugins/trellis/VERSION` bump**, or no cached consumer re-pulls. Collides with the TRL-33 session; resolve to the higher number.
- **`decisions/` is append-only, no `status:` field in new frontmatter** (`decision-0082`). Next free number is **0086** (0085 is taken).

**Reference — the measured baseline**, real firm payload, `rules-a.toml` + N foreign rows, reproduced on `3f44620` before any change:

| N | session 1 | file the mandate produces | session 2 |
|---|---|---|---|
| 5 | degraded, 9010 B | 2097 B | 8831 B delivered |
| 8 | degraded, 9133 B | 2670 B | 9404 B delivered |
| **9** | degraded, 9174 B | **2861 B** | **refused, `context-over-budget`, nothing injected** |
| 12 | degraded, 9299 B | 3434 B | refused |

Each foreign row costs the file 191 B, of which 148 B is the quarantine note. Stripping notes takes the per-row cost to 43 B and moves the cliff from N ≥ 9 to roughly N ≈ 37.

---

### Task 1: Provenance strings become templates with a matching reader

**Files:**
- Modify: `plugins/trellis/hooks/codex-context.mjs` — `quarantineNote` / `addedHeader` (`:387-395`), inserting `stripPersistedProvenance` after them
- Test: `cli/codex_hook_test.go`

**Interfaces:**
- Produces: **`stripPersistedProvenance(source: string) => string`** — returns `source` with (a) every line that is entirely an `addedHeader` removed, and (b) the `quarantineNote` suffix taken off any line that ends with one. Everything else, including the commented-out row a note annotates, is returned byte-for-byte. Line endings are preserved as found (the function operates on whole lines split by `\n`, so a CRLF file keeps its `\r`).
- Produces: **`QUARANTINE_NOTE_TEMPLATE`**, **`ADDED_HEADER_TEMPLATE`** — literal strings carrying `{date}`, `{stamp}`, `{count}` placeholders. `quarantineNote(today, stamp)` and `addedHeader(count, today, stamp)` keep their existing signatures and return values; they now fill a template instead of concatenating.

This task is behaviour-neutral by design: nothing calls `stripPersistedProvenance` yet. Its test is the anti-drift pin — that the reader matches exactly what the writer produces.

- [ ] **Step 1: Write the failing test**

Append to `cli/codex_hook_test.go`:

```go
// TestCodexProvenanceStripperMatchesItsOwnWriter is the anti-drift pin for
// TRL-29's no-mismatch degradation. stripPersistedProvenance has to recognise
// the provenance an EARLIER session wrote — possibly by the other host, on an
// older date, against an older payload stamp — and the only thing that keeps a
// reader honest against a writer it never sees run is a test that runs both.
//
// Session 1 (a genuine mismatch, under budget) writes full provenance into the
// context; the agent's file is that text. Feeding that file back through a
// FORCED degradation must yield exactly what session 1 would have injected had
// IT been over budget: the same rows with no notes at all.
func TestCodexProvenanceStripperMatchesItsOwnWriter(t *testing.T) {
	base := payloadFiles()["rules-a.toml"]
	// One foreign row (quarantines) and one removed row (adds) — both provenance
	// forms in one fixture, which a single-kind fixture would not cover.
	fixture := strings.Replace(base,
		"inv-minimal-first         = { active = true }\n", "", 1)
	if fixture == base {
		t.Fatal("premise: fixture removed nothing — the case would prove nothing")
	}
	fixture += "inv-foreign-rule-a = { active = true }\n"

	full := claudeReconciledRows(t, fixture) + "\n"
	if !strings.Contains(full, "# quarantined ") || !strings.Contains(full, "# added 1 row(s) below on ") {
		t.Fatalf("premise: the reconciled file must carry BOTH provenance forms:\n%s", full)
	}

	stripped := codexStripProvenance(t, full)
	for _, banned := range []string{"# quarantined ", "# added "} {
		if strings.Contains(stripped, banned) {
			t.Errorf("stripPersistedProvenance left %q behind — the reader has drifted from the writer:\n%s", banned, stripped)
		}
	}
	// Quarantine never deletes: the commented row keeps its line and its value.
	if !strings.Contains(stripped, "# inv-foreign-rule-a = { active = true }") {
		t.Errorf("the quarantined row lost its line — quarantine never deletes:\n%s", stripped)
	}
	if !strings.Contains(stripped, "inv-minimal-first         = { active = true }") &&
		!strings.Contains(stripped, "inv-minimal-first = { active = true }") {
		t.Errorf("the added row did not survive the strip:\n%s", stripped)
	}
}

// codexStripProvenance calls stripPersistedProvenance directly, by importing
// codex-context.mjs's own source into a throwaway ES module. The hook is a
// top-level script (it reads stdin and exits), so it cannot be imported as-is;
// the helper slices out the function under test by name and evaluates just
// that. Direct rather than end-to-end BECAUSE the end-to-end path needs a
// fixture over 9500 B, which cannot isolate a stripper defect from a budget
// arithmetic defect — Task 3 covers the end-to-end path separately.
func codexStripProvenance(t *testing.T, source string) string {
	t.Helper()
	body, err := os.ReadFile("../plugins/trellis/hooks/codex-context.mjs")
	if err != nil {
		t.Fatal(err)
	}
	const from = "const QUARANTINE_NOTE_TEMPLATE"
	const to = "\n// reconcileRows mirrors"
	start := strings.Index(string(body), from)
	end := strings.Index(string(body), to)
	if start < 0 || end < 0 || end <= start {
		t.Fatal("could not slice the provenance-template region out of codex-context.mjs — the extraction is broken, and a helper that reads nothing proves nothing")
	}
	region := string(body)[start:end]
	if !strings.Contains(region, "function stripPersistedProvenance(") {
		t.Fatal("the sliced region does not contain stripPersistedProvenance — the extraction is broken")
	}
	dir := t.TempDir()
	mod := filepath.Join(dir, "strip.mjs")
	script := region + "\nprocess.stdout.write(stripPersistedProvenance(await new Promise((r) => {" +
		"let s = \"\"; process.stdin.setEncoding(\"utf8\");" +
		"process.stdin.on(\"data\", (c) => { s += c; }); process.stdin.on(\"end\", () => r(s)); })));\n"
	if err := os.WriteFile(mod, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", mod)
	cmd.Stdin = strings.NewReader(source)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("stripPersistedProvenance harness failed: %v\nstderr: %s", err, stderr.String())
	}
	return stdout.String()
}
```

- [ ] **Step 2: Run it and verify it fails**

Run: `cd cli && go test -count=1 -run TestCodexProvenanceStripperMatchesItsOwnWriter .`
Expected: FAIL — "could not slice the provenance-template region out of codex-context.mjs".

- [ ] **Step 3: Implement the templates and the stripper**

Replace `quarantineNote` / `addedHeader` in `plugins/trellis/hooks/codex-context.mjs` with:

```js
// The two provenance strings, as ONE source of truth for the writer below and
// the reader (stripPersistedProvenance) under it. A regex written out by hand
// next to a string built by hand is two statements of the same text that drift
// silently: the writer changes, the reader keeps matching yesterday's wording,
// and the degradation quietly stops degrading — which is the failure class
// TRL-29 is, arrived at from the other side. Deriving both from one template
// makes that drift impossible rather than merely tested for.
//
// staleness.sh:862 and :933 carry the identical text (both hosts write into one
// .trellis/rules.toml, so a file repaired on Claude must strip on Codex).
// TestBothHostsReconcileIdentically is what keeps those two in step; these
// templates keep the reader in step with THIS side's writer.
const QUARANTINE_NOTE_TEMPLATE =
  "  # quarantined {date}: not in {stamp}. If a newer Trellis" +
  " release ships this slug, update the Trellis plugin and uncomment this row.";
const ADDED_HEADER_TEMPLATE =
  "# added {count} row(s) below on {date} (missing from {stamp})";

const fillTemplate = (template, values) =>
  template.replace(/\{(\w+)\}/gu, (_match, key) => values[key]);

// A template's literal segments, escaped, joined by a same-line wildcard: the
// pattern matches what the template WROTE on any date, against any stamp, for
// any count. `[^\n]*` rather than `.*` so a pattern can never span a line even
// if a caller ever hands this a whole file.
const templatePattern = (template, wrap) =>
  new RegExp(
    wrap(
      template
        .split(/\{\w+\}/u)
        .map((literal) => literal.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&"))
        .join("[^\\n]*"),
    ),
    "u",
  );

// quarantineNote/addedHeader are the two provenance strings reconcileRows
// glues onto a row (or a block of rows) it could not match to the payload's
// slug set. Filled from the templates above so the `withProvenance` branches
// below (the ordinary call and TRL-29's degraded one, immediately after) can
// never drift from each other about what "with provenance" means.
function quarantineNote(today, stamp) {
  return fillTemplate(QUARANTINE_NOTE_TEMPLATE, { date: today, stamp });
}
function addedHeader(count, today, stamp) {
  return fillTemplate(ADDED_HEADER_TEMPLATE, { count, date: today, stamp });
}

// A quarantine note is a SUFFIX on a commented-out row; an added-rows header is
// a whole line of its own. Anchored accordingly.
const QUARANTINE_NOTE_PATTERN = templatePattern(
  QUARANTINE_NOTE_TEMPLATE,
  (body) => `(?:${body})[ \\t]*\\r?$`,
);
const ADDED_HEADER_PATTERN = templatePattern(
  ADDED_HEADER_TEMPLATE,
  (body) => `^[ \\t]*(?:${body})[ \\t]*\\r?$`,
);

// TRL-29: the degradation decision-0084 §6 built drops provenance the hook is
// about to WRITE. This drops provenance an earlier session already wrote into
// the file — the other half, and the one that made the degradation one-shot.
// A session that follows a repair has no mismatch, so nothing new is generated
// and there was nothing to leave off; the persisted comments were the only
// thing left to give up, and the gate above them never opened.
//
// Only Trellis's own two provenance forms come off. A comment the PROJECT
// wrote is the project's content and stays: this is a byte-budget concession
// on Trellis's own bookkeeping, not a licence to abbreviate a consumer's file.
//
// Never touches the file on disk (decision-0070 D4) and never touches a row's
// VALUE: a quarantined row keeps its commented-out line verbatim and loses only
// the note appended to it, which is exactly the shape reconcileRows produces
// with `withProvenance = false`. Quarantine still never deletes.
function stripPersistedProvenance(source) {
  const kept = [];
  for (const line of source.split("\n")) {
    if (ADDED_HEADER_PATTERN.test(line)) continue;
    kept.push(line.replace(QUARANTINE_NOTE_PATTERN, ""));
  }
  return kept.join("\n");
}
```

Note the `\r?$` in both anchors: a CRLF file splits on `"\n"` with the `\r` still on the line, and a note must still be recognised there. The `\r` itself is preserved by `replace` on the quarantine branch (it sits outside the matched group only if the pattern excludes it) — the pattern deliberately *consumes* the trailing `\r` on a stripped note so the resulting line matches what `reconcileRows` writes for the same input; the added-header branch drops the whole line including its `\r`.

- [ ] **Step 4: Run the test and verify it passes**

Run: `cd cli && go test -count=1 -run TestCodexProvenanceStripperMatchesItsOwnWriter .`
Expected: PASS.

- [ ] **Step 5: Mutation-prove the pin**

In `codex-context.mjs`, change `QUARANTINE_NOTE_TEMPLATE`'s wording (e.g. `# quarantined` → `# shelved`).
Run: `cd cli && go test -count=1 -run TestCodexProvenanceStripperMatchesItsOwnWriter .`
Expected: still PASS — the writer and reader moved together, which is the point. Then instead mutate only the reader: change `ADDED_HEADER_PATTERN`'s `wrap` to `^$` so it matches nothing.
Expected: FAIL with `stripPersistedProvenance left "# added " behind`. Restore.

- [ ] **Step 6: Run the whole suite and commit**

Run: `cd cli && go build ./... && go vet ./... && go test -count=1 ./...`
Expected: PASS (nothing calls the stripper yet; `TestBothHostsReconcileIdentically` unaffected).

```bash
git add plugins/trellis/hooks/codex-context.mjs cli/codex_hook_test.go
git commit -m "TRL-29: derive the provenance reader from the writer's own templates"
```

---

### Task 2: The over-budget branch degrades on both paths

**Files:**
- Modify: `plugins/trellis/hooks/codex-context.mjs` — the over-budget branch (`:870-912`), and a new `provenanceOmittedNotice` beside `repairMandate`
- Modify: `cli/plugin_hook_test.go` — `codexPayloadAssembly` (`:617-636`)
- Modify: `cli/codex_hook_test.go` — `codexReconciledRowsFromContext` (`:1498`)

**Interfaces:**
- Consumes: `stripPersistedProvenance` from Task 1.
- Produces: **`provenanceOmittedNotice() => string`** — the no-mismatch degraded announcement. Contains the shared marker sentence `Provenance was omitted above to fit the context budget`, and contains **no** write instruction.
- Produces: the heading `## Provenance comments were left out of the rows above`, which `codexReconciledRowsFromContext` must stop at.

- [ ] **Step 1: Write the failing test**

Append to `cli/codex_hook_test.go`:

```go
// TestCodexDegradesOnASecondSessionOverBudget is TRL-29's remaining half, as
// the sequence that produced it rather than as a unit assertion.
//
//  1. Session 1 sees a mismatched file, degrades, delivers, and mandates a
//     FULL-PROVENANCE write.
//  2. The agent complies. That file is what session 2 reads.
//  3. Session 2 has NO mismatch. Before this fix the degradation was gated on
//     `mismatch !== null`, so nothing was offered up and the hard refusal fired
//     — permanently, because nothing about that file changes again, while
//     staleness.sh governed happily from the identical bytes.
//
// The nine-foreign-row fixture is the measured cliff (decision-0084 §6 and its
// reproduction on 3f44620): at N = 8 session 2 delivers 9404 B unaided, at
// N = 9 it refused. Below the cliff this test would pass without the fix.
func TestCodexDegradesOnASecondSessionOverBudget(t *testing.T) {
	base := payloadFiles()["rules-a.toml"]
	letters := "abcdefghi"
	fixture := base
	for i := 0; i < 9; i++ {
		fixture += fmt.Sprintf("inv-foreign-rule-%c = { active = true }\n", letters[i])
	}

	// Session 1, and the file its mandate produces. claudeReconciledRows is the
	// full-provenance text both hosts agree on byte for byte
	// (TestBothHostsReconcileIdentically), which is what an agent following the
	// mandate writes — taking it from the Claude host deliberately, so this test
	// cannot be satisfied by the Codex hook agreeing with itself.
	session1Rows, session1Degraded := codexReconciledRowsAllowingDegraded(t, fixture)
	if !session1Degraded {
		t.Fatal("premise: session 1 must be over budget and degrade, or this fixture is not the reopened case")
	}
	repaired := claudeReconciledRows(t, fixture) + "\n"
	if !strings.Contains(repaired, "# quarantined ") {
		t.Fatalf("premise: the file the mandate produces must carry persisted provenance:\n%s", repaired)
	}

	// Session 2.
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	tomlPath := filepath.Join(project, ".trellis", "rules.toml")
	if err := os.WriteFile(tomlPath, []byte(repaired), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))

	if strings.Contains(raw, "context-over-budget") {
		t.Fatalf("session 2 refused on a file Trellis itself told the agent to write — that is TRL-29's remaining half:\n%s", raw)
	}
	if got.HookSpecificOutput == nil {
		t.Fatalf("session 2 injected nothing:\n%s", raw)
	}
	ctx := got.HookSpecificOutput.AdditionalContext
	for _, slug := range assessableSlugs {
		if !regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(slug) + `[ \t]*=`).MatchString(ctx) {
			t.Errorf("rule %s must still be delivered on the degraded no-mismatch path", slug)
		}
	}
	if n := len([]byte(ctx)); n > 9500 {
		t.Errorf("degraded session-2 context is %d bytes, still over the cap", n)
	}
	if !strings.Contains(ctx, codexDegradedMarker) {
		t.Errorf("the omission must be announced, not silent:\n%s", ctx)
	}

	// The file is the archive, the injection is the working set. Two halves:
	// the file keeps every note...
	after, err := os.ReadFile(tomlPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != repaired {
		t.Errorf("the hook wrote .trellis/rules.toml — decision-0070 D4 says it never does:\nbefore:\n%s\nafter:\n%s", repaired, after)
	}
	// ...and the agent is never told to write the abbreviated copy back. On this
	// path there is nothing to repair, so there must be no write instruction at
	// all — the mismatch path's own version of this property is pinned by the
	// "not the abbreviated ones shown above" clause asserted below.
	if strings.Contains(ctx, "Write .trellis/rules.toml") {
		t.Errorf("no repair ran, so nothing may instruct a write — an abbreviated file is the one outcome this whole branch exists to prevent:\n%s", ctx)
	}

	// The strongest statement of "same mechanism, different trigger": whatever
	// session 1 injected when IT degraded is what session 2 injects, byte for
	// byte. If these ever diverge, one of the two paths grew provenance the
	// other drops, and the archive/working-set split has a seam in it.
	session2Rows := codexReconciledRowsFromContext(t, ctx)
	if session1Rows != session2Rows {
		t.Errorf("the two degraded paths disagree about the same file\nsession 1:\n%s\nsession 2:\n%s", session1Rows, session2Rows)
	}
}

// TestCodexKeepsProvenanceWhenItFits is the over-refusal guard for Task 2: the
// strip is a budget concession, not a policy. A file carrying persisted
// provenance that assembles UNDER the cap must be injected verbatim, notes and
// all — degrading a session that had the bytes to spare would throw away the
// provenance for nothing and make every session's context depend on a
// threshold no reader can see.
func TestCodexKeepsProvenanceWhenItFits(t *testing.T) {
	base := payloadFiles()["rules-a.toml"]
	repaired := claudeReconciledRows(t, base+"inv-foreign-rule-a = { active = true }\n") + "\n"
	if !strings.Contains(repaired, "# quarantined ") {
		t.Fatalf("premise: the fixture must carry persisted provenance:\n%s", repaired)
	}

	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	if err := os.WriteFile(filepath.Join(project, ".trellis", "rules.toml"), []byte(repaired), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("an already-repaired file must govern:\n%s", raw)
	}
	ctx := got.HookSpecificOutput.AdditionalContext
	if n := len([]byte(ctx)); n > 9500 {
		t.Fatalf("premise: this fixture must fit, or it proves nothing about the fitting case (%d bytes)", n)
	}
	if !strings.Contains(ctx, "# quarantined ") {
		t.Errorf("provenance was dropped from a session that had the budget for it:\n%s", ctx)
	}
	if strings.Contains(ctx, codexDegradedMarker) {
		t.Errorf("a session under budget announced a degradation it did not perform:\n%s", ctx)
	}
}
```

- [ ] **Step 2: Run them and verify they fail**

Run: `cd cli && go test -count=1 -run 'TestCodexDegradesOnASecondSessionOverBudget|TestCodexKeepsProvenanceWhenItFits' .`
Expected: `TestCodexDegradesOnASecondSessionOverBudget` FAILS with "session 2 refused on a file Trellis itself told the agent to write". `TestCodexKeepsProvenanceWhenItFits` PASSES already (it pins behaviour the fix must not break).

- [ ] **Step 3: Add the no-mismatch notice**

Insert into `plugins/trellis/hooks/codex-context.mjs`, immediately after `repairMandate`:

```js
// provenanceOmittedNotice is repairMandate's counterpart on the path where
// there is nothing to repair (TRL-29): the file already matches the payload's
// slug set, and what does not fit is the provenance the file itself carries
// from an earlier repair.
//
// It shares repairMandate's degraded marker sentence VERBATIM — one sentence,
// two callers — because "the injected copy was abbreviated" is one fact
// however the session arrived at it, and cli/codex_hook_test.go's
// codexDegradedMarker matches on exactly that sentence to tell a degraded
// response from a full one. Two wordings would make that helper silently
// half-blind.
//
// It carries NO write instruction, and must not grow one. On the mismatch path
// the mandate has to say "write the full-provenance version, not the
// abbreviated rows above" because it is asking for a write at all; here nothing
// asked for one, the file on disk is already correct, and the only thing an
// instruction could achieve is the exact failure this branch exists to prevent
// — an agent helpfully rewriting .trellis/rules.toml from an abbreviated copy
// and permanently losing the provenance. TestCodexDegradesOnASecondSessionOverBudget
// asserts the string "Write .trellis/rules.toml" never appears on this path.
function provenanceOmittedNotice() {
  return (
    "\n## Provenance comments were left out of the rows above\n\n" +
    "Provenance was omitted above to fit the context budget and remains in full in .trellis/rules.toml, which matches the rules this payload ships and needs no repair this session. " +
    "The rows above are what governs; the file on disk is the archive of why each row reads the way it does. " +
    "Read the file itself if you need a quarantined row's reason or date.\n"
  );
}
```

- [ ] **Step 4: Rewrite the over-budget branch**

Replace `codex-context.mjs:870-912` (from `if (Buffer.byteLength(context, "utf8") > MAX_CONTEXT_BYTES) {` through its closing brace) with:

```js
if (Buffer.byteLength(context, "utf8") > MAX_CONTEXT_BYTES) {
  // TRL-29: refusing outright here used to be a self-inflicted blackout —
  // Codex's own documented behaviour on oversized hook output is to spill,
  // not reject (MAX_CONTEXT_BYTES' own comment, above), so failing closed was
  // strictly worse than the host's own degradation.
  //
  // What degrades is the INJECTED COPY, never the file: the file is the
  // archive, the injection is the working set. Both are dropped here —
  //
  //   * provenance this session would have GENERATED (reconcileRows'
  //     `withProvenance = false`), and
  //   * provenance the file ALREADY CARRIES from an earlier repair
  //     (stripPersistedProvenance).
  //
  // The second used to be unreachable, and that is the bug TRL-29 was
  // reopened for. This branch was gated on `mismatch !== null`, so it only
  // ran in a session that had something to reconcile; the session AFTER a
  // repair has no mismatch, generates no provenance, and had its persisted
  // provenance withheld from the strip — so the refusal below fired instead,
  // permanently, on a file Trellis itself told the agent to write. Measured
  // (decision-0086, reproduced on 3f44620): at nine foreign rows session 1
  // degraded and delivered 9174 B, the file its mandate produced was 2861 B,
  // and session 2 refused while staleness.sh governed from the identical
  // bytes. The trigger is now the budget, which is what the branch was
  // always about; the gate below only chooses which ANNOUNCEMENT fits.
  const stripped = stripPersistedProvenance(rulesToml);
  if (mismatch !== null) {
    // Stripped source, not the raw one: a file can carry BOTH persisted
    // provenance and a fresh mismatch, and leaving the persisted half on
    // would degrade that session strictly less than the same file with no
    // mismatch at all. reconcileRows classifies from uncommented `(inv|
    // floor)-... =` rows only, so removing comment text changes no
    // quarantine or addition decision and no count — `added`/`quarantined`
    // are identical either way, and so is what governs.
    const bare = reconcileRows(stripped, slugs, stamp, today, false);
    const repairSummary = `added ${bare.added} row(s); quarantined ${bare.quarantined} row(s)`;
    repairMandateText = repairMandate(mismatchCounts(mismatch), repairSummary, true);
    context = buildContext(bare.text, repairMandateText);
  } else {
    // Nothing to reconcile, so no mandate and no write instruction — see
    // provenanceOmittedNotice. effectiveRulesToml was the file verbatim on
    // this path; it is now the file minus Trellis's own bookkeeping.
    context = buildContext(stripped, provenanceOmittedNotice());
  }
  if (Buffer.byteLength(context, "utf8") > MAX_CONTEXT_BYTES) {
    // The runaway guard, and nothing more. Reached only when a context with
    // NO Trellis provenance in it at all — neither generated nor persisted —
    // is still over the cap, which takes roughly thirty-seven quarantined
    // rows on the real firm payload (each costs 43 B once its note is off,
    // against 191 B with it). It is deliberately NOT described as a state
    // with "nothing left to degrade": an earlier draft of this comment and
    // of decision-0084 said exactly that, it was measurably false, and the
    // limitation it concealed is what TRL-29 was reopened for. What is left
    // to degrade here is the consumer's own content — their comments, their
    // row set — and abbreviating that is not a call this hook may make on
    // its own. So it stops, loudly, at exit 0.
    fail("assembled-context", "context-over-budget");
    process.exit(0);
  }
}
```

- [ ] **Step 5: Teach the two Go extractors about the new section**

In `cli/codex_hook_test.go`, `codexReconciledRowsFromContext` (`:1498`), extend the stop alternation so the notice is not swallowed into the "row block":

```go
	m := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(invariantsTrigger) +
		`\n\n(.*?)(?:\n\n## Rule activation was reconciled this session` +
		`|\n\n## Provenance comments were left out of the rows above` +
		`|\nTrellis hook loaded installed overlay: )`).
		FindStringSubmatch(context)
```

In `cli/plugin_hook_test.go`, `codexPayloadAssembly` (`:623`), scan **both** agent-facing text functions rather than only `repairMandate` — a new channel into the agent's context that the deletion and destructive guards cannot see is exactly the hole those guards exist to close:

```go
func codexPayloadAssembly(t *testing.T, body string) string {
	t.Helper()
	var blocks []string
	// Every function in codex-context.mjs that concatenates literal text into
	// the agent's context. repairMandate was the first (TRL-30 task 3);
	// provenanceOmittedNotice is TRL-29's no-mismatch counterpart, and it is
	// scanned here for the same reason — the "no deletion verb reaches the
	// agent" argument only holds if every channel is enforced, not just the
	// one that existed when the guard was written.
	for _, name := range []string{"repairMandate", "provenanceOmittedNotice"} {
		marker := "\nfunction " + name + "("
		start := strings.Index(body, marker)
		if start < 0 {
			t.Fatalf("%s(...) not found in codex-context.mjs — the scan is broken", name)
		}
		rest := body[start:]
		end := strings.Index(rest, "\n}\n")
		if end < 0 {
			t.Fatalf("%s(...) has no closing '}' — the scan is broken", name)
		}
		blocks = append(blocks, rest[:end])
	}
	return strings.Join(blocks, "\n")
}
```

- [ ] **Step 6: Run the new tests and verify they pass**

Run: `cd cli && go test -count=1 -run 'TestCodexDegradesOnASecondSessionOverBudget|TestCodexKeepsProvenanceWhenItFits|TestCodexDegradesRatherThanRefusingOverBudget|TestBothHostsReconcileIdentically|TestCROnlyLineEndingsAreTheOneKnownDivergence|TestEveryDeletionInstructionIsGated|TestEveryDestructiveInstructionIsGated' .`
Expected: PASS.

- [ ] **Step 7: Mutation-prove each guard**

Run each mutation, confirm the named test goes red with the named symptom, then restore:

| Mutation in `codex-context.mjs` | Test that must go red | Symptom |
|---|---|---|
| Restore the `if (mismatch !== null)` gate around the whole degraded reassembly | `TestCodexDegradesOnASecondSessionOverBudget` | "session 2 refused on a file Trellis itself told the agent to write" |
| `const stripped = rulesToml;` (strip disabled) | `TestCodexDegradesOnASecondSessionOverBudget` | same refusal |
| `reconcileRows(rulesToml, …)` on the mismatch branch (persisted half left on) | `TestCodexDegradesOnASecondSessionOverBudget` | "the two degraded paths disagree about the same file" |
| Add `"Write .trellis/rules.toml with the rows shown above. "` to `provenanceOmittedNotice` | `TestCodexDegradesOnASecondSessionOverBudget` | "no repair ran, so nothing may instruct a write" |
| Change `provenanceOmittedNotice`'s marker sentence wording | `TestCodexDegradesOnASecondSessionOverBudget` | "the omission must be announced, not silent" |
| Run the strip unconditionally, before the budget check | `TestCodexKeepsProvenanceWhenItFits` | "provenance was dropped from a session that had the budget for it" |
| Add `"Delete the quarantined rows. "` to `provenanceOmittedNotice` | `TestEveryDeletionInstructionIsGated` | "the codex repair mandate text must never instruct a deletion" |

The last row is the proof that Step 5's `codexPayloadAssembly` change actually wired the new channel in — without it the mutation lands undetected.

- [ ] **Step 8: Full suite, then commit**

Run: `cd cli && go build ./... && go vet ./... && go test -count=1 ./...`
Expected: PASS.

```bash
git add plugins/trellis/hooks/codex-context.mjs cli/codex_hook_test.go cli/plugin_hook_test.go
git commit -m "TRL-29: degrade the injected copy whenever the assembly is over budget"
```

---

### Task 3: Strengthen the mismatch-path pin and sweep the legitimate shapes

**Files:**
- Modify: `cli/codex_hook_test.go` — `TestCodexDegradesRatherThanRefusingOverBudget` (`:1083`)
- Test: `cli/codex_hook_test.go` (new sweep)

**Interfaces:**
- Consumes: everything from Tasks 1 and 2. Adds no product code.

The existing mismatch-path degradation is pinned only by `strings.Contains(ctx, "provenance")`, which the *whole* degraded mandate satisfies. `decision-0084` claims the abbreviated-file property is "mutation-tested directly"; the assertion that carries that claim should name the clause, not a substring that survives most rewordings.

- [ ] **Step 1: Strengthen the existing assertion**

In `TestCodexDegradesRatherThanRefusingOverBudget`, replace:

```go
	if !strings.Contains(ctx, "provenance") {
		t.Errorf("the omission must be announced, not silent:\n%s", ctx)
	}
```

with:

```go
	if !strings.Contains(ctx, codexDegradedMarker) {
		t.Errorf("the omission must be announced, not silent:\n%s", ctx)
	}
	// The property decision-0084 §6 says is the whole point of the branch: the
	// context is abbreviated, the FILE the mandate asks for is not. A bare
	// Contains(ctx, "provenance") passed on any wording — including the
	// non-degraded mandate, if the flag ever regressed — so the clause is named
	// here instead.
	if !strings.Contains(ctx, "not the abbreviated ones shown above") {
		t.Errorf("the degraded mandate must tell the agent to write the FULL-provenance version, not the rows it can see:\n%s", ctx)
	}
	if strings.Contains(ctx, "Write .trellis/rules.toml with exactly the rows shown above") {
		t.Errorf("the degraded mandate fell back to the full-provenance wording — the file would silently lose its provenance:\n%s", ctx)
	}
```

- [ ] **Step 2: Write the over-refusal sweep**

Append to `cli/codex_hook_test.go`:

```go
// TestCodexDoesNotOverRefuseTheLegitimateShapes runs every rules.toml shape a
// real project can present through the real hook and requires that none of
// them reaches a refusal. Two guards on an earlier TRL-29 branch over-corrected
// into refusing healthy payloads, which is exactly as bad for a consumer as the
// silent case this issue is about — a broad, boring sweep is what catches that
// class, and it is cheap.
func TestCodexDoesNotOverRefuseTheLegitimateShapes(t *testing.T) {
	files := payloadFiles()
	firm := files["rules-a.toml"]
	adaptive := files["rules-b.toml"]

	renamed := strings.Replace(firm,
		"inv-minimal-first         = { active = true }",
		"inv-renamed-first         = { active = true }", 1)
	if renamed == firm {
		t.Fatal("premise: rename fixture changed nothing")
	}
	missingRow := strings.Replace(firm, "inv-minimal-first         = { active = true }\n", "", 1)
	if missingRow == firm {
		t.Fatal("premise: missing-row fixture changed nothing")
	}
	alreadyQuarantined := claudeReconciledRows(t, firm+"inv-foreign-rule-a = { active = true }\n") + "\n"
	if !strings.Contains(alreadyQuarantined, "# quarantined ") {
		t.Fatal("premise: already-quarantined fixture carries no provenance")
	}

	cases := map[string]string{
		"unchanged firm preset":     firm,
		"unchanged adaptive preset": adaptive,
		"missing row":               missingRow,
		"unknown row":               firm + "inv-foreign-rule-a = { active = true }\n",
		"duplicate row":             firm + "inv-minimal-first         = { active = false }\n",
		"renamed row":               renamed,
		"already quarantined":       alreadyQuarantined,
		"hand-written partial":      "strictness  = \"firm\"\n",
		"governed = false":          "governed = false\n",
	}
	for name, toml := range cases {
		t.Run(name, func(t *testing.T) {
			project := newGitProject(t)
			writeValidCodexOverlay(t, project)
			if err := os.WriteFile(filepath.Join(project, ".trellis", "rules.toml"), []byte(toml), 0o644); err != nil {
				t.Fatal(err)
			}
			raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
			if strings.Contains(raw, "context-over-budget") || strings.Contains(raw, "invalid-rules") {
				t.Fatalf("a legitimate shape was refused:\n%s", raw)
			}
			if name == "governed = false" {
				// decision-0070 D5: an opt-out is silence, not a refusal, and not
				// a delivery either. Asserted rather than skipped, because
				// "emitted nothing" and "refused" are the two outcomes this
				// sweep must tell apart.
				if raw != "" {
					t.Fatalf("an opt-out must emit nothing at all:\n%s", raw)
				}
				return
			}
			if got.HookSpecificOutput == nil {
				t.Fatalf("no context injected for a legitimate shape:\n%s", raw)
			}
			if n := len([]byte(got.HookSpecificOutput.AdditionalContext)); n > 9500 {
				t.Errorf("%s assembled to %d bytes, over the cap", name, n)
			}
		})
	}

	t.Run("no project rules.toml at all (shipped defaults)", func(t *testing.T) {
		project := newGitProject(t)
		writeValidCodexOverlay(t, project)
		if err := os.Remove(filepath.Join(project, ".trellis", "rules.toml")); err != nil {
			t.Fatal(err)
		}
		raw, _ := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
		if strings.Contains(raw, "context-over-budget") {
			t.Fatalf("a project with no rules.toml must not be refused for budget:\n%s", raw)
		}
	})
}
```

- [ ] **Step 3: Run the sweep**

Run: `cd cli && go test -count=1 -run 'TestCodexDoesNotOverRefuseTheLegitimateShapes|TestCodexDegradesRatherThanRefusingOverBudget' -v .`
Expected: PASS on every subtest. **If any subtest fails, that is the finding** — record the actual behaviour before adjusting the expectation, and never adjust an expectation to make a refusal acceptable.

- [ ] **Step 4: Mutation-prove the strengthened pin**

In `codex-context.mjs`'s over-budget mismatch branch, pass `false` instead of `true` as `repairMandate`'s `degraded` argument.
Run: `cd cli && go test -count=1 -run TestCodexDegradesRatherThanRefusingOverBudget .`
Expected: FAIL with "the degraded mandate must tell the agent to write the FULL-provenance version". Restore.

- [ ] **Step 5: Full suite, then commit**

Run: `cd cli && go build ./... && go vet ./... && go test -count=1 ./...`

```bash
git add cli/codex_hook_test.go
git commit -m "TRL-29: name the abbreviated-file clause, and sweep the shapes that must not refuse"
```

---

### Task 4: Record the decision, bump the payload, advance the manifest

**Files:**
- Create: `decisions/0086-<slug>.md`
- Modify: `decisions/0084-codex-reaches-reconciliation-parity.md` — add `superseded_in_part_by: [decision-0086]`
- Modify: `plugins/trellis/VERSION`
- Modify: `install.sh` — the `hooks/codex-context.mjs` manifest line

**Interfaces:** none — documentation and release bookkeeping.

- [ ] **Step 1: Write the decision record**

`decisions/0086-*.md`, following `decision-0084`'s frontmatter shape **minus any `status:` field** (`decision-0082`). It must state: what changed (the trigger moved from "a reconciliation ran" to "the assembly is over budget"); the measurement, before and after; that the archive/working-set split is the invariant and how it is pinned; that the hard refusal survives as a runaway guard and **what reaching it now means**, with no reprise of the corrected "pathologically large / nothing left to degrade" claim; and that this closes TRL-29's remaining half.

- [ ] **Step 2: Add the forward pointer to `decision-0084`**

`decisions/` is append-only, so the pointer is the only edit: add `superseded_in_part_by: [decision-0086]` to its frontmatter, with a trailing comment naming exactly the part superseded — the §6 sentences describing the degradation as one-shot and gated on `mismatch !== null`. The rest of §6, and all of the record's parity content, stands.

- [ ] **Step 3: Bump the payload version**

`plugins/trellis/VERSION` — increment the minor or patch component past whatever is on `origin/main` at rebase time. The TRL-33 session bumps the same single line; on conflict, resolve to the higher number.

- [ ] **Step 4: Advance the bundle manifest**

Run: `cd cli && go test -count=1 -run TestInstallScriptBundleManifestIsCurrent .`
Expected: FAIL, naming the stale `hooks/codex-context.mjs` hash. Recompute and paste it into `install.sh`'s `TRELLIS_BUNDLE_MANIFEST`, then re-run.
Expected: PASS.

- [ ] **Step 5: Run the corpus reviewer**

Invoke the repo-owned `corpus-reviewer` agent (`.claude/agents/`) over `decisions/0086-*.md` and the edited `0084`. It is read-only; fix what it reports.

- [ ] **Step 6: Full suite, commit, open the PR**

Run: `cd cli && go build ./... && go vet ./... && go test -count=1 ./...`

```bash
git add decisions/ plugins/trellis/VERSION install.sh docs/superpowers/plans/
git commit -m "decision-0086: the injected copy degrades whenever the assembly is over budget"
```

Branch name is `<category>/<slug>` per `AGENTS.md`; `feature/*` branches carry the issue key. Rebase on `origin/main` before opening. **Do not merge** — the maintainer merges (`floor-intent-gate`).

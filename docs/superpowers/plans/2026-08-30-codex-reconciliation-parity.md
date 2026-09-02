# Codex Reconciliation Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The Codex hook reconciles a mismatched `.trellis/rules.toml` exactly as the Claude hook does, instead of failing closed and delivering nothing.

**Architecture:** `parseRulesToml` stops being a pass/fail gate and becomes a classifier: slug-set mismatches (missing, unknown, duplicate) reconcile; a missing `strictness` falls back to adaptive as Claude already does; genuine syntax faults still fail closed. A new `reconcileRows()` mirrors the awk in `staleness.sh`, and a conformance test feeds one fixture table through **both** hooks asserting byte-identical output, so the two implementations cannot drift. TRL-29 folds in: over budget, the injected context drops provenance comments and says so, while the file the mandate writes keeps them.

**Tech Stack:** Node ESM (`codex-context.mjs`), POSIX shell + awk (`staleness.sh`, reference only), Go tests (`cli/`).

**Spec:** `docs/superpowers/specs/2026-08-30-codex-reconciliation-parity-design.md`

## Global Constraints

- **The hook never writes.** `decision-0070` D4. The hook emits text; the *agent* writes the file. No task adds a filesystem write to `codex-context.mjs`.
- **Quarantine, never delete.** Every repair is additive or commenting. No row's value is lost on any path.
- **Every exit is `process.exit(0)`.** A hook must never fail the session.
- **Reconcile only the slug set.** Malformed rows, unknown top-level keys, and a `strictness` present but not `firm`/`adaptive` still fail closed.
- **No deletion verb in any hook output string**, on either host.
- **Byte-identical reconciliation across hosts** — enforced by the Task 2 conformance guard, not by inspection.
- **The cap stays under Codex's ~2,500-token spill threshold.** Spilling hands the model a head-and-tail preview — partial governance, worse than an announced omission.
- **Run after every change:** `cd cli && go build ./... && go vet ./... && go test ./...`

**Reference — the Claude reconciler this must match** is the awk block in `plugins/trellis/hooks/staleness.sh` (search `reconciled=`). Read it before Task 1; it is the specification for `reconcileRows()`.

---

### Task 1: `reconcileRows()` and the classifier split

**Files:**
- Modify: `plugins/trellis/hooks/codex-context.mjs` — `parseRulesToml` (~`:231-300`), and the call site (~`:491-496`)
- Test: `cli/codex_hook_test.go`

**Interfaces:**
- Produces: **`parseRulesToml(source, slugs)`** returns `{ rows: Map<string,boolean>, mismatch: {missing: string[], unknown: string[], duplicate: string[]} | null }` on a parseable file, or **`null`** only for a genuine syntax fault. Callers must distinguish the two.
- Produces: **`reconcileRows(source, slugs, stamp, today)`** returns `{ text: string, added: number, quarantined: number }`.

- [ ] **Step 1: Write the failing tests**

Append to `cli/codex_hook_test.go`:

```go
// Codex used to fail closed on ANY mismatch, so a single bad row cost all
// sixteen rules — the TRL-20 blackout, still live on this host after the Claude
// side was fixed. It now reconciles, exactly as staleness.sh does.
func TestCodexReconcilesInsteadOfFailingClosed(t *testing.T) {
	pluginRoot := codexPluginRoot(t)
	files := payloadFiles()

	run := func(t *testing.T, toml string) (string, codexHookResult) {
		t.Helper()
		project := t.TempDir()
		writeValidCodexOverlay(t, project)
		p := filepath.Join(project, ".trellis", "rules.toml")
		if err := os.WriteFile(p, []byte(toml), 0o644); err != nil {
			t.Fatal(err)
		}
		return runCodexHook(t, pluginRoot, startupInput(t, project))
	}

	t.Run("a missing row no longer blacks out delivery", func(t *testing.T) {
		short := strings.Replace(files["rules-a.toml"],
			"inv-minimal-first         = { active = true }\n", "", 1)
		if short == files["rules-a.toml"] {
			t.Fatal("fixture removed nothing — the case would prove nothing")
		}
		raw, got := run(t, short)
		if strings.Contains(raw, "invalid-rules") {
			t.Errorf("a missing row must reconcile, not fail closed:\n%s", raw)
		}
		if got.HookSpecificOutput == nil {
			t.Fatalf("no context was injected:\n%s", raw)
		}
		ctx := got.HookSpecificOutput.AdditionalContext
		if !regexp.MustCompile(`(?m)^inv-minimal-first[ \t]*=[ \t]*\{[ \t]*active[ \t]*=[ \t]*true`).MatchString(ctx) {
			t.Errorf("the missing row must be reconciled to active = true:\n%s", ctx)
		}
	})

	t.Run("an unknown row is quarantined, never dropped", func(t *testing.T) {
		bogus := files["rules-a.toml"] + "inv-not-a-real-rule       = { active = false }\n"
		_, got := run(t, bogus)
		if got.HookSpecificOutput == nil {
			t.Fatal("an unknown row must reconcile, not fail closed")
		}
		ctx := got.HookSpecificOutput.AdditionalContext
		if !strings.Contains(ctx, "# inv-not-a-real-rule") {
			t.Errorf("the unknown row must survive as a commented row:\n%s", ctx)
		}
		if !strings.Contains(ctx, "quarantined") {
			t.Errorf("the quarantine must be labelled:\n%s", ctx)
		}
	})

	t.Run("a duplicate keeps the first occurrence", func(t *testing.T) {
		dup := files["rules-a.toml"] + "inv-minimal-first         = { active = false }\n"
		_, got := run(t, dup)
		if got.HookSpecificOutput == nil {
			t.Fatal("a duplicate must reconcile, not fail closed")
		}
		if !strings.Contains(got.HookSpecificOutput.AdditionalContext,
			"# inv-minimal-first         = { active = false }") {
			t.Errorf("the extra occurrence must be quarantined verbatim:\n%s",
				got.HookSpecificOutput.AdditionalContext)
		}
	})

	t.Run("a missing strictness falls back to adaptive, as Claude does", func(t *testing.T) {
		noStrict := strings.Replace(files["rules-a.toml"],
			"strictness  = \"firm\"\n", "", 1)
		if noStrict == files["rules-a.toml"] {
			t.Fatal("fixture removed nothing — the case would prove nothing")
		}
		raw, got := run(t, noStrict)
		if got.HookSpecificOutput == nil {
			t.Fatalf("a missing strictness must not be fatal — Claude defaults it:\n%s", raw)
		}
		if !strings.Contains(got.HookSpecificOutput.AdditionalContext, "**By default**") {
			t.Errorf("the adaptive header must be selected:\n%s",
				got.HookSpecificOutput.AdditionalContext)
		}
	})

	// The permissive direction is the dangerous one: reconciliation must never
	// paper over a file it cannot parse.
	t.Run("genuine syntax faults still fail closed", func(t *testing.T) {
		for name, toml := range map[string]string{
			"invalid strictness value": strings.Replace(files["rules-a.toml"],
				`strictness  = "firm"`, `strictness  = "bogus"`, 1),
			"unknown top-level key": "nonsense = \"x\"\n" + files["rules-a.toml"],
			"malformed row":         files["rules-a.toml"] + "inv-broken = notatable\n",
		} {
			t.Run(name, func(t *testing.T) {
				raw, got := run(t, toml)
				if got.HookSpecificOutput != nil {
					t.Errorf("a syntax fault must fail closed, not reconcile:\n%s", raw)
				}
				if !strings.Contains(raw, "invalid-rules") {
					t.Errorf("want invalid-rules, got:\n%s", raw)
				}
			})
		}
	})
}
```

Add the helper if `codexPluginRoot` does not already exist — check first; `cli/codex_hook_test.go` builds a plugin root in `TestCodexHookContract`, and reuse is preferred over a second copy:

```go
// codexPluginRoot stages the shipped payload in a temp plugin root.
func codexPluginRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range payloadFiles() {
		if err := os.WriteFile(filepath.Join(root, "reference", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./... -run TestCodexReconcilesInsteadOfFailingClosed -v`
Expected: the first four subtests FAIL with `invalid-rules` (the hook fails closed); the syntax-fault subtests PASS already (they are the behaviour being preserved — that is correct and expected).

- [ ] **Step 3: Split the classifier**

In `codex-context.mjs`, change `parseRulesToml` so slug-set problems are **collected** rather than returned as `null`:

- Replace `if (!row || rows.has(row[1]) || !slugSet.has(row[1])) return null;` — keep `return null` for `!row` (a malformed row is a syntax fault), but record `rows.has(row[1])` into a `duplicate` list and `!slugSet.has(row[1])` into an `unknown` list, continuing the scan.
- Replace the tail block: keep `return null` when `!rulesSectionSeen` or when `strictness` is present but is neither `firm` nor `adaptive`. **A missing `strictness` is no longer fatal** — leave it unset and let the caller default it, matching `staleness.sh:558-560`.
- Compute `missing` as the slugs with no row.
- Return `{ rows, mismatch }` where `mismatch` is `null` when all three lists are empty, else `{ missing, unknown, duplicate }`.

Add `reconcileRows(source, slugs, stamp, today)` mirroring the awk in `staleness.sh`: walk the source line by line; a row whose slug is unknown or already seen is emitted as `"# " + line + note`; every other line passes through verbatim; missing slugs are appended at the end as `slug + " = { active = true }"` under a single dated header comment; a `[rules]` header is emitted first when the file has none, matched leniently (`/^[ \t]*\[rules\]/`). Provenance strings must match the Claude side **byte for byte** — Task 2 pins this.

At the call site, when `mismatch` is non-null, reconcile and use the reconciled text in place of `rulesToml`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./...`
Expected: PASS. If `TestCodexHookContract`'s existing failure cases now behave differently, **stop and report** — that test pins the fail-closed contract and a change there is a design question, not an implementation detail.

- [ ] **Step 5: Commit**

```bash
git add plugins/trellis/hooks/codex-context.mjs cli/codex_hook_test.go
git commit -m "fix(codex): reconcile a slug-set mismatch instead of failing closed

parseRulesToml stops being a pass/fail gate and becomes a classifier: missing,
unknown and duplicate slugs reconcile as they do on the Claude hook, a missing
strictness falls back to adaptive, and genuine syntax faults still fail closed.
Closes the TRL-20 blackout on the second host."
```

---

### Task 2: The cross-host conformance guard

**Files:**
- Test: `cli/codex_hook_test.go` (new test)

**Interfaces:**
- Consumes: `reconcileRows()` via the Codex hook's output, and the Claude reconciler via `staleness.sh`'s output.
- Produces: nothing consumed downstream.

**Why this task exists.** Two implementations of one ratified semantic is a deliberate choice (`decision-0028`: a guard per pair). The guard is what makes it safe. Without it the two silently diverge and the divergence ships.

- [ ] **Step 1: Write the failing test**

```go
// Two implementations of one ratified semantic (decision-0083's table) are only
// safe with a guard that makes drift a test failure rather than a field report.
// This is decision-0028's "a guard per pair", applied to the reconciler.
func TestBothHostsReconcileIdentically(t *testing.T) {
	files := payloadFiles()
	base := files["rules-a.toml"]

	fixtures := map[string]string{
		"rename (missing + unknown together)": strings.Replace(base,
			"inv-minimal-first         = { active = true }",
			"inv-renamed-first         = { active = true }", 1),
		"indented [rules] table": strings.Replace(base, "[rules]", "  [rules]", 1),
		"already quarantined": base +
			"# inv-gone = { active = true }  # quarantined 2026-01-01: not in payload@old\n",
		"duplicate with a differing value": base +
			"inv-minimal-first         = { active = false }\n",
		"no [rules] table at all": "strictness  = \"firm\"\n",
		"empty file":              "",
		"missing strictness": strings.Replace(base, "strictness  = \"firm\"\n", "", 1),
	}

	for name, toml := range fixtures {
		t.Run(name, func(t *testing.T) {
			claude := claudeReconciledRows(t, toml)
			codex := codexReconciledRows(t, toml)
			if claude != codex {
				t.Errorf("the two hosts reconciled the same file differently — "+
					"decision-0083's table must apply identically to both\n"+
					"claude:\n%s\ncodex:\n%s", claude, codex)
			}
		})
	}
}
```

Both helpers run one hook against a fixture and return the reconciled TOML block from its injected context:

```go
// claudeReconciledRows runs staleness.sh against `toml` and returns the
// reconciled row block it injected.
func claudeReconciledRows(t *testing.T, toml string) string {
	t.Helper()
	out := rulesTomlRun(t)(t, toml)
	return reconciledRowsFromContext(t, out)
}

// codexReconciledRows runs codex-context.mjs against the same fixture and
// returns the same block, so the two can be compared byte for byte.
func codexReconciledRows(t *testing.T, toml string) string {
	t.Helper()
	project := t.TempDir()
	writeValidCodexOverlay(t, project)
	if err := os.WriteFile(filepath.Join(project, ".trellis", "rules.toml"),
		[]byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	_, got := runCodexHook(t, codexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("codex injected nothing for this fixture — it must reconcile, not refuse")
	}
	return codexReconciledRowsFromContext(t, got.HookSpecificOutput.AdditionalContext)
}
```

`rulesTomlRun` and `reconciledRowsFromContext` already exist in `cli/plugin_hook_test.go` — **read them first and reuse them; check `reconciledRowsFromContext`'s exact signature before calling it**, since it was narrowed during the predecessor branch. Write only `codexReconciledRowsFromContext`, locating the block between the payload prose and the mandate section in the Codex context.

Both hooks must be given the *same* posture so the comparison is meaningful: the fixtures above carry `strictness = "firm"` except where the fixture's own point is that it is missing.

- [ ] **Step 2: Run test to verify it fails meaningfully**

Run: `cd cli && go test ./... -run TestBothHostsReconcileIdentically -v`
Expected: it should PASS if Task 1 matched the awk byte for byte. **If it passes on the first run, prove it can fail**: change one character of the Codex provenance string, confirm the test goes red naming the fixture, then restore. Report that mutation's output — a conformance guard that has never failed is not known to work.

- [ ] **Step 3: Fix any divergence Task 1 introduced**

If fixtures fail, the Claude side is the reference — `staleness.sh` shipped first and is under review in PR #258. Change `codex-context.mjs` to match it, not the reverse. **Exception:** if the Claude side is demonstrably wrong on a fixture, stop and report rather than encoding the bug into a second implementation.

- [ ] **Step 4: Run the full suite**

Run: `cd cli && go test ./...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cli/codex_hook_test.go
git commit -m "test: pin byte-identical reconciliation across both hosts

Two implementations of decision-0083's table are safe only with a guard that
turns drift into a test failure (decision-0028: a guard per pair)."
```

---

### Task 3: The repair mandate, and extending the destructive-instruction guards

**Files:**
- Modify: `plugins/trellis/hooks/codex-context.mjs` (context assembly)
- Modify: `cli/plugin_hook_test.go` — `TestEveryDeletionInstructionIsGated`, `TestEveryDestructiveInstructionIsGated`
- Test: `cli/codex_hook_test.go`

**Interfaces:**
- Consumes: `reconcileRows()`'s `{text, added, quarantined}` from Task 1.
- Produces: nothing consumed downstream.

**Why the guard extension is in this task and not a later one.** Both guards scan the Claude hook only. This task puts an agent-facing instruction into the Codex payload for the first time. The design's safety argument is *the repair may be ungated because nothing destructive reaches the agent* — shipping the mandate without extending the guard leaves that argument unenforced on the new host, which is how it was nearly lost on the Claude side.

- [ ] **Step 1: Write the failing tests**

```go
// The repair is applied and REPORTED, not proposed and gated — safe only
// because quarantine loses nothing. What must never be lost is the loudness.
func TestCodexMandatesAndReportsTheRepair(t *testing.T) {
	pluginRoot := codexPluginRoot(t)
	files := payloadFiles()
	project := t.TempDir()
	writeValidCodexOverlay(t, project)
	short := strings.Replace(files["rules-a.toml"],
		"inv-minimal-first         = { active = true }\n", "", 1)
	if err := os.WriteFile(filepath.Join(project, ".trellis", "rules.toml"),
		[]byte(short), 0o644); err != nil {
		t.Fatal(err)
	}
	_, got := runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatal("no context injected")
	}
	ctx := got.HookSpecificOutput.AdditionalContext

	if !strings.Contains(ctx, "Write .trellis/rules.toml") {
		t.Errorf("the mandate must instruct the write:\n%s", ctx)
	}
	if !strings.Contains(ctx, "added 1 row(s)") {
		t.Errorf("the repair must be reported per row:\n%s", ctx)
	}
	for _, verb := range []string{"delete", "remove", "drop"} {
		if strings.Contains(strings.ToLower(ctx), verb+" those rows") {
			t.Errorf("no deletion instruction may reach the agent, found %q:\n%s", verb, ctx)
		}
	}
}
```

Then extend both guards in `cli/plugin_hook_test.go` to scan `codex-context.mjs`'s emitted strings as well as `staleness.sh`'s. Read how each currently collects its message set and add the Codex source alongside, keeping the existing floor assertions so a collapsed scan still fails.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./... -run 'TestCodexMandatesAndReportsTheRepair|Gated' -v`
Expected: `TestCodexMandatesAndReportsTheRepair` FAILS (no mandate yet). The two guards should PASS — they scan a hook with no destructive strings.

- [ ] **Step 3: Add the mandate**

Append to the assembled context, after the reconciled rows and before the stamp line, the Claude mandate's text host-adjusted: state that the rows above are reconciled and govern this session, that the file on disk still differs, that writing the shown rows loses nothing because a row the payload does not ship is commented with its reason and date rather than taken out, and that the agent must report what it reconciled row by row before substantive work. Include the stale-plugin remedy (`codex plugin update` equivalent, or the neutral wording if none exists — check `plugins/trellis/README.md` before inventing a command).

**No deletion verb may appear.** The guards from Step 1 now enforce this.

- [ ] **Step 4: Run tests, and prove the guard extension works**

Run: `cd cli && go test ./...`
Expected: PASS.

Then mutate: insert `"Delete the stale rows and write the rest."` into the Codex mandate string, re-run the two guards, and confirm at least one FAILS naming that string. Restore and confirm green. **Report the mutation output** — an extended guard that has never caught anything is not known to scan the new source.

- [ ] **Step 5: Commit**

```bash
git add plugins/trellis/hooks/codex-context.mjs cli/plugin_hook_test.go cli/codex_hook_test.go
git commit -m "feat(codex): mandate and report the reconciliation repair

Extends both destructive-instruction guards to scan codex-context.mjs, so the
safety argument — ungated because nothing destructive reaches the agent — is
enforced by construction on the new host too."
```

---

### Task 4: Degrade over budget instead of refusing (TRL-29)

**Files:**
- Modify: `plugins/trellis/hooks/codex-context.mjs` — the budget check (~`:505-508`)
- Test: `cli/codex_hook_test.go`

**Interfaces:**
- Consumes: the assembled context from Task 3.
- Produces: nothing consumed downstream.

- [ ] **Step 1: Write the failing test**

```go
// Refusing to emit is a self-inflicted blackout: Codex itself spills oversized
// context to disk with a preview rather than rejecting it, so failing closed is
// strictly worse than the host's own degradation. Governance survives; the
// provenance comments are what give way, and the omission is announced.
func TestCodexDegradesRatherThanRefusingOverBudget(t *testing.T) {
	pluginRoot := codexPluginRoot(t)
	files := payloadFiles()
	project := t.TempDir()
	writeValidCodexOverlay(t, project)

	// The worst case: every row foreign, so all 16 quarantine AND all 16 add.
	var b strings.Builder
	b.WriteString("strictness  = \"firm\"\n\n[rules]\n")
	for i := 0; i < 16; i++ {
		fmt.Fprintf(&b, "inv-foreign-rule-%02d = { active = true }\n", i)
	}
	if err := os.WriteFile(filepath.Join(project, ".trellis", "rules.toml"),
		[]byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))

	if strings.Contains(raw, "context-over-budget") {
		t.Fatalf("the hook must degrade, not refuse — refusing is the blackout:\n%s", raw)
	}
	if got.HookSpecificOutput == nil {
		t.Fatalf("no context injected:\n%s", raw)
	}
	ctx := got.HookSpecificOutput.AdditionalContext
	for _, slug := range assessableSlugs {
		if !regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(slug) + `[ \t]*=`).MatchString(ctx) {
			t.Errorf("rule %s must still be delivered when provenance is dropped", slug)
		}
	}
	if !strings.Contains(ctx, "provenance") {
		t.Errorf("the omission must be announced, not silent:\n%s", ctx)
	}
	if n := len([]byte(ctx)); n > 9500 {
		t.Errorf("degraded context is %d bytes, still over the cap", n)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./... -run TestCodexDegradesRatherThanRefusingOverBudget -v`
Expected: FAIL with `context-over-budget` — the hook still refuses.

- [ ] **Step 3: Implement the degradation**

Replace the `fail("assembled-context", "context-over-budget")` branch: when the assembled context exceeds `MAX_CONTEXT_BYTES`, re-assemble using a **provenance-free** variant of the reconciled text — the same rows, with the `# quarantined …` and `# added …` comments omitted — and append one line stating that provenance was omitted to fit the context budget and remains in full in the file the mandate instructs writing.

**The mandate still instructs writing the full-provenance text.** The file is the archive; only the injected working set is trimmed. Keep a final hard refusal if even the provenance-free assembly exceeds the cap — that is a runaway guard, not a governance decision.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./...`
Expected: PASS, including `TestReconciledCodexPayloadFitsContextBudget` and `TestBothHostsReconcileIdentically` — the degradation must not change what the *file* would contain.

- [ ] **Step 5: Commit**

```bash
git add plugins/trellis/hooks/codex-context.mjs cli/codex_hook_test.go
git commit -m "fix(codex): degrade over budget instead of refusing (TRL-29)

Codex spills oversized context rather than rejecting it, so failing closed was
strictly worse than the host's own behaviour. Governance now survives: the
provenance comments give way, the omission is announced, and the file the
mandate writes keeps them in full."
```

---

### Task 5: Record the decision and ship the payload

**Files:**
- Create: `decisions/0084-codex-reaches-reconciliation-parity.md`
- Modify: `decisions/0083-rules-toml-reconciles-itself.md` (frontmatter forward pointer only)
- Modify: `plugins/trellis/VERSION`, and the four parity annotations
- Modify: `README.md`, `plugins/trellis/README.md` if they scope reconciliation to Claude

**Interfaces:**
- Consumes: the shipped behaviour from Tasks 1-4.

- [ ] **Step 1: Write the decision record**

Create `decisions/0084-codex-reaches-reconciliation-parity.md`. **No `status:` field** (`decision-0082`); `depends_on: [decision-0083]`; `owner: agent`; `date: 2026-08-30`. It must cover:

1. **What changed** — the classifier split, and precisely which conditions still fail closed. Name the permissive-direction risk explicitly.
2. **Why Codex moved on `strictness` rather than the reconciler writing the key** — fixing the divergence at its root repairs every existing partial file, not only newly reconciled ones, and keeps the reconciler from inventing configuration.
3. **Why TRL-29 folded in** — reconciliation only adds bytes, so shipping it alone would rebuild the blackout at the budget boundary. Record that Codex *spills* rather than rejects, so the old refusal was strictly worse than the host's own behaviour, citing https://learn.chatgpt.com/docs/hooks.
4. **Two implementations with a guard, and its expiry** — `decision-0028`'s "a guard per pair"; extraction becomes correct at a third host.
5. **What this supersedes** — `decision-0083`'s Claude-only scoping, and only that.

- [ ] **Step 2: Add the forward pointer to decision-0083**

Append-only — frontmatter, no prose change:

```yaml
superseded_in_part_by: [decision-0084]  # 2026-08-30 — the Claude-only scoping of §1 and its parity open question only; decision-0084 gives codex-context.mjs the same reconciliation. Everything else in 0083 — the resolution table, the ungated-write argument, the quarantine semantics — stands unchanged.
```

- [ ] **Step 3: Close out the four parity annotations**

`docs/superpowers/plans/2026-08-30-rules-toml-self-repair.md` (two: the Architecture note and the Global Constraints note) and `docs/superpowers/specs/2026-08-30-rules-toml-self-repair-design.md` (two: the resolution-table note and the Codex-parity note) each say parity is owed and name TRL-30.

**Do not edit or delete them.** Add a second dated line beside each recording that the debt was paid and naming `decision-0084`, so a reader arriving at the "owed" note is not left believing it still stands.

- [ ] **Step 4: Bump VERSION and sweep the prose**

`plugins/trellis/VERSION` → `0.8.0`. A payload change that does not move it never reaches an installed copy. Check both READMEs for text scoping reconciliation to Claude and correct it.

- [ ] **Step 5: Run the suite and the corpus reviewer**

Run: `cd cli && go build ./... && go vet ./... && go test ./...`
Expected: PASS. Then invoke the repo-owned `corpus-reviewer` subagent against `decisions/`. Two pre-existing violations (`decisions/0044:5`, `research/0010:5`) are known and out of scope; anything else is yours to fix.

- [ ] **Step 6: Commit**

```bash
git add decisions/ plugins/trellis/VERSION README.md plugins/trellis/README.md docs/superpowers/
git commit -m "decision-0084: Codex reaches reconciliation parity"
```

**Do not push and do not open a PR** — PR #259 already exists as a draft for this branch. Report that the work is ready to be marked ready-for-review; the maintainer decides.

---

## Verification the plan depends on

Two guards do most of the work here and must both be proven by mutation rather than by a green run:

- **Task 2's conformance guard** — break one character of the Codex provenance and confirm it goes red naming the fixture.
- **Task 3's extended destructive-instruction guards** — insert a destructive mandate into the Codex string and confirm at least one guard fails.

The predecessor branch produced five tests that passed for the wrong reason, every one caught only by mutation. A guard that has never failed is not known to work.

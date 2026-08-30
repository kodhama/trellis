# Self-Repairing `.trellis/rules.toml` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A slug mismatch in `.trellis/rules.toml` stops blacking out the whole rule set — the hook reconciles in memory, delivers every rule, and hands the agent the exact repaired file to write and report.

**Architecture:** The hook gains one reconciler that turns the project's rows plus the payload's slug list into a single reconciled TOML text. That one text is used twice: injected as the activation rows so delivery never fails, and quoted in the emit as the literal content the agent writes back to disk. `missing:` rows are added `active = true`; `unknown:` and duplicate rows are commented out with dated provenance, never deleted. Because nothing is ever lost, the write needs no confirmation gate — only a report. Codex reaches the same semantics by deriving its slug set from `reference/rules.md` instead of a hardcoded array.

> **2026-08-30, after execution — this last sentence overstates what Task 3 delivered.** Deriving the slug set removed the drift hazard and the false `unknown:` reason; it did **not** give Codex the reconcile semantics. `parseRulesToml` still returns `null` on any mismatch and the hook still fails closed with `invalid-rules`, so the blackout this plan retires is retired on Claude only. Recorded here rather than edited away: the deferral was deliberate. See `decision-0083` §1 and its open questions; tracked as [TRL-30](https://linear.app/kodhama/issue/TRL-30).

**Tech Stack:** POSIX shell + awk (`staleness.sh`), Node ESM (`codex-context.mjs`), Go tests (`cli/`).

**Spec:** `docs/superpowers/specs/2026-08-30-rules-toml-self-repair-design.md`

## Global Constraints

- **The hook never writes.** `decision-0070` D4 stands; `cli/plugin_hook_test.go:1516` pins it. The hook emits text; the *agent* writes the file. No task adds a write to either hook.
- **Quarantine, never delete.** Every repair is additive or commenting. No row's value is ever lost.
- **No deletion verbs in any hook emit string.** `TestEveryDeletionInstructionIsGated` (`cli/plugin_hook_test.go:579`) and `TestEveryDestructiveInstructionIsGated` (`:506`) scan emit strings and require an `explicit confirmation` clause on each match. Keeping the new strings non-destructive is what keeps the repair ungated.
- **A hook must never fail the session** — always `exit 0`.
- **Injection budget stays 32768 bytes** (`staleness.sh:645`).
- **Both hosts, same semantics.** A behaviour added to `staleness.sh` is added to `codex-context.mjs` in Task 3.

  > **2026-08-30, after execution — this constraint was not met, and the note at line 9 above says why.** Task 3 delivered the slug derivation and the raised context cap; it did **not** give `codex-context.mjs` the reconcile semantics. `parseRulesToml` still returns `null` on any mismatch — missing, unknown and duplicate alike — and the hook still fails closed with `invalid-rules`, so reconciliation is **Claude-only** and the constraint as written above is unsatisfied. Recorded rather than edited away: what this plan originally required is part of the story, and the deferral was deliberate. The gap is tracked as [TRL-30](https://linear.app/kodhama/issue/TRL-30); see `decision-0083` §1 and its open questions.
- **Provenance format, exact:** `  # quarantined <YYYY-MM-DD>: not in <payload-stamp>. If a newer Trellis ships this slug, run \`claude plugin update trellis@kodhama\` and uncomment.`
- **Run after every change:** `cd cli && go test ./...`

---

### Task 1: The reconciler — a mismatch delivers every rule

**Files:**
- Modify: `plugins/trellis/hooks/staleness.sh:576-616` (slug validation and the blackout emit)
- Test: `cli/plugin_hook_test.go` (new tests appended)

**Interfaces:**
- Consumes: `$rules` (payload `reference/rules.md`), `$toml` (project file), `$current` (payload stamp), `$slug_report` — all already in scope at `staleness.sh:612`.
- Produces: shell variable **`$reconciled`** — the complete reconciled TOML text — and **`$repair_summary`**, a one-line human summary like `added 1 row; quarantined 1 row`. Task 2 consumes both.

- [ ] **Step 1: Write the failing test**

Append to `cli/plugin_hook_test.go`:

```go
// rulesTomlRun builds a plugin root from the shipped payload and returns a
// runner that writes `rows` to .trellis/rules.toml in a fresh project, then
// returns the hook's raw stdout.
func rulesTomlRun(t *testing.T) func(*testing.T, string) string {
	t.Helper()
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range payloadFiles() {
		if err := os.WriteFile(filepath.Join(pluginRoot, "reference", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return func(t *testing.T, rows string) string {
		t.Helper()
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "rules.toml"), []byte(rows), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, _ := cmd.CombinedOutput()
		return string(out)
	}
}

// hookSlugs returns every distinct rule slug appearing anywhere in the hook's
// output — the same shape as the `injected` closure at plugin_hook_test.go:1458.
func hookSlugs(out string) map[string]bool {
	slugs := map[string]bool{}
	for _, m := range regexp.MustCompile(`(inv|floor)-[a-z-]+`).FindAllString(out, -1) {
		slugs[m] = true
	}
	return slugs
}

// A slug mismatch used to inject NOTHING — one bad row cost all sixteen rules,
// every session, until a human edited the file (TRL-20). Delivery now
// reconciles in memory, so a mismatch degrades to a repair notice rather than
// a blackout.
func TestSlugMismatchStillDeliversEveryRule(t *testing.T) {
	run := rulesTomlRun(t)
	files := payloadFiles()

	t.Run("a missing row does not black out the other rules", func(t *testing.T) {
		short := strings.Replace(files["rules-b.toml"],
			"inv-minimal-first         = { active = true }\n", "", 1)
		if short == files["rules-b.toml"] {
			t.Fatal("fixture removed nothing — the case would prove nothing")
		}
		out := run(t, short)
		if strings.Contains(out, "Nothing was injected") {
			t.Errorf("a missing row must not black out delivery:\n%s", out)
		}
		got := hookSlugs(out)
		for _, slug := range assessableSlugs {
			if !got[slug] {
				t.Errorf("rule %s was not delivered after reconciliation; got %v", slug, keysOfBool(got))
			}
		}
	})

	t.Run("the missing row is reconciled to active = true", func(t *testing.T) {
		short := strings.Replace(files["rules-b.toml"],
			"inv-minimal-first         = { active = true }\n", "", 1)
		out := run(t, short)
		if !strings.Contains(out, "inv-minimal-first = { active = true }") {
			t.Errorf("the reconciled rows must add the missing slug as active:\n%s", out)
		}
	})

	t.Run("an unknown row is quarantined, never dropped", func(t *testing.T) {
		bogus := files["rules-b.toml"] + "inv-not-a-real-rule       = { active = false }\n"
		out := run(t, bogus)
		if !strings.Contains(out, "# inv-not-a-real-rule") {
			t.Errorf("an unknown row must survive as a commented-out row:\n%s", out)
		}
		if !strings.Contains(out, "quarantined") {
			t.Errorf("the quarantine must be labelled so a reader can tell why:\n%s", out)
		}
		if !strings.Contains(out, "claude plugin update trellis@kodhama") {
			t.Errorf("quarantine provenance must name the stale-plugin cause (TRL-27):\n%s", out)
		}
	})

	t.Run("a duplicate keeps the first occurrence and quarantines the extra", func(t *testing.T) {
		dup := files["rules-b.toml"] + "inv-minimal-first         = { active = false }\n"
		out := run(t, dup)
		if !strings.Contains(out, "# inv-minimal-first = { active = false }") {
			t.Errorf("the extra occurrence must be quarantined, not deleted:\n%s", out)
		}
		if !strings.Contains(out, "inv-minimal-first         = { active = true }") {
			t.Errorf("the FIRST occurrence must survive verbatim:\n%s", out)
		}
	})

	t.Run("a rename is both kinds at once and both are reconciled", func(t *testing.T) {
		renamed := strings.Replace(files["rules-b.toml"],
			"inv-minimal-first         = { active = true }",
			"inv-renamed-first         = { active = true }", 1)
		if renamed == files["rules-b.toml"] {
			t.Fatal("fixture did not rename anything — the case would prove nothing")
		}
		out := run(t, renamed)
		if !strings.Contains(out, "# inv-renamed-first") {
			t.Errorf("the stale slug must be quarantined:\n%s", out)
		}
		if !strings.Contains(out, "inv-minimal-first = { active = true }") {
			t.Errorf("the new slug must be added:\n%s", out)
		}
	})

	t.Run("a quarantined row is invisible next session — the repair is idempotent", func(t *testing.T) {
		quarantined := files["rules-b.toml"] +
			"# inv-not-a-real-rule = { active = false }  # quarantined 2026-08-30: not in payload@test\n"
		out := run(t, quarantined)
		if strings.Contains(out, "does not match the rules the installed plugin ships") {
			t.Errorf("an already-repaired file must draw no repair notice at all:\n%s", out)
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./... -run TestSlugMismatchStillDeliversEveryRule -v`
Expected: FAIL — the subtests report `Nothing was injected` and no reconciled rows, because `staleness.sh:614` still exits before assembling the payload.

- [ ] **Step 3: Write the reconciler**

In `plugins/trellis/hooks/staleness.sh`, replace the `if [ "$slug_report" != "ok" ]; then … exit 0; fi` block at `:612-616` with reconciliation. Insert **before** the `payload="$(` assembly:

```sh
# Reconcile rather than refuse. A mismatch used to inject nothing at all, so a
# single bad row cost all sixteen rules every session until a human edited the
# file (TRL-20). The rules the payload ships are still the authority; what
# changes is that an unmatched row is quarantined instead of blocking delivery.
#
# Quarantine — commenting the row out rather than deleting it — is what makes
# an ungated repair safe. `unknown:` has two causes with opposite repairs (the
# rule was retired, or the installed plugin is behind the project's config,
# TRL-27) and config-only mode carries no version stamp to tell them apart. A
# commented row is correct under both readings and loses nothing either way.
# It is also invisible to the validator above, which anchors rows at line
# start, so a repaired file draws no second notice.
reconciled=""
repair_summary=""
if [ "$slug_report" != "ok" ]; then
  today="$(date +%Y-%m-%d)"
  reconciled="$(
    awk -v want_src="$rules" -v stamp="$current" -v today="$today" '
      BEGIN {
        while ((getline line < want_src) > 0) {
          if (match(line, /`(inv|floor)-[a-z-]+`[[:space:]]*$/)) {
            s = substr(line, RSTART + 1, RLENGTH - 2)
            sub(/`[[:space:]]*$/, "", s)
            want[s] = 1
            order[++n] = s
          }
        }
        note = "  # quarantined " today ": not in " stamp ". If a newer Trellis" \
               " ships this slug, run `claude plugin update trellis@kodhama` and uncomment."
      }
      /^[[:space:]]*(inv|floor)-[a-z-]+[[:space:]]*=/ {
        row = $1
        sub(/[^a-z-].*$/, "", row)
        if (!(row in want) || (row in seen)) {
          print "# " $0 note
          quarantined++
          next
        }
        seen[row] = 1
        print
        next
      }
      { print }
      END {
        for (i = 1; i <= n; i++) {
          s = order[i]
          if (!(s in seen)) print s " = { active = true }  # added " today " from " stamp
        }
      }
    ' "$toml"
  )"
  # Counted from the reconciled text itself rather than passed out of awk — one
  # source of truth, and no second channel to keep in step with the first.
  added="$(printf '%s\n' "$reconciled" | grep -c '# added ' || true)"
  quarantined="$(printf '%s\n' "$reconciled" | grep -c '# quarantined ' || true)"
  repair_summary="added ${added:-0} row(s); quarantined ${quarantined:-0} row(s)"
fi
```

No temp file and no second output channel are needed: awk writes only the reconciled text to stdout.

Then change the row emission inside the `payload="$(…)"` assembly so the reconciled text wins when there is one — replace the bare `cat "$toml"` at `:637`:

```sh
  if [ -n "$reconciled" ]; then
    printf '%s\n' "$reconciled"
  else
    cat "$toml"
  fi
```

and, in the same `if/else` that chooses the row preamble, add a third arm before the two existing ones:

```sh
  if [ -n "$reconciled" ]; then
    printf 'Rows from this project'"'"'s .trellis/rules.toml, RECONCILED against the rules this payload ships (%s) — the file on disk still differs. Apply a rule only when its row says active = true; the two floor rules apply regardless of their row.\n\n' "$repair_summary"
  elif [ "${rows_are_default:-no}" = yes ]; then
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./... -run TestSlugMismatchStillDeliversEveryRule -v`
Expected: PASS, all six subtests.

Then the full suite: `cd cli && go test ./...`
Expected: `TestRepairRemedyCoversEveryMismatchKind` and `TestRowMismatchRemedyIsNotDestructive` now FAIL — they pin the retired blackout remedy. **Leave them failing; Task 2 retires them deliberately.** Every other test must still pass; if any other test fails, stop and report rather than adjusting it.

- [ ] **Step 5: Commit**

```bash
git add plugins/trellis/hooks/staleness.sh cli/plugin_hook_test.go
git commit -m "fix(hook): reconcile a rules.toml slug mismatch instead of blacking out delivery

A single unmatched row cost all sixteen rules every session until someone
hand-edited the file (TRL-20). The hook now reconciles in memory: missing
slugs are added active = true, unknown and duplicate rows are quarantined by
commenting them out with dated provenance. Nothing is deleted, so nothing is
lost, and a repaired file draws no second notice."
```

---

### Task 2: The repair mandate — announced, not asked

**Files:**
- Modify: `plugins/trellis/hooks/staleness.sh` (the emit that replaces the blackout)
- Modify: `cli/plugin_hook_test.go:620-704` (`TestRepairRemedyCoversEveryMismatchKind`), `:799` (`TestRowMismatchRemedyIsNotDestructive`)

**Interfaces:**
- Consumes: `$reconciled`, `$repair_summary`, `$slug_report` from Task 1.
- Produces: nothing consumed downstream — this task is the user-facing surface.

- [ ] **Step 1: Write the failing test**

Append to `cli/plugin_hook_test.go`:

```go
// The repair is applied and REPORTED, not proposed and gated. decision-0072's
// finding #6 — "retiring a confirm-gated writer silently retires the gate" —
// is answered by quarantine semantics: no prior value is ever lost, so the
// gate that guarded destructive writes is not engaged. What must never be
// lost is the loudness.
func TestReconciledRepairIsMandatedAndReported(t *testing.T) {
	run := rulesTomlRun(t)
	files := payloadFiles()
	short := strings.Replace(files["rules-b.toml"],
		"inv-minimal-first         = { active = true }\n", "", 1)
	out := run(t, short)

	t.Run("the agent is told to write the file, not to ask", func(t *testing.T) {
		if !strings.Contains(out, "Write .trellis/rules.toml") {
			t.Errorf("the emit must mandate the write:\n%s", out)
		}
		if strings.Contains(out, "get explicit confirmation before writing") {
			t.Errorf("the confirm-first remedy is retired by this change:\n%s", out)
		}
	})

	t.Run("the agent must report what changed, before other work", func(t *testing.T) {
		if !strings.Contains(out, "Tell the user") {
			t.Errorf("a silent repair is the failure this design exists to prevent:\n%s", out)
		}
		if !strings.Contains(out, "added 1 row(s)") {
			t.Errorf("the emit must state what it reconciled, per row:\n%s", out)
		}
	})

	t.Run("the emit carries the literal file content to write", func(t *testing.T) {
		// The agent re-deriving the repair is how a wrong one lands. The hook
		// computes it once and shows exactly the bytes to save.
		if !strings.Contains(out, "inv-minimal-first = { active = true }  # added") {
			t.Errorf("the emit must quote the exact reconciled file:\n%s", out)
		}
	})

	t.Run("no deletion verb enters the emit", func(t *testing.T) {
		// Keeping the repair non-destructive is what keeps it ungated —
		// TestEveryDeletionInstructionIsGated would otherwise demand a
		// confirmation clause and re-impose the gate this change removes.
		for _, verb := range []string{"delete those rows", "remove those rows", "drop the unknown"} {
			if strings.Contains(out, verb) {
				t.Errorf("the reconciled remedy must never instruct a deletion, found %q:\n%s", verb, out)
			}
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./... -run TestReconciledRepairIsMandatedAndReported -v`
Expected: FAIL — the emit still carries the retired confirm-first wording.

- [ ] **Step 3: Write the emit**

The repair notice rides **with** the delivered rules, not instead of them — there is no separate `emit` and no early `exit`. Add this inside the `payload="$(…)"` assembly, after the reconciled rows are printed and before the `Delivered by the Trellis plugin` provenance line:

```sh
  if [ -n "$reconciled" ]; then
    printf '\n## Rule activation was reconciled this session\n\n'
    printf 'This project'"'"'s .trellis/rules.toml did not match the rules this payload ships (%s). The rows above are the reconciled set and are what governs this session; the file on disk still differs. Reconciliation: %s.\n\n' "$slug_report" "$repair_summary"
    printf 'Write .trellis/rules.toml with exactly the rows shown above, so the file matches what governs. Nothing is lost by this: a row the payload does not ship is commented out with the reason and the date, never removed, and every value the project chose is preserved verbatim. Tell the user what you reconciled, row by row, before doing substantive work — a repair they did not see is the failure this reconciliation exists to prevent. If a quarantined slug is one a newer Trellis release added, the installed plugin is the stale side: `claude plugin update trellis@kodhama`, restart the session, and uncomment the row.\n'
  fi
```

- [ ] **Step 4: Retire the two pins that guarded the blackout remedy**

In `cli/plugin_hook_test.go`, `TestRepairRemedyCoversEveryMismatchKind`: the `governed = false` subtest is untouched (the opt-out is still absolute). Replace the two subtests that assert the retired remedy with their reconciliation counterparts, and record why above the function:

```go
// Retired by the reconciliation change (TRL-20/TRL-2/TRL-27): the remedy this
// guarded — "for missing:, add those slugs; for unknown:, remove those rows;
// for duplicate:, delete the extra occurrences" — no longer exists. Its intent
// survives in TestSlugMismatchStillDeliversEveryRule, which asserts every kind
// is RESOLVED rather than merely explained. The `governed = false` subtest is
// unrelated to the remedy and stays.
```

Delete the subtests `"a renamed slug reports BOTH categories, not the first"` and `"a duplicated slug is reported AND its repair is explained"` — both cases are now covered by the rename and duplicate subtests in Task 1, which assert the stronger property.

For `TestRowMismatchRemedyIsNotDestructive` (`:799`): read the whole test, and if its assertions are satisfied by the new emit, leave it untouched. If it pins retired wording, narrow it the same way with a comment naming this change. **Do not weaken an assertion to make it pass** — if it fails for a real reason, stop and report.

- [ ] **Step 5: Run the full suite**

Run: `cd cli && go test ./...`
Expected: PASS, everything. Confirm specifically that `TestEveryDestructiveInstructionIsGated` and `TestEveryDeletionInstructionIsGated` still pass — the new strings must not have tripped either.

- [ ] **Step 6: Commit**

```bash
git add plugins/trellis/hooks/staleness.sh cli/plugin_hook_test.go
git commit -m "fix(hook): mandate and report the rules.toml repair instead of gating it

Quarantine loses nothing, so the confirm gate that guarded destructive row
writes is no longer engaged (decision-0072 finding #6). The emit now carries
the exact reconciled file and requires the agent to report what changed before
substantive work. Retires the two pins that guarded the blackout remedy."
```

---

### Task 3: Codex parity — derive the slug set from the payload

**Files:**
- Modify: `plugins/trellis/hooks/codex-context.mjs:14-32` (delete `SLUGS`/`SLUG_SET`), `:231-300` (`parseRulesToml`), `:360-420` (resolution order)
- Test: `cli/codex_hook_test.go`

**Interfaces:**
- Consumes: `reference/rules.md`, already loaded at `codex-context.mjs:412`.
- Produces: `parseRulesToml(source, slugSet)` — same return shape as today (a `Map<string, boolean>` or `null`), with the slug set now passed in rather than closed over.

- [ ] **Step 1: Write the failing test**

Append to `cli/codex_hook_test.go`:

```go
// The Codex hook validated rows against a hardcoded 16-slug array while the
// Claude hook derived its set from the shipped rules.md, and nothing in CI
// compared the two. A payload upgrade therefore could not repair drift on
// Codex — worse, a stale array made an `unknown:` reason FALSE: the agent
// would quarantine a live row and cite a payload that does ship it.
func TestCodexDerivesItsSlugSetFromThePayload(t *testing.T) {
	src, err := os.ReadFile("../plugins/trellis/hooks/codex-context.mjs")
	if err != nil {
		t.Fatal(err)
	}
	body := string(src)
	for _, slug := range assessableSlugs {
		if strings.Contains(body, `"`+slug+`"`) {
			t.Errorf("codex-context.mjs still hardcodes %s; the slug set must come from reference/rules.md", slug)
		}
	}
	if strings.Contains(body, "const SLUGS = [") {
		t.Error("the hardcoded SLUGS array must be gone — it cannot be repaired by a payload upgrade")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./... -run TestCodexDerivesItsSlugSetFromThePayload -v`
Expected: FAIL — sixteen errors plus the array assertion.

- [ ] **Step 3: Derive the slug set**

In `codex-context.mjs`, delete the `SLUGS` and `SLUG_SET` constants at `:14-32` and add a deriver beside the payload read:

```js
// The slugs the payload actually ships, read from the same rules.md the Claude
// hook validates against. A hardcoded list here could not be repaired by a
// plugin upgrade, and a stale one made a quarantine reason false.
function slugsFromRules(rulesMd) {
  const found = [];
  for (const line of rulesMd.split(/\r?\n/u)) {
    const m = line.match(/`((?:inv|floor)-[a-z-]+)`[ \t]*$/u);
    if (m) found.push(m[1]);
  }
  return found;
}
```

Change `parseRulesToml(source)` to `parseRulesToml(source, slugs)`, take `const slugSet = new Set(slugs)` as its first statement, and replace the three references: `!SLUG_SET.has(row[1])` → `!slugSet.has(row[1])`; `rows.size !== SLUGS.length` → `rows.size !== slugs.length`; `SLUGS.some(...)` → `slugs.some(...)`.

**The reordering the spec flagged as unconfirmed.** `parseRulesToml` is called at `:369` and the payload resolves at `:408`, so the payload read must move above the parse. Read `:360-420` in full first, move the payload resolution block ahead of the parse call, and verify no intervening statement depends on the parse result. **If any does, stop and report rather than restructuring further** — that is a design question, not an implementation detail.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./...`
Expected: PASS. `TestCodexDerivesItsSlugSetFromThePayload` goes green, and every existing Codex test (`writeValidCodexOverlay` at `cli/codex_hook_test.go:58-75` writes the generated `rules-a.toml` and asserts success) still passes — that fixture is the implicit pin that the derived set matches the generated one.

- [ ] **Step 5: Commit**

```bash
git add plugins/trellis/hooks/codex-context.mjs cli/codex_hook_test.go
git commit -m "fix(codex): derive the slug set from reference/rules.md

The hardcoded SLUGS array could not be repaired by a payload upgrade, and a
stale one made a quarantine reason false — the agent would quarantine a live
row and cite a payload that ships it. rules.md was already loaded here."
```

---

### Task 4: Record the decision and ship the payload

**Files:**
- Create: `decisions/0083-rules-toml-reconciles-itself.md`
- Modify: `decisions/0072-retire-the-setup-skill.md` (frontmatter forward pointer only)
- Modify: `plugins/trellis/VERSION`, `README.md:204-215`, `plugins/trellis/README.md`

**Interfaces:**
- Consumes: the shipped behaviour from Tasks 1–3.
- Produces: the merge-ready change. Nothing downstream.

- [ ] **Step 1: Write the decision record**

Create `decisions/0083-rules-toml-reconciles-itself.md`. Frontmatter carries **no `status` field** (`decision-0082`); `depends_on: [decision-0051, decision-0070, decision-0072, decision-0074]`; `owner: agent`; `date: 2026-08-30`. The body must cover, each with its evidence:

1. **What changed** — reconciliation replaces refusal; the three resolutions.
2. **Why quarantine and not deletion** — TRL-27's live case, where the remedy would have deleted `inv-no-orphan-followups`, a ratified row, to match a 0.5.0 plugin. Name that the two causes are indistinguishable at runtime in config-only mode.
3. **Why the gate is not being dropped** — quote `decision-0072`'s finding #6 (*"retiring a confirm-gated writer silently retires the gate"*) and state the distinction: that gate guarded destructive writes; quarantine loses nothing, so `floor-intent-gate` is satisfied without an act.
4. **What this supersedes** — `decision-0072`'s confirm-first row-repair remedy, and only that. Its retirement of `/trellis:setup` stands.
5. **The two retired test pins**, named, with why each is safe to retire.
6. **Consequences** — `decision-0070` D4 (*"the hook never writes"*) is untouched and still pinned.

- [ ] **Step 2: Add the forward pointer to decision-0072**

`decisions/` is append-only — add the pointer to frontmatter, change no prose:

```yaml
superseded_in_part_by: [decision-0083]  # 2026-08-30 — the confirm-first row-repair remedy only (D2's three shapes and the reseed gate). The retirement of /trellis:setup, D1 and D3-D5, all stand.
```

- [ ] **Step 3: Bump VERSION and sweep the prose**

`plugins/trellis/VERSION` → `0.7.0`. A payload change that does not move VERSION never reaches an installed copy (`cli/assets/invariants.md:65`).

In `README.md:204-215` and the plugin README, add the quarantine convention beside the existing row-editing instructions: a commented row prefixed `#` is a quarantined row, inert and safe to leave; if a newer Trellis ships that slug, update the plugin and uncomment.

- [ ] **Step 4: Run the full suite and the corpus reviewer**

Run: `cd cli && go test ./... && go build ./... && go vet ./...`
Expected: PASS. `TestDocsClaimOnlyRealCommands` and the payload guards are the ones most likely to catch a missed surface.

Then invoke the repo-owned `corpus-reviewer` subagent (`.claude/agents/`) against `decisions/`. It is read-only and reports; fix what it finds and re-run.

- [ ] **Step 5: Commit and open the PR**

```bash
git add decisions/ plugins/trellis/VERSION README.md plugins/trellis/README.md
git commit -m "decision-0083: rules.toml reconciles itself; the repair is reported, not gated"
git push -u origin feature/trl-20-rules-toml-self-repair
gh pr create --title "TRL-20: rules.toml reconciles itself instead of blacking out delivery" --body "$(cat <<'EOF'
Closes TRL-20, TRL-2, TRL-27 — three arms of one defect.

- **TRL-20** — a catalog-changing upgrade left every existing project one row short and blacked out every session. Missing rows now reconcile to `active = true`.
- **TRL-2** — the curl path claimed governance while a row was missing. Delivery and the row set can no longer disagree: the injected rows are the reconciled set.
- **TRL-27** — `unknown:` had two causes and one destructive remedy. Rows are now quarantined, never deleted, and the notice names the stale-plugin cause.

Supersedes `decision-0072`'s confirm-first row-repair remedy in part; its retirement of `/trellis:setup` stands. `decision-0070` D4 ("the hook never writes") is untouched.

Owed and filed separately: the producer-side row-set sweep, and the inline block's frozen row copy.
EOF
)"
```

The PR opens as a normal (non-draft) PR — the maintainer asked for this work. **Do not merge** — `floor-intent-gate`: the maintainer's review is the approval act (`decision-0080`).

---

## Still owed, and deliberately not in this plan

Both are named in the spec's out-of-scope section. **File the first as a Linear issue on team Trellis before closing this work** — `inv-no-orphan-followups` requires a named consumer, and this plan is not one:

- **The producer-side sweep.** `cli/assets/invariants.md:55-67` makes a row-set change a ~15-surface obligation enforced by prose. This is the root cause upstream of all three issues and has no tracking issue.
- **The inline block's frozen row copy** (`cli/apply.go:207`, `README.md:210`) — the one derivative that genuinely goes stale.

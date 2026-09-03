# Payload-Read Gateway Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every read of a Trellis *payload* file, in both SessionStart hooks, goes through one guard that turns absent / empty / unreadable / unusable into a loud refusal — so the twelfth instance of the defect class is caught by construction rather than by whoever next thinks to try it.

**Architecture:** `staleness.sh` gains one function, `payload_read`, which classifies a read as `ok | missing | unreadable | empty` and is the only place in the file that opens a payload file. Call sites branch on the classification, so the shortest thing to write is also the safe thing. `codex-context.mjs`'s `readRequired` — already the model for this — gains a third argument stating what an empty read means, making its default loud without changing any existing failure class. Two guards hold it: a behavioural matrix that enumerates payload files from the bundle directory itself, and a source scan that rejects a bypass.

**Tech Stack:** bash + awk (`staleness.sh`), Node ESM (`codex-context.mjs`), Go tests (`cli/`).

**Spec:** `docs/superpowers/specs/2026-09-03-payload-read-gateway-design.md`

## Global Constraints

- **The hooks never write `.trellis/rules.toml`** (`decision-0070` D4, pinned at `cli/plugin_hook_test.go`). They emit text; the agent writes.
- **Quarantine never deletes.** No row's value is lost on any path.
- **Every exit is zero** (`exit 0` / `process.exit(0)`). A hook must never fail the session.
- **`go test -count=1`, always.** The hooks are external files Go's test cache does not track; a cached PASS replays over a hook mutation. `AGENTS.md` records this.
- **No destructive verb in a new `emit "…"` literal** unless it also carries `explicit confirmation`. The scanned list is `delete, drop, remove, overwrite, replace, reset, discard, "rm "` — substring-matched, so `preset` hits `reset`. `TestEveryDestructiveInstructionIsGated` enforces it.
- **Emit literals are complete strings on their own line**, never prose assembled into a variable and spliced in: the guard's regex is `(?m)^\s*emit "((?:[^"\\]|\\.)*)"`.
- **Injection budget stays 32768 bytes.**
- **A payload change requires a `plugins/trellis/VERSION` bump** and an advance of `install.sh`'s baked bundle manifest. Keep the VERSION bump to one line — another change is queued behind this one.
- **Mutation-prove every guard**: break it, confirm the covering test goes red with the expected symptom, restore.

---

### Task 1: The gateway, and a source guard that no read bypasses it

**Files:**
- Modify: `plugins/trellis/hooks/staleness.sh` — insert after `emit()` (currently ends ~line 110)
- Test: `cli/plugin_hook_test.go` (new test appended)

**Interfaces:**
- Produces: `payload_read <abs-path>` — returns 0 only when the file is readable and non-empty. Sets `payload_status` (`ok|missing|unreadable|empty`), `payload_why` (a clause for messages), `payload_text` (content, only when `ok`). Tasks 2-5 consume all three.

- [ ] **Step 1: Write the failing source guard**

Append to `cli/plugin_hook_test.go`:

```go
// TestNoPayloadReadBypassesTheGateway is the structural half of the guard
// TRL-33 asks for. Eleven defects on this hook shared one shape — a payload
// file that was absent, empty, truncated or unreadable reached downstream
// logic and the session ran ungoverned at exit 0. Each was fixed where it
// was found; none was found by this suite. The behavioural half is
// TestBrokenPayloadIsNeverSilent. This half is what makes a read added LATER
// safe: a new "$plugin/reference/..." operand that never reaches payload_read
// fails here, at the point it is written.
func TestNoPayloadReadBypassesTheGateway(t *testing.T) {
	body, err := os.ReadFile("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	src := string(body)

	gateStart := strings.Index(src, "\npayload_read() {")
	if gateStart < 0 {
		t.Fatal("payload_read() not found in staleness.sh — the scan is broken, and a guard that reads nothing passes")
	}
	gateEnd := strings.Index(src[gateStart:], "\n}\n")
	if gateEnd < 0 {
		t.Fatal("payload_read() has no closing brace — the scan is broken")
	}
	gate := src[gateStart : gateStart+gateEnd]

	// Every shell variable assigned a path under the installed plugin.
	assignRe := regexp.MustCompile(`(?m)^\s*([a-z_]+)="\$plugin/`)
	vars := map[string]bool{}
	for _, m := range assignRe.FindAllStringSubmatch(src, -1) {
		vars[m[1]] = true
	}
	if len(vars) < 3 {
		t.Fatalf("found only %d payload path variables — the scan is broken", len(vars))
	}

	// Each must reach the gateway.
	for v := range vars {
		if !strings.Contains(src, `payload_read "$`+v+`"`) {
			t.Errorf("$%s holds a payload path but is never passed to payload_read — every payload read goes through the gateway", v)
		}
	}

	// And nothing may open one behind the gateway's back.
	readers := regexp.MustCompile(`(?m)^\s*(?:[A-Za-z_]+=)?\$?\(?\s*(cat|head|tail|sed|awk|grep|wc|tr|sort)\b[^\n]*"\$(plugin|` + strings.Join(sortedKeys(vars), "|") + `)[/"]`)
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(gate, line) {
			continue
		}
		if readers.MatchString(line) {
			t.Errorf("a payload path is opened outside payload_read:\n  %s", strings.TrimSpace(line))
		}
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `cd cli && go test -count=1 -run TestNoPayloadReadBypassesTheGateway ./...`
Expected: FAIL — `payload_read() not found`.

- [ ] **Step 3: Add the gateway to `staleness.sh`**

Insert immediately after `emit()`'s closing brace:

```sh
# ======================================================== the payload gateway
# THE ONE PLACE THIS FILE OPENS A TRELLIS PAYLOAD FILE.
#
# Eleven defects across decision-0083 and decision-0084 shared one shape: an
# absent, empty, truncated or unreadable payload input reached downstream
# logic, and the session ran ungoverned at exit 0 with nothing signalling a
# problem. Each was fixed where it was found. Almost none was found by the
# suite or by reading the code -- every one was found by a reviewer RUNNING
# the hook against a deliberately broken input. That recurrence is the
# finding, and this function is the answer to it: a read added to this file
# later is guarded by where it is written, not by remembering. The guards that
# hold it are TestNoPayloadReadBypassesTheGateway (nothing opens a payload
# path behind its back) and TestBrokenPayloadIsNeverSilent (no break of any
# shipped payload file produces silence).
#
# It CLASSIFIES; it does not judge. Two of those eleven were the inverse
# defect -- a guard that refused a HEALTHY payload (a CRLF-terminated
# rules.md; an unreadable comparison preset reported as payload incoherence)
# -- and a consumer who sees TRELLIS_RULES_NOT_LOADED with nothing wrong to
# fix is as badly served as one governed by a broken payload. So the four
# outcomes are reported and the CALL SITE decides what each one costs.
#
#   missing     the path is not there, or its symlink target is gone
#   unreadable  it exists and could not be opened -- a permission mode, a
#               stale ACL, or a directory where a file must be
#   empty       it opened and yielded nothing
#   ok          content is in $payload_text
#
# missing and unreadable are told apart because their remedies differ
# (reinstall vs. fix the mode) -- but NEITHER is silent anywhere, which is
# TRL-33's whole finding: the two were handled differently for no reason.
#
# Returns 0 only for ok, so the shortest thing a caller can write --
# `payload_read "$f" || { emit "..."; exit 0; }` -- is also the safe thing.
# The result comes back in globals rather than on stdout ON PURPOSE: a
# `$(payload_read ...)` capture runs the function in a SUBSHELL, and every
# status it set would be discarded at the closing paren.
payload_status=""
payload_why=""
payload_text=""
payload_read() {
  payload_text=""
  if [ ! -e "$1" ]; then
    payload_status=missing
    payload_why="is missing"
    return 1
  fi
  if [ ! -f "$1" ]; then
    payload_status=unreadable
    payload_why="is not a readable file (a directory or a device sits at that path)"
    return 1
  fi
  # `-f` proves a file EXISTS, never that it can be READ, and the gap between
  # those two is where several of the eleven lived. Opening it is the only
  # test that settles it.
  if ! payload_text="$(cat "$1" 2>/dev/null)"; then
    payload_text=""
    payload_status=unreadable
    payload_why="exists but could not be read (a permission mode, a stale ACL, or a symlink whose target is gone)"
    return 1
  fi
  # Command substitution strips trailing newlines, so a file of nothing but
  # blank lines lands here too -- which is right: it is as unusable as a
  # zero-byte one, and both were shipped as separate defects.
  if [ -z "$payload_text" ]; then
    payload_status=empty
    payload_why="is empty"
    return 1
  fi
  payload_status=ok
  payload_why=""
  return 0
}
```

- [ ] **Step 4: Run the guard**

Run: `cd cli && go test -count=1 -run TestNoPayloadReadBypassesTheGateway ./...`
Expected: still FAIL — the existing reads have not been routed yet. That is Tasks 2-5.

- [ ] **Step 5: Commit**

```bash
git add plugins/trellis/hooks/staleness.sh cli/plugin_hook_test.go
git commit -m "TRL-33: add the payload gateway and the guard that no read bypasses it"
```

---

### Task 2: TRL-33 — a missing defaults preset refuses loudly

**Files:**
- Modify: `plugins/trellis/hooks/staleness.sh` — the `scoped_to_project = yes` arm (`toml="$plugin/reference/rules-b.toml"` / `[ -f "$toml" ] || exit 0`)
- Test: `cli/plugin_hook_test.go` — extend `TestEveryLegitimateShapeStillGoverns`'s corrupted-defaults states

**Interfaces:**
- Consumes: `payload_read`, `payload_why` from Task 1.

- [ ] **Step 1: Add the failing case**

In `TestEveryLegitimateShapeStillGoverns`, add `"absent"` to the corrupted-defaults state list and give it an `os.Remove` arm beside the existing `empty` / `no rows` / `unreadable` states. The existing assertions are unchanged.

- [ ] **Step 2: Run it and watch it fail**

Run: `cd cli && go test -count=1 -run TestEveryLegitimateShapeStillGoverns ./...`
Expected: FAIL — the `absent` case produces empty output, so `nudgeContext` fatals on "empty output" or the marker assertion fires.

- [ ] **Step 3: Route the read through the gateway**

Replace `[ -f "$toml" ] || exit 0` with a loud refusal reusing the wording of the `default_rows == 0` refusal further down (which is already tested), gaining `$payload_why`:

```sh
    toml="$plugin/reference/rules-b.toml"
    # TRL-33. This was `[ -f "$toml" ] || exit 0`: an absent preset produced
    # completely empty stdout at exit 0 and a session governed by nothing,
    # with no signal of any kind -- while the UNREADABLE sibling of the same
    # file was caught loudly two hundred lines downstream, by a message that
    # named .trellis/rules.toml to a project that has none. Absent-vs-
    # unreadable was an inconsistency, not a choice.
    if ! payload_read "$toml"; then
      emit "TRELLIS_RULES_NOT_LOADED — this project has no .trellis/rules.toml and is governed by the rule rows the Trellis plugin ships ($toml), and that file $payload_why. This project adopted Trellis (the plugin is vendored in this repository), but the session is running ungoverned and NO rules and NO rows were injected. The hook refused rather than treat a broken payload file as if it were this project's own settings. NOTHING is wrong with this project and there is nothing here to correct: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the fix. Tell the user before doing substantive work."
      exit 0
    fi
    rows_are_default=yes
```

- [ ] **Step 4: Run the case, then the whole suite**

Run: `cd cli && go test -count=1 -run TestEveryLegitimateShapeStillGoverns ./...` → PASS
Run: `cd cli && go test -count=1 ./...` → PASS

- [ ] **Step 5: Mutation-prove it**

Change `if ! payload_read "$toml"; then` back to `[ -f "$toml" ] || exit 0`; run the test; confirm it fails on the `absent` case; restore.

- [ ] **Step 6: Commit**

```bash
git add plugins/trellis/hooks/staleness.sh cli/plugin_hook_test.go
git commit -m "TRL-33: a missing vendored-defaults preset refuses loudly instead of exiting silently"
```

---

### Task 3: TRL-34 — an unusable `reference/version` says why, and stays governed

**Files:**
- Modify: `plugins/trellis/hooks/staleness.sh` — the `current=` read; path A's and the legacy path's `[ -n "$current" ] || exit 0`; path C's staleness compare; the path-B trailer and quarantine stamp
- Test: `cli/plugin_hook_test.go` — invert `TestStalenessHook`/"unreadable plugin reference is silent"; add the malformed-stamp cases

**Interfaces:**
- Consumes: `payload_read`, `payload_why`, `payload_text`.
- Produces: `$current` (the validated stamp, or `""`) and `$stamp_defect` (the reason, or `""`). Tasks 4 and 5 read both.

- [ ] **Step 1: Write the failing tests**

Rewrite the `"unreadable plugin reference is silent"` subtest as `"an unreadable plugin stamp says so instead of vanishing"`, asserting `TRELLIS_STALENESS_UNKNOWN` is present and `TRELLIS_RULES_NOT_LOADED` is **absent** (the session is still governed — reusing the blackout marker would be the over-correction). Add a table over `{absent, empty, mode-000, "payload@short", "plugin@abcdef123456"}`.

- [ ] **Step 2: Run and watch it fail**

Run: `cd cli && go test -count=1 -run TestStalenessHook ./...`
Expected: FAIL — output is empty.

- [ ] **Step 3: Implement**

Replace the two-line `current=` read with a gateway read plus a shape check matching `codex-context.mjs`'s (`^payload@[0-9a-f]{12}$`), and give each of the three consumers a disposition. Path A and the legacy path emit `TRELLIS_STALENESS_UNKNOWN`; path C's stand-down gains a variant literal; path B substitutes a stable phrase for the empty stamp in the quarantine note and states the defect in the trailer.

- [ ] **Step 4: Run**

Run: `cd cli && go test -count=1 ./...` → PASS

- [ ] **Step 5: Mutation-prove it**

Restore `[ -n "$current" ] || exit 0` on path A; confirm the new subtest fails; restore.

- [ ] **Step 6: Commit**

```bash
git add plugins/trellis/hooks/staleness.sh cli/plugin_hook_test.go
git commit -m "TRL-34: an unreadable plugin version stamp says why staleness could not be checked"
```

---

### Task 4: The vendored overlay is payload too

**Files:**
- Modify: `plugins/trellis/hooks/staleness.sh` — path A's `[ ! -f "$ver" ]`, the `for f in trellis.md rules.md` loop, `overlay=`, and the legacy flat branch
- Test: `cli/plugin_hook_test.go` — split `"empty stamp is silent"`; add the mode-000 overlay-file case

- [ ] **Step 1: Write the failing tests** — an empty `.trellis/internal/version` and a mode-000 `.trellis/internal/rules.md` must each draw `TRELLIS_RULES_NOT_LOADED` naming the file; the legacy flat empty stamp must still draw its migration nudge.
- [ ] **Step 2: Run and watch them fail.**
- [ ] **Step 3:** Route all three overlay files through `payload_read` in one loop, replacing the `-f`/`-s` pair.
- [ ] **Step 4:** `cd cli && go test -count=1 ./...` → PASS
- [ ] **Step 5: Mutation-prove** — restore `[ ! -s "$internal/$f" ]`; confirm the mode-000 case fails; restore.
- [ ] **Step 6: Commit.**

---

### Task 5: The remaining plugin-side reads

**Files:**
- Modify: `plugins/trellis/hooks/staleness.sh` — `$rules` and `$header` existence checks and the `header_prose` read; `$preset`'s `[ -f ]`

- [ ] **Step 1:** Route `$rules`, `$header` and `$preset` through `payload_read`. `$preset` keeps its **silent skip** on failure — unchanged, and pinned by `TestAnUnusablePresetSkipsTheCoherenceCheckRatherThanBlackingOut`, which exists because this check once over-corrected.
- [ ] **Step 2:** `cd cli && go test -count=1 ./...` → PASS
- [ ] **Step 3:** `go test -count=1 -run TestNoPayloadReadBypassesTheGateway ./...` → PASS (Task 1's guard now has nothing to complain about).
- [ ] **Step 4: Commit.**

---

### Task 6: The behavioural matrix — no break of any payload file is silent

**Files:**
- Test: `cli/plugin_hook_test.go` — new `TestBrokenPayloadIsNeverSilent`

- [ ] **Step 1:** Enumerate the payload files by reading `plugins/trellis/reference/` with `os.ReadDir` — **never a hardcoded list**, so a payload file added later joins the matrix without anyone remembering.
- [ ] **Step 2:** Cross with break shapes `{healthy, crlf, absent, zero-byte, mode-000, truncated-to-half}` and project shapes `{config-only, vendored-defaults}` — the two paths on which the hook *delivers*, so "silence is always wrong" holds unconditionally.
- [ ] **Step 3:** Assert, for every cell: exit 0; stdout non-empty; and **either** a complete governed injection (the `<!-- trellis:rules-loaded -->` sentinel **and** at least one `deliveredRow`) **or** a message carrying a `TRELLIS_` marker — never a marker-free partial, which is the shape that ships an activation list with no rules under it.
- [ ] **Step 4:** Assert the `healthy` and `crlf` rows produce a complete injection with **no** marker. This is the discrimination proof, on exactly the shape that caused one of the two over-corrections.
- [ ] **Step 5:** Guard the premise — skip mode-000 rows under `os.Geteuid() == 0`, and re-read the file after chmod to confirm it is genuinely unreadable, matching the four existing skip-guards in this file.
- [ ] **Step 6: Mutation-prove the matrix** — reintroduce `[ -f "$toml" ] || exit 0`; confirm the matrix reports the silent cell; restore.
- [ ] **Step 7: Commit.**

---

### Task 7: Codex — make the default loud without changing a single failure class

**Files:**
- Modify: `plugins/trellis/hooks/codex-context.mjs` — `readRequired` signature; its four call sites; the two hardcoded `fail` labels
- Test: `cli/codex_hook_test.go` — new `TestEveryReadRequiredStatesWhatEmptyMeans`

- [ ] **Step 1:** Write the source guard: every `readRequired(` call site passes a third argument.
- [ ] **Step 2:** Run it and watch it fail.
- [ ] **Step 3:** Add `options` to `readRequired`; return `{ error: options.emptyError ?? "empty-file" }` on a zero-length read unless `options.emptyIsValid` is set. Pass `{ emptyError: "empty-prose" }` for prose and rules, `{ emptyError: "invalid-version" }` for version, `{ emptyIsValid: true }` for `.trellis/rules.toml` with a comment saying why (empty is the supported hand-written-partial shape). **Every existing class is preserved byte-for-byte** — `TestCodexHookFailureVocabularyAndIsolation` asserts exact stdout and must not need an edit.
- [ ] **Step 4:** Change `fail(".trellis/internal/trellis.md", …)` → `fail(sources.prose, …)` and `fail(".trellis/internal/rules.md", "invalid-rules")` → `fail(sources.rules, …)`. On the vendored fixture these resolve to the same strings the tests already assert; on the plugin-native path they stop naming a file that was never read.
- [ ] **Step 5:** `cd cli && go test -count=1 ./...` → PASS, **with no edit to any existing Codex assertion**. An assertion that needs editing is a behaviour change to justify or back out.
- [ ] **Step 6: Mutation-prove** — drop the third argument from the version call site; confirm the new guard fails; restore.
- [ ] **Step 7: Commit.**

---

### Task 8: Ship it — version, manifest, record

**Files:**
- Modify: `plugins/trellis/VERSION` (one line), `install.sh` (bundle manifest)
- Create: `decisions/0086-one-gateway-for-every-payload-read.md`

- [ ] **Step 1:** Bump `plugins/trellis/VERSION` — a payload change no cached consumer would otherwise re-pull. **One line only**, so the queued change behind this one rebases cleanly.
- [ ] **Step 2:** Advance the baked manifest hashes in `install.sh` for `hooks/staleness.sh`, `hooks/codex-context.mjs` and `VERSION`. `TestInstallScriptBundleManifestIsCurrent` names each stale line.
- [ ] **Step 3:** Write `decisions/0086`. Verify `0086` is free by listing `decisions/`. No `status:` field (`decision-0082`). Record: the defect class and its count; the gateway; why `missing` and `unreadable` are distinguished but neither is silent; the new `TRELLIS_STALENESS_UNKNOWN` marker as a contract change; the two inverted tests; and as an **open question** the Claude/Codex divergence on a malformed `reference/version` (Claude governs and annotates, Codex refuses) — named, not closed.
- [ ] **Step 4:** `cd cli && go test -count=1 ./... && go build ./... && go vet ./...` → all green.
- [ ] **Step 5:** Invoke the repo-owned `corpus-reviewer` agent against the new decision.
- [ ] **Step 6:** Rebase on `origin/main` (PR #260 touches `install.sh`'s manifest — expect a conflict there and rebase rather than fight it). Commit, push, open the PR. **Do not merge** (`floor-intent-gate`).

  > **2026-09-03, after execution — the record shipped as `decision-0087`, not `0086`.** Step 3's
  > instruction was followed exactly as written: `0085` was the highest on `main` when this branch
  > opened, so `0086` was free by the only test available to a branch. It was not free in the merge
  > queue — `kodhama/trellis#263` (TRL-29) also claimed `0086` and lands first — so this branch
  > renumbered before merge. **Decision ids are allocated by whichever branch merges first, and a
  > branch cut before the race cannot see it** (`decision-0078` recorded the same mechanism and
  > armed a trigger to file it on a third recurrence; this is the third, filed as
  > [TRL-40](https://linear.app/kodhama/issue/TRL-40)).
  >
  > **Appended rather than edited in place**, per `decision-0085` point 5 — *"the correction is a
  > dated note appended beside the original claim, never an edit that makes the record look
  > prescient."* An earlier version of this renumber did edit both lines in place, which made the
  > plan read as though it had instructed verifying `0087`; a corpus review caught it and the
  > original wording is restored above.

---

## Self-review

- **Spec coverage.** S1 → Task 3. S2 → Task 2. S3, S4, S5 → Task 5. S6, S7, S8 → Task 4. C1 is already loud and needs no change (stated in the spec). C2, C3 → Task 7. The two guards → Tasks 1 and 6. Ship mechanics → Task 8. No spec section is unclaimed.
- **Type consistency.** `payload_read` / `payload_status` / `payload_why` / `payload_text` are used under those exact names in Tasks 2-5; `$stamp_defect` is introduced in Task 3 and consumed only there.
- **Ordering.** Task 1's guard is expected to stay red until Task 5 completes; that is stated in Task 1 Step 4 rather than left as a surprise.

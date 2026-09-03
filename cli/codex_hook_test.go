package main

// Production-contract tests for the Codex live-rule delivery contract
// (`decision-0058`). The `spec-0007@v1 R##` / `S##` markers below are the
// requirement ids of the spec that recorded it; `decision-0079` retired the spec
// stage and deleted `specs/`, so these markers now name requirements whose only
// surviving statement is the tests themselves (the retired text remains in git
// history). Kept rather than stripped: the mapping is the information.
//
// The captured input shape is
// normalized from Codex's SessionStart request contract and decision-0058's live
// local positive control; volatile session/model fields are intentionally omitted
// because the handler contract consumes only hook_event_name, source, and cwd.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

const rulesLoadedSentinel = "<!-- trellis:rules-loaded -->"

type codexHookResult struct {
	HookSpecificOutput *struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput,omitempty"`
	SystemMessage string `json:"systemMessage,omitempty"`
}

func codexHookPath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs("../plugins/trellis/hooks/codex-context.mjs")
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func writeCodexPluginRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".codex-plugin", "plugin.json"), []byte(`{"name":"trellis"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func writeValidCodexOverlay(t *testing.T, project string) {
	t.Helper()
	files := payloadFiles()
	for rel, content := range map[string]string{
		".trellis/internal/trellis.md": files["trellis-a.md"],
		".trellis/internal/rules.md":   files["rules.md"],
		".trellis/internal/version":    files["version"],
		".trellis/rules.toml":          files["rules-a.toml"],
	} {
		path := filepath.Join(project, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func runCodexHook(t *testing.T, pluginRoot, stdin string) (string, codexHookResult) {
	t.Helper()
	cmd := exec.Command("node", codexHookPath(t))
	cmd.Env = append(os.Environ(), "PLUGIN_ROOT="+pluginRoot)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("hook exited non-zero: %v\nstderr: %s\nstdout: %s", err, stderr.String(), stdout.String())
	}
	raw := strings.TrimSpace(stdout.String())
	if raw == "" {
		return "", codexHookResult{}
	}
	var got codexHookResult
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("hook stdout is not one JSON object: %v\n%s", err, raw)
	}
	return raw, got
}

func startupInput(t *testing.T, cwd string) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"hook_event_name": "SessionStart",
		"source":          "startup",
		"cwd":             cwd,
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func setRuleActive(t *testing.T, source, slug string, active bool) string {
	t.Helper()
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), slug+" ") {
			start := strings.Index(line, "{ active = ")
			if start < 0 {
				t.Fatalf("row %s has unexpected shape: %q", slug, line)
			}
			end := strings.Index(line[start:], " }")
			if end < 0 {
				t.Fatalf("row %s has unexpected shape: %q", slug, line)
			}
			value := "true"
			if !active {
				value = "false"
			}
			lines[i] = line[:start] + "{ active = " + value + line[start+end:]
			return strings.Join(lines, "\n")
		}
	}
	t.Fatalf("row %s not found", slug)
	return ""
}

func newGitProject(t *testing.T) string {
	t.Helper()
	project := t.TempDir()
	if err := os.Mkdir(filepath.Join(project, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return project
}

// guards spec-0007@v1 R4, R5, R26, S9, S17
func TestCodexHookRegistrationIsStartupOnlyAndHostIsolated(t *testing.T) {
	codexManifest := readFileT(t, "../plugins/trellis/.codex-plugin/plugin.json")
	if !strings.Contains(codexManifest, `"hooks": "./hooks/codex-hooks.json"`) {
		t.Error("Codex plugin manifest must point at ./hooks/codex-hooks.json")
	}
	var registration struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	raw := readFileT(t, "../plugins/trellis/hooks/codex-hooks.json")
	if err := json.Unmarshal([]byte(raw), &registration); err != nil {
		t.Fatalf("parse Codex hook registration: %v", err)
	}
	if len(registration.Hooks) != 1 || len(registration.Hooks["SessionStart"]) != 1 {
		t.Fatalf("Codex registration must contain only one SessionStart group: %#v", registration.Hooks)
	}
	group := registration.Hooks["SessionStart"][0]
	if group.Matcher != "startup" {
		t.Errorf("Codex matcher must be exactly startup, got %q", group.Matcher)
	}
	if len(group.Hooks) != 1 || group.Hooks[0].Type != "command" ||
		group.Hooks[0].Command != `node "${PLUGIN_ROOT}/hooks/codex-context.mjs"` {
		t.Errorf("unexpected Codex hook command: %#v", group.Hooks)
	}

	claude := readFileT(t, "../plugins/trellis/hooks/hooks.json")
	if strings.Contains(claude, "codex-context") || strings.Contains(claude, "hookSpecificOutput") {
		t.Error("Claude hook registration must not contain Codex transport")
	}
}

// guards spec-0007@v1 R7, R8, R36, S6, S19
func TestCodexHookBoundsAuthoritativeFileReads(t *testing.T) {
	source := readFileT(t, "../plugins/trellis/hooks/codex-context.mjs")
	for _, required := range []string{
		"fs.readSync",
		"MAX_CONTEXT_BYTES + 1",
		"stat.size > MAX_CONTEXT_BYTES",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("Codex hook bounded-read implementation missing %q", required)
		}
	}
	if strings.Contains(source, "fs.readFileSync(absolute") {
		t.Error("Codex hook must not read an authoritative file wholly before enforcing its byte bound")
	}
}

// guards spec-0007@v1 R1, R2, R7, R10, R31, R34-R36, S1, S2, S19
func TestCodexHookValidStartupAndLiveRows(t *testing.T) {
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	pluginRoot := writeCodexPluginRoot(t)
	raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil || got.SystemMessage != "" {
		t.Fatalf("valid startup must emit hookSpecificOutput only: %s", raw)
	}
	if got.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Errorf("wrong hookEventName: %q", got.HookSpecificOutput.HookEventName)
	}
	context := got.HookSpecificOutput.AdditionalContext
	if len([]byte(context)) > 8000 {
		t.Fatalf("additionalContext is %d bytes, exceeds 8000", len([]byte(context)))
	}
	for _, unique := range []string{
		rulesLoadedSentinel,
		strings.TrimSpace(payloadFiles()["version"]),
		`strictness  = "firm"`,
	} {
		if n := strings.Count(context, unique); n != 1 {
			t.Errorf("assembled context must contain %q once, got %d", unique, n)
		}
	}
	if !strings.Contains(context, rulesLoadedSentinel+"\n\n---\n"+invariantsTrigger) {
		t.Error("assembled context must expose exactly the generated sentinel-plus-fixed-footer boundary")
	}
	if strings.Contains(context, "../plugins/trellis/reference") {
		t.Error("assembled context must not source rule content from the plugin payload")
	}

	rulesPath := filepath.Join(project, ".trellis", "rules.toml")
	rows := readFileT(t, rulesPath)
	rows = setRuleActive(t, rows, "inv-handover-points", false)
	if err := os.WriteFile(rulesPath, []byte(rows), 0o644); err != nil {
		t.Fatal(err)
	}
	_, edited := runCodexHook(t, pluginRoot, startupInput(t, project))
	if edited.HookSpecificOutput == nil ||
		!strings.Contains(edited.HookSpecificOutput.AdditionalContext, "inv-handover-points       = { active = false }") ||
		strings.Contains(edited.HookSpecificOutput.AdditionalContext, "inv-handover-points       = { active = true }") {
		t.Error("next startup must read the consumer's edited row without refresh")
	}
}

// guards spec-0007@v1 R6, R8, R9, R31, R33-R36, R41, S6, S8, S19
func TestCodexHookFailureVocabularyAndIsolation(t *testing.T) {
	pluginRoot := writeCodexPluginRoot(t)
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)

	assertFailure := func(t *testing.T, stdin, label, class string) {
		t.Helper()
		raw, got := runCodexHook(t, pluginRoot, stdin)
		want := fmt.Sprintf(`{"systemMessage":"Trellis hook did not load rules: %s: %s. The AGENTS.md bootstrap must attempt the installed overlay."}`, label, class)
		if raw != want {
			t.Errorf("failure mismatch\n got: %s\nwant: %s", raw, want)
		}
		if got.HookSpecificOutput != nil {
			t.Error("failure must not emit hookSpecificOutput")
		}
	}

	assertFailure(t, `{`, "stdin", "invalid-json")
	assertFailure(t, `{"hook_event_name":"Stop","source":"startup","cwd":"`+project+`"}`, "hook_event_name", "wrong-event")
	assertFailure(t, `{"hook_event_name":"SessionStart","source":"resume","cwd":"`+project+`"}`, "source", "wrong-event")
	assertFailure(t, `{"hook_event_name":"SessionStart","source":"startup","cwd":"relative"}`, "cwd", "invalid-cwd")
	assertFailure(t, startupInput(t, t.TempDir()), "project-root", "project-root-not-found")

	mutateAndFail := func(t *testing.T, rel, content, class string) {
		t.Helper()
		path := filepath.Join(project, filepath.FromSlash(rel))
		original := readFileT(t, path)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		assertFailure(t, startupInput(t, project), rel, class)
		if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mutateAndFail(t, ".trellis/internal/trellis.md", "", "empty-prose")
	mutateAndFail(t, ".trellis/internal/trellis.md", "no import\n", "invalid-placeholder-count")
	mutateAndFail(t, ".trellis/internal/trellis.md", "@rules.md\n@rules.md\n", "invalid-placeholder-count")
	mutateAndFail(t, ".trellis/internal/rules.md", "", "empty-prose")
	mutateAndFail(t, ".trellis/internal/rules.md", "no sentinel\n", "invalid-rules")
	mutateAndFail(t, ".trellis/internal/rules.md", rulesLoadedSentinel+"\nnot terminal\n", "invalid-rules")
	mutateAndFail(t, ".trellis/internal/rules.md", rulesLoadedSentinel+"\n"+rulesLoadedSentinel+"\n", "invalid-rules")
	for _, invalid := range []string{
		"",
		"payload@abcdef12345\n",
		"payload@abcdef123456\n\n",
		"payload@ABCDEF123456\n",
		"plugin@abcdef123456\n",
	} {
		mutateAndFail(t, ".trellis/internal/version", invalid, "invalid-version")
	}
	mutateAndFail(t, ".trellis/internal/version", "payload@ABCDEF123456\n", "invalid-version")
	mutateAndFail(t, ".trellis/rules.toml", "strictness = \"loose\"\n[rules]\n", "invalid-rules")
	// A renamed/missing/duplicate slug used to fail this file closed too (three
	// deleted assertions: rename, missing row, duplicate row). TRL-20 moved that
	// slug-set mismatch off the fail-closed path entirely — it now reconciles
	// instead of refusing, exactly as staleness.sh already did. See
	// TestCodexReconcilesInsteadOfFailingClosed for the reconciled behaviour
	// these three cases now exercise; only a genuine syntax fault (this test's
	// remaining assertions) still lands here.

	largeRules := strings.TrimSuffix(payloadFiles()["rules.md"], rulesLoadedSentinel+"\n") +
		strings.Repeat("é", 8001) + "\n" + rulesLoadedSentinel + "\n"
	rulesPath := filepath.Join(project, ".trellis", "internal", "rules.md")
	originalRules := readFileT(t, rulesPath)
	if err := os.WriteFile(rulesPath, []byte(largeRules), 0o644); err != nil {
		t.Fatal(err)
	}
	assertFailure(t, startupInput(t, project), "assembled-context", "context-over-budget")
	if err := os.WriteFile(rulesPath, []byte(originalRules), 0o644); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(project, ".trellis", "rules.toml")
	originalConfig := readFileT(t, configPath)
	oversizedConfig := originalConfig + "#" + strings.Repeat("x", 8001) + "\n"
	if err := os.WriteFile(configPath, []byte(oversizedConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	assertFailure(t, startupInput(t, project), "assembled-context", "context-over-budget")
	if err := os.WriteFile(configPath, []byte(originalConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	missing := filepath.Join(project, ".trellis", "internal", "rules.md")
	original := readFileT(t, missing)
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(missing, 0o755); err != nil {
		t.Fatal(err)
	}
	assertFailure(t, startupInput(t, project), ".trellis/internal/rules.md", "unreadable-file")
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(missing, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}
	assertFailure(t, startupInput(t, project), ".trellis/internal/rules.md", "missing-file")
	if err := os.WriteFile(missing, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{".trellis/internal/trellis.md", ".trellis/internal/version"} {
		path := filepath.Join(project, filepath.FromSlash(rel))
		original := readFileT(t, path)
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
		assertFailure(t, startupInput(t, project), rel, "missing-file")
		if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	nested := filepath.Join(project, "nested")
	if err := os.MkdirAll(filepath.Join(nested, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	assertFailure(t, startupInput(t, nested), "project-root", "project-root-not-found")

	missingCwd := filepath.Join(project, "does-not-exist")
	assertFailure(t, startupInput(t, missingCwd), "cwd", "invalid-cwd")
}

// guards spec-0007@v1 R32, S20
func TestCodexHookRejectsInvalidPluginRootWithoutFallback(t *testing.T) {
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	wrong := writeCodexPluginRoot(t)
	if err := os.WriteFile(filepath.Join(wrong, ".codex-plugin", "plugin.json"), []byte(`{"name":"other"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", codexHookPath(t))
	cmd.Env = append(os.Environ(), "PLUGIN_ROOT="+wrong, "CLAUDE_PLUGIN_ROOT="+writeCodexPluginRoot(t))
	cmd.Stdin = strings.NewReader(startupInput(t, project))
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	want := `{"systemMessage":"Trellis hook did not load rules: PLUGIN_ROOT: invalid-plugin-root. The AGENTS.md bootstrap must attempt the installed overlay."}` + "\n"
	if string(out) != want {
		t.Errorf("invalid Codex root must not fall through to Claude root\n got: %q\nwant: %q", out, want)
	}
}

// guards spec-0007@v1 R37, S22
func TestCodexHookFalseFloorRowsWarnButSucceed(t *testing.T) {
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	rulesPath := filepath.Join(project, ".trellis", "rules.toml")
	rows := readFileT(t, rulesPath)
	rows = setRuleActive(t, rows, "floor-intent-gate", false)
	rows = setRuleActive(t, rows, "floor-transparency", false)
	if err := os.WriteFile(rulesPath, []byte(rows), 0o644); err != nil {
		t.Fatal(err)
	}
	_, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatal("false floor rows must retain successful context delivery")
	}
	want := "Trellis warning: floor rows set active = false are overridden-by-floor and remain active: floor-intent-gate, floor-transparency."
	if got.SystemMessage != want {
		t.Errorf("floor warning mismatch\n got: %q\nwant: %q", got.SystemMessage, want)
	}
}

// guards spec-0007@v1 R7, R8, R20, R31, S1, S6, S12
func TestCodexHookStrictRulesTomlSchema(t *testing.T) {
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	pluginRoot := writeCodexPluginRoot(t)
	rulesPath := filepath.Join(project, ".trellis", "rules.toml")
	canonical := readFileT(t, rulesPath)

	literalStrings := strings.Replace(canonical,
		`seeded_from = "conductor"`, `seeded_from = 'conductor'`, 1)
	literalStrings = strings.Replace(literalStrings,
		`strictness  = "firm"`, `strictness  = 'firm'`, 1)
	if err := os.WriteFile(rulesPath, []byte(literalStrings), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("valid TOML literal strings must be accepted: %s", raw)
	}

	basicUnicode := strings.Replace(canonical,
		`strictness  = "firm"`, `strictness  = "\U00000066irm"`, 1)
	if err := os.WriteFile(rulesPath, []byte(basicUnicode), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, got = runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("valid TOML \\U basic-string escape must be accepted: %s", raw)
	}

	tabWhitespace := strings.Replace(canonical,
		`strictness  = "firm"`, "strictness\t=\t\"firm\"", 1)
	if err := os.WriteFile(rulesPath, []byte(tabWhitespace), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, got = runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("valid TOML space/tab whitespace must be accepted: %s", raw)
	}

	assertInvalid := func(t *testing.T, source string) {
		t.Helper()
		if err := os.WriteFile(rulesPath, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
		want := `{"systemMessage":"Trellis hook did not load rules: .trellis/rules.toml: invalid-rules. The AGENTS.md bootstrap must attempt the installed overlay."}`
		if raw != want || got.HookSpecificOutput != nil {
			t.Errorf("malformed or duplicate top-level Trellis TOML must fail exactly\n got: %s\nwant: %s", raw, want)
		}
	}

	beforeRules := strings.Replace(canonical, "[rules]", "seeded_from = 'duplicate'\n\n[rules]", 1)
	assertInvalid(t, beforeRules)
	beforeRules = strings.Replace(canonical, "[rules]", "strictness = 'adaptive'\n\n[rules]", 1)
	assertInvalid(t, beforeRules)
	assertInvalid(t, canonical+"\n[rules]\n")
	assertInvalid(t, canonical+"\n[other]\n")
	assertInvalid(t, strings.Replace(canonical, "[rules]", "unexpected = 'value'\n\n[rules]", 1))
	// A duplicate ROW (as opposed to the duplicate top-level keys/sections
	// above, which stay fatal) used to fail closed here too; TRL-20 moved it
	// onto the reconcile path instead — see
	// TestCodexReconcilesInsteadOfFailingClosed's "a duplicate keeps the first
	// occurrence" case.
	assertInvalid(t, strings.Replace(canonical,
		`seeded_from = "conductor"`, `seeded_from = "\/"`, 1))
	assertInvalid(t, strings.Replace(canonical,
		`strictness  = "firm"`, "strictness\u00a0= \"firm\"", 1))
	for _, invalidValue := range []string{
		`seeded_from = "\x41"`,
		`seeded_from = "\uD800"`,
		`seeded_from = "\U00110000"`,
		"seeded_from = \"bad" + string(rune(1)) + "\"",
	} {
		assertInvalid(t, strings.Replace(canonical,
			`seeded_from = "conductor"`, invalidValue, 1))
	}
}

// stripTOMLLine deletes an entire top-level assignment line (its trailing
// inline comment included) from a rendered rules.toml fixture, matching-and-
// removing rather than requiring the caller to spell out the payload's exact
// comment text (which carries a literal "·" the generator emits, easy to
// transcribe wrong). Used below to build a "no strictness line at all"
// fixture from the real firm/adaptive presets.
func stripTOMLLine(t *testing.T, source, key string) string {
	t.Helper()
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `[ \t]*=.*\n`)
	out := re.ReplaceAllString(source, "")
	if out == source {
		t.Fatalf("fixture removed nothing — %q was not found as a top-level line", key)
	}
	return out
}

// Codex used to fail closed on ANY mismatch, so a single bad row cost all
// sixteen rules — the TRL-20 blackout, still live on this host after the Claude
// side was fixed. It now reconciles, exactly as staleness.sh does.
func TestCodexReconcilesInsteadOfFailingClosed(t *testing.T) {
	pluginRoot := writeCodexPluginRoot(t)
	files := payloadFiles()

	run := func(t *testing.T, toml string) (string, codexHookResult) {
		t.Helper()
		project := newGitProject(t)
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

	// Fix round 1, minor: the mutation this covered used to live as a
	// fail-closed assertion in TestCodexHookFailureVocabularyAndIsolation
	// (deleted there in the same round, since a rename is exactly the
	// missing+unknown mismatch this task makes reconcilable) — nothing
	// otherwise pinned it after that deletion, though it already worked.
	t.Run("a rename is both kinds at once and both are reconciled", func(t *testing.T) {
		renamed := strings.Replace(files["rules-a.toml"],
			"inv-minimal-first         = { active = true }",
			"inv-renamed-first         = { active = true }", 1)
		if renamed == files["rules-a.toml"] {
			t.Fatal("fixture did not rename anything — the case would prove nothing")
		}
		raw, got := run(t, renamed)
		if strings.Contains(raw, "invalid-rules") {
			t.Errorf("a rename must reconcile, not fail closed:\n%s", raw)
		}
		if got.HookSpecificOutput == nil {
			t.Fatalf("no context was injected:\n%s", raw)
		}
		ctx := got.HookSpecificOutput.AdditionalContext
		if !strings.Contains(ctx, "# inv-renamed-first") {
			t.Errorf("the stale slug must be quarantined:\n%s", ctx)
		}
		if !regexp.MustCompile(`(?m)^inv-minimal-first[ \t]*=[ \t]*\{[ \t]*active[ \t]*=[ \t]*true`).MatchString(ctx) {
			t.Errorf("the new slug must be added as active:\n%s", ctx)
		}
	})

	// Fix round 1, IMPORTANT: parseRulesToml's row regex (matches any
	// [a-z][a-z-]* slug before this fix) and reconcileRows' row-detection
	// (always inv-/floor- only) disagreed. A row like this used to classify
	// as "unknown" here (triggering reconciliation) but reconcileRows never
	// recognised it as a row to quarantine — measured: context delivered,
	// zero quarantine notes, the row passed through uncommented, so the
	// mismatch never cleared and the hook re-reconciled every session to no
	// effect. The two now share the same (?:inv|floor)-[a-z-]+ grammar, so
	// this is a malformed row (fails closed) rather than a silent no-op.
	t.Run("a row not shaped like a rule slug fails closed, not a silent no-op reconcile", func(t *testing.T) {
		bogus := files["rules-a.toml"] + "bogus-rule = { active = true }\n"
		raw, got := run(t, bogus)
		if got.HookSpecificOutput != nil {
			t.Errorf("a row not shaped (inv|floor)-... must fail closed, not reconcile:\n%s", raw)
		}
		if !strings.Contains(raw, "invalid-rules") {
			t.Errorf("want invalid-rules, got:\n%s", raw)
		}
	})

	// Fix round 1, RULING (correcting the brief, which told the implementer
	// to keep !rulesSectionSeen fatal): a rules.toml with no [rules] table at
	// all is not a syntax fault, it is a slug set that is entirely missing —
	// reconcilable like any other slug-set mismatch. staleness.sh already
	// repairs this exact hand-written-partial shape (strictness alone, no
	// rows) into a full [rules] table plus every row; Codex must reach the
	// same repair instead of refusing the file outright. This is also what
	// makes reconcileRows' own `if (!hasRules)` insertion reachable at all —
	// dead code before this fix.
	t.Run("a rules.toml with no [rules] table at all reconciles, not fails closed", func(t *testing.T) {
		raw, got := run(t, "strictness  = \"firm\"\n")
		if strings.Contains(raw, "invalid-rules") {
			t.Errorf("a missing [rules] table must reconcile, not fail closed:\n%s", raw)
		}
		if got.HookSpecificOutput == nil {
			t.Fatalf("no context was injected:\n%s", raw)
		}
		ctx := got.HookSpecificOutput.AdditionalContext
		if !strings.Contains(ctx, "[rules]") {
			t.Errorf("a [rules] table must be opened where none existed:\n%s", ctx)
		}
		if !strings.Contains(ctx, "added 16 row(s)") {
			t.Errorf("all sixteen rows must be added — none had a table to belong to:\n%s", ctx)
		}
		for _, slug := range assessableSlugs {
			re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(slug) + `[ \t]*=[ \t]*\{[ \t]*active[ \t]*=[ \t]*true`)
			if !re.MatchString(ctx) {
				t.Errorf("rule %s must be reconciled to active = true:\n%s", slug, ctx)
			}
		}
	})

	// Fix round 1, IMPORTANT: staleness.sh reads `date +%Y-%m-%d` — the
	// process's LOCAL calendar date. Left as UTC (toISOString), this and the
	// Claude hook would disagree by |UTC-offset| hours a day on any non-UTC
	// machine, and Task 2's byte-identical comparison would go red for part
	// of every day and green only on UTC CI — a mismatch that reads as flake.
	t.Run("the reconciliation date is the local calendar date, not UTC", func(t *testing.T) {
		short := strings.Replace(files["rules-a.toml"],
			"inv-minimal-first         = { active = true }\n", "", 1)
		raw, got := run(t, short)
		if got.HookSpecificOutput == nil {
			t.Fatalf("no context was injected:\n%s", raw)
		}
		wantDate := time.Now().Format("2006-01-02")
		ctx := got.HookSpecificOutput.AdditionalContext
		if !strings.Contains(ctx, "below on "+wantDate+" (missing from") {
			t.Errorf("reconciliation date must be today's LOCAL calendar date (%s):\n%s", wantDate, ctx)
		}
	})

	// Ruling 1(a) on the task-1 brief: the original single subtest here
	// asserted the adaptive header ("**By default**") on the fixture
	// writeValidCodexOverlay builds, but that fixture creates
	// .trellis/internal/, so codex-context.mjs:720-727 takes the VENDORED branch —
	// prose comes from the overlay's own trellis.md, which
	// writeValidCodexOverlay hardcodes to trellis-a.md (firm). Posture
	// selection is bypassed entirely on that path, so the header assertion
	// could never pass. Split in two: non-fatality on the vendored fixture
	// here, posture selection on a plugin-native fixture below.
	t.Run("a missing strictness is not fatal — vendored path", func(t *testing.T) {
		noStrict := stripTOMLLine(t, files["rules-a.toml"], "strictness")
		raw, got := run(t, noStrict)
		if strings.Contains(raw, "invalid-rules") {
			t.Errorf("a missing strictness must reconcile, not fail closed:\n%s", raw)
		}
		if got.HookSpecificOutput == nil {
			t.Fatalf("a missing strictness must not be fatal — Claude defaults it:\n%s", raw)
		}
	})

	// Ruling 1(b): codex-context.mjs:717-719 already defaults posture to "b"
	// (adaptive) whenever strictness is not literally "firm" — that selection
	// logic is untouched by this task; the only change is that parseRulesToml
	// no longer treats an absent strictness as a syntax fault. This subtest
	// pins the existing default reaching a project on the PLUGIN-NATIVE path
	// (no .trellis/internal/, so codex-context.mjs:720-727 selects
	// reference/trellis-${posture}.md from PLUGIN_ROOT), which a missing
	// strictness could not reach before this task — parseRulesToml refused the
	// whole file first.
	t.Run("a missing strictness falls back to adaptive, as Claude does — posture", func(t *testing.T) {
		project := newGitProject(t)
		noStrict := stripTOMLLine(t, files["rules-b.toml"], "strictness")
		p := filepath.Join(project, ".trellis", "rules.toml")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(noStrict), 0o644); err != nil {
			t.Fatal(err)
		}
		raw, got := runCodexHook(t, vendoredBundleAbs(t), startupInput(t, project))
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

// TestCodexMandatesAndReportsTheRepair pins TRL-30 task 3 (decision-0083 host
// parity): the repair is applied and REPORTED, not proposed and gated — safe
// only because quarantine loses nothing. What must never be lost is the
// loudness. Uses writeCodexPluginRoot/newGitProject (this file's actual
// helpers), not the task brief's codexPluginRoot, which does not exist here.
func TestCodexMandatesAndReportsTheRepair(t *testing.T) {
	pluginRoot := writeCodexPluginRoot(t)
	files := payloadFiles()
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	short := strings.Replace(files["rules-a.toml"],
		"inv-minimal-first         = { active = true }\n", "", 1)
	if short == files["rules-a.toml"] {
		t.Fatal("premise: fixture removed nothing — the case would prove nothing")
	}
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

// Fix round 1, CRITICAL. The sentinel gate a few lines above where slugs is
// derived (rules.split(SENTINEL).length - 1 !== 1 && rules.endsWith(...))
// passes any rules.md carrying exactly one sentinel — it says nothing about
// whether the file has any SLUG TAGS at all. If every backtick-wrapped
// `inv-...`/`floor-...` tag is gone but the sentinel survives, slugsFromRules
// returns an empty array, slugSet is empty, and every row in the project's
// rules.toml classifies as "unknown" (nothing is ever in an empty set). That
// used to trigger reconciliation, which quarantined all sixteen rows and
// still emitted hookSpecificOutput with exit 0 — an ungoverned session that
// LOOKED governed. staleness.sh:642 already names this exact failure
// "no-slugs-in-payload", and staleness.sh:679-682 refuses to reconcile
// against it; this pins the equivalent Codex guard.
func TestCodexRefusesToReconcileAgainstAnEmptySlugSet(t *testing.T) {
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)

	rulesMdPath := filepath.Join(project, ".trellis", "internal", "rules.md")
	original := readFileT(t, rulesMdPath)
	// Strip every backtick-wrapped slug tag but leave everything else —
	// including the terminal sentinel line, which carries no such tag and so
	// is untouched — so this exercises ONLY the empty-slug-set path, not the
	// sentinel gate above it.
	tagless := regexp.MustCompile("`(?:inv|floor)-[a-z-]+`").ReplaceAllString(original, "")
	if tagless == original {
		t.Fatal("fixture stripped nothing — the case would prove nothing")
	}
	if !strings.Contains(tagless, rulesLoadedSentinel) {
		t.Fatal("fixture must still carry the sentinel — this pins the slug-set gate, not the sentinel gate")
	}
	if err := os.WriteFile(rulesMdPath, []byte(tagless), 0o644); err != nil {
		t.Fatal(err)
	}
	// A MINIMAL rules.toml (one row), not the real sixteen writeValidCodexOverlay
	// wrote: without this guard, reconciling one row against an empty slug set
	// stays well under MAX_CONTEXT_BYTES and delivers successfully — the exact
	// silently-ungoverned-but-exit-0 shape the guard exists to prevent. Using
	// the real sixteen-row preset here would let the UNRELATED byte-budget
	// guard (sixteen quarantine notes together overrun MAX_CONTEXT_BYTES) catch
	// the mutation by coincidence, which would mask whether this specific guard
	// fired — confirmed by running this test against a build with the guard
	// removed and the sixteen-row preset still in place: it failed for the
	// wrong reason (context-over-budget, not no-slugs-in-payload).
	minimal := "strictness  = \"firm\"\n\n[rules]\ninv-minimal-first         = { active = true }\n"
	rulesTomlPath := filepath.Join(project, ".trellis", "rules.toml")
	if err := os.WriteFile(rulesTomlPath, []byte(minimal), 0o644); err != nil {
		t.Fatal(err)
	}

	raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput != nil {
		t.Errorf("an empty derived slug set must fail closed, never reconcile:\n%s", raw)
	}
	if !strings.Contains(raw, "no-slugs-in-payload") {
		t.Errorf("want no-slugs-in-payload, got:\n%s", raw)
	}
}

// guards spec-0007@v1 R11-R16, R26, R35, S3-S5, S7, S17, S23
func TestCodexBootstrapPayloadContract(t *testing.T) {
	files := payloadFiles()
	block := files["block-codex.md"]
	for _, required := range []string{
		"<!-- trellis:codex-bootstrap:begin",
		"<!-- trellis:codex-bootstrap:end -->",
		rulesLoadedSentinel,
		".trellis/internal/trellis.md",
		".trellis/internal/rules.md",
		".trellis/internal/version",
		".trellis/rules.toml",
		"Trellis was not loaded",
		"best-effort",
		"sentinel alone",
		"diagnostic marker",
		"read only `.trellis/rules.toml`",
		"read only the three `.trellis/internal/` files",
		"read and validate all four installed inputs",
	} {
		if !strings.Contains(block, required) {
			t.Errorf("block-codex.md missing contract phrase %q", required)
		}
	}
	for _, forbidden := range []string{
		rulesAuthorityHeader,
		"active = true",
		"active = false",
		"**Firmly**",
		"**By default**",
	} {
		if strings.Contains(block, forbidden) {
			t.Errorf("block-codex.md embeds forbidden rule/row/posture content %q", forbidden)
		}
	}
	slugs := append([]string(nil), assessableSlugs...)
	sort.Strings(slugs)
	for _, slug := range slugs {
		if n := strings.Count(block, slug); n != 1 {
			t.Errorf("bootstrap must carry canonical slug %s exactly once, got %d", slug, n)
		}
	}
	if strings.Contains(files["trellis-a.md"], rulesLoadedSentinel) ||
		strings.Contains(files["trellis-b.md"], rulesLoadedSentinel) ||
		strings.Contains(files["block-claude.md"], rulesLoadedSentinel) {
		t.Error("sentinel belongs only at the terminal line of rules.md")
	}
	if !strings.HasSuffix(files["rules.md"], rulesLoadedSentinel+"\n") ||
		strings.Count(files["rules.md"], rulesLoadedSentinel) != 1 {
		t.Error("rules.md must end with exactly one completion sentinel")
	}
	for _, name := range []string{"trellis-a.md", "trellis-b.md"} {
		if !strings.Contains(files[name], "@rules.md\n---\n"+invariantsTrigger) {
			t.Errorf("%s must carry the fixed post-import footer", name)
		}
	}
}

// Guards what survives of spec-0007@v1's host-boundary contract after
// decision-0065. It no longer guards R17-R30/R38-R40/S10-S18/S21 — those
// covered setup's two-host vendoring contract, which the product removed. The
// comment is narrowed with the body rather than left overstating it: a test
// whose docstring claims more than it checks is how the delivery bug survived.
func TestPhaseOneSkillsAndDocsDeclareHostBoundaries(t *testing.T) {
	// Narrowed by decision-0065. This test asserted setup's two-host
	// vendoring contract: the preflight, the instruction-file inventory, the
	// canonical opposite-host block, byte-for-byte copying. Setup no longer
	// writes any of that — it writes .trellis/rules.toml and nothing else — so
	// those assertions now guard behaviour the product deliberately removed.
	//
	// What still holds is the docs' host boundary, so that is what remains here.
	// KNOWN GAP, recorded rather than hidden: with setup's vendored Codex block
	// gone and no Codex SessionStart hook shipped, Codex receives no Trellis
	// rules at all. See decision-0065 and the Codex delivery issue.
	remove := readFileT(t, "../plugins/trellis/skills/remove/SKILL.md")

	// Remove still operates over vendored state, because vendored overlays still
	// exist in the wild and are exactly what it has to clean up.
	if !strings.Contains(remove, ".trellis") {
		t.Error("remove skill must still name .trellis, the state it cleans up")
	}
}

// guards spec-0007@v1 R26, R30, R40, S17, S21
func TestCliCIProvidesNode20BeforeGoTests(t *testing.T) {
	workflow := readFileT(t, "../.github/workflows/cli-ci.yml")
	setupNode := strings.Index(workflow, "uses: actions/setup-node@v5")
	node20 := strings.Index(workflow, `node-version: "20"`)
	goTests := strings.Index(workflow, "run: go test -count=1 ./...")
	if setupNode < 0 || node20 < setupNode || goTests < node20 {
		t.Errorf("cli-ci must install Node.js 20 with actions/setup-node@v5 before Go tests execute the Codex hook")
	}
	// -count=1 is matched, not just `go test`, because the cache hazard this
	// suite lives inside is real: these tests execute the production hooks as
	// EXTERNAL files, which Go's test cache cannot see, so a hook mutation with
	// no .go change replays a stale `ok (cached)`. Reproduced by deleting
	// codex-context.mjs's empty-slug-set guard. Pinned here so dropping the flag
	// from the workflow is a red test rather than a silent loss of coverage.
	if !strings.Contains(workflow, "run: go test -count=1 ./...") {
		t.Error("cli-ci must run tests with -count=1 — the hook tests exec external files the Go test cache does not track")
	}
}

// decision-0070 D5, Codex half. `governed = false` must mean not governed on
// BOTH hosts — an opt-out one host ignores is not an opt-out.
//
// This existed unguarded: deleting the whole check from codex-context.mjs and
// refreshing the pinned manifest sha (which is what CI does) left `go test ./...`
// green. The only thing that noticed was the checksum, which fires for any byte
// change and says nothing about behaviour.
func TestCodexHookHonoursGovernedFalse(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not available")
	}
	bundle := vendoredBundleAbs(t)
	hook := filepath.Join(bundle, "hooks", "codex-context.mjs")

	run := func(t *testing.T, config string) string {
		t.Helper()
		proj := t.TempDir()
		initGitRepo(t, proj)
		writeFileT(t, filepath.Join(proj, ".trellis", "rules.toml"), config)
		cmd := exec.Command("node", hook)
		cmd.Dir = proj
		cmd.Stdin = strings.NewReader(`{"hook_event_name":"SessionStart","source":"startup","cwd":"` + proj + `"}`)
		cmd.Env = append(os.Environ(), "PLUGIN_ROOT="+bundle)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("codex hook exited non-zero (%v) — a hook must never fail the session: %s", err, out)
		}
		return string(out)
	}

	// Control: without the key the hook governs, so a silent result below cannot
	// be mistaken for the hook simply not working in this fixture.
	if got := run(t, readFileT(t, filepath.Join(bundle, "reference", "rules-b.toml"))); !strings.Contains(got, "inv-directional-flow") {
		t.Fatalf("control failed — the Codex hook delivered no rules for a normal project, so this test cannot detect the opt-out:\n%s", got)
	}
	if got := run(t, "governed = false\n"); strings.TrimSpace(got) != "" {
		t.Errorf("decision-0070 D5: a project declaring governed = false must get nothing on Codex either; got:\n%s", got)
	}
}

// TestReconciledCodexPayloadFitsContextBudget pins Ruling 6 (TRL-20 task 3,
// fix round 1): reconciled payloads must still fit inside Codex's own
// MAX_CONTEXT_BYTES once combined with the REAL payload
// (reference/trellis-a.md + reference/rules.md, not the minimal placeholders
// every other reconciliation test uses — see TestReconciledRowsParseForCodexToo's
// doc comment for why those stay minimal).
//
// The cap moved from 8000 to 9500 in fix round 1: 8000 had no recorded
// rationale, and review established Codex's actual default is ~2500 TOKENS
// (documented at https://learn.chatgpt.com/docs/hooks), not bytes, and Codex
// does not even reject over that limit — it spills to a file and gives the
// model a preview. 8000 B measured the wrong unit against a limit that fails
// open, not closed; this hook's refusal at 8000 was self-inflicted. See
// codex-context.mjs's MAX_CONTEXT_BYTES comment for the full accounting.
//
// Two cases, both real reachability review found, not just the one the
// original round named:
//   - the worst case named in the original ruling: a hand-written partial
//     file (just strictness, no rows at all), firm posture, all sixteen rows
//     missing — reconciliation adds all sixteen.
//   - the case round-1 review actually found reachable at the OLD 8000 cap:
//     one quarantined row on an otherwise untouched firm install (baseline
//     7876 B, 124 B headroom at the old cap; one quarantine note adds ~191 B
//     — enough on its own to blow an 8000 B budget, with no missing rows at
//     all).
func TestReconciledCodexPayloadFitsContextBudget(t *testing.T) {
	t.Run("worst case: hand-written partial file, firm posture, all sixteen rows missing", func(t *testing.T) {
		run := rulesTomlRun(t)
		out := run(t, "strictness  = \"firm\"\n")
		if !strings.Contains(out, "added 16 row(s)") {
			t.Fatalf("premise: all sixteen rows must be missing at firm posture, or this is not the worst case Ruling 6 names:\n%s", out)
		}
		context := nudgeContext(t, out)
		assertReconciledFitsCodexBudget(t, reconciledRowsFromContext(t, context))
	})

	t.Run("one quarantined row on an otherwise untouched firm install", func(t *testing.T) {
		run := rulesTomlRun(t)
		// The real, unmodified firm preset plus exactly one row it does not
		// recognize — every real row survives untouched (added 0), and the
		// extra row is the only thing quarantined (quarantined 1).
		withExtraRow := payloadFiles()["rules-a.toml"] + "inv-bogus-extra-rule = { active = true }\n"
		out := run(t, withExtraRow)
		if !strings.Contains(out, "added 0 row(s)") || !strings.Contains(out, "quarantined 1 row(s)") {
			t.Fatalf("premise: exactly one unrecognized row on an otherwise-untouched firm install, or this is not the case round-1 review found reachable:\n%s", out)
		}
		context := nudgeContext(t, out)
		assertReconciledFitsCodexBudget(t, reconciledRowsFromContext(t, context))
	})
}

// assertReconciledFitsCodexBudget writes reconciled — what an agent applying
// the repair actually writes to .trellis/rules.toml — alongside the REAL
// payload (not minimal placeholders: the byte budget is exactly what these
// cases exist to prove) and requires the real Codex hook to govern from it
// within MAX_CONTEXT_BYTES. The 9500 literal is deliberately not read from
// codex-context.mjs's own constant — it must reflect the value review
// actually intended, not silently track whatever the source happens to say.
func assertReconciledFitsCodexBudget(t *testing.T, reconciled string) {
	t.Helper()
	project := newGitProject(t)
	internal := filepath.Join(project, ".trellis", "internal")
	if err := os.MkdirAll(internal, 0o755); err != nil {
		t.Fatal(err)
	}
	files := payloadFiles()
	for rel, content := range map[string]string{
		"trellis.md": files["trellis-a.md"],
		"rules.md":   files["rules.md"],
		"version":    files["version"],
	} {
		if err := os.WriteFile(filepath.Join(internal, rel), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(project, ".trellis", "rules.toml"), []byte(reconciled), 0o644); err != nil {
		t.Fatal(err)
	}

	raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput == nil || got.SystemMessage != "" {
		t.Fatalf("the reconciled payload must still fit and govern under Codex, not fail closed: %s", raw)
	}
	if n := len([]byte(got.HookSpecificOutput.AdditionalContext)); n > 9500 {
		t.Errorf("reconciled payload is %d bytes, exceeds Codex MAX_CONTEXT_BYTES (9500) — Ruling 6 fix round 1 is unresolved", n)
	}
}

// TestCodexDegradesRatherThanRefusingOverBudget pins TRL-29: refusing to emit
// is a self-inflicted blackout — Codex itself spills oversized context to
// disk with a preview rather than rejecting it, so failing closed is
// strictly worse than the host's own degradation. Governance survives; the
// provenance comments are what give way, and the omission is announced.
// Uses writeCodexPluginRoot/newGitProject (see codexReconciledRows above),
// not the task brief's codexPluginRoot/t.TempDir(), which do not work here.
func TestCodexDegradesRatherThanRefusingOverBudget(t *testing.T) {
	pluginRoot := writeCodexPluginRoot(t)
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)

	// The worst case: every row foreign, so all 16 quarantine AND all 16 add.
	// Letter suffixes, not the brief's zero-padded digits: parseRulesToml's row
	// regex is `(?:inv|floor)-[a-z-]+` (codex-context.mjs:319) — no digits — so
	// a slug like "inv-foreign-rule-00" fails to match a row at all and the
	// whole file is rejected as malformed (invalid-rules) before reconciliation
	// is ever reached. Measured: the brief's literal fixture never exercises
	// this task's degradation path.
	letters := "abcdefghijklmnop"
	var b strings.Builder
	b.WriteString("strictness  = \"firm\"\n\n[rules]\n")
	for i := 0; i < 16; i++ {
		fmt.Fprintf(&b, "inv-foreign-rule-%c = { active = true }\n", letters[i])
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
	if !strings.Contains(ctx, codexDegradedMarker) {
		t.Errorf("the omission must be announced, not silent:\n%s", ctx)
	}
	// decision-0084 §6 calls this "the whole point of the branch": the CONTEXT is
	// abbreviated, the FILE the mandate asks for is not. The assertion carrying
	// that claim used to be a bare Contains(ctx, "provenance"), which survives
	// almost any rewording and would also pass on text that told the agent to
	// write the abbreviated rows back. The clause is named here instead.
	if !strings.Contains(ctx, "not the abbreviated ones shown above") {
		t.Errorf("the degraded mandate must tell the agent to write the FULL-provenance version, not the rows it can see:\n%s", ctx)
	}
	if strings.Contains(ctx, "Write .trellis/rules.toml with exactly the rows shown above") {
		t.Errorf("the degraded mandate fell back to the full-provenance wording — the file would silently lose its provenance:\n%s", ctx)
	}
	if n := len([]byte(ctx)); n > 9500 {
		t.Errorf("degraded context is %d bytes, still over the cap", n)
	}
}

// TestCodexToleratesADuplicateSlugTagInRulesMd — round-1 fix 2. slugSet
// (membership, inside parseRulesToml) already treated the derived slugs as a
// set, but the completeness check (slugs.length / slugs.some) did not, so a
// rules.md that ever tagged one slug twice made rows.size !== slugs.length
// permanently true: every Codex project would read
// .trellis/rules.toml: invalid-rules while Claude — whose own want[] in
// staleness.sh is already a set — kept governing normally from the identical
// file. Not reachable with the current payload (every one of the sixteen
// tags occurs exactly once); this guards the shape directly, by duplicating
// one tag line in the payload rules.md itself, so a future catalog edit that
// accidentally reused a slug tag does not turn into the same
// blame-the-consumer mislabel this whole task exists to close.
func TestCodexToleratesADuplicateSlugTagInRulesMd(t *testing.T) {
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)

	rulesMdPath := filepath.Join(project, ".trellis", "internal", "rules.md")
	original := readFileT(t, rulesMdPath)
	tagLine := "`inv-minimal-first`\n"
	if !strings.Contains(original, tagLine) {
		t.Fatalf("premise: fixture rules.md must carry the tag line this test duplicates: %q", tagLine)
	}
	duplicated := strings.Replace(original, tagLine, tagLine+tagLine, 1)
	if strings.Count(duplicated, tagLine) != 2 {
		t.Fatalf("premise: the duplication must actually produce two occurrences of the tag, got:\n%s", duplicated)
	}
	if err := os.WriteFile(rulesMdPath, []byte(duplicated), 0o644); err != nil {
		t.Fatal(err)
	}

	raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput == nil || got.SystemMessage != "" {
		t.Fatalf("a duplicated slug tag in rules.md must not fail every row closed: %s", raw)
	}
}

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

// TestCodexRejectsAnEmptyDerivedSlugSet is the Codex half of the same
// governance-blackout class staleness.sh refuses as `no-slugs-in-payload`.
//
// Deriving the slug set from the payload (rather than hardcoding it) opened a
// hole the hardcoded array could not have: `slugsFromRules` returns [] for a
// rules.md that keeps its sentinel but carries no trailing backticked slug on
// any line, and the sentinel gate above the derivation cannot see that — it
// checks the marker, not the tags. With `slugs` empty, parseRulesToml's two
// completeness checks pass VACUOUSLY (rows.size 0 === slugs.length 0, and
// slugs.some() over an empty array is false), so a config holding nothing but
// `strictness` and an empty `[rules]` table was ACCEPTED and the hook emitted a
// successful "loaded installed overlay" response with zero activation rows —
// a silently ungoverned session at exit 0, on the host where success is what it
// looks like. The refusal must fire before anything consumes `slugs`.
func TestCodexRejectsAnEmptyDerivedSlugSet(t *testing.T) {
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)

	// Well-formed by every check that runs BEFORE the derivation — non-empty,
	// exactly one sentinel, terminated by it — and carrying no slug tag at all.
	brokenRules := "# Trellis rules\n\nThis payload lost every trailing backticked slug tag.\n" +
		rulesLoadedSentinel + "\n"
	if strings.Count(brokenRules, rulesLoadedSentinel) != 1 || !strings.HasSuffix(brokenRules, rulesLoadedSentinel+"\n") {
		t.Fatal("premise: the fixture must still satisfy the sentinel gate, or it would fail as invalid-rules for the wrong reason")
	}
	if err := os.WriteFile(filepath.Join(project, ".trellis", "internal", "rules.md"), []byte(brokenRules), 0o644); err != nil {
		t.Fatal(err)
	}
	// The one rules.toml shape an empty slug set ACCEPTS: any actual row would
	// be rejected as unknown (slugSet is empty), which fails loudly on its own
	// — mislabelled, but loudly. This shape is the silent one.
	if err := os.WriteFile(filepath.Join(project, ".trellis", "rules.toml"), []byte("strictness = \"adaptive\"\n[rules]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput != nil {
		t.Fatalf("a payload with no derivable slugs must not deliver a successful, row-free context — that is a silent governance blackout:\n%s", raw)
	}
	want := `{"systemMessage":"Trellis hook did not load rules: .trellis/internal/rules.md: no-slugs-in-payload. The AGENTS.md bootstrap must attempt the installed overlay."}`
	if raw != want {
		t.Errorf("failure mismatch\n got: %s\nwant: %s", raw, want)
	}
}

// Two implementations of one ratified semantic (decision-0083's table) are
// only safe with a guard that makes drift a test failure rather than a field
// report. This is decision-0028's "a guard per pair", applied to the
// reconciler: both hosts reconcile the SAME fixture and their reconciled row
// blocks must be byte-identical.
func TestBothHostsReconcileIdentically(t *testing.T) {
	files := payloadFiles()
	base := files["rules-a.toml"]

	// An indented `[rules]` ALONE never reaches the reconciler: parseRulesToml
	// trims every line before matching its section regex (codex-context.mjs),
	// so "  [rules]" parses as the identical table — zero mismatch — and
	// reconcileRows is gated on `if (mismatch !== null)`. A fixture that only
	// indents the header therefore proves nothing about reconciler parity:
	// reviewer-verified by deleting `[ \t]*` from reconcileRows' own
	// `rulesHeader` regex, which left an indent-only fixture green (it never
	// runs the mutated code at all). Pairing the indent with a removed row
	// forces a real slug-set mismatch, so the reconciler must actually find
	// the indented header — to decide whether to open a second one — for this
	// fixture to exercise anything.
	indentedWithMissingRow := strings.Replace(base, "[rules]", "  [rules]", 1)
	indentedWithMissingRow = strings.Replace(indentedWithMissingRow,
		"inv-minimal-first         = { active = true }\n", "", 1)
	if indentedWithMissingRow == base {
		t.Fatal("fixture changed nothing — the case would prove nothing")
	}

	// Hoisted out of the map below so the CRLF fixture can be built from it.
	renamed := strings.Replace(base,
		"inv-minimal-first         = { active = true }",
		"inv-renamed-first         = { active = true }", 1)
	if renamed == base {
		t.Fatal("fixture renamed nothing — the case would prove nothing")
	}
	// CRLF is the divergence a HUMAN comparison found before any fixture
	// reached it (decision-0084 §5) — and, until this fixture, decision-0083's
	// "byte identity ... for LF and CRLF input" pointed at a guard with no
	// CRLF case in it. awk's RS is "\n", so a CRLF line arrives with its `\r`
	// still attached to $0; without staleness.sh:876's `{ sub(/\r$/, "") }`,
	// `print "# " $0 note` on a QUARANTINED row emits a bare CR mid-line,
	// before the note, while the JS splitter (/\r?\n/) consumes the pair and
	// never leaves one. Built on the rename, not on a plain copy of base, for
	// exactly that reason: a fixture with nothing to quarantine never reaches
	// the line the strip protects and would be vacuous. Verified by deleting
	// `{ sub(/\r$/, "") }` from staleness.sh — this subtest goes red, the
	// LF ones stay green.
	crlfRenamed := strings.ReplaceAll(renamed, "\n", "\r\n")
	if crlfRenamed == renamed || !strings.Contains(crlfRenamed, "\r\n") {
		t.Fatal("fixture is not CRLF — the case would prove nothing")
	}

	fixtures := map[string]string{
		// Reconciling fixtures: parseRulesToml finds a genuine slug-set
		// mismatch, so reconcileRows actually runs and these six compare its
		// real output byte for byte.
		"rename (missing + unknown together)":       renamed,
		"CRLF line endings, plus a rename":          crlfRenamed,
		"indented [rules] table plus a missing row": indentedWithMissingRow,
		"duplicate with a differing value": base +
			"inv-minimal-first         = { active = false }\n",
		"no [rules] table at all": "strictness  = \"firm\"\n",
		"empty file":              "",

		// Pass-through fixtures: parseRulesToml finds NO mismatch, so
		// reconcileRows is never called on either host and the compared block
		// is just the input verbatim (minus its own trailing newline — see
		// reconciledRowsFromContext / codexReconciledRowsFromContext). Kept
		// deliberately unreconciled, not upgraded to match the six above:
		// "already quarantined" pins idempotency (an already-repaired file
		// draws no second notice on either host); "missing strictness" pins
		// non-fatality (an absent strictness must not block delivery on
		// either host). Neither is meant to exercise reconcileRows itself.
		"already quarantined": base +
			"# inv-gone = { active = true }  # quarantined 2026-01-01: not in payload@old\n",
		// The brief's literal `strings.Replace(base, "strictness  = \"firm\"\n",
		// "", 1)` is a silent no-op: the real line carries a trailing comment
		// ("strictness  = \"firm\"  # firm (a·conductor) | ..."), so that exact
		// substring never occurs and Replace returns base unchanged, proving
		// nothing. stripTOMLLine strips the whole logical line by key and
		// fails loudly if it removed nothing.
		"missing strictness": stripTOMLLine(t, base, "strictness"),
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

// claudeReconciledRows runs staleness.sh against `toml` and returns the
// reconciled row block it injected. rulesTomlRun returns raw hook stdout
// (JSON with newlines escaped), and reconciledRowsFromContext's regex matches
// against real newline bytes — so the raw output must go through nudgeContext
// first, exactly as every other caller of reconciledRowsFromContext does.
func claudeReconciledRows(t *testing.T, toml string) string {
	t.Helper()
	out := rulesTomlRun(t)(t, toml)
	context := nudgeContext(t, out)
	return reconciledRowsFromContext(t, context)
}

// codexReconciledRows runs codex-context.mjs against the same fixture and
// returns the same block, so the two can be compared byte for byte.
// writeCodexPluginRoot (not the brief's nonexistent codexPluginRoot) only
// needs a syntactically valid .codex-plugin/plugin.json here: writeValidCodexOverlay
// puts a full vendored overlay in `project`, so codex-context.mjs reads the
// project's own .trellis/internal/* and never touches pluginRoot/reference.
// The project must sit under a git boundary (nearestGitBoundary) or the hook
// fails at project-root-not-found before ever reaching the reconciler — the
// brief's plain t.TempDir() has no such boundary, so newGitProject is used
// instead.
func codexReconciledRows(t *testing.T, toml string) string {
	t.Helper()
	rows, degraded := codexReconciledRowsAllowingDegraded(t, toml)
	// The byte-for-byte comparison is only about host parity while BOTH hosts
	// are on the full-provenance path. Over MAX_CONTEXT_BYTES (9500) the Codex
	// hook re-reconciles without provenance and injects that instead, so its
	// rows lose the `# added N row(s)` header and every quarantine note while
	// staleness.sh keeps them — a guaranteed, expected difference that has
	// nothing to do with the resolution table decision-0083 pins. The headroom
	// is thin enough for this to matter: the `rename` fixture already assembles
	// to 8939 B. Without this check, the first fixture to cross 9500 B fails
	// with "the two hosts reconciled the same file differently", which blames
	// the wrong thing and sends the next reader after a parity bug that is not
	// there.
	if degraded {
		t.Fatalf("this fixture crossed MAX_CONTEXT_BYTES, so the Codex hook degraded to " +
			"provenance-free rows — the two row blocks are expected to differ and comparing " +
			"them proves nothing about host parity. Shrink the fixture, or move it to " +
			"TestCROnlyLineEndingsAreTheOneKnownDivergence, which asserts the degraded shape " +
			"on purpose.")
	}
	return rows
}

// codexDegradedMarker is the one sentence only repairMandate's degraded branch
// emits (codex-context.mjs). Matched rather than recomputing the byte budget
// here: the budget lives in the hook, and a test that re-derives it drifts.
const codexDegradedMarker = "Provenance was omitted above to fit the context budget"

// codexReconciledRowsAllowingDegraded is codexReconciledRows without the
// full-provenance assertion, reporting instead whether the hook degraded, so
// the one test that expects degradation can say so explicitly.
func codexReconciledRowsAllowingDegraded(t *testing.T, toml string) (string, bool) {
	t.Helper()
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	tomlPath := filepath.Join(project, ".trellis", "rules.toml")
	if err := os.WriteFile(tomlPath, []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	_, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatalf("codex injected nothing for this fixture — it must reconcile, not refuse")
	}
	// decision-0070 D4: the hook computes a repair in memory and reports it,
	// but never writes it — the mirror of staleness.sh's own pin
	// (plugin_hook_test.go:1841, "the hook wrote .trellis/rules.toml —
	// \"the hook never writes\" is the half of decision-0070 D4 that stands").
	// That Claude-side pin only covers a project with no rules.toml at all;
	// every fixture through this helper already carries a genuine mismatch
	// that gets reconciled in the response, so this is the stronger case —
	// the file on disk must still read back as the UNRECONCILED fixture,
	// byte for byte, after a run that just told the agent to write the
	// reconciled text over it. Holds today by construction (no writeFile/
	// appendFile call in codex-context.mjs); pinning it behaviourally so a
	// regression is a red test, not a code-reading exercise.
	if after, err := os.ReadFile(tomlPath); err != nil {
		t.Fatalf("could not re-read .trellis/rules.toml after the hook ran: %v", err)
	} else if string(after) != toml {
		t.Errorf("the codex hook wrote .trellis/rules.toml — \"the hook never writes\" is the half of decision-0070 D4 that stands:\nbefore:\n%s\nafter:\n%s", toml, after)
	}
	ctx := got.HookSpecificOutput.AdditionalContext
	return codexReconciledRowsFromContext(t, ctx), strings.Contains(ctx, codexDegradedMarker)
}

// The one divergence class decision-0083's byte-identity claim does not cover,
// pinned here rather than left to be rediscovered.
//
// A CR-only file (classic-Mac line endings) is ONE line to both reconcilers —
// staleness.sh's awk has RS="\n", and codex-context.mjs splits on /\r?\n/ —
// so neither finds a single row, both classify all sixteen slugs as missing,
// and both reconcile by appending all sixteen. They then differ in two ways,
// measured, not assumed:
//
//  1. staleness.sh:876's `sub(/\r$/, "")` strips the record's trailing CR;
//     the Codex splitter leaves it, so Codex emits `...this row\r\n[rules]`
//     where Claude emits `...this row\n[rules]`.
//  2. Sixteen added rows on top of an intact sixteen-row file assemble to
//     9724 B WITH full provenance — over MAX_CONTEXT_BYTES, which is 9500 —
//     so Codex silently takes the provenance-free path and omits the
//     `# added 16 row(s) below on <date>` header that Claude writes. What it
//     delivers on that path is 9481 B, UNDER the cap: that is the degradation
//     working, not a contradiction. Keep the two numbers apart — an earlier
//     wording ("9481 B degraded — over MAX_CONTEXT_BYTES") collapsed them and
//     the claim was copied into decision-0084 as 9481 > 9500. Measured
//     2026-08-31 by running the hook on this fixture with the cap as shipped
//     and again with it raised.
//
// Both hosts still deliver and still govern; what diverges is the text of the
// repair. This test exists so that closing either divergence is a deliberate
// act with a red test to update, not a silent change — and so the claim in
// decision-0084 §"What this supersedes" has a fixture behind it.
func TestCROnlyLineEndingsAreTheOneKnownDivergence(t *testing.T) {
	base := payloadFiles()["rules-a.toml"]
	crOnly := strings.ReplaceAll(base, "\n", "\r")
	if crOnly == base || strings.Contains(crOnly, "\n") {
		t.Fatal("fixture is not CR-only — the case would prove nothing")
	}

	claude := claudeReconciledRows(t, crOnly)
	codex, degraded := codexReconciledRowsAllowingDegraded(t, crOnly)

	if !degraded {
		t.Errorf("Codex no longer degrades on the CR-only fixture — the byte budget or the " +
			"reconciler changed. Re-measure and update this test and decision-0084's " +
			"qualification of the byte-identity claim.")
	}
	for _, slug := range []string{"inv-minimal-first", "floor-intent-gate"} {
		for host, rows := range map[string]string{"claude": claude, "codex": codex} {
			if !strings.Contains(rows, slug+" = { active = true }") {
				t.Errorf("%s did not reconcile the CR-only file — %s is missing; both hosts must still deliver every rule", host, slug)
			}
		}
	}
	if claude == codex {
		t.Errorf("the CR-only divergence has closed. That is an improvement, not a failure — " +
			"delete this test and drop decision-0084's qualification of the byte-identity claim.")
	}
	if !strings.Contains(codex, "this row\r\n[rules]") {
		t.Errorf("expected Codex to keep the record's trailing CR before the appended table; got:\n%q", codex)
	}
	if strings.Contains(claude, "this row\r\n[rules]") {
		t.Errorf("expected staleness.sh:876 to strip the trailing CR; got:\n%q", claude)
	}
	if strings.Contains(codex, "# added ") {
		t.Errorf("Codex reported degraded but still emitted the added-rows header; got:\n%q", codex)
	}
	if !strings.Contains(claude, "# added 16 row(s) below on ") {
		t.Errorf("expected staleness.sh to write the added-rows header; got:\n%q", claude)
	}
}

// codexReconciledRowsFromContext extracts the reconciled `.trellis/rules.toml`
// text from a decoded Codex additionalContext (runCodexHook's
// HookSpecificOutput.AdditionalContext is already decoded — real newlines,
// no nudgeContext needed here).
//
// TRL-30 task 3 gave codex-context.mjs its own "## Rule activation was
// reconciled this session" mandate section, appended after the row block and
// before the fixed "Trellis hook loaded installed overlay: <stamp>" footer —
// mirroring reconciledRowsFromContext's own two-way stop on the Claude side
// (plugin_hook_test.go:3271, "apply regardless of their row\.\n\n(.*?)\n\n
// (?:## Rule activation...|Delivered by...)"). Before that task the row block
// ran straight into the footer with nothing between them, so a single
// unconditional stop at the footer was correct; left unconditional now, it
// would swallow the mandate text into the "row block" this function returns,
// and TestBothHostsReconcileIdentically would compare Claude's bare rows
// against Codex's rows-plus-mandate — a drift this extractor exists to catch,
// not cause. The row block is the same whether or not this session
// reconciled anything (the mandate section only exists on the reconciling
// path, same asymmetry reconciledRowsFromContext already handles on Claude).
func codexReconciledRowsFromContext(t *testing.T, context string) string {
	t.Helper()
	// TRL-29 added a second section in the same slot: on the degraded
	// NO-mismatch path there is no repair to mandate, so what follows the row
	// block is provenanceOmittedNotice's own heading instead. Left out of this
	// alternation it would be swallowed into the "row block" this function
	// returns, and TestCodexDegradesOnASecondSessionOverBudget's byte-identity
	// comparison would be comparing rows-plus-notice against bare rows.
	m := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(invariantsTrigger) +
		`\n\n(.*?)(?:\n\n## Rule activation was reconciled this session` +
		`|\n\n## Provenance comments were left out of the rows above` +
		`|\nTrellis hook loaded installed overlay: )`).
		FindStringSubmatch(context)
	if m == nil {
		t.Fatalf("could not find the row block in the codex hook's decoded context:\n%s", context)
	}
	return m[1]
}

// TestCodexProvenanceStripperMatchesItsOwnWriter is the anti-drift pin for
// TRL-29's no-mismatch degradation. stripPersistedProvenance has to recognise
// provenance an EARLIER session wrote — possibly by the other host, on an older
// date, against an older payload stamp — and the only thing that keeps a reader
// honest against a writer it never sees run is a test that runs both.
//
// Session 1 (a genuine mismatch, under budget) writes full provenance into the
// context; the file the agent then writes IS that text. Feeding that file back
// through the stripper must leave the rows with no Trellis provenance on them at
// all, which is the shape reconcileRows produces with `withProvenance = false`.
//
// The fixture carries BOTH provenance forms — a quarantine note (a suffix on a
// commented-out row) and an added-rows header (a line of its own) — because they
// are anchored differently and a single-kind fixture would pin only one of them.
func TestCodexProvenanceStripperMatchesItsOwnWriter(t *testing.T) {
	base := payloadFiles()["rules-a.toml"]
	fixture := strings.Replace(base,
		"inv-minimal-first         = { active = true }\n", "", 1)
	if fixture == base {
		t.Fatal("premise: fixture removed nothing — the case would prove nothing")
	}
	fixture += "inv-foreign-rule-a = { active = true }\n"

	// Taken from the CLAUDE host on purpose: the file an agent writes is text
	// both hosts agree on byte for byte (TestBothHostsReconcileIdentically), so
	// this cannot be satisfied by the Codex reader agreeing with a writer it
	// shares a typo with.
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
	if !strings.Contains(stripped, "inv-minimal-first = { active = true }") {
		t.Errorf("the added row did not survive the strip:\n%s", stripped)
	}
	// Everything that is not Trellis's own bookkeeping is the project's content
	// and must come through untouched — the strip is a byte-budget concession on
	// what Trellis wrote, not a licence to abbreviate a consumer's file.
	for _, kept := range []string{
		`seeded_from = "conductor"  # provenance only`,
		"[rules]  # one row per assessable catalog slug",
		"floor-intent-gate         = { active = true }  # floor — applies regardless of this row",
	} {
		if !strings.Contains(stripped, kept) {
			t.Errorf("the strip took out %q, which Trellis did not write as provenance:\n%s", kept, stripped)
		}
	}
}

// codexStripProvenance calls stripPersistedProvenance directly, by slicing its
// self-contained region out of codex-context.mjs into a throwaway ES module.
// The hook is a top-level script (it reads stdin and exits at once), so it
// cannot be imported as-is.
//
// Direct rather than end-to-end BECAUSE the end-to-end path needs a fixture
// over MAX_CONTEXT_BYTES, where a stripper defect and a budget-arithmetic
// defect are indistinguishable. TestCodexDegradesOnASecondSessionOverBudget
// covers the wired-up path; this covers the reader against its writer.
func codexStripProvenance(t *testing.T, source string) string {
	t.Helper()
	raw, err := os.ReadFile("../plugins/trellis/hooks/codex-context.mjs")
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	start := strings.Index(body, "const QUARANTINE_NOTE_TEMPLATE")
	if start < 0 {
		t.Fatal("QUARANTINE_NOTE_TEMPLATE not found in codex-context.mjs — the extraction is broken, and a helper that reads nothing proves nothing")
	}
	fn := strings.Index(body, "function stripPersistedProvenance(")
	if fn < start {
		t.Fatal("stripPersistedProvenance is not below the templates it derives from — the extraction is broken")
	}
	closing := strings.Index(body[fn:], "\n}\n")
	if closing < 0 {
		t.Fatal("stripPersistedProvenance has no closing '}' — the extraction is broken")
	}
	region := body[start : fn+closing+len("\n}\n")]

	dir := t.TempDir()
	mod := filepath.Join(dir, "strip.mjs")
	// Reads stdin rather than embedding the fixture, so no escaping of the
	// fixture's own quotes and backslashes can quietly change what is measured.
	script := region + `
const input = await new Promise((resolve) => {
  let s = "";
  process.stdin.setEncoding("utf8");
  process.stdin.on("data", (c) => { s += c; });
  process.stdin.on("end", () => resolve(s));
});
process.stdout.write(stripPersistedProvenance(input));
`
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

// TestCodexDegradesOnASecondSessionOverBudget is TRL-29's remaining half, run as
// the sequence that produced it rather than asserted as a unit.
//
//  1. Session 1 sees a mismatched file, degrades, delivers, and mandates a
//     FULL-PROVENANCE write.
//  2. The agent complies. That file is what session 2 reads.
//  3. Session 2 has NO mismatch. The degradation used to be gated on
//     `mismatch !== null`, so nothing was offered up and the hard refusal fired
//     instead — permanently, because nothing about that file changes again,
//     while staleness.sh governed happily from the identical bytes.
//
// Nine foreign rows is the measured cliff (decision-0084 §6, reproduced on
// 3f44620 before this fix: at N = 8 session 2 delivered 9404 B unaided, at N = 9
// it refused). Below the cliff this test would pass without the fix, which is
// why the fixture is not smaller.
func TestCodexDegradesOnASecondSessionOverBudget(t *testing.T) {
	fixture := payloadFiles()["rules-a.toml"]
	letters := "abcdefghi"
	for i := 0; i < len(letters); i++ {
		fixture += fmt.Sprintf("inv-foreign-rule-%c = { active = true }\n", letters[i])
	}

	session1Rows, session1Degraded := codexReconciledRowsAllowingDegraded(t, fixture)
	if !session1Degraded {
		t.Fatal("premise: session 1 must be over budget and degrade, or this fixture is not the reopened case")
	}
	// The file the mandate produces, taken from the CLAUDE host: it is text both
	// hosts agree on byte for byte (TestBothHostsReconcileIdentically), and
	// sourcing it there keeps this test from being satisfiable by the Codex hook
	// agreeing with itself.
	repaired := claudeReconciledRows(t, fixture) + "\n"
	if !strings.Contains(repaired, "# quarantined ") {
		t.Fatalf("premise: the file the mandate produces must carry persisted provenance:\n%s", repaired)
	}

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

	// The file is the archive, the injection is the working set — two halves.
	// The file keeps every note...
	after, err := os.ReadFile(tomlPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != repaired {
		t.Errorf("the hook wrote .trellis/rules.toml — decision-0070 D4 says it never does:\nbefore:\n%s\nafter:\n%s", repaired, after)
	}
	// ...and the agent is never told to write the abbreviated copy back. On this
	// path nothing asked for a repair, so there must be no write instruction at
	// all; the mismatch path's counterpart of this property is the "not the
	// abbreviated ones shown above" clause TestCodexDegradesRatherThanRefusingOverBudget
	// names.
	if strings.Contains(ctx, "Write .trellis/rules.toml") {
		t.Errorf("no repair ran, so nothing may instruct a write — an abbreviated file is the one outcome this whole branch exists to prevent:\n%s", ctx)
	}

	// The strongest statement of "same mechanism, different trigger": what
	// session 1 injected when IT degraded is what session 2 injects, byte for
	// byte. If these ever diverge, one path is dropping provenance the other
	// keeps and the archive/working-set split has grown a seam.
	session2Rows := codexReconciledRowsFromContext(t, ctx)
	if session1Rows != session2Rows {
		t.Errorf("the two degraded paths disagree about the same file\nsession 1:\n%s\nsession 2:\n%s", session1Rows, session2Rows)
	}
}

// TestCodexKeepsProvenanceWhenItFits is the over-refusal guard for the branch
// above: the strip is a byte-budget concession, not a policy. A file carrying
// persisted provenance that assembles UNDER the cap must be injected verbatim,
// notes and all. Degrading a session that had the bytes to spare would throw the
// provenance away for nothing and make every session's context depend on a
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

// TestCodexDegradesPersistedProvenanceOnTheMismatchPathToo is the same blackout
// one step to the left, and the reason the over-budget branch strips before it
// reconciles rather than only on the no-mismatch side.
//
// The degraded MISMATCH path leaves provenance off the notes it would GENERATE.
// That says nothing about notes the file already carries: a project that was
// repaired once and then drifts again arrives here with both. Reconciling from
// the raw file leaves every persisted note in the injected copy, so this session
// degrades strictly LESS than the identical file with nothing to reconcile —
// and at nine persisted rows plus one new foreign row that is the difference
// between governing and injecting nothing.
//
// Mutation-proven: `reconcileRows(rulesToml, ...)` in place of
// `reconcileRows(stripped, ...)` makes this refuse.
func TestCodexDegradesPersistedProvenanceOnTheMismatchPathToo(t *testing.T) {
	fixture := payloadFiles()["rules-a.toml"]
	letters := "abcdefghi"
	for i := 0; i < len(letters); i++ {
		fixture += fmt.Sprintf("inv-foreign-rule-%c = { active = true }\n", letters[i])
	}
	// An already-repaired file: nine rows quarantined, each carrying its note.
	repaired := claudeReconciledRows(t, fixture) + "\n"
	if strings.Count(repaired, "# quarantined ") != len(letters) {
		t.Fatalf("premise: the fixture must carry one persisted note per foreign row:\n%s", repaired)
	}
	// ...which then drifts again. One new foreign row is all it takes.
	drifted := repaired + "inv-foreign-rule-j = { active = true }\n"

	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	tomlPath := filepath.Join(project, ".trellis", "rules.toml")
	if err := os.WriteFile(tomlPath, []byte(drifted), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))

	if strings.Contains(raw, "context-over-budget") {
		t.Fatalf("a repaired file that drifted again refused — the degraded mismatch path must give up the file's persisted provenance too:\n%s", raw)
	}
	if got.HookSpecificOutput == nil {
		t.Fatalf("nothing injected:\n%s", raw)
	}
	ctx := got.HookSpecificOutput.AdditionalContext
	for _, slug := range assessableSlugs {
		if !regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(slug) + `[ \t]*=`).MatchString(ctx) {
			t.Errorf("rule %s must still be delivered", slug)
		}
	}
	if n := len([]byte(ctx)); n > 9500 {
		t.Errorf("degraded context is %d bytes, still over the cap", n)
	}
	// Both provenance kinds are off: the notes this session would have written
	// AND the notes the file already carried.
	if strings.Contains(ctx, "# quarantined ") {
		t.Errorf("the degraded mismatch path kept persisted provenance it could have given up:\n%s", ctx)
	}
	// Every quarantined row still keeps its line and its value, persisted ones
	// included — the degradation drops notes, never rows.
	for i := 0; i <= len(letters); i++ {
		slug := fmt.Sprintf("inv-foreign-rule-%c", byte('a')+byte(i))
		if !strings.Contains(ctx, "# "+slug+" = { active = true }") {
			t.Errorf("quarantined row %s lost its line — quarantine never deletes:\n%s", slug, ctx)
		}
	}
	// The repair still describes what THIS session did, not a running total:
	// nine rows were already commented out and are invisible to the counters,
	// so exactly one row is newly quarantined.
	if !strings.Contains(ctx, "added 0 row(s); quarantined 1 row(s)") {
		t.Errorf("the strip changed the reconciliation counts — it must only remove comment text:\n%s", ctx)
	}
	// The file is the archive: it keeps every note.
	after, err := os.ReadFile(tomlPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != drifted {
		t.Errorf("the hook wrote .trellis/rules.toml — decision-0070 D4 says it never does")
	}
}

// TestCodexDoesNotOverRefuseTheLegitimateShapes runs every rules.toml shape a
// real project can present through the real hook and requires that none of them
// reaches a refusal.
//
// Over-correction is the failure mode this class of work keeps producing: two
// guards on an earlier TRL-29 branch tightened into refusing healthy payloads,
// which costs a consumer exactly what the silent case does. The sweep is broad
// and boring on purpose — a narrow test of the newly-changed path would not have
// caught either of them.
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
			if got.HookSpecificOutput == nil {
				t.Fatalf("no context injected for a legitimate shape:\n%s", raw)
			}
			ctx := got.HookSpecificOutput.AdditionalContext
			if n := len([]byte(ctx)); n > 9500 {
				t.Errorf("%s assembled to %d bytes, over the cap", name, n)
			}
			for _, slug := range assessableSlugs {
				if !regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(slug) + `[ \t]*=`).MatchString(ctx) {
					t.Errorf("rule %s was not delivered for a legitimate shape", slug)
				}
			}
		})
	}

	// decision-0070 D5: an opt-out is silence, not a refusal — and not a
	// delivery either. Asserted rather than folded into the table above,
	// because "emitted nothing" and "refused" are exactly the two outcomes
	// this sweep exists to tell apart, and a shared assertion would blur them.
	t.Run("governed = false", func(t *testing.T) {
		project := newGitProject(t)
		writeValidCodexOverlay(t, project)
		if err := os.WriteFile(filepath.Join(project, ".trellis", "rules.toml"), []byte("governed = false\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		raw, _ := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
		if raw != "" {
			t.Fatalf("an opt-out must emit nothing at all — not a refusal, not a delivery:\n%s", raw)
		}
	})

	// A project with no .trellis/rules.toml is not a Trellis project on this
	// host: nearestOverlay walks up looking for exactly that file, so the hook
	// stops at project-root-not-found before any budget arithmetic runs. What
	// matters for this sweep is that it is NOT a budget refusal — the byte cap
	// must not be what a project without an overlay hears about.
	t.Run("no project rules.toml at all", func(t *testing.T) {
		project := newGitProject(t)
		writeValidCodexOverlay(t, project)
		if err := os.Remove(filepath.Join(project, ".trellis", "rules.toml")); err != nil {
			t.Fatal(err)
		}
		raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
		if strings.Contains(raw, "context-over-budget") {
			t.Fatalf("a project with no rules.toml must never be refused for budget:\n%s", raw)
		}
		if got.HookSpecificOutput != nil {
			t.Fatalf("a project with no overlay must not be governed from one:\n%s", raw)
		}
		if !strings.Contains(raw, "project-root-not-found") {
			t.Errorf("expected the overlay-not-found path, got:\n%s", raw)
		}
	})
}

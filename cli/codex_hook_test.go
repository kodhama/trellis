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
	// .trellis/internal/, so codex-context.mjs:424 takes the VENDORED branch —
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

	// Ruling 1(b): codex-context.mjs:421-423 already defaults posture to "b"
	// (adaptive) whenever strictness is not literally "firm" — that selection
	// logic is untouched by this task; the only change is that parseRulesToml
	// no longer treats an absent strictness as a syntax fault. This subtest
	// pins the existing default reaching a project on the PLUGIN-NATIVE path
	// (no .trellis/internal/, so codex-context.mjs:424 selects
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

// Fix round 1, CRITICAL. The sentinel gate a few lines above where slugs is
// derived (rules.split(SENTINEL).length - 1 !== 1 && rules.endsWith(...))
// passes any rules.md carrying exactly one sentinel — it says nothing about
// whether the file has any SLUG TAGS at all. If every backtick-wrapped
// `inv-...`/`floor-...` tag is gone but the sentinel survives, slugsFromRules
// returns an empty array, slugSet is empty, and every row in the project's
// rules.toml classifies as "unknown" (nothing is ever in an empty set). That
// used to trigger reconciliation, which quarantined all sixteen rows and
// still emitted hookSpecificOutput with exit 0 — an ungoverned session that
// LOOKED governed. staleness.sh:625-636 already names this exact failure
// "no-slugs-in-payload" and refuses to reconcile against it; this pins the
// equivalent Codex guard.
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
	goTests := strings.Index(workflow, "run: go test ./...")
	if setupNode < 0 || node20 < setupNode || goTests < node20 {
		t.Errorf("cli-ci must install Node.js 20 with actions/setup-node@v5 before Go tests execute the Codex hook")
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

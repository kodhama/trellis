package main

// Production-contract tests for spec-0007@v1. The captured input shape is
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
	"sort"
	"strings"
	"testing"
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
	mutateAndFail(t, ".trellis/rules.toml", strings.Replace(payloadFiles()["rules-a.toml"], "inv-minimal-first", "inv-unknown-rule", 1), "invalid-rules")
	mutateAndFail(t, ".trellis/rules.toml", strings.Replace(payloadFiles()["rules-a.toml"], "inv-minimal-first         = { active = true }\n", "", 1), "invalid-rules")
	mutateAndFail(t, ".trellis/rules.toml", payloadFiles()["rules-a.toml"]+"inv-minimal-first = { active = true }\n", "invalid-rules")

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
			t.Errorf("malformed/duplicate Trellis TOML must fail exactly\n got: %s\nwant: %s", raw, want)
		}
	}

	beforeRules := strings.Replace(canonical, "[rules]", "seeded_from = 'duplicate'\n\n[rules]", 1)
	assertInvalid(t, beforeRules)
	beforeRules = strings.Replace(canonical, "[rules]", "strictness = 'adaptive'\n\n[rules]", 1)
	assertInvalid(t, beforeRules)
	assertInvalid(t, canonical+"\n[rules]\n")
	assertInvalid(t, canonical+"\n[other]\n")
	assertInvalid(t, strings.Replace(canonical, "[rules]", "unexpected = 'value'\n\n[rules]", 1))
	assertInvalid(t, canonical+"inv-minimal-first = { active = true }\n")
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

// guards spec-0007@v1 R17-R30, R38-R40, S10-S18, S21
func TestPhaseOneSkillsAndDocsDeclareHostBoundaries(t *testing.T) {
	setup := readFileT(t, "../plugins/trellis/skills/setup/SKILL.md")
	remove := readFileT(t, "../plugins/trellis/skills/remove/SKILL.md")
	readme := readFileT(t, "../plugins/trellis/README.md")
	rootReadme := readFileT(t, "../README.md")

	for _, required := range []string{
		"PLUGIN_ROOT",
		"CLAUDE_PLUGIN_ROOT",
		"Node.js 20",
		"bootstrap-only",
		"preflight",
		"AGENTS.md",
		"block-codex.md",
		"byte-for-byte",
		"never overwrite",
		"fresh trusted local Codex startup",
	} {
		if !strings.Contains(setup, required) {
			t.Errorf("setup skill missing Phase 1 contract phrase %q", required)
		}
	}
	instructionFiles := []string{
		"CLAUDE.md",
		"AGENTS.md",
		"GEMINI.md",
		".github/copilot-instructions.md",
		".clinerules",
	}
	for _, name := range instructionFiles {
		if !strings.Contains(setup, name) {
			t.Errorf("setup preflight must inventory documented instruction file %q", name)
		}
		if !strings.Contains(remove, name) {
			t.Errorf("remove preflight must inventory documented instruction file %q", name)
		}
	}
	for _, required := range []string{
		"legacy/manual",
		"inline/full-rule",
		"selected target",
		"outside the selected target",
		"explicit consent",
		"same transaction",
		"before the first project write",
		"canonical opposite-host block",
		"allowed and byte-preserved",
		"When Codex is selected",
		"When Claude is selected",
	} {
		if !strings.Contains(setup, required) {
			t.Errorf("setup migration/atomicity contract missing %q", required)
		}
	}
	for _, required := range []string{
		"every documented instruction file",
		"before any edit",
		"all recognized managed blocks",
		"delete the shared `.trellis/` overlay last",
	} {
		if !strings.Contains(remove, required) {
			t.Errorf("remove inventory/atomicity contract missing %q", required)
		}
	}
	for _, required := range []string{
		"AGENTS.md",
		"CLAUDE.md",
		"preflight",
		"before",
		"ambiguous",
		"already absent",
	} {
		if !strings.Contains(remove, required) {
			t.Errorf("remove skill missing product-wide cleanup phrase %q", required)
		}
	}
	for _, doc := range []string{readme, rootReadme} {
		for _, required := range []string{
			"trusted local Codex",
			"fresh startup",
			"Node.js 20",
			"best-effort",
			"resume",
			"compact",
			"subagent",
			"desktop",
			"IDE",
			"cloud",
			"product-wide",
			"preset",
			"Claude-hook replacement",
			"host-native transport",
			"`seed`",
			"`custom`",
		} {
			if !strings.Contains(doc, required) {
				t.Errorf("README support boundary missing %q", required)
			}
		}
	}
}

// guards spec-0007@v1 R26, R30, R40, S17, S21
func TestCliCIProvidesNode20BeforeGoTests(t *testing.T) {
	workflow := readFileT(t, "../.github/workflows/cli-ci.yml")
	setupNode := strings.Index(workflow, "uses: actions/setup-node@v4")
	node20 := strings.Index(workflow, `node-version: "20"`)
	goTests := strings.Index(workflow, "run: go test ./...")
	if setupNode < 0 || node20 < setupNode || goTests < node20 {
		t.Errorf("cli-ci must install Node.js 20 with actions/setup-node@v4 before Go tests execute the Codex hook")
	}
}

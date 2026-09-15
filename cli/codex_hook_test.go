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
	"slices"
	"sort"
	"strings"
	"testing"
)

const rulesLoadedSentinel = "<!-- trellis:rules-loaded -->"

// The generated header's own terminator (TRL-10), spelled out here rather than
// imported from apply.go: a test that reads the production const cannot catch
// the production const changing.
const proseCompleteMarker = "<!-- trellis:prose-complete -->"

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

// frozenFirmOverlayHeader is the header a vendored overlay carries when it was
// written before TRL-97 retired the posture presets: reference/trellis-a.md as it
// last shipped (d4d76e0). The plugin no longer ships it and vendored projects keep
// theirs (KTD7), so the vendored fixtures freeze it here instead of reading a
// payload key that is gone.
const frozenFirmOverlayHeader = "# How to work in this project\n\n" +
	"You are working in a project that follows **Trellis** — a small, load-bearing set of working rules on top of the project's own process. **Follow the rules below as you work here.** They add guardrails; they don't replace this project's own instructions.\n\n" +
	"**How strictly to follow them:** **Firmly** — treat these as hard requirements. Follow them as written; don't skip or soften one without the human's explicit say-so.\n\n" +
	"@rules.md\n---\n" +
	"If a rule seems ambiguous, or in tension with this project's own instructions, read its entry in `.trellis/internal/invariants.md` — the description and with/without examples — before deviating.\n" +
	"<!-- trellis:prose-complete -->\n"

// legacyFirmRulesToml is reference/rules-a.toml as it last shipped (d4d76e0):
// the file every project seeded from the firm preset still holds, sixteen inert
// `active = true` rows and two top-level keys that no longer mean anything.
const legacyFirmRulesToml = "# Rows govern rule activation live (see the authority note in the project instructions).\n\n" +
	"seeded_from = \"conductor\"  # provenance only — the rows below win if they diverge\n" +
	"strictness  = \"firm\"  # firm (a·conductor) | adaptive (b·author-adapt)\n\n" +
	"[rules]  # one row per assessable catalog slug (signature-catalog-v1)\n" +
	"inv-directional-flow      = { active = true }\n" +
	"inv-handover-points       = { active = true }\n" +
	"inv-intent-locus          = { active = true }\n" +
	"inv-ratifiable-artifacts  = { active = true }\n" +
	"inv-graph-maintenance     = { active = true }\n" +
	"inv-self-improvement      = { active = true }\n" +
	"inv-deliberate-succession = { active = true }\n" +
	"inv-no-orphan-followups   = { active = true }\n" +
	"inv-gate-at-handover      = { active = true }\n" +
	"inv-independent-judgment  = { active = true }\n" +
	"inv-auditable-archive     = { active = true }\n" +
	"inv-bounded-context       = { active = true }\n" +
	"inv-minimal-first         = { active = true }\n" +
	"inv-clarify-before-commit = { active = true }\n" +
	"floor-transparency        = { active = true }  # floor — applies regardless of this row\n" +
	"floor-intent-gate         = { active = true }  # floor — applies regardless of this row\n"

// writeValidCodexOverlay writes a vendored overlay: the frozen header, the
// current rules payload and version stamp, and the sparse project file every
// plugin-native fixture uses.
func writeValidCodexOverlay(t *testing.T, project string) {
	t.Helper()
	for rel, content := range map[string]string{
		".trellis/internal/trellis.md": frozenFirmOverlayHeader,
		".trellis/internal/rules.md":   payloadFile(t, "rules.md"),
		".trellis/internal/version":    payloadFile(t, "version"),
		".trellis/rules.toml":          configOnlyProjectRules,
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

// replaceOnce is strings.Replace for a fixture, failing when old is absent so a
// case that changed nothing cannot pass by testing the unchanged fixture.
func replaceOnce(t *testing.T, s, old, replacement string) string {
	t.Helper()
	out := strings.Replace(s, old, replacement, 1)
	if out == s {
		t.Fatalf("fixture drift: %q not found, so the case would test the unchanged fixture", old)
	}
	return out
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
	// SCOPED TO readRequired's OWN BODY, not to the whole file. TRL-72 added a
	// second fs.readSync — payloadDefect's chunked scan of the CONSULTED
	// invariants copy — and a bare Contains over the file was then satisfiable
	// by that one alone, so this stopped proving the AUTHORITATIVE reader is
	// still bounded. Found by review of #296; the mutant is deleting
	// readRequired's readSync loop while payloadDefect keeps the file passing.
	body := functionBody(t, source, "function readRequired(")
	// TRL-97 gave the project's rules.toml its own read bound, so the bound is a
	// value here rather than the literal cap. What is pinned is unchanged: the
	// buffer is one byte over the bound, the size is checked before the read, and
	// the bound defaults to MAX_CONTEXT_BYTES for every caller that does not
	// state its own.
	for _, required := range []string{
		"fs.readSync",
		"options.maxBytes ?? MAX_CONTEXT_BYTES",
		"Buffer.alloc(maxBytes + 1)",
		"stat.size > maxBytes",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("readRequired's bounded-read implementation is missing %q", required)
		}
	}
	if strings.Contains(source, "fs.readFileSync(absolute") {
		t.Error("Codex hook must not read an authoritative file wholly before enforcing its byte bound")
	}
}

// functionBody returns the source of the function opening with `decl`, from its
// declaration to the first closing brace at column 0. That brace convention is
// how every top-level function in codex-context.mjs ends, and it is the same
// scan TestNoPayloadReadBypassesTheGateway uses on staleness.sh's payload_read.
//
// Fatal rather than empty when the declaration or its terminator is missing: a
// scan that silently returns "" would make every assertion built on it fail for
// the wrong reason, or — worse, if the assertions were negative — pass.
func functionBody(t *testing.T, source, decl string) string {
	t.Helper()
	start := strings.Index(source, decl)
	if start < 0 {
		t.Fatalf("%q not found — the scan is broken, and a guard that reads nothing proves nothing", decl)
	}
	end := strings.Index(source[start:], "\n}\n")
	if end < 0 {
		t.Fatalf("%q has no closing brace at column 0 — the scan is broken", decl)
	}
	return source[start : start+end]
}

// guards spec-0007@v1 R1, R2, R7, R10, R31, R34-R36, S1, S2, S19
func TestCodexHookValidStartupAndLiveRows(t *testing.T) {
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	rulesPath := filepath.Join(project, ".trellis", "rules.toml")
	// The sparse file TRL-97 seeds: a table and no rows.
	writeFileT(t, rulesPath, "[rules]\n")
	pluginRoot := writeCodexPluginRoot(t)
	raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
	if got.HookSpecificOutput == nil || warningsBesideVendoredInvariants(t, got.SystemMessage) != "" {
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
		strings.TrimSpace(payloadFile(t, "version")),
		"\n" + activationHeading + "\n",
		expectedActivationSentence(nil),
	} {
		if n := strings.Count(context, unique); n != 1 {
			t.Errorf("assembled context must contain %q once, got %d", unique, n)
		}
	}
	if strings.Contains(context, "strictness") {
		t.Errorf("a sparse project file carries no strictness, and nothing the hook adds may mention one:\n%s", context)
	}
	// KTD7. This overlay's frozen text still says a rule applies only when its
	// row says active = true, and the file has no rows. The computed sentence is
	// what tells the model every rule applies; no row is added to the echoed
	// file and nothing asks for one, now that nothing reconciles.
	if region, ok := activationRegion(context, "\nTrellis hook loaded installed overlay: "); !ok ||
		region != expectedActivationSentence(nil)+"\n\n[rules]\n" {
		t.Errorf("a vendored overlay whose file lacks every row must get the sentence and the file verbatim, nothing added\n got: %q\nwant: %q", region, expectedActivationSentence(nil)+"\n\n[rules]\n")
	}
	// TRL-10: the boundary the Codex agent is told to look for is two markers in
	// order, so that is what the injected context must expose. This used to
	// assert the sentinel was followed by `---` and the invariants sentence —
	// the prose landmarks block-codex.md keyed on, asserted from the same side
	// of the contract, which is why neither could catch the other drifting.
	if i := strings.Index(context, rulesLoadedSentinel); i < 0 ||
		!strings.Contains(context[i:], "\n"+proseCompleteMarker+"\n") {
		t.Errorf("assembled context must expose the sentinel-plus-end-marker boundary, in that order:\n%s", context)
	}
	// The footer prose still ships between them (decision-0053), but as prose
	// nothing keys on rather than as the boundary itself.
	if !strings.Contains(context, rulesLoadedSentinel+"\n\n---\n"+invariantsTrigger) {
		t.Error("assembled context lost the generated post-import footer")
	}
	if strings.Contains(context, "../plugins/trellis/reference") {
		t.Error("assembled context must not source rule content from the plugin payload")
	}

	writeFileT(t, rulesPath, "[rules]\ninv-handover-points = { active = false }\n")
	_, edited := runCodexHook(t, pluginRoot, startupInput(t, project))
	if edited.HookSpecificOutput == nil ||
		!strings.Contains(edited.HookSpecificOutput.AdditionalContext, "\ninv-handover-points = { active = false }\n") ||
		!strings.Contains(edited.HookSpecificOutput.AdditionalContext, expectedActivationSentence([]string{"inv-handover-points"})) {
		t.Error("next startup must read the consumer's edited row without refresh, and the computed sentence must name the rule it switches off")
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

	// TRL-97: an unrecognised strictness used to refuse the whole file here as
	// invalid-rules while Claude read it as adaptive. strictness means nothing
	// now, so the file delivers every rule and nothing is said about the key. A
	// project file refuses only when it cannot be read (the parity table's
	// unreadable row) or when it outgrows the runaway guard below.
	configPath := filepath.Join(project, ".trellis", "rules.toml")
	originalConfig := readFileT(t, configPath)
	writeFileT(t, configPath, "strictness = \"loose\"\n[rules]\n")
	if raw, got := runCodexHook(t, pluginRoot, startupInput(t, project)); got.HookSpecificOutput == nil ||
		!strings.Contains(got.HookSpecificOutput.AdditionalContext, expectedActivationSentence(nil)) ||
		warningsBesideVendoredInvariants(t, got.SystemMessage) != "" {
		t.Errorf("an unrecognised strictness must deliver every rule with nothing said about it: %s", raw)
	}
	writeFileT(t, configPath, originalConfig)

	largeRules := strings.TrimSuffix(payloadFile(t, "rules.md"), rulesLoadedSentinel+"\n") +
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

	// The project file is read up to its own runaway guard and echoed only while
	// it and its warnings fit the shared bound (TRL-97). A file past the old 9500-byte read
	// bound used to refuse here; it now delivers the too-large line instead, and
	// only a file past MAX_PROJECT_CONFIG_BYTES (one MiB) is refused.
	writeFileT(t, configPath, originalConfig+"#"+strings.Repeat("x", 8001)+"\n")
	if raw, got := runCodexHook(t, pluginRoot, startupInput(t, project)); got.HookSpecificOutput == nil ||
		!strings.Contains(got.HookSpecificOutput.AdditionalContext, expectedTooLargeLine()) {
		t.Errorf("a project file past the old read bound must deliver the too-large line, not refuse: %s", raw)
	}
	writeFileT(t, configPath, "#"+strings.Repeat("x", 1024*1024)+"\n")
	assertFailure(t, startupInput(t, project), ".trellis/rules.toml", "context-over-budget")
	writeFileT(t, configPath, originalConfig)

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
	writeFileT(t, filepath.Join(project, ".trellis", "rules.toml"),
		"[rules]\nfloor-intent-gate = { active = false }\nfloor-transparency = { active = false }\n")
	_, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput == nil {
		t.Fatal("false floor rows must retain successful context delivery")
	}
	// TRL-97 folded the floor-row warning into the shared rules.toml warnings:
	// one per row, in line order, inside the context and mirrored to
	// systemMessage, in the words staleness.sh uses too.
	want := []string{warnFloorRow(2, "floor-intent-gate"), warnFloorRow(3, "floor-transparency")}
	if beside := warningsBesideVendoredInvariants(t, got.SystemMessage); beside != strings.Join(want, " ") {
		t.Errorf("floor warning mismatch in systemMessage\n got: %q\nwant: %q", beside, strings.Join(want, " "))
	}
	context := got.HookSpecificOutput.AdditionalContext
	if !strings.Contains(context, strings.Join(want, "\n")+"\n") {
		t.Errorf("the floor warnings must be in the injected context, in line order:\n%s", context)
	}
	if !strings.Contains(context, expectedActivationSentence(nil)) {
		t.Errorf("a floor row set false switches nothing off, and the computed sentence must say so:\n%s", context)
	}
}

// guards spec-0007@v1 R7, R8, R20, R31, S1, S6, S12
//
// TRL-97 turned this from a strict schema into a tolerant one, on the vendored
// branch where a legacy preset is most likely to live. Everything the old presets
// wrote still delivers silently. The top-level keys that no longer mean anything
// are silent whatever they hold, including values the old parser refused. What
// used to refuse the whole file now delivers, with one warning naming the entry.
// Every row shape is pinned on both hosts by TestBothHostsClassifyRulesRowsIdentically.
func TestCodexHookRulesTomlSchema(t *testing.T) {
	project := newGitProject(t)
	writeValidCodexOverlay(t, project)
	pluginRoot := writeCodexPluginRoot(t)
	rulesPath := filepath.Join(project, ".trellis", "rules.toml")
	canonical := legacyFirmRulesToml

	assertDelivers := func(t *testing.T, name, source string, warnings ...string) {
		t.Helper()
		writeFileT(t, rulesPath, source)
		raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
		if got.HookSpecificOutput == nil {
			t.Errorf("%s: must deliver, not refuse: %s", name, raw)
			return
		}
		want := strings.Join(warnings, " ")
		if beside := warningsBesideVendoredInvariants(t, got.SystemMessage); beside != want {
			t.Errorf("%s: warnings\n got: %q\nwant: %q", name, beside, want)
		}
	}

	assertDelivers(t, "the legacy firm preset", canonical)
	assertDelivers(t, "TOML literal strings", replaceOnce(t, replaceOnce(t, canonical,
		`seeded_from = "conductor"`, `seeded_from = 'conductor'`),
		`strictness  = "firm"`, `strictness  = 'firm'`))
	assertDelivers(t, "a \\U basic-string escape", replaceOnce(t, canonical,
		`strictness  = "firm"`, `strictness  = "\U00000066irm"`))
	assertDelivers(t, "tab whitespace", replaceOnce(t, canonical,
		`strictness  = "firm"`, "strictness\t=\t\"firm\""))
	assertDelivers(t, "a second seeded_from", replaceOnce(t, canonical, "[rules]", "seeded_from = 'duplicate'\n\n[rules]"))
	assertDelivers(t, "a second strictness", replaceOnce(t, canonical, "[rules]", "strictness = 'adaptive'\n\n[rules]"))
	for _, invalidValue := range []string{
		`seeded_from = "\/"`,
		`seeded_from = "\x41"`,
		`seeded_from = "\uD800"`,
		`seeded_from = "\U00110000"`,
		"seeded_from = \"bad" + string(rune(1)) + "\"",
	} {
		assertDelivers(t, "a seeded_from the old parser refused", replaceOnce(t, canonical,
			`seeded_from = "conductor"`, invalidValue))
	}

	// The legacy preset is 22 lines, so a table appended after one blank line
	// opens on line 24, and a key inserted before [rules] sits on line 6.
	assertDelivers(t, "a second [rules] table", canonical+"\n[rules]\n", warnRulesAgain(24))
	assertDelivers(t, "another table", canonical+"\n[other]\n", warnOtherTable(24))
	assertDelivers(t, "an unknown top-level key",
		replaceOnce(t, canonical, "[rules]", "unexpected = 'value'\n\n[rules]"), warnUnknownKey(6, "unexpected"))
	assertDelivers(t, "an NBSP between a key and its equals sign",
		replaceOnce(t, canonical, `strictness  = "firm"`, "strictness\u00a0= \"firm\""), warnTopLevelLine(4))
}

// Fix round 1, CRITICAL, kept through TRL-97. The sentinel gate a few lines above
// where slugs is derived (rules.split(SENTINEL).length - 1 !== 1 &&
// rules.endsWith(...)) passes any rules.md carrying exactly one sentinel — it
// says nothing about whether the file has any SLUG TAGS at all. If every
// backtick-wrapped `inv-...`/`floor-...` tag is gone but the sentinel survives,
// slugsFromRules returns an empty array and every row in the project's
// rules.toml names an unknown slug. That once quarantined all sixteen rows and
// still emitted hookSpecificOutput with exit 0 — an ungoverned session that
// LOOKED governed — and under the classifier it would switch nothing off and
// warn about every row, a session governed by rules the model cannot identify.
// staleness.sh names this failure "no-slugs-in-payload". This pins the Codex
// guard beside a project file carrying a row;
// TestCodexRejectsAnEmptyDerivedSlugSet pins it beside one carrying none.
func TestCodexRefusesAnEmptySlugSetWhateverTheRows(t *testing.T) {
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
	// One row, small enough that no other guard (the byte budget above all) can
	// catch the mutation by coincidence and mask whether this one fired.
	minimal := "strictness  = \"firm\"\n\n[rules]\ninv-minimal-first         = { active = true }\n"
	rulesTomlPath := filepath.Join(project, ".trellis", "rules.toml")
	if err := os.WriteFile(rulesTomlPath, []byte(minimal), 0o644); err != nil {
		t.Fatal(err)
	}

	raw, got := runCodexHook(t, writeCodexPluginRoot(t), startupInput(t, project))
	if got.HookSpecificOutput != nil {
		t.Errorf("an empty derived slug set must fail closed, never deliver:\n%s", raw)
	}
	if !strings.Contains(raw, "no-slugs-in-payload") {
		t.Errorf("want no-slugs-in-payload, got:\n%s", raw)
	}
}

// guards spec-0007@v1 R11-R16, R26, R35, S3-S5, S7, S17, S23
func TestCodexBootstrapPayloadContract(t *testing.T) {
	block := payloadFile(t, "block-codex.md")
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
		// TRL-97: the two retired top-level keys are ignored without comment, so
		// the fallback names neither; naming one invites the agent to act on it.
		"strictness",
		"seeded_from",
		// Nothing repairs the file and nothing is set aside inside it any more.
		"reconcil",
		"commented out",
	} {
		if strings.Contains(block, forbidden) {
			t.Errorf("block-codex.md embeds forbidden rule/row/posture content %q", forbidden)
		}
	}
	// TRL-97: the bootstrap must teach the opt-out-only row semantics both hooks
	// now deliver. Only a row set to false switches a rule off, a rule with no
	// row applies, a bad entry costs that entry and never the file, and nothing
	// rewrites the file. The bootstrap is the fallback the agent follows when no
	// hook ran, so prose that still refused a mismatched file, or repaired it,
	// tells that agent to do what the hook beside it no longer does. TRL-31 fixed
	// that defect once for reconciliation; this change retires reconciliation and
	// moves the fallback with it.
	//
	// Pinned as WORDING, not behaviour, deliberately: this artifact has no
	// runtime, so its only enforceable contract is the bytes it ships
	// (decision-0053 — the tested wording is the shipped wording). The forbidden
	// half matters as much as the required half: without it a retired predicate
	// can come back one careless regeneration later, which is exactly how the
	// all-or-nothing one survived two decision records.
	// Every phrase here is a semantic the fallback agent gets wrong without it:
	//
	//   "A single top-level `governed = false` ... not a row set", and its
	//   "no rule applies including the two floor rules"
	//                                   the Codex hook emits NOTHING on the opt-out,
	//                                   which hands the file to this bootstrap. Read as
	//                                   a row set it holds no `false` row, so every rule
	//                                   would apply and a project that declined Trellis
	//                                   would be governed. Pinned as the WHOLE clause:
	//                                   "single top-level" and the floors half are each
	//                                   load-bearing and were each droppable while a
	//                                   shorter pin stayed green
	//   "Only a row whose boolean is `false` ..." and "a canonical slug with no row applies"
	//                                   the activation rule itself; the retired
	//                                   predicate wanted an active row for every rule
	//   "a rule with any `false` row is off, whatever its other rows say" and
	//   "a repeated row for a slug is not an error"
	//                                   any false row wins on both hooks, so without
	//                                   it a later true row can read as switching a
	//                                   deliberate disable back on, and a repeated
	//                                   row can read as one more entry to ignore
	//   "a `false` row for either floor rule" and
	//   "a row naming a slug not in the list below, which another plugin version may ship"
	//                                   the two ignored rows a project writes on
	//                                   purpose; both hooks ignore each with a warning,
	//                                   so the fallback must honour neither
	//   "lowercase letters and hyphens only" / "followed by `= { active = <boolean> }`"
	//                                   the row shape, written with a placeholder so the
	//                                   forbidden literals above stay out of the block
	//   "no entry in it makes the file invalid"
	//                                   a bad entry costs that entry, never the session
	//   "never edit or rewrite the file"
	//                                   no host rewrites the file any more, and this
	//                                   block is scanned by neither
	//                                   TestEveryDestructiveInstructionIsGated nor
	//                                   TestEveryDeletionInstructionIsGated (both read
	//                                   only the two hook files), so a write instruction
	//                                   landing here would be ungated
	for _, required := range []string{
		"A single top-level `governed = false` is an opt-out, not a row set",
		"no rule applies including the two floor rules",
		"Only a row whose boolean is `false` switches its rule off",
		"a canonical slug with no row applies",
		"a rule with any `false` row is off, whatever its other rows say",
		"a repeated row for a slug is not an error",
		"a `false` row for either floor rule",
		"a row naming a slug not in the list below, which another plugin version may ship",
		"lowercase letters and hyphens only",
		"followed by `= { active = <boolean> }`",
		"no entry in it makes the file invalid",
		"never edit or rewrite the file",
		"An ignored row is not a failure to load",
		"tell the user which rows you ignored, and why",
	} {
		if !strings.Contains(block, required) {
			t.Errorf("block-codex.md missing activation phrase %q — it must describe the opt-out-only row semantics (TRL-97), not refusal or repair", required)
		}
	}
	// Scoped to the ROW-SET predicate on purpose. A bare "occurs exactly once"
	// would also fire on correct future wording about the terminal sentinel,
	// which item 1 legitimately describes as single-occurrence.
	for _, retired := range []string{
		"Activation TOML is complete only when",
		"slug below occurs exactly once",
		"no unknown or duplicate slug",
		"complete activation predicate",
		"genuine syntax fault",
		// Retired with first-row-decides: any false row wins (TRL-97, Q4).
		"the first row for a slug decides",
		"a later row for a slug that already has one",
	} {
		if strings.Contains(block, retired) {
			t.Errorf("block-codex.md carries a retired activation predicate %q — a bad entry costs that entry, never the file (TRL-97)", retired)
		}
	}
	slugs := append([]string(nil), assessableSlugs...)
	sort.Strings(slugs)
	for _, slug := range slugs {
		if n := strings.Count(block, slug); n != 1 {
			t.Errorf("bootstrap must carry canonical slug %s exactly once, got %d", slug, n)
		}
	}
	header, rulesBody := payloadFile(t, "trellis.md"), payloadFile(t, "rules.md")
	if strings.Contains(header, rulesLoadedSentinel) ||
		strings.Contains(payloadFile(t, "block-claude.md"), rulesLoadedSentinel) {
		t.Error("sentinel belongs only at the terminal line of rules.md")
	}
	if !strings.HasSuffix(rulesBody, rulesLoadedSentinel+"\n") ||
		strings.Count(rulesBody, rulesLoadedSentinel) != 1 {
		t.Error("rules.md must end with exactly one completion sentinel")
	}
	if !strings.Contains(header, "@rules.md\n---\n"+invariantsTrigger) {
		t.Error("trellis.md must carry the fixed post-import footer")
	}
}

// TRL-10. The bootstrap's delivery boundary must be expressed in markers the
// GENERATOR writes, never in payload prose.
//
// This is the same defect #212 fixed on the Claude side, still live here. There,
// the hook decided whether `.claude/rules/trellis.md` had been rendered by
// matching the invariants sentence, the posture note and the activation heading —
// all payload text an editor may legitimately reword. One reword flipped the
// predicate, every freshly installed project got a permanent false "not governed"
// warning, and the whole suite stayed green. install.sh:888-893 says so in a
// comment and prints `<!-- trellis:rendered-begin -->` / `<!-- trellis:rendered-footer -->`
// instead; staleness.sh:619-621 matches those as whole lines.
//
// block-codex.md carried the old shape: generated prose counted as delivered only
// when the sentinel was followed by "the fixed footer whose first nonblank line is
// `---` and whose next text is the ambiguity/fallback sentence" — two landmarks in
// renderHeader's tail. The reader (a Codex agent following this block) and the
// writer (renderHeader) shared no contract. So renderHeader now terminates the
// header with `<!-- trellis:prose-complete -->` and the block keys on that.
//
// On decision-0053, stated precisely rather than conceded loosely: its "tested
// wording is the shipped wording" bullet is a CONTEXT bullet (0053:51; ## Context
// opens at :28, ## Decision at :57) and it names three validated artifacts — "a
// specific authority header, a rows-inlined-below-the-rules layout for the inline
// channel, and live-rows seed comments (`header_arm_toml`)". renderHeader's tail
// is not among them, and could not be: the experiment put trellis-a.md ON DISK
// (eval/experiments/annotation-vs-absence/run.sh:113) while assembling the tested
// context from block-inline-a-head.md + the readout + rows + header_arm_tail
// (run.sh:128-138), so these bytes were never in it. The divergence proves it —
// 0053's tail transform landed on block-inline-tail.md, which today carries a
// live-rows clause invariantsTrigger does not. The pin does not reach this file;
// no validated byte moves here, and the marker lands after even the untested tail.
//
// Three parts, because the defect has three faces and a pin on one alone lets the
// other two come back:
//
//	(a) every literal item 1 tells the agent to match is writer-owned, AND both
//	    markers are still named — one marker alone proves only that delivery
//	    started, since the sentinel ends the imported rules.md and so sits
//	    mid-document once expanded;
//	(b) no UNBACKTICKED prose landmark either — item 1's worst dependency ("the
//	    ambiguity/fallback sentence") named prose without quoting it, so (a)
//	    alone would have scored the old wording as one violation, not two. Note
//	    honestly that (b) is a BLACKLIST, against (a)'s whitelist: unbackticked
//	    prose cannot be enumerated, so (b) catches the three landmarks the old
//	    predicate used and cannot catch one nobody has written yet;
//	(c) the property (a) and (b) exist to buy, exercised against the real
//	    delivered payload rather than against the block in isolation. With (a)
//	    holding, the landmark set is the two markers and the footer reword
//	    cannot touch them — so (c) passes by construction TODAY and is a
//	    regression net, not the primary guard: it goes red if the delivery stops
//	    carrying a landmark at all (rules.md losing its sentinel, the header
//	    losing its marker, @rules.md ceasing to expand).
//
// The writer's half is pinned here too. A reader keyed on a marker no writer
// emits is the same blackout with the blame reversed.
func TestCodexBootstrapBoundaryIsMachineOwned(t *testing.T) {
	block := payloadFile(t, "block-codex.md")
	headerProse, rulesBody := payloadFile(t, "trellis.md"), payloadFile(t, "rules.md")
	item := codexAssessmentItemOne(t, block)

	// (a) Whitelist, not blacklist. A blacklist of known-bad landmarks passes
	// every landmark nobody thought of, which is how prose validation survives.
	writerOwned := map[string]string{
		rulesLoadedSentinel: "the terminal line of generated rules.md",
		proseCompleteMarker: "the terminal line of the generated header",
	}
	landmarks := backtickedLiterals(item)
	if len(landmarks) == 0 {
		t.Fatalf("premise: assessment item 1 quotes no landmark at all, so this test proves nothing:\n%s", item)
	}
	for _, landmark := range landmarks {
		if _, ok := writerOwned[landmark]; !ok {
			t.Errorf("assessment item 1 keys the completeness boundary on %q, which no generator writes as a marker — payload prose may be reworded and the predicate moves with it (TRL-10)", landmark)
		}
	}

	// The OTHER direction, and it is not redundant. The loop above proves every
	// landmark item 1 names is writer-owned; it says nothing about item 1 naming
	// FEWER. An item 1 rewritten to the sentinel alone quotes none but writer-owned
	// literals, so it passes that loop, passes the prose check below (it names no
	// prose), and passes the reword case (the sentinel survives any reword) — the
	// whole guard green over the predicate this block itself calls insufficient:
	// "A sentinel alone, an end marker alone, ... is not completion". The sentinel
	// ends the imported rules.md and so sits MID-document once expanded; alone it
	// proves only that delivery started, which is the half TRL-10 exists to close.
	for want, role := range writerOwned {
		if !slices.Contains(landmarks, want) {
			t.Errorf("assessment item 1 no longer names %s (%s) — with one marker the predicate proves only that delivery started, not that the tail arrived (TRL-10)", want, role)
		}
	}

	// Named in item 1 and NOWHERE else, which pins the deliberate asymmetry at
	// renderCodexBootstrap: the four-inputs paragraph validates the FILES and is
	// left alone on purpose. An overlay vendored before this release carries no
	// end marker, so requiring one there would turn a stale-but-correct install
	// into a false "Trellis was not loaded" — the very failure this fixes. That
	// carve-out lives only in a comment today; a later "tighten it up" would
	// silently orphan every old overlay, and this is what goes red when it does.
	if n := strings.Count(block, proseCompleteMarker); n != 1 {
		t.Errorf("block-codex.md names %s %d times; it belongs to assessment item 1 alone — the file-level test must keep validating an overlay vendored before this release (TRL-10)", proseCompleteMarker, n)
	}

	// (b) The unquoted half. Scoped to the WHOLE BLOCK on purpose, not to item 1
	// like (a) and the reverse loop: a prose landmark that migrates out of item 1
	// into the fallback table or the four-inputs paragraph is the same defect at
	// a different address, and matching `item` would let it walk there. The cost
	// of the wider net is a misleading message if a future item 2 ever wants one
	// of these three phrases for TOML shape rather than for the boundary — strict,
	// never loose, which is the right way round for this one.
	for _, prose := range []string{"fixed footer", "first nonblank line", "ambiguity/fallback sentence"} {
		if strings.Contains(block, prose) {
			t.Errorf("block-codex.md still identifies the delivery boundary by the prose landmark %q — a reworded payload then fails a correct delivery, or passes a wrong one (TRL-10, mirroring #212)", prose)
		}
	}

	// (c) The generated prose a Codex agent actually receives: the header with
	// its sibling expanded, exactly what codex-context.mjs injects
	// (`trellis.replace("@rules.md", rules)`).
	delivered := strings.Replace(headerProse, "@rules.md", rulesBody, 1)
	if delivered == headerProse {
		t.Fatal("premise: the header carries no @rules.md expansion point, so nothing was delivered")
	}
	// A legitimate editorial reword of the footer: same meaning, same pointer,
	// same delivery. Both landmarks the retired predicate named are in here —
	// the sentence, and the horizontal rule above it (`---` occurs exactly once
	// in the header and never in rules.md, so this genuinely removes it).
	if n := strings.Count(delivered, "\n---\n"); n != 1 {
		t.Fatalf("premise: the reword below assumes the header's `---` is the ONLY one in the delivered prose, so that replacing the first occurrence removes the landmark it names; got %d. A `---` reaching rules.md would send the replacement to the wrong line and this case would stop exercising the reword it claims to", n)
	}
	reworded := strings.Replace(delivered, invariantsTrigger,
		"If a rule is unclear, or pulls against this project's own instructions, read its entry in `.trellis/internal/invariants.md` — the description and with/without examples — before you deviate.", 1)
	reworded = strings.Replace(reworded, "\n---\n", "\n***\n", 1)
	if reworded == delivered {
		t.Fatal("premise: the reword changed nothing, so the case would pass vacuously")
	}
	for _, landmark := range landmarks {
		if !strings.Contains(delivered, landmark) {
			t.Errorf("assessment item 1 tells the agent to match %q, which the delivered prose does not contain at all", landmark)
		}
		if !strings.Contains(reworded, landmark) {
			t.Errorf("rewording the footer prose removed %q, a landmark item 1 keys on — delivery is unchanged, so the agent would report a correctly delivered payload as \"Trellis was not loaded\" (TRL-10)", landmark)
		}
	}

	// The writer's half.
	header := headerProse
	if !strings.HasSuffix(header, "\n"+proseCompleteMarker+"\n") {
		t.Errorf("trellis.md must END with %s — the marker proves the header's tail arrived, which it cannot do from the middle: %q", proseCompleteMarker, header)
	}
	if n := strings.Count(header, proseCompleteMarker); n != 1 {
		t.Errorf("trellis.md must carry %s exactly once, got %d — a second copy makes the boundary ambiguous", proseCompleteMarker, n)
	}
	// rules.md has its own terminator; two end markers in one delivered prose
	// would let a truncation between them pass.
	if strings.Contains(rulesBody, proseCompleteMarker) {
		t.Errorf("rules.md must not carry %s — it ends at the sentinel, and the header ends after it", proseCompleteMarker)
	}
}

// codexAssessmentItemOne returns block-codex.md's numbered assessment item 1 —
// the sentence that defines when generated prose counts as delivered complete.
// Extracted from the shipped block rather than restated here, so the test reads
// the same bytes the Codex agent does (decision-0053: this artifact has no
// runtime, so its bytes are its only contract).
func codexAssessmentItemOne(t *testing.T, block string) string {
	t.Helper()
	start := strings.Index(block, "\n1. ")
	end := strings.Index(block, "\n2. ")
	if start < 0 || end <= start {
		t.Fatalf("block-codex.md no longer carries a numbered assessment item 1 followed by item 2; this test cannot locate the completeness predicate:\n%s", block)
	}
	return block[start+1 : end]
}

var backtickedLiteralRe = regexp.MustCompile("`([^`]+)`")

// backtickedLiterals returns every backtick-quoted literal in s, in order. In
// block-codex.md a backticked string is an EXACT string the agent is told to
// match, so this is the set of landmarks a predicate depends on.
func backtickedLiterals(s string) []string {
	var out []string
	for _, m := range backtickedLiteralRe.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
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
	// The action's major version is not what this guards: Node 20 installed before
	// the Go tests is (spec-0007 R40, the Codex hook needs Node >=20). Matching any
	// setup-node release keeps a dependabot bump from turning this red.
	setupNode := strings.Index(workflow, "uses: actions/setup-node@")
	node20 := strings.Index(workflow, `node-version: "20"`)
	goTests := strings.Index(workflow, "run: go test -count=1 -coverprofile=")
	if setupNode < 0 || node20 < setupNode || goTests < node20 {
		t.Errorf("cli-ci must install Node.js 20 with actions/setup-node before Go tests execute the Codex hook")
	}
	// -count=1 is matched, not just `go test`, because the cache hazard this
	// suite lives inside is real: these tests execute the production hooks as
	// EXTERNAL files, which Go's test cache cannot see, so a hook mutation with
	// no .go change replays a stale `ok (cached)`. Reproduced by deleting
	// codex-context.mjs's empty-slug-set guard. Pinned here so dropping the flag
	// from the workflow is a red test rather than a silent loss of coverage.
	// -coverprofile= is pinned with it: the coverage-floor step consumes that
	// profile, so dropping the flag must be red here rather than a silent loss.
	if !strings.Contains(workflow, "run: go test -count=1 -coverprofile=") {
		t.Error("cli-ci must run tests with -count=1 and a coverprofile — the hook tests exec external files the Go test cache does not track, and the coverage floor consumes the profile")
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
	if got := run(t, "[rules]\n"); !strings.Contains(got, "inv-directional-flow") {
		t.Fatalf("control failed — the Codex hook delivered no rules for a normal project, so this test cannot detect the opt-out:\n%s", got)
	}
	if got := run(t, "governed = false\n"); strings.TrimSpace(got) != "" {
		t.Errorf("decision-0070 D5: a project declaring governed = false must get nothing on Codex either; got:\n%s", got)
	}
}

// TestCodexToleratesADuplicateSlugTagInRulesMd — round-1 fix 2. A rules.md that
// ever tagged one slug twice once made every Codex project read
// .trellis/rules.toml: invalid-rules, because the old parser's completeness check
// counted the derived slugs as a list, while Claude — whose own want[] in
// staleness.sh is a set — kept governing from the identical file. The derived
// set is de-duplicated at the source, and TRL-97's classifier has no
// completeness check left, but a duplicated tag must still cost nothing. Not
// reachable with the current payload (every one of the sixteen tags occurs
// exactly once); this guards the shape directly, by duplicating one tag line in
// the payload rules.md itself.
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
	if got.HookSpecificOutput == nil || warningsBesideVendoredInvariants(t, got.SystemMessage) != "" {
		t.Fatalf("a duplicated slug tag in rules.md must not fail every row closed: %s", raw)
	}
}

// The Codex hook validated rows against a hardcoded 16-slug array while the
// Claude hook derived its set from the shipped rules.md, and nothing in CI
// compared the two. A payload upgrade therefore could not repair drift on
// Codex — worse, a stale array made an `unknown:` reason FALSE: the agent was
// told a live row named no rule, citing a payload that does ship it. Floors are
// found by their `floor-` prefix for the same reason (TRL-97).
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
// checks the marker, not the tags. With `slugs` empty, the old parser's two
// completeness checks passed VACUOUSLY, so a config holding nothing but
// `strictness` and an empty `[rules]` table was ACCEPTED and the hook emitted a
// successful "loaded installed overlay" response with zero activation rows —
// a silently ungoverned session at exit 0, on the host where success is what it
// looks like. TRL-97's classifier would accept the same file as switching
// nothing off. The refusal must fire before anything consumes `slugs`.
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
	// The shape with no row at all, so not even an unknown-slug warning could
	// hint that something is wrong. This is the silent one.
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
	firm := legacyFirmRulesToml
	// A file an older release reconciled and an agent wrote back: an added-rows
	// header and a quarantined row with its note. TRL-97 retired the writer, and
	// files carrying its output stay in consumer repositories.
	alreadyQuarantined := firm +
		"# added 1 row(s) below on 2026-01-01 (missing from payload@000000000000)\n" +
		"# inv-foreign-rule-a = { active = true }  # quarantined 2026-01-01: not in payload@000000000000." +
		" If a newer Trellis release ships this slug, update the Trellis plugin and uncomment this row.\n"

	cases := map[string]string{
		"legacy firm preset":     firm,
		"legacy adaptive preset": replaceOnce(t, firm, `strictness  = "firm"`, `strictness  = "adaptive"`),
		"sparse file":            configOnlyProjectRules,
		"empty file":             "",
		"missing row":            replaceOnce(t, firm, "inv-minimal-first         = { active = true }\n", ""),
		"unknown row":            firm + "inv-foreign-rule-a = { active = true }\n",
		"duplicate row":          firm + "inv-minimal-first         = { active = false }\n",
		"renamed row": replaceOnce(t, firm,
			"inv-minimal-first         = { active = true }",
			"inv-renamed-first         = { active = true }"),
		"already quarantined":  alreadyQuarantined,
		"hand-written partial": "strictness  = \"firm\"\n",
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
			if n := len([]byte(ctx)); n > codexContextCap {
				t.Errorf("%s assembled to %d bytes, over the cap", name, n)
			}
			if !strings.Contains(ctx, rulesLoadedSentinel) || !strings.Contains(ctx, "\n"+activationHeading+"\n\n") {
				t.Errorf("the rules and the activation section were not delivered for a legitimate shape:\n%s", ctx)
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

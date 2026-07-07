package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestStalenessHook exercises the plugin's SessionStart staleness hook (decision-0039):
// it stays silent unless a plugin-generated overlay is behind the installed plugin, and
// then emits a valid {"additionalContext": …} refresh nudge. Runs the real shell script
// against a temp "plugin root" git repo at a known HEAD.
func TestStalenessHook(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(hook); err != nil {
		t.Fatalf("hook script missing: %v", err)
	}

	pluginRoot := t.TempDir()
	git := func(args ...string) string {
		out, err := exec.Command("git", append([]string{"-C", pluginRoot}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("config", "user.email", "t@example.com")
	git("config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(pluginRoot, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-qm", "init")
	head := git("rev-parse", "--short", "HEAD")

	run := func(t *testing.T, stamp string, writeStamp bool) string {
		proj := t.TempDir()
		if writeStamp {
			if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(proj, ".trellis", "version"), []byte(stamp+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v) — a hook must never fail the session: %s", err, out)
		}
		return strings.TrimSpace(string(out))
	}

	t.Run("no overlay is silent", func(t *testing.T) {
		if out := run(t, "", false); out != "" {
			t.Errorf("want silent, got %q", out)
		}
	})
	t.Run("current plugin stamp is silent", func(t *testing.T) {
		if out := run(t, "plugin@"+head, true); out != "" {
			t.Errorf("want silent, got %q", out)
		}
	})
	t.Run("older plugin stamp surfaces a refresh nudge", func(t *testing.T) {
		out := run(t, "plugin@0000000", true)
		if out == "" {
			t.Fatal("want a staleness message, got silence")
		}
		var v struct {
			AdditionalContext string `json:"additionalContext"`
		}
		if err := json.Unmarshal([]byte(out), &v); err != nil {
			t.Fatalf("output is not valid JSON: %v (%q)", err, out)
		}
		if !strings.Contains(v.AdditionalContext, "/trellis:setup") {
			t.Errorf("message should point at /trellis:setup: %q", v.AdditionalContext)
		}
	})
	t.Run("plugin@unknown is silent", func(t *testing.T) {
		if out := run(t, "plugin@unknown", true); out != "" {
			t.Errorf("want silent, got %q", out)
		}
	})
	t.Run("CLI-version stamp is silent", func(t *testing.T) {
		if out := run(t, "0.2.16", true); out != "" {
			t.Errorf("want silent (trellis status is the CLI surface), got %q", out)
		}
	})
}

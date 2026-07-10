package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestStalenessHook exercises the plugin's SessionStart staleness hook — decision-0039
// rule 1 (the surface is a SessionStart hook emitting additionalContext) as reworked by
// decision-0043 / kodhama-0007 slice 4 (#120): the check is a binary-free, git-free
// file-to-file comparison of the project's .trellis/version stamp against the installed
// plugin's ${CLAUDE_PLUGIN_ROOT}/reference/version payload stamp. Silent when they
// match (or nothing is comparable); a refresh nudge when they differ. With `trellis
// status` retired (#120), this hook is the only user-facing drift surface, so legacy
// stamps (plugin@<sha>, CLI semver) now draw the nudge too instead of deferring to the
// binary — decision-0035's "drift is made visible, not silent" must survive the
// binary's retirement. Runs the real shell script against a temp "plugin root".
func TestStalenessHook(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(hook); err != nil {
		t.Fatalf("hook script missing: %v", err)
	}

	// The payload stamp the plugin actually ships (kept current by
	// TestVendoredPayloadIsCurrent); the hook compares against this file.
	shipped := payloadFiles()["version"]
	current := strings.TrimSpace(shipped)

	// A plain-directory plugin root — no git repo, on purpose: the file compare
	// must not depend on the plugin being a git checkout (decision-0043).
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "reference"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginRoot, "reference", "version"), []byte(shipped), 0o644); err != nil {
		t.Fatal(err)
	}

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

	nudge := func(t *testing.T, out string) string {
		t.Helper()
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
		return v.AdditionalContext
	}

	t.Run("no overlay is silent", func(t *testing.T) {
		if out := run(t, "", false); out != "" {
			t.Errorf("want silent, got %q", out)
		}
	})
	t.Run("current payload stamp is silent", func(t *testing.T) {
		if out := run(t, current, true); out != "" {
			t.Errorf("want silent, got %q", out)
		}
	})
	t.Run("older payload stamp surfaces a refresh nudge", func(t *testing.T) {
		msg := nudge(t, run(t, "payload@000000000000", true))
		if !strings.Contains(msg, "payload@000000000000") || !strings.Contains(msg, current) {
			t.Errorf("message should name both stamps: %q", msg)
		}
	})
	t.Run("legacy plugin@sha stamp surfaces a refresh nudge", func(t *testing.T) {
		// Pre-#120 installs are stamped plugin@<short-sha> (decision-0039 rule 2,
		// superseded in part by decision-0043); they differ from the payload stamp,
		// so the one-time nudge migrates them onto the payload vocabulary.
		nudge(t, run(t, "plugin@0000000", true))
	})
	t.Run("legacy CLI-version stamp surfaces a refresh nudge", func(t *testing.T) {
		// Pre-#120 CLI installs are stamped with the binary's semver; their old
		// surface (`trellis status`) retired with the binary, so the hook is now
		// their surface too (decision-0043 — visibility survives the retirement).
		nudge(t, run(t, "0.2.16", true))
	})
	t.Run("empty stamp is silent", func(t *testing.T) {
		if out := run(t, "", true); out != "" {
			t.Errorf("want silent (nothing to compare), got %q", out)
		}
	})
	t.Run("unreadable plugin reference is silent", func(t *testing.T) {
		proj := t.TempDir()
		if err := os.MkdirAll(filepath.Join(proj, ".trellis"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, ".trellis", "version"), []byte("payload@abc\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+t.TempDir())
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v): %s", err, out)
		}
		if strings.TrimSpace(string(out)) != "" {
			t.Errorf("want silent (can't read the installed payload stamp), got %q", out)
		}
	})
}

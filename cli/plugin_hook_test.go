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
// decision-0043 / kodhama-0007 slice 4 (#120), with the compared path moved by
// decision-0051 (the authority split): the check is a binary-free, git-free
// file-to-file comparison of the project's .trellis/internal/version stamp against the
// installed plugin's ${CLAUDE_PLUGIN_ROOT}/reference/version payload stamp. Silent
// when they match (or nothing is comparable); a refresh nudge when they differ. A
// stamp found only at the legacy flat path (.trellis/version — pre-decision-0051
// layouts, and before them the plugin@<sha>/CLI-semver stamps) always draws the
// nudge: the layout itself is stale, and a refresh is the migration vehicle. With
// `trellis status` retired (#120), this hook is the only user-facing drift surface
// (decision-0035: drift is made visible, not silent). Runs the real shell script
// against a temp "plugin root".
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

	// run executes the hook in a fresh project dir; stampRel names where the stamp
	// file is written (".trellis/internal/version", the legacy ".trellis/version",
	// or "" for no overlay at all).
	run := func(t *testing.T, stampRel, stamp string) string {
		proj := t.TempDir()
		if stampRel != "" {
			p := filepath.Join(proj, filepath.FromSlash(stampRel))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(stamp+"\n"), 0o644); err != nil {
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
		if out := run(t, "", ""); out != "" {
			t.Errorf("want silent, got %q", out)
		}
	})
	t.Run("current stamp at internal/version is silent", func(t *testing.T) {
		if out := run(t, ".trellis/internal/version", current); out != "" {
			t.Errorf("want silent, got %q", out)
		}
	})
	t.Run("older stamp at internal/version surfaces a refresh nudge", func(t *testing.T) {
		msg := nudge(t, run(t, ".trellis/internal/version", "payload@000000000000"))
		if !strings.Contains(msg, "payload@000000000000") || !strings.Contains(msg, current) {
			t.Errorf("message should name both stamps: %q", msg)
		}
	})
	t.Run("legacy flat-layout stamp surfaces a migration nudge", func(t *testing.T) {
		// A stamp at the pre-decision-0051 path means the overlay itself predates
		// the internal/ layout — the nudge fires even if the stamp text happens to
		// match the shipped payload, because the layout is what's stale.
		nudge(t, run(t, ".trellis/version", current))
		nudge(t, run(t, ".trellis/version", "payload@000000000000"))
	})
	t.Run("legacy plugin@sha stamp surfaces a refresh nudge", func(t *testing.T) {
		// Pre-#120 installs are stamped plugin@<short-sha> (decision-0039 rule 2,
		// superseded in part by decision-0043); they sit at the flat path, so the
		// one-time nudge migrates them onto the payload vocabulary and layout.
		nudge(t, run(t, ".trellis/version", "plugin@0000000"))
	})
	t.Run("legacy CLI-version stamp surfaces a refresh nudge", func(t *testing.T) {
		nudge(t, run(t, ".trellis/version", "0.2.16"))
	})
	t.Run("internal stamp wins over a leftover legacy file", func(t *testing.T) {
		// Mid-migration robustness: when both exist, the new layout's stamp is the
		// one compared — a current internal/version stays silent even if the old
		// flat file was not cleaned up.
		proj := t.TempDir()
		for rel, stamp := range map[string]string{
			".trellis/internal/version": current,
			".trellis/version":          "payload@000000000000",
		} {
			p := filepath.Join(proj, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(stamp+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(hook)
		cmd.Dir = proj
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj, "CLAUDE_PLUGIN_ROOT="+pluginRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero (%v): %s", err, out)
		}
		if strings.TrimSpace(string(out)) != "" {
			t.Errorf("want silent (internal/version is current; the leftover flat file must not fire), got %q", out)
		}
	})
	t.Run("empty stamp is silent", func(t *testing.T) {
		if out := run(t, ".trellis/internal/version", ""); out != "" {
			t.Errorf("want silent (nothing to compare), got %q", out)
		}
		if out := run(t, ".trellis/version", ""); out != "" {
			t.Errorf("want silent (empty legacy stamp), got %q", out)
		}
	})
	t.Run("unreadable plugin reference is silent", func(t *testing.T) {
		proj := t.TempDir()
		p := filepath.Join(proj, ".trellis", "internal", "version")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("payload@abc\n"), 0o644); err != nil {
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

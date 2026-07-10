package main

import (
	"bytes"
	"strings"
	"testing"
)

// run2 invokes run with the given stdin and args, returning captured stdout + error.
func run2(stdin string, args ...string) (string, error) {
	var buf bytes.Buffer
	err := run(strings.NewReader(stdin), &buf, args)
	return buf.String(), err
}

func TestRunVersion(t *testing.T) {
	out, err := run2("", "version")
	if err != nil {
		t.Fatalf("version returned error: %v", err)
	}
	if !strings.Contains(out, "trellis ") {
		t.Errorf("version output = %q, want it to contain %q", out, "trellis ")
	}
}

func TestRunHelpAndNoArgs(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}} {
		out, err := run2("", args...)
		if err != nil {
			t.Fatalf("run(%v) returned error: %v", args, err)
		}
		if !strings.Contains(out, "trellis payload") {
			t.Errorf("run(%v) usage did not mention the payload command: %q", args, out)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	if _, err := run2("", "nope"); err == nil {
		t.Fatal("expected an error for an unknown command, got nil")
	}
}

// TestGeneratorOnlyCommandSurface encodes decision-0043 (kodhama-0007 slice 4, #120):
// the Go code survives as the release-time payload generator ONLY. The end-user
// commands — setup, status, remove, uninstall — retired with the binary channel;
// their live homes are the plugin skills (/trellis:setup, /trellis:remove), the
// bundled staleness hook, and the documented manual copy path. A retired command
// must be an unknown-command error, and the usage text must not advertise it.
func TestGeneratorOnlyCommandSurface(t *testing.T) {
	for _, retired := range []string{"setup", "status", "remove", "uninstall"} {
		if _, err := run2("", retired); err == nil {
			t.Errorf("`trellis %s` retired with the binary channel (#120) — it must be an unknown command", retired)
		}
	}
	usage, err := run2("", "help")
	if err != nil {
		t.Fatal(err)
	}
	for _, retired := range []string{"setup", "status", "remove", "uninstall"} {
		if strings.Contains(usage, "trellis "+retired) {
			t.Errorf("usage still advertises retired command `trellis %s`: %q", retired, usage)
		}
	}
}

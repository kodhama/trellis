package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestCheckTodosMarkerForms runs scripts/check-todos.sh over one tracked sample
// file per debt marker. Every marker form the script's error message and
// README.md show must pass, so the filter and those two copies of its accepted
// forms stay in step (decision-0028's guard for that pair); malformed record ids
// must fail. The markers are assembled at run time because check-todos.sh scans
// this file too.
func TestCheckTodosMarkerForms(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "scripts", "check-todos.sh"))
	if err != nil {
		t.Fatal(err)
	}
	todo, fixme := "TO"+"DO", "FIX"+"ME"
	documented := regexp.MustCompile(todo + `\([^)]*\)`)

	for _, src := range []string{script, filepath.Join("..", "README.md")} {
		examples := documented.FindAllString(readFileT(t, src), -1)
		if len(examples) == 0 {
			t.Fatalf("%s shows no %s(...) marker form", src, todo)
		}
		for _, marker := range examples {
			if out, err := runCheckTodos(t, script, marker); err != nil {
				t.Errorf("%s shows %s, which check-todos.sh rejects: %v\n%s", src, marker, err, out)
			}
		}
	}

	accepted := fixme + "(decision-2026-09-20-rules-toml-is-optional)"
	if out, err := runCheckTodos(t, script, accepted); err != nil {
		t.Errorf("check-todos.sh rejects %s: %v\n%s", accepted, err, out)
	}

	for _, ref := range []string{"", "(decision-)", "(decision-2026-09-20-)", "(decision-2026-09-20-Rules)", "(decision-2026-9-20-rules)"} {
		marker := todo + ref
		out, err := runCheckTodos(t, script, marker)
		if err == nil {
			t.Errorf("check-todos.sh accepts %s", marker)
		} else if !strings.Contains(out, marker) {
			t.Errorf("check-todos.sh failed on %s without naming it: %v\n%s", marker, err, out)
		}
	}
}

// runCheckTodos runs check-todos.sh in a new repository whose only tracked file
// is a comment holding marker, and returns the script's combined output.
func runCheckTodos(t *testing.T, script, marker string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	initGitRepo(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "sample.go"), []byte("// "+marker+"\n"), 0o644); err != nil {
		t.Fatalf("writing sample: %v", err)
	}
	add := exec.Command("git", "add", "sample.go")
	add.Dir = dir
	if out, err := add.CombinedOutput(); err != nil {
		t.Fatalf("git add in %s: %v: %s", dir, err, out)
	}
	cmd := exec.Command("/bin/sh", script)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// executablePath is indirected so tests can point uninstall at a temp file instead
// of the real test binary.
var executablePath = os.Executable

// uninstall removes the trellis binary itself (spec-0004 §1). It does not touch any
// project's .trellis/ — that is `remove`'s job. On Unix, unlinking a running binary
// is safe (the inode frees when the process exits).
func uninstall(in io.Reader, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fs.SetOutput(w)
	yes := fs.Bool("yes", false, "skip the confirmation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	path, err := executablePath()
	if err != nil {
		return fmt.Errorf("cannot locate the trellis binary: %w", err)
	}

	// If Homebrew manages this binary, deleting it by hand leaves brew inconsistent
	// (metadata says installed, file gone). Defer to `brew uninstall` (decision-0032).
	if homebrewManaged(path) {
		fmt.Fprintf(w, "trellis here was installed with Homebrew (%s).\n", path)
		fmt.Fprintln(w, "Remove it cleanly with:  brew uninstall trellis")
		fmt.Fprintln(w, "(Deleting the binary by hand would leave Homebrew's records inconsistent.)")
		return nil
	}

	fmt.Fprintf(w, "This removes the trellis binary at %s.\n", path)
	fmt.Fprintln(w, "(It does not touch any project — use `trellis remove` inside a project for that.)")
	if !*yes && !askYesNo(in, w, "Remove it?") {
		fmt.Fprintln(w, "cancelled — nothing removed")
		return nil
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(w, "already gone — nothing to remove")
			return nil
		}
		return fmt.Errorf("removing %s: %w", path, err)
	}
	fmt.Fprintf(w, "removed %s\n", path)
	return nil
}

// homebrewManaged reports whether the binary at path lives under a Homebrew Cellar —
// directly, or via the `<prefix>/bin` symlink that resolves into it. Homebrew always
// installs into `.../Cellar/<formula>/<version>/`, so that path segment is the tell.
func homebrewManaged(path string) bool {
	if strings.Contains(path, "/Cellar/") {
		return true
	}
	if real, err := filepath.EvalSymlinks(path); err == nil {
		return strings.Contains(real, "/Cellar/")
	}
	return false
}

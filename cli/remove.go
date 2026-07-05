package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// remove undoes `setup` on a project (spec-0004 §2). M1 (overlay) reverses
// deterministically — delete .trellis/ and strip the block from whichever instruction
// file setup attached to (CLAUDE.md, AGENTS.md, …), leaving the rest byte-for-byte.
// M2 (morph) rewrote the project's own files and cannot be
// cleanly reversed, so remove warns loudly and reports the rollback ref recorded at
// apply time; it never mutates git history.
func remove(in io.Reader, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	fs.SetOutput(w)
	dir := fs.String("dir", ".", "project directory")
	yes := fs.Bool("yes", false, "skip the confirmation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	tdir := filepath.Join(*dir, ".trellis")
	rollbackPath := filepath.Join(tdir, "rollback")

	// The overlay may have attached to any known instruction file (decision-0029), not
	// just CLAUDE.md — strip the block from every one that carries it.
	blocked := instructionFilesWithBlock(*dir)
	overlay := fileExists(filepath.Join(tdir, "trellis.md")) || len(blocked) > 0 // M1
	morph := fileExists(rollbackPath)                                            // M2 marker

	if !overlay && !morph {
		fmt.Fprintf(w, "no Trellis install found in %q — nothing to remove\n", *dir)
		return nil
	}

	// M2: git rollback path — warn, report the ref, do not reverse.
	if morph {
		ref := strings.TrimSpace(readFileOr(rollbackPath))
		fmt.Fprintln(w, "!  This project was set up with M2 (morph): Trellis rewrote your own files.")
		fmt.Fprintln(w, "   That cannot be cleanly reversed. To roll back, use git:")
		if ref != "" {
			fmt.Fprintf(w, "     git reset --hard %s        (or: git revert)\n", ref)
		} else {
			fmt.Fprintln(w, "     (no rollback ref recorded — check `git log` and the trellis/morph branch)")
		}
		fmt.Fprintln(w, "   That is the limit of what remove can do for a morph.")
	}

	fmt.Fprintf(w, "\nWill delete %s", tdir)
	if len(blocked) > 0 {
		fmt.Fprintf(w, " and strip the Trellis block from %s", strings.Join(blocked, ", "))
	}
	fmt.Fprintln(w, ".")
	if !*yes && !askYesNo(in, w, "Proceed?") {
		fmt.Fprintln(w, "cancelled — nothing removed")
		return nil
	}

	if err := os.RemoveAll(tdir); err != nil {
		return fmt.Errorf("removing .trellis/: %w", err)
	}
	for _, name := range blocked {
		if err := stripBlockFromFile(filepath.Join(*dir, name)); err != nil {
			return err
		}
	}
	fmt.Fprintln(w, "removed the Trellis overlay.")
	return nil
}

// instructionFilesWithBlock returns the known instruction files in dir that carry the
// Trellis block (by relative name), so remove strips exactly those.
func instructionFilesWithBlock(dir string) []string {
	var out []string
	for _, f := range instructionFiles {
		if fileHasBlock(filepath.Join(dir, f.Name)) {
			out = append(out, f.Name)
		}
	}
	return out
}

// stripBlockFromFile removes the delimited Trellis block from an instruction file,
// preserving everything outside the markers. If nothing but the block remains, the file
// (which setup created) is removed; otherwise the host's content is written back.
func stripBlockFromFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading %s: %w", path, err)
	}
	out := stripBlock(string(b))
	if strings.TrimSpace(out) == "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing %s: %w", path, err)
		}
		return nil
	}
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// stripBlock is the reverse of upsertBlock: it removes the block between the markers
// (inclusive) and normalizes the surrounding blank lines. Content outside is exact.
func stripBlock(content string) string {
	i := strings.Index(content, trellisBegin)
	j := strings.Index(content, trellisEnd)
	if i < 0 || j < 0 || j < i {
		return content
	}
	before := strings.TrimRight(content[:i], "\n")
	after := strings.TrimLeft(content[j+len(trellisEnd):], "\n")
	switch {
	case before == "":
		return after
	case after == "":
		return before + "\n"
	default:
		return before + "\n\n" + after
	}
}

func fileHasBlock(path string) bool {
	b, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(b), trellisBegin)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readFileOr(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

// askYesNo reads a single confirmation line from in; empty or anything but y/yes is no.
func askYesNo(in io.Reader, w io.Writer, prompt string) bool {
	fmt.Fprintf(w, "%s [y/N]: ", prompt)
	sc := bufio.NewScanner(in)
	if !sc.Scan() {
		return false
	}
	a := strings.ToLower(strings.TrimSpace(sc.Text()))
	return a == "y" || a == "yes"
}

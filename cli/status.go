package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// status reports whether a project's Trellis overlay is current with THIS binary — the
// user-facing surface for overlay↔product drift (decision-0035). The overlay is a
// snapshot (no runtime dependency, decision-0010), so it can lag the installed tool;
// this makes that lag visible (D1) instead of silent. Read-only.
func status(in io.Reader, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(w)
	dir := fs.String("dir", ".", "project directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	verPath := filepath.Join(*dir, ".trellis", "version")
	b, err := os.ReadFile(verPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(w, "no Trellis overlay in %q — run `trellis setup` to install one.\n", *dir)
			return nil
		}
		return fmt.Errorf("reading %s: %w", verPath, err)
	}
	overlayVer := strings.TrimSpace(string(b))

	fmt.Fprintf(w, "Trellis overlay in %q\n", *dir)
	fmt.Fprintf(w, "  generated with: trellis %s\n", overlayVer)
	fmt.Fprintf(w, "  this binary:    trellis %s\n", version)
	if overlayVer == version {
		fmt.Fprintln(w, "  → current.")
	} else {
		fmt.Fprintln(w, "  → the overlay differs from this binary — re-run `trellis setup` to refresh it.")
	}
	return nil
}

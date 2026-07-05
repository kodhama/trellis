// Command trellis is the setup CLI for the Trellis governance layer.
//
// It is *setup tooling, not a runtime* (decision-0010): you run it once to pick an
// install mode, detect what that mode needs (a harness only for the M2 rewrite), pick
// an expression profile, and compose Trellis onto the project; your agents then follow
// the resulting instructions with no dependency on this binary. See specs/0003 §2b.
package main

import (
	"fmt"
	"io"
	"os"
)

// version is stamped at release time via -ldflags "-X main.version=...".
var version = "0.0.0-dev"

func main() {
	if err := run(os.Stdin, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "trellis: "+err.Error())
		os.Exit(1)
	}
}

// run is the testable entrypoint: it reads interactive answers from in, writes
// user-facing output to out, and returns a non-nil error on failure (which main
// turns into a stderr line + exit 1).
func run(in io.Reader, out io.Writer, args []string) error {
	if len(args) == 0 {
		usage(out)
		return nil
	}
	switch args[0] {
	case "version", "-v", "--version":
		fmt.Fprintln(out, "trellis "+version)
		return nil
	case "help", "-h", "--help":
		usage(out)
		return nil
	}
	if h, ok := commands[args[0]]; ok {
		return h(in, out, args[1:])
	}
	return fmt.Errorf("unknown command %q (try `trellis help`)", args[0])
}

// commands is the canonical set of trellis subcommands — the single source the
// dispatch, the usage text, and the docs-consistency check all read (decision-0025).
var commands = map[string]func(in io.Reader, out io.Writer, args []string) error{
	"setup":     setup,
	"remove":    remove,
	"uninstall": uninstall,
}

// commandNames returns every valid command word, including the built-in version/help.
func commandNames() map[string]bool {
	names := map[string]bool{"version": true, "help": true}
	for k := range commands {
		names[k] = true
	}
	return names
}

func usage(w io.Writer) {
	fmt.Fprintln(w, `trellis — setup CLI for the Trellis governance layer

usage:
  trellis setup      interactive setup: pick a mode, detect what it needs (harness for m2), a profile, a model
  trellis remove     undo setup in a project (removes the .trellis overlay)
  trellis uninstall  remove the trellis binary
  trellis version    print the version
  trellis help       show this message`)
}

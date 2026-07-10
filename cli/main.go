// Command trellis is the release-time payload generator for the Trellis governance
// layer.
//
// It is release tooling, not an end-user installer (decision-0043; kodhama-0007
// slice 4, kodhama/trellis#120): `trellis payload` renders the pre-built M1 payload +
// checksum manifest that ships vendored in plugins/trellis/reference/, and the tests
// in this package are the CI guards that keep the vendored payload, the repo's own
// overlay, and the docs in sync with the render. End users install Trellis via the
// Claude Code plugin (/trellis:setup) or the documented manual copy path — never this
// binary; the interactive setup/status/remove/uninstall commands retired with the
// homebrew/curl distribution channel.
package main

import (
	"fmt"
	"io"
	"os"
)

// version is the build stamp (release builds used to set it via -ldflags; CI/dev
// builds run as 0.0.0-dev — the payload carries its own content-derived stamp).
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
// Generator-only since #120 (decision-0043): the end-user commands live on as the
// plugin skills and the manual copy path, not here.
var commands = map[string]func(in io.Reader, out io.Writer, args []string) error{
	"payload": payload,
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
	fmt.Fprintln(w, `trellis — release-time payload generator for the Trellis governance layer

usage:
  trellis payload    render the pre-built plugin payload + checksum manifest (release tooling)
  trellis version    print the version
  trellis help       show this message

This is not the installer. Install Trellis via the Claude Code plugin
(/plugin marketplace add kodhama/kodhama → /plugin install trellis@kodhama →
/trellis:setup) or the manual copy path in the repo README.`)
}

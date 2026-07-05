package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// Terminal styling for the interactive setup flow (decision-0030). Colors reuse the
// landing page's accent green (docs/index.html, #1f9d68) and are applied to ACCENTS
// ONLY — the terminal's own foreground carries body text — so it reads on light and
// dark terminals without theme detection. Respects the NO_COLOR convention.
const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
	ansiDim   = "\x1b[2m"
	ansiGreen = "\x1b[38;2;31;157;104m" // #1f9d68 — the landing's accent
)

func colorEnabled() bool { return os.Getenv("NO_COLOR") == "" }

func paint(code, s string) string {
	if !colorEnabled() {
		return s
	}
	return code + s + ansiReset
}

// ttyPair returns the concrete files when BOTH in and out are terminals, so the
// interactive selector can render in place. Otherwise ok=false and the caller falls
// back to line input — which is every test, pipe, and CI run (they pass non-*os.File
// readers or non-terminals, so the deterministic path is preserved).
func ttyPair(in io.Reader, out io.Writer) (inF, outF *os.File, ok bool) {
	i, iok := in.(*os.File)
	o, ook := out.(*os.File)
	if !iok || !ook {
		return nil, nil, false
	}
	if !term.IsTerminal(int(i.Fd())) || !term.IsTerminal(int(o.Fd())) {
		return nil, nil, false
	}
	return i, o, true
}

// selectInteractive renders an arrow-navigable list (↑/↓ or j/k move, enter selects)
// and returns the chosen key. The terminal is put in raw mode and restored on every
// exit path, including Ctrl-C. Callers gate this behind ttyPair.
func selectInteractive(in, out *os.File, label string, opts []option, def string) (string, error) {
	cur := 0
	for i, o := range opts {
		if o.key == def {
			cur = i
		}
	}

	old, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(in.Fd()), old)

	draw := func(first bool) {
		if !first {
			fmt.Fprintf(out, "\x1b[%dA", len(opts)+1) // back up over the previous render
		}
		fmt.Fprintf(out, "\r\x1b[J%s\r\n", paint(ansiBold, label))
		for i, o := range opts {
			prefix, key := "  ", o.key
			if i == cur {
				prefix, key = paint(ansiGreen, "› "), paint(ansiGreen+ansiBold, o.key)
			}
			fmt.Fprintf(out, "%s%s%s\r\n", prefix, key, paint(ansiDim, "  "+o.label))
		}
	}
	draw(true)

	buf := make([]byte, 3)
	for {
		n, err := in.Read(buf)
		if err != nil {
			return "", err
		}
		b := buf[:n]
		switch {
		case n == 1 && (b[0] == '\r' || b[0] == '\n'):
			fmt.Fprintf(out, "\x1b[%dA\r\x1b[J", len(opts)+1) // collapse the list to one line
			fmt.Fprintf(out, "%s %s\r\n", paint(ansiBold, label), paint(ansiGreen, opts[cur].key))
			return opts[cur].key, nil
		case n == 1 && b[0] == 3: // Ctrl-C
			fmt.Fprint(out, "\r\n")
			return "", fmt.Errorf("cancelled")
		case n >= 3 && b[0] == 0x1b && b[1] == '[' && b[2] == 'A', n == 1 && b[0] == 'k': // up
			if cur > 0 {
				cur--
			}
			draw(false)
		case n >= 3 && b[0] == 0x1b && b[1] == '[' && b[2] == 'B', n == 1 && b[0] == 'j': // down
			if cur < len(opts)-1 {
				cur++
			}
			draw(false)
		}
	}
}

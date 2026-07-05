package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// Terminal styling for the interactive setup (decision-0030). The accent is the
// landing's plant green in its BRIGHT (dark-mode) form, #4ccb90, so it reads on dark
// terminals — with a 256-colour fallback and NO_COLOR honoured. Green appears only in
// the ❯ pointer; the selected label is bold + bright, so text is always high-contrast.
type palette struct{ green, bold, dim, reset string }

func newPalette() palette {
	if os.Getenv("NO_COLOR") != "" {
		return palette{}
	}
	green := "\x1b[38;5;78m" // 256-colour bright green (fallback)
	if ct := os.Getenv("COLORTERM"); ct == "truecolor" || ct == "24bit" {
		green = "\x1b[38;2;76;203;144m" // #4ccb90
	}
	return palette{green: green, bold: "\x1b[1m", dim: "\x1b[2m", reset: "\x1b[0m"}
}

func (p palette) g(s string) string { return wrap(p.green, s, p.reset) }
func (p palette) b(s string) string { return wrap(p.bold, s, p.reset) }
func (p palette) d(s string) string { return wrap(p.dim, s, p.reset) }

func wrap(code, s, reset string) string {
	if code == "" {
		return s
	}
	return code + s + reset
}

// ttyPair returns the concrete files when BOTH in and out are terminals, so the
// selector can render in place. Otherwise ok=false and the caller falls back to line
// input — every test, pipe, and CI run, so the deterministic path is preserved.
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

// selectInteractive renders a bold title, a dim context line, and an arrow-navigable
// list (↑/↓ or j/k move, enter selects, q/Ctrl-C cancels). Green lives only in the ❯
// pointer; the selected label is bold + bright so it stays readable. The terminal is
// restored on every exit path.
func selectInteractive(in, out *os.File, title, hint string, opts []option, def string) (string, error) {
	cur := 0
	for i, o := range opts {
		if o.key == def {
			cur = i
		}
	}
	p := newPalette()

	old, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(in.Fd()), old)

	printed := 1 + len(opts) // title + options
	if hint != "" {
		printed++
	}

	draw := func(first bool) {
		if !first {
			fmt.Fprintf(out, "\x1b[%dA", printed) // back up over the previous render
		}
		fmt.Fprintf(out, "\r\x1b[J%s\r\n", p.b(title))
		if hint != "" {
			fmt.Fprintf(out, "%s\r\n", p.d(hint))
		}
		for i, o := range opts {
			if i == cur {
				fmt.Fprintf(out, "%s%s   %s\r\n", p.g("❯ "), p.b(o.key), p.d(o.label))
			} else {
				fmt.Fprintf(out, "  %s   %s\r\n", o.key, p.d(o.label))
			}
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
			fmt.Fprintf(out, "\x1b[%dA\r\x1b[J", printed) // collapse to one confirmed line
			fmt.Fprintf(out, "%s %s\r\n", p.b(title), p.g(opts[cur].key))
			return opts[cur].key, nil
		case n == 1 && (b[0] == 3 || b[0] == 'q'): // Ctrl-C / q
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

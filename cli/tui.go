package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// Terminal styling for the interactive setup (decision-0030). Everything is drawn from
// the terminal's OWN palette, not fixed hex — the accent is ANSI green and the selected
// row's band is ANSI "bright black" (the theme's gray). So both adapt to the user's
// theme (dark, light, beige…) and stay legible on any background. NO_COLOR is honoured.
type palette struct {
	green, bold, dim, sel, reset string
	on                           bool
}

func newPalette() palette {
	if os.Getenv("NO_COLOR") != "" {
		return palette{}
	}
	return palette{
		green: "\x1b[32m",     // the terminal's own green
		bold:  "\x1b[1m",
		dim:   "\x1b[2m",
		sel:   "\x1b[48;5;8m", // background = the theme's gray (a subtle band)
		reset: "\x1b[0m",
		on:    true,
	}
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

func padTo(s string, n int) string {
	for len(s) < n {
		s += " "
	}
	return s
}

// selectInteractive renders a bold title, a dim context line, an arrow-navigable list,
// and a dim key hint. The selected row gets a full-width theme-gray band + a green ❯;
// its label is bold + default-foreground so it's always readable. ↑/↓ or j/k move,
// enter selects, q/Ctrl-C cancels. The terminal is restored on every exit path.
func selectInteractive(in, out *os.File, title, hint string, opts []option, def string) (string, error) {
	cur, maxName := 0, 0
	for i, o := range opts {
		if o.key == def {
			cur = i
		}
		if len(o.name) > maxName {
			maxName = len(o.name)
		}
	}
	p := newPalette()
	width := 76
	if w, _, err := term.GetSize(int(out.Fd())); err == nil && w > 24 {
		width = w
	}

	old, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(in.Fd()), old)

	footer := p.d("↑↓ move   ⏎ select   q quit")
	// title + [hint] + blank + options + blank + footer
	printed := 3 + len(opts)
	if hint != "" {
		printed++
	}

	drawRow := func(i int, o option) {
		name := padTo(o.name, maxName)
		if i != cur {
			fmt.Fprintf(out, "  %s   %s\r\n", name, p.d(o.desc))
			return
		}
		if !p.on {
			fmt.Fprintf(out, "❯ %s   %s\r\n", name, o.desc)
			return
		}
		// One SGR run with no mid-line reset, so the gray band spans the whole row:
		// gray bg · bold-green arrow · bold default-fg name · dim description · fill.
		fill := width - (2 + len(name) + 3 + len(o.desc))
		if fill < 0 {
			fill = 0
		}
		fmt.Fprintf(out, "%s\x1b[1;32m❯ \x1b[39m%s\x1b[22m   \x1b[2m%s\x1b[22m%s\x1b[0m\r\n",
			p.sel, name, o.desc, padTo("", fill))
	}

	draw := func(first bool) {
		if !first {
			fmt.Fprintf(out, "\x1b[%dA", printed)
		}
		fmt.Fprintf(out, "\r\x1b[J%s\r\n", p.b(title))
		if hint != "" {
			fmt.Fprintf(out, "%s\r\n", p.d(hint))
		}
		fmt.Fprint(out, "\r\n") // breathing room under the question
		for i, o := range opts {
			drawRow(i, o)
		}
		fmt.Fprint(out, "\r\n") // and above the footer
		fmt.Fprintf(out, "%s\r\n", footer)
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
			fmt.Fprintf(out, "%s %s\r\n", p.b(title), p.g(opts[cur].name))
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

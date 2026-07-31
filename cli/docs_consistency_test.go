package main

import (
	"html"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// docSurfaces are the user-facing files whose claims must match the shipped product
// (decision-0025). Paths are relative to this package dir (cli/).
// install.sh returned in #124 as a plugin vendor script — a different, much smaller
// artifact class than the end-user binary installer retired in #120/decision-0043
// (see the note appended to decision-0043 §4); it is a doc surface in its own right
// (its usage text references /trellis:setup) so it is checked here too.
var docSurfaces = []string{
	"../README.md",
	"../docs/index.html",
	"../docs/invariants.html",
	"../install.sh",
	// docs/lp-content.md is the LP's source of truth and was NOT checked here
	// until 2026-07-31. It is the copy an author edits; index.html is the page a
	// consumer reads. Leaving the source out meant the one file most likely to
	// be edited was the one file nothing verified.
	"../docs/lp-content.md",
}

// proseAfterTrellis are lowercase words that legitimately follow "trellis" in prose
// (e.g. "a trellis is structure that enables growth") and are NOT commands. Extend
// this if the docs add more prose — that's the intended, low-friction escape hatch.
var proseAfterTrellis = map[string]bool{
	"is": true, "a": true, "an": true, "the": true, "and": true, "or": true,
	"on": true, "in": true, "to": true, "as": true, "for": true, "with": true,
	"that": true, "governance": true, "mark": true,
}

// TestDocsClaimOnlyRealCommands enforces decision-0025: the docs must not advertise a
// `trellis <command>` the CLI does not have, nor a `/trellis:<skill>` the plugin lacks.
// This is decision-0020's "no claim without a rule behind it", generalized to the
// product surface — the guard that stops the docs drifting ahead of the code.
func TestDocsClaimOnlyRealCommands(t *testing.T) {
	cmds := commandNames()
	skills := pluginSkills(t)
	cmdRe := regexp.MustCompile(`trellis ([a-z][a-z-]+)`)
	skillRe := regexp.MustCompile(`/trellis:([a-z-]+)`)

	for _, f := range docSurfaces {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		text := string(b)

		for _, m := range cmdRe.FindAllStringSubmatch(text, -1) {
			word := m[1]
			if cmds[word] || proseAfterTrellis[word] {
				continue
			}
			t.Errorf("%s references `trellis %s`, which is not a real command. "+
				"Add the command, fix the docs, or (if it is prose) add %q to proseAfterTrellis.",
				f, word, word)
		}

		for _, m := range skillRe.FindAllStringSubmatch(text, -1) {
			if !skills[m[1]] {
				t.Errorf("%s references `/trellis:%s`, but the plugin has no such skill", f, m[1])
			}
		}
	}
}

func pluginSkills(t *testing.T) map[string]bool {
	t.Helper()
	entries, err := os.ReadDir("../plugins/trellis/skills")
	if err != nil {
		t.Fatalf("reading plugin skills: %v", err)
	}
	skills := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			skills[e.Name()] = true
		}
	}
	return skills
}

func readDocSurface(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}

// docs/index.html claims in its own header to be "Generated per kodhama/design-system's
// lp-generator.md contract … composed against this repo's own docs/lp-content.md".
// No generator exists: no build script, no workflow step, and nothing outside
// lp-content.md even references it. The two files are maintained BY HAND IN
// PARALLEL and had silently diverged.
//
// This asserts the install terminal is CONSISTENT four ways at once. An earlier
// version of this guard checked a hardcoded list of command strings and was
// weaker than it looked: it could not see a `data-tab` renamed without its
// `data-panel`, could not see a commands VALUE changed to point anywhere, and —
// worst — disarmed itself, because renaming a URL correctly in both files left a
// list entry matching neither, after which that entry silently checked nothing
// forever. The commit that introduced it shipped a drift it could not see.
//
// Comparing extracted structure instead of hardcoded strings has no list to go
// stale.
func TestInstallTerminalIsConsistentAcrossSourceRenderAndScript(t *testing.T) {
	page := readDocSurface(t, "../docs/index.html")
	source := readDocSurface(t, "../docs/lp-content.md")

	buttons := setOf(submatches(regexp.MustCompile(`data-tab="([^"]+)"`), page))
	panels := parseRenderedPanels(t, page)
	keys := parseCommandsObject(t, page)
	brief := parseBriefTabs(t, source)

	// 1. Every rendered tab button has a panel, and a copy-command key. A button
	//    with no key makes the copy button write the string "undefined"; a button
	//    with no panel activates nothing at all. Both have shipped.
	assertSameKeys(t, "data-tab buttons", buttons, "data-panel panels", setOf(keysOf(panels)))
	assertSameKeys(t, "data-tab buttons", buttons, "commands object keys", setOf(keysOf(keys)))

	// 2. The copy button must copy what the panel shows. Key PRESENCE is not
	//    enough — a key pointing at the wrong command is silent and worse than a
	//    missing one.
	for tab, cmds := range panels {
		want := strings.Join(runnable(cmds), "\n")
		if got := keys[tab]; got != want {
			t.Errorf("tab %q: the copy button copies\n  %q\nbut the panel shows\n  %q\nthe copy button must hand over what the user is looking at", tab, got, want)
		}
	}

	// 3. The content brief and the rendered page must agree, command for command.
	//    These are hand-maintained in parallel with no generator; this is the only
	//    thing that makes a one-sided edit visible.
	assertSameKeys(t, "docs/index.html panels", setOf(keysOf(panels)), "docs/lp-content.md tabs", setOf(keysOf(brief)))
	for tab, pageCmds := range panels {
		briefCmds, ok := brief[tab]
		if !ok {
			continue // already reported by assertSameKeys
		}
		if strings.Join(pageCmds, "\n") != strings.Join(briefCmds, "\n") {
			t.Errorf("tab %q differs between the LP source and the rendered page:\n  docs/lp-content.md:\n    %s\n  docs/index.html:\n    %s\nboth are hand-maintained; a change to one must be made in the other", tab, strings.Join(briefCmds, "\n    "), strings.Join(pageCmds, "\n    "))
		}
	}
}

func submatches(re *regexp.Regexp, s string) []string {
	var out []string
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}

func setOf(xs []string) map[string]bool {
	m := map[string]bool{}
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func keysOf[V any](m map[string]V) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func assertSameKeys(t *testing.T, aName string, a map[string]bool, bName string, b map[string]bool) {
	t.Helper()
	for k := range a {
		if !b[k] {
			t.Errorf("%q appears in %s but not in %s", k, aName, bName)
		}
	}
	for k := range b {
		if !a[k] {
			t.Errorf("%q appears in %s but not in %s", k, bName, aName)
		}
	}
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

// text strips markup and decodes entities, so a rendered command compares equal
// to the same command written as plain text in the brief.
func text(s string) string {
	return strings.TrimRight(html.UnescapeString(tagRe.ReplaceAllString(s, "")), " \t")
}

// runnable drops pure-comment lines. The copy button should hand over commands,
// not `#` annotations the user cannot execute.
func runnable(cmds []string) []string {
	var out []string
	for _, c := range cmds {
		body := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(c), "$>"))
		if body == "" || strings.HasPrefix(body, "#") {
			continue // a pure annotation line — nothing to run
		}
		if i := strings.Index(body, "    #"); i >= 0 {
			body = strings.TrimRight(body[:i], " ") // trailing annotation
		}
		out = append(out, body)
	}
	return out
}

func parseRenderedPanels(t *testing.T, page string) map[string][]string {
	t.Helper()
	re := regexp.MustCompile(`(?s)<div class="panel[^"]*" data-panel="([^"]+)">(.*?)</div>\s*(?:<div class="panel|</div>)`)
	ms := re.FindAllStringSubmatch(page, -1)
	if len(ms) == 0 {
		t.Fatal("no terminal panels found in docs/index.html — this test's premise has drifted")
	}
	out := map[string][]string{}
	code := regexp.MustCompile(`(?s)<code>(.*?)</code>`)
	for _, m := range ms {
		var cmds []string
		for _, c := range code.FindAllStringSubmatch(m[2], -1) {
			cmds = append(cmds, text(c[1]))
		}
		out[m[1]] = cmds
	}
	return out
}

func parseCommandsObject(t *testing.T, page string) map[string]string {
	t.Helper()
	body := regexp.MustCompile(`(?s)var commands = \{(.*?)\n\s*\};`).FindStringSubmatch(page)
	if body == nil {
		t.Fatal("could not find `var commands = {...};` in docs/index.html — if the copy-to-clipboard code was restructured, update this test in the same commit rather than letting it silently stop checking")
	}
	out := map[string]string{}
	for _, m := range regexp.MustCompile(`(?m)^\s*([A-Za-z0-9_-]+):\s*"((?:[^"\\]|\\.)*)"`).FindAllStringSubmatch(body[1], -1) {
		v := strings.ReplaceAll(m[2], `\n`, "\n")
		out[m[1]] = v
	}
	return out
}

func parseBriefTabs(t *testing.T, source string) map[string][]string {
	t.Helper()
	re := regexp.MustCompile("(?s)- `([a-z0-9_-]+)` \\(.*?\\):\n\\s*```\n(.*?)```")
	ms := re.FindAllStringSubmatch(source, -1)
	if len(ms) == 0 {
		t.Fatal("no install tabs found in docs/lp-content.md — this test's premise has drifted")
	}
	out := map[string][]string{}
	for _, m := range ms {
		var cmds []string
		for _, line := range strings.Split(m[2], "\n") {
			if s := strings.TrimSpace(line); s != "" {
				cmds = append(cmds, s)
			}
		}
		out[m[1]] = cmds
	}
	return out
}

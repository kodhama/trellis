package main

import (
	"os"
	"regexp"
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

// docs/index.html claims in its own header to be "Generated per kodhama/design-system's
// lp-generator.md contract … composed against this repo's own docs/lp-content.md".
// No generator exists: there is no build script, no workflow step, and nothing
// outside lp-content.md even references it. The two files are maintained BY HAND
// IN PARALLEL, and by 2026-07-31 they had silently diverged — lp-content.md
// documented both install scopes in the curl tab while index.html documented only
// project scope, and index.html's copy-to-clipboard object had lost its `curl` key
// entirely, so the curl tab's copy button wrote the string "undefined".
//
// Until a real generator exists, this is the cheapest thing that makes the drift
// visible: the install commands a consumer is told to run must appear in BOTH
// files. It deliberately checks commands rather than prose — prose should be
// allowed to differ between a content brief and rendered HTML; the commands
// must not.
func readDocSurface(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}

func TestLandingPageSourceAndRenderedPageAgreeOnCommands(t *testing.T) {
	source := readDocSurface(t, "../docs/lp-content.md")
	page := readDocSurface(t, "../docs/index.html")
	for _, cmd := range []string{
		"/plugin marketplace add kodhama/kodhama",
		"/plugin install trellis@kodhama",
		"/trellis:setup",
		"curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh",
		"git clone --depth 1 https://github.com/kodhama/trellis",
	} {
		inSource, inPage := strings.Contains(source, cmd), strings.Contains(page, cmd)
		if inSource != inPage {
			where := "docs/index.html but not docs/lp-content.md"
			if inSource {
				where = "docs/lp-content.md but not docs/index.html"
			}
			t.Errorf("install command %q appears in %s — the two are hand-maintained in parallel, so a change to one must be made in the other", cmd, where)
		}
	}
}

// The copy button reads commands[activeTab]; a tab with no key copies the literal
// string "undefined". That shipped on the curl tab. Every rendered tab must have
// a key.
func TestEveryTerminalTabHasACopyCommand(t *testing.T) {
	page := readDocSurface(t, "../docs/index.html")
	tabs := regexp.MustCompile(`data-panel="([a-z]+)"`).FindAllStringSubmatch(page, -1)
	if len(tabs) < 2 {
		t.Fatalf("found %d terminal panels; this test's premise has drifted", len(tabs))
	}
	obj := page[strings.Index(page, "var commands = {"):]
	obj = obj[:strings.Index(obj, "};")]
	seen := map[string]bool{}
	for _, m := range tabs {
		if seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		if !strings.Contains(obj, m[1]+":") {
			t.Errorf("terminal tab %q has no entry in the commands object — its copy button writes the string \"undefined\"", m[1])
		}
	}
}

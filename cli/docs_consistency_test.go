package main

import (
	"os"
	"regexp"
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

package main

import (
	"os"
	"path/filepath"
)

// InstructionFile is an agent-instructions file the M1 overlay can attach to.
// Imports marks native @import support (Claude Code): if true the overlay adds a
// one-line import of .trellis/trellis.md; if false it inlines the governance and
// rules directly, since there is nothing to import (decision-0029 follow-up).
type InstructionFile struct {
	Name    string
	Imports bool
}

// instructionFiles are the known agent-instruction files, in preference order:
// CLAUDE.md leads (native @import); AGENTS.md is the portable inline fallback.
var instructionFiles = []InstructionFile{
	{"CLAUDE.md", true},
	{"AGENTS.md", false},
}

func instructionFileByName(n string) (InstructionFile, bool) {
	for _, f := range instructionFiles {
		if f.Name == n {
			return f, true
		}
	}
	return InstructionFile{}, false
}

// detectInstructionFiles reports which known instruction files already exist in dir,
// in preference order — the forward half of "detect what the mode needs" for M1.
func detectInstructionFiles(dir string) []InstructionFile {
	var found []InstructionFile
	for _, f := range instructionFiles {
		if fi, err := os.Stat(filepath.Join(dir, f.Name)); err == nil && !fi.IsDir() {
			found = append(found, f)
		}
	}
	return found
}

func targetOptions() []option {
	opts := make([]option, len(instructionFiles))
	for i, f := range instructionFiles {
		opts[i] = option{f.Name, importKind(f)}
	}
	return opts
}

// importKind describes how the overlay attaches to a file, for display.
func importKind(f InstructionFile) string {
	if f.Imports {
		return "native @import"
	}
	return "inline (no @import)"
}

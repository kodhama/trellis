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

// instructionFiles are the known single-file agent-instruction targets, in preference
// order (research-0010). Only CLAUDE.md keeps Imports: true — inline can't silently
// fail to resolve (D1), so every other tool gets the rules inlined even where it could
// import. AGENTS.md is the cross-tool standard (Codex, Copilot, Devin/Cascade, Windsurf).
// Directory-based conventions (Cursor .cursor/rules/, Continue .continue/rules/) and
// opt-in ones (Aider CONVENTIONS.md) need a different apply path — see research-0010.
var instructionFiles = []InstructionFile{
	{"CLAUDE.md", true},                        // Claude Code (@import)
	{"AGENTS.md", false},                       // Codex, Copilot, Devin/Cascade, Windsurf
	{"GEMINI.md", false},                       // Gemini CLI
	{".github/copilot-instructions.md", false}, // GitHub Copilot
	{".clinerules", false},                     // Cline (single-file form)
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

// importKind describes how the overlay attaches to a file, for display.
func importKind(f InstructionFile) string {
	if f.Imports {
		return "native @import"
	}
	return "inline (no @import)"
}

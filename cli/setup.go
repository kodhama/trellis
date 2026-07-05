package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"strings"
)

// Plan is the outcome of onboarding — what a later `apply` step will act on.
type Plan struct {
	Harness Harness
	Target  InstructionFile // M1 only: the instruction file the overlay augments
	Profile Profile
	Mode    Mode
	Model   Model
}

// option is a single selectable choice: key is returned (and typed in the line path),
// name is the bold display label, desc the dim column beside it.
type option struct{ key, name, desc string }

// setup runs the interactive setup flow (spec-0003 §2b, decision-0029): mode first,
// because the mode decides what to detect — M2 (morph) drives a harness binary to
// rewrite the project, so it detects and requires one; M1 (overlay) edits instruction
// files deterministically and needs no binary. Then a profile and (for M2) a model.
// Each question can be preset with a flag (--mode/--profile/--model); anything omitted
// is prompted for on `in`. Applying the plan (M1 deterministic / M2 model-driven) is a
// later step.
func setup(in io.Reader, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(w)
	dir := fs.String("dir", ".", "project directory to set up")
	profileKey := fs.String("profile", "", "posture: a|b")
	modeKey := fs.String("mode", "", "install mode: m1|m2")
	targetKey := fs.String("target", "", "M1 instruction file to augment: CLAUDE.md|AGENTS.md")
	modelKey := fs.String("model", "", "model: high|balanced|cheap|none")
	applyFlag := fs.Bool("apply", false, "write the changes (default is a dry run)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	sc := bufio.NewScanner(in)

	// Mode first — it decides what the rest of setup even needs to detect.
	mKey, err := ask(in, sc, w, "How should Trellis attach to your project?",
		"holds a few invariants; respects your methodology otherwise", *modeKey, modeOptions(), "m1")
	if err != nil {
		return err
	}
	mode, _ := modeByKey(mKey)

	// Detect only what the chosen mode needs: M2 drives a harness binary to rewrite
	// the project; M1 augments an instruction file and needs no binary (decision-0029).
	var h Harness
	var target InstructionFile
	if mode.Key == "m2" {
		var ok bool
		if h, ok = detectHarness(*dir); !ok {
			if hasClaudeProjectFiles(*dir) {
				return fmt.Errorf("mode m2 rewrites via the harness, but the `claude` CLI isn't on PATH — install Claude Code and re-run (or use --mode m1)")
			}
			return fmt.Errorf("mode m2 needs a harness to drive the rewrite — v0 rides Claude Code; install the `claude` CLI and re-run, or use --mode m1 (looked in %q)", *dir)
		}
		fmt.Fprintf(w, "detected harness: %s (%s)\n\n", h.Name, h.Detail)
	} else if target, err = chooseTarget(in, sc, w, *targetKey, *dir); err != nil {
		return err
	}

	pKey, err := ask(in, sc, w, "How strict should Trellis be?",
		"the posture seeds your profile — tune it in .trellis/ afterward", *profileKey, profileOptions(), "b")
	if err != nil {
		return err
	}
	profile, _ := profileByKey(pKey)

	model, err := resolveModel(in, sc, w, *modelKey, mode)
	if err != nil {
		return err
	}

	plan := Plan{Harness: h, Target: target, Profile: profile, Mode: mode, Model: model}
	printPlan(w, plan)

	doApply := *applyFlag
	if !doApply {
		fmt.Fprint(w, "\napply now? [y/N]: ")
		if sc.Scan() {
			a := strings.ToLower(strings.TrimSpace(sc.Text()))
			doApply = a == "y" || a == "yes"
		}
	}
	if !doApply {
		fmt.Fprintln(w, "dry run — nothing written (re-run with --apply, or answer y)")
		return nil
	}

	switch mode.Key {
	case "m1":
		summary, err := applyM1(*dir, plan)
		if err != nil {
			return err
		}
		fmt.Fprint(w, summary)
		return nil
	default: // m2
		summary, err := applyM2(*dir, plan)
		if err != nil {
			return err
		}
		fmt.Fprint(w, summary)
		return nil
	}
}

// chooseTarget resolves the M1 instruction file to augment (decision-0029 follow-up):
// the --target flag if given, else a prompt over the known files, defaulting to one
// already present (else CLAUDE.md). M1 needs no harness — the target file is the story.
func chooseTarget(in io.Reader, sc *bufio.Scanner, w io.Writer, preset, dir string) (InstructionFile, error) {
	// --target wins (automation) and may name any known file (created if absent).
	if preset != "" {
		f, ok := instructionFileByName(preset)
		if !ok {
			return InstructionFile{}, fmt.Errorf("unknown --target %q (known: %s)", preset, knownTargetNames())
		}
		return f, nil
	}

	// Otherwise offer ONLY the instruction files actually present — the registry is just
	// the checklist of what to look for. None present → create CLAUDE.md or exit.
	detected := detectInstructionFiles(dir)
	switch len(detected) {
	case 0:
		key, err := ask(in, sc, w, "No agent-instructions file found here.",
			"Trellis needs one file to attach to.", "",
			[]option{
				{"create", "Create CLAUDE.md", "add the Trellis section to it"},
				{"exit", "Exit", "I'll add an instructions file myself"},
			}, "create")
		if err != nil {
			return InstructionFile{}, err
		}
		if key == "exit" {
			return InstructionFile{}, fmt.Errorf("no instructions file — nothing written; add one (e.g. CLAUDE.md) and re-run")
		}
		f, _ := instructionFileByName("CLAUDE.md")
		return f, nil
	case 1:
		fmt.Fprintf(w, "found %s — Trellis will extend it\n\n", detected[0].Name)
		return detected[0], nil
	default:
		opts := make([]option, len(detected))
		for i, f := range detected {
			opts[i] = option{f.Name, f.Name, importKind(f)}
		}
		key, err := ask(in, sc, w, "Which instructions file should Trellis extend?",
			"only files present here are offered", "", opts, detected[0].Name)
		if err != nil {
			return InstructionFile{}, err
		}
		f, _ := instructionFileByName(key)
		return f, nil
	}
}

func knownTargetNames() string {
	names := make([]string, len(instructionFiles))
	for i, f := range instructionFiles {
		names[i] = f.Name
	}
	return strings.Join(names, ", ")
}

// resolveModel picks the model for the chosen mode. M2 (morph) is model-driven, so
// it offers the reasoning tiers (default high) and rejects "none" — there is no
// deterministic rewrite. M1 (overlay) is deterministic, so there is no model to pick,
// and a real --model is a loud error rather than a silently-ignored choice.
func resolveModel(in io.Reader, sc *bufio.Scanner, w io.Writer, preset string, mode Mode) (Model, error) {
	if mode.Key != "m2" {
		if preset != "" && preset != "none" {
			return Model{}, fmt.Errorf("mode %s is a deterministic overlay; --model %q does not apply (only 'none')", mode.Key, preset)
		}
		fmt.Fprintln(w, "mode m1 → deterministic overlay, no model needed")
		m, _ := modelByKey("none")
		return m, nil
	}
	key, err := ask(in, sc, w, "Which model drives the rewrite?",
		"M2 rewrites your instructions on a git branch", preset, morphModelOptions(), "high")
	if err != nil {
		return Model{}, err
	}
	m, _ := modelByKey(key)
	return m, nil
}

// ask resolves one choice. If preset is non-empty it is validated and used with no
// prompt (the flag path); otherwise the options are printed and a line is read from
// sc, with empty input taking def. An out-of-set answer is a loud error (D1).
func ask(in io.Reader, sc *bufio.Scanner, w io.Writer, title, hint, preset string, opts []option, def string) (string, error) {
	keys := make([]string, len(opts))
	for i, o := range opts {
		keys[i] = o.key
	}

	if preset != "" {
		if !contains(keys, preset) {
			return "", fmt.Errorf("invalid choice %q (choose one of %s)", preset, strings.Join(keys, ", "))
		}
		return preset, nil
	}

	// A real terminal gets the arrow-key selector; pipes, CI, and tests fall through to
	// the line-based prompt below, so the deterministic path is unchanged (decision-0030).
	if inF, outF, ok := ttyPair(in, w); ok {
		key, err := selectInteractive(inF, outF, title, hint, opts, def)
		if err != nil {
			return "", err
		}
		if !contains(keys, key) {
			return "", fmt.Errorf("invalid choice %q", key)
		}
		return key, nil
	}

	fmt.Fprintf(w, "%s\n", title)
	if hint != "" {
		fmt.Fprintf(w, "  %s\n", hint)
	}
	for _, o := range opts {
		marker := "  "
		if o.key == def {
			marker = "* "
		}
		fmt.Fprintf(w, "%s%s  %s — %s\n", marker, o.key, o.name, o.desc)
	}
	fmt.Fprintf(w, "choose [%s] (default %s): ", strings.Join(keys, "/"), def)

	if !sc.Scan() {
		if def != "" {
			fmt.Fprintln(w, def)
			return def, nil
		}
		return "", fmt.Errorf("no input for %q and no default available", title)
	}
	ans := strings.TrimSpace(sc.Text())
	if ans == "" {
		ans = def
	}
	if !contains(keys, ans) {
		return "", fmt.Errorf("invalid choice %q (choose one of %s)", ans, strings.Join(keys, ", "))
	}
	fmt.Fprintln(w)
	return ans, nil
}

// printPlan renders the final summary in the same visual language as the dialogs:
// a bold header + dim, aligned labels. No trailing "what happens next" line.
func printPlan(w io.Writer, p Plan) {
	pal := paletteFor(w)
	fmt.Fprintf(w, "\n%s\n\n", pal.b("setup plan"))
	row := func(label, val string) {
		fmt.Fprintf(w, "  %s %s\n", pal.d(padTo(label, 8)), val)
	}
	if p.Harness.Name != "" {
		row("harness", p.Harness.Name)
	} else {
		row("target", p.Target.Name+"  "+pal.d(importKind(p.Target)))
	}
	row("profile", p.Profile.Short)
	row("mode", p.Mode.Short)
	row("model", p.Model.Name)
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

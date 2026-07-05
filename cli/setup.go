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
	Profile Profile
	Mode    Mode
	Model   Model
}

// option is a single selectable choice shown in an interactive prompt.
type option struct{ key, label string }

// setup runs the interactive setup flow (spec-0003 §2b): detect the harness, then
// ask for a profile, an install mode, and a model (in that order, so the mode
// informs the model suggestion). Each question can be answered non-interactively
// with a flag (--profile/--mode/--model); anything omitted is prompted for on `in`.
// Applying the plan (M1 deterministic / M2 model-driven) is a later step.
func setup(in io.Reader, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(w)
	dir := fs.String("dir", ".", "project directory to set up")
	profileKey := fs.String("profile", "", "posture: a|b|seed|custom")
	modeKey := fs.String("mode", "", "install mode: m1|m2")
	modelKey := fs.String("model", "", "model: high|balanced|cheap|none")
	if err := fs.Parse(args); err != nil {
		return err
	}

	h, ok := detectHarness(*dir)
	if !ok {
		if hasClaudeProjectFiles(*dir) {
			return fmt.Errorf("this looks like a Claude Code project, but the `claude` CLI isn't on PATH — install Claude Code and re-run")
		}
		return fmt.Errorf("no supported agent harness found — v0 rides Claude Code; install the `claude` CLI and re-run (looked in %q)", *dir)
	}
	fmt.Fprintf(w, "detected harness: %s (%s)\n\n", h.Name, h.Detail)

	sc := bufio.NewScanner(in)

	pKey, err := ask(sc, w, "profile", *profileKey, profileOptions(), "b")
	if err != nil {
		return err
	}
	profile, _ := profileByKey(pKey)

	mKey, err := ask(sc, w, "install mode", *modeKey, modeOptions(), "m1")
	if err != nil {
		return err
	}
	mode, _ := modeByKey(mKey)

	model, err := resolveModel(sc, w, *modelKey, mode)
	if err != nil {
		return err
	}

	printPlan(w, Plan{Harness: h, Profile: profile, Mode: mode, Model: model})
	return nil
}

// resolveModel picks the model for the chosen mode. M2 (morph) is model-driven, so
// it offers the reasoning tiers (default high) and rejects "none" — there is no
// deterministic rewrite. M1 (overlay) is deterministic, so there is no model to pick,
// and a real --model is a loud error rather than a silently-ignored choice.
func resolveModel(sc *bufio.Scanner, w io.Writer, preset string, mode Mode) (Model, error) {
	if mode.Key != "m2" {
		if preset != "" && preset != "none" {
			return Model{}, fmt.Errorf("mode %s is a deterministic overlay; --model %q does not apply (only 'none')", mode.Key, preset)
		}
		fmt.Fprintln(w, "mode m1 → deterministic overlay, no model needed")
		m, _ := modelByKey("none")
		return m, nil
	}
	key, err := ask(sc, w, "model", preset, morphModelOptions(), "high")
	if err != nil {
		return Model{}, err
	}
	m, _ := modelByKey(key)
	return m, nil
}

// ask resolves one choice. If preset is non-empty it is validated and used with no
// prompt (the flag path); otherwise the options are printed and a line is read from
// sc, with empty input taking def. An out-of-set answer is a loud error (D1).
func ask(sc *bufio.Scanner, w io.Writer, label, preset string, opts []option, def string) (string, error) {
	keys := make([]string, len(opts))
	for i, o := range opts {
		keys[i] = o.key
	}

	if preset != "" {
		if !contains(keys, preset) {
			return "", fmt.Errorf("invalid %s %q (choose one of %s)", label, preset, strings.Join(keys, ", "))
		}
		return preset, nil
	}

	fmt.Fprintf(w, "%s:\n", label)
	for _, o := range opts {
		marker := "  "
		if o.key == def {
			marker = "* "
		}
		fmt.Fprintf(w, "%s%s — %s\n", marker, o.key, o.label)
	}
	fmt.Fprintf(w, "choose [%s] (default %s): ", strings.Join(keys, "/"), def)

	if !sc.Scan() {
		if def != "" {
			fmt.Fprintln(w, def)
			return def, nil
		}
		return "", fmt.Errorf("no input for %s and no default available", label)
	}
	ans := strings.TrimSpace(sc.Text())
	if ans == "" {
		ans = def
	}
	if !contains(keys, ans) {
		return "", fmt.Errorf("invalid %s %q (choose one of %s)", label, ans, strings.Join(keys, ", "))
	}
	fmt.Fprintln(w)
	return ans, nil
}

func printPlan(w io.Writer, p Plan) {
	fmt.Fprintln(w, "setup plan:")
	fmt.Fprintf(w, "  harness: %s\n", p.Harness.Name)
	fmt.Fprintf(w, "  profile: %s — %s\n", p.Profile.Name, p.Profile.Description)
	fmt.Fprintf(w, "  mode:    %s — %s\n", p.Mode.Name, p.Mode.Description)
	fmt.Fprintf(w, "  model:   %s\n", p.Model.Name)
	fmt.Fprintln(w, "\nnext: apply — M1 composes deterministically; M2 rewrites on a branch with the model (upcoming step)")
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

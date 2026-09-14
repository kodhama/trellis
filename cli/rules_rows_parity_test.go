package main

// TRL-97 U2. Both hooks read the same .trellis/rules.toml on the plugin-native
// path, and R12 requires them to reach the same answer: the same rules switched
// off, the same file segment, the same warnings, in the same words. The two
// classifiers share no line of code (awk in staleness.sh, JS in
// codex-context.mjs), so the only thing that can be pinned is what each hands
// the model. This table is that pin (decision-0028, a guard per pair).
//
// Every row runs each hook on its own fresh project against one dual-host
// plugin root, asserts that host's delivery against the golden text below, and
// then compares the two hosts with each other. Subtests are named
// <row>/claude, <row>/codex and <row>/parity, so one host can be run alone.
//
// The golden text lives in this file and nowhere else in the test suite. A test
// that read the hook's own templates could not catch the templates changing.
//
// Three consumer-visible pieces were still open with the maintainer when this
// was written, and each sits behind one named place so an answer is a local
// edit rather than a rewrite of the rows:
//
//   - expectedActivationSentence, the computed sentence (KTD1)
//   - rulesEchoMaxBytes and expectedTooLargeLine, the shared threshold (KTD12)
//   - expectedWarnings's bound, the five named others (KTD4)
//
// Out of scope on purpose: CR-only line endings and invalid UTF-8, the two
// divergences TRL-97's follow-up issue records, and every shape where no hook
// classifies the file (the stand-down paths, KTD7).

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// rulesEchoMaxBytes is T (KTD12): a file of at most this many bytes is echoed
// verbatim under the framing; a larger one is replaced by the too-large line.
const rulesEchoMaxBytes = 1500

// codexContextCap is codex-context.mjs's MAX_CONTEXT_BYTES, written out rather
// than read from the hook so a changed cap is a visible edit here.
const codexContextCap = 9500

const activationHeading = "## Project rule activation"

// expectedActivationSentence is the second framing line (KTD1). It names the
// rules the file switches off in the order rules.md lists them.
func expectedActivationSentence(off []string) string {
	if len(off) == 0 {
		return "The project file .trellis/rules.toml switches no rule off, so every rule above applies."
	}
	return "The project file .trellis/rules.toml switches these rules off: " + strings.Join(off, ", ") + ". Every other rule above applies."
}

// expectedTooLargeLine stands in for a file larger than T (KTD12).
func expectedTooLargeLine() string {
	return fmt.Sprintf("The project file is larger than %d bytes, so it is not shown here; the sentence above names every rule it switches off.", rulesEchoMaxBytes)
}

// The warning texts (KTD4). Each names the line and the slug, the key, or only
// the entry type, and never quotes the raw line.
func warnLine(line int, rest string) string {
	return fmt.Sprintf("Trellis warning: line %d of .trellis/rules.toml %s", line, rest)
}

func warnUnknownSlug(line int, slug string) string {
	return warnLine(line, "sets "+slug+" to active = false, but this plugin ships no rule by that name, so the row is ignored; another plugin version may ship it.")
}

func warnFloorRow(line int, slug string) string {
	return warnLine(line, "sets the floor rule "+slug+" to active = false, but floor rules cannot be switched off, so the row is ignored and the rule applies.")
}

func warnDuplicateRow(line int, slug string) string {
	return warnLine(line, "is a later row for "+slug+", so it is ignored; the first row for a rule decides.")
}

func warnRowAboveRules(line int, slug string) string {
	return warnLine(line, "is a row for "+slug+" above the [rules] table, so it is ignored; rows count only under [rules].")
}

func warnRulesAgain(line int) string {
	return warnLine(line, "opens [rules] a second time; the rows under it still count.")
}

func warnOtherTable(line int) string {
	return warnLine(line, "opens a table other than [rules], so every line under it is ignored.")
}

func warnGoverned(line int) string {
	return warnLine(line, "sets governed but does not opt out; only one governed = false line, above every table, opts a project out, so this project stays governed.")
}

func warnUnknownKey(line int, key string) string {
	return warnLine(line, "sets "+key+", which is not a key this file defines, so it is ignored and switches no rule off; a row under [rules] such as slug = { active = false } switches a rule off.")
}

func warnTopLevelLine(line int) string {
	return warnLine(line, "is not an entry this file defines, so it is ignored and switches no rule off.")
}

// warnMalformedRow names the bare key before `=` when the line has one.
func warnMalformedRow(line int, key string) string {
	if key == "" {
		return warnLine(line, "is a malformed row, so it is ignored and switches no rule off.")
	}
	return warnLine(line, "is a malformed row naming "+key+", so it is ignored and switches no rule off; any rule it names still applies.")
}

func warnNulByte() string {
	return "Trellis warning: .trellis/rules.toml contains a NUL byte, so none of its rows take effect and it is not shown; every rule applies."
}

func warnCounted(n int) string {
	return fmt.Sprintf("Trellis warning: ignored entries in .trellis/rules.toml beyond those named above: %d. None of them switches a rule off.", n)
}

type rulesSegment int

const (
	segmentEcho rulesSegment = iota
	segmentTooLarge
	segmentNone
)

type rulesOutcome string

const (
	outcomeDeliver    rulesOutcome = "deliver"
	outcomeRefuse     rulesOutcome = "refuse"
	outcomeUngoverned rulesOutcome = "ungoverned"
)

type rulesRowsCase struct {
	name       string
	toml       string
	unreadable bool
	outcome    rulesOutcome // "" means deliver
	off        []string
	segment    rulesSegment
	warnings   []string
}

func (c rulesRowsCase) wantOutcome() rulesOutcome {
	if c.outcome == "" {
		return outcomeDeliver
	}
	return c.outcome
}

// expectedSegment is the file part of the region: the file verbatim (with a
// newline supplied when it has none), the too-large line, or nothing.
func (c rulesRowsCase) expectedSegment() string {
	switch c.segment {
	case segmentTooLarge:
		return expectedTooLargeLine() + "\n"
	case segmentNone:
		return ""
	}
	if c.toml == "" || strings.HasSuffix(c.toml, "\n") {
		return c.toml
	}
	return c.toml + "\n"
}

// expectedRegion is everything between the heading's blank line and the host's
// footer: the sentence, the segment, then the warnings, one blank line apart.
func (c rulesRowsCase) expectedRegion() string {
	parts := []string{expectedActivationSentence(c.off) + "\n"}
	if seg := c.expectedSegment(); seg != "" {
		parts = append(parts, seg)
	}
	if len(c.warnings) > 0 {
		parts = append(parts, strings.Join(c.warnings, "\n")+"\n")
	}
	return strings.Join(parts, "\n")
}

// rulesFileOfBytes pads body with one comment line to exactly n bytes.
func rulesFileOfBytes(t *testing.T, body string, n int) string {
	t.Helper()
	pad := n - len(body)
	if pad < 3 {
		t.Fatalf("cannot pad a %d-byte body to %d bytes with a comment line", len(body), n)
	}
	out := body + "# " + strings.Repeat("x", pad-3) + "\n"
	if len(out) != n {
		t.Fatalf("padded fixture is %d bytes, wanted %d", len(out), n)
	}
	return out
}

func rulesRowsCases(t *testing.T) []rulesRowsCase {
	t.Helper()

	var fourteen strings.Builder
	fourteen.WriteString("[rules]  # one row per assessable catalog slug\n")
	for _, slug := range assessableSlugs {
		// The fourteen-row set that predates these two rules.
		if slug == "inv-deliberate-succession" || slug == "inv-no-orphan-followups" {
			continue
		}
		fmt.Fprintf(&fourteen, "%-25s = { active = true }\n", slug)
	}

	thisRepo := readFileT(t, "../.trellis/rules.toml")
	if !strings.Contains(thisRepo, `strictness  = "firm"`) || !strings.Contains(thisRepo, "seeded_from") ||
		strings.Count(thisRepo, "{ active = true }") != 16 {
		t.Fatal("premise: AE8 is this repository's file with strictness = \"firm\", seeded_from and sixteen true rows; it no longer is")
	}

	const optOut = "[rules]\ninv-minimal-first = { active = false }\n"
	longSlug := "inv-" + strings.Repeat("a", 60)

	return []rulesRowsCase{{
		name: "AE1 a single false row switches that rule off",
		toml: optOut,
		off:  []string{"inv-minimal-first"},
	}, {
		name:     "AE3 an unknown slug set false is ignored and named",
		toml:     "[rules]\ninv-not-a-rule = { active = false }\n",
		warnings: []string{warnUnknownSlug(2, "inv-not-a-rule")},
	}, {
		name:     "AE4 a floor row set false is ignored",
		toml:     "[rules]\nfloor-transparency = { active = false }\n",
		warnings: []string{warnFloorRow(2, "floor-transparency")},
	}, {
		name:     "AE5 a malformed row costs only itself",
		toml:     "[rules]\ninv-bounded-context = off\ninv-minimal-first = { active = false }\n",
		off:      []string{"inv-minimal-first"},
		warnings: []string{warnMalformedRow(2, "inv-bounded-context")},
	}, {
		name: "AE7 fourteen true rows from an older release",
		toml: fourteen.String(),
	}, {
		name: "AE8 this repository's own file",
		toml: thisRepo,
	}, {
		name: "an empty file",
		toml: "",
	}, {
		name: "only a [rules] header",
		toml: "[rules]\n",
	}, {
		name: "a [rules] header with a trailing comment",
		toml: "[rules]  # one row per rule\n",
	}, {
		name: "CRLF line endings keep an opt-out",
		toml: "[rules]\r\ninv-minimal-first = { active = false }\r\n",
		off:  []string{"inv-minimal-first"},
	}, {
		name: "a byte-order mark before [rules] keeps an opt-out",
		toml: "\ufeff" + optOut,
		off:  []string{"inv-minimal-first"},
	}, {
		name: "a file with no trailing newline",
		toml: strings.TrimSuffix(optOut, "\n"),
		off:  []string{"inv-minimal-first"},
	}, {
		name: "accepted row spellings",
		toml: "  [rules]\n" +
			"inv-directional-flow={active=false}\n" +
			"inv-handover-points\t=\t{\tactive\t=\tfalse\t}\n" +
			"inv-intent-locus = { active = false }  # off for now\n" +
			"inv-ratifiable-artifacts = { active = false }   \n" +
			"\tinv-graph-maintenance = { active = false }\n",
		off: []string{"inv-directional-flow", "inv-handover-points", "inv-intent-locus", "inv-ratifiable-artifacts", "inv-graph-maintenance"},
	}, {
		name: "rejected row spellings",
		toml: "[rules]\n" +
			"\u00a0inv-graph-maintenance = { active = false }\n" +
			"inv-self-improvement = { active = False }\n" +
			"\"inv-deliberate-succession\" = { active = false }\n" +
			"inv-no-orphan-followups.active = false\n",
		warnings: []string{
			warnMalformedRow(2, ""),
			warnMalformedRow(3, "inv-self-improvement"),
			warnMalformedRow(4, ""),
			warnMalformedRow(5, ""),
		},
	}, {
		name: "a multi-line inline table is three malformed lines",
		toml: "[rules]\ninv-gate-at-handover = {\n  active = false\n}\n",
		warnings: []string{
			warnMalformedRow(2, "inv-gate-at-handover"),
			warnMalformedRow(3, "active"),
			warnMalformedRow(4, ""),
		},
	}, {
		name:     "a duplicate slug, false then true",
		toml:     optOut + "inv-minimal-first = { active = true }\n",
		off:      []string{"inv-minimal-first"},
		warnings: []string{warnDuplicateRow(3, "inv-minimal-first")},
	}, {
		name:     "a duplicate slug, true then false",
		toml:     "[rules]\ninv-minimal-first = { active = true }\ninv-minimal-first = { active = false }\n",
		warnings: []string{warnDuplicateRow(3, "inv-minimal-first")},
	}, {
		name:     "a second [rules] header continues the table",
		toml:     optOut + "[rules]\ninv-bounded-context = { active = false }\n",
		off:      []string{"inv-bounded-context", "inv-minimal-first"},
		warnings: []string{warnRulesAgain(3)},
	}, {
		name:     "a false row under another table",
		toml:     "[rules]\n[tool]\ninv-minimal-first = { active = false }\nname = \"x\"\n",
		warnings: []string{warnOtherTable(2)},
	}, {
		name:     "a false row above [rules]",
		toml:     "inv-minimal-first = { active = false }\n[rules]\n",
		warnings: []string{warnRowAboveRules(1, "inv-minimal-first")},
	}, {
		name:     "an unknown top-level key",
		toml:     "disabled = [\"inv-minimal-first\"]\n[rules]\n",
		warnings: []string{warnUnknownKey(1, "disabled")},
	}, {
		name:     "a top-level line that is not a key",
		toml:     "just some words\n[rules]\n",
		warnings: []string{warnTopLevelLine(1)},
	}, {
		name: "strictness and seeded_from are silent whatever they hold",
		toml: "strictness = \"loose\"\nseeded_from = 'anything'\nstrictness = 3\n" + optOut,
		off:  []string{"inv-minimal-first"},
	}, {
		name: "an unknown slug set true is silent",
		toml: "[rules]\ninv-retired-rule = { active = true }\n",
	}, {
		name: "commented-out rows and old quarantine notes are silent",
		toml: "[rules]\n# inv-gone = { active = false }  # quarantined 2026-01-01: not in payload@000000000000.\ninv-minimal-first = { active = true }\n",
	}, {
		name:     "a long unknown slug is named in part",
		toml:     "[rules]\n" + longSlug + " = { active = false }\n",
		warnings: []string{warnUnknownSlug(2, longSlug[:40]+"...")},
	}, {
		name:     "governed = false under [rules] does not opt out",
		toml:     optOut + "governed = false\n",
		off:      []string{"inv-minimal-first"},
		warnings: []string{warnGoverned(3)},
	}, {
		name:     "governed = false twice does not opt out",
		toml:     "governed = false\ngoverned = false\n[rules]\n",
		warnings: []string{warnGoverned(1), warnGoverned(2)},
	}, {
		name:     "governed = \"false\" does not opt out",
		toml:     "governed = \"false\"\n[rules]\n",
		warnings: []string{warnGoverned(1)},
	}, {
		name: "governed = true is silent",
		toml: "governed = true\n" + optOut,
		off:  []string{"inv-minimal-first"},
	}, {
		name:    "AE2 governed = false alone",
		toml:    "governed = false\n",
		outcome: outcomeUngoverned,
	}, {
		name:    "AE2 governed = false beside an unknown key",
		toml:    "governed = false\nprofile = \"b\"\n[rules]\n",
		outcome: outcomeUngoverned,
	}, {
		name:    "AE2 governed = false beside a malformed row",
		toml:    "governed = false\n[rules]\ninv-bounded-context = off\n",
		outcome: outcomeUngoverned,
	}, {
		name:    "AE2 governed = false beside a NUL byte",
		toml:    "governed = false\n[rules]\n\x00\n",
		outcome: outcomeUngoverned,
	}, {
		name:     "a NUL byte and no opt-out",
		toml:     optOut + "\x00\n",
		segment:  segmentNone,
		warnings: []string{warnNulByte()},
	}, {
		name: "twelve malformed rows name five and count the rest",
		toml: "[rules]\n" + strings.Repeat("{ active = false }\n", 12),
		warnings: []string{
			warnMalformedRow(2, ""), warnMalformedRow(3, ""), warnMalformedRow(4, ""),
			warnMalformedRow(5, ""), warnMalformedRow(6, ""), warnCounted(7),
		},
	}, {
		// Six ignored false rows for shipped slugs beside six other warnings. If
		// the six were folded into the five-name bound, the count would read 7.
		name: "ignored false rows for shipped slugs are always named",
		toml: "inv-directional-flow = { active = false }\n" +
			"inv-handover-points = { active = false }\n" +
			"[rules]\n" +
			"floor-transparency = { active = false }\n" +
			"floor-intent-gate = { active = false }\n" +
			"inv-intent-locus = { active = true }\n" +
			"inv-intent-locus = { active = false }\n" +
			"inv-graph-maintenance = { active = true }\n" +
			"inv-graph-maintenance = { active = false }\n" +
			strings.Repeat("not a row\n", 6),
		warnings: []string{
			warnRowAboveRules(1, "inv-directional-flow"),
			warnRowAboveRules(2, "inv-handover-points"),
			warnFloorRow(4, "floor-transparency"),
			warnFloorRow(5, "floor-intent-gate"),
			warnDuplicateRow(7, "inv-intent-locus"),
			warnDuplicateRow(9, "inv-graph-maintenance"),
			warnMalformedRow(10, ""), warnMalformedRow(11, ""), warnMalformedRow(12, ""),
			warnMalformedRow(13, ""), warnMalformedRow(14, ""),
			warnCounted(1),
		},
	}, {
		name: "a file of exactly T bytes is echoed",
		toml: rulesFileOfBytes(t, optOut, rulesEchoMaxBytes),
		off:  []string{"inv-minimal-first"},
	}, {
		name:    "a file one byte over T is not echoed and still switches its rule off",
		toml:    rulesFileOfBytes(t, optOut, rulesEchoMaxBytes+1),
		off:     []string{"inv-minimal-first"},
		segment: segmentTooLarge,
	}, {
		name:       "an unreadable file is refused",
		toml:       optOut,
		unreadable: true,
		outcome:    outcomeRefuse,
	}}
}

// hostRulesResult is what one hook did with one file.
type hostRulesResult struct {
	outcome       rulesOutcome
	context       string
	region        string
	systemMessage string
	raw           string
}

// activationRegion returns the text between the heading's blank line and the
// host's footer. The footer is matched from the end, because an echoed file is
// free to contain anything.
func activationRegion(context, footer string) (string, bool) {
	lead := "\n" + activationHeading + "\n\n"
	i := strings.Index(context, lead)
	if i < 0 {
		return "", false
	}
	start := i + len(lead)
	j := strings.LastIndex(context, footer)
	if j < start {
		return "", false
	}
	return context[start:j], true
}

func runRulesRowsHost(t *testing.T, host, pluginRoot, project string) hostRulesResult {
	t.Helper()
	switch host {
	case "claude":
		out := claudeStdoutFor(t, pluginRoot, project)
		r := hostRulesResult{raw: out}
		if out == "" {
			r.outcome = outcomeUngoverned
			return r
		}
		ctx := nudgeContext(t, out)
		r.context = ctx
		switch {
		case strings.Contains(ctx, rulesLoadedSentinel):
			r.outcome = outcomeDeliver
			r.region, _ = activationRegion(ctx, "\nDelivered by the Trellis plugin")
		case strings.Contains(ctx, "TRELLIS_NOT_GOVERNING"):
			r.outcome = outcomeUngoverned
		default:
			r.outcome = outcomeRefuse
		}
		return r
	case "codex":
		raw, got := runCodexHook(t, pluginRoot, startupInput(t, project))
		r := hostRulesResult{raw: raw, systemMessage: got.SystemMessage}
		switch {
		case raw == "":
			r.outcome = outcomeUngoverned
		case got.HookSpecificOutput == nil:
			r.outcome = outcomeRefuse
		default:
			r.outcome = outcomeDeliver
			r.context = got.HookSpecificOutput.AdditionalContext
			r.region, _ = activationRegion(r.context, "\nTrellis hook loaded installed overlay: ")
		}
		return r
	}
	t.Fatalf("unknown host %q", host)
	return hostRulesResult{}
}

// splitActivationRegion pulls the three parts back out of a region, so a
// mismatch says which part moved. The whole-region equality is the assertion;
// this is its diagnosis.
func splitActivationRegion(region string) (sentence, segment string, warnings []string) {
	lines := strings.SplitAfter(region, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return "", "", nil
	}
	sentence = strings.TrimSuffix(lines[0], "\n")
	rest := lines[1:]
	if len(rest) > 0 && rest[0] == "\n" {
		rest = rest[1:]
	}
	n := len(rest)
	for n > 0 && strings.HasPrefix(rest[n-1], "Trellis warning: ") {
		n--
	}
	for _, w := range rest[n:] {
		warnings = append(warnings, strings.TrimSuffix(w, "\n"))
	}
	body := rest[:n]
	if len(warnings) > 0 && len(body) > 0 && body[len(body)-1] == "\n" {
		body = body[:len(body)-1]
	}
	return sentence, strings.Join(body, ""), warnings
}

// writeInstructionRe is what the hooks' own framing and warnings may never say:
// an instruction to change the file, or the vocabulary of the retired repair.
var writeInstructionRe = regexp.MustCompile(`(?i)\b(write|rewrite|edit|uncomment)\b|reconcil|quarantin`)

func assertRulesRowsHost(t *testing.T, host string, c rulesRowsCase, r hostRulesResult) {
	t.Helper()
	if r.outcome != c.wantOutcome() {
		t.Fatalf("%s: want outcome %q, got %q\n%s", host, c.wantOutcome(), r.outcome, r.raw)
	}
	switch r.outcome {
	case outcomeRefuse:
		if host == "codex" && !strings.Contains(r.raw, ".trellis/rules.toml: unreadable-file") {
			t.Errorf("codex: an unreadable file must be refused as unreadable-file:\n%s", r.raw)
		}
		if host == "claude" && !strings.Contains(r.context, "TRELLIS_RULES_NOT_LOADED") {
			t.Errorf("claude: an unreadable file must be refused as TRELLIS_RULES_NOT_LOADED:\n%s", r.context)
		}
		return
	case outcomeUngoverned:
		// AE2: no rule loads, and nothing delivered says a floor rule applies.
		if strings.Contains(r.raw, "trellis:rules-loaded") || strings.Contains(r.raw, "Project rule activation") {
			t.Errorf("%s: a governed = false project was handed rules:\n%s", host, r.raw)
		}
		return
	}

	if !strings.Contains(r.context, rulesLoadedSentinel) {
		t.Fatalf("%s: the rules themselves were not delivered:\n%s", host, r.context)
	}
	// R14 and AE8: one posture sentence for every project, whatever the file says.
	if !strings.Contains(r.context, "**By default**") || strings.Contains(r.context, "**Firmly**") {
		t.Errorf("%s: every project receives the By default posture sentence and no other:\n%s", host, r.context)
	}
	if r.region == "" && !strings.Contains(r.context, "\n"+activationHeading+"\n\n") {
		t.Fatalf("%s: the context carries no %q section:\n%s", host, activationHeading, r.context)
	}

	want := c.expectedRegion()
	sentence, segment, warnings := splitActivationRegion(r.region)
	if wantSentence := expectedActivationSentence(c.off); sentence != wantSentence {
		t.Errorf("%s: computed sentence\n got: %q\nwant: %q", host, sentence, wantSentence)
	}
	if wantSegment := c.expectedSegment(); segment != wantSegment {
		t.Errorf("%s: file segment\n got: %q\nwant: %q", host, segment, wantSegment)
	}
	if strings.Join(warnings, "\n") != strings.Join(c.warnings, "\n") {
		t.Errorf("%s: warnings, in order\n got: %q\nwant: %q", host, warnings, c.warnings)
	}
	if r.region != want {
		t.Errorf("%s: the activation section differs from the golden text\n got: %q\nwant: %q", host, r.region, want)
	}

	// R15: nothing the hook itself wrote asks for the file to change. The echoed
	// file is the project's own text and is excluded.
	hookText := strings.Replace(r.region, c.expectedSegment(), "", 1)
	if m := writeInstructionRe.FindString(hookText); m != "" {
		t.Errorf("%s: the hook's framing or warnings say %q, and nothing may ask for .trellis/rules.toml to change:\n%s", host, m, hookText)
	}

	if host == "codex" {
		if n := len(r.context); n > codexContextCap {
			t.Errorf("codex: the context is %d bytes, over MAX_CONTEXT_BYTES (%d)", n, codexContextCap)
		}
		// KTD4: the warnings are mirrored to systemMessage, in the same order.
		at := 0
		for _, w := range c.warnings {
			i := strings.Index(r.systemMessage[at:], w)
			if i < 0 {
				t.Errorf("codex: systemMessage does not carry %q after position %d:\n%s", w, at, r.systemMessage)
				break
			}
			at += i + len(w)
		}
		if len(c.warnings) == 0 && strings.Contains(r.systemMessage, ".trellis/rules.toml") {
			t.Errorf("codex: a file with nothing to warn about drew a systemMessage about it:\n%s", r.systemMessage)
		}
	}
}

// TestBothHostsClassifyRulesRowsIdentically is R12 and KTD3 as one table.
func TestBothHostsClassifyRulesRowsIdentically(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	for _, c := range rulesRowsCases(t) {
		t.Run(c.name, func(t *testing.T) {
			results := map[string]hostRulesResult{}
			for _, host := range []string{"claude", "codex"} {
				t.Run(host, func(t *testing.T) {
					project := newGitProject(t)
					path := filepath.Join(project, ".trellis", "rules.toml")
					writeFileT(t, path, c.toml)
					if c.unreadable {
						if os.Geteuid() == 0 {
							t.Skip("running as root: mode 0000 does not deny reads")
						}
						if err := os.Chmod(path, 0o000); err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
					}

					r := runRulesRowsHost(t, host, pluginRoot, project)
					results[host] = r
					assertRulesRowsHost(t, host, c, r)

					// Neither hook writes the file, whatever it says about it.
					if !c.unreadable {
						if after := readFileT(t, path); after != c.toml {
							t.Errorf("%s: the hook changed .trellis/rules.toml\nbefore: %q\nafter:  %q", host, c.toml, after)
						}
					}
				})
			}
			t.Run("parity", func(t *testing.T) {
				claude, okClaude := results["claude"]
				codex, okCodex := results["codex"]
				if !okClaude || !okCodex {
					t.Fatal("a host produced no result to compare")
				}
				if claude.outcome != codex.outcome {
					t.Fatalf("the hosts disagree about the outcome: claude %q, codex %q", claude.outcome, codex.outcome)
				}
				if claude.outcome == outcomeDeliver && claude.region != codex.region {
					t.Errorf("the hosts delivered different activation sections for the same file\nclaude: %q\ncodex:  %q", claude.region, codex.region)
				}
			})
		})
	}
}

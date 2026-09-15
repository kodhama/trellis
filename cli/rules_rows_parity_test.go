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
// The maintainer's answers to the five open questions each sit behind one named
// place here:
//
//   - any false row wins, and a repeated row draws no warning (the rows below)
//   - five named warnings in total, real-rule attempts first (warnMore)
//   - one bound on the file plus its warnings (rulesShowMaxBytes,
//     expectedTooLargeLine)
//   - a symbolic link is classified and never shown (expectedSymlinkLine,
//     warnSymlinkCounted)
//
// The rows themselves are in rules_rows_parity_cases_test.go.
//
// Out of scope on purpose: CR-only line endings and invalid UTF-8, the two
// divergences TRL-97's follow-up issue records, and every shape where no hook
// classifies the file (the stand-down paths, KTD7). Two rows hold invalid
// UTF-8 anyway: one pins that it never truncates the Claude context, and one
// (decodedOverB) that it never costs a Codex session its rules.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// rulesShowMaxBytes is B, the bound both hooks share (codex-context.mjs
// RULES_ECHO_MAX_BYTES, staleness.sh rules_echo_max). The file is echoed
// verbatim only when its bytes plus the bytes of its rendered warning block fit
// B; otherwise the too-large line stands in for it. Measured so the largest
// section either branch can build fits codexContextCap on a plugin root of
// codexPluginRootMaxBytes (the two "largest" rows below).
const rulesShowMaxBytes = 1800

// rulesReadMaxBytes is the read bound both hooks put on the project file
// (codex-context.mjs MAX_PROJECT_CONFIG_BYTES, staleness.sh rules_file_max): a
// larger file is refused rather than read.
const rulesReadMaxBytes = 1024 * 1024

// codexContextCap is codex-context.mjs's MAX_CONTEXT_BYTES, written out rather
// than read from the hook so a changed cap is a visible edit here.
const codexContextCap = 9500

// codexPluginRootMaxBytes is the longest plugin root path the largest section
// must still fit under codexContextCap. The root enters the Codex context once
// per repointed invariants pointer, so the check is measured from the context
// against this suite's own root, whose length varies by runner, rather than
// against a fixed margin. The decision record states the same bound.
const codexPluginRootMaxBytes = 400

const activationHeading = "## Project rule activation"

// expectedActivationSentence is the second framing line (KTD1). It names the
// rules the file switches off in the order rules.md lists them.
func expectedActivationSentence(off []string) string {
	if len(off) == 0 {
		return "The project file .trellis/rules.toml switches no rule off, so every rule above applies."
	}
	return "The project file .trellis/rules.toml switches these rules off: " + strings.Join(off, ", ") + ". Every other rule above applies."
}

// expectedTooLargeLine stands in for a file whose bytes plus its warning block
// exceed B.
func expectedTooLargeLine() string {
	return "The project file and its warnings are too large to show here; the sentence above names every rule it switches off."
}

// expectedSymlinkLine stands in for a file reached through a symbolic link, at
// any size.
func expectedSymlinkLine() string {
	return "The project file is a symbolic link, so its contents are not shown here; the sentence above names every rule it switches off."
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

// warnMore is the count line after the five named warnings: n entries not
// named, m of which try to switch a rule off.
func warnMore(n, m int) string {
	const lead = "Trellis warning: .trellis/rules.toml has "
	if n == 1 {
		if m == 1 {
			return lead + "1 more ignored entry, and it tries to switch a rule off."
		}
		return lead + "1 more ignored entry, and it does not try to switch a rule off."
	}
	switch m {
	case 0:
		return fmt.Sprintf(lead+"%d more ignored entries, none of which tries to switch a rule off.", n)
	case 1:
		return fmt.Sprintf(lead+"%d more ignored entries, 1 of which tries to switch a rule off.", n)
	}
	return fmt.Sprintf(lead+"%d more ignored entries, %d of which try to switch a rule off.", n, m)
}

// warnSymlinkCounted is the only warning a symlinked file draws: a count, and
// nothing read from the file it points at.
func warnSymlinkCounted(n, m int) string {
	const lead = "Trellis warning: .trellis/rules.toml is a symbolic link, so its "
	if n == 1 {
		if m == 1 {
			return lead + "1 ignored entry is counted but not named; it tries to switch a rule off."
		}
		return lead + "1 ignored entry is counted but not named; it does not try to switch a rule off."
	}
	switch m {
	case 0:
		return fmt.Sprintf(lead+"%d ignored entries are counted but not named; none of them tries to switch a rule off.", n)
	case 1:
		return fmt.Sprintf(lead+"%d ignored entries are counted but not named; 1 of them tries to switch a rule off.", n)
	}
	return fmt.Sprintf(lead+"%d ignored entries are counted but not named; %d of them try to switch a rule off.", n, m)
}

// warningBlockBytes is what the bound charges for the warnings: every warning
// line with its newline, the count line included.
func warningBlockBytes(warnings []string) int {
	if len(warnings) == 0 {
		return 0
	}
	return len(strings.Join(warnings, "\n")) + 1
}

type rulesSegment int

const (
	segmentEcho rulesSegment = iota
	segmentTooLarge
	segmentSymlink
	segmentNone
)

// rulesLink says how the project reaches its rules file.
type rulesLink int

const (
	linkNone rulesLink = iota
	// linkFile: .trellis/rules.toml is a symbolic link to a file outside the project.
	linkFile
	// linkDir: .trellis is a symbolic link to a directory outside the project.
	linkDir
	// linkAncestor: the project itself is reached through a symbolic link, and
	// neither .trellis nor its rules file is one, so nothing is hidden.
	linkAncestor
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
	link       rulesLink
	outcome    rulesOutcome // "" means deliver
	off        []string
	segment    rulesSegment
	warnings   []string
	// codexRefusal is the failure class a refused file draws from Codex; ""
	// means unreadable-file. Claude refuses every such file as
	// TRELLIS_RULES_NOT_LOADED.
	codexRefusal string
	// absent is text the file holds that neither host may put anywhere in its
	// output, the Codex systemMessage included.
	absent []string
	// largest marks a row built to the largest section its branch can deliver:
	// the Codex context must fit the cap on a plugin root of
	// codexPluginRootMaxBytes.
	largest bool
	// decodedOverB marks an echoed file that fits B as read but not as decoded.
	// Codex decodes the file as UTF-8, so each byte that is not valid UTF-8
	// enters its context as U+FFFD, three bytes, and Codex charges B what enters
	// its context: it shows the too-large line where Claude echoes the file. The
	// hosts' regions differ here, as they do for invalid UTF-8 in any shown file
	// (TRL-100); both still deliver every rule.
	decodedOverB bool
}

func (c rulesRowsCase) wantOutcome() rulesOutcome {
	if c.outcome == "" {
		return outcomeDeliver
	}
	return c.outcome
}

// expectedSegment is the file part of the region on host: the file verbatim
// (with a newline supplied when it has none), the too-large line, the symlink
// line, or nothing.
func (c rulesRowsCase) expectedSegment(host string) string {
	segment := c.segment
	if c.decodedOverB && host == "codex" {
		segment = segmentTooLarge
	}
	switch segment {
	case segmentTooLarge:
		return expectedTooLargeLine() + "\n"
	case segmentSymlink:
		return expectedSymlinkLine() + "\n"
	case segmentNone:
		return ""
	}
	// The hook output is decoded as JSON, which turns every byte that is not
	// valid UTF-8 into U+FFFD, one per byte; converting through runes does the
	// same, and leaves a valid file unchanged.
	file := string([]rune(c.toml))
	if file == "" || strings.HasSuffix(file, "\n") {
		return file
	}
	return file + "\n"
}

// expectedRegion is everything between the heading's blank line and the host's
// footer: the sentence, the segment, then the warnings, one blank line apart.
func (c rulesRowsCase) expectedRegion(host string) string {
	parts := []string{expectedActivationSentence(c.off) + "\n"}
	if seg := c.expectedSegment(host); seg != "" {
		parts = append(parts, seg)
	}
	if len(c.warnings) > 0 {
		parts = append(parts, strings.Join(c.warnings, "\n")+"\n")
	}
	return strings.Join(parts, "\n")
}

// checkBoundPremise keeps the table honest about the bound: a row that declares
// the echo or the too-large line must be on that side of B, so a golden cannot
// pin the wrong branch by mistake. The raw bytes decide for every row; a
// decodedOverB row must also cross B once decoded.
func (c rulesRowsCase) checkBoundPremise(t *testing.T) {
	t.Helper()
	if c.wantOutcome() != outcomeDeliver || (c.link != linkNone && c.link != linkAncestor) || (c.segment != segmentEcho && c.segment != segmentTooLarge) {
		return
	}
	charged := len(c.toml) + warningBlockBytes(c.warnings)
	want := segmentEcho
	if charged > rulesShowMaxBytes {
		want = segmentTooLarge
	}
	if c.segment != want {
		t.Fatalf("premise: the file and its warnings are %d bytes against a bound of %d, so the row must declare segment %d, not %d", charged, rulesShowMaxBytes, want, c.segment)
	}
	if c.decodedOverB {
		decoded := len(string([]rune(c.toml))) + warningBlockBytes(c.warnings)
		if c.segment != segmentEcho || decoded <= rulesShowMaxBytes {
			t.Fatalf("premise: a decodedOverB row is echoed as read and over the bound decoded; it declares segment %d and decodes to %d bytes", c.segment, decoded)
		}
	}
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

// shippedNonFloorSlugs is the payload's non-floor slugs in rules.md order, the
// order the computed sentence names them in.
func shippedNonFloorSlugs(t *testing.T) []string {
	t.Helper()
	var slugs []string
	for _, m := range regexp.MustCompile("(?m)`((?:inv|floor)-[a-z-]+)`[ \t]*$").FindAllStringSubmatch(payloadFiles()["rules.md"], -1) {
		if !strings.HasPrefix(m[1], "floor-") {
			slugs = append(slugs, m[1])
		}
	}
	if len(slugs) != 14 {
		t.Fatalf("premise: the payload ships fourteen non-floor rules, found %d", len(slugs))
	}
	return slugs
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
	// Q5: nothing read from a symlinked file reaches either host's output, on
	// any outcome. Checked first, because a leak is worse than a wrong branch.
	for _, s := range c.absent {
		if !strings.Contains(c.toml, s) {
			t.Fatalf("premise: the file must hold %q for its absence to mean anything", s)
		}
		if strings.Contains(r.raw, s) || strings.Contains(r.context, s) || strings.Contains(r.systemMessage, s) {
			t.Errorf("%s: text from the file a symbolic link points at reached the output: %q\n%s", host, s, r.raw)
		}
	}
	if r.outcome != c.wantOutcome() {
		t.Fatalf("%s: want outcome %q, got %q\n%s", host, c.wantOutcome(), r.outcome, r.raw)
	}
	switch r.outcome {
	case outcomeRefuse:
		wantCodex := ".trellis/rules.toml: unreadable-file"
		if c.codexRefusal != "" {
			wantCodex = c.codexRefusal
		}
		if host == "codex" && !strings.Contains(r.raw, wantCodex) {
			t.Errorf("codex: this file must be refused as %q:\n%s", wantCodex, r.raw)
		}
		if host == "claude" && !strings.Contains(r.context, "TRELLIS_RULES_NOT_LOADED") {
			t.Errorf("claude: this file must be refused as TRELLIS_RULES_NOT_LOADED:\n%s", r.context)
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

	want := c.expectedRegion(host)
	sentence, segment, warnings := splitActivationRegion(r.region)
	if wantSentence := expectedActivationSentence(c.off); sentence != wantSentence {
		t.Errorf("%s: computed sentence\n got: %q\nwant: %q", host, sentence, wantSentence)
	}
	if wantSegment := c.expectedSegment(host); segment != wantSegment {
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
	hookText := strings.Replace(r.region, c.expectedSegment(host), "", 1)
	if m := writeInstructionRe.FindString(hookText); m != "" {
		t.Errorf("%s: the hook's framing or warnings say %q, and nothing may ask for .trellis/rules.toml to change:\n%s", host, m, hookText)
	}

	if host == "codex" {
		n := len(r.context)
		if n > codexContextCap {
			t.Errorf("codex: the context is %d bytes, over MAX_CONTEXT_BYTES (%d)", n, codexContextCap)
		}
		// Q1: no project file can cost a Codex session its rules, and the largest
		// section the bound allows still fits on a plugin root of
		// codexPluginRootMaxBytes. The root is read back from the repointed
		// invariants pointer, and every byte it grows by is charged once per time
		// the context names it.
		if c.largest {
			m := regexp.MustCompile("`([^`]+)/reference/invariants\\.md`").FindStringSubmatch(r.context)
			if m == nil {
				t.Fatalf("codex: the context names no repointed invariants pointer to measure the plugin root by:\n%s", r.context)
			}
			root := m[1]
			fits := len(root) + (codexContextCap-n)/strings.Count(r.context, root)
			t.Logf("codex: the largest section of this kind assembles a %d-byte context on a %d-byte plugin root, so it fits a plugin root of up to %d bytes", n, len(root), fits)
			if fits < codexPluginRootMaxBytes {
				t.Errorf("codex: the largest section fits a plugin root of only %d bytes, under the %d bytes the decision record promises", fits, codexPluginRootMaxBytes)
			}
		}
		// KTD4: the warnings are mirrored to systemMessage, in the same order and
		// nothing else: the plugin root here is complete, so no invariants report
		// shares the message, and an extra or repeated warning is a failure.
		if wantMessage := strings.Join(c.warnings, " "); r.systemMessage != wantMessage {
			t.Errorf("codex: systemMessage\n got: %q\nwant: %q", r.systemMessage, wantMessage)
		}
	}
}

// writeRulesRowsProject writes the case's file where the case says the project
// reaches it: in place, or through a symbolic link to a place outside the project.
func writeRulesRowsProject(t *testing.T, c rulesRowsCase) (project, path string) {
	t.Helper()
	project = newGitProject(t)
	path = filepath.Join(project, ".trellis", "rules.toml")
	switch c.link {
	case linkFile:
		target := filepath.Join(t.TempDir(), "credentials")
		writeFileT(t, target, c.toml)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, path); err != nil {
			t.Fatal(err)
		}
	case linkDir:
		outside := filepath.Join(t.TempDir(), "elsewhere")
		writeFileT(t, filepath.Join(outside, "rules.toml"), c.toml)
		if err := os.Symlink(outside, filepath.Join(project, ".trellis")); err != nil {
			t.Fatal(err)
		}
	case linkAncestor:
		writeFileT(t, path, c.toml)
		via := filepath.Join(t.TempDir(), "via")
		if err := os.Symlink(project, via); err != nil {
			t.Fatal(err)
		}
		return via, filepath.Join(via, ".trellis", "rules.toml")
	default:
		writeFileT(t, path, c.toml)
	}
	return project, path
}

// TestBothHostsClassifyRulesRowsIdentically is R12 and KTD3 as one table.
func TestBothHostsClassifyRulesRowsIdentically(t *testing.T) {
	pluginRoot := writeDualHostPluginRoot(t)
	for _, c := range rulesRowsCases(t) {
		t.Run(c.name, func(t *testing.T) {
			c.checkBoundPremise(t)
			results := map[string]hostRulesResult{}
			for _, host := range []string{"claude", "codex"} {
				t.Run(host, func(t *testing.T) {
					project, path := writeRulesRowsProject(t, c)
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
				// A decodedOverB row pins each host to its own region above, and
				// the two differ by design.
				if claude.outcome == outcomeDeliver && !c.decodedOverB && claude.region != codex.region {
					t.Errorf("the hosts delivered different activation sections for the same file\nclaude: %q\ncodex:  %q", claude.region, codex.region)
				}
			})
		})
	}
}

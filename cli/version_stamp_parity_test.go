package main

// TRL-39. The plugin's own reference/version is a provenance stamp, not part of
// the rules, and both hosts treat a broken one the same way on the plugin-native
// path: the rules are delivered, and the hook says it cannot name which payload
// build they came from. Before this, codex-context.mjs refused the whole
// session (`invalid-version`) where staleness.sh delivered and annotated, which
// is the failure direction decision-0086 records: a healthy payload refused over
// something that is not the payload.
//
// The model for the Codex report is staleness.sh's path-B provenance line, and
// not TRELLIS_STALENESS_UNKNOWN. That marker fires only where a project stamp
// is compared against the plugin's (a vendored overlay, a rendered rules file)
// and the hook injects nothing. On this path there is no comparison on either
// host, so "drift could not be checked" would name a check nobody makes.
//
// The two hooks share no line to diff (awk and shell there, JavaScript here), so
// the table below runs both on one project and pins them to each other on what
// they can be compared on: whether the rules arrived, whether the stamp was
// named, and the classification each gives the file.
//
// Three shapes have no row, because the hosts do not agree on them yet. That
// is a defect in staleness.sh, filed as TRL-112: a hash with uppercase A to E
// under a UTF-8 locale on bash 3.2, a second line holding a lone non-UTF-8
// byte, and a file whose only line is `#multi`. Both hosts deliver the rules on
// all three and differ on how the stamp is reported. The uppercase row below
// carries an F, which that hook rejects under a UTF-8 locale too.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The two phrases both hosts' reports carry, on whichever channel each host
// has. Everything between them is the path and the classification, asserted
// whole below.
const (
	stampReportLead = "own version stamp could not be read ("
	stampReportTail = "cannot name which payload build it came from"
)

// codexNoStampFooter is what the Codex context says in place of a stamp. It is
// shorter than a stamp on purpose: see TestCodexPluginNativeGovernsOnABadVersionStamp.
const codexNoStampFooter = "Trellis hook loaded installed overlay: no readable stamp\n"

type versionStampCase struct {
	name string
	// body is the file's content, built around the shipped stamp so a decoy or a
	// corrupted copy of it is told from the real one.
	body func(stamp string) string
	// breakIt replaces body for the shapes a write cannot produce.
	breakIt func(t *testing.T, path string)
	// defect is staleness.sh's classification of the shape; "" is a healthy stamp.
	defect string
}

func versionStampCases() []versionStampCase {
	const notAStamp = "is not a Trellis payload stamp"
	const multiLine = "carries more than one line of content, where a payload stamp is exactly one"
	return []versionStampCase{
		// Healthy. The last four are what an editor or a core.autocrlf=true
		// checkout produces, and staleness.sh forgives exactly these.
		{name: "a stamp and a newline", body: func(s string) string { return s + "\n" }},
		{name: "a stamp with no newline", body: func(s string) string { return s }},
		{name: "a CRLF line ending", body: func(s string) string { return s + "\r\n" }},
		{name: "trailing whitespace on the stamp line", body: func(s string) string { return s + " \t\n" }},
		{name: "trailing blank lines", body: func(s string) string { return s + "\n\n\n" }},
		{name: "blank lines before the stamp", body: func(s string) string { return "\n \n" + s + "\n" }},
		// codex-context.mjs scans the file 4096 bytes at a time, so a stamp that
		// starts in one read and ends in the next is the shape that loses its
		// first half if the scan forgets a line between reads.
		{name: "a stamp that straddles two reads", body: func(s string) string { return strings.Repeat("\n", 4090) + s + "\n" }},

		// The four classes a file can fail to yield anything in.
		{name: "no file", breakIt: removeFileT, defect: "is missing"},
		{name: "a zero-byte file", body: func(string) string { return "" }, defect: "is empty"},
		{name: "a file of newlines", body: func(string) string { return "\n\n" }, defect: "is empty"},
		{name: "a directory at the path", breakIt: func(t *testing.T, path string) {
			removeFileT(t, path)
			if err := os.Mkdir(path, 0o755); err != nil {
				t.Fatal(err)
			}
		}, defect: "is not a readable file — a directory or a device sits at that path"},
		{name: "a file at mode 000", breakIt: func(t *testing.T, path string) {
			if os.Geteuid() == 0 {
				t.Skip("root reads a mode-000 file, so the shape cannot be built")
			}
			if err := os.Chmod(path, 0o000); err != nil {
				t.Fatal(err)
			}
		}, defect: "exists but could not be read — a permission mode, a stale ACL, or a symlink whose target is gone"},

		// One line of content that is not a stamp.
		{name: "a stamp from the plugin@ era", body: func(string) string { return "plugin@abcdef123456\n" }, defect: notAStamp},
		{name: "uppercase hex", body: func(string) string { return "payload@ABCDEF123456\n" }, defect: notAStamp},
		{name: "a truncated hash", body: func(string) string { return "payload@abc\n" }, defect: notAStamp},
		{name: "a hash one digit too long", body: func(s string) string { return s + "0\n" }, defect: notAStamp},
		{name: "whitespace inside the stamp", body: func(s string) string { return s[:12] + " " + s[12:] + "\n" }, defect: notAStamp},
		{name: "whitespace before the stamp", body: func(s string) string { return " " + s + "\n" }, defect: notAStamp},
		{name: "a line of spaces", body: func(string) string { return "   \n" }, defect: notAStamp},

		// More than one line of content, wherever the real stamp sits in it.
		{name: "a stamp followed by garbage", body: func(s string) string { return s + "\nGARBAGE\n" }, defect: multiLine},
		{name: "a stamp followed by a decoy stamp", body: func(s string) string { return s + "\npayload@ffffffffffff\n" }, defect: multiLine},
		{name: "a stamp preceded by garbage", body: func(s string) string { return "GARBAGE\n" + s + "\n" }, defect: multiLine},
		// The second line is only counted when the file ends, which is a
		// different check from the one a newline triggers.
		{name: "a stamp followed by garbage with no final newline", body: func(s string) string { return s + "\nGARBAGE" }, defect: multiLine},
		{name: "a stamp preceded by garbage with no final newline", body: func(s string) string { return "GARBAGE\n" + s }, defect: multiLine},
		{name: "a second line of content one read later", body: func(s string) string { return s + "\n" + strings.Repeat("\n", 5000) + "GARBAGE\n" }, defect: multiLine},
	}
}

func TestBothHostsReportABadVersionStampIdentically(t *testing.T) {
	stamp := strings.TrimSpace(payloadFile(t, "version"))
	rules := payloadFile(t, "rules.md")
	for _, tc := range versionStampCases() {
		t.Run(tc.name, func(t *testing.T) {
			pluginRoot := writeDualHostPluginRoot(t)
			project := writeConfigOnlyProject(t)
			ref := filepath.Join(pluginRoot, "reference", "version")
			if tc.breakIt != nil {
				tc.breakIt(t, ref)
			} else {
				writeFileT(t, ref, tc.body(stamp))
			}

			claude := claudeContextFor(t, pluginRoot, project)
			raw, codex := runCodexHook(t, pluginRoot, startupInput(t, project))
			if codex.HookSpecificOutput == nil {
				t.Fatalf("Codex refused a session over the plugin's version stamp, whose rules payload is intact: %s", raw)
			}
			codexContext := codex.HookSpecificOutput.AdditionalContext

			// Governed, whatever the stamp. This is the assertion the defect failed.
			for host, context := range map[string]string{"staleness.sh (Claude)": claude, "codex-context.mjs (Codex)": codexContext} {
				if !strings.Contains(context, rules) {
					t.Errorf("%s: the rules body did not arrive whole:\n%s", host, context)
				}
			}

			if tc.defect == "" {
				if want := "Delivered by the Trellis plugin (" + stamp + ")."; !strings.Contains(claude, want) {
					t.Errorf("Claude: a healthy stamp must be named: want %q in:\n%s", want, claude)
				}
				if want := "Trellis hook loaded installed overlay: " + stamp + "\n"; !strings.HasSuffix(codexContext, want) {
					t.Errorf("Codex: a healthy stamp must be named: want the context to end %q, got:\n%s", want, codexContext)
				}
				if strings.Contains(claude, stampReportLead) || codex.SystemMessage != "" {
					t.Errorf("a healthy stamp was reported as broken, which is the over-correction\nClaude: %s\nCodex: %q", claude, codex.SystemMessage)
				}
				return
			}

			// The same file, the same classification, in the same words. Claude
			// has one channel and says it in the context; Codex says it on
			// systemMessage, outside the context budget.
			report := stampReportLead + ref + " " + tc.defect + ")"
			for host, got := range map[string]string{"staleness.sh (Claude)": claude, "codex-context.mjs (Codex)": codex.SystemMessage} {
				if !strings.Contains(got, report) {
					t.Errorf("%s: want the report to carry %q, got:\n%s", host, report, got)
				}
				if !strings.Contains(got, stampReportTail) {
					t.Errorf("%s: want the report to say it %s, got:\n%s", host, stampReportTail, got)
				}
			}
			// A stamp that could not be validated is not quoted as the build's
			// identity, on either host. The decoy and corrupted rows are why.
			if strings.Contains(claude, "Delivered by the Trellis plugin (") {
				t.Errorf("Claude named a stamp it could not validate:\n%s", claude)
			}
			if !strings.HasSuffix(codexContext, codexNoStampFooter) {
				t.Errorf("Codex: want the context to end %q, got:\n%s", codexNoStampFooter, codexContext)
			}
		})
	}
}

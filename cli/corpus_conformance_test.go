package main

// The corpus conformance check (TRL-88). It applies the mechanically decidable
// rules of core/rubrics/artifact-contract.md, checks 1–11, to the artifact
// corpus on every CI run that touches it, and reports each violation as
// `<path>:<line>: check <N> (<rule>): <detail>`.
//
// Three tests carry it:
//
//   - TestCorpusConformsToArtifactContract runs the check on the live corpus.
//     Findings are collected over the whole corpus first and reported in one
//     subtest per numbered check, so `go test -v` shows each check pass or fail.
//   - TestCorpusConformanceRejectsKnownBadFixture runs the same check on
//     core/fixtures/known-bad/, a standalone corpus of seeded violations and
//     deliberately valid constructs, and requires exactly the expected findings.
//     It calls newCorpusCheck and run itself rather than checkArtifactCorpus,
//     because it needs the check value. A check is trusted only after it
//     rejects the known-bad fixture (the rubric's "How it is graded").
//   - TestCorpusConformanceHaltsOnMissingInput proves the halt policy: a missing
//     or unparseable shared input stops the run by name, never a partial pass.
//
// Every part of every rubric check has a recorded outcome in
// contractOutcomeTable. The control test fails when an implemented rule has no
// seeded violation, or an accepts rule has no valid construct, and
// cli/artifact_contract_guard_test.go fails when a numbered rubric check has no
// rows or its text no longer matches the digest its rows carry.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
)

// corpusRoots is the one place this check names the corpus. Every rule keys on
// an artifact's `type`, never its directory, so moving `decisions/` and
// `research/` (TRL-89) changes the one line below and nothing else. The paths
// are `../`-prefixed literals on purpose: cli/ci_paths_guard_test.go extracts
// exactly that form, so a new root here fails that guard until cli-ci.yml's
// paths filter selects it. A loop joining bare directory names would hide the
// reads from it.
var corpusRoots = []string{
	"../decisions", "../research",
	"../core/invariants", "../core/rubrics", "../core/schemas", "../core/catalog", "../core/lexicon.md",
	"../profiles",
}

// knownBadRoot is the positive control, outside corpusRoots: the rubric
// excludes core/fixtures/ from normal runs.
const knownBadRoot = "../core/fixtures/known-bad"

// artifactContractPath is the rubric this check applies and reads data from.
const artifactContractPath = "../core/rubrics/artifact-contract.md"

// contractOutcome is the vocabulary of the plan's Check outcomes table.
type contractOutcome string

const (
	outcomeImplemented contractOutcome = "implemented" // reports findings; needs a seeded violation
	outcomeAccepts     contractOutcome = "accepts"     // only accepts or exempts input; needs a valid construct
	outcomeHalts       contractOutcome = "halts"       // proven by TestCorpusConformanceHaltsOnMissingInput
	outcomeCovered     contractOutcome = "covered"     // another rule or test enforces it; the reason names which
	outcomeDropped     contractOutcome = "dropped"     // no program can decide it, or it does not apply here
	outcomeGuidance    contractOutcome = "guidance"    // judgment, stated where a reviewer reads it
	outcomeNoRule      contractOutcome = "no rule"     // the rubric itself says it is not gated
	outcomeRetired     contractOutcome = "retired"     // the rubric retired the check
)

type contractRuleOutcome struct {
	rule    string // rule id, or clause name for the unnumbered clauses
	outcome contractOutcome
	reason  string
}

// contractCheckOutcomes groups one numbered rubric check's rule rows. digest is
// contractCheckDigest of the check's rubric text as those rows were last
// reviewed against. cli/artifact_contract_guard_test.go fails when the rubric
// text no longer matches it, so an edit anywhere in a check, prose included,
// forces its rows and rule code to be re-reviewed in the same change (KTD8).
type contractCheckOutcomes struct {
	check  int
	digest string
	rules  []contractRuleOutcome
}

// contractOutcomeTable records an outcome for every part of rubric checks 1–12,
// accepted by the maintainer on 2026-09-14 (TRL-88).
var contractOutcomeTable = []contractCheckOutcomes{
	{1, "f584533051ec8b80", []contractRuleOutcome{
		{"1a", outcomeImplemented, "the file opens with a `---` frontmatter block that closes, and every line in it is a single-line `key: value`"},
		{"1b", outcomeImplemented, "`id`, `type`, `depends_on` and `owner` each appear once with a value; `status` is never required or flagged"},
		{"1c", outcomeImplemented, "`depends_on` is a flow list; `informed_by`, `superseded_by`, `superseded_in_part_by` and `changes` are flow lists when present; `id`, `type` and `owner` are scalars"},
	}},
	{2, "cd40525c2890a982", []contractRuleOutcome{
		{"2a", outcomeImplemented, "the `type` value is a row of check 6's closed enumeration"},
		{"2b", outcomeImplemented, "a per-file `scope` is `core-methodology`, `trellis-product` or `trellis-meta`"},
		{"2c", outcomeDropped, "scope and rubric \"may be declared centrally\", but no central declaration exists (the rubric's own open question), and most files carry no per-file `scope`, so requiring one would reject valid artifacts"},
		{"2d", outcomeDropped, "the recognized typed artifacts and their scopes are descriptive; the typed checks key on `type`, and 2b validates any declared `scope`"},
	}},
	{3, "db2436267ba7fc2e", []contractRuleOutcome{
		{"3", outcomeImplemented, "an `id` declared by more than one file is one finding naming every file"},
	}},
	{4, "b528af1d8edcda16", []contractRuleOutcome{
		{"4a", outcomeAccepts, "an existing artifact id in the corpus"},
		{"4b", outcomeAccepts, "`brief-§` followed by a non-empty section token, on shape only"},
		{"4c", outcomeAccepts, "`<repo>/<id>` whose repo is in the registry list read from the rubric; not verified against the other repository"},
		{"4d", outcomeAccepts, "a retired id in the Identifiers section of the artifact whose id is `invariants-v1`"},
		{"4e", outcomeAccepts, "a retired artifact id in the 3a table of the artifact whose id is `decision-0079`"},
		{"4f", outcomeAccepts, "an `@version` pin is stripped before resolving; there is no pin-versus-upstream comparison"},
		{"4i", outcomeImplemented, "a `depends_on` entry that matches none of 4a–4f is one finding naming the entry"},
		{"4g", outcomeImplemented, "an `informed_by` entry resolves by the same forms"},
		{"4h", outcomeImplemented, "an `@version` pin on an `informed_by` entry is a category error, reported before stripping, and no second finding follows"},
	}},
	{5, "1dc718a76031d4d8", []contractRuleOutcome{
		{"5a", outcomeCovered, "covered by 4i: every `depends_on` entry resolves in the corpus, with no separate code"},
		{"5b", outcomeDropped, "the status-lifecycle form applies only where a methodology declares a lifecycle, and this repository declares none (`decision-0082`)"},
		{"5c", outcomeCovered, "covered by 1c: `changes` is checked for shape only, and no rule walks edges"},
		{"5d", outcomeGuidance, "whether an artifact's correctness rests on a source relabeled as `informed_by` is judgment (`decision-0047`); stated in the `AGENTS.md` conformance bullet"},
	}},
	{6, "34023c0a987ead35", []contractRuleOutcome{
		{"6a", outcomeImplemented, "each required section is an H2 outside fenced code whose text, case-insensitively, is the name alone or the name followed by ` (`; rows read from the rubric"},
		{"6b", outcomeAccepts, "`feedback` and `schema` are exempt"},
		{"6c", outcomeNoRule, "the rubric states that a research note's sources and confidence tags are not gated"},
	}},
	{7, "24127ddd4e98bb20", []contractRuleOutcome{
		{"7a", outcomeImplemented, "`superseded_by` entries resolve by check 4's forms"},
		{"7b", outcomeImplemented, "`superseded_in_part_by` entries resolve by check 4's forms"},
		{"7c", outcomeImplemented, "a non-`decision` artifact may not `depends_on` one carrying `superseded_by`, unless it is listed as that artifact's successor"},
	}},
	{8, "b698024e0925d3af", []contractRuleOutcome{
		{"8a", outcomeCovered, "covered by TestRowSetDerivativesFollowThePin, which fails when the catalog's entries or the invariant set's live entries differ from the pinned slug set in cli/payload_test.go, which holds no dial or collapsed slug"},
		{"8b", outcomeImplemented, "each catalog entry carries the ten fields, read through `·`-joined lines, wrapped lines and bolding"},
		{"8c", outcomeImplemented, "`honored` and `violated` form at least two pairs, equal in number, whose tags match position by position, compared exactly"},
		{"8d", outcomeCovered, "covered by TestRowSetDerivativesFollowThePin: a dial entry reads there as an extra slug"},
	}},
	{9, "6903f339c3624c8d", []contractRuleOutcome{
		{"9", outcomeImplemented, "every profile slug resolves to a signature-catalog entry"},
	}},
	{10, "ae19550ba788b2e9", []contractRuleOutcome{
		{"10a", outcomeImplemented, "an `active: true`, `basis: honored-implicitly` row carries `evidence` and a `confidence` of `verified`, `inferred` or `speculated` (`schema-typed-artifacts`)"},
		{"10b", outcomeGuidance, "whether the evidence pointer shows the tell is judgment; stated in the `AGENTS.md` conformance bullet"},
	}},
	{11, "c8ffd19e518966cc", []contractRuleOutcome{
		{"11", outcomeImplemented, "no profile row sets `C2: none` on a slug whose catalog entry has `intent_locus: true`"},
	}},
	{12, "d8d6290004c415ab", []contractRuleOutcome{
		{"12", outcomeRetired, "the rubric retired the version cross-check and keeps only its number"},
	}},
}

// contractClauseOutcomes records the rubric's unnumbered clauses, keyed by name.
var contractClauseOutcomes = []contractRuleOutcome{
	{"Honesty clause", outcomeHalts, "findings are listed per violation over the whole corpus, and a missing or unparseable shared input halts the run by name (KTD6), never a partial pass"},
	{"How it is graded", outcomeImplemented, "one subtest per numbered check with an implemented rule, and every finding names its file, line, check and rule; the section's consumer-facing agent wording stays"},
	{"Charter: \"derive your checklist yourself\"", outcomeDropped, "an instruction to an agent; the checklist is now code pinned to the rubric, and CI runs it whoever authored the change"},
}

// contractRuleRef is a numbered rule's check and outcome.
type contractRuleRef struct {
	check   int
	outcome contractOutcome
}

// contractRuleIndex maps each numbered rule id to its check and outcome.
func contractRuleIndex() map[string]contractRuleRef {
	idx := map[string]contractRuleRef{}
	for _, c := range contractOutcomeTable {
		for _, r := range c.rules {
			idx[r.rule] = contractRuleRef{c.check, r.outcome}
		}
	}
	return idx
}

// contractFinding is one violation. subject is the thing the finding names —
// the missing field or section, the dangling reference, the slug or the pair
// position — and is what the control test's expected set matches on.
type contractFinding struct {
	path    string
	line    int
	rule    string
	subject string
	detail  string
}

func (f contractFinding) String() string {
	return fmt.Sprintf("%s:%d: check %d (%s): %s", f.path, f.line, contractRuleIndex()[f.rule].check, f.rule, f.detail)
}

// artifactContract is what the check reads from the rubric as data (KTD3).
type artifactContract struct {
	sections map[string][]string // check 6: type → required sections; an exempt type maps to none
	repos    map[string]bool     // check 4: the recognized cross-repo registry
}

// loadArtifactContract reads check 6's type rows and check 4's registry list
// from the rubric, through the same windows and row counts the artifact-contract
// guard pins, so there is no second copy of either list to drift. A marker or
// row it cannot read halts the test.
func loadArtifactContract(t *testing.T) artifactContract {
	t.Helper()
	rubric := readFileT(t, artifactContractPath)
	c := artifactContract{sections: map[string][]string{}}
	for _, row := range contractSectionRuleRows(t, rubric) {
		typ, sections, err := parseSectionRule(row)
		if err != nil {
			t.Fatalf("%s: %v", artifactContractPath, err)
		}
		if _, dup := c.sections[typ]; dup {
			t.Fatalf("%s: check 6 names type %q twice", artifactContractPath, typ)
		}
		c.sections[typ] = sections
	}
	repos, err := parseRepoRegistry(contractRefFormRows(t, rubric))
	if err != nil {
		t.Fatalf("%s: %v", artifactContractPath, err)
	}
	c.repos = setOf(repos)
	return c
}

var sectionRuleShape = regexp.MustCompile("^`([a-z][a-z-]*)` → (.+)$")

// parseSectionRule reads one check 6 row: "`decision` → Context + Decision +
// Consequences", or "`schema` → exempt".
func parseSectionRule(row string) (string, []string, error) {
	m := sectionRuleShape.FindStringSubmatch(row)
	if m == nil {
		return "", nil, fmt.Errorf("check 6 row %q is not in the form \"`type` → Section + Section\" or \"`type` → exempt\"", row)
	}
	if m[2] == "exempt" {
		return m[1], nil, nil
	}
	var sections []string
	for _, s := range strings.Split(m[2], " + ") {
		s = strings.TrimSpace(s)
		if s == "" {
			return "", nil, fmt.Errorf("check 6 row %q names an empty section", row)
		}
		sections = append(sections, s)
	}
	return m[1], sections, nil
}

var repoRegistryShape = regexp.MustCompile(`recognized\s+registry\s+\(([^)]*)\)`)
var repoNameShape = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// parseRepoRegistry finds the one check 4 row that lists the recognized
// cross-repo registry and returns its members.
func parseRepoRegistry(rows []string) ([]string, error) {
	var lists []string
	for _, row := range rows {
		if m := repoRegistryShape.FindStringSubmatch(row); m != nil {
			lists = append(lists, m[1])
		}
	}
	if len(lists) != 1 {
		return nil, fmt.Errorf("check 4 carries %d rows naming a \"recognized registry (…)\" list, and this check needs exactly one to know which `<repo>/<id>` references to accept", len(lists))
	}
	var repos []string
	for _, r := range strings.Split(lists[0], ",") {
		r = strings.TrimSpace(r)
		if !repoNameShape.MatchString(r) {
			return nil, fmt.Errorf("check 4's recognized registry lists %q, which is not a repository name", r)
		}
		repos = append(repos, r)
	}
	return repos, nil
}

// corpusArtifact is one Markdown file of the corpus.
type corpusArtifact struct {
	path     string   // repo-relative when under the repository
	lines    []string // lines[0] is line 1
	opens    bool     // line 1 is `---`
	fmEnd    int      // index of the closing `---`; -1 when the block never closes
	fields   []frontmatterField
	badLines []int // indexes of frontmatter lines that are not `key: value`
}

// frontmatterField is one `key: value` line (KTD10). A value opening with `[`
// is a flow list that ends at its first `]`; the rest of that line may only be
// a comment, and a comment may contain brackets, colons and pipes.
type frontmatterField struct {
	key     string
	line    int      // 1-based
	blank   bool     // no value, or only a comment
	isList  bool     // the value opens with `[`
	list    []string // entries, when isList and listErr is empty
	listErr string   // why a list-shaped value is malformed
	scalar  string   // the value without its comment, when !isList
}

var frontmatterLineShape = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*):(?:[ \t]+(.*))?$`)

func parseCorpusArtifact(path, text string) *corpusArtifact {
	text = strings.TrimPrefix(strings.ReplaceAll(text, "\r\n", "\n"), string(rune(0xFEFF)))
	a := &corpusArtifact{path: path, lines: strings.Split(text, "\n"), fmEnd: -1}
	if strings.TrimRight(a.lines[0], " \t") != "---" {
		return a
	}
	a.opens = true
	for i := 1; i < len(a.lines); i++ {
		if strings.TrimRight(a.lines[i], " \t") == "---" {
			a.fmEnd = i
			break
		}
	}
	for i := 1; i < a.fmEnd; i++ {
		line := a.lines[i]
		if trimmed := strings.TrimSpace(line); trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		m := frontmatterLineShape.FindStringSubmatch(line)
		if m == nil {
			a.badLines = append(a.badLines, i)
			continue
		}
		a.fields = append(a.fields, parseFrontmatterValue(m[1], i+1, strings.TrimSpace(m[2])))
	}
	return a
}

func parseFrontmatterValue(key string, line int, raw string) frontmatterField {
	f := frontmatterField{key: key, line: line}
	if raw == "" || strings.HasPrefix(raw, "#") {
		f.blank = true
		return f
	}
	if !strings.HasPrefix(raw, "[") {
		if i := strings.Index(raw, " #"); i >= 0 {
			raw = raw[:i]
		}
		if i := strings.Index(raw, "\t#"); i >= 0 {
			raw = raw[:i]
		}
		f.scalar = strings.TrimSpace(raw)
		return f
	}
	f.isList = true
	end := strings.Index(raw, "]")
	if end < 0 {
		f.listErr = "the flow list never closes with `]`"
		return f
	}
	if rest := strings.TrimSpace(raw[end+1:]); rest != "" && !strings.HasPrefix(rest, "#") {
		f.listErr = fmt.Sprintf("the flow list is followed by %q, which is neither nothing nor a `#` comment", rest)
		return f
	}
	inner := strings.TrimSpace(raw[1:end])
	if inner == "" {
		return f
	}
	for _, e := range strings.Split(inner, ",") {
		e = strings.TrimSpace(e)
		if e == "" || strings.Contains(e, "[") {
			f.listErr = fmt.Sprintf("the flow list %q holds an empty or nested entry", raw[:end+1])
			f.list = nil
			return f
		}
		f.list = append(f.list, e)
	}
	return f
}

// frontmatter reports whether the file has a readable frontmatter block. A file
// without one is reported by rule 1a and read by no other rule.
func (a *corpusArtifact) frontmatter() bool { return a.opens && a.fmEnd > 0 }

// all returns every occurrence of key, in file order.
func (a *corpusArtifact) all(key string) []frontmatterField {
	var out []frontmatterField
	for _, f := range a.fields {
		if f.key == key {
			out = append(out, f)
		}
	}
	return out
}

// first returns the first occurrence of key.
func (a *corpusArtifact) first(key string) (frontmatterField, bool) {
	for _, f := range a.fields {
		if f.key == key {
			return f, true
		}
	}
	return frontmatterField{}, false
}

// scalarValue returns key's first value when it is a non-blank scalar.
func (a *corpusArtifact) scalarValue(key string) string {
	f, ok := a.first(key)
	if !ok || f.blank || f.isList {
		return ""
	}
	return f.scalar
}

// listValue returns key's first value when it is a well-formed flow list.
func (a *corpusArtifact) listValue(key string) (frontmatterField, bool) {
	f, ok := a.first(key)
	if !ok || f.blank || !f.isList || f.listErr != "" {
		return frontmatterField{}, false
	}
	return f, true
}

// typeLine is the line findings about the file's type point at.
func (a *corpusArtifact) typeLine() int {
	if f, ok := a.first("type"); ok {
		return f.line
	}
	return 1
}

// corpusDisplayPath is the path a finding names: repository-relative for a
// file under the repository, the path as given otherwise.
func corpusDisplayPath(p string) string {
	return strings.TrimPrefix(filepath.ToSlash(p), "../")
}

// loadCorpus reads every Markdown file under roots. A root that is missing or
// holds no Markdown, or a file that cannot be read, is a halt: the error names
// it, and no finding is returned for a partial corpus.
func loadCorpus(roots []string) ([]*corpusArtifact, error) {
	var arts []*corpusArtifact
	seen := map[string]bool{}
	for _, root := range roots {
		info, err := os.Stat(root)
		if err != nil {
			return nil, fmt.Errorf("corpus root %s cannot be read (%v), so the corpus is incomplete", corpusDisplayPath(root), err)
		}
		var paths []string
		if info.IsDir() {
			err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(p, ".md") {
					paths = append(paths, p)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walking corpus root %s: %v", corpusDisplayPath(root), err)
			}
		} else if strings.HasSuffix(root, ".md") {
			paths = []string{root}
		}
		if len(paths) == 0 {
			return nil, fmt.Errorf("corpus root %s holds no Markdown artifact, so it was moved, emptied or misnamed; a root that reads nothing would pass by checking nothing", corpusDisplayPath(root))
		}
		sort.Strings(paths)
		for _, p := range paths {
			clean := filepath.Clean(p)
			if seen[clean] {
				return nil, fmt.Errorf("%s is read twice, through overlapping corpus roots; every file would collide with itself on check 3", corpusDisplayPath(p))
			}
			seen[clean] = true
			b, err := os.ReadFile(p)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %v", corpusDisplayPath(p), err)
			}
			arts = append(arts, parseCorpusArtifact(corpusDisplayPath(p), string(b)))
		}
	}
	return arts, nil
}

var (
	fenceOpenShape = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
	h2Shape        = regexp.MustCompile(`^ {0,3}##(?:[ \t]+(.*?))?[ \t]*$`)
	closingHashes  = regexp.MustCompile(`(?:^|[ \t]+)#+$`)
)

// nextFence advances the fenced-code state by one line. open is the fence run
// in force before line, empty outside a fence. It returns the run in force
// after line, and whether line is fenced code: an opening fence, a line inside
// one, or its closing fence. A fence opens on fenceOpenShape, matched against
// the untrimmed line, and closes on a line that, trimmed, starts with the
// opening run and holds nothing but its character. eachBodyLine and the
// artifact-contract guard's contractNumberedChecks share it.
func nextFence(open, line string) (stillOpen string, isFenceLine bool) {
	if open != "" {
		if t := strings.TrimSpace(line); strings.HasPrefix(t, open) && strings.Trim(t, open[:1]) == "" {
			return "", true
		}
		return open, true
	}
	if m := fenceOpenShape.FindStringSubmatch(line); m != nil {
		return m[1], true
	}
	return "", false
}

// eachBodyLine calls fn for every line after the frontmatter that is outside a
// fenced code block, so an example inside a fence is never read as structure.
func (a *corpusArtifact) eachBodyLine(fn func(i int, line string)) {
	start, fence := 0, ""
	if a.frontmatter() {
		start = a.fmEnd + 1
	}
	for i := start; i < len(a.lines); i++ {
		var fenced bool
		if fence, fenced = nextFence(fence, a.lines[i]); !fenced {
			fn(i, a.lines[i])
		}
	}
}

type h2Heading struct {
	text  string
	index int
}

// h2Text is a heading's text from h2Shape's first group: closing hashes
// removed, then trimmed.
func h2Text(group string) string {
	return strings.TrimSpace(closingHashes.ReplaceAllString(group, ""))
}

// h2Headings returns the level-2 ATX headings outside fenced code.
func (a *corpusArtifact) h2Headings() []h2Heading {
	var out []h2Heading
	a.eachBodyLine(func(i int, line string) {
		if m := h2Shape.FindStringSubmatch(line); m != nil {
			out = append(out, h2Heading{h2Text(m[1]), i})
		}
	})
	return out
}

// headingNames is KTD9's match: case-insensitively, the heading is the section
// name alone or the name followed by ` (`. `Decision (draft)` names Decision;
// `Decision state` does not.
func headingNames(heading, name string) bool {
	h, n := strings.ToLower(heading), strings.ToLower(name)
	return h == n || strings.HasPrefix(h, n+" (")
}

// section returns the line range of the body under the first H2 naming name,
// up to the next H2.
func (a *corpusArtifact) section(name string) (start, end int, ok bool) {
	hs := a.h2Headings()
	for k, h := range hs {
		if headingNames(h.text, name) {
			end = len(a.lines)
			if k+1 < len(hs) {
				end = hs[k+1].index
			}
			return h.index + 1, end, true
		}
	}
	return 0, 0, false
}

type mdTable struct {
	index  int // the header row's line index
	header []string
	rows   []mdTableRow
}

type mdTableRow struct {
	index int
	cells []string
}

var tableDelimiterCell = regexp.MustCompile(`^:?-+:?$`)

// tables returns the pipe tables in the line range [start, end) outside fenced
// code: a header row, a delimiter row with as many cells, and body rows. Only
// the leading-pipe form is read, which is the form every corpus table uses.
func (a *corpusArtifact) tables(start, end int) []mdTable {
	var out []mdTable
	var block []int
	flush := func() {
		if len(block) >= 2 {
			header, delim := splitTableRow(a.lines[block[0]]), splitTableRow(a.lines[block[1]])
			ok := len(delim) == len(header)
			for _, d := range delim {
				ok = ok && tableDelimiterCell.MatchString(d)
			}
			if ok {
				t := mdTable{index: block[0], header: header}
				for _, i := range block[2:] {
					t.rows = append(t.rows, mdTableRow{i, splitTableRow(a.lines[i])})
				}
				out = append(out, t)
			}
		}
		block = nil
	}
	a.eachBodyLine(func(i int, line string) {
		if i < start || i >= end {
			return
		}
		isRow := strings.HasPrefix(strings.TrimSpace(line), "|")
		if !isRow || (len(block) > 0 && block[len(block)-1] != i-1) {
			flush()
		}
		if isRow {
			block = append(block, i)
		}
	})
	flush()
	return out
}

// splitTableRow splits a pipe-table row on unescaped pipes, as GFM does, and
// trims each cell.
func splitTableRow(line string) []string {
	s := strings.TrimPrefix(strings.TrimSpace(line), "|")
	if strings.HasSuffix(s, "|") && !strings.HasSuffix(s, `\|`) {
		s = s[:len(s)-1]
	}
	var cells []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '\\' && i+1 < len(s) && s[i+1] == '|':
			cur.WriteByte('|')
			i++
		case s[i] == '|':
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(s[i])
		}
	}
	return append(cells, strings.TrimSpace(cur.String()))
}

// cellValue strips code and bold markers from a table cell.
func cellValue(cell string) string {
	return strings.TrimSpace(strings.NewReplacer("`", "", "**", "").Replace(cell))
}

// corpusCheck holds one run: the corpus, what it reads from the rubric, the ids
// and registries references resolve against, and the findings so far.
type corpusCheck struct {
	contract       artifactContract
	arts           []*corpusArtifact
	ids            map[string]bool // every declared id, for form 4a
	retiredSetIDs  map[string]bool // form 4d, from `invariants-v1`'s Identifiers section
	retiredSpecIDs map[string]bool // form 4e, from `decision-0079`'s 3a table
	findings       []contractFinding
}

// artifactByID returns the first artifact, in path order, declaring id. A
// second one is check 3's finding.
func (c *corpusCheck) artifactByID(id string) *corpusArtifact {
	for _, a := range c.arts {
		if a.frontmatter() && a.scalarValue("id") == id {
			return a
		}
	}
	return nil
}

var resolvesToShape = regexp.MustCompile("resolves to `[^`]+`")

// readRetiredSetIDs reads form 4d's registry, found by id (KTD4): in the
// Identifiers section of `invariants-v1`, every cell of a column whose header
// names `retired` (the legacy codes and the retired artifact ids), and the slug
// of any row whose note says it `resolves to` a successor (a collapsed slug).
// A missing artifact, section or registry is a halt.
func (c *corpusCheck) readRetiredSetIDs() error {
	a := c.artifactByID("invariants-v1")
	if a == nil {
		return fmt.Errorf("no artifact in the corpus declares id `invariants-v1`, whose Identifiers section is check 4's registry of retired ids; if its frontmatter is malformed, that is the first thing to fix")
	}
	start, end, ok := a.section("Identifiers")
	if !ok {
		return fmt.Errorf("%s declares `invariants-v1` and has no `## Identifiers` section, which check 4 reads as its registry of retired ids", a.path)
	}
	ids := map[string]bool{}
	for _, t := range a.tables(start, end) {
		for col, h := range t.header {
			if !strings.Contains(strings.ToLower(h), "retired") {
				continue
			}
			for _, r := range t.rows {
				if col < len(r.cells) && cellValue(r.cells[col]) != "" {
					ids[cellValue(r.cells[col])] = true
				}
			}
		}
		for _, r := range t.rows {
			for _, cell := range r.cells {
				if resolvesToShape.MatchString(cell) && cellValue(r.cells[0]) != "" {
					ids[cellValue(r.cells[0])] = true
				}
			}
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("%s: the Identifiers section holds no table column whose header names `retired`, and no row that `resolves to` a successor, so check 4 has no retired ids to resolve against", a.path)
	}
	c.retiredSetIDs = ids
	return nil
}

// readRetiredSpecIDs reads form 4e's registry, found by id (KTD4): the
// `Retired id` column of the first table after the one `**3a.` marker in
// `decision-0079`, before the next H2. A missing artifact, marker, table or
// column is a halt.
func (c *corpusCheck) readRetiredSpecIDs() error {
	a := c.artifactByID("decision-0079")
	if a == nil {
		return fmt.Errorf("no artifact in the corpus declares id `decision-0079`, whose 3a table is check 4's registry of retired artifact ids; if its frontmatter is malformed, that is the first thing to fix")
	}
	markers := []int{}
	a.eachBodyLine(func(i int, line string) {
		if strings.HasPrefix(strings.TrimSpace(line), "**3a.") {
			markers = append(markers, i)
		}
	})
	if len(markers) != 1 {
		return fmt.Errorf("%s: found %d `**3a.` markers, and check 4 needs exactly one to locate the retired-artifacts registry", a.path, len(markers))
	}
	end := len(a.lines)
	for _, h := range a.h2Headings() {
		if h.index > markers[0] {
			end = h.index
			break
		}
	}
	tables := a.tables(markers[0]+1, end)
	if len(tables) == 0 {
		return fmt.Errorf("%s: no table follows the `**3a.` marker in its section, so check 4 has no retired artifact ids", a.path)
	}
	col := -1
	for j, h := range tables[0].header {
		if strings.EqualFold(cellValue(h), "Retired id") {
			col = j
		}
	}
	if col < 0 {
		return fmt.Errorf("%s: the 3a table has no `Retired id` column, so check 4 has no retired artifact ids", a.path)
	}
	ids := map[string]bool{}
	for _, r := range tables[0].rows {
		if col < len(r.cells) && cellValue(r.cells[col]) != "" {
			ids[cellValue(r.cells[col])] = true
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("%s: the 3a table's `Retired id` column is empty, so check 4 has no retired artifact ids", a.path)
	}
	c.retiredSpecIDs = ids
	return nil
}

// newCorpusCheck loads the corpus under roots. Its error is a halt.
func newCorpusCheck(roots []string, contract artifactContract) (*corpusCheck, error) {
	arts, err := loadCorpus(roots)
	if err != nil {
		return nil, err
	}
	c := &corpusCheck{contract: contract, arts: arts, ids: map[string]bool{}}
	for _, a := range arts {
		if id := a.scalarValue("id"); a.frontmatter() && id != "" {
			c.ids[id] = true
		}
	}
	if err := c.readRetiredSetIDs(); err != nil {
		return nil, err
	}
	if err := c.readRetiredSpecIDs(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *corpusCheck) report(a *corpusArtifact, line int, rule, subject, format string, args ...any) {
	c.findings = append(c.findings, contractFinding{path: a.path, line: line, rule: rule, subject: subject, detail: fmt.Sprintf(format, args...)})
}

// run applies every implemented rule and returns the findings in file order.
func (c *corpusCheck) run() ([]contractFinding, error) {
	c.checkFrontmatterBlock()
	c.checkRequiredFields()
	c.checkFieldTypes()
	c.checkTypeDeclared()
	c.checkScope()
	c.checkUniqueIDs()
	c.checkReferences()
	c.checkSections()
	c.checkSupersededDependencies()
	if err := c.checkTypedArtifacts(); err != nil {
		return nil, err
	}
	sort.SliceStable(c.findings, func(i, j int) bool {
		a, b := c.findings[i], c.findings[j]
		if a.path != b.path {
			return a.path < b.path
		}
		if a.line != b.line {
			return a.line < b.line
		}
		return a.rule < b.rule
	})
	return c.findings, nil
}

// checkFrontmatterBlock applies rule 1a. A file with no readable block is read
// by no other rule, so one malformed file yields one finding, not a cascade.
func (c *corpusCheck) checkFrontmatterBlock() {
	for _, a := range c.arts {
		switch {
		case !a.opens:
			c.report(a, 1, "1a", "frontmatter", "does not open with a `---` frontmatter block")
		case a.fmEnd < 0:
			c.report(a, 1, "1a", "closing ---", "the frontmatter block opened on line 1 is never closed by a `---` line")
		}
		for _, i := range a.badLines {
			c.report(a, i+1, "1a", strings.TrimSpace(a.lines[i]), "frontmatter line %q is not a single-line `key: value`", strings.TrimSpace(a.lines[i]))
		}
	}
}

// requiredFields are rule 1b's fields. `status` is not among them, and nothing
// here reads it: decision-0082 retired it, and a legacy line is history.
var requiredFields = []string{"id", "type", "depends_on", "owner"}

// checkRequiredFields applies rule 1b: each required field appears once, with a
// value. An empty flow list, `depends_on: []`, is a value.
func (c *corpusCheck) checkRequiredFields() {
	for _, a := range c.arts {
		if !a.frontmatter() {
			continue
		}
		for _, key := range requiredFields {
			occ := a.all(key)
			switch {
			case len(occ) == 0:
				c.report(a, 1, "1b", key, "missing required field `%s`", key)
			case len(occ) > 1:
				c.report(a, occ[1].line, "1b", key, "field `%s` appears %d times, and must appear once", key, len(occ))
			case occ[0].blank:
				c.report(a, occ[0].line, "1b", key, "required field `%s` has no value", key)
			}
		}
	}
}

// listFields are the relations rule 1c requires to be flow lists when present.
// `depends_on` is checked with them; its absence is rule 1b's. `changes` is
// checked for shape only and never walked as an edge (rubric check 5).
var listFields = []string{"depends_on", "informed_by", "superseded_by", "superseded_in_part_by", "changes"}

// scalarFields are the fields rule 1c requires to be scalars. Keys the rubric
// does not type, such as `supersedes` or `date`, are not validated.
var scalarFields = []string{"id", "type", "owner"}

// checkFieldTypes applies rule 1c.
func (c *corpusCheck) checkFieldTypes() {
	for _, a := range c.arts {
		if !a.frontmatter() {
			continue
		}
		for _, key := range listFields {
			occ := a.all(key)
			if len(occ) == 0 || (key == "depends_on" && (len(occ) > 1 || occ[0].blank)) {
				continue // rule 1b reports a repeated or empty depends_on
			}
			if len(occ) > 1 {
				c.report(a, occ[1].line, "1c", key, "field `%s` appears %d times; a relation is one flow list", key, len(occ))
				continue
			}
			f := occ[0]
			switch {
			case !f.isList:
				c.report(a, f.line, "1c", key, "`%s` must be a flow list `[…]`, and is %q", key, f.scalar)
			case f.listErr != "":
				c.report(a, f.line, "1c", key, "`%s` is malformed: %s", key, f.listErr)
			}
		}
		for _, key := range scalarFields {
			for _, f := range a.all(key) {
				if f.isList || strings.HasPrefix(f.scalar, "{") {
					c.report(a, f.line, "1c", key, "`%s` must be a scalar, and is a collection", key)
				}
			}
		}
	}
}

// checkTypeDeclared applies rule 2a: the type is a row of check 6's closed
// enumeration. An unknown type has no required sections, so rule 6a stays
// silent for it and this is its only finding.
func (c *corpusCheck) checkTypeDeclared() {
	for _, a := range c.arts {
		typ := a.scalarValue("type")
		if !a.frontmatter() || typ == "" {
			continue
		}
		if _, ok := c.contract.sections[typ]; !ok {
			c.report(a, a.typeLine(), "2a", typ, "type `%s` is not in rubric check 6's closed enumeration of types", typ)
		}
	}
}

// scopeValues are rubric check 2's scopes. A file need not declare one (2c is
// dropped); one it declares must be among them.
var scopeValues = map[string]bool{"core-methodology": true, "trellis-product": true, "trellis-meta": true}

// checkScope applies rule 2b.
func (c *corpusCheck) checkScope() {
	for _, a := range c.arts {
		if !a.frontmatter() {
			continue
		}
		for _, f := range a.all("scope") {
			if f.isList || !scopeValues[f.scalar] {
				c.report(a, f.line, "2b", f.scalar, "scope %q is not `core-methodology`, `trellis-product` or `trellis-meta`", f.scalar)
			}
		}
	}
}

// checkUniqueIDs applies check 3: an id declared by more than one file is one
// finding, at the second file in path order, naming every file.
func (c *corpusCheck) checkUniqueIDs() {
	byID := map[string][]*corpusArtifact{}
	var order []string
	for _, a := range c.arts {
		id := a.scalarValue("id")
		if !a.frontmatter() || id == "" {
			continue
		}
		if len(byID[id]) == 0 {
			order = append(order, id)
		}
		byID[id] = append(byID[id], a)
	}
	for _, id := range order {
		arts := byID[id]
		if len(arts) < 2 {
			continue
		}
		var paths []string
		for _, a := range arts {
			paths = append(paths, a.path)
		}
		f, _ := arts[1].first("id")
		c.report(arts[1], f.line, "3", id, "id `%s` is declared by %d files: %s", id, len(arts), strings.Join(paths, ", "))
	}
}

// resolveRef reports which of rubric check 4's accepted forms, 4a–4e, accepts
// ref, and whether an `@version` pin was stripped first (4f). The empty form
// means none does. A pin is stripped on shape only: nothing compares it with
// the referent's current version.
func (c *corpusCheck) resolveRef(ref string) (form string, pinned bool) {
	if strings.Contains(ref, "@") {
		m := pinnedRefShape.FindStringSubmatch(ref)
		if m == nil {
			return "", false
		}
		ref, pinned = m[1], true
	}
	switch {
	case c.ids[ref]:
		return "4a", pinned
	case briefRefShape.MatchString(ref):
		return "4b", pinned
	case c.retiredSetIDs[ref]:
		return "4d", pinned
	case c.retiredSpecIDs[ref]:
		return "4e", pinned
	}
	if m := crossRepoRefShape.FindStringSubmatch(ref); m != nil && c.contract.repos[m[1]] {
		return "4c", pinned
	}
	return "", pinned
}

// pinnedRefShape is form 4f's shape: a bare reference, `@`, and a version token.
var pinnedRefShape = regexp.MustCompile(`^([^@\s]+)@([A-Za-z0-9][A-Za-z0-9._-]*)$`)

// briefRefShape is form 4b: `brief-§` and a non-empty section token.
var briefRefShape = regexp.MustCompile(`^brief-§[^\s§@]+$`)

// crossRepoRefShape is form 4c's shape, `<repo>/<id>`; the repo must also be in
// the rubric's registry. Shape and membership only: the id is not looked up in
// the other repository.
var crossRepoRefShape = regexp.MustCompile(`^([a-z0-9][a-z0-9-]*)/([A-Za-z0-9][A-Za-z0-9._-]*)$`)

// checkReferences applies rule 4i to `depends_on`, and rule 4h to
// `informed_by`.
func (c *corpusCheck) checkReferences() {
	for _, a := range c.arts {
		if !a.frontmatter() {
			continue
		}
		if f, ok := a.listValue("depends_on"); ok {
			for _, e := range f.list {
				if form, _ := c.resolveRef(e); form == "" {
					c.report(a, f.line, "4i", e, "`depends_on` entry `%s` resolves to no corpus id, `brief-§` section, registry-member `<repo>/<id>`, or retired id", e)
				}
			}
		}
		if f, ok := a.listValue("informed_by"); ok {
			for _, e := range f.list {
				// Before stripping: stripping would swallow the pin silently. The
				// pin is the finding, so no resolution finding follows for it.
				if strings.Contains(e, "@") {
					c.report(a, f.line, "4h", e, "`informed_by` entry `%s` carries an `@version` pin; `informed_by` is provenance and never drifts, so a pin is a category error", e)
					continue
				}
				if form, _ := c.resolveRef(e); form == "" {
					c.report(a, f.line, "4g", e, "`informed_by` entry `%s` resolves to no corpus id, `brief-§` section, registry-member `<repo>/<id>`, or retired id", e)
				}
			}
		}
		// A dangling supersession pointer is a check 7 finding, not a check 4
		// one (KTD5), though it resolves by check 4's forms.
		if f, ok := a.listValue("superseded_by"); ok {
			for _, e := range f.list {
				if form, _ := c.resolveRef(e); form == "" {
					c.report(a, f.line, "7a", e, "`superseded_by` entry `%s` resolves to nothing, so the forward pointer leads nowhere", e)
				}
			}
		}
		if f, ok := a.listValue("superseded_in_part_by"); ok {
			for _, e := range f.list {
				if form, _ := c.resolveRef(e); form == "" {
					c.report(a, f.line, "7b", e, "`superseded_in_part_by` entry `%s` resolves to nothing, so the partial forward pointer leads nowhere", e)
				}
			}
		}
	}
}

// checkSections applies rule 6a: every section check 6 requires for the file's
// type is an H2 outside fenced code, matched by headingNames (KTD9). One finding
// per missing section. An unknown type is rule 2a's and has no sections here;
// an exempt type (6b) requires none.
func (c *corpusCheck) checkSections() {
	for _, a := range c.arts {
		typ := a.scalarValue("type")
		if !a.frontmatter() || typ == "" {
			continue
		}
		headings := a.h2Headings()
		for _, name := range c.contract.sections[typ] {
			if !slices.ContainsFunc(headings, func(h h2Heading) bool { return headingNames(h.text, name) }) {
				c.report(a, a.typeLine(), "6a", name, "type `%s` requires a `## %s` section, and the file has none outside fenced code", typ, name)
			}
		}
	}
}

// bareRef strips a well-formed `@version` pin from a reference.
func bareRef(ref string) string {
	if m := pinnedRefShape.FindStringSubmatch(ref); m != nil {
		return m[1]
	}
	return ref
}

// checkSupersededDependencies applies rule 7c. Supersession is identified by
// the forward pointer (decision-0082): an artifact carrying `superseded_by` is
// superseded. A revise-in-place artifact — anything but a `decision` — that
// still depends on it has not re-pointed to the successor. Exempt: an
// append-only `decision`, which may keep the version current when it was
// written, and a listed successor referencing its predecessor.
func (c *corpusCheck) checkSupersededDependencies() {
	successors := map[string][]string{}
	for _, a := range c.arts {
		id := a.scalarValue("id")
		f, ok := a.listValue("superseded_by")
		if !a.frontmatter() || id == "" || !ok || len(f.list) == 0 {
			continue
		}
		for _, e := range f.list {
			successors[id] = append(successors[id], bareRef(e))
		}
	}
	for _, a := range c.arts {
		typ, id := a.scalarValue("type"), a.scalarValue("id")
		f, ok := a.listValue("depends_on")
		if !a.frontmatter() || typ == "" || typ == "decision" || !ok {
			continue
		}
		for _, e := range f.list {
			bare := bareRef(e)
			succ, superseded := successors[bare]
			if !superseded {
				continue
			}
			listed := id != "" && slices.Contains(succ, id)
			if !listed {
				c.report(a, f.line, "7c", bare, "a `%s` depends on `%s`, which carries `superseded_by: [%s]`; a revise-in-place artifact re-points to the successor", typ, bare, strings.Join(succ, ", "))
			}
		}
	}
}

// checkArtifactCorpus is the entry point: the live test calls it on
// corpusRoots, and the halt test on altered copies of the fixture corpus. The
// control test calls newCorpusCheck and run itself, because it needs the check
// value. A non-nil error is a halt.
func checkArtifactCorpus(roots []string, contract artifactContract) ([]contractFinding, error) {
	c, err := newCorpusCheck(roots, contract)
	if err != nil {
		return nil, err
	}
	return c.run()
}

// reportFindingsPerCheck fails one subtest per numbered check that has an
// implemented rule, listing that check's findings, so a run shows which checks
// hold. A finding under a rule the outcome table does not mark implemented is a
// defect in the check itself and fails the parent test.
func reportFindingsPerCheck(t *testing.T, findings []contractFinding) {
	t.Helper()
	idx := contractRuleIndex()
	byCheck := map[int][]contractFinding{}
	for _, f := range findings {
		r, ok := idx[f.rule]
		if !ok || r.outcome != outcomeImplemented {
			t.Errorf("%s — rule %q is not an implemented row of contractOutcomeTable, so the check reports something its outcome table does not account for", f, f.rule)
			continue
		}
		byCheck[r.check] = append(byCheck[r.check], f)
	}
	for _, c := range contractOutcomeTable {
		if !slices.ContainsFunc(c.rules, func(r contractRuleOutcome) bool { return r.outcome == outcomeImplemented }) {
			continue
		}
		t.Run(fmt.Sprintf("check %d", c.check), func(t *testing.T) {
			for _, f := range byCheck[c.check] {
				t.Error(f)
			}
		})
	}
}

// TestCorpusConformsToArtifactContract is the live run: the corpus under
// corpusRoots yields no finding.
func TestCorpusConformsToArtifactContract(t *testing.T) {
	findings, err := checkArtifactCorpus(corpusRoots, loadArtifactContract(t))
	if err != nil {
		t.Fatalf("the corpus conformance check halted: %v", err)
	}
	reportFindingsPerCheck(t, findings)
}

// expectedContractFinding is one entry of the control test's expected multiset:
// a file under knownBadRoot, a rule id, the subject the finding names, and the
// line it points at.
type expectedContractFinding struct {
	file, rule, subject string
	line                int
}

// knownBadExpected is every finding the check must report on the fixture
// corpus, and nothing else. Two findings of one rule in one file are two
// entries. A rule with several ways to fail seeds each of them, and each entry
// carries its line, so a finding sent to the wrong line fails here.
var knownBadExpected = []expectedContractFinding{
	{"known-bad.md", "1b", "owner", 1},
	{"known-bad.md", "4i", "decision-9999", 4},
	{"known-bad.md", "6a", "Acceptance criteria", 3},
	{"known-bad.md", "6a", "Open questions", 3},
	{"known-bad.md", "7a", "decision-9998", 5},
	{"no-frontmatter.md", "1a", "frontmatter", 1},
	{"unterminated-frontmatter.md", "1a", "closing ---", 1},
	{"malformed-frontmatter-line.md", "1a", "this line is not a key and a value", 6},
	{"repeated-and-empty-fields.md", "1b", "owner", 6},
	{"repeated-and-empty-fields.md", "1b", "type", 3},
	{"scalar-depends-on.md", "1c", "depends_on", 4},
	{"malformed-list.md", "1c", "depends_on", 4},
	{"collection-id.md", "1c", "id", 2},
	{"repeated-relation.md", "1c", "superseded_in_part_by", 6},
	{"unknown-type.md", "2a", "memo", 3},
	{"bad-scope.md", "2b", "trellis-core", 6},
	{"duplicate-id-b.md", "3", "decision-0300", 2},
	{"dangling-refs.md", "4i", "notarepo/x", 4},
	{"dangling-refs.md", "4i", "brief-§", 4},
	{"dangling-refs.md", "4i", "spec-0009", 4},
	{"informed-by.md", "4h", "decision-0001@v1", 5},
	{"informed-by.md", "4h", "research-9998@v1", 5},
	{"informed-by.md", "4g", "research-9999", 5},
	{"decision-state-heading.md", "6a", "Decision", 3},
	{"suffixed-and-fenced-headings.md", "6a", "Consequences", 3},
	{"dangling-partial-supersession.md", "7b", "decision-9997", 6},
	{"stale-note.md", "7c", "decision-0100", 4},
	{"catalog.md", "8b", "This line belongs to no entry.", 53},
	{"catalog.md", "8b", "inv-no-why why", 55},
	{"catalog.md", "8b", "inv-no-c2 default_C2", 68},
	{"catalog.md", "8b", "inv-empty-why why", 108},
	{"catalog.md", "8b", "inv-twice class", 122},
	{"catalog.md", "8c", "inv-one-pair pairs", 82},
	{"catalog.md", "8c", "inv-misaligned pair 2", 104},
	{"catalog.md", "8c", "inv-untagged pair 1", 146},
	{"catalog.md", "8c", "inv-untagged pair 2", 147},
	{"profile.md", "9", "inv-unknown", 28},
	{"profile.md", "10a", "inv-no-why evidence", 26},
	{"profile.md", "10a", "inv-no-c2 confidence", 27},
	{"profile.md", "11", "inv-gate", 23},
	{"unreadable-profile-table.md", "11", "column c2", 20},
}

// acceptedContractConstruct is a deliberately valid fixture construct for an
// accepts rule: the file carries value in field, and yields no finding at all.
type acceptedContractConstruct struct {
	rule, file, field, value string
}

var knownBadAccepted = []acceptedContractConstruct{
	{"4a", "valid-refs.md", "depends_on", "decision-0001"},
	{"4b", "valid-refs.md", "depends_on", "brief-§3"},
	{"4c", "valid-refs.md", "depends_on", "kodhama/kodhama-0004-uniform-lifecycle"},
	{"4d", "valid-refs.md", "depends_on", "invariants-v0"},
	{"4d", "valid-refs.md", "depends_on", "inv-old-slug"},
	{"4d", "valid-refs.md", "depends_on", "B8"},
	{"4e", "valid-refs.md", "depends_on", "spec-0007"},
	{"4f", "valid-refs.md", "depends_on", "spec-0001@v1"},
	{"6b", "schema.md", "type", "schema"},
}

// TestCorpusConformanceRejectsKnownBadFixture is the positive control (R7).
func TestCorpusConformanceRejectsKnownBadFixture(t *testing.T) {
	contract := loadArtifactContract(t)
	checkOutcomeTableShape(t)
	c, err := newCorpusCheck([]string{knownBadRoot}, contract)
	if err != nil {
		t.Fatalf("the check halted on the fixture corpus: %v", err)
	}
	findings, err := c.run()
	if err != nil {
		t.Fatalf("the check halted on the fixture corpus: %v", err)
	}
	fixturePrefix := corpusDisplayPath(knownBadRoot) + "/"

	got := map[expectedContractFinding][]contractFinding{}
	for _, f := range findings {
		k := expectedContractFinding{strings.TrimPrefix(f.path, fixturePrefix), f.rule, f.subject, f.line}
		got[k] = append(got[k], f)
	}
	want := map[expectedContractFinding]int{}
	for _, e := range knownBadExpected {
		want[e]++
	}
	for _, e := range knownBadExpected {
		if n := len(got[e]); n < want[e] {
			t.Errorf("missing expected finding: %s:%d rule %s naming %q (want %d, got %d) — the rule no longer reports its seeded violation, or reports it on another line", e.file, e.line, e.rule, e.subject, want[e], n)
			want[e] = n // report each shortfall once
		}
	}
	for k, fs := range got {
		for i := want[k]; i < len(fs); i++ {
			t.Errorf("unexpected finding: %s — the fixture corpus seeds no such violation; the rule over-reports, or the expected set is stale", fs[i])
		}
	}
	// One full rendering pins the `path:line: check N (rule): detail` format and
	// the rule-to-check lookup, which the multiset above does not read.
	const wantRendered = "core/fixtures/known-bad/known-bad.md:4: check 4 (4i): "
	if !slices.ContainsFunc(findings, func(f contractFinding) bool { return strings.HasPrefix(f.String(), wantRendered) }) {
		t.Errorf("no finding renders with the prefix %q; the finding format or the rule-to-check lookup changed", wantRendered)
	}

	idx := contractRuleIndex()
	seeded := map[string]bool{}
	for _, e := range knownBadExpected {
		seeded[e.rule] = true
		if idx[e.rule].outcome != outcomeImplemented {
			t.Errorf("knownBadExpected names rule %q, which is not an implemented row of contractOutcomeTable", e.rule)
		}
	}
	accepted := map[string]bool{}
	for _, a := range knownBadAccepted {
		accepted[a.rule] = true
		if idx[a.rule].outcome != outcomeAccepts {
			t.Errorf("knownBadAccepted names rule %q, which is not an accepts row of contractOutcomeTable", a.rule)
		}
	}
	for _, chk := range contractOutcomeTable {
		for _, r := range chk.rules {
			if r.outcome == outcomeImplemented && !seeded[r.rule] {
				t.Errorf("rule %s is implemented but has no seeded violation in knownBadExpected, so deleting its logic would leave this test green (R7)", r.rule)
			}
			if r.outcome == outcomeAccepts && !accepted[r.rule] {
				t.Errorf("rule %s accepts input but has no deliberately valid construct in knownBadAccepted, so a regression that rejects it would go unseen", r.rule)
			}
		}
	}
	checkAcceptedConstructs(t, c, findings, fixturePrefix)
}

// liveCorpusCheck loads the live corpus under corpusRoots for a test that pins
// a parser to live shapes, and fails the test when the load halts.
func liveCorpusCheck(t *testing.T) *corpusCheck {
	t.Helper()
	c, err := newCorpusCheck(corpusRoots, loadArtifactContract(t))
	if err != nil {
		t.Fatalf("the check halted on the live corpus: %v", err)
	}
	return c
}

// liveArtifact returns the live artifact declaring id, and fails the test when
// none does.
func liveArtifact(t *testing.T, c *corpusCheck, id string) *corpusArtifact {
	t.Helper()
	a := c.artifactByID(id)
	if a == nil {
		t.Fatalf("no live artifact declares %s", id)
	}
	return a
}

// TestCorpusConformanceReadsLiveShapes pins the parser to the live constructs
// most likely to break it, so a brittle read fails here by name instead of
// surfacing as a finding against a valid file, or as a silent pass.
func TestCorpusConformanceReadsLiveShapes(t *testing.T) {
	c := liveCorpusCheck(t)

	// A trailing comment holding brackets, colons and pipes ends nothing early
	// and adds nothing to the list.
	want := "decision-0037 decision-0047 decision-0057 decision-0068 decision-0070 decision-0073 decision-0090"
	if f, ok := liveArtifact(t, c, "decision-0091").listValue("depends_on"); !ok || strings.Join(f.list, " ") != want {
		t.Errorf("decision-0091's depends_on parses as %q (well-formed %v), want %q", strings.Join(f.list, " "), ok, want)
	}
	for _, line := range liveArtifact(t, c, "decision-0091").lines[:6] {
		if strings.HasPrefix(line, "depends_on:") && !strings.ContainsAny(line[strings.Index(line, "]")+1:], "[:|") {
			t.Errorf("decision-0091's depends_on comment no longer carries brackets, colons or pipes, so this case tests nothing: %q", line)
		}
	}

	// An untyped scalar is read as written, and a legacy status loses its comment.
	if got := liveArtifact(t, c, "invariants-v1").scalarValue("supersedes"); got != "invariants-v0" {
		t.Errorf("invariants-v1's supersedes reads as %q, want invariants-v0", got)
	}
	if got := liveArtifact(t, c, "schema-typed-artifacts").scalarValue("status"); got != "approved" {
		t.Errorf("schema-typed-artifacts' legacy status reads as %q, want approved with its comment stripped", got)
	}

	// Each live reference form resolves through the form the rubric names.
	if f, ok := liveArtifact(t, c, "decision-0044").listValue("depends_on"); !ok || !strings.Contains(strings.Join(f.list, " "), "kodhama/kodhama-0004-uniform-lifecycle") {
		t.Errorf("decision-0044's depends_on no longer carries kodhama/kodhama-0004-uniform-lifecycle, so the case below tests nothing")
	}
	for _, tc := range []struct {
		ref, form string
		pinned    bool
	}{
		{"decision-0001", "4a", false},
		{"brief-§9.1", "4b", false},
		{"kodhama/kodhama-0004-uniform-lifecycle", "4c", false},
		{"kodhama/kodhama-spec-0003-marketplace-test-observation@v1", "4c", true},
		{"invariants-v0", "4d", false},
		{"spec-0007@v1", "4e", true},
		{"spec-0008@v2", "4e", true},
	} {
		if form, pinned := c.resolveRef(tc.ref); form != tc.form || pinned != tc.pinned {
			t.Errorf("%s resolves as form %q (pinned %v), want form %q (pinned %v)", tc.ref, form, pinned, tc.form, tc.pinned)
		}
	}
}

// copyKnownBadCorpus copies the fixture corpus into a temporary directory, so a
// halt case can remove or rewrite one shared input and keep every other input.
func copyKnownBadCorpus(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	entries, err := os.ReadDir(knownBadRoot)
	if err != nil {
		t.Fatalf("reading the fixture corpus: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			t.Fatalf("the fixture corpus holds a subdirectory, %s, and this copy is flat", e.Name())
		}
		writeFileT(t, filepath.Join(dir, e.Name()), readFileT(t, filepath.Join(knownBadRoot, e.Name())))
	}
	return dir
}

// rewriteFixtureT replaces old with replacement in one copied fixture file, and
// fails if old is absent, so a halt case cannot silently stop mutating.
func rewriteFixtureT(t *testing.T, path, old, replacement string) {
	t.Helper()
	text := readFileT(t, path)
	if !strings.Contains(text, old) {
		t.Fatalf("%s does not contain %q, so this halt case no longer removes the input it names", path, old)
	}
	writeFileT(t, path, strings.Replace(text, old, replacement, 1))
}

// TestCorpusConformanceHaltsOnMissingInput proves the halt policy (KTD6), the
// Honesty clause's "missing/unparseable input → halt loudly, never a partial
// pass". Each case removes one shared input from a copy of the fixture corpus
// and requires an error naming it, with no findings.
func TestCorpusConformanceHaltsOnMissingInput(t *testing.T) {
	contract := loadArtifactContract(t)
	cases := []struct {
		name  string
		setup func(t *testing.T, dir string) []string
		want  string
	}{
		{"a corpus root that does not exist", func(t *testing.T, dir string) []string {
			return []string{dir, filepath.Join(dir, "absent")}
		}, "absent"},
		{"a corpus root that holds no Markdown", func(t *testing.T, dir string) []string {
			empty := t.TempDir()
			writeFileT(t, filepath.Join(empty, "notes.txt"), "not an artifact\n")
			return []string{dir, empty}
		}, "holds no Markdown"},
		{"no artifact declares invariants-v1", func(t *testing.T, dir string) []string {
			removeFileT(t, filepath.Join(dir, "invariants.md"))
			return []string{dir}
		}, "invariants-v1"},
		{"invariants-v1 has no Identifiers section", func(t *testing.T, dir string) []string {
			rewriteFixtureT(t, filepath.Join(dir, "invariants.md"), "## Identifiers (stable slugs)", "## Names")
			return []string{dir}
		}, "Identifiers"},
		{"the Identifiers section has no retired-id column", func(t *testing.T, dir string) []string {
			p := filepath.Join(dir, "invariants.md")
			rewriteFixtureT(t, p, "| Slug | Legacy code (retired) | Note |", "| Slug | Legacy code | Note |")
			rewriteFixtureT(t, p, "| Retired id | → Successor |", "| Old id | → Successor |")
			rewriteFixtureT(t, p, "id resolves to `inv-kept`", "id maps to `inv-kept`")
			return []string{dir}
		}, "retired"},
		{"no artifact declares decision-0079", func(t *testing.T, dir string) []string {
			removeFileT(t, filepath.Join(dir, "decision-0079.md"))
			return []string{dir}
		}, "decision-0079"},
		{"decision-0079 has no 3a registry", func(t *testing.T, dir string) []string {
			rewriteFixtureT(t, filepath.Join(dir, "decision-0079.md"), "**3a. Retired-artifacts registry.**", "**Registry.**")
			return []string{dir}
		}, "3a"},
		{"decision-0079's 3a table has no Retired id column", func(t *testing.T, dir string) []string {
			rewriteFixtureT(t, filepath.Join(dir, "decision-0079.md"), "| Retired id | Was |", "| Id | Was |")
			return []string{dir}
		}, "Retired id"},
		{"a catalog with no Entries section while a profile needs it", func(t *testing.T, dir string) []string {
			rewriteFixtureT(t, filepath.Join(dir, "catalog.md"), "## Entries", "## Items")
			return []string{dir}
		}, "catalog.md"},
		{"no signature-catalog while a profile needs one", func(t *testing.T, dir string) []string {
			removeFileT(t, filepath.Join(dir, "catalog.md"))
			return []string{dir}
		}, "profile.md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			roots := tc.setup(t, copyKnownBadCorpus(t))
			findings, err := checkArtifactCorpus(roots, contract)
			if err == nil {
				t.Fatalf("the check did not halt, and returned %d findings as if the corpus were complete — a partial pass", len(findings))
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the check halted, but its error does not name %q:\n  %v", tc.want, err)
			}
			if findings != nil {
				t.Errorf("a halted run returned %d findings; a halt reports the missing input and nothing else", len(findings))
			}
		})
	}
}

// checkOutcomeTableShape fails when an outcome row, numbered or clause, names
// an outcome outside the vocabulary, carries no reason (R5), or repeats an id.
func checkOutcomeTableShape(t *testing.T) {
	t.Helper()
	vocabulary := map[contractOutcome]bool{
		outcomeImplemented: true, outcomeAccepts: true, outcomeHalts: true, outcomeCovered: true,
		outcomeDropped: true, outcomeGuidance: true, outcomeNoRule: true, outcomeRetired: true,
	}
	rows := append([]contractRuleOutcome{}, contractClauseOutcomes...)
	for _, c := range contractOutcomeTable {
		rows = append(rows, c.rules...)
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if seen[r.rule] {
			t.Errorf("outcome table: %q has two rows", r.rule)
		}
		seen[r.rule] = true
		if !vocabulary[r.outcome] {
			t.Errorf("outcome table: %q has outcome %q, which is not in the Check outcomes vocabulary", r.rule, r.outcome)
		}
		if strings.TrimSpace(r.reason) == "" {
			t.Errorf("outcome table: %q records no reason for its outcome (R5)", r.rule)
		}
	}
}

// checkAcceptedConstructs proves each accepts construct is real: present in its
// file, in a file that yields no finding, and exercising the form it names.
// The unexpected findings themselves are already listed above, so a finding in
// a construct's file is reported here once per file rather than per construct.
func checkAcceptedConstructs(t *testing.T, c *corpusCheck, findings []contractFinding, prefix string) {
	t.Helper()
	reported := map[string]bool{}
	for _, ac := range knownBadAccepted {
		var art *corpusArtifact
		for _, a := range c.arts {
			if a.path == prefix+ac.file {
				art = a
			}
		}
		if art == nil {
			t.Errorf("accepts construct for rule %s: %s is not in the fixture corpus", ac.rule, ac.file)
			continue
		}
		present := false
		for _, f := range art.all(ac.field) {
			present = present || f.scalar == ac.value
			for _, e := range f.list {
				present = present || e == ac.value
			}
		}
		if !present {
			t.Errorf("accepts construct for rule %s: %s carries no %q in `%s`, so the construct proves nothing", ac.rule, ac.file, ac.value, ac.field)
		}
		n := 0
		for _, f := range findings {
			if f.path == art.path {
				n++
			}
		}
		if n > 0 && !reported[ac.file] {
			reported[ac.file] = true
			t.Errorf("%s holds deliberately valid constructs and must yield no finding, and yields %d", ac.file, n)
		}
		switch ac.rule {
		case "6b":
			if sections, ok := c.contract.sections[ac.value]; !ok || len(sections) > 0 {
				t.Errorf("accepts construct for rule 6b: type %q is not exempt in rubric check 6", ac.value)
			}
		case "4f":
			if form, pinned := c.resolveRef(ac.value); form == "" || !pinned {
				t.Errorf("accepts construct for rule 4f: %q should resolve after its pin is stripped, and resolves as form %q, pinned %v", ac.value, form, pinned)
			}
		default:
			if form, pinned := c.resolveRef(ac.value); form != ac.rule || pinned {
				t.Errorf("accepts construct for rule %s: %q resolves as form %q (pinned %v), so it does not exercise the form it names", ac.rule, ac.value, form, pinned)
			}
		}
	}
}

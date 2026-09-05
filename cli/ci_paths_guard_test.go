package main

// The CI-coverage guard (TRL-50). cli-ci.yml is the only workflow that runs
// `go test`, and it is gated on a `paths:` filter. Every file this test suite
// reads out of the parent tree therefore has to appear in that filter, or a PR
// touching only that file skips the single job that covers it — the test still
// passes on `main` later, so nothing is ever red, and the guard silently stops
// guarding for the change that most needed it.
//
// This has now happened twice. First `.github/scripts/decision-id-guard.sh` and
// `decision-id-guard.yml` (cli/decision_id_guard_test.go reads both), fixed by
// adding them to the filter and writing the reasoning into the workflow's header
// comment. Then `core/invariants/trellis-invariants-v1.md` and
// `profiles/trellis-self.md` (cli/row_set_guard_test.go reads both), demonstrated
// live on PR #278: a first commit touching only the invariants file produced
// `parity`, `review` and `release-guard` checks and no `build-test`.
//
// A comment that states a rule is not a check that enforces it (decision-0028: a
// guard per source↔derivative pair). This test is that check. The pair is
// "what the suite reads" ↔ "what the filter lets the suite run on".
//
// HOW READ PATHS ARE FOUND. Three forms, all found empirically by grepping `"../`
// across cli/*_test.go and classifying every hit:
//
//  1. a literal beginning `../` — `readFileT(t, "../profiles/trellis-self.md")`,
//     `os.ReadFile("../README.md")`, `filepath.Abs("../install.sh")`. The helper
//     does not matter; the literal is the evidence.
//  2. `filepath.Join` rooted at `".."` (or at an identifier bound to it, as
//     surface_matrix_guard_test.go's `repoRoot`) with literal components —
//     `filepath.Join("..", ".trellis", "rules.toml")`.
//  3. a call to a helper that itself prefixes the repo root — selfapply_test.go's
//     `readRepoFile`, whose body is `os.ReadFile(filepath.Join("..", name))`, so
//     `readRepoFile("AGENTS.md")` reads AGENTS.md. Such helpers are detected by
//     shape, not by name.
//
// Form 3 has a limit worth stating plainly: when such a helper is called with a
// value the guard cannot resolve statically (selfapply_test.go ranges a map and
// calls `readRepoFile(name)`), guessing would under-report. Instead that FILE is
// swept: every string literal in it that names a path present in the working tree
// is treated as read. Over-approximating a file that already reads the repo root
// dynamically is safe — the worst case is a filter entry we did not strictly need.
//
// os.Stat probes count as reads. selfapply_test.go fails if `.grove/` returns
// (decision-0076); for that to be true of the PR that brings it back, the filter
// has to name it.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// ciWorkflowFile is the workflow whose paths filter gates `go test`.
const ciWorkflowFile = "../.github/workflows/cli-ci.yml"

// ciPathFilterAllowlist holds repo-relative read paths that are deliberately NOT
// in cli-ci.yml's paths filter, each with the reason. It is empty, and the
// intent is that it stays that way: the fix for a new entry here is almost
// always one more filter entry. An allowlist exists at all so that "this read is
// intentionally uncovered" has to be written down and reviewed, rather than
// happening by a filter nobody re-read.
var ciPathFilterAllowlist = map[string]string{}

// readSite is one place the suite reads a repo-relative path.
type readSite struct {
	file string
	line int
}

func (s readSite) String() string { return fmt.Sprintf("%s:%d", s.file, s.line) }

// TestCIPathFilterCoversEveryPathTheSuiteReads is the guard. It fails naming the
// read path, the test file that reads it, and the consequence.
func TestCIPathFilterCoversEveryPathTheSuiteReads(t *testing.T) {
	reads := repoRelativeReads(t)
	if len(reads) == 0 {
		t.Fatal("extracted no repo-relative read paths from cli/*_test.go — the extractor is broken, not the suite; a guard that finds nothing cannot fail")
	}

	workflow := readFileT(t, ciWorkflowFile)
	for _, trigger := range []string{"pull_request", "push"} {
		entries := ciPathFilter(t, workflow, trigger)
		if len(entries) == 0 {
			t.Fatalf("cli-ci.yml's %s trigger declares no paths filter this test can read", trigger)
		}
		matchers := make([]*regexp.Regexp, 0, len(entries))
		for _, e := range entries {
			matchers = append(matchers, ciFilterMatcher(t, e))
		}

		for _, path := range sortedReadPaths(reads) {
			if _, ok := ciPathFilterAllowlist[path]; ok {
				continue
			}
			if ciFilterCovers(matchers, path) {
				continue
			}
			t.Errorf("cli-ci.yml's %s paths filter does not cover %s, which the suite reads at %s — a PR editing %s alone would not run build-test, the only job that runs `go test`, so the guard covering it would not fire on the change that needed it. Add an entry covering %s to BOTH paths lists (or, if the read is meant to be uncovered, an explained entry in ciPathFilterAllowlist)",
				trigger, path, strings.Join(siteStrings(reads[path]), ", "), path, path)
		}
	}
}

// ciFilterCovers reports whether any filter entry selects path. A read path that
// names a directory (or a tree asserted absent, like .grove) is covered when the
// filter selects things INSIDE it, so both shapes are tried.
func ciFilterCovers(matchers []*regexp.Regexp, path string) bool {
	for _, m := range matchers {
		if m.MatchString(path) || m.MatchString(path+"/probe") {
			return true
		}
	}
	return false
}

// ciPathFilter returns the quoted entries of the paths list under the named
// trigger. cli-ci.yml is parsed textually — the cli module is dependency-free
// (decision-0043), so there is no YAML library to reach for, and the two trigger
// sections are bounded by the same landmarks selfapply_test.go uses.
func ciPathFilter(t *testing.T, workflow, trigger string) []string {
	t.Helper()
	start := strings.Index(workflow, "\n  "+trigger+":\n")
	if start < 0 {
		t.Fatalf("cli-ci.yml has no %q trigger section", trigger)
	}
	rest := workflow[start+1:]
	end := len(rest)
	for _, landmark := range []string{"\n  pull_request:\n", "\n  push:\n", "\njobs:\n"} {
		if i := strings.Index(rest, landmark); i > 0 && i < end {
			end = i
		}
	}
	section := rest[:end]
	line := regexp.MustCompile(`(?m)^\s*paths:\s*\[(.*)\]\s*$`).FindStringSubmatch(section)
	if line == nil {
		t.Fatalf("cli-ci.yml's %s trigger has no single-line `paths: [...]` filter; this guard parses that form", trigger)
	}
	var entries []string
	for _, m := range regexp.MustCompile(`"([^"]*)"`).FindAllStringSubmatch(line[1], -1) {
		entries = append(entries, m[1])
	}
	return entries
}

// ciFilterMatcher compiles one GitHub paths-filter entry. Only the constructs
// this repo's filter actually uses are modelled — `**` (any characters), `*`
// (any but `/`), `?`, and literals. Anything else fails loudly rather than being
// matched wrongly: a guard that quietly mis-models its input is worse than none.
func ciFilterMatcher(t *testing.T, entry string) *regexp.Regexp {
	t.Helper()
	if strings.ContainsAny(entry, "![]+") {
		t.Fatalf("cli-ci.yml paths entry %q uses filter syntax this guard does not model — teach ciFilterMatcher the construct before using it, so coverage is not assumed", entry)
	}
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(entry); i++ {
		switch {
		case strings.HasPrefix(entry[i:], "**"):
			b.WriteString(".*")
			i++
		case entry[i] == '*':
			b.WriteString("[^/]*")
		case entry[i] == '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(entry[i])))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

// repoRelativeReads extracts every repo-relative path cli/*_test.go reads, keyed
// by path, valued by the sites that read it. See this file's header for the three
// forms and the one over-approximation.
func repoRelativeReads(t *testing.T) map[string][]readSite {
	t.Helper()
	files, err := filepath.Glob("*_test.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("globbing cli/*_test.go: %v (%d files)", err, len(files))
	}
	sort.Strings(files)

	fset := token.NewFileSet()
	parsed := map[string]*ast.File{}
	for _, f := range files {
		af, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", f, err)
		}
		parsed[f] = af
	}

	// Identifiers bound to ".." anywhere in the package, so `filepath.Join(repoRoot, …)`
	// resolves the same as `filepath.Join("..", …)`.
	repoRootIdents := map[string]bool{}
	for _, af := range parsed {
		ast.Inspect(af, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.ValueSpec:
				for i, name := range d.Names {
					if i < len(d.Values) && stringLit(d.Values[i]) == ".." {
						repoRootIdents[name.Name] = true
					}
				}
			case *ast.AssignStmt:
				for i, lhs := range d.Lhs {
					id, ok := lhs.(*ast.Ident)
					if ok && i < len(d.Rhs) && stringLit(d.Rhs[i]) == ".." {
						repoRootIdents[id.Name] = true
					}
				}
			}
			return true
		})
	}

	// Form 3: helpers that prefix the repo root onto a parameter.
	rootHelpers := map[string]int{}
	for _, af := range parsed {
		ast.Inspect(af, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.FuncDecl:
				if i, ok := rootPrefixingParam(d.Type, d.Body, repoRootIdents); ok {
					rootHelpers[d.Name.Name] = i
				}
			case *ast.AssignStmt:
				for i, rhs := range d.Rhs {
					lit, ok := rhs.(*ast.FuncLit)
					if !ok || i >= len(d.Lhs) {
						continue
					}
					id, ok := d.Lhs[i].(*ast.Ident)
					if !ok {
						continue
					}
					if p, ok := rootPrefixingParam(lit.Type, lit.Body, repoRootIdents); ok {
						rootHelpers[id.Name] = p
					}
				}
			}
			return true
		})
	}

	reads := map[string][]readSite{}
	add := func(path, file string, pos token.Pos) {
		path = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(path)), "./")
		if path == "" || path == "." || path == ".." || strings.HasPrefix(path, "../") || strings.HasPrefix(path, "/") {
			return // the repo root itself, or outside it: not a filterable path
		}
		if strings.HasPrefix(path, "cli/") || path == "cli" {
			return // cli/** is in the filter by construction
		}
		reads[path] = append(reads[path], readSite{file: file, line: fset.Position(pos).Line})
	}

	sweep := map[string]bool{}
	for _, f := range files {
		ast.Inspect(parsed[f], func(n ast.Node) bool {
			// Form 1.
			if bl, ok := n.(*ast.BasicLit); ok {
				if s, ok := unquote(bl); ok && strings.HasPrefix(s, "../") && looksLikePath(s) {
					add(strings.TrimPrefix(s, "../"), f, bl.Pos())
				}
				return true
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			// Form 2.
			if isFilepathJoin(call.Fun) && len(call.Args) > 1 && rootExpr(call.Args[0], repoRootIdents) {
				parts := []string{}
				for _, a := range call.Args[1:] {
					s := stringLit(a)
					if s == "" {
						parts = nil
						break
					}
					parts = append(parts, s)
				}
				if len(parts) > 0 {
					add(strings.Join(parts, "/"), f, call.Pos())
				}
			}
			// Form 3.
			if id, ok := call.Fun.(*ast.Ident); ok {
				if idx, ok := rootHelpers[id.Name]; ok && idx < len(call.Args) {
					if s := stringLit(call.Args[idx]); s != "" && looksLikePath(s) {
						add(strings.TrimPrefix(s, "../"), f, call.Pos())
					} else if s == "" {
						sweep[f] = true // unresolvable argument: over-approximate this file
					}
				}
			}
			return true
		})
	}

	// The over-approximation, applied only to files that read the repo root through
	// a value this guard could not resolve.
	for f := range sweep {
		ast.Inspect(parsed[f], func(n ast.Node) bool {
			bl, ok := n.(*ast.BasicLit)
			if !ok {
				return true
			}
			s, ok := unquote(bl)
			if !ok || !looksLikePath(s) {
				return true
			}
			candidate := strings.TrimPrefix(s, "../")
			if candidate == "" || strings.HasPrefix(candidate, "..") {
				return true
			}
			if _, err := os.Stat(filepath.Join("..", candidate)); err == nil {
				add(candidate, f, bl.Pos())
			}
			return true
		})
	}
	return reads
}

// rootPrefixingParam reports the index of the parameter a function joins onto the
// repo root, i.e. the shape `filepath.Join("..", name)`.
func rootPrefixingParam(sig *ast.FuncType, body *ast.BlockStmt, roots map[string]bool) (int, bool) {
	if sig == nil || body == nil || sig.Params == nil {
		return 0, false
	}
	var names []string
	for _, field := range sig.Params.List {
		for _, n := range field.Names {
			names = append(names, n.Name)
		}
	}
	found, idx := false, 0
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isFilepathJoin(call.Fun) || len(call.Args) != 2 || !rootExpr(call.Args[0], roots) {
			return true
		}
		arg, ok := call.Args[1].(*ast.Ident)
		if !ok {
			return true
		}
		for i, name := range names {
			if name == arg.Name {
				found, idx = true, i
				return false
			}
		}
		return true
	})
	return idx, found
}

func isFilepathJoin(fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Join" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "filepath"
}

func rootExpr(e ast.Expr, roots map[string]bool) bool {
	if stringLit(e) == ".." {
		return true
	}
	id, ok := e.(*ast.Ident)
	return ok && roots[id.Name]
}

// stringLit returns the value of a string literal expression, or "" if e is not one.
func stringLit(e ast.Expr) string {
	bl, ok := e.(*ast.BasicLit)
	if !ok {
		return ""
	}
	s, ok := unquote(bl)
	if !ok {
		return ""
	}
	return s
}

func unquote(bl *ast.BasicLit) (string, bool) {
	if bl.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(bl.Value)
	if err != nil {
		return "", false
	}
	return s, true
}

// looksLikePath rejects prose, shell fixtures and multi-line blobs, which is what
// most string literals in this suite are.
func looksLikePath(s string) bool {
	if s == "" || len(s) > 200 {
		return false
	}
	return !strings.ContainsAny(s, " \t\n\r\"'`$;|<>()") && s != ".." && s != "../"
}

func sortedReadPaths(m map[string][]readSite) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func siteStrings(sites []readSite) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range sites {
		if !seen[s.file] {
			seen[s.file] = true
			out = append(out, s.String())
		}
	}
	sort.Strings(out)
	return out
}

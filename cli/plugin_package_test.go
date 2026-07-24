package main

// TestPluginPackageParity guards decision-0061 §§1–4 and
// kodhama/kodhama-spec-0003-marketplace-test-observation@v1 R1/R5.
// It is deliberately a repository-local parity guard, not a release engine:
// no other product version, tag, marketplace, or network service is read.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

const trellisPluginDir = "../plugins/trellis"

var (
	semverPattern     = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	surfaceIDPattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]{0,127}$`)
	repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	shaPattern        = regexp.MustCompile(`^[0-9a-f]{40}$`)
	stepIDPattern     = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]{0,127}$`)
	observedAtPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$`)
)

type pluginManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Skills  string `json:"skills,omitempty"`
	Hooks   string `json:"hooks,omitempty"`
}

type surfaceContract struct {
	SchemaVersion int          `json:"schema_version"`
	Version       string       `json:"version"`
	Rows          []surfaceRow `json:"rows"`
}

type surfaceRow struct {
	SurfaceID                   string   `json:"surface_id"`
	Host                        string   `json:"host"`
	BehaviorState               string   `json:"behavior_state"`
	BehaviorContracts           []string `json:"behavior_contracts"`
	MarketplaceTestObservations []string `json:"marketplace_test_observations"`
	Disclosure                  string   `json:"disclosure"`
}

type marketplaceObservation struct {
	SchemaVersion int                    `json:"schema_version"`
	Host          string                 `json:"host"`
	SurfaceID     string                 `json:"surface_id"`
	Marketplace   observationMarketplace `json:"marketplace"`
	Execution     observationExecution   `json:"execution"`
	ObservedAt    string                 `json:"observed_at"`
}

type observationMarketplace struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
}

type observationExecution struct {
	Repository  string `json:"repository"`
	Commit      string `json:"commit"`
	Workflow    string `json:"workflow"`
	Job         string `json:"job"`
	RunID       int64  `json:"run_id"`
	RunAttempt  int64  `json:"run_attempt"`
	SetupStepID string `json:"setup_step_id"`
}

func readPluginFile(t *testing.T, rel string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(trellisPluginDir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read plugin file %s: %v", rel, err)
	}
	return content
}

func decodeJSON[T any](raw []byte, label string, closed bool) (T, error) {
	var value T
	if !utf8.Valid(raw) {
		return value, fmt.Errorf("%s: JSON is not valid UTF-8", label)
	}
	if err := validateJSONUnicodeEscapes(raw); err != nil {
		return value, fmt.Errorf("%s: %w", label, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if closed {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("%s: %w", label, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return value, fmt.Errorf("%s: trailing JSON value", label)
		}
		return value, fmt.Errorf("%s: trailing data: %w", label, err)
	}
	return value, nil
}

func jsonHex4(raw []byte) (uint16, bool) {
	if len(raw) < 4 {
		return 0, false
	}
	var value uint16
	for _, digit := range raw[:4] {
		value *= 16
		switch {
		case digit >= '0' && digit <= '9':
			value += uint16(digit - '0')
		case digit >= 'a' && digit <= 'f':
			value += uint16(digit-'a') + 10
		case digit >= 'A' && digit <= 'F':
			value += uint16(digit-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

func validateJSONUnicodeEscapes(raw []byte) error {
	inString := false
	for index := 0; index < len(raw); index++ {
		switch raw[index] {
		case '"':
			inString = !inString
		case '\\':
			if !inString || index+1 >= len(raw) {
				continue
			}
			if raw[index+1] != 'u' {
				index++
				continue
			}
			codePoint, ok := jsonHex4(raw[index+2:])
			if !ok {
				return fmt.Errorf("invalid Unicode escape")
			}
			index += 5
			switch {
			case codePoint >= 0xd800 && codePoint <= 0xdbff:
				if index+6 >= len(raw) || raw[index+1] != '\\' || raw[index+2] != 'u' {
					return fmt.Errorf("unpaired high-surrogate escape")
				}
				low, ok := jsonHex4(raw[index+3:])
				if !ok || low < 0xdc00 || low > 0xdfff {
					return fmt.Errorf("unpaired high-surrogate escape")
				}
				index += 6
			case codePoint >= 0xdc00 && codePoint <= 0xdfff:
				return fmt.Errorf("unpaired low-surrogate escape")
			}
		}
	}
	return nil
}

func validateManifestPath(root, declared, label string) error {
	if declared == "" {
		return nil
	}
	clean := filepath.Clean(filepath.FromSlash(declared))
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s escapes the plugin root: %q", label, declared)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("%s plugin root: %w", label, err)
	}
	realRoot, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return fmt.Errorf("%s plugin root: %w", label, err)
	}
	realTarget, err := filepath.EvalSymlinks(filepath.Join(rootAbs, clean))
	if err != nil {
		return fmt.Errorf("%s target %q: %w", label, declared, err)
	}
	relative, err := filepath.Rel(realRoot, realTarget)
	if err != nil {
		return fmt.Errorf("%s target %q: %w", label, declared, err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s target %q resolves outside the plugin root", label, declared)
	}
	return nil
}

func validateObservation(raw []byte, label, host, surfaceID string) error {
	observation, err := decodeJSON[marketplaceObservation](raw, label, true)
	if err != nil {
		return err
	}
	if observation.SchemaVersion != 1 {
		return fmt.Errorf("%s: schema_version must be 1", label)
	}
	if observation.Host != host || observation.SurfaceID != surfaceID {
		return fmt.Errorf("%s: host/surface_id do not match their surface row", label)
	}
	if host != "claude" && host != "codex" {
		return fmt.Errorf("%s: invalid host %q", label, host)
	}
	if !surfaceIDPattern.MatchString(surfaceID) {
		return fmt.Errorf("%s: invalid surface_id %q", label, surfaceID)
	}
	if strings.TrimSpace(observation.Marketplace.Name) == "" ||
		!utf8.ValidString(observation.Marketplace.Name) ||
		len([]byte(observation.Marketplace.Name)) > 128 {
		return fmt.Errorf("%s: invalid marketplace name", label)
	}
	if !repositoryPattern.MatchString(observation.Marketplace.Repository) ||
		!repositoryPattern.MatchString(observation.Execution.Repository) {
		return fmt.Errorf("%s: invalid repository", label)
	}
	if !shaPattern.MatchString(observation.Marketplace.Revision) ||
		!shaPattern.MatchString(observation.Execution.Commit) {
		return fmt.Errorf("%s: invalid commit", label)
	}
	workflow := observation.Execution.Workflow
	cleanWorkflow := filepath.ToSlash(filepath.Clean(filepath.FromSlash(workflow)))
	if cleanWorkflow != workflow ||
		strings.Contains(workflow, `\`) ||
		(!strings.HasSuffix(workflow, ".yml") && !strings.HasSuffix(workflow, ".yaml")) ||
		!strings.HasPrefix(workflow, ".github/workflows/") {
		return fmt.Errorf("%s: invalid workflow path", label)
	}
	if strings.TrimSpace(observation.Execution.Job) == "" ||
		!utf8.ValidString(observation.Execution.Job) ||
		len([]byte(observation.Execution.Job)) > 128 {
		return fmt.Errorf("%s: invalid job", label)
	}
	if observation.Execution.RunID <= 0 || observation.Execution.RunAttempt <= 0 {
		return fmt.Errorf("%s: run identifiers must be positive", label)
	}
	if !stepIDPattern.MatchString(observation.Execution.SetupStepID) {
		return fmt.Errorf("%s: invalid setup_step_id", label)
	}
	if !observedAtPattern.MatchString(observation.ObservedAt) {
		return fmt.Errorf("%s: observed_at must use exact millisecond UTC syntax", label)
	}
	if _, err := time.Parse("2006-01-02T15:04:05.000Z", observation.ObservedAt); err != nil {
		return fmt.Errorf("%s: observed_at must be a real millisecond UTC instant", label)
	}
	return nil
}

func validateInitialSurfaceRows(rows []surfaceRow) error {
	if len(rows) != 2 {
		return fmt.Errorf("surfaces.json must contain exactly two Phase-1 rows, got %d", len(rows))
	}
	wantRows := map[string]struct {
		host               string
		disclosureFragment string
	}{
		"claude-interactive": {
			host:               "claude",
			disclosureFragment: "marketplace registration for 0.2.0 is not yet evidenced",
		},
		"codex-cli-local-startup": {
			host:               "codex",
			disclosureFragment: "marketplace registration for 0.2.0 is not yet evidenced",
		},
	}
	seen := map[string]bool{}
	for index, row := range rows {
		label := fmt.Sprintf("surfaces.json rows[%d]", index)
		want, ok := wantRows[row.SurfaceID]
		if !ok || seen[row.SurfaceID] {
			return fmt.Errorf("%s: unexpected or duplicate surface_id %q", label, row.SurfaceID)
		}
		seen[row.SurfaceID] = true
		if !surfaceIDPattern.MatchString(row.SurfaceID) || row.Host != want.host {
			return fmt.Errorf("%s: invalid surface identity or host", label)
		}
		if row.BehaviorState != "supported" {
			return fmt.Errorf("%s: initial Phase-1 row must be supported", label)
		}
		if len(row.BehaviorContracts) != 2 ||
			row.BehaviorContracts[0] != "decision-0058" ||
			row.BehaviorContracts[1] != "spec-0007@v1" {
			return fmt.Errorf("%s: behavior contracts drifted: %v", label, row.BehaviorContracts)
		}
		if row.MarketplaceTestObservations == nil {
			return fmt.Errorf("%s: marketplace_test_observations is required and must be an array", label)
		}
		if strings.TrimSpace(row.Disclosure) == "" ||
			!strings.Contains(row.Disclosure, want.disclosureFragment) {
			return fmt.Errorf("%s: disclosure must distinguish marketplace evidence", label)
		}
	}
	return nil
}

func validObservationJSON() []byte {
	return []byte(`{
  "schema_version": 1,
  "host": "codex",
  "surface_id": "codex-cli-local-startup",
  "marketplace": {
    "name": "kodhama",
    "repository": "kodhama/stewards",
    "revision": "0123456789abcdef0123456789abcdef01234567"
  },
  "execution": {
    "repository": "kodhama/trellis",
    "commit": "89abcdef0123456789abcdef0123456789abcdef",
    "workflow": ".github/workflows/ci.yml",
    "job": "codex-marketplace",
    "run_id": 123456789,
    "run_attempt": 1,
    "setup_step_id": "stewards_marketplace_kodhama_codex"
  },
  "observed_at": "2026-07-24T12:34:56.789Z"
}`)
}

func TestPackageValidatorsRejectMalformedMetadata(t *testing.T) {
	t.Run("SemVer permits digit-leading alphanumeric prerelease identifiers", func(t *testing.T) {
		if !semverPattern.MatchString("1.2.3-1alpha") {
			t.Fatal("canonical SemVer must accept a nonnumeric prerelease identifier that begins with a digit")
		}
	})

	t.Run("manifest symlink cannot escape plugin root", func(t *testing.T) {
		root := t.TempDir()
		outside := t.TempDir()
		target := filepath.Join(outside, "hooks.json")
		if err := os.WriteFile(target, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(root, "hooks.json")); err != nil {
			t.Fatal(err)
		}
		if err := validateManifestPath(root, "hooks.json", "hooks"); err == nil {
			t.Fatal("manifest symlink escaping the plugin root must fail")
		}
	})

	t.Run("surface contracts remain two distinct entries", func(t *testing.T) {
		rows := []surfaceRow{
			{
				SurfaceID:                   "claude-interactive",
				Host:                        "claude",
				BehaviorState:               "supported",
				BehaviorContracts:           []string{"decision-0058,spec-0007@v1"},
				MarketplaceTestObservations: []string{},
				Disclosure:                  "marketplace registration for 0.2.0 is not yet evidenced",
			},
			{
				SurfaceID:                   "codex-cli-local-startup",
				Host:                        "codex",
				BehaviorState:               "supported",
				BehaviorContracts:           []string{"decision-0058", "spec-0007@v1"},
				MarketplaceTestObservations: []string{},
				Disclosure:                  "marketplace registration for 0.2.0 is not yet evidenced",
			},
		}
		if err := validateInitialSurfaceRows(rows); err == nil {
			t.Fatal("joined contract identifiers must not masquerade as two entries")
		}
	})

	t.Run("observation array is required", func(t *testing.T) {
		rows := []surfaceRow{
			{
				SurfaceID:                   "claude-interactive",
				Host:                        "claude",
				BehaviorState:               "supported",
				BehaviorContracts:           []string{"decision-0058", "spec-0007@v1"},
				MarketplaceTestObservations: nil,
				Disclosure:                  "marketplace registration for 0.2.0 is not yet evidenced",
			},
			{
				SurfaceID:                   "codex-cli-local-startup",
				Host:                        "codex",
				BehaviorState:               "supported",
				BehaviorContracts:           []string{"decision-0058", "spec-0007@v1"},
				MarketplaceTestObservations: []string{},
				Disclosure:                  "marketplace registration for 0.2.0 is not yet evidenced",
			},
		}
		if err := validateInitialSurfaceRows(rows); err == nil {
			t.Fatal("missing or null marketplace_test_observations must fail")
		}
	})

	t.Run("timestamp requires period separator", func(t *testing.T) {
		raw := bytes.Replace(validObservationJSON(), []byte(".789Z"), []byte(",789Z"), 1)
		if err := validateObservation(raw, "observation", "codex", "codex-cli-local-startup"); err == nil {
			t.Fatal("comma fractional-second separator must fail")
		}
	})

	t.Run("raw JSON must be valid UTF-8", func(t *testing.T) {
		raw := bytes.Replace(validObservationJSON(), []byte("kodhama"), []byte{0xff}, 1)
		if err := validateObservation(raw, "observation", "codex", "codex-cli-local-startup"); err == nil {
			t.Fatal("malformed UTF-8 must fail before JSON replacement")
		}
	})

	t.Run("JSON escapes require valid surrogate pairs", func(t *testing.T) {
		for _, invalid := range []string{`\ud800`, `\udfff`} {
			raw := bytes.Replace(validObservationJSON(), []byte("kodhama"), []byte(invalid), 1)
			if err := validateObservation(raw, "observation", "codex", "codex-cli-local-startup"); err == nil {
				t.Errorf("unpaired surrogate escape %q must fail", invalid)
			}
		}
		paired := bytes.Replace(validObservationJSON(), []byte("kodhama"), []byte(`\ud83d\ude80`), 1)
		if err := validateObservation(paired, "observation", "codex", "codex-cli-local-startup"); err != nil {
			t.Errorf("valid surrogate pair must pass: %v", err)
		}
	})
}

func TestPluginPackageParity(t *testing.T) {
	versionBytes := readPluginFile(t, "VERSION")
	if string(versionBytes) != "0.2.0\n" {
		t.Fatalf("VERSION must contain exactly 0.2.0 plus LF, got %q", versionBytes)
	}
	version := strings.TrimSuffix(string(versionBytes), "\n")
	if !semverPattern.MatchString(version) {
		t.Fatalf("VERSION is not canonical SemVer: %q", version)
	}

	for _, rel := range []string{
		".claude-plugin/plugin.json",
		".codex-plugin/plugin.json",
	} {
		manifest, err := decodeJSON[pluginManifest](readPluginFile(t, rel), rel, false)
		if err != nil {
			t.Fatal(err)
		}
		if manifest.Name != "trellis" || manifest.Version != version {
			t.Errorf("%s must declare trellis@%s, got %s@%s", rel, version, manifest.Name, manifest.Version)
		}
		if err := validateManifestPath(trellisPluginDir, manifest.Skills, rel+" skills"); err != nil {
			t.Error(err)
		}
		if err := validateManifestPath(trellisPluginDir, manifest.Hooks, rel+" hooks"); err != nil {
			t.Error(err)
		}
	}

	surfaces, err := decodeJSON[surfaceContract](readPluginFile(t, "surfaces.json"), "surfaces.json", true)
	if err != nil {
		t.Fatal(err)
	}
	if surfaces.SchemaVersion != 1 || surfaces.Version != version {
		t.Fatalf("surfaces.json must carry schema 1 and package version %s", version)
	}
	if err := validateInitialSurfaceRows(surfaces.Rows); err != nil {
		t.Fatal(err)
	}
	for index, row := range surfaces.Rows {
		label := fmt.Sprintf("surfaces.json rows[%d]", index)
		for observationIndex, rel := range row.MarketplaceTestObservations {
			observationLabel := fmt.Sprintf("%s observation[%d]", label, observationIndex)
			clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(rel)))
			if filepath.IsAbs(filepath.FromSlash(rel)) ||
				clean == ".." ||
				strings.HasPrefix(clean, "../") ||
				strings.Contains(rel, `\`) ||
				clean != rel ||
				!strings.HasSuffix(rel, ".json") {
				t.Errorf("%s: path must be normalized repository-relative JSON: %q", observationLabel, rel)
				continue
			}
			raw, err := os.ReadFile(filepath.Join("..", filepath.FromSlash(rel)))
			if err != nil {
				t.Errorf("%s: %v", observationLabel, err)
				continue
			}
			if err := validateObservation(raw, observationLabel, row.Host, row.SurfaceID); err != nil {
				t.Error(err)
			}
		}
	}
}

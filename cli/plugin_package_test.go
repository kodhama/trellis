package main

// TestPluginPackageParity guards what survives of decision-0061 §§1-4 after
// decision-0066 superseded §3 and three of §4's six bullets in part: VERSION as
// canonical SemVer, and cross-manifest identity and version equality for the two
// host manifests. The surface matrix the other three bullets validated is retired,
// so nothing here reads a surface row, a behavior state, or a marketplace
// observation. It is deliberately a repository-local parity guard, not a release
// engine: no other product version, tag, marketplace, or network service is read.

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
	"unicode/utf8"
)

const trellisPluginDir = "../plugins/trellis"

var (
	semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
)

type pluginManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Skills  string `json:"skills,omitempty"`
	Hooks   string `json:"hooks,omitempty"`
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

// validPluginManifestJSON is the fixture the two JSON-hardening subtests below
// use. It replaces the retired marketplace-observation fixture (decision-0066 §4):
// those subtests never exercised anything observation-specific — what they test is
// validateJSONUnicodeEscapes and decodeJSON's UTF-8 check, and decodeJSON stays
// live for both host manifests in TestPluginPackageParity. Re-pointing them at a
// plugin manifest keeps the coverage on the code that still runs.
func validPluginManifestJSON() []byte {
	return []byte(`{
  "name": "trellis",
  "version": "0.2.0",
  "skills": "skills",
  "hooks": "hooks/hooks.json"
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

	t.Run("raw JSON must be valid UTF-8", func(t *testing.T) {
		raw := bytes.Replace(validPluginManifestJSON(), []byte("trellis"), []byte{0xff}, 1)
		if _, err := decodeJSON[pluginManifest](raw, "plugin.json", false); err == nil {
			t.Fatal("malformed UTF-8 must fail before JSON replacement")
		}
	})

	t.Run("JSON escapes require valid surrogate pairs", func(t *testing.T) {
		for _, invalid := range []string{`\ud800`, `\udfff`} {
			raw := bytes.Replace(validPluginManifestJSON(), []byte("trellis"), []byte(invalid), 1)
			if _, err := decodeJSON[pluginManifest](raw, "plugin.json", false); err == nil {
				t.Errorf("unpaired surrogate escape %q must fail", invalid)
			}
		}
		paired := bytes.Replace(validPluginManifestJSON(), []byte("trellis"), []byte(`\ud83d\ude80`), 1)
		if _, err := decodeJSON[pluginManifest](paired, "plugin.json", false); err != nil {
			t.Errorf("valid surrogate pair must pass: %v", err)
		}
	})
}

func TestPluginPackageParity(t *testing.T) {
	versionBytes := readPluginFile(t, "VERSION")
	if string(versionBytes) != "0.4.0\n" {
		t.Fatalf("VERSION must contain exactly 0.4.0 plus LF, got %q", versionBytes)
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
}

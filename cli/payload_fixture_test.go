package main

import "testing"

// payloadFile returns one generated payload file by name and fails the test when
// the payload no longer ships that name (TRL-97, KTD15).
//
// A bare payloadFiles()[name] yields "" for a key the payload retired, and an
// empty .trellis/rules.toml is a valid file meaning every rule applies. A fixture
// built on a retired preset therefore stopped testing what it named and went on
// asserting the defaults, which is how the retirement of rules-a.toml and
// trellis-a.md turned forty tests red for reasons none of them was about. Every
// lookup goes through here so the next retirement fails at the lookup instead.
func payloadFile(t *testing.T, name string) string {
	t.Helper()
	body, ok := payloadFiles()[name]
	if !ok {
		t.Fatalf("the payload ships no %s, so a fixture built on it would silently read an empty file", name)
	}
	return body
}

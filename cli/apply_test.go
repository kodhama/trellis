package main

import (
	"testing"
)

// TestExtractEntriesSectionFailsLoud guards decision-0055 point 1,
// code-reviewer's confirmed HIGH-severity finding on the decision-0055 build:
// either boundary heading going missing — "## Entries" not found at all, or
// found but "## Acceptance criteria" not matched exactly after it (renamed,
// re-cased, reformatted — exactly what decision-0056 will eventually do to
// this same catalog file) — must fail exactly as loudly as the other, never
// silently return a best-effort partial or full string that leaks the
// catalog's stale tail (Open questions etc.) into the payload. These are
// independent, isolated assertions on the function itself — deliberately not
// derived via a second call to extractEntriesSection the way sync_test.go's
// "want" is, which is exactly the shared-oracle blind spot that let the
// original defect ship undetected. Table-driven (code-reviewer's low-severity
// finding on the decision-0055 fix pass: the two cases were near-identical
// defer/recover/type-assert bodies before this consolidation).
//
// Each case asserts the FULL expected panic message, not a bare substring:
// both sibling messages mention "Entries" (one names "## Entries" as the
// missing heading, the other names "## Acceptance criteria" but still refers
// to "## Entries" in its "not found after" clause), so a substring check on
// "Entries" alone would not catch a message-mixup regression (e.g. swapping
// the two panic call sites) — code-reviewer proved this empirically by
// swapping the two panic messages and showing a prior "Entries"-substring
// assertion still passed on the wrong message. Full-message equality is what
// actually discriminates the two (re-verified against this consolidated form
// the same way: swap the two panic messages in apply.go, confirm both cases
// below fail, then revert).
func TestExtractEntriesSectionFailsLoud(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "missing ## Acceptance criteria after a present ## Entries",
			// "## Entries" present and matched; "## Acceptance criteria" deliberately
			// renamed — replicating code-reviewer's synthetic probe.
			in: "preamble text\n\n## Entries\nentry body here\n\n" +
				"## Acceptance Criteria (renamed)\nstale tail content that must never leak through\n",
			want: `extractEntriesSection: "## Acceptance criteria" heading not found after "## Entries" — the catalog's shape changed underneath the extraction boundary (decision-0055 point 1)`,
		},
		{
			name: "missing ## Entries entirely (the symmetric first boundary)",
			in:   "no entries heading anywhere in this document\n\n## Acceptance criteria\nsomething\n",
			want: `extractEntriesSection: "## Entries" heading not found — the catalog's shape changed underneath the extraction boundary (decision-0055 point 1)`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("extractEntriesSection did not fail on the missing heading — " +
						"it silently returned a best-effort string instead of failing loudly")
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("panic value is not a string: %v", r)
				}
				if msg != tc.want {
					t.Fatalf("panic message = %q, want %q", msg, tc.want)
				}
			}()
			got := extractEntriesSection(tc.in)
			t.Fatalf("extractEntriesSection returned a value instead of panicking: %q", got)
		})
	}
}

// TestExtractEntriesSectionHappyPath guards decision-0055 point 1's success
// path stays correct while the failure path is hardened: a well-formed input
// with both boundary headings present still yields exactly the text strictly
// between "## Entries" (inclusive) and "## Acceptance criteria" (exclusive).
func TestExtractEntriesSectionHappyPath(t *testing.T) {
	in := "preamble text\n\n## Entries\nentry body here\n\n## Acceptance criteria\ntail content\n"
	want := "## Entries\nentry body here\n\n"
	if got := extractEntriesSection(in); got != want {
		t.Fatalf("extractEntriesSection(happy path) = %q, want %q", got, want)
	}
}

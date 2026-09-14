# `cli/testdata/` — positive controls

Deliberately **broken** artifacts. Their job is to be **rejected**: the conformance check is
trusted only after it flags every violation here (`rubric-artifact-contract` — *the verifier must be
demonstrably able to fail*, the B3 positive-control lesson logged from our CI episode).

**Excluded from normal corpus runs** — they are test data, not real artifacts.

## The fixture corpus — `known-bad/`

The positive control is the fixture corpus under `known-bad/`. It is self-contained: it carries its
own `invariants-v1`, `decision-0079`, catalog and profile, so the check needs nothing from the live
corpus to run over it. Most files seed violations. A few are deliberately valid constructs that must
yield no finding, such as the accepted reference forms in `valid-refs.md`. `known-bad.md`, the
original four-violation fixture, is one file among them.

## Answer key

The answer key is code, not prose (`decision-0098`): the exact expected-findings set in
`TestCorpusConformanceRejectsKnownBadFixture`, in `cli/corpus_conformance_test.go`. That test runs
the check over `known-bad/` and fails on any expected finding that is missing and on any finding the
set does not name. It also fails when an implemented rule has no seeded violation here, so deleting
a rule's logic turns it red. A violation seeded here without its entry in the set fails the test,
and so does an entry with no violation behind it.

---
title: A dated note moves line citations that a one-line search misses
date: 2026-10-10
category: workflow-issues
module: docs/decisions
problem_type: workflow_issue
component: documentation
severity: medium
applies_when:
  - Adding a dated note under a decision record's title
  - Adding or removing any line above the body of a record that other files cite by line number
  - Reviewing a PR that corrects line citations into a record
tags: [dated-note, line-citations, decision-records, append-only, review]
---

# A dated note moves line citations that a one-line search misses

## Context

A dated note goes directly under a record's title, so the note and its blank line move the record's whole body down two lines. `AGENTS.md` requires the same PR to correct every live line citation into that record. No test checks line citations, so a missed one passes CI.

TRL-83 (PR #327) added notes to `decision-0078` and `decision-0083`. Its first sweep searched for the record's id followed, on the same line, by a `:NN` token. That found seven citations into `decision-0078` and none into `decision-0083`. It missed one into `decision-0083`, which an independent review then caught:

```text
docs/decisions/0087-one-gateway-for-every-payload-read.md:64-65
  project's rows were both perfectly well (`decision-0083` records the general property at
  `:117-118`; the specific fix is in the hook).
```

The id is on one line and the range on the next. Three separate one-line searches missed it: the author's, and two of the reviewer's.

The same review found a second fault. `decision-0084:366` cited `decision-0078` one line above the quote it cites. It had been one short since PR #266 added a frontmatter line to `decision-0078`. PR #316 and the first commit of PR #327 each added the right delta to the number, and so each carried the old error forward.

## Guidance

1. **Search for bare tokens as well as attached ones.** A citation can continue on the next line, or follow an earlier citation as `` `:127-128` `` with no id beside it. This lists every bare line-range token outside the plans:

   ```bash
   git grep -nE '(^|[ (])`:[0-9]{2,3}(-[0-9]{1,3})?`' -- . \
     ':(exclude)docs/plans' ':(exclude)docs/superpowers' ':(exclude)docs/ideation'
   ```

   It returned 83 hits at `60c3640`, the head of PR #327 before this document was added, and `decision-0087:65` was among them. This document's own examples add 4 more. Read the lines above each hit to see which record it cites.

2. **Open the target, never add the delta.** For each citation, open the cited record at the new line numbers and read the text. Adding two to the old number keeps any error the old number already had.

3. **Use the last note's PR as the precedent for what is live.** `git show <merge commit> -- docs/decisions` on the PR that added the record's previous note shows which citations that PR moved. PR #316 (`6f292aa`) moved exactly the citations into `decision-0078` and `decision-0083` that PR #327 had to move again.

4. **Leave dated citations alone.** These are not live under `AGENTS.md`, and PR #316 left them too:
   - a citation that dates itself, as `decision-0087:48-49` does with "line numbers as of this record";
   - a sweep table, as at `decision-0092:211`;
   - a record's account of a past review, as at `decision-0089:236-240`;
   - anything under `docs/plans/`, `docs/superpowers/` or `docs/ideation/`.

5. **Correcting a citation inside a record is allowed.** `AGENTS.md` names "correcting a line citation" as one of the few permitted edits to a record's existing text. `decision-0078` cites its own body from its frontmatter, so its own note moved that citation too.

## Why This Matters

A wrong line citation sends the reader to a nearby sentence that does not say what the citing record claims. The record is append-only, so the reader has no other signal that the pointer drifted. Each uncorrected citation also becomes the base the next note's author adds a delta to.

## When to Apply

Any change that adds a dated note, a `superseded_by` line, or any other line above text that another file cites by number. The same applies to a non-record file that is cited by line, such as the catalog (`decision-0081` cites `core/catalog/signature-catalog-v1.md:189-193`, and dates the citation).

## Examples

PR #327's corrections, each checked against the target lines:

| Citing line | Before | After | Target text |
|---|---|---|---|
| `decision-0087:65` | `:117-118` | `:119-120` | "That distinction is what keeps the guard from over-refusing" |
| `decision-0084:366` | `0078:49` | `0078:52` | the line where the quoted sentence begins |
| `decision-0087:22`, `decision-0089:37,102,210,214`, `decision-0078:7` | `:149-158` | `:151-160` | the "Renumbered 0077 → 0078 mid-flight" bullet |

## Related

- `AGENTS.md`, the append-only rule and "The PR that adds a note also corrects every live line citation into that record".
- A related trap from the same fix: `cli/assets/invariants.md` is a byte-identical copy of `core/catalog/signature-catalog-v1.md` (`TestBundledCatalogInSync`), while `plugins/trellis/reference/invariants.md` copies the entries section only. An edit above `## Entries` needs the CLI copy regenerated and is not a payload change, so it needs no `plugins/trellis/VERSION` bump.

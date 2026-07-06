---
title: Export design
status: draft
---

# Export design  *(DRAFT — not yet approved)*

## `GET /export`

Return the full list of items for offline backup.

- Path: `GET /export`
- Auth: none for now.
- Output format: **TODO — still deciding between CSV and JSON. Undecided.** Product wants CSV for
  spreadsheet users; eng prefers JSON for round-tripping. Not resolved yet.
- Filename / headers: TBD once the format is chosen.

> This document is a draft. The output-format decision above is open and will change the response shape.

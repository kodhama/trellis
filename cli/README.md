# Trellis setup CLI

The `trellis` command sets Trellis up in a host project (`specs/0003` §2b). It is **setup
tooling, not a runtime** (`decision-0010`): run it once to detect your agent harness, pick an
expression profile, and compose Trellis onto the project — your agents then follow the resulting
instructions with no dependency on this binary.

Built in **Go** (`decision-0023`): a single static binary, no package manager.

## Build & test

```sh
cd cli
go build ./...
go test ./...
```

CI runs `go build` / `go vet` / `go test` on every PR touching `cli/` (`.github/workflows/cli-ci.yml`).

## Install (once releases are published)

```sh
curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
```

Downloads a single binary from GitHub Releases — no npm/registry. Until releases exist, build
from source (above).

## Status

Scaffolding only. Command surface: `version`, `help`, and a `setup` **stub** that fails loudly
until the interactive flow lands (the next stacked PR): detect harness → pick profile
(A / B / seed / Custom) → install mode (M1 alongside / M2 rewrite-on-a-branch).

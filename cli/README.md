# Trellis payload generator

The `trellis` command is **release tooling, not an end-user installer**: `trellis payload`
renders the complete pre-built M1 bundle — every posture variant, the managed blocks, the
expression seed skeletons, a content-derived version stamp, and a `shasum`-compatible checksum
manifest — into the vendored payload home, `plugins/trellis/reference/` (`kodhama-0007` rule 1:
render once, at release; `#117`).

**Generator-only, by decision.** `kodhama-0007` rule 5 retired the end-user binary channel
(Homebrew/curl) and left the Go code's fate open; `decision-0043` (#120) resolved it: the code
survives exactly as this generator plus its CI guards. The interactive `setup` TUI, `status`,
`remove`, `uninstall`, and the binary's M2 path are gone — their live homes are the Claude Code
plugin (`/trellis:setup`, `/trellis:remove`, the bundled staleness hook) and the manual copy path
in the repo README. Still no runtime, same as ever (`decision-0010`): consuming projects depend on
plain files, never on this binary.

## Build & test

```sh
cd cli
go build ./...
go test ./...
```

Dependency-free. CI runs `go build` / `go vet` / `go test` on every PR touching `cli/`, `docs/`,
`plugins/`, or the catalog (`.github/workflows/cli-ci.yml`).

## Regenerating the payload

```sh
cd cli
go generate ./...                                    # re-sync the catalog into assets/
go run . payload --out ../plugins/trellis/reference  # re-render the vendored payload
```

The tests in this package are the sync-guards that make drift impossible rather than merely
visible (`decision-0035`, mechanism per `kodhama-0007` rule 3):

- `TestVendoredPayloadIsCurrent` — the vendored payload is byte-identical to a fresh render.
- `TestRepoOverlayIsCurrent` — this repo's own committed `.trellis/` overlay matches the payload
  (self-application through the install boundary).
- `TestVendoredPayloadManifestVerifies` — `shasum -c` in Go: the manifest verifies the files as
  they sit on disk.
- `TestDocsClaimOnlyRealCommands` — the docs never advertise a command or skill that doesn't
  exist (`decision-0025`).
- `TestStalenessHook` — the plugin's SessionStart hook contract: a file-to-file
  `.trellis/version` ↔ `reference/version` comparison (`decision-0043`).

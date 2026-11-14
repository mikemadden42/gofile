# TODO

Findings from a project review (2026-11-14).

## Bugs and behavior (`main.go`)

- [x] **No feedback on failure.** A missing file or unknown format prints nothing and exits 0; errors only go to the log file. Print errors to stderr and exit non-zero.
- [x] **Log file written to the current directory.** `baseBinary + ".log"` (`main.go:20`) is relative, so every directory you run in gets a `gofile.log`, even with no arguments. Remove it, or put it behind a `-v` flag that writes to stderr.
- [x] **Noisy log.** Every run tries all four parsers and logs the expected failures, even when the file was identified.
- [x] **Parsed files never closed.** `macho.Open`, `macho.OpenFat`, `elf.Open`, `pe.Open` return `io.Closer`s; add `defer f.Close()`.
- [x] **Argument and error handling:**
  - `os.Executable()` error is discarded (`main.go:18`).
  - "Need more args" is also printed for too *many* arguments; no usage line.
  - Only one file accepted per run (`file(1)` accepts many).
  - `log.Fatal` skips the deferred `Close`, making it pointless.
- [ ] **Uneven output:**
  - Mach-O prints `CpuArm64` (Go constant name) instead of `arm64`.
  - PE prints raw hex (`0x8664`) instead of a name like `amd64`.
  - No 32/64-bit or file type (executable, shared library, object).
  - Parsers keep running after one succeeds; read magic bytes once and dispatch.

## Tooling and config

- [x] **`.golangci.yml` invalid for golangci-lint v2** (`golangci-lint config verify` fails):
  - `version: 2` must be the string `"2"`.
  - `disable-all` → `linters.default: none`.
  - `linters-settings` → `linters.settings`.
  - `issues.exclude-dirs` → `linters.exclusions.paths`.
  - goimports `local-prefixes` → `formatters.settings.goimports.local-prefixes` (list).
- [x] **`.goreleaser.yaml` fails `goreleaser check`:**
  - Add `version: 2` at the top.
  - `snapshot.name_template` → `version_template`.
  - Remove the unused `go generate ./...` hook.
- [x] **Go versions disagree:** `go.mod` says 1.18, CI tests 1.23–1.25, last commit says 1.21–1.23. Pick a minimum, set `go.mod` to it, test from there to the current release.
- [x] **CI gaps:**
  - No tests, so `go test` does nothing.
  - Lint doesn't run in CI.
  - `setup-go` runs before `checkout`.
  - Bump `actions/checkout` and `actions/setup-go` to current majors (now v7).

## Smaller things

- [ ] Expand the README with usage and example output.
- [x] Add a table test over small fixture binaries (or `os.Executable()` itself).

## Suggested order

1. User-facing behavior: errors to stderr, non-zero exit, no log file in the current directory.
2. Fix both config files so lint and release use your settings.
3. Align Go versions; add lint and a basic test to CI.

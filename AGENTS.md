# AGENTS.md

This file provides context for AI coding agents working in this repository.

## Version Control

This project uses **[jj (Jujutsu)](https://martinvonz.github.io/jj/)**, not git.

- Track new files: `jj file track <path>`
- Nix flakes only see VCS-tracked files — new files must be tracked before `nix build` can use them.

## Build and Test

Tool management and task running use [mise](https://mise.jdx.dev). Use `mise exec --` to run mise-managed tools from non-activated shells.

### Project-wide commands

```bash
mise run test        # Run tests with gotestsum (coverage → bin/coverage.out)
mise run lint        # Run golangci-lint with auto-fix
mise run generate    # Run go generate (stringer)
mise run mock        # Generate mocks with mockery
mise run build       # Build to dist/
mise run dev         # Full pipeline: generate → mock → lint → test → snapshot
```

### File-scoped commands

```bash
# Lint a single package
golangci-lint run ./pkg/exercise/...

# Type-check a single package
go vet ./pkg/exercise/...

# Run a single test
go test -run TestFunctionName ./path/to/package

# Run all tests in one package
go test -race ./pkg/exercise/...

# Format (applied automatically by golangci-lint, but can be run directly)
goimports -w path/to/file.go
```

## Code Style

- Linter config: `.golangci.yml` (golangci-lint v2, very strict)
- Max line length: **120** characters (enforced by `golines`)
- Imports grouped: stdlib, third-party, then `github.com/asphaltbuffet/elf` (enforced by `goimports`)
- Naked returns forbidden; named returns forbidden
- Magic numbers flagged — extract to named constants
- No `init()` functions; no global loggers (use `log/slog` with context)
- Comments end with a period
- Test files are excluded from `bodyclose`, `dupl`, `errcheck`, `funlen`, `goconst`, `gosec`, `noctx`, `wrapcheck`

## Architecture

**elf** is a CLI tool for managing programming challenge exercises (Advent of Code, Exercism). It downloads challenges, runs solutions in multiple languages, and benchmarks implementations.

### Package map

| Package | Purpose |
|---------|---------|
| `cmd/` | Cobra CLI commands (solve, test, benchmark, download, analyze) |
| `pkg/exercise/` | Exercise management — downloading, solving, testing, benchmarking |
| `pkg/config/` | Configuration via Viper — config files, env vars (`ELF_*`), defaults |
| `pkg/runners/` | Language runner abstraction — Go (`go/`) and Python (`py/`) |
| `pkg/tasks/` | Task types (Solve, Test, Benchmark, Visualize) and result handling |
| `pkg/analyze/` | Benchmark analysis and graph generation |
| `internal/utilities/` | Internal string helpers |

### `pkg/exercise/` file organization

The largest package splits files by responsibility:

| File | Contents |
|------|----------|
| `advent.go` | `Exercise` struct, constructor, options, `loadInfo` |
| `solver.go` | `Solve`, `runMainTasks`, `makeMainTasks` |
| `tester.go` | `Test`, `runTests`, `makeTestTasks` |
| `result.go` | `handleTaskResult` (shared by solver, tester, benchmarker), `testTask` type |
| `benchmarker.go` | `Benchmarker` struct, `NewBenchmarker`, `Benchmark`, `runBenchmark`, `NormalizationFactor` |
| `benchmarker_data.go` | `BenchmarkData`, `ImplementationData`, `PartData` types, `calculateMetrics` |
| `downloader.go` | `Downloader` struct, constructor, options, `validate`, `Download`, URL parsing, path resolution |
| `downloader_http.go` | `getPage`, `getCachedPage`, `downloadPage`, `getCachedInput`, `downloadInput`, `getInput` |
| `downloader_files.go` | `go:embed` templates, `addMissingFiles`, `writeInputFile`, `writeInfoFile`, `addTemplatedFile` |
| `helpers_test.go` | Shared test fixtures: `setupTestCase`, `setupSubTest`, `FileExists`, `goldenValue`, package-level vars (`testFs`, `mockDlr`) |

Test files mirror source files (e.g., `downloader_http_test.go` tests HTTP/caching functions).

### Exercise directory layout

```
exercises/<year>/<day>-<title>/
├── info.json      # Metadata (year, day, title, URL)
├── input.txt      # Puzzle input
├── README.md      # Problem description
├── go/exercise.go
└── py/exercise.py
```

### Runner system

`runners.Runner` interface abstracts language execution (Start/Stop/Run/Cleanup). Solutions communicate results via a standardized protocol in `pkg/runners/comm.go`.

### Configuration

- Config file: `elf.toml` in cwd or `~/.config/elf/`
- Environment: `ELF_ADVENT_TOKEN`, `ELF_LANGUAGE`
- Defaults: `pkg/config/defaults.go`

## Testing

### Cobra command testing pattern

All `cmd/` packages use **factory variables** to make `RunE` handlers testable. Package-level `var` functions wrap constructors so tests can swap them with mock-returning functions without changing cobra signatures.

Tests must be **internal** (`package download`, not `package download_test`) to access unexported `runXxxCmd` functions and factory variables.

Each test file uses a `resetState` helper to restore package-level state:

```go
func resetState(t *testing.T, origMakeConfig func(string) (config.Config, error), ...) {
    t.Helper()
    t.Cleanup(func() {
        downloadCmd = nil    // reset singleton
        language = ""        // reset flag vars
        makeConfig = origMakeConfig  // restore factory
    })
}
```

Standard test cases for each `cmd/` handler:
1. Config creation error
2. Domain object creation error (e.g., `NewDownloader` fails)
3. Operation error (e.g., `Download()` fails)
4. Happy path (verify output and return value)
5. Flag propagation (verify each flag reaches the factory/mock)

Mockery-generated mocks exist in `mocks/` for `Challenge`, `ChallengeTester`, `Benchmarker`, `Downloader`, `Analyzer`, `ConfigurationReader`, `DownloadConfiguration`, `ExerciseConfiguration`, and `Runner`.

### Gotchas

- **pflag resets variables on flag creation**: `StringVarP(&variable, ..., "", ...)` immediately sets the variable to the default. In tests, set flag-bound variables **after** calling `GetXxxCmd()`, not before.
- **testifylint require-error**: Use `require` variants (not `assert`) for the **first** error assertion in a chain — applies to `ErrorContains`, `Error`, and `NoError` alike. The linter enforces this to prevent nil-pointer panics in subsequent assertions.
- **Singleton commands**: `GetXxxCmd()` caches in a package-level `var`. Reset it (`downloadCmd = nil`) in `t.Cleanup` between tests to avoid stale flag state.

## Code Generation

- **stringer**: `//go:generate stringer` — generates `String()` for enums. Run `mise run generate`.
- **mockery**: configured in `.mockery.yaml`. Run `mise run mock`.

Generated files should be committed. Re-run generation after changing interfaces or enum types.

## Nix

- Flake uses `gomod2nix` with `buildGoApplication` (not `buildGoModule`)
- Dependency lockfile: `gomod2nix.toml` — regenerate with `mise run nix-hash` or `gomod2nix generate`
- Source filtering via `lib.fileset` — includes `.go`, `go.mod`, `go.sum`, `gomod2nix.toml`, and `go:embed` templates
- `mod-tidy` auto-runs `nix-hash` as post-dependency
- Home-manager module at `nix/home-manager.nix` — option defaults must stay in sync with `pkg/config/defaults.go`

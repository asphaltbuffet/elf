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
golangci-lint run ./pkg/advent/...

# Type-check a single package
go vet ./pkg/advent/...

# Run a single test
go test -run TestFunctionName ./path/to/package

# Run all tests in one package
go test -race ./pkg/runners/...

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

**elf** is a CLI tool for managing programming challenge exercises (Advent of Code, Project Euler, Exercism). It downloads challenges, runs solutions in multiple languages, and benchmarks implementations.

### Package map

| Package | Purpose |
|---------|---------|
| `cmd/` | Cobra CLI commands (solve, test, benchmark, download, analyze) |
| `pkg/advent/` | Advent of Code — downloading, solving, testing, benchmarking |
| `pkg/euler/` | Project Euler support (WIP) |
| `pkg/krampus/` | Configuration via Viper — config files, env vars (`ELF_*`), defaults |
| `pkg/runners/` | Language runner abstraction — Go (`go/`) and Python (`py/`) |
| `pkg/tasks/` | Task types (Solve, Test, Benchmark, Visualize) and result handling |
| `pkg/analysis/` | Benchmark analysis and graph generation |

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
- Defaults: `pkg/krampus/defaults.go`

## Testing

### Cobra command testing pattern

The `cmd/` packages use **factory variables** to make `RunE` handlers testable. Package-level `var` functions wrap constructors so tests can swap them with mock-returning functions without changing cobra signatures.

```go
// Production: factory variable wraps real constructor.
var makeChallenge = func(cfg krampus.ExerciseConfiguration, lang, dir, inputFile string) (Challenge, error) {
    return advent.New(cfg, ...)
}

// Test: swap factory to return mock.
makeChallenge = func(...) (Challenge, error) { return mockCh, nil }
```

Mockery-generated mocks exist in `mocks/` for `Challenge`, `ChallengeTester`, `Benchmarker`, and `Downloader`.

### Gotchas

- **pflag resets variables on flag creation**: `StringVarP(&variable, ..., "", ...)` immediately sets the variable to the default. In tests, set flag-bound variables **after** calling `GetXxxCmd()`, not before.
- **testifylint require-error**: Use `require.NoError` (not `assert.NoError`) when subsequent assertions depend on no error. The linter enforces this.
- **Singleton commands**: `GetXxxCmd()` caches in a package-level `var`. Reset it (`solveCmd = nil`) in `t.Cleanup` between tests to avoid stale flag state.

## Code Generation

- **stringer**: `//go:generate stringer` — generates `String()` for enums. Run `mise run generate`.
- **mockery**: configured in `.mockery.yaml`. Run `mise run mock`.

Generated files should be committed. Re-run generation after changing interfaces or enum types.

## Nix

- Flake uses `gomod2nix` with `buildGoApplication` (not `buildGoModule`)
- Dependency lockfile: `gomod2nix.toml` — regenerate with `mise run nix-hash` or `gomod2nix generate`
- Source filtering via `lib.fileset` — includes `.go`, `go.mod`, `go.sum`, `gomod2nix.toml`, and `go:embed` templates
- `mod-tidy` auto-runs `nix-hash` as post-dependency
- Home-manager module at `nix/home-manager.nix` — option defaults must stay in sync with `pkg/krampus/defaults.go`

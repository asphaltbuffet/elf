# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## TODO

- [ ] Consider using `version-file` in golangci-lint-action to read version from `.tool-versions` instead of hardcoding in workflow
- [ ] Bump pinned actionlint version in `actionlint.yml` (currently `1.7.10`, check [releases](https://github.com/rhysd/actionlint/releases))
- [ ] Migrate internal mockery-generated mocks toward hand-rolled fakes (prefer fakes for interfaces defined within this codebase; reserve mockery for external dependencies)

## Issue Tracker

Issues live in GitHub Issues at `github.com/asphaltbuffet/elf`. Use the `gh` CLI for all operations.

## Important

- This project uses **jj** (Jujutsu) for version control, not git. Use `jj file track` instead of `git add`.
- Nix flakes only see VCS-tracked files. New files must be tracked with `jj file track` before `nix build` can use them.
- Use `mise exec --` to run mise-managed tools (e.g. `mise exec -- gomod2nix generate`) from non-activated shells.
- `gomod2nix` is a mise `go:` tool (built from source — it ships no prebuilt binaries, so no aqua/ubi backend). On NixOS the built binary hardcodes a `/nix/store` glibc interpreter that breaks after `nix-collect-garbage` (`cannot execute: required file not found`); fix by reinstalling: `mise install "go:github.com/nix-community/gomod2nix"`. The flake devShell also provides gomod2nix as a GC-safe fallback.

## Build and Development Commands

This project uses [mise](https://mise.jdx.dev) for tool management and task running.

```bash
mise run test        # Run tests with gotestsum (generates coverage in bin/coverage.out)
mise run lint        # Run golangci-lint (read-only, used in CI)
mise run lint-fix    # Run golangci-lint with auto-fix (use locally to apply fixes)
mise run generate    # Run go generate (stringer)
mise run mock        # Generate mocks with mockery
mise run build       # Build to dist/
mise run snapshot    # Build release snapshot with goreleaser
mise run update-deps # Update direct dependencies to latest (then mod-tidy → nix-hash)
mise run dev         # Full dev pipeline: generate, mock, lint, test, snapshot
```

Always use `mise run dev` for full verification (generate, mock, lint, test, snapshot), not just `go test`/`go build`.

Run a single test:
```bash
go test -run TestFunctionName ./path/to/package
```

## Architecture

**elf** is a CLI tool that helps manage programming challenge exercises (Advent of Code, Exercism). It downloads challenges, runs solutions in multiple languages, and benchmarks implementations.

### Core Packages

- **cmd/**: Cobra CLI commands (solve, test, benchmark, download, analyze)
- **pkg/exercise/**: Exercise management - downloading, solving, testing, benchmarking
- **pkg/config/**: Configuration management via Viper - handles config files, environment variables (ELF_* prefix), and defaults
- **pkg/runners/**: Language runner abstraction - executes solutions in Go (`go/`) or Python (`py/`) subdirectories
- **pkg/tasks/**: Task types (Solve, Test, Benchmark, Visualize) and result handling
- **pkg/analyze/**: Benchmark analysis and graph generation
- **internal/tui/**: Bubbletea TUI mode (launched when `elf` runs with no subcommand)
- **internal/utilities/**: Internal string helpers

### `pkg/exercise/` File Organization

The largest package splits files by responsibility:

| File | Contents |
|------|----------|
| `advent.go` | `Exercise` struct, constructor, options, `loadInfo` |
| `solver.go` | `Solve`, `runMainTasks`, `makeMainTasks` |
| `tester.go` | `Test`, `runTests`, `makeTestTasks` |
| `result.go` | `buildResult` (pure data), `renderResult` (CLI styling), `handleTaskResult` wrapper, `testTask` type |
| `benchmarker.go` | `Benchmarker` struct, `NewBenchmarker`, `Benchmark`, `runBenchmark`, `NormalizationFactor` |
| `benchmarker_data.go` | `BenchmarkData`, `ImplementationData`, `PartData` types, `calculateMetrics` |
| `downloader.go` | `Downloader` struct, constructor, options, `validate`, `Download`, URL parsing, path resolution |
| `downloader_http.go` | `getPage`, `getCachedPage`, `downloadPage`, `getCachedInput`, `downloadInput`, `getInput` |
| `downloader_files.go` | `go:embed` templates, `addMissingFiles`, `writeInputFile`, `writeInfoFile`, `addTemplatedFile` |
| `helpers_test.go` | Shared test fixtures: `setupTestCase`, `setupSubTest`, `FileExists`, `goldenValue`, package-level vars (`testFs`, `mockDlr`) |

Test files mirror source files (e.g., `downloader_http_test.go` tests HTTP/caching functions).

### Exercise Structure

Exercises follow this directory layout:
```
exercises/<year>/<day>-<title>/
├── info.json      # Metadata (year, day, title, URL)
├── input.txt      # Puzzle input
├── README.md      # Problem description
├── go/            # Go implementation
│   └── exercise.go
└── py/            # Python implementation
    └── exercise.py
```

### Runner System

The `runners.Runner` interface abstracts language execution. Each runner (Go, Python) implements Start/Stop/Run/Cleanup methods. Solutions communicate results back via a standardized protocol in `pkg/runners/comm.go`.

### TUI Architecture

The TUI uses [bubbletea](https://github.com/charmbracelet/bubbletea) with a navigation stack pattern:

| Package | Purpose |
|---------|---------|
| `internal/tui/` | Root `App` model with nav stack (`[]tea.Model`), `Run()` entry point |
| `internal/tui/dashboard/` | Year list with progress bars, config summary |
| `internal/tui/yearview/` | Exercise table for a single year, action keybindings (s/t/b/a) |
| `internal/tui/exerciseview/` | Runs solve/test/benchmark with result streaming via channel |
| `internal/tui/components/` | Reusable: `ResultList` (viewport), `Progress` (spinner), `OpenFile` (xdg-open) |
| `internal/tui/discover/` | Filesystem scanner: reads `info.json` files, groups exercises by year |
| `internal/tui/nav/` | Shared message types (`PushScreenMsg`, `PopScreenMsg`) to prevent import cycles |

**Key patterns:**
- **Result streaming**: `exercise.WithResultCallback` sends `tasks.Result` through a channel. `exerciseview` reads it via `waitForResult` tea.Cmd, dispatching `resultMsg` for each item.
- **Suppressed CLI output**: TUI passes `exercise.WithWriter(io.Discard)` to suppress the CLI's direct stdout rendering. Results flow only through the callback.
- **ANSI-aware columns**: Use `lipgloss.Style.Width()` (not `fmt.Sprintf` `%-Ns`) for fixed-width columns — `fmt` counts ANSI escape bytes, breaking alignment for styled text.

**Gotchas:**
- `fmt.Sprintf("%-Ns")` counts ANSI escape bytes, not visible chars — always use `lipgloss.Style.Width()` for fixed columns
- Bubbletea `Init()` uses value receivers — cannot mutate model state; initialize channels/timers in constructors
- `lipgloss.Style.Inherit(other)` copies color/bold but preserves layout (width, align, padding) — use for styled table cells

### Configuration

Configuration uses Viper with:
- Config file: `elf.toml` in current dir or `~/.config/elf/`
- Environment: `ELF_ADVENT_TOKEN`, `ELF_LANGUAGE`
- Cache: `~/.cache/elf/` (or platform equivalent)

### Testing Cobra Commands

All `cmd/` packages (`solve`, `test`, `benchmark`, `download`, `analyze`) use a **factory variable pattern** for testability. Package-level `var` functions wrap constructors so tests can swap them with mock-returning functions. This avoids changing cobra `RunE` signatures while enabling mock injection.

Each `cmd/` test file follows this structure:
1. **Internal test** (`package solve`, not `package solve_test`) for access to unexported `runXxxCmd` and factory vars
2. **`resetState` helper** that restores all package-level vars and factory functions via `t.Cleanup`
3. **Factory variable swaps** to inject mocks or error-returning stubs

Example factory variables from `cmd/download/`:
```go
var makeConfig = func(cf string) (config.Config, error) {
    return config.NewConfig(config.WithFile(cf))
}
var makeDownloader = func(cfg config.DownloadConfiguration, url, lang string, forced *exercise.Overwrites) (Downloader, error) {
    return exercise.NewDownloader(cfg, ...)
}
```

**pflag gotcha**: `StringVarP(&variable, ...)` sets the variable to the default value immediately. In tests, always set package-level flag variables **after** `GetXxxCmd()`, not before.

Mockery-generated mocks exist in `mocks/` for `Challenge`, `ChallengeTester`, `Benchmarker`, `Downloader`, `Analyzer`, `ConfigurationReader`, `DownloadConfiguration`, `ExerciseConfiguration`, and `Runner` — use them with the factory variable pattern.

### Code Generation

- **stringer**: Generates String() methods for enums (`//go:generate stringer`)
- **mockery**: Generates test mocks (configured in `.mockery.yaml`)

### Linter Gotchas

- `noctx`: Use `exec.CommandContext(ctx, ...)` not `exec.Command(...)` — even with `context.Background()`
- Adding `func` fields to structs breaks `==` comparison — use field-level zero checks instead
- `govet shadow`: inner `:=` shadowing outer `err` is caught — rename inner variable (e.g., `graphErr`)
- `mnd`: Extract magic numbers to named constants, even for visual widths/padding

### Shell Tooling — `sd` Newline Limitation

`sd` does **not** match across newlines. A pattern like `'foo,\n'` silently matches nothing (exits 0, no change). Always write patterns that match within a single line:

```bash
# Wrong — \n never matches, silent no-op
sd 'SomeField:\s+pkg\.SomeType\w+,\n' '' file.go

# Right — omit the \n; accept that a blank line may remain
sd 'SomeField:\s+pkg\.SomeType\w+,' '' file.go
```

If you need to remove an entire line (including the newline), use the `Edit` tool instead of `sd`.

### Nix Flake

- Uses `gomod2nix` with `buildGoApplication` (not `buildGoModule`)
- `gomod2nix.toml` is the dependency lockfile — regenerate with `mise run nix-hash` or `gomod2nix generate`
- Source filtering via `lib.fileset` — only `.go` files, `go.mod`, `go.sum`, `gomod2nix.toml`, and `go:embed` templates (`pkg/exercise/templates/`, `pkg/runners/interface/`) are included
- `mod-tidy` automatically runs `nix-hash` as a post-dependency

Flake outputs:
- `packages.default` — the elf binary (per-system)
- `devShells.default` — development shell with Go, mise, gomod2nix (per-system)
- `overlays.default` — adds `pkgs.elf` to nixpkgs
- `homeManagerModules.default` — home-manager module (`nix/home-manager.nix`) providing `programs.elf.*` options

#### Home-Manager Module

- Located at `nix/home-manager.nix`, options live under `programs.elf`
- Option defaults mirror `pkg/config/defaults.go` — keep them in sync when defaults change
- `config-dir` and `cache-dir` default to `null` (omitted from TOML) so elf's runtime XDG logic applies
- Config file generated via `pkgs.formats.toml {}` and placed at `xdg.configFile."elf/elf.toml"`
- Environment variables (`ELF_ADVENT_TOKEN`, `ELF_LANGUAGE`) mapped via `home.sessionVariables`

## Agent skills

### Issue tracker

Issues live in GitHub Issues at `github.com/asphaltbuffet/elf`. See `docs/agents/issue-tracker.md`.

### Triage labels

Default label vocabulary (needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout — one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.

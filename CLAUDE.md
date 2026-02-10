# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Important

- This project uses **jj** (Jujutsu) for version control, not git. Use `jj file track` instead of `git add`.
- Nix flakes only see VCS-tracked files. New files must be tracked with `jj file track` before `nix build` can use them.
- Use `mise exec --` to run mise-managed tools (e.g. `mise exec -- gomod2nix generate`) from non-activated shells.

## Build and Development Commands

This project uses [mise](https://mise.jdx.dev) for tool management and task running.

```bash
mise run test        # Run tests with gotestsum (generates coverage in bin/coverage.out)
mise run lint        # Run golangci-lint with auto-fix
mise run generate    # Run go generate (stringer)
mise run mock        # Generate mocks with mockery
mise run build       # Build to dist/
mise run snapshot    # Build release snapshot with goreleaser
mise run update-deps # Update direct dependencies to latest (then mod-tidy → nix-hash)
mise run dev         # Full dev pipeline: generate, mock, lint, test, snapshot
mise run ci          # CI pipeline: generate, mock, mod-tidy, test, cover, build, diff
```

Run a single test:
```bash
go test -run TestFunctionName ./path/to/package
```

## Architecture

**elf** is a CLI tool that helps manage programming challenge exercises (Advent of Code, Project Euler, Exercism). It downloads challenges, runs solutions in multiple languages, and benchmarks implementations.

### Core Packages

- **cmd/**: Cobra CLI commands (solve, test, benchmark, download, analyze)
- **pkg/advent/**: Advent of Code implementation - downloading, solving, testing, benchmarking
- **pkg/euler/**: Project Euler support (WIP)
- **pkg/krampus/**: Configuration management via Viper - handles config files, environment variables (ELF_* prefix), and defaults
- **pkg/runners/**: Language runner abstraction - executes solutions in Go (`go/`) or Python (`py/`) subdirectories
- **pkg/tasks/**: Task types (Solve, Test, Benchmark, Visualize) and result handling
- **pkg/analysis/**: Benchmark analysis and graph generation

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

### Configuration

Configuration uses Viper with:
- Config file: `elf.toml` in current dir or `~/.config/elf/`
- Environment: `ELF_ADVENT_TOKEN`, `ELF_LANGUAGE`
- Cache: `~/.cache/elf/` (or platform equivalent)

### Code Generation

- **stringer**: Generates String() methods for enums (`//go:generate stringer`)
- **mockery**: Generates test mocks (configured in `.mockery.yaml`)

### Nix Flake

- Uses `gomod2nix` with `buildGoApplication` (not `buildGoModule`)
- `gomod2nix.toml` is the dependency lockfile — regenerate with `mise run nix-hash` or `gomod2nix generate`
- Source filtering via `lib.fileset` — only `.go` files, `go.mod`, `go.sum`, `gomod2nix.toml`, and `go:embed` templates are included
- `mod-tidy` automatically runs `nix-hash` as a post-dependency

Flake outputs:
- `packages.default` — the elf binary (per-system)
- `devShells.default` — development shell with Go, mise, gomod2nix (per-system)
- `overlays.default` — adds `pkgs.elf` to nixpkgs
- `homeManagerModules.default` — home-manager module (`nix/home-manager.nix`) providing `programs.elf.*` options

#### Home-Manager Module

- Located at `nix/home-manager.nix`, options live under `programs.elf`
- Option defaults mirror `pkg/krampus/defaults.go` — keep them in sync when defaults change
- `config-dir` and `cache-dir` default to `null` (omitted from TOML) so elf's runtime XDG logic applies
- Config file generated via `pkgs.formats.toml {}` and placed at `xdg.configFile."elf/elf.toml"`
- Environment variables (`ELF_ADVENT_TOKEN`, `ELF_LANGUAGE`) mapped via `home.sessionVariables`

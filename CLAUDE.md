# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

This project uses [mise](https://mise.jdx.dev) for tool management and task running.

```bash
mise run test        # Run tests with gotestsum (generates coverage in bin/coverage.out)
mise run lint        # Run golangci-lint with auto-fix
mise run generate    # Run go generate (stringer)
mise run mock        # Generate mocks with mockery
mise run build       # Build to dist/
mise run snapshot    # Build release snapshot with goreleaser
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

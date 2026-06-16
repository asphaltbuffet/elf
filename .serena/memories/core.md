# elf — Core

CLI tool for managing programming challenge exercises (Advent of Code, Exercism). Downloads challenges, runs solutions in multiple languages, benchmarks implementations.

## Source map

- `cmd/` — Cobra CLI commands (analyze, benchmark, config, download, man, solve, test)
- `pkg/analyze/` — benchmark analysis and graph generation
- `pkg/config/` — Viper-based configuration (env prefix `ELF_`, file `elf.toml`)
- `pkg/exercise/` — exercise management (downloading, solving, testing, benchmarking)
- `pkg/runners/` — language runner abstraction (Go, Python)
- `pkg/tasks/` — task types (Solve, Test, Benchmark, Visualize) and Result types
- `internal/tui/` — Bubbletea TUI (launched with no subcommand); see `mem:tui`
- `internal/utilities/` — string helpers

## Project-wide invariants

- VCS: **jj** (Jujutsu). Use `jj file track` instead of `git add`. New files must be tracked before `nix build` sees them.
- Build: Nix flake with `gomod2nix` (`buildGoApplication`). `gomod2nix.toml` is the dep lockfile.
- Tool management: **mise**. Use `mise exec --` from non-activated shells.
- Issue tracker: GitHub Issues at `github.com/asphaltbuffet/elf`. Use `gh` CLI.

See `mem:tech_stack`, `mem:suggested_commands`, `mem:conventions`, `mem:task_completion`, `mem:tui`.

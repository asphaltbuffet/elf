# elf — Core

CLI tool for managing programming challenge exercises (Advent of Code, Exercism).

## Package map
- `cmd/` — Cobra commands (solve, test, benchmark, download, analyze)
- `pkg/exercise/` — exercise lifecycle (download, solve, test, benchmark, scaffold)
- `pkg/tasks/` — task types (Solve/Test/Benchmark/Visualize), result, status
- `pkg/runners/` — language runner abstraction (Go, Python)
- `pkg/config/` — Viper config (file + ELF_* env vars)
- `pkg/analyze/` — benchmark analysis + graph gen
- `internal/render/` — event-stream renderers (Plain + Live/bubbletea)
- `internal/utilities/` — internal string helpers

## Key invariants
- VCS is jujutsu (jj), not git. Do not use `git` commands.
- Build system: Nix flake with gomod2nix (`buildGoApplication`, not `buildGoModule`).
- Tool runner: mise (see `mem:suggested_commands`).
- Three stringer-generated files are gitignored / untracked; nix build fails without them (see `mem:nix`).

References: `mem:tech_stack`, `mem:nix`, `mem:conventions`, `mem:suggested_commands`, `mem:task_completion`

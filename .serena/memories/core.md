# Core

**elf** is a Go CLI tool (cobra) for managing programming challenge exercises (Advent of Code). It downloads challenges, runs solutions in multiple languages, and benchmarks implementations.

## Top-level commands

| Command | Short | Purpose |
|---|---|---|
| `solve` | `s` | Run solution + optional tests |
| `test` | `t` | Run test cases only |
| `benchmark` | `b` | Benchmark all implementations |
| `visualize` | `vis`, `v` | Run visualization output to disk |
| `download` | `d` | Download challenge from URL |
| `analyze` | `a`, `analyse` | Graph benchmark run-time data |
| `config` | — | Manage config (init, check, update-token) |
| `runners` | — | Manage runner plugins (list, install) |
| `version` | — | Print version |
| `man` | — | Generate man pages |

## Source layout

- `cmd/` — Cobra commands (one subdir per command)
- `pkg/exercise/` — Download, solve, test, benchmark logic
- `pkg/runners/` — Language runner abstraction (descriptor-driven plugin system)
- `pkg/config/` — Viper-based config (file + env)
- `pkg/tasks/` — Task types and result handling
- `pkg/analyze/` — Benchmark graph generation
- `internal/render/` — Event-stream rendering (Plain + Live bubbletea renderers)
- `internal/utilities/` — Internal string helpers

See `mem:conventions` for code patterns. See `mem:tech_stack` for build tools.

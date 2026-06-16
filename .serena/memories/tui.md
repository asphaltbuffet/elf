# TUI Architecture

Bubbletea-based TUI launched when `elf` runs with no subcommand.

## Package layout

| Package | Purpose |
|---------|---------|
| `internal/tui/` | Root `App` model, nav stack (`[]tea.Model`), `Run()` entry |
| `internal/tui/dashboard/` | Year list with progress bars, config summary |
| `internal/tui/yearview/` | Exercise table for a single year, action keybindings (s/t/b/a) |
| `internal/tui/exerciseview/` | Runs solve/test/benchmark with result streaming via channel |
| `internal/tui/components/` | Reusable: `ResultList` (viewport), `Progress` (spinner), `OpenFile` (xdg-open) |
| `internal/tui/discover/` | Filesystem scanner: reads `info.json`, groups exercises by year |
| `internal/tui/nav/` | Shared message types (`PushScreenMsg`, `PopScreenMsg`) to prevent import cycles |

## Key patterns

- **Result streaming**: `exercise.WithResultCallback` → channel → `waitForResult` tea.Cmd → `resultMsg` per item.
- **Suppressed CLI output**: TUI passes `exercise.WithWriter(io.Discard)`.
- **huh forms** (download view): heap-allocated pointers for form values; override keymap for `esc`.

## Gotchas

- `fmt.Sprintf("%-Ns")` counts ANSI bytes → use `lipgloss.Style.Width()`.
- Bubbletea `Init()` uses value receivers — cannot mutate state; initialize channels in constructors.
- `lipgloss.Style.Inherit(other)` copies color/bold but preserves layout.
- huh `Form.View()` renders empty until `WindowSizeMsg` — send one first in tests.
- nested `if ok` type assertions need unique names (`wsOk`, `fOk`) for govet shadow.

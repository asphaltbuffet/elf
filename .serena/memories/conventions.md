# Conventions

## cmd/ pattern (factory variable pattern)

Each `cmd/` package uses package-level `var` functions wrapping constructors so tests can swap them with mock-returning stubs. Tests are internal (`package solve`, not `package solve_test`) for access to unexported symbols. Always use a `resetState` helper via `t.Cleanup`.

**pflag gotcha**: `StringVarP(&variable, ...)` sets the variable to the default immediately. In tests, always set flag variables **after** `GetXxxCmd()`.

## exercise/ file split

`advent.go` (struct/ctor), `solver.go`, `tester.go`, `result.go`, `benchmarker.go`, `benchmarker_data.go`, `downloader.go`, `downloader_http.go`, `downloader_files.go`, `helpers_test.go`.

## TUI patterns

- Result streaming via `exercise.WithResultCallback` → channel → `waitForResult` tea.Cmd → `resultMsg`.
- `exercise.WithWriter(io.Discard)` suppresses CLI stdout in TUI mode.
- Use `lipgloss.Style.Width()` for fixed-width columns — NOT `fmt.Sprintf("%-Ns")` (counts ANSI bytes).
- huh form values must be heap-allocated pointers — bubbletea copies models on every Update.
- huh default quit key is only `ctrl+c`; override keymap to add `esc`.
- `lipgloss.Style.Inherit(other)` copies color/bold but preserves layout (width, align, padding).

## Linter gotchas

- `noctx`: always use `exec.CommandContext(ctx, ...)`.
- Adding `func` fields to structs breaks `==` — use field-level zero checks.
- `govet shadow`: inner `:=` shadowing outer `err` caught — rename inner var.
- `mnd`: extract magic numbers to named constants.

## sd limitation

`sd` does NOT match across newlines. Use `Edit` tool for whole-line removal.

## Mocks

Prefer hand-rolled fakes for interfaces defined within this codebase; reserve mockery for external dependencies.

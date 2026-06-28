# Conventions

## cmd/ pattern
- Factory variable pattern for testability: package-level `var make* = func(...)` wraps constructors
- Internal tests (`package solve`, not `package solve_test`) for access to unexported vars
- `resetState` helper + `t.Cleanup` to restore factory vars between tests
- `pflag` gotcha: `StringVarP` sets default immediately — set flag vars AFTER `GetXxxCmd()` in tests

## Error handling
- `noctx`: always `exec.CommandContext(ctx, ...)` even with `context.Background()`
- No `func` fields in structs compared with `==` — use field-level zero checks
- `govet shadow`: rename inner `:=` err to avoid shadowing (e.g. `graphErr`)
- `mnd`: magic numbers → named constants

## sd (in-place substitution)
- Does NOT match across newlines — patterns with `\n` silently no-op
- To remove a whole line including newline, use the Edit tool

## Runner system
- `RunnerDescriptor` struct drives language execution via config-declared `[[runner]]` blocks
- Built-in templates: Go, Python, Bash, Rust, Fortran 77
- `runners install` writes templates + prints config blocks to add to elf.toml

## Rendering
- Domain emits `tasks.Event` via `cb func(tasks.Event)` callback
- `render.Run` feeds events into renderer, returns `([]tasks.Result, error)`
- `render.New(w, h, plain)`: `plain==true` or non-TTY → Plain; otherwise → Live

# Conventions

## Code style
- Factory variable pattern in `cmd/` packages for testability (package-level `var` funcs swapped in tests)
- Internal tests (`package foo`, not `package foo_test`) for access to unexported vars
- `resetState` helper in each cmd test restores factory vars via `t.Cleanup`
- pflag `StringVarP` sets default immediately — always set flag vars AFTER `GetXxxCmd()` in tests

## Linter gotchas
- `noctx`: use `exec.CommandContext(ctx, ...)` not `exec.Command(...)`
- `mnd`: extract magic numbers to named constants
- `govet shadow`: rename inner `:=` vars that shadow outer `err`
- Adding `func` fields to structs breaks `==` — use field-level zero checks instead

## sd limitations
- `sd` does not match across newlines — patterns with `\n` silently no-op
- To remove entire lines (including newline), use Edit tool instead

## Mocks
- mockery-generated mocks in `mocks/` for major interfaces
- Prefer hand-rolled fakes for interfaces defined within this codebase

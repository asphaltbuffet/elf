# TUI uses narrow interfaces; cmd/ does not

The TUI is a distinct presentation layer and should not import the application struct directly.
Each TUI screen that needs to trigger a domain operation receives a narrow interface (e.g.
`Solver` with a single `Solve` method) satisfied by `*App`. This enforces the seam
architecturally, not just for testability.

`cmd/` packages do not use interfaces or factory variables. They construct a real `App` and
call methods on it directly. The domain behavior is tested in domain packages; `cmd/` tests
cover only flag parsing and argument resolution.

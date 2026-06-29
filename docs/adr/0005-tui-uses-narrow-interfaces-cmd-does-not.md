# TUI uses narrow interfaces; cmd/ does not

> **Status: Superseded by [ADR 0010](0010-presentation-as-event-stream.md).** The TUI
> (`internal/tui/`) was removed entirely — it was effectively unused and its nav-stack /
> narrow-interface machinery complicated every change to the run/render path. The
> CLI/TUI duality this ADR governed no longer exists; presentation is now a single
> event-stream-driven layer shared by all `cmd/` operations. The decision below is kept
> for historical context only.

The TUI is a distinct presentation layer and should not import the application struct directly.
Each TUI screen that needs to trigger a domain operation receives a narrow interface (e.g.
`Solver` with a single `Solve` method) satisfied by `*App`. This enforces the seam
architecturally, not just for testability.

`cmd/` packages do not use interfaces or factory variables. They construct a real `App` and
call methods on it directly. The domain behavior is tested in domain packages; `cmd/` tests
cover only flag parsing and argument resolution.

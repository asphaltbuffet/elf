# Exercise operations as standalone functions

Solve, Test, and Benchmark are standalone functions that accept an `Exercise` value and explicit
dependencies (Runner, writer, logger, result callback) rather than methods on a shared mutable
type. This keeps `Exercise` as pure data and makes the execution context visible at the call
site. A session/executor type was considered but rejected: callers (cmd/, TUI) already own the
dependency context and don't reuse sessions across calls, so the indirection added no value.

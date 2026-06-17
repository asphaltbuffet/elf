# App owns invariant infrastructure; per-call dependencies are parameters

Domain operations (Solve, Test, Benchmark) are methods on an `App` struct rather than
standalone functions or methods on domain types. `App` holds infrastructure that does not
change across the application lifetime: filesystem, logger, and resolved config values.
Per-call dependencies that vary between invocations — Exercise, Runner, writer, result
callback — are explicit method parameters. This keeps `App` free of the mutation problems
that arise when dynamic state (e.g. the current runner) is stored on a shared struct, while
avoiding the parameter-threading overhead of passing fs and logger to every function.

# Per-task timeout applied at call sites, not inside the runner

Each `Runner.Run` call (one Task: one Part, one input) is individually guarded by a
`context.WithTimeout`. When the timeout fires the subprocess is killed; the existing
`defer Close/Cleanup` teardown handles the rest.

The timeout is applied in `runMainTasks`, `runTests`, and `runBenchmark` — not inside
`descriptorRunner.Run` — because the CLI flag override (`--timeout`/`-t`) lives at the
`cmd/` layer and needs to flow down as a resolved `time.Duration` via a
`WithTaskTimeout` option function on `Exercise`/`Benchmarker`. Putting the wrapping
inside the runner would require injecting config into the runner, coupling it to the
config package and making the flag-override path awkward.

A per-session timeout (one budget for the entire Prepare→Open→Run*→Close sequence) was
considered and rejected: it would make benchmark runs with many iterations unpredictably
expire, and the real hazard being guarded is a single hung or infinite-looping solution,
not total wall-clock time.

The config key is `task.timeout` (a Go duration string, default `"2m"`). A value `<=0`
disables the timeout entirely. The flag is bound to the config key via Viper's
`BindPFlag` so the config value is the automatic default when the flag is absent.

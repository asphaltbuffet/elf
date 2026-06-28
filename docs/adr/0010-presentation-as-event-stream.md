# Presentation is a pure function of a task lifecycle event stream

`Solve`, `Test`, and `Benchmark` no longer format any output themselves. Instead they emit
a stream of `tasks.Event` values through a single callback, `cb func(tasks.Event)`. An
event has one of three kinds — **Planned** (this task will run), **Started** (this task is
now executing), **Finished** (this task is done; carries the authoritative
`*tasks.Result`). Planned and Started are emitted per task (not as a batch), so the stream
is uniform and a renderer can build its view incrementally without assuming the shape of a
run (e.g. test-then-solve, or benchmark's per-implementation layout).

Two renderers consume the same event stream:

- a **live renderer** (bubbles/v2 + lipgloss/v2) that owns stdout, shows all tasks up
  front as `<not started>`, animates a spinner and a wall-clock elapsed timer for the
  `<running>` task, and settles each row to PASS/FAIL/TIMEOUT with the runner-measured
  `Duration` on Finished; and
- a **plain renderer** that reproduces the previous synchronous output (one line per
  Finished result, header box and section labels on the appropriate first events) for
  non-interactive output.

The renderer is chosen by TTY auto-detection on stdout (overridable with `--plain`).
Diagnostic `DBG` output is routed to stderr so it never corrupts the live frame.

## Why

The previous design split presentation between a synchronous `io.Writer` path (used by
`cmd/`) and a callback path (used by the now-removed TUI), with the domain itself calling
`fmt.Fprintln`. A live spinner/timer cannot interleave with synchronous writes — whatever
owns the redraw region must own stdout — and showing `<not started>` rows requires knowing
the task list before any result exists. Both needs point to the same answer: make the
domain announce task lifecycle as events and make every renderer a pure function of that
stream. This also gives `solve`, `test`, and (later) `benchmark` one shared rendering
vocabulary, the stated goal of a common output theme.

## Trade-offs and what was rejected

- **Domain still returns `([]tasks.Result, error)`.** Events are display-only; the return
  value remains the single source of truth for exit code and error propagation. The two
  are complementary and must not disagree, so outcome logic lives only in the return path,
  never in a renderer.
- **`tasks.Result` is unchanged.** `Event` *wraps* a `Result` on Finished rather than
  overloading `Result` with synthetic "running"/"pending" statuses, which would have
  forced golden-file and benchmark-serialization consumers to special-case non-real
  states.
- **Live timer is not reconciled with `Duration`.** While running, the row shows wall-clock
  `time.Since(Started)` in a coarse format ("how long you've waited"); on Finished it shows
  the authoritative `Duration` in the existing precise format ("how long the solution
  took"). The nondeterministic wall-clock value never appears in settled output, keeping
  final frames reproducible.
- **An up-front "planned tasks" query (instead of events)** was rejected: without a real
  Started event the running timer would start at the previous task's finish, mismeasuring
  the gap (runner restart, I/O).

This supersedes [ADR 0005](0005-tui-uses-narrow-interfaces-cmd-does-not.md): the TUI was
removed, so the CLI/TUI interface duality it governed no longer applies.

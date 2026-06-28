# Benchmark renders as progress bars by aggregating the per-iteration event stream

Benchmark joins Solve and Test on the [ADR-0010](0010-presentation-as-event-stream.md) event stream.
The domain emits the same uniform per-iteration stream (Planned/Started/Finished per iteration per
Part); the renderer aggregates it into one **progress bar per (Runner, Part)** — a 100-iteration,
two-Runner benchmark shows four bars, not 400 lines. A bar advances one tick per Finished event for
its key. While running it shows live wall-clock elapsed; settled it shows the sum (Σ) of the
iteration durations it already received. The Plain renderer shows the same as one settled line per
(Runner, Part).

To attribute an iteration event to a Runner, `tasks.Event` gains a `Language` field (the
human-readable Runner name, matching `Meta.Language`); benchmark sets it, Solve/Test leave it empty.

## Why

ADR-0010 made presentation a pure function of the lifecycle stream and named benchmark as a future
consumer. Aggregating the *unchanged* stream into bars cashes that in without giving the domain any
opinion about how it is drawn. The only new datum crossing the domain→renderer boundary is
`Event.Language`; fraction, Σ, and elapsed are all derived from events the renderer already receives.

## Trade-offs

- **Stream stays per-iteration; aggregation is renderer-side.** Collapsing it into bar-shaped events
  would re-couple the domain to the view — rejected per ADR-0010.
- **Bar/row keyed off `Type == Benchmark`,** not a presentation hint on the event (the domain must
  not dictate the view). `Language` is an explicit field, not parsed back out of the opaque TaskID.
- **`tasks.Result` is unchanged.** Plain buffers the Language from the Finished event rather than
  adding it to the serialized `Result` type ADR-0010 froze.
- **Settled bar shows Σ, not mean/stddev.** Full statistics live in `benchmark.json` / analyze;
  mirroring them would need `calculateMetrics` piped across the event boundary.
- **Empty-output fix (precondition).** A bar must reach 100%, so every planned iteration must emit
  exactly one Finished. The old `Ok && Output != ""` guard dropped the event *and* the duration
  sample for an empty-output run; since a benchmark's measurement is its duration (`buildResult`
  already forces `StatusPassed`), the guard is removed. An iteration is now either a timeout or a
  measured duration — making benchmark stats strictly more correct.

This extends [ADR-0010](0010-presentation-as-event-stream.md); it does not supersede it.

# All domain operations are App methods; cmd/ holds no domain seams

Every user-facing domain operation is a method on `App`: `Solve`, `Test`, `Benchmark`,
`Visualize`, and now `Add` (the `download` command) and `Analyze` (the `analyze` command).
`App` builds the underlying domain object (Exercise+Runner, `exercise.Adder`,
`analyze.Analyzer`) from its own invariant infrastructure plus the per-call arguments, and
returns the result data; presentation stays in `cmd/`.

This completes the direction ADR-0005 set for `cmd/`: a command constructs a real `App`
and calls a method on it directly. `cmd/` packages define **no domain interfaces and no
domain factory variables**. The two exceptions that previously diverged — `download`
(a local `Downloader`/`Adder` interface + `makeAdder`) and `analyze` (a local `Analyzer`
interface + `makeAnalyzer`) — diverged only because `Add`/`Analyze` were not yet `App`
methods. Promoting them removed the divergence; their generated mocks
(`mocks/download`, `mocks/analyze`) were deleted.

`App` holds the resolved `config.Config` (alongside its other resolved fields) because
`exercise.NewAdder` is config-driven — it reads eight values off the config. Mirroring all
of them as discrete `App` fields would turn `App` into a config clone; holding the config
object (itself invariant infrastructure, consistent with ADR-0004) keeps `App.Add` a thin
wrapper.

**`makeConfig` is the one sanctioned `cmd/` factory variable.** Every command keeps it so
tests can stub configuration without touching the real filesystem or environment. It is not
a domain seam — it constructs `config.Config`, not a domain object — so it does not
contradict the "no factory variables" rule, which targets *domain* operation seams.

**Testing follows the seam.** Domain behavior is tested where it lives: `App.Add`/
`App.Analyze` in `pkg/app`, and the `Adder`/`Analyzer` themselves in `pkg/exercise`/
`pkg/analyze`. `cmd/` tests cover command construction and the `makeConfig` error path only.

The deliberate cost: with the domain seam gone, `cmd/` tests can no longer assert that a
flag value (e.g. `--lang`, `--graph`) reached the operation's argument without running the
real operation. Those flag→argument assertions were dropped rather than reintroduce a
domain seam solely for them; a broken flag now surfaces end-to-end and via the `pkg/app`
tests, not in a focused `cmd/` unit test. This trade was accepted to keep `cmd/` free of
domain seams; revisit it only if flag-resolution regressions actually appear.

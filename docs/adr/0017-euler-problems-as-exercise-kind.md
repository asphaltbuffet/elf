# Project Euler problems as a second Exercise Kind under a unified `add` verb

elf was built for one challenge source (Advent of Code): an [[Exercise]]'s identity is year + day,
it has a source URL and a per-user input, and it always has two parts. Supporting Project Euler —
problems identified by a bare number, with no URL, usually no input, and a single answer — required
deciding whether Euler is *the same application with a second kind of exercise* or *a different
application that happens to share code*.

We chose **same application, two Kinds**. `Exercise` becomes an umbrella value type with a `Kind`
discriminator: a **Puzzle** (AoC — year/day/url, two parts, per-user input) or a **Problem** (Euler
— number only, one part, optional input). Everything below identity — input data, test cases,
answer verification, Runner, Task, Result, benchmarking, analysis — is kind-agnostic and reused
unchanged. The number of parts an exercise has is **declared by the exercise** and iterated by the
solve/test/benchmark driver, rather than hardcoded to two; a Problem declares only Part One and so
never runs a phantom Part Two. The fixed wire values of `Part` (1/2/3, ADR-0002) are untouched — a
Problem simply reuses Part One.

On the CLI, exercise creation unifies under `add`, one subcommand per kind: `elf add aoc <url>`
(routes to the [[Exercise Adder]]) and `elf add euler <number>` (routes to the new
[[Problem Adder]]). Both orchestrators produce a finished Exercise and hand it to the *same*
[[Exercise Scaffold]]; they differ only in how the Exercise is obtained (network fetch vs. a number
and title the user supplies). `elf download <url>` is kept as a **deprecated alias** for
`elf add aoc <url>`.

Euler solutions live in the *same solutions repo* as AoC, in a sibling `euler/<number>/<language>/`
tree (unpadded number). *(The exact location — originally derived from the AoC base dir — is
superseded by ADR-0018, which makes the Euler directory an independent `euler.dir` config key.)* Because they share the repo's Go module but sit in different directories,
compiled-language stubs need Euler-specific templates that do not borrow AoC's `common` base and
pick a valid per-exercise package/crate identity. Go (bare-file) and Rust (manifest-based) are the
first two languages scaffolded, deliberately exercising both scaffolding shapes from ADR-0016
before the interpreted languages follow.

**Delivery is phased.** The first cut supports `add euler`, `solve`, and `test` (Go and Rust).
`benchmark` and `analyze` are deferred because they are structurally two-part (`ImplementationData`
has literal `PartOne`/`PartTwo` fields; the analyze grid renders a Part Two column) — making them
Euler-correct is a larger, separable change. Until it lands, both commands detect `KindProblem` and
return `ErrEulerUnsupported` rather than emitting a misleading Part Two measurement. The deferral is
an explicit guard, not silent breakage.

## Considered options

**A separate Project Euler application**: rejected — the execution machinery (Runner protocol,
Task/Result, benchmarking, analysis, the four commands `solve`/`test`/`benchmark`/`analyze`) is
entirely identity-agnostic and transfers to Euler unchanged. Forking it would duplicate the most
valuable and most-tested part of elf to avoid relaxing a narrow band of identity/validation code.
The domains proved *not* too different once "two parts" and "has a URL/input" were recognized as
per-kind facts rather than universal ones.

**Force Euler into the existing year/day shape** (e.g. `Day` = problem number, `Year` = 0):
rejected — a field named `Day` holding `42` misleads every future reader, and `Year == 0` is
exactly the sentinel `loadInfo` already treats as invalid. An explicit `Kind` + `Number` with
per-kind validation keeps `Day`/`Year` meaning what they say.

**Arbitrary N-part model** (Answers/TestCases as maps, `Part` enum grows, N-row analyze grids):
rejected for now — it revisits the ADR-0002 wire protocol and every runner template for a
generality neither source needs (AoC has 2 parts, Euler has 1). The **declared part-set** model
(an exercise names which of the fixed parts it has) gets single-part Euler for free without
touching the protocol or the language wrappers.

**Keep `download` as the AoC verb, add `add euler` beside it**: rejected — `add` and `download`
would then name the same category of action asymmetrically. Unifying under `add` makes the CLI
match the domain name ("Exercise Adder") that CONTEXT.md already committed to; the alias preserves
existing usage.

## Consequences

- `loadInfo` branches on `Kind`: a Puzzle requires year/day/url; a Problem requires a number. The
  old unconditional four-field check is now the Puzzle branch only.
- The solve/test/benchmark driver iterates an exercise's *declared* parts. Any code that assumed
  exactly `{PartOne, PartTwo}` (solver, tester, benchmarker, and the analyze grid) reads the
  declared set instead. The analyze exercise-scope grid shows only declared parts, so a Problem
  renders a single-part column rather than an empty Part Two.
- Enumerating Problems (for `analyze euler/`) must sort **numerically**, because the unpadded
  directory names sort lexically wrong (`1, 10, 100, 2`). AoC's padded days never needed this.
- `euler/` behaves like a year for Analysis Year scope: `elf analyze euler/` produces a
  cross-problem comparison, the Euler analogue of a year graph.
- Adding the remaining languages (Python, Bash, C#, Fortran-77, Lua) to Euler follows the same
  pattern: a kind-aware solution stub and, for compiled languages, wrapper/manifest handling that
  does not reference the AoC `common` base.
- The `Adder`/`Downloader` renaming already recorded in ADR-0014 and CONTEXT.md reaches its
  conclusion here: `add` is the domain-aligned verb, `download` survives only as a compatibility
  alias.

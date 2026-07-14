# Domain Glossary

## App

The application object. Owns infrastructure that is invariant across the application lifetime:
filesystem (`afero.Fs`), logger, and resolved config values (base dir, cache dir, token,
default language). Exposes domain operations as methods. Per-call dependencies — Exercise,
Runner, writer, result callback — are explicit method parameters, not fields on App.

## Exercise

A challenge — its identity, optional input data, and test cases. A value type *produced from a
source*: either read from disk (an existing `info.json`) or assembled from a download. Either way
it is a finished value before anything writes it to disk — it is never built up in place across
stages.

An Exercise comes in one of two [[Kind]]s — a [[Puzzle]] (Advent of Code) or a [[Problem]]
(Project Euler) — which differ only in *identity and origin*, not in how they are executed. The
shared machinery below identity (input data, test cases, expected answers, Runner, Task, Result,
benchmarking, analysis) is kind-agnostic.

**Fields common to both kinds:** kind, title, the [[Part]]s it declares, input data (optional),
test cases, expected answers, filesystem path.

**Not** a session or runner context. An Exercise has no opinion about how it is executed, fetched,
or scaffolded.

## Kind

Which family of challenge an [[Exercise]] belongs to. Exactly one of:

- **[[Puzzle]]** — Advent of Code. Identified by year + day; has a source URL and a per-user input
  file; always declares two [[Part]]s.
- **[[Problem]]** — Project Euler. Identified by a bare problem number; has no URL and no input by
  default; declares a single [[Part]].

Kind is a discriminator stored in `info.json` and drives per-kind validation when an Exercise is
loaded from disk (a Puzzle requires year/day/url; a Problem requires a number). It does **not**
change execution: `solve`, `test`, `benchmark`, and `analyze` operate on any Kind.

_Avoid_: type, category, source

## Puzzle

The Advent of Code [[Kind]] of [[Exercise]]. Identity is year + day; carries a source URL and a
per-user `input.txt`. Always declares [[Part]]s One and Two. Laid out at
`exercises/<year>/<day>-<title>/<language>/`. This is the kind the [[Exercise Adder]] and
[[Page Fetcher]] produce.

_Avoid_: AoC exercise, advent puzzle (redundant), day

## Problem

The Project Euler [[Kind]] of [[Exercise]]. Identity is a single **problem number** (e.g. 42) — no
year, no day, no URL. Declares exactly one [[Part]] (reusing Part One). Input is **optional** and
defaults to none: most Euler problems are self-contained in their prose, so no `input.txt` is
written; the minority that reference a data file get one supplied manually via the custom-input
mechanism. Laid out at `<euler-dir>/<number>/<language>/` with the number **unpadded**
(`euler/42/`, `euler/100/`) — so anything enumerating problems must sort *numerically*, not
lexically. The Euler directory is configured **independently** of the AoC exercise directory (its
own `euler.dir` key, default `euler`), so the two trees are siblings a user places as they like —
they are not derived from a shared base path. Its
`info.json` carries kind, number, title, and declared parts; it is never downloaded (the user
reads the prompt on the Project Euler site themselves). `euler/` as a whole behaves like a year for
[[Analysis]] Year scope: a cross-problem comparison.

_Avoid_: puzzle, exercise (too generic), euler day, kata

## Exercise Adder

Makes a [[Puzzle]] exist in the workspace. Given a puzzle URL, it produces an Exercise — fetching
the puzzle page and input via the [[Page Fetcher]] on a cache miss, or reading an existing
`info.json` from disk — and then lays it out via the [[Exercise Scaffold]]. Owns the puzzle URL,
the implementation language, and the overwrite policy; holds the resulting filesystem path and the
scaffold report once it has run. Surface: `Add() → error`, then `FilePath()` and `Report()`.

It is the AoC-kind orchestrator; its Euler-kind sibling is the [[Problem Adder]]. Both produce a
finished Exercise and hand it to the *same* [[Exercise Scaffold]] — the difference is entirely in
*how they obtain the Exercise* (network fetch vs. a number and a title the user supplies), never in
how it is laid out.

The name is deliberately *not* "download": on a cache hit nothing is fetched, and the durable
result is an exercise on disk ready to solve, not a network transfer. The per-file results of an
Add are [[Scaffold Outcome]]s (Added / Skipped / Replaced).

_Avoid_: downloader, fetcher, getter

## Problem Adder

Makes a [[Problem]] exist in the workspace. Given a problem number and a language, it builds a
Problem-kind Exercise in memory (no network, no [[Page Fetcher]] — the user supplies the title;
the prose lives on the Project Euler site) and lays it out via the *same* [[Exercise Scaffold]] as
the [[Exercise Adder]]. Writes no `input.txt` by default (a Problem's input is optional). Resolves
the target directory from its **own** `euler.dir` config key (default `euler`), independent of the
AoC exercise directory, and falls back to the configured default language when none is given on the
CLI — mirroring how the [[Exercise Adder]] seeds its language. The Euler-kind counterpart to the
[[Exercise Adder]].

_Avoid_: downloader (nothing is downloaded), euler fetcher, problem creator

## Add (CLI verb)

The user-facing command that makes an Exercise exist, one subcommand per [[Kind]]:
`elf add aoc <url>` (routes to the [[Exercise Adder]]) and `elf add euler <number>` (routes to the
[[Problem Adder]]). `add` is now the umbrella verb, matching the domain name "[[Exercise Adder]]".
`elf download <url>` is retained as a **deprecated alias** for `elf add aoc <url>` so existing
usage and the AoC solving workflow keep working. See [[ADR 0017]].

_Avoid_: download (deprecated as the primary verb), new, create, fetch

## Page Fetcher

Fetches puzzle page HTML and puzzle input from Advent of Code, with on-disk caching. Owns the HTTP
client, the session token, and the cache directory. It builds and fully configures its own client
(base URL and User-Agent) at construction — callers never configure it externally. Surface:
`fetchPage(year, day) → bytes`, `fetchInput(year, day) → bytes` — a cache check followed by an HTTP
request on miss. Knows the AoC URL shape and the `cacheDir/pages` + `cacheDir/inputs` cache layout,
and nothing about the exercise directory on disk.

_Avoid_: HTTP client, downloader, page loader

## Exercise Scaffold

Lays a finished Exercise out on disk: creates the implementation directory, writes `input.txt` and
`info.json`, and renders the language template files. Owns the filesystem, the input file name, and
the overwrite policy. Surface: `write(exercise) → (report, error)`, where the report records the
[[Scaffold Outcome]] of each file laid out. Reads the Exercise; never mutates it.
Knows the exercise directory layout and nothing about HTTP or the cache.

_Avoid_: file writer, template renderer, downloader

## Scaffold Outcome

What the Exercise Scaffold did with one file. Exactly one of:

- **Added** — the file did not exist; elf wrote it.
- **Skipped** — the file already existed and the overwrite policy left it untouched.
- **Replaced** — the file already existed and the overwrite policy replaced it.

These describe the *action elf took*, not the file's prior state. Template files (README, the
language solution stub) are never replaced under the current policy, so they only ever report Added
or Skipped.

`encrypt` and `decrypt` report per-file in the same vocabulary. `encrypt` writes each `.age`
sibling: **Added** when none existed, **Replaced** when refreshing an existing one (the plaintext is
the source of truth and may have changed since the last encrypt). `decrypt` writes plaintext from an
`.age`: **Added** on a fresh clone that has only ciphertext, **Skipped** when plaintext already
exists (elf does not clobber the working copy; `--force` makes it **Replaced**). Neither verb ever
removes the other side — plaintext and ciphertext coexist.

_Avoid_: status, already-existed, overwritten, created


## Solution Set

The files of an [[Exercise]] that reveal its answer: the `info.json` (which carries the expected
answers) and every language subdirectory (`go/`, `py/`, …, the solution source). It is
[[Kind]]-agnostic — a [[Puzzle]] and a [[Problem]] both have one — and it deliberately **excludes**
`input.txt` (already gitignored; not the user's answer) and `README.md` (not sensitive). This is
the set `encrypt` encrypts (to per-file `.age` siblings) and `decrypt` restores. elf treats the
`.age` files as **derived artifacts** and never removes the plaintext: the user commits the `.age`
files and gitignores the plaintext, so a public repository carries the ciphertext without publishing
the answer — satisfying Project Euler's "do not share solutions" rule (the motivating case, though
the concept is not Euler-specific). Because plaintext stays authoritative on the working machine, a
mid-encrypt crash can never lose data; `decrypt` exists to reconstruct plaintext on a fresh clone
that has only the committed `.age` files. See [ADR-0019](docs/adr/0019-encrypt-leaves-plaintext-in-place.md)
for the artifact model and [ADR-0020](docs/adr/0020-ssh-keys-as-age-recipients.md) for the SSH-key
recipient/identity model.

_Avoid_: solution files, secrets, encrypted files, sealed files


## Runner

Abstracts language execution for any language. Lifecycle: `Prepare` (render wrapper template,
run build commands) → `Open` (start subprocess) → `Run` (one task → one result) → `Close`
(stop subprocess) → `Cleanup` (remove artifacts). Communicates via line-delimited JSON over
stdin/stdout — a public protocol, not a runner internal. Runners are not compiled into elf;
they are instantiated at runtime from Runner Descriptors read from config.

_Avoid_: plugin, language runner, implementation

## Runner Descriptor

A config entry (`[[runner]]` in `elf.toml`) that fully specifies how to build and launch a
Runner for one language. Contains: a key (used as both the registry key and the exercise
subdirectory name), a display name, an optional wrapper template path, optional static template
variables, optional ordered build commands, optional `cleanup_paths` (build-output trees removed
by `Cleanup`, e.g. `bin`/`obj` for C# or `target` for Rust), and an open spec (interpreter or
compiled binary). Populated into the runner registry at startup. The key is the only required
field that must be unique across all descriptors.

Some compiled languages are **manifest-based** (Rust, C#): the Exercise Scaffold writes a project
manifest (`Cargo.toml`, `.csproj`) alongside the solution file, pinning a fixed package/assembly
name so the descriptor's build/open tokens resolve to a static path. Others are **bare-file**
(Go, C): the compiler runs directly against the wrapper and solution files, no manifest involved.
See [ADR-0016](docs/adr/0016-per-exercise-project-manifest.md).

_Avoid_: plugin descriptor, runner config, runner definition

## Runner Registry

The runtime map from language key to `RunnerCreator`, populated at startup from Runner
Descriptors in config. Consulted whenever elf needs to execute an exercise. Previously
compile-time; now config-driven.

_Avoid_: available runners, runner map

## ExerciseMeta

The identity fields of an Exercise passed to a `RunnerCreator` at construction time: year, day,
title, and root directory path. Used by plugin runners to substitute token values into templates
and build commands without parsing the filesystem path.

_Avoid_: exercise context, runner context

## Runner Token

A placeholder (e.g. `{year}`, `{lang_dir}`, `{wrapper_file}`) substituted by elf into
descriptor fields at runtime. The fixed vocabulary is: `{exercise_dir}`, `{rel_exercise_dir}`,
`{lang_dir}`, `{wrapper_file}` (only valid when `template_path` is set — extension derived from it),
`{binary_file}`, `{year}`, `{day}`, `{title}`. User-defined values go in `template_vars`, not
new tokens.

`{rel_exercise_dir}` is the exercise directory relative to the nearest ancestor `go.mod` (the
exercise path resolved to absolute, then made relative to the enclosing Go module root). Unlike
`{exercise_dir}` — which is the CLI-supplied path verbatim, so `elf solve .` yields `.` — it is
invariant to the working directory and to how the path is spelled on the command line, giving a
stable segment for a Go import path across both [[Kind]]s (`exercises/2019/12-foo`, `euler/42`). It
is **Go-specific by design**: only the Go runner needs a module-relative import path (Rust and C#
pin crate/assembly names instead), so the anchor is `go.mod` rather than a language-neutral
workspace root.

_Avoid_: template variable, placeholder, interpolation variable

## Task

A unit of work sent to a Runner: a (Part, input) pair with a TaskID. Maps to one call to
`Runner.Run`.

## Part

One of: PartOne, PartTwo, Visualize. Identifies which sub-problem of an [[Exercise]] a Task
addresses. The wire values are fixed (1=PartOne, 2=PartTwo, 3=Visualize; see ADR-0002) and shared
by every language solution.

Which parts an Exercise *has* is per-[[Kind]] and **declared by the Exercise**, not universal: a
[[Puzzle]] declares One and Two; a [[Problem]] declares only One. The solve/test/benchmark driver
iterates the declared set rather than assuming two, so a Problem never runs a phantom Part Two.
Visualize is orthogonal — a separate, opt-in Task, not a declared solve part.

## Result

The outcome of a Task: TaskID, success flag, output string, duration.

## Visualization

The artifact produced by running a Visualize Task against an exercise. Always a file on disk; the exercise implementation decides the filename and format (SVG, HTML, PNG, etc.). The output directory is passed to the exercise via `Task.OutputDir`; the exercise writes its artifact there. The Result output string is informational only and its meaning is per-runner (the Go runner reports the output directory; other runners may report a status string) — elf displays it but never opens or resolves it as a path. The `elf visualize` command defaults `OutputDir` to the exercise's own directory (see [[ADR 0015]]) and exposes `--outdir`/`-o` to override it. A Visualization Result carries `StatusUnverified` (no expected answer exists). The file path is reported to the user; elf never opens it automatically.

_Avoid_: visualization output, vis output, render

## Iteration

One repetition of a benchmark Task. Benchmarking runs each (Runner, Part) `iterations` times to
gather a duration sample per run; the iteration index is the Task's SubPart. An iteration has
exactly two outcomes — it **times out** (yielding a timeout Result, after which the Runner is
restarted) or it **completes with a measured duration** (the duration is the sample, regardless of
the output string). There is no third "empty output" outcome: for benchmarking the duration is the
measurement, so a completed iteration always contributes a sample.

_Avoid_: run, repeat, trial, sample (a sample is the *duration* an iteration yields, not the
iteration itself)

## Progress Bar

The live (TTY) presentation of a benchmark's iterations: one bar per (Runner, Part), advancing one
tick per completed Iteration until it reaches `iterations/iterations`. Replaces the one-line-per-
iteration view — a benchmark of 100 iterations across two Runners shows four bars, not 400 lines.
While running, the bar's trailing metric is live wall-clock elapsed; settled, it shows the sum of
the iteration durations (Σ). The non-TTY (Plain) renderer shows the same information as one settled
line per (Runner, Part) rather than an animated bar. The bar is a pure view over the existing
per-iteration event stream (see [[ADR 0011]]); the Runner is identified by the event's Language.

_Avoid_: spinner, status line, per-iteration row

## Presentation Mode

How a run's [[Result]]s are surfaced to whoever invoked it. Every mode is a view over the same
lifecycle event stream (see [[ADR 0010]]); the domain formats nothing itself. Modes:

- **Live** — the animated TTY view (spinner, live timer, [[Progress Bar]]s). The default when
  stdout is a terminal.
- **Plain** — settled human text for non-TTY sinks (pipes, files, CI). Buffers results and emits
  aligned columns on close. Selected automatically off-TTY, or forced with `--plain`.
- **Machine** — a structured, stable rendering intended for programmatic consumers (agents,
  scripts). Selected with `--json`, which is mutually exclusive with `--plain`. Emits one JSON
  summary object per run: exercise metadata (omitted when absent) plus a `results` array;
  benchmark results aggregate per (runner, part). It is the same buffer-then-emit shape as Plain,
  marshalled as data rather than formatted as columns — so a caller reads answers, pass/fail, and
  runner errors as fields instead of grepping Plain text.

The distinction is *audience*, not *content*: all three carry identical [[Result]] information.
Live and Plain optimise for a human reader; Machine optimises for a parser.

_Avoid_: format, output style, JSON dump (the mode is named by audience, not by wire format)

## Analysis

Renders run-time graphs from persisted benchmark data. One operation with two scopes, inferred
from the shape of the target directory — no mode flag:

- **Exercise scope** — the target *is* an Exercise (has `info.json`). Compares language against
  language for that one exercise, across its **declared Parts**. Rendered as an **R×N grid of
  per-language consistency facets**: rows are the Parts present in the data, columns are languages,
  and each cell is an independently auto-scaled box plot of *consistency* (each sample as a
  percentage of that language's own median). It deliberately does **not** compare speed — that
  lives in the year graph — because the cross-language speed gap and each language's own spread
  cannot share one axis. The row set is **data-present-driven**: a Part row exists only if some
  implementation has data for it, so a single-Part [[Problem]] renders a 1×N grid (no empty Part
  Two row). Within the rows that exist, a missing (language, part) is a blank-but-present cell so
  the grid stays aligned. See *Comparison vocabulary*.
- **Year scope** — the target *contains* Exercise subdirectories (is a year). Compares day against
  day across the year, with languages overlaid. Rendered as a line graph of running time vs. day,
  one line per language. Year scope is an **AoC-only** view: it proceeds only when the target holds
  at least one child and *every* child is a [[Puzzle]]. A Euler tree (children are Problems), a
  mixed tree, or an empty tree is refused — Euler Problems are a sparse, heterogeneous set with no
  meaningful cross-problem season to plot (unlike a curated AoC year). Euler is analyzable only at
  exercise scope, one Problem at a time.

In both scopes, **color always encodes language**, and a given language maps to the same color in
every graph (derived from the language key, not from its position in the data). The exercise box
plot groups by Part, with one box per language inside each Part group. Because the two scopes are
one operation, they share a single visual identity: non-color styling (fonts, grid, axis weight,
legend, background) is applied uniformly to every plot rather than tuned per scope, so a box plot
and a line graph read as siblings from the same tool.

Reads `benchmark.json` files produced by benchmarking; never produces benchmark data itself.

### Comparison vocabulary

These terms are distinct and must not be conflated — in particular, two different
"normalizations" exist:

- **Normalization factor** — a *machine-speed calibration* value persisted with each
  benchmark (the time to run a fixed synthetic workload on the benchmarking machine).
  It makes durations from different machines comparable. It is about the *machine*, not
  about languages, and it cancels out of any same-machine, same-day language comparison.
- **Reference language** — for cross-language comparison, the one language chosen as the
  anchor (its time is defined as 1.0 on every day). About the *languages*, on one machine.
  By default the reference is the *fastest* language for that exercise/day, so relative
  runtime reads as "how far behind the leader." (Anchoring instead to the user's configured
  `language` is a possible future override, not the default.)
- **Relative runtime** — a language's running time expressed as a dimensionless multiple of
  the reference language on the same day and part (e.g. "210× the reference"). Because it is
  a same-day ratio, it is inherently machine-independent and does not need the normalization
  factor applied. This is the comparison primitive for finding where a language is unusually
  strong or weak relative to the others, which absolute (log-scale) timing obscures. The
  baseline is computed *per part*: every sample in a part-group is divided by that part's
  fastest mean, so the reference language (the one sitting at 1×) may differ between the
  Part One and Part Two groups — that change is itself a visible strength signal.
- **Consistency** — how tightly a language's repeated samples for one (day, part) cluster.
  A separate axis of "strength" from speed: a language can be slower on average yet far more
  predictable. Consistency is encoded as each sample's **percentage of that language's own
  median** (sample ÷ median × 100), so the spread is dimensionless and centred at 100%. The
  median (robust to outliers) is the normaliser; the absolute median is shown alongside for
  context but is not the comparison axis. This deliberately discards cross-language *speed*
  comparison from the axis — speed lives in the relative runtime view and the year graph —
  because the ~4-orders-of-magnitude speed gap between languages and each language's own
  modest spread cannot share one axis without crushing the spread to an illegible sliver.

Scope is detected by looking at the target and at most one level below it: an `info.json` in the
target means Exercise scope; otherwise an `info.json` in an immediate child means Year scope;
anything else (e.g. a base directory of year directories) is an error rather than a silent
multi-year merge. Detection never recurses across years.

Missing data is tolerated, absent data is not: a scope with *some* benchmark data renders whatever
exists (a year with 3 of 25 days benchmarked yields a 3-point graph). A scope with *no* benchmark
data is an error that names the benchmark command as the fix, rather than producing an empty graph.

The graph is written into the target directory by default (next to the data it describes): the
exercise folder for Exercise scope, the year folder for Year scope. Year scope produces exactly
one year-level graph — it does not descend into day folders. An explicit override may redirect the
output elsewhere. On success, the resolved filepath of the written image is returned to the caller
and printed to stdout by the CLI.

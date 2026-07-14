# Domain Glossary

The shared language of **elf**, a CLI that manages programming-challenge exercises
(Advent of Code, Project Euler) — downloading them, running solutions in multiple languages,
benchmarking, and analysing the results. This file is a glossary only: it defines what each
term *is*. The *why* behind a decision lives in an ADR under `docs/adr/`, linked where relevant.

## Language

### Exercises and their kinds

**App**:
The application object. Owns the infrastructure that is invariant across the application
lifetime (filesystem, logger, resolved config) and exposes the domain operations as methods.
See [ADR-0004](docs/adr/0004-app-owns-invariant-infrastructure.md).
_Avoid_: context, service, container

**Exercise**:
A challenge — its identity, optional input data, and test cases — as a finished value produced
from a source (read from disk or assembled from a download), never built up in place. Comes in
one of two **Kind**s and is agnostic about how it is executed, fetched, or scaffolded.
_Avoid_: kata, task, session

**Kind**:
Which family a **Exercise** belongs to — exactly one of **Puzzle** or **Problem**. A
discriminator that drives per-kind identity and validation but never changes execution. See
[ADR-0017](docs/adr/0017-euler-problems-as-exercise-kind.md).
_Avoid_: type, category, source

**Puzzle**:
The Advent of Code **Kind** of **Exercise**: identity is year + day, carries a source URL and a
per-user input, and always declares **Part**s One and Two. The kind produced by the
**Exercise Adder**.
_Avoid_: AoC exercise, advent puzzle, day

**Problem**:
The Project Euler **Kind** of **Exercise**: identity is a single problem number, with no URL,
input optional, and a single declared **Part** (reusing Part One). Its tree is configured
independently of the AoC one and behaves as a peer of a year for **Analysis**. See
[ADR-0018](docs/adr/0018-euler-directory-as-independent-config-key.md).
_Avoid_: puzzle, exercise (too generic), euler day, kata

### Making an exercise exist

**Add** (CLI verb):
The user-facing command that makes an **Exercise** exist, one subcommand per **Kind**
(`elf add aoc <url>`, `elf add euler <number>`); `elf download` survives as a deprecated alias.
See [ADR-0017](docs/adr/0017-euler-problems-as-exercise-kind.md).
_Avoid_: download (deprecated as the primary verb), new, create, fetch

**Exercise Adder**:
The orchestrator that makes a **Puzzle** exist: obtains its **Exercise** (via the **Page
Fetcher** on a cache miss, or from an existing `info.json`) and lays it out via the **Exercise
Scaffold**. The name is deliberately not "download" — on a cache hit nothing is fetched.
_Avoid_: downloader, fetcher, getter

**Problem Adder**:
The Euler-kind counterpart to the **Exercise Adder**: builds a **Problem** **Exercise** in memory
from a number and title the user supplies (no network) and lays it out via the same **Exercise
Scaffold**.
_Avoid_: downloader, euler fetcher, problem creator

**Page Fetcher**:
The component that fetches Advent of Code puzzle-page HTML and puzzle input, with on-disk caching.
Knows the AoC URL shape and the cache layout, and nothing about the exercise directory on disk.
_Avoid_: HTTP client, downloader, page loader

**Exercise Scaffold**:
The component that lays a finished **Exercise** out on disk: the implementation directory, the
input and info files, and the rendered language templates. Reads the **Exercise** and never
mutates it; knows the directory layout and nothing about HTTP or the cache. See
[ADR-0016](docs/adr/0016-per-exercise-project-manifest.md).
_Avoid_: file writer, template renderer, downloader

**Scaffold Outcome**:
What elf did with one file it laid out — exactly one of **Added** (wrote a file that did not
exist), **Skipped** (left an existing file untouched), or **Replaced** (overwrote an existing
file). Describes the action elf took, not the file's prior state; `encrypt`/`decrypt` report in
the same vocabulary. See [ADR-0019](docs/adr/0019-encrypt-leaves-plaintext-in-place.md).
_Avoid_: status, already-existed, overwritten, created

**Solution Set**:
The files of an **Exercise** that reveal its answer — the `info.json` plus every language
subdirectory — deliberately excluding input and README. This is the set `encrypt` seals to `.age`
siblings and `decrypt` restores. See
[ADR-0019](docs/adr/0019-encrypt-leaves-plaintext-in-place.md) and
[ADR-0020](docs/adr/0020-ssh-keys-as-age-recipients.md).
_Avoid_: solution files, secrets, encrypted files, sealed files

### Running an exercise

**Runner**:
The abstraction of language execution: a lifecycle of Prepare → Open → Run → Close → Cleanup,
communicating over a public line-delimited JSON protocol. Instantiated at runtime from **Runner
Descriptor**s, not compiled into elf. See
[ADR-0003](docs/adr/0003-runner-lifecycle-stages.md) and
[ADR-0002](docs/adr/0002-runner-protocol-as-public-contract.md).
_Avoid_: plugin, language runner, implementation

**Runner Descriptor**:
A config entry (`[[runner]]` in `elf.toml`) that fully specifies how to build and launch a
**Runner** for one language, keyed by a unique language key. A compiled language may be
manifest-based (Rust, C#) or bare-file (Go, C). See
[ADR-0006](docs/adr/0006-runner-plugin-system.md) and
[ADR-0016](docs/adr/0016-per-exercise-project-manifest.md).
_Avoid_: plugin descriptor, runner config, runner definition

**Runner Registry**:
The runtime map from language key to runner, populated at startup from **Runner Descriptor**s and
consulted whenever elf executes an exercise. See
[ADR-0006](docs/adr/0006-runner-plugin-system.md).
_Avoid_: available runners, runner map

**ExerciseMeta**:
The identity fields of an **Exercise** (year, day, title, root path) handed to a runner at
construction so it can substitute **Runner Token**s without parsing the filesystem path.
_Avoid_: exercise context, runner context

**Runner Token**:
A placeholder (e.g. `{year}`, `{lang_dir}`, `{wrapper_file}`) that elf substitutes into a
**Runner Descriptor**'s fields at runtime; user-defined values go in `template_vars`, not new
tokens. `{rel_exercise_dir}` is the Go-specific, module-relative variant used for import paths.
See [ADR-0021](docs/adr/0021-rel-exercise-dir-token-anchored-to-go-mod.md).
_Avoid_: template variable, placeholder, interpolation variable

**Task**:
A unit of work sent to a **Runner**: a (**Part**, input) pair with a TaskID, mapping to one
`Runner.Run` call.
_Avoid_: job, unit, request

**Part**:
One of PartOne, PartTwo, or Visualize — which sub-problem of an **Exercise** a **Task** addresses.
Which parts an **Exercise** *has* is declared by the exercise and iterated by the driver, not
universal. Wire values are fixed (1/2/3). See
[ADR-0002](docs/adr/0002-runner-protocol-as-public-contract.md).
_Avoid_: stage, phase, step

**Result**:
The outcome of a **Task**: TaskID, success flag, output string, and duration.
_Avoid_: outcome, response, answer

**Iteration**:
One repetition of a benchmark **Task**. Each (Runner, Part) is run `iterations` times; an
iteration either times out or completes with a measured duration — there is no third outcome.
_Avoid_: run, repeat, trial, sample (a sample is the *duration* an iteration yields)

**Visualization**:
The file-on-disk artifact produced by running a Visualize **Task**; the exercise decides its
filename and format and writes it to a passed-in output directory. Its **Result** is always
unverified and its path is reported but never opened by elf. See
[ADR-0015](docs/adr/0015-visualize-outputs-beside-the-exercise.md).
_Avoid_: visualization output, vis output, render

### Presenting and analysing results

**Presentation Mode**:
How a run's **Result**s are surfaced — **Live** (animated TTY), **Plain** (settled text for
non-TTY), or **Machine** (structured JSON for parsers). The distinction is audience, not content:
all three carry identical **Result** information. Every mode is a view over one lifecycle event
stream. See [ADR-0010](docs/adr/0010-presentation-as-event-stream.md).
_Avoid_: format, output style, JSON dump

**Progress Bar**:
The live (TTY) presentation of a benchmark's **Iteration**s: one bar per (Runner, Part), advancing
one tick per completed iteration. A pure view over the per-iteration event stream. See
[ADR-0011](docs/adr/0011-benchmark-progress-bars-as-stream-aggregation.md).
_Avoid_: spinner, status line, per-iteration row

**Analysis**:
The operation that renders run-time graphs from persisted benchmark data, in one of two scopes
inferred from the target directory — **Exercise scope** (per-language consistency facets for one
exercise) or **Year scope** (running-time-vs-day line graph across an AoC year, AoC-only). Never
produces benchmark data itself. See
[ADR-0007](docs/adr/0007-relative-runtime-box-plot.md),
[ADR-0008](docs/adr/0008-per-language-consistency-facets.md), and
[ADR-0022](docs/adr/0022-euler-analyze-exercise-scope-only.md).
_Avoid_: report, plotting, charting

**Normalization factor**:
A machine-speed calibration value persisted with each benchmark, making durations from different
machines comparable. About the *machine*, not languages; cancels out of any same-machine
comparison. See [ADR-0007](docs/adr/0007-relative-runtime-box-plot.md).
_Avoid_: normalization (ambiguous — see also Consistency), calibration, scaling factor

**Reference language**:
For cross-language comparison, the one language chosen as the anchor (defined as 1.0 on every day),
by default the fastest for that exercise/day. About the *languages*, on one machine. See
[ADR-0007](docs/adr/0007-relative-runtime-box-plot.md).
_Avoid_: baseline language, anchor, reference implementation

**Relative runtime**:
A language's running time expressed as a dimensionless multiple of the **Reference language** on
the same day and part (e.g. "210× the reference") — inherently machine-independent. See
[ADR-0007](docs/adr/0007-relative-runtime-box-plot.md).
_Avoid_: speedup, ratio, normalized time

**Consistency**:
How tightly a language's repeated samples for one (day, part) cluster — a strength axis separate
from speed (a language can be slower on average yet more predictable). See
[ADR-0008](docs/adr/0008-per-language-consistency-facets.md).
_Avoid_: variance, spread, normalization (ambiguous — see also Normalization factor)

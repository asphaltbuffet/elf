# Domain Glossary

## App

The application object. Owns infrastructure that is invariant across the application lifetime:
filesystem (`afero.Fs`), logger, and resolved config values (base dir, cache dir, token,
default language). Exposes domain operations as methods. Per-call dependencies — Exercise,
Runner, writer, result callback — are explicit method parameters, not fields on App.

## Exercise

A puzzle — its identity, input data, and test cases. A value type *produced from a source*:
either read from disk (an existing `info.json`) or assembled from a download. Either way it is a
finished value before anything writes it to disk — it is never built up in place across stages.

**Fields:** year, day, title, URL, input data, test cases, expected answers, filesystem path.

**Not** a session or runner context. An Exercise has no opinion about how it is executed, fetched,
or scaffolded.

## Page Fetcher

Fetches puzzle page HTML and puzzle input from Advent of Code, with on-disk caching. Owns the HTTP
client, the session token, and the cache directory. Surface: `fetchPage(year, day) → bytes`,
`fetchInput(year, day) → bytes` — a cache check followed by an HTTP request on miss. Knows the AoC
URL shape and the `cacheDir/pages` + `cacheDir/inputs` cache layout, and nothing about the exercise
directory on disk.

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

_Avoid_: status, already-existed, overwritten, created


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
variables, optional ordered build commands, and an open spec (interpreter or compiled binary).
Populated into the runner registry at startup. The key is the only required field that must be
unique across all descriptors.

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
descriptor fields at runtime. The fixed vocabulary is: `{exercise_dir}`, `{lang_dir}`,
`{wrapper_file}` (only valid when `template_path` is set — extension derived from it),
`{binary_file}`, `{year}`, `{day}`, `{title}`. User-defined values go in `template_vars`, not
new tokens.

_Avoid_: template variable, placeholder, interpolation variable

## Task

A unit of work sent to a Runner: a (Part, input) pair with a TaskID. Maps to one call to
`Runner.Run`.

## Part

One of: PartOne, PartTwo, Visualize. Identifies which sub-problem of a puzzle a Task addresses.

## Result

The outcome of a Task: TaskID, success flag, output string, duration.

## Analysis

Renders run-time graphs from persisted benchmark data. One operation with two scopes, inferred
from the shape of the target directory — no mode flag:

- **Exercise scope** — the target *is* an Exercise (has `info.json`). Compares language against
  language for that one puzzle, across both Parts. Rendered as a box plot grouped by language,
  using the raw timing samples so the distribution (median, quartiles, outliers) is visible — not
  just the mean.
- **Year scope** — the target *contains* Exercise subdirectories (is a year). Compares day against
  day across the year, with languages overlaid. Rendered as a line graph of running time vs. day,
  one line per language.

In both scopes, **color always encodes language**, and a given language maps to the same color in
every graph (derived from the language key, not from its position in the data). The exercise box
plot groups by Part, with one box per language inside each Part group. Because the two scopes are
one operation, they share a single visual identity: non-color styling (fonts, grid, axis weight,
legend, background) is applied uniformly to every plot rather than tuned per scope, so a box plot
and a line graph read as siblings from the same tool.

Reads `benchmark.json` files produced by benchmarking; never produces benchmark data itself.

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
output elsewhere.

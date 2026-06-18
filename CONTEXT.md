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
the overwrite policy. Surface: `write(exercise) → error`. Reads the Exercise; never mutates it.
Knows the exercise directory layout and nothing about HTTP or the cache.

_Avoid_: file writer, template renderer, downloader


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

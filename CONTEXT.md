# Domain Glossary

## App

The application object. Owns infrastructure that is invariant across the application lifetime:
filesystem (`afero.Fs`), logger, and resolved config values (base dir, cache dir, token,
default language). Exposes domain operations as methods. Per-call dependencies — Exercise,
Runner, writer, result callback — are explicit method parameters, not fields on App.

## Exercise

A puzzle — its identity, input data, and test cases. A value type loaded from disk.

**Fields:** year, day, title, URL, input data, test cases, expected answers, filesystem path.

**Not** a session or runner context. An Exercise has no opinion about how it is executed.


## Runner

Abstracts language execution (Go, Python, or any out-of-process executable).
Lifecycle: `Prepare` (build/write artifacts) → `Open` (start subprocess) → `Run` (one task →
one result) → `Close` (stop subprocess) → `Cleanup` (remove artifacts). `Prepare` and
`Cleanup` are no-ops for pre-built or PATH-discovered runners. Communicates via
line-delimited JSON over stdin/stdout — a public protocol, not a runner internal.

## Task

A unit of work sent to a Runner: a (Part, input) pair with a TaskID. Maps to one call to
`Runner.Run`.

## Part

One of: PartOne, PartTwo, Visualize. Identifies which sub-problem of a puzzle a Task addresses.

## Result

The outcome of a Task: TaskID, success flag, output string, duration.

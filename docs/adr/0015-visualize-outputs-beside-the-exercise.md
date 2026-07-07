# `visualize` writes beside the exercise, not into the current directory

`elf visualize` defaults `Task.OutputDir` to the target exercise's own directory
(where its `info.json` lives), not to the process's current working directory.
`--outdir`/`-o` still overrides it.

## Context

The original default was the current working directory. Command-history analysis of
real usage (both interactive and agent-driven) showed `--outdir` supplied on 205 of 209
`visualize` invocations — a ~98% override rate — and the supplied value was almost always
the exercise's own directory or a scratch path. A default that is overridden 98% of the
time is the wrong default: it forces every caller to compute and pass the path the tool
could have derived itself, and it is the single largest source of ceremony in the
`visualize` command line.

The current-directory default also produces a surprising failure mode: running
`elf visualize exercises/2018/17-reservoirResearch` from the repo root silently drops the
artifact in the repo root, detached from the exercise it describes, rather than next to it.

## Decision

Default `OutputDir` to the resolved exercise directory. A [[Visualization]] is
conceptually an artifact *of* an exercise, so co-locating it with the exercise's other
files (`info.json`, `input.txt`, language sources) is the least-surprising home. Callers
who want it elsewhere (a scratch preview, a docs folder) still say so with `--outdir`.

## Consequences

This is a behavior change to a default, so it is mildly hard to reverse: any script or
habit that relied on the artifact landing in cwd will now find it beside the exercise.
Because the artifact path is always reported on stdout (elf never opens it automatically),
a caller that reads the reported path rather than assuming cwd is unaffected.

## Alternatives considered

**Keep cwd, document better.** Rejected: documentation does not fix a default that is
wrong 98% of the time; the friction is in the typing, not the understanding.

**Default to cwd but special-case agents.** Rejected: there is no reliable "am I an agent"
signal, and the exercise directory is the better default for humans too — the override
rate is high in interactive history as well.

# Euler solutions directory is an independent config key, not derived from the AoC base dir

ADR-0017 laid [[Problem]]s out "in a sibling `euler/<number>/` tree" and the first implementation
expressed this as `filepath.Join(baseDir, "euler", number)`, where `baseDir` is the AoC exercise
directory (`advent.dir`). That coupling is wrong: a user whose repo keeps `exercises/` and `euler/`
as **independent siblings under the repo root** cannot express it. One config value (`advent.dir`)
was being asked to place two independent trees, so setting it to `euler` to fix the Euler path
(`euler/2`) simultaneously broke AoC (`euler/2024/...`) and stacked the hardcoded literal into
`euler/euler/2`.

We decided the Euler directory is its **own** configuration key, `euler.dir` (default `euler`),
read via `GetEulerDir()`. The [[Problem Adder]]'s path becomes `filepath.Join(eulerDir, number)` —
the hardcoded `"euler"` literal is removed; the directory name now *is* the config value. AoC keeps
`advent.dir` (`exercises`) untouched. The two trees are configured, and therefore placed,
independently.

This **supersedes the path/layout claim in ADR-0017** (the sibling-derived-from-base-dir wording);
everything else in ADR-0017 — Euler as a second Exercise Kind, the unified `add` verb, the declared
part-set — stands.

## Considered options

- **Separate `euler.dir` key (chosen).** Smallest change that decouples the two trees; matches the
  existing per-concern key layout (`advent.dir`, `advent.token`). AoC config is untouched.
- **Single "solutions root" + per-kind subdirs** (`root` + `aoc-subdir` + `euler-subdir`): more
  coherent long-term but a larger redesign that churns the AoC config and every `GetBaseDir()`
  caller for generality the two-source workspace does not yet need. Deferred.
- **Drop the literal and reuse `advent.dir`:** rejected — it forces AoC and Euler to share one
  value, which is exactly the coupling that produced the bug.

## Consequences

- New public config surface: `euler.dir` key, `GetEulerDir()` getter, a `[euler]` section in the
  generated default config, and a matching `programs.elf.euler.dir` option in the home-manager
  module (kept in sync with `pkg/config/defaults.go` per the module's contract).
- `newProblemFromSource` no longer contains a hardcoded `"euler"` path segment; the segment is
  supplied entirely by configuration, so `euler.dir = "."` places bare-numbered problem dirs at the
  repo root and any other value renames the subtree.

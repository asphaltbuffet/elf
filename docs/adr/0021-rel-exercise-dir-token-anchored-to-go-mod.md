# `{rel_exercise_dir}` runner token, anchored to the nearest `go.mod`

The Go runner wrapper `import`s the user's exercise package by its **module-qualified path**, so a Go
[[Runner Descriptor]] carries `import_path = "github.com/.../{year}/{dir_name}/go"`. Encoding the
location as AoC identity fields breaks once a second [[Kind]] exists: a [[Problem]] (Euler) has no
year and lives under a different root (`euler.dir`, ADR-0018). One `import_path` string can name only
one family's layout, so users toggled two commented lines — and a stale toggle silently produced
`.../euler/<name>/go` for AoC exercises, which fail to build.

The existing `{exercise_dir}` token carries the family root for both Kinds, but it is the
**CLI-supplied path verbatim**: `elf solve .` yields `.`, producing the invalid import `..././go`.
The import path is anchored to the **Go module root**, and no existing token speaks in those
coordinates.

We add **`{rel_exercise_dir}`**: the exercise directory made relative to the nearest ancestor
`go.mod`. It yields `exercises/2019/12-foo` / `euler/42` for a shared-root module and `.` for a
per-exercise module (Exercism-style). One line now serves every Kind and layout:
`import_path = "github.com/.../{rel_exercise_dir}/go"`.

Two scoping choices: it is **Go-specific** (only Go needs a module-relative import; Rust/C# pin
crate/assembly names), so the anchor is `go.mod`, not a workspace root. And it is resolved in
`Prepare` — the first token doing filesystem I/O that can fail — **lazily**, only when a descriptor
references it, so non-Go runners with no module are untouched and `substituteTokens` stays pure.

## Considered options

- **`{rel_exercise_dir}` anchored to `go.mod` (chosen).** No new config; correct regardless of
  working directory or `elf solve .`. Go-specific, which is honest — Go alone has the problem.
- **Reuse `{exercise_dir}`, no code change.** Rejected: fails for `elf solve .` (yields `.`), a
  workflow the user wants.
- **A workspace-root config key.** Rejected/deferred — the "single solutions root" redesign ADR-0018
  already deferred; no non-Go runner needs it.
- **`error`-returning `substituteTokens`, or fall back to `{exercise_dir}` on no-`go.mod`.** Rejected:
  the first churns a pure function every token flows through; the second silently reintroduces the
  broken-import bug this token exists to kill. `Prepare` fails loudly instead.

## Consequences

- New entry in the [[Runner Token]] vocabulary; CONTEXT.md glossary updated.
- `Prepare` gains a lazy step: if a descriptor references the token, walk up from the absolute
  exercise dir to the first `go.mod`, compute `filepath.Rel(root, dir)`, and pass it into
  `substituteTokens`. No reference → no walk; match but no `go.mod` → `Prepare` errors.
- The walk uses real `os`/`filepath` (not `afero.Fs`), consistent with the runner layer, which shells
  out to a real compiler.
- `.` is a valid output (per-exercise module). A shared-URL template over a per-exercise module would
  render an invalid `.../ ./go`, but that combination is nonsensical and unsupported — not
  special-cased.

# Per-exercise project manifest for ecosystem-idiomatic compiled languages

Some compiled languages don't compile a bare source file the way Go/C do — their toolchain is
built around a project manifest (`Cargo.toml`, `.csproj`) that names dependencies and pins build
output. For these languages, the Exercise Scaffold emits **two** files per exercise instead of one:
the manifest and the solution source, both under the language's `{lang_dir}`. The manifest pins a
fixed package/assembly name (Rust: `name = "solution"`; C#: `<AssemblyName>wrapper</AssemblyName>`)
so the Runner Descriptor's build/open tokens resolve to a static path regardless of the exercise's
year/day/title. Rust established this shape first (`rs-cargo.tmpl` + `rs-solution.tmpl`); C# is the
second adopter (`Solution.csproj` + `Solution.cs`).

## Considered options

**One shared manifest for all exercises in a language**: rejected — breaks the "each exercise
directory is self-contained" pattern every other language follows (Go, Python, Bash all scaffold
fully independent exercise directories), and requires a mechanism to swap which exercise's source
file the shared manifest points at per run.

**Bare-file compile, no manifest** (`csc`/`rustc` invoked directly on the wrapper + solution
files): rejected for languages whose ecosystem convention is manifest-based — fighting the
toolchain's idiomatic path instead of using it, and forecloses dependency support later.

## Consequences

- ADR-0006's compiled/interpreted split is not the only fork in the runner model: within
  "compiled," a language may additionally be manifest-based (Rust, C#) or bare-file (Go, C). This
  ADR documents that second fork explicitly since it wasn't previously written down.
- The scaffold's wrapper harness (`runtime-wrapper.rs`, `Wrapper.cs`) is rendered by the runner's
  `PrepareSpec` as a sibling file inside `{lang_dir}`, not scaffolded — only the manifest and the
  user's solution file are scaffolded per exercise.
- Adding a future manifest-based language (e.g. Java/Maven, JS/`package.json`) should follow this
  same shape: pinned output name in the manifest, sibling harness rendered at Prepare time.

---
name: creating-runners
description: Use when adding a new language runner to elf (Rust, Ruby, C, Zig, etc.) via the plugin system, when a user asks to "run my solutions in <language>", or when wiring up a [[runner]] descriptor. Covers both the config-driven descriptor AND the Go-side exercise scaffold that adding a language requires.
---

# Creating a Runner

## Overview

A runner lets elf execute solutions in one language. Adding one is a **two-part job**, and the
parts live in different places — this is the thing agents miss:

1. **Execution** is config-driven: a `[[runner]]` Runner Descriptor in `elf.toml` (no recompile).
2. **Exercise scaffolding is NOT config-driven**: `pkg/exercise/downloader_files.go` has a hardcoded
   `switch ex.Language` (only `go`/`py`). A downloaded exercise in a new language has **no solution
   file** until you add a `case` + an embedded template + `jj file track`. This is Go code.

ADR-0006 says "add a language without recompiling." That is true for **execution only**. Scaffolding
was out of that ADR's scope. A runner that runs but can't be scaffolded is half a runner.

**Always grill the user before building.** The baseline failure for this task is an agent that
guesses the toolchain, function signatures, and build strategy instead of asking. Those guesses are
expensive to unwind.

## When to Use

- User: "I want to write solutions in Rust / Ruby / C / Zig."
- Adding or editing a `[[runner]]` block.
- "Why isn't my `rs` exercise downloading a solution file?" (the scaffold half is missing.)

Use the project's Runner vocabulary from `CONTEXT.md`: **Runner**, **Runner Descriptor**, **Runner
Registry**, **Runner Token**, **ExerciseMeta**. Respect ADR-0002 (protocol is a public contract),
ADR-0003 (lifecycle stages), ADR-0006 (config-file descriptors).

## Step 1 — Grill first (do not skip)

Do not create any file until you have the user's answer to each of these. Ask them together, up
front. Guessing any of these is the failure this skill exists to prevent.

| Must extract | Why it changes the build |
|---|---|
| **Interpreted or compiled?** | Compiled → `build_commands` + `open.binary`. Interpreted → `open.interpreter`, no build. |
| **Single-file or project/deps?** (e.g. `rustc` vs `cargo`, needs `regex`?) | Deps → build command runs a project tool (`cargo build`), open spec points at the produced binary. Single-file → direct compiler invocation. **This is the most-guessed fact.** |
| **Toolchain actually installed?** | If absent, you cannot run the end-to-end verification in Step 4. Surface this NOW, not at the end. |
| **Solution function names + signatures** | The wrapper template calls them. e.g. `part_one(&str) -> String`. Don't invent them. |
| **Which Parts?** (PartOne, PartTwo, Visualize) | Determines the dispatch arms in the harness. Visualize takes an output dir. |

If the user says "just get something working" under time pressure: still ask. The grill is ~5
questions and saves the unwind. A wrong toolchain assumption means everything downstream is wrong.

## Step 2 — The Runner Descriptor (`elf.toml`)

A `[[runner]]` block: `key` (Registry key + exercise subdir name, must be unique), `name`,
`[runner.prepare]`, `[runner.open]`. Compiled langs use `build_commands` + `open.binary`;
interpreted langs use `open.interpreter` with no build.

**REQUIRED READ before writing the TOML:** [references/descriptor.md](references/descriptor.md) —
the full field reference, the compiled-vs-interpreted shapes, and the fixed **Runner Token**
vocabulary (`{wrapper_file}`, `{binary_file}`, `{lang_dir}`, …). Do not invent tokens; user values
go in `template_vars`. The shape you pick here is determined by the Step 1 grill answers.

## Step 3 — The wrapper template + protocol harness

The wrapper template renders to `{wrapper_file}` at Prepare time and must speak elf's wire protocol
(ADR-0002). Model it on the reference harnesses `pkg/runners/interface/go.tmpl` and `python.templ` —
read them.

**REQUIRED READ before writing the harness:** [references/protocol.md](references/protocol.md) —
the Task/Result JSON shapes, the part numbers, stdin/stdout framing, and the panic-handling
requirement. A harness that gets the framing wrong hangs the run; do not write it from memory.

The harness includes the user's solution file and dispatches `part` → their function (names and
signatures come from the Step 1 grill — do not invent them).

## Step 4 — The exercise scaffold (Go change — the part agents miss)

So `elf download` produces a starter solution for the new language. In `pkg/exercise/downloader_files.go`:

1. Add a `case "<key>":` to the `switch ex.Language` (alongside `go`/`py`), appending a `tmplFile`.
2. Add the embedded template: `//go:embed templates/<key>.tmpl` + a `var`, and create that file.
3. **`jj file track` the new template** — Nix flakes only see VCS-tracked files (the fileset already
   includes `pkg/exercise/templates`). Without tracking, `nix build` won't see it.
4. Without this step, a downloaded exercise hits the `default` branch → `ErrInvalidLanguage`, or has
   no solution file to compile. A `.example` file the user must copy by hand is a workaround, not done.

## Step 5 — Verify end-to-end (not "list says ok")

`elf runners list` reporting `ok` validates **config/template-path resolution only** — NOT that the
generated code compiles or runs. A runner is not done until:

- A real solution in the new language **solves a downloaded exercise** through `elf solve`, OR
- If the toolchain is absent (Step 1 surfaced this), you have **explicitly told the user** it is
  untested and what they must install — you do not silently claim it works.

Run `mise run lint` + `mise run test` if you touched Go (the scaffold change).

## Completeness contract — a runner is ALL of these

- [ ] Grill answers obtained (Step 1) — toolchain, deps, signatures, parts
- [ ] Runner Descriptor in `elf.toml` (correct compiled/interpreted shape)
- [ ] Wrapper template implementing the protocol (Step 3)
- [ ] Exercise scaffold `case` + embedded template + `jj file track` (Step 4)
- [ ] End-to-end run, OR explicit "untested, install X" to the user (Step 5)

Missing the scaffold (Step 4) or verification (Step 5) is the most common way this is left half-done.

## Red Flags — STOP

- About to create the descriptor without having asked interpreted-vs-compiled or cargo-vs-single-file
- Inventing solution function names instead of asking
- Treating `elf runners list` → `ok` as "it works"
- Shipping a `.example` solution file instead of wiring the scaffold `case`
- Saying "done" when the toolchain was never run

Each of these means: go back to the step you skipped.

## Rationalization Table

| Excuse | Reality |
|---|---|
| "User is in a hurry, skip the questions" | The 5 grill questions are faster than unwinding a wrong toolchain guess. |
| "ADR-0006 says no recompile, so it's config-only" | True for execution; scaffolding (Step 4) is Go code. Half a runner otherwise. |
| "`elf runners list` says ok" | That only checks config resolution, not compilation. Run a real solution. |
| "I'll add a `.example` file so they have a solution" | Workaround. The scaffold `case` is the actual fix. |
| "No toolchain installed, but config looks right" | Then it is UNTESTED. Say so explicitly; don't imply it works. |
| "cargo vs rustc doesn't matter, I'll pick one" | It changes build_commands and the open spec. Ask. |

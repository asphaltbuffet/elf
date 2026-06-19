# Runner Descriptor & Token Reference

A `[[runner]]` block in `elf.toml` fully specifies how to build and launch a Runner for one
language (ADR-0006). **Source of truth:** `pkg/runners/descriptor.go`. Confirm field names there if
anything looks stale.

## Top-level fields

| Field | Meaning |
|-------|---------|
| `key` | Registry key **and** exercise subdirectory name (e.g. `rs`). Must be unique across descriptors. |
| `name` | Display name (e.g. `Rust`). |
| `[runner.prepare]` | PrepareSpec — render wrapper + build (compiled langs). |
| `[runner.open]` | OpenSpec — how to launch the subprocess. |

## `[runner.prepare]` (PrepareSpec)

| Field | Meaning |
|-------|---------|
| `template_path` | Path to the wrapper template file. Empty = no template. |
| `wrapper_ext` | Output extension for the rendered wrapper (e.g. `.rs`). Drives `{wrapper_file}`. |
| `wrapper_subdir` | Subdir within `{lang_dir}` to write the wrapper; affects `{wrapper_file}`/`{binary_file}`. |
| `template_vars` | User-defined variables substituted into the template. Put custom values here — **do not invent new tokens.** |
| `build_commands` | Ordered list of argv arrays; tokens substituted at Prepare time. Compiled langs only. |

## `[runner.open]` (OpenSpec)

| Field | Meaning |
|-------|---------|
| `interpreter` | Interpreter name/path, looked up via `$PATH` (e.g. `python3`, `ruby`). Interpreted langs. |
| `args` | Arguments; tokens substituted at Open time. |
| `env` | Additional `KEY=VALUE` env vars; tokens substituted at Open time. |
| `binary` | Path to the compiled binary; tokens substituted at Open time. Compiled langs. |

## Compiled vs interpreted (the shape the grill determines)

**Interpreted** (Python, Ruby): `[runner.open] interpreter = ...`, no `build_commands`. The
interpreter runs the rendered wrapper directly.

```toml
[runner.open]
interpreter = "ruby"
args = ["{wrapper_file}"]
```

The shipped **Bash** runner is a concrete interpreted example — `wrapper_ext = ".sh"`, no
`build_commands`:

```toml
[[runner]]
key = "bash"
name = "Bash"

[runner.prepare]
template_path = "~/.config/elf/runners/bash.tmpl"
wrapper_ext = ".sh"

[runner.open]
interpreter = "bash"
args = ["{wrapper_file}"]
```

Its wrapper renders to `{lang_dir}/runtime-wrapper.sh` and **sources the sibling solution file**
`{lang_dir}/exercise.sh` (the file the scaffold writes), then dispatches `part` to the user's
`part_one`/`part_two` shell functions — each reads the puzzle input on stdin and echoes the answer.
This sibling-source idiom is the interpreted analogue of Python's `from py import Exercise`; it works
because the wrapper and solution share `{lang_dir}`. The wrapper requires **`jq`** for robust
JSON parse/encode (pure-bash JSON parsing breaks on escaped newlines in puzzle input).

**Compiled** (Rust, Go, C): `[runner.prepare] build_commands = [...]` produces a binary, and
`[runner.open] binary = "{binary_file}"` launches it.

```toml
[runner.prepare]
build_commands = [["rustc", "--edition", "2021", "-O", "-o", "{binary_file}", "{wrapper_file}"]]
[runner.open]
binary = "{binary_file}"
```

## Runner Tokens (fixed vocabulary — do not invent new ones)

Substituted by elf into descriptor fields at runtime. User-defined values go in `template_vars`,
**not** new tokens.

| Token | Resolves to |
|-------|-------------|
| `{exercise_dir}` | The exercise's root directory. |
| `{lang_dir}` | The language subdirectory (named by `key`). |
| `{wrapper_file}` | The rendered wrapper path. **Only valid when `template_path` is set**; extension from `wrapper_ext`. |
| `{binary_file}` | The compiled binary path (compiled runners). |
| `{year}` | Exercise year. |
| `{day}` | Exercise day. |
| `{title}` | Exercise title. |

These map to **ExerciseMeta** (year, day, title, root dir) passed to the RunnerCreator at construction.

## Where it lives

`elf.toml` in the project dir or `~/.config/elf/`. Template files conventionally in
`~/.config/elf/runners/`. `elf runners install` writes the built-in templates and prints the TOML
blocks to paste.

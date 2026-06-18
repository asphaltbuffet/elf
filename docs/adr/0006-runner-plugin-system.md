# Runner plugin system via config-file descriptors

Runners (Go, Python, and any future language) are configured as `[[runner]]` table arrays in
`elf.toml` rather than being compiled into the binary. Each descriptor names a registry key
(which doubles as the exercise subdirectory name), a display name, an optional wrapper template
file path, optional static template variables (substituted using a fixed token vocabulary at
`Prepare` time), optional ordered build commands (for compiled languages), and an open spec
(interpreter + args + env, or a binary path for compiled runners). The two built-in runners are
removed from the binary; users run `elf runners install` to write the built-in template files to
`~/.config/elf/runners/` and receive the TOML blocks to paste into their config. A missing runner
descriptor produces `ErrNoRunner` with a message referencing that command.

## Considered options

**Go plugin (.so)**: ruled out — fragile, requires same Go version, doesn't satisfy "no rebuild."

**Directory scan**: descriptors discovered by scanning a well-known directory. Rejected in favour
of config-file because it introduces a conflict-resolution problem (scan vs config entries for the
same key) and is harder to manage from Nix/home-manager.

**Bundled fallback for built-ins**: Go and Python retain embedded templates as a silent fallback.
Rejected — it papers over broken configs and delays users learning the new model. The hard-fail +
`elf runners install` gives a clear, actionable upgrade path instead.

**Introspection hooks for template variables** (e.g. elf runs `go list -m` automatically): rejected
in favour of static user-supplied `template_vars`. The user configuring a runner knows their module
name; elf should not need language-specific introspection strategies per runner.

## Consequences

- `RunnerCreator` signature changes from `func(dir string) Runner` to
  `func(meta ExerciseMeta) Runner`, where `ExerciseMeta` carries year, day, title, and root path.
  This breaks the existing Go and Python runners, which become the first plugin descriptors.
- The runner registry (`runners.Available`) is populated at startup from config, not at compile time.
- `elf runners` becomes a new top-level command with `install` and `list` subcommands.

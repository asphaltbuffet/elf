# Tech Stack

- Language: Go (see go.mod for version)
- Build: Nix flake + gomod2nix (`buildGoApplication`)
- Tool runner: mise (`.mise.toml`)
- VCS: jujutsu (jj)
- CLI framework: Cobra
- Config: Viper (file: `elf.toml`, env prefix `ELF_`)
- TUI: bubbletea v2 + lipgloss v2 (render package)
- Code generation: `stringer` (enums), `mockery` (mocks)
- Test helpers: gotestsum, golangci-lint
- Dependency lock: `gomod2nix.toml` (regenerate via `mise run nix-hash`)

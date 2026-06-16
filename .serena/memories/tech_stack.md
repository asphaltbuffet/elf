# Tech Stack

- **Language**: Go (see `go.mod` for version)
- **CLI framework**: Cobra
- **Config**: Viper (env prefix `ELF_`, file `elf.toml`)
- **TUI**: Bubbletea v1 + lipgloss v1 + bubbles v1 + huh v0.8.0
- **Filesystem abstraction**: afero
- **Build**: Nix flake with `gomod2nix` (`buildGoApplication`, not `buildGoModule`)
- **Tool manager**: mise (tasks in `mise.toml`)
- **Linter**: golangci-lint (config in `.golangci.yml`)
- **Mocks**: mockery (config in `.mockery.yaml`); generated in `mocks/`
- **Code gen**: stringer (`//go:generate stringer`)
- **Test runner**: gotestsum
- **Snapshot releases**: goreleaser

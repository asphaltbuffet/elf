# Tech Stack

- **Language**: Go (see `go.mod` for version)
- **CLI framework**: Cobra
- **Config**: Viper (`pkg/config/`)
- **TUI**: bubbletea v2 + lipgloss v2 (Live renderer); huh v0.8.0 (download form)
- **Build/task runner**: mise (`mise.toml`)
- **Nix**: flake with `gomod2nix` + `buildGoApplication` (not `buildGoModule`)
- **Linter**: golangci-lint (via mise)
- **Mocks**: mockery (`.mockery.yaml`); migrating toward hand-rolled fakes for internal interfaces
- **Code gen**: stringer for enums (`//go:generate stringer`)
- **Test runner**: gotestsum

Runner templates available: Go, Python, Bash, Rust, Fortran 77 (`pkg/runners/templates.go`).

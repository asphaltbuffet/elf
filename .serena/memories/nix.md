# Nix Build

- `buildGoApplication` (gomod2nix), not `buildGoModule`
- Source filtered via `lib.fileset` — only VCS-tracked files are visible in sandbox
- `gomod2nix.toml` must be in fileset AND referenced via `modules` attribute
- gomod2nix input uses `follows` for nixpkgs and flake-utils

## Stringer files (#149, resolved — ADR-0012)
Three `*_string.go` files are gitignored/untracked (generated code not committed on principle).
Fix: `pkgs.gotools` in `nativeBuildInputs` + `go generate ./...` in `preBuild` hook.
Nix uses nixpkgs-pinned gotools stringer (not mise's `latest`) — accepted divergence.
Affected files: `pkg/exercise/outcome_string.go`, `pkg/tasks/taskstatus_string.go`, `pkg/tasks/tasktype_string.go`.

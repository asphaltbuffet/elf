# Suggested Commands

## Build / Dev
- `mise run dev` — full pipeline: generate, mock, lint, test, snapshot
- `mise run build` — build to dist/
- `mise run generate` — go generate (stringer)
- `mise run mock` — mockery mocks
- `mise run lint` — golangci-lint (read-only)
- `mise run lint-fix` — golangci-lint with auto-fix
- `mise run test` — gotestsum (coverage → bin/coverage.out)
- `mise run snapshot` — goreleaser snapshot
- `mise run nix-hash` — gomod2nix generate
- `mise run mod-tidy` — go mod tidy + nix-hash

## Single test
`go test -run TestFunctionName ./path/to/package`

## Nix
- `nix build` — build elf binary via flake
- `nix develop` — enter dev shell

## VCS (jj)
- `jj file track <path>` — only for new .nix files (flake visibility)
- Never use `git` commands directly

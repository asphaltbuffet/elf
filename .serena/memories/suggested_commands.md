# Suggested Commands

## Development
```
mise run dev         # Full pipeline: generate, mock, lint, test, snapshot
mise run test        # Run tests with coverage (bin/coverage.out)
mise run lint        # Lint read-only (CI)
mise run lint-fix    # Lint with auto-fix (local)
mise run generate    # go generate (stringer)
mise run mock        # Generate mocks (mockery)
mise run build       # Build to dist/
mise run snapshot    # goreleaser snapshot
mise run update-deps # Update direct deps → mod-tidy → nix-hash
```

## Single test
```
go test -run TestFunctionName ./path/to/package
```

## Nix
```
mise run nix-hash    # Regenerate gomod2nix.toml + update vendorHash
nix build            # Verify flake build
```

## VCS (jujutsu)
Use `jj` not `git`. `jj file track` only needed for new `.nix` files (nix flake visibility).

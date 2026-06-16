# Suggested Commands

## mise tasks (always prefer over raw go/nix commands)

```
mise run test        # run tests (gotestsum, coverage → bin/coverage.out)
mise run lint        # golangci-lint read-only (CI)
mise run lint-fix    # golangci-lint with auto-fix (local)
mise run generate    # go generate (stringer)
mise run mock        # generate mocks (mockery)
mise run build       # build → dist/
mise run snapshot    # goreleaser snapshot
mise run update-deps # update direct deps → mod-tidy → nix-hash
mise run mod-tidy    # go mod tidy + auto-runs nix-hash
mise run nix-hash    # gomod2nix generate
mise run dev         # full pipeline: generate, mock, lint, test, snapshot
```

## Single test

```
go test -run TestFunctionName ./path/to/package
```

## mise-managed tools from non-activated shell

```
mise exec -- <tool> <args>   # e.g. mise exec -- gomod2nix generate
```

## VCS

```
jj file track <path>   # track new files (required before nix build sees them)
jj st / jj diff / jj log   # status, diff, log
```

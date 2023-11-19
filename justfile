go_path := `go env GOPATH`

# test and lint
check: test lint

# dev build pipeline
dev: generate snapshot lint test

# CI build pipeline
ci: mod-tidy install generate build test diff

# remove files created during build pipeline
clean:
    rm -rf dist
    rm -rf bin

# go clean + remove build artifacts
[no-exit-message]
nuke: clean
    go clean -i -cache -testcache -modcache -fuzzcache -x

# go mod tidy
[no-exit-message]
mod-tidy:
    go mod tidy

# install tools
install:
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b {{ go_path }}/bin v1.55.2
    go install mvdan.cc/gofumpt@v0.4.0
    go install gotest.tools/gotestsum@v1.11.0

# go generate
[no-exit-message]
generate:
    go generate ./...

# go build
[no-exit-message]
build:
    go build -o dist/ ./...

# goreleaser build snapshot
[no-exit-message]
snapshot:
    goreleaser build --clean --single-target --snapshot

# golangci-lint
lint:
    -golangci-lint run --fix

# go test
[no-exit-message]
test:
    @mkdir -p bin
    gotestsum --format=dots-v2 -- -race -covermode=atomic -coverprofile=bin/coverage.out ./...
    go tool cover -html=bin/coverage.out -o bin/coverage.html
    go tool cover -func=bin/coverage.out

# git diff
_diff:
    git diff --exit-code

diff: _diff
    #!/usr/bin/env bash
    set -euxo pipefail
    RES="$(git status --porcelain)"
    if [ -n "$RES" ]
    then 
        echo $RES && exit 1
    fi

alias gen := generate
alias ss := snapshot
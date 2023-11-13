# dev build - no tools adjustments
dev: gen build-local lint test

# build pipeline
all: mod inst gen build test

# CI build pipeline
ci: all diff

# remove files created during build pipeline
clean:
    -rm -rf dist
    -rm -rf bin
    go clean -i -cache -testcache -modcache -fuzzcache -x

# go mod tidy
mod:
    go mod tidy

# go install tools
inst:
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2
    go install mvdan.cc/gofumpt@v0.4.0

# go generate
gen:
    go generate ./...

# go build
build:
    go build -o dist/ ./...

# goreleaser build
build-local:
    goreleaser build --clean --single-target --snapshot

# golangci-lint
lint:
    -golangci-lint run --fix

# go test
test:
    @mkdir -p bin
    go test -race -covermode=atomic -coverprofile=bin/coverage.out ./...
    go tool cover -html=bin/coverage.out -o bin/coverage.html
    go tool cover -func=bin/coverage.out

# git diff
diff:
    git diff --exit-code
    RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi

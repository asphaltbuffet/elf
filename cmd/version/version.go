package versioncmd

// Version is the elf version, set via -ldflags -X at build time
// (see flake.nix and .goreleaser.yml). It defaults to "dev" for plain
// `go build` invocations that do not inject it.
var Version = "dev"

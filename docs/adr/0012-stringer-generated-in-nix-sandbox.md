# Stringer files generated in-sandbox during nix build

Three enums (`Outcome`, `TaskStatus`, `TaskType`) use `go:generate stringer` to produce `*_string.go` files. These files are gitignored and never committed — generated code does not belong in VCS. Because the nix flake's `lib.fileset` source filter only sees VCS-tracked files, the sandbox has no `String()` methods at compile time and `nix build` fails.

We resolved this by adding `pkgs.gotools` to `nativeBuildInputs` and running `go generate ./...` in a `preBuild` hook, so stringer produces the files inside the sandbox before compilation. The alternatives were: (1) track the generated files in VCS — rejected on principle; (2) hand-write `String()` methods — rejected because they silently drift when enum values are added.

The nix sandbox uses nixpkgs-pinned `gotools` (not mise's `latest` stringer) — acceptable because stringer output for simple `iota` enums is stable across versions.

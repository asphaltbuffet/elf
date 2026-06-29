# Flake version derived from changie data, not git rev

The compiled-in version (`cmd/version.version`, injected via `-X` ldflags) was reporting `dev` for every Nix-built binary. Two independent causes: (1) both `flake.nix` and `.goreleaser.yml` targeted the wrong symbol path — `.../cmd/version/version.version` (an extra `/version` segment) instead of `.../cmd/version.version`; Go's linker silently ignores an `-X` for an unresolved symbol, so injection was a no-op. (2) The flake derived its version from `self.shortRev`, which Nix deliberately does **not** expose for a dirty working copy (a reproducibility guard), so local builds fell through to the `dev` default.

We decided the flake reads the released version from changie's data — the topmost `## [x.y.z]` heading in `CHANGELOG.md` — using **pure Nix** (`builtins.readFile` + `builtins.match`) rather than shelling out to `changie latest`. The Nix build sandbox has no network and no mise tools, so invoking `changie` would mean threading it in as a `nativeBuildInput`; reading the file it already has in `src` returns the identical string with zero added build inputs and is dirty-tree-safe. This is the same principle as [ADR-0012](0012-stringer-generated-in-nix-sandbox.md): prefer a self-contained in-sandbox mechanism over a mise/network tool. Changie remains the single human-curated source of truth — `changie latest` is still used in CI/devShell (where the tool exists) to drive tagging and goreleaser.

The version string distinguishes build provenance. Nix's `self` does not portably expose git *tags* (only a clean `shortRev` or a `dirtyShortRev`), so the flake cannot tell a tagged release from any other clean commit — every flake build therefore carries the rev for traceability: a clean build reports `0.4.1+g<shortRev>`, a dirty tree reports `0.4.1+g<dirtyShortRev>-dirty`, and a source with no rev at all falls back to `0.4.1-dirty`. The bare version (`0.4.1`) is produced only by the tag-driven goreleaser release, not by the flake.

## Considered options

- **`self.rev`/`self.shortRev` as the version** — rejected: absent for dirty working-copy builds (the common local case) and carries no human-meaningful release number.
- **`changie latest` inside the derivation** — rejected: requires adding changie to the sandbox as a build input to read a file already present in `src`.
- **A committed `VERSION` file** — rejected: a third place to update on release, duplicating what changie already owns.

## Consequences

- The `.go` distribution path is the gomod2nix flake (source build), suitable for NUR registration. GoReleaser is intentionally **not** used to generate a Nix expression — that would create a second, divergent (`buildGoModule` + prebuilt-tarball) build definition alongside the flake. NUR registration itself is a follow-up.
- The `-X` symbol path is `github.com/asphaltbuffet/elf/cmd/version.version`. Reintroducing the `/version` typo silently reverts the binary to reporting `dev` with no build error — guard against it when editing ldflags in `flake.nix` or `.goreleaser.yml`.

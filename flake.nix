{
  description = "elf - A CLI helper for programming exercises";

  # Advertise the project's Cachix cache so consumers substitute prebuilt
  # builds (pushed by the release workflow) instead of compiling from source.
  # `extra-` prefixes append to the user's config rather than replacing it.
  nixConfig = {
    extra-substituters = [ "https://asphaltbuffet.cachix.org" ];
    extra-trusted-public-keys = [
      "asphaltbuffet.cachix.org-1:X7blz7HiaFpaq9Om6pYaKHWAaq7jAbjdQdareQQpJmU="
    ];
  };

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
    # NUR provides goreleaser-pro (the unfree pro build used in CI); the free
    # nixpkgs goreleaser would diverge from CI's config parsing/features.
    nur.url = "github:nix-community/NUR";
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix, nur }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default nur.overlays.default ];
          # goreleaser-pro is unfree; allow just that package (used in the devShell
          # to match CI). Scoped predicate rather than a blanket allowUnfree.
          config.allowUnfreePredicate = pkg:
            builtins.elem (pkgs.lib.getName pkg) [ "goreleaser-pro" ];
        };
        lib = pkgs.lib;

        # Released version is the topmost `## [x.y.z]` heading in CHANGELOG.md
        # (changie prepends, so the first match is always the latest release).
        # Read with pure Nix — `changie latest` is not available in the build
        # sandbox; the file already lives in `src`. See ADR-0013.
        # POSIX ERE (builtins.match) matches per-string, not across newlines, and
        # a literal `[` must be written as the bracket expression `[[]`, not `\[`.
        releaseVersion =
          let
            lines = lib.splitString "\n" (builtins.readFile ./CHANGELOG.md);
            matched = builtins.filter (m: m != null)
              (map (builtins.match "## [[]([0-9]+\\.[0-9]+\\.[0-9]+)[]].*") lines);
          in
          if matched == [ ] then "dev" else builtins.head (builtins.head matched);

        # Distinguish build provenance: a clean tagged/release build reports the
        # bare version; a clean commit appends the short rev; a dirty tree marks
        # it `-dirty`; a source with no rev falls back to a bare `-dirty`.
        version =
          if (self ? shortRev) then "${releaseVersion}+g${self.shortRev}"
          else if (self ? dirtyShortRev) then "${releaseVersion}+g${self.dirtyShortRev}-dirty"
          else "${releaseVersion}-dirty";
      in {
        packages.default = pkgs.buildGoApplication {
          pname = "elf";
          inherit version;

          src = lib.fileset.toSource {
            root = ./.;
            fileset = lib.fileset.unions [
              ./go.mod
              ./go.sum
              ./gomod2nix.toml
              (lib.fileset.fileFilter (file: lib.hasSuffix ".go" file.name) ./.)
              ./pkg/exercise/templates
              ./pkg/runners/interface
            ];
          };

          modules = ./gomod2nix.toml;

          # Only build the main binary
          subPackages = [ "." ];

          ldflags = [
            "-s" "-w"
            "-X github.com/asphaltbuffet/elf/cmd/version.Version=${version}"
          ];

          postInstall = ''
            # Generate and install shell completions
            installShellCompletion --cmd elf \
              --bash <($out/bin/elf completion bash) \
              --fish <($out/bin/elf completion fish) \
              --zsh <($out/bin/elf completion zsh)

            # Generate and install man pages
            mkdir -p manpages
            $out/bin/elf man
            mkdir -p $out/share/man/man1
            for f in manpages/*.1; do
              ${pkgs.gzip}/bin/gzip -c "$f" > "$out/share/man/man1/$(basename $f).gz"
            done
          '';

          nativeBuildInputs = [ pkgs.installShellFiles pkgs.gotools ];

          preBuild = ''
            go generate ./...
          '';

          meta = with lib; {
            description = "A CLI helper for programming exercises";
            homepage = "https://github.com/asphaltbuffet/elf";
            license = licenses.mit;
            mainProgram = "elf";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            jujutsu
            jjui
            mise
            ripgrep
            fd
            sd
            gopls
            nixd
            gh
            pkgs.nur.repos.goreleaser.goreleaser-pro
            gomod2nix.packages.${system}.default
          ];

          shellHook = ''
            mise trust --all
          '';
        };
      }
    ) // {
      overlays.default = final: prev: {
        elf = self.packages.${prev.system}.default;
      };

      homeManagerModules.default = import ./nix/home-manager.nix;
    };
}

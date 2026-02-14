{
  description = "elf - A CLI helper for programming exercises";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };
        lib = pkgs.lib;
        version = if (self ? shortRev) then self.shortRev else "dev";
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
            "-X github.com/asphaltbuffet/elf/cmd/version/version.version=${version}"
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

          nativeBuildInputs = [ pkgs.installShellFiles ];

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
            mise
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

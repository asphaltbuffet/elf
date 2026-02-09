{
  description = "elf - A CLI helper for programming exercises";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = if (self ? shortRev) then self.shortRev else "dev";
      in {
        packages.default = pkgs.buildGoModule {
          pname = "elf";
          inherit version;
          src = ./.;

          vendorHash = "sha256-Y0AlaqsvTevLqppXnJKoQBqFG1TFDTwR6H2+gGwyvE8=";

          # Only build the main binary, not the tools/ submodule
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

          meta = with pkgs.lib; {
            description = "A CLI helper for programming exercises";
            homepage = "https://github.com/asphaltbuffet/elf";
            license = licenses.mit;
            mainProgram = "elf";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            mise  # Handles all other dev tools (golangci-lint, gotestsum, mockery, etc.)
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
    };
}

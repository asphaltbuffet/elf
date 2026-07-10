# home-manager module for elf
# Usage: import this module and set programs.elf.enable = true
{
  config,
  lib,
  pkgs,
  ...
}: let
  cfg = config.programs.elf;
  tomlFormat = pkgs.formats.toml {};

  # Build the settings attrset, omitting null/empty values so elf's
  # runtime defaults (XDG dirs, etc.) are not overridden.
  settingsToml =
    {}
    // lib.optionalAttrs (cfg.settings.language != null) {language = cfg.settings.language;}
    // lib.optionalAttrs (cfg.settings.input-file != null) {input-file = cfg.settings.input-file;}
    // lib.optionalAttrs (cfg.settings.task-timeout != null) {task = {timeout = cfg.settings.task-timeout;};}
    // lib.optionalAttrs (cfg.settings.config-dir != null) {config-dir = cfg.settings.config-dir;}
    // lib.optionalAttrs (cfg.settings.cache-dir != null) {cache-dir = cfg.settings.cache-dir;}
    // lib.optionalAttrs (cfg.settings.advent.token != "" || cfg.settings.advent.dir != null) {
      advent =
        {}
        // lib.optionalAttrs (cfg.settings.advent.token != "") {token = cfg.settings.advent.token;}
        // lib.optionalAttrs (cfg.settings.advent.dir != null) {dir = cfg.settings.advent.dir;};
    }
    // lib.optionalAttrs (cfg.settings.euler.dir != null) {
      euler = {dir = cfg.settings.euler.dir;};
    }
    // lib.optionalAttrs (cfg.settings.runners != []) {
      runner = map (r:
        {inherit (r) key name;}
        // lib.optionalAttrs (r.prepare.template_path != null || r.prepare.template_vars != {} || r.prepare.build_commands != []) {
          prepare =
            {}
            // lib.optionalAttrs (r.prepare.template_path != null) {template_path = r.prepare.template_path;}
            // lib.optionalAttrs (r.prepare.template_vars != {}) {template_vars = r.prepare.template_vars;}
            // lib.optionalAttrs (r.prepare.build_commands != []) {build_commands = r.prepare.build_commands;};
        }
        // lib.optionalAttrs (r.open.interpreter != null || r.open.args != [] || r.open.env != [] || r.open.binary != null) {
          open =
            {}
            // lib.optionalAttrs (r.open.interpreter != null) {interpreter = r.open.interpreter;}
            // lib.optionalAttrs (r.open.args != []) {args = r.open.args;}
            // lib.optionalAttrs (r.open.env != []) {env = r.open.env;}
            // lib.optionalAttrs (r.open.binary != null) {binary = r.open.binary;};
        }
      ) cfg.settings.runners;
    };
in {
  options.programs.elf = {
    enable = lib.mkEnableOption "elf, a CLI helper for programming exercises";

    package = lib.mkPackageOption pkgs "elf" {};

    settings = {
      language = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = "go";
        description = "Default implementation language for solutions.";
      };

      input-file = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = "input.txt";
        description = "Default input file name for exercises.";
      };

      task-timeout = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = ''
          Per-task execution timeout. Uses Go duration syntax (e.g. "30s", "2m").
          When null, elf uses its built-in default (2m).
          Set to "0" to disable the timeout entirely.
        '';
      };

      config-dir = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = ''
          Path to the elf configuration directory.
          When null, elf uses the platform's default config directory
          (e.g. ~/.config/elf on Linux).
        '';
      };

      cache-dir = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = ''
          Path to the elf cache directory.
          When null, elf uses the platform's default cache directory
          (e.g. ~/.cache/elf on Linux).
        '';
      };

      advent = {
        token = lib.mkOption {
          type = lib.types.str;
          default = "";
          description = ''
            Advent of Code session token.

            WARNING: This value will be written to the TOML config file in the
            Nix store, which is world-readable. For secret management, prefer
            setting the ELF_ADVENT_TOKEN environment variable via sops-nix,
            agenix, or the dedicated option below.
          '';
        };

        dir = lib.mkOption {
          type = lib.types.nullOr lib.types.str;
          default = "exercises";
          description = "Directory for Advent of Code exercises.";
        };
      };

      euler = {
        dir = lib.mkOption {
          type = lib.types.nullOr lib.types.str;
          default = "euler";
          description = "Directory for Project Euler problems (sibling of the AoC exercise directory).";
        };
      };

      runners = lib.mkOption {
        type = lib.types.listOf (lib.types.submodule {
          options = {
            key = lib.mkOption {
              type = lib.types.str;
              description = "Registry key and exercise subdirectory name (e.g. \"py\").";
            };
            name = lib.mkOption {
              type = lib.types.str;
              description = "Display name (e.g. \"Python\").";
            };
            prepare = lib.mkOption {
              type = lib.types.submodule {
                options = {
                  template_path = lib.mkOption {
                    type = lib.types.nullOr lib.types.str;
                    default = null;
                    description = "Path to wrapper template file. Null means no template.";
                  };
                  template_vars = lib.mkOption {
                    type = lib.types.attrsOf lib.types.str;
                    default = {};
                    description = "Static variables substituted into the template.";
                  };
                  build_commands = lib.mkOption {
                    type = lib.types.listOf (lib.types.listOf lib.types.str);
                    default = [];
                    description = "Ordered build commands. Tokens substituted at Prepare time.";
                  };
                };
              };
              default = {};
              description = "Prepare phase specification.";
            };
            open = lib.mkOption {
              type = lib.types.submodule {
                options = {
                  interpreter = lib.mkOption {
                    type = lib.types.nullOr lib.types.str;
                    default = null;
                    description = "Interpreter binary name looked up via PATH (e.g. \"python3\").";
                  };
                  args = lib.mkOption {
                    type = lib.types.listOf lib.types.str;
                    default = [];
                    description = "Arguments passed to interpreter. Tokens substituted at Open time.";
                  };
                  env = lib.mkOption {
                    type = lib.types.listOf lib.types.str;
                    default = [];
                    description = "Additional env vars in KEY=VALUE form. Tokens substituted at Open time.";
                  };
                  binary = lib.mkOption {
                    type = lib.types.nullOr lib.types.str;
                    default = null;
                    description = "Path to compiled binary. Tokens substituted at Open time. Used instead of interpreter for compiled runners.";
                  };
                };
              };
              default = {};
              description = "Open phase specification.";
            };
          };
        });
        default = [];
        description = "List of runner plugin descriptors. Each entry becomes a [[runner]] block in elf.toml.";
      };
    };

    ELF_ADVENT_TOKEN = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      description = ''
        Set the ELF_ADVENT_TOKEN environment variable.
        This is the recommended way to provide your AoC session token,
        as it avoids writing the token to the Nix store.
        For truly secret handling, use sops-nix or agenix to populate
        this value.
      '';
    };

    ELF_LANGUAGE = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      description = ''
        Set the ELF_LANGUAGE environment variable.
        Overrides the language setting from the config file.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [cfg.package];

    xdg.configFile."elf/elf.toml" = lib.mkIf (settingsToml != {}) {
      source = tomlFormat.generate "elf.toml" settingsToml;
    };

    home.sessionVariables =
      {}
      // lib.optionalAttrs (cfg.ELF_ADVENT_TOKEN != null) {
        ELF_ADVENT_TOKEN = cfg.ELF_ADVENT_TOKEN;
      }
      // lib.optionalAttrs (cfg.ELF_LANGUAGE != null) {
        ELF_LANGUAGE = cfg.ELF_LANGUAGE;
      };
  };
}

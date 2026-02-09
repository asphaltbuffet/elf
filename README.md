# elf - a helper for trivial things

<div align="center">

[![GitHub release (with filter)](https://img.shields.io/github/v/release/asphaltbuffet/elf)](https://github.com/asphaltbuffet/elf/releases)
[![go.mod](https://img.shields.io/github/go-mod/go-version/asphaltbuffet/elf)](go.mod)
[![GitHub License](https://img.shields.io/github/license/asphaltbuffet/elf)](LICENSE)

[![Common Changelog](https://common-changelog.org/badge.svg)](https://common-changelog.org)
[![wakatime](https://wakatime.com/badge/user/09307b0e-8348-4b4e-9b67-0026db3fe1f5/project/74e79fcc-61aa-4b15-835d-4b3b37346599.svg)](https://wakatime.com/@asphaltbuffet/projects/bueionwowo)
[![CodeQL](https://github.com/asphaltbuffet/elf/workflows/CodeQL/badge.svg)](https://app.codecov.io/gh/asphaltbuffet/elf)

</div>

`elf` is a helper app for several programming practice sites that attempts to reduce the overhead needed to test and solve puzzles.

Currently supporting:

- [Advent of Code](https://adventofcode.com/)
- *WIP* [Project Euler](https://projecteuler.net/)
- *WIP* [Exercism](https://exercism.org/)


## Features

- `Download` challenge information with local caching
- `Solve` challenge with multiple language implementations
- `Test` solution with implementation-agnostic test cases
- Show debug output inline with solution output
- Write `visualization` for solutions to disk
- `Benchmark` with graphs to compare implementations

## Requirements

- Language toolchains must be installed separately (Go, Python, etc.)
- Site-specific authorization tokens (see Configuration)

## Install

### From GitHub Releases

Download the latest release from [GitHub Releases](https://github.com/asphaltbuffet/elf/releases) and unpack the binary into your preferred location. Release packages include:

- Pre-built binaries for Linux, MacOS, and Windows
- Shell completions for bash, fish, and zsh
- Man pages

### Using Nix

If you have [Nix](https://nixos.org/) with flakes enabled:

```bash
# Run directly without installing
nix run github:asphaltbuffet/elf

# Install to your profile
nix profile install github:asphaltbuffet/elf
```

```nix
# Add to your flake.nix as an input
{
  inputs.elf.url = "github:asphaltbuffet/elf";
}
```

## Configuration

Configuration can be set via config file or environment variables (with `ELF_` prefix).

### Config File

```bash
elf config init           # Create elf.toml in current directory
elf config init --global  # Create in user config directory
elf config check          # Display and validate current configuration
elf config update-token   # Update Advent of Code session token
```

Files may also be managed directly. Create `elf.toml` in your working directory or the user config directory:

- Windows: `%AppData%\elf\elf.toml`
- Linux: `$XDG_CONFIG_HOME/elf/elf.toml` or `~/.config/elf/elf.toml`
- MacOS: `~/Library/Application Support/elf/elf.toml`

### Environment Variables

Environment variables override config file settings:

- `ELF_ADVENT_TOKEN` — Session token for Advent of Code (required to download inputs)
- `ELF_LANGUAGE` — Default language for solutions (e.g., `go`, `py`)

## Site-specific Details

### Advent of Code

Requires a session token to download puzzle inputs. Get your token from the `session` cookie after logging in at [adventofcode.com](https://adventofcode.com) (browser dev tools → Application → Cookies).

Set the token via:
- `config` subcommand: `elf config update-token`
- Config file: `advent.token = "your-session-token"`
- Environment: `export ELF_ADVENT_TOKEN="your-session-token"`

#### Exercise Directory Structure (Example)

```text
exercises/<year>/<day>-<title>/
├── info.json        # Puzzle metadata (year, day, title, URL)
├── input.txt        # Your puzzle input
├── README.md        # Problem description
├── <lang>/          # <lang> implementation files
│   ├── <file1>
│   ├── <...>
│   └── <file_n>
└── benchmark.json   # Benchmark results (generated)
```

## Caching

Elf caches downloaded data to reduce load on source sites. Cache locations vary by OS:

- Windows: `%LocalAppData%\elf`
- Linux: `$XDG_CACHE_HOME/elf` or `~/.cache/elf`
- MacOS: `~/Library/Caches/elf`

## Development

This project uses [mise](https://mise.jdx.dev) for tool management and task running.

### Quick Start

```bash
# Enter the development shell (if using Nix)
nix develop

# Or install mise and let it manage tools
mise trust
mise install
```

### Common Tasks

```bash
mise run dev         # Full dev pipeline: generate, mock, lint, test, snapshot
mise run test        # Run tests with coverage
mise run lint        # Lint with auto-fix
mise run snapshot    # Build release snapshot for your OS (output in ./dist/)

mise tasks           # List all available tasks
```

### Nix Flake Maintenance

When Go dependencies change (e.g., after `go get` or modifying `go.mod`), the Nix flake's `vendorHash` must be updated:

```bash
mise run nix-hash    # Automatically update vendorHash in flake.nix
nix build            # Verify the build succeeds
```

The `nix-hash` task attempts a build, captures the expected hash from the error output, and updates `flake.nix` automatically.

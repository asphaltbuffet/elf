# elf - a helper for trivial things

<div align="center">

[![GitHub release (with filter)](https://img.shields.io/github/v/release/asphaltbuffet/elf)](https://github.com/asphaltbuffet/elf/releases)
[![go.mod](https://img.shields.io/github/go-mod/go-version/asphaltbuffet/elf)](go.mod)
[![GitHub License](https://img.shields.io/github/license/asphaltbuffet/elf)](LICENSE)
[![Common Changelog](https://common-changelog.org/badge.svg)](https://common-changelog.org)
[![wakatime](https://wakatime.com/badge/user/09307b0e-8348-4b4e-9b67-0026db3fe1f5/project/74e79fcc-61aa-4b15-835d-4b3b37346599.svg)](https://wakatime.com/@asphaltbuffet/projects/bueionwowo)

[![CodeQL](https://github.com/asphaltbuffet/elf/workflows/CodeQL/badge.svg)](https://app.codecov.io/gh/asphaltbuffet/elf)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=asphaltbuffet_elf&metric=coverage)](https://sonarcloud.io/summary/new_code?id=asphaltbuffet_elf)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=asphaltbuffet_elf&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=asphaltbuffet_elf)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=asphaltbuffet_elf&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=asphaltbuffet_elf)

</div>

`elf`` is a helper app for several programming practice sites that attempts to reduce the overhead needed to test and solve puzzles.

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

## Demo

TBD

## Requirements

See site-specific sections for details on unique requirements.

- Toolchain must be maintained separately (compiler installation, etc.)
- Site-specific authorization may need to be set in config file or ENV

## Install

Manually download and unpack the elf binary into your preferred location.

## Configuration

The default location for this data may vary based on OS and personal settings.

- Windows: `%AppData%/elf`
- Unix: `$XDG_CONFIG_HOME/elf` (if non-empty), else `$HOME/.config/elf`
- Darwin: `$HOME/Library/Application Support/elf`

## Site-specific details

### Advent of Code

```text
.
└─ exercises
   └─ <year>
     └─ <day>-<title>
       ├─ go
       │  └─ exercise.go
       ├─ info.json
       ├─ input.txt
       ├─ README.md
       └─ benchmark.json
```

## Caching

Elf caches downloaded information from source sites to reduce load on their servers. The default location for this data may vary based on OS and personal settings.

- Windows: `%AppData%/elf`
- *nix: `$XDG_CONFIG_HOME/elf` (if non-empty), else `$HOME/.config/elf`
- Darwin: `$HOME/Library/Application Support/elf`

## Build

Build scripts use [Task](https://taskfile.dev). There is a [Makefile](./Makefile) and [justfile](./justfile) with similar arguments. These may be removed at some point though.

Install the necessary dependencies with `task install`. This should only be necessary to do once.

`task snapshot` will create a local build for your OS in `./dist/elf-<OS name>/`.

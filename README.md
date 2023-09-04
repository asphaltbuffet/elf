# elf - a helper for trivial things

<div align="center">

![GitHub release)](https://img.shields.io/github/downloads-pre/asphaltbuffet/elf/latest/total?label=latest)
[![go.mod](https://img.shields.io/github/go-mod/go-version/asphaltbuffet/elf)](go.mod)
[![Common Changelog](https://common-changelog.org/badge.svg)](https://common-changelog.org)
[![wakatime](https://wakatime.com/badge/user/09307b0e-8348-4b4e-9b67-0026db3fe1f5/project/74e79fcc-61aa-4b15-835d-4b3b37346599.svg)](https://wakatime.com/badge/user/09307b0e-8348-4b4e-9b67-0026db3fe1f5/project/74e79fcc-61aa-4b15-835d-4b3b37346599)

![CodeQL](https://github.com/asphaltbuffet/elf/workflows/CodeQL/badge.svg)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=asphaltbuffet_elf&metric=coverage)](https://sonarcloud.io/summary/new_code?id=asphaltbuffet_elf)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=asphaltbuffet_elf&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=asphaltbuffet_elf)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=asphaltbuffet_elf&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=asphaltbuffet_elf)

</div>

`elf`` is a helper app for [Advent of Code](https://adventofcode.com) that attempts to reduce the overhead needed to test and solve puzzles.

## Installation

## Features

- Can test/run multiple language solutions (Go and Python... so far)
- Runs test cases
- Supports showing debug prints
- Solution visualization subcommand
- Benchmarking and graphs
- Templates for adding new puzzle days
- Downloads puzzle input (mindful of AoC site load)

## Example

## Requirements

Elf assumes that your solutions are in a specific structure:

```text
.
└── exercises
    └── <year>
        └── <day>-<title>
            ├── <implementation #1>
            │   └── <implementation files>
            ├── info.json
            ├── input.txt
            ├── README.md
            └── benchmark.json


```

## Install

## Use

### Caching

Elf caches downloaded information from the Advent of Code website locally to reduce load on their servers. The default location for this data may vary based on OS and personal settings.

- Windows: `%AppData%/elf`
- Unix: `$XDG_CONFIG_HOME/elf` (if non-empty), else `$HOME/.config/elf`
- Darwin: `$HOME/Library/Application Support/elf`
- Plan 9: `$home/lib/elf`

## Build

Install the necessary dependencies with `make install`. This should only be necessary to do once.

`make build` will create a local build for your OS in `./dist/elf-<OS name>/`.

# Plan: Refactor package structure from domain-based to functionality-based

## Context

The project was originally designed to support multiple programming challenge sites (Advent of Code, Exercism, etc.), but will now focus exclusively on Advent of Code. The `pkg/advent/` package nesting is unnecessary — there will never be a `pkg/exercism/` alongside it. Similarly, `pkg/analysis/` exists as a one-interface package, and `pkg/krampus/` uses a thematic name that obscures its purpose.

**Goal:** Reorganize packages so the structure reflects functionality, not site-specific domains.

## Target structure

```
pkg/
├── exercise/     # renamed from pkg/advent/       (package exercise)
├── analyze/      # moved from pkg/advent/analyze/  (package analyze)
│                 # + merged pkg/analysis/ interface
├── config/       # renamed from pkg/krampus/       (package config)
├── runners/      # unchanged
└── tasks/        # unchanged
internal/
├── common/       # unchanged
└── utilities/    # moved from pkg/utilities/
mocks/
├── analyze/      # was mocks/analysis/
├── config/       # was mocks/krampus/
├── (others unchanged)
```

---

## Complete import change map

| Old import path | New import path | Importers |
|---|---|---|
| `pkg/advent` | `pkg/exercise` | cmd/solve, cmd/test, cmd/benchmark, cmd/download, pkg/advent/analyze (6 .go files + tests) |
| `pkg/advent/analyze` | `pkg/analyze` | cmd/analyze (1 file) |
| `pkg/krampus` | `pkg/config` | cmd/root, cmd/config/*, cmd/solve, cmd/test, cmd/benchmark, cmd/download, cmd/analyze, pkg/advent, pkg/advent/analyze, pkg/advent/benchmarker (16 files) |
| `pkg/analysis` | `pkg/analyze` | cmd/analyze (1 file) |
| `pkg/utilities` | `internal/utilities` | pkg/advent/downloader (1 file) |

**Package name collision:** `cmd/config/` is already `package config`. Files there that import `pkg/config` will use alias `elfcfg`.

---

## Implementation phases

### Phase 1: `pkg/utilities/` → `internal/utilities/`

Smallest change, 1 importer. Builds confidence.

1. `mkdir -p internal/utilities`
2. Move `pkg/utilities/*.go` → `internal/utilities/` (package name stays `utilities`)
3. Update 1 import in `pkg/advent/downloader.go`
4. Delete `pkg/utilities/`
5. `jj file track internal/utilities/`
6. Verify: `go build ./...` + `go test ./internal/utilities/...`

**Files touched:** `internal/utilities/strings.go`, `internal/utilities/strings_test.go`, `pkg/advent/downloader.go`

### Phase 2: Merge `pkg/analysis/` into `pkg/advent/analyze/`

Eliminate the one-interface package before it moves.

1. Add `Analyzer` interface to `pkg/advent/analyze/analyzer.go` (above the struct)
2. Update `cmd/analyze/analyze.go`:
   - Remove `"github.com/asphaltbuffet/elf/pkg/analysis"` import
   - Change `var aa analysis.Analyzer` → `var aa advent.Analyzer` (the `advent` alias already points to `pkg/advent/analyze`)
3. Update `.mockery.yaml`: remove `pkg/analysis` entry, add interface under `pkg/advent/analyze`
4. Delete `pkg/analysis/` and `mocks/analysis/`
5. Regenerate mocks: `mise run mock`
6. Verify: `go build ./...` + `go test ./cmd/analyze/...`

**Files touched:** `pkg/advent/analyze/analyzer.go`, `cmd/analyze/analyze.go`, `.mockery.yaml`

### Phase 3: `pkg/krampus/` → `pkg/config/`

Medium blast radius (16 importers). Done before the big advent rename to reduce variables.

1. `mkdir -p pkg/config`
2. Move `pkg/krampus/*.go` → `pkg/config/`
3. In all moved files: `package krampus` → `package config`
4. Bulk-update imports with `sed` in all .go files **except** `cmd/config/`
5. **Special handling for `cmd/config/`**: manually update to use alias `elfcfg "github.com/asphaltbuffet/elf/pkg/config"`, then replace `krampus.` → `elfcfg.` in those files
6. Update `.mockery.yaml`: `pkg/krampus` → `pkg/config`
7. Delete `pkg/krampus/` and `mocks/krampus/`
8. Regenerate mocks: `mise run mock`
9. `jj file track pkg/config/ mocks/config/`
10. Verify: `go build ./...` + `go test ./...`

**Files with alias (cmd/config/):**
- `cmd/config/init.go` — uses `krampus.DefaultConfigFileBase`, `krampus.DefaultConfigExt`, `krampus.GenerateDefaultConfig`
- `cmd/config/check.go` — uses `krampus.NewConfig`, `krampus.WithFile`, `krampus.MaskToken`, config key constants
- `cmd/config/update_token.go` — uses `krampus.NewConfig`, `krampus.WithFile`
- `cmd/config/config.go` — check if it imports krampus

**All other importers (bulk sed):**
- `cmd/root.go`, `cmd/solve/solve.go`, `cmd/solve/solve_test.go`, `cmd/test/test.go`, `cmd/test/test_test.go`, `cmd/benchmark/benchmark.go`, `cmd/download/download.go`, `cmd/analyze/analyze.go`
- `pkg/advent/advent.go`, `pkg/advent/downloader.go`, `pkg/advent/benchmarker.go`
- `pkg/advent/analyze/analyzer.go`

### Phase 4: `pkg/advent/` → `pkg/exercise/` AND `pkg/advent/analyze/` → `pkg/analyze/`

Biggest change. Both moves happen atomically since `analyze` is a child of `advent`.

1. `mkdir -p pkg/exercise pkg/analyze`
2. Move `pkg/advent/*.go` → `pkg/exercise/`
3. Move `pkg/advent/templates/` → `pkg/exercise/templates/`
4. Move `pkg/advent/testdata/` → `pkg/exercise/testdata/`
5. Move `pkg/advent/analyze/*.go` → `pkg/analyze/`
6. Move `pkg/advent/analyze/testdata/` → `pkg/analyze/testdata/`
7. In `pkg/exercise/*.go`: `package advent` → `package exercise`
8. In `pkg/analyze/*.go`: package name already `analyze` — no change needed
9. Bulk-update imports (order matters — do the longer path first):
   - `pkg/advent/analyze` → `pkg/analyze` (in `cmd/analyze/analyze.go` — currently aliased as `advent`)
   - `pkg/advent` → `pkg/exercise` (in cmd/*, pkg/analyze/*)
10. Update `cmd/analyze/analyze.go` specifically:
    - Change alias from `advent "github.com/asphaltbuffet/elf/pkg/advent/analyze"` to just `"github.com/asphaltbuffet/elf/pkg/analyze"`
    - Update references: `advent.NewAnalyzer` → `analyze.NewAnalyzer`, `advent.WithDirectory` → `analyze.WithDirectory`, `advent.WithOutput` → `analyze.WithOutput`
    - The `Analyzer` interface is now in `pkg/analyze`, so `var aa analyze.Analyzer` works
11. Update `.mockery.yaml`: `pkg/advent/analyze` → `pkg/analyze`
12. Update `flake.nix` line 35: `./pkg/advent/templates` → `./pkg/exercise/templates`
13. Delete `pkg/advent/` entirely
14. `jj file track pkg/exercise/ pkg/analyze/`
15. Regenerate mocks: `mise run mock`
16. Verify: `go build ./...` + `go test ./...`

**Import updates in pkg/analyze/ files** (these import both `pkg/advent` and `pkg/krampus`):
- `analyzer.go`: `pkg/advent` → `pkg/exercise`, `pkg/krampus` → `pkg/config` (already done in Phase 3)
- `graph.go`, `boxplot.go`: `pkg/advent` → `pkg/exercise`
- Test files: `analyzer_test.go`, `graph_test.go`, `boxplot_test.go`, `helpers_test.go`: `pkg/advent` → `pkg/exercise`

**Import updates in cmd/ files:**
- `cmd/solve/solve.go`: `pkg/advent` → `pkg/exercise`, update `advent.New` → `exercise.New`, `advent.WithLanguage` → `exercise.WithLanguage`, etc.
- `cmd/test/test.go`: same pattern
- `cmd/benchmark/benchmark.go`: `pkg/advent` → `pkg/exercise`, update `advent.NewBenchmarker` → `exercise.NewBenchmarker`, etc.
- `cmd/download/download.go`: `pkg/advent` → `pkg/exercise`, update `advent.Overwrites` → `exercise.Overwrites`, `advent.NewDownloader` → `exercise.NewDownloader`, etc.

### Phase 5: Simplify `cmd/download/download.go`

Remove the multi-site URL dispatch switch since only AoC is supported.

1. Remove the `switch` block, directly call `exercise.NewDownloader`
2. Remove unused `strings` import
3. Verify: `go build ./...` + `go test ./cmd/download/...`

### Phase 6: Update documentation

1. Update `CLAUDE.md`: all package references (`pkg/advent` → `pkg/exercise`, `pkg/krampus` → `pkg/config`, file organization table)
2. Update `AGENTS.md`: same package references and table
3. Update `README.md`: remove Exercism mention, update any package references
4. Run `goimports -w .` to ensure clean import formatting

### Phase 7: Full verification

```bash
go build ./...
go test ./... -count=1 -race
mise run lint
mise run test
nix build
```

Check no stale imports remain:
```bash
grep -r "pkg/advent" --include="*.go" .
grep -r "pkg/krampus" --include="*.go" .
grep -r "pkg/analysis" --include="*.go" .
grep -r "pkg/utilities" --include="*.go" .
```

---

## Risk notes

- **go:embed paths are relative** — testdata must move with the test files (handled by moving entire directories)
- **Nix fileset** — `flake.nix` line 35 explicitly references `./pkg/advent/templates`; must update or Nix build breaks
- **jj file tracking** — new files must be tracked before Nix can see them
- **cmd/config collision** — requires import alias `elfcfg`; all other packages can use `config` directly since they don't have a package name conflict
- **Mockery regeneration** — after each `.mockery.yaml` update, regenerate mocks to keep them in sync
- **Rollback** — `jj restore` can undo any phase if something goes wrong

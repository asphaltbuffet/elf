# Runner Plugin System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace compile-time runner registration with config-file-driven Runner Descriptors, enabling any language to be added as a plugin without recompiling elf.

**Architecture:** A new `RunnerDescriptor` type in `pkg/runners/` is loaded from `[[runner]]` TOML table arrays at startup and populates `runners.Available`. A generic `descriptorRunner` implements the `Runner` interface using the descriptor's template path, token substitution, build commands, and open spec. The existing Go and Python runners are removed; their templates become files installed via `elf runners install`.

**Tech Stack:** Go stdlib (`text/template`, `os/exec`, `path/filepath`), Cobra (CLI), Viper (config), afero (filesystem), TOML, Nix (home-manager module).

## Global Constraints

- All Go code must pass `mise run lint` (golangci-lint) with zero warnings — use `exec.CommandContext` not `exec.Command`, no magic numbers (`mnd`), no shadowed `err` variables.
- Tests use `github.com/stretchr/testify` (`assert`, `require`). Table-driven tests with `t.Run`. No mocking of the filesystem — use `afero.NewMemMapFs()` or `t.TempDir()`.
- VCS is **jujutsu** (`jj`). Track new files with `jj file track <path>` instead of `git add`.
- Run full verification with `mise run dev` before each commit (generate → mock → lint → test → snapshot).
- `mise run lint-fix` before manually fixing linter errors.
- New `cmd/` packages must use the factory variable pattern (see CLAUDE.md §Testing Cobra Commands).
- Line-delimited JSON stdin/stdout protocol is a public contract — do not change `pkg/protocol/`.
- Config keys follow the `config.Key` type pattern (string constants in `pkg/config/keys.go`).

---

## File Map

### New files
| File | Responsibility |
|------|----------------|
| `pkg/runners/descriptor.go` | `RunnerDescriptor`, `ExerciseMeta`, `OpenSpec`, token substitution, `RunnerDescriptor.ToCreator()` |
| `pkg/runners/descriptor_runner.go` | `descriptorRunner` struct implementing `Runner` interface |
| `pkg/runners/descriptor_runner_test.go` | Unit tests for `descriptorRunner` lifecycle |
| `pkg/runners/descriptor_test.go` | Unit tests for token substitution |
| `pkg/runners/interface/python.templ` | **Unchanged** — existing file, moved to install target |
| `pkg/runners/interface/go.tmpl` | **Unchanged** — existing file, moved to install target |
| `cmd/runners/runners.go` | `GetRunnersCmd()` parent command |
| `cmd/runners/install.go` | `getInstallCmd()`, `runInstallCmd()` |
| `cmd/runners/install_test.go` | Tests for install command |
| `cmd/runners/list.go` | `getListCmd()`, `runListCmd()` |
| `cmd/runners/list_test.go` | Tests for list command |

### Modified files
| File | Change |
|------|--------|
| `pkg/runners/runners.go` | Change `RunnerCreator` signature; change `Available` initialization; export `RegisterFromDescriptors()` |
| `pkg/runners/golangRunner.go` | **Deleted** — replaced by descriptor |
| `pkg/runners/golangRunner_test.go` | **Deleted** |
| `pkg/runners/pythonRunner.go` | **Deleted** — replaced by descriptor |
| `pkg/runners/pythonRunner_test.go` | **Deleted** |
| `pkg/config/keys.go` | Add `RunnersKey` constant |
| `pkg/config/krampus.go` | Add `GetRunners() []RunnerDescriptor` method |
| `pkg/exercise/advent.go` | Update `RunnerCreator` call sites to pass `ExerciseMeta` |
| `pkg/exercise/benchmarker.go` | Update `RunnerCreator` call site to pass `ExerciseMeta` |
| `pkg/app/app.go` | Update `RunnerCreator` call sites to pass `ExerciseMeta`; accept runners registry as parameter or load from config |
| `pkg/config/writer.go` | Update `GenerateDefaultConfig()` to remove default language hint; note runners must be configured |
| `cmd/root.go` | Add `runners.GetRunnersCmd()` |
| `nix/home-manager.nix` | Add `runners` list option and `settingsToml` mapping |

---

## Task 1: `ExerciseMeta` and new `RunnerCreator` signature

**Files:**
- Modify: `pkg/runners/runners.go`
- Modify: `pkg/exercise/advent.go`
- Modify: `pkg/exercise/benchmarker.go`
- Modify: `pkg/app/app.go`

**Interfaces:**
- Produces: `ExerciseMeta` struct; updated `RunnerCreator` type used by all subsequent tasks.

- [ ] **Step 1: Write failing test for ExerciseMeta construction**

In a new temporary test (add to `pkg/runners/runners_test.go` — create it):

```go
package runners_test

import (
    "testing"
    "github.com/asphaltbuffet/elf/pkg/runners"
    "github.com/stretchr/testify/assert"
)

func TestExerciseMeta_LangDir(t *testing.T) {
    meta := runners.ExerciseMeta{
        Year:  2015,
        Day:   1,
        Title: "not-quite-lisp",
        Dir:   "/home/user/exercises/2015/01-not-quite-lisp",
        Key:   "py",
    }
    assert.Equal(t, "/home/user/exercises/2015/01-not-quite-lisp/py", meta.LangDir())
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/runners/... -run TestExerciseMeta
```

Expected: FAIL — `ExerciseMeta` undefined.

- [ ] **Step 3: Define `ExerciseMeta` in `pkg/runners/runners.go`**

Replace the existing `RunnerCreator` type and `Available` var with:

```go
// ExerciseMeta carries the identity fields of an Exercise to a RunnerCreator.
type ExerciseMeta struct {
    Year  int
    Day   int
    Title string
    Dir   string // exercise root path (e.g. "exercises/2015/01-foo")
    Key   string // language key / subdirectory name (e.g. "py")
}

// LangDir returns the language subdirectory path: Dir/Key.
func (m ExerciseMeta) LangDir() string {
    return filepath.Join(m.Dir, m.Key)
}

// RunnerCreator constructs a Runner for a given exercise.
type RunnerCreator func(meta ExerciseMeta) Runner

// Available is the runtime runner registry, populated at startup from config.
var Available = map[string]RunnerCreator{}
```

Also add `"path/filepath"` import if not present.

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./pkg/runners/... -run TestExerciseMeta
```

Expected: PASS.

- [ ] **Step 5: Update all RunnerCreator call sites**

In `pkg/app/app.go`, change `rc(path)` → `rc(runners.ExerciseMeta{Year: ex.Year, Day: ex.Day, Title: ex.Title, Dir: path, Key: language})` in both `Solve` and `Test` methods.

In `pkg/exercise/benchmarker.go` line ~72, change `implRunner(b.Path)` →:
```go
runner := implRunner(runners.ExerciseMeta{
    Year:  b.Year,
    Day:   b.Day,
    Title: b.Title,
    Dir:   b.Path,
    Key:   impl,
})
```

In `pkg/exercise/advent.go`, any direct `RunnerCreator` calls get the same treatment.

- [ ] **Step 6: Delete the old runner implementations**

```bash
rm pkg/runners/golangRunner.go pkg/runners/golangRunner_test.go
rm pkg/runners/pythonRunner.go pkg/runners/pythonRunner_test.go
```

Track the deletions:
```bash
jj file track pkg/runners/
```

`Available` is now empty — the build will compile but all runner lookups return `ErrNoRunner`. This is intentional; descriptors come in Task 2.

- [ ] **Step 7: Run lint and tests**

```bash
mise run lint-fix
mise run test
```

Expected: All tests pass. Tests that previously tested the Go/Python runner directly are gone. Tests in `pkg/exercise/` and `pkg/app/` that mock runners still pass because they use the `Runner` interface, not the concrete types.

- [ ] **Step 8: Commit**

```bash
jj file track pkg/runners/runners_test.go
jj commit -m "refactor(runners): ExerciseMeta replaces dir string in RunnerCreator"
```

---

## Task 2: `RunnerDescriptor` and token substitution

**Files:**
- Create: `pkg/runners/descriptor.go`
- Create: `pkg/runners/descriptor_test.go`

**Interfaces:**
- Consumes: `ExerciseMeta` (Task 1)
- Produces:
  - `RunnerDescriptor` struct (used by Tasks 3, 4, 7)
  - `RunnerDescriptor.ToCreator() RunnerCreator` (used by Task 3)
  - `substituteTokens(s string, meta ExerciseMeta, wrapperExt string) string` (internal, tested directly)

- [ ] **Step 1: Write failing tests for token substitution**

Create `pkg/runners/descriptor_test.go`:

```go
package runners

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSubstituteTokens(t *testing.T) {
    meta := ExerciseMeta{
        Year:  2015,
        Day:   1,
        Title: "not-quite-lisp",
        Dir:   "/home/user/exercises/2015/01-not-quite-lisp",
        Key:   "py",
    }

    tests := []struct {
        name       string
        input      string
        wrapperExt string
        want       string
    }{
        {
            name:       "exercise_dir token",
            input:      "{exercise_dir}/input.txt",
            wrapperExt: ".py",
            want:       "/home/user/exercises/2015/01-not-quite-lisp/input.txt",
        },
        {
            name:       "lang_dir token",
            input:      "{lang_dir}/wrapper.py",
            wrapperExt: ".py",
            want:       "/home/user/exercises/2015/01-not-quite-lisp/py/wrapper.py",
        },
        {
            name:       "wrapper_file token",
            input:      "{wrapper_file}",
            wrapperExt: ".py",
            want:       "/home/user/exercises/2015/01-not-quite-lisp/py/runtime-wrapper.py",
        },
        {
            name:       "binary_file token",
            input:      "{binary_file}",
            wrapperExt: "",
            want:       "/home/user/exercises/2015/01-not-quite-lisp/py/runtime-wrapper",
        },
        {
            name:       "year day title tokens",
            input:      "{year}/{day}/{title}",
            wrapperExt: "",
            want:       "2015/1/not-quite-lisp",
        },
        {
            name:       "multiple tokens",
            input:      "github.com/me/aoc/{year}/{day}-{title}/go",
            wrapperExt: "",
            want:       "github.com/me/aoc/2015/1-not-quite-lisp/go",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := substituteTokens(tt.input, meta, tt.wrapperExt)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./pkg/runners/... -run TestSubstituteTokens
```

Expected: FAIL — `substituteTokens` undefined.

- [ ] **Step 3: Implement `descriptor.go`**

Create `pkg/runners/descriptor.go`:

```go
package runners

import (
    "fmt"
    "path/filepath"
    "strings"
)

const wrapperBaseName = "runtime-wrapper"

// OpenSpec describes how to launch the runner subprocess.
// Either Interpreter or Binary must be set, not both.
type OpenSpec struct {
    Interpreter string   // path/name of interpreter (e.g. "python3"); looked up via $PATH
    Args        []string // arguments; tokens substituted at Open time
    Env         []string // additional env vars (KEY=VALUE); tokens substituted at Open time
    Binary      string   // path to compiled binary; tokens substituted at Open time
}

// PrepareSpec describes the Prepare phase.
type PrepareSpec struct {
    TemplatePath  string            // path to wrapper template file; empty = no template
    TemplateVars  map[string]string // user-defined variables substituted into the template
    BuildCommands [][]string        // ordered build commands; tokens substituted at Prepare time
}

// RunnerDescriptor is a config entry that fully specifies how to build and launch a Runner.
type RunnerDescriptor struct {
    Key     string      // registry key and exercise subdirectory name (e.g. "py")
    Name    string      // display name (e.g. "Python")
    Prepare PrepareSpec
    Open    OpenSpec
}

// ToCreator returns a RunnerCreator that constructs a descriptorRunner from this descriptor.
func (d RunnerDescriptor) ToCreator() RunnerCreator {
    return func(meta ExerciseMeta) Runner {
        meta.Key = d.Key
        return &descriptorRunner{desc: d, meta: meta}
    }
}

// substituteTokens replaces built-in token placeholders in s with values derived from meta.
// wrapperExt is the file extension for {wrapper_file} (e.g. ".py"); empty string is valid.
func substituteTokens(s string, meta ExerciseMeta, wrapperExt string) string {
    langDir := meta.LangDir()
    wrapperFile := filepath.Join(langDir, wrapperBaseName+wrapperExt)
    binaryFile := filepath.Join(langDir, wrapperBaseName)

    replacer := strings.NewReplacer(
        "{exercise_dir}", meta.Dir,
        "{lang_dir}", langDir,
        "{wrapper_file}", wrapperFile,
        "{binary_file}", binaryFile,
        "{year}", fmt.Sprintf("%d", meta.Year),
        "{day}", fmt.Sprintf("%d", meta.Day),
        "{title}", meta.Title,
    )

    return replacer.Replace(s)
}

// substituteSlice applies substituteTokens to every element of a string slice.
func substituteSlice(ss []string, meta ExerciseMeta, wrapperExt string) []string {
    out := make([]string, len(ss))
    for i, s := range ss {
        out[i] = substituteTokens(s, meta, wrapperExt)
    }
    return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/runners/... -run TestSubstituteTokens
```

Expected: PASS.

- [ ] **Step 5: Write failing test for `ToCreator`**

Add to `pkg/runners/descriptor_test.go`:

```go
func TestRunnerDescriptor_ToCreator(t *testing.T) {
    desc := RunnerDescriptor{
        Key:  "py",
        Name: "Python",
        Prepare: PrepareSpec{
            TemplatePath: "/tmp/python.templ",
        },
        Open: OpenSpec{
            Interpreter: "python3",
            Args:        []string{"-B", "{wrapper_file}"},
        },
    }

    creator := desc.ToCreator()
    assert.NotNil(t, creator)

    meta := ExerciseMeta{
        Year: 2015, Day: 1, Title: "foo",
        Dir: "/exercises/2015/01-foo",
    }
    r := creator(meta)
    assert.NotNil(t, r)
    assert.Equal(t, "Python", r.String())
}
```

- [ ] **Step 6: Run tests to verify they fail**

```bash
go test ./pkg/runners/... -run TestRunnerDescriptor_ToCreator
```

Expected: FAIL — `descriptorRunner` undefined.

- [ ] **Step 7: Create stub `pkg/runners/descriptor_runner.go`**

```go
package runners

// descriptorRunner implements Runner using a RunnerDescriptor.
type descriptorRunner struct {
    desc RunnerDescriptor
    meta ExerciseMeta
}

func (r *descriptorRunner) String() string { return r.desc.Name }
```

Leave lifecycle methods unimplemented for now (they come in Task 4).

- [ ] **Step 8: Run tests to verify they pass**

```bash
go test ./pkg/runners/... -run TestRunnerDescriptor_ToCreator
```

Expected: PASS (interface not yet satisfied, but `String()` test passes; full interface implemented in Task 4).

- [ ] **Step 9: Commit**

```bash
jj file track pkg/runners/descriptor.go pkg/runners/descriptor_test.go pkg/runners/descriptor_runner.go
jj commit -m "feat(runners): RunnerDescriptor type with token substitution"
```

---

## Task 3: Config loading — `[[runner]]` blocks → registry

**Files:**
- Modify: `pkg/config/keys.go`
- Modify: `pkg/config/krampus.go`
- Modify: `pkg/runners/runners.go`
- Create: `pkg/config/runners_test.go` (new test file)

**Interfaces:**
- Consumes: `RunnerDescriptor`, `PrepareSpec`, `OpenSpec` (Task 2)
- Produces:
  - `config.GetRunners() []runners.RunnerDescriptor`
  - `runners.RegisterFromDescriptors(descs []RunnerDescriptor)`

- [ ] **Step 1: Add config key**

In `pkg/config/keys.go`, add:

```go
RunnersKey Key = "runner" // Configuration key for the [[runner]] table array.
```

- [ ] **Step 2: Write failing test for `GetRunners`**

Create `pkg/config/runners_test.go`:

```go
package config

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/spf13/afero"
)

func TestGetRunners_ParsesDescriptors(t *testing.T) {
    tomlContent := `
[[runner]]
key = "py"
name = "Python"

[runner.prepare]
template_path = "/home/user/.config/elf/runners/python.templ"

[runner.open]
interpreter = "python3"
args = ["-B", "{wrapper_file}"]
env = ["PYTHONPATH={lang_dir}/../../../lib:{lang_dir}"]

[[runner]]
key = "go"
name = "Go"

[runner.prepare]
template_path = "/home/user/.config/elf/runners/go.tmpl"
template_vars = { import_path = "github.com/me/aoc/{year}/{day}-{title}/go" }
build_commands = [["go", "mod", "tidy"], ["go", "build", "-o", "{binary_file}", "{wrapper_file}"]]

[runner.open]
binary = "{binary_file}"
`
    fs := afero.NewMemMapFs()
    require.NoError(t, afero.WriteFile(fs, "elf.toml", []byte(tomlContent), 0o644))

    cfg, err := NewConfig(WithFs(fs), WithFile("elf.toml"))
    require.NoError(t, err)

    runners := cfg.GetRunners()
    require.Len(t, runners, 2)

    assert.Equal(t, "py", runners[0].Key)
    assert.Equal(t, "Python", runners[0].Name)
    assert.Equal(t, "/home/user/.config/elf/runners/python.templ", runners[0].Prepare.TemplatePath)
    assert.Equal(t, "python3", runners[0].Open.Interpreter)
    assert.Equal(t, []string{"-B", "{wrapper_file}"}, runners[0].Open.Args)

    assert.Equal(t, "go", runners[1].Key)
    assert.Equal(t, "{binary_file}", runners[1].Open.Binary)
    assert.Equal(t, "github.com/me/aoc/{year}/{day}-{title}/go", runners[1].Prepare.TemplateVars["import_path"])
}

func TestGetRunners_Empty(t *testing.T) {
    cfg, err := NewConfig()
    require.NoError(t, err)
    assert.Empty(t, cfg.GetRunners())
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
go test ./pkg/config/... -run TestGetRunners
```

Expected: FAIL — `GetRunners` undefined.

- [ ] **Step 4: Add `GetRunners` to `pkg/config/krampus.go`**

Add a method that reads the Viper config into `[]runners.RunnerDescriptor`. Viper reads `[[runner]]` table arrays as `[]map[string]any`, so unmarshal via `mapstructure`:

```go
import (
    "github.com/asphaltbuffet/elf/pkg/runners"
    "github.com/mitchellh/mapstructure"
)

// GetRunners returns the list of runner descriptors from config.
func (c Config) GetRunners() []runners.RunnerDescriptor {
    raw := c.viper.Get(string(RunnersKey))
    if raw == nil {
        return nil
    }

    var descs []runners.RunnerDescriptor
    decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
        Result:           &descs,
        WeaklyTypedInput: true,
        TagName:          "mapstructure",
    })
    if err != nil || decoder.Decode(raw) != nil {
        return nil
    }

    return descs
}
```

Add `mapstructure` struct tags to `RunnerDescriptor`, `PrepareSpec`, and `OpenSpec` in `pkg/runners/descriptor.go`:

```go
type RunnerDescriptor struct {
    Key     string      `mapstructure:"key"`
    Name    string      `mapstructure:"name"`
    Prepare PrepareSpec `mapstructure:"prepare"`
    Open    OpenSpec    `mapstructure:"open"`
}

type PrepareSpec struct {
    TemplatePath  string            `mapstructure:"template_path"`
    TemplateVars  map[string]string `mapstructure:"template_vars"`
    BuildCommands [][]string        `mapstructure:"build_commands"`
}

type OpenSpec struct {
    Interpreter string   `mapstructure:"interpreter"`
    Args        []string `mapstructure:"args"`
    Env         []string `mapstructure:"env"`
    Binary      string   `mapstructure:"binary"`
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./pkg/config/... -run TestGetRunners
```

Expected: PASS.

- [ ] **Step 6: Write failing test for `RegisterFromDescriptors`**

Add to `pkg/runners/runners_test.go`:

```go
func TestRegisterFromDescriptors(t *testing.T) {
    // Save and restore Available
    orig := Available
    t.Cleanup(func() { Available = orig })
    Available = map[string]RunnerCreator{}

    descs := []RunnerDescriptor{
        {Key: "py", Name: "Python", Open: OpenSpec{Interpreter: "python3"}},
        {Key: "rb", Name: "Ruby", Open: OpenSpec{Interpreter: "ruby"}},
    }

    RegisterFromDescriptors(descs)

    assert.Contains(t, Available, "py")
    assert.Contains(t, Available, "rb")
    assert.Len(t, Available, 2)
}
```

- [ ] **Step 7: Run test to verify it fails**

```bash
go test ./pkg/runners/... -run TestRegisterFromDescriptors
```

Expected: FAIL.

- [ ] **Step 8: Implement `RegisterFromDescriptors` in `pkg/runners/runners.go`**

```go
// RegisterFromDescriptors populates Available from a slice of RunnerDescriptors.
// Existing entries are overwritten.
func RegisterFromDescriptors(descs []RunnerDescriptor) {
    for _, d := range descs {
        desc := d // capture loop variable
        Available[desc.Key] = desc.ToCreator()
    }
}
```

- [ ] **Step 9: Run test to verify it passes**

```bash
go test ./pkg/runners/... -run TestRegisterFromDescriptors
```

Expected: PASS.

- [ ] **Step 10: Wire registration into app startup**

In `pkg/app/app.go`, import config and call `RegisterFromDescriptors` at the point `App` is constructed (or wherever `NewConfig` result is available). Add a `RegisterRunners(cfg config.Config)` function that `main` (or `cmd/root.go`) calls after config load:

```go
// RegisterRunners populates runners.Available from the given config.
// Must be called before any exercise operation.
func RegisterRunners(cfg elfconfig.Config) {
    runners.RegisterFromDescriptors(cfg.GetRunners())
}
```

In `cmd/root.go`'s `RunE`, after `elfconfig.NewConfig(...)`, add:
```go
app.RegisterRunners(cfg)
```

Do the same in each `cmd/` subcommand that creates a config and runs exercise ops (`solve`, `test`, `benchmark`, `download`, `analyze`). Each one calls `NewConfig` then must call `app.RegisterRunners(cfg)` before using `runners.Available`.

- [ ] **Step 11: Run full suite**

```bash
mise run lint-fix
mise run test
```

Expected: All tests pass. Exercise operations still return `ErrNoRunner` in practice (no descriptors in test configs), but the plumbing is correct.

- [ ] **Step 12: Commit**

```bash
jj file track pkg/config/runners_test.go
jj commit -m "feat(config): GetRunners reads [[runner]] descriptors; RegisterFromDescriptors wires registry"
```

---

## Task 4: `descriptorRunner` — full `Runner` interface implementation

**Files:**
- Modify: `pkg/runners/descriptor_runner.go`
- Create: `pkg/runners/descriptor_runner_test.go`

**Interfaces:**
- Consumes: `RunnerDescriptor`, `ExerciseMeta`, `substituteTokens`, `substituteSlice` (Task 2); `comm.go` helpers `setupBuffers`, `readJSONFromCommand` (existing, unexported — tests use real subprocesses or fakes)
- Produces: complete `Runner` implementation

- [ ] **Step 1: Write failing tests for `Prepare` (template writing)**

Create `pkg/runners/descriptor_runner_test.go`:

```go
package runners

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestDescriptorRunner_Prepare_WritesTemplate(t *testing.T) {
    templateContent := "hello {year} from {title}"
    templateFile := filepath.Join(t.TempDir(), "test.templ")
    require.NoError(t, os.WriteFile(templateFile, []byte(templateContent), 0o600))

    exerciseDir := t.TempDir()

    desc := RunnerDescriptor{
        Key:  "py",
        Name: "Python",
        Prepare: PrepareSpec{
            TemplatePath: templateFile,
        },
        Open: OpenSpec{Interpreter: "python3"},
    }

    meta := ExerciseMeta{
        Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "py",
    }

    runner := &descriptorRunner{desc: desc, meta: meta}
    require.NoError(t, runner.Prepare(context.Background()))

    langDir := filepath.Join(exerciseDir, "py")
    wrapperPath := filepath.Join(langDir, "runtime-wrapper.templ")
    content, err := os.ReadFile(wrapperPath)
    require.NoError(t, err)
    assert.Equal(t, "hello 2015 from foo", string(content))
}

func TestDescriptorRunner_Prepare_NoTemplate(t *testing.T) {
    exerciseDir := t.TempDir()

    desc := RunnerDescriptor{
        Key:  "go",
        Name: "Go",
        Prepare: PrepareSpec{
            BuildCommands: [][]string{{"echo", "build"}},
        },
        Open: OpenSpec{Binary: "{binary_file}"},
    }

    meta := ExerciseMeta{
        Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "go",
    }

    runner := &descriptorRunner{desc: desc, meta: meta}
    // With no template_path and a build command that succeeds (echo), Prepare should not error
    require.NoError(t, runner.Prepare(context.Background()))
}

func TestDescriptorRunner_Prepare_TemplateVarsSubstituted(t *testing.T) {
    templateContent := "import {{ .import_path }}"
    templateFile := filepath.Join(t.TempDir(), "test.tmpl")
    require.NoError(t, os.WriteFile(templateFile, []byte(templateContent), 0o600))

    exerciseDir := t.TempDir()

    desc := RunnerDescriptor{
        Key:  "go",
        Name: "Go",
        Prepare: PrepareSpec{
            TemplatePath: templateFile,
            TemplateVars: map[string]string{
                "import_path": "github.com/me/aoc/{year}/{day}-{title}/go",
            },
        },
        Open: OpenSpec{Binary: "{binary_file}"},
    }

    meta := ExerciseMeta{
        Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "go",
    }

    runner := &descriptorRunner{desc: desc, meta: meta}
    require.NoError(t, runner.Prepare(context.Background()))

    langDir := filepath.Join(exerciseDir, "go")
    wrapperPath := filepath.Join(langDir, "runtime-wrapper.tmpl")
    content, err := os.ReadFile(wrapperPath)
    require.NoError(t, err)
    assert.Equal(t, "import github.com/me/aoc/2015/1-foo/go", string(content))
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./pkg/runners/... -run TestDescriptorRunner_Prepare
```

Expected: FAIL — methods undefined on `descriptorRunner`.

- [ ] **Step 3: Implement `Prepare`, `Cleanup` in `descriptor_runner.go`**

```go
package runners

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "syscall"
    "text/template"

    "github.com/asphaltbuffet/elf/pkg/protocol"
)

type descriptorRunner struct {
    desc        RunnerDescriptor
    meta        ExerciseMeta
    cmd         *exec.Cmd
    stdin       io.WriteCloser
    wrapperFile string // absolute path to written wrapper; empty if no template
    binaryFile  string // absolute path to compiled binary; empty if not compiled
}

func (r *descriptorRunner) String() string { return r.desc.Name }

func (r *descriptorRunner) wrapperExt() string {
    return filepath.Ext(r.desc.Prepare.TemplatePath)
}

func (r *descriptorRunner) Prepare(ctx context.Context) error {
    langDir := r.meta.LangDir()

    if err := os.MkdirAll(langDir, 0o755); err != nil {
        return fmt.Errorf("creating lang dir: %w", err)
    }

    ext := r.wrapperExt()

    if r.desc.Prepare.TemplatePath != "" {
        templateBytes, err := os.ReadFile(r.desc.Prepare.TemplatePath)
        if err != nil {
            return fmt.Errorf("reading template %s: %w", r.desc.Prepare.TemplatePath, err)
        }

        // Substitute built-in tokens into template_vars values first
        vars := make(map[string]string, len(r.desc.Prepare.TemplateVars))
        for k, v := range r.desc.Prepare.TemplateVars {
            vars[k] = substituteTokens(v, r.meta, ext)
        }

        // Render template_vars into the template content using text/template
        tpl, err := template.New("").Parse(string(templateBytes))
        if err != nil {
            return fmt.Errorf("parsing template: %w", err)
        }

        var buf bytes.Buffer
        if err = tpl.Execute(&buf, vars); err != nil {
            return fmt.Errorf("executing template: %w", err)
        }

        r.wrapperFile = filepath.Join(langDir, wrapperBaseName+ext)
        if err = os.WriteFile(r.wrapperFile, buf.Bytes(), 0o600); err != nil {
            return fmt.Errorf("writing wrapper: %w", err)
        }
    }

    r.binaryFile = filepath.Join(langDir, wrapperBaseName)

    for _, cmdArgs := range r.desc.Prepare.BuildCommands {
        if len(cmdArgs) == 0 {
            continue
        }

        substituted := substituteSlice(cmdArgs, r.meta, ext)
        //nolint:gosec // build commands come from user config, not untrusted external input
        cmd := exec.CommandContext(ctx, substituted[0], substituted[1:]...)
        cmd.Dir = langDir

        var stderr bytes.Buffer
        cmd.Stderr = &stderr

        if err := cmd.Run(); err != nil {
            return fmt.Errorf("build command %q failed: %w: %s", substituted[0], err, stderr.String())
        }
    }

    return nil
}

func (r *descriptorRunner) Cleanup() error {
    var wErr, bErr error

    if r.wrapperFile != "" {
        wErr = os.Remove(r.wrapperFile)
        if errors.Is(wErr, os.ErrNotExist) {
            wErr = nil
        }
    }

    if r.binaryFile != "" && r.desc.Open.Binary != "" {
        bErr = os.Remove(r.binaryFile)
        if errors.Is(bErr, os.ErrNotExist) {
            bErr = nil
        }
    }

    return errors.Join(wErr, bErr)
}
```

- [ ] **Step 4: Run tests to verify Prepare tests pass**

```bash
go test ./pkg/runners/... -run TestDescriptorRunner_Prepare
```

Expected: PASS.

- [ ] **Step 5: Write failing tests for Open, Close, Run**

Add to `descriptor_runner_test.go`:

```go
func TestDescriptorRunner_Open_Interpreter(t *testing.T) {
    // Use a real interpreter that just echoes JSON back — "cat" reads stdin and writes stdout
    // We simulate the runner protocol with a shell script
    scriptContent := `#!/bin/sh
while IFS= read -r line; do
  echo "$line"
done
`
    scriptFile := filepath.Join(t.TempDir(), "echo.sh")
    require.NoError(t, os.WriteFile(scriptFile, []byte(scriptContent), 0o700))

    exerciseDir := t.TempDir()
    require.NoError(t, os.MkdirAll(filepath.Join(exerciseDir, "sh"), 0o755))

    desc := RunnerDescriptor{
        Key:  "sh",
        Name: "Shell",
        Open: OpenSpec{
            Interpreter: "sh",
            Args:        []string{scriptFile},
        },
    }

    meta := ExerciseMeta{Dir: exerciseDir, Key: "sh"}
    runner := &descriptorRunner{desc: desc, meta: meta}

    ctx := context.Background()
    require.NoError(t, runner.Open(ctx))
    require.NotNil(t, runner.cmd)
    require.NotNil(t, runner.stdin)

    _ = runner.Close(ctx)
}
```

- [ ] **Step 6: Run test to verify it fails**

```bash
go test ./pkg/runners/... -run TestDescriptorRunner_Open
```

Expected: FAIL.

- [ ] **Step 7: Implement `Open`, `Close`, `Run` in `descriptor_runner.go`**

```go
func (r *descriptorRunner) Open(ctx context.Context) error {
    ext := r.wrapperExt()

    var cmd *exec.Cmd

    if r.desc.Open.Binary != "" {
        binaryPath := substituteTokens(r.desc.Open.Binary, r.meta, ext)
        absPath, err := filepath.Abs(binaryPath)
        if err != nil {
            return fmt.Errorf("resolving binary path: %w", err)
        }
        //nolint:gosec // binary path comes from user config
        cmd = exec.CommandContext(ctx, absPath)
    } else {
        args := substituteSlice(r.desc.Open.Args, r.meta, ext)
        //nolint:gosec // interpreter comes from user config
        cmd = exec.CommandContext(ctx, r.desc.Open.Interpreter, args...)
    }

    cmd.Dir = r.meta.LangDir()
    cmd.Env = append(os.Environ(), substituteSlice(r.desc.Open.Env, r.meta, ext)...)

    var err error
    r.stdin, err = setupBuffers(cmd)
    if err != nil {
        return fmt.Errorf("setting up buffers: %w", err)
    }

    r.cmd = cmd

    return cmd.Start()
}

func (r *descriptorRunner) Close(ctx context.Context) error {
    if r.cmd == nil || r.cmd.Process == nil {
        return nil
    }

    if err := r.cmd.Process.Signal(syscall.SIGTERM); err != nil {
        return fmt.Errorf("failed to send SIGTERM to %s process: %w", r.desc.Name, err)
    }

    done := make(chan error, 1)
    go func() {
        _, waitErr := r.cmd.Process.Wait()
        done <- waitErr
    }()

    select {
    case <-ctx.Done():
        if err := r.cmd.Process.Kill(); err != nil {
            return fmt.Errorf("failed to kill %s process: %w", r.desc.Name, err)
        }
    case err := <-done:
        if err != nil {
            return fmt.Errorf("failed to stop %s process: %w", r.desc.Name, err)
        }
    }

    return nil
}

func (r *descriptorRunner) Run(_ context.Context, task *protocol.Task) (*protocol.Result, error) {
    taskJSON, err := json.Marshal(task)
    if err != nil {
        return nil, fmt.Errorf("marshalling task: %w", err)
    }

    if _, err = r.stdin.Write(append(taskJSON, '\n')); err != nil {
        return nil, fmt.Errorf("writing task to stdin: %w", err)
    }

    result := new(protocol.Result)
    if err = readJSONFromCommand(result, r.cmd); err != nil {
        return nil, err
    }

    return result, nil
}
```

- [ ] **Step 8: Run all runner tests**

```bash
go test ./pkg/runners/... -v
```

Expected: All pass.

- [ ] **Step 9: Run full lint and test suite**

```bash
mise run lint-fix
mise run test
```

Expected: PASS.

- [ ] **Step 10: Commit**

```bash
jj commit -m "feat(runners): descriptorRunner implements full Runner interface"
```

---

## Task 5: Updated `ErrNoRunner` message

**Files:**
- Modify: `pkg/exercise/advent.go`

**Interfaces:**
- Consumes: nothing new
- Produces: updated error message string used by `pkg/app/` and `pkg/exercise/`

- [ ] **Step 1: Write failing test**

Add to `pkg/exercise/advent_test.go` (or the relevant test file):

```go
func TestLoad_NoRunner_ErrorMessage(t *testing.T) {
    // Use a language key with no registered runner
    _, err := Load("/some/path", "xyz", "", afero.NewMemMapFs(), slog.Default())
    require.Error(t, err)
    assert.ErrorIs(t, err, ErrNoRunner)
    assert.Contains(t, err.Error(), "elf runners install")
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/exercise/... -run TestLoad_NoRunner_ErrorMessage
```

Expected: FAIL — error message does not contain "elf runners install".

- [ ] **Step 3: Update error message in `pkg/exercise/advent.go`**

Change the `ErrNoRunner` format string:

```go
// Before:
return nil, fmt.Errorf("%s: %w", language, ErrNoRunner)

// After:
return nil, fmt.Errorf("no runner configured for %q: run 'elf runners install' to install built-in runner templates, then add [[runner]] blocks to your elf.toml: %w", language, ErrNoRunner)
```

Apply the same message pattern to `pkg/exercise/benchmarker.go` and `pkg/app/app.go` where `ErrNoRunner` is returned.

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./pkg/exercise/... -run TestLoad_NoRunner_ErrorMessage
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
jj commit -m "feat(exercise): ErrNoRunner message references 'elf runners install'"
```

---

## Task 6: `elf runners` command — `install` and `list` subcommands

**Files:**
- Create: `cmd/runners/runners.go`
- Create: `cmd/runners/install.go`
- Create: `cmd/runners/install_test.go`
- Create: `cmd/runners/list.go`
- Create: `cmd/runners/list_test.go`
- Modify: `cmd/root.go`

**Interfaces:**
- Consumes: `runners.RunnerDescriptor` (Task 2); embedded template bytes from `pkg/runners/interface/` (existing `//go:embed`); `config.GetRunners()` (Task 3)
- Produces: `GetRunnersCmd() *cobra.Command`

- [ ] **Step 1: Write failing test for `GetRunnersCmd`**

Create `cmd/runners/runners.go` (stub) and `cmd/runners/install_test.go`:

```go
package runners_test  // note: external test package

import (
    "testing"
    "github.com/asphaltbuffet/elf/cmd/runners"
    "github.com/stretchr/testify/assert"
)

func TestGetRunnersCmd(t *testing.T) {
    cmd := runners.GetRunnersCmd()
    assert.NotNil(t, cmd)
    assert.Equal(t, "runners", cmd.Use)
    assert.NotEmpty(t, cmd.Short)
}

func TestGetRunnersCmd_HasInstallSubcommand(t *testing.T) {
    cmd := runners.GetRunnersCmd()
    sub, _, err := cmd.Find([]string{"install"})
    assert.NoError(t, err)
    assert.Equal(t, "install", sub.Use)
}

func TestGetRunnersCmd_HasListSubcommand(t *testing.T) {
    cmd := runners.GetRunnersCmd()
    sub, _, err := cmd.Find([]string{"list"})
    assert.NoError(t, err)
    assert.Equal(t, "list", sub.Use)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./cmd/runners/... -run TestGetRunnersCmd
```

Expected: FAIL — package does not exist.

- [ ] **Step 3: Create `cmd/runners/runners.go`**

```go
package runners

import "github.com/spf13/cobra"

var runnersCmd *cobra.Command

// GetRunnersCmd returns the runners parent command.
func GetRunnersCmd() *cobra.Command {
    if runnersCmd == nil {
        runnersCmd = &cobra.Command{
            Use:               "runners",
            Short:             "Manage elf runner plugins",
            Long:              "Manage runner plugins that execute exercise solutions in different languages.",
            Args:              cobra.NoArgs,
            ValidArgsFunction: cobra.NoFileCompletions,
        }

        runnersCmd.AddCommand(getInstallCmd())
        runnersCmd.AddCommand(getListCmd())
    }

    return runnersCmd
}
```

- [ ] **Step 4: Create `cmd/runners/install.go`**

The install command writes the two built-in template files (re-embedded here) to `~/.config/elf/runners/` and prints the TOML blocks to paste:

```go
package runners

import (
    _ "embed"
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
)

//go:embed ../../pkg/runners/interface/python.templ
var pythonTemplate []byte

//go:embed ../../pkg/runners/interface/go.tmpl
var goTemplate []byte

var installCmd *cobra.Command

func getInstallCmd() *cobra.Command {
    if installCmd == nil {
        installCmd = &cobra.Command{
            Use:               "install",
            Short:             "Install built-in runner template files",
            Long:              "Writes the built-in Go and Python runner wrapper templates to ~/.config/elf/runners/ and prints the [[runner]] config blocks to add to elf.toml.",
            Args:              cobra.NoArgs,
            ValidArgsFunction: cobra.NoFileCompletions,
            RunE:              runInstallCmd,
            Example:           "elf runners install",
        }

        installCmd.Flags().BoolP("force", "f", false, "overwrite existing template files")
    }

    return installCmd
}

func runInstallCmd(cmd *cobra.Command, _ []string) error {
    force, _ := cmd.Flags().GetBool("force")

    configDir, err := os.UserConfigDir()
    if err != nil {
        return fmt.Errorf("getting user config directory: %w", err)
    }

    runnersDir := filepath.Join(configDir, "elf", "runners")
    if err = os.MkdirAll(runnersDir, 0o755); err != nil {
        return fmt.Errorf("creating runners directory: %w", err)
    }

    templates := []struct {
        filename string
        content  []byte
    }{
        {"python.templ", pythonTemplate},
        {"go.tmpl", goTemplate},
    }

    for _, tmpl := range templates {
        dest := filepath.Join(runnersDir, tmpl.filename)

        if _, statErr := os.Stat(dest); statErr == nil && !force {
            cmd.Printf("Skipping %s (already exists; use --force to overwrite)\n", dest)
            continue
        }

        if writeErr := os.WriteFile(dest, tmpl.content, 0o644); writeErr != nil {
            return fmt.Errorf("writing %s: %w", dest, writeErr)
        }

        cmd.Printf("Wrote %s\n", dest)
    }

    cmd.Println()
    cmd.Println("Add the following to your elf.toml:")
    cmd.Println()
    cmd.Printf(`[[runner]]
key = "py"
name = "Python"

[runner.prepare]
template_path = %q

[runner.open]
interpreter = "python3"
args = ["-B", "{wrapper_file}"]
env = ["PYTHONPATH={lang_dir}/../../../lib:{lang_dir}"]

[[runner]]
key = "go"
name = "Go"

[runner.prepare]
template_path = %q
template_vars = { import_path = "YOUR_MODULE/{year}/{day}-{title}/go" }
build_commands = [
  ["go", "mod", "tidy"],
  ["go", "build", "-tags", "runtime", "-o", "{binary_file}", "{wrapper_file}"],
]

[runner.open]
binary = "{binary_file}"
`,
        filepath.Join(runnersDir, "python.templ"),
        filepath.Join(runnersDir, "go.tmpl"),
    )

    cmd.Println("Replace YOUR_MODULE with your Go module name (from go.mod, e.g. github.com/you/advent-of-code).")

    return nil
}
```

- [ ] **Step 5: Create `cmd/runners/list.go`**

```go
package runners

import (
    "fmt"
    "os"

    elfcfg "github.com/asphaltbuffet/elf/pkg/config"
    "github.com/spf13/cobra"
)

var listCmd *cobra.Command

func getListCmd() *cobra.Command {
    if listCmd == nil {
        listCmd = &cobra.Command{
            Use:               "list",
            Short:             "List configured runners",
            Long:              "Lists all runners configured in elf.toml and whether their template files exist on disk.",
            Args:              cobra.NoArgs,
            ValidArgsFunction: cobra.NoFileCompletions,
            RunE:              runListCmd,
            Example:           "elf runners list",
        }

        listCmd.Flags().StringP("config-file", "c", "", "configuration file to read")
    }

    return listCmd
}

func runListCmd(cmd *cobra.Command, _ []string) error {
    cfgFile, _ := cmd.Flags().GetString("config-file")

    cfg, err := elfcfg.NewConfig(elfcfg.WithFile(cfgFile))
    if err != nil {
        return fmt.Errorf("loading configuration: %w", err)
    }

    descs := cfg.GetRunners()

    if len(descs) == 0 {
        cmd.Println("No runners configured.")
        cmd.Println("Run 'elf runners install' to install built-in runner templates.")
        return nil
    }

    cmd.Printf("%-8s  %-16s  %-6s  %s\n", "KEY", "NAME", "STATUS", "TEMPLATE PATH")
    cmd.Println("--------  ----------------  ------  -----------------------------------")

    for _, d := range descs {
        status := "ok"
        templatePath := d.Prepare.TemplatePath

        if templatePath == "" {
            status = "no tmpl"
        } else if _, statErr := os.Stat(templatePath); statErr != nil {
            status = "missing"
        }

        cmd.Printf("%-8s  %-16s  %-6s  %s\n", d.Key, d.Name, status, templatePath)
    }

    return nil
}
```

- [ ] **Step 6: Write tests for `install` and `list` commands**

Add to `cmd/runners/install_test.go`:

```go
package runners_test

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/asphaltbuffet/elf/cmd/runners"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRunInstallCmd_WritesFiles(t *testing.T) {
    tmpDir := t.TempDir()
    t.Setenv("XDG_CONFIG_HOME", tmpDir) // on Linux, UserConfigDir uses this

    // Reset command singleton
    runners.ResetForTest()

    cmd := runners.GetRunnersCmd()
    var out bytes.Buffer
    cmd.SetOut(&out)
    cmd.SetErr(&out)

    err := cmd.Execute()
    // "runners" with no subcommand shows help, not an error
    _ = err

    installCmd, _, findErr := cmd.Find([]string{"install"})
    require.NoError(t, findErr)
    installCmd.SetOut(&out)

    err = installCmd.RunE(installCmd, nil)
    require.NoError(t, err)

    runnersDir := filepath.Join(tmpDir, "elf", "runners")
    assert.FileExists(t, filepath.Join(runnersDir, "python.templ"))
    assert.FileExists(t, filepath.Join(runnersDir, "go.tmpl"))
    assert.Contains(t, out.String(), "[[runner]]")
    assert.Contains(t, out.String(), "YOUR_MODULE")
}
```

Add `ResetForTest()` to `cmd/runners/runners.go` (pattern from existing cmd packages):

```go
// ResetForTest resets command singletons. For use in tests only.
func ResetForTest() {
    runnersCmd = nil
    installCmd = nil
    listCmd = nil
}
```

- [ ] **Step 7: Run tests**

```bash
go test ./cmd/runners/... -v
```

Expected: PASS.

- [ ] **Step 8: Register command in `cmd/root.go`**

In `GetRootCommand()`, add:

```go
import runnerspkg "github.com/asphaltbuffet/elf/cmd/runners"
// ...
rootCmd.AddCommand(runnerspkg.GetRunnersCmd())
```

- [ ] **Step 9: Run full suite**

```bash
mise run lint-fix
mise run test
```

Expected: PASS.

- [ ] **Step 10: Commit**

```bash
jj file track cmd/runners/
jj commit -m "feat(cmd): elf runners install/list subcommands"
```

---

## Task 7: Home-manager module — `[[runner]]` option

**Files:**
- Modify: `nix/home-manager.nix`

**Interfaces:**
- Consumes: agreed `RunnerDescriptor` TOML shape (from grill session)
- Produces: `programs.elf.settings.runners` list option generating `[[runner]]` table array in `elf.toml`

- [ ] **Step 1: Add `runners` option to `options.programs.elf.settings`**

In `nix/home-manager.nix`, inside the `settings = { ... }` block, add after the `advent` block:

```nix
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
```

- [ ] **Step 2: Map runners into `settingsToml`**

In the `settingsToml` let-binding, add:

```nix
// lib.optionalAttrs (cfg.settings.runners != []) {
  runner = map (r:
    { inherit (r) key name; }
    // lib.optionalAttrs (r.prepare.template_path != null || r.prepare.template_vars != {} || r.prepare.build_commands != []) {
      prepare =
        {}
        // lib.optionalAttrs (r.prepare.template_path != null) { template_path = r.prepare.template_path; }
        // lib.optionalAttrs (r.prepare.template_vars != {}) { template_vars = r.prepare.template_vars; }
        // lib.optionalAttrs (r.prepare.build_commands != []) { build_commands = r.prepare.build_commands; };
    }
    // lib.optionalAttrs (r.open.interpreter != null || r.open.args != [] || r.open.env != [] || r.open.binary != null) {
      open =
        {}
        // lib.optionalAttrs (r.open.interpreter != null) { interpreter = r.open.interpreter; }
        // lib.optionalAttrs (r.open.args != []) { args = r.open.args; }
        // lib.optionalAttrs (r.open.env != []) { env = r.open.env; }
        // lib.optionalAttrs (r.open.binary != null) { binary = r.open.binary; };
    }
  ) cfg.settings.runners;
}
```

Note: `pkgs.formats.toml {}` renders Nix lists of attrsets as TOML table arrays (`[[runner]]`). The key `runner` (singular) matches `[[runner]]` in TOML.

- [ ] **Step 3: Verify with nix flake check**

```bash
nix flake check
```

Expected: No errors. The warning `unknown flake output 'homeManagerModules'` is expected and harmless (documented in CLAUDE.md).

- [ ] **Step 4: Commit**

```bash
jj commit -m "feat(nix): home-manager runners option generates [[runner]] config blocks"
```

---

## Task 8: Integration smoke test

**Files:**
- No new files — runs the built binary end-to-end

**Goal:** Verify the full path from `elf runners install` → `elf.toml` config → exercise solve works.

- [ ] **Step 1: Build the binary**

```bash
mise run build
```

Expected: `dist/elf` produced.

- [ ] **Step 2: Run `elf runners install` and verify output**

```bash
./dist/elf runners install
```

Expected: Two files written to `~/.config/elf/runners/` and TOML blocks printed to stdout containing `[[runner]]`, `key = "py"`, `key = "go"`, `YOUR_MODULE`.

- [ ] **Step 3: Run `elf runners list` with no config**

```bash
./dist/elf runners list
```

Expected: "No runners configured. Run 'elf runners install'..."

- [ ] **Step 4: Run `elf runners list` with a config that has bad template path**

Create a temp `elf.toml`:
```toml
[[runner]]
key = "py"
name = "Python"

[runner.prepare]
template_path = "/nonexistent/python.templ"

[runner.open]
interpreter = "python3"
args = ["-B", "{wrapper_file}"]
```

```bash
./dist/elf runners list -c /tmp/test-elf.toml
```

Expected: Table output showing `py`, `Python`, `missing`, `/nonexistent/python.templ`.

- [ ] **Step 5: Verify `ErrNoRunner` message**

```bash
./dist/elf solve --language xyz /some/path 2>&1 || true
```

Expected: Error output contains `"elf runners install"`.

- [ ] **Step 6: Run full dev pipeline**

```bash
mise run dev
```

Expected: All steps pass (generate → mock → lint → test → snapshot).

- [ ] **Step 7: Commit**

```bash
jj commit -m "chore: integration smoke test verified — runner plugin system complete"
```

---

## Self-Review

**Spec coverage check:**

| Requirement | Task |
|---|---|
| `RunnerCreator` signature change to `ExerciseMeta` | Task 1 |
| Remove Go/Python hardcoded runners | Task 1 |
| `RunnerDescriptor` type with all fields | Task 2 |
| Token substitution (all 7 tokens) | Task 2 |
| `[[runner]]` config loading via Viper | Task 3 |
| `RegisterFromDescriptors` populates registry | Task 3 |
| Registration wired into app startup | Task 3 |
| `descriptorRunner` full `Runner` interface | Task 4 |
| `ErrNoRunner` message references install command | Task 5 |
| `elf runners install` writes templates + prints TOML | Task 6 |
| `elf runners list` shows status table | Task 6 |
| Home-manager `runners` option | Task 7 |
| End-to-end smoke test | Task 8 |

**Placeholder scan:** No TBDs, no "similar to Task N" shortcuts, no missing code blocks found.

**Type consistency check:** `RunnerDescriptor`, `PrepareSpec`, `OpenSpec` used consistently across Tasks 2, 3, 4, 6, 7. `ExerciseMeta` defined in Task 1, consumed in Tasks 2, 3, 4. `substituteTokens` / `substituteSlice` defined in Task 2, used in Task 4. `GetRunners()` defined in Task 3, consumed in Task 6.

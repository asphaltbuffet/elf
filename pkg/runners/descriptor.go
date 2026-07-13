package runners

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const wrapperBaseName = "runtime-wrapper"

// OpenSpec describes how to launch the runner subprocess.
// Either Interpreter or Binary must be set, not both.
type OpenSpec struct {
	Interpreter string   `mapstructure:"interpreter"` // path/name of interpreter (e.g. "python3"); looked up via $PATH
	Args        []string `mapstructure:"args"`        // arguments; tokens substituted at Open time
	Env         []string `mapstructure:"env"`         // additional env vars (KEY=VALUE); tokens substituted at Open time
	Binary      string   `mapstructure:"binary"`      // path to compiled binary; tokens substituted at Open time
}

// PrepareSpec describes the Prepare phase.
type PrepareSpec struct {
	TemplatePath  string            `mapstructure:"template_path"`  // path to wrapper template file; empty = no template
	WrapperExt    string            `mapstructure:"wrapper_ext"`    // output file extension for the rendered wrapper (e.g. ".go"); empty = no extension
	WrapperSubdir string            `mapstructure:"wrapper_subdir"` // subdirectory within lang_dir to write the wrapper; affects {wrapper_file} and {binary_file} tokens
	TemplateVars  map[string]string `mapstructure:"template_vars"`  // user-defined variables substituted into the template
	BuildCommands [][]string        `mapstructure:"build_commands"` // ordered build commands; tokens substituted at Prepare time
	CleanupPaths  []string          `mapstructure:"cleanup_paths"`  // dirs/files (token-substitutable, relative to lang_dir) removed by Cleanup; for build-output trees like bin/obj (C#) or target (Rust)
}

// RunnerDescriptor is a config entry that fully specifies how to build and launch a Runner.
type RunnerDescriptor struct {
	Key     string      `mapstructure:"key"`     // registry key and exercise subdirectory name (e.g. "py")
	Name    string      `mapstructure:"name"`    // display name (e.g. "Python")
	Prepare PrepareSpec `mapstructure:"prepare"` // prepare phase spec
	Open    OpenSpec    `mapstructure:"open"`    // open phase spec
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
// wrapperSubdir, when non-empty, places {wrapper_file} and {binary_file} in a subdirectory of lang_dir.
// relExerciseDir is the precomputed {rel_exercise_dir} value (exercise dir relative to the nearest
// go.mod; see ADR-0021). It is resolved by the caller (Prepare) so this function stays pure; callers
// with no need for it pass "".
func substituteTokens(s string, meta ExerciseMeta, wrapperExt, wrapperSubdir, relExerciseDir string) string {
	langDir := meta.LangDir()
	wrapperDir := langDir
	if wrapperSubdir != "" {
		wrapperDir = filepath.Join(langDir, wrapperSubdir)
	}
	wrapperFile := filepath.Join(wrapperDir, wrapperBaseName+wrapperExt)
	binaryFile := filepath.Join(wrapperDir, wrapperBaseName)

	replacer := strings.NewReplacer(
		"{exercise_dir}", meta.Dir,
		"{rel_exercise_dir}", relExerciseDir,
		"{dir_name}", filepath.Base(meta.Dir),
		"{lang_dir}", langDir,
		"{wrapper_file}", wrapperFile,
		"{binary_file}", binaryFile,
		"{year}", strconv.Itoa(meta.Year),
		"{day}", fmt.Sprintf("%02d", meta.Day),
		"{title}", meta.Title,
	)

	return replacer.Replace(s)
}

// substituteSlice applies substituteTokens to every element of a string slice.
func substituteSlice(ss []string, meta ExerciseMeta, wrapperExt, wrapperSubdir, relExerciseDir string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = substituteTokens(s, meta, wrapperExt, wrapperSubdir, relExerciseDir)
	}

	return out
}

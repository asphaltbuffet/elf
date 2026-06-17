package runners

import (
	"path/filepath"
	"strconv"
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
	Key     string // registry key and exercise subdirectory name (e.g. "py")
	Name    string // display name (e.g. "Python")
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
		"{year}", strconv.Itoa(meta.Year),
		"{day}", strconv.Itoa(meta.Day),
		"{title}", meta.Title,
	)

	return replacer.Replace(s)
}

// substituteSlice applies substituteTokens to every element of a string slice.
// Used by descriptorRunner lifecycle methods added in Task 4.
//
//nolint:unused // referenced by Task 4 (descriptor_runner lifecycle methods)
func substituteSlice(ss []string, meta ExerciseMeta, wrapperExt string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = substituteTokens(s, meta, wrapperExt)
	}

	return out
}

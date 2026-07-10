package exercise

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

// Sentinel errors for exercise construction and loading.
var (
	ErrEmptyLanguage     = errors.New("no language specified")
	ErrNotFound          = afero.ErrFileNotFound
	ErrNotImplemented    = errors.New("not implemented")
	ErrNoRunner          = errors.New("no runner available")
	ErrInvalidData       = errors.New("invalid data")
	ErrNoImplementations = errors.New("no implementations found")
	ErrLoadInfo          = errors.New("load info")

	// ErrEulerUnsupported marks an operation not yet implemented for Project
	// Euler Problems. Benchmark and analyze remain two-part-shaped; see
	// ADR-0017.
	ErrEulerUnsupported = errors.New("not yet supported for Project Euler problems")
)

// errNoRunner builds the user-facing error for an unregistered language, wrapping
// ErrNoRunner and pointing at 'elf runners install'.
func errNoRunner(language string) error {
	return fmt.Errorf(
		"no runner configured for %q: run 'elf runners install' to install built-in runner templates, then add [[runner]] blocks to your elf.toml: %w",
		language,
		ErrNoRunner,
	)
}

// WithTaskTimeout sets the per-task execution timeout. A value <=0 disables the timeout.
func WithTaskTimeout(d time.Duration) func(*Exercise) {
	return func(e *Exercise) { e.taskTimeout = d }
}

// Load creates an Exercise from explicit parameters and loads metadata from info.json in fs.
// language must be a key in runners.Available. customInput overrides the default input file
// when non-empty.
func Load(
	exercisePath, language, customInput string,
	fs afero.Fs,
	logger *slog.Logger,
	opts ...func(*Exercise),
) (*Exercise, error) {
	if language == "" {
		return nil, ErrEmptyLanguage
	}

	if _, ok := runners.Available[language]; !ok {
		return nil, errNoRunner(language)
	}

	if exercisePath == "" {
		return nil, fmt.Errorf("instantiate exercise: %w", ErrNotFound)
	}

	e := &Exercise{
		Path:        exercisePath,
		Language:    language,
		customInput: customInput,
	}

	if err := e.loadInfo(fs, logger.With(slog.String("fn", "exercise"))); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}

// KindAt probes dir for an info.json and, if present, returns the Kind it
// declares. The second return value reports whether an info.json was found;
// callers use it to distinguish "not a Problem" from "not an exercise dir at
// all" (e.g. a year directory), both of which should fall through to normal
// handling. Errors reading or parsing info.json are treated the same as
// "not present" — KindAt is a lightweight probe, not a full load.
func KindAt(fs afero.Fs, logger *slog.Logger, dir string) (Kind, bool) {
	probe := &Exercise{Path: dir}

	if err := probe.loadInfo(fs, logger); err != nil {
		return "", false
	}

	return probe.Kind, true
}

func (e *Exercise) loadInfo(fs afero.Fs, logger *slog.Logger) error {
	logger.Debug("populating exercise from info file", "path", e.Path)

	fn := filepath.Join(e.Path, "info.json")

	data, err := afero.ReadFile(fs, path.Clean(fn))
	if err != nil {
		logger.Error("reading info file", tint.Err(err), slog.String("path", fn))
		return fmt.Errorf("%w: %w", ErrLoadInfo, err)
	}

	if err = json.Unmarshal(data, e); err != nil {
		logger.Error("unmarshal json into info struct", tint.Err(err), slog.String("path", fn))
		return fmt.Errorf("%w: %w", ErrLoadInfo, err)
	}

	if e.Kind == "" {
		e.Kind = KindPuzzle
	}

	var incomplete bool
	switch e.Kind {
	case KindProblem:
		incomplete = e.Number == 0 || e.Title == ""
	case KindPuzzle:
		incomplete = e.Day == 0 || e.Year == 0 || e.Title == "" || e.URL == ""
	default:
		incomplete = e.Day == 0 || e.Year == 0 || e.Title == "" || e.URL == ""
	}

	if incomplete {
		logger.Error("incomplete info data", slog.Any("data", e.LogValue()))
		return fmt.Errorf("%w: %w", ErrLoadInfo, ErrInvalidData)
	}

	// A Project Euler Problem may legitimately declare no input file. Don't let a
	// caller's *defaulted* input filename (the CLI always supplies one) override
	// that intentional emptiness — otherwise every Problem solve fails trying to
	// read a nonexistent input.txt. A Problem that DOES declare an input file
	// (Data.InputFileName already set from info.json) still honors the override.
	if e.customInput != "" && (e.Kind != KindProblem || e.Data.InputFileName != "") {
		e.Data.InputFileName = e.customInput
	}

	return nil
}

// Dir returns the base of the exercise directory.
// It will return an empty string if the exercise does not exist.
//
// Example: 01-someExerciseTitle.
func (e *Exercise) Dir() string {
	if e.Path == "" {
		return ""
	}

	return filepath.Base(e.Path)
}

func makeExerciseID(year, day int) string {
	return fmt.Sprintf("%d-%02d", year, day)
}

// exerciseSource carries the already-fetched facts needed to build an Exercise from a downloaded
// puzzle. It exists so newExerciseFromSource has a readable call site rather than a long argument
// list; the Adder fills it after fetching the page and input.
type exerciseSource struct {
	baseDir       string
	language      string
	url           string
	title         string
	input         string
	inputFileName string
	year          int
	day           int
}

// newExerciseFromSource builds a finished Exercise value from already-fetched puzzle data. It is
// pure construction with no I/O — the caller (the Adder) is responsible for fetching the page and
// input first. This is the download counterpart to loadInfo, which builds an Exercise from disk.
func newExerciseFromSource(s exerciseSource) *Exercise {
	return &Exercise{
		ID:       makeExerciseID(s.year, s.day),
		Title:    s.title,
		Language: s.language,
		Year:     s.year,
		Day:      s.day,
		URL:      s.url,
		Path:     makeExercisePath(s.baseDir, s.year, s.day, s.title),
		Data: &Data{
			InputData:     s.input,
			InputFileName: s.inputFileName,
			TestCases: TestCase{
				One: []*Test{{Input: "", Expected: ""}},
				Two: []*Test{{Input: "", Expected: ""}},
			},
			Answers: Answer{One: "", Two: ""},
		},
	}
}

// GetImplementations returns a list of available implementations for the exercise.
func (e *Exercise) GetImplementations(fs afero.Fs) ([]string, error) {
	dirEntries, err := afero.ReadDir(fs, e.Path)
	if err != nil {
		return nil, err
	}

	impls := []string{}

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			continue
		}

		name := strings.ToLower(entry.Name())

		if _, ok := runners.Available[name]; ok {
			impls = append(impls, name)
		}
	}

	if len(impls) == 0 {
		return nil, fmt.Errorf("search %s: %w", e.Path, ErrNoImplementations)
	}

	return impls, nil
}

// readInput reads the exercise input file from fs.
//
// A Project Euler Problem may declare no input file at all (Data.InputFileName
// == ""); in that case there is nothing to read, so this returns an empty
// string with no error rather than trying to read the exercise directory itself.
func (e *Exercise) readInput(fs afero.Fs) (string, error) {
	if e.Data.InputFileName == "" {
		return "", nil
	}

	inputFile := filepath.Join(e.Path, e.Data.InputFileName)

	data, err := afero.ReadFile(fs, inputFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

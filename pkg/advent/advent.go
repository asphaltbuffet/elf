package advent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/krampus"
	"github.com/asphaltbuffet/elf/pkg/runners"
)

var (
	ErrEmptyLanguage     = fmt.Errorf("no language specified")
	ErrNotFound          = afero.ErrFileNotFound
	ErrNotImplemented    = fmt.Errorf("not implemented")
	ErrNoRunner          = fmt.Errorf("no runner available")
	ErrInvalidData       = fmt.Errorf("invalid data")
	ErrNoImplementations = fmt.Errorf("no implementations found")
	ErrLoadInfo          = fmt.Errorf("load info")
)

func New(config krampus.ExerciseConfiguration, options ...func(*Exercise)) (*Exercise, error) {
	e := &Exercise{
		logger: config.GetLogger().With(slog.String("fn", "exercise")),
		writer: os.Stdout,
	}

	for _, option := range options {
		option(e)
	}

	e.appFs = config.GetFs()

	switch {
	case e.Language == "":
		return nil, ErrEmptyLanguage

	case e.Path != "":
		if err := e.loadInfo(); err != nil {
			return nil, err
		}

	case e.URL != "":
		if err := e.loadFromURL(); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("instantiate exercise: %w", ErrNotFound)
	}

	return e, nil
}

func WithDir(dir string) func(*Exercise) {
	return func(e *Exercise) {
		e.Path = dir
	}
}

func WithLanguage(lang string) func(*Exercise) {
	return func(e *Exercise) {
		e.Language = lang
	}
}

func WithInputFile(file string) func(*Exercise) {
	return func(e *Exercise) {
		e.customInput = file
	}
}

func (e *Exercise) loadInfo() error {
	logger := e.logger.With(slog.String("fn", "loadInfo"))
	logger.Debug("populating exercise from info file", "path", e.Path)

	// populate exercise info from info.json
	fn := filepath.Join(e.Path, "info.json")

	data, err := afero.ReadFile(e.appFs, path.Clean(fn))
	if err != nil {
		logger.Error("reading info file", tint.Err(err), slog.String("path", fn))
		return fmt.Errorf("%w: %w", ErrLoadInfo, err)
	}

	err = json.Unmarshal(data, e)
	if err != nil {
		logger.Error("unmarshal json into info struct", tint.Err(err), slog.String("path", fn))
		return fmt.Errorf("%w: %w", ErrLoadInfo, err)
	}

	if e.Day == 0 || e.Year == 0 || e.Title == "" || e.URL == "" {
		logger.Error("incomplete info data", slog.Any("data", e.LogValue()))
		return fmt.Errorf("%w: %w", ErrLoadInfo, ErrInvalidData)
	}

	// replace input file if custom input is set
	if e.customInput != "" {
		e.Data.InputFileName = e.customInput
	}

	// instantiate runner for language
	rc, ok := runners.Available[e.Language]
	if !ok {
		return fmt.Errorf("%s: %w", e.Language, ErrNoRunner)
	}

	e.runner = rc(e.Path)

	return nil
}

func (e *Exercise) loadFromURL() error {
	return ErrNotImplemented
}

// Dir returns the base of the exercise directory.
// It will return an empty string if the exercise does not exist.
//
// Example: 01-someExerciseTitle.
func (e *Exercise) Dir() string {
	if e.Path == "" {
		e.logger.Error("no path available")
		return ""
	}

	return filepath.Base(e.Path)
}

func makeExerciseID(year, day int) string {
	return fmt.Sprintf("%d-%02d", year, day)
}

// GetImplementations returns a list of available implementations for the exercise.
func (e *Exercise) GetImplementations() ([]string, error) {
	dirEntries, err := afero.ReadDir(e.appFs, e.Path)
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

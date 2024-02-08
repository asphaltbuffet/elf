package advent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/utilities"
)

var exerciseBaseDir = "exercises"

func New(options ...func(*Exercise)) (*Exercise, error) {
	e := &Exercise{}

	for _, option := range options {
		option(e)
	}

	switch {
	case e.Language == "":
		return nil, fmt.Errorf("no language specified")

	case e.Path != "":
		if err := e.loadInfo(); err != nil {
			return nil, err
		}

	case e.URL != "":
		if err := e.loadFromURL(); err != nil {
			return nil, err
		}

	default:
		err := fmt.Errorf("no exercise path or URL specified")
		slog.Error("instantiating exercise", tint.Err(err), slog.Any("options", options))
		return nil, err
	}

	return e, nil
}

func WithDir(dir string) func(*Exercise) {
	return func(e *Exercise) {
		e.Path = dir
	}
}

func WithURL(url string) func(*Exercise) {
	return func(e *Exercise) {
		e.URL = url
	}
}

func WithLanguage(lang string) func(*Exercise) {
	return func(e *Exercise) {
		e.Language = lang
	}
}

func (e *Exercise) loadInfo() error {
	slog.Debug("populating exercise from info file", "path", e.Path)

	// populate exercise info from info.json
	fn := filepath.Join(e.Path, "info.json")

	data, err := os.ReadFile(path.Clean(fn))
	if err != nil {
		slog.Error("reading info file", tint.Err(err), slog.String("path", fn))
		return fmt.Errorf("read info file %q: %w", fn, err)
	}

	err = json.Unmarshal(data, e)
	if err != nil {
		slog.Error("unmarshal json into info struct", tint.Err(err), slog.String("path", fn))
		return fmt.Errorf("unmarshal info file %s: %w", fn, err)
	}

	if e.Day == 0 || e.Year == 0 || e.Title == "" || e.URL == "" {
		slog.Error("incomplete info data", slog.Any("data", e.LogValue()))
		return fmt.Errorf("loading data: %s", fn)
	}

	// instantiate runner for language
	rc, ok := runners.Available[e.Language]
	if !ok {
		return fmt.Errorf("no runner available for language %q", e.Language)
	}

	e.runner = rc(e.Path)

	return nil
}

func (e *Exercise) loadFromURL() error {
	return fmt.Errorf("loading exercise directly from URL not implemented")
}

// Dir returns the path to the exercise directory.
// It will return an empty string if the exercise does not exist.
//
// Example: 01-someExerciseTitle.
func (e *Exercise) Dir() string {
	if e == nil || e.Path == "" {
		slog.Error("nil exercise or no path available")
		return ""
	}

	return filepath.Base(e.Path)
}

func makeExerciseID(year, day int) string {
	return fmt.Sprintf("%d-%02d", year, day)
}

func makeExercisePath(year, day int, title string) string {
	return filepath.Join(exerciseBaseDir, strconv.Itoa(year), fmt.Sprintf("%02d-%s", day, utilities.ToCamel(title)))
}

// GetImplementations returns a list of available implementations for the exercise.
func (e *Exercise) GetImplementations() ([]string, error) {
	dirEntries, err := os.ReadDir(e.Path)
	if err != nil {
		return nil, fmt.Errorf("checking %s: %w", e.Path, err)
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

	return impls, nil
}

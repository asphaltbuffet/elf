package advent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/utilities"
)

var exerciseBaseDir = "exercises"

func New(lang string, options ...func(*Exercise)) (*Exercise, error) {
	e := &Exercise{Language: lang}

	for _, option := range options {
		option(e)
	}

	switch {
	case e.Language == "":
		slog.Error("no language specified")
		return nil, fmt.Errorf("no language specified")

	case e.path != "":
		if err := e.loadInfo(); err != nil {
			slog.Error("filling exercise from info file", tint.Err(err))
			return nil, err
		}

	case e.URL != "":
		if err := e.loadFromURL(); err != nil {
			slog.Error("filling exercise from URL", tint.Err(err))
			return nil, err
		}

	default:
		slog.Error("no exercise path or URL specified")
		return nil, fmt.Errorf("no exercise path or URL specified")
	}

	return e, nil
}

func WithDir(dir string) func(*Exercise) {
	return func(e *Exercise) {
		e.path = dir
	}
}

func WithURL(url string) func(*Exercise) {
	return func(e *Exercise) {
		e.URL = url
	}
}

func (e *Exercise) loadInfo() error {
	slog.Debug("populating exercise from info file", "path", e.path)

	// populate exercise info from info.json
	fn := filepath.Join(e.path, "info.json")

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
		slog.Error("incomplete info data", slog.Any("data", e))
		return fmt.Errorf("incomplete info data: %v", e)
	}

	// instatiate runner for language
	rc, ok := runners.Available[e.Language]
	if !ok {
		return fmt.Errorf("no runner available for language %q", e.Language)
	}

	e.runner = rc(e.path)

	return nil
}

func (e *Exercise) loadFromURL() error {
	return fmt.Errorf("not implemented")
}

// Dir returns the path to the exercise directory.
// It will return an empty string if the exercise does not exist.
//
// Example: exercises/2020/01-someExerciseTitle.
func (e *Exercise) Dir() string {
	if e == nil || e.path == "" {
		slog.Error("nil exercise or no path available")
		return ""
	}

	return filepath.Join(exerciseBaseDir, strconv.Itoa(e.Year), filepath.Base(e.path))
}

func makeExerciseID(year, day int) string {
	return fmt.Sprintf("%d-%02d", year, day)
}

func makeExercisePath(year, day int, title string) string {
	return filepath.Join(exerciseBaseDir, strconv.Itoa(year), fmt.Sprintf("%02d-%s", day, utilities.ToCamel(title)))
}

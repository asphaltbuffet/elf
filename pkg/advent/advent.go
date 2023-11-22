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

	"github.com/asphaltbuffet/elf/pkg/utilities"
)

func NewFromDir(dir, lang string) (*Exercise, error) {
	logger := slog.With(slog.String("src", "NewFromDir"))
	logger.Debug("creating new advent exercise", "dir", dir, "language", lang)

	e, err := NewExerciseFromInfo(dir)
	if err != nil {
		return nil, err
	}

	e.Language = lang
	e.path = dir

	slog.Debug("new advent exercise", slog.Any("exercise", e))

	return e, nil
}

func (e *Exercise) SetLanguage(lang string) {
	e.Language = lang
}

// Dir returns the path to the exercise directory.
// It will return an empty string if the exercise does not exist.
//
// Example: exercises/2020/01-someExerciseTitle.
func (e *Exercise) Dir() string {
	if e == nil || e.path == "" {
		slog.Error("nil exercise or no path available")
		panic("no exercise path available")
	}

	return filepath.Join("exercises", strconv.Itoa(e.Year), filepath.Base(e.path))
}

func NewExerciseFromInfo(dir string) (*Exercise, error) {
	fn := filepath.Join(dir, "info.json")

	data, err := os.ReadFile(path.Clean(fn))
	if err != nil {
		slog.Error("reading info file", tint.Err(err), slog.String("path", fn))
		return nil, fmt.Errorf("read info file %q: %w", fn, err)
	}

	d := &Exercise{}

	err = json.Unmarshal(data, d)
	if err != nil {
		slog.Error("unmarshal json into info struct", tint.Err(err), slog.String("path", fn))
		return nil, fmt.Errorf("unmarshal info file %s: %w", fn, err)
	}

	if d.Day == 0 || d.Year == 0 || d.Title == "" || d.URL == "" {
		slog.Error("incomplete info data", slog.Any("data", d))
		return nil, fmt.Errorf("incomplete info data: %v", d)
	}

	return d, nil
}

func makeExerciseID(year, day int) string {
	return fmt.Sprintf("%d-%02d", year, day)
}

func makeExercisePath(year, day int, title string) string {
	return filepath.Join("exercises", strconv.Itoa(year), fmt.Sprintf("%02d-%s", day, utilities.ToCamel(title)))
}

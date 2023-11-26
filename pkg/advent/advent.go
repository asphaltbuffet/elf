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

var exerciseBaseDir = "exercises"

func New(options ...func(*Exercise)) *Exercise {
	e := &Exercise{}

	for _, option := range options {
		option(e)
	}

	return e
}

func WithDir(dir string) func(*Exercise) {
	return func(e *Exercise) {
		e.path = dir
	}
}

func WithLanguage(lang string) func(*Exercise) {
	return func(e *Exercise) {
		e.Language = lang
	}
}

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

func loadExisting(path string) (*Exercise, error) {
	var (
		err error
		e   *Exercise
	)

	infoPath := filepath.Join(path, "info.json")

	_, err = appFs.Stat(infoPath)
	if err == nil {
		// exercise exists, we may need to update it
		logger.Info("update existing exercise", slog.String("dir", path))

		// TODO: a bad info.json will cause this to behave unpredictably
		// TODO: if this fails, try to create a new exercise, or tell user to delete file(s)
		e, err = NewExerciseFromInfo(path)
		if err != nil {
			logger.Error("creating exercise from info", slog.String("dir", path), tint.Err(err))
			return nil, fmt.Errorf("loading exercise from info: %w", err)
		}
	}

	return e, nil
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

	return filepath.Join(exerciseBaseDir, strconv.Itoa(e.Year), filepath.Base(e.path))
}

func makeExerciseID(year, day int) string {
	return fmt.Sprintf("%d-%02d", year, day)
}

func makeExercisePath(year, day int, title string) string {
	return filepath.Join(exerciseBaseDir, strconv.Itoa(year), fmt.Sprintf("%02d-%s", day, utilities.ToCamel(title)))
}

package advent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/runners"
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

	slog.Debug("created advent exercise",
		"id", e.ID,
		"language", e.Language,
		"year", e.Year,
		"day", e.Day,
		"url", e.URL,
		"path", e.path)

	return e, nil
}

func (e *Exercise) SetLanguage(lang string) {
	e.Language = lang
}

func (e *Exercise) Solve() error {
	logger := slog.With(slog.String("fn", "Solve"), slog.String("exercise", e.Title))
	logger.Debug("solving", slog.String("language", e.Language))

	input := e.Data.Input

	runner := e.runner

	if err := runner.Start(); err != nil {
		logger.Error("starting runner", slog.String("path", e.Data.InputFile), tint.Err(err))
		return err
	}

	defer func() {
		_ = runner.Stop()
		_ = runner.Cleanup()
	}()

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		Foreground(lipgloss.Color("5"))

	fmt.Println(headerStyle.Render(e.String()))

	if err := runTests(runner, e.Data); err != nil {
		logger.Error("running tests", tint.Err(err))
		return err
	}

	if err := runMainTasks(runner, string(input)); err != nil {
		logger.Error("running main tasks", tint.Err(err))
		return err
	}

	return nil
}

func (e *Exercise) Test() error {
	logger := slog.With(slog.String("fn", "Solve"), slog.String("exercise", e.Title))
	logger.Debug("solving", slog.String("language", e.Language))

	runner := runners.Available[e.Language](e.path)

	if err := runner.Start(); err != nil {
		logger.Error("starting runner", slog.String("path", e.Data.InputFile), tint.Err(err))
		return err
	}

	defer func() {
		_ = runner.Stop()
		_ = runner.Cleanup()
	}()

	headerStyle := lipgloss.NewStyle().Bold(true).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("5"))

	fmt.Println(headerStyle.Render(e.String()))

	if err := runTests(runner, e.Data); err != nil {
		logger.Error("running tests", tint.Err(err))
		return err
	}

	return nil
}

// String returns a string representation of the exercise in the format:
// `Advent of Code YYYY, Day DD: TITLE (LANGUAGE)`.
//
// Example: Advent of Code: 2020-01 (Go).
func (e *Exercise) String() string {
	if e == nil {
		slog.Error("nil exercise")
		return "Advent of Code: INVALID EXERCISE"
	}

	name, ok := runners.RunnerNames[e.Language]
	if !ok {
		slog.Warn("unknown language", slog.String("language", e.Language))

		name = "INVALID LANGUAGE"
	}

	return fmt.Sprintf("Advent of Code %d, Day %d: %s (%s)", e.Year, e.Day, e.Title, name)
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

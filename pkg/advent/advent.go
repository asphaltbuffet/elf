package advent

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/lipgloss"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

var baseDir = "exercises"

type Exercise struct {
	ID       string
	Language string
	Year     int
	Day      int
}

// Data contains the relative path to exercise input and the specific test case data for an exercise.
type Data struct {
	InputFile string   `json:"inputFile"`
	TestCases TestCase `json:"testCases"`
}

// TestCase contains the test case for each part of an exercise.
type TestCase struct {
	One []*Test `json:"one"`
	Two []*Test `json:"two"`
}

// Test contains the input and expected output for a test case.
type Test struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

func New(id, lang string) (*Exercise, error) {
	var y, d int

	if n, err := fmt.Sscanf(id, "%d-%d", &y, &d); err != nil || n != 2 {
		return nil, fmt.Errorf("invalid exercise ID: %s", id)
	}

	// allow shorthand for years; we'll validate it's in range later
	if y < 1000 {
		y += 2000
	}

	return &Exercise{
		ID:       fmt.Sprintf("%d-%02d", y, d),
		Language: lang,
		Year:     y,
		Day:      d,
	}, nil
}

func (e *Exercise) SetLanguage(lang string) {
	e.Language = lang
}

func (e *Exercise) Solve() error {
	data, err := loadData(e.Dir())
	if err != nil {
		return err
	}

	input, err := os.ReadFile(filepath.Join(e.Dir(), data.InputFile))
	if err != nil {
		return err
	}

	runner := runners.Available[e.Language](e.Dir())

	if err = runner.Start(); err != nil {
		return err
	}

	defer func() {
		_ = runner.Stop()
		_ = runner.Cleanup()
	}()

	headerStyle := lipgloss.NewStyle().Bold(true).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("5"))

	fmt.Println(headerStyle.Render(e.String()))

	if err = runTests(runner, data); err != nil {
		return err
	}

	if err := runMainTasks(runner, string(input)); err != nil {
		return err
	}

	return nil
}

func (e *Exercise) Test() error {
	data, err := loadData(e.Dir())
	if err != nil {
		return err
	}

	runner := runners.Available[e.Language](e.Dir())

	if err = runner.Start(); err != nil {
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

	if err = runTests(runner, data); err != nil {
		return err
	}

	return nil
}

// String returns a string representation of the exercise in the format:
// `Advent of Code: YYYY-DD (LANGUAGE)`.
//
// Example: Advent of Code: 2020-01 (Go).
func (e *Exercise) String() string {
	if e == nil {
		return "Advent of Code: INVALID EXERCISE"
	}

	name, ok := runners.RunnerNames[e.Language]
	if !ok {
		name = "INVALID LANGUAGE"
	}

	return fmt.Sprintf("Advent of Code: %04d-%02d (%s)", e.Year, e.Day, name)
}

// Dir returns the path to the exercise directory.
// It will return an empty string if the exercise does not exist.
//
// Example: exercises/2020/01-someExerciseTitle.
func (e *Exercise) Dir() string {
	entries, _ := os.ReadDir(filepath.Join(baseDir, fmt.Sprintf("%d", e.Year)))

	for _, entry := range entries {
		if entry.IsDir() && entry.Name()[:2] == fmt.Sprintf("%02d", e.Day) {
			return filepath.Join(baseDir, fmt.Sprintf("%d", e.Year), entry.Name())
		}
	}

	return ""
}

func loadData(p string) (*Data, error) {
	fn := filepath.Join(p, "info.json")

	data, err := os.ReadFile(path.Clean(fn))
	if err != nil {
		return nil, fmt.Errorf("read data file %q: %w", fn, err)
	}

	d := &Data{}

	err = json.Unmarshal(data, d)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data file %s: %w", fn, err)
	}

	return d, nil
}

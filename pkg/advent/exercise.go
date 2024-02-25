package advent

import (
	"fmt"
	"log/slog"

	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

type Exercise struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Language string `json:"-"`
	Year     int    `json:"year"`
	Day      int    `json:"day"`
	URL      string `json:"url"`
	Data     *Data  `json:"data"`
	Path     string `json:"-"`

	runner runners.Runner `json:"-"`
	appFs  afero.Fs       `json:"-"`
	logger *slog.Logger   `json:"-"`
}

// Data contains the relative path to exercise input and the specific test case data for an exercise.
type Data struct {
	InputData     string   `json:"-"`
	InputFileName string   `json:"inputFile"`
	TestCases     TestCase `json:"testCases"`
	Answers       Answer   `json:"answers,omitempty"`
}

// TestCase contains the test case for each part of an exercise.
type TestCase struct {
	One []*Test `json:"one"`
	Two []*Test `json:"two"`
}

// Answer contains the answer for each part of an exercise.
type Answer struct {
	One string `json:"a"`
	Two string `json:"b"`
}

// Test contains the input and expected output for a test case.
type Test struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

func (e *Exercise) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", e.ID),
		slog.String("dir", e.Dir()),
		slog.String("lang", e.Language),
	)
}

func (e *Exercise) String() string {
	if *e == *(&Exercise{}) { //nolint:staticcheck // this is needed for the comparison
		return "INVALID EXERCISE"
	}

	if e.runner == nil {
		return fmt.Sprintf("Advent of Code %d, Day %d: %s (?)", e.Year, e.Day, e.Title)
	}

	return fmt.Sprintf("Advent of Code %d, Day %d: %s (%s)", e.Year, e.Day, e.Title, e.runner.String())
}

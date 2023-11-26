package advent

import (
	"log/slog"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

type Exercise struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Language string         `json:"-"`
	Year     int            `json:"year"`
	Day      int            `json:"day"`
	URL      string         `json:"url"`
	Data     *Data          `json:"data"`
	path     string         `json:"-"`
	runner   runners.Runner `json:"-"`
}

// Data contains the relative path to exercise input and the specific test case data for an exercise.
type Data struct {
	Input     string   `json:"-"`
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

func (e *Exercise) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", e.ID),
		slog.String("title", e.Title),
		slog.String("url", e.URL),
		slog.String("dir", e.Dir()),
		slog.String("lang", e.Language),
	)
}

func (e *Exercise) SetLanguage(lang string) {
	e.Language = lang
}

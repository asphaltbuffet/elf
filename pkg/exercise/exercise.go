package exercise

import (
	"fmt"
	"log/slog"
)

// Exercise represents a single programming challenge with its metadata, runner, and I/O configuration.
type Exercise struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Language string `json:"-"`
	Year     int    `json:"year"`
	Day      int    `json:"day"`
	URL      string `json:"url"`
	Data     *Data  `json:"data"`
	Path     string `json:"-"`

	customInput string `json:"-"`
}

// Data contains the relative path to exercise input and the specific test case data for an exercise.
type Data struct {
	InputData     string   `json:"-"`
	InputFileName string   `json:"inputFile"`
	TestCases     TestCase `json:"testCases"`
	Answers       Answer   `json:"answers"`
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

// LogValue implements [slog.LogValuer], emitting id, dir, and language as a group.
func (e *Exercise) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", e.ID),
		slog.String("dir", e.Dir()),
		slog.String("lang", e.Language),
	)
}

func (e *Exercise) String() string {
	if e.Year == 0 && e.Day == 0 && e.Title == "" {
		return "INVALID EXERCISE"
	}

	return fmt.Sprintf("Advent of Code %d, Day %d: %s", e.Year, e.Day, e.Title)
}

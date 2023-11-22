package advent

import (
	"fmt"
	"log/slog"
)

// String returns a string representation of the exercise in the format:
// `Advent of Code YYYY, Day DD: TITLE (LANGUAGE)`.
//
// Example: Advent of Code: 2020-01 (Go).
func (e *Exercise) String() string {
	if e == nil || e.ID == "" {
		slog.Error("nil or empty exercise")
		return "Advent of Code: INVALID EXERCISE"
	}

	if e.runner == nil {
		return fmt.Sprintf("Advent of Code %d, Day %d: %s (INVALID LANGUAGE)", e.Year, e.Day, e.Title)
	}

	return fmt.Sprintf("Advent of Code %d, Day %d: %s (%s)", e.Year, e.Day, e.Title, e.runner.String())
}

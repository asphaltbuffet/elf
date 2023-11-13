//go:build runtime
// +build runtime

// Package exercises contains the exercise for Advent of Code 2015 day 1.
package exercises

import (
	"fmt"
	"os"
	"strings"

	"github.com/asphaltbuffet/elf/internal/common"
)

// Exercise for Advent of Code 2015 day 1.
type Exercise struct {
	common.BaseExercise
}

// One returns the answer to the first part of the exercise.
func (e Exercise) One(in string) (any, error) {
	// the test exercise converts the input to lowercase
	return strings.ToLower(in), nil
}

// Two returns the answer to the second part of the exercise.
func (e Exercise) Two(in string) (any, error) {
	if in == "die" {
		return nil, fmt.Errorf("example fake error")
	}

	// the test exercise converts the input to uppercase
	return strings.ToUpper(in), nil
}

// Visualize returns the visualization of the exercise.
func (e Exercise) Vis(in string, out string) error {
	// the test exercise writes the input as-is
	return os.WriteFile(out, []byte(in), 0o600)
}

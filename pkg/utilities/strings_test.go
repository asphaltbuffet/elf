// Package utilities contains helper functions used by the application.
package utilities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamelToTitle(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"test data", "testDayOne", "Test Day One"},
		{"multiple words", "notQuiteLisp", "Not Quite Lisp"},
		{"single-letter capitalized word", "allInASingleNight", "All In A Single Night"},
		{"hyphen and apostrophe", "doesn'tHeHaveIntern-ElvesForThis", "Doesn't He Have Intern-Elves For This"},
		{"single word", "matchsticks", "Matchsticks"},
		{"hyphen", "cathode-RayTube", "Cathode-Ray Tube"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CamelToTitle(tt.arg)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToCamel(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"test data", "Test Day One", "testDayOne"},
		{"multiple words", "Not Quite Lisp", "notQuiteLisp"},
		{"single-letter capitalized word", "All In A Single Night", "allInASingleNight"},
		{"hyphen and apostrophe", "Doesn't He Have Intern-Elves For This", "doesn'tHeHaveIntern-ElvesForThis"},
		{"single word", "Matchsticks", "matchsticks"},
		{"hyphen", "Cathode-Ray Tube", "cathode-RayTube"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToCamel(tt.arg))
		})
	}
}

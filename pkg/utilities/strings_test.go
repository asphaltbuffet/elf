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

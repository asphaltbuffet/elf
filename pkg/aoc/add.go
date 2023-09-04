package aoc

import (
	"fmt"
	"path/filepath"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

func (ac *AOCClient) AddExercise(year int, day int, language string) (*exercise.Exercise, error) {
	// check for year
	if err := isValidYear(year); err != nil {
		// make year
	}

	// check for day/exercise
	e, err := ac.GetExercise(year, day)
	if err != nil {
		// make day/exercise
	}

	info, err := fs.Stat(filepath.Join(e.Dir, language))
	if err == nil {
		return e, fmt.Errorf("exercise already exists: %s", info.Name())
	}

	return nil, fmt.Errorf("not implemented")
}

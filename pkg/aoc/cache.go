package aoc

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

func getCachedPuzzlePage(year int, day int) ([]byte, error) {
	f, err := afero.ReadFile(appFs, filepath.Join(cfgDir, "puzzle_pages", fmt.Sprintf("%d-%d.txt", year, day)))
	if err != nil {
		return nil, fmt.Errorf("reading puzzle page: %w", err)
	}

	return f, nil
}

func getCachedInput(year, day int) ([]byte, error) {
	f, err := afero.ReadFile(appFs, filepath.Join(cfgDir, "inputs", fmt.Sprintf("%d-%d.txt", year, day)))
	if err != nil {
		return nil, fmt.Errorf("reading cached input: %w", err)
	}

	return f, nil
}

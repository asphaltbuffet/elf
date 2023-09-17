package exercise

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

// GetImplementations returns a list of available implementations for the exercise.
func (e *Exercise) GetImplementations(fs afero.Fs) ([]string, error) {
	dirEntries, err := afero.ReadDir(fs, e.Path)
	if err != nil {
		return nil, fmt.Errorf("checking %s: %w", e.Path, err)
	}

	impls := []string{}

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			continue
		}

		// TODO: should check if the implementation is more than just template
		if _, ok := runners.Available[strings.ToLower(entry.Name())]; ok {
			impls = append(impls, entry.Name())
		}
	}

	return impls, nil
}

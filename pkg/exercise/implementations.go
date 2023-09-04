package exercise

import (
	"fmt"
	"os"
	"strings"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

// GetImplementations returns a list of available implementations for the exercise.
func (e *Exercise) GetImplementations() ([]string, error) {
	dirEntries, err := os.ReadDir(e.Dir)
	if err != nil {
		return nil, fmt.Errorf("getting implementations for exercise: %w", err)
	}

	var impls []string

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

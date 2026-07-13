package runners

import (
	"fmt"
	"os"
	"path/filepath"
)

// moduleRelDir returns exerciseDir relative to the directory of the nearest
// ancestor go.mod (the enclosing Go module root). exerciseDir may be relative,
// ".", or absolute; it is resolved against the process working directory first.
//
// This is the resolver behind the {rel_exercise_dir} runner token (ADR-0021).
// It is Go-specific by design: only the Go runner needs a module-relative import
// path, so the anchor is go.mod. The walk starts at the exercise dir itself, so a
// per-exercise module (go.mod inside the exercise dir) correctly yields ".".
func moduleRelDir(exerciseDir string) (string, error) {
	abs, err := filepath.Abs(exerciseDir)
	if err != nil {
		return "", fmt.Errorf("resolving exercise dir %q: %w", exerciseDir, err)
	}

	for dir := abs; ; {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			rel, relErr := filepath.Rel(dir, abs)
			if relErr != nil {
				return "", fmt.Errorf("relativizing %q to module root %q: %w", abs, dir, relErr)
			}
			return rel, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir { // reached filesystem root
			return "", fmt.Errorf("no go.mod found above %q", abs)
		}
		dir = parent
	}
}

package analyze

import (
	"fmt"
	"os"
	"path/filepath"
)

// Scope is which kind of target the analyze operation is pointed at.
type Scope int

const (
	// ScopeExercise means the target directory is a single exercise (has info.json).
	ScopeExercise Scope = iota
	// ScopeYear means the target directory contains exercise subdirectories.
	ScopeYear
)

// detectScope classifies dir by looking at the target and at most one level
// below it. A directory of year directories (info.json two levels down) is an
// error, not a silent multi-year merge.
func detectScope(dir string) (Scope, error) {
	if fileExists(filepath.Join(dir, "info.json")) {
		return ScopeExercise, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("reading directory: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() && fileExists(filepath.Join(dir, e.Name(), "info.json")) {
			return ScopeYear, nil
		}
	}

	return 0, fmt.Errorf("expected an exercise or year directory; got %s", dir)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

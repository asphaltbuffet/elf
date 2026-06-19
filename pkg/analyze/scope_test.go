package analyze

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o600))
}

func Test_detectScope(t *testing.T) {
	t.Run("exercise scope when target has info.json", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "info.json"))

		got, err := detectScope(dir)
		require.NoError(t, err)
		assert.Equal(t, ScopeExercise, got)
	})

	t.Run("year scope when a child has info.json", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "01-puzzle", "info.json"))

		got, err := detectScope(dir)
		require.NoError(t, err)
		assert.Equal(t, ScopeYear, got)
	})

	t.Run("errors on a directory of year directories", func(t *testing.T) {
		dir := t.TempDir() // base/<year>/<day>/info.json — info.json is two levels down
		writeFile(t, filepath.Join(dir, "2015", "01-puzzle", "info.json"))

		_, err := detectScope(dir)
		require.Error(t, err)
		assert.ErrorContains(t, err, "exercise or year directory")
	})

	t.Run("errors on an empty directory", func(t *testing.T) {
		_, err := detectScope(t.TempDir())
		require.Error(t, err)
	})

	t.Run("errors on a nonexistent directory", func(t *testing.T) {
		_, err := detectScope(filepath.Join(t.TempDir(), "nope"))
		require.Error(t, err)
	})
}

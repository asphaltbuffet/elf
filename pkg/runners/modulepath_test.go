package runners

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeGoMod creates dir (and parents) and puts an empty go.mod in it.
func writeGoMod(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0o600))
}

func TestModuleRelDir_SharedRootModule(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root)
	ex := filepath.Join(root, "exercises", "2019", "12-foo")
	require.NoError(t, os.MkdirAll(ex, 0o755))

	got, err := moduleRelDir(ex)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join("exercises", "2019", "12-foo"), got)
}

func TestModuleRelDir_EulerSibling(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root)
	ex := filepath.Join(root, "euler", "42")
	require.NoError(t, os.MkdirAll(ex, 0o755))

	got, err := moduleRelDir(ex)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join("euler", "42"), got)
}

func TestModuleRelDir_PerExerciseModule(t *testing.T) {
	root := t.TempDir()
	ex := filepath.Join(root, "tracks", "go", "hamming")
	writeGoMod(t, ex) // go.mod IS in the exercise dir

	got, err := moduleRelDir(ex)

	require.NoError(t, err)
	assert.Equal(t, ".", got)
}

func TestModuleRelDir_NoGoMod(t *testing.T) {
	root := t.TempDir() // no go.mod anywhere
	ex := filepath.Join(root, "exercises", "2019", "12-foo")
	require.NoError(t, os.MkdirAll(ex, 0o755))

	_, err := moduleRelDir(ex)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no go.mod found")
}

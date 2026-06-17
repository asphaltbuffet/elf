package discover

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func setupTestFs(t *testing.T) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	// 2015 - one exercise with go impl, has part 1 answer
	require.NoError(t, fs.MkdirAll("exercises/2015/01-notCalledIt/go", 0o755))
	require.NoError(t, afero.WriteFile(fs, "exercises/2015/01-notCalledIt/info.json", []byte(`{
		"year": 2015, "day": 1, "title": "Not Called It",
		"data": {"answers": {"a": "42", "b": ""}}
	}`), 0o644))

	// 2015 - second exercise with go and py impls, both answers
	require.NoError(t, fs.MkdirAll("exercises/2015/02-inverse/go", 0o755))
	require.NoError(t, fs.MkdirAll("exercises/2015/02-inverse/py", 0o755))
	require.NoError(t, afero.WriteFile(fs, "exercises/2015/02-inverse/info.json", []byte(`{
		"year": 2015, "day": 2, "title": "Inverse",
		"data": {"answers": {"a": "100", "b": "200"}}
	}`), 0o644))

	// 2023 - one exercise, no impls, no answers
	require.NoError(t, fs.MkdirAll("exercises/2023/05-gears/", 0o755))
	require.NoError(t, afero.WriteFile(fs, "exercises/2023/05-gears/info.json", []byte(`{
		"year": 2023, "day": 5, "title": "Gears"
	}`), 0o644))

	// malformed info.json should be skipped
	require.NoError(t, fs.MkdirAll("exercises/2023/06-broken/", 0o755))
	require.NoError(t, afero.WriteFile(fs, "exercises/2023/06-broken/info.json", []byte(`{invalid`), 0o644))

	// empty info.json should be skipped
	require.NoError(t, fs.MkdirAll("exercises/2023/07-empty/", 0o755))
	require.NoError(t, afero.WriteFile(fs, "exercises/2023/07-empty/info.json", []byte(`{}`), 0o644))

	return fs
}

func TestScan(t *testing.T) {
	// Seed the runner registry so detectLanguages recognises "go" and "py".
	restore := runners.ResetRegistry(map[string]runners.RunnerCreator{
		"go": func(_ runners.ExerciseMeta) runners.Runner { return nil },
		"py": func(_ runners.ExerciseMeta) runners.Runner { return nil },
	})
	t.Cleanup(restore)

	fs := setupTestFs(t)

	result, err := Scan(fs, "exercises")
	require.NoError(t, err)

	assert.Len(t, result, 2, "should have 2 years")

	// 2015: 2 exercises sorted by day
	require.Len(t, result[2015], 2)

	assert.Equal(t, 1, result[2015][0].Day)
	assert.Equal(t, "Not Called It", result[2015][0].Title)
	assert.Equal(t, []string{"go"}, result[2015][0].Langs)
	assert.True(t, result[2015][0].HasP1)
	assert.False(t, result[2015][0].HasP2)

	assert.Equal(t, 2, result[2015][1].Day)
	assert.Equal(t, "Inverse", result[2015][1].Title)
	assert.Equal(t, []string{"go", "py"}, result[2015][1].Langs)
	assert.True(t, result[2015][1].HasP1)
	assert.True(t, result[2015][1].HasP2)

	// 2023: only the valid exercise (broken and empty skipped)
	require.Len(t, result[2023], 1)

	assert.Equal(t, 5, result[2023][0].Day)
	assert.Equal(t, "Gears", result[2023][0].Title)
	assert.Empty(t, result[2023][0].Langs)
	assert.False(t, result[2023][0].HasP1)
	assert.False(t, result[2023][0].HasP2)
}

func TestScan_EmptyRoot(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll("exercises", 0o755))

	result, err := Scan(fs, "exercises")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestScan_MissingRoot(t *testing.T) {
	fs := afero.NewMemMapFs()

	_, err := Scan(fs, "nonexistent")
	assert.Error(t, err)
}

package exercise

import (
	"io"
	"log/slog"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("valid exercise", func(t *testing.T) {
		teardownSubTest := setupSubTest(t)
		defer teardownSubTest(t)

		got, err := Load("exercises/2017/01-fakeFullDay", "py", "", testFs, logger)

		require.NoError(t, err)
		assert.Equal(t, "2017-01", got.ID)
		assert.Equal(t, "Fake Full Day", got.Title)
		assert.Equal(t, "py", got.Language)
		assert.Equal(t, 2017, got.Year)
		assert.Equal(t, 1, got.Day)
		assert.Equal(t, "https://fake.fk/2017/day/1", got.URL)
		assert.Equal(t, "input.txt", got.Data.InputFileName)
	})

	t.Run("custom input file", func(t *testing.T) {
		teardownSubTest := setupSubTest(t)
		defer teardownSubTest(t)

		got, err := Load("exercises/2017/01-fakeFullDay", "py", "custom.txt", testFs, logger)

		require.NoError(t, err)
		assert.Equal(t, "custom.txt", got.Data.InputFileName)
	})

	t.Run("empty language", func(t *testing.T) {
		teardownSubTest := setupSubTest(t)
		defer teardownSubTest(t)

		_, err := Load("exercises/2017/01-fakeFullDay", "", "", testFs, logger)

		require.Error(t, err)
	})

	t.Run("invalid language", func(t *testing.T) {
		teardownSubTest := setupSubTest(t)
		defer teardownSubTest(t)

		_, err := Load("exercises/2017/01-fakeFullDay", "fake", "", testFs, logger)

		require.Error(t, err)
	})

	t.Run("no runner error message", func(t *testing.T) {
		teardownSubTest := setupSubTest(t)
		defer teardownSubTest(t)

		_, err := Load("exercises/2017/01-fakeFullDay", "xyz", "", testFs, logger)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrNoRunner)
		require.Contains(t, err.Error(), "elf runners install")
	})

	t.Run("missing exercise directory", func(t *testing.T) {
		teardownSubTest := setupSubTest(t)
		defer teardownSubTest(t)

		_, err := Load("exercises/2015/01-doesNotExist", "go", "", testFs, logger)

		require.Error(t, err)
	})
}

func TestLoadInfo_PerKind(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("valid problem info loads", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fs, "euler/42/info.json",
			[]byte(`{"kind":"euler","number":42,"title":"Test","data":{"inputFile":"input.txt"}}`), 0o600))

		ex := &Exercise{Path: "euler/42"}
		err := ex.loadInfo(fs, logger)

		require.NoError(t, err)
		assert.Equal(t, KindProblem, ex.Kind)
		assert.Equal(t, 42, ex.Number)
	})

	t.Run("problem missing number is rejected", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fs, "euler/0/info.json",
			[]byte(`{"kind":"euler","title":"Test","data":{}}`), 0o600))

		ex := &Exercise{Path: "euler/0"}
		err := ex.loadInfo(fs, logger)

		require.ErrorIs(t, err, ErrInvalidData)
	})

	t.Run("puzzle still requires year day url", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fs, "exercises/x/info.json",
			[]byte(`{"kind":"aoc","title":"Test","year":2015,"data":{}}`), 0o600))

		ex := &Exercise{Path: "exercises/x"}
		err := ex.loadInfo(fs, logger)

		require.ErrorIs(t, err, ErrInvalidData)
	})
}

func TestKindAt(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("returns KindProblem for a scaffolded problem dir", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fs, "euler/42/info.json",
			[]byte(`{"kind":"euler","number":42,"title":"Test","data":{"inputFile":"input.txt"}}`), 0o600))

		kind, ok := KindAt(fs, logger, "euler/42")

		require.True(t, ok)
		assert.Equal(t, KindProblem, kind)
	})

	t.Run("returns false for a dir with no info.json", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		kind, ok := KindAt(fs, logger, "exercises/2015")

		require.False(t, ok)
		assert.Empty(t, kind)
	})
}

package exercise

import (
	"io"
	"log/slog"
	"testing"

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

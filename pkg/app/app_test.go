package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/exercise"
)

// testApp builds a minimal App for unit tests.
func testApp(t *testing.T) *app.App {
	t.Helper()

	return &app.App{
		FS:     afero.NewMemMapFs(),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

// TestApp_Solve_ErrNoExercise verifies that Solve returns an error when the
// exercise directory does not exist in the filesystem.
func TestApp_Solve_ErrNoExercise(t *testing.T) {
	a := testApp(t)

	_, err := a.Solve(context.Background(), "exercises/2015/01-not-there", "go", "", nil, false)
	require.Error(t, err)
}

// TestApp_Test_ErrNoExercise verifies that Test returns an error when the
// exercise directory does not exist.
func TestApp_Test_ErrNoExercise(t *testing.T) {
	a := testApp(t)

	_, err := a.Test(context.Background(), "exercises/2015/01-not-there", "go", "", nil)
	require.Error(t, err)
}

// TestApp_Benchmark_ErrNoExercise verifies that Benchmark returns an error
// when the exercise directory does not exist.
func TestApp_Benchmark_ErrNoExercise(t *testing.T) {
	a := testApp(t)

	_, err := a.Benchmark(context.Background(), "exercises/2015/01-not-there", "go", nil, 1)
	require.Error(t, err)
}

// TestApp_Solve_ErrInvalidLanguage verifies that Solve returns an error for an
// unknown language.
func TestApp_Solve_ErrInvalidLanguage(t *testing.T) {
	a := testApp(t)

	_, err := a.Solve(context.Background(), "exercises/2015/01-hello", "brainfuck", "", nil, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, exercise.ErrNoRunner)
}

// TestApp_Solve_ErrEmptyLanguage verifies that an empty language is reported as
// ErrEmptyLanguage (via exercise.Load) rather than ErrNoRunner. App no longer
// pre-checks the registry; Load owns both guards and reports the empty-language
// case first.
func TestApp_Solve_ErrEmptyLanguage(t *testing.T) {
	a := testApp(t)

	_, err := a.Solve(context.Background(), "exercises/2015/01-hello", "", "", nil, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, exercise.ErrEmptyLanguage)
}

package app_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/runners"
	"github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/runners"
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

// TestApp_Benchmark_Euler verifies that Benchmark runs against a Project Euler
// Problem: it iterates the Problem's single declared part and returns results
// (no phantom Part Two, no refusal). See ADR-0022.
func TestApp_Benchmark_Euler(t *testing.T) {
	mockRunner := mocks.NewMockRunner(t)
	mockRunner.EXPECT().String().Return("MOCK").Maybe()
	mockRunner.EXPECT().Prepare(mock.Anything).Return(nil)
	mockRunner.EXPECT().Open(mock.Anything).Return(nil)
	mockRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID:   "benchmark.1.0",
		Ok:       true,
		Output:   "42",
		Duration: 0.001,
	}, nil).Times(1) // a Problem declares a single part for 1 iteration
	mockRunner.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	mockRunner.EXPECT().Cleanup().Return(nil).Maybe()

	restore := runners.ResetRegistry(map[string]runners.RunnerCreator{
		"go": func(_ runners.ExerciseMeta) runners.Runner { return mockRunner },
	})
	defer restore()

	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	a := app.New(cfg)

	_, path, err := a.AddProblem(1, "go", "Test")
	require.NoError(t, err)

	_, err = a.Benchmark(context.Background(), path, "go", nil, 1)

	require.NoError(t, err)
}

// TestApp_Analyze_SingleProblem_NotRefused verifies that analyzing a single
// Euler Problem directory is no longer refused by the guard. (It may still
// error later for lack of benchmark.json — we assert only that it is not the
// unsupported-target refusal.) See ADR-0022.
func TestApp_Analyze_SingleProblem_NotRefused(t *testing.T) {
	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	a := app.New(cfg)

	_, path, err := a.AddProblem(1, "go", "Test")
	require.NoError(t, err)

	_, err = a.Analyze(path, "")

	require.NotErrorIs(t, err, exercise.ErrUnsupportedAnalysis)
}

// TestApp_Analyze_EulerTree_Refused verifies that analyzing the containing
// euler/ tree (a directory of Problems) is refused: cross-problem analysis is
// not supported. See ADR-0022.
func TestApp_Analyze_EulerTree_Refused(t *testing.T) {
	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	a := app.New(cfg)

	// Scaffold two Problems under the configured euler dir, then point analyze
	// at their parent directory.
	_, p1, err := a.AddProblem(1, "go", "One")
	require.NoError(t, err)
	_, _, err = a.AddProblem(2, "go", "Two")
	require.NoError(t, err)

	eulerTree := filepath.Dir(p1) // parent of euler/1 → euler/

	_, err = a.Analyze(eulerTree, "")

	require.ErrorIs(t, err, exercise.ErrUnsupportedAnalysis)
}

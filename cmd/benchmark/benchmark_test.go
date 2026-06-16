package benchmark

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/benchmark"
	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetBenchmarkCmd(t *testing.T) {
	t.Cleanup(func() { benchmarkCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetBenchmarkCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetBenchmarkCmd()
		assert.Equal(t, cmd, GetBenchmarkCmd())
	})
}

func resetState(
	t *testing.T,
	origMakeConfig func(string) (config.Config, error),
	origMakeBenchmarker func(string, string, afero.Fs, *slog.Logger) (Benchmarker, error),
) {
	t.Helper()

	t.Cleanup(func() {
		benchmarkCmd = nil
		iterations = 0
		makeConfig = origMakeConfig
		makeBenchmarker = origMakeBenchmarker
	})
}

func Test_runBenchmarkCmd(t *testing.T) {
	origMakeConfig := makeConfig
	origMakeBenchmarker := makeBenchmarker

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeBenchmarker)

		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetBenchmarkCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runBenchmarkCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad config")
	})

	t.Run("benchmarker creation error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeBenchmarker)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}
		makeBenchmarker = func(_, _ string, _ afero.Fs, _ *slog.Logger) (Benchmarker, error) {
			return nil, errors.New("bad dir")
		}

		cmd := GetBenchmarkCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runBenchmarkCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad dir")
	})

	t.Run("benchmark error prints to stderr", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeBenchmarker)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		mockBm := mocks.NewMockBenchmarker(t)
		mockBm.EXPECT().
			Benchmark(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("runner crashed"))

		makeBenchmarker = func(_, _ string, _ afero.Fs, _ *slog.Logger) (Benchmarker, error) {
			return mockBm, nil
		}

		cmd := GetBenchmarkCmd()
		cmd.SetOut(&bytes.Buffer{})
		errBuf := &bytes.Buffer{}
		cmd.SetErr(errBuf)

		err := runBenchmarkCmd(cmd, []string{"."})
		require.NoError(t, err, "runBenchmarkCmd should return nil even when Benchmark fails")
		assert.Contains(t, errBuf.String(), "benchmark failed")
		assert.Contains(t, errBuf.String(), "runner crashed")
	})

	t.Run("happy path", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeBenchmarker)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		mockBm := mocks.NewMockBenchmarker(t)
		mockBm.EXPECT().
			Benchmark(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, nil)

		makeBenchmarker = func(_, _ string, _ afero.Fs, _ *slog.Logger) (Benchmarker, error) {
			return mockBm, nil
		}

		cmd := GetBenchmarkCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runBenchmarkCmd(cmd, []string{"."})
		assert.NoError(t, err)
	})

	t.Run("iterations flag passed to benchmarker", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeBenchmarker)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		mockBm := mocks.NewMockBenchmarker(t)
		mockBm.EXPECT().Benchmark(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 42).
			Return(nil, nil)

		makeBenchmarker = func(_, _ string, _ afero.Fs, _ *slog.Logger) (Benchmarker, error) {
			return mockBm, nil
		}

		cmd := GetBenchmarkCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		// Set iterations AFTER GetBenchmarkCmd() — flag creation resets the variable.
		iterations = 42

		err := runBenchmarkCmd(cmd, []string{"."})
		assert.NoError(t, err)
	})
}

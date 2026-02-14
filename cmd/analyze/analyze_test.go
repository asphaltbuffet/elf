package analyze

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/analyze"
	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetAnalyzeCmd(t *testing.T) {
	t.Cleanup(func() { analyzeCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetAnalyzeCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetAnalyzeCmd()
		assert.Equal(t, cmd, GetAnalyzeCmd())
	})
}

// resetState restores package-level variables and factory functions to defaults.
func resetState(
	t *testing.T,
	origMakeConfig func(string) (config.Config, error),
	origMakeAnalyzer func(config.ExerciseConfiguration, string, string) (Analyzer, error),
) {
	t.Helper()

	t.Cleanup(func() {
		analyzeCmd = nil
		outFile = ""
		makeConfig = origMakeConfig
		makeAnalyzer = origMakeAnalyzer
	})
}

func Test_runAnalyzeCmd(t *testing.T) {
	origMakeConfig := makeConfig
	origMakeAnalyzer := makeAnalyzer

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAnalyzer)

		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetAnalyzeCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runAnalyzeCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad config")
	})

	t.Run("analyzer creation error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAnalyzer)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}
		makeAnalyzer = func(_ config.ExerciseConfiguration, _, _ string) (Analyzer, error) {
			return nil, errors.New("bad analyzer config")
		}

		cmd := GetAnalyzeCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runAnalyzeCmd(cmd, []string{"."})
		require.ErrorContains(t, err, "creating grapher")
		assert.ErrorContains(t, err, "bad analyzer config")
	})

	t.Run("graph error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAnalyzer)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		mockAn := mocks.NewMockAnalyzer(t)
		mockAn.EXPECT().Graph().Return(errors.New("render failed"))

		makeAnalyzer = func(_ config.ExerciseConfiguration, _, _ string) (Analyzer, error) {
			return mockAn, nil
		}

		cmd := GetAnalyzeCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runAnalyzeCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "render failed")
	})

	t.Run("happy path", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAnalyzer)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		mockAn := mocks.NewMockAnalyzer(t)
		mockAn.EXPECT().Graph().Return(nil)

		makeAnalyzer = func(_ config.ExerciseConfiguration, _, _ string) (Analyzer, error) {
			return mockAn, nil
		}

		cmd := GetAnalyzeCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runAnalyzeCmd(cmd, []string{"."})
		assert.NoError(t, err)
	})

	t.Run("graph flag passed to analyzer", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAnalyzer)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		var gotOut string

		mockAn := mocks.NewMockAnalyzer(t)
		mockAn.EXPECT().Graph().Return(nil)

		makeAnalyzer = func(_ config.ExerciseConfiguration, _, out string) (Analyzer, error) {
			gotOut = out
			return mockAn, nil
		}

		cmd := GetAnalyzeCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		// Set outFile AFTER GetAnalyzeCmd() — flag creation resets the variable.
		outFile = "/tmp/custom-graph.png"

		err := runAnalyzeCmd(cmd, []string{"."})
		require.NoError(t, err)
		assert.Equal(t, "/tmp/custom-graph.png", gotOut)
	})
}

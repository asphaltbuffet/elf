package download

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/download"
	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
)

func TestGetDownloadCmd(t *testing.T) {
	t.Cleanup(func() { downloadCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetDownloadCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetDownloadCmd()
		assert.Equal(t, cmd, GetDownloadCmd())
	})
}

// resetState restores package-level variables and factory functions to defaults.
func resetState(
	t *testing.T,
	origMakeConfig func(string) (config.Config, error),
	origMakeAdder func(config.Config, string, string, *exercise.Overwrites) (Adder, error),
) {
	t.Helper()

	t.Cleanup(func() {
		downloadCmd = nil
		language = ""
		forceInput = false
		makeConfig = origMakeConfig
		makeAdder = origMakeAdder
	})
}

func Test_runDownloadCmd(t *testing.T) {
	origMakeConfig := makeConfig
	origMakeAdder := makeAdder

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAdder)

		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetDownloadCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		assert.ErrorContains(t, err, "bad config")
	})

	t.Run("downloader creation error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAdder)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}
		makeAdder = func(_ config.Config, _, _ string, _ *exercise.Overwrites) (Adder, error) {
			return nil, errors.New("invalid URL")
		}

		cmd := GetDownloadCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		require.ErrorContains(t, err, "creating adder")
		assert.ErrorContains(t, err, "invalid URL")
	})

	t.Run("download error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAdder)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		mockAdder := mocks.NewMockAdder(t)
		mockAdder.EXPECT().Add().Return(errors.New("network timeout"))

		makeAdder = func(_ config.Config, _, _ string, _ *exercise.Overwrites) (Adder, error) {
			return mockAdder, nil
		}

		cmd := GetDownloadCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		require.ErrorContains(t, err, "adding challenge")
		assert.ErrorContains(t, err, "network timeout")
	})

	t.Run("happy path prints file path", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAdder)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		mockAdder := mocks.NewMockAdder(t)
		mockAdder.EXPECT().Add().Return(nil)
		mockAdder.EXPECT().FilePath().Return("/exercises/2023/01-trebuchet")
		mockAdder.EXPECT().Report().Return(exercise.Report{
			{Path: "input.txt", Outcome: exercise.Added},
			{Path: "info.json", Outcome: exercise.Skipped},
		})

		makeAdder = func(_ config.Config, _, _ string, _ *exercise.Overwrites) (Adder, error) {
			return mockAdder, nil
		}

		cmd := GetDownloadCmd()
		outBuf := &bytes.Buffer{}
		cmd.SetOut(outBuf)
		cmd.SetErr(&bytes.Buffer{})

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		require.NoError(t, err)
		out := outBuf.String()
		assert.Contains(t, out, "/exercises/2023/01-trebuchet")
		assert.Contains(t, out, "input.txt")
		assert.Contains(t, out, "added")
		assert.Contains(t, out, "info.json")
		assert.Contains(t, out, "skipped")
	})

	t.Run("language flag passed to downloader", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAdder)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		var gotLang string

		mockAdder := mocks.NewMockAdder(t)
		mockAdder.EXPECT().Add().Return(nil)
		mockAdder.EXPECT().FilePath().Return("/some/path")
		mockAdder.EXPECT().Report().Return(nil)

		makeAdder = func(_ config.Config, _, lang string, _ *exercise.Overwrites) (Adder, error) {
			gotLang = lang
			return mockAdder, nil
		}

		cmd := GetDownloadCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		// Set language AFTER GetDownloadCmd() — flag creation resets the variable.
		language = "py"

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		require.NoError(t, err)
		assert.Equal(t, "py", gotLang)
	})

	t.Run("force-input flag passed to downloader", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeAdder)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		var gotForced *exercise.Overwrites

		mockAdder := mocks.NewMockAdder(t)
		mockAdder.EXPECT().Add().Return(nil)
		mockAdder.EXPECT().FilePath().Return("/some/path")
		mockAdder.EXPECT().Report().Return(nil)

		makeAdder = func(_ config.Config, _, _ string, forced *exercise.Overwrites) (Adder, error) {
			gotForced = forced
			return mockAdder, nil
		}

		cmd := GetDownloadCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		// Set forceInput AFTER GetDownloadCmd() — flag creation resets the variable.
		forceInput = true

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		require.NoError(t, err)
		require.NotNil(t, gotForced)
		assert.True(t, gotForced.Input)
	})
}

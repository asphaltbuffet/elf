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
	origMakeDownloader func(config.Config, string, string, *exercise.Overwrites) (Downloader, error),
) {
	t.Helper()

	t.Cleanup(func() {
		downloadCmd = nil
		language = ""
		forceInput = false
		makeConfig = origMakeConfig
		makeDownloader = origMakeDownloader
	})
}

func Test_runDownloadCmd(t *testing.T) {
	origMakeConfig := makeConfig
	origMakeDownloader := makeDownloader

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeDownloader)

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
		resetState(t, origMakeConfig, origMakeDownloader)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}
		makeDownloader = func(_ config.Config, _, _ string, _ *exercise.Overwrites) (Downloader, error) {
			return nil, errors.New("invalid URL")
		}

		cmd := GetDownloadCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		require.ErrorContains(t, err, "creating downloader")
		assert.ErrorContains(t, err, "invalid URL")
	})

	t.Run("download error", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeDownloader)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		mockDl := mocks.NewMockDownloader(t)
		mockDl.EXPECT().Download().Return(errors.New("network timeout"))

		makeDownloader = func(_ config.Config, _, _ string, _ *exercise.Overwrites) (Downloader, error) {
			return mockDl, nil
		}

		cmd := GetDownloadCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		require.ErrorContains(t, err, "downloading challenge")
		assert.ErrorContains(t, err, "network timeout")
	})

	t.Run("happy path prints file path", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeDownloader)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		mockDl := mocks.NewMockDownloader(t)
		mockDl.EXPECT().Download().Return(nil)
		mockDl.EXPECT().FilePath().Return("/exercises/2023/01-trebuchet")

		makeDownloader = func(_ config.Config, _, _ string, _ *exercise.Overwrites) (Downloader, error) {
			return mockDl, nil
		}

		cmd := GetDownloadCmd()
		outBuf := &bytes.Buffer{}
		cmd.SetOut(outBuf)
		cmd.SetErr(&bytes.Buffer{})

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		require.NoError(t, err)
		assert.Contains(t, outBuf.String(), "/exercises/2023/01-trebuchet")
	})

	t.Run("language flag passed to downloader", func(t *testing.T) {
		resetState(t, origMakeConfig, origMakeDownloader)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		var gotLang string

		mockDl := mocks.NewMockDownloader(t)
		mockDl.EXPECT().Download().Return(nil)
		mockDl.EXPECT().FilePath().Return("/some/path")

		makeDownloader = func(_ config.Config, _, lang string, _ *exercise.Overwrites) (Downloader, error) {
			gotLang = lang
			return mockDl, nil
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
		resetState(t, origMakeConfig, origMakeDownloader)

		makeConfig = func(_ string) (config.Config, error) {
			return config.NewConfig()
		}

		var gotForced *exercise.Overwrites

		mockDl := mocks.NewMockDownloader(t)
		mockDl.EXPECT().Download().Return(nil)
		mockDl.EXPECT().FilePath().Return("/some/path")

		makeDownloader = func(_ config.Config, _, _ string, forced *exercise.Overwrites) (Downloader, error) {
			gotForced = forced
			return mockDl, nil
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

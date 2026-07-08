package download

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// Test_runDownloadCmd covers the cmd-level concerns: flag/arg parsing and config
// construction. The download operation itself (fetch + scaffold) is exercised in
// pkg/app (App.Add) and pkg/exercise (Adder); see ADR-0005.
func Test_runDownloadCmd(t *testing.T) {
	origMakeConfig := makeConfig

	t.Cleanup(func() {
		downloadCmd = nil
		language = ""
		forceInput = false
		makeConfig = origMakeConfig
	})

	t.Run("config error", func(t *testing.T) {
		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetDownloadCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		assert.ErrorContains(t, err, "bad config")
	})
}

func Test_downloadAcceptsTwoArgs(t *testing.T) {
	origMakeConfig := makeConfig
	t.Cleanup(func() {
		downloadCmd = nil
		language = ""
		forceInput = false
		makeConfig = origMakeConfig
	})

	makeConfig = func(_ string) (config.Config, error) {
		return config.Config{}, errors.New("stub config error")
	}

	cmd := GetDownloadCmd()
	cmd.SetArgs([]string{"2015", "1"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	// Two args are now accepted by Args, so we reach RunE and hit the stubbed
	// config error — NOT a cobra "accepts 1 arg(s)" argument-count error.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stub config error")
}

func Test_downloadRejectsOutOfRangeDay(t *testing.T) {
	origMakeConfig := makeConfig
	t.Cleanup(func() {
		downloadCmd = nil
		language = ""
		forceInput = false
		makeConfig = origMakeConfig
	})

	// A zero-value config constructs without error, so control passes
	// RegisterRunners/New and reaches resolveTarget. resolveTarget rejects the
	// day before Add touches the config, so the empty config never matters.
	makeConfig = func(_ string) (config.Config, error) {
		return config.Config{}, nil
	}

	cmd := GetDownloadCmd()
	cmd.SetArgs([]string{"2015", "26"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

func Test_resolveTarget_URLPassThrough(t *testing.T) {
	got, err := resolveTarget([]string{"https://adventofcode.com/2015/day/1"})
	require.NoError(t, err)
	assert.Equal(t, "https://adventofcode.com/2015/day/1", got)
}

func Test_resolveTarget_AssemblesURL(t *testing.T) {
	got, err := resolveTarget([]string{"2015", "1"})
	require.NoError(t, err)
	assert.Equal(t, "https://adventofcode.com/2015/day/1", got)

	// The assembled URL must satisfy the domain's own parser.
	year, day, perr := exercise.ParseURL(got)
	require.NoError(t, perr)
	assert.Equal(t, 2015, year)
	assert.Equal(t, 1, day)
}

func Test_resolveTarget_InvalidInputs(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"non-numeric year", []string{"twenty", "1"}, "year"},
		{"non-numeric day", []string{"2015", "xmas"}, "day"},
		{"day too high", []string{"2015", "26"}, "out of range"},
		{"day zero", []string{"2015", "0"}, "out of range"},
		{"year too early", []string{"2014", "1"}, "before Advent of Code"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveTarget(tc.args)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
			assert.Empty(t, got, "no URL should be returned on error")
		})
	}
}

package visualize

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetVisualizeCmd(t *testing.T) {
	t.Cleanup(func() { visualizeCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetVisualizeCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetVisualizeCmd()
		assert.Equal(t, cmd, GetVisualizeCmd())
	})
}

func resetState(t *testing.T, origMakeConfig func(string) (config.Config, error)) {
	t.Helper()

	t.Cleanup(func() {
		visualizeCmd = nil
		language = ""
		outdir = ""
		plainFlag = false
		jsonFlag = false
		makeConfig = origMakeConfig
	})
}

func Test_runVisualizeCmd(t *testing.T) {
	origMakeConfig := makeConfig

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig)

		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetVisualizeCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runVisualizeCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad config")
	})
}

func Test_resolveOutdir(t *testing.T) {
	t.Run("empty outdir defaults to exercise dir", func(t *testing.T) {
		got, err := resolveOutdir("", "/abs/exercises/2015/01-notQuiteLisp")
		require.NoError(t, err)
		assert.Equal(t, "/abs/exercises/2015/01-notQuiteLisp", got)
	})

	t.Run("explicit absolute outdir is used as-is", func(t *testing.T) {
		got, err := resolveOutdir("/tmp/out", "/abs/exercises/2015/01-notQuiteLisp")
		require.NoError(t, err)
		assert.Equal(t, "/tmp/out", got)
	})

	t.Run("explicit relative outdir is made absolute", func(t *testing.T) {
		got, err := resolveOutdir("sub/out", "/abs/exercises/2015/01-notQuiteLisp")
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(got), "expected absolute path, got %q", got)
		assert.Equal(t, "out", filepath.Base(got))
	})
}

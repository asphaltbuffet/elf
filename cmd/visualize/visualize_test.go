package visualize

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

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

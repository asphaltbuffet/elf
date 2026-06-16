package test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetTestCmd(t *testing.T) {
	t.Cleanup(func() { testCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetTestCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetTestCmd()
		assert.Equal(t, cmd, GetTestCmd())
	})
}

// resetState restores package-level variables and factory functions to defaults.
func resetState(t *testing.T, origMakeConfig func(string) (config.Config, error)) {
	t.Helper()

	t.Cleanup(func() {
		testCmd = nil
		language = ""
		makeConfig = origMakeConfig
	})
}

func Test_runTestCmd(t *testing.T) {
	origMakeConfig := makeConfig

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig)

		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetTestCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runTestCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad config")
	})
}

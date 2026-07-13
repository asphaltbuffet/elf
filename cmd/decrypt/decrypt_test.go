package decrypt

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetDecryptCmd(t *testing.T) {
	t.Cleanup(func() { decryptCmd = nil })

	t.Run("new command", func(t *testing.T) {
		cmd := GetDecryptCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "decrypt", cmd.Name())
		assert.NotNil(t, cmd.RunE)
		assert.NotNil(t, cmd.Flags().Lookup("identity"))
		assert.NotNil(t, cmd.Flags().Lookup("force"))
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetDecryptCmd()
		assert.Equal(t, cmd, GetDecryptCmd())
	})
}

// Test_runDecryptCmd covers the cmd-level concerns: config construction and arg
// handling. The decrypt operation itself is exercised in pkg/app (App.Decrypt) and
// pkg/secret (Decrypt).
func Test_runDecryptCmd(t *testing.T) {
	origMakeConfig := makeConfig

	t.Cleanup(func() {
		decryptCmd = nil
		makeConfig = origMakeConfig
	})

	t.Run("config error", func(t *testing.T) {
		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetDecryptCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runDecryptCmd(cmd, []string{"some/dir"})
		assert.ErrorContains(t, err, "bad config")
	})
}

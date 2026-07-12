package encrypt

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetEncryptCmd(t *testing.T) {
	t.Cleanup(func() { encryptCmd = nil })

	t.Run("new command", func(t *testing.T) {
		cmd := GetEncryptCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "encrypt", cmd.Name())
		assert.NotNil(t, cmd.RunE)
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetEncryptCmd()
		assert.Equal(t, cmd, GetEncryptCmd())
	})
}

// Test_runEncryptCmd covers the cmd-level concerns: config construction and arg
// handling. The encrypt operation itself is exercised in pkg/app (App.Encrypt) and
// pkg/secret (Encrypt).
func Test_runEncryptCmd(t *testing.T) {
	origMakeConfig := makeConfig

	t.Cleanup(func() {
		encryptCmd = nil
		makeConfig = origMakeConfig
	})

	t.Run("config error", func(t *testing.T) {
		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetEncryptCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runEncryptCmd(cmd, []string{"some/dir"})
		assert.ErrorContains(t, err, "bad config")
	})
}

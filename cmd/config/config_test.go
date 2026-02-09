package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/cmd/config"
)

func TestGetConfigCmd(t *testing.T) {
	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, config.GetConfigCmd())
	})

	t.Run("existing command returns same instance", func(t *testing.T) {
		cmd := config.GetConfigCmd()
		assert.Equal(t, cmd, config.GetConfigCmd())
	})
}

func TestGetConfigCmd_Subcommands(t *testing.T) {
	cmd := config.GetConfigCmd()

	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	// Check subcommand names
	names := make([]string, len(subcommands))
	for i, sc := range subcommands {
		names[i] = sc.Name()
	}

	assert.Contains(t, names, "init")
	assert.Contains(t, names, "check")
	assert.Contains(t, names, "update-token")
}

func TestGetConfigCmd_Usage(t *testing.T) {
	cmd := config.GetConfigCmd()

	assert.Equal(t, "config", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

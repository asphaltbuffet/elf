package versioncmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	versioncmd "github.com/asphaltbuffet/elf/cmd/version"
)

func TestNewVersionCmd(t *testing.T) {
	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, versioncmd.NewVersionCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := versioncmd.NewVersionCmd()
		assert.NotEqual(t, cmd, versioncmd.NewVersionCmd())
	})
}

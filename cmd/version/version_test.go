package version_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/cmd/version"
)

func TestNewVersionCmd(t *testing.T) {
	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, version.NewVersionCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := version.NewVersionCmd()
		assert.NotEqual(t, cmd, version.NewVersionCmd())
	})
}

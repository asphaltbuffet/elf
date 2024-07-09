package download_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/cmd/download"
)

func TestGetDownloadCmd(t *testing.T) {
	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, download.GetDownloadCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := download.GetDownloadCmd()
		assert.Equal(t, cmd, download.GetDownloadCmd())
	})
}

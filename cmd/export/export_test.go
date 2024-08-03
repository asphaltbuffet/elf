package export_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/cmd/export"
	mocks "github.com/asphaltbuffet/elf/mocks/krampus"
)

func TestGetAnalyzeCmd(t *testing.T) {
	mockConfig := mocks.NewMockExerciseConfiguration(t)

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, export.NewCommand(mockConfig))
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := export.NewCommand(mockConfig)
		assert.NotEqual(t, cmd, export.NewCommand(mockConfig))
	})
}

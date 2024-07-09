package analyze_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/cmd/analyze"
)

func TestGetAnalyzeCmd(t *testing.T) {
	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, analyze.GetAnalyzeCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := analyze.GetAnalyzeCmd()
		assert.Equal(t, cmd, analyze.GetAnalyzeCmd())
	})
}

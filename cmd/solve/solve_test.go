package solve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/cmd/solve"
)

func TestGetSolveCmd(t *testing.T) {
	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, solve.GetSolveCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := solve.GetSolveCmd()
		assert.Equal(t, cmd, solve.GetSolveCmd())
	})
}

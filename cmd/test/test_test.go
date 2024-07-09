package test_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/cmd/test"
)

func TestGetTestCmd(t *testing.T) {
	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, test.GetTestCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := test.GetTestCmd()
		assert.Equal(t, cmd, test.GetTestCmd())
	})
}

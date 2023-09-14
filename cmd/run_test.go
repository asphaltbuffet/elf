package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetRunCmd(t *testing.T) {
	got := GetRunCmd()

	assert.NotNil(t, got)
	assert.IsType(t, &cobra.Command{}, got)
	assert.Equal(t, "run", got.Name())
}

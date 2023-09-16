package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetVisualizeCmd(t *testing.T) {
	got := GetVisualizeCmd()

	assert.NotNil(t, got)
	assert.IsType(t, &cobra.Command{}, got)
	assert.Equal(t, "visualize", got.Name())
}

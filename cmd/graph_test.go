package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetGraphCmd(t *testing.T) {
	got := GetGraphCmd()

	assert.NotNil(t, got)
	assert.IsType(t, &cobra.Command{}, got)
	assert.Equal(t, "graph", got.Name())
}

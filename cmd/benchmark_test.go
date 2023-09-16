package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetBenchmarkCmd(t *testing.T) {
	got := GetBenchmarkCmd()

	assert.NotNil(t, got)
	assert.IsType(t, &cobra.Command{}, got)
	assert.Equal(t, "benchmark", got.Name())
}

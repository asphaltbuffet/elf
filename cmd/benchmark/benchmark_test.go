package benchmark_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/cmd/benchmark"
)

func TestGetBenchmarkCmd(t *testing.T) {
	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, benchmark.GetBenchmarkCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := benchmark.GetBenchmarkCmd()
		assert.Equal(t, cmd, benchmark.GetBenchmarkCmd())
	})
}

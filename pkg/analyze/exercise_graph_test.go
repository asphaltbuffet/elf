package analyze

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_generateBoxPlot(t *testing.T) {
	t.Run("writes a png for a single exercise", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "run-times.png")
		data := makeBenchmarkData(2015, 1) // one day, Golang + Python, both parts

		err := generateBoxPlot(data, out)
		require.NoError(t, err)

		info, statErr := os.Stat(out)
		require.NoError(t, statErr)
		assert.Positive(t, info.Size())
	})

	t.Run("handles a missing part two", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "run-times.png")
		data := makeBenchmarkDataNilPartTwo(2015, 1)

		err := generateBoxPlot(data, out)
		require.NoError(t, err)
	})

	t.Run("errors with no data", func(t *testing.T) {
		err := generateBoxPlot(nil, filepath.Join(t.TempDir(), "x.png"))
		require.Error(t, err)
		assert.ErrorContains(t, err, "no benchmark data")
	})
}

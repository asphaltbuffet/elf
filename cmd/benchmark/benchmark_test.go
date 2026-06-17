package benchmark

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetBenchmarkCmd(t *testing.T) {
	t.Cleanup(func() { benchmarkCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetBenchmarkCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetBenchmarkCmd()
		assert.Equal(t, cmd, GetBenchmarkCmd())
	})
}

func resetState(t *testing.T, origMakeConfig func(string) (config.Config, error)) {
	t.Helper()

	t.Cleanup(func() {
		benchmarkCmd = nil
		iterations = 0
		makeConfig = origMakeConfig
	})
}

func Test_runBenchmarkCmd(t *testing.T) {
	origMakeConfig := makeConfig

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig)

		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetBenchmarkCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runBenchmarkCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad config")
	})
}

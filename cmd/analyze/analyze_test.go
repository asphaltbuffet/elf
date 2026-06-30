package analyze

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetAnalyzeCmd(t *testing.T) {
	t.Cleanup(func() { analyzeCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetAnalyzeCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetAnalyzeCmd()
		assert.Equal(t, cmd, GetAnalyzeCmd())
	})
}

// Test_runAnalyzeCmd covers the cmd-level concerns: config construction and arg
// handling. The analysis operation itself (load + graph) is exercised in pkg/app
// (App.Analyze) and pkg/analyze (Analyzer); see ADR-0005.
func Test_runAnalyzeCmd(t *testing.T) {
	origMakeConfig := makeConfig

	t.Cleanup(func() {
		analyzeCmd = nil
		outFile = ""
		makeConfig = origMakeConfig
	})

	t.Run("config error", func(t *testing.T) {
		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetAnalyzeCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runAnalyzeCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad config")
	})
}

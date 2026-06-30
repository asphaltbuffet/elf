package download

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetDownloadCmd(t *testing.T) {
	t.Cleanup(func() { downloadCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetDownloadCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetDownloadCmd()
		assert.Equal(t, cmd, GetDownloadCmd())
	})
}

// Test_runDownloadCmd covers the cmd-level concerns: flag/arg parsing and config
// construction. The download operation itself (fetch + scaffold) is exercised in
// pkg/app (App.Add) and pkg/exercise (Adder); see ADR-0005.
func Test_runDownloadCmd(t *testing.T) {
	origMakeConfig := makeConfig

	t.Cleanup(func() {
		downloadCmd = nil
		language = ""
		forceInput = false
		makeConfig = origMakeConfig
	})

	t.Run("config error", func(t *testing.T) {
		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetDownloadCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runDownloadCmd(cmd, []string{"https://adventofcode.com/2023/day/1"})
		assert.ErrorContains(t, err, "bad config")
	})
}

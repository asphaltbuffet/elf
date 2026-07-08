package solve

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestGetSolveCmd(t *testing.T) {
	t.Cleanup(func() { solveCmd = nil })

	t.Run("new command", func(t *testing.T) {
		assert.NotNil(t, GetSolveCmd())
	})

	t.Run("existing command", func(t *testing.T) {
		cmd := GetSolveCmd()
		assert.Equal(t, cmd, GetSolveCmd())
	})
}

func resetState(t *testing.T, origMakeConfig func(string) (config.Config, error)) {
	t.Helper()

	t.Cleanup(func() {
		solveCmd = nil
		language = ""
		input = ""
		noTest = false
		jsonFlag = false
		makeConfig = origMakeConfig
	})
}

func Test_solveHasJSONFlag(t *testing.T) {
	t.Cleanup(func() { solveCmd = nil })

	cmd := GetSolveCmd()
	f := cmd.Flags().Lookup("json")
	require.NotNil(t, f, "solve should register a --json flag")
	assert.Equal(t, "false", f.DefValue)
}

func Test_solvePlainAndJSONMutuallyExclusive(t *testing.T) {
	t.Cleanup(func() { solveCmd = nil; plainFlag = false; jsonFlag = false })

	cmd := GetSolveCmd()
	cmd.SetArgs([]string{"--plain", "--json", "somepath"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "if any flags in the group [plain json] are set none of the others can be")
}

func Test_runSolveCmd(t *testing.T) {
	origMakeConfig := makeConfig

	t.Run("config error", func(t *testing.T) {
		resetState(t, origMakeConfig)

		makeConfig = func(_ string) (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		}

		cmd := GetSolveCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := runSolveCmd(cmd, []string{"."})
		assert.ErrorContains(t, err, "bad config")
	})
}

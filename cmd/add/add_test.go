package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAddCmd_HasSubcommands(t *testing.T) {
	cmd := GetAddCmd()

	names := map[string]bool{}
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}

	assert.True(t, names["aoc"], "add must have an aoc subcommand")
	assert.True(t, names["euler"], "add must have a euler subcommand")
}

func TestEuler_RejectsNonNumericArg(t *testing.T) {
	err := runEulerCmd(GetAddCmd(), []string{"abc"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a number")
}

func Test_eulerCmd_hasNoTitleFlag(t *testing.T) {
	addCmd = nil // reset package-level singleton
	t.Cleanup(func() { addCmd = nil })

	c := eulerCmd()

	assert.Nil(t, c.Flags().Lookup("title"), "euler command must not expose a --title flag")
}

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

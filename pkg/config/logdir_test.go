package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStateDir_XDGStateHomeSet(t *testing.T) {
	// On Linux, an explicit XDG_STATE_HOME is honored verbatim.
	if runtimeGOOS() != "linux" {
		t.Skip("XDG_STATE_HOME semantics are Linux-only")
	}

	t.Setenv("XDG_STATE_HOME", "/tmp/custom-state")

	got, err := userStateDir()
	require.NoError(t, err)
	assert.Equal(t, "/tmp/custom-state", got)
}

func TestUserStateDir_XDGStateHomeUnset(t *testing.T) {
	// On Linux with XDG_STATE_HOME unset, fall back to ~/.local/state.
	if runtimeGOOS() != "linux" {
		t.Skip("XDG_STATE_HOME semantics are Linux-only")
	}

	t.Setenv("XDG_STATE_HOME", "")

	home, err := homeDir()
	require.NoError(t, err)

	got, err := userStateDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".local", "state"), got)
}

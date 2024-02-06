package krampus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	got, err := New()

	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "exercises", got.GetString("advent.dir"))
}

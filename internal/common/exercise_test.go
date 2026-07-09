package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseExercise_One(t *testing.T) {
	t.Parallel()

	e := BaseExercise{}

	got, err := e.One("fake")

	require.Error(t, err)
	assert.Nil(t, got)
}

func TestBaseExercise_Two(t *testing.T) {
	t.Parallel()

	e := BaseExercise{}

	got, err := e.Two("fake")

	require.Error(t, err)
	assert.Nil(t, got)
}

func TestBaseExercise_Vis(t *testing.T) {
	t.Parallel()

	e := BaseExercise{}

	err := e.Vis("fake", "/tmp")
	require.Error(t, err)
}

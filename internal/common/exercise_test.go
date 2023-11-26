// Package common contains the base struct for all exercises.
package common

import (
	"bytes"
	"io"
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
	var b bytes.Buffer

	w := io.Writer(&b)
	e := BaseExercise{}

	require.Error(t, e.Vis("fake", &w))
	assert.Empty(t, b.String())
}

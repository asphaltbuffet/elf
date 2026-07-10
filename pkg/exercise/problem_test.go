package exercise

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

func TestExercise_declaredParts(t *testing.T) {
	t.Run("problem declares only part one", func(t *testing.T) {
		ex := &Exercise{Kind: KindProblem}
		assert.Equal(t, []protocol.Part{protocol.PartOne}, ex.declaredParts())
	})

	t.Run("puzzle declares part one and two", func(t *testing.T) {
		ex := &Exercise{Kind: KindPuzzle}
		assert.Equal(t, []protocol.Part{protocol.PartOne, protocol.PartTwo}, ex.declaredParts())
	})

	t.Run("empty kind defaults to two parts", func(t *testing.T) {
		ex := &Exercise{}
		assert.Equal(t, []protocol.Part{protocol.PartOne, protocol.PartTwo}, ex.declaredParts())
	})
}

func TestMakeProblemID(t *testing.T) {
	assert.Equal(t, "euler-42", makeProblemID(42))
	assert.Equal(t, "euler-100", makeProblemID(100))
	assert.Equal(t, "euler-7", makeProblemID(7))
}

func TestNewProblemFromSource(t *testing.T) {
	ex := newProblemFromSource(problemSource{
		baseDir:  "/work",
		language: "go",
		title:    "Multiples of 3 or 5",
		number:   1,
	})

	assert.Equal(t, KindProblem, ex.Kind)
	assert.Equal(t, 1, ex.Number)
	assert.Equal(t, "euler-1", ex.ID)
	assert.Equal(t, "go", ex.Language)
	assert.Equal(t, filepath.Join("/work", "euler", "1"), ex.Path)
	assert.Empty(t, ex.Data.InputData)
	assert.Empty(t, ex.Data.InputFileName)
	assert.Nil(t, ex.Data.TestCases.Two)
	assert.Zero(t, ex.Year)
	assert.Zero(t, ex.Day)
	assert.Empty(t, ex.URL)
}

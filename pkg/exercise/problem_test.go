package exercise

import (
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

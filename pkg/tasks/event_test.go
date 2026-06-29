package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

func TestPlannedEvent(t *testing.T) {
	e := PlannedEvent("Test.1.1", Test, protocol.PartOne, 1, "")
	assert.Equal(t, EventPlanned, e.Kind)
	assert.Equal(t, "Test.1.1", e.ID)
	assert.Equal(t, Test, e.Type)
	assert.Equal(t, protocol.PartOne, e.Part)
	assert.Equal(t, 1, e.SubPart)
	assert.Empty(t, e.Language)
	assert.Nil(t, e.Result)
}

func TestPlannedEventCarriesLanguage(t *testing.T) {
	e := PlannedEvent("Benchmark.1.0", Benchmark, protocol.PartOne, 0, "Go")
	assert.Equal(t, "Go", e.Language)
}

func TestStartedEvent(t *testing.T) {
	e := StartedEvent("Solve.2", Solve, protocol.PartTwo, 0, "")
	assert.Equal(t, EventStarted, e.Kind)
	assert.Equal(t, "Solve.2", e.ID)
	assert.Empty(t, e.Language)
	assert.Nil(t, e.Result)
}

func TestStartedEventCarriesLanguage(t *testing.T) {
	e := StartedEvent("Benchmark.1.0", Benchmark, protocol.PartOne, 0, "Python")
	assert.Equal(t, "Python", e.Language)
}

func TestFinishedEvent(t *testing.T) {
	r := Result{ID: "Test.1.1", Type: Test, Part: protocol.PartOne, SubPart: 1, Status: StatusPassed, Duration: 0.5}
	e := FinishedEvent(r, "")
	assert.Equal(t, EventFinished, e.Kind)
	assert.Equal(t, "Test.1.1", e.ID)
	assert.Equal(t, Test, e.Type)
	assert.Equal(t, protocol.PartOne, e.Part)
	assert.Equal(t, 1, e.SubPart)
	assert.Empty(t, e.Language)
	require.NotNil(t, e.Result)
	assert.Equal(t, StatusPassed, e.Result.Status)
}

// FinishedEvent takes Language explicitly because tasks.Result is frozen and
// does not carry it (see ADR-0011); benchmark supplies the runner name here.
func TestFinishedEventCarriesLanguage(t *testing.T) {
	r := Result{
		ID:       "Benchmark.1.0",
		Type:     Benchmark,
		Part:     protocol.PartOne,
		SubPart:  0,
		Status:   StatusPassed,
		Duration: 0.01,
	}
	e := FinishedEvent(r, "Go")
	assert.Equal(t, "Go", e.Language)
}

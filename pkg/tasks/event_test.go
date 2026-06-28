package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

func TestPlannedEvent(t *testing.T) {
	e := PlannedEvent("Test.1.1", Test, protocol.PartOne, 1)
	assert.Equal(t, EventPlanned, e.Kind)
	assert.Equal(t, "Test.1.1", e.ID)
	assert.Equal(t, Test, e.Type)
	assert.Equal(t, protocol.PartOne, e.Part)
	assert.Equal(t, 1, e.SubPart)
	assert.Nil(t, e.Result)
}

func TestStartedEvent(t *testing.T) {
	e := StartedEvent("Solve.2", Solve, protocol.PartTwo, 0)
	assert.Equal(t, EventStarted, e.Kind)
	assert.Equal(t, "Solve.2", e.ID)
	assert.Nil(t, e.Result)
}

func TestFinishedEvent(t *testing.T) {
	r := Result{ID: "Test.1.1", Type: Test, Part: protocol.PartOne, SubPart: 1, Status: StatusPassed, Duration: 0.5}
	e := FinishedEvent(r)
	assert.Equal(t, EventFinished, e.Kind)
	assert.Equal(t, "Test.1.1", e.ID)
	assert.Equal(t, Test, e.Type)
	assert.Equal(t, protocol.PartOne, e.Part)
	assert.Equal(t, 1, e.SubPart)
	require.NotNil(t, e.Result)
	assert.Equal(t, StatusPassed, e.Result.Status)
}

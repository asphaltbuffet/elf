package render

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func TestRunPlainPathEmitsOutputAndReturnsResults(t *testing.T) {
	var buf bytes.Buffer // not a TTY -> plain renderer
	want := []tasks.Result{{
		ID: "Test.1.1", Type: tasks.Test, Part: protocol.PartOne, SubPart: 1,
		Status: tasks.StatusPassed, Duration: 0.071,
	}}

	op := func(cb func(tasks.Event)) ([]tasks.Result, error) {
		cb(tasks.PlannedEvent("Test.1.1", tasks.Test, protocol.PartOne, 1, ""))
		cb(tasks.StartedEvent("Test.1.1", tasks.Test, protocol.PartOne, 1, ""))
		cb(tasks.FinishedEvent(want[0], ""))
		return want, nil
	}

	got, err := Run(
		context.Background(),
		&buf,
		Header{Year: 2015, Day: 4, Title: "Day 4", Language: "Go"},
		false,
		false,
		op,
	)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Contains(t, buf.String(), "PASS", buf.String())
}

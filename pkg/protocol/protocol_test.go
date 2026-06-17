package protocol_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

// Part constants are part of the wire protocol — their numeric values must never change
// without a corresponding change to all runner implementations.
func TestPart_StableValues(t *testing.T) {
	assert.Equal(t, protocol.PartOne, protocol.Part(1))
	assert.Equal(t, protocol.PartTwo, protocol.Part(2))
	assert.Equal(t, protocol.Visualize, protocol.Part(3))
}

func TestTask_JSONRoundTrip(t *testing.T) {
	task := protocol.Task{
		TaskID: "solve.1",
		Part:   protocol.PartOne,
		Input:  "hello world",
	}

	data, err := json.Marshal(task)
	require.NoError(t, err)

	var got protocol.Task
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, task, got)
}

func TestTask_OutputDirOmittedWhenEmpty(t *testing.T) {
	task := protocol.Task{TaskID: "solve.1", Part: protocol.PartOne, Input: "x"}

	data, err := json.Marshal(task)
	require.NoError(t, err)

	assert.NotContains(t, string(data), "output_dir")
}

func TestResult_JSONRoundTrip(t *testing.T) {
	result := protocol.Result{
		TaskID:   "solve.1",
		Ok:       true,
		Output:   "42",
		Duration: 0.005,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var got protocol.Result
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, result, got)
}

package render

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func decodeSummary(t *testing.T, b []byte) jsonSummary {
	t.Helper()
	var s jsonSummary
	require.NoError(t, json.Unmarshal(b, &s))
	return s
}

func TestJSON_SolveSummary(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSON(&buf, Header{Year: 2015, Day: 1, Title: "Not Quite Lisp"})

	r.Handle(tasks.Event{Kind: tasks.EventFinished, Type: tasks.Solve, Language: "Go",
		Result: &tasks.Result{Part: 1, Status: tasks.StatusPassed, Output: "280", Duration: 0.0021}})
	r.Handle(tasks.Event{Kind: tasks.EventFinished, Type: tasks.Solve, Language: "Go",
		Result: &tasks.Result{Part: 2, Status: tasks.StatusUnverified, Output: "1797", Duration: 0.0044}})
	require.NoError(t, r.Close())

	s := decodeSummary(t, buf.Bytes())
	assert.Equal(t, 2015, s.Year)
	assert.Equal(t, 1, s.Day)
	assert.Equal(t, "Not Quite Lisp", s.Title)
	require.Len(t, s.Results, 2)

	assert.Equal(t, 1, s.Results[0].Part)
	assert.Equal(t, "PASS", s.Results[0].Status)
	assert.Equal(t, "280", s.Results[0].Output)
	assert.InDelta(t, 0.0021, s.Results[0].Duration, 1e-9)

	assert.Equal(t, "NEW", s.Results[1].Status)
	assert.Equal(t, "1797", s.Results[1].Output)
}

func TestJSON_OmitsHeaderWhenNoMetadata(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSON(&buf, Header{}) // Year == 0
	r.Handle(tasks.Event{Kind: tasks.EventFinished, Type: tasks.Solve, Language: "Go",
		Result: &tasks.Result{Part: 1, Status: tasks.StatusPassed, Output: "42", Duration: 0.001}})
	require.NoError(t, r.Close())

	// year/day/title use omitempty, so they must be absent from the raw JSON.
	raw := buf.String()
	assert.NotContains(t, raw, `"year"`)
	assert.NotContains(t, raw, `"day"`)
	assert.NotContains(t, raw, `"title"`)

	s := decodeSummary(t, buf.Bytes())
	require.Len(t, s.Results, 1)
	assert.Equal(t, "42", s.Results[0].Output)
}

func TestJSON_FailureIncludesExpected(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSON(&buf, Header{Year: 2020, Day: 2, Title: "X"})
	r.Handle(tasks.Event{
		Kind:     tasks.EventFinished,
		Type:     tasks.Test,
		Language: "Go",
		Result: &tasks.Result{
			Part:     1,
			SubPart:  3,
			Status:   tasks.StatusFailed,
			Output:   "wrong",
			Expected: "right",
			Duration: 0.01,
		},
	})
	require.NoError(t, r.Close())

	s := decodeSummary(t, buf.Bytes())
	require.Len(t, s.Results, 1)
	assert.Equal(t, "FAIL", s.Results[0].Status)
	assert.Equal(t, 3, s.Results[0].SubPart)
	assert.Equal(t, "wrong", s.Results[0].Output)
	assert.Equal(t, "right", s.Results[0].Expected)
}

func TestJSON_BenchmarkAggregatesPerRunnerPart(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSON(&buf, Header{Year: 2015, Day: 1, Title: "X"})

	// Two iterations of Go part 1, one of Go part 2.
	r.Handle(tasks.Event{Kind: tasks.EventFinished, Type: tasks.Benchmark, Language: "Go",
		Result: &tasks.Result{Part: 1, Status: tasks.StatusPassed, Duration: 0.10}})
	r.Handle(tasks.Event{Kind: tasks.EventFinished, Type: tasks.Benchmark, Language: "Go",
		Result: &tasks.Result{Part: 1, Status: tasks.StatusPassed, Duration: 0.20}})
	r.Handle(tasks.Event{Kind: tasks.EventFinished, Type: tasks.Benchmark, Language: "Go",
		Result: &tasks.Result{Part: 2, Status: tasks.StatusPassed, Duration: 0.05}})
	require.NoError(t, r.Close())

	s := decodeSummary(t, buf.Bytes())
	require.Len(t, s.Results, 2, "expected one entry per (runner, part) group")

	assert.Equal(t, 1, s.Results[0].Part)
	assert.Equal(t, "Go", s.Results[0].Runner)
	assert.Equal(t, 2, s.Results[0].Iterations)
	assert.InDelta(t, 0.30, s.Results[0].Duration, 1e-9)

	assert.Equal(t, 2, s.Results[1].Part)
	assert.Equal(t, 1, s.Results[1].Iterations)
	assert.InDelta(t, 0.05, s.Results[1].Duration, 1e-9)
}

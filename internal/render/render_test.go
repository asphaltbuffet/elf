package render

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func TestNew_JSONWhenJSONOut(t *testing.T) {
	var buf bytes.Buffer

	// jsonOut selects the JSON renderer. (plain+jsonOut both true is unreachable
	// at runtime — cobra rejects it — but the factory's guard order is checked
	// here in isolation to document the branch precedence defensively.)
	r := New(&buf, Header{}, true /*plain*/, true /*jsonOut*/)
	_, ok := r.(*JSON)
	assert.True(t, ok, "expected *JSON when jsonOut=true, got %T", r)
}

func TestNew_PlainWhenNotJSON(t *testing.T) {
	var buf bytes.Buffer
	r := New(&buf, Header{}, true /*plain*/, false /*jsonOut*/)
	_, ok := r.(*Plain)
	assert.True(t, ok, "expected *Plain when plain=true jsonOut=false, got %T", r)
}

func TestRun_JSONPath(t *testing.T) {
	var buf bytes.Buffer

	want := []tasks.Result{{Part: 1, Status: tasks.StatusPassed, Output: "7", Duration: 0.002}}

	got, err := Run(context.Background(), &buf, Header{Year: 2015, Day: 1, Title: "X"},
		false /*plain*/, true, /*jsonOut*/
		func(cb func(tasks.Event)) ([]tasks.Result, error) {
			cb(tasks.Event{Kind: tasks.EventFinished, Type: tasks.Solve, Language: "Go", Result: &want[0]})
			return want, nil
		})
	require.NoError(t, err)
	require.Equal(t, want, got)

	var s jsonSummary
	require.NoError(t, json.Unmarshal(buf.Bytes(), &s))
	require.Len(t, s.Results, 1)
	assert.Equal(t, "7", s.Results[0].Output)
	assert.Equal(t, "PASS", s.Results[0].Status)
}

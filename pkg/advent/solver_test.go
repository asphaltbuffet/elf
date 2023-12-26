package advent

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/mocks"
	"github.com/asphaltbuffet/elf/pkg/runners"
)

func Test_makeMainID(t *testing.T) {
	t.Parallel()
	type args struct {
		part runners.Part
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"single digit", args{part: 1}, "main.1"},
		{"double digit", args{part: 25}, "main.25"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, makeMainID(tt.args.part))
		})
	}
}

func Test_parseMainID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want runners.Part
	}{
		{"single digit", args{id: "main.1"}, 1},
		{"two digit", args{id: "main.25"}, 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseMainID(tt.args.id))
		})
	}
}

func TestParseMainIDWithPanic(t *testing.T) {
	type args struct {
		id string
	}

	tests := []struct {
		name string
		args args
	}{
		{"negative", args{id: "main.-1"}},
		{"too big", args{id: "main.9001"}},
		{"not a number", args{id: "main.foo"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() { parseMainID(tt.args.id) })
		})
	}
}

func Test_runMainTasks(t *testing.T) {
	runner := new(mocks.MockRunner)

	mockCall := runner.On("Run", mock.Anything).Return(&runners.Result{
		TaskID:   "main.1",
		Ok:       true,
		Output:   "FAKE OUTPUT",
		Duration: 0.042,
	}, nil)

	err := runMainTasks(runner, &Data{Input: "FAKE INPUT"})

	runner.AssertExpectations(t)
	require.NoError(t, err)

	mockCall.Unset()

	runner.On("Run", mock.Anything).Return(&runners.Result{}, fmt.Errorf("FAKE ERROR"))
	err = runMainTasks(runner, &Data{Input: "FAKE INPUT"})

	runner.AssertExpectations(t)
	require.Error(t, err)
}

func Test_handleMainResult(t *testing.T) {
	type args struct {
		r *runners.Result
	}
	tests := []struct {
		name  string
		args  args
		wantW string
	}{
		{
			name: "sucessful run",
			args: args{
				r: &runners.Result{
					TaskID:   "main.1",
					Ok:       true,
					Output:   "FAKE OUTPUT",
					Duration: 0.042,
				},
			},
			wantW: "  Part 1: FAKE OUTPUT in 42 ms\n",
		},
		{
			name: "not ok",
			args: args{
				r: &runners.Result{
					TaskID:   "main.1",
					Ok:       false,
					Output:   "FAKE ERROR",
					Duration: 0.042,
				},
			},
			wantW: "  Part 1: did not complete saying \"FAKE ERROR\"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			handleTaskResult(w, tt.args.r, "HACK: FIX THIS IN THE TEST INPUT")
			assert.Equal(t, tt.wantW, w.String())
		})
	}
}

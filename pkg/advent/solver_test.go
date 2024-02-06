package advent

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/Runner"
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
	mockRunner := mocks.NewMockRunner(t)

	mockCall := mockRunner.On("Run", mock.Anything).Return(&runners.Result{
		TaskID:   "main.1",
		Ok:       true,
		Output:   "FAKE OUTPUT",
		Duration: 0.042,
	}, nil)

	_, err := runMainTasks(mockRunner, &Data{Input: "FAKE INPUT"})

	require.NoError(t, err)

	mockCall.Unset()

	mockRunner.On("Run", mock.Anything).Return(&runners.Result{}, fmt.Errorf("FAKE ERROR"))
	_, err = runMainTasks(mockRunner, &Data{Input: "FAKE INPUT"})

	require.Error(t, err)
}

func Test_handleMainResult(t *testing.T) {
	type args struct {
		r *runners.Result
	}
	tests := []struct {
		name string
		args args
		want TaskResult
	}{
		{
			name: "sucessful run",
			args: args{
				r: &runners.Result{
					TaskID:   "main.1",
					Ok:       true,
					Output:   "good output",
					Duration: 0.042,
				},
			},
			want: TaskResult{
				ID:       "main.1",
				Type:     TaskMain,
				Part:     1,
				SubPart:  -1,
				Status:   Passed,
				Output:   "good output",
				Expected: "good output",
				Duration: 0.042,
			},
		},
		{
			name: "not ok",
			args: args{
				r: &runners.Result{
					TaskID:   "main.2",
					Ok:       false,
					Output:   "error text",
					Duration: 0.042,
				},
			},
			want: TaskResult{
				ID:       "main.2",
				Type:     TaskMain,
				Part:     2,
				SubPart:  -1,
				Status:   Error,
				Output:   "⤷ saying:error text",
				Expected: "",
				Duration: 0.042,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			got := handleTaskResult(w, tt.args.r, "good output")
			assert.Equal(t, tt.want, got)
		})
	}
}

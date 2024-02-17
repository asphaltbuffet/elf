package advent

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/runners"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func Test_runMainTasks(t *testing.T) {
	mockRunner := mocks.NewMockRunner(t)
	mockCall := mockRunner.EXPECT().Run(mock.Anything).Return(&runners.Result{
		TaskID:   "solve.1",
		Ok:       true,
		Output:   "FAKE OUTPUT",
		Duration: 0.042,
	}, nil).Times(2)

	_, err := runMainTasks(mockRunner, &Data{InputData: "FAKE INPUT"})

	require.NoError(t, err)

	mockCall.Unset()

	mockRunner.EXPECT().Run(mock.Anything).Return(&runners.Result{
		TaskID:   "fake.1",
		Ok:       false,
		Output:   "fakey fake",
		Duration: 0.666,
	}, fmt.Errorf("FAKE ERROR")).Once()
	_, err = runMainTasks(mockRunner, &Data{InputData: "FAKE INPUT"})

	require.Error(t, err)
}

func Test_handleMainResult(t *testing.T) {
	type args struct {
		r *runners.Result
	}
	tests := []struct {
		name string
		args args
		want tasks.Result
	}{
		{
			name: "sucessful run",
			args: args{
				r: &runners.Result{
					TaskID:   "solve.1",
					Ok:       true,
					Output:   "good output",
					Duration: 0.042,
				},
			},
			want: tasks.Result{
				ID:       "solve.1",
				Type:     tasks.Solve,
				Part:     1,
				SubPart:  0,
				Status:   tasks.StatusPassed,
				Output:   "good output",
				Expected: "good output",
				Duration: 0.042,
			},
		},
		{
			name: "not ok",
			args: args{
				r: &runners.Result{
					TaskID:   "solve.2",
					Ok:       false,
					Output:   "error text",
					Duration: 0.042,
				},
			},
			want: tasks.Result{
				ID:       "solve.2",
				Type:     tasks.Solve,
				Part:     2,
				SubPart:  0,
				Status:   tasks.StatusError,
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

package exercise

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func Test_handleMainResult(t *testing.T) {
	type args struct {
		r *protocol.Result
	}

	tests := []struct {
		name string
		args args
		want tasks.Result
	}{
		{
			name: "sucessful run",
			args: args{
				r: &protocol.Result{
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
				r: &protocol.Result{
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
			got := handleTaskResult(tt.args.r, "good output")

			assert.Equal(t, tt.want, got)
		})
	}
}

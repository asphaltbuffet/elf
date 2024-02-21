package tasks_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func Test_MakeTaskID(t *testing.T) {
	type args struct {
		task    tasks.TaskType
		part    runners.Part
		subPart []int
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "negative subpart",
			args: args{tasks.Test, runners.PartOne, []int{-1}},
			want: "test.1.-1",
		},
		{
			name: "positive subpart",
			args: args{tasks.Test, runners.PartTwo, []int{1}},
			want: "test.2.1",
		},
		{
			name: "solve part one",
			args: args{tasks.Solve, runners.PartTwo, nil},
			want: "solve.2",
		},
		{
			name: "solve part two",
			args: args{tasks.Solve, runners.PartTwo, nil},
			want: "solve.2",
		},
		{
			name: "visualize",
			args: args{tasks.Visualize, runners.PartOne, []int{25}},
			want: "visualize.1.25",
		},
		{
			name: "Benchmark with subpart",
			args: args{tasks.Benchmark, runners.PartOne, []int{-1}},
			want: "benchmark.1.-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.subPart == nil {
				assert.Equal(t, tt.want, tasks.MakeTaskID(tt.args.task, tt.args.part))
			} else {
				assert.Equal(t, tt.want, tasks.MakeTaskID(tt.args.task, tt.args.part, tt.args.subPart...))
			}
		})
	}
}

func Test_MakeTaskID_WithPanic(t *testing.T) {
	type args struct {
		task    tasks.TaskType
		part    runners.Part
		subPart []int
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Solve with subpart",
			args: args{tasks.Solve, runners.PartOne, []int{-1}},
		},
		{
			name: "Test with no subpart",
			args: args{tasks.Test, runners.PartOne, []int{}},
		},
		{
			name: "Test with many subparts",
			args: args{tasks.Test, runners.PartOne, []int{1, 2, 3}},
		},
		{
			name: "invalid task type",
			args: args{42, runners.PartOne, []int{-1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				tasks.MakeTaskID(tt.args.task, tt.args.part, tt.args.subPart...)
			})
		})
	}
}

func Test_ParseTaskID(t *testing.T) {
	type args struct {
		id string
	}

	type wants struct {
		taskType tasks.TaskType
		part     runners.Part
		subPart  int
	}

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{"valid test", args{id: "test.1.2"}, wants{tasks.Test, runners.PartOne, 2}},
		{"valid benchmark", args{id: "benchmark.1.2"}, wants{tasks.Benchmark, 1, 2}},
		{"valid test", args{id: "solve.2"}, wants{tasks.Solve, runners.PartTwo, 0}},
		{"invalid short", args{id: "visualize.2"}, wants{tasks.Invalid, 0, 0}},
		{"invalid type", args{id: "fake.2.1"}, wants{tasks.Invalid, 0, 0}},
		{"no parts", args{id: "test"}, wants{tasks.Invalid, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotPart, gotSubPart := tasks.ParseTaskID(tt.args.id)

			assert.Equal(t, tt.wants.taskType.String(), gotType.String()) // compare strings to help debug failures
			assert.Equal(t, tt.wants.part, gotPart, "part")
			assert.Equal(t, tt.wants.subPart, gotSubPart, "subpart")
		})
	}
}

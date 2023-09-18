package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func TestGetBenchmarkCmd(t *testing.T) {
	got := GetBenchmarkCmd()

	checkCommand(t, got, "benchmark")
}

func Test_makeBenchmarkID(t *testing.T) {
	type args struct {
		part    runners.Part
		subPart int
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"part one, first test", args{runners.PartOne, 0}, "benchmark.part.1.0"},
		{"part two, first test", args{runners.PartTwo, 0}, "benchmark.part.2.0"},
		{"part one, default test", args{runners.PartOne, 30}, "benchmark.part.1.30"},
		{"part two, default test", args{runners.PartTwo, 30}, "benchmark.part.2.30"},
		{"part one, no num", args{runners.PartOne, -1}, "benchmark.part.1"},
		{"part one, no num", args{runners.PartTwo, -1}, "benchmark.part.2"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, makeBenchmarkID(tt.args.part, tt.args.subPart))
		})
	}
}

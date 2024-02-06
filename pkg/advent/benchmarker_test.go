package advent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

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
		{
			name: "negative subpart",
			args: args{
				part:    runners.PartOne,
				subPart: -1,
			},
			want: "benchmark.1.-1",
		},
		{
			name: "part and subpart",
			args: args{
				part:    runners.PartTwo,
				subPart: 1,
			},
			want: "benchmark.2.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, makeBenchmarkID(tt.args.part, tt.args.subPart))
		})
	}
}

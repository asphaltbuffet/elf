package advent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func Test_makeTestID(t *testing.T) {
	type args struct {
		part runners.Part
		n    int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				part: runners.PartOne,
				n:    1,
			},
			want: "test.1.1",
		},
		{
			name: "empty",
			args: args{},
			want: "test.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, makeTestID(tt.args.part, tt.args.n))
		})
	}
}

func Test_parseTestID(t *testing.T) {
	type args struct {
		id string
	}

	type wants struct {
		part runners.Part
		n    int
	}

	tests := []struct {
		name string
		args args
		want wants
	}{
		{"success", args{id: "test.1.1"}, wants{part: runners.PartOne, n: 1}},
		{"part 2", args{id: "test.2.23"}, wants{part: runners.PartTwo, n: 23}},
		{"part 3", args{id: "test.3.23"}, wants{part: 3, n: 23}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPart, gotN := parseTestID(tt.args.id)

			assert.Equal(t, tt.want.part, gotPart)
			assert.Equal(t, tt.want.n, gotN)
		})
	}
}

func TestParseTestIDWithPanic(t *testing.T) {
	type args struct {
		id string
	}

	tests := []struct {
		name string
		args args
	}{
		{"negative", args{id: "test.-1.1"}},
		{"too big", args{id: "test.9001.1"}},
		{"not a part number", args{id: "test.foo.1"}},
		{"not a test number", args{id: "test.1.foo"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() { parseTestID(tt.args.id) })
		})
	}
}

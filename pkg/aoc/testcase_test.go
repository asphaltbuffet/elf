package aoc

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
			name: "valid 1-0",
			args: args{part: 1, n: 0},
			want: "test.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeTestID(tt.args.part, tt.args.n)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseTestID(t *testing.T) {
	type args struct {
		id string
	}

	tests := []struct {
		name     string
		args     args
		wantPart runners.Part
		wantSub  int
	}{
		{
			name:     "valid 1-0",
			args:     args{id: "test.1.0"},
			wantPart: 1,
			wantSub:  0,
		},
		{
			name:     "valid 2-4",
			args:     args{id: "test.2.4"},
			wantPart: 2,
			wantSub:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part, sub := parseTestID(tt.args.id)

			assert.Equal(t, tt.wantPart, part)
			assert.Equal(t, tt.wantSub, sub)
		})
	}
}

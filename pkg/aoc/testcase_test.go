package aoc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/asphaltbuffet/elf/mocks"
	"github.com/asphaltbuffet/elf/pkg/exercise"
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

func Test_runTests(t *testing.T) {
	m := new(mocks.MockRunner)

	type args struct {
		exInfo *exercise.Info
	}

	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "run error - part one",
			args: args{
				exInfo: &exercise.Info{
					InputFile: "fake.txt",
					TestCases: exercise.TestCase{
						One: []*exercise.Test{{Input: "fake", Expected: "FAKE"}},
						Two: []*exercise.Test{{Input: "fake", Expected: "FAKE"}},
					},
				},
			},
			assertion: assert.Error,
		},
		{
			name: "run error - part two",
			args: args{
				exInfo: &exercise.Info{
					InputFile: "fake.txt",
					TestCases: exercise.TestCase{
						One: []*exercise.Test{{Input: "", Expected: ""}},
						Two: []*exercise.Test{{Input: "fake", Expected: "FAKE"}},
					},
				},
			},
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.On("Run", mock.Anything).Return(&runners.Result{}, errors.New("mock error"))

			tt.assertion(t, runTests(m, tt.args.exInfo))
			m.AssertExpectations(t)
		})
	}
}

package advent

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/mocks"
	"github.com/asphaltbuffet/elf/pkg/runners"
)

func Test_Test(t *testing.T) {
	mockRunner := new(mocks.MockRunner)
	mockRunner.On("String").Return("MOCK")
	mockRunner.On("Stop").Return(nil)
	mockRunner.On("Cleanup").Return(nil)

	tests := []struct {
		name      string
		e         *Exercise
		mock1     func() *mock.Call
		mock2     func() *mock.Call
		assertion require.ErrorAssertionFunc
	}{
		{
			name:      "error on empty exercise",
			e:         &Exercise{},
			mock1:     func() *mock.Call { return mockRunner.On("Start").Return(t.Fatal) },
			mock2:     func() *mock.Call { return mockRunner.On("Run").Return(t.Fatal) },
			assertion: require.Error,
		},
		{
			name: "start error",
			e: &Exercise{
				ID:       "2015-01",
				Title:    "Fake Title",
				Language: "go",
				Year:     2015,
				Day:      1,
				URL:      "https://fake.url/",
				Data:     &Data{},
				Path:     "/fake/path/here",
				runner:   mockRunner,
			},
			mock1:     func() *mock.Call { return mockRunner.On("Start").Return(fmt.Errorf("fake start error")) },
			mock2:     func() *mock.Call { return mockRunner.On("Run").Return(t.Fatal) },
			assertion: require.Error,
		},
		{
			name: "run error",
			e: &Exercise{
				ID:       "2015-01",
				Title:    "Fake Title",
				Language: "go",
				Year:     2015,
				Day:      1,
				URL:      "https://fake.url/",
				Data: &Data{
					Input:     "FAKE\nINPUT",
					InputFile: "input.txt",
					TestCases: TestCase{
						One: []*Test{{Input: "", Expected: ""}},
						Two: []*Test{{Input: "", Expected: ""}},
					},
				},
				Path:   "/fake/path/here",
				runner: mockRunner,
			},
			mock1: func() *mock.Call { return mockRunner.On("Start").Return(nil) },
			mock2: func() *mock.Call {
				return mockRunner.On("Run", &runners.Task{
					TaskID:    "test.1.1",
					Part:      1,
					Input:     "",
					OutputDir: "",
				}).Return(&runners.Result{TaskID: "test.1.1"}, fmt.Errorf("fake run error"))
			},
			assertion: require.NoError,
		},
		{
			name: "success",
			e: &Exercise{
				ID:       "2015-01",
				Title:    "Fake Title",
				Language: "go",
				Year:     2015,
				Day:      1,
				URL:      "https://fake.url/",
				Data: &Data{
					Input:     "FAKE\nINPUT",
					InputFile: "input.txt",
					TestCases: TestCase{
						One: []*Test{{Input: "fake", Expected: "fake"}},
						Two: []*Test{{Input: "fake", Expected: "fake"}},
					},
				},
				Path:   "/fake/path/here",
				runner: mockRunner,
			},
			mock1: func() *mock.Call { return mockRunner.On("Start").Return(nil) },
			mock2: func() *mock.Call {
				return mockRunner.On("Run", mock.Anything).Return(&runners.Result{
					TaskID:   "test.1.1",
					Ok:       true,
					Output:   "fake",
					Duration: 0.06942,
				}, nil)
			},
			assertion: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m1 := tt.mock1()
			m2 := tt.mock2()

			tt.assertion(t, tt.e.Test())

			// reset mock calls
			m1.Unset()
			m2.Unset()
		})
	}
}

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

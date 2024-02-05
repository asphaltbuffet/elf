package advent

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/Runner"
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

func Test_runTestTasks(t *testing.T) {
	mockRunner := mocks.NewMockRunner(t)

	mockRunner.EXPECT().Run(mock.Anything).Return(&runners.Result{
		TaskID:   "test.1.1",
		Ok:       true,
		Output:   "FAKE OUTPUT",
		Duration: 0.042,
	}, nil)

	_, err := runTests(mockRunner, &Data{
		Input: "FAKE INPUT",
		TestCases: TestCase{
			One: []*Test{
				{
					Input:    "FAKE INPUT",
					Expected: "FAKE OUTPUT",
				},
			},
			Two: []*Test{
				{
					Input:    "FAKE INPUT",
					Expected: "FAKE OUTPUT",
				},
			},
		},
		Answers: Answer{},
	})

	require.NoError(t, err)
}

func Test_handleTestResult(t *testing.T) {
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
					TaskID:   "test.1.1",
					Ok:       true,
					Output:   "good output",
					Duration: 0.042,
				},
			},
			want: TaskResult{
				ID:       "test.1.1",
				Type:     TaskTest,
				Part:     1,
				SubPart:  1,
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
					TaskID:   "test.1.2",
					Ok:       false,
					Output:   "error text",
					Duration: 0.042,
				},
			},
			want: TaskResult{
				ID:       "test.1.2",
				Type:     TaskTest,
				Part:     1,
				SubPart:  2,
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

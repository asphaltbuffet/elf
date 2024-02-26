package advent

import (
	"fmt"
	"io"
	"log/slog"
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

	e := &Exercise{
		runner: mockRunner,
		Data:   &Data{InputData: "FAKE INPUT"},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		writer: io.Discard,
	}

	_, err := e.runMainTasks()

	require.NoError(t, err)

	mockCall.Unset()

	mockRunner.EXPECT().Run(mock.Anything).Return(&runners.Result{
		TaskID:   "fake.1",
		Ok:       false,
		Output:   "fakey fake",
		Duration: 0.666,
	}, fmt.Errorf("FAKE ERROR")).Once()

	_, err = e.runMainTasks()

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
			got := handleTaskResult(io.Discard, tt.args.r, "good output")

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSolve(t *testing.T) {
	type fields struct {
		inputFile string
	}

	type args struct {
		skipTests bool
	}

	tests := []struct {
		name      string
		setup     func(*mocks.MockRunner)
		fields    fields
		args      args
		want      []tasks.Result
		assertion require.ErrorAssertionFunc
		wantErr   error
	}{
		{
			name:  "missing input file",
			setup: func(_m *mocks.MockRunner) {},
			fields: fields{
				inputFile: "missingInput.fake",
			},
			args: args{
				skipTests: false,
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "runner start error",
			setup: func(_m *mocks.MockRunner) {
				_m.EXPECT().Start().Return(fmt.Errorf("FAKE ERROR"))
			},
			fields: fields{
				inputFile: "input.fake",
			},
			args: args{
				skipTests: false,
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "runner run error",
			setup: func(_m *mocks.MockRunner) {
				_m.EXPECT().Start().Return(nil)
				_m.EXPECT().Run(mock.Anything).Return(nil, fmt.Errorf("FAKE ERROR"))
				_m.EXPECT().String().Return("fakeRunner")
				_m.EXPECT().Stop().Return(nil)
				_m.EXPECT().Cleanup().Return(nil)
			},
			fields: fields{
				inputFile: "input.fake",
			},
			args: args{
				skipTests: false,
			},
			want:      nil,
			assertion: require.Error,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			f, err := testFs.Create("input.fake")
			require.NoError(t, err)
			_, err = f.WriteString("fake input data")
			require.NoError(t, err)
			f.Close()

			// set up mocks
			mockRunner := mocks.NewMockRunner(t)
			tt.setup(mockRunner)

			e := &Exercise{
				ID:       "1111-22",
				Title:    "Fake Title",
				Language: "fakeLang",
				Year:     1111,
				Day:      22,
				Data: &Data{
					InputData:     "",
					InputFileName: tt.fields.inputFile,
					TestCases: TestCase{
						One: []*Test{
							{Input: "fake test 1.1", Expected: "fake result 1.1"},
							{Input: "fake test 1.2", Expected: "fake result 1.2"},
						},
						Two: []*Test{
							{Input: "fake test 2.1", Expected: "fake result 2.1"},
							{Input: "fake test 2.2", Expected: "fake result 2.2"},
						},
					},
					Answers: Answer{},
				},
				Path:   "",
				runner: mockRunner,
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				appFs:  testFs,
				writer: io.Discard,
			}

			// execute the function under test
			// skipTests == false
			got, err := e.Solve(false)

			// verify results
			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}

			// skipTests == true
			got, err = e.Solve(true)

			// verify results
			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

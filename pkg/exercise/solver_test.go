package exercise

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/runners"
	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func Test_runMainTasks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockRunner := mocks.NewMockRunner(t)
	mockCall := mockRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID:   "solve.1",
		Ok:       true,
		Output:   "FAKE OUTPUT",
		Duration: 0.042,
	}, nil).Times(2)

	e := &Exercise{
		Data: &Data{InputData: "FAKE INPUT"},
	}

	_, err := e.runMainTasks(t.Context(), mockRunner, io.Discard, nil)

	require.NoError(t, err)

	mockCall.Unset()

	mockRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID:   "fake.1",
		Ok:       false,
		Output:   "fakey fake",
		Duration: 0.666,
	}, errors.New("FAKE ERROR")).Once()

	_, err = e.runMainTasks(t.Context(), mockRunner, io.Discard, nil)

	require.Error(t, err)

	_ = logger // used in TestSolve
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
			setup: func(_ *mocks.MockRunner) {},
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
				_m.EXPECT().Prepare(mock.Anything).Return(errors.New("FAKE ERROR"))
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
				_m.EXPECT().Prepare(mock.Anything).Return(nil)
				_m.EXPECT().Open(mock.Anything).Return(nil)
				_m.EXPECT().Run(mock.Anything, mock.Anything).Return(nil, errors.New("FAKE ERROR"))
				_m.EXPECT().String().Return("fakeRunner")
				_m.EXPECT().Close(mock.Anything).Return(nil)
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

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			f, err := testFs.Create("input.fake")
			require.NoError(t, err)
			_, err = f.WriteString("fake input data")
			require.NoError(t, err)
			f.Close()

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
				Path: "",
			}

			// skipTests == false
			got, err := e.Solve(t.Context(), testFs, logger, mockRunner, io.Discard, nil, false)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}

			// skipTests == true
			got, err = e.Solve(t.Context(), testFs, logger, mockRunner, io.Discard, nil, true)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

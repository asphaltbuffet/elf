package advent

import (
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/spf13/afero"
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

	_, err := runMainTasks(mockRunner, &Data{InputData: "FAKE INPUT"})

	require.NoError(t, err)

	mockCall.Unset()

	mockRunner.EXPECT().Run(mock.Anything).Return(&runners.Result{
		TaskID:   "fake.1",
		Ok:       false,
		Output:   "fakey fake",
		Duration: 0.666,
	}, fmt.Errorf("FAKE ERROR")).Once()

	_, err = runMainTasks(mockRunner, &Data{InputData: "FAKE INPUT"})

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

func TestExercise_SolveMissingInput(t *testing.T) {
	base := afero.NewBasePathFs(afero.NewOsFs(), "testdata")
	roBase = afero.NewReadOnlyFs(base)

	testFs = afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())

	e := &Exercise{
		ID:       "1111-22",
		Title:    "Fake Title",
		Language: "fakeLang",
		Year:     1111,
		Day:      22,
		Data: &Data{
			InputData:     "",
			InputFileName: "missingInput.txt",
			TestCases:     TestCase{},
			Answers:       Answer{},
		},
		Path:   "",
		runner: nil,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		appFs:  testFs,
	}

	// execute the function under test
	_, err := e.Solve(false)

	require.Error(t, err)
}

func TestExercise_SolveRunnerStartError(t *testing.T) {
	base := afero.NewBasePathFs(afero.NewOsFs(), "testdata")
	roBase = afero.NewReadOnlyFs(base)

	testFs = afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())
	f, err := testFs.Create("input.fake")
	require.NoError(t, err)
	_, err = f.WriteString("fake input data")
	require.NoError(t, err)
	f.Close()

	// set up mock runner
	mockRunner := mocks.NewMockRunner(t)
	mockRunner.EXPECT().Start().Return(fmt.Errorf("FAKE ERROR")).Once()

	e := &Exercise{
		ID:       "1111-22",
		Title:    "Fake Title",
		Language: "fakeLang",
		Year:     1111,
		Day:      22,
		Data: &Data{
			InputData:     "",
			InputFileName: "input.fake",
			TestCases:     TestCase{},
			Answers:       Answer{},
		},
		Path:   "",
		runner: mockRunner,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		appFs:  testFs,
	}

	// execute the function under test
	_, err = e.Solve(false)

	require.Error(t, err)
}

func TestExercise_SolveRunnerRunError(t *testing.T) {
	base := afero.NewBasePathFs(afero.NewOsFs(), "testdata")
	roBase = afero.NewReadOnlyFs(base)

	testFs = afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())
	f, err := testFs.Create("input.fake")
	require.NoError(t, err)
	_, err = f.WriteString("fake input data")
	require.NoError(t, err)
	f.Close()

	// set up mock runner
	mockRunner := mocks.NewMockRunner(t)
	mockRunner.EXPECT().Start().Return(nil).Once()
	mockRunner.EXPECT().Run(mock.Anything).Return(nil, fmt.Errorf("FAKE ERROR"))
	mockRunner.EXPECT().String().Return("fakeRunner")
	mockRunner.EXPECT().Stop().Return(nil).Once()
	mockRunner.EXPECT().Cleanup().Return(nil).Once()

	e := &Exercise{
		ID:       "1111-22",
		Title:    "Fake Title",
		Language: "fakeLang",
		Year:     1111,
		Day:      22,
		Data: &Data{
			InputData:     "",
			InputFileName: "input.fake",
			TestCases:     TestCase{},
			Answers:       Answer{},
		},
		Path:   "",
		runner: mockRunner,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		appFs:  testFs,
	}

	// execute the function under test
	_, err = e.Solve(true)

	require.Error(t, err)
}

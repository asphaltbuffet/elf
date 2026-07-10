package exercise

import (
	"context"
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

func Test_Test(t *testing.T) {
	type fields struct {
		data *Data
	}

	tests := []struct {
		name      string
		setup     func(*mocks.MockRunner)
		fields    fields
		want      []tasks.Result
		assertion require.ErrorAssertionFunc
		wantErr   error
	}{
		{
			name: "runner start error",
			setup: func(_m *mocks.MockRunner) {
				_m.EXPECT().Prepare(mock.Anything).Return(errors.New("FAKE ERROR"))
				_m.EXPECT().Close(mock.Anything).Return(nil).Maybe()
				_m.EXPECT().Cleanup().Return(nil).Maybe()
			},
			fields:    fields{},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "runner run error",
			setup: func(_m *mocks.MockRunner) {
				_m.EXPECT().Prepare(mock.Anything).Return(nil)
				_m.EXPECT().Open(mock.Anything).Return(nil)
				_m.EXPECT().Run(mock.Anything, mock.Anything).Return(nil, errors.New("FAKE ERROR"))
				_m.EXPECT().Close(mock.Anything).Return(nil)
				_m.EXPECT().Cleanup().Return(nil)
			},
			fields: fields{
				data: &Data{
					InputData:     "",
					InputFileName: "input.fake",
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

			mockRunner := mocks.NewMockRunner(t)
			mockRunner.EXPECT().String().Return("MOCK").Maybe()
			tt.setup(mockRunner)

			e := &Exercise{
				ID:       "1111-22",
				Title:    "Fake Title",
				Language: "fakeLang",
				Year:     1111,
				Day:      22,
				Data:     tt.fields.data,
				Path:     "",
			}

			got, err := e.Test(t.Context(), logger, mockRunner, nil)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExercise_Test_EmitsLifecycleEvents(t *testing.T) {
	mockRunner := mocks.NewMockRunner(t)
	mockRunner.EXPECT().String().Return("MOCK").Maybe()
	mockRunner.EXPECT().Prepare(mock.Anything).Return(nil)
	mockRunner.EXPECT().Open(mock.Anything).Return(nil)
	mockRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID:   "test.1.0",
		Ok:       true,
		Output:   "FAKE OUTPUT",
		Duration: 0.042,
	}, nil)
	mockRunner.EXPECT().Close(mock.Anything).Return(nil)
	mockRunner.EXPECT().Cleanup().Return(nil)

	e := &Exercise{
		ID:       "1111-22",
		Title:    "Fake Title",
		Language: "fakeLang",
		Year:     1111,
		Day:      22,
		Data: &Data{
			InputData: "FAKE INPUT",
			TestCases: TestCase{
				One: []*Test{
					{Input: "FAKE INPUT", Expected: "FAKE OUTPUT"},
				},
			},
			Answers: Answer{},
		},
		Path: "",
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	var events []tasks.Event
	cb := func(ev tasks.Event) { events = append(events, ev) }

	_, err := e.Test(context.Background(), logger, mockRunner, cb)
	require.NoError(t, err)

	// At least one Planned precedes the first Finished.
	var sawPlanned, sawFinished bool
	for _, ev := range events {
		switch ev.Kind {
		case tasks.EventMeta:
			require.NotNil(t, ev.Meta)
			assert.False(t, sawPlanned, "Meta must precede task events")
		case tasks.EventPlanned:
			sawPlanned = true
			assert.False(t, sawFinished, "Planned must precede Finished")
		case tasks.EventStarted:
			// no assertion needed; just consume to satisfy exhaustive
		case tasks.EventFinished:
			sawFinished = true
			require.NotNil(t, ev.Result)
		}
	}
	assert.True(t, sawPlanned)
	assert.True(t, sawFinished)
}

func Test_runTests(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockRunner := mocks.NewMockRunner(t)

	mockRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID:   "test.1.1",
		Ok:       true,
		Output:   "FAKE OUTPUT",
		Duration: 0.042,
	}, nil)

	e := &Exercise{
		Data: &Data{
			InputData: "FAKE INPUT",
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
		},
	}

	_, err := e.runTests(t.Context(), mockRunner, nil)

	require.NoError(t, err)

	_ = logger
}

func TestRunTests_ProblemRunsOnePartTestOnly(t *testing.T) {
	mockRunner := mocks.NewMockRunner(t)
	mockRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID:   "test.1.1",
		Ok:       true,
		Output:   "FAKE OUTPUT",
		Duration: 0.042,
	}, nil).Once()

	e := &Exercise{
		Kind: KindProblem,
		Data: &Data{
			TestCases: TestCase{
				One: []*Test{{Input: "10", Expected: "23"}},
				Two: nil,
			},
		},
	}

	results, err := e.runTests(t.Context(), mockRunner, nil)

	require.NoError(t, err)
	assert.Len(t, results, 1)

	for _, r := range results {
		assert.NotEqual(t, protocol.PartTwo, r.Part, "problem must not emit Part Two test tasks")
	}
}

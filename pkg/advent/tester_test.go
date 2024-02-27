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
				_m.EXPECT().Start().Return(fmt.Errorf("FAKE ERROR"))
			},
			fields:    fields{},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "runner run error",
			setup: func(_m *mocks.MockRunner) {
				_m.EXPECT().Start().Return(nil)
				_m.EXPECT().Run(mock.Anything).Return(nil, fmt.Errorf("FAKE ERROR"))
				_m.EXPECT().Stop().Return(nil)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			// set up mocks
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
				runner:   mockRunner,
				logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
				appFs:    testFs,
				writer:   io.Discard,
			}

			// execute the function under test
			got, err := e.Test()

			// verify results
			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_runTests(t *testing.T) {
	mockRunner := mocks.NewMockRunner(t)

	mockRunner.EXPECT().Run(mock.Anything).Return(&runners.Result{
		TaskID:   "test.1.1",
		Ok:       true,
		Output:   "FAKE OUTPUT",
		Duration: 0.042,
	}, nil)

	e := &Exercise{
		runner: mockRunner,
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
		writer: io.Discard,
	}

	_, err := e.runTests()

	require.NoError(t, err)
}

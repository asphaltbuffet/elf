package exercise

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadExerciseInfo(t *testing.T) {
	goodInfo := []byte(
		`{
"inputFile": "fake.txt",
"testCases": {
	"one": [
		{
			"input": "abcd",
			"expected": "test 1"
		}
	],
	"two": [
		{
			"input": "1234",
			"expected": "test 2"
		}
	]
}
}`)

	testFs = afero.NewMemMapFs()

	require.NoError(t, afero.WriteFile(testFs, "not_json_info.json", []byte("not json"), 0o600))
	require.NoError(t, afero.WriteFile(testFs, "bad_json_info.json", []byte("{}"), 0o600))
	require.NoError(t, afero.WriteFile(testFs, "info.json", goodInfo, 0o600))

	type args struct {
		fname string
	}

	tests := []struct {
		name      string
		args      args
		want      *Info
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "no file",
			args: args{
				fname: "fakey_fake.json",
			},
			want:      nil,
			assertion: assert.Error,
			errText:   "file does not exist",
		},
		{
			name: "info file isn't json",
			args: args{
				fname: "not_json_info.json",
			},
			want:      nil,
			assertion: assert.Error,
			errText:   "unmarshal info file",
		},
		{
			name: "incomplete json",
			args: args{
				fname: "bad_json_info.json",
			},
			want:      &Info{},
			assertion: assert.NoError,
		},
		{
			name: "good json with data",
			args: args{
				fname: "info.json",
			},
			want: &Info{
				InputFile: "fake.txt",
				TestCases: TestCase{
					One: []*Test{{Input: "abcd", Expected: "test 1"}},
					Two: []*Test{{Input: "1234", Expected: "test 2"}},
				},
			},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadExerciseInfo(testFs, tt.args.fname)

			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
}

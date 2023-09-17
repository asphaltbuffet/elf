package cmd

import (
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGraphCmd(t *testing.T) {
	got := GetGraphCmd()

	checkCommand(t, got, "graph")
}

func Test_readBenchmarkFile(t *testing.T) {
	type args struct {
		path    string
		content string
	}

	tests := []struct {
		name      string
		args      args
		want      *BenchmarkData
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "valid benchmark file",
			args: args{
				path: "benchmark.json",
				content: `{
					"date": "2023-08-30T21:05:47Z",
					"dir": "testdata/2015/01-testDayOne",
					"year": "2015",
					"day": 1,
					"numRuns": 30,
					"implementations": [
					  {
						"name": "Golang",
						"part-one": {
						  "mean": 0.123,
						  "median": 0.123,
						  "min": 0.12,
						  "max": 0.23
						},
						"part-two": {
						  "mean": 0.567,
						  "median": 0.567,
						  "min": 0.56,
						  "max": 0.67
						}
					  }
					]
				  }`,
			},
			want: &BenchmarkData{
				Date: time.Date(2023, 8, 30, 21, 5, 47, 0, time.UTC),
				Dir:  "testdata/2015/01-testDayOne",
				Year: "2015",
				Day:  1,
				Runs: 30,
				Implementations: []*ImplementationData{
					{
						Name: "Golang",
						PartOne: &PartData{
							Mean:   0.123,
							Median: 0.123,
							Min:    0.12,
							Max:    0.23,
						},
						PartTwo: &PartData{
							Mean:   0.567,
							Median: 0.567,
							Min:    0.56,
							Max:    0.67,
						},
					},
				},
			},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			require.NoError(t, afero.WriteFile(fs, tt.args.path, []byte(tt.args.content), 0o600))

			got, err := readBenchmarkFile(fs, tt.args.path)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

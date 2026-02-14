package exercise

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func TestBenchmarkDataString(t *testing.T) {
	type fields struct {
		Date            time.Time
		Title           string
		Year            int
		Day             int
		Runs            int
		Normalization   float64
		Implementations []*ImplementationData
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "no impl data",
			fields: fields{
				Date:            time.Date(2021, time.December, 25, 12, 34, 56, 0, time.UTC),
				Title:           "fake title",
				Year:            2015,
				Day:             2,
				Runs:            69,
				Normalization:   0.42,
				Implementations: []*ImplementationData{},
			},
			want: "BenchmarkData{Date: 2021-12-25, AOC 2015/02, Runs:  69, Normalization: 0.420000, Implementations: []}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BenchmarkData{
				Date:            tt.fields.Date,
				Title:           tt.fields.Title,
				Year:            tt.fields.Year,
				Day:             tt.fields.Day,
				Runs:            tt.fields.Runs,
				Normalization:   tt.fields.Normalization,
				Implementations: tt.fields.Implementations,
			}
			assert.Equal(t, tt.want, b.String())
		})
	}
}

func Test_calculateMetrics(t *testing.T) {
	type args struct {
		results map[runners.Part][]float64
	}

	tests := []struct {
		name      string
		args      args
		want      map[runners.Part]*PartData
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "empty results",
			args: args{
				results: map[runners.Part][]float64{},
			},
			want:      map[runners.Part]*PartData{},
			assertion: require.NoError,
		},
		{
			name: "one result",
			args: args{
				results: map[runners.Part][]float64{ //nolint:exhaustive // not testing visualize
					runners.PartOne: {1.0},
					runners.PartTwo: {2.0},
				},
			},
			want: map[runners.Part]*PartData{ //nolint:exhaustive // not testing visualize
				runners.PartOne: {Mean: 1.0, Min: 1.0, Max: 1.0, Data: []float64{1.0}},
				runners.PartTwo: {Mean: 2.0, Min: 2.0, Max: 2.0, Data: []float64{2.0}},
			},
			assertion: require.NoError,
		},
		{
			name: "multiple results",
			args: args{
				results: map[runners.Part][]float64{ //nolint:exhaustive // not testing visualize
					runners.PartOne: {1.0, 2.0, 3.0},
					runners.PartTwo: {2.0, 3.0, 4.0},
				},
			},
			want: map[runners.Part]*PartData{ //nolint:exhaustive // not testing visualize
				runners.PartOne: {Mean: 2.0, Min: 1.0, Max: 3.0, Data: []float64{1.0, 2.0, 3.0}},
				runners.PartTwo: {Mean: 3.0, Min: 2.0, Max: 4.0, Data: []float64{2.0, 3.0, 4.0}},
			},
			assertion: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateMetrics(tt.args.results)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

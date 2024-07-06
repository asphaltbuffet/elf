package analyze_test

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/krampus"
	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/advent/analyze"
)

func Test_NewAnalyzer(t *testing.T) {
	type args struct {
		opts []func(*analyze.Analyzer)
	}

	tests := []struct {
		name      string
		setup     func(*mocks.MockExerciseConfiguration)
		args      args
		want      *analyze.Analyzer
		assertion require.ErrorAssertionFunc
	}{
		{
			name:  "no dir",
			setup: func(_ *mocks.MockExerciseConfiguration) {},
			args: args{
				opts: []func(*analyze.Analyzer){},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "with directory",
			setup: func(_ *mocks.MockExerciseConfiguration) {
				// _m.EXPECT().GetFs().Return(afero.NewMemMapFs())
			},
			args: args{
				opts: []func(*analyze.Analyzer){
					analyze.WithDirectory("foo/bar"),
				},
			},
			want: &analyze.Analyzer{
				Data:      []*advent.BenchmarkData{},
				Dir:       "foo/bar",
				GraphType: 1,
				Output:    "",
			},
			assertion: require.NoError,
		},
		{
			name: "with yearly",
			setup: func(_ *mocks.MockExerciseConfiguration) {
				// _m.EXPECT().GetFs().Return(afero.NewMemMapFs())
			},
			args: args{
				opts: []func(*analyze.Analyzer){
					analyze.WithDirectory("foo/bar"),
					analyze.WithYearly(true),
				},
			},
			want: &analyze.Analyzer{
				Data:      []*advent.BenchmarkData{},
				Dir:       "foo/bar",
				GraphType: 1,
				Output:    "",
			},
			assertion: require.NoError,
		},
		{
			name: "with daily",
			setup: func(_ *mocks.MockExerciseConfiguration) {
				// _m.EXPECT().GetFs().Return(afero.NewMemMapFs())
			},
			args: args{
				opts: []func(*analyze.Analyzer){
					analyze.WithDirectory("foo/bar"),
					analyze.WithDaily(true),
				},
			},
			want: &analyze.Analyzer{
				Data:      []*advent.BenchmarkData{},
				Dir:       "foo/bar",
				GraphType: 1,
				Output:    "",
			},
			assertion: require.NoError,
		},
		{
			name: "with output",
			setup: func(_ *mocks.MockExerciseConfiguration) {
				// _m.EXPECT().GetFs().Return(afero.NewMemMapFs())
			},
			args: args{
				opts: []func(*analyze.Analyzer){
					analyze.WithDirectory("foo/bar"),
					analyze.WithOutput("fakeOutput.png"),
				},
			},
			want: &analyze.Analyzer{
				Data:      []*advent.BenchmarkData{},
				Dir:       "foo/bar",
				GraphType: 1,
				Output:    "fakeOutput.png",
			},
			assertion: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up mocks
			mockConfig := mocks.NewMockExerciseConfiguration(t)
			mockConfig.EXPECT().GetLogger().Return(slog.New(slog.NewTextHandler(io.Discard, nil)))
			tt.setup(mockConfig)

			// execute function under test
			got, err := analyze.NewAnalyzer(mockConfig, tt.args.opts...)

			// verify results
			tt.assertion(t, err)
			if err == nil {
				assert.EqualExportedValues(t, *tt.want, *got)
			}
		})
	}
}

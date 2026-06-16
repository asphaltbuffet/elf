package analyze

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

func Test_NewAnalyzer(t *testing.T) {
	type args struct {
		opts []func(*Analyzer)
	}

	tests := []struct {
		name      string
		args      args
		want      *Analyzer
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "no dir",
			args: args{
				opts: []func(*Analyzer){},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "with directory",
			args: args{
				opts: []func(*Analyzer){
					WithDirectory("foo/bar"),
				},
			},
			want: &Analyzer{
				Data:   []*exercise.BenchmarkData{},
				Dir:    "foo/bar",
				Output: "",
			},
			assertion: require.NoError,
		},
		{
			name: "with output",
			args: args{
				opts: []func(*Analyzer){
					WithDirectory("foo/bar"),
					WithOutput("fakeOutput.png"),
				},
			},
			want: &Analyzer{
				Data:   []*exercise.BenchmarkData{},
				Dir:    "foo/bar",
				Output: "fakeOutput.png",
			},
			assertion: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAnalyzer(testLogger(), tt.args.opts...)

			tt.assertion(t, err)
			if err == nil {
				assert.EqualExportedValues(t, *tt.want, *got)
			}
		})
	}
}

func Test_readBenchmarkFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		got, err := readBenchmarkFile("testdata/valid_benchmark.json")
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, 2024, got[0].Year)
		assert.Equal(t, 1, got[0].Day)
		assert.Len(t, got[0].Implementations, 2)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := readBenchmarkFile("testdata/does_not_exist.json")
		require.Error(t, err)
		assert.ErrorContains(t, err, "reading file")
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := readBenchmarkFile("testdata/invalid_benchmark.json")
		require.Error(t, err)
		assert.ErrorContains(t, err, "unmarshalling json")
	})
}

func Test_getBenchmarkFiles(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		dir := t.TempDir()
		got, err := getBenchmarkFiles(dir)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("nested benchmark files", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "2024", "01-test")
		require.NoError(t, os.MkdirAll(sub, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(sub, "benchmark.json"), []byte("[]"), 0o644))

		got, err := getBenchmarkFiles(dir)
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Contains(t, got[0], "benchmark.json")
	})

	t.Run("non-benchmark files ignored", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "other.json"), []byte("{}"), 0o644))

		got, err := getBenchmarkFiles(dir)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("nonexistent dir", func(t *testing.T) {
		got, err := getBenchmarkFiles("/nonexistent/path/that/does/not/exist")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func Test_Load(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		a := &Analyzer{
			Dir:    t.TempDir(),
			logger: testLogger(),
		}

		err := a.Load()
		require.NoError(t, err)
		assert.Empty(t, a.Data)
	})

	t.Run("dir with valid benchmark", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "day01")
		require.NoError(t, os.MkdirAll(sub, 0o755))

		src, err := os.ReadFile("testdata/valid_benchmark.json")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(sub, "benchmark.json"), src, 0o644))

		a := &Analyzer{
			Dir:    dir,
			logger: testLogger(),
		}

		err = a.Load()
		require.NoError(t, err)
		assert.NotEmpty(t, a.Data)
	})

	t.Run("dir with invalid benchmark", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "benchmark.json"), []byte("{bad"), 0o644))

		a := &Analyzer{
			Dir:    dir,
			logger: testLogger(),
		}

		err := a.Load()
		require.Error(t, err)
	})
}

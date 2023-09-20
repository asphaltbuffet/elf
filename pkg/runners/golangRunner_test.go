package runners

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newGolangRunner(t *testing.T) {
	type args struct {
		dir string
	}

	tests := []struct {
		name string
		args args
		want Runner
	}{
		{
			name: "standard",
			args: args{
				dir: filepath.Join("testdata", "2015", "01-testDayOne", "go"),
			},
			want: &golangRunner{
				dir:                filepath.Join("testdata", "2015", "01-testDayOne", "go"),
				cmd:                nil,
				wrapperFilepath:    "",
				executableFilepath: "",
				stdin:              nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, newGolangRunner(tt.args.dir))
		})
	}
}

func Test_golangRunner_Cleanup(t *testing.T) {
	tf, err := os.MkdirTemp("", "test-go")
	require.NoError(t, err)

	defer assert.NoError(t, os.RemoveAll(tf))

	exDir := filepath.Join(tf, "2015", "01-testDayOne", "go")
	require.NoError(t, os.MkdirAll(exDir, 0o750))

	tests := []struct {
		name      string
		g         *golangRunner
		assertion assert.ErrorAssertionFunc
		err       error
	}{
		{
			name: "all files exist",
			g: &golangRunner{
				dir:                exDir,
				cmd:                nil,
				wrapperFilepath:    filepath.Join(exDir, golangWrapperFilename),
				executableFilepath: filepath.Join(exDir, golangWrapperExecutableFilename),
				stdin:              nil,
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, os.WriteFile(tt.g.wrapperFilepath, []byte("fake test data"), 0o600))
			require.NoError(t, os.WriteFile(tt.g.executableFilepath, []byte("fake binary"), 0o600))

			err := tt.g.Cleanup()

			assert.NoFileExists(t, tt.g.wrapperFilepath)
			assert.NoFileExists(t, tt.g.executableFilepath)
			require.DirExists(t, tt.g.dir)

			tt.assertion(t, err)
		})
	}
}

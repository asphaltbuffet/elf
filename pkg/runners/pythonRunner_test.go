package runners

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newPythonRunner(t *testing.T) {
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
				dir: filepath.Join("testdata", "2015", "01-testDayOne", "py"),
			},
			want: &pythonRunner{
				dir:             filepath.Join("testdata", "2015", "01-testDayOne", "py"),
				cmd:             nil,
				wrapperFilepath: "",
				stdin:           nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, newPythonRunner(tt.args.dir))
		})
	}
}

func Test_pythonRunner_Cleanup(t *testing.T) {
	tf, err := os.MkdirTemp("", "test-py")
	require.NoError(t, err)

	defer assert.NoError(t, os.RemoveAll(tf))

	exDir := filepath.Join(tf, "2015", "01-testDayOne", "py")
	require.NoError(t, os.MkdirAll(exDir, 0o750))

	tests := []struct {
		name             string
		p                *pythonRunner
		writeWrapperFile bool
		assertion        assert.ErrorAssertionFunc
		err              error
	}{
		{
			name: "existing wrapper file",
			p: &pythonRunner{
				dir:             exDir,
				cmd:             nil,
				wrapperFilepath: filepath.Join(exDir, pythonWrapperFilename),
				stdin:           nil,
			},
			writeWrapperFile: true,
			assertion:        assert.NoError,
		},
		{
			name: "no file",
			p: &pythonRunner{
				dir:             exDir,
				cmd:             nil,
				wrapperFilepath: filepath.Join(exDir, pythonWrapperFilename),
				stdin:           nil,
			},
			writeWrapperFile: false,
			assertion:        assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.writeWrapperFile {
				require.NoError(t, os.WriteFile(tt.p.wrapperFilepath, []byte("fake test data"), 0o600))
			}
			require.DirExists(t, tt.p.dir)

			err := tt.p.Cleanup()

			tt.assertion(t, err)
		})
	}
}

func Test_pythonRunner_Stop(t *testing.T) {
	tests := []struct {
		name      string
		p         *pythonRunner
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "cmd is nil",
			p:         &pythonRunner{},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.p.Stop())
		})
	}
}

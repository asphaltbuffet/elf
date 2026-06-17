package runners

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

func Test_customWriter_Write(t *testing.T) {
	type args struct {
		b []byte
	}

	type output struct {
		n       int
		entries [][]byte
		pending []byte
	}

	tests := []struct {
		name      string
		c         *customWriter
		args      args
		want      output
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "single line",
			c:    &customWriter{},
			args: args{
				b: []byte("fake\n"),
			},
			want: output{
				n:       5,
				entries: [][]byte{[]byte("fake")},
				pending: nil,
			},
			assertion: assert.NoError,
		},
		{
			name: "multiple lines",
			c:    &customWriter{},
			args: args{
				b: []byte("line1\nline2\nline3\n"),
			},
			want: output{
				n:       18,
				entries: [][]byte{[]byte("line1"), []byte("line2"), []byte("line3")},
				pending: nil,
			},
			assertion: assert.NoError,
		},
		{
			name: "no newline leaves pending",
			c:    &customWriter{},
			args: args{
				b: []byte("partial"),
			},
			want: output{
				n:       7,
				entries: nil,
				pending: []byte("partial"),
			},
			assertion: assert.NoError,
		},
		{
			name: "empty input",
			c:    &customWriter{},
			args: args{
				b: []byte{},
			},
			want: output{
				n:       0,
				entries: nil,
				pending: nil,
			},
			assertion: assert.NoError,
		},
		{
			name: "appends to existing pending",
			c:    &customWriter{pending: []byte("hello ")},
			args: args{
				b: []byte("world\n"),
			},
			want: output{
				n:       6,
				entries: [][]byte{[]byte("hello world")},
				pending: nil,
			},
			assertion: assert.NoError,
		},
		{
			name: "trailing content after last newline",
			c:    &customWriter{},
			args: args{
				b: []byte("done\nleftover"),
			},
			want: output{
				n:       13,
				entries: [][]byte{[]byte("done")},
				pending: []byte("leftover"),
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Write(tt.args.b)
			cw := tt.c

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want.entries, cw.entries)
				assert.Equal(t, tt.want.pending, cw.pending)
				assert.Equal(t, tt.want.n, got)
			}
		})
	}
}

func Test_customWriter_GetEntry(t *testing.T) {
	tests := []struct {
		name      string
		entries   [][]byte
		want      []byte
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{"no entries to get", [][]byte{}, nil, assert.Error, "no entries"},
		{"single entry", [][]byte{[]byte("test data")}, []byte("test data"), assert.NoError, ""},
		{
			"multiple entries",
			[][]byte{
				[]byte("test data 1"),
				[]byte("test data 2"),
				[]byte("fake data 3"),
			},
			[]byte("test data 1"), assert.NoError, "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &customWriter{entries: tt.entries}

			got, err := c.GetEntry()

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got, "expected %q, got %q", tt.want, got)
				assert.NotContains(t, c.entries, got)
			}
		})
	}
}

func TestSetupBuffers(t *testing.T) {
	tests := []struct {
		name      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "Set buffers correctly",
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &exec.Cmd{}

			got, err := setupBuffers(c)

			tt.assertion(t, err)

			if err == nil {
				assert.IsType(t, &customWriter{}, c.Stdout)
				assert.IsType(t, &bytes.Buffer{}, c.Stderr)
				assert.NotNil(t, got)
			}
		})
	}
}

func Test_checkWait(t *testing.T) {
	t.Run("returns entry when available", func(t *testing.T) {
		cw := &customWriter{
			entries: [][]byte{[]byte(`{"task_id":"1","ok":true}`)},
		}
		cmd := exec.Command("true")
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		got, err := checkWait(cmd)

		require.NoError(t, err)
		assert.JSONEq(t, `{"task_id":"1","ok":true}`, string(got))
	})

	t.Run("returns error when process exits with no entries", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "echo 'something went wrong' >&2; exit 42")
		cw := &customWriter{}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		_ = cmd.Run() // sets ProcessState

		got, err := checkWait(cmd)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "exit code 42")
		assert.Contains(t, err.Error(), "something went wrong")
	})

	t.Run("returns error with empty stderr on process exit", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "exit 1")
		cw := &customWriter{}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		_ = cmd.Run()

		got, err := checkWait(cmd)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "exit code 1")
	})

	t.Run("prefers entries over process state", func(t *testing.T) {
		// Even if the process has exited, available entries are returned first.
		cmd := exec.Command("true")
		cw := &customWriter{
			entries: [][]byte{[]byte("result data")},
		}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		_ = cmd.Run() // ProcessState is set, but entries exist

		got, err := checkWait(cmd)

		require.NoError(t, err)
		assert.Equal(t, []byte("result data"), got)
	})
}

func Test_readJSONFromCommand(t *testing.T) {
	t.Run("unmarshals valid JSON result", func(t *testing.T) {
		cmd := exec.Command("true")
		cw := &customWriter{
			entries: [][]byte{
				[]byte(`{"task_id":"abc","ok":true,"output":"42","duration":1.23}`),
			},
		}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		_ = cmd.Run()

		var result protocol.Result
		err := readJSONFromCommand(&result, cmd)

		require.NoError(t, err)
		assert.Equal(t, "abc", result.TaskID)
		assert.True(t, result.Ok)
		assert.Equal(t, "42", result.Output)
		assert.InDelta(t, 1.23, result.Duration, 0.001)
	})

	t.Run("skips non-JSON debug lines before valid JSON", func(t *testing.T) {
		cmd := exec.Command("true")
		cw := &customWriter{
			entries: [][]byte{
				[]byte("debug: initializing solver"),
				[]byte("debug: reading input"),
				[]byte(`{"task_id":"t1","ok":true,"output":"answer","duration":0.5}`),
			},
		}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		_ = cmd.Run()

		var result protocol.Result
		err := readJSONFromCommand(&result, cmd)

		require.NoError(t, err)
		assert.Equal(t, "answer", result.Output)
	})

	t.Run("returns error when process exits without valid JSON", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "exit 1")
		cw := &customWriter{
			entries: [][]byte{
				[]byte("not json at all"),
			},
		}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		_ = cmd.Run()

		var result protocol.Result
		err := readJSONFromCommand(&result, cmd)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "exit code 1")
	})

	t.Run("works with generic target type", func(t *testing.T) {
		cmd := exec.Command("true")
		cw := &customWriter{
			entries: [][]byte{
				[]byte(`{"name":"test","value":123}`),
			},
		}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		_ = cmd.Run()

		var raw map[string]json.RawMessage
		err := readJSONFromCommand(&raw, cmd)

		require.NoError(t, err)
		assert.Contains(t, raw, "name")
		assert.Contains(t, raw, "value")
	})
}

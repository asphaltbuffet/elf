package runners

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"testing"
	"time"

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

// reaped runs cmd to completion (populating ProcessState) and returns a closed
// channel, modelling the post-Wait state the reaper goroutine signals via the
// runner's exited channel.
func reaped(cmd *exec.Cmd) <-chan struct{} {
	_ = cmd.Run()
	ch := make(chan struct{})
	close(ch)

	return ch
}

func Test_checkWait(t *testing.T) {
	t.Run("returns entry when available", func(t *testing.T) {
		cw := &customWriter{
			entries: [][]byte{[]byte(`{"task_id":"1","ok":true}`)},
		}
		cmd := exec.Command("true")
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		// Process still running: an open channel never fires the exit path.
		got, err := checkWait(context.Background(), cmd, make(chan struct{}))

		require.NoError(t, err)
		assert.JSONEq(t, `{"task_id":"1","ok":true}`, string(got))
	})

	t.Run("returns error when process exits with no entries", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "echo 'something went wrong' >&2; exit 42")
		cw := &customWriter{}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		got, err := checkWait(context.Background(), cmd, reaped(cmd))

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

		got, err := checkWait(context.Background(), cmd, reaped(cmd))

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "exit code 1")
	})

	t.Run("returns error when process dies without producing output", func(t *testing.T) {
		// Real-world path: the subprocess crashes on startup (writes to stderr,
		// nothing to stdout). checkWait must detect exit via the closed channel
		// and return its error rather than spinning forever. The timeout guard
		// fails loudly if it hangs.
		cmd := exec.Command("sh", "-c", "echo 'startup crash' >&2; exit 3")
		cw := &customWriter{}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		exited := reaped(cmd)
		done := make(chan struct{})

		var (
			got []byte
			err error
		)

		go func() {
			got, err = checkWait(context.Background(), cmd, exited)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("checkWait hung on a dead subprocess")
		}

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "exit code 3")
		assert.Contains(t, err.Error(), "startup crash")
	})

	t.Run("prefers entries over process exit", func(t *testing.T) {
		// Even if the process has exited, available entries are returned first.
		cmd := exec.Command("true")
		cw := &customWriter{
			entries: [][]byte{[]byte("result data")},
		}
		cmd.Stdout = cw
		cmd.Stderr = new(bytes.Buffer)

		got, err := checkWait(context.Background(), cmd, reaped(cmd))

		require.NoError(t, err)
		assert.Equal(t, []byte("result data"), got)
	})
}

func Test_checkWait_contextCancelled(t *testing.T) {
	const (
		taskTimeout  = 50 * time.Millisecond
		guardTimeout = 2 * time.Second
	)

	// Start a subprocess that never writes output so checkWait blocks.
	cmd := exec.Command("sleep", "60")
	cw := &customWriter{}
	cmd.Stdout = cw
	cmd.Stderr = new(bytes.Buffer)
	require.NoError(t, cmd.Start())

	exited := make(chan struct{})

	go func() {
		_ = cmd.Wait()
		close(exited)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	start := time.Now()

	got, err := checkWait(ctx, cmd, exited)

	require.Less(t, time.Since(start), guardTimeout, "checkWait should return promptly on context cancellation")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func Test_readJSONFromCommand_contextCancelled(t *testing.T) {
	const (
		taskTimeout  = 50 * time.Millisecond
		guardTimeout = 2 * time.Second
	)

	// Start a subprocess that never writes output so readJSONFromCommand blocks.
	cmd := exec.Command("sleep", "60")
	cw := &customWriter{}
	cmd.Stdout = cw
	cmd.Stderr = new(bytes.Buffer)
	require.NoError(t, cmd.Start())

	exited := make(chan struct{})

	go func() {
		_ = cmd.Wait()
		close(exited)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	start := time.Now()

	var result protocol.Result
	err := readJSONFromCommand(ctx, &result, cmd, exited)

	require.Less(
		t,
		time.Since(start),
		guardTimeout,
		"readJSONFromCommand should return promptly on context cancellation",
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
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

		exited := reaped(cmd)

		var result protocol.Result
		err := readJSONFromCommand(context.Background(), &result, cmd, exited)

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

		exited := reaped(cmd)

		var result protocol.Result
		err := readJSONFromCommand(context.Background(), &result, cmd, exited)

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

		exited := reaped(cmd)

		var result protocol.Result
		err := readJSONFromCommand(context.Background(), &result, cmd, exited)

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

		exited := reaped(cmd)

		var raw map[string]json.RawMessage
		err := readJSONFromCommand(context.Background(), &raw, cmd, exited)

		require.NoError(t, err)
		assert.Contains(t, raw, "name")
		assert.Contains(t, raw, "value")
	})
}

package runners

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type customWriter struct {
	pending []byte
	entries [][]byte
	mux     sync.Mutex
}

// Write writes the given bytes to the custom writer.
//
// Newline characters ('\n') are used to flush the pending buffer and append
// the current contents to the list of entries.
func (c *customWriter) Write(b []byte) (int, error) {
	var n int

	c.mux.Lock()
	for _, x := range b {
		if x == '\n' {
			c.entries = append(c.entries, c.pending)
			c.pending = nil
		} else {
			c.pending = append(c.pending, x)
		}
		n++
	}
	c.mux.Unlock()

	return n, nil
}

// GetEntry returns the next entry from the custom writer, or nil if there are no more entries.
//
// If there are no more entries, the function returns an error.
func (c *customWriter) GetEntry() ([]byte, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if len(c.entries) == 0 {
		return nil, errors.New("no entries")
	}

	var x []byte
	x, c.entries = c.entries[0], c.entries[1:]

	return x, nil
}

func setupBuffers(cmd *exec.Cmd) (io.WriteCloser, error) {
	stdoutWriter := &customWriter{}
	cmd.Stdout = stdoutWriter
	cmd.Stderr = new(bytes.Buffer)

	return cmd.StdinPipe()
}

// checkWait returns the next stdout entry from the command, blocking until one
// is available. The exited channel must be closed by the sole owner of
// cmd.Wait() once the process terminates; checkWait uses it to detect a
// subprocess that died without producing output (e.g. a startup crash) and
// returns its exit code and stderr rather than spinning forever.
func checkWait(ctx context.Context, cmd *exec.Cmd, exited <-chan struct{}) ([]byte, error) {
	const checkWaitDelay time.Duration = 10 * time.Millisecond

	c := cmd.Stdout.(*customWriter) //nolint:errcheck // we will handle errors in the loop

	for {
		e, err := c.GetEntry()
		if err == nil {
			return e, nil
		}

		select {
		case <-ctx.Done():
			if cmd.Process != nil {
				// Kill the whole process group, not just the leader: a wrapper
				// runner's real work runs in a forked child that would otherwise
				// survive and keep the stdout pipe open, blocking cmd.Wait().
				_ = signalGroup(cmd.Process.Pid, syscall.SIGKILL)
			}

			return nil, fmt.Errorf("task timed out: %w", ctx.Err())
		case <-exited:
			// Re-check for a final entry that may have landed alongside exit, so
			// a process that wrote its result and then exited still succeeds.
			if finalEntry, getErr := c.GetEntry(); getErr == nil {
				return finalEntry, nil
			}

			stderrBuf, _ := cmd.Stderr.(*bytes.Buffer)
			return nil, fmt.Errorf(
				"run failed with exit code %d: %s",
				cmd.ProcessState.ExitCode(),
				stderrBuf.String())
		default:
			time.Sleep(checkWaitDelay)
		}
	}
}

func readJSONFromCommand(ctx context.Context, res any, cmd *exec.Cmd, exited <-chan struct{}) error {
	for {
		inp, err := checkWait(ctx, cmd, exited)
		if err != nil {
			return err
		}

		err = json.Unmarshal(inp, res)
		if err != nil {
			// anything returned as an error is considered a debug message
			fmt.Fprintf(os.Stderr, "[DBG] %s\n", strings.TrimSpace(string(inp)))
		} else {
			break
		}
	}

	return nil
}

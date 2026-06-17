package runners

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
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

func checkWait(cmd *exec.Cmd) ([]byte, error) {
	const checkWaitDelay time.Duration = 10 * time.Millisecond

	c := cmd.Stdout.(*customWriter) //nolint:errcheck // we will handle errors in the loop

	for {
		e, err := c.GetEntry()
		if err == nil {
			return e, nil
		}

		if cmd.ProcessState != nil {
			stderrBuf, _ := cmd.Stderr.(*bytes.Buffer)
			return nil, fmt.Errorf(
				"run failed with exit code %d: %s",
				cmd.ProcessState.ExitCode(),
				stderrBuf.String())
		}

		time.Sleep(checkWaitDelay)
	}
}

func readJSONFromCommand(res any, cmd *exec.Cmd) error {
	for {
		inp, err := checkWait(cmd)
		if err != nil {
			return err
		}

		err = json.Unmarshal(inp, res)
		if err != nil {
			// anything returned as an error is considered a debug message
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			//nolint:forbidigo // intentional debug output to user
			fmt.Printf("[%s] %v\n", style.Render("DBG"), strings.TrimSpace(string(inp)))
		} else {
			break
		}
	}

	return nil
}

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

// A Task represents a unit of work to be performed.
type Task struct {
	// TaskID is the unique identifier for the task.
	TaskID string `json:"task_id"`

	// Part is the part of the work that the task should perform.
	Part Part `json:"part"`

	// Input is the input data for the task.
	Input string `json:"input"`

	// OutputDir is the directory where the task should store its output.
	// This field is optional.
	OutputDir string `json:"output_dir,omitempty"`
}

// A Result represents the outcome of a Task.
type Result struct {
	// TaskID is the unique identifier for the task that produced this result.
	TaskID string `json:"task_id"`
	// Ok indicates whether the task was successful.
	Ok bool `json:"ok"`
	// Output is the output of the task, if successful.
	Output string `json:"output"`
	// Duration is the amount of time it took for the task to complete.
	Duration float64 `json:"duration"`
}

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
			return nil, fmt.Errorf(
				"run failed with exit code %d: %s",
				cmd.ProcessState.ExitCode(),
				cmd.Stderr.(*bytes.Buffer).String())
		}

		time.Sleep(checkWaitDelay)
	}
}

func readJSONFromCommand(res interface{}, cmd *exec.Cmd) error {
	for {
		inp, err := checkWait(cmd)
		if err != nil {
			return err
		}

		err = json.Unmarshal(inp, res)
		if err != nil {
			// anything returned as an error is considered a debug message
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			fmt.Printf("[%s] %v\n", style.Render("DBG"), strings.TrimSpace(string(inp)))
		} else {
			break
		}
	}

	return nil
}

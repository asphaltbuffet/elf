//go:build unix

package runners

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignalGroup_KillsForkedChild reproduces the bash-wrapper hang: a shell
// that forks a long-lived child (mimicking `$(... | part_one)`). Killing only
// the leader would orphan the child; signalGroup must take down the whole tree.
func TestSignalGroup_KillsForkedChild(t *testing.T) {
	// Parent shell spawns a `sleep 60` child, prints the child PID, then waits.
	script := `sleep 60 & echo $!; wait`
	cmd := exec.Command("sh", "-c", script)
	setProcessGroup(cmd)

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())

	// Read the child PID the script printed.
	var childPID int
	line, err := bufio.NewReader(stdout).ReadString('\n')
	require.NoError(t, err)
	_, err = fmt.Sscan(line, &childPID)
	require.NoError(t, err)
	require.Positive(t, childPID)

	// Kill the whole group and reap the leader.
	require.NoError(t, signalGroup(cmd.Process.Pid, syscall.SIGKILL))
	_ = cmd.Wait()

	// The forked child must be gone too. Give the kernel a moment to reap.
	assert.Eventually(t, func() bool {
		return !processAlive(childPID)
	}, 2*time.Second, 20*time.Millisecond, "forked child %d survived group kill", childPID)
}

// TestSignalGroup_AlreadyDead confirms signaling a reaped group is a benign
// no-op (os.ErrProcessDone), not a hard error.
func TestSignalGroup_AlreadyDead(t *testing.T) {
	cmd := exec.Command("true")
	setProcessGroup(cmd)
	require.NoError(t, cmd.Start())
	require.NoError(t, cmd.Wait()) // fully reaped

	err := signalGroup(cmd.Process.Pid, syscall.SIGKILL)
	assert.ErrorIs(t, err, os.ErrProcessDone)
}

// processAlive reports whether pid refers to a live (non-reaped) process.
// signal 0 performs error checking without actually sending a signal.
func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

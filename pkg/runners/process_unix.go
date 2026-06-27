//go:build unix

package runners

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// setProcessGroup places the command in its own process group so that
// signalGroup can later terminate the whole tree (the leader plus any children
// a wrapper runner forks).
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// signalGroup sends sig to the entire process group led by pid (started with
// SysProcAttr.Setpgid). Signaling the group (negative pid) reaches a wrapper's
// forked children, not just the leader. [syscall.ESRCH] ("no such process") is
// normalized to [os.ErrProcessDone] so callers can treat an already-dead group
// as a benign no-op.
func signalGroup(pid int, sig syscall.Signal) error {
	if err := syscall.Kill(-pid, sig); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}

		return err
	}

	return nil
}

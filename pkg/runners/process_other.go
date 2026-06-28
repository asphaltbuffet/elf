//go:build !unix

package runners

import (
	"os"
	"os/exec"
	"syscall"
)

// setProcessGroup is a no-op on platforms without POSIX process groups.
func setProcessGroup(_ *exec.Cmd) {}

// signalGroup falls back to signaling the single leader process, since process
// groups are unavailable. The sig argument is ignored: [os.Process] only
// supports Kill on these platforms. [os.ErrProcessDone] is returned for an
// already-exited process so callers can treat it as a benign no-op.
func signalGroup(pid int, _ syscall.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := proc.Kill(); err != nil {
		return err
	}

	return nil
}

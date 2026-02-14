package components

import (
	"context"
	"os/exec"
	"runtime"
)

// OpenFile opens a file with the system's default application.
func OpenFile(path string) error {
	ctx := context.Background()

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", path)
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", path)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "", path)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", path)
	}

	return cmd.Start()
}

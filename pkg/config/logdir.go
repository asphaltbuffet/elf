package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// runtimeGOOS is a seam over [runtime.GOOS] so tests can reason about the current
// platform without importing runtime.
func runtimeGOOS() string { return runtime.GOOS }

// homeDir returns the current user's home directory.
func homeDir() (string, error) { return os.UserHomeDir() }

// userStateDir returns the per-user base directory for elf's logs, following
// the platform convention. The caller appends "elf" and the log filename.
//
//   - Linux:   $XDG_STATE_HOME, else ~/.local/state
//   - macOS:   ~/Library/Logs
//   - Windows: %LocalAppData% (os.UserCacheDir on Windows)
//
// It returns an error only when the home/base directory cannot be determined.
func userStateDir() (string, error) {
	switch runtimeGOOS() {
	case "windows":
		// os.UserCacheDir returns %LocalAppData% on Windows.
		return os.UserCacheDir()

	case "darwin":
		home, err := homeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, "Library", "Logs"), nil

	default: // linux and other unix
		if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
			return xdg, nil
		}

		home, err := homeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, ".local", "state"), nil
	}
}

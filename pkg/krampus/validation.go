package krampus

import (
	"errors"
	"fmt"
)

// Validation errors.
var (
	ErrTokenNotSet      = errors.New("advent token not configured")
	ErrTokenPlaceholder = errors.New("advent token is placeholder value")
	ErrDirectoryMissing = errors.New("directory does not exist")
)

// ValidateToken checks if the advent token is properly configured.
// Returns an error if the token is empty or is the default placeholder.
func (c Config) ValidateToken() error {
	token := c.GetToken()

	if token == "" {
		return ErrTokenNotSet
	}

	if token == defaults[AdventTokenKey] {
		return ErrTokenPlaceholder
	}

	return nil
}

// ValidateDirectories checks if configured directories exist.
// Returns a slice of errors for each missing directory.
func (c Config) ValidateDirectories() []error {
	var errs []error

	dirs := map[string]string{
		"config":   c.GetConfigDir(),
		"cache":    c.GetCacheDir(),
		"exercise": c.GetBaseDir(),
	}

	for name, dir := range dirs {
		if dir == "" {
			continue
		}

		info, err := c.fs.Stat(dir)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s directory %q: %w", name, dir, ErrDirectoryMissing))

			continue
		}

		if !info.IsDir() {
			errs = append(errs, fmt.Errorf("%s path %q is not a directory", name, dir))
		}
	}

	return errs
}

// Validate performs full validation of the configuration.
// Returns all validation errors found.
func (c Config) Validate() []error {
	var errs []error

	if err := c.ValidateToken(); err != nil {
		errs = append(errs, err)
	}

	errs = append(errs, c.ValidateDirectories()...)

	return errs
}

// MaskToken returns a masked version of the token for display.
// Shows first 4 and last 4 characters with "..." in between.
// Returns empty string if token is empty.
// Returns "****" if token is too short to mask properly.
func MaskToken(token string) string {
	if token == "" {
		return ""
	}

	const minLen = 10

	if len(token) < minLen {
		return "****"
	}

	return token[:4] + "..." + token[len(token)-4:]
}

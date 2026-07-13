package secret

import (
	"fmt"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/spf13/afero"
)

// LoadIdentity reads an unencrypted on-disk SSH private key at path (through fs)
// and parses it into an age identity. Passphrase-protected keys are not
// supported in this version and return a parse error (see ADR-0020).
func LoadIdentity(fs afero.Fs, path string) (age.Identity, error) {
	pemBytes, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, fmt.Errorf("reading identity %q: %w", path, err)
	}

	id, err := agessh.ParseIdentity(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing identity %q (passphrase-protected keys are unsupported): %w", path, err)
	}

	return id, nil
}

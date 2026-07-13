package secret

import (
	"errors"
	"fmt"

	"filippo.io/age"
	"filippo.io/age/agessh"
)

// ErrNoRecipients is returned when no encryption recipients are configured.
var ErrNoRecipients = errors.New("no recipients configured (set encrypt.recipients in config)")

// ParseRecipients parses SSH public-key strings into age recipients. It is
// all-or-nothing: an empty list or any single unparseable key returns an error
// and no recipients, so encryption never seals to a partial set (see ADR-0020).
func ParseRecipients(keys []string) ([]age.Recipient, error) {
	if len(keys) == 0 {
		return nil, ErrNoRecipients
	}

	recipients := make([]age.Recipient, 0, len(keys))
	for i, key := range keys {
		r, err := agessh.ParseRecipient(key)
		if err != nil {
			return nil, fmt.Errorf("parsing recipient %d (%q): %w", i, key, err)
		}
		recipients = append(recipients, r)
	}

	return recipients, nil
}

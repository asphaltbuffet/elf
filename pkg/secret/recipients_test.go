package secret_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/secret"
)

// A syntactically valid ssh-ed25519 public key (throwaway, generated for tests).
const testPubKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBPVqN5VN24U3qAsA2C1vX7lrsG90jbwsXe1H2sHup49 test@elf"

func TestParseRecipients_Valid(t *testing.T) {
	t.Parallel()

	rs, err := secret.ParseRecipients([]string{testPubKey})
	require.NoError(t, err)
	require.Len(t, rs, 1)
}

func TestParseRecipients_Empty(t *testing.T) {
	t.Parallel()

	_, err := secret.ParseRecipients(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no recipients")
}

func TestParseRecipients_OneInvalidFailsAll(t *testing.T) {
	t.Parallel()

	_, err := secret.ParseRecipients([]string{testPubKey, "not-a-key"})
	require.Error(t, err)
}

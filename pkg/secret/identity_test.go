package secret_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/secret"
)

// An unencrypted OpenSSH ed25519 private key generated for tests only.
const testPrivKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACAT1ajeVTduFN6gLANgtb1+5a7BvdI28LF3tR9rB7qePQAAAJC78wBQu/MA
UAAAAAtzc2gtZWQyNTUxOQAAACAT1ajeVTduFN6gLANgtb1+5a7BvdI28LF3tR9rB7qePQ
AAAEA+E1IkDnE92aRZud0x8dDOyOFFLjq6s46lRmL8LHamiRPVqN5VN24U3qAsA2C1vX7l
rsG90jbwsXe1H2sHup49AAAACHRlc3RAZWxmAQIDBAU=
-----END OPENSSH PRIVATE KEY-----
`

func TestLoadIdentity_Valid(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/home/u/.ssh/id_ed25519", []byte(testPrivKey), 0o600)
	require.NoError(t, err)

	id, err := secret.LoadIdentity(fs, "/home/u/.ssh/id_ed25519")
	require.NoError(t, err)
	require.NotNil(t, id)
}

func TestLoadIdentity_Missing(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_, err := secret.LoadIdentity(fs, "/nope/id_ed25519")
	require.Error(t, err)
}

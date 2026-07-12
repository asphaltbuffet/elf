package secret_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/secret"
)

func TestDecrypt_RoundTrip(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	seedExercise(t, fs)
	rs, err := secret.ParseRecipients([]string{testPubKey})
	require.NoError(t, err)

	_, err = secret.Encrypt(fs, rs, "/ex/42", []string{"go", "py"})
	require.NoError(t, err)

	// Remove plaintext to simulate a fresh clone (only .age present).
	for _, p := range []string{"/ex/42/info.json", "/ex/42/go/exercise.go", "/ex/42/py/exercise.py"} {
		require.NoError(t, fs.Remove(p))
	}

	require.NoError(t, afero.WriteFile(fs, "/id", []byte(testPrivKey), 0o600))
	id, err := secret.LoadIdentity(fs, "/id")
	require.NoError(t, err)

	report, err := secret.Decrypt(fs, id, "/ex/42", false)
	require.NoError(t, err)

	got, err := afero.ReadFile(fs, "/ex/42/go/exercise.go")
	require.NoError(t, err)
	assert.Equal(t, "package main", string(got))

	// .age retained.
	exists, _ := afero.Exists(fs, "/ex/42/go/exercise.go.age")
	assert.True(t, exists, ".age file must be retained after decrypt")

	assert.Len(t, report, 3)
}

func TestDecrypt_SkipsExistingPlaintext(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	seedExercise(t, fs)
	rs, err := secret.ParseRecipients([]string{testPubKey})
	require.NoError(t, err)

	_, err = secret.Encrypt(fs, rs, "/ex/42", []string{"go", "py"})
	require.NoError(t, err)
	// plaintext still present (not removed).

	require.NoError(t, afero.WriteFile(fs, "/id", []byte(testPrivKey), 0o600))
	id, err := secret.LoadIdentity(fs, "/id")
	require.NoError(t, err)

	report, err := secret.Decrypt(fs, id, "/ex/42", false)
	require.NoError(t, err)

	for _, e := range report {
		assert.Equal(t, secret.Skipped, e.Outcome, "entry %s", e.Path)
	}
}

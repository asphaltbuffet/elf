package secret_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/secret"
)

func seedExercise(t *testing.T, fs afero.Fs) {
	t.Helper()
	files := map[string]string{
		"/ex/42/info.json":      `{"kind":"problem","number":42}`,
		"/ex/42/README.md":      "# 42",
		"/ex/42/go/exercise.go": "package main",
		"/ex/42/py/exercise.py": "print(1)",
	}
	for p, c := range files {
		require.NoError(t, afero.WriteFile(fs, p, []byte(c), 0o644))
	}
}

func TestEncrypt_SealsSolutionSetOnly(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	seedExercise(t, fs)

	rs, err := secret.ParseRecipients([]string{testPubKey})
	require.NoError(t, err)

	report, err := secret.Encrypt(fs, rs, "/ex/42", []string{"go", "py"})
	require.NoError(t, err)

	// info.json + go/exercise.go + py/exercise.py are sealed.
	for _, p := range []string{"/ex/42/info.json.age", "/ex/42/go/exercise.go.age", "/ex/42/py/exercise.py.age"} {
		exists, _ := afero.Exists(fs, p)
		assert.True(t, exists, "expected %s to exist", p)
	}

	// README.md is NOT sealed; plaintext is retained.
	exists, _ := afero.Exists(fs, "/ex/42/README.md.age")
	assert.False(t, exists, "README.md should not be sealed")

	exists, _ = afero.Exists(fs, "/ex/42/info.json")
	assert.True(t, exists, "plaintext info.json must be retained")

	assert.Len(t, report, 3)
	for _, e := range report {
		assert.Equal(t, secret.Added, e.Outcome, "entry %s", e.Path)
	}
}

func TestEncrypt_RefreshReplaces(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	seedExercise(t, fs)
	rs, err := secret.ParseRecipients([]string{testPubKey})
	require.NoError(t, err)

	_, err = secret.Encrypt(fs, rs, "/ex/42", []string{"go", "py"})
	require.NoError(t, err)

	// Re-encrypt: existing .age files are Replaced, not Added.
	report, err := secret.Encrypt(fs, rs, "/ex/42", []string{"go", "py"})
	require.NoError(t, err)

	for _, e := range report {
		assert.Equal(t, secret.Replaced, e.Outcome, "entry %s", e.Path)
	}
}

func TestEncrypt_MissingInfoJSONErrors(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/ex/42/go/exercise.go", []byte("package main"), 0o644))
	rs, err := secret.ParseRecipients([]string{testPubKey})
	require.NoError(t, err)

	_, err = secret.Encrypt(fs, rs, "/ex/42", []string{"go"})
	assert.Error(t, err, "expected error when info.json is missing")
}

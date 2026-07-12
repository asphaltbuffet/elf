package app_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
)

// testPubKey is a throwaway ed25519 SSH public key used only to exercise the
// recipients config path in these tests. Declared fresh here since pkg/app
// tests cannot import pkg/secret's unexported test constants.
const testPubKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBPVqN5VN24U3qAsA2C1vX7lrsG90jbwsXe1H2sHup49 test@elf"

// TestApp_Encrypt_Seals verifies that App.Encrypt seals the exercise's
// plaintext files into per-file .age ciphertext using configured recipients
// and language keys.
func TestApp_Encrypt_Seals(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/ex/42/info.json", []byte(`{"kind":"problem","number":42}`), 0o644))
	require.NoError(t, afero.WriteFile(fs, "/ex/42/go/exercise.go", []byte("package main"), 0o644))

	cfg, err := config.NewConfig(config.WithFs(fs))
	require.NoError(t, err)

	cfg.Viper().Set(config.RecipientsKey.String(), []string{testPubKey})
	cfg.Viper().Set(config.RunnersKey.String(), []map[string]any{{"key": "go", "name": "Go"}})

	a := app.New(cfg)

	report, err := a.Encrypt("/ex/42")
	require.NoError(t, err)

	exists, err := afero.Exists(fs, "/ex/42/info.json.age")
	require.NoError(t, err)
	assert.True(t, exists, "expected info.json.age")

	assert.NotEmpty(t, report, "expected non-empty report")
}

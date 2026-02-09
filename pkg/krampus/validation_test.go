package krampus

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateToken_Empty(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Set empty token
	cfg.SetToken("")

	err = cfg.ValidateToken()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenNotSet)
}

func TestValidateToken_Placeholder(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Default token is the placeholder
	err = cfg.ValidateToken()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenPlaceholder)
}

func TestValidateToken_Valid(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	cfg.SetToken("valid-real-token-value")

	err = cfg.ValidateToken()
	require.NoError(t, err)
}

func TestValidateDirectories_AllExist(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create the directories
	require.NoError(t, fs.MkdirAll("/home/user/.config/elf", 0o755))
	require.NoError(t, fs.MkdirAll("/home/user/.cache/elf", 0o755))
	require.NoError(t, fs.MkdirAll("exercises", 0o755))

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Set to our test directories
	cfg.SetValue(ConfigDirKey, "/home/user/.config/elf")
	cfg.SetValue(CacheDirKey, "/home/user/.cache/elf")
	cfg.SetValue(AdventDirKey, "exercises")

	errs := cfg.ValidateDirectories()
	assert.Empty(t, errs)
}

func TestValidateDirectories_MissingDir(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Set to non-existent directories
	cfg.SetValue(ConfigDirKey, "/nonexistent/config")
	cfg.SetValue(CacheDirKey, "/nonexistent/cache")
	cfg.SetValue(AdventDirKey, "/nonexistent/exercises")

	errs := cfg.ValidateDirectories()
	assert.Len(t, errs, 3)

	for _, err := range errs {
		assert.ErrorIs(t, err, ErrDirectoryMissing)
	}
}

func TestValidateDirectories_NotADirectory(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create files instead of directories for all paths
	f, err := fs.Create("/tmp/notadir-config")
	require.NoError(t, err)
	f.Close()

	f, err = fs.Create("/tmp/notadir-cache")
	require.NoError(t, err)
	f.Close()

	f, err = fs.Create("/tmp/notadir-exercises")
	require.NoError(t, err)
	f.Close()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	cfg.SetValue(ConfigDirKey, "/tmp/notadir-config")
	cfg.SetValue(CacheDirKey, "/tmp/notadir-cache")
	cfg.SetValue(AdventDirKey, "/tmp/notadir-exercises")

	errs := cfg.ValidateDirectories()
	require.Len(t, errs, 3) // config + cache + exercise dirs

	// All errors should mention "not a directory"
	for _, err := range errs {
		assert.Contains(t, err.Error(), "not a directory")
	}
}

func TestValidate_AllErrors(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Default config has placeholder token and non-existent directories
	errs := cfg.Validate()

	// Should have token error + directory errors
	assert.NotEmpty(t, errs)

	// Token error should be present
	hasTokenErr := false

	for _, err := range errs {
		if errors.Is(err, ErrTokenPlaceholder) || errors.Is(err, ErrTokenNotSet) {
			hasTokenErr = true

			break
		}
	}

	assert.True(t, hasTokenErr, "should have token validation error")
}

func TestValidate_NoErrors(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create all required directories
	require.NoError(t, fs.MkdirAll("/config", 0o755))
	require.NoError(t, fs.MkdirAll("/cache", 0o755))
	require.NoError(t, fs.MkdirAll("/exercises", 0o755))

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Set valid token and directories
	cfg.SetToken("real-valid-token-here")
	cfg.SetValue(ConfigDirKey, "/config")
	cfg.SetValue(CacheDirKey, "/cache")
	cfg.SetValue(AdventDirKey, "/exercises")

	errs := cfg.Validate()
	assert.Empty(t, errs)
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "empty token",
			token: "",
			want:  "",
		},
		{
			name:  "short token",
			token: "short",
			want:  "****",
		},
		{
			name:  "exactly 10 chars",
			token: "1234567890",
			want:  "1234...7890",
		},
		{
			name:  "normal token",
			token: "abcdefghijklmnop",
			want:  "abcd...mnop",
		},
		{
			name:  "long session token",
			token: "53616c7465645f5f8e91e2ac123456789abcdef",
			want:  "5361...cdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskToken(tt.token)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaskToken_BoundaryConditions(t *testing.T) {
	// 9 chars - should mask
	assert.Equal(t, "****", MaskToken("123456789"))

	// 10 chars - should show partial
	assert.Equal(t, "1234...7890", MaskToken("1234567890"))

	// 11 chars - should show partial
	assert.Equal(t, "1234...890a", MaskToken("1234567890a"))
}

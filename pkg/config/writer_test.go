package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteConfig(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		setup     func(afero.Fs)
		assertion require.ErrorAssertionFunc
	}{
		{
			name:      "write to specified path",
			path:      "/tmp/test/elf.toml",
			setup:     func(_ afero.Fs) {},
			assertion: require.NoError,
		},
		{
			name:      "write creates parent directories",
			path:      "/tmp/nested/deep/dir/elf.toml",
			setup:     func(_ afero.Fs) {},
			assertion: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tt.setup(fs)

			cfg, err := NewConfig(WithFs(fs))
			require.NoError(t, err)

			err = cfg.WriteConfig(tt.path)
			tt.assertion(t, err)

			if err == nil {
				exists, statErr := afero.Exists(fs, tt.path)
				require.NoError(t, statErr)
				assert.True(t, exists, "config file should exist")
			}
		})
	}
}

func TestWriteConfig_EmptyPath(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	err = cfg.WriteConfig("")
	require.NoError(t, err)

	// Should write to default config dir
	expectedPath := filepath.Join(cfg.GetConfigDir(), DefaultConfigFileBase+"."+DefaultConfigExt)
	exists, err := afero.Exists(fs, expectedPath)
	require.NoError(t, err)
	assert.True(t, exists, "config file should exist at default path")
}

func TestSafeWriteConfig_FileExists(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	path := "/tmp/elf.toml"

	// Create existing file
	err = afero.WriteFile(fs, path, []byte("existing"), 0o644)
	require.NoError(t, err)

	// SafeWriteConfig should fail
	err = cfg.SafeWriteConfig(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestSafeWriteConfig_NewFile(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	path := "/tmp/new-elf.toml"

	err = cfg.SafeWriteConfig(path)
	require.NoError(t, err)

	exists, err := afero.Exists(fs, path)
	require.NoError(t, err)
	assert.True(t, exists, "config file should be created")
}

func TestSetToken(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Initially has default token
	assert.Equal(t, defaults[AdventTokenKey], cfg.GetToken())

	// Set new token
	cfg.SetToken("new-test-token")
	assert.Equal(t, "new-test-token", cfg.GetToken())
}

func TestSetValue(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Set language
	cfg.SetValue(LanguageKey, "python")
	assert.Equal(t, "python", cfg.GetLanguage())

	// Set base dir
	cfg.SetValue(AdventDirKey, "my-exercises")
	assert.Equal(t, "my-exercises", cfg.GetBaseDir())
}

func TestGetAllSettings(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	settings := cfg.GetAllSettings()

	assert.NotEmpty(t, settings)
	assert.Contains(t, settings, "language")
	assert.Contains(t, settings, "advent")
}

func TestGenerateDefaultConfig(t *testing.T) {
	config := GenerateDefaultConfig()

	assert.Contains(t, config, "language")
	assert.Contains(t, config, "input-file")
	assert.Contains(t, config, "[advent]")
	assert.Contains(t, config, "token")
}

func TestGenerateDefaultConfig_ContainsDefaults(t *testing.T) {
	config := GenerateDefaultConfig()

	// Check it includes the default values
	assert.Contains(t, config, defaults[LanguageKey])
	assert.Contains(t, config, defaults[InputFileKey])
	assert.Contains(t, config, defaults[AdventDirKey])
}

func TestWriteConfig_PersistsValues(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Set custom values
	cfg.SetToken("my-secret-token")
	cfg.SetValue(LanguageKey, "rust")

	path := "/tmp/persist-test.toml"
	err = cfg.WriteConfig(path)
	require.NoError(t, err)

	// Read the file content
	content, err := afero.ReadFile(fs, path)
	require.NoError(t, err)

	assert.Contains(t, string(content), "my-secret-token")
	assert.Contains(t, string(content), "rust")
}

func TestWriteConfig_DirectoryCreation(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// Deep nested path that doesn't exist
	path := "/a/very/deep/nested/directory/structure/elf.toml"
	err = cfg.WriteConfig(path)
	require.NoError(t, err)

	// Verify parent directory was created
	dir := filepath.Dir(path)
	info, err := fs.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestSafeWriteConfig_EmptyPath(t *testing.T) {
	fs := afero.NewMemMapFs()

	cfg, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	// First call should succeed
	err = cfg.SafeWriteConfig("")
	require.NoError(t, err)

	expectedPath := filepath.Join(cfg.GetConfigDir(), DefaultConfigFileBase+"."+DefaultConfigExt)
	exists, err := afero.Exists(fs, expectedPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Second call should fail (file exists)
	cfg2, err := NewConfig(WithFs(fs))
	require.NoError(t, err)

	err = cfg2.SafeWriteConfig("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGenerateDefaultConfig_UsesSystemDirs(t *testing.T) {
	config := GenerateDefaultConfig()

	// Should contain user config/cache dir paths
	userConfigDir, err := os.UserConfigDir()
	if err == nil {
		assert.Contains(t, config, filepath.Join(userConfigDir, "elf"))
	}

	userCacheDir, err := os.UserCacheDir()
	if err == nil {
		assert.Contains(t, config, filepath.Join(userCacheDir, "elf"))
	}
}

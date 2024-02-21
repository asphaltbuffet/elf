package krampus

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	roBase afero.Fs
	testFs afero.Fs
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Helper()

	// base := afero.NewBasePathFs(afero.NewOsFs(), "testdata")
	base := afero.NewOsFs()
	roBase = afero.NewReadOnlyFs(base)

	return func(t *testing.T) {
		t.Helper()
	}
}

func setupSubTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	testFs = afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())
	f, _ := testFs.Create("fakeFileTmp.toml")
	f.Close()

	return func(t *testing.T) {
		t.Helper()

		// t.Log("teardown sub-test")
	}
}

func TestWithFile(t *testing.T) {
	type args struct {
		f   string
		cfg *Config
	}

	type wants struct {
		filename string
		ext      string
	}

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "no filename given",
			args: args{
				f:   "",
				cfg: &Config{},
			},
			wants: wants{DefaultConfigFileBase + "." + DefaultConfigExt, DefaultConfigExt},
		},
		{
			name: "filename with toml extension",
			args: args{
				f:   "fakeFile.toml",
				cfg: &Config{},
			},
			wants: wants{"fakeFile.toml", "toml"},
		},
		{
			name: "dotfile",
			args: args{
				f:   ".fakeConfigFile",
				cfg: &Config{},
			},
			wants: wants{".fakeConfigFile", DefaultConfigExt},
		},
		{
			name: "filename with many dots",
			args: args{
				f:   "fake.file.with.dots.toml",
				cfg: &Config{},
			},
			wants: wants{"fake.file.with.dots.toml", "toml"},
		},
		{
			name: "filename without extension",
			args: args{
				f:   "fakeFile",
				cfg: &Config{},
			},
			wants: wants{"fakeFile", DefaultConfigExt},
		},
		{
			name: "change existing config values",
			args: args{
				f: "newFakeFile.toml",
				cfg: &Config{
					cfgFile:     "oldFile",
					cfgFileType: "oldExt",
				},
			},
			wants: wants{"newFakeFile.toml", "toml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.args.cfg)

			WithFile(tt.args.f)(tt.args.cfg)

			assert.Equal(t, tt.wants.filename, tt.args.cfg.cfgFile)
			assert.Equal(t, tt.wants.ext, tt.args.cfg.cfgFileType)
		})
	}
}

func TestNewConfig(t *testing.T) {
	type args struct {
		options []func(*Config)
	}

	type wants struct {
		filename    string
		filetype    string
		fsAssertion assert.ValueAssertionFunc
	}

	tests := []struct {
		name      string
		args      args
		wants     wants
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "config in current directory",
			args: args{
				options: []func(*Config){
					WithFile("./fakeFile.toml"),
					WithFs(testFs),
				},
			},
			wants: wants{
				filename:    "fakeFile.toml",
				filetype:    "toml",
				fsAssertion: assert.NotNil,
			},
			assertion: require.NoError,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up testing
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			f, err := testFs.Create("fakeFile.toml")
			require.NoError(t, err)
			f.Close()

			got, err := NewConfig(tt.args.options...)

			tt.assertion(t, err)
			if err == nil {
				wantPath, _ := filepath.Abs(tt.wants.filename)
				assert.Equal(t, wantPath, got.viper.ConfigFileUsed())

				assert.Equal(t, tt.wants.filetype, got.cfgFileType)
				tt.wants.fsAssertion(t, got.GetFs())
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	tfs := afero.NewMemMapFs()

	const (
		wantToken    = "default-placeholder"
		wantLanguage = "go"
		wantBaseDir  = "exercises"
	)

	cfgPath, err := os.UserConfigDir()
	require.NoError(t, err)

	wantConfigDir := filepath.Join(cfgPath, "elf")

	cachePath, err := os.UserCacheDir()
	require.NoError(t, err)

	wantCacheDir := filepath.Join(cachePath, "elf")

	// execute function under test
	got, err := NewConfig(WithFs(tfs))

	require.NoError(t, err)
	if err == nil {
		assert.Empty(t, got.GetConfigFileUsed(), "no config file used")

		assert.Equal(t, wantToken, got.GetToken(), "default token")
		assert.Equal(t, wantLanguage, got.GetLanguage(), "default language")
		assert.Equal(t, wantConfigDir, got.GetConfigDir(), "default config dir")
		assert.Equal(t, wantCacheDir, got.GetCacheDir(), "default cache dir")
		assert.Equal(t, wantBaseDir, got.GetBaseDir(), "default base dir")

		assert.NotNil(t, got.GetLogger(), "default logger should not be nil")
		assert.NotNil(t, got.GetFs(), "default fs should not be nil")
	}
}

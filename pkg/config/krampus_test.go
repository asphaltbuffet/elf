package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestConfig_GetRecipients(t *testing.T) {
	t.Parallel()

	cfg, err := NewConfig(WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)
	cfg.Viper().Set(RecipientsKey.String(), []string{"ssh-ed25519 AAAA... a@b"})

	got := cfg.GetRecipients()
	require.Len(t, got, 1)
	assert.Equal(t, "ssh-ed25519 AAAA... a@b", got[0])
}

func TestNewConfig_WritesLogToFile(t *testing.T) {
	// Point the log dir at a writable temp location and assert a logged line
	// lands in the file, not on the console.
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	cfg, err := NewConfig(WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	cfg.GetLogger().Info("hello-from-test")

	logPath := filepath.Join(stateDir, "elf", "elf.log")
	data, readErr := os.ReadFile(logPath)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "hello-from-test")
}

func TestNewConfig_NoConsoleOutput(t *testing.T) {
	// The core regression guard: constructing config and logging an ERROR must
	// produce nothing on stdout or stderr.
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW                             //nolint:reassign // test needs to intercept console output
	t.Cleanup(func() { os.Stdout, os.Stderr = origOut, origErr }) //nolint:reassign // restore console output

	cfg, err := NewConfig(WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)
	cfg.GetLogger().Error("should-not-appear-on-console")

	_ = outW.Close()
	_ = errW.Close()
	os.Stdout, os.Stderr = origOut, origErr //nolint:reassign // restore console output

	var outBuf, errBuf strings.Builder
	_, _ = io.Copy(&outBuf, outR)
	_, _ = io.Copy(&errBuf, errR)

	assert.Empty(t, outBuf.String(), "nothing should be written to stdout")
	assert.Empty(t, errBuf.String(), "nothing should be written to stderr")
}

func TestNewConfig_LogOpenFailureIsSilent(t *testing.T) {
	// If the log dir cannot be created, NewConfig still succeeds and logging is
	// a no-op — the command must never break.
	// Point XDG_STATE_HOME at a path whose parent is a FILE, so MkdirAll fails.
	blocker := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))
	t.Setenv("XDG_STATE_HOME", blocker) // MkdirAll(blocker/elf) fails: parent is a file

	cfg, err := NewConfig(WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	// Logger is usable (writes go to io.Discard); this must not panic or error.
	cfg.GetLogger().Info("discarded")
}

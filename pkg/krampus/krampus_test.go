package krampus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

			assert.NotNil(t, tt.args.cfg.viper)
			assert.Equal(t, tt.wants.filename, tt.args.cfg.cfgFile)
			assert.Equal(t, tt.wants.ext, tt.args.cfg.cfgFileType)
		})
	}
}

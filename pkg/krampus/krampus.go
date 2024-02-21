package krampus

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	ElfEnvPrefix          string = "ELF"
	DefaultConfigFileBase string = "elf"
	DefaultConfigExt      string = "toml"
)

type Config struct {
	viper       *viper.Viper
	cfgFile     string
	cfgFileType string
	logger      *slog.Logger
	fs          afero.Fs
}

func NewConfig(options ...func(*Config)) (Config, error) {
	cfg := Config{
		viper: viper.New(),
	}

	for _, option := range options {
		option(&cfg)
	}

	// set up fs
	if cfg.fs == nil {
		cfg.fs = afero.NewOsFs()
	}
	cfg.viper.SetFs(cfg.fs)

	// set up logger
	w := os.Stderr
	cfg.logger = slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.StampMilli,
		}),
	)
	slog.SetDefault(cfg.logger)

	// set up viper
	cfg.setViperConfigFile()

	cfg.viper.SetEnvPrefix(ElfEnvPrefix)

	_ = cfg.viper.BindEnv(string(AdventTokenKey), "ELF_ADVENT_TOKEN")
	cfg.viper.SetDefault(string(AdventTokenKey), "default-placeholder")

	cfg.viper.SetDefault(string(AdventUserKey), "")
	cfg.viper.SetDefault(string(EulerDirKey), "problems")
	cfg.viper.SetDefault(string(AdventDirKey), "exercises")

	_ = cfg.viper.BindEnv(string(LanguageKey), "ELF_LANGUAGE")
	cfg.viper.SetDefault(string(LanguageKey), "go")

	configDir, err := os.UserConfigDir()
	if err != nil {
		slog.Error("get default user config dir", "error", tint.Err(err))
		return Config{}, err
	}
	cfg.viper.SetDefault(string(ConfigDirKey), filepath.Join(configDir, "elf"))

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		slog.Error("get default user cache dir", "error", tint.Err(err))
		return Config{}, err
	}

	cfg.viper.SetDefault(string(CacheDirKey), filepath.Join(cacheDir, "elf"))

	// set config sources
	cfg.viper.AddConfigPath(".")

	userCfg, err := os.UserConfigDir()
	if err == nil {
		cfg.viper.AddConfigPath(filepath.Join(userCfg, "elf"))
	}

	err = cfg.viper.ReadInConfig()
	if err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			// only return error if it's not a missing config file
			slog.Error("failed to read config file", "error", err, "config", cfg.cfgFile)
			return Config{}, err
		}

		slog.Warn("no config file found", slog.String("file", cfg.cfgFile), tint.Err(err))
		// return cfg, err
	} else {
		slog.Debug("starting with config file", "config", cfg.viper.ConfigFileUsed())
	}

	return cfg, nil
}

func WithFile(f string) func(*Config) {
	return func(c *Config) {
		file := filepath.Base(f)
		ext := filepath.Ext(f)

		// deal with dotfiles
		// foo -> "foo" + "" (false)
		// foo.bar -> "foo.bar" + ".bar" (false)
		// .foo.bar -> ".foo.bar" + ".bar" (false)
		// .foo.foo -> ".foo.foo" + ".foo" (false)
		// .foo -> ".foo" + ".foo" (true)
		// "" -> "" + "" (true)
		if file == ext {
			ext = ""
		}

		// remove leading dot
		ext = strings.TrimPrefix(ext, ".")

		switch {
		case file != "." && ext == "":
			// filename without extension; use default
			c.cfgFile = file
			c.cfgFileType = DefaultConfigExt

		case file != ".":
			// filename with extension; set type as well
			c.cfgFile = file
			c.cfgFileType = ext

		default:
			// lazy; only support one dot for now
			c.cfgFile = DefaultConfigFileBase + "." + DefaultConfigExt
			c.cfgFileType = DefaultConfigExt
		}
	}
}

func (c *Config) setViperConfigFile() {
	if c.viper == nil {
		panic("viper not initialized")
	}

	if c.cfgFile == "" {
		c.cfgFile = DefaultConfigFileBase + "." + DefaultConfigExt
	}

	c.viper.SetConfigName(c.cfgFile)

	if c.cfgFileType != "" {
		c.viper.SetConfigType(c.cfgFileType)
	}
}

func WithFs(fs afero.Fs) func(*Config) {
	return func(c *Config) {
		c.fs = fs
	}
}

func (c Config) GetConfigFileUsed() string {
	return c.viper.ConfigFileUsed()
}

func (c Config) GetToken() string {
	return c.viper.GetString(string(AdventTokenKey))
}

func (c Config) GetLanguage() string {
	return c.viper.GetString(string(LanguageKey))
}

func (c Config) GetConfigDir() string {
	return c.viper.GetString(string(ConfigDirKey))
}

func (c Config) GetCacheDir() string {
	return c.viper.GetString(string(CacheDirKey))
}

func (c Config) GetLogger() *slog.Logger {
	return c.logger
}

func (c Config) GetFs() afero.Fs {
	return c.fs
}

func (c Config) GetBaseDir() string {
	return c.viper.GetString(string(AdventDirKey))
}

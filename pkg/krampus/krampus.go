package krampus

import (
	"errors"
	"fmt"
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
	_ = cfg.viper.BindEnv(string(LanguageKey), "ELF_LANGUAGE")

	for k, v := range defaults {
		cfg.viper.SetDefault(string(k), v)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		cfg.logger.Error("get default user config dir", "error", tint.Err(err))
		return Config{}, err
	}
	cfg.viper.SetDefault(string(ConfigDirKey), filepath.Join(configDir, "elf"))

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cfg.logger.Error("get default user cache dir", "error", tint.Err(err))
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
			cfg.logger.Error("failed to read config file", "error", err, "config", cfg.cfgFile)
			return Config{}, err
		}

		cfg.logger.Warn("no config file found", slog.String("file", cfg.cfgFile), tint.Err(err))
		// return cfg, err
	} else {
		cfg.logger.Debug("starting with config file", "config", cfg.viper.ConfigFileUsed())
	}

	return cfg, nil
}

// WithFile sets the configuration file and type.
//
// If the file is empty, the default file name and type are used.
func WithFile(f string) func(*Config) {
	return func(c *Config) {
		file := filepath.Base(f)
		ext := filepath.Ext(f)

		// handle dotfiles
		// foo 		-> "foo" + "" 			(false)
		// foo.bar 	-> "foo.bar" + ".bar" 	(false)
		// .foo.bar -> ".foo.bar" + ".bar" 	(false)
		// .foo.foo -> ".foo.foo" + ".foo" 	(false)
		// .foo 	-> ".foo" + ".foo" 		(true)
		// "" 		-> "." + "" 			(false)
		if file == ext {
			ext = ""
		}

		// remove leading dot from extension
		ext = strings.TrimPrefix(ext, ".")

		switch {
		// filepath.Base returns "." for empty path
		case file == ".":
			// no filename; use defaults
			c.cfgFile = fmt.Sprintf("%s.%s", DefaultConfigFileBase, DefaultConfigExt)
			c.cfgFileType = DefaultConfigExt
		case file != "." && ext == "":
			// filename without extension; use default extension
			c.cfgFile = file
			c.cfgFileType = DefaultConfigExt

		case file != "." && ext != "":
			// filename with extension; set type as well
			c.cfgFile = file
			c.cfgFileType = ext

		default:
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

// WithFs sets the file system.
//
// If the file system is nil, a new OS file system is used.
func WithFs(fs afero.Fs) func(*Config) {
	return func(c *Config) {
		c.fs = fs
	}
}

// GetConfigFileUsed returns the configuration file used.
//
// If no configuration file is loaded, an empty string is returned. Failure to read a
// configuration file does not cause an error and will still result in an empty string.
func (c Config) GetConfigFileUsed() string {
	return c.viper.ConfigFileUsed()
}

// GetToken returns the authentication token for downloading exercises.
func (c Config) GetToken() string {
	return c.viper.GetString(string(AdventTokenKey))
}

// GetLanguage returns the configured default implementation language.
//
// If no language is configured, an empty string is returned.
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

func (c Config) GetInputFilename() string {
	return c.viper.GetString(string(InputFileKey))
}

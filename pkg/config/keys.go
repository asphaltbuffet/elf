package config

// Key is a type for configuration keys.
type Key string

const (
	// General configuration keys.

	LanguageKey  Key = "language"   // Configuration key for the the default implementation language.
	ConfigDirKey Key = "config-dir" // Configuration key for application configuration files.
	CacheDirKey  Key = "cache-dir"  // Configuration key for cached application data.
	InputFileKey Key = "input-file" // InputFileKey is the configuration key for the default input file name.

	// Advent of Code configuration keys.

	AdventTokenKey Key = "advent.token" // Configuration key for the Advent of Code auth token.
	AdventUserKey  Key = "advent.user"  // Configuration key for the Advent of Code user name.
	AdventDirKey   Key = "advent.dir"   // Configuration key for the Advent of Code exercise directory.
)

func (k Key) String() string {
	return string(k)
}

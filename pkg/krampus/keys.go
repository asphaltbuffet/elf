package krampus

// ConfigKey is a type for configuration keys.
type ConfigKey string

const (
	// General configuration keys.

	LanguageKey  ConfigKey = "language"   // Configuration key for the the default implementation language.
	ConfigDirKey ConfigKey = "config-dir" // Configuration key for application configuration files.
	CacheDirKey  ConfigKey = "cache-dir"  // Configuration key for cached application data.
	InputFileKey ConfigKey = "input-file" // InputFileKey is the configuration key for the default input file name.

	// Advent of Code configuration keys.

	AdventTokenKey ConfigKey = "advent.token" // Configuration key for the Advent of Code auth token.
	AdventUserKey  ConfigKey = "advent.user"  // Configuration key for the Advent of Code user name.
	AdventDirKey   ConfigKey = "advent.dir"   // Configuration key for the Advent of Code exercise directory.

	// Project Euler configuration keys.

	EulerUserKey ConfigKey = "euler.user" // Configuration key for the Project Euler user name.
	EulerDirKey  ConfigKey = "euler.dir"  // Configuration key for Project Euler problem directory.
)

func (k ConfigKey) String() string {
	return string(k)
}

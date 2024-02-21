package krampus

type ConfigKey string

const (
	// General configuration keys.
	LanguageKey  ConfigKey = "language"
	ConfigDirKey ConfigKey = "config-dir"
	CacheDirKey  ConfigKey = "cache-dir"

	// Advent of Code configuration keys.
	AdventTokenKey ConfigKey = "advent.token"
	AdventUserKey  ConfigKey = "advent.user"
	AdventDirKey   ConfigKey = "advent.dir"

	// Project Euler configuration keys.
	EulerUserKey ConfigKey = "euler.user"
	EulerDirKey  ConfigKey = "euler.dir"
)

func (k ConfigKey) String() string {
	return string(k)
}

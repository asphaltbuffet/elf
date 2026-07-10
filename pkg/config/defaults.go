package config

var defaults = map[Key]string{ //nolint: exhaustive // not all keys have defaults
	InputFileKey:   "input.txt",
	AdventDirKey:   "exercises",
	EulerDirKey:    "euler",
	AdventTokenKey: "default-placeholder",
	LanguageKey:    "go",
	TaskTimeoutKey: "2m",
}

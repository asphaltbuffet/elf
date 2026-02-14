package config

var defaults = map[Key]string{ //nolint: exhaustive // not all keys have defaults
	InputFileKey:   "input.txt",
	AdventDirKey:   "exercises",
	AdventTokenKey: "default-placeholder",
	LanguageKey:    "go",
}

package krampus

var defaults = map[ConfigKey]string{ //nolint: exhaustive // not all keys have defaults
	InputFileKey:   "input.txt",
	AdventDirKey:   "exercises",
	EulerDirKey:    "problems",
	AdventTokenKey: "default-placeholder",
	LanguageKey:    "go",
}

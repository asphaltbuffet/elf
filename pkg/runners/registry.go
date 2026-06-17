package runners

// Register adds a RunnerCreator to the Available registry under the given key.
// It is intended for use at program startup (e.g. from config loading) and in tests.
func Register(key string, rc RunnerCreator) {
	Available[key] = rc
}

// ResetRegistry replaces the Available map with m and returns a function that
// restores the previous map. Intended for use in tests.
func ResetRegistry(m map[string]RunnerCreator) func() {
	prev := Available
	Available = m

	return func() { Available = prev }
}

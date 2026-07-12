package secret

//go:generate stringer -type=Outcome -linecomment

// Outcome records what encrypt or decrypt did with a single file.
type Outcome int

const (
	// Added means the target did not exist and was written.
	Added Outcome = iota // added
	// Skipped means the target already existed and was left untouched.
	Skipped // skipped
	// Replaced means the target already existed and was overwritten.
	Replaced // replaced
)

// Entry pairs a file (path relative to the exercise directory) with its Outcome.
type Entry struct {
	Path    string
	Outcome Outcome
}

// Report is the per-file record of a single encrypt or decrypt run,
// in the order files were considered.
type Report []Entry

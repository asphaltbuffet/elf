package exercise

//go:generate stringer -type=Outcome -linecomment

// Outcome records what the Exercise Scaffold did with a single file.
type Outcome int

const (
	// Added means the file did not exist and was written.
	Added Outcome = iota // added
	// Skipped means the file already existed and the overwrite policy left it untouched.
	Skipped // skipped
	// Replaced means the file already existed and the overwrite policy replaced it.
	Replaced // replaced
)

// Entry pairs a scaffolded file (path relative to the exercise directory) with its Outcome.
type Entry struct {
	Path    string
	Outcome Outcome
}

// Report is the per-file record of a single scaffold run, in the order files were considered.
type Report []Entry

// defaultReportCap pre-sizes a Report for the files a scaffold run considers:
// input.txt, info.json, README.md, and one language solution stub.
const defaultReportCap = 4

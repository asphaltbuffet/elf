package tasks

import "github.com/asphaltbuffet/elf/pkg/protocol"

// TaskStatus represents the outcome of a single task execution.
//
//go:generate stringer -type=TaskStatus --linecomment
type TaskStatus int

// Task status constants representing the result of a task execution.
const (
	StatusInvalid    TaskStatus = iota // Invalid
	StatusPassed                       // Passed
	StatusUnverified                   // Unverified
	StatusFailed                       // Failed
	StatusError                        // Error
	StatusTimeout                      // Timeout
)

// Result holds the output and metadata from a single task execution.
type Result struct {
	ID       string
	Type     TaskType
	Part     protocol.Part
	SubPart  int
	Status   TaskStatus
	Output   string
	Expected string
	Duration float64
}

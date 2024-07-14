package tasks

import "github.com/asphaltbuffet/elf/pkg/runners"

//go:generate stringer -type=TaskStatus --linecomment
type TaskStatus int

const (
	StatusInvalid    TaskStatus = iota // Invalid
	StatusPassed                       // Passed
	StatusUnverified                   // Unverified
	StatusFailed                       // Failed
	StatusError                        // Error
)

type Result struct {
	ID       string
	Type     TaskType
	Part     runners.Part
	SubPart  int
	Status   TaskStatus
	Output   string
	Expected string
	Duration float64
}

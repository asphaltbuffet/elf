package tasks

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

// TaskType represents the type of task to be executed.
type TaskType int

// Task type constants identifying the kind of operation to perform.
//
//go:generate stringer -type=TaskType --linecomment
const (
	Invalid   TaskType = iota // invalid
	Solve                     // solve
	Test                      // test
	Benchmark                 // benchmark
	Visualize                 // visualize
)

// MakeTaskID returns a unique identifier for a task.
//
// Examples:
//
//	MakeTaskID(Test, protocol.PartOne, 1) => "Test.1.1"
//	MakeTaskID(Solve, protocol.PartTwo) => "Solve.2"
func MakeTaskID(name TaskType, part protocol.Part, subparts ...int) string {
	switch name {
	case Benchmark, Test, Visualize:
		if len(subparts) != 1 {
			panic("unexpected subpart")
		}

		return fmt.Sprintf("%s.%d.%d", name, part, subparts[0])

	case Solve:
		if len(subparts) != 0 {
			panic("unexpected subpart")
		}

		return fmt.Sprintf("%s.%d", name, part)
	case Invalid:
		panic("invalid task")
	default:
		panic("unexpected task type")
	}
}

// ParseTaskID parses a task ID string into its component TaskType, Part, and sub-part index.
func ParseTaskID(id string) (TaskType, protocol.Part, int) {
	tokens := strings.Split(id, ".")

	switch t := StringToTaskType(tokens[0]); t {
	case Benchmark, Test, Visualize:
		if len(tokens) != 3 { //nolint:mnd // 2 is the expected length
			break
		}

		p, err := strconv.ParseUint(tokens[1], 10, 8)
		if err != nil {
			break
		}

		n, err := strconv.Atoi(tokens[2])
		if err != nil {
			break
		}

		return t, protocol.Part(p), n

	case Solve:
		if len(tokens) != 2 { //nolint:mnd // 2 is the expected length
			break
		}

		p, err := strconv.ParseUint(tokens[1], 10, 8)
		if err != nil {
			break
		}

		return t, protocol.Part(p), 0

	case Invalid:
		break
	}

	return Invalid, protocol.Part(0), 0
}

// StringToTaskType converts a lowercase task name string to its TaskType constant.
func StringToTaskType(s string) TaskType {
	switch s {
	case "solve":
		return Solve

	case "test":
		return Test

	case "benchmark":
		return Benchmark

	case "visualize":
		return Visualize

	default:
		return Invalid
	}
}

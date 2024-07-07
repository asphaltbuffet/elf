package tasks

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

// TaskType represents the type of task to be executed.
type TaskType int

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=TaskType -linecomment
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
//	MakeTaskID(Test, runners.PartOne, 1) => "Test.1.1"
//	MakeTaskID(Solve, runners.PartTwo) => "Solve.2"
func MakeTaskID(name TaskType, part runners.Part, subparts ...int) string {
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

func ParseTaskID(id string) (TaskType, runners.Part, int) {
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

		return t, runners.Part(p), n

	case Solve:
		if len(tokens) != 2 { //nolint:mnd // 2 is the expected length
			break
		}

		p, err := strconv.ParseUint(tokens[1], 10, 8)
		if err != nil {
			break
		}

		return t, runners.Part(p), 0

	case Invalid:
		break
	}

	return Invalid, runners.Part(0), 0
}

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

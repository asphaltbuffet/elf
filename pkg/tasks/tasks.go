package tasks

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

// TaskType represents the type of task to be executed.
type TaskType int

//go:generate stringer -type=TaskType -linecomment
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
	case Test, Visualize:
		if len(subparts) != 1 {
			panic(fmt.Sprintf("unexpected subpart for %s: %d", name, subparts))
		}

		return fmt.Sprintf("%s.%d.%d", name, part, subparts[0])

	case Benchmark, Solve:
		if len(subparts) != 0 {
			panic(fmt.Sprintf("unexpected subpart for %s: %d", name, subparts))
		}

		return fmt.Sprintf("%s.%d", name, part)

	default:
		panic(fmt.Sprint("unexpected task type:", name))
	}
}

func ParseTaskID(id string) (TaskType, runners.Part, int) {
	tokens := strings.Split(id, ".")

	switch t := StringToTaskType(tokens[0]); t {
	case Test, Visualize:
		if len(tokens) != 3 {
			break
		}

		p, err := strconv.ParseUint(tokens[1], 10, 8)
		if err != nil {
			slog.Error("invalid part type", slog.String("id", id))
			break
		}

		n, err := strconv.Atoi(tokens[2])
		if err != nil {
			slog.Error("invalid sub-test number", slog.String("id", id))
			break
		}

		return t, runners.Part(p), n

	case Solve, Benchmark:
		if len(tokens) != 2 {
			break
		}

		p, err := strconv.ParseUint(tokens[1], 10, 8)
		if err != nil {
			slog.Error("invalid part type", slog.String("id", id))
			break
		}

		return t, runners.Part(p), 0

	case Invalid:
		break
	}

	slog.Error("invalid task type", slog.String("id", id))

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

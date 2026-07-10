package exercise

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

// declaredParts returns the Parts this Exercise actually has, by Kind. A Project
// Euler Problem has a single answer, so it declares only Part One; an AoC Puzzle
// declares both. The solve/test/benchmark driver iterates this set rather than
// assuming two parts, so a Problem never runs a phantom Part Two.
func (e *Exercise) declaredParts() []protocol.Part {
	if e.Kind == KindProblem {
		return []protocol.Part{protocol.PartOne}
	}

	return []protocol.Part{protocol.PartOne, protocol.PartTwo}
}

// makeProblemID builds the identity string for a Project Euler Problem. Unlike
// an AoC ID (year-day), it is a single unpadded number prefixed with the kind.
func makeProblemID(number int) string {
	return fmt.Sprintf("euler-%d", number)
}

// problemSource carries the inputs needed to assemble a Problem Exercise.
type problemSource struct {
	baseDir  string
	language string
	title    string
	number   int
}

// newProblemFromSource assembles a finished Project Euler Problem Exercise in
// memory. It writes no input (a Problem's input is optional and defaults to
// none) and declares only a single empty Part One test placeholder.
func newProblemFromSource(s problemSource) *Exercise {
	return &Exercise{
		ID:       makeProblemID(s.number),
		Kind:     KindProblem,
		Title:    s.title,
		Language: s.language,
		Number:   s.number,
		Path:     filepath.Join(s.baseDir, "euler", strconv.Itoa(s.number)),
		Data: &Data{
			InputData:     "",
			InputFileName: "",
			TestCases: TestCase{
				One: []*Test{{Input: "", Expected: ""}},
				Two: nil,
			},
			Answers: Answer{One: "", Two: ""},
		},
	}
}

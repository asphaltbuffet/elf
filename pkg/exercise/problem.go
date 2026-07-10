package exercise

import "github.com/asphaltbuffet/elf/pkg/protocol"

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

package euler

import (
	"fmt"
	"log/slog"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

type Problem struct {
	ID       int
	Language string
}

func New(id int, lang string) *Problem {
	return &Problem{
		ID:       id,
		Language: lang,
	}
}

func (p *Problem) Solve() error {
	fmt.Println("Solving", p)
	return nil
}

func (p *Problem) String() string {
	if p == nil {
		return "Project Euler: NIL PROBLEM"
	}

	name, ok := runners.RunnerNames[p.Language]
	if !ok {
		name = "INVALID LANGUAGE"
	}

	return fmt.Sprintf("Project Euler: %03d (%s)", p.ID, name)
}

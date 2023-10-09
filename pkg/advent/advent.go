package advent

import (
	"fmt"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

var (
	baseDir = "exercises"
	exDir   = "%d/%02d-%s"
)

type Exercise struct {
	ID       string
	Language string
	Year     int
	Day      int
}

func New(id, lang string) (*Exercise, error) {
	var y, d int

	if n, err := fmt.Sscanf(id, "%d-%d", &y, &d); err != nil || n != 2 {
		return nil, fmt.Errorf("invalid exercise ID: %s", id)
	}

	// allow shorthand for years; we'll validate it's in range later
	if y < 1000 {
		y += 2000
	}

	return &Exercise{
		ID:       fmt.Sprintf("%d-%02d", y, d),
		Language: lang,
		Year:     y,
		Day:      d,
	}, nil
}

func (e *Exercise) SetLanguage(lang string) {
	e.Language = lang
}

func (e *Exercise) Solve() error {
	fmt.Println("Solving", e)
	return nil
}

func (e *Exercise) String() string {
	if e == nil {
		return "Advent of Code: INVALID EXERCISE"
	}

	name, ok := runners.RunnerNames[e.Language]
	if !ok {
		name = "INVALID LANGUAGE"
	}

	return fmt.Sprintf("Advent of Code: %04d-%02d (%s)", e.Year, e.Day, name)
}

package advent

import "fmt"

type Exercise struct {
	ID string

	year     int
	day      int
	language string
}

func New(id, lang string) *Exercise {
	var y, d int

	if _, err := fmt.Sscanf(id, "%d-%d", &y, &d); err != nil {
		panic(err)
	}

	return &Exercise{
		ID:       id,
		language: lang,
		year:     y,
		day:      d,
	}
}

func (e *Exercise) SetLanguage(lang string) {
	e.Language = lang
}

func (e *Exercise) Solve() error {
	return nil
}

func (e *Exercise) String() string {
	return fmt.Sprintf("%d-%2d (%s)", e.year, e.day, e.language)
}

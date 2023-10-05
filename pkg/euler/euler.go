package euler

import "fmt"

type Problem struct {
	ID int

	lang string
}

func New(id int, lang string) *Problem {
	return &Problem{
		ID:   id,
		lang: lang,
	}
}

func (p *Problem) Solve() error {
	return nil
}

func (p *Problem) String() string {
	return fmt.Sprintf("Project Euler: %03d (%s)", p.ID, p.lang)
}

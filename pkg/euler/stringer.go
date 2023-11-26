package euler

import "fmt"

func (p *Problem) String() string {
	if p == nil {
		return "Project Euler: NIL PROBLEM"
	}

	if p.Runner == nil {
		return fmt.Sprintf("Project Euler: %03d (INVALID LANGUAGE)", p.ID)
	}

	return fmt.Sprintf("Project Euler: %03d (%s)", p.ID, p.Runner.String())
}

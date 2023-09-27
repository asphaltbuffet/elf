package exercise

import (
	"fmt"
)

type AdventExercise struct {
	Year     int
	Day      int
	Title    string
	Dir      string
	Path     string
	Language string
}

func (e *AdventExercise) String() string {
	return fmt.Sprintf("%d - %s", e.Day, e.Title)
}

package exercise

import (
	"fmt"
)

type Exercise struct {
	Year  int
	Day   int
	Title string
	Dir   string
	Path  string
}

func (e *Exercise) String() string {
	return fmt.Sprintf("%d - %s", e.Day, e.Title)
}

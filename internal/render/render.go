package render

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/x/term"

	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Renderer consumes a task lifecycle event stream and produces user-facing output.
type Renderer interface {
	Handle(tasks.Event)
	Close() error
}

// Header carries the static chrome a renderer draws above the task rows.
type Header struct {
	Year     int
	Day      int
	Number   int
	Title    string
	Language string
}

// hasMeta reports whether the header carries exercise identity worth drawing a
// box for. An Advent of Code puzzle has a non-zero Year; a Project Euler problem
// has a non-zero Number instead (its Year and Day are zero).
func (h Header) hasMeta() bool {
	return h.Year != 0 || h.Number != 0
}

// title formats the header's exercise identity. Project Euler problems (Number
// non-zero) render as "#21: Amicable Numbers"; Advent of Code puzzles render as
// "2023 Day 1: Trebuchet".
func (h Header) title() string {
	if h.Number != 0 {
		return fmt.Sprintf("#%d: %s", h.Number, h.Title)
	}

	return fmt.Sprintf("%d Day %d: %s", h.Year, h.Day, h.Title)
}

// isTTY reports whether w is a terminal. It returns true only when w is an
// [*os.File] whose file descriptor passes [term.IsTerminal].
func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	return term.IsTerminal(f.Fd())
}

// New returns the appropriate [Renderer] for w. Precedence: jsonOut selects the
// machine-output [JSON] renderer (even if plain is also set). Otherwise, when
// plain is true or w is not a TTY, a [Plain] renderer is returned; else [Live].
func New(w io.Writer, h Header, plain, jsonOut bool) Renderer {
	if jsonOut {
		return NewJSON(w, h)
	}

	if plain || !isTTY(w) {
		return NewPlain(w, h)
	}

	return NewLive(h)
}

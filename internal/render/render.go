package render

import (
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
	Title    string
	Language string
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

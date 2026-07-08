package render

import (
	"context"
	"io"

	tea "charm.land/bubbletea/v2"

	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Run orchestrates a domain operation op while feeding its events to a [Renderer].
//
// Selection: jsonOut runs the synchronous machine-output path (a [JSON]
// renderer). Otherwise, plain=true or a non-TTY w runs the synchronous [Plain]
// path; a TTY runs the [Live] bubbletea path. op receives a callback that emits
// one [tasks.Event] per call; Run returns op's (results, error).
func Run(
	ctx context.Context,
	w io.Writer,
	h Header,
	plain, jsonOut bool,
	op func(cb func(tasks.Event)) ([]tasks.Result, error),
) ([]tasks.Result, error) {
	if jsonOut {
		return runSync(NewJSON(w, h), op)
	}

	if plain || !isTTY(w) {
		return runSync(NewPlain(w, h), op)
	}

	return runLive(ctx, w, h, op)
}

// runSync executes op synchronously, delivering events directly to r.Handle and
// closing r afterward. Used by the Plain and JSON paths (both are buffer-then-
// emit renderers that are not goroutine-driven).
func runSync(
	r Renderer,
	op func(cb func(tasks.Event)) ([]tasks.Result, error),
) ([]tasks.Result, error) {
	results, err := op(r.Handle)
	_ = r.Close()

	return results, err
}

// runLive executes the operation in a goroutine while the [tea.Program] runs on
// the calling goroutine. Events are forwarded via [tea.Program.Send].
func runLive(
	ctx context.Context,
	w io.Writer,
	h Header,
	op func(cb func(tasks.Event)) ([]tasks.Result, error),
) ([]tasks.Result, error) {
	model := NewLive(h)
	p := tea.NewProgram(model, tea.WithContext(ctx), tea.WithOutput(w))

	type opResult struct {
		results []tasks.Result
		err     error
	}

	done := make(chan opResult, 1)

	go func() {
		results, err := op(func(ev tasks.Event) {
			p.Send(eventMsg{ev})
		})
		p.Quit()
		done <- opResult{results, err}
	}()

	if _, progErr := p.Run(); progErr != nil {
		// ErrProgramKilled is expected when the context is cancelled; surface
		// it only if op itself did not produce an error.
		r := <-done
		if r.err == nil {
			r.err = progErr
		}

		return r.results, r.err
	}

	r := <-done

	return r.results, r.err
}

package render

import (
	"context"
	"io"

	tea "charm.land/bubbletea/v2"

	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Run orchestrates a domain operation op while feeding its events to a [Renderer].
//
// op receives a callback cb; each call to cb emits one [tasks.Event]. When op
// returns its (results, error), Run closes the renderer and propagates the
// values to the caller. Exit-code logic is the caller's responsibility.
//
// Plain path (plain=true or w is not a TTY): op runs synchronously on the
// calling goroutine. Events are delivered directly via [Plain.Handle].
//
// Live path: a [tea.Program] is started on the calling goroutine. op runs in a
// separate goroutine and delivers events via [tea.Program.Send] — never by
// calling [Live.Handle] directly, which is not goroutine-safe. When op returns
// it signals the program to quit; Run waits for the program to exit before
// returning.
func Run(
	ctx context.Context,
	w io.Writer,
	h Header,
	plain bool,
	op func(cb func(tasks.Event)) ([]tasks.Result, error),
) ([]tasks.Result, error) {
	if plain || !isTTY(w) {
		return runPlain(w, h, op)
	}

	return runLive(ctx, w, h, op)
}

// runPlain executes the operation synchronously using a [Plain] renderer.
func runPlain(
	w io.Writer,
	h Header,
	op func(cb func(tasks.Event)) ([]tasks.Result, error),
) ([]tasks.Result, error) {
	r := NewPlain(w, h)
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

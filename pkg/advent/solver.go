package advent

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func (e *Exercise) Solve(skipTests bool) error {
	solverLog := slog.With(slog.String("fn", "Solve"), slog.String("exercise", e.Title))
	solverLog.Debug("solving", slog.String("language", e.Language))

	input, err := os.ReadFile(e.Data.InputFile)
	if err != nil {
		solverLog.Error("reading input file", slog.String("path", e.Data.InputFile), tint.Err(err))
		return err
	}

	e.Data.Input = string(input)

	if err = e.runner.Start(); err != nil {
		solverLog.Error("starting runner", slog.String("path", e.Data.InputFile), tint.Err(err))
		return err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.Border{
			Top:    "─",
			Bottom: "─",
		}).
		Foreground(lipgloss.Color("5"))

	fmt.Fprintln(os.Stdout, headerStyle.Render(
		fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)),
	)

	if !skipTests {
		if err = runTests(e.runner, e.Data); err != nil {
			solverLog.Error("running tests", tint.Err(err))
			return err
		}
	}

	if err = runMainTasks(e.runner, e.Data.Input); err != nil {
		solverLog.Error("running main tasks", tint.Err(err))
		return err
	}

	return nil
}

func makeMainID(part runners.Part) string {
	return fmt.Sprintf("main.%d", part)
}

func parseMainID(id string) runners.Part {
	var p runners.Part

	_, err := fmt.Sscanf(id, "main.%d", &p)
	if err != nil {
		slog.Error("parsing main id", slog.Group("task", "id", id), tint.Err(err))
		panic(err)
	}

	return p
}

func runMainTasks(runner runners.Runner, input string) error {
	for part := runners.PartOne; part <= runners.PartTwo; part++ {
		id := makeMainID(part)

		result, err := runner.Run(&runners.Task{
			TaskID: id,
			Part:   part,
			Input:  input,
		})
		if err != nil {
			slog.Error("running main tasks",
				slog.Group("result", "id", result.TaskID, "ok", result.Ok, "output", result.Output),
				tint.Err(err))
			return err
		}

		handleMainResult(os.Stdout, result)
	}

	return nil
}

func handleMainResult(w io.Writer, r *runners.Result) {
	part := parseMainID(r.TaskID)

	name := taskStyle(int(part), -1)

	var status, followUpText lipgloss.Style

	if r.Ok {
		status = mainResultStyle(r.Output, r.Ok)
		followUpText = mainNoteStyle(humanize.SIWithDigits(r.Duration, 1, "s"), r.Ok)
	} else {
		status = mainResultStyle("did not complete", r.Ok)
		followUpText = mainNoteStyle(r.Output, r.Ok)
	}

	slog.Debug("handling main result", slog.Group("result", "id", r.TaskID, "ok", r.Ok, "output", r.Output))

	fmt.Fprintln(w, name, status, followUpText)
}

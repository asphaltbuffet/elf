package advent

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func makeMainID(part runners.Part) string {
	return fmt.Sprintf("main.%d", part)
}

func parseMainID(id string) runners.Part {
	tokens := strings.Split(id, ".")

	p, err := strconv.ParseUint(tokens[1], 10, 8)
	if err != nil {
		panic(err)
	}

	return runners.Part(uint8(p))
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
			return err
		}

		handleMainResult(os.Stdout, result)
	}

	return nil
}

func handleMainResult(w io.Writer, r *runners.Result) {
	part := parseMainID(r.TaskID)

	mainStyle := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("57")).SetString(fmt.Sprintf("Part %d:", part))

	var status, followUpText lipgloss.Style

	if r.Ok {
		status = lipgloss.NewStyle().Bold(true).SetString(r.Output)
		followUpText = lipgloss.NewStyle().
			Faint(true).
			Italic(true).
			Foreground(lipgloss.Color("242")).
			SetString(fmt.Sprintf("in %s", humanize.SIWithDigits(r.Duration, 1, "s")))
	} else {
		status = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("227")).SetString("did not complete")
		followUpText = lipgloss.NewStyle().
			Faint(true).
			Italic(true).
			Foreground(lipgloss.Color("242")).
			SetString(fmt.Sprintf("saying %q", r.Output))
	}

	fmt.Fprintln(w, mainStyle, status, followUpText)
}

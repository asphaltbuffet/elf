package advent

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

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

	fmt.Fprint(w, "Part ")       //nolint:errcheck,gosec // printing to stdout
	fmt.Fprintf(w, "%d: ", part) //nolint:errcheck,gosec // printing to stdout

	if !r.Ok {
		fmt.Fprint(w, "did not complete")
		fmt.Fprintf(w, " saying %q\n", r.Output) //nolint:errcheck,gosec // printing to stdout
	} else {
		fmt.Fprint(w, r.Output)                                               //nolint:errcheck,gosec // printing to stdout
		fmt.Fprintf(w, " in %s\n", humanize.SIWithDigits(r.Duration, 1, "s")) //nolint:errcheck,gosec // printing to stdout
	}
}

package aoc

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/asphaltbuffet/elf/pkg/exercise"
	"github.com/asphaltbuffet/elf/pkg/runners"
)

func (ac *AOCClient) RunExercise(year int, day int, lang string) error {
	e, err := ac.GetExercise(year, day)
	if err != nil {
		return fmt.Errorf("getting exercise: %w", err)
	}

	runner, err := GetRunner(e, lang)
	if err != nil {
		return fmt.Errorf("getting runner: %w", err)
	}

	if runErr := runner.Start(); runErr != nil {
		return fmt.Errorf("starting runner: %w", runErr)
	}

	defer func() {
		_ = runner.Stop()
		_ = runner.Cleanup()
	}()

	info, err := ac.GetExerciseInfo(year, day)
	if err != nil {
		return fmt.Errorf("getting exercise info: %w", err)
	}

	//nolint:errcheck,gosec // printing to stdout
	bold.Printf("%d-%d %s (%s)\n\n",
		year,
		day,
		e.Name,
		runners.RunnerNames[lang],
	)

	fmt.Print("Running...\n\n")

	if ac.RunMode == RunModeTestOnly || ac.RunMode == RunModeAll {
		if runTestErr := runTests(runner, info); runTestErr != nil {
			return runTestErr
		}
	}

	input, err := ac.GetInput(year, day)
	if err != nil {
		return fmt.Errorf("getting input for %d-%d: %w", year, day, err)
	}

	if ac.RunMode == RunModeNoTest || ac.RunMode == RunModeAll {
		if err := runMainTasks(runner, input); err != nil {
			return err
		}
	}

	return fmt.Errorf("not implemented")
}

func GetRunner(e *exercise.Exercise, lang string) (runners.Runner, error) {
	impls, err := e.GetImplementations()
	if err != nil {
		return nil, fmt.Errorf("getting implementations for exercise: %w", err)
	}

	if !slices.Contains(impls, lang) {
		return nil, fmt.Errorf("implementation path not found: %s", filepath.Join(e.Dir, lang))
	}

	runner := runners.Available[lang](e.Dir)

	return runner, nil
}

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
	for part := runners.PartOne; part <= runners.PartTwo; part += 1 {
		id := makeMainID(part)

		result, err := runner.Run(&runners.Task{
			TaskID: id,
			Part:   part,
			Input:  input,
		})
		if err != nil {
			return err
		}

		handleMainResult(result)
	}

	return nil
}

func handleMainResult(r *runners.Result) {
	part := parseMainID(r.TaskID)

	bold.Print("Part ")             //nolint:errcheck,gosec // printing to stdout
	boldYellow.Printf("%d: ", part) //nolint:errcheck,gosec // printing to stdout

	if !r.Ok {
		fmt.Print(incompleteLabel)
		dimmed.Printf(" saying %q\n", r.Output) //nolint:errcheck,gosec // printing to stdout
	} else {
		brightBlue.Print(r.Output)                                           //nolint:errcheck,gosec // printing to stdout
		dimmed.Printf(" in %s\n", humanize.SIWithDigits(r.Duration, 1, "s")) //nolint:errcheck,gosec // printing to stdout
	}
}

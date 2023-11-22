package advent

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func (e *Exercise) Test() error {
	testerLog := slog.With(slog.String("fn", "Solve"), slog.String("exercise", e.Title))
	testerLog.Debug("solving", slog.String("language", e.Language))

	runner := runners.Available[e.Language](e.path)

	if err := runner.Start(); err != nil {
		testerLog.Error("starting runner", slog.String("path", e.Data.InputFile), tint.Err(err))
		return err
	}

	defer func() {
		_ = runner.Stop()
		_ = runner.Cleanup()
	}()

	headerStyle := lipgloss.NewStyle().Bold(true).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("5"))

	fmt.Println(headerStyle.Render(e.String()))

	if err := runTests(runner, e.Data); err != nil {
		testerLog.Error("running tests", tint.Err(err))
		return err
	}

	return nil
}

func makeTestID(part runners.Part, n int) string {
	return fmt.Sprintf("test.%d.%d", part, n)
}

func parseTestID(id string) (runners.Part, int) {
	tokens := strings.Split(id, ".")

	p, err := strconv.ParseUint(tokens[1], 10, 8)
	if err != nil {
		panic(err)
	}

	n, _ := strconv.Atoi(tokens[2])

	return runners.Part(uint8(p)), n
}

func runTests(runner runners.Runner, exInfo *Data) error {
	for i, testCase := range exInfo.TestCases.One {
		id := makeTestID(runners.PartOne, i)

		if testCase.Input == "" && testCase.Expected == "" {
			handleTestResult(&runners.Result{
				TaskID: id,
				Ok:     false,
				Output: "empty input or expected output",
			}, testCase)

			continue
		}

		result, err := runner.Run(&runners.Task{
			TaskID: id,
			Part:   runners.PartOne,
			Input:  testCase.Input,
		})
		if err != nil {
			return err
		}

		handleTestResult(result, testCase)
	}

	for i, testCase := range exInfo.TestCases.Two {
		id := makeTestID(runners.PartTwo, i)

		if testCase.Input == "" && testCase.Expected == "" {
			handleTestResult(&runners.Result{
				TaskID: id,
				Ok:     false,
				Output: "empty input or expected output",
			}, testCase)

			continue
		}

		result, err := runner.Run(&runners.Task{
			TaskID: id,
			Part:   runners.PartTwo,
			Input:  testCase.Input,
		})
		if err != nil {
			return err
		}

		handleTestResult(result, testCase)
	}

	return nil
}

func handleTestResult(r *runners.Result, testCase *Test) {
	part, n := parseTestID(r.TaskID)

	testStyle := lipgloss.NewStyle().
		PaddingLeft(2). //nolint:gomnd // hard-coded padding for now
		Foreground(lipgloss.Color("69")).
		SetString(fmt.Sprintf("Test %d.%d:", part, n))

	passed := r.Output == testCase.Expected
	missing := testCase.Input == "" && testCase.Expected == ""

	var status, followUpText lipgloss.Style

	switch {
	case missing:
		status = lipgloss.NewStyle().Faint(true).SetString("EMPTY")

	case !r.Ok:
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("202")).
			SetString("DID NOT COMPLETE")

		followUpText = lipgloss.NewStyle().
			Faint(true).
			Italic(true).
			Foreground(lipgloss.Color("242")).
			SetString(fmt.Sprintf("saying %q", r.Output))

	case passed:
		status = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("48")).SetString("PASS")

	default:
		status = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("161")).SetString("FAIL")
	}

	if followUpText.String() == "" && !missing {
		followUpText = lipgloss.NewStyle().
			Faint(true).
			Italic(true).
			Foreground(lipgloss.Color("242")).
			SetString(fmt.Sprintf("in %s", humanize.SIWithDigits(r.Duration, 1, "s")))
	}

	fmt.Println(testStyle, status, followUpText)

	if !passed && r.Ok {
		extra := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("161")).
			PaddingLeft(4). //nolint:gomnd // hard-coded padding for now
			SetString(fmt.Sprintf("⤷ expected %q, got %q", testCase.Expected, r.Output))

		fmt.Println(extra)
	}
}

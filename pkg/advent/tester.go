package advent

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

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

	fmt.Print("Test ")           //nolint:errcheck,gosec // printing to stdout
	fmt.Printf("%d.%d", part, n) //nolint:errcheck,gosec // printing to stdout
	fmt.Print(": ")              //nolint:errcheck,gosec // printing to stdout

	passed := r.Output == testCase.Expected
	missing := testCase.Input == "" && testCase.Expected == ""

	var status, followUpText string

	switch {
	case missing:
		status = "EMPTY"

	case !r.Ok:
		status = "DID NOT COMPLETE"
		followUpText = fmt.Sprintf(" saying %q", r.Output)

	case passed:
		status = "PASS"

	default:
		status = "FAIL"
	}

	if followUpText == "" && !missing {
		followUpText = fmt.Sprintf(" in %s", humanize.SIWithDigits(r.Duration, 1, "s"))
	}

	fmt.Print(status)
	fmt.Println(followUpText) //nolint:errcheck,gosec // printing to stdout

	if !passed && r.Ok {
		fmt.Printf(" └ Expected %s, got %s\n", testCase.Expected, r.Output)
	}
}

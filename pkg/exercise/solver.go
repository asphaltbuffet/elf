package exercise

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Solve runs the exercise solution and optionally skips the pre-solve test run.
func (e *Exercise) Solve(skipTests bool) ([]tasks.Result, error) {
	logger := e.logger.With(slog.String("exercise", e.Title))
	logger.Debug("solving", slog.String("language", e.Language))

	results := []tasks.Result{}

	inputFile := filepath.Join(e.Path, e.Data.InputFileName)
	input, err := afero.ReadFile(e.appFs, inputFile)
	if err != nil {
		logger.Error("reading input file", slog.String("path", inputFile), tint.Err(err))
		return nil, err
	}

	e.Data.InputData = string(input)

	if err = e.runner.Start(); err != nil {
		logger.Error("starting runner", tint.Err(err))
		return nil, err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	fmt.Fprintln(e.writer, headerStyle(fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)))

	if !skipTests {
		fmt.Fprintf(e.writer, "Testing (%s)...\n", e.runner)

		var tr []tasks.Result

		tr, err = e.runTests()
		if err != nil {
			logger.Error("running tests", tint.Err(err))
			return nil, err
		}

		results = append(results, tr...)
	}

	fmt.Fprintf(e.writer, "Solving (%s)...\n", e.runner)

	mainResults, err := e.runMainTasks()
	if err != nil {
		logger.Error("running main tasks", tint.Err(err))
		return nil, err
	}

	results = append(results, mainResults...)

	return results, nil
}

func (e *Exercise) runMainTasks() ([]tasks.Result, error) {
	var solveTasks []testTask

	solveTasks = append(solveTasks, makeMainTasks(protocol.PartOne, e.Data)...)
	solveTasks = append(solveTasks, makeMainTasks(protocol.PartTwo, e.Data)...)

	results := make([]tasks.Result, 0, len(solveTasks))

	for _, t := range solveTasks {
		result, err := e.runner.Run(t.task)
		if err != nil {
			return nil, err
		}

		r := handleTaskResult(e.writer, result, t.expected)
		if e.onResult != nil {
			e.onResult(r)
		}
		results = append(results, r)
	}

	return results, nil
}

func makeMainTasks(part protocol.Part, data *Data) []testTask {
	var solveTasks []testTask
	var expected string

	if part == protocol.PartOne {
		expected = data.Answers.One
	} else {
		expected = data.Answers.Two
	}

	solveTasks = append(solveTasks, testTask{
		task: &protocol.Task{
			TaskID:    tasks.MakeTaskID(tasks.Solve, part),
			Part:      part,
			Input:     data.InputData,
			OutputDir: "",
		},
		expected: expected,
	})

	return solveTasks
}

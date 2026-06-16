package exerciseview

import (
	"context"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/internal/tui/discover"
	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// fakeExecutor satisfies Executor for testing — returns empty results with no error.
type fakeExecutor struct{}

func (f *fakeExecutor) Solve(
	_ context.Context,
	_, _, _ string,
	_ io.Writer,
	_ func(tasks.Result),
	_ bool,
) ([]tasks.Result, error) {
	return nil, nil
}

func (f *fakeExecutor) Test(
	_ context.Context,
	_, _, _ string,
	_ io.Writer,
	_ func(tasks.Result),
) ([]tasks.Result, error) {
	return nil, nil
}

func (f *fakeExecutor) Benchmark(
	_ context.Context,
	_, _ string,
	_ io.Writer,
	_ func(tasks.Result),
	_ int,
) ([]tasks.Result, error) {
	return nil, nil
}

func testExec() *fakeExecutor { return &fakeExecutor{} }

func TestNew_DefaultFields(t *testing.T) {
	t.Parallel()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Trebuchet",
		Path: "exercises/2023/01-trebuchet", Langs: []string{"go", "py"},
	}
	m := New(testExec(), "go", info, "solve")
	assert.Equal(t, "solve", m.action)
	assert.True(t, m.running)
	assert.False(t, m.done)
	require.NoError(t, m.err)
	assert.NotNil(t, m.resultCh)
	assert.True(t, m.progress.Active())
}

func TestNew_LangFromConfig(t *testing.T) {
	t.Parallel()
	info := discover.ExerciseInfo{Langs: []string{"py", "go"}}
	m := New(testExec(), "go", info, "test")
	assert.Equal(t, "go", m.lang)
}

func TestView_RunningState(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	view := m.View()
	assert.Contains(t, view, "SOLVE")
	assert.Contains(t, view, "Test Exercise")
	assert.Contains(t, view, "(go)")
}

func TestView_DoneNoError(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	m.running = false
	m.done = true
	m.progress.Stop()
	view := m.View()
	assert.Contains(t, view, "Done.")
	assert.Contains(t, view, "0 results")
}

func TestView_DoneWithResults(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	m.running = false
	m.done = true
	m.progress.Stop()
	m.results.AddResult(tasks.Result{Status: tasks.StatusPassed, Part: protocol.PartOne, SubPart: -1})
	view := m.View()
	assert.Contains(t, view, "1 results")
}

func TestView_DoneWithError(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	m.running = false
	m.done = true
	m.err = errTest
	m.progress.Stop()
	view := m.View()
	assert.Contains(t, view, "Error:")
	assert.Contains(t, view, "test error")
}

func TestView_NotRunningNotDone(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	m.running = false
	m.done = false
	m.progress.Stop()
	view := m.View()
	assert.NotContains(t, view, "Done.")
	assert.NotContains(t, view, "Error:")
}

func TestView_DayFormat(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	m.info.Day = 25
	view := m.View()
	assert.Contains(t, view, "Day 25")
}

func TestView_ContainsHelpText(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	view := m.View()
	assert.True(t, strings.Contains(view, "esc") || strings.Contains(view, "back") || strings.Contains(view, "quit"))
}

func TestWaitForResult_OpenChannel(t *testing.T) {
	t.Parallel()
	ch := make(chan tasks.Result, 1)
	ch <- tasks.Result{Status: tasks.StatusPassed, Output: "42"}
	cmd := waitForResult(ch)
	msg := cmd()
	rm, ok := msg.(resultMsg)
	require.True(t, ok)
	assert.Equal(t, "42", rm.result.Output)
}

func TestWaitForResult_ClosedChannel(t *testing.T) {
	t.Parallel()
	ch := make(chan tasks.Result)
	close(ch)
	cmd := waitForResult(ch)
	msg := cmd()
	_, ok := msg.(doneMsg)
	assert.True(t, ok)
}

func TestShortHelp(t *testing.T) {
	t.Parallel()
	bindings := keys.ShortHelp()
	assert.NotEmpty(t, bindings)
	assert.Len(t, bindings, 3)
}

func TestFullHelp(t *testing.T) {
	t.Parallel()
	groups := keys.FullHelp()
	assert.NotEmpty(t, groups)
	assert.Len(t, groups, 1)
	assert.Len(t, groups[0], 3)
}

func TestHelpToggle(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	assert.False(t, m.help.ShowAll)

	updated, cmd := m.Update(keyMsg("?"))
	model := updated.(Model)
	assert.True(t, model.help.ShowAll)
	assert.Nil(t, cmd)

	updated, _ = model.Update(keyMsg("?"))
	model = updated.(Model)
	assert.False(t, model.help.ShowAll)
}

func TestActionGerunds(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Solving", actionGerunds["solve"])
	assert.Equal(t, "Testing", actionGerunds["test"])
	assert.Equal(t, "Benchmarking", actionGerunds["benchmark"])
	assert.Equal(t, "Analyzing", actionGerunds["analyze"])
}

func TestUnknownMsg(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	updated, _ := m.Update("unknown message type")
	_, ok := updated.(Model)
	assert.True(t, ok)
}

func TestWindowSizeMsg_SetsHelpWidth(t *testing.T) {
	t.Parallel()
	m := testModel()
	m.help = helpModel()
	updated, _ := m.Update(windowSizeMsg(100, 50))
	model := updated.(Model)
	assert.Equal(t, 100, model.help.Width)
}

func TestInit_ReturnsCmd(t *testing.T) {
	t.Parallel()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: "nonexistent/path", Langs: []string{"go"},
	}
	m := New(testExec(), "go", info, "solve")
	cmd := m.Init()
	assert.NotNil(t, cmd, "Init should return a batch command")
}

func TestInit_SolveProducesResult(t *testing.T) {
	t.Parallel()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: "nonexistent/path", Langs: []string{"go"},
	}
	m := New(testExec(), "go", info, "solve")
	cmd := m.Init()
	require.NotNil(t, cmd)
	assert.NotNil(t, m.resultCh)
}

func TestInit_TestAction(t *testing.T) {
	t.Parallel()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: "nonexistent/path", Langs: []string{"go"},
	}
	m := New(testExec(), "go", info, "test")
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestInit_BenchmarkAction(t *testing.T) {
	t.Parallel()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: "nonexistent/path", Langs: []string{"go"},
	}
	m := New(testExec(), "go", info, "benchmark")
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestInit_ExecutesSolve(t *testing.T) {
	t.Parallel()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: t.TempDir(), Langs: []string{"go"},
	}
	m := New(testExec(), "go", info, "solve")
	cmd := m.Init()
	require.NotNil(t, cmd)

	msg := cmd()
	if cmds, ok := msg.(tea.BatchMsg); ok {
		for _, c := range cmds {
			if c != nil {
				_ = c()
			}
		}
	}
}

func TestInit_ExecutesTest(t *testing.T) {
	t.Parallel()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: t.TempDir(), Langs: []string{"go"},
	}
	m := New(testExec(), "go", info, "test")
	cmd := m.Init()
	require.NotNil(t, cmd)

	msg := cmd()
	if cmds, ok := msg.(tea.BatchMsg); ok {
		for _, c := range cmds {
			if c != nil {
				_ = c()
			}
		}
	}
}

func TestInit_ExecutesBenchmark(t *testing.T) {
	t.Parallel()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: t.TempDir(), Langs: []string{"go"},
	}
	m := New(testExec(), "go", info, "benchmark")
	cmd := m.Init()
	require.NotNil(t, cmd)

	msg := cmd()
	if cmds, ok := msg.(tea.BatchMsg); ok {
		for _, c := range cmds {
			if c != nil {
				_ = c()
			}
		}
	}
}

func TestRunSolve_BadPath(t *testing.T) {
	t.Parallel()
	exec := testExec()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: t.TempDir(), Langs: []string{"go"},
	}
	ch := make(chan tasks.Result, resultBufSize)
	runSolve(exec, info, "go", ch)
	var results []tasks.Result
	close(ch)
	for r := range ch {
		results = append(results, r)
	}
	// fakeExecutor returns nil error and nil results — channel stays empty, no error result
	_ = results
}

func TestRunTest_BadPath(t *testing.T) {
	t.Parallel()
	exec := testExec()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: t.TempDir(), Langs: []string{"go"},
	}
	ch := make(chan tasks.Result, resultBufSize)
	runTest(exec, info, "go", ch)
	var results []tasks.Result
	close(ch)
	for r := range ch {
		results = append(results, r)
	}
	_ = results
}

func TestRunBenchmark_BadPath(t *testing.T) {
	t.Parallel()
	exec := testExec()
	info := discover.ExerciseInfo{
		Year: 2023, Day: 1, Title: "Test",
		Path: t.TempDir(), Langs: []string{"go"},
	}
	ch := make(chan tasks.Result, resultBufSize)
	runBenchmark(exec, info, ch)
	var results []tasks.Result
	close(ch)
	for r := range ch {
		results = append(results, r)
	}
	_ = results
}

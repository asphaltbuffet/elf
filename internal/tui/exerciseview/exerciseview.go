// Package exerciseview is the TUI screen that runs and streams results for a single exercise action.
package exerciseview

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/asphaltbuffet/elf/internal/tui/components"
	"github.com/asphaltbuffet/elf/internal/tui/discover"
	"github.com/asphaltbuffet/elf/internal/tui/nav"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Executor is the narrow interface that exerciseview uses to run domain operations.
// *app.App satisfies this interface.
type Executor interface {
	Solve(
		ctx context.Context,
		path, language, customInput string,
		w io.Writer,
		cb func(tasks.Result),
		skipTests bool,
	) ([]tasks.Result, error)
	Test(
		ctx context.Context,
		path, language, customInput string,
		w io.Writer,
		cb func(tasks.Result),
	) ([]tasks.Result, error)
	Benchmark(
		ctx context.Context,
		path, language string,
		w io.Writer,
		cb func(tasks.Result),
		iterations int,
	) ([]tasks.Result, error)
}

type resultMsg struct {
	result tasks.Result
}

type doneMsg struct {
	err error
}

// Model is the exercise view TUI model.
type Model struct {
	exec     Executor
	info     discover.ExerciseInfo
	action   string
	lang     string
	results  components.ResultList
	progress components.Progress
	resultCh chan tasks.Result
	running  bool
	done     bool
	err      error
	width    int
	height   int
	help     help.Model
}

const (
	headerHeight    = 4
	footerHeight    = 3
	resultBufSize   = 32
	defaultBenchItr = 10
)

// New creates a new exercise view model.
func New(exec Executor, lang string, info discover.ExerciseInfo, action string) Model {
	if lang == "" && len(info.Langs) > 0 {
		lang = info.Langs[0]
	}

	return Model{
		exec:     exec,
		info:     info,
		action:   action,
		lang:     lang,
		results:  components.NewResultList(80, 20), //nolint:mnd // initial size, will be resized
		progress: components.StartedProgress(fmt.Sprintf("%s %s...", actionGerunds[action], info.Title)),
		resultCh: make(chan tasks.Result, resultBufSize),
		running:  true,
		help:     help.New(),
	}
}

// Init starts the spinner and launches the exercise action in a goroutine, returning the first result.
func (m Model) Init() tea.Cmd {
	spinCmd := m.progress.InitialTick()

	ch := m.resultCh
	exec := m.exec
	info := m.info
	lang := m.lang
	action := m.action

	runCmd := func() tea.Msg {
		go func() {
			defer close(ch)

			switch action {
			case "solve":
				runSolve(exec, info, lang, ch)
			case "test":
				runTest(exec, info, lang, ch)
			case "benchmark":
				runBenchmark(exec, info, ch)
			}
		}()

		result, ok := <-ch
		if !ok {
			return doneMsg{}
		}

		return resultMsg{result: result}
	}

	return tea.Batch(spinCmd, runCmd)
}

// Update streams results from the channel and handles completion, resize, and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.results.SetSize(msg.Width, msg.Height-headerHeight-footerHeight)

		return m, nil

	case resultMsg:
		m.results.AddResult(msg.result)

		return m, waitForResult(m.resultCh)

	case doneMsg:
		m.running = false
		m.done = true
		m.progress.Stop()

		if msg.err != nil {
			m.err = msg.err
		}

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.help):
			m.help.ShowAll = !m.help.ShowAll

			return m, nil

		case key.Matches(msg, keys.back):
			return m, func() tea.Msg { return nav.PopScreenMsg{} }

		case key.Matches(msg, keys.quit):
			return m, tea.Quit
		}
	}

	cmd := m.progress.Update(msg)
	rlCmd := m.results.Update(msg)

	return m, tea.Batch(cmd, rlCmd)
}

// View renders the result list, progress spinner, and completion status.
func (m Model) View() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Render(
		fmt.Sprintf("Day %d: %s", m.info.Day, m.info.Title),
	)

	actionLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Render(
		strings.ToUpper(m.action),
	)

	langLabel := lipgloss.NewStyle().Faint(true).Render(
		fmt.Sprintf("(%s)", m.lang),
	)

	fmt.Fprintf(&b, "  %s  %s %s\n\n", title, actionLabel, langLabel)

	b.WriteString(m.results.View())
	b.WriteString("\n")

	if m.running {
		b.WriteString(m.progress.View())
		b.WriteString("\n")
	} else if m.done {
		if m.err != nil {
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
			fmt.Fprintf(&b, "  %s %v\n", errStyle.Render("Error:"), m.err)
		} else {
			doneStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
			fmt.Fprintf(&b, "  %s  %d results\n", doneStyle.Render("Done."), m.results.Count())
		}
	}

	b.WriteString("\n  " + m.help.View(keys) + "\n")

	return b.String()
}

func runSolve(exec Executor, info discover.ExerciseInfo, lang string, ch chan<- tasks.Result) {
	cb := func(r tasks.Result) { ch <- r }

	if _, err := exec.Solve(context.Background(), info.Path, lang, "", io.Discard, cb, false); err != nil {
		ch <- tasks.Result{Status: tasks.StatusError, Output: err.Error()}
	}
}

func runTest(exec Executor, info discover.ExerciseInfo, lang string, ch chan<- tasks.Result) {
	cb := func(r tasks.Result) { ch <- r }

	if _, err := exec.Test(context.Background(), info.Path, lang, "", io.Discard, cb); err != nil {
		ch <- tasks.Result{Status: tasks.StatusError, Output: err.Error()}
	}
}

func runBenchmark(exec Executor, info discover.ExerciseInfo, ch chan<- tasks.Result) {
	cb := func(r tasks.Result) { ch <- r }

	if _, err := exec.Benchmark(
		context.Background(),
		info.Path,
		info.Langs[0],
		io.Discard,
		cb,
		defaultBenchItr,
	); err != nil {
		ch <- tasks.Result{Status: tasks.StatusError, Output: err.Error()}
	}
}

func waitForResult(ch <-chan tasks.Result) tea.Cmd {
	return func() tea.Msg {
		result, ok := <-ch
		if !ok {
			return doneMsg{}
		}

		return resultMsg{result: result}
	}
}

var actionGerunds = map[string]string{
	"solve":     "Solving",
	"test":      "Testing",
	"benchmark": "Benchmarking",
	"analyze":   "Analyzing",
}

// keyMap defines the exercise view key bindings and implements key.Map for bubbles/help.
type keyMap struct {
	back, quit, help key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.back, k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.back, k.help, k.quit},
	}
}

var keys = keyMap{
	back: key.NewBinding(
		key.WithKeys("esc", "h"),
		key.WithHelp("esc/h", "back"),
	),
	help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
}

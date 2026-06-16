// Package tui is the root bubbletea application managing the navigation stack.
package tui

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/asphaltbuffet/elf/internal/tui/components"
	"github.com/asphaltbuffet/elf/internal/tui/dashboard"
	"github.com/asphaltbuffet/elf/internal/tui/exerciseview"
	"github.com/asphaltbuffet/elf/internal/tui/nav"
	"github.com/asphaltbuffet/elf/internal/tui/yearview"
	"github.com/asphaltbuffet/elf/pkg/analyze"
	"github.com/asphaltbuffet/elf/pkg/config"
)

// App is the root TUI model managing a navigation stack.
type App struct {
	stack  []tea.Model
	cfg    config.Config
	width  int
	height int
}

// Run creates and starts the TUI program.
func Run(cfg config.Config) error {
	dash := dashboard.New(cfg)

	app := App{
		cfg:   cfg,
		stack: []tea.Model{dash},
	}

	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return nil
}

// Init delegates to the bottom-most screen on the navigation stack.
func (a App) Init() tea.Cmd {
	if len(a.stack) > 0 {
		return a.stack[0].Init()
	}

	return nil
}

// Update handles navigation stack messages (push, pop, analyze) and forwards all others to the top screen.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		if len(a.stack) > 0 {
			var cmd tea.Cmd
			a.stack[len(a.stack)-1], cmd = a.stack[len(a.stack)-1].Update(msg)

			return a, cmd
		}

		return a, nil

	case nav.PushScreenMsg:
		a.stack = append(a.stack, msg.Screen)

		return a, msg.Screen.Init()

	case nav.PopScreenMsg:
		if len(a.stack) > 1 {
			a.stack = a.stack[:len(a.stack)-1]

			var cmd tea.Cmd
			a.stack[len(a.stack)-1], cmd = a.stack[len(a.stack)-1].Update(tea.WindowSizeMsg{
				Width:  a.width,
				Height: a.height,
			})

			return a, cmd
		}

		return a, tea.Quit

	case yearview.ActionMsg:
		ev := exerciseview.New(msg.Cfg, msg.Exercise, msg.Action)
		a.stack = append(a.stack, ev)

		return a, ev.Init()

	case yearview.AnalyzeMsg:
		return a, runAnalyze(msg.Cfg, msg.YearDir)

	case analyzeDoneMsg:
		if msg.err != nil {
			// TODO: show error in status bar
			return a, nil
		}

		return a, tea.Exec(components.NewImageDisplay(msg.path), func(err error) tea.Msg {
			return imageDisplayDoneMsg{err: err}
		})

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	}

	if len(a.stack) > 0 {
		var cmd tea.Cmd
		a.stack[len(a.stack)-1], cmd = a.stack[len(a.stack)-1].Update(msg)

		return a, cmd
	}

	return a, nil
}

type analyzeDoneMsg struct {
	path string
	err  error
}

type imageDisplayDoneMsg struct {
	err error
}

func runAnalyze(cfg config.Config, yearDir string) tea.Cmd {
	return func() tea.Msg {
		outFile := filepath.Join(yearDir, "run-times.png")

		aa, err := analyze.NewAnalyzer(&cfg,
			analyze.WithDirectory(yearDir),
			analyze.WithOutput(outFile),
		)
		if err != nil {
			return analyzeDoneMsg{err: fmt.Errorf("creating analyzer: %w", err)}
		}

		if graphErr := aa.Graph(); graphErr != nil {
			return analyzeDoneMsg{err: fmt.Errorf("generating graph: %w", graphErr)}
		}

		return analyzeDoneMsg{path: outFile}
	}
}

// View renders the top screen on the navigation stack.
func (a App) View() string {
	if len(a.stack) == 0 {
		return "Loading..."
	}

	return a.stack[len(a.stack)-1].View()
}

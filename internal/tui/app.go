package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/asphaltbuffet/elf/internal/tui/dashboard"
	"github.com/asphaltbuffet/elf/internal/tui/exerciseview"
	"github.com/asphaltbuffet/elf/internal/tui/nav"
	"github.com/asphaltbuffet/elf/internal/tui/yearview"
	"github.com/asphaltbuffet/elf/pkg/config"
)

// App is the root TUI model managing a navigation stack and optional modal overlay.
type App struct {
	stack  []tea.Model
	modal  tea.Model
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

func (a App) Init() tea.Cmd {
	if len(a.stack) > 0 {
		return a.stack[0].Init()
	}

	return nil
}

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

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	}

	if a.modal != nil {
		var cmd tea.Cmd
		a.modal, cmd = a.modal.Update(msg)

		return a, cmd
	}

	if len(a.stack) > 0 {
		var cmd tea.Cmd
		a.stack[len(a.stack)-1], cmd = a.stack[len(a.stack)-1].Update(msg)

		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	if len(a.stack) == 0 {
		return "Loading..."
	}

	view := a.stack[len(a.stack)-1].View()

	if a.modal != nil {
		view = a.modal.View()
	}

	return view
}

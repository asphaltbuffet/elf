// Package dashboard is the TUI screen showing the year list with exercise progress.
package dashboard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/asphaltbuffet/elf/internal/tui/discover"
	"github.com/asphaltbuffet/elf/internal/tui/nav"
	"github.com/asphaltbuffet/elf/internal/tui/yearview"
	"github.com/asphaltbuffet/elf/pkg/config"
)

const progressBarWidth = 20

type scanCompleteMsg struct {
	years map[int][]discover.ExerciseInfo
	err   error
}

// Model is the dashboard TUI model.
type Model struct {
	cfg       config.Config
	years     []int
	exercises map[int][]discover.ExerciseInfo
	cursor    int
	width     int
	height    int
	loading   bool
	err       error
	help      help.Model
}

// New creates a new dashboard model.
func New(cfg config.Config) Model {
	return Model{
		cfg:     cfg,
		loading: true,
		help:    help.New(),
	}
}

// Init triggers an async filesystem scan for exercises.
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		result, err := discover.Scan(m.cfg.GetFs(), m.cfg.GetBaseDir())
		return scanCompleteMsg{years: result, err: err}
	}
}

// Update handles the scan result, window resize, and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		return m, nil

	case scanCompleteMsg:
		m.loading = false

		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		m.exercises = msg.years
		m.years = sortedYears(msg.years)

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, keys.down):
			if m.cursor < len(m.years)-1 {
				m.cursor++
			}

		case key.Matches(msg, keys.help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, keys.quit):
			return m, tea.Quit

		case key.Matches(msg, keys.enter, keys.right):
			if len(m.years) > 0 {
				year := m.years[m.cursor]
				yv := yearview.New(m.cfg, year, m.exercises[year])

				return m, func() tea.Msg {
					return nav.PushScreenMsg{Screen: yv}
				}
			}
		}
	}

	return m, nil
}

// View renders the year list with progress bars and config summary.
func (m Model) View() string {
	if m.loading {
		return "  Scanning exercises..."
	}

	if m.err != nil {
		return fmt.Sprintf("  Error: %v\n\n  Press q to quit.", m.err)
	}

	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render("🎄 Elf — Advent of Code Dashboard")

	b.WriteString(title + "\n\n")

	// config summary
	cfgInfo := lipgloss.NewStyle().Faint(true).Render(
		fmt.Sprintf("  Language: %s  •  Exercises: %s", m.cfg.GetLanguage(), m.cfg.GetBaseDir()),
	)
	b.WriteString(cfgInfo + "\n\n")

	if len(m.years) == 0 {
		b.WriteString("  No exercises found.\n")
		fmt.Fprintf(&b, "  Check your exercise directory: %s\n", m.cfg.GetBaseDir())
	} else {
		for i, year := range m.years {
			exercises := m.exercises[year]
			completed := countCompleted(exercises)
			total := len(exercises)

			cursor := "  "
			style := lipgloss.NewStyle()
			if i == m.cursor {
				cursor = "▸ "
				style = style.Bold(true).Foreground(lipgloss.Color("6"))
			}

			progress := renderProgress(completed, total, progressBarWidth)

			line := fmt.Sprintf("%s%s  %s  %d/%d exercises",
				cursor,
				style.Render(strconv.Itoa(year)),
				progress,
				completed,
				total,
			)
			b.WriteString(line + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString("  " + m.help.View(keys) + "\n")

	return b.String()
}

func countCompleted(exercises []discover.ExerciseInfo) int {
	count := 0

	for _, e := range exercises {
		if e.HasP1 && e.HasP2 {
			count++
		}
	}

	return count
}

func renderProgress(completed, total, width int) string {
	if total == 0 {
		return strings.Repeat("░", width)
	}

	filled := completed * width / total

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render(bar)
}

func sortedYears(m map[int][]discover.ExerciseInfo) []int {
	years := make([]int, 0, len(m))

	for y := range m {
		years = append(years, y)
	}

	// sort descending (most recent first)
	for i := range years {
		for j := i + 1; j < len(years); j++ {
			if years[j] > years[i] {
				years[i], years[j] = years[j], years[i]
			}
		}
	}

	return years
}

// keyMap defines the dashboard key bindings and implements key.Map for bubbles/help.
type keyMap struct {
	up, down, quit, enter, right, help key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.up, k.down, k.enter, k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.up, k.down, k.enter, k.right},
		{k.help, k.quit},
	}
}

var keys = keyMap{
	up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "open"),
	),
	help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	quit: key.NewBinding(
		key.WithKeys("q", "esc"),
		key.WithHelp("q", "quit"),
	),
}

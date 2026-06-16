package yearview

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/asphaltbuffet/elf/internal/tui/discover"
	"github.com/asphaltbuffet/elf/internal/tui/nav"
	"github.com/asphaltbuffet/elf/pkg/config"
)

// ActionMsg is sent when the user triggers an action on an exercise.
type ActionMsg struct {
	Action   string // "solve", "test", "benchmark"
	Exercise discover.ExerciseInfo
	Cfg      config.Config
}

// AnalyzeMsg is sent when the user triggers analysis for the year.
type AnalyzeMsg struct {
	YearDir string
	Cfg     config.Config
}

// Model is the year view TUI model showing exercises for a single year.
type Model struct {
	cfg       config.Config
	year      int
	exercises []discover.ExerciseInfo
	cursor    int
	width     int
	height    int
	help      help.Model
}

// New creates a new year view model.
func New(cfg config.Config, year int, exercises []discover.ExerciseInfo) Model {
	return Model{
		cfg:       cfg,
		year:      year,
		exercises: exercises,
		help:      help.New(),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles window resize and key events, dispatching actions to the app via messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, keys.down):
		if m.cursor < len(m.exercises)-1 {
			m.cursor++
		}

	case key.Matches(msg, keys.help):
		m.help.ShowAll = !m.help.ShowAll

	case key.Matches(msg, keys.back):
		return m, popScreen

	case key.Matches(msg, keys.quit):
		return m, tea.Quit

	case key.Matches(msg, keys.solve):
		if len(m.exercises) > 0 {
			return m, m.actionCmd("solve")
		}

	case key.Matches(msg, keys.test):
		if len(m.exercises) > 0 {
			return m, m.actionCmd("test")
		}

	case key.Matches(msg, keys.bench):
		if len(m.exercises) > 0 {
			return m, m.actionCmd("benchmark")
		}

	case key.Matches(msg, keys.analyze):
		if len(m.exercises) > 0 {
			return m, m.analyzeCmd()
		}

	case key.Matches(msg, keys.enter, keys.right):
		if len(m.exercises) > 0 {
			return m, m.actionCmd("solve")
		}
	}

	return m, nil
}

const (
	separatorWidth = 60
	dayColWidth    = 5
	titleColWidth  = 30
	langColWidth   = 10
)

// View renders the exercise table for the selected year.
func (m Model) View() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render(fmt.Sprintf("🎄 Advent of Code %d", m.year))

	b.WriteString(title + "\n\n")

	if len(m.exercises) == 0 {
		b.WriteString("  No exercises found for this year.\n")
	} else {
		// header
		dimStyle := lipgloss.NewStyle().Faint(true)
		header := dimStyle.Render(
			fmt.Sprintf("  %-5s %-30s %-10s %s", "Day", "Title", "Langs", "Status"),
		)
		b.WriteString(header + "\n")

		sep := dimStyle.Render(
			"  " + strings.Repeat("─", separatorWidth),
		)
		b.WriteString(sep + "\n")

		dayCol := lipgloss.NewStyle().Width(dayColWidth).Align(lipgloss.Right)
		titleCol := lipgloss.NewStyle().Width(titleColWidth)
		langCol := lipgloss.NewStyle().Width(langColWidth)

		for i, ex := range m.exercises {
			cursor := "  "
			style := lipgloss.NewStyle()

			if i == m.cursor {
				cursor = "▸ "
				style = style.Bold(true).Foreground(lipgloss.Color("6"))
			}

			langs := strings.Join(ex.Langs, ",")
			if langs == "" {
				langs = "-"
			}

			status := statusIndicator(ex)

			dayText := dayCol.Inherit(style).Render(strconv.Itoa(ex.Day))
			titleText := titleCol.Inherit(style).Render(truncateStr(ex.Title, maxTitleWidth))
			langText := langCol.Inherit(style).Render(langStyle(langs))

			line := fmt.Sprintf("%s%s %s %s %s", cursor, dayText, titleText, langText, status)
			b.WriteString(line + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString("  " + m.help.View(keys) + "\n")

	return b.String()
}

const maxTitleWidth = 28

func truncateStr(s string, limit int) string {
	if len(s) <= limit {
		return s
	}

	return s[:limit-1] + "…"
}

func statusIndicator(ex discover.ExerciseInfo) string {
	p1 := partStatus(ex.HasP1)
	p2 := partStatus(ex.HasP2)

	return fmt.Sprintf("P1:%s P2:%s", p1, p2)
}

func partStatus(solved bool) string {
	if solved {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46")).Render("✓")
	}

	return lipgloss.NewStyle().Faint(true).Render("·")
}

func langStyle(langs string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render(langs)
}

func (m Model) actionCmd(action string) tea.Cmd {
	ex := m.exercises[m.cursor]

	return func() tea.Msg {
		return ActionMsg{
			Action:   action,
			Exercise: ex,
			Cfg:      m.cfg,
		}
	}
}

func (m Model) analyzeCmd() tea.Cmd {
	// Use the parent directory of the first exercise as the year directory.
	yearDir := filepath.Dir(m.exercises[0].Path)

	return func() tea.Msg {
		return AnalyzeMsg{
			YearDir: yearDir,
			Cfg:     m.cfg,
		}
	}
}

func popScreen() tea.Msg {
	return nav.PopScreenMsg{}
}

// keyMap defines the year view key bindings and implements key.Map for bubbles/help.
type keyMap struct {
	up, down, quit, enter, right, back, help key.Binding
	solve, test, bench, analyze              key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.up, k.down, k.solve, k.test, k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.up, k.down, k.enter, k.right},
		{k.solve, k.test, k.bench, k.analyze},
		{k.back, k.help, k.quit},
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
		key.WithHelp("enter", "select"),
	),
	right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "open"),
	),
	back: key.NewBinding(
		key.WithKeys("esc", "h"),
		key.WithHelp("esc/h", "back"),
	),
	solve: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "solve"),
	),
	test: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "test"),
	),
	bench: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "benchmark"),
	),
	analyze: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "analyze"),
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

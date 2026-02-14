package yearview

import (
	"fmt"
	"strconv"
	"strings"

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

// Model is the year view TUI model showing exercises for a single year.
type Model struct {
	cfg       config.Config
	year      int
	exercises []discover.ExerciseInfo
	cursor    int
	width     int
	height    int
}

// New creates a new year view model.
func New(cfg config.Config, year int, exercises []discover.ExerciseInfo) Model {
	return Model{
		cfg:       cfg,
		year:      year,
		exercises: exercises,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

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

	help := lipgloss.NewStyle().Faint(true).Render(
		"  j/k navigate • s solve • t test • b benchmark • esc back • q quit",
	)
	b.WriteString(help + "\n")

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

func popScreen() tea.Msg {
	return nav.PopScreenMsg{}
}

var keys = struct {
	up, down, quit, enter, right, back key.Binding
	solve, test, bench                 key.Binding
}{
	up:    key.NewBinding(key.WithKeys("up", "k")),
	down:  key.NewBinding(key.WithKeys("down", "j")),
	quit:  key.NewBinding(key.WithKeys("q")),
	enter: key.NewBinding(key.WithKeys("enter")),
	right: key.NewBinding(key.WithKeys("right", "l")),
	back:  key.NewBinding(key.WithKeys("esc", "h")),
	solve: key.NewBinding(key.WithKeys("s")),
	test:  key.NewBinding(key.WithKeys("t")),
	bench: key.NewBinding(key.WithKeys("b")),
}

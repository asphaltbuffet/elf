package dashboard

import (
	"fmt"
	"strconv"
	"strings"

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
}

// New creates a new dashboard model.
func New(cfg config.Config) Model {
	return Model{
		cfg:     cfg,
		loading: true,
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		result, err := discover.Scan(m.cfg.GetFs(), m.cfg.GetBaseDir())
		return scanCompleteMsg{years: result, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

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

	help := lipgloss.NewStyle().Faint(true).Render(
		"  j/k navigate • enter open • q quit",
	)
	b.WriteString(help + "\n")

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

// local key bindings referencing the shared keys.
var keys = struct {
	up, down, quit, enter, right key.Binding
}{
	up: key.NewBinding(
		key.WithKeys("up", "k"),
	),
	down: key.NewBinding(
		key.WithKeys("down", "j"),
	),
	quit: key.NewBinding(
		key.WithKeys("q", "esc"),
	),
	enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	right: key.NewBinding(
		key.WithKeys("right", "l"),
	),
}

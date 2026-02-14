package help

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CloseModalMsg signals the app to dismiss the modal overlay.
type CloseModalMsg struct{}

// Model is the help overlay showing keybinding reference.
type Model struct {
	width  int
	height int
}

// New creates a new help overlay model.
func New() Model {
	return Model{}
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
		if key.Matches(msg, closeKeys.close, closeKeys.quit) {
			return m, func() tea.Msg { return CloseModalMsg{} }
		}
	}

	return m, nil
}

const (
	helpBoxWidth  = 50
	sectionGap    = 1
	keyColWidth   = 12
	helpBoxPadVer = 1
	helpBoxPadHor = 2
)

func (m Model) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).MarginTop(sectionGap)
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Width(keyColWidth)
	descStyle := lipgloss.NewStyle().Faint(true)

	var b strings.Builder

	b.WriteString(titleStyle.Render("Keybinding Reference"))
	b.WriteString("\n\n")

	b.WriteString(sectionStyle.Render("Navigation"))
	b.WriteString("\n")

	for _, entry := range navigationBindings {
		b.WriteString("  " + keyStyle.Render(entry.keys) + descStyle.Render(entry.desc) + "\n")
	}

	b.WriteString(sectionStyle.Render("Actions (Year View)"))
	b.WriteString("\n")

	for _, entry := range actionBindings {
		b.WriteString("  " + keyStyle.Render(entry.keys) + descStyle.Render(entry.desc) + "\n")
	}

	b.WriteString(sectionStyle.Render("Global"))
	b.WriteString("\n")

	for _, entry := range globalBindings {
		b.WriteString("  " + keyStyle.Render(entry.keys) + descStyle.Render(entry.desc) + "\n")
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(helpBoxPadVer, helpBoxPadHor).
		Width(helpBoxWidth).
		Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

type bindingEntry struct {
	keys string
	desc string
}

var navigationBindings = []bindingEntry{
	{"↑ / k", "Move up"},
	{"↓ / j", "Move down"},
	{"← / h / esc", "Go back"},
	{"→ / l / enter", "Open / select"},
}

var actionBindings = []bindingEntry{
	{"s", "Solve exercise"},
	{"t", "Test exercise"},
	{"b", "Benchmark exercise"},
	{"a", "Analyze benchmarks"},
}

var globalBindings = []bindingEntry{
	{"?", "Toggle this help"},
	{"q", "Quit"},
	{"ctrl+c", "Force quit"},
}

var closeKeys = struct {
	close, quit key.Binding
}{
	close: key.NewBinding(key.WithKeys("?", "esc")),
	quit:  key.NewBinding(key.WithKeys("q")),
}

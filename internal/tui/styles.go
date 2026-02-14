package tui

import "github.com/charmbracelet/lipgloss"

// Colors reused from pkg/exercise/styles.go palette.
var (
	colorRed     = lipgloss.Color("9")
	colorGreen   = lipgloss.Color("46")
	colorBlue    = lipgloss.AdaptiveColor{Light: "#000080", Dark: "#0000ff"}
	colorCyan    = lipgloss.Color("6")
	colorMagenta = lipgloss.Color("5")
	colorGray    = lipgloss.AdaptiveColor{Light: "#808080", Dark: "#808080"}
)

// TitleStyle renders the application header.
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorMagenta).
	MarginBottom(1)

// SelectedStyle highlights the currently focused row.
var SelectedStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorCyan)

// DimStyle renders inactive or secondary text.
var DimStyle = lipgloss.NewStyle().
	Faint(true).
	Foreground(colorGray)

// StatusBadge returns a styled badge string for a given status.
func StatusBadge(status string) string {
	var style lipgloss.Style

	switch status {
	case "PASS":
		style = lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	case "FAIL", "ERROR":
		style = lipgloss.NewStyle().Bold(true).Foreground(colorRed)
	case "NEW":
		style = lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	default:
		style = lipgloss.NewStyle().Faint(true)
	}

	return style.Render(status)
}

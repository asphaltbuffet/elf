package exercise

import (
	lipgloss "charm.land/lipgloss/v2"
)

// Column width and padding constants for CLI result rendering.
const (
	HeaderWidth int = 40

	// ReportPathWidth is the fixed width of the file-path column in a scaffold report row.
	ReportPathWidth int = 20
	// ReportIndent is the left padding applied to each scaffold report row.
	ReportIndent int = 2
)

var (
	good = lipgloss.Color("#00ff00") // green (dark terminal default)
	warn = lipgloss.Color("#ffff00") // yellow (dark terminal default)

	minor = lipgloss.Color("#808080") // gray
)

// scaffoldOutcomeStyle returns the color styling for a scaffold Outcome:
// green for Added, yellow for Replaced, and a faint gray for Skipped.
func scaffoldOutcomeStyle(o Outcome) lipgloss.Style {
	style := lipgloss.NewStyle()

	switch o {
	case Added:
		style = style.Foreground(good)
	case Replaced:
		style = style.Foreground(warn)
	case Skipped:
		style = style.Faint(true).Foreground(minor)
	}

	return style.SetString(o.String())
}

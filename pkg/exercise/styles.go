package exercise

import (
	"github.com/charmbracelet/lipgloss"
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
	good = lipgloss.AdaptiveColor{Light: "#008000", Dark: "#00ff00"} // green
	warn = lipgloss.AdaptiveColor{Light: "#808000", Dark: "#ffff00"} // yellow

	minor = lipgloss.AdaptiveColor{Light: "#808080", Dark: "#808080"} // gray
)

// func taskHeaderStyle(s string) lipgloss.Style {
// 	headerStyle := lipgloss.NewStyle().
// 		Italic(true).
// 		MarginTop(1).
// 		Foreground(lipgloss.Color("5"))

// 	return headerStyle.SetString(s)
// }

// func mainResultStyle(status string, success bool) lipgloss.Style {
// 	style := lipgloss.NewStyle().Bold(true)

// 	switch {
// 	case status == "":
// 		status = "..."
// 		fallthrough
// 	case !success:
// 		style.Foreground(bad)
// 	case success:
// 		style.Foreground(good)
// 	}

// 	return style.SetString(status)
// }

// func mainNoteStyle(note string, success bool) lipgloss.Style {
// 	style := lipgloss.NewStyle().Faint(true).Italic(true)

// 	if success {
// 		style = style.Foreground(minor).SetString("in", note)
// 	} else {
// 		style = style.Foreground(warn).SetString("saying", note)
// 	}

// 	return style.SetString(note)
// }

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

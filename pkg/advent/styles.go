package advent

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	good = lipgloss.AdaptiveColor{Light: "#008000", Dark: "#00ff00"} // green
	warn = lipgloss.AdaptiveColor{Light: "#808000", Dark: "#ffff00"} // yellow
	bad  = lipgloss.AdaptiveColor{Light: "#800000", Dark: "#ff0000"} // red

	minor = lipgloss.AdaptiveColor{Light: "#808080", Dark: "#808080"} // gray
	// info  = lipgloss.AdaptiveColor{Light: "#000080", Dark: "#0000ff"} // blue.

	theme = lipgloss.AdaptiveColor{Light: "#800080", Dark: "#ff00ff"} // magenta
)

func headerStyle(s string) lipgloss.Style {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		BorderStyle(lipgloss.Border{
			Top:    "─",
			Bottom: "─",
		}).
		Foreground(lipgloss.Color("5"))

	return headerStyle.SetString(s)
}

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

func taskStyle(part, subpart int) lipgloss.Style {
	style := lipgloss.NewStyle().Align(lipgloss.Right).Width(6).Foreground(lipgloss.Color("6"))

	// TODO: return a []style so "Part" is different formatting from numbers
	if subpart >= 0 {
		style = style.SetString(fmt.Sprintf("%d.%d:", part, subpart+1))
	} else {
		style = style.SetString(fmt.Sprintf("%d:", part))
	}

	return style
}

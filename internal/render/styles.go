// Package render provides shared styles and rendering helpers for the elf CLI
// output layer. Styles are expressed in [charm.land/lipgloss/v2].
package render

import (
	"image/color"
	"os"

	lipgloss "charm.land/lipgloss/v2"
)

// Column width and padding constants for CLI result rendering.
const (
	// StatusWidth is the fixed width of the status column (e.g. "✓", "✗").
	StatusWidth = 4
	// TaskWidth is the fixed width of the task-label column (e.g. "  1.1:").
	TaskWidth = 6
	// ExtraPadding is the left padding applied to extra/note lines.
	ExtraPadding = 6
	// TimeWidth is the fixed width of the elapsed-time column.
	TimeWidth = 20
	// HeaderWidth is the fixed width of the AoC header box.
	HeaderWidth = 40
)

// lightDark is the package-level adaptive-color selector, initialised once
// from the real terminal. Renderers may call [SetDarkBackground] to override.
var lightDark = lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))

// SetDarkBackground lets callers (e.g. Bubble Tea renderers) override the
// terminal-background detection with the value received from
// [tea.BackgroundColorMsg].
func SetDarkBackground(isDark bool) {
	lightDark = lipgloss.LightDark(isDark)
}

// Adaptive palette — call as a function with (light, dark) to get the
// appropriate [color.Color] for the current terminal background.

// Good returns green: bright on dark terminals, darker on light ones.
func Good() color.Color {
	return lightDark(lipgloss.Color("#008000"), lipgloss.Color("#00ff00"))
}

// Warn returns yellow: muted on dark terminals, olive on light ones.
func Warn() color.Color {
	return lightDark(lipgloss.Color("#808000"), lipgloss.Color("#ffff00"))
}

// Bad returns ANSI red (color 9) on both light and dark terminals.
func Bad() color.Color {
	return lipgloss.Color("9")
}

// NewAns returns blue: navy on light terminals, bright blue on dark ones.
func NewAns() color.Color {
	return lightDark(lipgloss.Color("#000080"), lipgloss.Color("#0000ff"))
}

// Minor returns gray on both light and dark terminals.
func Minor() color.Color {
	return lipgloss.Color("#808080")
}

// StatusStyle returns a bold, fixed-width style for a status symbol.
func StatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Width(StatusWidth)
}

// ExtraStyle returns an italic, left-padded style for supplementary output.
func ExtraStyle() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true).PaddingLeft(ExtraPadding)
}

// TimeStyle returns a faint, italic, right-aligned style for elapsed time.
func TimeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Faint(true).
		Italic(true).
		Foreground(Minor()).
		Width(TimeWidth).
		Align(lipgloss.Right)
}

// headerStyle returns a bordered, centred header box containing s — the
// AoC year/title banner displayed above solution output.
func headerStyle(s string) lipgloss.Style {
	border := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "─",
		Right:       "─",
		TopLeft:     "─",
		TopRight:    "─",
		BottomLeft:  "─",
		BottomRight: "─",
	}

	return lipgloss.NewStyle().
		Width(HeaderWidth).
		Bold(true).
		Align(lipgloss.Center).
		BorderStyle(border).
		Foreground(lipgloss.Color("5")).
		SetString(s)
}

// TaskStyle returns a fixed-width label style for a task's part and sub-part
// (e.g. "  1.1:"). The label is padded to [TaskWidth] characters.
// The part and subPart parameters are accepted for future label formatting.
func TaskStyle(part int, subPart int) lipgloss.Style {
	_ = part
	_ = subPart

	return lipgloss.NewStyle().Width(TaskWidth)
}

// HeaderStyle is the exported wrapper around [headerStyle].
func HeaderStyle(s string) lipgloss.Style {
	return headerStyle(s)
}

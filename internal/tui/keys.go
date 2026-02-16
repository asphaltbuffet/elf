package tui

import "github.com/charmbracelet/bubbles/key"

// Keys defines the shared key bindings for the TUI.
var Keys = struct {
	Up, Down, Left, Right key.Binding
	Enter, Quit, Help     key.Binding
	Solve, Test, Bench    key.Binding
	Analyze, Download     key.Binding
}{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "back"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "open"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Solve: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "solve"),
	),
	Test: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "test"),
	),
	Bench: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "benchmark"),
	),
	Analyze: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "analyze"),
	),
	Download: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "download"),
	),
}

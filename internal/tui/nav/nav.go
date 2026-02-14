package nav

import tea "github.com/charmbracelet/bubbletea"

// PushScreenMsg requests the app to push a new screen onto the navigation stack.
type PushScreenMsg struct {
	Screen tea.Model
}

// PopScreenMsg requests the app to pop the current screen from the navigation stack.
type PopScreenMsg struct{}

package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Progress shows a spinner with operation name and elapsed time.
type Progress struct {
	spinner   spinner.Model
	operation string
	startTime time.Time
	active    bool
}

// NewProgress creates a new progress spinner component.
func NewProgress() Progress {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	return Progress{
		spinner: s,
	}
}

// Start begins the progress animation and returns the initial tick command.
func (p *Progress) Start(operation string) tea.Cmd {
	p.operation = operation
	p.startTime = time.Now()
	p.active = true

	return p.spinner.Tick
}

// InitialTick returns the spinner tick command without mutating state.
// Use this from value-receiver Init() methods after constructing with StartedProgress.
func (p *Progress) InitialTick() tea.Cmd {
	return p.spinner.Tick
}

// Stop ends the progress animation.
func (p *Progress) Stop() {
	p.active = false
}

// Active returns whether the spinner is currently running.
func (p *Progress) Active() bool {
	return p.active
}

// Update handles spinner tick messages.
func (p *Progress) Update(msg tea.Msg) tea.Cmd {
	if !p.active {
		return nil
	}

	var cmd tea.Cmd
	p.spinner, cmd = p.spinner.Update(msg)

	return cmd
}

// View renders the spinner with operation name and elapsed time.
func (p *Progress) View() string {
	if !p.active {
		return ""
	}

	elapsed := time.Since(p.startTime).Truncate(time.Millisecond)

	timing := lipgloss.NewStyle().Faint(true).Render(elapsed.String())

	return fmt.Sprintf("  %s %s  %s", p.spinner.View(), p.operation, timing)
}

// StartedProgress creates a Progress that is already active with the given operation.
func StartedProgress(operation string) Progress {
	p := NewProgress()
	p.operation = operation
	p.startTime = time.Now()
	p.active = true

	return p
}

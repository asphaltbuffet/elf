package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

const extraPadding = 6

var (
	colorRed   = lipgloss.Color("9")
	colorGreen = lipgloss.Color("46")
	colorBlue  = lipgloss.AdaptiveColor{Light: "#000080", Dark: "#0000ff"}
	colorCyan  = lipgloss.Color("6")
	colorGray  = lipgloss.AdaptiveColor{Light: "#808080", Dark: "#808080"}
)

// ResultList renders a scrollable list of task results.
type ResultList struct {
	viewport viewport.Model
	results  []tasks.Result
	width    int
	height   int
}

// NewResultList creates a new result list component.
func NewResultList(width, height int) ResultList {
	vp := viewport.New(width, height)

	return ResultList{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

// AddResult appends a result and auto-scrolls to bottom.
func (r *ResultList) AddResult(result tasks.Result) {
	r.results = append(r.results, result)
	r.viewport.SetContent(r.renderAll())
	r.viewport.GotoBottom()
}

// SetSize updates the viewport dimensions.
func (r *ResultList) SetSize(width, height int) {
	r.width = width
	r.height = height
	r.viewport.Width = width
	r.viewport.Height = height
	r.viewport.SetContent(r.renderAll())
}

// Update handles viewport messages (scrolling).
func (r *ResultList) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	r.viewport, cmd = r.viewport.Update(msg)

	return cmd
}

// View renders the result list.
func (r *ResultList) View() string {
	return r.viewport.View()
}

// Count returns the number of results.
func (r *ResultList) Count() int {
	return len(r.results)
}

func (r *ResultList) renderAll() string {
	var b strings.Builder

	for _, result := range r.results {
		b.WriteString(renderOneResult(result))
		b.WriteString("\n")
	}

	return b.String()
}

func renderOneResult(result tasks.Result) string {
	badge := statusBadge(result.Status)
	label := taskLabel(result.Part, result.SubPart)

	dur, err := time.ParseDuration(fmt.Sprintf("%fs", result.Duration))
	if err != nil {
		dur = 0
	}

	timing := lipgloss.NewStyle().Faint(true).Foreground(colorGray).Render(dur.String())
	output := strings.TrimSpace(result.Output)

	line := fmt.Sprintf("  %s  %s  %s", label, badge, timing)

	if result.Status == tasks.StatusError ||
		result.Status == tasks.StatusFailed ||
		result.Status == tasks.StatusUnverified {
		extra := lipgloss.NewStyle().Italic(true).PaddingLeft(extraPadding).Render(output)
		line += "\n" + extra
	} else if result.Type == tasks.Solve && result.Status == tasks.StatusPassed {
		extra := lipgloss.NewStyle().
			Faint(true).
			PaddingLeft(extraPadding).
			Render("⤷ " + output)

		line += "\n" + extra
	}

	return line
}

func statusBadge(status tasks.TaskStatus) string {
	switch status {
	case tasks.StatusPassed:
		return lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("PASS")
	case tasks.StatusFailed:
		return lipgloss.NewStyle().Bold(true).Foreground(colorRed).Render("FAIL")
	case tasks.StatusError:
		return lipgloss.NewStyle().Bold(true).Foreground(colorRed).Render("ERROR")
	case tasks.StatusUnverified:
		return lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("NEW")
	case tasks.StatusInvalid:
		return lipgloss.NewStyle().Faint(true).Render("???")
	}

	return lipgloss.NewStyle().Faint(true).Render("???")
}

func taskLabel(part protocol.Part, subpart int) string {
	style := lipgloss.NewStyle().Foreground(colorCyan).Width(6).Align(lipgloss.Right) //nolint:mnd // visual width

	if subpart >= 0 {
		return style.Render(fmt.Sprintf("%d.%d:", part, subpart+1))
	}

	return style.Render(fmt.Sprintf("%d:", part))
}

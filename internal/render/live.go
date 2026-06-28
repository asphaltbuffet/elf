package render

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// tickInterval is the repaint cadence for the live renderer.
const tickInterval = 100 * time.Millisecond

// rowStatus tracks the lifecycle stage of a single task row.
type rowStatus int

const (
	rowNotStarted rowStatus = iota
	rowRunning
	rowDone
)

// row holds the display state for one task.
type row struct {
	id        string
	taskType  tasks.TaskType
	part      int
	subPart   int
	status    rowStatus
	startedAt time.Time
	result    *tasks.Result
}

// eventMsg carries a [tasks.Event] from the driver into the Live model.
type eventMsg struct {
	ev tasks.Event
}

// tickMsg triggers a repaint at [tickInterval].
type tickMsg time.Time

// Live is a bubbletea/v2 model that renders task rows with an animated spinner
// and live elapsed timer for running tasks, settling each row on completion.
type Live struct {
	header   Header
	spinner  spinner.Model
	rows     []*row
	rowIndex map[string]int   // task ID → index into rows
	sections []tasks.TaskType // first-seen ordering of task types
	seenType map[tasks.TaskType]bool
}

// NewLive returns a new [Live] model for the given header.
func NewLive(h Header) *Live {
	return &Live{
		header:   h,
		spinner:  spinner.New(spinner.WithSpinner(spinner.Dot)),
		rows:     nil,
		rowIndex: make(map[string]int),
		sections: nil,
		seenType: make(map[tasks.TaskType]bool),
	}
}

// Init satisfies [tea.Model]. Returns the spinner tick and the repaint tick.
func (m *Live) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) }),
	)
}

// Update satisfies [tea.Model]. Processes event, spinner tick, repaint tick, and quit key.
func (m *Live) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.BackgroundColorMsg:
		SetDarkBackground(typedMsg.IsDark())
		return m, nil

	case eventMsg:
		m.handleEvent(typedMsg.ev)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		return m, tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) })

	case tea.KeyMsg:
		if typedMsg.Key().Code == 'c' && typedMsg.Key().Mod == tea.ModCtrl {
			return m, tea.Quit
		}
		return m, nil
	}

	return m, nil
}

// handleEvent mutates rows in response to a [tasks.Event].
func (m *Live) handleEvent(ev tasks.Event) {
	switch ev.Kind {
	case tasks.EventMeta:
		if ev.Meta != nil {
			m.header = Header(*ev.Meta)
		}

	case tasks.EventPlanned:
		if !m.seenType[ev.Type] {
			m.seenType[ev.Type] = true
			m.sections = append(m.sections, ev.Type)
		}

		idx := len(m.rows)
		m.rows = append(m.rows, &row{
			id:       ev.ID,
			taskType: ev.Type,
			part:     int(ev.Part),
			subPart:  ev.SubPart,
			status:   rowNotStarted,
		})
		m.rowIndex[ev.ID] = idx

	case tasks.EventStarted:
		if idx, ok := m.rowIndex[ev.ID]; ok {
			m.rows[idx].status = rowRunning
			m.rows[idx].startedAt = time.Now()
		}

	case tasks.EventFinished:
		if idx, ok := m.rowIndex[ev.ID]; ok {
			m.rows[idx].status = rowDone
			m.rows[idx].result = ev.Result
		}
	}
}

// View satisfies [tea.Model]. Renders the header, section labels, and all task rows.
func (m *Live) View() tea.View {
	var sb strings.Builder

	// The status column shares one width across the run so columns line up between
	// sections; each section picks its own duration unit (below) so fast tests
	// are not crushed to 0.000s by a multi-second solve. Recomputed each frame
	// (cheap at AoC scale).
	statusWidth := m.statusColumnWidth()

	// Header box — omitted when Year is zero (no exercise metadata available).
	if m.header.Year != 0 {
		title := fmt.Sprintf("%d Day %d: %s", m.header.Year, m.header.Day, m.header.Title)
		sb.WriteString(HeaderStyle(title).String())
		sb.WriteByte('\n')
	}

	// Render each section and its rows.
	for _, sectionType := range m.sections {
		label := sectionLabels[sectionType]
		if label != "" {
			fmt.Fprintf(&sb, "\n%s (%s)…\n", label, m.header.Language)
		}

		unit := pickDurationUnit(m.maxDoneDuration(sectionType))
		for _, r := range m.rows {
			if r.taskType == sectionType {
				m.renderRow(&sb, r, unit, statusWidth)
			}
		}
	}

	return tea.NewView(sb.String())
}

// renderRow writes a single task row: not-started, running (spinner + elapsed),
// or done (status + duration, plus any detail lines such as the NEW answer).
func (m *Live) renderRow(sb *strings.Builder, r *row, unit durationUnit, statusWidth int) {
	taskLabel := fmt.Sprintf("%d.%d:", r.part, r.subPart)
	taskCol := TaskStyle(r.part, r.subPart).SetString(taskLabel).String()

	switch r.status {
	case rowNotStarted:
		notStartedStr := ExtraStyle().SetString("<not started>").String()
		fmt.Fprintf(sb, "%s%s\n", taskCol, notStartedStr)

	case rowRunning:
		spinnerStr := m.spinner.View()
		elapsedStr := TimeStyle().SetString(formatElapsed(time.Since(r.startedAt))).String()
		fmt.Fprintf(sb, "%s%s%s\n", taskCol, spinnerStr, elapsedStr)

	case rowDone:
		if r.result == nil {
			return
		}

		statusCol, extraLines := resultStatusCols(r.result, statusWidth)
		timeCol := TimeStyle().SetString(formatDuration(secondsToDuration(r.result.Duration), unit)).String()
		fmt.Fprintf(sb, "%s%s%s\n", taskCol, statusCol, timeCol)

		// Detail lines (the NEW answer, or got/expected on failure) carry the
		// solve answer the user submits, so Live shows them too, not just Plain.
		for _, line := range extraLines {
			sb.WriteString(ExtraStyle().SetString(line).String())
			sb.WriteByte('\n')
		}
	}
}

// maxDoneDuration returns the largest duration among settled rows of the given
// section type, used to pick that section's display unit.
func (m *Live) maxDoneDuration(sectionType tasks.TaskType) time.Duration {
	var maxDur time.Duration
	for _, r := range m.rows {
		if r.taskType == sectionType && r.status == rowDone && r.result != nil {
			if d := secondsToDuration(r.result.Duration); d > maxDur {
				maxDur = d
			}
		}
	}

	return maxDur
}

// statusColumnWidth returns a width wide enough for the longest settled status
// label, never narrower than [StatusWidth], so time columns align.
func (m *Live) statusColumnWidth() int {
	width := StatusWidth
	for _, r := range m.rows {
		if r.status == rowDone && r.result != nil {
			if n := len(statusLabel(r.result.Status)); n > width {
				width = n
			}
		}
	}

	return width
}

// Handle satisfies [Renderer]. Wraps the event in an [eventMsg] for the driver
// to push into the tea.Program. Since Live is also the tea.Model, the driver
// calls p.Send(eventMsg{ev}) directly; this method is provided for interface
// compatibility and should NOT be called outside of tests.
func (m *Live) Handle(ev tasks.Event) {
	m.handleEvent(ev)
}

// Close satisfies [Renderer]. It is a no-op; the driver owns tea.Program lifecycle.
func (m *Live) Close() error {
	return nil
}

// formatElapsed formats d as a coarse human-readable string.
// For durations under one minute: "%.1fs" (e.g. "23.1s").
// For one minute or more: "%dm%02ds" (e.g. "1m05s").
func formatElapsed(d time.Duration) string {
	const secondsPerMinute = 60

	secs := d.Seconds()
	if secs < secondsPerMinute {
		return fmt.Sprintf("%.1fs", secs)
	}

	mins := int(secs) / secondsPerMinute
	rem := int(secs) % secondsPerMinute

	return fmt.Sprintf("%dm%02ds", mins, rem)
}

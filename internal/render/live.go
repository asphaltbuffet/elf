package render

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
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

// barKey groups benchmark iteration events into one progress bar per
// (language, part). See ADR-0011.
type barKey struct {
	lang string
	part int
}

// bar is the live progress of one (runner, part) benchmark: total iterations
// planned, how many have finished, the summed finished durations (in seconds),
// and when the first iteration started (for the running wall-clock metric).
type bar struct {
	lang      string
	part      int
	planned   int
	completed int
	sumSecs   float64
	startedAt time.Time
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
	progress progress.Model // shared bar renderer; ViewAs is stateless per fill
	rows     []*row
	rowIndex map[string]int   // task ID → index into rows
	bars     []*bar           // benchmark progress bars, one per (language, part)
	barIndex map[barKey]int   // (language, part) → index into bars
	sections []tasks.TaskType // first-seen ordering of task types
	seenType map[tasks.TaskType]bool
}

// NewLive returns a new [Live] model for the given header.
func NewLive(h Header) *Live {
	return &Live{
		header:  h,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		// Fixed-width bar; we render our own "n/m" count, so suppress the
		// component's percentage suffix. Full blocks (not the default half-block
		// '▌') avoid the sub-cell rendering that stripes in many terminals.
		progress: progress.New(
			progress.WithWidth(barWidth),
			progress.WithoutPercentage(),
			progress.WithFillCharacters('█', '░'),
		),
		rows:     nil,
		rowIndex: make(map[string]int),
		bars:     nil,
		barIndex: make(map[barKey]int),
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
		m.markSection(ev.Type)

		if ev.Type == tasks.Benchmark {
			m.barFor(ev.Language, int(ev.Part)).planned++
			return
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
		if ev.Type == tasks.Benchmark {
			b := m.barFor(ev.Language, int(ev.Part))
			if b.startedAt.IsZero() {
				b.startedAt = time.Now()
			}
			return
		}

		if idx, ok := m.rowIndex[ev.ID]; ok {
			m.rows[idx].status = rowRunning
			m.rows[idx].startedAt = time.Now()
		}

	case tasks.EventFinished:
		if ev.Type == tasks.Benchmark {
			b := m.barFor(ev.Language, int(ev.Part))
			b.completed++
			if ev.Result != nil {
				b.sumSecs += ev.Result.Duration
			}
			return
		}

		if idx, ok := m.rowIndex[ev.ID]; ok {
			m.rows[idx].status = rowDone
			m.rows[idx].result = ev.Result
		}
	}
}

// markSection records the first appearance of a task type so sections render in
// arrival order.
func (m *Live) markSection(t tasks.TaskType) {
	if !m.seenType[t] {
		m.seenType[t] = true
		m.sections = append(m.sections, t)
	}
}

// barFor returns the progress bar for a (language, part), creating it on first
// reference.
func (m *Live) barFor(lang string, part int) *bar {
	key := barKey{lang: lang, part: part}
	if idx, ok := m.barIndex[key]; ok {
		return m.bars[idx]
	}

	b := &bar{lang: lang, part: part}
	m.barIndex[key] = len(m.bars)
	m.bars = append(m.bars, b)

	return b
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
		// Benchmark renders as one progress bar per (runner, part), not rows.
		if sectionType == tasks.Benchmark {
			m.renderBars(&sb)
			continue
		}

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

// barWidth is the number of glyph cells in a progress bar.
const barWidth = 20

// barLabelWidth pads the "<lang> P<part>" column so bars across runners align.
const barLabelWidth = 12

// renderBars writes the benchmark section: a label and one progress bar per
// (runner, part), in first-seen order. A running bar shows live wall-clock
// elapsed; a completed bar shows the summed iteration durations (ADR-0011).
func (m *Live) renderBars(sb *strings.Builder) {
	label := sectionLabels[tasks.Benchmark]
	if label != "" {
		fmt.Fprintf(sb, "\n%s…\n", label)
	}

	for _, b := range m.bars {
		var ratio float64
		if b.planned > 0 {
			ratio = float64(b.completed) / float64(b.planned)
		}

		var metric string
		if b.completed >= b.planned && b.planned > 0 {
			// Settled: summed iteration durations.
			unit := pickDurationUnit(secondsToDuration(b.sumSecs))
			metric = formatDuration(secondsToDuration(b.sumSecs), unit)
		} else if !b.startedAt.IsZero() {
			// Running: live wall-clock elapsed.
			metric = formatElapsed(time.Since(b.startedAt))
		}

		labelCol := fmt.Sprintf("%-*s", barLabelWidth, fmt.Sprintf("%s P%d", b.lang, b.part))
		countCol := fmt.Sprintf("%d/%d", b.completed, b.planned)
		fmt.Fprintf(sb, "%s %s %s %s\n",
			labelCol, m.progress.ViewAs(ratio), countCol, TimeStyle().SetString(metric).String())
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

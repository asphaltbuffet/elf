package render

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"time"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// sectionLabels maps a [tasks.TaskType] to its human-readable section label prefix.
var sectionLabels = map[tasks.TaskType]string{
	tasks.Invalid:   "",
	tasks.Test:      "Testing",
	tasks.Solve:     "Solving",
	tasks.Benchmark: "Benchmarking",
	tasks.Visualize: "Visualizing",
}

// section groups the finished results of one [tasks.TaskType] in arrival order.
type section struct {
	typ     tasks.TaskType
	results []*tasks.Result
	// langs[i] is the runner name for results[i], captured from the Finished
	// event because tasks.Result does not carry it (ADR-0011). Empty for
	// Solve/Test; set for Benchmark so Close can group by (runner, part).
	langs []string
}

// Plain is a non-TTY renderer. It buffers finished results per section and emits
// them on Close so the whole run can share one duration unit and aligned
// columns. It ignores [tasks.EventPlanned] and [tasks.EventStarted].
type Plain struct {
	w        io.Writer
	header   Header
	sections []section
	index    map[tasks.TaskType]int // typ -> sections index
}

// NewPlain returns a new [Plain] renderer that writes to w. When w is not a TTY,
// output is wrapped with [colorprofile.NewWriter] to strip ANSI escape sequences.
// Rows are buffered and emitted on Close.
func NewPlain(w io.Writer, h Header) *Plain {
	if !isTTY(w) {
		w = colorprofile.NewWriter(w, os.Environ())
	}

	return &Plain{
		w:      w,
		header: h,
		index:  make(map[tasks.TaskType]int),
	}
}

// Handle processes a single lifecycle event. Meta updates the header/language;
// Finished buffers a result. Planned and Started are ignored.
func (p *Plain) Handle(e tasks.Event) {
	switch e.Kind {
	case tasks.EventMeta:
		if e.Meta != nil {
			p.header = Header(*e.Meta)
		}

	case tasks.EventFinished:
		if e.Result == nil {
			return
		}

		i, ok := p.index[e.Type]
		if !ok {
			i = len(p.sections)
			p.index[e.Type] = i
			p.sections = append(p.sections, section{typ: e.Type})
		}
		p.sections[i].results = append(p.sections[i].results, e.Result)
		p.sections[i].langs = append(p.sections[i].langs, e.Language)

	case tasks.EventPlanned, tasks.EventStarted:
		// Plain output is settled-only; progress events are ignored.
	}
}

// Close emits the buffered header, sections, and result rows. Each section
// chooses its own duration unit from its own largest value (so fast tests are
// not crushed to 0.000s by a multi-second solve), while the status column shares
// one width across the run so the columns line up between sections.
func (p *Plain) Close() error {
	p.printHeader()

	statusWidth := p.statusColumnWidth()

	for _, s := range p.sections {
		// Benchmark collapses to one aggregated line per (runner, part) rather
		// than one line per iteration (ADR-0011).
		if s.typ == tasks.Benchmark {
			p.printBenchmarkSection(s)
			continue
		}

		if label, ok := sectionLabels[s.typ]; ok && label != "" {
			fmt.Fprintf(p.w, "\n%s (%s)…\n", label, p.header.Language)
		}

		unit := pickDurationUnit(maxResultDuration(s.results))
		for _, r := range s.results {
			p.printResult(r, unit, statusWidth)
		}
	}

	return nil
}

// benchAgg accumulates a (runner, part) group's iteration count and summed
// duration for the Plain benchmark summary.
type benchAgg struct {
	lang    string
	part    int
	count   int
	sumSecs float64
}

// printBenchmarkSection writes one settled line per (runner, part): the runner
// name, part, iteration count, and summed duration. Groups are emitted in
// first-seen order so output is deterministic.
func (p *Plain) printBenchmarkSection(s section) {
	if label, ok := sectionLabels[s.typ]; ok && label != "" {
		fmt.Fprintf(p.w, "\n%s…\n", label)
	}

	type key struct {
		lang string
		part int
	}

	order := make([]key, 0)
	groups := make(map[key]*benchAgg)

	for i, r := range s.results {
		k := key{lang: s.langs[i], part: int(r.Part)}

		g, ok := groups[k]
		if !ok {
			g = &benchAgg{lang: k.lang, part: k.part}
			groups[k] = g
			order = append(order, k)
		}

		g.count++
		g.sumSecs += r.Duration
	}

	for _, k := range order {
		g := groups[k]
		unit := pickDurationUnit(secondsToDuration(g.sumSecs))
		fmt.Fprintf(p.w, "%s P%d: %d iterations, %s\n",
			g.lang, g.part, g.count, formatDuration(secondsToDuration(g.sumSecs), unit))
	}
}

// maxResultDuration returns the largest duration among the given results.
func maxResultDuration(results []*tasks.Result) time.Duration {
	var maxDur time.Duration
	for _, r := range results {
		if d := secondsToDuration(r.Duration); d > maxDur {
			maxDur = d
		}
	}

	return maxDur
}

// statusColumnWidth returns a width wide enough for the longest status label
// seen, never narrower than [StatusWidth], so every row's time column aligns.
func (p *Plain) statusColumnWidth() int {
	width := StatusWidth
	for _, s := range p.sections {
		for _, r := range s.results {
			if n := len(statusLabel(r.Status)); n > width {
				width = n
			}
		}
	}

	return width
}

// printHeader renders the AoC title box. When Year is zero (no exercise
// metadata) the box is omitted.
func (p *Plain) printHeader() {
	if p.header.Year == 0 {
		return
	}

	title := fmt.Sprintf("%d Day %d: %s", p.header.Year, p.header.Day, p.header.Title)
	fmt.Fprintln(p.w, HeaderStyle(title).String())
}

// printResult writes one settled output line for r using the run-wide duration
// unit and status-column width.
func (p *Plain) printResult(r *tasks.Result, unit durationUnit, statusWidth int) {
	taskLabel := fmt.Sprintf("%d.%d:", r.Part, r.SubPart)
	taskCol := TaskStyle(int(r.Part), r.SubPart).SetString(taskLabel).String()

	statusCol, extraLines := resultStatusCols(r, statusWidth)

	timeCol := TimeStyle().SetString(formatDuration(secondsToDuration(r.Duration), unit)).String()

	fmt.Fprintf(p.w, "%s%s%s\n", taskCol, statusCol, timeCol)

	for _, line := range extraLines {
		fmt.Fprintln(p.w, ExtraStyle().SetString(line).String())
	}
}

// statusLabel returns the human-readable label for a status (e.g. "PASS",
// "TIMEOUT"). It is the single source of truth for label text so renderers can
// size the status column to the longest label across a run.
func statusLabel(s tasks.TaskStatus) string {
	switch s {
	case tasks.StatusPassed:
		return "PASS"
	case tasks.StatusFailed:
		return "FAIL"
	case tasks.StatusError:
		return "ERROR"
	case tasks.StatusTimeout:
		return "TIMEOUT"
	case tasks.StatusUnverified:
		return "NEW"
	case tasks.StatusInvalid:
		return s.String()
	default:
		return s.String()
	}
}

// statusColor returns the foreground color for a status.
func statusColor(s tasks.TaskStatus) color.Color {
	switch s {
	case tasks.StatusPassed:
		return Good()
	case tasks.StatusFailed, tasks.StatusError:
		return Bad()
	case tasks.StatusTimeout:
		return Warn()
	case tasks.StatusUnverified:
		return NewAns()
	case tasks.StatusInvalid:
		return Minor()
	default:
		return Minor()
	}
}

// resultExtraLines returns the indented detail lines shown below a result row
// (got/expected on failure, output on error/new).
func resultExtraLines(r *tasks.Result) []string {
	var extra []string

	switch r.Status {
	case tasks.StatusFailed:
		if r.Output != "" {
			extra = append(extra, fmt.Sprintf("got:      %s", r.Output))
		}
		if r.Expected != "" {
			extra = append(extra, fmt.Sprintf("expected: %s", r.Expected))
		}
	case tasks.StatusError, tasks.StatusUnverified:
		if r.Output != "" {
			extra = append(extra, r.Output)
		}
	case tasks.StatusInvalid, tasks.StatusPassed, tasks.StatusTimeout:
		// no extra detail
	}

	return extra
}

// resultStatusCols returns the rendered status column (sized to width) and any
// extra output lines for a result.
func resultStatusCols(r *tasks.Result, width int) (string, []string) {
	col := lipgloss.NewStyle().
		Bold(true).
		Width(width).
		Foreground(statusColor(r.Status)).
		SetString(statusLabel(r.Status)).
		String()

	return col, resultExtraLines(r)
}

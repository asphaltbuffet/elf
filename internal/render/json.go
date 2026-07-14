package render

import (
	"encoding/json"
	"io"

	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// jsonResult is one settled result in the machine-output summary. Benchmark
// entries set Runner and Iterations; other task types leave them zero.
type jsonResult struct {
	Part       int     `json:"part"`
	SubPart    int     `json:"sub_part,omitempty"`
	Status     string  `json:"status"`
	Output     string  `json:"output,omitempty"`
	Expected   string  `json:"expected,omitempty"`
	Duration   float64 `json:"duration"`
	Runner     string  `json:"runner,omitempty"`
	Iterations int     `json:"iterations,omitempty"`
}

// jsonSummary is the single object emitted per run. Kind names the challenge
// family ("aoc"/"euler") so consumers branch on it rather than sniffing which
// identity fields are present. Identity fields are omitted when zero: an Advent
// of Code puzzle carries Year/Day, a Project Euler problem carries Number
// instead. Title is present for either.
type jsonSummary struct {
	Kind    string       `json:"kind,omitempty"`
	Year    int          `json:"year,omitempty"`
	Day     int          `json:"day,omitempty"`
	Number  int          `json:"number,omitempty"`
	Title   string       `json:"title,omitempty"`
	Results []jsonResult `json:"results"`
}

// finished pairs a buffered result with its runner name and task type, so Close
// can format non-benchmark results directly and aggregate benchmark results.
type finished struct {
	typ    tasks.TaskType
	lang   string
	result *tasks.Result
}

// JSON is a machine-output renderer. It buffers finished results in Handle and
// marshals one jsonSummary in Close — the same buffer-then-emit shape as Plain.
// It ignores Planned and Started events.
type JSON struct {
	w      io.Writer
	header Header
	items  []finished
}

// NewJSON returns a JSON renderer writing to w.
func NewJSON(w io.Writer, h Header) *JSON {
	return &JSON{w: w, header: h}
}

// Handle buffers finished results and updates the header from Meta. Planned and
// Started events are ignored (settled-only output).
func (j *JSON) Handle(e tasks.Event) {
	switch e.Kind {
	case tasks.EventMeta:
		if e.Meta != nil {
			j.header = Header(*e.Meta)
		}
	case tasks.EventFinished:
		if e.Result == nil {
			return
		}
		j.items = append(j.items, finished{typ: e.Type, lang: e.Language, result: e.Result})
	case tasks.EventPlanned, tasks.EventStarted:
		// settled-only
	}
}

// Close marshals the buffered results as one jsonSummary and writes it to w.
// Benchmark results aggregate per (runner, part); other task types emit one
// entry per result.
func (j *JSON) Close() error {
	s := jsonSummary{Results: []jsonResult{}}
	if j.header.hasMeta() {
		s.Kind = j.header.Kind
		s.Year = j.header.Year
		s.Day = j.header.Day
		s.Number = j.header.Number
		s.Title = j.header.Title
	}

	type benchKey struct {
		lang string
		part int
	}
	benchOrder := make([]benchKey, 0)
	benchGroups := make(map[benchKey]*jsonResult)

	for _, it := range j.items {
		r := it.result
		if it.typ == tasks.Benchmark {
			k := benchKey{lang: it.lang, part: int(r.Part)}
			g, ok := benchGroups[k]
			if !ok {
				g = &jsonResult{Part: k.part, Runner: it.lang, Status: statusLabel(r.Status)}
				benchGroups[k] = g
				benchOrder = append(benchOrder, k)
			}
			g.Iterations++
			g.Duration += r.Duration
			continue
		}

		s.Results = append(s.Results, jsonResult{
			Part:     int(r.Part),
			SubPart:  r.SubPart,
			Status:   statusLabel(r.Status),
			Output:   r.Output,
			Expected: r.Expected,
			Duration: r.Duration,
		})
	}

	for _, k := range benchOrder {
		s.Results = append(s.Results, *benchGroups[k])
	}

	enc := json.NewEncoder(j.w)
	enc.SetIndent("", "  ")

	return enc.Encode(s)
}

package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// TestPlainNoANSIOnNonTTY asserts that writing to a [bytes.Buffer] (non-TTY)
// produces no ANSI escape sequences in the output.
func TestPlainNoANSIOnNonTTY(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{Year: 2015, Day: 4, Title: "The Ideal Stocking Stuffer", Language: "Go"})

	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.1", Type: tasks.Test, Part: protocol.PartOne, SubPart: 1,
		Status: tasks.StatusPassed, Duration: 0.071143,
	}, ""))
	_ = p.Close()

	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Errorf("output contains ANSI escape sequences on non-TTY writer:\n%q", out)
	}
}

// TestPlainEmptyHeaderOmitsAoCBox asserts that Year==0 suppresses the AoC header box,
// while a populated header renders the box correctly.
func TestPlainEmptyHeaderOmitsAoCBox(t *testing.T) {
	t.Run("zero year omits box", func(t *testing.T) {
		var buf bytes.Buffer
		p := NewPlain(&buf, Header{Language: "Go"})
		_ = p.Close()
		out := buf.String()
		if strings.Contains(out, "ADVENT OF CODE 0") || strings.Contains(out, "Day 0") {
			t.Errorf("header with Year=0 should not render AoC box, got:\n%q", out)
		}
	})

	t.Run("populated header renders box", func(t *testing.T) {
		var buf bytes.Buffer
		p := NewPlain(&buf, Header{Year: 2015, Day: 4, Title: "The Ideal Stocking Stuffer", Language: "Go"})
		_ = p.Close()
		out := buf.String()
		if !strings.Contains(out, "The Ideal Stocking Stuffer") {
			t.Errorf("populated header should render AoC box, got:\n%q", out)
		}
	})
}

func TestPlainRendersHeaderAndPassLine(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{Year: 2015, Day: 4, Title: "The Ideal Stocking Stuffer", Language: "Go"})

	p.Handle(tasks.PlannedEvent("Test.1.1", tasks.Test, protocol.PartOne, 1, "")) // ignored
	p.Handle(tasks.StartedEvent("Test.1.1", tasks.Test, protocol.PartOne, 1, "")) // ignored
	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.1", Type: tasks.Test, Part: protocol.PartOne, SubPart: 1,
		Status: tasks.StatusPassed, Duration: 0.071143,
	}, ""))
	_ = p.Close()

	out := buf.String()
	if !strings.Contains(out, "The Ideal Stocking Stuffer") {
		t.Errorf("missing title in:\n%s", out)
	}
	if !strings.Contains(out, "Testing (Go)") {
		t.Errorf("missing section label in:\n%s", out)
	}
	if !strings.Contains(out, "PASS") {
		t.Errorf("missing PASS in:\n%s", out)
	}
}

func TestPlainPlannedAndStartedEmitNothing(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{Year: 2015, Day: 1, Title: "Not Quite Lisp", Language: "Go"})

	// header is emitted on construction, not on events; capture post-construction state
	headerLen := buf.Len()

	p.Handle(tasks.PlannedEvent("Test.1.1", tasks.Test, protocol.PartOne, 1, ""))
	afterPlanned := buf.Len()

	p.Handle(tasks.StartedEvent("Test.1.1", tasks.Test, protocol.PartOne, 1, ""))
	afterStarted := buf.Len()

	if afterPlanned != headerLen {
		t.Errorf("PlannedEvent wrote %d bytes (expected 0 new bytes after header)", afterPlanned-headerLen)
	}
	if afterStarted != headerLen {
		t.Errorf("StartedEvent wrote %d bytes (expected 0 new bytes after header)", afterStarted-headerLen)
	}
}

func TestPlainTimeoutLine(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{Year: 2020, Day: 6, Title: "Custom Customs", Language: "Go"})

	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.1", Type: tasks.Test, Part: protocol.PartOne, SubPart: 1,
		Status: tasks.StatusTimeout, Duration: 5.0,
	}, ""))
	_ = p.Close()

	out := buf.String()
	if !strings.Contains(out, "TIMEOUT") {
		t.Errorf("missing TIMEOUT in:\n%s", out)
	}
}

func TestPlainSolveSectionLabel(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{Year: 2021, Day: 1, Title: "Sonar Sweep", Language: "py"})

	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Solve.1", Type: tasks.Solve, Part: protocol.PartOne,
		Status: tasks.StatusPassed, Duration: 0.001,
	}, ""))
	_ = p.Close()

	out := buf.String()
	if !strings.Contains(out, "Solving (py)") {
		t.Errorf("missing Solving section label in:\n%s", out)
	}
}

func TestPlainFailLine(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{Year: 2022, Day: 3, Title: "Rucksack Reorganization", Language: "Go"})

	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.1", Type: tasks.Test, Part: protocol.PartOne, SubPart: 1,
		Status: tasks.StatusFailed, Output: "42", Expected: "99", Duration: 0.005,
	}, ""))
	_ = p.Close()

	out := buf.String()
	if !strings.Contains(out, "FAIL") {
		t.Errorf("missing FAIL in:\n%s", out)
	}
}

func TestPlainSectionLabelOnlyOncePerType(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{Year: 2023, Day: 1, Title: "Trebuchet?!", Language: "Go"})

	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.1", Type: tasks.Test, Part: protocol.PartOne, SubPart: 1,
		Status: tasks.StatusPassed, Duration: 0.001,
	}, ""))
	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.2", Type: tasks.Test, Part: protocol.PartOne, SubPart: 2,
		Status: tasks.StatusPassed, Duration: 0.002,
	}, ""))
	_ = p.Close()

	out := buf.String()
	count := strings.Count(out, "Testing (Go)")
	if count != 1 {
		t.Errorf("expected exactly 1 section label, got %d in:\n%s", count, out)
	}
}

func TestPlainMetaEventSetsHeaderAndLanguage(t *testing.T) {
	var buf bytes.Buffer
	// Constructed with no metadata (the CLI builds the renderer pre-Load).
	p := NewPlain(&buf, Header{})

	// Domain emits Meta first with the resolved identity + pretty runner name.
	p.Handle(tasks.MetaEvent(tasks.Meta{Year: 2015, Day: 11, Title: "Corporate Policy", Language: "Rust"}))
	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.0", Type: tasks.Test, Part: protocol.PartOne, SubPart: 0,
		Status: tasks.StatusPassed, Duration: 0.000190984,
	}, ""))
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "2015 Day 11: Corporate Policy") {
		t.Errorf("header missing/wrong:\n%s", out)
	}
	if !strings.Contains(out, "Testing (Rust)") {
		t.Errorf("section label should use pretty name 'Rust', not a key:\n%s", out)
	}
	if strings.Contains(out, "(rs)") {
		t.Errorf("section label leaked the lang key 'rs':\n%s", out)
	}
}

func TestPlainUniformDurationUnitAndAlignment(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{})
	p.Handle(tasks.MetaEvent(tasks.Meta{Year: 2015, Day: 11, Title: "X", Language: "Rust"}))
	// Mix of µs- and ms-scale durations; max is ~15ms so ALL should render in ms.
	durs := []float64{0.000190984, 0.011634847, 0.000185025, 0.015325656}
	for i, d := range durs {
		p.Handle(tasks.FinishedEvent(tasks.Result{
			ID: "Test.1." + string(rune('0'+i)), Type: tasks.Test,
			Part: protocol.PartOne, SubPart: i, Status: tasks.StatusPassed, Duration: d,
		}, ""))
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	out := buf.String()
	// Every duration uses ms (uniform magnitude); none uses µs.
	if strings.Contains(out, "µs") {
		t.Errorf("expected uniform ms unit, found µs:\n%s", out)
	}
	for _, want := range []string{"0.191ms", "11.635ms", "0.185ms", "15.326ms"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing uniform-unit duration %q in:\n%s", want, out)
		}
	}
}

func TestPlainDurationUnitIsPerSection(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{Year: 2015, Day: 6, Title: "Fire Hazard", Language: "Go"})
	// Testing: all sub-ms → section should render in µs or ms, NOT seconds.
	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.0", Type: tasks.Test, Part: protocol.PartOne, SubPart: 0,
		Status: tasks.StatusPassed, Duration: 0.000200,
	}, ""))
	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.2.0", Type: tasks.Test, Part: protocol.PartTwo, SubPart: 0,
		Status: tasks.StatusPassed, Duration: 0.000353,
	}, ""))
	// Solving: multi-second → section renders in seconds.
	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Solve.1", Type: tasks.Solve, Part: protocol.PartOne, SubPart: 0,
		Status: tasks.StatusUnverified, Output: "569999", Duration: 8.652,
	}, ""))
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	out := buf.String()
	// Fast tests must NOT collapse to 0.000s — they keep a sub-second unit.
	if strings.Contains(out, "0.000s") {
		t.Errorf("fast test crushed to 0.000s; section unit should be sub-second:\n%s", out)
	}
	// The slow solve uses seconds, and its answer is shown.
	if !strings.Contains(out, "8.652s") {
		t.Errorf("solve section should render seconds:\n%s", out)
	}
	if !strings.Contains(out, "569999") {
		t.Errorf("NEW answer must be shown:\n%s", out)
	}
}

// Plain collapses benchmark iterations into one settled line per (runner, Part)
// — the same "not every iteration its own line" intent as the live bars, minus
// animation (ADR-0011).
func TestPlainBenchmarkAggregatesPerRunnerPart(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{Year: 2015, Day: 1, Title: "Day 1"})

	const iters = 3

	for i := range iters {
		p.Handle(tasks.FinishedEvent(tasks.Result{
			ID: tasks.MakeTaskID(tasks.Benchmark, protocol.PartOne, i), Type: tasks.Benchmark,
			Part: protocol.PartOne, SubPart: i, Status: tasks.StatusPassed, Duration: 0.5,
		}, "Go"))
	}

	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	out := buf.String()
	// One aggregated line: runner name, count, and summed duration.
	if !strings.Contains(out, "Go") {
		t.Errorf("benchmark line should name the runner:\n%s", out)
	}
	if !strings.Contains(out, "3") {
		t.Errorf("benchmark line should show the iteration count:\n%s", out)
	}
	if !strings.Contains(out, "1.500s") {
		t.Errorf("benchmark line should show summed duration 1.500s (3 × 0.5):\n%s", out)
	}
	// Must NOT print one line per iteration: no per-iteration task labels.
	if strings.Contains(out, "1.0:") || strings.Contains(out, "1.1:") || strings.Contains(out, "1.2:") {
		t.Errorf("benchmark iterations must not render as per-iteration lines:\n%s", out)
	}
}

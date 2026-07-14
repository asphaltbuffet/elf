package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	assert.NotContains(t, out, "\x1b[", "output contains ANSI escape sequences on non-TTY writer")
}

// TestPlainEmptyHeaderOmitsAoCBox asserts that Year==0 suppresses the AoC header box,
// while a populated header renders the box correctly.
func TestPlainEmptyHeaderOmitsAoCBox(t *testing.T) {
	t.Run("zero year omits box", func(t *testing.T) {
		var buf bytes.Buffer
		p := NewPlain(&buf, Header{Language: "Go"})
		_ = p.Close()
		out := buf.String()
		assert.NotContains(t, out, "ADVENT OF CODE 0", "header with Year=0 should not render AoC box")
		assert.NotContains(t, out, "Day 0", "header with Year=0 should not render AoC box")
	})

	t.Run("populated header renders box", func(t *testing.T) {
		var buf bytes.Buffer
		p := NewPlain(&buf, Header{Year: 2015, Day: 4, Title: "The Ideal Stocking Stuffer", Language: "Go"})
		_ = p.Close()
		out := buf.String()
		assert.Contains(t, out, "The Ideal Stocking Stuffer", "populated header should render AoC box")
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
	assert.Contains(t, out, "The Ideal Stocking Stuffer", "missing title")
	assert.Contains(t, out, "Testing (Go)", "missing section label")
	assert.Contains(t, out, "PASS", "missing PASS")
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

	assert.Equal(t, headerLen, afterPlanned, "PlannedEvent wrote unexpected bytes after header")
	assert.Equal(t, headerLen, afterStarted, "StartedEvent wrote unexpected bytes after header")
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
	assert.Contains(t, out, "TIMEOUT", "missing TIMEOUT")
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
	assert.Contains(t, out, "Solving (py)", "missing Solving section label")
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
	assert.Contains(t, out, "FAIL", "missing FAIL")
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
	assert.Equal(t, 1, strings.Count(out, "Testing (Go)"), "expected exactly 1 section label")
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
	require.NoError(t, p.Close())

	out := buf.String()
	assert.Contains(t, out, "2015 Day 11: Corporate Policy", "header missing/wrong")
	assert.Contains(t, out, "Testing (Rust)", "section label should use pretty name 'Rust', not a key")
	assert.NotContains(t, out, "(rs)", "section label leaked the lang key 'rs'")
}

func TestPlainEulerHeaderUsesNumber(t *testing.T) {
	var buf bytes.Buffer
	p := NewPlain(&buf, Header{})

	// A Project Euler problem carries a bare Number (no Year/Day); the header
	// should read "#21: Amicable Numbers".
	p.Handle(tasks.MetaEvent(tasks.Meta{Number: 21, Title: "Amicable Numbers", Language: "Go"}))
	p.Handle(tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.0", Type: tasks.Test, Part: protocol.PartOne, SubPart: 0,
		Status: tasks.StatusPassed, Duration: 0.000190984,
	}, ""))
	require.NoError(t, p.Close())

	out := buf.String()
	assert.Contains(t, out, "#21: Amicable Numbers", "Euler header should render '#N: Title'")
	assert.NotContains(t, out, "Day", "Euler header must not use the AoC 'Day' format")
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
	require.NoError(t, p.Close())

	out := buf.String()
	assert.NotContains(t, out, "µs", "expected uniform ms unit, found µs")
	for _, want := range []string{"0.191ms", "11.635ms", "0.185ms", "15.326ms"} {
		assert.Contains(t, out, want, "missing uniform-unit duration")
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
	require.NoError(t, p.Close())

	out := buf.String()
	assert.NotContains(t, out, "0.000s", "fast test crushed to 0.000s; section unit should be sub-second")
	assert.Contains(t, out, "8.652s", "solve section should render seconds")
	assert.Contains(t, out, "569999", "NEW answer must be shown")
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

	require.NoError(t, p.Close())

	out := buf.String()
	assert.Contains(t, out, "Go", "benchmark line should name the runner")
	assert.Contains(t, out, "3", "benchmark line should show the iteration count")
	assert.Contains(t, out, "1.500s", "benchmark line should show summed duration 1.500s (3 × 0.5)")
	assert.NotContains(t, out, "1.0:", "benchmark iterations must not render as per-iteration lines")
	assert.NotContains(t, out, "1.1:", "benchmark iterations must not render as per-iteration lines")
	assert.NotContains(t, out, "1.2:", "benchmark iterations must not render as per-iteration lines")
}

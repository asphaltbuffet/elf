package render

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func TestLiveShowsNotStartedThenSettles(t *testing.T) {
	m := NewLive(Header{Year: 2015, Day: 4, Title: "Day 4", Language: "Go"})

	// Planned -> row appears as not-started.
	mi, _ := m.Update(eventMsg{tasks.PlannedEvent("Test.1.1", tasks.Test, protocol.PartOne, 1, "")})
	m = mi.(*Live)

	require.Contains(t, m.View().Content, "not started", "expected '<not started>' row")

	// Finished -> row settles to PASS.
	mi, _ = m.Update(eventMsg{tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.1", Type: tasks.Test, Part: protocol.PartOne, SubPart: 1,
		Status: tasks.StatusPassed, Duration: 0.071,
	}, "")})
	m = mi.(*Live)

	require.Contains(t, m.View().Content, "PASS", "expected PASS after Finished")
}

func TestFormatElapsed(t *testing.T) {
	assert.Equal(t, "23.1s", formatElapsed(23100*time.Millisecond))
	assert.Equal(t, "1m05s", formatElapsed(65*time.Second))
}

func TestLiveMetaEventSetsHeaderAndLanguage(t *testing.T) {
	m := NewLive(Header{})
	mi, _ := m.Update(eventMsg{tasks.MetaEvent(tasks.Meta{
		Year: 2015, Day: 11, Title: "Corporate Policy", Language: "Rust",
	})})
	m = mi.(*Live)
	mi, _ = m.Update(eventMsg{tasks.PlannedEvent("Test.1.0", tasks.Test, protocol.PartOne, 0, "")})
	m = mi.(*Live)

	out := m.View().Content
	assert.Contains(t, out, "2015 Day 11: Corporate Policy", "live header missing/wrong")
	assert.Contains(t, out, "Testing (Rust)", "live section label should use pretty name 'Rust'")
}

// Benchmark iterations collapse into one progress bar per (runner, Part): the
// bar advances as each iteration finishes rather than each iteration being its
// own row (ADR-0011).
func TestLiveBenchmarkShowsOneBarPerRunnerPart(t *testing.T) {
	m := NewLive(Header{Year: 2015, Day: 1, Title: "Day 1"})

	const iters = 3

	// Plan three PartOne iterations for the "Go" runner.
	for i := range iters {
		mi, _ := m.Update(eventMsg{tasks.PlannedEvent(
			tasks.MakeTaskID(tasks.Benchmark, protocol.PartOne, i),
			tasks.Benchmark, protocol.PartOne, i, "Go",
		)})
		m = mi.(*Live)
	}

	// Finish two of the three iterations.
	for i := range 2 {
		mi, _ := m.Update(eventMsg{tasks.FinishedEvent(tasks.Result{
			ID: tasks.MakeTaskID(tasks.Benchmark, protocol.PartOne, i), Type: tasks.Benchmark,
			Part: protocol.PartOne, SubPart: i, Status: tasks.StatusPassed, Duration: 0.01,
		}, "Go")})
		m = mi.(*Live)
	}

	out := m.View().Content

	assert.Contains(t, out, "Go", "bar should be labelled with the runner name")
	assert.Contains(t, out, "2/3", "bar should show 2 of 3 iterations complete")
	// Three iterations must NOT produce three separate task rows.
	assert.NotContains(t, out, "1.0:", "benchmark iterations must not render as per-iteration rows")
	assert.NotContains(t, out, "1.1:", "benchmark iterations must not render as per-iteration rows")
	assert.NotContains(t, out, "1.2:", "benchmark iterations must not render as per-iteration rows")
}

func TestLiveBenchmarkSettledBarShowsSumDuration(t *testing.T) {
	m := NewLive(Header{Year: 2015, Day: 1, Title: "Day 1"})

	const iters = 2

	for i := range iters {
		mi, _ := m.Update(eventMsg{tasks.PlannedEvent(
			tasks.MakeTaskID(tasks.Benchmark, protocol.PartOne, i),
			tasks.Benchmark, protocol.PartOne, i, "Go",
		)})
		m = mi.(*Live)
	}

	for i := range iters {
		mi, _ := m.Update(eventMsg{tasks.FinishedEvent(tasks.Result{
			ID: tasks.MakeTaskID(tasks.Benchmark, protocol.PartOne, i), Type: tasks.Benchmark,
			Part: protocol.PartOne, SubPart: i, Status: tasks.StatusPassed, Duration: 0.5,
		}, "Go")})
		m = mi.(*Live)
	}

	out := m.View().Content
	assert.Contains(t, out, "2/2", "completed bar should show 2/2")
	// Sum of the two 0.5s samples is 1.000s.
	assert.Contains(t, out, "1.000", "settled bar should show summed duration 1.000s")
}

// The benchmark bar uses the bubbles/v2 progress component, which renders a
// colored gradient fill — so a partially-filled bar carries ANSI color escapes
// within it (the hand-rolled glyph bar did not).
func TestLiveBenchmarkBarUsesGradientFill(t *testing.T) {
	m := NewLive(Header{Year: 2015, Day: 1, Title: "Day 1"})

	const iters = 10
	for i := range iters {
		mi, _ := m.Update(eventMsg{tasks.PlannedEvent(
			tasks.MakeTaskID(tasks.Benchmark, protocol.PartOne, i),
			tasks.Benchmark, protocol.PartOne, i, "Go",
		)})
		m = mi.(*Live)
	}
	for i := range 5 {
		mi, _ := m.Update(eventMsg{tasks.FinishedEvent(tasks.Result{
			ID: tasks.MakeTaskID(tasks.Benchmark, protocol.PartOne, i), Type: tasks.Benchmark,
			Part: protocol.PartOne, SubPart: i, Status: tasks.StatusPassed, Duration: 0.01,
		}, "Go")})
		m = mi.(*Live)
	}

	out := m.View().Content
	require.Contains(t, out, "5/10", "bar should still show the 5/10 count")

	// Isolate the bar region: everything on the bar's line before the "5/10"
	// count. The bubbles gradient puts ANSI color escapes there; the old
	// hand-rolled glyph bar did not.
	var barLine string
	for line := range strings.SplitSeq(out, "\n") {
		if strings.Contains(line, "5/10") {
			barLine = line

			break
		}
	}

	prefix, _, _ := strings.Cut(barLine, "5/10")
	assert.Contains(t, prefix, "\x1b[",
		"progress bar fill should be colored (ANSI escapes before the count)")
	// Use solid full blocks, not half-blocks: the half-block sub-cell trick
	// renders as distracting vertical stripes in many terminal/font combos.
	assert.Contains(t, prefix, "█", "bar fill should use the solid full block")
	assert.NotContains(t, prefix, "▌", "bar must not use half-blocks (they stripe)")
}

func TestLiveDoneRowShowsNewAnswer(t *testing.T) {
	m := NewLive(Header{})
	mi, _ := m.Update(eventMsg{tasks.PlannedEvent("Solve.1", tasks.Solve, protocol.PartOne, 0, "")})
	m = mi.(*Live)
	mi, _ = m.Update(eventMsg{tasks.FinishedEvent(tasks.Result{
		ID: "Solve.1", Type: tasks.Solve, Part: protocol.PartOne, SubPart: 0,
		Status: tasks.StatusUnverified, Output: "1321131112", Duration: 3.382,
	}, "")})
	m = mi.(*Live)

	out := m.View().Content
	assert.Contains(t, out, "NEW", "expected NEW status")
	assert.Contains(t, out, "1321131112", "expected the NEW answer to be shown so the user can submit it")
}

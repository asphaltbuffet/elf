package render

import (
	"strings"
	"testing"
	"time"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func TestLiveShowsNotStartedThenSettles(t *testing.T) {
	m := NewLive(Header{Year: 2015, Day: 4, Title: "Day 4", Language: "Go"})

	// Planned -> row appears as not-started.
	mi, _ := m.Update(eventMsg{tasks.PlannedEvent("Test.1.1", tasks.Test, protocol.PartOne, 1)})
	m = mi.(*Live)

	if !strings.Contains(m.View().Content, "not started") {
		t.Fatalf("expected '<not started>' row, got:\n%s", m.View().Content)
	}

	// Finished -> row settles to PASS.
	mi, _ = m.Update(eventMsg{tasks.FinishedEvent(tasks.Result{
		ID: "Test.1.1", Type: tasks.Test, Part: protocol.PartOne, SubPart: 1,
		Status: tasks.StatusPassed, Duration: 0.071,
	})})
	m = mi.(*Live)

	if !strings.Contains(m.View().Content, "PASS") {
		t.Fatalf("expected PASS after Finished, got:\n%s", m.View().Content)
	}
}

func TestFormatElapsed(t *testing.T) {
	if got := formatElapsed(23100 * time.Millisecond); got != "23.1s" {
		t.Errorf("got %q want 23.1s", got)
	}

	if got := formatElapsed(65 * time.Second); got != "1m05s" {
		t.Errorf("got %q want 1m05s", got)
	}
}

func TestLiveMetaEventSetsHeaderAndLanguage(t *testing.T) {
	m := NewLive(Header{})
	mi, _ := m.Update(eventMsg{tasks.MetaEvent(tasks.Meta{
		Year: 2015, Day: 11, Title: "Corporate Policy", Language: "Rust",
	})})
	m = mi.(*Live)
	mi, _ = m.Update(eventMsg{tasks.PlannedEvent("Test.1.0", tasks.Test, protocol.PartOne, 0)})
	m = mi.(*Live)

	out := m.View().Content
	if !strings.Contains(out, "2015 Day 11: Corporate Policy") {
		t.Errorf("live header missing/wrong:\n%s", out)
	}
	if !strings.Contains(out, "Testing (Rust)") {
		t.Errorf("live section label should use pretty name 'Rust':\n%s", out)
	}
}

func TestLiveDoneRowShowsNewAnswer(t *testing.T) {
	m := NewLive(Header{})
	mi, _ := m.Update(eventMsg{tasks.PlannedEvent("Solve.1", tasks.Solve, protocol.PartOne, 0)})
	m = mi.(*Live)
	mi, _ = m.Update(eventMsg{tasks.FinishedEvent(tasks.Result{
		ID: "Solve.1", Type: tasks.Solve, Part: protocol.PartOne, SubPart: 0,
		Status: tasks.StatusUnverified, Output: "1321131112", Duration: 3.382,
	})})
	m = mi.(*Live)

	out := m.View().Content
	if !strings.Contains(out, "NEW") {
		t.Errorf("expected NEW status:\n%s", out)
	}
	if !strings.Contains(out, "1321131112") {
		t.Errorf("expected the NEW answer to be shown so the user can submit it:\n%s", out)
	}
}

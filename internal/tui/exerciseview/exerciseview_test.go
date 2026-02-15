package exerciseview

import (
	"testing"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/asphaltbuffet/elf/internal/tui/components"
	"github.com/asphaltbuffet/elf/internal/tui/discover"
	"github.com/asphaltbuffet/elf/internal/tui/nav"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func keyMsg(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func specialKeyMsg(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

// testModel returns a minimal exercise view model suitable for testing
// Update behavior without triggering the actual exercise runner.
func testModel() Model {
	return Model{
		info:     discover.ExerciseInfo{Year: 2015, Day: 1, Title: "Test Exercise"},
		action:   "solve",
		lang:     "go",
		results:  components.NewResultList(80, 20),
		progress: components.StartedProgress("Solving Test Exercise..."),
		resultCh: make(chan tasks.Result, resultBufSize),
		running:  true,
	}
}

func TestResultMsgAddsResult(t *testing.T) {
	m := testModel()

	result := tasks.Result{
		Type:   tasks.Solve,
		Status: tasks.StatusPassed,
		Output: "42",
	}

	updated, cmd := m.Update(resultMsg{result: result})
	model := updated.(Model)

	if model.results.Count() != 1 {
		t.Errorf("expected 1 result, got %d", model.results.Count())
	}

	// resultMsg should return a waitForResult command
	if cmd == nil {
		t.Error("expected non-nil cmd after resultMsg")
	}
}

func TestMultipleResultMsgs(t *testing.T) {
	m := testModel()

	r1 := tasks.Result{Status: tasks.StatusPassed, Output: "42"}
	r2 := tasks.Result{Status: tasks.StatusPassed, Output: "100"}

	updated, _ := m.Update(resultMsg{result: r1})
	model := updated.(Model)
	updated, _ = model.Update(resultMsg{result: r2})
	model = updated.(Model)

	if model.results.Count() != 2 {
		t.Errorf("expected 2 results, got %d", model.results.Count())
	}
}

func TestDoneMsgStopsSpinner(t *testing.T) {
	m := testModel()

	if !m.running {
		t.Fatal("model should start running")
	}

	updated, cmd := m.Update(doneMsg{})
	model := updated.(Model)

	if model.running {
		t.Error("expected running to be false after doneMsg")
	}

	if !model.done {
		t.Error("expected done to be true after doneMsg")
	}

	if cmd != nil {
		t.Errorf("expected nil cmd after doneMsg, got non-nil")
	}
}

func TestDoneMsgWithError(t *testing.T) {
	m := testModel()

	updated, _ := m.Update(doneMsg{err: errTest})
	model := updated.(Model)

	if model.running {
		t.Error("expected running to be false")
	}

	if !model.done {
		t.Error("expected done to be true")
	}

	if model.err == nil {
		t.Error("expected error to be set")
	}
}

var errTest = &testError{}

type testError struct{}

func (e *testError) Error() string { return "test error" }

func TestEscGoesBack(t *testing.T) {
	m := testModel()

	_, cmd := m.Update(specialKeyMsg(tea.KeyEsc))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for esc")
	}

	msg := cmd()
	if _, ok := msg.(nav.PopScreenMsg); !ok {
		t.Fatalf("expected PopScreenMsg, got %T", msg)
	}
}

func TestHGoesBack(t *testing.T) {
	m := testModel()

	_, cmd := m.Update(keyMsg("h"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for h key")
	}

	msg := cmd()
	if _, ok := msg.(nav.PopScreenMsg); !ok {
		t.Fatalf("expected PopScreenMsg, got %T", msg)
	}
}

func TestQQuits(t *testing.T) {
	m := testModel()

	_, cmd := m.Update(keyMsg("q"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for quit")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := testModel()

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(Model)

	if model.width != 120 {
		t.Errorf("expected width 120, got %d", model.width)
	}

	if model.height != 40 {
		t.Errorf("expected height 40, got %d", model.height)
	}

	// WindowSizeMsg should return nil cmd
	if cmd != nil {
		t.Errorf("expected nil cmd for WindowSizeMsg")
	}
}

func helpModel() help.Model {
	return help.New()
}

func windowSizeMsg(w, h int) tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: w, Height: h}
}

func TestProgressStopsAfterDone(t *testing.T) {
	m := testModel()

	if !m.progress.Active() {
		t.Fatal("progress should be active initially")
	}

	updated, _ := m.Update(doneMsg{})
	model := updated.(Model)

	if model.progress.Active() {
		t.Error("progress should be inactive after doneMsg")
	}
}

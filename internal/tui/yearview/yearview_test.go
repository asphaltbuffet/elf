package yearview

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/asphaltbuffet/elf/internal/tui/discover"
	"github.com/asphaltbuffet/elf/internal/tui/nav"
)

func testExercises() []discover.ExerciseInfo {
	return []discover.ExerciseInfo{
		{Year: 2015, Day: 1, Title: "Not Called It", Path: "exercises/2015/01-notCalledIt", Langs: []string{"go"}},
		{Year: 2015, Day: 2, Title: "Inverse", Path: "exercises/2015/02-inverse", Langs: []string{"go", "py"}},
		{Year: 2015, Day: 3, Title: "Perfectly Spherical", Path: "exercises/2015/03-perfectly", Langs: []string{"go"}},
	}
}

func keyMsg(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func specialKeyMsg(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

func TestNew(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	if m.year != 2015 {
		t.Errorf("expected year 2015, got %d", m.year)
	}

	if len(m.exercises) != 3 {
		t.Errorf("expected 3 exercises, got %d", len(m.exercises))
	}

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}
}

func TestCursorDown(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	// Move down once
	updated, _ := m.Update(keyMsg("j"))
	model := updated.(Model)

	if model.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", model.cursor)
	}

	// Move down again
	updated, _ = model.Update(keyMsg("j"))
	model = updated.(Model)

	if model.cursor != 2 {
		t.Errorf("expected cursor at 2, got %d", model.cursor)
	}
}

func TestCursorUp(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)
	m.cursor = 2

	updated, _ := m.Update(keyMsg("k"))
	model := updated.(Model)

	if model.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", model.cursor)
	}
}

func TestCursorDownAtEnd(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)
	m.cursor = 2 // last exercise

	updated, _ := m.Update(keyMsg("j"))
	model := updated.(Model)

	if model.cursor != 2 {
		t.Errorf("expected cursor to stay at 2, got %d", model.cursor)
	}
}

func TestCursorUpAtZero(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	updated, _ := m.Update(keyMsg("k"))
	model := updated.(Model)

	if model.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", model.cursor)
	}
}

func TestCursorDownArrowKey(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	updated, _ := m.Update(specialKeyMsg(tea.KeyDown))
	model := updated.(Model)

	if model.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", model.cursor)
	}
}

func TestCursorUpArrowKey(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)
	m.cursor = 1

	updated, _ := m.Update(specialKeyMsg(tea.KeyUp))
	model := updated.(Model)

	if model.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", model.cursor)
	}
}

func TestSolveAction(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	_, cmd := m.Update(keyMsg("s"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for solve action")
	}

	msg := cmd()
	action, ok := msg.(ActionMsg)

	if !ok {
		t.Fatalf("expected ActionMsg, got %T", msg)
	}

	if action.Action != "solve" {
		t.Errorf("expected action 'solve', got %q", action.Action)
	}

	if action.Exercise.Day != 1 {
		t.Errorf("expected exercise day 1, got %d", action.Exercise.Day)
	}
}

func TestTestAction(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)
	m.cursor = 1

	_, cmd := m.Update(keyMsg("t"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for test action")
	}

	msg := cmd()
	action, ok := msg.(ActionMsg)

	if !ok {
		t.Fatalf("expected ActionMsg, got %T", msg)
	}

	if action.Action != "test" {
		t.Errorf("expected action 'test', got %q", action.Action)
	}

	if action.Exercise.Day != 2 {
		t.Errorf("expected exercise day 2, got %d", action.Exercise.Day)
	}
}

func TestBenchmarkAction(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	_, cmd := m.Update(keyMsg("b"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for benchmark action")
	}

	msg := cmd()
	action, ok := msg.(ActionMsg)

	if !ok {
		t.Fatalf("expected ActionMsg, got %T", msg)
	}

	if action.Action != "benchmark" {
		t.Errorf("expected action 'benchmark', got %q", action.Action)
	}
}

func TestAnalyzeAction(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	_, cmd := m.Update(keyMsg("a"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for analyze action")
	}

	msg := cmd()
	analyze, ok := msg.(AnalyzeMsg)

	if !ok {
		t.Fatalf("expected AnalyzeMsg, got %T", msg)
	}

	if analyze.YearDir != "exercises/2015" {
		t.Errorf("expected year dir 'exercises/2015', got %q", analyze.YearDir)
	}
}

func TestBackEsc(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	_, cmd := m.Update(specialKeyMsg(tea.KeyEsc))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for esc")
	}

	msg := cmd()
	if _, ok := msg.(nav.PopScreenMsg); !ok {
		t.Fatalf("expected PopScreenMsg, got %T", msg)
	}
}

func TestBackH(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	_, cmd := m.Update(keyMsg("h"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for h key")
	}

	msg := cmd()
	if _, ok := msg.(nav.PopScreenMsg); !ok {
		t.Fatalf("expected PopScreenMsg, got %T", msg)
	}
}

func TestQuit(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	_, cmd := m.Update(keyMsg("q"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for quit")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestActionsWithEmptyExercises(t *testing.T) {
	m := New(nil, "", nil, 2015, nil)

	actions := []string{"s", "t", "b", "a"}

	for _, a := range actions {
		_, cmd := m.Update(keyMsg(a))

		if cmd != nil {
			t.Errorf("action %q on empty exercises should return nil cmd", a)
		}
	}
}

func TestEnterAction(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	_, cmd := m.Update(specialKeyMsg(tea.KeyEnter))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for enter")
	}

	msg := cmd()
	action, ok := msg.(ActionMsg)

	if !ok {
		t.Fatalf("expected ActionMsg, got %T", msg)
	}

	if action.Action != "solve" {
		t.Errorf("expected enter to trigger 'solve', got %q", action.Action)
	}
}

func TestRightAction(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	_, cmd := m.Update(keyMsg("l"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for l key")
	}

	msg := cmd()
	action, ok := msg.(ActionMsg)

	if !ok {
		t.Fatalf("expected ActionMsg, got %T", msg)
	}

	if action.Action != "solve" {
		t.Errorf("expected l to trigger 'solve', got %q", action.Action)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(Model)

	if cmd != nil {
		t.Errorf("expected nil cmd for WindowSizeMsg, got %v", cmd)
	}

	if model.width != 120 {
		t.Errorf("expected width 120, got %d", model.width)
	}

	if model.height != 40 {
		t.Errorf("expected height 40, got %d", model.height)
	}
}

func TestSolveActionCursorPosition(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)
	m.cursor = 2

	_, cmd := m.Update(keyMsg("s"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	msg := cmd()
	action := msg.(ActionMsg)

	if action.Exercise.Day != 3 {
		t.Errorf("expected exercise day 3 at cursor 2, got %d", action.Exercise.Day)
	}

	if action.Exercise.Title != "Perfectly Spherical" {
		t.Errorf("expected title 'Perfectly Spherical', got %q", action.Exercise.Title)
	}
}

func TestInit(t *testing.T) {
	exercises := testExercises()
	m := New(nil, "", nil, 2015, exercises)

	cmd := m.Init()
	if cmd != nil {
		t.Errorf("expected nil from Init(), got non-nil")
	}
}

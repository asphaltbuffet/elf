package dashboard

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/internal/tui/discover"
	"github.com/asphaltbuffet/elf/internal/tui/nav"
)

func keyMsg(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func specialKeyMsg(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

// populatedModel returns a dashboard model with years and exercises already loaded,
// bypassing the async scan.
func populatedModel() Model {
	exercises := map[int][]discover.ExerciseInfo{
		2023: {
			{Year: 2023, Day: 1, Title: "Trebuchet", Path: "exercises/2023/01-trebuchet"},
			{Year: 2023, Day: 2, Title: "Cube Conundrum", Path: "exercises/2023/02-cube"},
		},
		2015: {
			{Year: 2015, Day: 1, Title: "Not Called It", Path: "exercises/2015/01-notCalledIt"},
		},
	}

	return Model{
		app:       &fakeDashboardInfo{fs: afero.NewMemMapFs(), lang: "go", baseDir: "exercises"},
		years:     []int{2023, 2015}, // descending order
		exercises: exercises,
		loading:   false,
	}
}

func TestCursorDown(t *testing.T) {
	m := populatedModel()

	updated, _ := m.Update(keyMsg("j"))
	model := updated.(Model)

	if model.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", model.cursor)
	}
}

func TestCursorUp(t *testing.T) {
	m := populatedModel()
	m.cursor = 1

	updated, _ := m.Update(keyMsg("k"))
	model := updated.(Model)

	if model.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", model.cursor)
	}
}

func TestCursorDownAtEnd(t *testing.T) {
	m := populatedModel()
	m.cursor = 1 // last year

	updated, _ := m.Update(keyMsg("j"))
	model := updated.(Model)

	if model.cursor != 1 {
		t.Errorf("expected cursor to stay at 1, got %d", model.cursor)
	}
}

func TestCursorUpAtZero(t *testing.T) {
	m := populatedModel()

	updated, _ := m.Update(keyMsg("k"))
	model := updated.(Model)

	if model.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", model.cursor)
	}
}

func TestCursorDownArrow(t *testing.T) {
	m := populatedModel()

	updated, _ := m.Update(specialKeyMsg(tea.KeyDown))
	model := updated.(Model)

	if model.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", model.cursor)
	}
}

func TestCursorUpArrow(t *testing.T) {
	m := populatedModel()
	m.cursor = 1

	updated, _ := m.Update(specialKeyMsg(tea.KeyUp))
	model := updated.(Model)

	if model.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", model.cursor)
	}
}

func TestEnterPushesYearView(t *testing.T) {
	m := populatedModel()

	_, cmd := m.Update(specialKeyMsg(tea.KeyEnter))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for enter")
	}

	msg := cmd()
	push, ok := msg.(nav.PushScreenMsg)

	if !ok {
		t.Fatalf("expected PushScreenMsg, got %T", msg)
	}

	if push.Screen == nil {
		t.Fatal("expected non-nil Screen in PushScreenMsg")
	}
}

func TestLPushesYearView(t *testing.T) {
	m := populatedModel()

	_, cmd := m.Update(keyMsg("l"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for l key")
	}

	msg := cmd()
	if _, ok := msg.(nav.PushScreenMsg); !ok {
		t.Fatalf("expected PushScreenMsg, got %T", msg)
	}
}

func TestRightArrowPushesYearView(t *testing.T) {
	m := populatedModel()

	_, cmd := m.Update(specialKeyMsg(tea.KeyRight))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for right arrow")
	}

	msg := cmd()
	if _, ok := msg.(nav.PushScreenMsg); !ok {
		t.Fatalf("expected PushScreenMsg, got %T", msg)
	}
}

func TestEnterWithNoYears(t *testing.T) {
	m := Model{
		years:   nil,
		loading: false,
	}

	_, cmd := m.Update(specialKeyMsg(tea.KeyEnter))

	if cmd != nil {
		t.Errorf("expected nil cmd when no years present")
	}
}

func TestQuit(t *testing.T) {
	m := populatedModel()

	_, cmd := m.Update(keyMsg("q"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for quit")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestEscQuit(t *testing.T) {
	m := populatedModel()

	_, cmd := m.Update(specialKeyMsg(tea.KeyEsc))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for esc")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := populatedModel()

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model := updated.(Model)

	if cmd != nil {
		t.Errorf("expected nil cmd for WindowSizeMsg")
	}

	if model.width != 100 {
		t.Errorf("expected width 100, got %d", model.width)
	}

	if model.height != 50 {
		t.Errorf("expected height 50, got %d", model.height)
	}
}

func TestScanCompleteMsg(t *testing.T) {
	m := Model{loading: true}

	exercises := map[int][]discover.ExerciseInfo{
		2015: {{Year: 2015, Day: 1, Title: "Test"}},
	}

	updated, _ := m.Update(scanCompleteMsg{years: exercises})
	model := updated.(Model)

	if model.loading {
		t.Error("expected loading to be false after scanCompleteMsg")
	}

	if len(model.years) != 1 {
		t.Errorf("expected 1 year, got %d", len(model.years))
	}

	if model.years[0] != 2015 {
		t.Errorf("expected year 2015, got %d", model.years[0])
	}
}

func TestScanCompleteMsgWithError(t *testing.T) {
	m := Model{loading: true}

	updated, _ := m.Update(scanCompleteMsg{err: errTest})
	model := updated.(Model)

	if model.loading {
		t.Error("expected loading to be false")
	}

	if model.err == nil {
		t.Error("expected error to be set")
	}
}

var errTest = &testError{}

type testError struct{}

func (e *testError) Error() string { return "test error" }

func TestSortedYears(t *testing.T) {
	m := map[int][]discover.ExerciseInfo{
		2015: nil,
		2023: nil,
		2020: nil,
	}

	years := sortedYears(m)

	if len(years) != 3 {
		t.Fatalf("expected 3 years, got %d", len(years))
	}

	if years[0] != 2023 || years[1] != 2020 || years[2] != 2015 {
		t.Errorf("expected descending order [2023 2020 2015], got %v", years)
	}
}

func TestCountCompleted(t *testing.T) {
	exercises := []discover.ExerciseInfo{
		{HasP1: true, HasP2: true},
		{HasP1: true, HasP2: false},
		{HasP1: false, HasP2: false},
	}

	count := countCompleted(exercises)

	if count != 1 {
		t.Errorf("expected 1 completed, got %d", count)
	}
}

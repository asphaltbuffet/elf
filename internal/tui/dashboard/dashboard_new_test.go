package dashboard

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/internal/tui/discover"
	"github.com/asphaltbuffet/elf/pkg/config"
)

func testConfig(t *testing.T) config.Config {
	t.Helper()

	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	return cfg
}

func populatedModelWithConfig(t *testing.T) Model {
	t.Helper()

	cfg := testConfig(t)

	exercises := map[int][]discover.ExerciseInfo{
		2023: {
			{Year: 2023, Day: 1, Title: "Trebuchet", Path: "exercises/2023/01-trebuchet", HasP1: true, HasP2: true},
			{Year: 2023, Day: 2, Title: "Cube Conundrum", Path: "exercises/2023/02-cube", HasP1: true, HasP2: false},
		},
		2015: {
			{
				Year:  2015,
				Day:   1,
				Title: "Not Called It",
				Path:  "exercises/2015/01-notCalledIt",
				HasP1: false,
				HasP2: false,
			},
		},
	}

	return Model{
		cfg:       cfg,
		years:     []int{2023, 2015},
		exercises: exercises,
		loading:   false,
		help:      help.New(),
	}
}

func TestNew(t *testing.T) {
	cfg := testConfig(t)
	m := New(cfg)

	if !m.loading {
		t.Error("expected loading to be true on new model")
	}

	if m.years != nil {
		t.Errorf("expected nil years, got %v", m.years)
	}

	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}

	_ = m.help.View(keys)
}

func TestInit(t *testing.T) {
	cfg := testConfig(t)
	m := New(cfg)

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected non-nil cmd from Init")
	}

	msg := cmd()
	if _, ok := msg.(scanCompleteMsg); !ok {
		t.Fatalf("expected scanCompleteMsg, got %T", msg)
	}
}

func TestViewLoading(t *testing.T) {
	m := Model{loading: true}
	view := m.View()

	if !strings.Contains(view, "Scanning") {
		t.Errorf("loading view should contain Scanning, got: %s", view)
	}
}

func TestViewError(t *testing.T) {
	m := Model{
		loading: false,
		err:     errors.New("something went wrong"),
	}

	view := m.View()

	if !strings.Contains(view, "Error:") {
		t.Errorf("error view should contain Error:, got: %s", view)
	}

	if !strings.Contains(view, "something went wrong") {
		t.Errorf("error view should contain error message, got: %s", view)
	}
}

func TestViewEmptyYears(t *testing.T) {
	cfg := testConfig(t)
	m := Model{
		cfg:       cfg,
		loading:   false,
		years:     []int{},
		exercises: map[int][]discover.ExerciseInfo{},
		help:      help.New(),
	}

	view := m.View()

	if !strings.Contains(view, "No exercises") {
		t.Errorf("empty years view should contain No exercises, got: %s", view)
	}
}

func TestViewPopulated(t *testing.T) {
	m := populatedModelWithConfig(t)
	view := m.View()

	if !strings.Contains(view, "2023") {
		t.Errorf("view should contain 2023, got: %s", view)
	}

	if !strings.Contains(view, "2015") {
		t.Errorf("view should contain 2015, got: %s", view)
	}
}

func TestViewPopulatedCursorSecond(t *testing.T) {
	m := populatedModelWithConfig(t)
	m.cursor = 1
	view := m.View()

	if !strings.Contains(view, "2023") {
		t.Errorf("view should contain 2023, got: %s", view)
	}

	if !strings.Contains(view, "2015") {
		t.Errorf("view should contain 2015, got: %s", view)
	}
}

func TestRenderProgressZeroTotal(t *testing.T) {
	result := renderProgress(0, 0, 10)

	if len(result) == 0 {
		t.Error("expected non-empty progress bar")
	}
}

func TestRenderProgressPartial(t *testing.T) {
	result := renderProgress(5, 10, 20)

	if !strings.Contains(result, "\u2588") {
		t.Errorf("expected filled blocks in partial progress, got: %s", result)
	}

	if !strings.Contains(result, "\u2591") {
		t.Errorf("expected empty blocks in partial progress, got: %s", result)
	}
}

func TestRenderProgressFull(t *testing.T) {
	result := renderProgress(10, 10, 20)

	if !strings.Contains(result, "\u2588") {
		t.Errorf("expected filled blocks, got: %s", result)
	}
}

func TestRenderProgressEmpty(t *testing.T) {
	result := renderProgress(0, 10, 20)

	if strings.Contains(result, "\u2588") {
		t.Errorf("expected no filled blocks for zero completed, got: %s", result)
	}
}

func TestShortHelp(t *testing.T) {
	bindings := keys.ShortHelp()

	if len(bindings) == 0 {
		t.Error("expected non-empty ShortHelp bindings")
	}
}

func TestFullHelp(t *testing.T) {
	groups := keys.FullHelp()

	if len(groups) == 0 {
		t.Fatal("expected non-empty FullHelp groups")
	}

	totalBindings := 0
	for _, group := range groups {
		totalBindings += len(group)
	}

	shortCount := len(keys.ShortHelp())
	if totalBindings <= shortCount {
		t.Errorf("FullHelp (%d) should have more than ShortHelp (%d)", totalBindings, shortCount)
	}
}

func TestHelpToggle(t *testing.T) {
	m := populatedModelWithConfig(t)

	if m.help.ShowAll {
		t.Fatal("expected ShowAll to start as false")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model := updated.(Model)

	if !model.help.ShowAll {
		t.Error("expected ShowAll to be true after pressing ?")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model = updated.(Model)

	if model.help.ShowAll {
		t.Error("expected ShowAll to be false after pressing ? again")
	}
}

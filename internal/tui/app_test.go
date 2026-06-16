package tui

import (
	"errors"
	"log/slog"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/internal/tui/discover"
	"github.com/asphaltbuffet/elf/internal/tui/nav"
	"github.com/asphaltbuffet/elf/internal/tui/yearview"
)

// mockScreen is a minimal tea.Model that records calls for testing.
type mockScreen struct {
	initCalled  bool
	initCmd     tea.Cmd
	lastMsg     tea.Msg
	updateModel tea.Model
	updateCmd   tea.Cmd
	viewStr     string
}

func (m *mockScreen) Init() tea.Cmd {
	m.initCalled = true
	return m.initCmd
}

func (m *mockScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.lastMsg = msg
	if m.updateModel != nil {
		return m.updateModel, m.updateCmd
	}
	return m, m.updateCmd
}

func (m *mockScreen) View() string {
	return m.viewStr
}

func ctrlC() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyCtrlC}
}

func TestApp_Init_EmptyStack(t *testing.T) {
	t.Parallel()

	a := App{}
	cmd := a.Init()

	assert.Nil(t, cmd)
}

func TestApp_Init_WithScreen(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{}
	a := App{stack: []tea.Model{screen}}

	a.Init()

	assert.True(t, screen.initCalled)
}

func TestApp_Init_WithScreenCmd(t *testing.T) {
	t.Parallel()

	called := false
	screen := &mockScreen{initCmd: func() tea.Msg {
		called = true
		return nil
	}}
	a := App{stack: []tea.Model{screen}}

	cmd := a.Init()

	require.NotNil(t, cmd)
	cmd()
	assert.True(t, called)
}

func TestApp_Update_WindowSizeMsg_EmptyStack(t *testing.T) {
	t.Parallel()

	a := App{}
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}

	model, cmd := a.Update(msg)
	result := model.(App)

	assert.Equal(t, 80, result.width)
	assert.Equal(t, 24, result.height)
	assert.Nil(t, cmd)
}

func TestApp_Update_WindowSizeMsg_WithScreen(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "hello"}
	a := App{stack: []tea.Model{screen}}
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	model, _ := a.Update(msg)
	result := model.(App)

	assert.Equal(t, 120, result.width)
	assert.Equal(t, 40, result.height)
	assert.Equal(t, msg, screen.lastMsg)
}

func TestApp_Update_PushScreenMsg(t *testing.T) {
	t.Parallel()

	first := &mockScreen{viewStr: "first"}
	second := &mockScreen{viewStr: "second"}
	a := App{stack: []tea.Model{first}}

	model, _ := a.Update(nav.PushScreenMsg{Screen: second})
	result := model.(App)

	assert.Len(t, result.stack, 2)
	assert.True(t, second.initCalled)
	assert.Equal(t, "second", result.View())
}

func TestApp_Update_PopScreenMsg_MultipleScreens(t *testing.T) {
	t.Parallel()

	first := &mockScreen{viewStr: "first"}
	second := &mockScreen{viewStr: "second"}
	a := App{
		stack:  []tea.Model{first, second},
		width:  80,
		height: 24,
	}

	model, _ := a.Update(nav.PopScreenMsg{})
	result := model.(App)

	assert.Len(t, result.stack, 1)
	assert.Equal(t, "first", result.View())

	wsMsg, ok := first.lastMsg.(tea.WindowSizeMsg)
	require.True(t, ok, "expected WindowSizeMsg, got %T", first.lastMsg)
	assert.Equal(t, 80, wsMsg.Width)
	assert.Equal(t, 24, wsMsg.Height)
}

func TestApp_Update_PopScreenMsg_SingleScreen(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "only"}
	a := App{stack: []tea.Model{screen}}

	_, cmd := a.Update(nav.PopScreenMsg{})

	require.NotNil(t, cmd)
	msg := cmd()
	assert.NotNil(t, msg)
}

func TestApp_Update_KeyMsg_CtrlC(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "test"}
	a := App{stack: []tea.Model{screen}}

	_, cmd := a.Update(ctrlC())

	require.NotNil(t, cmd)
	msg := cmd()
	assert.NotNil(t, msg)
}

func TestApp_Update_DefaultRouting(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "test"}
	a := App{stack: []tea.Model{screen}}

	type customMsg struct{ value int }
	msg := customMsg{value: 42}

	_, _ = a.Update(msg)

	assert.Equal(t, msg, screen.lastMsg)
}

func TestApp_Update_DefaultRouting_EmptyStack(t *testing.T) {
	t.Parallel()

	a := App{}

	type customMsg struct{ value int }
	model, cmd := a.Update(customMsg{value: 1})

	assert.Nil(t, cmd)
	assert.Equal(t, App{}, model)
}

func TestApp_Update_ActionMsg(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "dashboard"}
	a := App{stack: []tea.Model{screen}}

	msg := yearview.ActionMsg{
		Action:   "solve",
		Exercise: discover.ExerciseInfo{Year: 2015, Day: 1, Title: "Test", Langs: []string{"go"}},
		App:      nil,
		Lang:     "go",
	}

	model, _ := a.Update(msg)
	result := model.(App)

	assert.Len(t, result.stack, 2)
}

func TestApp_Update_AnalyzeDoneMsg_Error(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "test"}
	a := App{stack: []tea.Model{screen}}

	msg := analyzeDoneMsg{err: errors.New("analysis failed")}

	_, cmd := a.Update(msg)

	assert.Nil(t, cmd)
}

func TestApp_Update_AnalyzeDoneMsg_Success(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "test"}
	a := App{stack: []tea.Model{screen}}

	msg := analyzeDoneMsg{path: "/tmp/test-graph.png"}

	_, cmd := a.Update(msg)

	assert.NotNil(t, cmd)
}

func TestApp_View_EmptyStack(t *testing.T) {
	t.Parallel()

	a := App{}

	assert.Equal(t, "Loading...", a.View())
}

func TestApp_View_WithScreen(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "dashboard content"}
	a := App{stack: []tea.Model{screen}}

	assert.Equal(t, "dashboard content", a.View())
}

func TestApp_View_TopOfStack(t *testing.T) {
	t.Parallel()

	first := &mockScreen{viewStr: "first"}
	second := &mockScreen{viewStr: "second"}
	a := App{stack: []tea.Model{first, second}}

	assert.Equal(t, "second", a.View())
}

func TestApp_Update_KeyMsg_NonCtrlC(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "test"}
	a := App{stack: []tea.Model{screen}}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, _ = a.Update(msg)

	assert.Equal(t, msg, screen.lastMsg)
}

func TestStatusBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status string
		want   string
	}{
		{"PASS", "PASS"},
		{"FAIL", "FAIL"},
		{"ERROR", "ERROR"},
		{"NEW", "NEW"},
		{"UNKNOWN", "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			t.Parallel()
			result := StatusBadge(tt.status)
			assert.Contains(t, result, tt.want)
		})
	}
}

func TestRunAnalyze_BadDir(t *testing.T) {
	t.Parallel()

	cmd := runAnalyze(slog.Default(), t.TempDir())
	require.NotNil(t, cmd)

	msg := cmd()
	done, ok := msg.(analyzeDoneMsg)
	require.True(t, ok)
	assert.Error(t, done.err)
}

func TestApp_Update_AnalyzeMsg(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "test"}
	a := App{stack: []tea.Model{screen}}

	msg := yearview.AnalyzeMsg{YearDir: t.TempDir(), Logger: slog.Default()}
	_, cmd := a.Update(msg)

	assert.NotNil(t, cmd)
}

func TestApp_Update_ImageDisplayDoneMsg(t *testing.T) {
	t.Parallel()

	screen := &mockScreen{viewStr: "test"}
	a := App{stack: []tea.Model{screen}}

	_, cmd := a.Update(imageDisplayDoneMsg{})
	assert.Nil(t, cmd)
}

func TestApp_Update_WindowSizeMsg_PropagatedToTopOnly(t *testing.T) {
	t.Parallel()

	first := &mockScreen{viewStr: "first"}
	second := &mockScreen{viewStr: "second"}
	a := App{stack: []tea.Model{first, second}}

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	_, _ = a.Update(msg)

	assert.Nil(t, first.lastMsg)
	assert.Equal(t, msg, second.lastMsg)
}

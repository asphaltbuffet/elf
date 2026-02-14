package help

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func specialKeyMsg(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

func TestNew(t *testing.T) {
	m := New()

	if m.width != 0 {
		t.Errorf("expected width 0, got %d", m.width)
	}

	if m.height != 0 {
		t.Errorf("expected height 0, got %d", m.height)
	}
}

func TestEscClosesModal(t *testing.T) {
	m := New()

	_, cmd := m.Update(specialKeyMsg(tea.KeyEsc))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for esc")
	}

	msg := cmd()
	if _, ok := msg.(CloseModalMsg); !ok {
		t.Fatalf("expected CloseModalMsg, got %T", msg)
	}
}

func TestQuestionMarkClosesModal(t *testing.T) {
	m := New()

	_, cmd := m.Update(keyMsg("?"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for ?")
	}

	msg := cmd()
	if _, ok := msg.(CloseModalMsg); !ok {
		t.Fatalf("expected CloseModalMsg, got %T", msg)
	}
}

func TestQClosesModal(t *testing.T) {
	m := New()

	_, cmd := m.Update(keyMsg("q"))

	if cmd == nil {
		t.Fatal("expected non-nil cmd for q")
	}

	msg := cmd()
	if _, ok := msg.(CloseModalMsg); !ok {
		t.Fatalf("expected CloseModalMsg, got %T", msg)
	}
}

func TestOtherKeyDoesNotClose(t *testing.T) {
	m := New()

	_, cmd := m.Update(keyMsg("x"))

	if cmd != nil {
		t.Errorf("expected nil cmd for unbound key, got non-nil")
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := New()

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model := updated.(Model)

	if cmd != nil {
		t.Errorf("expected nil cmd for WindowSizeMsg")
	}

	if model.width != 80 {
		t.Errorf("expected width 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("expected height 24, got %d", model.height)
	}
}

func TestInit(t *testing.T) {
	m := New()

	cmd := m.Init()
	if cmd != nil {
		t.Errorf("expected nil from Init(), got non-nil")
	}
}

func TestNonKeyMsgPassesThrough(t *testing.T) {
	m := New()

	type customMsg struct{}

	updated, cmd := m.Update(customMsg{})
	model := updated.(Model)

	if cmd != nil {
		t.Errorf("expected nil cmd for non-key msg")
	}

	// Model should be unchanged
	if model.width != 0 || model.height != 0 {
		t.Errorf("model should be unchanged after non-key msg")
	}
}

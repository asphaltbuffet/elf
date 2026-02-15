package yearview

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/internal/tui/discover"
)

func TestView_Empty(t *testing.T) {
	t.Parallel()
	m := New(zeroCfg, 2015, nil)
	view := m.View()
	assert.Contains(t, view, "No exercises found")
	assert.Contains(t, view, "2015")
}

func TestView_Populated(t *testing.T) {
	t.Parallel()
	m := New(zeroCfg, 2015, testExercises())
	view := m.View()
	assert.Contains(t, view, "Not Called It")
	assert.Contains(t, view, "Inverse")
	assert.Contains(t, view, "Perfectly Spherical")
	assert.Contains(t, view, "Day")
	assert.Contains(t, view, "Title")
}

func TestView_CursorHighlights(t *testing.T) {
	t.Parallel()
	m := New(zeroCfg, 2015, testExercises())
	m.cursor = 1
	view := m.View()
	assert.Contains(t, view, "▸")
}

func TestView_StatusIndicators(t *testing.T) {
	t.Parallel()
	exercises := []discover.ExerciseInfo{
		{Year: 2015, Day: 1, Title: "Solved Both", Path: "p", HasP1: true, HasP2: true, Langs: []string{"go"}},
		{Year: 2015, Day: 2, Title: "Only P1", Path: "p", HasP1: true, HasP2: false, Langs: []string{"py"}},
	}
	m := New(zeroCfg, 2015, exercises)
	view := m.View()
	assert.Contains(t, view, "P1:")
	assert.Contains(t, view, "P2:")
}

func TestView_EmptyLangs(t *testing.T) {
	t.Parallel()
	exercises := []discover.ExerciseInfo{
		{Year: 2015, Day: 1, Title: "No Langs", Path: "p", Langs: nil},
	}
	m := New(zeroCfg, 2015, exercises)
	view := m.View()
	assert.Contains(t, view, "-")
}

func TestView_ContainsHelp(t *testing.T) {
	t.Parallel()
	m := New(zeroCfg, 2015, testExercises())
	view := m.View()
	// Help bar should contain at least one of the key hints
	assert.True(t, strings.Contains(view, "↑/k") || strings.Contains(view, "quit"))
}

func Test_truncateStr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		limit int
		want  string
	}{
		{"short", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"long", "hello world", 8, "hello w…"},
		{"one over", "abcdef", 5, "abcd…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, truncateStr(tt.input, tt.limit))
		})
	}
}

func Test_statusIndicator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ex   discover.ExerciseInfo
		want string
	}{
		{"both solved", discover.ExerciseInfo{HasP1: true, HasP2: true}, "✓"},
		{"none solved", discover.ExerciseInfo{HasP1: false, HasP2: false}, "·"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := statusIndicator(tt.ex)
			assert.Contains(t, s, tt.want)
		})
	}
}

func Test_partStatus(t *testing.T) {
	t.Parallel()
	solved := partStatus(true)
	assert.Contains(t, solved, "✓")

	unsolved := partStatus(false)
	assert.Contains(t, unsolved, "·")
}

func Test_langStyle(t *testing.T) {
	t.Parallel()
	s := langStyle("go,py")
	assert.Contains(t, s, "go,py")
}

func TestShortHelp(t *testing.T) {
	t.Parallel()
	bindings := keys.ShortHelp()
	assert.NotEmpty(t, bindings)
	assert.Len(t, bindings, 6)
}

func TestFullHelp(t *testing.T) {
	t.Parallel()
	groups := keys.FullHelp()
	assert.NotEmpty(t, groups)
	assert.Len(t, groups, 3)

	total := 0
	for _, g := range groups {
		total += len(g)
	}
	assert.Greater(t, total, len(keys.ShortHelp()))
}

func TestHelpToggle(t *testing.T) {
	t.Parallel()
	m := New(zeroCfg, 2015, testExercises())
	assert.False(t, m.help.ShowAll)

	updated, _ := m.Update(keyMsg("?"))
	model := updated.(Model)
	assert.True(t, model.help.ShowAll)

	updated, _ = model.Update(keyMsg("?"))
	model = updated.(Model)
	assert.False(t, model.help.ShowAll)
}

func TestEnterEmpty(t *testing.T) {
	t.Parallel()
	m := New(zeroCfg, 2015, nil)
	_, cmd := m.Update(keyMsg("l"))
	assert.Nil(t, cmd)
}

func TestRightEmpty(t *testing.T) {
	t.Parallel()
	m := New(zeroCfg, 2015, nil)
	_, cmd := m.Update(specialKeyMsg(0)) // unknown msg
	assert.Nil(t, cmd)
}

func TestView_LongTitle(t *testing.T) {
	t.Parallel()
	exercises := []discover.ExerciseInfo{
		{
			Year:  2015,
			Day:   1,
			Title: "A Very Very Long Exercise Title That Exceeds Limit",
			Path:  "p",
			Langs: []string{"go"},
		},
	}
	m := New(zeroCfg, 2015, exercises)
	view := m.View()
	assert.Contains(t, view, "…")
}

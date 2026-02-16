package components

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func Test_NewResultList(t *testing.T) {
	t.Parallel()
	rl := NewResultList(80, 24)
	assert.Equal(t, 0, rl.Count())
	assert.Equal(t, 80, rl.width)
	assert.Equal(t, 24, rl.height)
}

func Test_ResultList_AddResult(t *testing.T) {
	t.Parallel()
	rl := NewResultList(80, 24)
	rl.AddResult(tasks.Result{Status: tasks.StatusPassed, Part: runners.PartOne, SubPart: -1, Duration: 0.5})
	assert.Equal(t, 1, rl.Count())
	rl.AddResult(tasks.Result{Status: tasks.StatusFailed, Part: runners.PartTwo, SubPart: 0, Duration: 1.2})
	assert.Equal(t, 2, rl.Count())
}

func Test_ResultList_SetSize(t *testing.T) {
	t.Parallel()
	rl := NewResultList(80, 24)
	rl.SetSize(120, 40)
	assert.Equal(t, 120, rl.width)
	assert.Equal(t, 40, rl.height)
}

func Test_ResultList_View(t *testing.T) {
	t.Parallel()
	rl := NewResultList(80, 24)
	rl.AddResult(tasks.Result{Status: tasks.StatusPassed, Part: runners.PartOne, SubPart: -1, Duration: 0.001})
	view := rl.View()
	assert.NotEmpty(t, view)
}

func Test_renderOneResult(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		result tasks.Result
		check  func(t *testing.T, s string)
	}{
		{
			name:   "passed",
			result: tasks.Result{Status: tasks.StatusPassed, Part: runners.PartOne, SubPart: -1, Duration: 0.5},
			check:  func(t *testing.T, s string) { t.Helper(); assert.Contains(t, s, "PASS") },
		},
		{
			name: "failed",
			result: tasks.Result{
				Status:   tasks.StatusFailed,
				Part:     runners.PartOne,
				SubPart:  0,
				Duration: 1.0,
				Output:   "wrong answer",
			},
			check: func(t *testing.T, s string) {
				t.Helper()
				assert.Contains(t, s, "FAIL")
				assert.Contains(t, s, "wrong answer")
			},
		},
		{
			name: "error",
			result: tasks.Result{
				Status:   tasks.StatusError,
				Part:     runners.PartTwo,
				SubPart:  0,
				Duration: 0.1,
				Output:   "runtime error",
			},
			check: func(t *testing.T, s string) {
				t.Helper()
				assert.Contains(t, s, "ERROR")
				assert.Contains(t, s, "runtime error")
			},
		},
		{
			name: "unverified",
			result: tasks.Result{
				Status:   tasks.StatusUnverified,
				Part:     runners.PartOne,
				SubPart:  0,
				Duration: 0.2,
				Output:   "new result",
			},
			check: func(t *testing.T, s string) {
				t.Helper()
				assert.Contains(t, s, "NEW")
				assert.Contains(t, s, "new result")
			},
		},
		{
			name:   "invalid",
			result: tasks.Result{Status: tasks.StatusInvalid, Part: runners.PartOne, SubPart: -1},
			check:  func(t *testing.T, s string) { t.Helper(); assert.Contains(t, s, "???") },
		},
		{
			name: "solve passed with output",
			result: tasks.Result{
				Type:     tasks.Solve,
				Status:   tasks.StatusPassed,
				Part:     runners.PartOne,
				SubPart:  -1,
				Duration: 0.3,
				Output:   "42",
			},
			check: func(t *testing.T, s string) {
				t.Helper()
				assert.Contains(t, s, "PASS")
				assert.Contains(t, s, "42")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := renderOneResult(tt.result)
			tt.check(t, s)
		})
	}
}

func Test_statusBadge(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status tasks.TaskStatus
		want   string
	}{
		{tasks.StatusPassed, "PASS"},
		{tasks.StatusFailed, "FAIL"},
		{tasks.StatusError, "ERROR"},
		{tasks.StatusUnverified, "NEW"},
		{tasks.StatusInvalid, "???"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, statusBadge(tt.status), tt.want)
		})
	}

	unknown := statusBadge(tasks.TaskStatus(99))
	assert.Contains(t, unknown, "???")
}

func Test_taskLabel(t *testing.T) {
	t.Parallel()
	t.Run("no subpart", func(t *testing.T) {
		t.Parallel()
		label := taskLabel(runners.PartOne, -1)
		assert.Contains(t, label, "1:")
	})
	t.Run("with subpart", func(t *testing.T) {
		t.Parallel()
		label := taskLabel(runners.PartTwo, 0)
		assert.Contains(t, label, "2.1:")
	})
	t.Run("subpart 2", func(t *testing.T) {
		t.Parallel()
		label := taskLabel(runners.PartOne, 4)
		assert.Contains(t, label, "1.5:")
	})
}

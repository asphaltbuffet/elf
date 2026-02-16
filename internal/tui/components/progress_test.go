package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewProgress(t *testing.T) {
	t.Parallel()
	p := NewProgress()
	assert.False(t, p.Active())
	assert.Empty(t, p.View())
}

func Test_Progress_Start(t *testing.T) {
	t.Parallel()
	p := NewProgress()
	cmd := p.Start("testing")
	assert.True(t, p.Active())
	assert.NotNil(t, cmd)
	assert.Equal(t, "testing", p.operation)
}

func Test_Progress_Stop(t *testing.T) {
	t.Parallel()
	p := NewProgress()
	p.Start("testing")
	assert.True(t, p.Active())
	p.Stop()
	assert.False(t, p.Active())
	assert.Empty(t, p.View())
}

func Test_Progress_View_active(t *testing.T) {
	t.Parallel()
	p := NewProgress()
	p.Start("solving")
	view := p.View()
	assert.Contains(t, view, "solving")
}

func Test_Progress_Update_inactive(t *testing.T) {
	t.Parallel()
	p := NewProgress()
	cmd := p.Update(nil)
	assert.Nil(t, cmd)
}

func Test_StartedProgress(t *testing.T) {
	t.Parallel()
	p := StartedProgress("benchmarking")
	assert.True(t, p.Active())
	assert.Equal(t, "benchmarking", p.operation)
	assert.NotNil(t, p.InitialTick())
}

func Test_Progress_Active(t *testing.T) {
	t.Parallel()
	p := NewProgress()
	assert.False(t, p.Active())
	p.Start("op")
	assert.True(t, p.Active())
	p.Stop()
	assert.False(t, p.Active())
}

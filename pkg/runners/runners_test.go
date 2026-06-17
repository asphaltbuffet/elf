package runners_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func TestExerciseMeta_LangDir(t *testing.T) {
	meta := runners.ExerciseMeta{
		Year:  2015,
		Day:   1,
		Title: "not-quite-lisp",
		Dir:   "/home/user/exercises/2015/01-not-quite-lisp",
		Key:   "py",
	}
	assert.Equal(t, "/home/user/exercises/2015/01-not-quite-lisp/py", meta.LangDir())
}

func TestRegisterFromDescriptors(t *testing.T) {
	restore := runners.ResetRegistry(map[string]runners.RunnerCreator{})
	t.Cleanup(restore)

	descs := []runners.RunnerDescriptor{
		{Key: "py", Name: "Python", Open: runners.OpenSpec{Interpreter: "python3"}},
		{Key: "rb", Name: "Ruby", Open: runners.OpenSpec{Interpreter: "ruby"}},
	}

	runners.RegisterFromDescriptors(descs)

	assert.Contains(t, runners.Available, "py")
	assert.Contains(t, runners.Available, "rb")
	assert.Len(t, runners.Available, 2)
}

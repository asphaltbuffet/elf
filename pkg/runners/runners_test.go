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

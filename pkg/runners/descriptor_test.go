package runners

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubstituteTokens(t *testing.T) {
	meta := ExerciseMeta{
		Year:  2015,
		Day:   1,
		Title: "not-quite-lisp",
		Dir:   "/home/user/exercises/2015/01-not-quite-lisp",
		Key:   "py",
	}

	tests := []struct {
		name       string
		input      string
		wrapperExt string
		want       string
	}{
		{
			name:       "exercise_dir token",
			input:      "{exercise_dir}/input.txt",
			wrapperExt: ".py",
			want:       "/home/user/exercises/2015/01-not-quite-lisp/input.txt",
		},
		{
			name:       "lang_dir token",
			input:      "{lang_dir}/wrapper.py",
			wrapperExt: ".py",
			want:       "/home/user/exercises/2015/01-not-quite-lisp/py/wrapper.py",
		},
		{
			name:       "wrapper_file token",
			input:      "{wrapper_file}",
			wrapperExt: ".py",
			want:       "/home/user/exercises/2015/01-not-quite-lisp/py/runtime-wrapper.py",
		},
		{
			name:       "binary_file token",
			input:      "{binary_file}",
			wrapperExt: "",
			want:       "/home/user/exercises/2015/01-not-quite-lisp/py/runtime-wrapper",
		},
		{
			name:       "year day title tokens",
			input:      "{year}/{day}/{title}",
			wrapperExt: "",
			want:       "2015/01/not-quite-lisp",
		},
		{
			name:       "multiple tokens",
			input:      "github.com/me/aoc/{year}/{day}-{title}/go",
			wrapperExt: "",
			want:       "github.com/me/aoc/2015/01-not-quite-lisp/go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substituteTokens(tt.input, meta, tt.wrapperExt, "")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRunnerDescriptor_ToCreator(t *testing.T) {
	desc := RunnerDescriptor{
		Key:  "py",
		Name: "Python",
		Prepare: PrepareSpec{
			TemplatePath: "/tmp/python.templ",
		},
		Open: OpenSpec{
			Interpreter: "python3",
			Args:        []string{"-B", "{wrapper_file}"},
		},
	}

	creator := desc.ToCreator()
	assert.NotNil(t, creator)

	meta := ExerciseMeta{
		Year: 2015, Day: 1, Title: "foo",
		Dir: "/exercises/2015/01-foo",
	}
	r := creator(meta)
	assert.NotNil(t, r)
	assert.Equal(t, "Python", r.String())
}

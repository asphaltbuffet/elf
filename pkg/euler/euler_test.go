package euler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/euler"
)

func TestNew(t *testing.T) {
	type args struct {
		id   int
		lang string
	}

	tests := []struct {
		name string
		args args
		want *euler.Problem
	}{
		{
			name: "first problem",
			args: args{id: 1, lang: "go"},
			want: &euler.Problem{ID: 1, Language: "go"},
		},
		{
			name: "3-digit problem",
			args: args{id: 857, lang: "go"},
			want: &euler.Problem{ID: 857, Language: "go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, euler.New(tt.args.id, tt.args.lang))
		})
	}
}

func TestProblem_String(t *testing.T) {
	tests := []struct {
		name string
		p    *euler.Problem
		want string
	}{
		{"single digit problem in go", &euler.Problem{ID: 1, Language: "go"}, "Project Euler: 001 (Go)"},
		{"double digit problem in py", &euler.Problem{ID: 69, Language: "py"}, "Project Euler: 069 (Python)"},
		{"triple digit problem in py", &euler.Problem{ID: 666, Language: "py"}, "Project Euler: 666 (Python)"},
		{"invalid language", &euler.Problem{ID: 666, Language: "foo"}, "Project Euler: 666 (INVALID LANGUAGE)"},
		{"empty problem", &euler.Problem{}, "Project Euler: 000 (INVALID LANGUAGE)"},
		{"nil problem", nil, "Project Euler: NIL PROBLEM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.p.String())
		})
	}
}

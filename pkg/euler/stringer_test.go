package euler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func TestProblem_String(t *testing.T) {
	tests := []struct {
		name string
		p    *Problem
		want string
	}{
		{"single digit problem in go", &Problem{
			ID:       1,
			Language: "go",
			Runner:   runners.Available["go"]("foo"),
		}, "Project Euler: 001 (Go)"},
		{"double digit problem in py", &Problem{
			ID:       69,
			Language: "py",
			Runner:   runners.Available["py"]("foo"),
		}, "Project Euler: 069 (Python)"},
		{"triple digit problem in py", &Problem{
			ID:       666,
			Language: "py",
			Runner:   runners.Available["py"]("foo"),
		}, "Project Euler: 666 (Python)"},
		{"invalid language", &Problem{
			ID:       666,
			Language: "foo",
			Runner:   nil,
		}, "Project Euler: 666 (INVALID LANGUAGE)"},
		{"empty problem", &Problem{}, "Project Euler: 000 (INVALID LANGUAGE)"},
		{"nil problem", nil, "Project Euler: NIL PROBLEM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.p.String())
		})
	}
}

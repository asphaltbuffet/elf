package euler_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/euler"
	"github.com/asphaltbuffet/elf/pkg/runners"
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
			want: &euler.Problem{
				ID:       1,
				Language: "go",
				Runner:   runners.Available["go"](filepath.Join("problems", "001", "go")),
			},
		},
		{
			name: "3-digit problem",
			args: args{id: 857, lang: "go"},
			want: &euler.Problem{
				ID:       857,
				Language: "go",
				Runner:   runners.Available["go"](filepath.Join("problems", "857", "go")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, euler.New(tt.args.id, tt.args.lang))
		})
	}
}

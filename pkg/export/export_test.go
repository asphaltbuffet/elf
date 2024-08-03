package export_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/export"
)

func TestGetFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  export.Format
	}{
		{
			name:  "JSON",
			input: "json",
			want:  export.JSON,
		},
		{
			name:  "TOML",
			input: "toml",
			want:  export.TOML,
		},
		{
			name:  "Table",
			input: "table",
			want:  export.Table,
		},
		{
			name:  "CSV",
			input: "csv",
			want:  export.CSV,
		},
		{
			name:  "YAML",
			input: "yaml",
			want:  export.YAML,
		},
		{
			name:  "Text",
			input: "txt",
			want:  export.Text,
		},
		{
			name:  "Unknown",
			input: "unknown",
			want:  export.Invalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want.String(), export.GetFormat(tt.input).String())
		})
	}
}

package exercise

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_extractProblemTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		page    string
		want    string
		wantErr bool
	}{
		{
			name: "single h2 title",
			page: `<html><body><h2>Coded Triangle Numbers</h2><p>body</p></body></html>`,
			want: "Coded Triangle Numbers",
		},
		{
			name: "title with surrounding whitespace is trimmed",
			page: "<html><body><h2>\n  Multiples of 3 or 5  \n</h2></body></html>",
			want: "Multiples of 3 or 5",
		},
		{
			name:    "no h2 means bad number",
			page:    `<html><body><p>no heading here</p></body></html>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := extractProblemTitle([]byte(tt.page))
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidData)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

package runners

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_customWriter_Write(t *testing.T) {
	type args struct {
		b []byte
	}

	type output struct {
		n       int
		entries [][]byte
		pending []byte
	}

	tests := []struct {
		name      string
		c         *customWriter
		args      args
		want      output
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "single line",
			c:    &customWriter{},
			args: args{
				b: []byte("fake\n"),
			},
			want: output{
				n:       5,
				entries: [][]byte{[]byte("fake")},
				pending: nil,
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Write(tt.args.b)
			cw := tt.c

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want.entries, cw.entries)
				assert.Equal(t, tt.want.pending, cw.pending)
				assert.Equal(t, tt.want.n, got)
			}
		})
	}
}

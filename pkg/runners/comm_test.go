package runners

import (
	"bytes"
	"fmt"
	"os/exec"
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

func Test_customWriter_GetEntry(t *testing.T) {
	tests := []struct {
		name      string
		entries   [][]byte
		want      []byte
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{"no entries to get", [][]byte{}, nil, assert.Error, "no entries"},
		{"single entry", [][]byte{[]byte("test data")}, []byte("test data"), assert.NoError, ""},
		{
			"multiple entries",
			[][]byte{
				[]byte("test data 1"),
				[]byte("test data 2"),
				[]byte("fake data 3"),
			},
			[]byte("test data 1"), assert.NoError, "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &customWriter{entries: tt.entries}

			got, err := c.GetEntry()

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText) //nolint:testifylint // error message is not a constant
			} else {
				assert.Equal(t, tt.want, got, fmt.Sprintf("expected %q, got %q", tt.want, got))
				assert.NotContains(t, c.entries, got)
			}
		})
	}
}

func TestSetupBuffers(t *testing.T) {
	tests := []struct {
		name      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "Set buffers correctly",
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &exec.Cmd{}

			got, err := setupBuffers(c)

			tt.assertion(t, err)

			if err == nil {
				assert.IsType(t, &customWriter{}, c.Stdout)
				assert.IsType(t, &bytes.Buffer{}, c.Stderr)
				assert.NotNil(t, got)
			}
		})
	}
}

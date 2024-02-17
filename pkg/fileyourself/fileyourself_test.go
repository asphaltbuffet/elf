package fileyourself

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExt(t *testing.T) {
	type args struct {
		path string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"no dots", args{"fakefile"}, ""},
		{"one dot and extension", args{"fakefile.toml"}, "toml"},
		{"one dot at end", args{"fakefile."}, ""},
		{"one dot at beginning", args{".fakefile"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Ext(tt.args.path))
		})
	}
}

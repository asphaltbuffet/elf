package krampus_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/krampus"
)

func TestConfigKey_String(t *testing.T) {
	tests := []struct {
		name string
		key  krampus.ConfigKey
		want string
	}{
		{"one word", krampus.LanguageKey, "language"},
		{"nested", krampus.AdventDirKey, "advent.dir"},
		{"unknown", "fake", "fake"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.key.String())
		})
	}
}

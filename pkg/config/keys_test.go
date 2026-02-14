package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/config"
)

func TestKey_String(t *testing.T) {
	tests := []struct {
		name string
		key  config.Key
		want string
	}{
		{"one word", config.LanguageKey, "language"},
		{"nested", config.AdventDirKey, "advent.dir"},
		{"unknown", "fake", "fake"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.key.String())
		})
	}
}

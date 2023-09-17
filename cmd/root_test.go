// Package cmd contains all CLI commands used by the application.
package cmd

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_haveValidYearFlag(t *testing.T) {
	tests := []struct {
		name      string
		year      int
		assertion assert.BoolAssertionFunc
	}{
		{"too low", 2014, assert.False},
		{"too high", 2030, assert.False},
		{"lowest good", 2015, assert.True},
		{"higher good", 2022, assert.True},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yearArg = tt.year

			tt.assertion(t, haveValidYearFlag())
		})
	}
}

func Test_haveValidDayFlag(t *testing.T) {
	tests := []struct {
		name      string
		day       int
		assertion assert.BoolAssertionFunc
	}{
		{"too low", 0, assert.False},
		{"too high", 26, assert.False},
		{"lowest good", 1, assert.True},
		{"highest good", 25, assert.True},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dayArg = tt.day

			tt.assertion(t, haveValidDayFlag())
		})
	}
}

func Test_initialize(t *testing.T) {
	type args struct {
		year int
		day  int
		lang string
	}

	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{"good", args{2015, 1, "go"}, assert.NoError},
		{"bad year", args{0, 1, "go"}, assert.Error},
		{"bad day", args{2015, 0, "go"}, assert.Error},
		// {"bad lang", args{2015, 1, "bad"}, assert.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll("testdata", 0o755))
			yearArg, dayArg, langArg = tt.args.year, tt.args.day, tt.args.lang

			tt.assertion(t, initialize(fs))
		})
	}
}

func TestGetRootCommand(t *testing.T) {
	got := GetRootCommand()

	checkCommand(t, got, "elf")
}

func checkCommand(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()

	assert.IsType(t, &cobra.Command{}, cmd)
	assert.NotEmpty(t, cmd)
	assert.NotNil(t, cmd)
	assert.Equal(t, name, cmd.Name())
}

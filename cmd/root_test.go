// Package cmd contains all CLI commands used by the application.
package cmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	type args struct {
		v string
		d string
	}

	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Execute(tt.args.v, tt.args.d)
		})
	}
}

func TestGetRootCommand(t *testing.T) {
	tests := []struct {
		name string
		want *cobra.Command
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRootCommand(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRootCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_initialize(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initialize(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

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

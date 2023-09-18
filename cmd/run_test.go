package cmd

import (
	"testing"
)

func TestGetRunCmd(t *testing.T) {
	got := GetRunCmd()

	checkCommand(t, got, "run")
}

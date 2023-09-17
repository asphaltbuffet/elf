package cmd

import (
	"testing"
)

func TestGetVisualizeCmd(t *testing.T) {
	got := GetVisualizeCmd()

	checkCommand(t, got, "visualize")
}

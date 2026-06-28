package render

import (
	"strings"
	"testing"
)

func TestHeaderStyleRendersTitle(t *testing.T) {
	out := headerStyle("ADVENT OF CODE 2015").Render()
	if !strings.Contains(out, "ADVENT OF CODE 2015") {
		t.Fatalf("header missing title: %q", out)
	}
}

package render

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeaderStyleRendersTitle(t *testing.T) {
	out := headerStyle("ADVENT OF CODE 2015").Render()
	require.Contains(t, out, "ADVENT OF CODE 2015", "header missing title: %q", out)
}

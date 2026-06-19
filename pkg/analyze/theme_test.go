package analyze

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	xfont "golang.org/x/image/font"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
)

func Test_applyTheme(t *testing.T) {
	p := plot.New()
	applyTheme(p)

	t.Run("white background", func(t *testing.T) {
		assert.Equal(t, color.White, p.BackgroundColor)
	})

	t.Run("bold title at title size", func(t *testing.T) {
		assert.Equal(t, xfont.WeightBold, p.Title.TextStyle.Font.Weight)
		assert.InEpsilon(t, float64(vg.Points(titleFontPt)), float64(p.Title.TextStyle.Font.Size), 1e-9)
	})

	t.Run("thin axis lines", func(t *testing.T) {
		assert.InEpsilon(t, float64(vg.Points(axisLineWidthPt)), float64(p.X.LineStyle.Width), 1e-9)
		assert.InEpsilon(t, float64(vg.Points(axisLineWidthPt)), float64(p.Y.LineStyle.Width), 1e-9)
	})

	t.Run("legend has padding", func(t *testing.T) {
		assert.InEpsilon(t, float64(vg.Points(legendPaddingPt)), float64(p.Legend.Padding), 1e-9)
	})
}

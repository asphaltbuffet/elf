package analyze

import (
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"path/filepath"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// redlineColor is the color of the running-time reference line.
func redlineColor() color.Color {
	return color.RGBA{R: 213, G: 94, B: 0, A: 255} //nolint:mnd // palette vermillion
}

// saveGridPNG tiles a rows×cols grid of plots into a single PNG at outfile.
// nil cells are left blank, so callers can omit a (row, col) by leaving it nil.
func saveGridPNG(plots [][]*plot.Plot, outfile string, w, h font.Length, dpi int) error {
	const pad = 12 // points of padding around each tile

	rows := len(plots)
	cols := 0
	if rows > 0 {
		cols = len(plots[0])
	}

	img := vgimg.NewWith(vgimg.UseWH(w, h), vgimg.UseDPI(dpi))
	dc := draw.New(img)

	t := draw.Tiles{
		Rows:      rows,
		Cols:      cols,
		PadX:      vg.Points(pad),
		PadY:      vg.Points(pad),
		PadTop:    vg.Points(pad),
		PadBottom: vg.Points(pad),
		PadLeft:   vg.Points(pad),
		PadRight:  vg.Points(pad),
	}

	canvases := plot.Align(plots, t, dc)
	for r := range plots {
		for c := range plots[r] {
			if plots[r][c] != nil {
				plots[r][c].Draw(canvases[r][c])
			}
		}
	}

	path, err := filepath.Abs(outfile)
	if err != nil {
		return fmt.Errorf("resolving output path: %w", err)
	}

	slog.Info("writing graph", slog.String("path", path)) //nolint:sloglint // standalone function, no logger context

	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("creating image file: %w", err)
	}
	defer f.Close()

	png := vgimg.PngCanvas{Canvas: img}
	if _, err = png.WriteTo(f); err != nil {
		return fmt.Errorf("writing image file: %w", err)
	}

	return nil
}

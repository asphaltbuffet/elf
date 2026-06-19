package analyze

import (
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"path/filepath"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// redlineColor is the color of the running-time reference line.
func redlineColor() color.Color {
	return color.RGBA{R: 213, G: 94, B: 0, A: 255} //nolint:mnd // palette vermillion
}

// savePlotPNG draws a single plot to a PNG file at the given dimensions.
func savePlotPNG(p *plot.Plot, outfile string, w, h font.Length, dpi int) error {
	img := vgimg.NewWith(vgimg.UseWH(w, h), vgimg.UseDPI(dpi))
	dc := draw.New(img)
	p.Draw(dc)

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

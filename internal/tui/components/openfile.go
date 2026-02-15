package components

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/BourgeoisBear/rasterm"
)

// ImageDisplay implements tea.ExecCommand to render an image in the terminal.
// bubbletea's tea.Exec temporarily releases the terminal (exits alt screen),
// lets us write graphics escape sequences, then restores the TUI on return.
type ImageDisplay struct {
	path   string
	stdin  io.Reader
	stdout io.Writer
}

const maxWidth uint32 = 120

// NewImageDisplay creates a tea.ExecCommand that tries terminal graphics
// protocols (Kitty, iTerm2, Sixel) before falling back to the system opener.
func NewImageDisplay(path string) *ImageDisplay {
	return &ImageDisplay{path: path}
}

func (d *ImageDisplay) SetStdin(r io.Reader)  { d.stdin = r }
func (d *ImageDisplay) SetStdout(w io.Writer) { d.stdout = w }
func (d *ImageDisplay) SetStderr(io.Writer)   {}

func (d *ImageDisplay) Run() error {
	if rendered := d.tryTerminalGraphics(); rendered {
		fmt.Fprintln(d.stdout, "\nPress any key to continue...")

		// Wait for a keypress before returning to the TUI.
		buf := make([]byte, 1)
		_, _ = d.stdin.Read(buf)

		return nil
	}

	return openFile(d.path)
}

// tryTerminalGraphics attempts Kitty, iTerm2, then Sixel rendering.
// Returns true if any succeeded.
func (d *ImageDisplay) tryTerminalGraphics() bool {
	img, err := loadPNG(d.path)
	if err != nil {
		return false
	}

	if rasterm.IsKittyCapable() {
		if kittyErr := rasterm.KittyWriteImage(
			d.stdout,
			img,
			rasterm.KittyImgOpts{DstCols: maxWidth},
		); kittyErr == nil {
			return true
		}
	}

	if rasterm.IsItermCapable() {
		if itermErr := rasterm.ItermWriteImage(d.stdout, img); itermErr == nil {
			return true
		}
	}

	if sixelOK, _ := rasterm.IsSixelCapable(); sixelOK {
		paletted := toPaletted(img)
		if sixelErr := rasterm.SixelWriteImage(d.stdout, paletted); sixelErr == nil {
			return true
		}
	}

	return false
}

const paletteBits = 8 // 256 colors for sixel quantization

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return png.Decode(bufio.NewReader(f))
}

// toPaletted converts an image to a paletted format for sixel rendering.
func toPaletted(img image.Image) *image.Paletted {
	bounds := img.Bounds()
	q := MedianCutQuantizer{NumColor: 1 << paletteBits}
	paletted := image.NewPaletted(bounds, nil)
	q.Quantize(paletted, bounds, img, image.Point{})

	return paletted
}

// openFile opens a file with the system's default application.
func openFile(path string) error {
	ctx := context.Background()

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", path)
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", path)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "", path)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", path)
	}

	return cmd.Start()
}

// MedianCutQuantizer implements a simple median-cut color quantizer.
type MedianCutQuantizer struct {
	NumColor int
}

// Quantize fills dst with a paletted version of src.
func (q MedianCutQuantizer) Quantize(dst *image.Paletted, r image.Rectangle, src image.Image, sp image.Point) {
	// Collect all colors from the source image.
	colors := make([]rgbColor, 0, r.Dx()*r.Dy())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			cr, cg, cb, _ := src.At(x+sp.X, y+sp.Y).RGBA()
			colors = append(colors, rgbColor{
				r: uint8(cr >> 8), //nolint:mnd,gosec // 16-bit to 8-bit color; value ≤ 255
				g: uint8(cg >> 8), //nolint:mnd,gosec // 16-bit to 8-bit color; value ≤ 255
				b: uint8(cb >> 8), //nolint:mnd,gosec // 16-bit to 8-bit color; value ≤ 255
			})
		}
	}

	palette := medianCut(colors, q.NumColor)
	dst.Palette = palette

	draw.Draw(dst, r, src, sp, draw.Src)
}

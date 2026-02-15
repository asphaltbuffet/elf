package components

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewImageDisplay(t *testing.T) {
	t.Parallel()
	d := NewImageDisplay("/tmp/test.png")
	require.NotNil(t, d)
	assert.Equal(t, "/tmp/test.png", d.path)
}

func Test_ImageDisplay_SetStdin(t *testing.T) {
	t.Parallel()
	d := NewImageDisplay("/tmp/test.png")
	r := strings.NewReader("hello")
	d.SetStdin(r)
	assert.Equal(t, r, d.stdin)
}

func Test_ImageDisplay_SetStdout(t *testing.T) {
	t.Parallel()
	d := NewImageDisplay("/tmp/test.png")
	var buf bytes.Buffer
	d.SetStdout(&buf)
	assert.Equal(t, &buf, d.stdout)
}

func Test_ImageDisplay_SetStderr(t *testing.T) {
	t.Parallel()
	d := NewImageDisplay("/tmp/test.png")
	d.SetStderr(os.Stderr)
}

func createTestPNG(t *testing.T, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: uint8(x * 10), G: uint8(y * 10), B: 128, A: 255})
		}
	}
	path := filepath.Join(t.TempDir(), "test.png")
	f, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(f, img))
	require.NoError(t, f.Close())
	return path
}

func Test_loadPNG(t *testing.T) {
	t.Parallel()

	t.Run("valid png", func(t *testing.T) {
		t.Parallel()
		path := createTestPNG(t, 4, 4)
		img, err := loadPNG(path)
		require.NoError(t, err)
		require.NotNil(t, img)
		assert.Equal(t, 4, img.Bounds().Dx())
		assert.Equal(t, 4, img.Bounds().Dy())
	})

	t.Run("nonexistent file", func(t *testing.T) {
		t.Parallel()
		_, err := loadPNG("/tmp/nonexistent_test_file_12345.png")
		assert.Error(t, err)
	})
}

func Test_toPaletted(t *testing.T) {
	t.Parallel()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			img.Set(x, y, color.RGBA{R: uint8(x * 30), G: uint8(y * 30), B: 100, A: 255})
		}
	}
	paletted := toPaletted(img)
	require.NotNil(t, paletted)
	assert.Equal(t, image.Rect(0, 0, 8, 8), paletted.Bounds())
	assert.NotEmpty(t, paletted.Palette)
}

func Test_MedianCutQuantizer_Quantize(t *testing.T) {
	t.Parallel()
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			src.Set(x, y, color.RGBA{R: uint8(x * 60), G: uint8(y * 60), B: 128, A: 255})
		}
	}
	q := MedianCutQuantizer{NumColor: 4}
	dst := image.NewPaletted(src.Bounds(), nil)
	q.Quantize(dst, src.Bounds(), src, image.Point{})
	assert.NotEmpty(t, dst.Palette)
	assert.LessOrEqual(t, len(dst.Palette), 4)
}

func Test_tryTerminalGraphics_no_file(t *testing.T) {
	t.Parallel()
	d := NewImageDisplay("/tmp/nonexistent.png")
	var buf bytes.Buffer
	d.SetStdout(&buf)
	assert.False(t, d.tryTerminalGraphics())
}

func Test_tryTerminalGraphics_valid_image(t *testing.T) {
	t.Parallel()
	path := createTestPNG(t, 2, 2)
	d := NewImageDisplay(path)
	var buf bytes.Buffer
	d.SetStdout(&buf)
	// Result depends on terminal capabilities — just verify no panic.
	_ = d.tryTerminalGraphics()
}

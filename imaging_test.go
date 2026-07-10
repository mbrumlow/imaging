package imaging

import (
	"image"
	"image/color"
	"testing"
)

// solid returns a w x h image filled entirely with c.
func solid(w, h int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestCalcBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	tests := []struct {
		x, y, b        int
		sx, sy, ex, ey int
	}{
		{5, 5, 1, 4, 4, 6, 6}, // interior box stays centered
		{0, 0, 1, 0, 0, 1, 1}, // top-left clamps the low side
		{9, 9, 1, 8, 8, 9, 9}, // bottom-right clamps the high side
		{5, 5, 5, 0, 0, 9, 9}, // radius larger than the image clamps both sides
	}

	for _, tt := range tests {
		sx, sy, ex, ey := CalcBounds(img, tt.x, tt.y, tt.b)
		if sx != tt.sx || sy != tt.sy || ex != tt.ex || ey != tt.ey {
			t.Errorf("CalcBounds(_, %d, %d, %d) = (%d,%d,%d,%d), want (%d,%d,%d,%d)",
				tt.x, tt.y, tt.b, sx, sy, ex, ey, tt.sx, tt.sy, tt.ex, tt.ey)
		}
	}
}

func TestCalcLum(t *testing.T) {
	white := solid(3, 3, color.RGBA{255, 255, 255, 255})
	black := solid(3, 3, color.RGBA{0, 0, 0, 255})

	// The luminance coefficients sum to 1, so white is ~255 (allow for
	// float rounding) and black is exactly 0.
	if got := CalcLum(white, 1, 1); got < 254 || got > 255 {
		t.Errorf("white luminance = %d, want 254..255", got)
	}
	if got := CalcLum(black, 1, 1); got != 0 {
		t.Errorf("black luminance = %d, want 0", got)
	}

	// Out-of-range coordinates clamp to the nearest edge pixel rather than
	// panicking or sampling a zero pixel outside the image.
	for _, p := range []image.Point{{-5, -5}, {100, 100}, {-1, 1}, {1, 100}} {
		if got := CalcLum(white, p.X, p.Y); got < 254 || got > 255 {
			t.Errorf("clamped luminance at %v = %d, want 254..255", p, got)
		}
	}
}

func TestGaussianUniformIsNoOp(t *testing.T) {
	c := color.RGBA{100, 150, 200, 255}
	src := solid(8, 8, c)

	// Blurring a uniform image must leave every pixel unchanged, including on
	// the border where the kernel window is clamped.
	out := Gaussian(src, 3)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			r, g, b, _ := out.At(x, y).RGBA()
			if uint8(r>>8) != 100 || uint8(g>>8) != 150 || uint8(b>>8) != 200 {
				t.Fatalf("pixel (%d,%d) = (%d,%d,%d), want (100,150,200)",
					x, y, r>>8, g>>8, b>>8)
			}
		}
	}
}

func TestGaussianZeroCopiesImage(t *testing.T) {
	c := color.RGBA{10, 20, 30, 255}
	src := solid(4, 4, c)

	out := Gaussian(src, 0)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			r, g, b, _ := out.At(x, y).RGBA()
			if uint8(r>>8) != 10 || uint8(g>>8) != 20 || uint8(b>>8) != 30 {
				t.Fatalf("pixel (%d,%d) = (%d,%d,%d), want (10,20,30)",
					x, y, r>>8, g>>8, b>>8)
			}
		}
	}
}

func TestEdgeBlankIsAllBlack(t *testing.T) {
	src := solid(16, 16, color.RGBA{50, 50, 50, 255})

	// A uniform image has no gradients, so no edges should be detected.
	out := Edge(src, 100, 0)
	b := out.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := out.At(x, y).RGBA()
			if r>>8 != 0 || g>>8 != 0 || bl>>8 != 0 {
				t.Fatalf("blank edge pixel (%d,%d) = (%d,%d,%d), want black",
					x, y, r>>8, g>>8, bl>>8)
			}
		}
	}
}

func TestEdgeEmptyImage(t *testing.T) {
	// A zero-sized image must not panic.
	src := image.NewRGBA(image.Rect(0, 0, 0, 0))
	if out := Edge(src, 100, 0); out.Bounds().Dx() != 0 || out.Bounds().Dy() != 0 {
		t.Errorf("empty edge output bounds = %v, want empty", out.Bounds())
	}
}

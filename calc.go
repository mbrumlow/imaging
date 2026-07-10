// Package imaging provides small image-processing primitives: a Gaussian blur
// and Canny edge detection, along with the pixel helpers they are built on.
package imaging

import "image"

// CalcBounds returns the start (sx, sy) and end (ex, ey) coordinates of a
// square window of "radius" b centered on (x, y). The window spans (x-b, y-b)
// to (x+b, y+b), but the returned coordinates are clamped to the image bounds
// so callers can iterate the window without going out of range near an edge.
func CalcBounds(img image.Image, x, y, b int) (int, int, int, int) {

	sx := x - b
	sy := y - b
	ex := x + b
	ey := y + b

	// Clamp the top-left corner to the origin.
	if sx < 0 {
		sx = 0
	}

	if sy < 0 {
		sy = 0
	}

	// Bounds().Max is exclusive, and callers iterate the returned range
	// inclusively (<= ex/ey), so the last valid pixel is Max-1.
	if ex > img.Bounds().Max.X-1 {
		ex = img.Bounds().Max.X - 1
	}

	if ey > img.Bounds().Max.Y-1 {
		ey = img.Bounds().Max.Y - 1
	}

	return sx, sy, ex, ey
}

// CalcLum calculates the perceived luminance of the pixel at (x, y) using the
// Rec. 709 coefficients (0.2126*R + 0.7152*G + 0.0722*B). The result is a
// grayscale intensity in the 0–255 range.
func CalcLum(img image.Image, x, y int) int {

	// image.Color.RGBA returns 16-bit channels; mask off the low byte to work
	// with the 8-bit values this package operates on.
	r, g, b, _ := img.At(x, y).RGBA()
	r &= 0x0000FF
	g &= 0x0000FF
	b &= 0x0000FF
	out := 0.2126*float32(uint8(r)) + 0.7152*float32(uint8(g)) + 0.0722*float32(uint8(b))
	return int(out)

}

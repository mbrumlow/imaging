// Package imaging provides pixel-level calculation helpers for images,
// including bounding a box around a coordinate and computing a pixel's
// luminance.
package imaging

import "image"

// CalcBounds returns the start x,y and end x,y for the box size centered at the x,y coordinates.
// Return coordinates are ensured to be within the images bounds.
func CalcBounds(img image.Image, x, y, b int) (int, int, int, int) {

	sx := x - b
	sy := y - b
	ex := x + b
	ey := y + b

	if sx < 0 {
		sx = 0
	}

	if sy < 0 {
		sy = 0
	}

	if ex > img.Bounds().Max.X-1 {
		ex = img.Bounds().Max.X - 1
	}

	if ey > img.Bounds().Max.Y-1 {
		ey = img.Bounds().Max.Y - 1
	}

	return sx, sy, ex, ey
}

// clamp returns v limited to the inclusive range [lo, hi].
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// CalcLum calculates the luminance of a given pixel. Coordinates that fall
// outside the image bounds are clamped to the nearest edge pixel.
func CalcLum(img image.Image, x, y int) int {

	// Clamp the coordinates so out-of-range values sample the nearest edge
	// pixel instead of a zero-valued (transparent) pixel outside the image.
	bounds := img.Bounds()
	x = clamp(x, bounds.Min.X, bounds.Max.X-1)
	y = clamp(y, bounds.Min.Y, bounds.Max.Y-1)

	// 0.2126*R + 0.7152*G + 0.0722*B
	r, g, b, _ := img.At(x, y).RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	out := 0.2126*float32(uint8(r)) + 0.7152*float32(uint8(g)) + 0.0722*float32(uint8(b))
	return int(out)

}

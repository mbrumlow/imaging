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

	if ex > img.Bounds().Max.X {
		ex = img.Bounds().Max.X
	}

	if ey > img.Bounds().Max.Y {
		ey = img.Bounds().Max.Y
	}

	return sx, sy, ex, ey
}

// CalcLum calculates the luminance of a given pixel.
func CalcLum(img image.Image, x, y int) int {

	// 0.2126*R + 0.7152*G + 0.0722*B
	r, g, b, _ := img.At(x, y).RGBA()
	r &= 0x0000FF
	g &= 0x0000FF
	b &= 0x0000FF
	out := 0.2126*float32(uint8(r)) + 0.7152*float32(uint8(g)) + 0.0722*float32(uint8(b))
	return int(out)

}

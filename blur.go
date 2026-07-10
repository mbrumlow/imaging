package imaging

import "image"

// Gaussian returns a new RGBA image produced by applying a 3x3 Gaussian blur to
// img. Passing n > 1 applies the blur repeatedly, with each pass operating on
// the output of the previous one, which widens the effective blur radius.
//
// As a special case, n == 0 performs a single straight copy (no blur), which is
// what the edge detector requests when blurring is disabled.
func Gaussian(img image.Image, n int) *image.RGBA {

	out := image.NewRGBA(img.Bounds())

	blur := gaussianBlur

	// n == 0 means "don't blur": swap in a kernel-free copy and run one pass.
	if n == 0 {
		blur = func(i image.Image, x, y int) (uint8, uint8, uint8) {
			r, g, b, _ := i.At(x, y).RGBA()
			return uint8(r), uint8(g), uint8(b)
		}
		n = 1
	}

	for i := 0; i < n; i++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {

				r, g, b := blur(img, x, y)

				// Index into the packed RGBA pixel buffer (4 bytes per pixel).
				o := (y*img.Bounds().Max.X + x) * 4

				out.Pix[o+0] = r
				out.Pix[o+1] = g
				out.Pix[o+2] = b
				out.Pix[o+3] = 0xFF // fully opaque

			}
		}

		// Feed this pass's output into the next pass.
		if i+1 < n {
			img = out
			out = image.NewRGBA(img.Bounds())
		}
	}

	return out
}

// gaussianBlur computes the blurred RGB value for the pixel at (x, y) by
// convolving its 3x3 neighborhood with the Gaussian kernel:
//
//	1 2 1
//	2 4 2
//	1 2 1
//
// The kernel weights sum to 16, so the accumulated totals are divided by 16 to
// normalize. Near an image edge CalcBounds clamps the window, so fewer than 9
// samples may contribute.
func gaussianBlur(img image.Image, x, y int) (uint8, uint8, uint8) {

	k := []int{1, 2, 1, 2, 4, 2, 1, 2, 1}

	// c walks the kernel; xgr/xgg/xgb accumulate the weighted channel sums.
	c, xgr, xgg, xgb := 0, 0, 0, 0
	sx, sy, ex, ey := CalcBounds(img, x, y, 1)

	for y := sy; y <= ey; y++ {
		for x := sx; x <= ex; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r &= 0x0000FF
			g &= 0x0000FF
			b &= 0x0000FF
			xgr += k[c] * int(r)
			xgg += k[c] * int(g)
			xgb += k[c] * int(b)
			c++
		}
	}

	// Normalize by the kernel weight total.
	xgr = xgr / 16
	xgg = xgg / 16
	xgb = xgb / 16

	return uint8(xgr), uint8(xgg), uint8(xgb)
}

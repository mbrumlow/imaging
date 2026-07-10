package imaging

import "image"

// Gaussian returns a new image that has been blurred n times with a 3x3
// Gaussian kernel. A larger n produces a stronger blur. When n == 0 the image
// is copied through a single pass without blurring.
func Gaussian(img image.Image, n int) *image.RGBA {

	out := image.NewRGBA(img.Bounds())

	blur := gaussianBlur

	if n == 0 {
		blur = func(i image.Image, x, y int) (uint8, uint8, uint8) {
			r, g, b, _ := i.At(x, y).RGBA()
			return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)
		}
		n = 1
	}

	for i := 0; i < n; i++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {

				r, g, b := blur(img, x, y)
				o := out.PixOffset(x, y)

				out.Pix[o+0] = r
				out.Pix[o+1] = g
				out.Pix[o+2] = b
				out.Pix[o+3] = 0xFF

			}
		}

		if i+1 < n {
			img = out
			out = image.NewRGBA(img.Bounds())
		}
	}

	return out
}

// gaussianBlur computes the blurred R, G, B values for the pixel at x,y by
// convolving it with a 3x3 Gaussian kernel. Coordinates that fall outside the
// image are clamped to the nearest edge pixel (edge replication), so every
// pixel — including those on the border — is convolved with the full kernel and
// stays correctly weighted.
func gaussianBlur(img image.Image, x, y int) (uint8, uint8, uint8) {

	k := []int{1, 2, 1, 2, 4, 2, 1, 2, 1}
	b := img.Bounds()

	c, xgr, xgg, xgb := 0, 0, 0, 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			sx := clamp(x+dx, b.Min.X, b.Max.X-1)
			sy := clamp(y+dy, b.Min.Y, b.Max.Y-1)
			r, g, bl, _ := img.At(sx, sy).RGBA()
			xgr += k[c] * int(r>>8)
			xgg += k[c] * int(g>>8)
			xgb += k[c] * int(bl>>8)
			c++
		}
	}

	return uint8(xgr / 16), uint8(xgg / 16), uint8(xgb / 16)
}

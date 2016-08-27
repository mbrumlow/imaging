package imaging

import "image"

// Gaussian returns a new image has been blured n times with the gaussian blur
func Gaussian(img image.Image, n int) *image.RGBA {

	out := image.NewRGBA(img.Bounds())

	blur := gaussianBlur

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
				o := (y*img.Bounds().Max.X + x) * 4

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

func gaussianBlur(img image.Image, x, y int) (uint8, uint8, uint8) {

	k := []int{1, 2, 1, 2, 4, 2, 1, 2, 1}

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

	xgr = xgr / 16
	xgg = xgg / 16
	xgb = xgb / 16

	return uint8(xgr), uint8(xgg), uint8(xgb)
}
